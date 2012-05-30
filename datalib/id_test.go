package githubdata

import (
	"encoding/json"
	"testing"
)

var sample = []byte(`{
   "actor": "bobkarreman",
   "actor_attributes": {
       "gravatar_id": "d4ce883db263cb0eb6d8f4e9a10efa1c",
       "login": "bobkarreman",
       "name": "Bob Karreman",
       "type": "User"
   },
   "created_at": "2012/04/05 03:40:13 -0700",
   "payload": {
       "head": "284f9a67ac271074c799eff58f7ae0cd8a980adf",
       "ref": "refs/heads/develop",
       "shas": [
           [
               "284f9a67ac271074c799eff58f7ae0cd8a980adf",
               "bob@karreman.org",
               "Add event handler in pubsub for messge_event and fixed node to be Optional in create node",
               "Bob Karreman",
               true
           ]
       ],
       "size": 1
   },
   "public": true,
   "repository": {
       "created_at": "2011/12/19 04:40:59 -0800",
       "description": "Implementation of XEP-0080",
       "fork": true,
       "forks": 0,
       "has_downloads": true,
       "has_issues": false,
       "has_wiki": true,
       "homepage": "",
       "language": "Python",
       "name": "SleekXMPP--XEP-0080-",
       "open_issues": 0,
       "organization": "calendar42",
       "owner": "calendar42",
       "private": false,
       "pushed_at": "2012/04/05 03:40:12 -0700",
       "size": 124,
       "url": "https://github.com/calendar42/SleekXMPP--XEP-0080-",
       "watchers": 1
   },
   "type": "PushEvent",
   "url": "https://github.com/calendar42/SleekXMPP--XEP-0080-/compare/3163bd1177...284f9a67ac"
}`)

func getDoc(t *testing.T) map[string]interface{} {
	doc := map[string]interface{}{}
	err := json.Unmarshal(sample, &doc)
	if err != nil {
		t.Errorf("Error unmarshaling sample json: %v", err)
	}
	return doc
}

func TestIDGeneration(t *testing.T) {
	doc := getDoc(t)
	got := GenerateId(doc)
	expected := "2012-04-05T03-40-13-bc07103e24aec34c"
	if got != expected {
		t.Fatalf("Expected `%v', got `%v'", expected, got)
	}

}

type timeTestData struct {
	Expected string
	Input    string
}

var testData = []timeTestData{
	{"2012 05 29 22 45 20", "2012-05-29T22:45:20Z"},
	{"2012 04 05 03 40 12", "2012/04/05 03:40:12 -0700"},
	{"2012 05 29 15 10 06", "2012-05-29T15:10:06-07:00"},
}

func TestFormats(t *testing.T) {
	for _, ttd := range testData {
		date, err := ParseDate(ttd.Input)
		s := date.Format("2006 01 02 15 04 05")
		if err != nil {
			t.Fatalf("Error parsing %v (%v)",
				ttd.Input, err)
		} else if s != ttd.Expected {
			t.Fatalf("Expected %v for %v, got %v",
				ttd.Expected, ttd.Input, s)
		}
	}
}

func TestSupplementalFields(t *testing.T) {
	doc := getDoc(t)
	dow := GetDayOfWeek(doc)
	if dow != "Thursday" {
		t.Fatalf("Expected Tuesday, got %v", dow)
	}
	got := ExplodeDate(doc)
	exp := []int{2012, 4, 5, 3, 40, 13}
	if len(got) != len(exp) {
		t.Fatalf("Expected %d items, got %d - %v",
			len(exp), len(got), got)
	}
	for i := range exp {
		if exp[i] != got[i] {
			t.Fatalf("Expected %d, got %d at %d: %v / %v",
				exp[i], got[i], i, exp, got)
		}
	}
}
