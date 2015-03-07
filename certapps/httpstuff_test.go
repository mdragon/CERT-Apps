package certapps

import (
	//"reflect"
	"testing"

	"appengine"
	"appengine/aetest"
)

var c = getContext(nil)

func TestGetIdFromPath_trailingSlash(t *testing.T) {
	path := "/api/comfort-stations/"

	intID, err := getIdFromPath(path, c)

	if intID != 0 {
		t.Errorf("Expecting an zero value, didn't get one: %d, %v", intID, err)
	}

	if err == nil {
		t.Errorf("Expecting an error, didn't get one: %d, %v", intID, err)
	} else {
		t.Logf("Got error %v", err)
	}
}

func TestGetIdFromPath_good(t *testing.T) {
	path := "/api/comfort-stations/12345"

	intID, err := getIdFromPath(path, c)

	t.Logf("Got back intID: %d", intID)

	if intID == 0 || err != nil {
		t.Errorf("Expecting a real value, no error, didn't get one: %d, %v", intID, err)
	}
}

func TestGetIdFromPath_notAnInt(t *testing.T) {
	path := "/api/comfort-stations/abc"

	intID, err := getIdFromPath(path, c)

	if intID != 0 || err == nil {
		t.Errorf("Expecting an error, didn't get one: %d, %v", intID, err)
	}
}

func getContext(t *testing.T) appengine.Context {
	c, err := aetest.NewContext(nil)
	if err != nil {
		if t != nil {
			t.Fatal(err)
		}
	}

	return c
}

/*
func TestComposeNewsletter(t *testing.T) {
	want := &mail.Message{
		Sender:  "newsletter@appspot.com",
		To:      []string{"User <user@example.com>"},
		Subject: "Weekly App Engine Update",
		Body:    "Don't forget to test your code!",
	}

	if msg := composeNewsletter(); !reflect.DeepEqual(msg, want) {
		t.Errorf("composeMessage() = %+v, want %+v", msg, want)
	}
}
*/
