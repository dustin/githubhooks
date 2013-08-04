package main

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/couchbaselabs/cbfs/client"
)

type cbfsStore struct {
	client *cbfsclient.Client
	path   string
	known  map[string]bool
}

func newCBFSStore(ustr string) (cbfsStore, error) {
	rv := cbfsStore{known: map[string]bool{}}

	u, err := url.Parse(ustr)
	if err != nil {
		return rv, err
	}
	rv.path = u.Path
	u.Path = "/"

	client, err := cbfsclient.New(u.String())
	if err != nil {
		return rv, err
	}

	l, err := client.List(rv.path)
	if err != nil {
		return rv, err
	}

	for fn := range l.Files {
		rv.known[fn] = true
	}

	rv.client = client

	return rv, nil
}

func (c cbfsStore) exists(fn string) bool {
	return c.known[fn]
}

func (c cbfsStore) store(fn string) (io.WriteCloser, error) {
	pr, pw := io.Pipe()

	dest := c.client.URLFor(c.path + fn)
	req, err := http.NewRequest("PUT", dest, pr)
	if err != nil {
		return nil, err
	}

	log.Printf("Storing in %v", dest)

	go func() {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			pr.CloseWithError(err)
		}
		if res.StatusCode != 204 {
			pr.CloseWithError(err)
		}
	}()

	return pw, nil
}
