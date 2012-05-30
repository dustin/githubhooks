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

func Dochash(doc map[string]interface{}) string {
	h := fnv.New64()
	fields := []string{
		"actor", "created_at",
		"description", "type", "url",
	}
	for _, f := range fields {
		fmt.Fprintf(h, "%v", doc[f])
	}
	switch repo := doc["repository"].(type) {
	case map[string]interface{}:
		morefields := []string{"description",
			"name", "owner", "organization",
			"pushed_at",
		}
		for _, f := range morefields {
			fmt.Fprintf(h, "%v", repo[f])
		}
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
