package main

import (
	"github.com/rif/cgrates/timespans"
	"log"
	"net/rpc/jsonrpc"
	"time"
	"flag"
)

var (
	runs = flag.Int("runs", 10000, "stress cycle number")	
)


func main() {
	flag.Parse()
	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := timespans.CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	result := timespans.CallCost{}
	client, _ := jsonrpc.Dial("tcp", "localhost:5090")
	i := 0
	for j := 0; j < *runs; j++ {
		client.Call("Storage.GetCost", cd, &result)
	}
	log.Println(result)
	log.Println(i)
	client.Close()
}
