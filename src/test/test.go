package main

import (
	"net/rpc"
	"fmt"
) 


func main(){
	client, _ := rpc.DialHTTPPath("tcp", "localhost:2000", "/rpc")
	var reply string
	client.Call("Responder.Get", "test", &reply)
	fmt.Println(reply)
}


