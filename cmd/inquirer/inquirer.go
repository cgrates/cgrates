package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"errors"
)

var raterList *RaterList

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><ol>")
	for addr, _ := range raterList.clientAddresses {
		fmt.Fprint(w, fmt.Sprintf("<li>%s</li>", addr))
	}
	fmt.Fprint(w, "</ol></body></html>")
}

func callRater(key string) (reply string) {
	err := errors.New("") //not nil value
	for err != nil {
		client:= raterList.Balance()
		err = client.Call("Storage.Get", key, &reply)
		if err != nil {
			log.Printf("Got en error from rater: %v", err)
		}			
	}
	return 
}

func main() {
	raterList = NewRaterList()
	rpc.Register(raterList)
	rpc.HandleHTTP()
	
	responder := new(Responder)
	srvr := rpc.NewServer()
	srvr.Register(responder)
	f1 := func(w http.ResponseWriter, req *http.Request) {
  		srvr.ServeHTTP(w, req)
	}
	http.HandleFunc("/rpc", f1)
	
	http.HandleFunc("/", handler)	
	log.Print("The server is listening...")
	http.ListenAndServe(":2000", nil)
}
