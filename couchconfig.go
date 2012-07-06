package main

import (
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"code.google.com/p/dsallings-couch-go"
)

var confMonitor sync.Once
var couchConfStaleness *bool

type Doc struct {
	Trigger string
	Target  string
	Url     string
	Events  *[]string
}

func updateDBConf(il interestingList, doc Doc) {
	l, ok := il[doc.Target]
	if !ok {
		l = make([]string, 0, 1)
	}
	if doc.Events != nil && len(*doc.Events) > 1 {
		for _, e := range *doc.Events {
			il[doc.Target+":"+e] = append(l, doc.Url)
		}
	} else {
		il[doc.Target] = append(l, doc.Url)
	}
}

func watchConfigChanges(dburl string) {
	log.Printf("Watching for config changes in couchdb.")
	first := true
	largest := int64(0)
	for {
		if !first {
			time.Sleep(time.Second)
		}
		first = false

		db, err := couch.Connect(dburl)
		if err != nil {
			log.Printf("Config watcher db connect error: %v", err)
			continue
		}

		if largest == 0 {
			info, err := db.GetInfo()
			if err != nil {
				log.Printf("Config watcher info error: %v", err)
				continue
			}
			largest = info.UpdateSeq
		}

		err = db.Changes(func(r io.Reader) int64 {
			d := json.NewDecoder(r)

			for {
				out := struct{ Seq int64 }{}
				err := d.Decode(&out)
				if err != nil {
					log.Printf("Stream error from config: %v", err)
					return largest
				}
				log.Printf("Signaling a config change")
				*couchConfStaleness = true
			}
			panic("This can't happen")
		},
			map[string]interface{}{
				"since": largest,
				"feed":  "continuous",
			})
	}
}

func loadInterestingCouch(dburl string) Subscriber {
	db, err := couch.Connect(dburl)
	maybefatal(err, "Error connecting: %v", err)

	type hookResults struct {
		Rows []struct {
			Doc Doc `json:"doc"`
		}
	}

	results := hookResults{}
	err = db.Query("_design/app/_view/hooks", map[string]interface{}{
		"include_docs": true,
	}, &results)
	maybefatal(err, "Problem loading hooks from DB: %v", err)

	conf := jsonConfig{}
	conf.Actor = make(interestingList)
	conf.Owner = make(interestingList)
	conf.OwnerRepo = make(interestingList)

	for _, doc := range results.Rows {
		switch doc.Doc.Trigger {
		case "ByActor":
			updateDBConf(conf.Actor, doc.Doc)
		case "ByOwner":
			updateDBConf(conf.Owner, doc.Doc)
		case "ByOwnerRepo":
			updateDBConf(conf.OwnerRepo, doc.Doc)
		default:
			log.Fatalf("Unknown trigger: %v", doc.Doc.Trigger)
		}
	}

	couchConfStaleness = &conf.stale

	confMonitor.Do(func() {
		go watchConfigChanges(dburl)
	})

	return &conf
}
