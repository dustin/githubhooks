package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"code.google.com/p/dsallings-couch-go"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/dustin/githubhooks/datalib"
)

var mcdServer = flag.String("memcached", "localhost:11211", "Memcached to use.")

type event map[string]interface{}

func getData(url string) (rv []byte, err error) {
	mc := memcache.New(*mcdServer)

	itm, err := mc.Get(url)
	if err != nil {
		log.Printf("Fetching %v", url)
		resp, err := http.Get(url)
		if err != nil {
			return rv, err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)

		itm = &memcache.Item{
			Key:        url,
			Value:      data,
			Expiration: (86400 * 15),
		}
		err = mc.Set(itm)
		if err != nil {
			log.Printf("Error caching %v: %v", url, err)
		}
	}
	return itm.Value, nil
}

func fillRepository(repo map[string]interface{}) (interface{}, error) {
	url := repo["url"].(string)

	data, err := getData(url)
	if err != nil {
		return nil, err
	}
	rm := json.RawMessage(data)
	return &rm, nil
}

func process(r io.Reader, ch chan<- event) (dups int) {
	mc := memcache.New(*mcdServer)

	stuff := []event{}
	err := json.NewDecoder(r).Decode(&stuff)
	if err != nil {
		log.Printf("Error decoding stuff: %v", err)
	}
	for _, e := range stuff {
		switch i := e["actor"].(type) {
		case map[string]interface{}:
			e["actor_attributes"] = i
			actorName, ok := i["login"].(string)
			if ok {
				e["actor"] = actorName
			} else {
				e["actor"] = ""
				log.Printf("No actor name in %#v from\n%#v\n",
					i, e)
			}
		}
		switch i := e["repo"].(type) {
		case map[string]interface{}:
			val, err := fillRepository(i)
			if err == nil {
				e["repository"] = val
			}
		}
		githubdata.UpdateWithCustomFields(e)
		stringed := fmt.Sprintf("%v", e["_id"])
		_, err = mc.Get(stringed)
		if err != nil {
			ch <- e
			itm := &memcache.Item{
				Key:        stringed,
				Value:      []byte{},
				Expiration: 300,
			}
			mc.Set(itm)
		} else {
			dups++
		}
	}
	return
}

func watchGithub(ch chan<- event) {
	for {
		dups := 0
		page := 0

		for dups == 0 && page < 5 {
			url := "https://api.github.com/events?per_page=100"
			if page > 0 {
				url = fmt.Sprintf("%v&page=%d", url, page)
			}
			log.Printf("Fetching %v", url)

			page++
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Error fetching from github: %v", err)
				break
			}
			defer resp.Body.Close()
			log.Printf("rate limit: %v/%v remaining",
				resp.Header.Get("X-RateLimit-Remaining"),
				resp.Header.Get("X-RateLimit-Limit"))

			dups = process(resp.Body, ch)
			if dups == 0 {
				log.Printf("No dups!  Need another page")
			}
			log.Printf("found %d dups", dups)
		}

		time.Sleep(5 * time.Second)
	}
}

func store(db couch.Database, doc event) {
	id, _, err := db.Insert(doc)
	if err == nil {
		log.Printf("Stored %v", id)
	} else if err.Error() == "409 Conflict" {
		log.Printf("Conflict on %v", doc["_id"])
	} else {
		log.Fatalf("Error storing %v: %v", doc, err)
	}
}

func logger(dburl string, ch <-chan event) {
	db, err := couch.Connect(dburl)
	if err != nil {
		log.Fatalf("Could not connect to DB")
	}

	for {
		doc := <-ch
		go store(db, doc)
	}
}

func main() {
	flag.Parse()

	ch := make(chan event, 1000)
	go logger(flag.Arg(0), ch)

	watchGithub(ch)
}
