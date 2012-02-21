package main

import(
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"github.com/rif/cgrates/timespans"
)

type IncorrectParameters struct {
	Error string
}

/*
Handler for the statistics web client
*/
func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><ol>")
	for _, addr := range raterList.clientAddresses {
		fmt.Fprint(w, fmt.Sprintf("<li>Client: %v</li>", addr))
	}
	fmt.Fprint(w, "</ol></body></html>")
}

/*
curl "http://127.0.0.1:8000/getcost?cstmid=vdf&subj=rif&dest=0257"
*/
func getCostHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	r.ParseForm()
	cstmid, ok1 := r.Form["cstmid"]
	subj, ok2 :=r.Form["subj"]
	dest, ok3 := r.Form["dest"]
	if !ok1 || !ok2 || !ok3 {
		enc.Encode(IncorrectParameters{"Incorrect parameters"})
		return
	}
	arg := &timespans.CallDescriptor{CstmId: cstmid[0], Subject: subj[0], DestinationPrefix: dest[0]}
	callCost := CallRater(arg)
	enc.Encode(callCost)
}

func listenToHttpRequests(){
	http.HandleFunc("/", statusHandler)
	http.HandleFunc("/getcost", getCostHandler)
	log.Print("The server is listening on ", *httpApiAddress)
	http.ListenAndServe(*httpApiAddress, nil)
}
