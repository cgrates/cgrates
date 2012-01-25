package main

import (
	"net/rpc"
	"fmt"
) 


func main(){
	client, _ := rpc.DialHTTPPath("tcp", "localhost:2000", "/rpc")
	var reply string
	for i:= 0; i < 5 * 10e6; i++ {
		client.Call("Responder.Get", "test", &reply)
	}
	fmt.Println(reply)
}


