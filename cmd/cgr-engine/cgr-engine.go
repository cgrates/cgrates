/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"reflect"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/balancer2go"
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
	cfgDir       = flag.String("config_dir", utils.CONFIG_DIR, "Configuration directory path.")
	version      = flag.Bool("version", false, "Prints the application version.")
	raterEnabled = flag.Bool("rater", false, "Enforce starting of the rater daemon overwriting config")
	schedEnabled = flag.Bool("scheduler", false, "Enforce starting of the scheduler daemon .overwriting config")
	cdrsEnabled  = flag.Bool("cdrs", false, "Enforce starting of the cdrs daemon overwriting config")
	pidFile      = flag.String("pid", "", "Write pid file")
	cpuprofile   = flag.String("cpuprofile", "", "write cpu profile to file")
	singlecpu    = flag.Bool("singlecpu", false, "Run on single CPU core")

	cfg   *config.CGRConfig
	smRpc *v1.SessionManagerV1
	err   error
)

func startCdrcs(internalCdrSChan chan *engine.CdrServer, internalRaterChan chan *engine.Responder, exitChan chan bool) {
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
			var cdrcCfg *config.CdrcConfig
			for _, cdrcCfg = range cdrcCfgs { // Take a random config out since they should be the same
				break
			}
			if cdrcCfg.Enabled == false {
				continue // Ignore not enabled
			}
			go startCdrc(internalCdrSChan, internalRaterChan, cdrcCfgs, cfg.HttpSkipTlsVerify, cdrcChildrenChan, exitChan)
		}
		cdrcInitialized = true // Initialized

	}
}

// Fires up a cdrc instance
func startCdrc(internalCdrSChan chan *engine.CdrServer, internalRaterChan chan *engine.Responder, cdrcCfgs map[string]*config.CdrcConfig, httpSkipTlsCheck bool,
	closeChan chan struct{}, exitChan chan bool) {
	var cdrsConn rpcclient.RpcClientConnection
	var cdrcCfg *config.CdrcConfig
	for _, cdrcCfg = range cdrcCfgs { // Take the first config out, does not matter which one
		break
	}
	if cdrcCfg.Cdrs == utils.INTERNAL {
		cdrsChan := <-internalCdrSChan // This will signal that the cdrs part is populated in internalRaterChan
		internalCdrSChan <- cdrsChan   // Put it back for other components
		resp := <-internalRaterChan
		cdrsConn = resp
		internalRaterChan <- resp
	} else {
		conn, err := rpcclient.NewRpcClient("tcp", cdrcCfg.Cdrs, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRC> Could not connect to CDRS via RPC: %v", err))
			exitChan <- true
			return
		}
		cdrsConn = conn
	}
	cdrc, err := cdrc.NewCdrc(cdrcCfgs, httpSkipTlsCheck, cdrsConn, closeChan, cfg.DefaultTimezone)
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

func startSmGeneric(internalSMGChan chan rpcclient.RpcClientConnection, internalRaterChan chan *engine.Responder, server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS SM-Generic service.")
	raterConn := rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST)
	cdrsConn := rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST)
	var client *rpcclient.RpcClient
	var err error
	// Connect to rater
	for _, raterCfg := range cfg.SmGenericConfig.HaRater {
		if raterCfg.Server == utils.INTERNAL {
			resp := <-internalRaterChan
			raterConn.AddClient(resp)
			internalRaterChan <- resp
		} else {
			client, err = rpcclient.NewRpcClient("tcp", raterCfg.Server, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
			if err != nil { //Connected so no need to reiterate
				utils.Logger.Crit(fmt.Sprintf("<SM-Generic> Could not connect to Rater via RPC: %v", err))
				exitChan <- true
				return
			}
			raterConn.AddClient(client)
		}
	}
	// Connect to CDRS
	if reflect.DeepEqual(cfg.SmGenericConfig.HaCdrs, cfg.SmGenericConfig.HaRater) {
		cdrsConn = raterConn
	} else if len(cfg.SmGenericConfig.HaCdrs) != 0 {
		for _, cdrsCfg := range cfg.SmGenericConfig.HaCdrs {
			if cdrsCfg.Server == utils.INTERNAL {
				resp := <-internalRaterChan
				cdrsConn.AddClient(client)
				internalRaterChan <- resp
			} else {
				client, err = rpcclient.NewRpcClient("tcp", cdrsCfg.Server, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
				if err != nil {
					utils.Logger.Crit(fmt.Sprintf("<SM-Generic> Could not connect to CDRS via RPC: %v", err))
					exitChan <- true
					return
				}
				cdrsConn.AddClient(client)
			}
		}
	}
	smg_econns := sessionmanager.NewSMGExternalConnections()
	sm := sessionmanager.NewSMGeneric(cfg, raterConn, cdrsConn, cfg.DefaultTimezone, smg_econns)
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Generic> error: %s!", err))
	}
	// Register RPC handler
	smgRpc := v1.NewSMGenericV1(sm)
	server.RpcRegister(smgRpc)
	internalSMGChan <- smgRpc
	// Register BiRpc handlers
	smgBiRpc := v1.NewSMGenericBiRpcV1(sm)
	for method, handler := range smgBiRpc.Handlers() {
		server.BijsonRegisterName(method, handler)
	}
	// Register OnConnect handlers so we can intercept connections for session disconnects
	server.BijsonRegisterOnConnect(smg_econns.OnClientConnect)
	server.BijsonRegisterOnDisconnect(smg_econns.OnClientDisconnect)
}

