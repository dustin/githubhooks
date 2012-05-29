package main

import (
	"log"
)

func maybefatal(err error, msg string, args ...interface{}) {
	if err != nil {
		log.Fatalf(msg, args...)
	}
}
