package certapps

import (
	"encoding/json"
	"errors"
	"fmt"
	//	"html/template"
	"io"
	//"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	textT "text/template"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"appengine/mail"
	"appengine/urlfetch"
	"appengine/user"

	"github.com/mjibson/appstats"
)

type Greeting struct {
	Author  string
	Content string
	Date    time.Time
}

type AuthContext struct {
	LogInOutLink string
	LogInOutText string
	LoggedIn     bool
}

type MainAppContext struct {
	Member *Member
	Team   *Team
	//User   *user.User

	AuthContext
}

type JSONContext struct {
	Data string
}

type ErrorContext struct {
	Message string
	Error   bool
}

type RemindersToSend struct {
	All                 bool
	RespondedYesOrMaybe bool
	Unsent              bool
	NoResponse          bool
}

type Location struct {
	Line1 string
	Line2 string
	City  string
	State string
	Zip   string

	Latitude        float64
	Longitude       float64
	PublicLatitude  float64
	PublicLongitude float64
}

type Audit struct {
	Created    time.Time
	CreatedBy  *datastore.Key
	Modified   time.Time
	ModifiedBy *datastore.Key

	Key   *datastore.Key `datastore:"-"`
	KeyID int64          `datastore:"-"`

	Deleted bool
}

type Team struct {
	Name string
	Location

	MembersEmail  string
	OfficersEmail string

	GoogleAPIKey string

	Audit
}

type TeamMember struct {
	TeamKey   *datastore.Key
	MemberKey *datastore.Key

	Audit
}

type Person struct {
	FirstName string
	LastName  string
	Cell      string
	HomePhone string

	Email  string
	Email2 string

	Location
	UserID string

	Enabled bool

	LastLogin time.Time

	Audit
}

type Member struct {
	ShowCell  bool
	ShowEmail bool
	OKToText  bool

	RadioID string

	HomeAddress *datastore.Key

	//TODO: this would need to move to TeamMember when multiple teams was supported
	CanLookup bool `datastore:"-"`
	Town      bool
	OEM       bool
	Officer   bool
	Active    bool

	Person
}

type MemberLogin struct {
	MemberKey *datastore.Key
	Login     time.Time
	IP        string
}

type Event struct {
	TeamKey *datastore.Key

	Summary     string
	Description string

	Deployment bool
	Exercise   bool
	Meeting    bool
	Training   bool

	MeetTime    time.Time
	EventStart  time.Time
	EventFinish time.Time

	EventLocation   Location
	ParkingLocation Location

	Link1 string //Deployment: IAP, Meeting: Agenda
	Link2 string //Deployment: Roster

	Responses []*MemberEvent `datastore:"-"`

	Audit
}

type MemberEvent struct {
	MemberKey *datastore.Key
	EventKey  *datastore.Key

	Attending bool
	Sure      bool

	FirstReminder  time.Time
	FirstResponded time.Time
	FirstViewed    time.Time
	LastReminder   time.Time
	LastResponded  time.Time
	LastViewed     time.Time
	Arrive         time.Time
	Depart         time.Time

	Audit
}

type Reminder struct {
	MemberEvent *datastore.Key

	Subject string
	Body    string

	Sent time.Time

	Audit
}

type Certification struct {
	Name        string `json:"name"`
	MonthsValid int64  `json:"monthsValid"`

	Audit
}

type CertificationClass struct {
	CertificationKey *datastore.Key
	TeamKey          *datastore.Key

	Name      string    `json:"name"`
	Scheduled time.Time `json:"scheduled"`
	Cancelled bool      `json:"cancelled"`

	Location

	Audit
}

type TrainingTopic struct {
	CertificationKey *datastore.Key

	Name string `json:"name"`

	EffectiveDate time.Time `json:"effectiveDate"`
	SunsetDate    time.Time `json:"sunsetDate"`

	Audit
}

type MemberClassContent struct {
	TrainingTopicKey *datastore.Key
	MemberKey        *datastore.Key

	Attended bool

	Audit
}

type MemberCertification struct {
	CertificationKey *datastore.Key
	MemberKey        *datastore.Key

	Effective time.Time

	Audit
}

type TeamCertification struct {
	TeamKey          *datastore.Key
	CertificationKey *datastore.Key
	Abbr             string
	Hidden           bool

	Audit
}

type CertificationAndTopics struct {
	Certification *Certification
	Topics        []*TrainingTopic
}

// API types

type GoogleLocationResults struct {
	Results []GoogleLocation
	Status  string
}

type GoogleLocation struct {
	Geometry          GoogleLocationGeometry
	Formatted_address string
}

type GoogleLocationGeometry struct {
	Location LatLong
}

type LatLong struct {
	Lat float64
	Lng float64
}

// end types

var jsonTemplate = textT.Must(textT.New("json").Parse("{\"data\": {{.Data}} }"))
var jsonErrTemplate = textT.Must(textT.New("jsonErr").Parse("{\"error\": true, {{.Data}} }"))
var zeroDate = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

var delayEmailReminders2 = delay.Func("sendEmailReminders2", sendEmailReminders)

func init() {
	http.Handle("/api/trainingTopic/save", appstats.NewHandler(apiTrainingTopicSave))
	http.Handle("/api/certificationClass", appstats.NewHandler(apiCertificationClassGet))
	http.Handle("/api/certificationClass/all", appstats.NewHandler(apiCertificationClassGetAll))
	http.Handle("/api/certificationClass/save", appstats.NewHandler(apiCertificationClassSave))

	http.Handle("/api/member/search", appstats.NewHandler(apiMemberSearch))

	http.Handle("/audit", appstats.NewHandler(audit))
	http.Handle("/certification", appstats.NewHandler(certificationGet))
	http.Handle("/certifications/all", appstats.NewHandler(certificationsGetAll))
	http.Handle("/certification/save", appstats.NewHandler(certificationSave))
	http.Handle("/eventA", appstats.NewHandler(eventA))
	http.Handle("/event", appstats.NewHandler(event))
	http.Handle("/event/reminders/send", appstats.NewHandler(remindersSend))
	http.Handle("/event/save", appstats.NewHandler(eventSave))
	http.Handle("/events", appstats.NewHandler(events))
	http.Handle("/member", appstats.NewHandler(memberData))
	http.Handle("/member/save", appstats.NewHandler(memberSave))
	http.Handle("/response", appstats.NewHandler(response))
	http.Handle("/response/save", appstats.NewHandler(responseSave))
	http.Handle("/responses/create", appstats.NewHandler(responsesCreate))
	http.Handle("/team", appstats.NewHandler(teamData))
	http.Handle("/team/roster", appstats.NewHandler(teamRoster))
	http.Handle("/team/roster/import", appstats.NewHandler(membersImport))
	http.Handle("/fix/events/without/team", appstats.NewHandler(fixData))
	http.Handle("/fix/members/geocode", appstats.NewHandler(resaveMembers))
	http.Handle("/address", appstats.NewHandler(locationLookupHandler))
	http.Handle("/setup", appstats.NewHandler(initialSetup))
	http.Handle("/", appstats.NewHandler(root))

	rand.Seed(time.Now().UnixNano())
}

func initialSetup(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	member, errM := getMemberFromUser(c, u, w, r)

	if noErrMsg(errM, w, c, "Loading member") {
		errT, team := getTeam(0, member, c, w, r)

		if noErrMsg(errT, w, c, "Loading team in initialSetup") {
			//addErr := member.addToTeam(team.Key, c)

			//checkErr(addErr, w, c, "Failed adding member to team")

			if team.GoogleAPIKey == "" {
				if team.Key == nil {
					team.Key = datastore.NewKey(c, "Team", "", team.KeyID, nil)
				}
				team.GoogleAPIKey = "xyz"
				datastore.Put(c, team.Key, team)
			}
		}
	}
}

func fixData(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	//u := user.Current(c)

	var query *datastore.Query
	var results []*Event

	var keys []*datastore.Key
	var err error
	var teamKey = datastore.NewKey(c, "Team", "", 6067380039974912, nil)
	var wrongKey = datastore.NewKey(c, "Event", "", 6067380039974912, nil)

	query = datastore.NewQuery("Event").Filter("TeamKey =", wrongKey)

	c.Infof("calling GetAll")
	keys, err = query.GetAll(c, &results)

	if noErrMsg(err, w, c, "Getting Events with no TeamKey") {
		for idx, _ := range results {
			e := results[idx]
			key := keys[idx]

			e.TeamKey = teamKey

			c.Infof("fix team key for id %d, event %+v", key.IntID(), e)

			datastore.Put(c, key, e)
		}
	}

	context := struct {
		Whee bool
	}{false}

	returnJSONorErrorToResponse(context, c, w, r)
}

// guestbookKey returns the key used for all guestbook entries.
func guestbookKey(c appengine.Context) *datastore.Key {
	// The string "default_guestbook" here could be varied to have multiple guestbooks.
	return datastore.NewKey(c, "Guestbook", "default_guestbook", 0, nil)
}

func root(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/static/html/app.htm", http.StatusFound)
}

