package main

import (
	"fmt"
	"net/http"
	"log"
	"io"
)

func jsonHandler (w http.ResponseWriter, r *http.Request) {
	if _, err := io.Copy (w, r.Body); err != nil {
		fmt.Printf ("%s\n", err);
	}
}

func main () {
	fmt.Printf ("Starting http server")
	http.HandleFunc ("/", jsonHandler)
	log.Fatal (http.ListenAndServe (":80", nil))
}
