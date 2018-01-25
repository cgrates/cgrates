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
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/cdrc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

const (
	JSON     = "json"
	GOB      = "gob"
	POSTGRES = "postgres"
	MYSQL    = "mysql"
	MONGO    = "mongo"
	REDIS    = "redis"
	SAME     = "same"
	FS       = "freeswitch"
	KAMAILIO = "kamailio"
	OSIPS    = "opensips"
)

var (
	cfgDir            = flag.String("config_dir", utils.CONFIG_DIR, "Configuration directory path.")
	version           = flag.Bool("version", false, "Prints the application version.")
	pidFile           = flag.String("pid", "", "Write pid file")
	cpuprofile        = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile        = flag.String("memprofile", "", "write memory profile to file")
	scheduledShutdown = flag.String("scheduled_shutdown", "", "shutdown the engine after this duration")
	singlecpu         = flag.Bool("singlecpu", false, "Run on single CPU core")
	syslogger         = flag.String("logger", "", "logger <*syslog|*stdout>")
	nodeID            = flag.String("node_id", "", "The node ID of the engine")
	logLevel          = flag.Int("log_level", -1, "Log level (0-emergency to 7-debug)")

	cfg *config.CGRConfig
)

func startCdrcs(internalCdrSChan, internalRaterChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	cdrcInitialized := false           // Control whether the cdrc was already initialized (so we don't reload in that case)
	var cdrcChildrenChan chan struct{} // Will use it to communicate with the children of one fork
	for {
		select {
		case <-exitChan: // Stop forking CDRCs
			break
		case <-cfg.ConfigReloads[utils.CDRC]: // Consume the load request and wait for a new one
			if cdrcInitialized {
				utils.Logger.Info("<CDRC> Configuration reload")
				close(cdrcChildrenChan) // Stop all the children of the previous run
			}
			cdrcChildrenChan = make(chan struct{})
		}
		// Start CDRCs
		for _, cdrcCfgs := range cfg.CdrcProfiles {
			var enabledCfgs []*config.CdrcConfig
			for _, cdrcCfg := range cdrcCfgs { // Take a random config out since they should be the same
				if cdrcCfg.Enabled {
					enabledCfgs = append(enabledCfgs, cdrcCfg)
				}
			}

			if len(enabledCfgs) != 0 {
				go startCdrc(internalCdrSChan, internalRaterChan, enabledCfgs, cfg.HttpSkipTlsVerify, cdrcChildrenChan, exitChan)
			} else {
				utils.Logger.Info("<CDRC> No enabled CDRC clients")
			}
		}
		cdrcInitialized = true // Initialized
	}
}

// Fires up a cdrc instance
func startCdrc(internalCdrSChan, internalRaterChan chan rpcclient.RpcClientConnection, cdrcCfgs []*config.CdrcConfig, httpSkipTlsCheck bool,
	closeChan chan struct{}, exitChan chan bool) {
	var cdrcCfg *config.CdrcConfig
	for _, cdrcCfg = range cdrcCfgs { // Take the first config out, does not matter which one
		break
	}
	cdrsConn, err := engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
		cdrcCfg.CdrsConns, internalCdrSChan, cfg.InternalTtl)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRC> Could not connect to CDRS via RPC: %s", err.Error()))
		exitChan <- true
		return
	}
	cdrc, err := cdrc.NewCdrc(cdrcCfgs, httpSkipTlsCheck, cdrsConn, closeChan, cfg.DefaultTimezone, cfg.RoundingDecimals)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("Cdrc config parsing error: %s", err.Error()))
		exitChan <- true
		return
	}
	if err := cdrc.Run(); err != nil {
		utils.Logger.Crit(fmt.Sprintf("Cdrc run error: %s", err.Error()))
		exitChan <- true // If run stopped, something is bad, stop the application
		return
	}
}

