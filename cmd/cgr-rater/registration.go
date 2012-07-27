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
)

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
	client, err := rpc.Dial("tcp", rater_balancer)
	if err != nil {
		log.Print("Cannot contact the balancer!")
		exitChan <- true
		return
	}
	var reply int
	log.Print("Unregistering from balancer ", rater_balancer)
	client.Call("Responder.UnRegisterRater", rater_listen, &reply)
	if err := client.Close(); err != nil {
		log.Print("Could not close balancer unregistration!")
		exitChan <- true
	}
}

/*
Connects to the balancer and rehisters the rater to the server.
*/
func registerToBalancer() {
	client, err := rpc.Dial("tcp", rater_balancer)
	if err != nil {
		log.Print("Cannot contact the balancer!")
		exitChan <- true
		return
	}
	var reply int
	log.Print("Registering to balancer ", rater_balancer)
	client.Call("Responder.RegisterRater", rater_listen, &reply)
	if err := client.Close(); err != nil {
		log.Print("Could not close balancer registration!")
		exitChan <- true
	}
	log.Print("Registration finished!")
}

// Listens for the HUP system signal and gracefuly reloads the timers from database.
func reloadSchedulerSingnalHandler(getter timespans.StorageGetter) {
	timespans.Logger.Info("Handling HUP signal...")
	for {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGHUP)
		sig := <-c

		timespans.Logger.Info(fmt.Sprintf("Caught signal %v, reloading action timings.\n", sig))
		loadActionTimings(getter)
		// check the tip of the queue for new actions
		restartLoop <- 1
		timer.Stop()
	}
}
