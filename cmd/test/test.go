package main

import (
	"net/rpc"
	"fmt"
) 


func main(){
	client, _ := rpc.DialHTTPPath("tcp", "localhost:2000", "/rpc")
	var reply string
	i:= 0
	for ; i < 5 * 10e4; i++ {
		client.Call("Responder.Get", "test", &reply)
	}
	fmt.Println(i, reply)
}