func startSessionS(internalSMGChan, internalRaterChan, internalResourceSChan, internalSupplierSChan,
	internalAttrSChan, internalCDRSChan chan rpcclient.RpcClientConnection, server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS Session service.")
	var err error
	var ralsConns, resSConns, suplSConns, attrSConns, cdrsConn *rpcclient.RpcClientPool
	if len(cfg.SessionSCfg().RALsConns) != 0 {
		ralsConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SessionSCfg().RALsConns, internalRaterChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to RALs: %s", utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SessionSCfg().ResSConns) != 0 {
		resSConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SessionSCfg().ResSConns, internalResourceSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to ResourceS: %s", utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SessionSCfg().SupplSConns) != 0 {
		suplSConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SessionSCfg().SupplSConns, internalSupplierSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to SupplierS: %s", utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SessionSCfg().AttrSConns) != 0 {
		attrSConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SessionSCfg().AttrSConns, internalAttrSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to AttributeS: %s", utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SessionSCfg().CDRsConns) != 0 {
		cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SessionSCfg().CDRsConns, internalCDRSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to RALs: %s", utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	smgReplConns, err := sessionmanager.NewSessionReplicationConns(cfg.SessionSCfg().SessionReplicationConns, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to SMGReplicationConnection error: <%s>", utils.SessionS, err.Error()))
		exitChan <- true
		return
	}
	sm := sessionmanager.NewSMGeneric(cfg, ralsConns, resSConns, suplSConns,
		attrSConns, cdrsConn, smgReplConns, cfg.DefaultTimezone)
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.SessionS, err))
	}
	// Pass internal connection via BiRPCClient
	internalSMGChan <- sm
	// Register RPC handler
	smgRpc := v1.NewSMGenericV1(sm)
	server.RpcRegister(smgRpc)
	server.RpcRegister(&v2.SMGenericV2{*smgRpc})
	ssv1 := v1.NewSessionSv1(sm) // methods with multiple options
	server.RpcRegister(ssv1)
	// Register BiRpc handlers
	if cfg.SessionSCfg().ListenBijson != "" {
		smgBiRpc := v1.NewSMGenericBiRpcV1(sm)
		for method, handler := range smgBiRpc.Handlers() {
			server.BiRPCRegisterName(method, handler)
		}
		for method, handler := range ssv1.Handlers() {
			server.BiRPCRegisterName(method, handler)
		}
		server.ServeBiJSON(cfg.SessionSCfg().ListenBijson)
		exitChan <- true
	}
}

func startAsteriskAgent(internalSMGChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	utils.Logger.Info("Starting Asterisk agent")
	smgRpcConn := <-internalSMGChan
	internalSMGChan <- smgRpcConn
	birpcClnt := utils.NewBiRPCInternalClient(smgRpcConn.(*sessionmanager.SMGeneric))
	for connIdx := range cfg.AsteriskAgentCfg().AsteriskConns { // Instantiate connections towards asterisk servers
		sma, err := agents.NewSMAsterisk(cfg, connIdx, birpcClnt)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMAsterisk> error: %s!", err))
			exitChan <- true
			return
		}
		if err = sma.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMAsterisk> runtime error: %s!", err))
		}
	}
	exitChan <- true
}

func startDiameterAgent(internalSMGChan, internalPubSubSChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	var err error
	utils.Logger.Info("Starting CGRateS DiameterAgent service")
	var smgConn, pubsubConn *rpcclient.RpcClientPool
	if len(cfg.DiameterAgentCfg().SessionSConns) != 0 {
		smgConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.DiameterAgentCfg().SessionSConns, internalSMGChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<DiameterAgent> Could not connect to SMG: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.DiameterAgentCfg().PubSubConns) != 0 {
		pubsubConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.DiameterAgentCfg().PubSubConns, internalPubSubSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<DiameterAgent> Could not connect to PubSubS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	da, err := agents.NewDiameterAgent(cfg, smgConn, pubsubConn)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> error: %s!", err))
		exitChan <- true
		return
	}
	if err = da.ListenAndServe(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> error: %s!", err))
	}
	exitChan <- true
}

