package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type interestingList map[string][]string

type jsonConfig struct {
	Actor     interestingList `json:"ByActor"`
	Owner     interestingList `json:"ByOwner"`
	OwnerRepo interestingList `json:"ByOwnerRepo"`

	stale bool
}

type Subscriber interface {
	ByActor(ev event) []string
	ByOwner(ev event) []string
	ByOwnerRepo(ev event) []string

	Stale() bool
}

func fromIndex(i interestingList, key string, ev event) []string {
	l, ok := i[key]
	if !ok {
		l = []string{}
	}

	l2, ok := i[key+":"+ev.eventType]
	if !ok {
		l2 = []string{}
	}
	return append(l, l2...)
}

func (a *jsonConfig) ByActor(ev event) []string {
	return fromIndex(a.Actor, ev.actor, ev)
}

func (a *jsonConfig) ByOwner(ev event) []string {
	return fromIndex(a.Owner, ev.owner, ev)
}

func (a *jsonConfig) ByOwnerRepo(ev event) []string {
	return fromIndex(a.OwnerRepo, fmt.Sprintf("%s/%s", ev.owner, ev.repo),
		ev)
}

func (a jsonConfig) Stale() bool {
	return a.stale
}

func loadInterestingFile(path string) Subscriber {
	f, err := os.Open(path)
	maybefatal(err, "Error opening config: %v", err)
	defer f.Close()

	rv := jsonConfig{}

	d := json.NewDecoder(f)
	err = d.Decode(&rv)
	maybefatal(err, "Error reading config: %v", err)
	return &rv
}

func loadInteresting(path string) Subscriber {
	if strings.HasPrefix(path, "http:") {
		return loadInterestingCouch(path)
	}
	return loadInterestingFile(path)
}
