package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

const beginning = 1331452800
const archiveu = "http://data.githubarchive.org/"

type storage interface {
	exists(string) bool
	store(string) (io.WriteCloser, error)
}

var st storage

var concurrency = flag.Int("c", 4, "Number of concurrent downlaods.")

var wg = sync.WaitGroup{}

func download(fn string) error {
	if st.exists(fn) {
		return nil
	}

	start := time.Now()
	u := archiveu + fn
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP Error: %v", res.Status)
	}

	out, err := st.store(fn)
	if err != nil {
		return err
	}

	defer out.Close()
	n, err := io.Copy(out, res.Body)
	if err == nil {
		log.Printf("Downloaded %v (%s) in %v", fn,
			humanize.Bytes(uint64(n)), time.Since(start))
	}
	return err
}

func downloader(ch chan time.Time) {
	defer wg.Done()
	for t := range ch {
		fn := formatDate(t) + ".json.gz"
		if err := download(fn); err != nil {
			log.Printf("Error on %v: %v", fn, err)
		}
	}
}

func main() {
	flag.Parse()

	st = fileStore{}

	start := time.Unix(beginning, 0)
	end := time.Now()

	ch := make(chan time.Time)
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go downloader(ch)
	}

	for _, d := range genDates(start, end, time.Hour) {
		ch <- d
	}
	close(ch)

	wg.Wait()
}
