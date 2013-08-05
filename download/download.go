package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/syslog"
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
var cbfspath = flag.String("cbfs", "", "Path to store in cbfs")
var useSyslog = flag.Bool("syslog", false, "log to syslog")

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

func initLogger(slog bool) {
	if slog {
		lw, err := syslog.New(syslog.LOG_INFO, "github-dl")
		if err != nil {
			log.Fatalf("Can't initialize syslog: %v", err)
		}
		log.SetOutput(lw)
		log.SetFlags(0)
	}
}

func main() {
	flag.Parse()
	initLogger()

	st = fileStore{}
	if *cbfspath != "" {
		t, err := newCBFSStore(*cbfspath)
		if err != nil {
			log.Fatalf("Couldn't initialize cbfs store: %v", err)
		}
		st = t
	}

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