func startDiameterAgent(internalSMGChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS DiameterAgent service.")
	var smgConn *rpcclient.RpcClient
	var err error
	if cfg.DiameterAgentCfg().SMGeneric == utils.INTERNAL {
		smgRpc := <-internalSMGChan
		internalSMGChan <- smgRpc
		smgConn, err = rpcclient.NewRpcClient("", "", 0, 0, rpcclient.INTERNAL_RPC, smgRpc)
	} else {
		smgConn, err = rpcclient.NewRpcClient("tcp", cfg.DiameterAgentCfg().SMGeneric, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
	}
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<DiameterAgent> Could not connect to SMG: %s", err.Error()))
		exitChan <- true
		return
	}
	da, err := agents.NewDiameterAgent(cfg, smgConn)
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

func startSmFreeSWITCH(internalRaterChan chan *engine.Responder, cdrDb engine.CdrStorage, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS SM-FreeSWITCH service.")
	raterConn := rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST)
	cdrsConn := rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST)
	var client *rpcclient.RpcClient
	var err error
	// Connect to rater
	for _, raterCfg := range cfg.SmFsConfig.HaRater {
		if raterCfg.Server == utils.INTERNAL {
			resp := <-internalRaterChan
			raterConn.AddClient(resp)
			internalRaterChan <- resp
		} else {
			client, err = rpcclient.NewRpcClient("tcp", raterCfg.Server, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
			if err != nil { //Connected so no need to reiterate
				utils.Logger.Crit(fmt.Sprintf("<SM-FreeSWITCH> Could not connect to rater via RPC: %v", err))
				exitChan <- true
				return
			}
			raterConn.AddClient(client)
		}
	}
	// Connect to CDRS
	if reflect.DeepEqual(cfg.SmFsConfig.HaCdrs, cfg.SmFsConfig.HaRater) {
		cdrsConn = raterConn
	} else if len(cfg.SmFsConfig.HaCdrs) != 0 {
		for _, cdrsCfg := range cfg.SmFsConfig.HaCdrs {
			if cdrsCfg.Server == utils.INTERNAL {
				resp := <-internalRaterChan
				cdrsConn.AddClient(resp)
				internalRaterChan <- resp
			} else {
				client, err = rpcclient.NewRpcClient("tcp", cdrsCfg.Server, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
				if err != nil {
					utils.Logger.Crit(fmt.Sprintf("<SM-FreeSWITCH> Could not connect to CDRS via RPC: %v", err))
					exitChan <- true
					return
				}
				cdrsConn.AddClient(client)
			}
		}
	}
	sm := sessionmanager.NewFSSessionManager(cfg.SmFsConfig, raterConn, cdrsConn, cfg.DefaultTimezone)
	smRpc.SMs = append(smRpc.SMs, sm)
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-FreeSWITCH> error: %s!", err))
	}
	exitChan <- true
}

