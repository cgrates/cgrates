package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"time"
)

type RaterList struct {
	clients map[string]*rpc.Client
}

var raterList *RaterList

func (rl *RaterList) RegisterRater(clientAddress string, replay *byte) error {
	time.Sleep(1 * time.Second) // wait a second for Rater to start serving
	client, err := rpc.Dial("tcp", clientAddress)
	if err != nil {
		log.Panic("Could not connect to client!")
	}
	rl.clients[clientAddress] = client
	log.Print(fmt.Sprintf("Server %v registered succesfully", clientAddress))
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><ol>")
	for addr, _ := range raterList.clients {
		fmt.Fprint(w, fmt.Sprintf("<li>%s</li>", addr))
	}
	fmt.Fprint(w, "</ol></body></html>")
}

func callRater(rl *RaterList) {
	var reply float64
	arg := 9.0

	log.Print("Starting client polling.")
	for {
		for addr, client := range rl.clients {
			err := client.Call("Sumer.Square", arg, &reply)
			if err != nil {
				log.Print("Closing client!")
				delete(rl.clients, addr)
			}
			fmt.Println(fmt.Sprintf("Result from rater(%v): %v", addr, reply))
		}
		time.Sleep(5 * time.Second)
	}
}

func main() {
	raterList = &RaterList{clients: make(map[string]*rpc.Client)}
	go callRater(raterList)
	rpc.Register(raterList)
	rpc.HandleHTTP()
	http.HandleFunc("/", handler)
	log.Print("The server is listening...")
	http.ListenAndServe(":2000", nil)
}
