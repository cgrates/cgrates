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
	"reflect"

	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// Starts rater and reports on chan
func startRater(internalRaterChan, internalApierv1, internalApierv2, internalThdSChan, internalStatSChan,
	internalCacheSChan, internalSchedulerSChan, internalAttributeSChan, internalDispatcherSChan chan rpcclient.RpcClientConnection,
	serviceManager *servmanager.ServiceManager, server *utils.Server,
	dm *engine.DataManager, loadDb engine.LoadStorage, cdrDb engine.CdrStorage,
	chS *engine.CacheS, // separate from channel for optimization
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	var waitTasks []chan struct{}
	cacheTaskChan := make(chan struct{})
	waitTasks = append(waitTasks, cacheTaskChan)
	go func() { //Wait for cache load
		defer close(cacheTaskChan)
		<-chS.GetPrecacheChannel(utils.CacheDestinations)
		<-chS.GetPrecacheChannel(utils.CacheReverseDestinations)
		<-chS.GetPrecacheChannel(utils.CacheRatingPlans)
		<-chS.GetPrecacheChannel(utils.CacheRatingProfiles)
		<-chS.GetPrecacheChannel(utils.CacheActions)
		<-chS.GetPrecacheChannel(utils.CacheActionPlans)
		<-chS.GetPrecacheChannel(utils.CacheAccountActionPlans)
		<-chS.GetPrecacheChannel(utils.CacheActionTriggers)
		<-chS.GetPrecacheChannel(utils.CacheSharedGroups)
		<-chS.GetPrecacheChannel(utils.CacheTimings)
	}()

	var dispatcherConn rpcclient.RpcClientConnection
	isDispatcherEnabled := cfg.DispatcherSCfg().Enabled
	if isDispatcherEnabled {
		dispatcherConn = <-internalDispatcherSChan
		internalDispatcherSChan <- dispatcherConn
	}

	var thdS rpcclient.RpcClientConnection
	if isDispatcherEnabled {
		thdS = dispatcherConn
	} else if len(cfg.RalsCfg().RALsThresholdSConns) != 0 { // Connections to ThresholdS
		thdsTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, thdsTaskChan)
		go func() {
			defer close(thdsTaskChan)
			var err error
			thdS, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.TlsCfg().ClientKey,
				cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
				cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
				cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
				cfg.RalsCfg().RALsThresholdSConns, internalThdSChan, false)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to ThresholdS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	var stats rpcclient.RpcClientConnection
	if isDispatcherEnabled {
		stats = dispatcherConn
	} else if len(cfg.RalsCfg().RALsStatSConns) != 0 { // Connections to StatS
		statsTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, statsTaskChan)
		go func() {
			defer close(statsTaskChan)
			var err error
			stats, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.TlsCfg().ClientKey,
				cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
				cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
				cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
				cfg.RalsCfg().RALsStatSConns, internalStatSChan, false)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to StatS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	//create cache connection
	var cacheSrpc rpcclient.RpcClientConnection
	if isDispatcherEnabled {
		cacheSrpc = dispatcherConn
	} else if len(cfg.ApierCfg().CachesConns) != 0 {
		cachesTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, cachesTaskChan)
		go func() {
			defer close(cachesTaskChan)
			var err error
			cacheSrpc, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.TlsCfg().ClientKey,
				cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
				cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
				cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
				cfg.ApierCfg().CachesConns, internalCacheSChan, false)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<APIer> Could not connect to CacheS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	//create scheduler connection
	var schedulerSrpc rpcclient.RpcClientConnection
	if isDispatcherEnabled {
		schedulerSrpc = dispatcherConn
	} else if len(cfg.ApierCfg().SchedulerConns) != 0 {
		schedulerSTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, schedulerSTaskChan)
		go func() {
			defer close(schedulerSTaskChan)
			var err error
			schedulerSrpc, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.TlsCfg().ClientKey,
				cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
				cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
				cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
				cfg.ApierCfg().SchedulerConns, internalSchedulerSChan, false)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<APIer> Could not connect to SchedulerS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	//create scheduler connection
	var attributeSrpc rpcclient.RpcClientConnection
	if isDispatcherEnabled {
		attributeSrpc = dispatcherConn
	} else if len(cfg.ApierCfg().SchedulerConns) != 0 {
		attributeSTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, attributeSTaskChan)
		go func() {
			defer close(attributeSTaskChan)
			var err error
			attributeSrpc, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.TlsCfg().ClientKey,
				cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
				cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
				cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
				cfg.ApierCfg().AttributeSConns, internalAttributeSChan, false)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<APIer> Could not connect to AttributeS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	// Wait for all connections to complete before going further
	for _, chn := range waitTasks {
		<-chn
	}

	responder := &engine.Responder{
		ExitChan:         exitChan,
		MaxComputedUsage: cfg.RalsCfg().RALsMaxComputedUsage}

	// correct reflect on cacheS since there is no APIer init
	if cacheSrpc != nil && reflect.ValueOf(cacheSrpc).IsNil() {
		cacheSrpc = nil
	}
	// correct reflect on schedulerS since there is no APIer init
	if schedulerSrpc != nil && reflect.ValueOf(schedulerSrpc).IsNil() {
		schedulerSrpc = nil
	}
	// correct reflect on schedulerS since there is no APIer init
	if attributeSrpc != nil && reflect.ValueOf(attributeSrpc).IsNil() {
		attributeSrpc = nil
	}
	apierRpcV1 := &v1.ApierV1{
		StorDb:      loadDb,
		DataManager: dm,
		CdrDb:       cdrDb,
		Config:      cfg,
		Responder:   responder,
		ServManager: serviceManager,
		HTTPPoster: engine.NewHTTPPoster(cfg.GeneralCfg().HttpSkipTlsVerify,
			cfg.GeneralCfg().ReplyTimeout),
		FilterS:    filterS,
		CacheS:     cacheSrpc,
		SchedulerS: schedulerSrpc,
		AttributeS: attributeSrpc}

	if thdS != nil {
		engine.SetThresholdS(thdS) // temporary architectural fix until we will have separate AccountS
	}
	if stats != nil {
		engine.SetStatS(stats)
	}

	apierRpcV2 := &v2.ApierV2{
		ApierV1: *apierRpcV1}

	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(responder)
		server.RpcRegister(apierRpcV1)
		server.RpcRegister(apierRpcV2)
	}

	utils.RegisterRpcParams("", &v1.CDRsV1{})
	utils.RegisterRpcParams("", &v2.CDRsV2{})
	utils.RegisterRpcParams("", &v1.SMGenericV1{})
	utils.RegisterRpcParams("", responder)
	utils.RegisterRpcParams("", apierRpcV1)
	utils.RegisterRpcParams("", apierRpcV2)
	utils.GetRpcParams("")

	internalApierv1 <- apierRpcV1
	internalApierv2 <- apierRpcV2
	internalRaterChan <- responder // Rater done
}
