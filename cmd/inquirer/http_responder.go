package main

import(
	"fmt"
	"log"
	"net/http"
)

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

func listenToHttpRequests(){
	http.HandleFunc("/status", statusHandler)
	log.Print("The server is listening on ", *httpApiAddress)
	http.ListenAndServe(*httpApiAddress, nil)
}
