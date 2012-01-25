package main

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
	"sync"
)

type RaterList struct {
	Clients map[string]*rpc.Client
	Balancer chan *rpc.Client
	balancer_mutex sync.Mutex
}

func NewRaterList() *RaterList {
	r:= &RaterList{
		Clients: make(map[string]*rpc.Client),
		Balancer: make(chan *rpc.Client),
		}
	r.startBalance()
	return r
}

func (rl *RaterList) RegisterRater(clientAddress string, replay *byte) error {
	time.Sleep(1 * time.Second) // wait a second for Rater to start serving
	client, err := rpc.Dial("tcp", clientAddress)
	if err != nil {
		log.Panic("Could not connect to client!")
	}
	rl.Clients[clientAddress] = client
	log.Print(fmt.Sprintf("Server %v registered succesfully", clientAddress))
	rl.balancer_mutex.Unlock()
	return nil
}

func (rl *RaterList) UnRegisterRater(clientAddress string, replay *byte) error {
	client := rl.Clients[clientAddress]
	client.Close()	
	delete(rl.Clients, clientAddress)
	log.Print(fmt.Sprintf("Server %v unregistered succesfully", clientAddress))		
	return nil
}

func (rl *RaterList) startBalance() {	
	rl.balancer_mutex.Lock()
	go func(){		
		for {
			rl.balancer_mutex.Lock()
			log.Print("balancing")
			for addr, client := range rl.Clients {
				log.Printf("using server %s:", addr)
				rl.Balancer <- client			
			}
			if len(rl.Clients) != 0 {			
				rl.balancer_mutex.Unlock()
			}			
		}
	}()
}
