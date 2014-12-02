package certapps

import (
	"encoding/json"
	"io"
	"net/http"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

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

func apiCertificationClassAttendeeAdd(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}

	var postData struct {
		CClass *CertificationClass
		Member *Member
	}

	c.Infof("certificationClassAddAttendee")

	mem, memberErr := getMemberFromUser(c, u, w, r)

	if noErrMsg(memberErr, w, c, "getMemberFromUser") {
		decoder := json.NewDecoder(r.Body)
		jsonDecodeErr := decoder.Decode(&postData)

		if jsonDecodeErr == io.EOF {
			c.Infof("EOF, should it be?")
		} else if noErrMsg(jsonDecodeErr, nil, c, "Parsing json from body") {
			c.Infof("JSON from request: KeyIDs Member: %d, CClass: %d", postData.Member.KeyID, postData.CClass.KeyID)

			classLookupErr := postData.CClass.lookup(postData.CClass.KeyID, mem, c)

			if noErrMsg(classLookupErr, w, c, "Looking up Certification Class") {
				postData.Member.Key = datastore.NewKey(c, "Member", "", postData.Member.KeyID, nil)

				if postData.Member.Key.Incomplete() == false {

					unique := true
					for _, attendee := range postData.CClass.Attendees {
						if attendee.IntID() == postData.Member.KeyID {
							unique = false
							break
						}
					}

					if unique {
						postData.CClass.Attendees = append(postData.CClass.Attendees, postData.Member.Key)

						classSaveErr := postData.CClass.save(mem, c)

						if noErrMsg(classSaveErr, w, c, "Saving CClass with Member") {
							context = struct {
								Added bool
							}{
								true,
							}
						}
					}
				} else {
					c.Errorf("Member passed is not a valid Key ID, it is incomplete, ie 0")
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
	}

	returnJSONorErrorToResponse(context, c, w, r)
}
