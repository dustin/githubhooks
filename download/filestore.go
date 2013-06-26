package main

import (
	"io"
	"os"
)

type fileStore struct{}

func (f fileStore) exists(fn string) bool {
	_, err := os.Stat(fn)
	return err == nil
}

func (f fileStore) store(fn string) (io.WriteCloser, error) {
	return os.Create(fn)
}
