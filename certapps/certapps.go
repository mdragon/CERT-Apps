package certapps

import (
	//	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	//	"html/template"
	"io"
	//	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	textT "text/template"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
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
	AuthContext
}

type JSONContext struct {
	Data string
}

type Location struct {
	Line1 string
	Line2 string
	City  string
	State string
	Zip   string
}

type Audit struct {
	Created    time.Time
	CreatedBy  *datastore.Key
	Modified   time.Time
	ModifiedBy *datastore.Key

	Key   *datastore.Key `datastore:"-"`
	KeyID int64          `datastore:"-"`
}

type Team struct {
	Name string
	Location

	MembersEmail  string
	OfficersEmail string

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

	IAPLink    string
	RosterLink string

	Audit
}

type MemberEvent struct {
	MemberKey *datastore.Key
	EventKey  *datastore.Key

	Attending bool
	Sure      bool

	Responded time.Time
	Arrive    time.Time
	Depart    time.Time

	Audit
}

var jsonTemplate = textT.Must(textT.New("json").Parse("{\"data\": {{.Data}} }"))
var jsonErrTemplate = textT.Must(textT.New("jsonErr").Parse("{\"error\": true, {{.Data}} }"))

func init() {
	http.HandleFunc("/audit", audit)
	http.HandleFunc("/event", event)
	http.HandleFunc("/event/save", eventSave)
	http.HandleFunc("/member", memberData)
	http.HandleFunc("/member/save", memberSave)
	http.HandleFunc("/team", teamData)
	http.HandleFunc("/team/roster", teamRoster)
	http.HandleFunc("/team/roster/import", membersImport)
	http.HandleFunc("/del/broken/teamMember/links", delBroken)
	http.HandleFunc("/", root)
}

func delBroken(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	//u := user.Current(c)

	var query *datastore.Query
	query = datastore.NewQuery("TeamMember")
	var results []*TeamMember

	var keys []*datastore.Key
	var err error

	c.Infof("calling GetAll")
	keys, err = query.GetAll(c, &results)

	c.Infof("after GetAll")

	if noErrMsg(err, w, c, "GetAll") {
		var brokenKeys []*datastore.Key
		for idx := range results {
			res := results[idx]

			if res.MemberKey.IntID() == 0 {
				key := keys[idx]
				c.Infof("broken %+v", key)
				brokenKeys = append(brokenKeys, key)
			} else {
				c.Infof("ok %+v", res)
			}
		}

		c.Infof("Deleting %d broken Entities", len(brokenKeys))

		if len(brokenKeys) > 0 {
			delErr := datastore.DeleteMulti(c, brokenKeys)

			if noErrMsg(delErr, w, c, "DeleteMulti") {

				c.Infof("Did delete")
			} else {
				c.Infof("Error on DeleteMulti")
			}
		}
	}

	context := struct {
		Whee bool
	}{false}

	returnJSONorErrorToResponse(context, c, r, w)
}

// guestbookKey returns the key used for all guestbook entries.
func guestbookKey(c appengine.Context) *datastore.Key {
	// The string "default_guestbook" here could be varied to have multiple guestbooks.
	return datastore.NewKey(c, "Guestbook", "default_guestbook", 0, nil)
}

func root(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/static/html/app.htm", http.StatusFound)
}

func audit(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)

	rawurl := r.FormValue("finalURL")
	c.Debugf("rawurl %v", rawurl)
	url, err := url.QueryUnescape(rawurl)

	if u != nil {
		mem := getMember(c, u, r, w)

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

func auth(w http.ResponseWriter, r *http.Request) (appengine.Context, *user.User, MainAppContext) {
	c := appengine.NewContext(r)
	u := user.Current(c)

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
		//context.Member = getMember(c, u, context, r, w)
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

func teamRoster(w http.ResponseWriter, r *http.Request) {
	c, u, mainContext := auth(w, r)

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
			member = getMember(c, u, r, w)
		}

		teamKey := r.FormValue("team")
		if teamKey == "" {
			teamKey = "0"
		}

		intKey, _ := strconv.ParseInt(teamKey, 0, 0)
		var getTeamErr error
		var team *Team
		getTeamErr, team = getTeam(intKey, member, c, r, w)

		if noErrMsg(getTeamErr, w, c, "GetTeam with Key: "+teamKey) {
			context.Members = getMembersByTeam(team.Key.IntID(), member, c, r, w)

			context.Team = team
		}
	}

	returnJSONorErrorToResponse(context, c, r, w)

	return
}

