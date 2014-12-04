package certapps

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type ComfortStation struct {
	Name  string
	Notes string

	EditKeys []int64
	TeamKey  *datastore.Key

	Location

	Audit
}

type ComfortStationHours struct {
	Open  time.Time
	Close time.Time

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

	var postData struct {
		ComfortStation *ComfortStation
		Team           *Team
	}

	decoder := json.NewDecoder(r.Body)
	jsonDecodeErr := decoder.Decode(&postData)

	if jsonDecodeErr == io.EOF {
		c.Infof("EOF, should it be?")
	} else if noErrMsg(jsonDecodeErr, nil, c, "Parsing json from body") {
		c.Infof("JSON from request: %+v", postData)
	}

	return postData
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
	/*	var mem *Member
		var postData struct {
			CClass *CertificationClass
			Team   *Team
		}*/

	c.Infof("apiComfortStationsAll")

	path := r.URL.Path
	lastSlashIdx := strings.LastIndex(path, "/") + 1
	strId := path[lastSlashIdx:]
	intId, _ := strconv.ParseInt(strId, 0, 0)

	c.Debugf("lastSlashIdx: %d, strId: %s, intId: %d", lastSlashIdx, strId, intId)

	context = struct {
		Locations []*ComfortStation `json:"locations"`
	}{}

	returnJSONorErrorToResponse(context, c, w, r)
}

func (cs *ComfortStation) save(c appengine.Context) {

}