func startSmKamailio(internalRaterChan chan *engine.Responder, cdrDb engine.CdrStorage, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS SM-Kamailio service.")
	raterConn := rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST)
	cdrsConn := rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST)
	var client *rpcclient.RpcClient
	var err error
	// Connect to rater
	for _, raterCfg := range cfg.SmKamConfig.HaRater {
		if raterCfg.Server == utils.INTERNAL {
			resp := <-internalRaterChan
			raterConn.AddClient(resp)
			internalRaterChan <- resp
		} else {
			client, err = rpcclient.NewRpcClient("tcp", raterCfg.Server, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
			if err != nil { //Connected so no need to reiterate
				utils.Logger.Crit(fmt.Sprintf("<SM-Kamailio> Could not connect to rater via RPC: %v", err))
				exitChan <- true
				return
			}
			raterConn.AddClient(client)
		}
	}
	// Connect to CDRS
	if reflect.DeepEqual(cfg.SmKamConfig.HaCdrs, cfg.SmKamConfig.HaRater) {
		cdrsConn = raterConn
	} else if len(cfg.SmKamConfig.HaCdrs) != 0 {
		for _, cdrsCfg := range cfg.SmKamConfig.HaCdrs {
			if cdrsCfg.Server == utils.INTERNAL {
				resp := <-internalRaterChan
				cdrsConn.AddClient(resp)
				internalRaterChan <- resp
			} else {
				client, err = rpcclient.NewRpcClient("tcp", cdrsCfg.Server, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
				if err != nil {
					utils.Logger.Crit(fmt.Sprintf("<SM-Kamailio> Could not connect to CDRS via RPC: %v", err))
					exitChan <- true
					return
				}
				cdrsConn.AddClient(client)
			}
		}
	}
	sm, _ := sessionmanager.NewKamailioSessionManager(cfg.SmKamConfig, raterConn, cdrsConn, cfg.DefaultTimezone)
	smRpc.SMs = append(smRpc.SMs, sm)
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> error: %s!", err))
	}
	exitChan <- true
}

func startSmOpenSIPS(internalRaterChan chan *engine.Responder, cdrDb engine.CdrStorage, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS SM-OpenSIPS service.")
	raterConn := rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST)
	cdrsConn := rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST)
	var client *rpcclient.RpcClient
	var err error
	// Connect to rater
	for _, raterCfg := range cfg.SmOsipsConfig.HaRater {
		if raterCfg.Server == utils.INTERNAL {
			resp := <-internalRaterChan
			raterConn.AddClient(resp)
			internalRaterChan <- resp
		} else {
			client, err = rpcclient.NewRpcClient("tcp", raterCfg.Server, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
			if err != nil { //Connected so no need to reiterate
				utils.Logger.Crit(fmt.Sprintf("<SM-OpenSIPS> Could not connect to rater via RPC: %v", err))
				exitChan <- true
				return
			}
			raterConn.AddClient(client)
		}
	}
	// Connect to CDRS
	if reflect.DeepEqual(cfg.SmOsipsConfig.HaCdrs, cfg.SmOsipsConfig.HaRater) {
		cdrsConn = raterConn
	} else if len(cfg.SmOsipsConfig.HaCdrs) != 0 {
		for _, cdrsCfg := range cfg.SmOsipsConfig.HaCdrs {
			if cdrsCfg.Server == utils.INTERNAL {
				resp := <-internalRaterChan
				cdrsConn.AddClient(resp)
				internalRaterChan <- resp
			} else {
				client, err = rpcclient.NewRpcClient("tcp", cdrsCfg.Server, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
				if err != nil {
					utils.Logger.Crit(fmt.Sprintf("<SM-OpenSIPS> Could not connect to CDRS via RPC: %v", err))
					exitChan <- true
					return
				}
				cdrsConn.AddClient(client)
			}
		}
	}
	sm, _ := sessionmanager.NewOSipsSessionManager(cfg.SmOsipsConfig, cfg.Reconnects, raterConn, cdrsConn, cfg.DefaultTimezone)
	smRpc.SMs = append(smRpc.SMs, sm)
	if err := sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> error: %s!", err))
	}
	exitChan <- true
}