func teamData(w http.ResponseWriter, r *http.Request) {
	c, u, context := auth(w, r)

	if context.LoggedIn {
		member := getMember(c, u, r, w)

		teamKey := r.FormValue("team")
		if teamKey == "" {
			teamKey = "0"
		}

		intKey, _ := strconv.ParseInt(teamKey, 0, 0)
		var getTeamErr error
		getTeamErr, context.Team = getTeam(intKey, member, c, r, w)

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
	returnJSONorErrorToResponse(context, c, r, w)

	return
}

func memberData(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
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
		//context.Member = getMember(c, u, context, r, w)
		context.LoggedIn = true

		url, err := user.LogoutURL(c, returnUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		context.LogInOutLink = url
		context.LogInOutText = "Log Out"

		member := getMember(c, u, r, w)

		memKey := r.FormValue("member")
		if memKey != "" {
			intKey, _ := strconv.ParseInt(memKey, 0, 0)
			context.Member = getMemberByIntKey2(intKey, member, c, r, w)
		} else {
			context.Member = member
		}

		lookupOthers(context.Member)
	}

	//bArr, memberJSONerr := json.MarshalIndent(context, "", "\t")
	returnJSONorErrorToResponse(context, c, r, w)
}

func returnJSONorErrorToResponse(context interface{}, c appengine.Context, r *http.Request, w http.ResponseWriter) {
	jsonC := JSONContext{}
	bArr, memberJSONerr := json.Marshal(context)
	if noErrMsg(memberJSONerr, w, c, "json.Marshall of Member") {
		c.Debugf("getting length")
		//n := bytes.Index(bArr, []byte{0})
		n := len(bArr)

		if n > 0 {
			c.Debugf("getting string for: %d bytes", n)
			jsonC.Data = string(bArr[:n])

			c.Debugf("jsonTemplate.ExecuteTemplate: %+v", jsonC)
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

func getMember(c appengine.Context, u *user.User, r *http.Request, w http.ResponseWriter) *Member {
	var memberQ *datastore.Query
	memberQ = datastore.NewQuery("Member").Filter("UserID =", u.ID)
	var member []Member

	var keys []*datastore.Key
	var err error

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
				if len(member) == 0 {
					c.Infof("existing Member not found for user.Email: %v, count:%d", u.Email, len(member))
					/*
						l := &Location{
							Line1: "123 Main St",
							City:  "Anytown",
							State: "NJ",
							Zip:   "55555",

							Created:  time.Now(),
							Modified: time.Now(),
							// need to be nil because we don't have the member key to use yet
							CreatedBy:  nil,
							ModifiedBy: nil,
						}
					*/
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
							},
						},
					}

					//lKey := datastore.NewKey(c, "Location", "", 0, nil)
					mKey := datastore.NewKey(c, "Member", "", 0, nil)

					c.Infof("Putting Member: %v", mKey)
					outMKey, mErr := datastore.Put(c, mKey, m)

					if noErrMsg(mErr, w, c, "Trying to put Member") {
						//l.CreatedBy = outMKey
						//l.ModifiedBy = outMKey

						// c.Infof("Putting Location: %v", lKey)
						// outLKey, lErr := datastore.Put(c, lKey, l)

						// if noErrMsg(lErr, w, c, "Trying to put Location") {
						c.Infof("no error on 1st member put")

						//					m.HomeAddress = outLKey
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
				} else {
					// found by email
					mem = &member[0]
					mem.setKey(keys[0])

					found = true
					mem.UserID = u.ID

					c.Infof("Adding User.ID: %v, to Member with Key: %v", mem.UserID, mem.Key)
					datastore.Put(c, mem.Key, mem)
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

		} else {
			c.Infof("no mem")
		}
	}

	return mem
}

func getMemberByKey2(key string, c appengine.Context, r *http.Request, w http.ResponseWriter) *Member {
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

func getMemberByIntKey2(key int64, currentMem *Member, c appengine.Context, r *http.Request, w http.ResponseWriter) *Member {
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

func getTeam(teamID int64, member *Member, c appengine.Context, r *http.Request, w http.ResponseWriter) (error, *Team) {
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

func getMembersByTeam(teamID int64, member *Member, c appengine.Context, r *http.Request, w http.ResponseWriter) []Member {
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

	checkErr(memberErr, w, c, "Error calling GetMulti with Keys")

	lookup := lookupOthers(member)
	for idx := range members {
		m := &members[idx]
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

			m.Line1 = "123 My Street"
			m.Line2 = "Apt 1"
			m.City = "Anytown"
			m.Zip = "11111"
		} else {
			c.Debugf("Has lookup rights, not resetting hidden email/cell")
		}
	}

	return members
}

type MemberSaveContext struct {
	Whee string
}

func memberSave(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
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

	save(&saveMember, nil, c, u, w, r)
}

func save(saveMember *Member, curMember *Member, c appengine.Context, u *user.User, w http.ResponseWriter, r *http.Request) (*datastore.Key, error) {

	if curMember == nil {
		c.Debugf("Looking up curMember because it is nil")
		curMember = getMember(c, u, r, w)
	}
	c.Debugf("Looking up allow for curMember")
	allow := lookupOthers(curMember)

	var saveMemberKey *datastore.Key

	if saveMember.Key != nil {
		saveMemberKey = saveMember.Key
	} else {
		saveMember.Key = datastore.NewKey(c, "Member", "", 0, nil)
	}

	c.Debugf("Checking Keys || allow: %s", allow)
	if curMember.Key == saveMemberKey || allow {
		saveMember.setAudits(curMember)

		keyOut, mErr := datastore.Put(c, saveMember.Key, saveMember)

		checkErr(mErr, w, c, "Failed to update Member for memberSave: ")

		context := struct {
			Member *Member
		}{
			saveMember,
		}
		returnJSONorErrorToResponse(context, c, r, w)

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
		returnJSONorErrorToResponse(context, c, r, w)
	}

	return nil, nil
}

func membersImport(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
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

	curMember := getMember(c, u, r, w)
	allow := lookupOthers(curMember)

	if allow {

		for idx := range importData.Members {
			c.Debugf("Importing idx: %d", idx)
			saveMember := importData.Members[idx]

			c.Debugf("Saving Imported Member: %+v", saveMember)
			tk := datastore.NewKey(c, "Team", "", importData.TeamID, nil)

			teamErr, team := getTeam(tk.IntID(), curMember, c, r, w)

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

				saveMember.Active = true

				mKeyOut, mErr := save(&saveMember, curMember, c, u, w, r)

				if noErrMsg(mErr, w, c, "Save member in import") {
					tm := &TeamMember{
						TeamKey:   tk,
						MemberKey: mKeyOut,
					}

					tm.setAudits(curMember)

					tm.Key = datastore.NewKey(c, "TeamMember", "", 0, nil)
					_, tmErr := datastore.Put(c, tm.Key, tm)

					checkErr(tmErr, w, c, "Creating TeamMember record for member: ")
				}
			}
		}

		context := struct {
			Whee bool
		}{
			true,
		}
		returnJSONorErrorToResponse(context, c, r, w)
	} else {
		c.Errorf("Only OEM, Town, Officer can update a different Member's record: %+v, tried to save: %+v", curMember, saveMember)
		context := struct {
			error   bool
			message string
		}{
			true,
			"Only OEM, Town, Officer can update a different Member's record",
		}
		returnJSONorErrorToResponse(context, c, r, w)
	}
}

func (a *Audit) setAudits(m *Member) {
	a.ModifiedBy = m.Key
	a.Modified = time.Now()

	if a.CreatedBy == nil {
		a.Created = m.Modified
		a.CreatedBy = m.CreatedBy
	}
}

func (a *Audit) setKey(key *datastore.Key) {
	a.Key = key
	if key != nil {
		a.KeyID = key.IntID()
	}
}

func event(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	var event Event

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&event)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Parsing event json from body") {
		c.Infof("JSON event: %+v", event)
		event.lookup(nil, c, u, w, r)
	} else {

	}

	context := struct {
		Event Event
	}{
		event,
	}

	returnJSONorErrorToResponse(context, c, r, w)
}

