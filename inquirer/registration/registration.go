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
		log.Print("Could not connect to client!")
		return err
	}
	rl.Clients[clientAddress] = client
	log.Print(fmt.Sprintf("Rater %v registered succesfully.", clientAddress))
	if len(rl.Clients) == 1 {
		// unlock the balancer on first rater
		rl.balancer_mutex.Unlock()
	}
	return nil
}

func (rl *RaterList) UnRegisterRater(clientAddress string, replay *byte) error {
	
	client, ok := rl.Clients[clientAddress]
	if ok {
		client.Close()	
		delete(rl.Clients, clientAddress)
		log.Print(fmt.Sprintf("Rater %v unregistered succesfully.", clientAddress))		
	} else {
		log.Print(fmt.Sprintf("Server %v was not on my watch!", clientAddress))		
	}
	return nil
}

func (rl *RaterList) startBalance() {	
	rl.balancer_mutex.Lock()
	go func(){		
		for {
			rl.balancer_mutex.Lock()
			for _, client := range rl.Clients {
				//log.Printf("using server %s:", addr)
				rl.Balancer <- client			
			}
			if len(rl.Clients) != 0 {			
				rl.balancer_mutex.Unlock()
			}			
		}
	}()
}
