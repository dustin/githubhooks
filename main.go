package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var configpath = flag.String("config", "config.json", "Path to config.")
var statePath = flag.String("statepath", "", "Path to store sequence state.")

type event struct {
	Seq int64
	Id  string
	Doc map[string]interface{}

	actor     string
	owner     string
	repo      string
	eventType string
}

func sendHook(urls []string, which string, ev event) {
	if len(urls) == 0 {
		return
	}
	bytes, err := json.Marshal(ev.Doc)
	if err != nil {
		log.Printf("Error encoding doc: %v", err)
		return
	}
	jsonstring := string(bytes)
	log.Printf("Sending %v hooks for %v %v on %v/%v with actor %v",
		which, ev.Id, ev.eventType, ev.owner, ev.repo, ev.actor)
	for _, u := range urls {
		resp, err := http.PostForm(u, url.Values{"payload": {jsonstring}})
		if err != nil {
			log.Printf("Error posting to %v: %v", u, err)
		} else {
			resp.Body.Close()
		}
	}

}

func byActor(s Subscriber, ch <-chan event) {
	for thing := range ch {
		sendHook(s.ByActor(thing), "byactor", thing)
	}
}

func byOwnerRepo(s Subscriber, ch <-chan event) {
	for thing := range ch {
		sendHook(s.ByOwnerRepo(thing), "byownerrepo", thing)
	}
}

func byOwner(s Subscriber, ch <-chan event) {
	for thing := range ch {
		sendHook(s.ByOwner(thing), "byowner", thing)
	}
}

func dispatcher(s Subscriber, ch <-chan event) {

	channels := map[string]chan event{}
	handlers := map[string]func(Subscriber, <-chan event){
		"byactor":     byActor,
		"byowner":     byOwner,
		"byownerrepo": byOwnerRepo,
	}

	for name, fun := range handlers {
		c := make(chan event, 1000)
		go fun(s, c)
		channels[name] = c
	}

	for thing := range ch {
		if s.Stale() {
			log.Printf("Reloading config.")
			s = loadInteresting(*configpath)
		}

		switch i := thing.Doc["repository"].(type) {
		case map[string]interface{}:
			switch o := i["owner"].(type) {
			case string:
				thing.owner = o
			case map[string]interface{}:
				thing.owner = fmt.Sprintf("%v", o["login"])
			}
			thing.repo = fmt.Sprintf("%v", i["name"])
		}
		switch i := thing.Doc["actor"].(type) {
		case string:
			thing.actor = i
		}
		et, ok := thing.Doc["type"].(string)
		if ok {
			thing.eventType = et
			for _, c := range channels {
				c <- thing
			}
		}
	}
}

func main() {
	flag.Parse()

	s := loadInteresting(*configpath)
	ch := make(chan event, 1000)
	go dispatcher(s, ch)

	dburl := flag.Arg(0)

	monitorDB(dburl, *statePath, ch)
}
