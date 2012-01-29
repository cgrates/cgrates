package main

import (
	"net/rpc/jsonrpc"
	"fmt"
	//"time"
) 


func main(){
	client, _ := jsonrpc.Dial("tcp", "localhost:5090")
	var reply string
	i:= 0
	for ; i < 5 * 10e3; i++ {
		client.Call("Responder.Get", "test", &reply)
	//time.Sleep(1*time.Second)
	}
	fmt.Println(i, reply)
}


