package main

import (
	"os/signal"
	"github.com/rif/cgrates/timespans"
	"log"
	"net/rpc"
	"os"
	"syscall"
)

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and  gracefuly unregister from inquirer and closes the storage before exiting.
*/
func StopSingnalHandler(server, listen *string, sg timespans.StorageGetter) {
	log.Print("Handling stop signals...")
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)	
	sig := <-c

	log.Printf("Caught signal %v, unregistering from server\n", sig)
	unregisterFromServer(server, listen)
	sg.Close()
	os.Exit(1)
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