func startRadiusAgent(internalSMGChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	var err error
	utils.Logger.Info("Starting CGRateS RadiusAgent service")
	var smgConn *rpcclient.RpcClientPool
	if len(cfg.RadiusAgentCfg().SessionSConns) != 0 {
		smgConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts,
			cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.RadiusAgentCfg().SessionSConns, internalSMGChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<RadiusAgent> Could not connect to SMG: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	ra, err := agents.NewRadiusAgent(cfg, smgConn)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<RadiusAgent> error: <%s>", err.Error()))
		exitChan <- true
		return
	}
	if err = ra.ListenAndServe(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<RadiusAgent> error: <%s>", err.Error()))
	}
	exitChan <- true
}

func startFsAgent(internalSMGChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	var err error
	utils.Logger.Info("Starting FreeSWITCH agent")
	smgRpcConn := <-internalSMGChan
	internalSMGChan <- smgRpcConn
	birpcClnt := utils.NewBiRPCInternalClient(smgRpcConn.(*sessionmanager.SMGeneric))
	sm := agents.NewFSSessionManager(cfg.FsAgentCfg(), birpcClnt, cfg.DefaultTimezone)
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
	}
	exitChan <- true
}

func startSmKamailio(internalRaterChan, internalCDRSChan, internalRsChan chan rpcclient.RpcClientConnection, cdrDb engine.CdrStorage, exitChan chan bool) {
	var err error
	utils.Logger.Info("Starting CGRateS SMKamailio service.")
	var ralsConn, cdrsConn, rlSConn *rpcclient.RpcClientPool
	if len(cfg.SmKamConfig.RALsConns) != 0 {
		ralsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SmKamConfig.RALsConns, internalRaterChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<SMKamailio> Could not connect to RAL: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SmKamConfig.CDRsConns) != 0 {
		cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SmKamConfig.CDRsConns, internalCDRSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<SMKamailio> Could not connect to CDRs: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SmKamConfig.RLsConns) != 0 {
		rlSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SmKamConfig.RLsConns, internalRsChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<SMKamailio> Could not connect to RLsConns: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	sm, _ := sessionmanager.NewKamailioSessionManager(cfg.SmKamConfig, ralsConn, cdrsConn, rlSConn, cfg.DefaultTimezone)
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMKamailio> error: %s!", err))
	}
	exitChan <- true
}

