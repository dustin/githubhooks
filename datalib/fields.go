// Tools for handling github data.
package githubdata

import (
	"fmt"
	"hash/fnv"
	"log"
	"strconv"
	"strings"
	"time"
)

func maybeFatal(err error, msg ...string) {
	if err != nil {
		if len(msg) > 0 {
			log.Fatalf("Error: %v (%v)", err, msg)
		} else {
			log.Fatalf("Error:  %v", err)
		}
	}
}

func findItem(key []string, m map[string]interface{}) string {
	if len(key) == 1 {
		return fmt.Sprintf("%v", m[key[0]])
	}
	switch v := m[key[0]].(type) {
	case nil:
		return "<nil>"
	case map[string]interface{}:
		return findItem(key[1:], v)
	default:
		return ""
	}
	panic("How did we get here?")
}

func Dochash(doc map[string]interface{}) string {
	h := fnv.New64()
	fields := [][]string{
		{"actor"}, {"created_at"},
		{"description"}, {"type"}, {"url"},
		{"repository", "description"},
		{"repository", "name"},
		{"repository", "owner"},
		{"repository", "organization"},
		{"payload", "head"},
		{"payload", "ref"},
		{"payload", "ref_type"},
		{"payload", "size"},
		{"payload", "target", "login"},
		{"payload", "forkee", "url"},
		{"payload", "action"},
	}
	for _, f := range fields {
		fmt.Fprintf(h, "%s", findItem(f, doc))
	}
	return fmt.Sprintf("%x", h.Sum64())
}

func GenerateId(doc map[string]interface{}) (rv string) {
	parts := []string{}
	ts := doc["created_at"].(string)
	t, err := ParseDate(ts)
	maybeFatal(err)

	parts = append(parts, t.Format("2006-01-02T15-04-05"))

	parts = append(parts, Dochash(doc))

	rv = strings.Join(parts, "-")

	return
}

func ParseDate(s string) (time.Time, error) {
	formats := []string{
		"2006/01/02 15:04:05 -0700",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
	}
	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("Could not parse %v", s)
}

func GetDayOfWeek(doc map[string]interface{}) string {
	date, err := ParseDate(doc["created_at"].(string))
	maybeFatal(err)
	return date.Format("Monday")
}

func ExplodeDate(doc map[string]interface{}) []int {
	date, err := ParseDate(doc["created_at"].(string))
	maybeFatal(err)
	s := date.Format("2006 01 02 15 04 05")
	sexp := strings.Split(s, " ")
	rv := make([]int, 0, len(sexp))
	for _, part := range sexp {
		i, err := strconv.ParseInt(part, 10, 32)
		maybeFatal(err)
		rv = append(rv, int(i))
	}
	return rv
}

func UpdateWithCustomFields(doc map[string]interface{}) {
	doc["_id"] = GenerateId(doc)
	doc["dow"] = GetDayOfWeek(doc)
	doc["expldate"] = ExplodeDate(doc)
}
