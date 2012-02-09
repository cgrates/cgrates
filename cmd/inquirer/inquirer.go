package main

import (
	"errors"
	"fmt"
	"github.com/rif/cgrates/timespans"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"runtime"
	"time"
)

var raterList        *RaterList

/*
Handler for the statistics web client
*/
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><ol>")
	for _, addr := range raterList.clientAddresses {
		fmt.Fprint(w, fmt.Sprintf("<li>Client: %v</li>", addr))
	}
	fmt.Fprint(w, fmt.Sprintf("<li>Gorutines: %v</li>", runtime.Goroutines()))
	fmt.Fprint(w, "</ol></body></html>")
}

/*
The function that gets the information from the raters using balancer.
*/
func CallRater(key *timespans.CallDescriptor) (reply *timespans.CallCost) {
	err := errors.New("") //not nil value
	for err != nil {
		client := raterList.Balance()
		if client == nil {
			log.Print("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			reply = &timespans.CallCost{}
			err = client.Call("Storage.GetCost", *key, reply)
			if err != nil {
				log.Printf("Got en error from rater: %v", err)
			}
		}
	}
	return
}

func listenToTheWorld() {
	l, err := net.Listen("tcp", ":5090")
	defer l.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Print("listening:", l.Addr())

	responder := new(Responder)
	rpc.Register(responder)

	for {
		c, err := l.Accept()
		if err != nil {
			log.Printf("accept error: %s", c)
			continue
		}

		log.Printf("connection started: %v", c.RemoteAddr())
		go jsonrpc.ServeConn(c)
	}
}

func main() {
	raterList = NewRaterList()
	raterServer := new(RaterServer)
	rpc.Register(raterServer)
	rpc.HandleHTTP()

	go StopSingnalHandler()
	go listenToTheWorld()
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	http.HandleFunc("/", handler)
	log.Print("The server is listening...")
	http.ListenAndServe(":2000", nil)
}