func audit(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)

	rawurl := r.FormValue("finalURL")
	c.Debugf("rawurl %v", rawurl)
	url, err := url.QueryUnescape(rawurl)

	if u != nil {
		mem, _ := getMemberFromUser(c, u, w, r)

		c.Debugf("Got member: %s", mem.Key.StringID())

		audit := &MemberLogin{
			MemberKey: mem.Key,
			Login:     time.Now(),
			IP:        r.RemoteAddr,
		}

		aKey := datastore.NewKey(c, "Audit", "", 0, nil)

		c.Infof("Putting MemberLogin: %v", aKey)
		_, aErr := datastore.Put(c, aKey, audit)

		if noErrMsg(aErr, w, c, "Trying to put MemberLogin") {
			//l.CreatedBy = outMKey
			//l.ModifiedBy = outMKey

			// c.Infof("Putting Location: %v", lKey)
			// outLKey, lErr := datastore.Put(c, lKey, l)

			// if noErrMsg(lErr, w, c, "Trying to put Location") {
			c.Infof("no error on MemberLogin")

			mem.LastLogin = audit.Login

			_, mErr := datastore.Put(c, mem.Key, mem)

			checkErr(mErr, w, c, "Failed to update Last Login for member: "+mem.Key.StringID())
		}
	}

	if noErrMsg(err, w, c, "could not unescape url") {
		c.Debugf("url %v", rawurl)

		http.Redirect(w, r, url, http.StatusFound)
	}
}