func startCDRS(internalCdrSChan chan *engine.CdrServer, logDb engine.LogStorage, cdrDb engine.CdrStorage,
	internalRaterChan chan *engine.Responder, internalPubSubSChan chan rpcclient.RpcClientConnection,
	internalUserSChan chan rpcclient.RpcClientConnection, internalAliaseSChan chan rpcclient.RpcClientConnection,
	internalCdrStatSChan chan rpcclient.RpcClientConnection, server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS CDRS service.")
	var err error
	var client *rpcclient.RpcClient
	// Rater connection init
	var raterConn rpcclient.RpcClientConnection
	if cfg.CDRSRater == utils.INTERNAL {
		responder := <-internalRaterChan // Wait for rater to come up before start querying
		raterConn = responder
		internalRaterChan <- responder // Put back the connection since there might be other entities waiting for it
	} else if len(cfg.CDRSRater) != 0 {
		client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSRater, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to rater: %s", err.Error()))
			exitChan <- true
			return
		}
		raterConn = client
	}
	// Pubsub connection init
	var pubSubConn rpcclient.RpcClientConnection
	if cfg.CDRSPubSub == utils.INTERNAL {
		pubSubs := <-internalPubSubSChan
		pubSubConn = pubSubs
		internalPubSubSChan <- pubSubs
	} else if len(cfg.CDRSPubSub) != 0 {
		if cfg.CDRSRater == cfg.CDRSPubSub {
			pubSubConn = client
		} else {
			client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSPubSub, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to pubsub server: %s", err.Error()))
				exitChan <- true
				return
			}
			pubSubConn = client
		}
	}
	// Users connection init
	var usersConn rpcclient.RpcClientConnection
	if cfg.CDRSUsers == utils.INTERNAL {
		userS := <-internalUserSChan
		usersConn = userS
		internalUserSChan <- userS
	} else if len(cfg.CDRSUsers) != 0 {
		if cfg.CDRSRater == cfg.CDRSUsers {
			usersConn = client
		} else {
			client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSUsers, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to users server: %s", err.Error()))
				exitChan <- true
				return
			}
			usersConn = client
		}
	}
	// Aliases connection init
	var aliasesConn rpcclient.RpcClientConnection
	if cfg.CDRSAliases == utils.INTERNAL {
		aliaseS := <-internalAliaseSChan
		aliasesConn = aliaseS
		internalAliaseSChan <- aliaseS
	} else if len(cfg.CDRSAliases) != 0 {
		if cfg.CDRSRater == cfg.CDRSAliases {
			aliasesConn = client
		} else {
			client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSAliases, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to aliases server: %s", err.Error()))
				exitChan <- true
				return
			}
			aliasesConn = client
		}
	}
	// Stats connection init
	var statsConn rpcclient.RpcClientConnection
	if cfg.CDRSStats == utils.INTERNAL {
		statS := <-internalCdrStatSChan
		statsConn = statS
		internalCdrStatSChan <- statS
	} else if len(cfg.CDRSStats) != 0 {
		if cfg.CDRSRater == cfg.CDRSStats {
			statsConn = client
		} else {
			client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSStats, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB, nil)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to stats server: %s", err.Error()))
				exitChan <- true
				return
			}
			statsConn = client
		}
	}

	cdrServer, _ := engine.NewCdrServer(cfg, cdrDb, raterConn, pubSubConn, usersConn, aliasesConn, statsConn)
	utils.Logger.Info("Registering CDRS HTTP Handlers.")
	cdrServer.RegisterHandlersToServer(server)
	utils.Logger.Info("Registering CDRS RPC service.")
	cdrSrv := v1.CdrsV1{CdrSrv: cdrServer}
	server.RpcRegister(&cdrSrv)
	server.RpcRegister(&v2.CdrsV2{CdrsV1: cdrSrv})
	// Make the cdr server available for internal communication
	responder := <-internalRaterChan // Retrieve again the responder
	responder.CdrSrv = cdrServer     // Attach connection to cdrServer in responder, so it can be used later
	internalRaterChan <- responder   // Put back the connection for the rest of the system
	internalCdrSChan <- cdrServer    // Signal that cdrS is operational
}

