package certapps

import (
	//	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
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

type MainAppContext struct {
	LogInOutLink string
	LogInOutText string
	Member       *Member
	LoggedIn     bool
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

	Created    time.Time
	CreatedBy  *datastore.Key
	Modified   time.Time
	ModifiedBy *datastore.Key
}

type Member struct {
	FirstName string
	LastName  string
	Cell      string
	HomePhone string

	Email  string
	Email2 string

	ShowCell  bool
	ShowEmail bool

	HomeAddress *datastore.Key
	UserID      string

	Created    time.Time
	CreatedBy  *datastore.Key
	Modified   time.Time
	ModifiedBy *datastore.Key

	LastLogin time.Time

	Key *datastore.Key `datastore:"-"`
}

type MemberLogin struct {
	MemberKey *datastore.Key
	Login     time.Time
	IP        string
}

var jsonTemplate = textT.Must(textT.New("json").Parse("{\"data\": {{.Data}} }"))
var jsonErrTemplate = textT.Must(textT.New("jsonErr").Parse("{\"error\": true, {{.Data}} }"))

func init() {
	http.HandleFunc("/memberData", memberData)
	http.HandleFunc("/audit", audit)
	http.HandleFunc("/sign", sign)
	http.HandleFunc("/", root)
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

			http.Redirect(w, r, url, http.StatusFound)
		}
	}

	if noErrMsg(err, w, c, "could not unescape url") {
		c.Debugf("url %v", rawurl)
	}
}

func memberData(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)

	context := MainAppContext{
		LogInOutLink: "boo",
		LogInOutText: "foo",
		LoggedIn:     false,
	}

	jsonC := JSONContext{}
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

		context.Member = getMember(c, u, r, w)
	}

	//bArr, memberJSONerr := json.MarshalIndent(context, "", "\t")
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
						FirstName: "",
						LastName:  "",
						Email:     u.Email,
						Cell:      "",

						UserID: u.ID,

						Created:  time.Now(),
						Modified: time.Now(),
						// need to be nil because we don't have the member key to use yet
						CreatedBy:  nil,
						ModifiedBy: nil,
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
					mem.Key = keys[0]
					found = true
					mem.UserID = u.ID

					c.Infof("Adding User.ID: %v, to Member with Key: %v", mem.UserID, mem.Key)
					datastore.Put(c, mem.Key, mem)
				}
			}
		} else {
			// found by id
			mem = &member[0]
			mem.Key = keys[0]
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
			c.Infof("existing Member found: %d", len(member))
			c.Debugf("with key: %v", mem.Key)
			c.Infof("existing: %+v", mem)
			c.Infof("existing.CreatedBy: %+v", mem.CreatedBy)
			c.Infof("existing.HomeAddress: %+v", mem.HomeAddress)
			c.Infof("user: %+v", *u)
		} else {
			c.Infof("no mem")
		}
	}

	return mem
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

var guestbookTemplate = template.Must(template.New("book").Parse(guestbookTemplateHTML))

const guestbookTemplateHTML = `
<html>
  <body>
  	<a href="{{.LogInOutLink}}">{{.LogInOutText}}</a>
  	  {{range .Greetings}}
      {{with .Author}}
        <p><b>{{.}}</b> wrote:</p>
      {{else}}
        <p>An anonymous person wrote:</p>
      {{end}}
      <pre>{{.Content}}</pre>
    {{end}}
    <form action="/sign" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Sign Guestbook"></div>
    </form>
  </body>
</html>
`

func sign(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	g := Greeting{
		Content: r.FormValue("content"),
		Date:    time.Now(),
	}
	if u := user.Current(c); u != nil {
		g.Author = u.String()
	}
	// We set the same parent key on every Greeting entity to ensure each Greeting
	// is in the same entity group. Queries across the single entity group
	// will be consistent. However, the write rate to a single entity group
	// should be limited to ~1/second.
	key := datastore.NewIncompleteKey(c, "Greeting", guestbookKey(c))
	_, err := datastore.Put(c, key, &g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
