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
	internalThdSChan, internalStatSChan, internalPubSubSChan,
	internalUserSChan, internalAliaseSChan chan rpcclient.RpcClientConnection,
	serviceManager *servmanager.ServiceManager, server *utils.Server,
	dm *engine.DataManager, loadDb engine.LoadStorage, cdrDb engine.CdrStorage, stopHandled *bool,
	exitChan chan bool, filterSChan chan *engine.FilterS) {
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
		<-cacheS.GetPrecacheChannel(utils.CacheDerivedChargers)
		<-cacheS.GetPrecacheChannel(utils.CacheAliases)
		<-cacheS.GetPrecacheChannel(utils.CacheReverseAliases)
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
				cfg.GeneralCfg().InternalTtl)
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
				cfg.GeneralCfg().InternalTtl)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to StatS, error: %s", err.Error()))
				exitChan <- true
				return
			}
		}()
	}

	if len(cfg.RalsCfg().RALsPubSubSConns) != 0 { // Connection to pubsubs
		pubsubTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, pubsubTaskChan)
		go func() {
			defer close(pubsubTaskChan)
			if pubSubSConns, err := engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.TlsCfg().ClientKey,
				cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
				cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
				cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
				cfg.RalsCfg().RALsPubSubSConns, internalPubSubSChan,
				cfg.GeneralCfg().InternalTtl); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to PubSubS: %s", err.Error()))
				exitChan <- true
				return
			} else {
				engine.SetPubSub(pubSubSConns)
			}
		}()
	}

	if len(cfg.RalsCfg().RALsAliasSConns) != 0 { // Connection to AliasService
		aliasesTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, aliasesTaskChan)
		go func() {
			defer close(aliasesTaskChan)
			if aliaseSCons, err := engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.TlsCfg().ClientKey,
				cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
				cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
				cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
				cfg.RalsCfg().RALsAliasSConns, internalAliaseSChan,
				cfg.GeneralCfg().InternalTtl); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to AliaseS, error: %s", err.Error()))
				exitChan <- true
				return
			} else {
				engine.SetAliasService(aliaseSCons)
			}
		}()
	}

	var usersConns rpcclient.RpcClientConnection
	if len(cfg.RalsCfg().RALsUserSConns) != 0 { // Connection to UserService
		usersTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, usersTaskChan)
		go func() {
			defer close(usersTaskChan)
			var err error
			if usersConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.TlsCfg().ClientKey,
				cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
				cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
				cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
				cfg.RalsCfg().RALsUserSConns, internalUserSChan,
				cfg.GeneralCfg().InternalTtl); err != nil {
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
	responder := &engine.Responder{
		ExitChan:         exitChan,
		MaxComputedUsage: cfg.RalsCfg().RALsMaxComputedUsage}
	responder.SetTimeToLive(cfg.GeneralCfg().ResponseCacheTTL, nil)
	apierRpcV1 := &v1.ApierV1{
		StorDb:      loadDb,
		DataManager: dm,
		CdrDb:       cdrDb,
		Config:      cfg,
		Responder:   responder,
		ServManager: serviceManager,
		HTTPPoster: engine.NewHTTPPoster(cfg.GeneralCfg().HttpSkipTlsVerify,
			cfg.GeneralCfg().ReplyTimeout),
		FilterS: filterS}
	if thdS != nil {
		engine.SetThresholdS(thdS) // temporary architectural fix until we will have separate AccountS
	}
	if stats != nil {
		engine.SetStatS(stats)
	}
	if usersConns != nil {
		apierRpcV1.Users = usersConns
	}
	apierRpcV2 := &v2.ApierV2{
		ApierV1: *apierRpcV1}

	server.RpcRegister(responder)
	server.RpcRegister(apierRpcV1)
	server.RpcRegister(apierRpcV2)

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
