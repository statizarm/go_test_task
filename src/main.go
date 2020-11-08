package main

import (
	"fmt"
	"net/http"
	"log"
	"io"
	"golang.org/x/text/encoding/charmap"
)

func jsonMaker (runeCh <-chan rune, errorCount *uint32, doneCh chan<- bool) {
	<- runeCh
	*errorCount += 1
	doneCh <- true
}

func jsonHandler (w http.ResponseWriter, r *http.Request) {
	var runeCh = make (chan rune, 20)
	var doneCh = make (chan bool)

	var errorCount uint32 = 0

	go jsonMaker (runeCh, &errorCount, doneCh)

	for {
		var bytes = make ([]byte, 16)
		n, err := r.Body.Read (bytes);

		for i := 0; i < n; i += 1 {
			runeCh <- charmap.Windows1251.DecodeByte (bytes[i])
		}

		if err != nil {

			if err != io.EOF {
				fmt.Printf ("%s\n", err.Error ())
			}

			break
		}
	}

	<- doneCh

	fmt.Fprintf (w, "%d\n", errorCount)
}

func main () {
	fmt.Printf ("Starting http server")
	http.HandleFunc ("/", jsonHandler)
	log.Fatal (http.ListenAndServe (":80", nil))
}
