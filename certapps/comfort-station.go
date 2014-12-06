package certapps

import (
	"encoding/json"
	//"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type ComfortStation struct {
	Name  string `json:"name"`
	Notes string `json:"notes"`

	EditKeys []*datastore.Key
	TeamKey  *datastore.Key

	Contact string `json:"contact"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`

	Location

	Audit
}

type ComfortStationHours struct {
	Open  time.Time `json:"open"`
	Close time.Time `json:"close"`

	TeamKey           *datastore.Key
	ComfortStationKey *datastore.Key

	Audit
}

func apiComfortStation(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	var context interface{}
	/*	var mem *Member
		var postData struct {
			CClass *CertificationClass
			Team   *Team
		}*/

	c.Infof("apiComfortStation")

	switch r.Method {
	case "POST":
		context = apiComfortStationSave(u, c, w, r)
	default:
		context = apiComfortStationGet(u, c, w, r)
	}

	// err := parseJSON(postData, r, c)
	// if noErrMsg(err, w, c, "Parsing JSON") {
	/*
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

	*/
	returnJSONorErrorToResponse(context, c, w, r)
}

func apiComfortStationSave(u *user.User, c appengine.Context, w http.ResponseWriter, r *http.Request) interface{} {
	// err := parseJSON(postData, r, c)
	// if noErrMsg(err, w, c, "Parsing JSON") {

	c.Debugf("apiComfortStationSave")

	var pd struct {
		Station *ComfortStation
		Team    *Team
	}

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&pd)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, nil, c, "Parsing json from body") {
		c.Infof("JSON from request: Station: %+v, Team: %+v", pd.Station, pd.Team)

		mem, err := getMemberFromUser(c, u, w, r)
		if noErrMsg(err, w, c, "Getting member from user") {
			pd.Station.TeamKey = datastore.NewKey(c, "Team", "", pd.Team.KeyID, nil)

			err = pd.Station.save(mem, c)

			if noErrMsg(err, w, c, "Saving ComfortStation") {

			}
		}
	}

	return pd
}

func apiComfortStationGetAll(u *user.User, c appengine.Context, w http.ResponseWriter, r *http.Request) interface{} {
	c.Debugf("apiComfortStationGetAll")
	return struct{ GetAll bool }{true}
}

func apiComfortStationGet(u *user.User, c appengine.Context, w http.ResponseWriter, r *http.Request) interface{} {
	c.Debugf("apiComfortStationGet")
	c.Debugf("Request.URL %+v, LastIndex: %d, len: %d", r.URL, strings.LastIndex(r.URL.String(), "/"), len(r.URL.String()))

	if strings.LastIndex(r.URL.String(), "/") == len(r.URL.String())-1 {
		return apiComfortStationGetAll(u, c, w, r)
	}

	return struct{ Get bool }{true}
}

func apiComfortStationsAll(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	//u := user.Current(c)
	var context interface{}
	var stations []*ComfortStation
	var keys []*datastore.Key
	/*	var mem *Member
		var postData struct {
			CClass *CertificationClass
			Team   *Team
		}*/

	c.Infof("apiComfortStationsAll")

	intId, err := getIdFromRequest(r, c)

	if noErrMsg(err, w, c, "Getting ID from Request") {
		c.Debugf("load for %d", intId)

		teamKey := datastore.NewKey(c, "Team", "", intId, nil)

		q := datastore.NewQuery("ComfortStation").Filter("TeamKey =", teamKey)
		keys, err = q.GetAll(c, &stations)

		for x, _ := range stations {
			s := stations[x]
			s.setKey(keys[x])
		}

		context = struct {
			Locations []*ComfortStation `json:"locations"`
		}{
			stations,
		}
	}

	returnJSONorErrorToResponse(context, c, w, r)
}

func (cs *ComfortStation) save(mem *Member, c appengine.Context) error {
	c.Infof("ComfortStation.save")

	if cs.Key == nil {
		cs.Key = datastore.NewKey(c, "ComfortStation", "", cs.KeyID, nil)
		c.Debugf("Using KeyID %d, to create Key %+v", cs.KeyID, cs.Key)
	}

	cs.setAudits(mem)

	newKey, err := datastore.Put(c, cs.Key, cs)

	cs.setKey(newKey)

	return err
}
