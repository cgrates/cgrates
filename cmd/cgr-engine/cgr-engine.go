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
	"github.com/cgrates/cgrates/history"
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
	scheduledShutdown = flag.String("scheduled_shutdown", "", "shutdown the engine after this duration")
	singlecpu         = flag.Bool("singlecpu", false, "Run on single CPU core")
	syslogger         = flag.String("logger", "", "logger <*syslog|*stdout>")
	logLevel          = flag.Int("log_level", -1, "Log level (0-emergency to 7-debug)")

	cfg   *config.CGRConfig
	smRpc *v1.SessionManagerV1
	err   error
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

func startSmGeneric(internalSMGChan chan *sessionmanager.SMGeneric, internalRaterChan, internalCDRSChan chan rpcclient.RpcClientConnection, server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS SMGeneric service.")
	var ralsConns, cdrsConn *rpcclient.RpcClientPool
	if len(cfg.SmGenericConfig.RALsConns) != 0 {
		ralsConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SmGenericConfig.RALsConns, internalRaterChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<SMGeneric> Could not connect to RALs: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SmGenericConfig.CDRsConns) != 0 {
		cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SmGenericConfig.CDRsConns, internalCDRSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<SMGeneric> Could not connect to RALs: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	smgReplConns, err := sessionmanager.NewSMGReplicationConns(cfg.SmGenericConfig.SMGReplicationConns, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<SMGeneric> Could not connect to SMGReplicationConnection error: <%s>", err.Error()))
		exitChan <- true
		return
	}
	sm := sessionmanager.NewSMGeneric(cfg, ralsConns, cdrsConn, smgReplConns, cfg.DefaultTimezone)
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMGeneric> error: %s!", err))
	}
	// Pass internal connection via BiRPCClient
	internalSMGChan <- sm
	// Register RPC handler
	smgRpc := v1.NewSMGenericV1(sm)
	server.RpcRegister(smgRpc)
	server.RpcRegister(&v2.SMGenericV2{*smgRpc})
	// Register BiRpc handlers
	if cfg.SmGenericConfig.ListenBijson != "" {
		smgBiRpc := v1.NewSMGenericBiRpcV1(sm)
		for method, handler := range smgBiRpc.Handlers() {
			server.BiRPCRegisterName(method, handler)
		}
		server.ServeBiJSON(cfg.SmGenericConfig.ListenBijson)
		exitChan <- true
	}
}

