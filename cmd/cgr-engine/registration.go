/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"os"
	"os/signal"
	"syscall"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and closes the storage before exiting.
*/
func stopRaterSignalHandler(internalCdrStatSChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-c

	utils.Logger.Info(fmt.Sprintf("Caught signal %v", sig))
	var dummyInt int
	select {
	case cdrStats := <-internalCdrStatSChan:
		cdrStats.Call("CDRStatsV1.Stop", dummyInt, &dummyInt)
	default:
	}
	exitChan <- true
}

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and shuts down the session manager.
*/
func shutdownSessionmanagerSingnalHandler(exitChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-c
	exitChan <- true
}
