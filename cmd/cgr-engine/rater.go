/*
Real-time Charging System for Telecom & ISP environments
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
	"time"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/balancer2go"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func startBalancer(internalBalancerChan chan *balancer2go.Balancer, stopHandled *bool, exitChan chan bool) {
	bal := balancer2go.NewBalancer()
	go stopBalancerSignalHandler(bal, exitChan)
	*stopHandled = true
	internalBalancerChan <- bal
}

// Starts rater and reports on chan
func startRater(internalRaterChan chan *engine.Responder, internalBalancerChan chan *balancer2go.Balancer, internalSchedulerChan chan *scheduler.Scheduler,
	internalCdrStatSChan chan rpcclient.RpcClientConnection, internalHistorySChan chan rpcclient.RpcClientConnection,
	internalPubSubSChan chan rpcclient.RpcClientConnection, internalUserSChan chan rpcclient.RpcClientConnection, internalAliaseSChan chan rpcclient.RpcClientConnection,
	server *utils.Server,
	ratingDb engine.RatingStorage, accountDb engine.AccountingStorage, loadDb engine.LoadStorage, cdrDb engine.CdrStorage, logDb engine.LogStorage,
	stopHandled *bool, exitChan chan bool) {
	waitTasks := make([]chan struct{}, 0)

	//Cache load
	cacheTaskChan := make(chan struct{})
	waitTasks = append(waitTasks, cacheTaskChan)
	go func() {
		defer close(cacheTaskChan)
		if err := ratingDb.CacheRatingAll(); err != nil {
			utils.Logger.Crit(fmt.Sprintf("Cache rating error: %s", err.Error()))
			exitChan <- true
			return
		}
		if err := accountDb.CacheAccountingPrefixes(); err != nil { // Used to cache load history
			utils.Logger.Crit(fmt.Sprintf("Cache accounting error: %s", err.Error()))
			exitChan <- true
			return
		}

	}()

	// Retrieve scheduler for it's API methods
	var sched *scheduler.Scheduler // Need the scheduler in APIer
	if cfg.SchedulerEnabled {
		schedTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, schedTaskChan)
		go func() {
			defer close(schedTaskChan)
			select {
			case sched = <-internalSchedulerChan:
				internalSchedulerChan <- sched
			case <-time.After(cfg.InternalTtl):
				utils.Logger.Crit("<Rater>: Internal scheduler connection timeout.")
				exitChan <- true
				return
			}

		}()
	}

	// Connection to balancer
	var bal *balancer2go.Balancer
	if cfg.RaterBalancer != "" {
		balTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, balTaskChan)
		go func() {
			defer close(balTaskChan)
			if cfg.RaterBalancer == utils.INTERNAL {
				select {
				case bal = <-internalBalancerChan:
					internalBalancerChan <- bal // Put it back if someone else is interested about
				case <-time.After(cfg.InternalTtl):
					utils.Logger.Crit("<Rater>: Internal balancer connection timeout.")
					exitChan <- true
					return
				}
			} else {
				go registerToBalancer(exitChan)
				go stopRaterSignalHandler(internalCdrStatSChan, exitChan)
				*stopHandled = true
			}
		}()
	}

	// Connection to CDRStats
	var cdrStats rpcclient.RpcClientConnection
	if cfg.RaterCdrStats != "" {
		cdrstatTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, cdrstatTaskChan)
		go func() {
			defer close(cdrstatTaskChan)
			if cfg.RaterCdrStats == utils.INTERNAL {
				select {
				case cdrStats = <-internalCdrStatSChan:
					internalCdrStatSChan <- cdrStats
				case <-time.After(cfg.InternalTtl):
					utils.Logger.Crit("<Rater>: Internal cdrstats connection timeout.")
					exitChan <- true
					return
				}
			} else if cdrStats, err = rpcclient.NewRpcClient("tcp", cfg.RaterCdrStats, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<Rater> Could not connect to cdrstats, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	// Connection to HistoryS
	if cfg.RaterHistoryServer != "" {
		histTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, histTaskChan)
		go func() {
			defer close(histTaskChan)
			var scribeServer rpcclient.RpcClientConnection
			if cfg.RaterHistoryServer == utils.INTERNAL {
				select {
				case scribeServer = <-internalHistorySChan:
					internalHistorySChan <- scribeServer
				case <-time.After(cfg.InternalTtl):
					utils.Logger.Crit("<Rater>: Internal historys connection timeout.")
					exitChan <- true
					return
				}
			} else if scribeServer, err = rpcclient.NewRpcClient("tcp", cfg.RaterHistoryServer, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<Rater> Could not connect historys, error: %s", err.Error()))
				exitChan <- true
				return
			}
			engine.SetHistoryScribe(scribeServer) // ToDo: replace package sharing with connection based one
		}()
	}

	// Connection to pubsubs
	if cfg.RaterPubSubServer != "" {
		pubsubTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, pubsubTaskChan)
		go func() {
			defer close(pubsubTaskChan)
			var pubSubServer rpcclient.RpcClientConnection
			if cfg.RaterPubSubServer == utils.INTERNAL {
				select {
				case pubSubServer = <-internalPubSubSChan:
					internalPubSubSChan <- pubSubServer
				case <-time.After(cfg.InternalTtl):
					utils.Logger.Crit("<Rater>: Internal pubsub connection timeout.")
					exitChan <- true
					return
				}
			} else if pubSubServer, err = rpcclient.NewRpcClient("tcp", cfg.RaterPubSubServer, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<Rater> Could not connect to pubsubs: %s", err.Error()))
				exitChan <- true
				return
			}
			engine.SetPubSub(pubSubServer) // ToDo: replace package sharing with connection based one
		}()
	}

	// Connection to AliasService
	if cfg.RaterAliasesServer != "" {
		aliasesTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, aliasesTaskChan)
		go func() {
			defer close(aliasesTaskChan)
			var aliasesServer rpcclient.RpcClientConnection
			if cfg.RaterAliasesServer == utils.INTERNAL {
				select {
				case aliasesServer = <-internalAliaseSChan:
					internalAliaseSChan <- aliasesServer
				case <-time.After(cfg.InternalTtl):
					utils.Logger.Crit("<Rater>: Internal aliases connection timeout.")
					exitChan <- true
					return
				}
			} else if aliasesServer, err = rpcclient.NewRpcClient("tcp", cfg.RaterAliasesServer, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<Rater> Could not connect to aliases, error: %s", err.Error()))
				exitChan <- true
				return
			}
			engine.SetAliasService(aliasesServer) // ToDo: replace package sharing with connection based one
		}()
	}

	// Connection to UserService
	var userServer rpcclient.RpcClientConnection
	if cfg.RaterUserServer != "" {
		usersTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, usersTaskChan)
		go func() {
			defer close(usersTaskChan)
			if cfg.RaterUserServer == utils.INTERNAL {
				select {
				case userServer = <-internalUserSChan:
					internalUserSChan <- userServer
				case <-time.After(cfg.InternalTtl):
					utils.Logger.Crit("<Rater>: Internal users connection timeout.")
					exitChan <- true
					return
				}
			} else if userServer, err = rpcclient.NewRpcClient("tcp", cfg.RaterUserServer, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<Rater> Could not connect users, error: %s", err.Error()))
				exitChan <- true
				return
			}
			engine.SetUserService(userServer)
		}()
	}

	// Wait for all connections to complete before going further
	for _, chn := range waitTasks {
		<-chn
	}

	responder := &engine.Responder{Bal: bal, ExitChan: exitChan, Stats: cdrStats}
	apierRpcV1 := &v1.ApierV1{StorDb: loadDb, RatingDb: ratingDb, AccountDb: accountDb, CdrDb: cdrDb, LogDb: logDb, Sched: sched,
		Config: cfg, Responder: responder, CdrStatsSrv: cdrStats, Users: userServer}
	apierRpcV2 := &v2.ApierV2{
		ApierV1: *apierRpcV1}
	// internalSchedulerChan shared here
	server.RpcRegister(responder)
	server.RpcRegister(apierRpcV1)
	server.RpcRegister(apierRpcV2)
	internalRaterChan <- responder // Rater done
}