func startCDRS(internalCdrSChan chan rpcclient.RpcClientConnection,
	cdrDb engine.CdrStorage, dm *engine.DataManager,
	internalRaterChan, internalPubSubSChan, internalAttributeSChan, internalUserSChan, internalAliaseSChan,
	internalCdrStatSChan, internalThresholdSChan, internalStatSChan chan rpcclient.RpcClientConnection,
	server *utils.Server, exitChan chan bool) {
	var err error
	utils.Logger.Info("Starting CGRateS CDRS service.")
	var ralConn, pubSubConn, usersConn, attrSConn, aliasesConn, cdrstatsConn, thresholdSConn, statsConn *rpcclient.RpcClientPool
	if len(cfg.CDRSRaterConns) != 0 { // Conn pool towards RAL
		ralConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.CDRSRaterConns, internalRaterChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to RAL: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CDRSPubSubSConns) != 0 { // Pubsub connection init
		pubSubConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.CDRSPubSubSConns, internalPubSubSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to PubSubSystem: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CDRSAttributeSConns) != 0 { // Users connection init
		attrSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.CDRSAttributeSConns, internalAttributeSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
				utils.AttributeS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CDRSUserSConns) != 0 { // Users connection init
		usersConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.CDRSUserSConns, internalUserSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to UserS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CDRSAliaseSConns) != 0 { // Aliases connection init
		aliasesConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.CDRSAliaseSConns, internalAliaseSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to AliaseS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CDRSCDRStatSConns) != 0 { // Stats connection init
		cdrstatsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.CDRSCDRStatSConns, internalCdrStatSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to CDRStatS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CDRSThresholdSConns) != 0 { // Stats connection init
		thresholdSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.CDRSThresholdSConns, internalThresholdSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to ThresholdS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CDRSStatSConns) != 0 { // Stats connection init
		statsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.CDRSStatSConns, internalStatSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to StatS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	cdrServer, _ := engine.NewCdrServer(cfg, cdrDb, dm, ralConn, pubSubConn,
		attrSConn, usersConn, aliasesConn, cdrstatsConn, thresholdSConn, statsConn)
	cdrServer.SetTimeToLive(cfg.ResponseCacheTTL, nil)
	utils.Logger.Info("Registering CDRS HTTP Handlers.")
	cdrServer.RegisterHandlersToServer(server)
	utils.Logger.Info("Registering CDRS RPC service.")
	cdrSrv := v1.CdrsV1{CdrSrv: cdrServer}
	server.RpcRegister(&cdrSrv)
	server.RpcRegister(&v2.CdrsV2{CdrsV1: cdrSrv})
	// Make the cdr server available for internal communication
	server.RpcRegister(cdrServer) // register CdrServer for internal usage (TODO: refactor this)
	internalCdrSChan <- cdrServer // Signal that cdrS is operational
}

func startScheduler(internalSchedulerChan chan *scheduler.Scheduler, cacheDoneChan chan struct{}, dm *engine.DataManager, exitChan chan bool) {
	// Wait for cache to load data before starting
	cacheDone := <-cacheDoneChan
	cacheDoneChan <- cacheDone
	utils.Logger.Info("Starting CGRateS Scheduler.")
	sched := scheduler.NewScheduler(dm)
	internalSchedulerChan <- sched

	sched.Loop()
	exitChan <- true // Should not get out of loop though
}

func startCdrStats(internalCdrStatSChan chan rpcclient.RpcClientConnection, dm *engine.DataManager, server *utils.Server) {
	cdrStats := engine.NewStats(dm, cfg.CDRStatsSaveInterval)
	server.RpcRegister(cdrStats)
	server.RpcRegister(&v1.CDRStatsV1{CdrStats: cdrStats}) // Public APIs
	internalCdrStatSChan <- cdrStats
}

func startPubSubServer(internalPubSubSChan chan rpcclient.RpcClientConnection, dm *engine.DataManager, server *utils.Server, exitChan chan bool) {
	pubSubServer, err := engine.NewPubSub(dm, cfg.HttpSkipTlsVerify)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<PubSubS> Could not start, error: %s", err.Error()))
		exitChan <- true
		return
	}
	server.RpcRegisterName("PubSubV1", pubSubServer)
	internalPubSubSChan <- pubSubServer
}

// ToDo: Make sure we are caching before starting this one
func startAliasesServer(internalAliaseSChan chan rpcclient.RpcClientConnection, dm *engine.DataManager, server *utils.Server, exitChan chan bool) {
	aliasesServer := engine.NewAliasHandler(dm)
	server.RpcRegisterName("AliasesV1", aliasesServer)
	loadHist, err := dm.DataDB().GetLoadHistory(1, true, utils.NonTransactional)
	if err != nil || len(loadHist) == 0 {
		utils.Logger.Info(fmt.Sprintf("could not get load history: %v (%v)", loadHist, err))
		internalAliaseSChan <- aliasesServer
		return
	}
	internalAliaseSChan <- aliasesServer
}

func startUsersServer(internalUserSChan chan rpcclient.RpcClientConnection, dm *engine.DataManager, server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting User service.")
	userServer, err := engine.NewUserMap(dm, cfg.UserServerIndexes)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<UsersService> Could not start, error: %s", err.Error()))
		exitChan <- true
		return
	}
	utils.Logger.Info("Started User service.")
	server.RpcRegisterName("UsersV1", userServer)
	internalUserSChan <- userServer
}

