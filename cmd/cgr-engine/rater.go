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
func startRater(internalRaterChan chan rpcclient.RpcClientConnection, cacheS *engine.CacheS,
	internalThdSChan, internalStatSChan chan rpcclient.RpcClientConnection,
	serviceManager *servmanager.ServiceManager, server *utils.Server,
	dm *engine.DataManager, loadDb engine.LoadStorage, cdrDb engine.CdrStorage, stopHandled *bool,
	exitChan chan bool, filterSChan chan *engine.FilterS, internalCacheSChan chan rpcclient.RpcClientConnection) {
	filterS := <-filterSChan
	filterSChan <- filterS
	var waitTasks []chan struct{}
	cacheTaskChan := make(chan struct{})
	waitTasks = append(waitTasks, cacheTaskChan)
	go func() { //Wait for cache load
		defer close(cacheTaskChan)
		<-cacheS.GetPrecacheChannel(utils.CacheDestinations)
		<-cacheS.GetPrecacheChannel(utils.CacheReverseDestinations)
		<-cacheS.GetPrecacheChannel(utils.CacheRatingPlans)
		<-cacheS.GetPrecacheChannel(utils.CacheRatingProfiles)
		<-cacheS.GetPrecacheChannel(utils.CacheActions)
		<-cacheS.GetPrecacheChannel(utils.CacheActionPlans)
		<-cacheS.GetPrecacheChannel(utils.CacheAccountActionPlans)
		<-cacheS.GetPrecacheChannel(utils.CacheActionTriggers)
		<-cacheS.GetPrecacheChannel(utils.CacheSharedGroups)
	}()

	var thdS *rpcclient.RpcClientPool
	if len(cfg.RalsCfg().RALsThresholdSConns) != 0 { // Connections to ThresholdS
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
				cfg.RalsCfg().RALsThresholdSConns, internalThdSChan,
				cfg.GeneralCfg().InternalTtl, false)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to ThresholdS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	var stats *rpcclient.RpcClientPool
	if len(cfg.RalsCfg().RALsStatSConns) != 0 { // Connections to StatS
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
				cfg.RalsCfg().RALsStatSConns, internalStatSChan,
				cfg.GeneralCfg().InternalTtl, false)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to StatS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	//create cache connection
	var caches *rpcclient.RpcClientPool
	if len(cfg.ApierCfg().CachesConns) != 0 {
		cachesTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, cachesTaskChan)
		go func() {
			defer close(cachesTaskChan)
			var err error
			caches, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.TlsCfg().ClientKey,
				cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
				cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
				cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
				cfg.ApierCfg().CachesConns, internalCacheSChan,
				cfg.GeneralCfg().InternalTtl, false)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<APIer> Could not connect to CacheS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}
	//add verification here
	if caches != nil && reflect.ValueOf(caches).IsNil() {
		caches = nil
	}

	// Wait for all connections to complete before going further
	for _, chn := range waitTasks {
		<-chn
	}
	responder := &engine.Responder{
		ExitChan:         exitChan,
		MaxComputedUsage: cfg.RalsCfg().RALsMaxComputedUsage}
	apierRpcV1 := &v1.ApierV1{
		StorDb:      loadDb,
		DataManager: dm,
		CdrDb:       cdrDb,
		Config:      cfg,
		Responder:   responder,
		ServManager: serviceManager,
		HTTPPoster: engine.NewHTTPPoster(cfg.GeneralCfg().HttpSkipTlsVerify,
			cfg.GeneralCfg().ReplyTimeout),
		FilterS: filterS,
		CacheS:  caches}
	if thdS != nil {
		engine.SetThresholdS(thdS) // temporary architectural fix until we will have separate AccountS
	}
	if stats != nil {
		engine.SetStatS(stats)
	}

	apierRpcV2 := &v2.ApierV2{
		ApierV1: *apierRpcV1}

	guardianSv1 := &v1.GuardianSv1{}

	server.RpcRegister(responder)
	server.RpcRegister(apierRpcV1)
	server.RpcRegister(apierRpcV2)
	server.RpcRegister(guardianSv1)

	utils.RegisterRpcParams("", &v1.CDRsV1{})
	utils.RegisterRpcParams("", &v2.CDRsV2{})
	utils.RegisterRpcParams("", &v1.SMGenericV1{})
	utils.RegisterRpcParams("", responder)
	utils.RegisterRpcParams("", apierRpcV1)
	utils.RegisterRpcParams("", apierRpcV2)
	utils.RegisterRpcParams("", guardianSv1)
	utils.GetRpcParams("")
	internalRaterChan <- responder // Rater done
}
