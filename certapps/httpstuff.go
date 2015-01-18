package certapps

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	textT "text/template"

	"appengine"
)

var jsonTemplate = textT.Must(textT.New("json").Parse("{\"data\": {{.Data}} }"))
var jsonErrTemplate = textT.Must(textT.New("jsonErr").Parse("{\"error\": true, {{.Data}} }"))

func getIdFromRequest(r *http.Request, c appengine.Context) (int64, error) {
	return getIdFromURL(r.URL, c)
}

func getIdFromURL(u *url.URL, c appengine.Context) (int64, error) {
	path := u.Path

	return getIdFromPath(path, c)
}

func getIdFromPath(path string, c appengine.Context) (int64, error) {
	lastSlashIdx := strings.LastIndex(path, "/") + 1
	strId := path[lastSlashIdx:]

	c.Debugf("getIdFromURL path: %s, lastSlash: %d, len: %d, strId: %s", path, lastSlashIdx, len(path), strId)

	if lastSlashIdx == len(path) {
		return 0, errors.New("Path has no portion after the last slash: " + path)
	}

	intId, err := strconv.ParseInt(strId, 0, 0)

	c.Debugf("lastSlashIdx: %d, strId: %s, intId: %d", lastSlashIdx, strId, intId)

	return intId, err
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
