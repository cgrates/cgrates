package main

import (
	"errors"
	"fmt"
	"github.com/rif/cgrates/timespans"
	"log"
	"net/http"
	"net/rpc"
	"runtime"
	"time"
	"flag"
)

var (
	raterAddress     = flag.String("rateraddr", ":2000", "Rater server address (localhost:2000)")
	jsonRpcAddress     = flag.String("jsonrpcaddr", ":2001", "Json RPC server address (localhost:2001)")
	htpApiAddress     = flag.String("httpapiaddr", ":2002", "Http API server address (localhost:2002)")
	raterList        *RaterList
)

/*
Handler for the statistics web client
*/
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><ol>")
	for _, addr := range raterList.clientAddresses {
		fmt.Fprint(w, fmt.Sprintf("<li>Client: %v</li>", addr))
	}
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



func main() {
	flag.Parse()
	raterList = NewRaterList()
	raterServer := new(RaterServer)
	rpc.Register(raterServer)
	rpc.HandleHTTP()

	go StopSingnalHandler()
	go listenToJsonRPCRequests()

	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	http.HandleFunc("/", handler)
	log.Print("The server is listening...")
	http.ListenAndServe(*raterAddress, nil)
}