func (event *Event) lookup(member *Member, c appengine.Context, u *user.User, w http.ResponseWriter, r *http.Request) error {
	c.Infof("*Event.lookup %d", event.KeyID)

	if event.KeyID != 0 {
		event.Key = datastore.NewKey(c, "Event", "", event.KeyID, nil)

		//member := getMember(c, u, r, w)

		err := datastore.Get(c, event.Key, event)

		event.setKey(event.Key)

		return err
	}

	return nil
}

func eventSave(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	var event Event

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&event)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, w, c, "Parasing event json from body") {
		c.Infof("JSON event: %+v", event)

		saveErr := event.save(nil, c, u, w, r)

		if saveErr != nil {
			context := struct {
				error   bool
				message string
			}{
				true,
				"Only OEM, Town, Officer can update an Event",
			}
			returnJSONorErrorToResponse(context, c, r, w)
		}
	} else {
		c.Infof("partial JSON event: %+v", event)
	}

	context := struct {
		Event
	}{
		event,
	}

	returnJSONorErrorToResponse(context, c, r, w)
}

func (event *Event) save(member *Member, c appengine.Context, u *user.User, w http.ResponseWriter, r *http.Request) error {
	var err error

	c.Infof("*Event.save %d", event.KeyID)

	event.Key = datastore.NewKey(c, "Event", "", event.KeyID, nil)

	if member == nil {
		member = getMember(c, u, r, w)
	}

	if lookupOthers(member) {
		event.setAudits(member)

		var key *datastore.Key

		c.Infof("Will put event ID: %d", event.KeyID)
		key, err = datastore.Put(c, event.Key, event)

		if noErrMsg(err, w, c, "Put event") {
			c.Infof("Set Key: %d, %d", event.KeyID, key.IntID())
			event.setKey(key)
		}
	} else {
		c.Errorf("Only OEM, Town, Officer can update an event record: %+v, tried to save: %+v", member, event)

		err = errors.New("Only Town, OEM, Officers can save events")
	}

	return err
}

func locationLookup(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	url := "https://maps.googleapis.com/maps/api/geocode/json?key="
	key := "AIzaSyDKyDk-VkRFZIs27NAGmHEbrC17s7ylTYE"
	address := "address=13%20sheridan%20ave,%2007052"

	c.Debugf("url: %s, key: %s, address: %s", url, key, address)

}

func checkErr(err error, w http.ResponseWriter, c appengine.Context, msg string) bool {
	retval := false
	if err != nil {
		c.Errorf("While attempting: %v Error: %+v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
