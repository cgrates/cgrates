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

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// Starts rater and reports on chan
func startRater(internalRaterChan chan rpcclient.RpcClientConnection, cacheS *engine.CacheS,
	internalThdSChan, internalCdrStatSChan, internalStatSChan, internalPubSubSChan,
	internalAttributeSChan, internalUserSChan, internalAliaseSChan chan rpcclient.RpcClientConnection,
	serviceManager *servmanager.ServiceManager, server *utils.Server,
	dm *engine.DataManager, loadDb engine.LoadStorage, cdrDb engine.CdrStorage, stopHandled *bool, exitChan chan bool) {
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
		<-cacheS.GetPrecacheChannel(utils.CacheActionTriggers)
		<-cacheS.GetPrecacheChannel(utils.CacheSharedGroups)
		<-cacheS.GetPrecacheChannel(utils.CacheLCRRules)
		<-cacheS.GetPrecacheChannel(utils.CacheDerivedChargers)
		<-cacheS.GetPrecacheChannel(utils.CacheAliases)
		<-cacheS.GetPrecacheChannel(utils.CacheReverseAliases)
	}()

	var thdS *rpcclient.RpcClientPool
	if len(cfg.RALsThresholdSConns) != 0 { // Connections to ThresholdS
		thdsTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, thdsTaskChan)
		go func() {
			defer close(thdsTaskChan)
			var err error
			thdS, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsThresholdSConns, internalThdSChan, cfg.InternalTtl)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to ThresholdS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	var cdrStats *rpcclient.RpcClientPool
	if len(cfg.RALsCDRStatSConns) != 0 { // Connections to CDRStats
		cdrstatTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, cdrstatTaskChan)
		go func() {
			defer close(cdrstatTaskChan)
			var err error
			cdrStats, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsCDRStatSConns, internalCdrStatSChan, cfg.InternalTtl)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to CDRStatS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	var stats *rpcclient.RpcClientPool
	if len(cfg.RALsStatSConns) != 0 { // Connections to CDRStats
		statsTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, statsTaskChan)
		go func() {
			defer close(statsTaskChan)
			var err error
			stats, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsStatSConns, internalStatSChan, cfg.InternalTtl)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to StatS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	if len(cfg.RALsPubSubSConns) != 0 { // Connection to pubsubs
		pubsubTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, pubsubTaskChan)
		go func() {
			defer close(pubsubTaskChan)
			if pubSubSConns, err := engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsPubSubSConns, internalPubSubSChan, cfg.InternalTtl); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to PubSubS: %s", err.Error()))
				exitChan <- true
				return
			} else {
				engine.SetPubSub(pubSubSConns)
			}
		}()
	}

	var attrS *rpcclient.RpcClientPool
	if len(cfg.RALsAttributeSConns) != 0 { // Connections to AttributeS
		attrsTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, attrsTaskChan)
		go func() {
			defer close(attrsTaskChan)
			var err error
			attrS, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts,
				cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsAttributeSConns, internalAttributeSChan, cfg.InternalTtl)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to %s, error: %s",
					utils.AttributeS, err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	if len(cfg.RALsAliasSConns) != 0 { // Connection to AliasService
		aliasesTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, aliasesTaskChan)
		go func() {
			defer close(aliasesTaskChan)
			if aliaseSCons, err := engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
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
			var err error
			if usersConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsUserSConns, internalUserSChan, cfg.InternalTtl); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect UserS, error: %s", err.Error()))
				exitChan <- true
				return
			} else {
				engine.SetUserService(usersConns)
			}
		}()
	}
	// Wait for all connections to complete before going further
	for _, chn := range waitTasks {
		<-chn
	}
	responder := &engine.Responder{ExitChan: exitChan, MaxComputedUsage: cfg.RALsMaxComputedUsage}
	responder.SetTimeToLive(cfg.ResponseCacheTTL, nil)
	apierRpcV1 := &v1.ApierV1{StorDb: loadDb, DataManager: dm, CdrDb: cdrDb,
		Config: cfg, Responder: responder, ServManager: serviceManager,
		HTTPPoster: utils.NewHTTPPoster(cfg.HttpSkipTlsVerify, cfg.ReplyTimeout)}
	if thdS != nil {
		engine.SetThresholdS(thdS) // temporary architectural fix until we will have separate AccountS
	}
	if attrS != nil {
		responder.AttributeS = attrS
	}
	if cdrStats != nil { // ToDo: Fix here properly the init of stats
		responder.CdrStats = cdrStats
		apierRpcV1.CdrStatsSrv = cdrStats
	}
	if usersConns != nil {
		apierRpcV1.Users = usersConns
	}
	apierRpcV2 := &v2.ApierV2{
		ApierV1: *apierRpcV1}

	server.RpcRegister(responder)
	server.RpcRegister(apierRpcV1)
	server.RpcRegister(apierRpcV2)

	utils.RegisterRpcParams("", &engine.Stats{})
	utils.RegisterRpcParams("", &v1.CDRStatsV1{})
	utils.RegisterRpcParams("PubSubV1", &engine.PubSub{})
	utils.RegisterRpcParams("AliasesV1", &engine.AliasHandler{})
	utils.RegisterRpcParams("UsersV1", &engine.UserMap{})
	utils.RegisterRpcParams("", &v1.CdrsV1{})
	utils.RegisterRpcParams("", &v2.CdrsV2{})
	utils.RegisterRpcParams("", &v1.SMGenericV1{})
	utils.RegisterRpcParams("", responder)
	utils.RegisterRpcParams("", apierRpcV1)
	utils.RegisterRpcParams("", apierRpcV2)
	utils.GetRpcParams("")
	internalRaterChan <- responder // Rater done
}
