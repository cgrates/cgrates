package main

import (
	"log"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
)

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and  gracefuly unregister from inquirer and closes the storage before exiting.
*/
func StopSingnalHandler(server, listen *string, getter *KyotoStorage) {
	log.Print("Handling stop signals...")
	sig := <-signal.Incoming
	if usig, ok := sig.(os.UnixSignal); ok {
		switch usig {
		case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
			log.Printf("Caught signal %v, unregistering from server\n", usig)
			unregisterFromServer(server, listen)
			getter.Close()
			os.Exit(1)
		}
	}
}

/*
Connects to the inquirer and calls unregister RPC method.
*/
func unregisterFromServer(server, listen *string) {
	client, err := rpc.DialHTTP("tcp", *server)
	if err != nil {
		log.Print("Cannot contact the server!")
		os.Exit(1)
	}
	var reply byte
	log.Print("Unregistering from server ", *server)
	client.Call("RaterServer.UnRegisterRater", *listen, &reply)
	if err := client.Close(); err != nil {
		log.Print("Could not close server unregistration!")
		os.Exit(1)
	}
}

/*
Connects to the inquirer and rehisters the rater to the server.
*/
func RegisterToServer(server, listen *string) {
	client, err := rpc.DialHTTP("tcp", *server)
	if err != nil {
		log.Print("Cannot contact the server!")
		os.Exit(1)
	}
	var reply byte
	log.Print("Registering to server ", *server)
	client.Call("RaterServer.RegisterRater", *listen, &reply)
	if err := client.Close(); err != nil {
		log.Print("Could not close server registration!")
		os.Exit(1)
	}
}
