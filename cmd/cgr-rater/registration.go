/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/rater"
	"github.com/cgrates/cgrates/scheduler"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
)

/*
Listens for SIGTERM, SIGINT, SIGQUIT system signals and shuts down all the registered raters.
*/
func stopBalancerSingnalHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-c
	rater.Logger.Info(fmt.Sprintf("Caught signal %v, sending shutdownto raters\n", sig))
	bal.Shutdown("Responder.Shutdown")
	exitChan <- true
}

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and  gracefuly unregister from balancer and closes the storage before exiting.
*/
func stopRaterSingnalHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-c

	rater.Logger.Info(fmt.Sprintf("Caught signal %v, unregistering from balancer\n", sig))
	unregisterFromBalancer()
	exitChan <- true
}

/*
Connects to the balancer and calls unregister RPC method.
*/
func unregisterFromBalancer() {
	client, err := rpc.Dial("tcp", cfg.RaterBalancer)
	if err != nil {
		rater.Logger.Crit("Cannot contact the balancer!")
		exitChan <- true
		return
	}
	var reply int
	rater.Logger.Info(fmt.Sprintf("Unregistering from balancer %s", cfg.RaterBalancer))
	client.Call("Responder.UnRegisterRater", cfg.RaterListen, &reply)
	if err := client.Close(); err != nil {
		rater.Logger.Crit("Could not close balancer unregistration!")
		exitChan <- true
	}
}

/*
Connects to the balancer and rehisters the rater to the server.
*/
func registerToBalancer() {
	client, err := rpc.Dial("tcp", cfg.RaterBalancer)
	if err != nil {
		rater.Logger.Crit(fmt.Sprintf("Cannot contact the balancer: %v", err))
		exitChan <- true
		return
	}
	var reply int
	rater.Logger.Info(fmt.Sprintf("Registering to balancer %s", cfg.RaterBalancer))
	client.Call("Responder.RegisterRater", cfg.RaterListen, &reply)
	if err := client.Close(); err != nil {
		rater.Logger.Crit("Could not close balancer registration!")
		exitChan <- true
	}
	rater.Logger.Info("Registration finished!")
}

// Listens for the HUP system signal and gracefuly reloads the timers from database.
func reloadSchedulerSingnalHandler(sched *scheduler.Scheduler, getter rater.DataStorage) {
	for {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGHUP)
		sig := <-c

		rater.Logger.Info(fmt.Sprintf("Caught signal %v, reloading action timings.\n", sig))
		sched.LoadActionTimings(getter)
		// check the tip of the queue for new actions
		sched.Restart()
	}
}

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and shuts down the session manager.
*/
func shutdownSessionmanagerSingnalHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-c

	if err := sm.Shutdown(); err != nil {
		rater.Logger.Warning(fmt.Sprintf("<SessionManager> %s", err))
	}
	exitChan <- true
}
