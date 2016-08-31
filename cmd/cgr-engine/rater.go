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
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

/*func init() {
	gob.Register(map[interface{}]struct{}{})
	gob.Register(engine.Actions{})
	gob.RegisterName("github.com/cgrates/cgrates/engine.ActionPlan", &engine.ActionPlan{})
	gob.Register([]*utils.LoadInstance{})
	gob.RegisterName("github.com/cgrates/cgrates/engine.RatingPlan", &engine.RatingPlan{})
	gob.RegisterName("github.com/cgrates/cgrates/engine.RatingProfile", &engine.RatingProfile{})
	gob.RegisterName("github.com/cgrates/cgrates/utils.DerivedChargers", &utils.DerivedChargers{})
	gob.Register(engine.AliasValues{})
}*/

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
	ratingDb engine.RatingStorage, accountDb engine.AccountingStorage, loadDb engine.LoadStorage, cdrDb engine.CdrStorage, stopHandled *bool, exitChan chan bool) {
	var waitTasks []chan struct{}

	//Cache load
	cacheTaskChan := make(chan struct{})
	waitTasks = append(waitTasks, cacheTaskChan)
	go func() {
		defer close(cacheTaskChan)

		loadHist, err := accountDb.GetLoadHistory(1, true, utils.NonTransactional)
		if err != nil || len(loadHist) == 0 {
			utils.Logger.Info(fmt.Sprintf("could not get load history: %v (%v)", loadHist, err))
			cacheDoneChan <- struct{}{}
			return
		}

		if err := ratingDb.PreloadRatingCache(); err != nil {
			utils.Logger.Crit(fmt.Sprintf("Cache rating error: %s", err.Error()))
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
	var bal *balancer2go.Balancer
	if cfg.RALsBalancer != "" { // Connection to balancer
		balTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, balTaskChan)
		go func() {
			defer close(balTaskChan)
			if cfg.RALsBalancer == utils.MetaInternal {
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
	var cdrStats *rpcclient.RpcClientPool
	if len(cfg.RALsCDRStatSConns) != 0 { // Connections to CDRStats
		cdrstatTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, cdrstatTaskChan)
		go func() {
			defer close(cdrstatTaskChan)
			cdrStats, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsCDRStatSConns, internalCdrStatSChan, cfg.InternalTtl)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to CDRStatS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}
	if len(cfg.RALsHistorySConns) != 0 { // Connection to HistoryS,
		histTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, histTaskChan)
		go func() {
			defer close(histTaskChan)
			if historySConns, err := engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsHistorySConns, internalHistorySChan, cfg.InternalTtl); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect HistoryS, error: %s", err.Error()))
				exitChan <- true
				return
			} else {
				engine.SetHistoryScribe(historySConns)
			}
		}()
	}
	if len(cfg.RALsPubSubSConns) != 0 { // Connection to pubsubs
		pubsubTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, pubsubTaskChan)
		go func() {
			defer close(pubsubTaskChan)
			if pubSubSConns, err := engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsPubSubSConns, internalPubSubSChan, cfg.InternalTtl); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to PubSubS: %s", err.Error()))
				exitChan <- true
				return
			} else {
				engine.SetPubSub(pubSubSConns)
			}
		}()
	}
	if len(cfg.RALsAliasSConns) != 0 { // Connection to AliasService
		aliasesTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, aliasesTaskChan)
		go func() {
			defer close(aliasesTaskChan)
			if aliaseSCons, err := engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsAliasSConns, internalAliaseSChan, cfg.InternalTtl); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to AliaseS, error: %s", err.Error()))
				exitChan <- true
				return
			} else {
				engine.SetAliasService(aliaseSCons)
			}
		}()
	}
	var usersConns rpcclient.RpcClientConnection
	if len(cfg.RALsUserSConns) != 0 { // Connection to UserService
		usersTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, usersTaskChan)
		go func() {
			defer close(usersTaskChan)
			if usersConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsUserSConns, internalUserSChan, cfg.InternalTtl); err != nil {
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
	responder := &engine.Responder{Bal: bal, ExitChan: exitChan}
	responder.SetTimeToLive(cfg.ResponseCacheTTL, nil)
	apierRpcV1 := &v1.ApierV1{StorDb: loadDb, RatingDb: ratingDb, AccountDb: accountDb, CdrDb: cdrDb, Sched: sched,
		Config: cfg, Responder: responder}
	if cdrStats != nil { // ToDo: Fix here properly the init of stats
		responder.Stats = cdrStats
		apierRpcV1.CdrStatsSrv = cdrStats
	}
	if usersConns != nil {
		apierRpcV1.Users = usersConns
	}
	apierRpcV2 := &v2.ApierV2{
		ApierV1: *apierRpcV1}

	// internalSchedulerChan shared here
	server.RpcRegister(responder)
	server.RpcRegister(apierRpcV1)
	server.RpcRegister(apierRpcV2)

	utils.RegisterRpcParams("", &engine.Stats{})
	utils.RegisterRpcParams("", &v1.CDRStatsV1{})
	utils.RegisterRpcParams("ScribeV1", &history.FileScribe{})
	utils.RegisterRpcParams("PubSubV1", &engine.PubSub{})
	utils.RegisterRpcParams("AliasesV1", &engine.AliasHandler{})
	utils.RegisterRpcParams("UsersV1", &engine.UserMap{})
	utils.RegisterRpcParams("", &v1.CdrsV1{})
	utils.RegisterRpcParams("", &v2.CdrsV2{})
	utils.RegisterRpcParams("", &v1.SessionManagerV1{})
	utils.RegisterRpcParams("", &v1.SMGenericV1{})
	utils.RegisterRpcParams("", responder)
	utils.RegisterRpcParams("", apierRpcV1)
	utils.RegisterRpcParams("", apierRpcV2)
	utils.GetRpcParams("")
	internalRaterChan <- responder // Rater done
}
