package certapps

import (
	"appengine/datastore"

	"github.com/GoogleCloudPlatform/go-endpoints/endpoints"
)

type WhereAmIService struct {
}

type WhereAmIList struct {
	Entries []*WhereAmI `json:"entries"`
}

type WhereAmIListReq struct {
	Limit int `json:"limit" endpoints:"d=500"`
}

// List responds with a list of all greetings ordered by Date field.
// Most recent greets come first.
func (gs *WhereAmIService) List(c endpoints.Context, r *WhereAmIListReq) (*WhereAmIList, error) {
	if r.Limit <= 0 {
		r.Limit = 500
	}

	q := datastore.NewQuery("WhereAmI").Order("-Entered").Limit(r.Limit)
	wheres := make([]*WhereAmI, 0, r.Limit)
	keys, err := q.GetAll(c, &wheres)
	if err != nil {
		return nil, err
	}

	for i, k := range keys {
		w := wheres[i]
		w.Key = k
		w.KeyID = k.IntID()
	}

	return &WhereAmIList{wheres}, nil
}
