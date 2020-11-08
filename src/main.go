package main

import (
	"fmt"
	"net/http"
	"log"
	"io"
	"io/ioutil"
	"golang.org/x/text/encoding/charmap"
)

func jsonMaker (r *io.PipeReader, errorCount *uint32, doneCh chan <- bool) {

	_,_ = ioutil.ReadAll (r)
	*errorCount += 1
	doneCh <- true
}

func jsonHandler (w http.ResponseWriter, r *http.Request) {
	var doneCh = make (chan bool)
	var errorCount uint32 = 0
	piper, pipew := io.Pipe ()

	go jsonMaker (piper, &errorCount, doneCh)

	decoder := charmap.Windows1251.NewDecoder ()
	reader := decoder.Reader (r.Body)

	var bytes = make ([]byte, 64)
	for {
		_, err := reader.Read (bytes)

		pipew.Write (bytes)

		if err != nil {
			if err != io.EOF {

				fmt.Printf ("%s\n", err.Error ())
			}

			pipew.Close ()
			break
		}
	}

	fmt.Println ("bytes perfoming excecuted")

	<- doneCh

	fmt.Fprintf (w, "%d\n", errorCount)
}

func main () {
	fmt.Printf ("Starting http server")
	http.HandleFunc ("/", jsonHandler)
	log.Fatal (http.ListenAndServe (":80", nil))
}
