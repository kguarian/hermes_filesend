package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const pattern string = ":3128/"
const errorioutilread string = "ioutil http.Request body read error."

func handler(w http.ResponseWriter, r *http.Request) {
	var variable *string
	buf, err := ioutil.ReadAll(r.Body)
	Errhandle_Exit(err, errorioutilread)
	fmt.Printf("%s\nEOF\n", buf)
	fmt.Sscanf(string(buf), "%s\n", variable)
	fmt.Printf("regexstring: %s\n", *variable)
}

func ReceiverRequest() {
	http.HandleFunc(pattern, handler)
}
