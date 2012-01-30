package main

import (
	"net/rpc/jsonrpc"
	"log"
	//"time"
) 


func main(){
	client, _ := jsonrpc.Dial("tcp", "localhost:5090")
	runs := int(5 * 10e3);
	i:= 0
	c := make(chan string)
	for ; i < runs; i++ {
		go func(){
			var reply string
			client.Call("Responder.Get", "test", &reply)
			c <- reply
		}()
	//time.Sleep(1*time.Second)
	}
	for j:=0; j < runs; j++ {
		<-c
	}
	log.Print(i)
	client.Close()
}