func startScheduler(internalSchedulerChan chan *scheduler.Scheduler, cacheDoneChan chan struct{}, ratingDb engine.RatingStorage, exitChan chan bool) {
	// Wait for cache to load data before starting
	cacheDone := <-cacheDoneChan
	cacheDoneChan <- cacheDone
	utils.Logger.Info("Starting CGRateS Scheduler.")
	sched := scheduler.NewScheduler()
	go reloadSchedulerSingnalHandler(sched, ratingDb)
	time.Sleep(1)
	internalSchedulerChan <- sched
	sched.LoadActionPlans(ratingDb)
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
	if err := accountDb.CacheAccountingPrefixes(utils.ALIASES_PREFIX); err != nil {
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

func startRpc(server *utils.Server, internalRaterChan chan *engine.Responder,
	internalCdrSChan chan *engine.CdrServer,
	internalCdrStatSChan chan rpcclient.RpcClientConnection,
	internalHistorySChan chan rpcclient.RpcClientConnection,
	internalPubSubSChan chan rpcclient.RpcClientConnection,
	internalUserSChan chan rpcclient.RpcClientConnection,
	internalAliaseSChan chan rpcclient.RpcClientConnection) {
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
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	cfg, err = config.NewCGRConfigFromFolder(*cfgDir)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("Could not parse config: %s exiting!", err))
		return
	}
	config.SetCgrConfig(cfg) // Share the config object
	if *raterEnabled {
		cfg.RaterEnabled = *raterEnabled
	}
	if *schedEnabled {
		cfg.SchedulerEnabled = *schedEnabled
	}
	if *cdrsEnabled {
		cfg.CDRSEnabled = *cdrsEnabled
	}
	var ratingDb engine.RatingStorage
	var accountDb engine.AccountingStorage
	var logDb engine.LogStorage
	var loadDb engine.LoadStorage
	var cdrDb engine.CdrStorage
	if cfg.RaterEnabled || cfg.SchedulerEnabled { // Only connect to dataDb if necessary
		ratingDb, err = engine.ConfigureRatingStorage(cfg.TpDbType, cfg.TpDbHost, cfg.TpDbPort,
			cfg.TpDbName, cfg.TpDbUser, cfg.TpDbPass, cfg.DBDataEncoding)
		if err != nil { // Cannot configure getter database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		}
		defer ratingDb.Close()
		engine.SetRatingStorage(ratingDb)
		accountDb, err = engine.ConfigureAccountingStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort,
			cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding)
		if err != nil { // Cannot configure getter database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		}
		defer accountDb.Close()
		engine.SetAccountingStorage(accountDb)
	}
	if cfg.RaterEnabled || cfg.CDRSEnabled || cfg.SchedulerEnabled { // Only connect to storDb if necessary
		logDb, err = engine.ConfigureLogStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort,
			cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding, cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns)
		if err != nil { // Cannot configure logger database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure logger database: %s exiting!", err))
			return
		}
		defer logDb.Close()
		engine.SetStorageLogger(logDb)
		// loadDb,cdrDb and logDb are all mapped on the same stordb storage
		loadDb = logDb.(engine.LoadStorage)
		cdrDb = logDb.(engine.CdrStorage)
		engine.SetCdrStorage(cdrDb)
	}

	engine.SetRoundingDecimals(cfg.RoundingDecimals)
	stopHandled := false

	// Rpc/http server
	server := new(utils.Server)

	// Async starts here, will follow cgrates.json start order
	exitChan := make(chan bool)

	// Define internal connections via channels
	internalBalancerChan := make(chan *balancer2go.Balancer, 1)
	internalRaterChan := make(chan *engine.Responder, 1)
	cacheDoneChan := make(chan struct{}, 1)
	internalSchedulerChan := make(chan *scheduler.Scheduler, 1)
	internalCdrSChan := make(chan *engine.CdrServer, 1)
	internalCdrStatSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalHistorySChan := make(chan rpcclient.RpcClientConnection, 1)
	internalPubSubSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalUserSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAliaseSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalSMGChan := make(chan rpcclient.RpcClientConnection, 1)
	// Start balancer service
	if cfg.BalancerEnabled {
		go startBalancer(internalBalancerChan, &stopHandled, exitChan) // Not really needed async here but to cope with uniformity
	}

	// Start rater service
	if cfg.RaterEnabled {
		go startRater(internalRaterChan, cacheDoneChan, internalBalancerChan, internalSchedulerChan, internalCdrStatSChan, internalHistorySChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan,
			server, ratingDb, accountDb, loadDb, cdrDb, logDb, &stopHandled, exitChan)
	}

	// Start Scheduler
	if cfg.SchedulerEnabled {
		go startScheduler(internalSchedulerChan, cacheDoneChan, ratingDb, exitChan)
	}

	// Start CDR Server
	if cfg.CDRSEnabled {
		go startCDRS(internalCdrSChan, logDb, cdrDb, internalRaterChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan, internalCdrStatSChan, server, exitChan)
	}

	// Start CDR Stats server
	if cfg.CDRStatsEnabled {
		go startCdrStats(internalCdrStatSChan, ratingDb, accountDb, server)
	}

	// Start CDRC components if necessary
	go startCdrcs(internalCdrSChan, internalRaterChan, exitChan)

	// Start SM-Generic
	if cfg.SmGenericConfig.Enabled {
		go startSmGeneric(internalSMGChan, internalRaterChan, server, exitChan)
	}
	// Start SM-FreeSWITCH
	if cfg.SmFsConfig.Enabled {
		go startSmFreeSWITCH(internalRaterChan, cdrDb, exitChan)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler(exitChan)
	}

	// Start SM-Kamailio
	if cfg.SmKamConfig.Enabled {
		go startSmKamailio(internalRaterChan, cdrDb, exitChan)
	}

	// Start SM-OpenSIPS
	if cfg.SmOsipsConfig.Enabled {
		go startSmOpenSIPS(internalRaterChan, cdrDb, exitChan)
	}

	// Register session manager service // FixMe: make sure this is thread safe
	if cfg.SmGenericConfig.Enabled || cfg.SmFsConfig.Enabled || cfg.SmKamConfig.Enabled || cfg.SmOsipsConfig.Enabled { // Register SessionManagerV1 service
		smRpc = new(v1.SessionManagerV1)
		server.RpcRegister(smRpc)
	}

	if cfg.DiameterAgentCfg().Enabled {
		go startDiameterAgent(internalSMGChan, exitChan)
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

	// Serve rpc connections
	go startRpc(server, internalRaterChan, internalCdrSChan, internalCdrStatSChan, internalHistorySChan,
		internalPubSubSChan, internalUserSChan, internalAliaseSChan)
	<-exitChan

	if *pidFile != "" {
		if err := os.Remove(*pidFile); err != nil {
			utils.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	utils.Logger.Info("Stopped all components. CGRateS shutdown!")
}
