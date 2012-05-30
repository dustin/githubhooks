package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"code.google.com/p/dsallings-couch-go"
	"github.com/dustin/githubhooks/datalib"
)

type event map[string]interface{}

func process(r io.Reader, inmap, outmap map[string]bool,
	ch chan<- event) (dups int) {

	stuff := []event{}
	err := json.NewDecoder(r).Decode(&stuff)
	if err != nil {
		log.Printf("Error decoding stuff: %v", err)
	}
	for _, e := range stuff {
		githubdata.UpdateWithCustomFields(e)
		switch i := e["actor"].(type) {
		case map[string]interface{}:
			e["actor_attributes"] = i
			e["actor"] = i["login"].(string)
		}
		stringed := fmt.Sprintf("%v", e["_id"])
		if _, ok := inmap[stringed]; !ok {
			ch <- e
		} else {
			dups++
		}
		outmap[stringed] = true
	}
	return
}

func watchGithub(ch chan<- event) {
	seen := map[string]bool{}
	for {
		dups := 0
		page := 0
		newmap := map[string]bool{}

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
			}
			defer resp.Body.Close()

			dups = process(resp.Body, seen, newmap, ch)
			if dups == 0 {
				log.Printf("No dups!  Need another page")
			}
			log.Printf("Now have %d dups tracked", len(newmap))
		}
		for k, v := range newmap {
			seen[k] = v
		}

		time.Sleep(2 * time.Second)
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
