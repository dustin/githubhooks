package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type interestingList map[string][]string

type allInteresting struct {
	Actor     interestingList `json:"ByActor"`
	Owner     interestingList `json:"ByOwner"`
	OwnerRepo interestingList `json:"ByOwnerRepo"`
}

type Subscriber interface {
	ByActor(ev event) []string
	ByOwner(ev event) []string
	ByOwnerRepo(ev event) []string
}

func fromIndex(i interestingList, key string, ev event) []string {
	if l, ok := i[key]; ok {
		return l
	}
	return []string{}
}

func (a *allInteresting) ByActor(ev event) []string {
	return fromIndex(a.Actor, ev.actor, ev)
}

func (a *allInteresting) ByOwner(ev event) []string {
	return fromIndex(a.Owner, ev.owner, ev)
}

func (a *allInteresting) ByOwnerRepo(ev event) []string {
	return fromIndex(a.OwnerRepo, fmt.Sprintf("%s/%s", ev.owner, ev.repo),
		ev)
}

func loadInteresting() Subscriber {
	f, err := os.Open("config.json")
	maybefatal(err, "Error opening config: %v", err)
	defer f.Close()

	rv := allInteresting{}

	d := json.NewDecoder(f)
	err = d.Decode(&rv)
	maybefatal(err, "Error reading config: %v", err)
	return &rv
}
