package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"time"
	"registration"
)

var raterList *registration.RaterList

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><ol>")
	for addr, _ := range raterList.Clients {
		fmt.Fprint(w, fmt.Sprintf("<li>%s</li>", addr))
	}
	fmt.Fprint(w, "</ol></body></html>")
}

func callRater(rl *registration.RaterList) {
	var reply float64
	arg := 9.0

	log.Print("Starting client polling.")
	for {
		for addr, client := range rl.Clients {
			err := client.Call("Sumer.Square", arg, &reply)
			if err != nil {
				log.Print("Closing client!")
				delete(rl.Clients, addr)
			}
			fmt.Println(fmt.Sprintf("Result from rater(%v): %v", addr, reply))
		}
		time.Sleep(5 * time.Second)
	}
}

func main() {
	raterList = &registration.RaterList{Clients: make(map[string]*rpc.Client)}
	go callRater(raterList)
	rpc.Register(raterList)
	rpc.HandleHTTP()
	http.HandleFunc("/", handler)
	log.Print("The server is listening...")
	http.ListenAndServe(":2000", nil)
}
