// Process github event data exports.
//
// Go here for more info:  http://www.githubarchive.org/
package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/githubhooks/datalib"
	"github.com/dustin/go-couch"
	"github.com/dustin/go-humanize"
)

var totalRead = int64(0)
var numerrors = int64(0)

type fileData struct {
	Filename  string    `json:"_id"`
	Timestamp time.Time `json:"ts"`
	Type      string    `json:"type"`
}

func maybeFatal(err error, msg ...string) {
	if err != nil {
		if len(msg) > 0 {
			log.Fatalf("Error: %v (%v)", err, msg)
		} else {
			log.Fatalf("Error:  %v", err)
		}
	}
}

func confirmResults(rs []couch.Response) (rv int64) {
	for _, r := range rs {
		if !r.Ok {
			rv += 1
			log.Printf("Error:  %#v", r)
		}
	}
	return
}

func dofile(fn string, db couch.Database) {
	docs := make([]interface{}, 0, 10000)

	fdata := fileData{Filename: filepath.Base(fn),
		Timestamp: time.Now().UTC(),
		Type:      "importghfile",
	}
	docs = append(docs, fdata)

	f, err := os.Open(fn)
	maybeFatal(err)
	defer f.Close()

	gz, err := gzip.NewReader(f)
	maybeFatal(err)
	defer gz.Close()

	d := json.NewDecoder(gz)
	eventcount := 0

	for {
		thing := map[string]interface{}{}
		if err := d.Decode(&thing); err != nil {
			if err != io.EOF {
				log.Printf("Error decoding %v: %v", fn, err)
			}
			log.Printf("Processed %s events from %v",
				humanize.Comma(int64(eventcount)), filepath.Base(fn))
			atomic.AddInt64(&totalRead, int64(eventcount))

			results, err := db.Bulk(docs)
			maybeFatal(err, "Error bulk updating")
			atomic.AddInt64(&numerrors, confirmResults(results))

			return
		}
		// This is where you do something exciting with the data.
		eventcount += 1
		githubdata.UpdateWithCustomFields(thing)
		docs = append(docs, thing)
	}
}

func processor(ch <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	db, err := couch.Connect("http://localhost:5984/github")
	maybeFatal(err)

	for fn := range ch {
		bn := filepath.Base(fn)
		var fdat fileData
		if db.Retrieve(bn, &fdat) != nil {
			dofile(fn, db)
		} else {
			log.Printf("Already saw %v", fn)
		}
	}
}

func main() {
	wg := sync.WaitGroup{}
	ch := make(chan string)

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go processor(ch, &wg)
	}
	for _, fn := range os.Args[1:] {
		ch <- fn
	}
	close(ch)
	wg.Wait()
	log.Printf("Processed %v events, %v duplicates",
		humanize.Comma(totalRead),
		humanize.Comma(numerrors))
}
