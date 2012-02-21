package main

import (
	"flag"
	"github.com/rif/cgrates/timespans"
	"log"
	"net/rpc/jsonrpc"
	"time"
)

var (
	runs     = flag.Int("runs", 10000, "stress cycle number")
	parallel = flag.Bool("parallel", false, "run requests in parallel")
)

func main() {
	flag.Parse()
	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := timespans.CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	result := timespans.CallCost{}
	client, _ := jsonrpc.Dial("tcp", "localhost:2001")
	i := 0
	if *parallel {
		c := make(chan string)
		for ; i < *runs; i++ {
			go func() {
				var reply string
				client.Call("Responder.Get", cd, &result)
				c <- reply
			}()
			//time.Sleep(1*time.Second)
		}
		for j := 0; j < *runs; j++ {
			<-c
		}
	} else {
		for j := 0; j < *runs; j++ {
			client.Call("Responder.Get", cd, &result)
		}
	}
	log.Println(result)
	log.Println(i)
	client.Close()
}
