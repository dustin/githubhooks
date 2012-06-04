package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

var sampleConfigString = `{
    "ByActor": {
        "dustin": [
            "http://localhost:6666/actor/dustin",
            "http://localhost:6666/group/couchbase"
        ]
    },
    "ByOwnerRepo": {
       "mxcl/homebrew": [
         "http://localhost:6666/ownerrepo/homebrew"
       ],
       "mxcl/homebrew:PushEvent": [
         "http://localhost:6666/ownerrepo/homebrew/push"
       ]
    }
}`

var sampleConfig jsonConfig

func init() {
	err := json.Unmarshal([]byte(sampleConfigString), &sampleConfig)
	if err != nil {
		panic("Couldn't parse json: " + err.Error())
	}
}

func TestEmpty(t *testing.T) {

	ev := event{0, "id", map[string]interface{}{},
		"x", "y", "z", "IssuesEvent"}

	l := sampleConfig.ByActor(ev)
	if len(l) != 0 {
		t.Fatalf("Expected empty result for actor x, got %v", l)
	}

	l = sampleConfig.ByOwner(ev)
	if len(l) != 0 {
		t.Fatalf("Expected empty result for owner x, got %v", l)
	}

}

func TestSimpleActor(t *testing.T) {
	ev := event{0, "id", map[string]interface{}{},
		"dustin", "y", "z", "IssuesEvent"}

	l := sampleConfig.ByActor(ev)
	if !reflect.DeepEqual(l, []string{
		"http://localhost:6666/actor/dustin",
		"http://localhost:6666/group/couchbase",
	}) {
		t.Fatalf("Got %v", l)
	}
}

func TestSimpleOwnerRepo(t *testing.T) {
	ev := event{0, "id", map[string]interface{}{},
		"bob", "mxcl", "homebrew", "IssuesEvent"}

	l := sampleConfig.ByOwnerRepo(ev)
	if !reflect.DeepEqual(l, []string{
		"http://localhost:6666/ownerrepo/homebrew",
	}) {
		t.Fatalf("Got %v", l)
	}
}

func TestSimpleOwnerRepoPush(t *testing.T) {
	ev := event{0, "id", map[string]interface{}{},
		"bob", "mxcl", "homebrew", "PushEvent"}

	l := sampleConfig.ByOwnerRepo(ev)
	if !reflect.DeepEqual(l, []string{
		"http://localhost:6666/ownerrepo/homebrew",
		"http://localhost:6666/ownerrepo/homebrew/push",
	}) {
		t.Fatalf("Got %v", l)
	}
}
