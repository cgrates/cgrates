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
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/timespans"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
)

/*
Listens for SIGTERM, SIGINT, SIGQUIT system signals and shuts down all the registered raters.
*/
func stopBalancerSingnalHandler() {
	timespans.Logger.Info("Handling stop signals...")
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-c
	timespans.Logger.Info(fmt.Sprintf("Caught signal %v, sending shutdownto raters\n", sig))
	bal.Shutdown("Responder.Shutdown")
	exitChan <- true
}

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and  gracefuly unregister from balancer and closes the storage before exiting.
*/
func stopRaterSingnalHandler() {
	timespans.Logger.Info("Handling stop signals...")
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-c

	timespans.Logger.Info(fmt.Sprintf("Caught signal %v, unregistering from balancer\n", sig))
	unregisterFromBalancer()
	exitChan <- true
}

/*
Connects to the balancer and calls unregister RPC method.
*/
func unregisterFromBalancer() {
	client, err := rpc.Dial("tcp", rater_balancer)
	if err != nil {
		timespans.Logger.Crit("Cannot contact the balancer!")
		exitChan <- true
		return
	}
	var reply int
	timespans.Logger.Info(fmt.Sprintf("Unregistering from balancer ", rater_balancer))
	client.Call("Responder.UnRegisterRater", rater_listen, &reply)
	if err := client.Close(); err != nil {
		timespans.Logger.Crit("Could not close balancer unregistration!")
		exitChan <- true
	}
}

/*
Connects to the balancer and rehisters the rater to the server.
*/
func registerToBalancer() {
	client, err := rpc.Dial("tcp", rater_balancer)
	if err != nil {
		timespans.Logger.Crit(fmt.Sprintf("Cannot contact the balancer!", err))
		exitChan <- true
		return
	}
	var reply int
	timespans.Logger.Info(fmt.Sprintf("Registering to balancer ", rater_balancer))
	client.Call("Responder.RegisterRater", rater_listen, &reply)
	if err := client.Close(); err != nil {
		timespans.Logger.Crit("Could not close balancer registration!")
		exitChan <- true
	}
	timespans.Logger.Info("Registration finished!")
}

// Listens for the HUP system signal and gracefuly reloads the timers from database.
func reloadSchedulerSingnalHandler(sched *scheduler.Scheduler, getter timespans.DataStorage) {
	timespans.Logger.Info("Handling HUP signal...")
	for {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGHUP)
		sig := <-c

		timespans.Logger.Info(fmt.Sprintf("Caught signal %v, reloading action timings.\n", sig))
		sched.LoadActionTimings(getter)
		// check the tip of the queue for new actions
		sched.Restart()
	}
}