// startAttributeService fires up the AttributeS
func startAttributeService(internalAttributeSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dm *engine.DataManager, server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	filterS := <-filterSChan
	filterSChan <- filterS
	aS, err := engine.NewAttributeService(dm, filterS,
		cfg.AttributeSCfg().StringIndexedFields, cfg.AttributeSCfg().PrefixIndexedFields)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.AttributeS, err.Error()))
		exitChan <- true
		return
	}
	go func() {
		if err := aS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.AttributeS, err.Error()))
		}
		aS.Shutdown()
		exitChan <- true
		return
	}()
	aSv1 := v1.NewAttributeSv1(aS)
	server.RpcRegister(aSv1)
	internalAttributeSChan <- aSv1
}

func startResourceService(internalRsChan, internalThresholdSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dm *engine.DataManager, server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	var err error
	var thdSConn *rpcclient.RpcClientPool
	filterS := <-filterSChan
	filterSChan <- filterS
	if len(cfg.ResourceSCfg().ThresholdSConns) != 0 { // Stats connection init
		thdSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.ResourceSCfg().ThresholdSConns, internalThresholdSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<ResourceS> Could not connect to ThresholdS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	rS, err := engine.NewResourceService(dm, cfg.ResourceSCfg().StoreInterval,
		thdSConn, filterS, cfg.ResourceSCfg().StringIndexedFields, cfg.ResourceSCfg().PrefixIndexedFields)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<ResourceS> Could not init, error: %s", err.Error()))
		exitChan <- true
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting Resource Service"))
	go func() {
		if err := rS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<ResourceS> Could not start, error: %s", err.Error()))

		}
		rS.Shutdown()
		exitChan <- true
		return
	}()
	rsV1 := v1.NewResourceSv1(rS)
	server.RpcRegister(rsV1)
	internalRsChan <- rsV1
}

// startStatService fires up the StatS
func startStatService(internalStatSChan, internalThresholdSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dm *engine.DataManager, server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	var err error
	var thdSConn *rpcclient.RpcClientPool
	filterS := <-filterSChan
	filterSChan <- filterS
	if len(cfg.StatSCfg().ThresholdSConns) != 0 { // Stats connection init
		thdSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.StatSCfg().ThresholdSConns, internalThresholdSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<StatS> Could not connect to ThresholdS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	sS, err := engine.NewStatService(dm, cfg.StatSCfg().StoreInterval,
		thdSConn, filterS, cfg.StatSCfg().StringIndexedFields, cfg.StatSCfg().PrefixIndexedFields)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<StatS> Could not init, error: %s", err.Error()))
		exitChan <- true
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting Stat Service"))
	go func() {
		if err := sS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<StatS> Error: %s listening for packets", err.Error()))
		}
		sS.Shutdown()
		exitChan <- true
		return
	}()
	stsV1 := v1.NewStatSv1(sS)
	server.RpcRegister(stsV1)
	internalStatSChan <- stsV1
}

// startThresholdService fires up the ThresholdS
func startThresholdService(internalThresholdSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dm *engine.DataManager, server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	filterS := <-filterSChan
	filterSChan <- filterS
	tS, err := engine.NewThresholdService(dm, cfg.ThresholdSCfg().StringIndexedFields,
		cfg.ThresholdSCfg().PrefixIndexedFields, cfg.ThresholdSCfg().StoreInterval, filterS)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<ThresholdS> Could not init, error: %s", err.Error()))
		exitChan <- true
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting Threshold Service"))
	go func() {
		if err := tS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<ThresholdS> Error: %s listening for packets", err.Error()))
		}
		tS.Shutdown()
		exitChan <- true
		return
	}()
	tSv1 := v1.NewThresholdSv1(tS)
	server.RpcRegister(tSv1)
	internalThresholdSChan <- tSv1
}

