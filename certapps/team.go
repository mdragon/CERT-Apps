package certapps

import (
	"encoding/json"
	//"errors"
	//"fmt"
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

func apiTeamSave(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	var context interface{}
	errMsg := ""

	c.Infof("apiTeamSave")

	u := user.Current(c)
	currentMem, curMemberErr := getMemberFromUser(c, u)

	if noErrMsg(curMemberErr, w, c, "Getting Member from User") {

		postData := struct {
			Team Team
		}{}

		decoder := json.NewDecoder(r.Body)
		jsonDecodeErr := decoder.Decode(&postData)

		if jsonDecodeErr == io.EOF {
			c.Infof("EOF, should it be?")
			errMsg = "JSON read error, was EOF"
		} else if noErrMsg(jsonDecodeErr, nil, c, "Parsing json from body") {
			c.Infof("JSON from request: %+v", postData)

			if lookupOthers(currentMem) {

				teamKey := datastore.NewKey(c, "Team", "", postData.Team.KeyID, nil)

				c.Debugf("Saving  Teamr for request from Member: %d, Lookup Team: %d ", currentMem.KeyID, teamKey.IntID())

				postData.Team.setAudits(currentMem)

				datastore.Put(c, teamKey, &postData.Team)

				context = struct {
					Success bool
				}{
					true,
				}
				// if noErrMsg(teamMemberErr, w, c, "Looking up TeamMember for Team, Member") {
				// 	callerKey := datastore.NewKey(c, "Member", "", postData.CalledBy, nil)

				// 	teamMember.CalledBy = callerKey
				// }
			} else {
				errMsg = "User does not have rights to perform action"
			}
		}
	}

	if len(errMsg) > 0 {
		context = struct {
			Error   bool
			Message string
		}{
			true,
			errMsg,
		}
	}
	returnJSONorErrorToResponse(context, c, w, r)
}
