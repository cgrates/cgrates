package main

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
	"os/signal"
	"os"
	"syscall"
)

/*
RPC Server that handles the registering and unregistering of raters.
*/
type RaterServer byte

/*
Listens for SIGTERM, SIGINT, SIGQUIT system signals and shuts down all the registered raters.
*/
func StopSingnalHandler() {
	log.Print("Handling stop signals...")
	sig := <-signal.Incoming
	if usig, ok := sig.(os.UnixSignal); ok {
		switch usig {
		case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
			log.Printf("Caught signal %v, sending shutdownto raters\n", usig)
			var reply string
			for i,client:= range raterList.clientConnections {
				client.Call("Storage.Shutdown", "", &reply)
				log.Printf("Shutdown rater %v: %v ", raterList.clientAddresses[i], reply)
			}
			os.Exit(1)
		}
	}
}

/*
RPC method that receives a rater address, connects to it and ads the pair to the rater list for balancing
*/
func (rs *RaterServer) RegisterRater(clientAddress string, replay *byte) error {
	time.Sleep(1 * time.Second) // wait a second for Rater to start serving
	client, err := rpc.Dial("tcp", clientAddress)
	if err != nil {
		log.Print("Could not connect to client!")
		return err
	}
	raterList.AddClient(clientAddress, client)
	log.Print(fmt.Sprintf("Rater %v registered succesfully.", clientAddress))
	return nil
}

/*
RPC method that recives a rater addres gets the connections and closes it and removes the pair from rater list.
*/
func (rs *RaterServer) UnRegisterRater(clientAddress string, replay *byte) error {	
	client, ok := raterList.GetClient(clientAddress)
	if ok {
		client.Close()	
		raterList.RemoveClient(clientAddress)
		log.Print(fmt.Sprintf("Rater %v unregistered succesfully.", clientAddress))		
	} else {
		log.Print(fmt.Sprintf("Server %v was not on my watch!", clientAddress))		
	}
	return nil
}
