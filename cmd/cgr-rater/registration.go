/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package main

import (
	"github.com/rif/cgrates/timespans"
	"log"
	"net/rpc"
	"os"
	"os/signal"
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
	log.Print("Registration finished!")
}
