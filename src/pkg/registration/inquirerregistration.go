package registration

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

type RaterList struct {
	Clients map[string]*rpc.Client
}

func (rl *RaterList) RegisterRater(clientAddress string, replay *byte) error {
	time.Sleep(1 * time.Second) // wait a second for Rater to start serving
	client, err := rpc.Dial("tcp", clientAddress)
	if err != nil {
		log.Panic("Could not connect to client!")
	}
	rl.Clients[clientAddress] = client
	log.Print(fmt.Sprintf("Server %v registered succesfully", clientAddress))
	return nil
}

func (rl *RaterList) UnRegisterRater(clientAddress string, replay *byte) error {
	client := rl.Clients[clientAddress]
	client.Close()	
	delete(rl.Clients, clientAddress)
	log.Print(fmt.Sprintf("Server %v unregistered succesfully", clientAddress))
	return nil
}