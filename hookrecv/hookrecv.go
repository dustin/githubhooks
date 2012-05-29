package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form:  %v", err)
		}
		fmt.Fprintf(w, "hi\n")
		log.Printf("%q:\n", r.URL.Path)
		log.Printf("%v", r.Form)
	})
	log.Fatal(http.ListenAndServe(":6666", nil))
}
