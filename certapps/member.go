package certapps

import (
	"encoding/json"
	"errors"
	"fmt"
	//	"html/template"
	"io"
	//"io/ioutil"
	//"math/rand"
	"net/http"
	//"net/url"
	//"strconv"
	//"strings"
	//"time"

	"appengine"
	"appengine/datastore"
	//"appengine/delay"
	//"appengine/mail"
	//"appengine/urlfetch"
	"appengine/user"
)

func apiMemberCalledBy(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	var context interface{}
	errMsg := ""

	c.Infof("apiMemberCalledBy")

	u := user.Current(c)
	currentMem, curMemberErr := getMemberFromUser(c, u)

	if noErrMsg(curMemberErr, w, c, "Getting Member from User") {

		postData := struct {
			Member   int64
			CalledBy int64
			Team     int64
		}{}

		decoder := json.NewDecoder(r.Body)
		jsonDecodeErr := decoder.Decode(&postData)

		if jsonDecodeErr == io.EOF {
			c.Infof("EOF, should it be?")
		} else if noErrMsg(jsonDecodeErr, nil, c, "Parsing json from body") {
			c.Infof("JSON from request: %+v", postData)

			if lookupOthers(currentMem) {
				var memberToChange *Member

				memberToChange, memberToChangeErr := getMemberByIntKey2(c, postData.Member, currentMem)

				if noErrMsg(memberToChangeErr, w, c, "Looking up member to change CalledBy for") {
					teamKey := datastore.NewKey(c, "Team", "", postData.Team, nil)

					c.Debugf("Looking up TeamMember for request from Member: %d, Lookup Team: %d = %d ?, Lookup Member: %d", currentMem.KeyID, teamKey.IntID(), memberToChange.KeyID)
					teamMember, teamMemberErr := getTeamMemberByTeamAndMember(c, teamKey, memberToChange.Key, currentMem)
					if noErrMsg(teamMemberErr, w, c, "Looking up TeamMember for Team, Member") {
						callerKey := datastore.NewKey(c, "Member", "", postData.CalledBy, nil)

						teamMember.CalledBy = callerKey
						teamMember.setAudits(currentMem)

						c.Debugf("Put TeamMember %+v", teamMember)
						datastore.Put(c, teamMember.Key, teamMember)
					} else {
						errMsg = "Could not find TeamMember"
					}
				} else {
					errMsg = "Could not find team"
				}
			} else {
				errMsg = "Could not find member to change CalledBy For"
			}

			context = struct {
				Success bool
			}{
				true,
			}

		} else {
			context = struct {
				Error            bool
				Message          string
				PermissionsError bool
			}{
				true,
				"User does not have rights to lookup other members",
				true,
			}
		}
	} // else jsonDecodeErr

	if context == nil {
		if errMsg != "" {
			context = struct {
				Error   bool
				Message string
			}{
				true,
				errMsg,
			}
		}
	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func apiMemberToggleActive(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	var context interface{}
	errMsg := ""

	c.Infof("apiMemberToggleActive")

	u := user.Current(c)
	currentMem, curMemberErr := getMemberFromUser(c, u)

	if noErrMsg(curMemberErr, w, c, "Getting Member from User") {

		postData := struct {
			Member int64
			Team   int64
			Active bool
		}{}

		decoder := json.NewDecoder(r.Body)
		jsonDecodeErr := decoder.Decode(&postData)

		if jsonDecodeErr == io.EOF {
			c.Infof("EOF, should it be?")
		} else if noErrMsg(jsonDecodeErr, nil, c, "Parsing json from body") {
			c.Infof("JSON from request: %+v", postData)

			if lookupOthers(currentMem) {
				var memberToChange *Member

				memberToChange, memberToChangeErr := getMemberByIntKey2(c, postData.Member, currentMem)

				if noErrMsg(memberToChangeErr, w, c, "Looking up member to toggle Active for") {
					if memberToChange.Active != postData.Active {
						memberToChange.Active = postData.Active
						memberToChange.setAudits(currentMem)

						c.Debugf("Changing Active for Member: %d to %t", memberToChange.Key.IntID(), memberToChange.Active)
						datastore.Put(c, memberToChange.Key, memberToChange)
						//if noErrMsg(putErr, w, c, "Putting member") {
						context = struct {
							Success bool
							Active  bool
						}{
							true,
							memberToChange.Active,
						}
						// } else {
						// 	errMsg = "Error saving user"
						// }
					}
				} else {
					errMsg = "Could not find member"
				}
			} else {
				errMsg = "Could not lookup other members"
			}

		} else {
			context = struct {
				Error            bool
				Message          string
				PermissionsError bool
			}{
				true,
				"User does not have rights to lookup other members",
				true,
			}
		}
	} // else jsonDecodeErr

	if context == nil {
		if errMsg != "" {
			context = struct {
				Error   bool
				Message string
			}{
				true,
				errMsg,
			}
		}
	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func getTeamMemberByTeamAndMember(c appengine.Context, teamKey *datastore.Key, memberKey *datastore.Key, currentMem *Member) (*TeamMember, error) {
	var teamMember *TeamMember
	var err error

	if lookupOthers(currentMem) {
		q := datastore.NewQuery("TeamMember").
			Filter("TeamKey =", teamKey).
			Filter("MemberKey =", memberKey).
			Limit(1)

		var teamMembers []TeamMember
		var keys []*datastore.Key

		keys, err = q.GetAll(c, &teamMembers)

		if len(teamMembers) == 1 {
			for idx, _ := range teamMembers {
				tm := teamMembers[idx]
				key := keys[idx]

				tm.setKey(key)

				teamMember = &tm
			}

		} else {
			err = errors.New(fmt.Sprintf("Could not find TeamMember for Team: %d, Member: %d", teamKey.IntID(), memberKey.IntID()))
		}
	}

	if teamMember != nil {
		c.Debugf("TeamMember found: %+v", teamMember)
	}

	return teamMember, err
}
