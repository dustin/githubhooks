package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

type interestingList map[string][]string

type allInteresting struct {
	ByActor     interestingList
	ByOwner     interestingList
	ByOwnerRepo interestingList
}

var interestingConfig allInteresting

type event struct {
	Seq int64
	Id  string
	Doc map[string]interface{}

	actor     string
	owner     string
	repo      string
	eventType string
}

func logEvent(ch <-chan event) {
	for _ = range ch {
		/*
			log.Printf("%v - %v/%v (%v)", thing.eventType, thing.owner,
				thing.repo, thing.actor)
		*/
	}
}

func sendHook(urls []string, ev event) {
	bytes, err := json.Marshal(ev.Doc)
	if err != nil {
		log.Printf("Error encoding doc: %v", err)
		return
	}
	jsonstring := string(bytes)
	for _, u := range urls {
		resp, err := http.PostForm(u, url.Values{"payload": {jsonstring}})
		if err != nil {
			log.Printf("Error posting to %v: %v", u, err)
		} else {
			resp.Body.Close()
		}
	}

}

func byActor(ch <-chan event) {
	interesting := interestingConfig.ByActor
	for thing := range ch {
		if list, ok := interesting[thing.actor]; ok {
			log.Printf("Handling byActor %v event for actor %v in %v/%v to %v",
				thing.eventType, thing.actor, thing.owner, thing.repo,
				list)
			sendHook(list, thing)
		}
	}
}

func byOwnerRepo(ch <-chan event) {
	interesting := interestingConfig.ByOwnerRepo
	for thing := range ch {
		k := fmt.Sprintf("%v/%v", thing.owner, thing.repo)
		if list, ok := interesting[k]; ok {
			log.Printf("Handling byOwnerRepo %v event in repo %v/%v by %v to %v",
				thing.eventType, thing.owner, thing.repo, thing.actor, list)
			sendHook(list, thing)
		}
	}
}

func byOwner(ch <-chan event) {
	interesting := interestingConfig.ByOwner
	for thing := range ch {
		if list, ok := interesting[thing.owner]; ok {
			log.Printf("Handling byOwner %v event in repo %v/%v by %v to %v",
				thing.eventType, thing.owner, thing.repo, thing.actor, list)
			sendHook(list, thing)
		}
	}
}

func dispatcher(ch <-chan event) {

	channels := map[string]chan event{}
	handlers := map[string]func(<-chan event){
		"log":         logEvent,
		"byactor":     byActor,
		"byowner":     byOwner,
		"byownerrepo": byOwnerRepo,
	}

	for name, fun := range handlers {
		c := make(chan event, 1000)
		go fun(c)
		channels[name] = c
	}

	for thing := range ch {
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

func loadInteresting() {
	f, err := os.Open("config.json")
	maybefatal(err, "Error opening config: %v", err)
	defer f.Close()

	d := json.NewDecoder(f)
	err = d.Decode(&interestingConfig)
	maybefatal(err, "Error reading config: %v", err)
}

func main() {
	loadInteresting()
	ch := make(chan event, 1000)
	go dispatcher(ch)

	flag.Parse()
	dburl := flag.Arg(0)

	monitorDB(dburl, ch)
}
