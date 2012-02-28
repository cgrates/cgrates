/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package main

import (
	"encoding/json"
	"fmt"
	"github.com/rif/cgrates/timespans"
	"log"
	"net/http"
	"strconv"
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
	subj, ok2 := r.Form["subj"]
	dest, ok3 := r.Form["dest"]
	if !ok1 || !ok2 || !ok3 {
		enc.Encode(IncorrectParameters{"Incorrect parameters"})
		return
	}
	arg := &timespans.CallDescriptor{CstmId: cstmid[0], Subject: subj[0], DestinationPrefix: dest[0]}
	callCost := GetCost(arg)
	enc.Encode(callCost)
}

/*
curl "http://127.0.0.1:8000/debitbalance?cstmid=vdf&subj=rif&dest=0257@amount=100"
*/
func debitBalanceHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	r.ParseForm()
	cstmid, ok1 := r.Form["cstmid"]
	subj, ok2 := r.Form["subj"]
	dest, ok3 := r.Form["dest"]
	amount_s, ok4 := r.Form["amount"]
	amount, err := strconv.ParseFloat(amount_s[0], 64)
	if !ok1 || !ok2 || !ok3 || ok4 || err != nil {
		enc.Encode(IncorrectParameters{"Incorrect parameters"})
		return
	}
	arg := &timespans.CallDescriptor{CstmId: cstmid[0], Subject: subj[0], DestinationPrefix: dest[0], Amount: amount}
	result := CallMethod(arg, "Storage.DebitCents")
	enc.Encode(result)
}

/*
curl "http://127.0.0.1:8000/debitsms?cstmid=vdf&subj=rif&dest=0257@amount=100"
*/
func debitSMSHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	r.ParseForm()
	cstmid, ok1 := r.Form["cstmid"]
	subj, ok2 := r.Form["subj"]
	dest, ok3 := r.Form["dest"]
	amount_s, ok4 := r.Form["amount"]
	amount, err := strconv.ParseFloat(amount_s[0], 64)
	if !ok1 || !ok2 || !ok3 || !ok4 || err != nil {
		enc.Encode(IncorrectParameters{"Incorrect parameters"})
		return
	}
	arg := &timespans.CallDescriptor{CstmId: cstmid[0], Subject: subj[0], DestinationPrefix: dest[0], Amount: amount}
	result := CallMethod(arg, "Storage.DebitSMS")
	enc.Encode(result)
}

/*
curl "http://127.0.0.1:8000/debitseconds?cstmid=vdf&subj=rif&dest=0257@amount=100"
*/
func debitSecondsHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	r.ParseForm()
	cstmid, ok1 := r.Form["cstmid"]
	subj, ok2 := r.Form["subj"]
	dest, ok3 := r.Form["dest"]
	amount_s, ok4 := r.Form["amount"]
	amount, err := strconv.ParseFloat(amount_s[0], 64)
	if !ok1 || !ok2 || !ok3 || !ok4 || err != nil {
		enc.Encode(IncorrectParameters{"Incorrect parameters"})
		return
	}
	arg := &timespans.CallDescriptor{CstmId: cstmid[0], Subject: subj[0], DestinationPrefix: dest[0], Amount: amount}
	result := CallMethod(arg, "Storage.DebitSeconds")
	enc.Encode(result)
}

/*
curl "http://127.0.0.1:8000/getmaxsessiontime?cstmid=vdf&subj=rif&dest=0257@amount=100"
*/
func getMaxSessionTimeHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	r.ParseForm()
	cstmid, ok1 := r.Form["cstmid"]
	subj, ok2 := r.Form["subj"]
	dest, ok3 := r.Form["dest"]
	amount_s, ok4 := r.Form["amount"]
	amount, err := strconv.ParseFloat(amount_s[0], 64)
	if !ok1 || !ok2 || !ok3 || !ok4 || err != nil {
		enc.Encode(IncorrectParameters{"Incorrect parameters"})
		return
	}
	arg := &timespans.CallDescriptor{CstmId: cstmid[0], Subject: subj[0], DestinationPrefix: dest[0], Amount: amount}
	result := CallMethod(arg, "Storage.GetMaxSessionTime")
	enc.Encode(result)
}

func listenToHttpRequests() {
	http.HandleFunc("/", statusHandler)
	http.HandleFunc("/getcost", getCostHandler)
	http.HandleFunc("/debitbalance", debitBalanceHandler)
	http.HandleFunc("/debitsms", debitSMSHandler)
	http.HandleFunc("/debitseconds", debitSecondsHandler)
	http.HandleFunc("/getmaxsessiontime", debitSecondsHandler)
	log.Print("The server is listening on ", *httpApiAddress)
	http.ListenAndServe(*httpApiAddress, nil)
}
