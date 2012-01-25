package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"time"
	"errors"
)

var raterList *RaterList

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><ol>")
	for addr, _ := range raterList.Clients {
		fmt.Fprint(w, fmt.Sprintf("<li>%s</li>", addr))
	}
	fmt.Fprint(w, "</ol></body></html>")
}

func callRater() {
	var reply string
	err := errors.New("") //not nil value
	for err != nil {
		client:= <-raterList.Balancer
		err = client.Call("Storage.Get", "test", &reply)
		if err != nil {
			log.Printf("Got en error from rater: %v", err)
		}			
	}
	fmt.Println(fmt.Sprintf("Result: %v", reply))
}

func testCallRater(){
	for {
		callRater()
		time.Sleep(1 * time.Second)
	}
}

func main() {
	raterList = NewRaterList()	
	go testCallRater()
	rpc.Register(raterList)
	rpc.HandleHTTP()
	http.HandleFunc("/", handler)
	log.Print("The server is listening...")
	http.ListenAndServe(":2000", nil)
}
