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
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"appengine/mail"
	"appengine/urlfetch"
	"appengine/user"
)

var zeroDate = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

var delayEmailReminders2 = delay.Func("sendEmailReminders2", sendEmailReminders)

func initialSetup(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	var context interface{}

	u := user.Current(c)

	if u.Admin {
		member, errM := getMemberFromUser(c, u)

		if noErrMsg(errM, w, c, "Loading member") {
			member.Officer = true

			_, errM = datastore.Put(c, member.Key, member)
			if noErrMsg(errM, w, c, "Making Member Officer") {

				errT, team := getTeam(0, member, c)

				if noErrMsg(errT, w, c, "Loading team in initialSetup") {
					//addErr := member.addToTeam(team.Key, c)

					if team == nil {
						// TODO setup the team
					}

					var tmRecords []*TeamMember
					tmQuery := datastore.NewQuery("TeamMember").Filter("TeamKey = ", team.Key).Filter("MemberKey = ", member.Key).KeysOnly()

					var tmErr error
					//var keys []*datastore.Key

					c.Infof("calling tmQuery.GetAll")
					_, tmErr = tmQuery.GetAll(c, &tmRecords)

					if noErrMsg(tmErr, w, c, fmt.Sprintf("looking for TeamMember for team: %+v, Member: %+v", team.Key, member.Key)) {
						if len(tmRecords) == 0 {

							tm := TeamMember{MemberKey: member.Key, TeamKey: team.Key}
							tm.Key = datastore.NewKey(c, "TeamMember", "", 0, nil)
							tm.setAudits(member)

							tm.Key, tmErr = datastore.Put(c, tm.Key, &tm)

							if noErrMsg(tmErr, w, c, fmt.Sprintf("Saving TeamMember record, team: %+v, Member: %+v", team.Key, member.Key)) {

							}
						}
					}

					//checkErr(addErr, w, c, "Failed adding member to team")
					if team != nil {
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
		}
		context = struct {
			DidSomeStuff bool
		}{true}
	} else {
		context = struct {
			MustLoginAsAdmin bool
		}{true}
	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func fixData(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)

	var query *datastore.Query
	var results []*TeamMember

	var keys []*datastore.Key
	var err error
	var context interface{}

	if u.Admin {
		query = datastore.NewQuery("TeamMember")

		c.Infof("calling GetAll")
		keys, err = query.GetAll(c, &results)

		if noErrMsg(err, w, c, "Getting Events with no TeamKey") {
			for idx, _ := range results {
				e := results[idx]
				key := keys[idx]

				if e.MemberKey.IntID() != 5045143463788544 {
					datastore.Delete(c, key)
				}
			}
		}

		context = struct {
			Whee bool
		}{false}
	} else {
		context = struct {
			MustBeAdmin bool
		}{true}
	}
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
		mem, memErr := getMemberFromUser(c, u)

		if noErrMsg(memErr, w, c, "Getting Member from User") {
			c.Debugf("Got member: %d", mem.Key.IntID())

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

				checkErr(mErr, w, c, fmt.Sprintf("Failed to update Last Login for member: %d", mem.Key.IntID()))
			}
		} else {
			c.Errorf("Could not load member from user %+v", u)
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
		//context.Member = getMemberFromUser(c, u, context)
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
			member, _ = getMemberFromUser(c, u)
		}

		teamKey := r.FormValue("team")
		if teamKey == "" {
			teamKey = "0"
		}

		intKey, _ := strconv.ParseInt(teamKey, 0, 0)
		var getTeamErr error
		var team *Team
		getTeamErr, team = getTeam(intKey, member, c)

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
		member, _ := getMemberFromUser(c, u)

		teamKey := r.FormValue("team")
		if teamKey == "" {
			teamKey = "0"
		}

		intKey, _ := strconv.ParseInt(teamKey, 0, 0)
		var getTeamErr error
		getTeamErr, context.Team = getTeam(intKey, member, c)

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

	context := MainAppContext{}

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
		//context.Member = getMemberFromUser(c, u, context)
		context.LoggedIn = true

		url, err := user.LogoutURL(c, returnUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		context.LogInOutLink = url
		context.LogInOutText = "Log Out"

		member, _ := getMemberFromUser(c, u)

		memKey := r.FormValue("member")
		if memKey != "" {
			intKey, _ := strconv.ParseInt(memKey, 0, 0)
			context.Member, err = getMemberByIntKey2(c, intKey, member)
		} else {
			context.Member = member
		}

		lookupOthers(context.Member)

		context.Member.setKey(context.Member.Key)
	}

	//bArr, memberJSONerr := json.MarshalIndent(context, "", "\t")
	returnJSONorErrorToResponse(context, c, w, r)
}

func getMemberFromUser(c appengine.Context, u *user.User) (*Member, error) {

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

	if err != nil && strings.Index(err.Error(), "cannot load field") > -1 {
		c.Debugf("overrode error because it was just a field error: %s", err.Error())
		err = nil
	}

	if noErrMsg(err, nil, c, "members for id") {
		c.Infof("checking len(members)")
		if len(member) == 0 {
			c.Infof("existing Member not found for user.ID: %v, count:%d", u.ID, len(member))

			lowerEmail := strings.ToLower(u.Email)
			memberQ = datastore.NewQuery("Member").Filter("Email =", lowerEmail)

			c.Infof("Got membersQ and members for Email: %v, calling GetAll", lowerEmail)
			keys, err = memberQ.GetAll(c, &member)

			if noErrMsg(err, nil, c, "members for email") {
				c.Infof("checking len(members)")
				if len(member) == 0 {
					c.Infof("existing Member not found for user.Email: %v, count:%d", lowerEmail, len(member))

					memberQ = datastore.NewQuery("Member").Filter("Email2 =", lowerEmail)

					c.Infof("Got membersQ and members for Email: %v, calling GetAll", lowerEmail)
					keys, err = memberQ.GetAll(c, &member)

					if noErrMsg(err, nil, c, "members for email2") {
						if len(member) == 0 {
							c.Infof("existing Member not found for user.Email2: %v, count:%d", lowerEmail, len(member))

							m := &Member{
								Person: Person{
									FirstName: strings.ToLower(u.Email),
									LastName:  "",
									Email:     strings.ToLower(u.Email),
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

								Active: true,
							}

							//lKey := datastore.NewKey(c, "Location", "", 0, nil)
							mKey := datastore.NewKey(c, "Member", "", 0, nil)

							c.Infof("Putting Member: %v", mKey)
							outMKey, mErr := datastore.Put(c, mKey, m)
							m.setKey(outMKey)

							if noErrMsg(mErr, nil, c, "Trying to put Member") {

								c.Infof("no error on 1st member put")

								m.CreatedBy = outMKey
								m.ModifiedBy = outMKey

								_, mErr2 := datastore.Put(c, outMKey, m)
								if noErrMsg(mErr2, nil, c, "Trying to put Member again") {
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
				lowerEmail := strings.ToLower(u.Email)
				mem.Email = lowerEmail

				_, err = datastore.Put(c, mem.Key, mem)
				_ = checkErr(err, nil, c, fmt.Sprintf("Updating Member: %v Email to User value: %v, %+v", mem.Key, lowerEmail, mem))
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

func getMemberByIntKey2(c appengine.Context, key int64, currentMem *Member) (*Member, error) {
	allow := lookupOthers(currentMem)
	c.Debugf("Lookup Member by key: %d, allowed to lookup? %t", key, allow)

	var m Member
	var retval *Member
	var err error
	retval = nil

	if allow {
		k := datastore.NewKey(c, "Member", "", key, nil)

		err = datastore.Get(c, k, &m)

		index := strings.Index(err.Error(), "cannot load field \"HomeAddress\"")
		c.Debugf("Index: %d, Err: %+v", index, err)
		if err != nil && index > -1 {
			err = nil
		}

		if err != nil {
			c.Errorf("datastore.Get member error: %v", err)

		} else {
			retval = &m
			retval.setKey(k)
		}
	}
	return retval, err
}

func getTeam(teamID int64, member *Member, c appengine.Context) (error, *Team) {
	var team Team
	if teamID != 0 {
		teamKey := datastore.NewKey(c, "Team", "", teamID, nil)
		c.Debugf("Calling Get Team with Key: %+v", teamKey)
		getTeamErr := datastore.Get(c, teamKey, &team)

		if _, ok := getTeamErr.(*datastore.ErrFieldMismatch); ok {
			getTeamErr = nil
		}

		if noErrMsg(getTeamErr, nil, c, "Error Get Team with Key") {
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

		if _, ok := getTeamErr.(*datastore.ErrFieldMismatch); ok {
			getTeamErr = nil
		}

		if noErrMsg(getTeamErr, nil, c, "Calling GetAll for Team") {
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

	memberCallerByIntID := make(map[int64]int64)
	c.Debugf("looping teamMembers which has: %d", len(teamMembers))
	for _, tm := range teamMembers {
		c.Debugf("teamMember: %+v", tm)
		memberKeys = append(memberKeys, tm.MemberKey)
		calledByKey := tm.CalledBy
		if calledByKey != nil {
			memberCallerByIntID[tm.MemberKey.IntID()] = calledByKey.IntID()
		}
	}
	members := make([]Member, len(teamMembers))

	c.Debugf("Calling GetMulti with Keys: %+v", memberKeys)

	memberErr := datastore.GetMulti(c, memberKeys, members)

	c.Debugf("memberErr: %+v", memberErr)
	if memberErr != nil {
		if memberErr.Error() != "datastore: no such entity" {

			if _, ok := memberErr.(*datastore.ErrFieldMismatch); ok {
				memberErr = nil
			}

			if strings.Index(memberErr.Error(), "cannot load field \"HomeAddress\"") == -1 {

				checkErr(memberErr, w, c, "Calling GetMulti with Keys")
			}
		}
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

			val := memberCallerByIntID[m.KeyID]
			c.Debugf("setting CalledBy to: %d, for Member: %d", val, m.KeyID)
			m.CalledBy = val
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
		curMember, _ = getMemberFromUser(c, u)
	}

	teamKey := r.FormValue("team")
	if teamKey == "" {
		teamKey = "0"
	}

	intKey, _ := strconv.ParseInt(teamKey, 0, 0)

	if intKey > 0 {
		var getTeamErr error
		var team *Team
		getTeamErr, team = getTeam(intKey, curMember, c)

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
	var err error

	if curMember == nil {
		c.Debugf("Looking up curMember because it is nil")
		curMember, _ = getMemberFromUser(c, u)
	}
	c.Debugf("Looking up allow for curMember")
	allow := lookupOthers(curMember)

	c.Debugf("Will Save Member to Key: %d", saveMember.KeyID)
	saveMember.Key = datastore.NewKey(c, "Member", "", saveMember.KeyID, nil)

	curMember.setKey(curMember.Key)

	c.Debugf("Checking Keys || allow: %s, %d == %d?", allow, curMember.KeyID, saveMember.KeyID)
	if curMember.KeyID == saveMember.KeyID || allow {

		err = saveMember.locationGeocode(c)
		if err != nil {
			c.Errorf("geocode err: %+v", err)
		} else {
			c.Debugf("saveMember.Location after geocode: %+v", saveMember.Location)
		}

		saveMember.setAudits(curMember)

		c.Debugf("Setting email(s) ToLower")
		saveMember.Email = strings.ToLower(saveMember.Email)
		saveMember.Email2 = strings.ToLower(saveMember.Email2)

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

	curMember, _ := getMemberFromUser(c, u)
	allow := lookupOthers(curMember)

	if allow {

		for idx := range importData.Members {
			c.Debugf("Importing idx: %d", idx)
			saveMember := importData.Members[idx]

			c.Debugf("Saving Imported Member: %+v", saveMember)
			tk := datastore.NewKey(c, "Team", "", importData.TeamID, nil)

			teamErr, team := getTeam(tk.IntID(), curMember, c)

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
				saveMember.Active = true

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
			member, _ = getMemberFromUser(c, u)
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
				member, _ = getMemberFromUser(c, u)
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
			member, _ = getMemberFromUser(c, u)
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
		member, _ = getMemberFromUser(c, u)
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
		member, _ = getMemberFromUser(c, u)
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

		mem, _ = getMemberFromUser(c, u)

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

		mem, _ := getMemberFromUser(c, u)

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

		mem, _ = getMemberFromUser(c, u)

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
		gl, err := geocode(address, c)

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

func getGoogleAPIKey(teamId int64, c appengine.Context) (string, error) {
	u := user.Current(c)
	mem, _ := getMemberFromUser(c, u)
	key := ""

	err, team := getTeam(teamId, mem, c)

	if err == nil {

		key = team.GoogleAPIKey

		c.Debugf("Found GoogleAPIKey %s", team.GoogleAPIKey)

		if len(key) < 4 {
			key = ""
			err = errors.New("Google API Key seemed invalid as it was less than 4 letters")
		}
	}

	return key, err
}

func geocode(address string, c appengine.Context) (*GoogleLocationResults, error) {
	apiURL := "https://maps.googleapis.com/maps/api/geocode/json"
	key, keyErr := getGoogleAPIKey(0, c)

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

		if results.Status == "REQUEST_DENIED" {
			c.Errorf("api results with error %+v", results)
			return nil, errors.New("Google API request denied, bad Key?")
		}

		c.Infof("json results %+v", results)

		return &results, nil
	} else {
		return nil, keyErr
	}
}

func (m *Member) locationGeocode(c appengine.Context) error {
	var err error

	m.Longitude = 0
	m.Latitude = 0
	m.PublicLatitude = 0
	m.PublicLongitude = 0

	if m.Line1 != "" {
		err = m.Location.geocode(c)
	}

	return err
}

func (l *Location) geocode(c appengine.Context) error {
	address := fmt.Sprintf("%s, %s, %s", l.Line1, l.City, l.Zip)

	googleLoc, err := geocode(address, c)

	if err == nil {
		l.Latitude = googleLoc.Results[0].Geometry.Location.Lat
		l.Longitude = googleLoc.Results[0].Geometry.Location.Lng

		l.PublicLatitude = fudgeGPS(l.Latitude, c)
		l.PublicLongitude = fudgeGPS(l.Longitude, c)
	}

	if err != nil {
		c.Errorf("Error calling geocode: %+v", err)
	} else {
		c.Debugf("Location after geocode: %+v", l)
	}

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

		mem, _ = getMemberFromUser(c, u)

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

	mem, _ = getMemberFromUser(c, u)

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

	mem, _ = getMemberFromUser(c, u)

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

		if candt != nil {
			if candt.Certification != nil {
				c.Debugf("\t for candt %d", candt.Certification.KeyID)

				candt.Topics = append(candt.Topics, topic)

				c.Debugf("\t after append %d, %+v", len(candt.Topics), topic)
			} else {
				c.Errorf("\t Certification was not found %+v, %+v", candt.Topics, topic.CertificationKey)
			}
		} else {
			c.Errorf("\t candt was not found %+v, %+v", topic.CertificationKey, topic)
		}
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

		mem, _ = getMemberFromUser(c, u)

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
	if member == nil {
		//TODO: Should lookup member here
		//member = getMemberFromUser(c, u)
	}

	t.Key = datastore.NewKey(c, "CertificationClass", "", id, nil)

	err := datastore.Get(c, t.Key, t)

	if noErrMsg(err, nil, c, fmt.Sprintf("Getting CertificationClass %d", id)) {
		t.setKey(t.Key)

		if len(t.Attendees) > 0 {
			t.MembersAttending = make([]Member, len(t.Attendees))

			for _, x := range t.Attendees {
				c.Debugf("Attendee: %+v", x)
			}

			if err = datastore.GetMulti(c, t.Attendees, t.MembersAttending); err != nil {
				if me, ok := err.(appengine.MultiError); ok {
					for i, merr := range me {
						if merr == datastore.ErrNoSuchEntity {
							c.Errorf("Member with Key: %d not found when finding Atendees", t.Attendees[i])
						}
					}
				} else {
					return err
				}
			}

		}
	}

	return err
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

			firstChar := postData.NameOrEmail[0:1]
			upperFirstChar := strings.ToUpper(firstChar)

			if firstChar == upperFirstChar {
				upperFirstChar = ""
			}
			queries += doMemberSearch(postData.NameOrEmail, membersChan, c)

			if upperFirstChar != "" {
				newValue := upperFirstChar + postData.NameOrEmail[1:]

				queries += doMemberSearch(newValue, membersChan, c)
			}

			queries++
			go func() {
				newValue := strings.ToLower(postData.NameOrEmail)
				members, err := memberEmailQuery(newValue, false, c)

				if err == nil {
					membersChan <- members
				} else {
					c.Errorf("%+v", err)
				}
			}()

		} else {
			queries++
			go func() {
				members, err := memberCellQuery(postData.Phone, false, c)

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

func doMemberSearch(value string, membersChan chan []*Member, c appengine.Context) int {

	go func() {
		members, err := memberFirstNameQuery(value, false, c)

		if err == nil {
			membersChan <- members
		} else {
			c.Errorf("%+v", err)
		}
	}()

	go func() {
		members, err := memberLastNameQuery(value, false, c)

		if err == nil {
			membersChan <- members
		} else {
			c.Errorf("%+v", err)
		}
	}()

	return 2
}

func memberFirstNameQuery(value string, equality bool, c appengine.Context) ([]*Member, error) {
	c.Infof("Member First Name Query: %s", value)

	mems, err := memberQuery("FirstName", value, equality, c)
	return mems, err
}

func memberLastNameQuery(value string, equality bool, c appengine.Context) ([]*Member, error) {
	c.Infof("Member Last Name Query: %s", value)

	mems, err := memberQuery("LastName", value, equality, c)
	return mems, err
}

func memberEmailQuery(value string, equality bool, c appengine.Context) ([]*Member, error) {
	c.Infof("Member Email Query: %s", value)

	mems, err := memberQuery("Email", value, equality, c)
	return mems, err
}

func memberCellQuery(value string, equality bool, c appengine.Context) ([]*Member, error) {
	c.Infof("Member Cell Query: %s", value)

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
		memberQ = memberQ.Filter(lessThan, valueNext)
		c.Debugf("Filter 2: %s, %s", lessThan, valueNext)
	}

	c.Infof("Got membersQ and members for calling GetAll: %+v", memberQ)
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
