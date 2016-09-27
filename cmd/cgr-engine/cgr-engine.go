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
	//	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/balancer2go"
	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/cdrc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/scheduler"
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
	raterEnabled      = flag.Bool("rater", false, "Enforce starting of the rater daemon overwriting config")
	schedEnabled      = flag.Bool("scheduler", false, "Enforce starting of the scheduler daemon .overwriting config")
	cdrsEnabled       = flag.Bool("cdrs", false, "Enforce starting of the cdrs daemon overwriting config")
	pidFile           = flag.String("pid", "", "Write pid file")
	cpuprofile        = flag.String("cpuprofile", "", "write cpu profile to file")
	scheduledShutdown = flag.String("scheduled_shutdown", "", "shutdown the engine after this duration")
	singlecpu         = flag.Bool("singlecpu", false, "Run on single CPU core")

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
	sm := sessionmanager.NewSMGeneric(cfg, ralsConns, cdrsConn, cfg.DefaultTimezone)
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SMGeneric> error: %s!", err))
	}
	// Pass internal connection via BiRPCClient
	internalSMGChan <- sm
	// Register RPC handler
	smgRpc := v1.NewSMGenericV1(sm)
	server.RpcRegister(smgRpc)
	// Register BiRpc handlers
	//server.BiRPCRegister(v1.NewSMGenericBiRpcV1(sm))
	smgBiRpc := v1.NewSMGenericBiRpcV1(sm)
	for method, handler := range smgBiRpc.Handlers() {
		server.BiRPCRegisterName(method, handler)
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
	utils.Logger.Info("Starting CGRateS DiameterAgent service.")
	smgChan := make(chan rpcclient.RpcClientConnection, 1) // Use it to pass smg
	go func(internalSMGChan chan *sessionmanager.SMGeneric, smgChan chan rpcclient.RpcClientConnection) {
		// Need this to pass from *sessionmanager.SMGeneric to rpcclient.RpcClientConnection
		smg := <-internalSMGChan
		internalSMGChan <- smg
		smgChan <- smg
	}(internalSMGChan, smgChan)
	var smgConn, pubsubConn *rpcclient.RpcClientPool

	if len(cfg.DiameterAgentCfg().SMGenericConns) != 0 {
		smgConn, err = engine.NewRPCPool(rpcclient.POOL_BROADCAST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
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

func startSmFreeSWITCH(internalRaterChan, internalCDRSChan, rlsChan chan rpcclient.RpcClientConnection, cdrDb engine.CdrStorage, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS SMFreeSWITCH service.")
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

func startSmKamailio(internalRaterChan, internalCDRSChan chan rpcclient.RpcClientConnection, cdrDb engine.CdrStorage, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS SMKamailio service.")
	var ralsConn, cdrsConn *rpcclient.RpcClientPool
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
			utils.Logger.Crit(fmt.Sprintf("<SMKamailio> Could not connect to RAL: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	sm, _ := sessionmanager.NewKamailioSessionManager(cfg.SmKamConfig, ralsConn, cdrsConn, cfg.DefaultTimezone)
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

func startCDRS(internalCdrSChan chan rpcclient.RpcClientConnection, cdrDb engine.CdrStorage, dataDB engine.AccountingStorage,
	internalRaterChan chan rpcclient.RpcClientConnection, internalPubSubSChan chan rpcclient.RpcClientConnection,
	internalUserSChan chan rpcclient.RpcClientConnection, internalAliaseSChan chan rpcclient.RpcClientConnection,
	internalCdrStatSChan chan rpcclient.RpcClientConnection, server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS CDRS service.")
	var ralConn, pubSubConn, usersConn, aliasesConn, statsConn *rpcclient.RpcClientPool
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

	if len(cfg.CDRSStatSConns) != 0 { // Stats connection init
		statsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.CDRSStatSConns, internalCdrStatSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to StatS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	cdrServer, _ := engine.NewCdrServer(cfg, cdrDb, dataDB, ralConn, pubSubConn, usersConn, aliasesConn, statsConn)
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

func startScheduler(internalSchedulerChan chan *scheduler.Scheduler, cacheDoneChan chan struct{}, ratingDb engine.RatingStorage, exitChan chan bool) {
	// Wait for cache to load data before starting
	cacheDone := <-cacheDoneChan
	cacheDoneChan <- cacheDone
	utils.Logger.Info("Starting CGRateS Scheduler.")
	sched := scheduler.NewScheduler(ratingDb)
	go reloadSchedulerSingnalHandler(sched, ratingDb)
	time.Sleep(1)
	internalSchedulerChan <- sched
	sched.Reload(true)
	sched.Loop()
	exitChan <- true // Should not get out of loop though
}

func startCdrStats(internalCdrStatSChan chan rpcclient.RpcClientConnection, ratingDb engine.RatingStorage, accountDb engine.AccountingStorage, server *utils.Server) {
	cdrStats := engine.NewStats(ratingDb, accountDb, cfg.CDRStatsSaveInterval)
	server.RpcRegister(cdrStats)
	server.RpcRegister(&v1.CDRStatsV1{CdrStats: cdrStats}) // Public APIs
	internalCdrStatSChan <- cdrStats
}

func startHistoryServer(internalHistorySChan chan rpcclient.RpcClientConnection, server *utils.Server, exitChan chan bool) {
	scribeServer, err := history.NewFileScribe(cfg.HistoryDir, cfg.HistorySaveInterval)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<HistoryServer> Could not start, error: %s", err.Error()))
		exitChan <- true
	}
	server.RpcRegisterName("HistoryV1", scribeServer)
	internalHistorySChan <- scribeServer
}

func startPubSubServer(internalPubSubSChan chan rpcclient.RpcClientConnection, accountDb engine.AccountingStorage, server *utils.Server) {
	pubSubServer := engine.NewPubSub(accountDb, cfg.HttpSkipTlsVerify)
	server.RpcRegisterName("PubSubV1", pubSubServer)
	internalPubSubSChan <- pubSubServer
}

// ToDo: Make sure we are caching before starting this one
func startAliasesServer(internalAliaseSChan chan rpcclient.RpcClientConnection, accountDb engine.AccountingStorage, server *utils.Server, exitChan chan bool) {
	aliasesServer := engine.NewAliasHandler(accountDb)
	server.RpcRegisterName("AliasesV1", aliasesServer)
	loadHist, err := accountDb.GetLoadHistory(1, true, utils.NonTransactional)
	if err != nil || len(loadHist) == 0 {
		utils.Logger.Info(fmt.Sprintf("could not get load history: %v (%v)", loadHist, err))
		internalAliaseSChan <- aliasesServer
		return
	}

	if err := accountDb.PreloadAccountingCache(); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<Aliases> Could not start, error: %s", err.Error()))
		exitChan <- true
		return
	}

	internalAliaseSChan <- aliasesServer
}

func startUsersServer(internalUserSChan chan rpcclient.RpcClientConnection, accountDb engine.AccountingStorage, server *utils.Server, exitChan chan bool) {
	userServer, err := engine.NewUserMap(accountDb, cfg.UserServerIndexes)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<UsersService> Could not start, error: %s", err.Error()))
		exitChan <- true
		return
	}
	server.RpcRegisterName("UsersV1", userServer)
	internalUserSChan <- userServer
}

func startResourceLimiterService(internalRLSChan, internalCdrStatSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	accountDb engine.AccountingStorage, server *utils.Server, exitChan chan bool) {
	var statsConn *rpcclient.RpcClientPool
	if len(cfg.ResourceLimiterCfg().CDRStatConns) != 0 { // Stats connection init
		statsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
			cfg.ResourceLimiterCfg().CDRStatConns, internalCdrStatSChan, cfg.InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<RLs> Could not connect to StatS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	rls, err := engine.NewResourceLimiterService(cfg, accountDb, statsConn)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<RLs> Could not init, error: %s", err.Error()))
		exitChan <- true
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting ResourceLimiter service"))
	if err := rls.ListenAndServe(); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<RLs> Could not start, error: %s", err.Error()))
		exitChan <- true
		return
	}
	server.RpcRegisterName("RLsV1", rls)
	internalRLSChan <- rls
}

func startRpc(server *utils.Server, internalRaterChan,
	internalCdrSChan, internalCdrStatSChan, internalHistorySChan, internalPubSubSChan, internalUserSChan,
	internalAliaseSChan chan rpcclient.RpcClientConnection, internalSMGChan chan *sessionmanager.SMGeneric) {
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
	}
	go server.ServeJSON(cfg.RPCJSONListen)
	go server.ServeGOB(cfg.RPCGOBListen)
	go server.ServeHTTP(cfg.HTTPListen)
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

func main() {
	flag.Parse()
	if *version {
		fmt.Println("CGRateS " + utils.VERSION)
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
		}()
	}
	cfg, err = config.NewCGRConfigFromFolder(*cfgDir)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("Could not parse config: %s exiting!", err))
		return
	}
	config.SetCgrConfig(cfg) // Share the config object
	cache2go.NewCache(cfg.CacheConfig)

	if *raterEnabled {
		cfg.RALsEnabled = *raterEnabled
	}
	if *schedEnabled {
		cfg.SchedulerEnabled = *schedEnabled
	}
	if *cdrsEnabled {
		cfg.CDRSEnabled = *cdrsEnabled
	}
	var ratingDb engine.RatingStorage
	var accountDb engine.AccountingStorage
	var loadDb engine.LoadStorage
	var cdrDb engine.CdrStorage
	if cfg.RALsEnabled || cfg.SchedulerEnabled || cfg.CDRStatsEnabled { // Only connect to dataDb if necessary
		ratingDb, err = engine.ConfigureRatingStorage(cfg.TpDbType, cfg.TpDbHost, cfg.TpDbPort,
			cfg.TpDbName, cfg.TpDbUser, cfg.TpDbPass, cfg.DBDataEncoding, cfg.CacheConfig, cfg.LoadHistorySize)
		if err != nil { // Cannot configure getter database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		}
		defer ratingDb.Close()
		engine.SetRatingStorage(ratingDb)
	}
	if cfg.RALsEnabled || cfg.CDRStatsEnabled || cfg.PubSubServerEnabled || cfg.AliasesServerEnabled || cfg.UserServerEnabled {
		accountDb, err = engine.ConfigureAccountingStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort,
			cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding, cfg.CacheConfig, cfg.LoadHistorySize)
		if err != nil { // Cannot configure getter database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		}
		defer accountDb.Close()
		engine.SetAccountingStorage(accountDb)
		if err := engine.CheckVersion(nil); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
	if cfg.RALsEnabled || cfg.CDRSEnabled || cfg.SchedulerEnabled { // Only connect to storDb if necessary
		storDb, err := engine.ConfigureStorStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort,
			cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding, cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns, cfg.StorDBCDRSIndexes)
		if err != nil { // Cannot configure logger database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure logger database: %s exiting!", err))
			return
		}
		defer storDb.Close()
		// loadDb,cdrDb and storDb are all mapped on the same stordb storage
		loadDb = storDb.(engine.LoadStorage)
		cdrDb = storDb.(engine.CdrStorage)
		engine.SetCdrStorage(cdrDb)
	}

	engine.SetRoundingDecimals(cfg.RoundingDecimals)
	engine.SetRpSubjectPrefixMatching(cfg.RpSubjectPrefixMatching)
	engine.SetLcrSubjectPrefixMatching(cfg.LcrSubjectPrefixMatching)
	stopHandled := false

	// Rpc/http server
	server := new(utils.Server)

	// Async starts here, will follow cgrates.json start order

	// Define internal connections via channels
	internalBalancerChan := make(chan *balancer2go.Balancer, 1)
	internalRaterChan := make(chan rpcclient.RpcClientConnection, 1)
	cacheDoneChan := make(chan struct{}, 1)
	internalSchedulerChan := make(chan *scheduler.Scheduler, 1)
	internalCdrSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalCdrStatSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalHistorySChan := make(chan rpcclient.RpcClientConnection, 1)
	internalPubSubSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalUserSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAliaseSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalSMGChan := make(chan *sessionmanager.SMGeneric, 1)
	internalRLSChan := make(chan rpcclient.RpcClientConnection, 1)
	// Start balancer service
	if cfg.BalancerEnabled {
		go startBalancer(internalBalancerChan, &stopHandled, exitChan) // Not really needed async here but to cope with uniformity
	}

	// Start rater service
	if cfg.RALsEnabled {
		go startRater(internalRaterChan, cacheDoneChan, internalBalancerChan, internalSchedulerChan, internalCdrStatSChan, internalHistorySChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan,
			server, ratingDb, accountDb, loadDb, cdrDb, &stopHandled, exitChan)
	}

	// Start Scheduler
	if cfg.SchedulerEnabled {
		go startScheduler(internalSchedulerChan, cacheDoneChan, ratingDb, exitChan)
	}

	// Start CDR Server
	if cfg.CDRSEnabled {
		go startCDRS(internalCdrSChan, cdrDb, accountDb,
			internalRaterChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan, internalCdrStatSChan, server, exitChan)
	}

	// Start CDR Stats server
	if cfg.CDRStatsEnabled {
		go startCdrStats(internalCdrStatSChan, ratingDb, accountDb, server)
	}

	// Start CDRC components if necessary
	go startCdrcs(internalCdrSChan, internalRaterChan, exitChan)

	// Start SM-Generic
	if cfg.SmGenericConfig.Enabled {
		go startSmGeneric(internalSMGChan, internalRaterChan, internalCdrSChan, server, exitChan)
	}
	// Start SM-FreeSWITCH
	if cfg.SmFsConfig.Enabled {
		go startSmFreeSWITCH(internalRaterChan, internalCdrSChan, internalRLSChan, cdrDb, exitChan)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler(exitChan)
	}

	// Start SM-Kamailio
	if cfg.SmKamConfig.Enabled {
		go startSmKamailio(internalRaterChan, internalCdrSChan, cdrDb, exitChan)
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

	// Start HistoryS service
	if cfg.HistoryServerEnabled {
		go startHistoryServer(internalHistorySChan, server, exitChan)
	}

	// Start PubSubS service
	if cfg.PubSubServerEnabled {
		go startPubSubServer(internalPubSubSChan, accountDb, server)
	}

	// Start Aliases service
	if cfg.AliasesServerEnabled {
		go startAliasesServer(internalAliaseSChan, accountDb, server, exitChan)
	}

	// Start users service
	if cfg.UserServerEnabled {
		go startUsersServer(internalUserSChan, accountDb, server, exitChan)
	}

	// Start RL service
	if cfg.ResourceLimiterCfg().Enabled {
		go startResourceLimiterService(internalRLSChan, internalCdrStatSChan, cfg, accountDb, server, exitChan)
	}

	// Serve rpc connections
	go startRpc(server, internalRaterChan, internalCdrSChan, internalCdrStatSChan, internalHistorySChan,
		internalPubSubSChan, internalUserSChan, internalAliaseSChan, internalSMGChan)
	<-exitChan

	if *pidFile != "" {
		if err := os.Remove(*pidFile); err != nil {
			utils.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	utils.Logger.Info("Stopped all components. CGRateS shutdown!")
}
