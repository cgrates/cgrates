package history

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

var (
	dataFile   = flag.String("file", "store.json", "data store file name")
	hostname   = flag.String("host", "localhost:8080", "http host name")
	masterAddr = flag.String("master", "", "RPC master address")
)

var store Store

func start() {

	flag.Parse()
	if *masterAddr != "" {
		store, err := NewProxyStore(*masterAddr)
	} else {
		store, err := NewFileStore(*dataFile)
	}
	rpc.RegisterName("Store", store)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}
