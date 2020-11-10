package main

import (
	"fmt"
	"net/http"
	"io"
	"encoding/json"
	"unicode"
	"unicode/utf8"
	"golang.org/x/text/encoding/charmap"
)

type worker struct {
	doneCh chan bool
}

type jsonChecker struct {
	worker
	invalidValues uint32
}

type jsonMaker worker

type jsonDocType = map[string]string

// TODO: move symbol validations to another package, a. g. valid.go
func isWrongSym (r rune) bool {
	return unicode.In (r, unicode.Z, unicode.P, unicode.C, unicode.M)
}

func containsWrongSym (str string) bool {
	for _, r := range str {
		if isWrongSym (r) {
			return true
		}
	}
	return false
}
func isValid (doc *jsonDocType) bool {
	// TODO: add more symbol calsses
	for k,v := range *doc {
		if containsWrongSym (k) || containsWrongSym (v) {
			return false
		}
	}

	return true
}

func (jc *jsonChecker) checkJson (docCh <- chan jsonDocType) {
	for {
		if doc, more := <- docCh; more {
			if isValid (&doc) == false {
				jc.invalidValues += 1

				// If we want to fix invalid json we must place here fixJson function
			}
		} else {
			break;
		}
	}
	jc.doneCh <- true
}

func (jm *jsonMaker) jsonMaker (r *io.PipeReader, errorCount *uint32) {
	var jc = new (jsonChecker)
	jc.doneCh = make (chan bool)
	jc.invalidValues = 0

	var docCh = make (chan jsonDocType)

	go jc.checkJson(docCh)

	dec := json.NewDecoder (r)
	if _, err := dec.Token (); err != nil {
		fmt.Println ("no open bracket");
	}
	for dec.More () {
		var doc jsonDocType

		if err := dec.Decode (&doc); err != nil {
			fmt.Println ("syntax error")
			// TODO: add syntax error handler for counting invalid json array values
			break
		}
		docCh <- doc

		fmt.Println (doc)
	}

	if _, err := dec.Token (); err != nil {
		fmt.Println ("no close bracket")
	}

	close (docCh)
	r.Close ()

	<- jc.doneCh
	*errorCount += jc.invalidValues
	jm.doneCh <- true
}

func jsonHandler (w http.ResponseWriter, r *http.Request) {
	var jm = new (jsonMaker)
	jm.doneCh = make (chan bool)

	var errorCount uint32 = 0
	piper, pipew := io.Pipe ()

	go jm.jsonMaker (piper, &errorCount)

	decoder := charmap.Windows1251.NewDecoder ()
	reader := decoder.Reader (r.Body)

	var bytes = make ([]byte, 64)
	for {
		n, err := reader.Read (bytes)

		pipew.Write (bytes[:n])
		fmt.Println (utf8.Valid(bytes[:n]))

		if err != nil {
			if err != io.EOF {

				fmt.Printf ("%s\n", err.Error ())
			}

			pipew.Close ()
			break
		}
	}

	<- jm.doneCh

	fmt.Fprintf (w, "%d\n", errorCount)
}

func main () {
	fmt.Printf ("Starting http server")
	http.HandleFunc ("/", jsonHandler)
	http.ListenAndServe (":80", nil)
}
