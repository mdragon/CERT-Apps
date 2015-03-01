package certapps

import (
	"time"

	"appengine/datastore"
)

type Greeting struct {
	Author  string
	Content string
	Date    time.Time
}

type AuthContext struct {
	LogInOutLink string
	LogInOutText string
	LoggedIn     bool
}

type MainAppContext struct {
	Member *Member
	Team   *Team
	//User   *user.User

	AuthContext
}

type JSONContext struct {
	Data string
}

type ErrorContext struct {
	Message string
	Error   bool
}

type RemindersToSend struct {
	All                 bool
	RespondedYesOrMaybe bool
	Unsent              bool
	NoResponse          bool
}

type Location struct {
	Line1 string `json:"line1"`
	Line2 string `json:"line2"`
	City  string `json:"city"`
	State string `json:"state"`
	Zip   string `json:"zip"`

	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	PublicLatitude  float64 `json:"publicLatitude"`
	PublicLongitude float64 `json:"publicLongitude"`
}

type Team struct {
	Name string
	Location

	MembersEmail  string
	OfficersEmail string

	GoogleAPIKey string

	Audit
}

type TeamMember struct {
	TeamKey   *datastore.Key
	MemberKey *datastore.Key
	CalledBy  *datastore.Key // who would call this member would call when phone tree is activated

	Audit
}

type Person struct {
	FirstName string
	LastName  string
	Cell      string
	HomePhone string

	Email  string
	Email2 string

	Location
	UserID string

	Enabled bool

	LastLogin time.Time

	Audit
}

type Member struct {
	ShowCell  bool
	ShowEmail bool
	OKToText  bool

	RadioID string

	Town    bool
	OEM     bool
	Officer bool
	Active  bool

	CalledBy int64 `json:"calledBy", datastore:"-"`
	//TODO: this would need to move to TeamMember when multiple teams was supported
	CanLookup bool `datastore:"-"`

	Person
}

type MemberLogin struct {
	MemberKey *datastore.Key
	Login     time.Time
	IP        string
}

type Event struct {
	TeamKey *datastore.Key

	Summary     string
	Description string

	Deployment bool
	Exercise   bool
	Meeting    bool
	Training   bool

	MeetTime    time.Time
	EventStart  time.Time
	EventFinish time.Time

	EventLocation   Location
	ParkingLocation Location

	Link1 string //Deployment: IAP, Meeting: Agenda
	Link2 string //Deployment: Roster

	Responses []*MemberEvent `datastore:"-"`

	Audit
}

type MemberEvent struct {
	MemberKey *datastore.Key
	EventKey  *datastore.Key

	Attending bool
	Sure      bool

	FirstReminder  time.Time
	FirstResponded time.Time
	FirstViewed    time.Time
	LastReminder   time.Time
	LastResponded  time.Time
	LastViewed     time.Time
	Arrive         time.Time
	Depart         time.Time

	Audit
}

type Reminder struct {
	MemberEvent *datastore.Key

	Subject string
	Body    string

	Sent time.Time

	Audit
}

type Certification struct {
	Name        string `json:"name"`
	MonthsValid int64  `json:"monthsValid"`

	Audit
}

type CertificationClass struct {
	CertificationKey *datastore.Key
	TeamKey          *datastore.Key

	Name      string    `json:"name"`
	Scheduled time.Time `json:"scheduled"`
	Cancelled bool      `json:"cancelled"`

	Location

	Attendees []*datastore.Key `json:"-"`

	MembersAttending []Member `json:"attendees", datastore:"-"`

	Audit
}

type TrainingTopic struct {
	CertificationKey *datastore.Key

	Name string `json:"name"`

	EffectiveDate time.Time `json:"effectiveDate"`
	SunsetDate    time.Time `json:"sunsetDate"`

	Audit
}

type MemberClassContent struct {
	TrainingTopicKey *datastore.Key
	MemberKey        *datastore.Key

	Attended bool

	Audit
}

type MemberCertification struct {
	CertificationKey *datastore.Key
	MemberKey        *datastore.Key

	Effective time.Time

	Audit
}

type TeamCertification struct {
	TeamKey          *datastore.Key
	CertificationKey *datastore.Key
	Abbr             string
	Hidden           bool

	Audit
}

type CertificationAndTopics struct {
	Certification *Certification
	Topics        []*TrainingTopic
}

// API types

type GoogleLocationResults struct {
	Results []GoogleLocation
	Status  string
}

type GoogleLocation struct {
	Geometry          GoogleLocationGeometry
	Formatted_address string
}

type GoogleLocationGeometry struct {
	Location LatLong
}

type LatLong struct {
	Lat float64
	Lng float64
}

// end types
