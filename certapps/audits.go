package certapps

import (
	"time"

	"appengine/datastore"
)

type Audit struct {
	Created    time.Time
	CreatedBy  *datastore.Key
	Modified   time.Time
	ModifiedBy *datastore.Key

	Key   *datastore.Key `datastore:"-"`
	KeyID int64          `datastore:"-"`

	Deleted bool
}

func (a *Audit) setAudits(m *Member) {
	a.ModifiedBy = m.Key
	a.Modified = time.Now()

	if a.CreatedBy == nil {
		a.Created = a.Modified
		a.CreatedBy = a.ModifiedBy
	}
}

func (a *Audit) setKey(key *datastore.Key) {
	a.Key = key
	if key != nil {
		a.KeyID = key.IntID()
	}
}