// startSupplierService fires up the ThresholdS
func startSupplierService(internalSupplierSChan, internalRsChan, internalStatSChan chan rpcclient.RpcClientConnection,
	cfg *config.CGRConfig, dm *engine.DataManager, server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	var err error
	filterS := <-filterSChan
	filterSChan <- filterS
	var resourceSConn, statSConn *rpcclient.RpcClientPool
	if len(cfg.SupplierSCfg().ResourceSConns) != 0 {
		resourceSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.ConnectAttempts, cfg.Reconnects,
			cfg.ConnectTimeout, cfg.ReplyTimeout, cfg.SupplierSCfg().ResourceSConns,
			internalRsChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to ResourceS: %s",
				utils.SupplierS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SupplierSCfg().StatSConns) != 0 {
		statSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.ConnectAttempts, cfg.Reconnects,
			cfg.ConnectTimeout, cfg.ReplyTimeout, cfg.SupplierSCfg().StatSConns,
			internalStatSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to StatS: %s",
				utils.SupplierS, err.Error()))
			exitChan <- true
			return
		}
	}
	splS, err := engine.NewSupplierService(dm, cfg.DefaultTimezone, filterS, cfg.SupplierSCfg().StringIndexedFields,
		cfg.SupplierSCfg().PrefixIndexedFields, resourceSConn, statSConn)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s",
			utils.SupplierS, err.Error()))
		exitChan <- true
		return
	}
	go func() {
		if err := splS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets",
				utils.SupplierS, err.Error()))
		}
		splS.Shutdown()
		exitChan <- true
		return
	}()
	splV1 := v1.NewSupplierSv1(splS)
	server.RpcRegister(splV1)
	internalSupplierSChan <- splV1
}

// startFilterService fires up the FilterS
func startFilterService(filterSChan chan *engine.FilterS,
	internalStatSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dm *engine.DataManager, exitChan chan bool) {

	filterSChan <- engine.NewFilterS(cfg, internalStatSChan, dm)

}

func startRpc(server *utils.Server, internalRaterChan,
	internalCdrSChan, internalCdrStatSChan, internalPubSubSChan, internalUserSChan,
	internalAliaseSChan, internalRsChan, internalStatSChan, internalSMGChan chan rpcclient.RpcClientConnection) {
	select { // Any of the rpc methods will unlock listening to rpc requests
	case resp := <-internalRaterChan:
		internalRaterChan <- resp
	case cdrs := <-internalCdrSChan:
		internalCdrSChan <- cdrs
	case cdrstats := <-internalCdrStatSChan:
		internalCdrStatSChan <- cdrstats
	case pubsubs := <-internalPubSubSChan:
		internalPubSubSChan <- pubsubs
	case users := <-internalUserSChan:
		internalUserSChan <- users
	case aliases := <-internalAliaseSChan:
		internalAliaseSChan <- aliases
	case smg := <-internalSMGChan:
		internalSMGChan <- smg
	case rls := <-internalRsChan:
		internalRsChan <- rls
	case statS := <-internalStatSChan:
		internalStatSChan <- statS
	}
	go server.ServeJSON(cfg.RPCJSONListen)
	go server.ServeGOB(cfg.RPCGOBListen)
	go server.ServeHTTP(
		cfg.HTTPListen,
		cfg.HTTPJsonRPCURL,
		cfg.HTTPWSURL,
		cfg.HTTPUseBasicAuth,
		cfg.HTTPAuthUsers,
	)
}

func writePid() {
	utils.Logger.Info(*pidFile)
	f, err := os.Create(*pidFile)
	if err != nil {
		log.Fatal("Could not write pid file: ", err)
	}
	f.WriteString(strconv.Itoa(os.Getpid()))
	if err := f.Close(); err != nil {
		log.Fatal("Could not write pid file: ", err)
	}
}