func startSMAsterisk(internalSMGChan chan *sessionmanager.SMGeneric, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS SMAsterisk service.")
	/*
		var smgConn *rpcclient.RpcClientPool
		if len(cfg.SMAsteriskCfg().SMGConns) != 0 {
			smgConn, err = engine.NewRPCPool(rpcclient.POOL_BROADCAST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
				cfg.SMAsteriskCfg().SMGConns, internalSMGChan, cfg.InternalTtl)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<SMAsterisk> Could not connect to SMG: %s", err.Error()))
				exitChan <- true
				return
			}
		}
	*/
	smg := <-internalSMGChan
	internalSMGChan <- smg
	birpcClnt := utils.NewBiRPCInternalClient(smg)
	for connIdx := range cfg.SMAsteriskCfg().AsteriskConns { // Instantiate connections towards asterisk servers
		sma, err := sessionmanager.NewSMAsterisk(cfg, connIdx, birpcClnt)
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

func startDiameterAgent(internalSMGChan chan *sessionmanager.SMGeneric, internalPubSubSChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS DiameterAgent service")
	smgChan := make(chan rpcclient.RpcClientConnection, 1) // Use it to pass smg
	go func(internalSMGChan chan *sessionmanager.SMGeneric, smgChan chan rpcclient.RpcClientConnection) {
		// Need this to pass from *sessionmanager.SMGeneric to rpcclient.RpcClientConnection
		smg := <-internalSMGChan
		internalSMGChan <- smg
		smgChan <- smg
	}(internalSMGChan, smgChan)
	var smgConn, pubsubConn *rpcclient.RpcClientPool

	if len(cfg.DiameterAgentCfg().SMGenericConns) != 0 {
		smgConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.DiameterAgentCfg().SMGenericConns, smgChan, cfg.InternalTtl)
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

func startRadiusAgent(internalSMGChan chan *sessionmanager.SMGeneric, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS RadiusAgent service")
	smgChan := make(chan rpcclient.RpcClientConnection, 1) // Use it to pass smg
	go func(internalSMGChan chan *sessionmanager.SMGeneric, smgChan chan rpcclient.RpcClientConnection) {
		// Need this to pass from *sessionmanager.SMGeneric to rpcclient.RpcClientConnection
		smg := <-internalSMGChan
		internalSMGChan <- smg
		smgChan <- smg
	}(internalSMGChan, smgChan)
	var smgConn *rpcclient.RpcClientPool
	if len(cfg.RadiusAgentCfg().SMGenericConns) != 0 {
		smgConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.RadiusAgentCfg().SMGenericConns, smgChan, cfg.InternalTtl)
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

func startSmFreeSWITCH(internalRaterChan, internalCDRSChan, rlsChan chan rpcclient.RpcClientConnection, cdrDb engine.CdrStorage, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS SMFreeSWITCH service")
	var ralsConn, cdrsConn, rlsConn *rpcclient.RpcClientPool
	if len(cfg.SmFsConfig.RALsConns) != 0 {
		ralsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SmFsConfig.RALsConns, internalRaterChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<SMFreeSWITCH> Could not connect to RAL: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SmFsConfig.CDRsConns) != 0 {
		cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SmFsConfig.CDRsConns, internalCDRSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<SMFreeSWITCH> Could not connect to RAL: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SmFsConfig.RLsConns) != 0 {
		rlsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SmFsConfig.RLsConns, rlsChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<SMFreeSWITCH> Could not connect to RLs: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	sm := sessionmanager.NewFSSessionManager(cfg.SmFsConfig, ralsConn, cdrsConn, rlsConn, cfg.DefaultTimezone)
	smRpc.SMs = append(smRpc.SMs, sm)
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMFreeSWITCH> error: %s!", err))
	}
	exitChan <- true
}

func startSmKamailio(internalRaterChan, internalCDRSChan, internalRsChan chan rpcclient.RpcClientConnection, cdrDb engine.CdrStorage, exitChan chan bool) {
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
	smRpc.SMs = append(smRpc.SMs, sm)
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMKamailio> error: %s!", err))
	}
	exitChan <- true
}

func startSmOpenSIPS(internalRaterChan, internalCDRSChan chan rpcclient.RpcClientConnection, cdrDb engine.CdrStorage, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS SMOpenSIPS service.")
	var ralsConn, cdrsConn *rpcclient.RpcClientPool
	if len(cfg.SmOsipsConfig.RALsConns) != 0 {
		ralsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SmOsipsConfig.RALsConns, internalRaterChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<SMOpenSIPS> Could not connect to RALs: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SmOsipsConfig.CDRsConns) != 0 {
		cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.SmOsipsConfig.CDRsConns, internalCDRSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<SMOpenSIPS> Could not connect to CDRs: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	sm, _ := sessionmanager.NewOSipsSessionManager(cfg.SmOsipsConfig, cfg.Reconnects, ralsConn, cdrsConn, cfg.DefaultTimezone)
	smRpc.SMs = append(smRpc.SMs, sm)
	if err := sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> error: %s!", err))
	}
	exitChan <- true
}

func startCDRS(internalCdrSChan chan rpcclient.RpcClientConnection,
	cdrDb engine.CdrStorage, dataDB engine.DataDB,
	internalRaterChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan,
	internalCdrStatSChan, internalStatSChan chan rpcclient.RpcClientConnection,
	server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS CDRS service.")
	var ralConn, pubSubConn, usersConn, aliasesConn, cdrstatsConn, statsConn *rpcclient.RpcClientPool
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
	if len(cfg.CDRSStatSConns) != 0 { // Stats connection init
		statsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.CDRSStatSConns, internalStatSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to StatS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	cdrServer, _ := engine.NewCdrServer(cfg, cdrDb, dataDB, ralConn, pubSubConn,
		usersConn, aliasesConn, cdrstatsConn, statsConn)
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

func startScheduler(internalSchedulerChan chan *scheduler.Scheduler, cacheDoneChan chan struct{}, dataDB engine.DataDB, exitChan chan bool) {
	// Wait for cache to load data before starting
	cacheDone := <-cacheDoneChan
	cacheDoneChan <- cacheDone
	utils.Logger.Info("Starting CGRateS Scheduler.")
	sched := scheduler.NewScheduler(dataDB)
	internalSchedulerChan <- sched

	sched.Loop()
	exitChan <- true // Should not get out of loop though
}

func startCdrStats(internalCdrStatSChan chan rpcclient.RpcClientConnection, dataDB engine.DataDB, server *utils.Server) {
	cdrStats := engine.NewStats(dataDB, cfg.CDRStatsSaveInterval)
	server.RpcRegister(cdrStats)
	server.RpcRegister(&v1.CDRStatsV1{CdrStats: cdrStats}) // Public APIs
	internalCdrStatSChan <- cdrStats
}

func startHistoryServer(internalHistorySChan chan rpcclient.RpcClientConnection, server *utils.Server, exitChan chan bool) {
	scribeServer, err := history.NewFileScribe(cfg.HistoryDir, cfg.HistorySaveInterval)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<HistoryServer> Could not start, error: %s", err.Error()))
		exitChan <- true
		return
	}
	server.RpcRegisterName("HistoryV1", scribeServer)
	internalHistorySChan <- scribeServer
}

func startPubSubServer(internalPubSubSChan chan rpcclient.RpcClientConnection, dataDB engine.DataDB, server *utils.Server, exitChan chan bool) {
	pubSubServer, err := engine.NewPubSub(dataDB, cfg.HttpSkipTlsVerify)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<PubSubS> Could not start, error: %s", err.Error()))
		exitChan <- true
		return
	}
	server.RpcRegisterName("PubSubV1", pubSubServer)
	internalPubSubSChan <- pubSubServer
}

// ToDo: Make sure we are caching before starting this one
func startAliasesServer(internalAliaseSChan chan rpcclient.RpcClientConnection, dataDB engine.DataDB, server *utils.Server, exitChan chan bool) {
	aliasesServer := engine.NewAliasHandler(dataDB)
	server.RpcRegisterName("AliasesV1", aliasesServer)
	loadHist, err := dataDB.GetLoadHistory(1, true, utils.NonTransactional)
	if err != nil || len(loadHist) == 0 {
		utils.Logger.Info(fmt.Sprintf("could not get load history: %v (%v)", loadHist, err))
		internalAliaseSChan <- aliasesServer
		return
	}
	internalAliaseSChan <- aliasesServer
}

func startUsersServer(internalUserSChan chan rpcclient.RpcClientConnection, dataDB engine.DataDB, server *utils.Server, exitChan chan bool) {
	userServer, err := engine.NewUserMap(dataDB, cfg.UserServerIndexes)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<UsersService> Could not start, error: %s", err.Error()))
		exitChan <- true
		return
	}
	server.RpcRegisterName("UsersV1", userServer)
	internalUserSChan <- userServer
}

func startResourceService(internalRsChan, internalStatSConn chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dataDB engine.DataDB, server *utils.Server, exitChan chan bool) {
	var statsConn *rpcclient.RpcClientPool
	if len(cfg.ResourceSCfg().StatSConns) != 0 { // Stats connection init
		statsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.ResourceSCfg().StatSConns, internalStatSConn, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<ResourceS> Could not connect to StatS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	rS, err := engine.NewResourceService(dataDB, cfg.ResourceSCfg().StoreInterval, statsConn)
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
	rsV1 := v1.NewResourceSV1(rS)
	server.RpcRegister(rsV1)
	internalRsChan <- rsV1
}

// startStatService fires up the StatS
func startStatService(internalStatSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dm *engine.DataManager, server *utils.Server, exitChan chan bool) {
	sS, err := engine.NewStatService(dm, cfg.StatSCfg().StoreInterval)
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
	stsV1 := v1.NewStatSV1(sS)
	server.RpcRegister(stsV1)
	internalStatSChan <- stsV1
}

func startRpc(server *utils.Server, internalRaterChan,
	internalCdrSChan, internalCdrStatSChan, internalHistorySChan, internalPubSubSChan, internalUserSChan,
	internalAliaseSChan, internalRsChan, internalStatSChan chan rpcclient.RpcClientConnection, internalSMGChan chan *sessionmanager.SMGeneric) {
	select { // Any of the rpc methods will unlock listening to rpc requests
	case resp := <-internalRaterChan:
		internalRaterChan <- resp
	case cdrs := <-internalCdrSChan:
		internalCdrSChan <- cdrs
	case cdrstats := <-internalCdrStatSChan:
		internalCdrStatSChan <- cdrstats
	case hist := <-internalHistorySChan:
		internalHistorySChan <- hist
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
	err := utils.Newlogger(sylogger)
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
		shutdownDur, err := utils.ParseDurationWithSecs(*scheduledShutdown)
		if err != nil {
			log.Fatal(err)
		}
		go func() { // Schedule shutdown
			time.Sleep(shutdownDur)
			exitChan <- true
			return
		}()
	}
	// Init config
	cfg, err = config.NewCGRConfigFromFolder(*cfgDir)
	if err != nil {
		log.Fatalf("Could not parse config: ", err)
		return
	}
	config.SetCgrConfig(cfg) // Share the config object
	// init syslog
	if err = initLogger(cfg); err != nil {
		log.Fatalf("Could not initialize sylog connection, err: <%s>", err.Error())
		return
	}
	lgLevel := cfg.LogLevel
	if *logLevel != -1 { // Modify the log level if provided by command arguments
		lgLevel = *logLevel
	}
	utils.Logger.SetLogLevel(lgLevel)

	// Init cache
	cache.NewCache(cfg.CacheConfig)

	var dataDB engine.DataDB
	var loadDb engine.LoadStorage
	var cdrDb engine.CdrStorage
	var dm *engine.DataManager

	if cfg.RALsEnabled || cfg.CDRStatsEnabled || cfg.PubSubServerEnabled || cfg.AliasesServerEnabled || cfg.UserServerEnabled || cfg.SchedulerEnabled {
		dataDB, err = engine.ConfigureDataStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort,
			cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding, cfg.CacheConfig, cfg.LoadHistorySize)
		if err != nil { // Cannot configure getter database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		}
		defer dataDB.Close()
		engine.SetDataStorage(dataDB)
		if err := engine.CheckVersions(dataDB); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
	if cfg.RALsEnabled || cfg.CDRSEnabled || cfg.SchedulerEnabled { // Only connect to storDb if necessary
		storDb, err := engine.ConfigureStorStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort,
			cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding, cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns, cfg.StorDBConnMaxLifetime, cfg.StorDBCDRSIndexes)
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

	dm = engine.NewDataManager(dataDB)
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
	internalHistorySChan := make(chan rpcclient.RpcClientConnection, 1)
	internalPubSubSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalUserSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAliaseSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalSMGChan := make(chan *sessionmanager.SMGeneric, 1)
	internalRsChan := make(chan rpcclient.RpcClientConnection, 1)
	internalStatSChan := make(chan rpcclient.RpcClientConnection, 1)

	// Start ServiceManager
	srvManager := servmanager.NewServiceManager(cfg, dataDB, exitChan, cacheDoneChan)

	// Start rater service
	if cfg.RALsEnabled {
		go startRater(internalRaterChan, cacheDoneChan, internalCdrStatSChan, internalStatSChan,
			internalHistorySChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan,
			srvManager, server, dataDB, loadDb, cdrDb, &stopHandled, exitChan)
	}

	// Start Scheduler
	if cfg.SchedulerEnabled {
		go srvManager.StartScheduler(true)
	}

	// Start CDR Server
	if cfg.CDRSEnabled {
		go startCDRS(internalCdrSChan, cdrDb, dataDB,
			internalRaterChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan,
			internalCdrStatSChan, internalStatSChan, server, exitChan)
	}

	// Start CDR Stats server
	if cfg.CDRStatsEnabled {
		go startCdrStats(internalCdrStatSChan, dataDB, server)
	}

	// Start CDRC components if necessary
	go startCdrcs(internalCdrSChan, internalRaterChan, exitChan)

	// Start SM-Generic
	if cfg.SmGenericConfig.Enabled {
		go startSmGeneric(internalSMGChan, internalRaterChan, internalCdrSChan, server, exitChan)
	}
	// Start SM-FreeSWITCH
	if cfg.SmFsConfig.Enabled {
		go startSmFreeSWITCH(internalRaterChan, internalCdrSChan, internalRsChan, cdrDb, exitChan)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler(exitChan)
	}

	// Start SM-Kamailio
	if cfg.SmKamConfig.Enabled {
		go startSmKamailio(internalRaterChan, internalCdrSChan, internalRsChan, cdrDb, exitChan)
	}

	// Start SM-OpenSIPS
	if cfg.SmOsipsConfig.Enabled {
		go startSmOpenSIPS(internalRaterChan, internalCdrSChan, cdrDb, exitChan)
	}

	// Register session manager service // FixMe: make sure this is thread safe
	if cfg.SmGenericConfig.Enabled || cfg.SmFsConfig.Enabled || cfg.SmKamConfig.Enabled || cfg.SmOsipsConfig.Enabled || cfg.SMAsteriskCfg().Enabled { // Register SessionManagerV1 service
		smRpc = new(v1.SessionManagerV1)
		server.RpcRegister(smRpc)
	}

	if cfg.SMAsteriskCfg().Enabled {
		go startSMAsterisk(internalSMGChan, exitChan)
	}

	if cfg.DiameterAgentCfg().Enabled {
		go startDiameterAgent(internalSMGChan, internalPubSubSChan, exitChan)
	}

	if cfg.RadiusAgentCfg().Enabled {
		go startRadiusAgent(internalSMGChan, exitChan)
	}

	// Start HistoryS service
	if cfg.HistoryServerEnabled {
		go startHistoryServer(internalHistorySChan, server, exitChan)
	}

	// Start PubSubS service
	if cfg.PubSubServerEnabled {
		go startPubSubServer(internalPubSubSChan, dataDB, server, exitChan)
	}

	// Start Aliases service
	if cfg.AliasesServerEnabled {
		go startAliasesServer(internalAliaseSChan, dataDB, server, exitChan)
	}

	// Start users service
	if cfg.UserServerEnabled {
		go startUsersServer(internalUserSChan, dataDB, server, exitChan)
	}

	// Start RL service
	if cfg.ResourceSCfg().Enabled {
		go startResourceService(internalRsChan,
			internalStatSChan, cfg, dataDB, server, exitChan)
	}

	if cfg.StatSCfg().Enabled {
		go startStatService(internalStatSChan, cfg, dm, server, exitChan)
	}

	// Serve rpc connections
	go startRpc(server, internalRaterChan, internalCdrSChan, internalCdrStatSChan, internalHistorySChan,
		internalPubSubSChan, internalUserSChan, internalAliaseSChan, internalRsChan, internalStatSChan, internalSMGChan)
	<-exitChan

	if *pidFile != "" {
		if err := os.Remove(*pidFile); err != nil {
			utils.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	utils.Logger.Info("Stopped all components. CGRateS shutdown!")
}
