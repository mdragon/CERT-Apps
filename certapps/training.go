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
		CClass        *CertificationClass
		Team          *Team
		Certification *Certification
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

		mem, _ = getMemberFromUser(c, u)

		teamKey := datastore.NewKey(c, "Team", "", postData.Team.KeyID, nil)
		certKey := datastore.NewKey(c, "Certification", "", postData.Certification.KeyID, nil)

		postData.CClass.TeamKey = teamKey
		postData.CClass.CertificationKey = certKey

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

	mem, _ = getMemberFromUser(c, u)

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

func (team *Team) getCertificationClasses(c appengine.Context) ([]*CertificationClass, error) {
	if team.Key == nil {
		team.Key = datastore.NewKey(c, "Team", "", team.KeyID, nil)
	}
	var results []*CertificationClass

	query := datastore.NewQuery("CertificationClass").Filter("TeamKey =", team.Key)

	keys, err := query.GetAll(c, &results)

	err = checkCannotLoadField(err, c)

	if noErrMsg(err, nil, c, "Getting All CertificationClasses for Team") {
		for idx, _ := range results {
			e := results[idx]
			key := keys[idx]

			e.setKey(key)
		}
	}

	return results, err
}

func getCertifications(c appengine.Context) ([]*Certification, error) {
	var results []*Certification

	query := datastore.NewQuery("Certification")

	keys, err := query.GetAll(c, &results)

	err = checkCannotLoadField(err, c)

	if noErrMsg(err, nil, c, "Getting All Certifications") {
		for idx, _ := range results {
			e := results[idx]
			key := keys[idx]

			e.setKey(key)
		}
	}

	return results, err
}

func apiCertificationClassGetAll(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	var mem *Member

	id := parseIntFormVal("teamID", r)

	c.Infof("certificationClassGetAll, id: %d", id)

	if id != 0 {
		mem, _ = getMemberFromUser(c, u)

		if mem != nil {
			team := &Team{Audit: Audit{KeyID: id}}

			classesC := make(chan []*CertificationClass)
			certsC := make(chan []*Certification)
			errC := make(chan error)

			go func() {
				classes, err := team.getCertificationClasses(c)

				classesC <- classes
				errC <- err
			}()

			classes := <-classesC

			go func() {
				certs, err := getCertifications(c)

				certsC <- certs
				errC <- err
			}()

			certifications := <-certsC

			err1 := <-errC
			err2 := <-errC

			if noErrMsg(err1, w, c, "apiCertificationClassGetAll error 1") {
				if noErrMsg(err2, w, c, "apiCertificationClassGetAll error 2") {
					context = struct {
						CClasses       []*CertificationClass
						Certifications []*Certification
					}{
						classes,
						certifications,
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

	c.Infof("apiCertificationClassAddAttendee")

	mem, memberErr := getMemberFromUser(c, u)

	if noErrMsg(memberErr, w, c, "getMemberFromUser") {
		decoder := json.NewDecoder(r.Body)
		jsonDecodeErr := decoder.Decode(&postData)

		if jsonDecodeErr == io.EOF {
			c.Infof("EOF, should it be?")
		} else if noErrMsg(jsonDecodeErr, nil, c, "Parsing json from body") {
			c.Infof("JSON from request: KeyIDs Member: %d, CClass: %d", postData.Member.KeyID, postData.CClass.KeyID)

			classLookupErr := postData.CClass.lookup(postData.CClass.KeyID, mem, c)

			classLookupErr = checkCannotLoadField(classLookupErr, c)

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
