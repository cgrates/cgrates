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
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// Starts rater and reports on chan
func startRater(internalRaterChan chan rpcclient.RpcClientConnection, cacheDoneChan chan struct{},
	internalThdSChan, internalCdrStatSChan, internalStatSChan, internalHistorySChan,
	internalPubSubSChan, internalUserSChan, internalAliaseSChan chan rpcclient.RpcClientConnection,
	serviceManager *servmanager.ServiceManager, server *utils.Server,
	dm *engine.DataManager, loadDb engine.LoadStorage, cdrDb engine.CdrStorage, stopHandled *bool, exitChan chan bool) {
	var waitTasks []chan struct{}

	//Cache load
	cacheTaskChan := make(chan struct{})
	waitTasks = append(waitTasks, cacheTaskChan)
	go func() {
		defer close(cacheTaskChan)
		var dstIDs, rvDstIDs, rplIDs, rpfIDs, actIDs, aplIDs, aapIDs, atrgIDs, sgIDs, lcrIDs, dcIDs, alsIDs, rvAlsIDs, rspIDs, resIDs, stqIDs, stqpIDs, thIDs, thpIDs, fltrIDs []string
		if cCfg, has := cfg.CacheConfig[utils.CacheDestinations]; !has || !cCfg.Precache {
			dstIDs = make([]string, 0) // Don't cache any
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheReverseDestinations]; !has || !cCfg.Precache {
			rvDstIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheRatingPlans]; !has || !cCfg.Precache {
			rplIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheRatingProfiles]; !has || !cCfg.Precache {
			rpfIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheActions]; !has || !cCfg.Precache {
			actIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheActionPlans]; !has || !cCfg.Precache {
			aplIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheAccountActionPlans]; !has || !cCfg.Precache {
			aapIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheActionTriggers]; !has || !cCfg.Precache {
			atrgIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheSharedGroups]; !has || !cCfg.Precache {
			sgIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheLCRRules]; !has || !cCfg.Precache {
			lcrIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheDerivedChargers]; !has || !cCfg.Precache {
			dcIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheAliases]; !has || !cCfg.Precache {
			alsIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheReverseAliases]; !has || !cCfg.Precache {
			rvAlsIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheResourceProfiles]; !has || !cCfg.Precache {
			rspIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheResources]; !has || !cCfg.Precache {
			resIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheStatQueues]; !has || !cCfg.Precache {
			stqIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheStatQueueProfiles]; !has || !cCfg.Precache {
			stqpIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheThresholds]; !has || !cCfg.Precache {
			thIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheThresholdProfiles]; !has || !cCfg.Precache {
			thpIDs = make([]string, 0)
		}
		if cCfg, has := cfg.CacheConfig[utils.CacheFilters]; !has || !cCfg.Precache {
			fltrIDs = make([]string, 0)
		}

		// ToDo: Add here timings
		if err := dm.LoadDataDBCache(dstIDs, rvDstIDs, rplIDs, rpfIDs, actIDs, aplIDs, aapIDs, atrgIDs, sgIDs, lcrIDs, dcIDs, alsIDs, rvAlsIDs, rspIDs, resIDs, stqIDs, stqpIDs, thIDs, thpIDs, fltrIDs); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<RALs> Cache rating error: %s", err.Error()))
			exitChan <- true
			return
		}
		cacheDoneChan <- struct{}{}
	}()

	var thdS *rpcclient.RpcClientPool
	if len(cfg.RALsThresholdSConns) != 0 { // Connections to ThresholdS
		thdsTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, thdsTaskChan)
		go func() {
			defer close(thdsTaskChan)
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
			stats, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.RALsStatSConns, internalStatSChan, cfg.InternalTtl)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<RALs> Could not connect to StatS, error: %s", err.Error()))
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
	responder := &engine.Responder{ExitChan: exitChan}
	responder.SetTimeToLive(cfg.ResponseCacheTTL, nil)
	apierRpcV1 := &v1.ApierV1{StorDb: loadDb, DataManager: dm, CdrDb: cdrDb,
		Config: cfg, Responder: responder, ServManager: serviceManager,
		HTTPPoster: utils.NewHTTPPoster(cfg.HttpSkipTlsVerify, cfg.ReplyTimeout)}
	if thdS != nil {
		engine.SetThresholdS(thdS) // temporary architectural fix until we will have separate AccountS
	}
	if cdrStats != nil { // ToDo: Fix here properly the init of stats
		responder.Stats = cdrStats
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
