package certapps

import (
	"html/template"
	"net/http"
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

type TContext struct {
	Greetings    []Greeting
	LogInOutLink string
	LogInOutText string
}

type Location struct {
	Line1 string
	Line2 string
	City  string
	State string
	Zip   string
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

	HomeAddress Location

	Created    time.Time
	CreatedBy  user.User
	Modified   time.Time
	ModifiedBy user.User
}

func init() {
	http.HandleFunc("/guest", guest)
	http.HandleFunc("/sign", sign)
	http.HandleFunc("/", root)
}

// guestbookKey returns the key used for all guestbook entries.
func guestbookKey(c appengine.Context) *datastore.Key {
	// The string "default_guestbook" here could be varied to have multiple guestbooks.
	return datastore.NewKey(c, "Guestbook", "default_guestbook", 0, nil)
}

func root(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/html/app.htm", http.StatusFound)
}

func guest(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)

	// Ancestor queries, as shown here, are strongly consistent with the High
	// Replication Datastore. Queries that span entity groups are eventually
	// consistent. If we omitted the .Ancestor from this query there would be
	// a slight chance that Greeting that had just been written would not
	// show up in a query.
	q := datastore.NewQuery("Greeting").Ancestor(guestbookKey(c)).Order("-Date").Limit(10)
	greetings := make([]Greeting, 0, 10)
	context := TContext{
		Greetings:    greetings,
		LogInOutLink: "boo",
		LogInOutText: "foo",
	}

	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		context.LogInOutLink = url
		context.LogInOutText = "Log In"
	} else {
		url, err := user.LogoutURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		context.LogInOutLink = url
		context.LogInOutText = "Log Out"

		l := Location{
			Line1: "123 Main St",
			City:  "Anytown",
			State: "NJ",
			Zip:   "55555",
		}
		m := Member{
			FirstName:   "Matt",
			LastName:    "Dragon",
			Email:       "foo@example.com",
			Cell:        "555-555-5555",
			HomeAddress: l,

			Created:    time.Now,
			CreatedBy:  u,
			Modified:   time.Now,
			ModifiedBy: u,
		}

	}

	if _, err := q.GetAll(c, &context.Greetings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := guestbookTemplate.Execute(w, context); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func checkErr(e error) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
