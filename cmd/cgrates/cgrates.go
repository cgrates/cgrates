package main

import (
	"flag"
	"github.com/rif/cgrates/timespans"
	"log"	
	"net/rpc/jsonrpc"
	"time"
)

var (
	balancer = flag.String("balancer", "127.0.0.1:2001", "balancer address host:port")
)



func main(){
	flag.Parse()
	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := timespans.CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	result := timespans.CallCost{}
	client, _ := jsonrpc.Dial("tcp", "localhost:2001")
	client.Call("Responder.GetCost", cd, &result)
	log.Println(result)
	client.Close()
	log.Print("done!")
}

