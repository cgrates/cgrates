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
	"fmt"

	"github.com/cgrates/cgrates/timespans"
	"log"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type RaterServer struct{}

/*
RPC method that receives a rater address, connects to it and ads the pair to the rater list for balancing
*/
func (rs *RaterServer) RegisterRater(clientAddress string, replay *int) error {
	log.Printf("Started rater %v registration...", clientAddress)
	time.Sleep(2 * time.Second) // wait a second for Rater to start serving
	client, err := rpc.Dial("tcp", clientAddress)
	if err != nil {
		log.Print("Could not connect to client!")
		return err
	}
	bal.AddClient(clientAddress, client)
	log.Printf("Rater %v registered succesfully.", clientAddress)
	return nil
}

/*
RPC method that recives a rater addres gets the connections and closes it and removes the pair from rater list.
*/
func (rs *RaterServer) UnRegisterRater(clientAddress string, replay *int) error {
	client, ok := bal.GetClient(clientAddress)
	if ok {
		client.Close()
		bal.RemoveClient(clientAddress)
		log.Print(fmt.Sprintf("Rater %v unregistered succesfully.", clientAddress))
	} else {
		log.Print(fmt.Sprintf("Server %v was not on my watch!", clientAddress))
	}
	return nil
}

/*
Listens for SIGTERM, SIGINT, SIGQUIT system signals and shuts down all the registered raters.
*/
func stopBalancerSingnalHandler() {
	log.Print("Handling stop signals...")
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-c
	log.Printf("Caught signal %v, sending shutdownto raters\n", sig)
	bal.Shutdown()
	exitChan <- true
}

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and  gracefuly unregister from balancer and closes the storage before exiting.
*/
func stopRaterSingnalHandler() {
	log.Print("Handling stop signals...")
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-c

	log.Printf("Caught signal %v, unregistering from balancer\n", sig)
	unregisterFromBalancer()
	exitChan <- true
}

/*
Connects to the balancer and calls unregister RPC method.
*/
func unregisterFromBalancer() {
	client, err := rpc.Dial("tcp", rater_balancer_server)
	if err != nil {
		log.Print("Cannot contact the balancer!")
		exitChan <- true
		return
	}
	var reply int
	log.Print("Unregistering from balancer ", rater_balancer_server)
	client.Call("RaterServer.UnRegisterRater", rater_listen, &reply)
	if err := client.Close(); err != nil {
		log.Print("Could not close balancer unregistration!")
		exitChan <- true
	}
}

/*
Connects to the balancer and rehisters the rater to the server.
*/
func registerToBalancer() {
	client, err := rpc.Dial("tcp", rater_balancer_server)
	if err != nil {
		log.Print("Cannot contact the balancer!")
		exitChan <- true
		return
	}
	var reply int
	log.Print("Registering to balancer ", rater_balancer_server)
	client.Call("RaterServer.RegisterRater", rater_listen, &reply)
	if err := client.Close(); err != nil {
		log.Print("Could not close balancer registration!")
		exitChan <- true
	}
	log.Print("Registration finished!")
}

// Listens for the HUP system signal and gracefuly reloads the timers from database.
func reloadSchedulerSingnalHandler() {
	timespans.Logger.Info("Handling HUP signal...")
	for {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGHUP)
		sig := <-c

		timespans.Logger.Info(fmt.Sprintf("Caught signal %v, reloading action timings.\n", sig))
		loadActionTimings()
		// check the tip of the queue for new actions
		restartLoop <- 1
		timer.Stop()
	}
}
