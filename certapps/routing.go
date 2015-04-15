package certapps

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"appengine"
	//	"github.com/mjibson/appstats"

	"github.com/GoogleCloudPlatform/go-endpoints/endpoints"
)

type handler struct {
	f func(appengine.Context, http.ResponseWriter, *http.Request)
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	h.f(c, w, r)
}

func NewHandler(f func(appengine.Context, http.ResponseWriter, *http.Request)) http.Handler {
	return handler{
		f: f,
	}
}

func newHandler(f func(appengine.Context, http.ResponseWriter, *http.Request)) http.Handler {
	//return appstats.NewHandler(f)
	return NewHandler(f)
}

func init() {
	http.Handle("/api/certification", newHandler(apiCertificationGet))

	http.Handle("/api/certificationClass", newHandler(apiCertificationClassGet))
	http.Handle("/api/certificationClass/all", newHandler(apiCertificationClassGetAll))
	http.Handle("/api/certificationClass/save", newHandler(apiCertificationClassSave))
	http.Handle("/api/certificationClass/attendee/add", newHandler(apiCertificationClassAttendeeAdd))

	http.Handle("/api/comfort-station/", newHandler(apiComfortStation))
	http.Handle("/api/comfort-stations/", newHandler(apiComfortStationsAll))

	http.Handle("/api/member", newHandler(memberData))
	http.Handle("/api/member/save", newHandler(apiMemberSave))
	http.Handle("/api/member/toggle-active", newHandler(apiMemberToggleActive))
	http.Handle("/api/member/toggle-enabled", newHandler(apiMemberToggleEnabled))
	http.Handle("/api/member/search", newHandler(apiMemberSearch))
	http.Handle("/api/member/calledBy", newHandler(apiMemberCalledBy))

	http.Handle("/api/team/save", newHandler(apiTeamSave))

	http.Handle("/api/trainingTopic/save", newHandler(apiTrainingTopicSave))

	//http.Handle("/api/where/lookup", newHandler(apiWhereLookup))

	http.Handle("/audit", newHandler(audit))
	http.Handle("/certifications/all", newHandler(certificationsGetAll))
	http.Handle("/certification/save", newHandler(certificationSave))
	http.Handle("/eventA", newHandler(eventA))
	http.Handle("/event", newHandler(event))
	http.Handle("/event/reminders/send", newHandler(remindersSend))
	http.Handle("/event/save", newHandler(eventSave))
	http.Handle("/events", newHandler(events))
	http.Handle("/response", newHandler(response))
	http.Handle("/response/save", newHandler(responseSave))
	http.Handle("/responses/create", newHandler(responsesCreate))
	http.Handle("/team", newHandler(teamData))
	http.Handle("/team/roster", newHandler(teamRoster))
	http.Handle("/team/roster/import", newHandler(membersImport))
	http.Handle("/fix/teamMembers/without/Member", newHandler(fixData))
	http.Handle("/fix/members/geocode", newHandler(resaveMembers))
	http.Handle("/address", newHandler(locationLookupHandler))
	http.Handle("/setup", newHandler(initialSetup))
	http.Handle("/", newHandler(root))

	http.Handle("/whereami", newHandler(whereAmI))
	http.Handle("/whereami/save", newHandler(whereAmISave))

	whereService := &WhereAmIService{}
	// was api, err
	_, err := endpoints.RegisterService(whereService, "where", "v1", "Where Am I? API", true)
	if err != nil {
		log.Fatalf("Register service: %v", err)
	}

	// register := func(orig, name, method, path, desc string) {
	// 	m := api.MethodByName(orig)
	// 	if m == nil {
	// 		log.Fatalf("Missing method %s", orig)
	// 	}
	// 	i := m.Info()
	// 	i.Name, i.HTTPMethod, i.Path, i.Desc = name, method, path, desc
	// }

	// register("List", "greets.list", "GET", "greetings", "List most recent greetings.")
	// register("Add", "greets.add", "PUT", "greetings", "Add a greeting.")
	// register("Count", "greets.count", "GET", "greetings/count", "Count all greetings.")
	endpoints.HandleHTTP()

	rand.Seed(time.Now().UnixNano())
}
