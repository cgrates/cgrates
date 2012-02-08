package main

import (
	"net/rpc/jsonrpc"
	"log"
	"github.com/rif/cgrates/timespans"
	"time"
) 

func main(){
	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := timespans.CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	result := timespans.CallCost{}
	client, _ := jsonrpc.Dial("tcp", "localhost:5090")
	runs := int(5 * 1e4);
	i:= 0
	c := make(chan string)
	for ; i < runs; i++ {
		go func(){
			var reply string
			client.Call("Responder.Get", cd, &result)
			c <- reply
		}()
	//time.Sleep(1*time.Second)
	}
	for j:=0; j < runs; j++ {
		<-c
	}
	log.Println(result)
	log.Println(i)
	client.Close()
}


