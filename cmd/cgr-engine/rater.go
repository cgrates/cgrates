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
	"sync"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/balancer2go"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
)

func startBalancer(internalBalancerChan chan *balancer2go.Balancer, stopHandled *bool, exitChan chan bool) {
	bal := balancer2go.NewBalancer()
	go stopBalancerSignalHandler(bal, exitChan)
	*stopHandled = true
	internalBalancerChan <- bal
}

func cacheRaterData(doneChan chan struct{}, ratingDb engine.RatingStorage, accountDb engine.AccountingStorage, exitChan chan bool) {
	if err := ratingDb.CacheRatingAll(); err != nil {
		engine.Logger.Crit(fmt.Sprintf("Cache rating error: %s", err.Error()))
		exitChan <- true
		return
	}
	if err := accountDb.CacheAccountingAll(); err != nil {
		engine.Logger.Crit(fmt.Sprintf("Cache accounting error: %s", err.Error()))
		exitChan <- true
		return
	}
	close(doneChan)
}

// Starts rater and reports on chan
func startRater(internalRaterChan chan *engine.Responder, internalBalancerChan chan *balancer2go.Balancer, internalSchedulerChan chan *scheduler.Scheduler,
	internalCdrStatSChan chan engine.StatsInterface, internalHistorySChan chan history.Scribe,
	internalPubSubSChan chan engine.PublisherSubscriber, internalUserSChan chan engine.UserService, internalAliaseSChan chan engine.AliasService,
	cacheChan chan struct{}, server *engine.Server,
	ratingDb engine.RatingStorage, accountDb engine.AccountingStorage, loadDb engine.LoadStorage, cdrDb engine.CdrStorage, logDb engine.LogStorage,
	stopHandled *bool, exitChan chan bool) {
	var wg sync.WaitGroup // Sync all external connections in a group

	var sched *scheduler.Scheduler // Need the scheduler in APIer
	if cfg.SchedulerEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sched = <-internalSchedulerChan
			internalSchedulerChan <- sched
		}()
	}

	// Connection to balancer
	var bal *balancer2go.Balancer
	if cfg.RaterBalancer != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if cfg.RaterBalancer == utils.INTERNAL {
				bal = <-internalBalancerChan
				internalBalancerChan <- bal // Put it back if someone else is interested about
			} else {
				go registerToBalancer(exitChan)
				go stopRaterSignalHandler(internalCdrStatSChan, exitChan)
				*stopHandled = true
			}
		}()
	}

	// Connection to CDRStats
	var cdrStats engine.StatsInterface
	if cfg.RaterCdrStats != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if cfg.RaterCdrStats == utils.INTERNAL {
				cdrStats = <-internalCdrStatSChan
				internalCdrStatSChan <- cdrStats
			} else if cdrStats, err = engine.NewProxyStats(cfg.RaterCdrStats, cfg.ConnectAttempts, -1); err != nil {
				engine.Logger.Crit(fmt.Sprintf("<CdrStats> Could not connect to the server, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	// Connection to HistoryS
	if cfg.RaterHistoryServer != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var scribeServer history.Scribe
			if cfg.RaterHistoryServer == utils.INTERNAL {
				scribeServer = <-internalHistorySChan
				internalHistorySChan <- scribeServer
			} else if scribeServer, err = history.NewProxyScribe(cfg.RaterHistoryServer, cfg.ConnectAttempts, -1); err != nil {
				engine.Logger.Crit(fmt.Sprintf("<HistoryServer> Could not connect to the server, error: %s", err.Error()))
				exitChan <- true
				return
			}
			engine.SetHistoryScribe(scribeServer) // ToDo: replace package sharing with connection based one
		}()
	}

	// Connection to pubsubs
	if cfg.RaterPubSubServer != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var pubSubServer engine.PublisherSubscriber
			if cfg.RaterPubSubServer == utils.INTERNAL {
				pubSubServer = <-internalPubSubSChan
				internalPubSubSChan <- pubSubServer
			} else if pubSubServer, err = engine.NewProxyPubSub(cfg.RaterPubSubServer, cfg.ConnectAttempts, -1); err != nil {
				engine.Logger.Crit(fmt.Sprintf("<PubSubServer> Could not connect to the server, error: %s", err.Error()))
				exitChan <- true
				return
			}
			engine.SetPubSub(pubSubServer) // ToDo: replace package sharing with connection based one
		}()
	}

	// Connection to AliasService
	if cfg.RaterAliasesServer != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var aliasesServer engine.AliasService
			if cfg.RaterAliasesServer == utils.INTERNAL {
				aliasesServer = <-internalAliaseSChan
				internalAliaseSChan <- aliasesServer
			} else if aliasesServer, err = engine.NewProxyAliasService(cfg.RaterAliasesServer, cfg.ConnectAttempts, -1); err != nil {
				engine.Logger.Crit(fmt.Sprintf("<AliasesServer> Could not connect to the server, error: %s", err.Error()))
				exitChan <- true
				return
			}
			engine.SetAliasService(aliasesServer) // ToDo: replace package sharing with connection based one
		}()
	}

	// Connection to UserService
	var userServer engine.UserService
	if cfg.RaterUserServer != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if cfg.RaterUserServer == utils.INTERNAL {
				userServer = <-internalUserSChan
				internalUserSChan <- userServer
			} else if userServer, err = engine.NewProxyUserService(cfg.RaterUserServer, cfg.ConnectAttempts, -1); err != nil {
				engine.Logger.Crit(fmt.Sprintf("<UserServer> Could not connect to the server, error: %s", err.Error()))
				exitChan <- true
				return
			}
			engine.SetUserService(userServer)
		}()
	}

	// Wait for all connections to complete before going further
	wg.Wait()

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