// initLogger will initialize syslog writter, needs to be called after config init
func initLogger(cfg *config.CGRConfig) error {
	sylogger := cfg.Logger
	if *syslogger != "" { // Modify the log level if provided by command arguments
		sylogger = *syslogger
	}
	err := utils.Newlogger(sylogger, cfg.NodeID)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}
	if *pidFile != "" {
		writePid()
	}
	if *singlecpu {
		runtime.GOMAXPROCS(1) // Having multiple cpus may slow down computing due to CPU management, to be reviewed in future Go releases
	}
	exitChan := make(chan bool)
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *scheduledShutdown != "" {
		shutdownDur, err := utils.ParseDurationWithNanosecs(*scheduledShutdown)
		if err != nil {
			log.Fatal(err)
		}
		go func() { // Schedule shutdown
			time.Sleep(shutdownDur)
			exitChan <- true
			return
		}()
	}
	var err error
	// Init config
	cfg, err = config.NewCGRConfigFromFolder(*cfgDir)
	if err != nil {
		log.Fatalf("Could not parse config: <%s>", err.Error())
		return
	}
	if *nodeID != "" {
		cfg.NodeID = *nodeID
	}
	config.SetCgrConfig(cfg)       // Share the config object
	cache.NewCache(cfg.CacheCfg()) // init cache
	// init syslog
	if err = initLogger(cfg); err != nil {
		log.Fatalf("Could not initialize syslog connection, err: <%s>", err.Error())
		return
	}
	lgLevel := cfg.LogLevel
	if *logLevel != -1 { // Modify the log level if provided by command arguments
		lgLevel = *logLevel
	}
	utils.Logger.SetLogLevel(lgLevel)
	var loadDb engine.LoadStorage
	var cdrDb engine.CdrStorage
	var dm *engine.DataManager

	if cfg.RALsEnabled || cfg.CDRStatsEnabled || cfg.PubSubServerEnabled ||
		cfg.AliasesServerEnabled || cfg.UserServerEnabled || cfg.SchedulerEnabled {
		dm, err = engine.ConfigureDataStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort,
			cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding, cfg.CacheCfg(), cfg.LoadHistorySize)
		if err != nil { // Cannot configure getter database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		}
		defer dm.DataDB().Close()
		engine.SetDataStorage(dm)
		if err := engine.CheckVersions(dm.DataDB()); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
	if cfg.RALsEnabled || cfg.CDRSEnabled || cfg.SchedulerEnabled { // Only connect to storDb if necessary
		storDb, err := engine.ConfigureStorStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort,
			cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding, cfg.StorDBMaxOpenConns,
			cfg.StorDBMaxIdleConns, cfg.StorDBConnMaxLifetime, cfg.StorDBCDRSIndexes)
		if err != nil { // Cannot configure logger database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure logger database: %s exiting!", err))
			return
		}
		defer storDb.Close()
		// loadDb,cdrDb and storDb are all mapped on the same stordb storage
		loadDb = storDb.(engine.LoadStorage)
		cdrDb = storDb.(engine.CdrStorage)
		engine.SetCdrStorage(cdrDb)
		if err := engine.CheckVersions(storDb); err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	// Done initing DBs
	engine.SetRoundingDecimals(cfg.RoundingDecimals)
	engine.SetRpSubjectPrefixMatching(cfg.RpSubjectPrefixMatching)
	engine.SetLcrSubjectPrefixMatching(cfg.LcrSubjectPrefixMatching)
	stopHandled := false

	// Rpc/http server
	server := new(utils.Server)

	// Async starts here, will follow cgrates.json start order

	// Define internal connections via channels
	internalRaterChan := make(chan rpcclient.RpcClientConnection, 1)
	cacheDoneChan := make(chan struct{}, 1)
	internalCdrSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalCdrStatSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalPubSubSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalUserSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAliaseSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalSMGChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAttributeSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalRsChan := make(chan rpcclient.RpcClientConnection, 1)
	internalStatSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalThresholdSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalSupplierSChan := make(chan rpcclient.RpcClientConnection, 1)
	filterSChan := make(chan *engine.FilterS, 1)

	// Start ServiceManager
	srvManager := servmanager.NewServiceManager(cfg, dm, exitChan, cacheDoneChan)

	// Start rater service
	if cfg.RALsEnabled {
		go startRater(internalRaterChan, cacheDoneChan, internalThresholdSChan,
			internalCdrStatSChan, internalStatSChan,
			internalPubSubSChan, internalAttributeSChan,
			internalUserSChan, internalAliaseSChan,
			srvManager, server, dm, loadDb, cdrDb, &stopHandled, exitChan)
	}

	// Start Scheduler
	if cfg.SchedulerEnabled {
		go srvManager.StartScheduler(true)
	}

	// Start CDR Server
	if cfg.CDRSEnabled {
		go startCDRS(internalCdrSChan, cdrDb, dm,
			internalRaterChan, internalPubSubSChan, internalAttributeSChan,
			internalUserSChan, internalAliaseSChan, internalCdrStatSChan,
			internalThresholdSChan, internalStatSChan, server, exitChan)
	}

	// Start CDR Stats server
	if cfg.CDRStatsEnabled {
		go startCdrStats(internalCdrStatSChan, dm, server)
	}

	// Start CDRC components if necessary
	go startCdrcs(internalCdrSChan, internalRaterChan, exitChan)

	// Start SM-Generic
	if cfg.SessionSCfg().Enabled {
		go startSessionS(internalSMGChan, internalRaterChan, internalRsChan,
			internalSupplierSChan, internalAttributeSChan, internalCdrSChan, server, exitChan)
	}
	// Start FreeSWITCHAgent
	if cfg.FsAgentCfg().Enabled {
		go startFsAgent(internalSMGChan, exitChan)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler(exitChan)
	}

	// Start SM-Kamailio
	if cfg.SmKamConfig.Enabled {
		go startSmKamailio(internalRaterChan, internalCdrSChan, internalRsChan, cdrDb, exitChan)
	}

	if cfg.AsteriskAgentCfg().Enabled {
		go startAsteriskAgent(internalSMGChan, exitChan)
	}

	if cfg.DiameterAgentCfg().Enabled {
		go startDiameterAgent(internalSMGChan, internalPubSubSChan, exitChan)
	}

	if cfg.RadiusAgentCfg().Enabled {
		go startRadiusAgent(internalSMGChan, exitChan)
	}

	// Start PubSubS service
	if cfg.PubSubServerEnabled {
		go startPubSubServer(internalPubSubSChan, dm, server, exitChan)
	}

	// Start Aliases service
	if cfg.AliasesServerEnabled {
		go startAliasesServer(internalAliaseSChan, dm, server, exitChan)
	}

	// Start users service
	if cfg.UserServerEnabled {
		go startUsersServer(internalUserSChan, dm, server, exitChan)
	}
	// Start FilterS
	go startFilterService(filterSChan, internalStatSChan, cfg, dm, exitChan)

	if cfg.AttributeSCfg().Enabled {
		go startAttributeService(internalAttributeSChan, cfg, dm, server, exitChan, filterSChan)
	}

	// Start RL service
	if cfg.ResourceSCfg().Enabled {
		go startResourceService(internalRsChan,
			internalThresholdSChan, cfg, dm, server, exitChan, filterSChan)
	}

	if cfg.StatSCfg().Enabled {
		go startStatService(internalStatSChan, internalThresholdSChan, cfg, dm, server, exitChan, filterSChan)
	}

	if cfg.ThresholdSCfg().Enabled {
		go startThresholdService(internalThresholdSChan, cfg, dm, server, exitChan, filterSChan)
	}

	if cfg.SupplierSCfg().Enabled {
		go startSupplierService(internalSupplierSChan, internalRsChan, internalStatSChan,
			cfg, dm, server, exitChan, filterSChan)
	}

	// Serve rpc connections
	go startRpc(server, internalRaterChan, internalCdrSChan, internalCdrStatSChan,
		internalPubSubSChan, internalUserSChan, internalAliaseSChan, internalRsChan, internalStatSChan, internalSMGChan)
	<-exitChan

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile file: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}

	if *pidFile != "" {
		if err := os.Remove(*pidFile); err != nil {
			utils.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	utils.Logger.Info("Stopped all components. CGRateS shutdown!")
}
