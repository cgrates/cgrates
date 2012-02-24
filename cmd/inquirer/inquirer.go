package main

import (
	"errors"
	"flag"
	"github.com/rif/cgrates/timespans"
	"log"
	"runtime"
	"time"
)

var (
	raterAddress   = flag.String("rateraddr", "127.0.0.1:2000", "Rater server address (localhost:2000)")
	jsonRpcAddress = flag.String("jsonrpcaddr", "127.0.0.1:2001", "Json RPC server address (localhost:2001)")
	httpApiAddress = flag.String("httpapiaddr", "127.0.0.1:8000", "Http API server address (localhost:2002)")
	raterList      *RaterList
)

/*
The function that gets the information from the raters using balancer.
*/
func GetCost(key *timespans.CallDescriptor) (reply *timespans.CallCost) {
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

/*
The function that gets the information from the raters using balancer.
*/
func CallMethod(key *timespans.CallDescriptor, method string) (reply float64) {
	err := errors.New("") //not nil value
	for err != nil {
		client := raterList.Balance()
		if client == nil {
			log.Print("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			err = client.Call(method, *key, &reply)
			if err != nil {
				log.Printf("Got en error from rater: %v", err)
			}
		}
	}
	return
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	raterList = NewRaterList()

	go StopSingnalHandler()
	go listenToRPCRaterRequests()
	go listenToJsonRPCRequests()

	listenToHttpRequests()
}
