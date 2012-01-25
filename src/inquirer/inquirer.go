package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"registration"
	//"time"
	"errors"
)

var raterList *registration.RaterList

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><ol>")
	for addr, _ := range raterList.Clients {
		fmt.Fprint(w, fmt.Sprintf("<li>%s</li>", addr))
	}
	fmt.Fprint(w, "</ol></body></html>")
}

func callRater() {
	var reply float64
	arg := 9.0		
	err := errors.New("") //not nil value
	for err != nil {
		client:= <-raterList.Balancer
		err = client.Call("Sumer.Square", arg, &reply)
		if err != nil {
			log.Print("Got en error from rater... recovering")
		}			
	}
	fmt.Println(fmt.Sprintf("Result: %v", reply))
}

func testCallRater(){
	for i:= 0; i<10; i++ {
		go callRater()
		//time.Sleep(1 * time.Second)
	}
}

func main() {
	raterList = registration.NewRaterList()	
	go testCallRater()
	rpc.Register(raterList)
	rpc.HandleHTTP()
	http.HandleFunc("/", handler)
	log.Print("The server is listening...")
	http.ListenAndServe(":2000", nil)
}
