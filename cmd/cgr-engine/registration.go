/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/cgrates/cgrates/balancer2go"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
)

/*
Listens for SIGTERM, SIGINT, SIGQUIT system signals and shuts down all the registered engines.
*/
func stopBalancerSignalHandler(bal *balancer2go.Balancer, exitChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-c
	engine.Logger.Info(fmt.Sprintf("Caught signal %v, sending shutdown to engines\n", sig))
	bal.Shutdown("Responder.Shutdown")
	exitChan <- true
}

func generalSignalHandler(internalCdrStatSChan chan engine.StatsInterface, exitChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-c
	engine.Logger.Info(fmt.Sprintf("Caught signal %v, shuting down cgr-engine\n", sig))
	var dummyInt int
	select {
	case cdrStats := <-internalCdrStatSChan:
		cdrStats.Stop(dummyInt, &dummyInt)
	default:
	}

	exitChan <- true
}

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and  gracefuly unregister from balancer and closes the storage before exiting.
*/
func stopRaterSignalHandler(internalCdrStatSChan chan engine.StatsInterface, exitChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-c

	engine.Logger.Info(fmt.Sprintf("Caught signal %v, unregistering from balancer\n", sig))
	unregisterFromBalancer(exitChan)
	var dummyInt int
	select {
	case cdrStats := <-internalCdrStatSChan:
		cdrStats.Stop(dummyInt, &dummyInt)
	default:
	}
	exitChan <- true
}

/*
Connects to the balancer and calls unregister RPC method.
*/
func unregisterFromBalancer(exitChan chan bool) {
	client, err := rpc.Dial("tcp", cfg.RaterBalancer)
	if err != nil {
		engine.Logger.Crit("Cannot contact the balancer!")
		exitChan <- true
		return
	}
	var reply int
	engine.Logger.Info(fmt.Sprintf("Unregistering from balancer %s", cfg.RaterBalancer))
	client.Call("Responder.UnRegisterRater", cfg.RPCGOBListen, &reply)
	if err := client.Close(); err != nil {
		engine.Logger.Crit("Could not close balancer unregistration!")
		exitChan <- true
	}
}

/*
Connects to the balancer and rehisters the engine to the server.
*/
func registerToBalancer(exitChan chan bool) {
	client, err := rpc.Dial("tcp", cfg.RaterBalancer)
	if err != nil {
		engine.Logger.Crit(fmt.Sprintf("Cannot contact the balancer: %v", err))
		exitChan <- true
		return
	}
	var reply int
	engine.Logger.Info(fmt.Sprintf("Registering to balancer %s", cfg.RaterBalancer))
	client.Call("Responder.RegisterRater", cfg.RPCGOBListen, &reply)
	if err := client.Close(); err != nil {
		engine.Logger.Crit("Could not close balancer registration!")
		exitChan <- true
	}
	engine.Logger.Info("Registration finished!")
}

// Listens for the HUP system signal and gracefuly reloads the timers from database.
func reloadSchedulerSingnalHandler(sched *scheduler.Scheduler, getter engine.RatingStorage) {
	for {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGHUP)
		sig := <-c

		engine.Logger.Info(fmt.Sprintf("Caught signal %v, reloading action timings.\n", sig))
		sched.LoadActionPlans(getter)
		// check the tip of the queue for new actions
		sched.Restart()
	}
}

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and shuts down the session manager.
*/
func shutdownSessionmanagerSingnalHandler(exitChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-c

	for _, sm := range sms {
		if err := sm.Shutdown(); err != nil {
			engine.Logger.Warning(fmt.Sprintf("<SessionManager> %s", err))
		}
	}
	exitChan <- true
}