func auth(c appengine.Context, w http.ResponseWriter, r *http.Request) (appengine.Context, *user.User, MainAppContext) {
	u := user.Current(c)
	// context.User = u
	// c.Debugf("context.User %+v", context.User)

	var context MainAppContext

	returnUrl := r.Referer()
	if returnUrl == "" {
		returnUrl = r.URL.String()
	}

	returnUrl = "/audit?finalURL=" + url.QueryEscape(returnUrl)

	if u == nil {
		url, err := user.LoginURL(c, returnUrl)
		if checkErr(err, w, c, "Error getting LoginURL") {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		context.LogInOutLink = url
		context.LogInOutText = "Log In"

		c.Infof("not logged in: %v", url)
	} else {
		//context.Member = getMemberFromUser(c, u, context, w, r)
		context.LoggedIn = true

		url, err := user.LogoutURL(c, returnUrl)
		if checkErr(err, w, c, "Error getting LogoutURL") {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		context.LogInOutLink = url
		context.LogInOutText = "Log Out"
	}

	return c, u, context
}

func updateAudit(audit *Audit, person Person) {
	updateAuditWithKey(audit, person.Key)
}

func updateAuditWithKey(audit *Audit, key *datastore.Key) {
	if audit == nil {
		audit = &Audit{}
	}

	audit.Modified = time.Now()
	audit.ModifiedBy = key

	if audit.CreatedBy == nil || audit.Created.IsZero() {
		audit.Created = audit.Modified
		audit.CreatedBy = audit.ModifiedBy
	}
}

func teamRoster(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	_, u, mainContext := auth(c, w, r)

	context := struct {
		Team    *Team
		Members []Member
	}{}

	if mainContext.LoggedIn {

		var member *Member

		if mainContext.Member != nil {
			member = mainContext.Member
			c.Debugf("using mainContext.Member")
		} else {
			member, _ = getMemberFromUser(c, u, w, r)
		}

		teamKey := r.FormValue("team")
		if teamKey == "" {
			teamKey = "0"
		}

		intKey, _ := strconv.ParseInt(teamKey, 0, 0)
		var getTeamErr error
		var team *Team
		getTeamErr, team = getTeam(intKey, member, c, w, r)

		if noErrMsg(getTeamErr, w, c, "GetTeam with Key: "+teamKey) {
			context.Members = getMembersByTeam(team.Key.IntID(), member, c, w, r)

			context.Team = team
		}
	}

	returnJSONorErrorToResponse(context, c, w, r)

	return
}

func teamData(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	_, u, context := auth(c, w, r)

	if context.LoggedIn {
		member, _ := getMemberFromUser(c, u, w, r)

		teamKey := r.FormValue("team")
		if teamKey == "" {
			teamKey = "0"
		}

		intKey, _ := strconv.ParseInt(teamKey, 0, 0)
		var getTeamErr error
		getTeamErr, context.Team = getTeam(intKey, member, c, w, r)

		if noErrMsg(getTeamErr, w, c, "GetTeam with Key: "+teamKey) {
			if (context.Team == nil || context.Team.Key == nil) && intKey == 0 {
				var audit Audit
				updateAudit(&audit, member.Person)

				context.Team = &Team{
					Name:     "Default CERT Team",
					Location: member.Location,
					Audit:    audit,
				}

				teamKey := datastore.NewKey(c, "Team", "", 0, nil)
				teamOutKey, teamPutError := datastore.Put(c, teamKey, context.Team)

				if noErrMsg(teamPutError, w, c, "Failed putting Default team") {
					context.Team.setKey(teamOutKey)

					updateAudit(&audit, member.Person)

					teamMember := TeamMember{
						MemberKey: member.Key,
						TeamKey:   context.Team.Key,
						Audit:     audit,
					}

					teamMemberKey := datastore.NewKey(c, "TeamMember", "", 0, nil)
					teamMemberOutKey, teamMemberPutError := datastore.Put(c, teamMemberKey, &teamMember)

					if noErrMsg(teamMemberPutError, w, c, "Failed putting TeamMember for Member to Team") {
						teamMember.setKey(teamMemberOutKey)
					}

					c.Infof("Created Default Team: %+v, %+v", context.Team.Key, context.Team)
				}
			}
		}
	}

	//bArr, memberJSONerr := json.MarshalIndent(context, "", "\t")
	returnJSONorErrorToResponse(context, c, w, r)

	return
}

func memberData(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)

	var context MainAppContext

	returnUrl := r.Referer()
	if returnUrl == "" {
		returnUrl = r.URL.String()
	}

	returnUrl = "/audit?finalURL=" + url.QueryEscape(returnUrl)

	if u == nil {
		url, err := user.LoginURL(c, returnUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		context.LogInOutLink = url
		context.LogInOutText = "Log In"

		c.Infof("not logged in: %v", url)
	} else {
		//context.Member = getMemberFromUser(c, u, context, w, r)
		context.LoggedIn = true

		url, err := user.LogoutURL(c, returnUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		context.LogInOutLink = url
		context.LogInOutText = "Log Out"

		member, _ := getMemberFromUser(c, u, w, r)

		memKey := r.FormValue("member")
		if memKey != "" {
			intKey, _ := strconv.ParseInt(memKey, 0, 0)
			context.Member = getMemberByIntKey2(intKey, member, c, w, r)
		} else {
			context.Member = member
		}

		lookupOthers(context.Member)

		context.Member.setKey(context.Member.Key)
	}

	//bArr, memberJSONerr := json.MarshalIndent(context, "", "\t")
	returnJSONorErrorToResponse(context, c, w, r)
}

func returnJSONorErrorToResponse(context interface{}, c appengine.Context, w http.ResponseWriter, r *http.Request) {
	jsonC := JSONContext{}
	bArr, memberJSONerr := json.Marshal(context)
	if noErrMsg(memberJSONerr, w, c, "json.Marshall of Member") {
		//c.Debugf("getting length")
		//n := bytes.Index(bArr, []byte{0})
		n := len(bArr)

		if n > 0 {
			//c.Debugf("getting string for: %d bytes", n)
			jsonC.Data = string(bArr[:n])

			//c.Debugf("jsonTemplate.ExecuteTemplate: %+v", jsonC)
			jsonTemplate.ExecuteTemplate(w, "json", jsonC)
		} else {
			c.Infof("whoops, no bytes in our array m:%d, when marshalling context: %+v", n, context)

			errData := struct{ message string }{"No bytes after Marshalling context"}

			bArr, memberJSONerr = json.Marshal(errData)
			if noErrMsg(memberJSONerr, w, c, "json.Marshall of Member") {
				c.Debugf("getting length for json of %+v", errData)
				//n := bytes.Index(bArr, []byte{0})
				n := len(bArr)

				c.Debugf("length for member JSON bytes %d", n)
				if n > 0 {
					c.Debugf("getting string for: %d bytes", n)
					jsonC.Data = string(bArr[:n])
				} else {
					jsonC.Data = "\"message\": \"could not form error JSON using template\""
				}

				c.Debugf("jsonErrTemplate.ExecuteTemplate: %+v", jsonC)
				jsonErrTemplate.ExecuteTemplate(w, "jsonErr", jsonC)
			}
		}
	}
}

func getMemberFromUser(c appengine.Context, u *user.User, w http.ResponseWriter, r *http.Request) (*Member, error) {

	var memberQ *datastore.Query
	var member []Member
	var keys []*datastore.Key
	var err error
	var retErr error

	memberQ = datastore.NewQuery("Member").Filter("UserID =", u.ID)

	c.Infof("Got membersQ and members for ID: %v, calling GetAll", u.ID)
	keys, err = memberQ.GetAll(c, &member)

	c.Infof("after GetAll")

	var mem *Member

	found := false

	if noErrMsg(err, w, c, "members for id") {
		c.Infof("checking len(members)")
		if len(member) == 0 {
			c.Infof("existing Member not found for user.ID: %v, count:%d", u.ID, len(member))

			memberQ = datastore.NewQuery("Member").Filter("Email =", u.Email)

			c.Infof("Got membersQ and members for Email: %v, calling GetAll", u.Email)
			keys, err = memberQ.GetAll(c, &member)

			if noErrMsg(err, w, c, "members for email") {
				c.Infof("checking len(members)")
				if len(member) == 0 {
					c.Infof("existing Member not found for user.Email: %v, count:%d", u.Email, len(member))

					memberQ = datastore.NewQuery("Member").Filter("Email2 =", u.Email)

					c.Infof("Got membersQ and members for Email: %v, calling GetAll", u.Email)
					keys, err = memberQ.GetAll(c, &member)

					if noErrMsg(err, w, c, "members for email2") {
						if len(member) == 0 {
							c.Infof("existing Member not found for user.Email2: %v, count:%d", u.Email, len(member))

							m := &Member{
								Person: Person{
									FirstName: "",
									LastName:  "",
									Email:     u.Email,
									Cell:      "",

									UserID: u.ID,

									Audit: Audit{
										Created:  time.Now(),
										Modified: time.Now(),
										// need to be nil because we don't have the member key to use yet
										CreatedBy:  nil,
										ModifiedBy: nil,

										Deleted: false,
									},

									Enabled: true,
								},
							}

							//lKey := datastore.NewKey(c, "Location", "", 0, nil)
							mKey := datastore.NewKey(c, "Member", "", 0, nil)

							c.Infof("Putting Member: %v", mKey)
							outMKey, mErr := datastore.Put(c, mKey, m)

							if noErrMsg(mErr, w, c, "Trying to put Member") {

								c.Infof("no error on 1st member put")

								m.CreatedBy = outMKey
								m.ModifiedBy = outMKey

								_, mErr2 := datastore.Put(c, outMKey, m)
								if noErrMsg(mErr2, w, c, "Trying to put Member again") {
									c.Infof("no error on 2nd member put")
								}
								//}
							}

							c.Infof("Member: %v", m)
							mem = m
							found = true
						}
					}
				}

				if len(member) == 1 {
					// found by email
					mem = &member[0]
					mem.setKey(keys[0])
					if mem.UserID == "" {
						found = true
						mem.UserID = u.ID

						c.Infof("Adding User.ID: %v, to Member with Key: %v", mem.UserID, mem.Key)
						datastore.Put(c, mem.Key, mem)
					} else {
						retErr = errors.New("User found by email already associated with another account")
					}
				}
			}
		} else {
			// found by id
			mem = &member[0]
			mem.setKey(keys[0])

			found = true

			if mem.Email == "" {
				mem.Email = u.Email

				_, err = datastore.Put(c, mem.Key, mem)
				_ = checkErr(err, w, c, fmt.Sprintf("Updating Member: %v Email to User value: %v, %+v", mem.Key, u.Email, mem))
			} else {
				c.Debugf("email was already set: %v, not updating to: %v", mem.Email, u.Email)
			}
		}

		if found {
			keyStuff := struct {
				IntID int64
				Key   *datastore.Key
			}{}

			if mem != nil {
				keyStuff.Key = mem.Key
				if mem.Key != nil {
					keyStuff.IntID = mem.Key.IntID()
				}
			}
			c.Infof("existing Member found: %d", len(member))
			c.Debugf("with key: %+v", keyStuff)
			c.Infof("existing: %+v", mem)
			c.Infof("existing.CreatedBy: %+v", mem.CreatedBy)
			c.Infof("user: %+v", *u)

			if mem.Enabled == false {
				retErr = errors.New("Found a Member for the User, but it is disabled")
				c.Errorf(fmt.Sprintf("%s, %+v", retErr, mem))
			}

		} else {
			c.Infof("no mem")
		}
	}

	return mem, retErr
}

func getMemberByKey2(key string, c appengine.Context, w http.ResponseWriter, r *http.Request) *Member {
	c.Debugf("Lookup Member by key: %s", key)

	k := datastore.NewKey(c, "Member", key, 0, nil)
	m := new(Member)
	if err := datastore.Get(c, k, m); err != nil {
		c.Errorf("datastore.Get member error: %v", err)
		return nil
	}
	return m
}

func lookupOthers(member *Member) bool {
	member.CanLookup = member.OEM || member.Town || member.Officer

	return member.CanLookup
}

func getMemberByIntKey2(key int64, currentMem *Member, c appengine.Context, w http.ResponseWriter, r *http.Request) *Member {
	allow := lookupOthers(currentMem)
	c.Debugf("Lookup Member by key: %d, allowed to lookup? %s", key, allow)

	var m Member
	var retval *Member
	retval = nil
	if allow {
		k := datastore.NewKey(c, "Member", "", key, nil)
		if err := datastore.Get(c, k, &m); err != nil {
			c.Errorf("datastore.Get member error: %v", err)

		} else {
			retval = &m
		}
	}
	return retval
}

func getTeam(teamID int64, member *Member, c appengine.Context, w http.ResponseWriter, r *http.Request) (error, *Team) {
	var team Team
	if teamID != 0 {
		teamKey := datastore.NewKey(c, "Team", "", teamID, nil)
		c.Debugf("Calling Get Team with Key: %+v", teamKey)
		getTeamErr := datastore.Get(c, teamKey, &team)
		if noErrMsg(getTeamErr, w, c, "Error Get Team with Key") {
			team.setKey(teamKey)
		} else {
			return fmt.Errorf("No team found for ID passed: %d", teamID), nil
		}

	} else {
		c.Debugf("Calling GetAll team because no teamID supplied")

		var teams []Team = nil
		teamQ := datastore.NewQuery("Team").
			Limit(2)

		keys, getTeamErr := teamQ.GetAll(c, &teams)

		if noErrMsg(getTeamErr, w, c, "Failed while calling GetAll") {
			lenTeams := len(teams)
			if lenTeams == 1 {
				team = teams[0]
				team.setKey(keys[0])
			} else if lenTeams > 1 {
				return fmt.Errorf("Too many Teams found to not specify an ID, only works if there is a single Team"), nil
			}
		}
	}

	if team.Key != nil {
		c.Debugf("Team found with Key: %+v", team.Key)
	} else {
		c.Warningf("No team found search key: %d", teamID)
	}

	return nil, &team
}

func getMembersByTeam(teamID int64, member *Member, c appengine.Context, w http.ResponseWriter, r *http.Request) []Member {
	teamKey := datastore.NewKey(c, "Team", "", teamID, nil)
	teamQ := datastore.NewQuery("TeamMember").
		Filter("TeamKey =", teamKey) //.Project("MemberKey")

	var teamMembers []TeamMember

	c.Debugf("TeamMember TeamKey query: %d, %+v, %+v", teamID, teamKey, teamQ)

	//keys, err
	_, err := teamQ.GetAll(c, &teamMembers)
	if err != nil {
		c.Errorf("fetching members: %v", err)
		return nil
	} else {
	}

	var memberKeys []*datastore.Key
	c.Debugf("looping teamMembers which has: %d", len(teamMembers))
	for _, tm := range teamMembers {
		c.Debugf("teamMember: %+v", tm)
		memberKeys = append(memberKeys, tm.MemberKey)
	}
	members := make([]Member, len(teamMembers))

	c.Debugf("Calling GetMulti with Keys: %+v", memberKeys)

	memberErr := datastore.GetMulti(c, memberKeys, members)

	c.Debugf("memberErr: %+v, %s", memberErr, memberErr.Error())
	if memberErr.Error() != "datastore: no such entity" {
		checkErr(memberErr, w, c, "Error calling GetMulti with Keys")
	}

	lookup := lookupOthers(member)
	for idx := range members {
		m := &members[idx]

		m.setKey(memberKeys[idx])

		if !lookup {
			c.Infof("Doesn't have lookup rights, potentially resetting hidden email/cell")
			if !m.ShowCell {
				c.Debugf("Resetting Cell")
				m.Cell = "555-555-5555"

				c.Debugf("Cell Reset: %+v", m)
			}
			if !m.ShowEmail {
				c.Debugf("Resetting Email")
				m.Email = "good-try@example.com"
			}
			m.Email2 = "good-try@example.com"

			m.Line1 = "123 My Street"
			m.Line2 = "Apt 1"
			m.City = "Anytown"
			m.Zip = "11111"

			m.Latitude = 0.0
			m.Longitude = 0.0
		} else {
			c.Debugf("Has lookup rights, not resetting hidden email/cell")
		}
	}

	return members
}

type MemberSaveContext struct {
	Whee string
}

func memberSave(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var saveMember Member

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&saveMember)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Failed to parse member json from body") {
		c.Infof("JSON mem: %+v", saveMember)
	} else {

	}

	saveMem(&saveMember, nil, c, u, w, r)
}

func resaveMembers(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	c.Debugf("resaveMembers")

	u := user.Current(c)
	var curMember *Member

	if curMember == nil {
		c.Debugf("Looking up curMember because it is nil")
		curMember, _ = getMemberFromUser(c, u, w, r)
	}

	teamKey := r.FormValue("team")
	if teamKey == "" {
		teamKey = "0"
	}

	intKey, _ := strconv.ParseInt(teamKey, 0, 0)

	if intKey > 0 {
		var getTeamErr error
		var team *Team
		getTeamErr, team = getTeam(intKey, curMember, c, w, r)

		if noErrMsg(getTeamErr, w, c, "GetTeam with Key: "+teamKey) {
			members := getMembersByTeam(team.Key.IntID(), curMember, c, w, r)

			var mem *Member
			for idx := range members {
				mem = &members[idx]

				_, err := saveMem(mem, curMember, c, u, w, r)
				checkErr(err, w, c, fmt.Sprintf("Re-Saving member %+v", mem))
			}
		}
	}
}

func saveMem(saveMember *Member, curMember *Member, c appengine.Context, u *user.User, w http.ResponseWriter, r *http.Request) (*datastore.Key, error) {

	if curMember == nil {
		c.Debugf("Looking up curMember because it is nil")
		curMember, _ = getMemberFromUser(c, u, w, r)
	}
	c.Debugf("Looking up allow for curMember")
	allow := lookupOthers(curMember)

	c.Debugf("Will Save Member to Key: %d", saveMember.KeyID)
	saveMember.Key = datastore.NewKey(c, "Member", "", saveMember.KeyID, nil)

	curMember.setKey(curMember.Key)

	c.Debugf("Checking Keys || allow: %s, %d == %d?", allow, curMember.KeyID, saveMember.KeyID)
	if curMember.KeyID == saveMember.KeyID || allow {

		saveMember.locationGeocode(c, w, r)

		c.Debugf("saveMember.Location after geocode: %+v", saveMember.Location)

		saveMember.setAudits(curMember)

		keyOut, mErr := datastore.Put(c, saveMember.Key, saveMember)

		checkErr(mErr, w, c, "Failed to update Member for memberSave: ")
		c.Debugf("Setting saveMember KeyID to keyOut: %+v, %d", keyOut, keyOut.IntID())
		saveMember.setKey(keyOut)

		context := struct {
			Member *Member
		}{
			saveMember,
		}
		returnJSONorErrorToResponse(context, c, w, r)

		return keyOut, mErr
	} else {
		c.Errorf("Only OEM, Town, Officer can update a different Member's record: %+v, tried to save: %+v", curMember, saveMember)
		context := struct {
			error   bool
			message string
		}{
			true,
			"Only OEM, Town, Officer can update a different Member's record",
		}
		returnJSONorErrorToResponse(context, c, w, r)
	}

	return nil, nil
}

func membersImport(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var saveMember Member

	importData := struct {
		TeamID  int64
		Members []Member
	}{}

	c.Debugf("creating decoder")
	decoder := json.NewDecoder(r.Body)
	c.Debugf("decoder.Decode")
	jsonDecodeErr := decoder.Decode(&importData)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Failed to parse member json from body") {
		c.Infof("JSON parsed for import")
	} else {

	}

	curMember, _ := getMemberFromUser(c, u, w, r)
	allow := lookupOthers(curMember)

	if allow {

		for idx := range importData.Members {
			c.Debugf("Importing idx: %d", idx)
			saveMember := importData.Members[idx]

			c.Debugf("Saving Imported Member: %+v", saveMember)
			tk := datastore.NewKey(c, "Team", "", importData.TeamID, nil)

			teamErr, team := getTeam(tk.IntID(), curMember, c, w, r)

			if noErrMsg(teamErr, w, c, "Lookup team") {

				if saveMember.City == "" {
					saveMember.City = team.City
				}

				if saveMember.State == "" {
					saveMember.State = team.State
				}

				if saveMember.Zip == "" {
					saveMember.Zip = team.Zip
				}

				saveMember.Enabled = true

				_, mErr := saveMem(&saveMember, curMember, c, u, w, r)

				if noErrMsg(mErr, w, c, "Save member in import") {
					saveMember.addToTeam(tk, c)
				}
			}
		}

		context := struct {
			Whee bool
		}{
			true,
		}
		returnJSONorErrorToResponse(context, c, w, r)
	} else {
		c.Errorf("Only OEM, Town, Officer can update a different Member's record: %+v, tried to save: %+v", curMember, saveMember)
		context := struct {
			error   bool
			message string
		}{
			true,
			"Only OEM, Town, Officer can update a different Member's record",
		}
		returnJSONorErrorToResponse(context, c, w, r)
	}
}

