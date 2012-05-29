package main

import (
	"code.google.com/p/dsallings-couch-go"
	"encoding/json"
	"io"
	"log"
)

func feedBody(r io.Reader, results chan<- event) int64 {

	largest := int64(0)

	d := json.NewDecoder(r)

	for {
		thing := event{}
		err := d.Decode(&thing)
		if err != nil {
			if err.Error() == "unexpected EOF" {
				return largest
			} else {
				log.Fatalf("Error decoding stuff: %#v", err)
			}
		}
		results <- thing
		largest = thing.Seq
	}

	return largest
}

func monitorDB(dburl string, ch chan<- event) {
	db, err := couch.Connect(dburl)
	maybefatal(err, "Error connecting: %v", err)

	info, err := db.GetInfo()
	maybefatal(err, "Error getting info: %v", err)
	log.Printf("Info %#v", info)

	err = db.Changes(func(r io.Reader) int64 {
		return feedBody(r, ch)
	},
		map[string]interface{}{
			"since":        info.UpdateSeq,
			"feed":         "continuous",
			"include_docs": true,
			"heartbeat":    5000,
		})
	maybefatal(err, "Error changesing: %v", err)
}
