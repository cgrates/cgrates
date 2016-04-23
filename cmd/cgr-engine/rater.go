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

func startRater(internalRaterChan chan rpcclient.RpcClientConnection, cacheDoneChan chan struct{}, internalBalancerChan chan *balancer2go.Balancer, internalSchedulerChan chan *scheduler.Scheduler,
	internalCdrStatSChan chan rpcclient.RpcClientConnection, internalHistorySChan chan rpcclient.RpcClientConnection,
	internalPubSubSChan chan rpcclient.RpcClientConnection, internalUserSChan chan rpcclient.RpcClientConnection, internalAliaseSChan chan rpcclient.RpcClientConnection,
	server *utils.Server,
	ratingDb engine.RatingStorage, accountDb engine.AccountingStorage, loadDb engine.LoadStorage, cdrDb engine.CdrStorage, logDb engine.LogStorage,
	stopHandled *bool, exitChan chan bool) {
	var waitTasks []chan struct{}

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
		cacheDoneChan <- struct{}{}
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
	if cfg.RALsBalancer != "" {
		balTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, balTaskChan)
		go func() {
			defer close(balTaskChan)
			if cfg.RALsBalancer == utils.INTERNAL {
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
	// Connections to CDRStats
	var cdrStats *rpcclient.RpcClientPool
	if len(cfg.RALsCDRStatSConns) != 0 {
		cdrstatTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, cdrstatTaskChan)
		go func() {
			defer close(cdrstatTaskChan)
			cdrStats, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB,
				cfg.CDRSRaterConns, internalCdrStatSChan, cfg.InternalTtl)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to CDRStatS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	// Connection to HistoryS,
	if len(cfg.RALsHistorySConns) != 0 {
		histTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, histTaskChan)
		go func() {
			defer close(histTaskChan)
			if historySConns, err := engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB,
				cfg.RALsHistorySConns, internalHistorySChan, cfg.InternalTtl); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect HistoryS, error: %s", err.Error()))
				exitChan <- true
				return
			} else {
				engine.SetHistoryScribe(historySConns)
			}
		}()
	}
	// Connection to pubsubs
	if len(cfg.RALsPubSubSConns) != 0 {
		pubsubTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, pubsubTaskChan)
		go func() {
			defer close(pubsubTaskChan)
			if pubSubSConns, err := engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB,
				cfg.RALsPubSubSConns, internalPubSubSChan, cfg.InternalTtl); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to PubSubS: %s", err.Error()))
				exitChan <- true
				return
			} else {
				engine.SetPubSub(pubSubSConns)
			}
		}()
	}
	// Connection to AliasService
	if len(cfg.RALsAliasSConns) != 0 {
		aliasesTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, aliasesTaskChan)
		go func() {
			defer close(aliasesTaskChan)
			if aliaseSCons, err := engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB,
				cfg.RALsAliasSConns, internalAliaseSChan, cfg.InternalTtl); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to AliaseS, error: %s", err.Error()))
				exitChan <- true
				return
			} else {
				engine.SetAliasService(aliaseSCons)
			}
		}()
	}
	// Connection to UserService
	var usersConns rpcclient.RpcClientConnection
	if len(cfg.RALsUserSConns) != 0 {
		usersTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, usersTaskChan)
		go func() {
			defer close(usersTaskChan)
			if usersConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB,
				cfg.RALsAliasSConns, internalAliaseSChan, cfg.InternalTtl); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect UserS, error: %s", err.Error()))
				exitChan <- true
				return
			}
			engine.SetUserService(usersConns)
		}()
	}

	// Wait for all connections to complete before going further
	for _, chn := range waitTasks {
		<-chn
	}

	responder := &engine.Responder{Bal: bal, ExitChan: exitChan, Stats: cdrStats}
	responder.SetTimeToLive(cfg.ResponseCacheTTL, nil)
	apierRpcV1 := &v1.ApierV1{StorDb: loadDb, RatingDb: ratingDb, AccountDb: accountDb, CdrDb: cdrDb, LogDb: logDb, Sched: sched,
		Config: cfg, Responder: responder, CdrStatsSrv: cdrStats, Users: usersConns}
	apierRpcV2 := &v2.ApierV2{
		ApierV1: *apierRpcV1}
	// internalSchedulerChan shared here
	server.RpcRegister(responder)
	server.RpcRegister(apierRpcV1)
	server.RpcRegister(apierRpcV2)
	internalRaterChan <- responder // Rater done
}
