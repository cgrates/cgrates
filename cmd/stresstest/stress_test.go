package main

import (
	"net/rpc"
	"fmt"
	//"time"
) 


func main(){
	client, _ := rpc.DialHTTPPath("tcp", "localhost:2000", "/rpc")
	var reply string
	i:= 0
	for ; i < 5 * 10e3; i++ {
		client.Call("Responder.Get", "test", &reply)
	//time.Sleep(1*time.Second)
	}
	fmt.Println(i, reply)
}


