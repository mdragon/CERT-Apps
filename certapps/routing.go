package certapps

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/mjibson/appstats"
)

func init() {
	http.Handle("/api/trainingTopic/save", appstats.NewHandler(apiTrainingTopicSave))
	http.Handle("/api/certificationClass", appstats.NewHandler(apiCertificationClassGet))
	http.Handle("/api/certificationClass/all", appstats.NewHandler(apiCertificationClassGetAll))
	http.Handle("/api/certificationClass/save", appstats.NewHandler(apiCertificationClassSave))
	http.Handle("/api/certificationClass/attendee/add", appstats.NewHandler(apiCertificationClassAttendeeAdd))

	http.Handle("/api/comfort-station/", appstats.NewHandler(apiComfortStation))
	http.Handle("/api/comfort-stations/", appstats.NewHandler(apiComfortStationsAll))

	http.Handle("/api/member/search", appstats.NewHandler(apiMemberSearch))
	http.Handle("/api/member/calledBy", appstats.NewHandler(apiMemberCalledBy))

	http.Handle("/audit", appstats.NewHandler(audit))
	http.Handle("/certification", appstats.NewHandler(certificationGet))
	http.Handle("/certifications/all", appstats.NewHandler(certificationsGetAll))
	http.Handle("/certification/save", appstats.NewHandler(certificationSave))
	http.Handle("/eventA", appstats.NewHandler(eventA))
	http.Handle("/event", appstats.NewHandler(event))
	http.Handle("/event/reminders/send", appstats.NewHandler(remindersSend))
	http.Handle("/event/save", appstats.NewHandler(eventSave))
	http.Handle("/events", appstats.NewHandler(events))
	http.Handle("/member", appstats.NewHandler(memberData))
	http.Handle("/member/save", appstats.NewHandler(memberSave))
	http.Handle("/response", appstats.NewHandler(response))
	http.Handle("/response/save", appstats.NewHandler(responseSave))
	http.Handle("/responses/create", appstats.NewHandler(responsesCreate))
	http.Handle("/team", appstats.NewHandler(teamData))
	http.Handle("/team/roster", appstats.NewHandler(teamRoster))
	http.Handle("/team/roster/import", appstats.NewHandler(membersImport))
	http.Handle("/fix/teamMembers/without/Memeber", appstats.NewHandler(fixData))
	http.Handle("/fix/members/geocode", appstats.NewHandler(resaveMembers))
	http.Handle("/address", appstats.NewHandler(locationLookupHandler))
	http.Handle("/setup", appstats.NewHandler(initialSetup))
	http.Handle("/", appstats.NewHandler(root))

	rand.Seed(time.Now().UnixNano())
}
