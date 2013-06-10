package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type SequencePersister struct {
	Path      string
	Frequency time.Duration

	latest  int64
	written int64
	ch      chan int64
}

func NewSequencePersister(path string, freq time.Duration) *SequencePersister {
	sp := SequencePersister{
		Path:      path,
		Frequency: freq,
		ch:        make(chan int64, 1000),
	}
	if path != "" {
		b, err := ioutil.ReadFile(path)
		if err == nil {
			_, err = fmt.Sscanf(string(b), "%v", &sp.latest)
			if err != nil {
				sp.latest = 0
			} else {
				sp.written = sp.latest
			}
		}
	}
	return &sp
}

func (s *SequencePersister) NewVal(i int64) {
	s.ch <- i
}

func (s *SequencePersister) Run() {
	t := time.Tick(s.Frequency)
	for {
		select {
		case s.latest = <-s.ch:
		case <-t:
			if s.latest != s.written {
				s.WriteValue(s.latest)
			}
		}
	}
}

func (s *SequencePersister) WriteValue(v int64) {
	if s.Path == "" {
		s.written = s.latest
		return
	}
	log.Printf("Remembering %v in %v", v, s.Path)
	err := ioutil.WriteFile(s.Path+".tmp", []byte(fmt.Sprintf("%v\n", v)), 0666)
	if err == nil {
		err = os.Rename(s.Path+".tmp", s.Path)
		if err == nil {
			s.written = s.latest
		} else {
			log.Printf("Error renaming persistence sequence file: %v",
				err)
		}
	} else {
		log.Printf("Error persisting sequence: %v", err)
	}
}

func (s *SequencePersister) Current() int64 {
	return s.latest
}