func (m *Member) addToTeam(teamKey *datastore.Key, c appengine.Context) error {
	tm := &TeamMember{
		TeamKey:   teamKey,
		MemberKey: m.Key,
	}

	tm.setAudits(m)

	tm.Key = datastore.NewKey(c, "TeamMember", "", 0, nil)
	_, tmErr := datastore.Put(c, tm.Key, tm)

	return tmErr
}

func (a *Audit) setAudits(m *Member) {
	a.ModifiedBy = m.Key
	a.Modified = time.Now()

	if a.CreatedBy == nil {
		a.Created = a.Modified
		a.CreatedBy = a.ModifiedBy
	}
}

func (a *Audit) setKey(key *datastore.Key) {
	a.Key = key
	if key != nil {
		a.KeyID = key.IntID()
	}
}

func event(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var event Event
	var context interface{}
	var mem *Member

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&event)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Parsing event json from body") {
		c.Infof("JSON event: %+v", event)
		lookupErr := event.lookup(mem, c, u, w, r)

		c.Debugf("event lookupErr %v", lookupErr)

		if noErrMsg(lookupErr, w, c, "event.lookup") {
			var responsesErr error
			event.Responses, responsesErr = event.responses(nil, mem, c, u, w, r)

			if checkErr(responsesErr, w, c, "Getting responses for event") {
				c.Infof("No errors looking up responses")
			}

			context = struct {
				Event Event
			}{
				event,
			}
		}
	} else {

	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func eventA(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var event Event
	var context interface{}
	var mem *Member

	c.Infof("event-async")

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&event)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Parsing event json from body") {
		c.Infof("JSON event: %+v", event)

		errc := make(chan error)
		lookupC := make(chan Event)
		responsesC := make(chan []*MemberEvent)

		go func() {
			lookupErr := event.lookup(mem, c, u, w, r)
			if noErrMsg(lookupErr, w, c, "event.lookup") {
			}
			lookupC <- event
		}()

		go func() {
			innerResp, errResp := event.responses(nil, mem, c, u, w, r)

			if checkErr(errResp, w, c, "Getting responses for event") {
				c.Infof("No errors looking up responses")
			}

			responsesC <- innerResp
			errc <- errResp
		}()

		event2 := <-lookupC
		responses := <-responsesC

		event.Responses = responses
		event2.Responses = responses

		context = struct {
			EventOrg Event
			Event    Event
		}{
			event,
			event2,
		}
	} else {

	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func response(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var mem *Member
	response := new(MemberEvent)
	var err error

	c.Infof("/response")

	//decoder := json.NewDecoder(r.Body)
	//jsonDecodeErr := decoder.Decode(&response)

	//if jsonDecodeErr == io.EOF {
	//	c.Infof("EOF, should it be?")
	//} else
	//if noErrMsg(jsonDecodeErr, w, c, "Parsing event json from body") {

	responseKey := r.FormValue("response")
	if responseKey != "" {
		response.KeyID, err = strconv.ParseInt(responseKey, 0, 0)
		if noErrMsg(err, w, c, "getting int for MemberEvent.KeyID") {
			c.Infof("JSON response: %+v", response)

			errC := make(chan error)
			//lookupC := make(chan Event)
			//responseC := make(chan *MemberEvent)

			go func() {
				lookupErr := response.lookup(mem, c, u, w, r)

				errC <- lookupErr
			}()

			/*
					go func() {
						innerResp, errResp := event.responses(mem, c, u, w, r)

						if checkErr(errResp, w, c, "Getting responses for event") {
							c.Infof("No errors looking up responses")
						}

						responsesC <- innerResp
						errc <- errResp
					}()
				event2 := <-lookupC
			*/

			err = <-errC

			if noErrMsg(err, w, c, "response.lookup") {

				event := new(Event)
				event.KeyID = response.EventKey.IntID()

				context = struct {
					Response *MemberEvent
					Event    *Event
				}{
					response,
					event,
				}
			}
		}
	} else {
		c.Warningf("/response called with no get ID value passed")
	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func (event *Event) lookup(member *Member, c appengine.Context, u *user.User, w http.ResponseWriter, r *http.Request) error {
	c.Infof("*Event.lookup %d", event.KeyID)
	var err error

	if event.KeyID != 0 {
		event.Key = datastore.NewKey(c, "Event", "", event.KeyID, nil)

		if member == nil {
			member, _ = getMemberFromUser(c, u, w, r)
		}

		err = datastore.Get(c, event.Key, event)

		if err != nil {
			err = event.checkMissingFields(c, err)
		}

		event.setKey(event.Key)
	} else {
		err = errors.New("Must pass an event with a KeyID")
	}

	return err
}

func (event *Event) checkMissingFields(c appengine.Context, err error) error {
	c.Infof("Event.checkMissingFields")

	var efm *datastore.ErrFieldMismatch
	var ok bool

	if err != nil {
		efm, ok = err.(*datastore.ErrFieldMismatch)

		c.Debugf("Was a ErrFieldMismatch?, %s, %s, %s", ok, err, efm)

		if efm != nil {
			c.Debugf("Was a ErrFieldMismatch, %s, %s", efm.FieldName, efm.Reason)

			clear := false

			switch efm.FieldName {
			case "RosterLink":
				clear = true
			case "IAPLink":
				clear = true
				break
			default:

			}

			if clear == true {
				err = nil
				efm = nil
			}
		} else {
			c.Errorf("Error loading Event")
		}
	}

	c.Debugf("checkMissingFields final err value %v", err)

	return err
}

func (me *MemberEvent) lookup(member *Member, c appengine.Context, u *user.User, w http.ResponseWriter, r *http.Request) error {
	c.Infof("*MemberEvent.lookup")
	var err error

	if me != nil {
		if me.KeyID != 0 {
			if me.Key == nil {
				me.Key = datastore.NewKey(c, "MemberEvent", "", me.KeyID, nil)
				c.Debugf("me.Key from ID %d", me.KeyID)
			} else {
				c.Debugf("me.Key was already set %+v", me.Key)
			}

			if member == nil {
				member, _ = getMemberFromUser(c, u, w, r)
			}

			err = datastore.Get(c, me.Key, me)

			me.setKey(me.Key)
		} else {
			err = errors.New("Must pass an MemberEvent with a KeyID")
		}
	} else {
		err = errors.New("lookup Called on nil MemberEvent")
	}

	return err
}

func (reminders *RemindersToSend) addFilters(query *datastore.Query) *datastore.Query {
	if reminders != nil {
		if reminders.All == false {
			if reminders.Unsent {
				query = query.Filter("FirstReminder =", zeroDate)
			}

			if reminders.RespondedYesOrMaybe {
				query = query.Filter("Attending =", true)
			}

			if reminders.NoResponse {
				query = query.Filter("FirstResponded =", zeroDate)
			}
		}
	}

	return query
}

func (event *Event) responses(reminders *RemindersToSend, member *Member, c appengine.Context, u *user.User, w http.ResponseWriter, r *http.Request) ([]*MemberEvent, error) {
	c.Infof("*Event.responses %d", event.KeyID)
	var results []*MemberEvent

	if event.KeyID != 0 {
		event.Key = datastore.NewKey(c, "Event", "", event.KeyID, nil)

		if member == nil {
			member, _ = getMemberFromUser(c, u, w, r)
		}

		query := datastore.NewQuery("MemberEvent").Filter("EventKey =", event.Key)
		var q2 *datastore.Query

		if reminders != nil {
			c.Infof("adding filters for: %+v", reminders)
			c.Infof("times %d, %s, %+v", time.Time{}, time.Time{}, time.Time{})
			q2 = reminders.addFilters(query)
		} else {
			q2 = query
		}

		c.Infof("calling GetAll on query %+v", q2)
		keys, err := q2.GetAll(c, &results)

		if noErrMsg(err, w, c, "Getting MemberEvent") {
			for idx, _ := range results {
				me := results[idx]
				key := keys[idx]

				me.setKey(key)
			}
		}

		return results, err
	}

	return nil, errors.New("Must pass an event with a KeyID")
}

func eventSave(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	jsonData := struct {
		Event *Event
		Team  *Team
	}{}

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&jsonData)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Parsing event and team json from body") {
		c.Infof("JSON event and team: %+v", jsonData)

		newEvent := (jsonData.Event.KeyID == 0)
		saveErr := jsonData.Event.save(nil, jsonData.Team, c, u, w, r)

		if saveErr != nil {
			context := struct {
				error   bool
				message string
			}{
				true,
				"Only OEM, Town, Officer can update an Event",
			}
			returnJSONorErrorToResponse(context, c, w, r)
		}

		if newEvent || 1 == 1 {
			responsesGenerate(jsonData.Event, jsonData.Team, u, c, w, r)
		}
	} else {
		c.Infof("partial JSON event: %+v", event)
	}

	context := struct {
		*Event
	}{
		jsonData.Event,
	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func (event *Event) save(member *Member, team *Team, c appengine.Context, u *user.User, w http.ResponseWriter, r *http.Request) error {
	var err error

	c.Infof("*Event.save %d, for team: %d", event.KeyID, team.KeyID)

	event.Key = datastore.NewKey(c, "Event", "", event.KeyID, nil)

	if member == nil {
		member, _ = getMemberFromUser(c, u, w, r)
	}

	if lookupOthers(member) {

		if team.hasMember(member, c, u, w, r) {
			event.setAudits(member)

			var key *datastore.Key

			event.TeamKey = team.Key
			event.Deleted = false

			c.Infof("Will put event ID: %d", event.KeyID)

			key, err = datastore.Put(c, event.Key, event)

			if noErrMsg(err, w, c, "Put event") {
				c.Infof("Set Key: %d, %d", event.KeyID, key.IntID())
				event.setKey(key)
			}
		} else {
			c.Errorf("You must be a member of a team to save an event record: %+v, tried to save: %+v", member, event)

			err = errors.New("You must be a member of a team to save an event record")
		}
	} else {
		c.Errorf("Only OEM, Town, Officer can update an event record: %+v, tried to save: %+v", member, event)

		err = errors.New("Only Town, OEM, Officers can save events")
	}

	return err
}

func events(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var team *Team
	var events []*Event
	var eventsErr error
	var context interface{}
	var responses []*MemberEvent

	team = &Team{}
	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(team)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Parsing team json from body") {
		c.Infof("JSON team: %+v", team)
		eventsErr, events, responses = team.events(nil, c, u, w, r)

		if noErrMsg(eventsErr, w, c, "Looking up team.events") {
			context = struct {
				Events    []*Event
				Responses []*MemberEvent
			}{
				events,
				responses,
			}
		} else {
			context = errorContextError(eventsErr)
		}

	} else {

	}

	returnJSONorErrorToResponse(context, c, w, r)

}

func (team *Team) events(member *Member, c appengine.Context, u *user.User, w http.ResponseWriter, r *http.Request) (error, []*Event, []*MemberEvent) {
	var eventList []*Event
	var eventKeys []*datastore.Key
	var err error
	var events []*Event
	var keys []*datastore.Key
	var responseList []*MemberEvent

	c.Infof("*Team.events %d", team.KeyID)

	if member == nil {
		member, _ = getMemberFromUser(c, u, w, r)
	}

	if team.KeyID != 0 {
		team.Key = datastore.NewKey(c, "Team", "", team.KeyID, nil)

		now := time.Now()
		histDate := now.AddDate(0, -2, 0)

		eventQ := datastore.NewQuery("Event").Filter("TeamKey =", team.Key).Filter("EventStart >=", histDate)

		c.Infof("Got eventQ and events for ID: %v, calling GetAll", team.KeyID)
		keys, err = eventQ.GetAll(c, &events)

		var storeErr error
		if err != nil {
			c.Debugf("Had an error loading events, looping events, err: %+v", err)
			for idx := range events {
				e := events[idx]

				newErr := e.checkMissingFields(c, err)

				c.Debugf("newErr %s, %s", newErr, err)

				if newErr != nil {
					c.Debugf("setting storeErr %s", newErr)
					storeErr = newErr
				}
			}
		}

		c.Debugf("storeErr %s", storeErr)
		if storeErr == nil {
			c.Debugf("reseting err = nil because storeErr %s", storeErr)

			err = nil
		}

		if noErrMsg(storeErr, w, c, "Calling GetAll for events by Team and EventStart") {
			c.Infof("Found: %d, Events for Team: %d", len(events), team.KeyID)

			errc := make(chan error)
			doneC := make(chan bool)
			responsesC := make(chan []*MemberEvent)

			for idx := range events {
				e := events[idx]
				k := keys[idx]

				e.setKey(k)

				eventList = append(eventList, e)

				if lookupOthers(member) {
					//var responses []MemberEvent
					c.Debugf("Finding Event Responses Calling GetMulti with Keys: %+v", eventKeys)

					go func() {
						var responses []*MemberEvent
						var responseList []*MemberEvent
						responseQ := datastore.NewQuery("MemberEvent").Filter("EventKey =", e.Key)

						c.Infof("Got responseQ for ID: %v, calling GetAll", e.KeyID)
						keys, memberEventsErr := responseQ.GetAll(c, &responses)

						if noErrMsg(memberEventsErr, w, c, "Calling GetMulti for MemberEvents for Event") {
							c.Infof("Found: %d, Responses for Event: %d", len(responses), e.KeyID)

							for idx := range responses {
								r := responses[idx]
								k := keys[idx]

								r.setKey(k)

								c.Debugf("Got a response in go routine %+v", r)

								//panic("Need to make this pass back slices not individual so counts range counts will match up also need to return something, empty slice? nil? if no responses are found")
								responseList = append(responseList, r)
							}

							c.Debugf("sending on channel list of %d responses", len(responseList))
							responsesC <- responseList
						} else {
							c.Errorf("Error in Response lookup: %v", memberEventsErr)
							errc <- memberEventsErr
						}
						//doneC <- true
					}()

				} else {
					c.Infof("Member cannot lookup, not sending Responses: %+v", member)
					responsesC <- nil
				}
			}

			for idx := range events {
				c.Debugf("select idx %d", idx)
				select {
				case response := <-responsesC:
					c.Debugf("received from channel list of %d responses", len(response))
					responseList = append(responseList, response...)
				case done := <-doneC:
					c.Debugf("done from channel")
					c.Infof("Done: %v", done)
				}
			}
		}
	}

	return err, eventList, responseList
}

func (team *Team) hasMember(member *Member, c appengine.Context, u *user.User, w http.ResponseWriter, r *http.Request) bool {
	c.Infof("Team.hasMember team: %d, member: %d", team.KeyID, member.KeyID)

	retval := false
	var teamMembers []TeamMember

	team.Key = datastore.NewKey(c, "Team", "", team.KeyID, nil)
	member.Key = datastore.NewKey(c, "Member", "", member.KeyID, nil)

	q := datastore.NewQuery("TeamMember").Filter("TeamKey =", team.Key).Filter("MemberKey =", member.Key).KeysOnly()

	c.Infof("Got q and events for team ID: %d, member ID: %d, calling GetAll", team.KeyID, member.KeyID)
	keys, err := q.GetAll(c, &teamMembers)

	if noErrMsg(err, w, c, "Looking for TeamMember records") {
		retval = len(keys) == 1
	}

	return retval
}

func responseSave(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var mem *Member
	var postData struct {
		Response *MemberEvent
		Event    *Event
	}

	c.Infof("responseSave")

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&postData)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Parsing event json from body") {
		c.Infof("JSON from request: Response: %+v, Event: %+v", postData.Response, postData.Event)

		postData.Response.Key = datastore.NewKey(c, "MemberEvent", "", postData.Response.KeyID, nil)

		c.Debugf("using MemberEvent KeyID: %d, Key: %+v", postData.Response.KeyID, postData.Response.Key)

		if postData.Response.FirstReminder == zeroDate {
			postData.Response.FirstResponded = time.Now()
		}

		postData.Response.EventKey = datastore.NewKey(c, "Event", "", postData.Event.KeyID, nil)
		postData.Response.LastResponded = time.Now()

		mem, _ = getMemberFromUser(c, u, w, r)

		postData.Response.MemberKey = mem.Key

		postData.Response.setAudits(mem)

		newKey, responseSaveErr := datastore.Put(c, postData.Response.Key, postData.Response)
		if noErrMsg(responseSaveErr, w, c, "Saving MemberEvent") {
			postData.Response.setKey(newKey)
		}

		context = struct {
			Response *MemberEvent
		}{
			postData.Response,
		}
	} else {

	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func responsesCreate(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var postData struct {
		Event *Event
	}

	c.Infof("responsesCreate")

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&postData)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Parsing event json from body") {
		c.Infof("JSON from request: %+v", postData)

		responsesGenerate(postData.Event, nil, u, c, w, r)

		context = struct {
			Whee bool
		}{
			true,
		}
	} else {

	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func responsesGenerate(event *Event, team *Team, u *user.User, c appengine.Context, w http.ResponseWriter, r *http.Request) {
	if event.KeyID != 0 {
		var responses []*MemberEvent
		var roster []*TeamMember

		if team == nil {
			c.Errorf("Need to lookup Team")
		}

		mem, _ := getMemberFromUser(c, u, w, r)

		if lookupOthers(mem) {
			done := make(chan error)
			respMap := make(map[int64]*MemberEvent)

			event.Key = datastore.NewKey(c, "Event", "", event.KeyID, nil)

			meQ := datastore.NewQuery("MemberEvent").Filter("EventKey =", event.Key).Project("MemberKey")

			c.Infof("Got q and responses for event ID: %d, calling GetAll", event.KeyID)

			go func() {
				_, err := meQ.GetAll(c, &responses)

				if noErrMsg(err, w, c, "GetAll MemberEvent") {
					c.Infof("Found %d responses for Event ID: %d", len(responses), event.KeyID)
				}

				done <- err
			}()

			team.Key = datastore.NewKey(c, "Team", "", team.KeyID, nil)
			mQ := datastore.NewQuery("TeamMember").Filter("TeamKey =", team.Key).Project("MemberKey")

			c.Infof("Got q and responses for event ID: %d, calling GetAll", event.KeyID)

			go func() {
				_, err := mQ.GetAll(c, &roster)

				if noErrMsg(err, w, c, "GetAll MemberEvent") {
					c.Infof("Found %d Members for Team ID: %d", len(roster), team.KeyID)
				}

				done <- err
			}()

			success := true

			for i := 0; i < 2; i++ {
				if err := <-done; err != nil {
					success = false
				}
			}

			if success {
				c.Infof("In success")
				emptyKey := datastore.NewKey(c, "MemberEvent", "", 0, nil)

				for idx := range responses {
					r := responses[idx]
					key := r.MemberKey.IntID()
					respMap[key] = r
					c.Debugf("Response item %d, %d, %+v", idx, key, r)
				}

				for idx := range roster {
					m := roster[idx]
					key := m.MemberKey.IntID()

					item := respMap[key]

					found := (item != nil)
					c.Debugf("Roster item found? %v, %d, %d, %+v", found, idx, key, m)

					if found == false {
						newMe := new(MemberEvent)
						newMe.MemberKey = m.MemberKey
						newMe.EventKey = event.Key
						newMe.setAudits(mem)

						outKey, putErr := datastore.Put(c, emptyKey, newMe)

						if noErrMsg(putErr, w, c, "Creating new MemberEvent") {
							c.Infof("Created new MemberEvent with ID: %d, for Event ID: %d, Member ID: %d", outKey.IntID(), event.KeyID, key)
						}
					}
				}
			}
		}
	}

}

func remindersSend(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var mem *Member
	var postData struct {
		Event           Event
		RemindersToSend RemindersToSend
		Team            Team
	}

	c.Infof("remindersSend")

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&postData)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Parsing event json from body") {
		c.Infof("JSON from request: %+v", postData)

		mem, _ = getMemberFromUser(c, u, w, r)

		var responses []*MemberEvent
		var errChan = make(chan error)
		go func() {
			var err error
			responses, err = postData.Event.responses(&postData.RemindersToSend, mem, c, u, w, r)

			errChan <- err
		}()

		go func() {
			var err error
			err = postData.Event.lookup(mem, c, u, w, r)

			errChan <- err
		}()

		var members []Member
		go func() {
			var err error
			members = getMembersByTeam(postData.Team.KeyID, mem, c, w, r)

			errChan <- err
		}()

		var errOut error
		for err := range errChan {
			if checkErr(err, w, c, "Getting reminders that need to be sent") {
				errOut = err
			}
		}

		memberByKey := make(map[int64]Member)

		for _, m := range members {
			memberByKey[m.KeyID] = m
		}

		if noErr(errOut, w, c) {
			for _, r := range responses {
				c.Debugf("Sending reminder: %d, %s, %+v", r.FirstResponded, r.FirstResponded, r)

				responseMember := memberByKey[r.MemberKey.IntID()]

				respond := postData.Event.Deployment || postData.Event.Exercise

				delayEmailReminders2.Call(c, postData.Event, r, responseMember, respond, *mem)
			}
		}

		context = struct {
			Response *MemberEvent
		}{
			nil,
		}
	} else {

	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func (event *Event) typeString() string {
	retval := "Event"

	if event.Deployment {
		retval = "Deployment"
	}

	if event.Meeting {
		if strings.Contains(strings.ToLower(event.Summary), "meeting") {
			retval = ""
		} else {
			retval = "Meeting"
		}
	}

	if event.Exercise {
		retval = "Exercise"
	}

	return retval
}

const reminderEmailText = `
%s

Time:
%s

Please respond with your availability using this link: %s

`

func (event *Event) timeInfo() string {
	const layout = "January 2, 2006 at 3:04pm"
	return event.EventStart.Format(layout)
}

func (response *MemberEvent) getURL(c appengine.Context) string {
	baseURL := appengine.DefaultVersionHostname(c)

	proto := "http"
	if appengine.IsDevAppServer() == false {
		proto = "https"
	}
	return fmt.Sprintf("%s://%s/response/%d", proto, baseURL, response.KeyID)
}

func (response *MemberEvent) sentFirstReminder() bool {
	return zeroDate.Equal(response.FirstReminder) == false
}

func (response *MemberEvent) save(c appengine.Context, member *Member) error {
	response.Key = datastore.NewKey(c, "MemberEvent", "", response.KeyID, nil)
	response.setAudits(member)

	key, err := datastore.Put(c, response.Key, response)

	response.setKey(key)

	return err
}

func sendEmailReminders(c appengine.Context, event Event, memberEvent *MemberEvent, emailMember Member, pleaseRespond bool, curMember Member) {
	msg := &mail.Message{
		Sender:  "CERT Apps <mdragon+cert-apps@gmail.com>",
		To:      []string{"mdragon+cert-apps-cc@gmail.com"},
		Subject: "CERT " + event.typeString() + " - " + event.Summary + " - Please Respond",
		Body:    fmt.Sprintf(reminderEmailText, event.Description, event.timeInfo(), memberEvent.getURL(c)),
	}

	if 1 == 2 || emailMember.Email == "mdragon@gmail.com" {
		msg.To = []string{emailMember.Email, emailMember.Email2}
		msg.Bcc = []string{"mdragon+cert-apps-cc@gmail.com"}
	}

	c.Debugf("Sending email: %+v", msg)
	if err := mail.Send(c, msg); err != nil {
		c.Errorf("Couldn't send email: %v, %+v", err, msg)
	} else {

		memberEvent.LastReminder = time.Now()
		if memberEvent.sentFirstReminder() {
			memberEvent.FirstReminder = memberEvent.LastReminder
		}

		memberEvent.save(c, &curMember)
	}
}

func errorContextString(message string) ErrorContext {
	context := ErrorContext{
		Message: message,
		Error:   true,
	}

	return context
}

func errorContextError(err error) ErrorContext {
	context := ErrorContext{
		Message: fmt.Sprintf("%s", err),
		Error:   true,
	}

	return context
}

func locationLookupHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	address := r.FormValue("address")
	if address == "" {
		err := errors.New("Must pass address as query string param")
		checkErr(err, w, c, "")
	} else {
		gl, err := geocode(address, c, w, r)

		if noErrMsg(err, w, c, "Geocode of address failed") {
			c.Infof("GoogleLocation %+v", gl)

			context := struct {
				Location *GoogleLocationResults
			}{
				gl,
			}
			returnJSONorErrorToResponse(context, c, w, r)
		}
	}
}

func getGoogleAPIKey(c appengine.Context, w http.ResponseWriter, r *http.Request) (string, error) {
	u := user.Current(c)
	mem, _ := getMemberFromUser(c, u, w, r)

	teamErr, team := getTeam(0, mem, c, w, r)

	if teamErr != nil {
		return "", teamErr
	}

	key := team.GoogleAPIKey

	c.Debugf("Found GoogleAPIKey %+v", team)

	return key, nil
}

func geocode(address string, c appengine.Context, w http.ResponseWriter, r *http.Request) (*GoogleLocationResults, error) {
	apiURL := "https://maps.googleapis.com/maps/api/geocode/json"
	key, keyErr := getGoogleAPIKey(c, w, r)

	if keyErr == nil {
		v := url.Values{}
		v.Set("key", key)
		v.Add("address", address)

		fullURL := fmt.Sprintf("%s?%s", apiURL, v.Encode())

		c.Debugf("full: %s, url: %s, key: %s, address: %s", fullURL, apiURL, key, address)

		client := urlfetch.Client(c)
		resp, err := client.Get(fullURL)

		if err != nil {
			return nil, err
		}
		c.Debugf("HTTP GET returned status %v, for url %s", resp.Status, fullURL)
		//c.Debugf("Header: %+v", resp.Header)

		var results GoogleLocationResults

		decoder := json.NewDecoder(resp.Body)

		jsonDecodeErr := decoder.Decode(&results)
		resp.Body.Close()

		if jsonDecodeErr != nil {
			return nil, jsonDecodeErr
		}

		c.Infof("json results %+v", results)

		return &results, nil
	} else {
		return nil, keyErr
	}
}

func (m *Member) locationGeocode(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	m.Longitude = 0
	m.Latitude = 0
	m.PublicLatitude = 0
	m.PublicLongitude = 0

	if m.Line1 != "" {
		m.Location.geocode(c, w, r)
	}

	c.Debugf("m.Location after geocode: %+v", m.Location)
}

func (l *Location) geocode(c appengine.Context, w http.ResponseWriter, r *http.Request) error {
	address := fmt.Sprintf("%s, %s, %s", l.Line1, l.City, l.Zip)

	googleLoc, err := geocode(address, c, w, r)

	if err == nil {
		l.Latitude = googleLoc.Results[0].Geometry.Location.Lat
		l.Longitude = googleLoc.Results[0].Geometry.Location.Lng

		l.PublicLatitude = fudgeGPS(l.Latitude, c)
		l.PublicLongitude = fudgeGPS(l.Longitude, c)
	}

	c.Debugf("Location after geocode: %+v", l)

	return err
}

func fudgeGPS(in float64, c appengine.Context) float64 {
	alter := rand.Float64() / 1000
	plusMinus := rand.Float32()

	abs := 1.0
	if plusMinus > 0.5 {
		abs = -1.0
	}

	if alter < 0.0001 {
		alter = 0.0001
	}

	alter = alter * abs

	out := in + alter

	c.Debugf("alter %f, abs %f, out %f, in %f", alter, abs, out, in)

	return out
}

func certificationSave(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var mem *Member
	var postData struct {
		Certification *Certification
	}

	c.Infof("certificationSave")

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&postData)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Parsing event json from body") {
		c.Infof("JSON from request: %+v", postData)

		mem, _ = getMemberFromUser(c, u, w, r)

		saveErr := postData.Certification.save(mem, c)

		c.Debugf("Certification after save: %+v", postData.Certification)

		if noErrMsg(saveErr, w, c, "Saving Certification") {
			context = struct {
				Certification *Certification
			}{
				postData.Certification,
			}
		}
	} else {

	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func certificationGet(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var mem *Member

	key := r.FormValue("id")

	if key == "" {
		key = "0"
	}

	id, _ := strconv.ParseInt(key, 0, 0)

	c.Infof("certificationGet %d", id)

	mem, _ = getMemberFromUser(c, u, w, r)

	results, err := getCertAndTopics(c, id, mem)

	if noErrMsg(err, w, c, "Certification lookup") {

		context = results
	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func certificationsGetAll(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var mem *Member

	c.Infof("certificationsGet")

	mem, _ = getMemberFromUser(c, u, w, r)

	if mem != nil {

		results, err1, err2 := getAllCertsAndTopics(c)

		if noErrMsg(err1, w, c, "GetAll Certifications") {
			if noErrMsg(err2, w, c, "GetAll Certifications") {
				context = struct {
					Certifications []*CertificationAndTopics
				}{
					results,
				}
			}
		}
	} else {
		context = struct {
			NotLoggedIn bool
		}{
			false,
		}
	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func getAllCertsAndTopics(c appengine.Context) ([]*CertificationAndTopics, error, error) {
	errC := make(chan error)
	topicC := make(chan []*TrainingTopic)
	certC := make(chan []*Certification)

	go func() {
		query := datastore.NewQuery("Certification").Filter("Deleted =", false)

		var lookupResults []*Certification
		keys, err := query.GetAll(c, &lookupResults)

		for idx, _ := range lookupResults {
			e := lookupResults[idx]
			key := keys[idx]

			c.Debugf("found Certification %d, %d, %d", idx, e.KeyID, key.IntID())

			e.setKey(key)
		}
		certC <- lookupResults
		errC <- err
	}()

	certs := <-certC

	go func() {
		query := datastore.NewQuery("TrainingTopic").Filter("Deleted =", false)

		var lookupResults []*TrainingTopic
		keys, err := query.GetAll(c, &lookupResults)

		for idx, _ := range lookupResults {
			e := lookupResults[idx]
			key := keys[idx]

			c.Debugf("found TrainingTopic %d, %d, %d", idx, e.KeyID, key.IntID())
			c.Debugf("\t for cert %d", e.CertificationKey.IntID())

			e.setKey(key)
		}
		topicC <- lookupResults
		errC <- err
	}()

	topics := <-topicC

	err1 := <-errC
	err2 := <-errC

	return processCertsAndTopics(c, certs, topics), err1, err2
}

func getCertAndTopics(c appengine.Context, id int64, member *Member) (*CertificationAndTopics, error) {
	errC := make(chan error)
	topicC := make(chan []*TrainingTopic)
	certC := make(chan []*Certification)

	certKey := datastore.NewKey(c, "Certification", "", id, nil)

	go func() {
		cert := new(Certification)

		err := cert.lookup(id, member, c)

		certC <- []*Certification{cert}
		errC <- err
	}()

	certs := <-certC

	go func() {
		query := datastore.NewQuery("TrainingTopic").Filter("Deleted =", false).Filter("CertificationKey = ", certKey)

		var lookupResults []*TrainingTopic
		keys, err := query.GetAll(c, &lookupResults)

		for idx, _ := range lookupResults {
			e := lookupResults[idx]
			key := keys[idx]

			c.Debugf("found TrainingTopic %d, %d, %d", idx, e.KeyID, key.IntID())
			c.Debugf("\t for cert %d", e.CertificationKey.IntID())

			e.setKey(key)
		}
		topicC <- lookupResults
		errC <- err
	}()

	topics := <-topicC

	err1 := <-errC
	err2 := <-errC

	if noErrMsg(err1, nil, c, "getCertAndTopics 1") {
		if noErrMsg(err2, nil, c, "getCertAndTopics 2") {
			results := processCertsAndTopics(c, certs, topics)
			if len(results) != 1 {
				err := errors.New(fmt.Sprintf("Expected 1 CertificationAndTopics object, but found: %d", len(results)))

				return nil, err
			}

			return results[0], nil
		} else {
			return nil, err2
		}
	}

	return nil, err1
}

func processCertsAndTopics(c appengine.Context, certs []*Certification, topics []*TrainingTopic) []*CertificationAndTopics {
	var results []*CertificationAndTopics

	certsByID := make(map[int64]*CertificationAndTopics)

	for idx, _ := range certs {
		cert := certs[idx]

		candt := CertificationAndTopics{
			Certification: cert,
			Topics:        make([]*TrainingTopic, 0),
		}

		results = append(results, &candt)

		certsByID[cert.KeyID] = &candt
	}

	for idx, _ := range topics {
		topic := topics[idx]

		c.Debugf("matching TrainingTopic %d, %d", idx, topic.KeyID)
		c.Debugf("\t for cert %d", topic.CertificationKey.IntID())

		candt := certsByID[topic.CertificationKey.IntID()]

		c.Debugf("\t for candt %d, %d", candt.Certification.KeyID)

		candt.Topics = append(candt.Topics, topic)

		c.Debugf("\t after append %d, %+v", len(candt.Topics), topic)
	}

	return results
}

func (t *Certification) save(member *Member, c appengine.Context) error {
	tKey := t.Key
	if tKey == nil {
		tKey = datastore.NewKey(c, "Certification", "", t.KeyID, nil)
	}

	t.setAudits(member)

	outKey, putErr := datastore.Put(c, tKey, t)

	if noErrMsg(putErr, nil, c, "Put Certification") {
		c.Debugf("Certification Key after Put: %d, %d", tKey.IntID(), outKey.IntID())
		t.setKey(outKey)

		c.Debugf("Certification after Put: %+v", t)
	}

	return putErr
}

func (t *Certification) lookup(id int64, member *Member, c appengine.Context) error {
	t.Key = datastore.NewKey(c, "Certification", "", id, nil)

	err := datastore.Get(c, t.Key, t)

	if noErrMsg(err, nil, c, fmt.Sprintf("Getting Certification %d", id)) {
		t.setKey(t.Key)
	}

	return err
}

func apiTrainingTopicSave(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var mem *Member
	var postData struct {
		TrainingTopic *TrainingTopic
		Certification *Certification
	}

	c.Infof("apiTrainingTopicSave")

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&postData)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Parsing json from body") {
		c.Infof("JSON from request: %+v", postData)

		mem, _ = getMemberFromUser(c, u, w, r)

		certKey := datastore.NewKey(c, "Certification", "", postData.Certification.KeyID, nil)
		postData.TrainingTopic.CertificationKey = certKey
		saveErr := postData.TrainingTopic.save(mem, c)

		if noErrMsg(saveErr, w, c, "Saving Certification") {
			context = postData
		}
	} else {

	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func (t *TrainingTopic) save(member *Member, c appengine.Context) error {
	tKey := t.Key
	if tKey == nil {
		tKey = datastore.NewKey(c, "TrainingTopic", "", t.KeyID, nil)
	}

	t.setAudits(member)

	outKey, putErr := datastore.Put(c, tKey, t)

	if noErrMsg(putErr, nil, c, "Put TrainingTopic") {
		t.setKey(outKey)
	}

	return putErr
}

func (obj *CertificationClass) save(member *Member, c appengine.Context) error {
	objType := "CertificationClass"

	objKey := obj.Key
	if objKey == nil {
		objKey = datastore.NewKey(c, objType, "", obj.KeyID, nil)
	}

	obj.setAudits(member)

	outKey, putErr := datastore.Put(c, objKey, obj)

	if noErrMsg(putErr, nil, c, "Put "+objType) {
		c.Debugf("Key after Put: %d, %d", objKey.IntID(), outKey.IntID())
		obj.setKey(outKey)

		c.Debugf("after Put: %+v", obj)
	}

	return putErr
}

func (t *CertificationClass) lookup(id int64, member *Member, c appengine.Context) error {
	t.Key = datastore.NewKey(c, "CertificationClass", "", id, nil)

	err := datastore.Get(c, t.Key, t)

	if noErrMsg(err, nil, c, fmt.Sprintf("Getting CertificationClass %d", id)) {
		t.setKey(t.Key)
	}

	return err
}

func apiCertificationClassSave(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var mem *Member
	var postData struct {
		CClass *CertificationClass
		Team   *Team
	}

	c.Infof("certificationClassSave")

	// err := parseJSON(postData, r, c)
	// if noErrMsg(err, w, c, "Parsing JSON") {

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&postData)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, nil, c, "Parsing json from body") {
		c.Infof("JSON from request: %+v", postData)

		mem, _ = getMemberFromUser(c, u, w, r)

		teamKey := datastore.NewKey(c, "Team", "", postData.Team.KeyID, nil)
		postData.CClass.TeamKey = teamKey

		saveErr := postData.CClass.save(mem, c)

		c.Debugf("CertificationClass after save: %+v", postData.CClass)

		if noErrMsg(saveErr, w, c, "Saving Certification") {
			context = struct {
				CClass *CertificationClass
			}{
				postData.CClass,
			}
		}
	} else {

	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func apiCertificationClassGet(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var mem *Member

	id := parseIntIDVal(r)

	c.Infof("certificationClassGet %d", id)

	mem, _ = getMemberFromUser(c, u, w, r)

	cClass := new(CertificationClass)
	err := cClass.lookup(id, mem, c)

	if noErrMsg(err, w, c, "Certification lookup") {
		context = struct {
			CertificationClass *CertificationClass
		}{
			cClass,
		}
	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func apiCertificationClassGetAll(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var mem *Member

	id := parseIntFormVal("teamID", r)

	c.Infof("certificationClassGetAll, id: %d", id)

	if id != 0 {
		mem, _ = getMemberFromUser(c, u, w, r)

		if mem != nil {
			var results []*CertificationClass
			var teamKey = datastore.NewKey(c, "Team", "", id, nil)

			query := datastore.NewQuery("CertificationClass").Filter("TeamKey =", teamKey)

			keys, err := query.GetAll(c, &results)

			if noErrMsg(err, w, c, "Getting All CertificationClasses for Team") {
				for idx, _ := range results {
					e := results[idx]
					key := keys[idx]

					e.setKey(key)
				}
			}

			context = struct {
				CClasses []*CertificationClass
			}{
				results,
			}
		} else {
			context = struct {
				NotLoggedIn bool
			}{
				false,
			}
		}
	} else {
		context = struct {
			Error   bool
			Message string
		}{
			true,
			"Must pass teamID",
		}
	}
	returnJSONorErrorToResponse(context, c, w, r)
}

func apiMemberSearch(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	var context interface{}
	queries := 0
	var members []*Member
	membersChan := make(chan []*Member)

	postData := struct {
		Phone       string
		NameOrEmail string
	}{}

	c.Infof("apiMemberSearch")

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&postData)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, nil, c, "Parsing json from body") {
		c.Infof("JSON from request: %+v", postData)

		if postData.Phone == "" {
			queries += 3
			memberFirstNameQuery(postData.NameOrEmail, false, c)
			memberLastNameQuery(postData.NameOrEmail, false, c)
			memberEmailQuery(postData.NameOrEmail, false, c)
		} else {
			queries++
			go func() {
				members, err := memberPhoneQuery(postData.Phone, false, c)

				if err == nil {
					membersChan <- members
				} else {
					c.Errorf("%+v", err)
				}
			}()
		}

		for queries > 0 {
			select {
			case newMembers := <-membersChan:
				for idx, _ := range newMembers {
					m := newMembers[idx]
					members = append(members, m)
				}
			}
			queries--
		}

		c.Infof("after Member query(s)")
		// if noErrMsg(err, w, c, "GetAll") {
		context = struct {
			Members []*Member
		}{
			members,
		}
	} // else jsonDecodeErr

	returnJSONorErrorToResponse(context, c, w, r)
}

func memberFirstNameQuery(value string, equality bool, c appengine.Context) ([]*Member, error) {
	c.Infof("Member First Name Query: %s", value)

	return nil, nil
}

func memberLastNameQuery(value string, equality bool, c appengine.Context) ([]*Member, error) {
	c.Infof("Member Last Name Query: %s", value)

	return nil, nil
}

func memberEmailQuery(value string, equality bool, c appengine.Context) ([]*Member, error) {
	c.Infof("Member Email Query: %s", value)

	return nil, nil
}

func memberPhoneQuery(value string, equality bool, c appengine.Context) ([]*Member, error) {
	c.Infof("Member Phone Query: %s", value)

	mems, err := memberQuery("Cell", value, equality, c)
	return mems, err
}

func memberQuery(field string, value string, equality bool, c appengine.Context) ([]*Member, error) {
	var members []*Member

	greaterThanSign := ""

	if equality == false {
		greaterThanSign = ">"
	}

	greaterThanEqual := fmt.Sprintf("%s %s=", field, greaterThanSign)
	lessThan := fmt.Sprintf("%s < ", field)

	// higest unicode value from here: http://stackoverflow.com/questions/47786/google-app-engine-is-it-possible-to-do-a-gql-like-query
	valueNext := value[:len(value)-1] + "\ufffd"

	memberQ := datastore.NewQuery("Member").Filter(greaterThanEqual, value)

	c.Debugf("Filter 1: %s, %s", greaterThanEqual, value)
	if equality == false {
		memberQ.Filter(lessThan, valueNext)
		c.Debugf("Filter 2: %s, %s", lessThan, valueNext)
	}

	c.Infof("Got membersQ and members for calling GetAll")
	keys, err := memberQ.GetAll(c, &members)

	for idx, _ := range members {
		m := members[idx]
		k := keys[idx]
		m.setKey(k)
	}

	return members, err
}

func checkErr(err error, w http.ResponseWriter, c appengine.Context, msg string) bool {
	retval := false
	if err != nil {
		c.Errorf("While attempting: %v Error: %+v", msg, err)
		if w != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		retval = true
	} else {
		c.Debugf("no error from %v", msg)
	}
	return retval
}

func noErr(err error, w http.ResponseWriter, c appengine.Context) bool {
	retval := checkErr(err, w, c, "")

	return !retval
}

func noErrMsg(err error, w http.ResponseWriter, c appengine.Context, msg string) bool {
	retval := checkErr(err, w, c, msg)

	return !retval
}

func parseIntFormVal(field string, r *http.Request) int64 {
	key := r.FormValue(field)

	if key == "" {
		key = "0"
	}

	id, _ := strconv.ParseInt(key, 0, 0)

	return id
}

func parseIntIDVal(r *http.Request) int64 {
	key := r.FormValue("id")

	if key == "" {
		key = "0"
	}

	id, _ := strconv.ParseInt(key, 0, 0)

	return id
}

func parseJSON(object interface{}, r *http.Request, c appengine.Context) error {

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&object)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, nil, c, "Parsing json from body") {
		c.Infof("JSON from request: %+v", object)
	}

	return jsonDecodeErr
}
