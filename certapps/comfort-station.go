package certapps

import (
	"encoding/json"
	"fmt"
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

type ComfortStationEditor struct {
	MemberKey *datastore.Key

	Audit
}

type ComfortStationHours struct {
	Open  time.Time `json:"open"`
	Close time.Time `json:"close"`

	TeamKey           *datastore.Key
	ComfortStationKey *datastore.Key
	EditorKey         *datastore.Key

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

		mem, err := getMemberFromUser(c, u)
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
	var stations []*ComfortStation
	var keys []*datastore.Key
	var context = struct {
		Locations []*ComfortStation `json:"locations"`
	}{}

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

		context.Locations = stations
	}

	return context
}

func apiComfortStationGet(u *user.User, c appengine.Context, w http.ResponseWriter, r *http.Request) interface{} {
	c.Debugf("apiComfortStationGet")
	c.Debugf("Request.URL %+v, LastIndex: %d, len: %d", r.URL, strings.LastIndex(r.URL.String(), "/"), len(r.URL.String()))

	var context interface{}

	if strings.LastIndex(r.URL.String(), "/") == len(r.URL.String())-1 {
		context = apiComfortStationGetAll(u, c, w, r)
	}

	intId, err := getIdFromRequest(r, c)

	context = struct {
		error   bool
		message string
	}{
		true,
		"Could not parse Id",
	}

	if noErrMsg(err, w, c, "Getting ID from Request") {
		station := new(ComfortStation)
		c.Debugf("load for %d", intId)

		csKey := datastore.NewKey(c, "ComfortStation", "", intId, nil)

		err = datastore.Get(c, csKey, station)

		if noErrMsg(err, w, c, fmt.Sprintf("Loading ComfortStation by id: %d", intId)) {
			station.setKey(csKey)

			context = struct {
				Station *ComfortStation `json:"station"`
			}{
				station,
			}
		} else {
			context = struct {
				error bool
			}{true}
		}
	}

	//returnJSONorErrorToResponse(context, c, w, r)

	return context
}

func apiComfortStationsAll(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	//u := user.Current(c)
	var context interface{}
	/*	var mem *Member
		var postData struct {
			CClass *CertificationClass
			Team   *Team
		}*/

	c.Infof("apiComfortStationsAll")
	u := user.Current(c)

	context = apiComfortStationGetAll(u, c, w, r)

	returnJSONorErrorToResponse(context, c, w, r)
}

func (cs *ComfortStation) save(mem *Member, c appengine.Context) error {
	c.Infof("ComfortStation.save")

	if cs.Key == nil {
		cs.Key = datastore.NewKey(c, "ComfortStation", "", cs.KeyID, nil)
		c.Debugf("Using KeyID %d, to create Key %+v", cs.KeyID, cs.Key)
	}

	if cs.Line1 != "" {
		err := cs.Location.geocode(c)
		checkErr(err, nil, c, "Geocoding Comfort Station address")
	}

	cs.setAudits(mem)

	newKey, err := datastore.Put(c, cs.Key, cs)

	cs.setKey(newKey)

	return err
}
