package main

import (
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/dustin/go-couch"
)

func feedBody(r io.Reader, sp *SequencePersister, results chan<- event) int64 {

	largest := int64(0)

	d := json.NewDecoder(r)

	for {
		thing := event{}
		err := d.Decode(&thing)
		if err != nil {
			log.Printf("Error decoding:  %v", err)
			time.Sleep(time.Second)
			return largest
		}
		results <- thing
		largest = thing.Seq
		sp.NewVal(largest)
	}
}

func monitorDB(dburl, statepath string, ch chan<- event) {
	db, err := couch.Connect(dburl)
	maybefatal(err, "Error connecting: %v", err)

	seqper := NewSequencePersister(statepath, time.Minute*5)

	startSeq := seqper.Current()
	go seqper.Run()

	if startSeq == 0 {
		info, err := db.GetInfo()
		maybefatal(err, "Error getting info: %v", err)
		log.Printf("Info %#v", info)
		startSeq = info.UpdateSeq
		seqper.NewVal(startSeq)
	}

	if startSeq != seqper.Current() {
		seqper.WriteValue(startSeq)
	}

	log.Printf("Starting changes from %v", startSeq)

	err = db.Changes(func(r io.Reader) int64 {
		return feedBody(r, seqper, ch)
	},
		map[string]interface{}{
			"since":        startSeq,
			"feed":         "continuous",
			"include_docs": true,
		})
	log.Fatalf("DB monitor exited with %v", err)
}
