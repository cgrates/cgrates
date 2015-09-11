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
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

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
	sms   []sessionmanager.SessionManager
	smRpc *v1.SessionManagerV1
	err   error
)

func startCdrcs(internalCdrSChan chan *engine.CdrServer, internalRaterChan chan *engine.Responder, exitChan chan bool) {
	cfgReloadsChan := cfg.GetConfigReloadsItem(utils.CDRC)
	for {
		for _, cdrcCfgs := range cfg.CdrcProfiles {
			var cdrcCfg *config.CdrcConfig
			for _, cdrcCfg = range cdrcCfgs { // Take a random config out since they should be the same
				break
			}
			if cdrcCfg.Enabled == false {
				continue // Ignore not enabled
			}
			go startCdrc(internalCdrSChan, internalRaterChan, cdrcCfgs, cfg.HttpSkipTlsVerify, cfgReloadsChan, exitChan)
		}
		select {
		case <-exitChan: // Stop forking CDRCs
			break
		case <-cfgReloadsChan: // Will release another start as soon as channel will be closed
			engine.Logger.Info("Reloading CDRC configuration")
			cfgReloadsChan = make(chan struct{}) // Schedule waiting again for another reload
			cfg.SetConfigReloadsItem(utils.CDRC, cfgReloadsChan)
			continue
		}
	}
}

// Fires up a cdrc instance
func startCdrc(internalCdrSChan chan *engine.CdrServer, internalRaterChan chan *engine.Responder, cdrcCfgs map[string]*config.CdrcConfig, httpSkipTlsCheck bool,
	closeChan chan struct{}, exitChan chan bool) {
	var cdrsConn engine.Connector
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
		conn, err := rpcclient.NewRpcClient("tcp", cdrcCfg.Cdrs, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<CDRC> Could not connect to CDRS via RPC: %v", err))
			exitChan <- true
			return
		}
		cdrsConn = &engine.RPCClientConnector{Client: conn}
	}
	cdrc, err := cdrc.NewCdrc(cdrcCfgs, httpSkipTlsCheck, cdrsConn, closeChan, cfg.DefaultTimezone)
	if err != nil {
		engine.Logger.Crit(fmt.Sprintf("Cdrc config parsing error: %s", err.Error()))
		exitChan <- true
		return
	}
	if err := cdrc.Run(); err != nil {
		engine.Logger.Crit(fmt.Sprintf("Cdrc run error: %s", err.Error()))
		exitChan <- true // If run stopped, something is bad, stop the application
	}
}

func startSmFreeSWITCH(internalRaterChan chan *engine.Responder, cdrDb engine.CdrStorage, exitChan chan bool) {
	engine.Logger.Info("Starting CGRateS SM-FreeSWITCH service.")
	var raterConn, cdrsConn engine.Connector
	var client *rpcclient.RpcClient
	if cfg.SmFsConfig.Rater == utils.INTERNAL {
		resp := <-internalRaterChan
		raterConn = resp
		internalRaterChan <- resp
	} else {
		var err error
		client, err = rpcclient.NewRpcClient("tcp", cfg.SmFsConfig.Rater, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
		if err != nil { //Connected so no need to reiterate
			engine.Logger.Crit(fmt.Sprintf("<SM-FreeSWITCH> Could not connect to rater via RPC: %v", err))
			exitChan <- true
			return
		}
		raterConn = &engine.RPCClientConnector{Client: client}
	}
	if cfg.SmFsConfig.Cdrs == cfg.SmFsConfig.Rater {
		cdrsConn = raterConn
	} else if cfg.SmFsConfig.Cdrs == utils.INTERNAL {
		resp := <-internalRaterChan
		cdrsConn = resp
		internalRaterChan <- resp
	} else if len(cfg.SmFsConfig.Cdrs) != 0 {
		client, err = rpcclient.NewRpcClient("tcp", cfg.SmFsConfig.Cdrs, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SM-FreeSWITCH> Could not connect to CDRS via RPC: %v", err))
			exitChan <- true
			return
		}
		cdrsConn = &engine.RPCClientConnector{Client: client}
	}
	sm := sessionmanager.NewFSSessionManager(cfg.SmFsConfig, raterConn, cdrsConn, cfg.DefaultTimezone)
	sms = append(sms, sm)
	smRpc.SMs = append(smRpc.SMs, sm)
	if err = sm.Connect(); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SessionManager> error: %s!", err))
	}
	exitChan <- true
}

func startSmKamailio(internalRaterChan chan *engine.Responder, cdrDb engine.CdrStorage, exitChan chan bool) {
	engine.Logger.Info("Starting CGRateS SM-Kamailio service.")
	var raterConn, cdrsConn engine.Connector
	var client *rpcclient.RpcClient
	if cfg.SmKamConfig.Rater == utils.INTERNAL {
		resp := <-internalRaterChan
		raterConn = resp
		internalRaterChan <- resp
	} else {
		var err error
		client, err = rpcclient.NewRpcClient("tcp", cfg.SmKamConfig.Rater, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SessionManager> Could not connect to rater: %v", err))
			exitChan <- true
		}
		raterConn = &engine.RPCClientConnector{Client: client}
	}
	if cfg.SmKamConfig.Cdrs == cfg.SmKamConfig.Rater {
		cdrsConn = raterConn
	} else if cfg.SmKamConfig.Cdrs == utils.INTERNAL {
		resp := <-internalRaterChan
		cdrsConn = resp
		internalRaterChan <- resp
	} else if len(cfg.SmKamConfig.Cdrs) != 0 {
		client, err = rpcclient.NewRpcClient("tcp", cfg.SmKamConfig.Cdrs, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SM-Kamailio> Could not connect to CDRS via RPC: %v", err))
			exitChan <- true
			return
		}
		cdrsConn = &engine.RPCClientConnector{Client: client}
	}
	sm, _ := sessionmanager.NewKamailioSessionManager(cfg.SmKamConfig, raterConn, cdrsConn, cfg.DefaultTimezone)
	sms = append(sms, sm)
	smRpc.SMs = append(smRpc.SMs, sm)
	if err = sm.Connect(); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SessionManager> error: %s!", err))
	}
	exitChan <- true
}

func startSmOpenSIPS(internalRaterChan chan *engine.Responder, cdrDb engine.CdrStorage, exitChan chan bool) {
	engine.Logger.Info("Starting CGRateS SM-OpenSIPS service.")
	var raterConn, cdrsConn engine.Connector
	var client *rpcclient.RpcClient
	if cfg.SmOsipsConfig.Rater == utils.INTERNAL {
		resp := <-internalRaterChan
		raterConn = resp
		internalRaterChan <- resp
	} else {
		var err error
		client, err = rpcclient.NewRpcClient("tcp", cfg.SmOsipsConfig.Rater, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SessionManager> Could not connect to rater: %v", err))
			exitChan <- true
		}
		raterConn = &engine.RPCClientConnector{Client: client}
	}
	if cfg.SmOsipsConfig.Cdrs == cfg.SmOsipsConfig.Rater {
		cdrsConn = raterConn
	} else if cfg.SmOsipsConfig.Cdrs == utils.INTERNAL {
		resp := <-internalRaterChan
		cdrsConn = resp
		internalRaterChan <- resp
	} else if len(cfg.SmOsipsConfig.Cdrs) != 0 {
		client, err = rpcclient.NewRpcClient("tcp", cfg.SmOsipsConfig.Cdrs, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SM-OpenSIPS> Could not connect to CDRS via RPC: %v", err))
			exitChan <- true
			return
		}
		cdrsConn = &engine.RPCClientConnector{Client: client}
	}
	sm, _ := sessionmanager.NewOSipsSessionManager(cfg.SmOsipsConfig, cfg.Reconnects, raterConn, cdrsConn, cfg.DefaultTimezone)
	sms = append(sms, sm)
	smRpc.SMs = append(smRpc.SMs, sm)
	if err := sm.Connect(); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> error: %s!", err))
	}
	exitChan <- true
}

func startCDRS(internalCdrSChan chan *engine.CdrServer, logDb engine.LogStorage, cdrDb engine.CdrStorage,
	internalRaterChan chan *engine.Responder, internalPubSubSChan chan engine.PublisherSubscriber,
	internalUserSChan chan engine.UserService, internalAliaseSChan chan engine.AliasService,
	internalCdrStatSChan chan engine.StatsInterface, server *engine.Server, exitChan chan bool) {
	engine.Logger.Info("Starting CGRateS CDRS service.")
	var err error
	var client *rpcclient.RpcClient
	// Rater connection init
	var raterConn engine.Connector
	if cfg.CDRSRater == utils.INTERNAL {
		responder := <-internalRaterChan // Wait for rater to come up before start querying
		raterConn = responder
		internalRaterChan <- responder // Put back the connection since there might be other entities waiting for it
	} else if len(cfg.CDRSRater) != 0 {
		client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSRater, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to rater: %s", err.Error()))
			exitChan <- true
			return
		}
		raterConn = &engine.RPCClientConnector{Client: client}
	}
	// Pubsub connection init
	var pubSubConn engine.PublisherSubscriber
	if cfg.CDRSPubSub == utils.INTERNAL {
		pubSubs := <-internalPubSubSChan
		pubSubConn = pubSubs
		internalPubSubSChan <- pubSubs
	} else if len(cfg.CDRSPubSub) != 0 {
		if cfg.CDRSRater == cfg.CDRSPubSub {
			pubSubConn = &engine.ProxyPubSub{Client: client}
		} else {
			client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSPubSub, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
			if err != nil {
				engine.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to pubsub server: %s", err.Error()))
				exitChan <- true
				return
			}
			pubSubConn = &engine.ProxyPubSub{Client: client}
		}
	}
	// Users connection init
	var usersConn engine.UserService
	if cfg.CDRSUsers == utils.INTERNAL {
		userS := <-internalUserSChan
		usersConn = userS
		internalUserSChan <- userS
	} else if len(cfg.CDRSUsers) != 0 {
		if cfg.CDRSRater == cfg.CDRSUsers {
			usersConn = &engine.ProxyUserService{Client: client}
		} else {
			client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSUsers, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
			if err != nil {
				engine.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to users server: %s", err.Error()))
				exitChan <- true
				return
			}
			usersConn = &engine.ProxyUserService{Client: client}
		}
	}
	// Aliases connection init
	var aliasesConn engine.AliasService
	if cfg.CDRSAliases == utils.INTERNAL {
		aliaseS := <-internalAliaseSChan
		aliasesConn = aliaseS
		internalAliaseSChan <- aliaseS
	} else if len(cfg.CDRSAliases) != 0 {
		if cfg.CDRSRater == cfg.CDRSAliases {
			aliasesConn = &engine.ProxyAliasService{Client: client}
		} else {
			client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSAliases, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
			if err != nil {
				engine.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to aliases server: %s", err.Error()))
				exitChan <- true
				return
			}
			aliasesConn = &engine.ProxyAliasService{Client: client}
		}
	}
	// Stats connection init
	var statsConn engine.StatsInterface
	if cfg.CDRSStats == utils.INTERNAL {
		statS := <-internalCdrStatSChan
		statsConn = statS
		internalCdrStatSChan <- statS
	} else if len(cfg.CDRSStats) != 0 {
		if cfg.CDRSRater == cfg.CDRSStats {
			statsConn = &engine.ProxyStats{Client: client}
		} else {
			client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSStats, cfg.ConnectAttempts, cfg.Reconnects, utils.GOB)
			if err != nil {
				engine.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to stats server: %s", err.Error()))
				exitChan <- true
				return
			}
			statsConn = &engine.ProxyStats{Client: client}
		}
	}

	cdrServer, _ := engine.NewCdrServer(cfg, cdrDb, raterConn, pubSubConn, usersConn, aliasesConn, statsConn)
	engine.Logger.Info("Registering CDRS HTTP Handlers.")
	cdrServer.RegisterHanlersToServer(server)
	engine.Logger.Info("Registering CDRS RPC service.")
	cdrSrv := v1.CdrsV1{CdrSrv: cdrServer}
	server.RpcRegister(&cdrSrv)
	server.RpcRegister(&v2.CdrsV2{CdrsV1: cdrSrv})
	// Make the cdr server available for internal communication
	responder := <-internalRaterChan // Retrieve again the responder
	responder.CdrSrv = cdrServer     // Attach connection to cdrServer in responder, so it can be used later
	internalRaterChan <- responder   // Put back the connection for the rest of the system
	internalCdrSChan <- cdrServer    // Signal that cdrS is operational
}

func startScheduler(internalSchedulerChan chan *scheduler.Scheduler, ratingDb engine.RatingStorage, exitChan chan bool) {
	engine.Logger.Info("Starting CGRateS Scheduler.")
	sched := scheduler.NewScheduler()
	go reloadSchedulerSingnalHandler(sched, ratingDb)
	time.Sleep(1)
	internalSchedulerChan <- sched
	sched.LoadActionPlans(ratingDb)
	sched.Loop()
	exitChan <- true // Should not get out of loop though
}

func startCdrStats(internalCdrStatSChan chan engine.StatsInterface, ratingDb engine.RatingStorage, accountDb engine.AccountingStorage, server *engine.Server) {
	cdrStats := engine.NewStats(ratingDb, accountDb, cfg.CDRStatsSaveInterval)
	server.RpcRegister(cdrStats)
	server.RpcRegister(&v1.CDRStatsV1{CdrStats: cdrStats}) // Public APIs
	internalCdrStatSChan <- cdrStats
}

func startHistoryServer(internalHistorySChan chan history.Scribe, server *engine.Server, exitChan chan bool) {
	scribeServer, err := history.NewFileScribe(cfg.HistoryDir, cfg.HistorySaveInterval)
	if err != nil {
		engine.Logger.Crit(fmt.Sprintf("<HistoryServer> Could not start, error: %s", err.Error()))
		exitChan <- true
	}
	server.RpcRegisterName("ScribeV1", scribeServer)
	internalHistorySChan <- scribeServer
}

func startPubSubServer(internalPubSubSChan chan engine.PublisherSubscriber, accountDb engine.AccountingStorage, server *engine.Server) {
	pubSubServer := engine.NewPubSub(accountDb, cfg.HttpSkipTlsVerify)
	server.RpcRegisterName("PubSubV1", pubSubServer)
	internalPubSubSChan <- pubSubServer
}

// ToDo: Make sure we are caching before starting this one
func startAliasesServer(internalAliaseSChan chan engine.AliasService, accountDb engine.AccountingStorage, server *engine.Server, exitChan chan bool) {
	aliasesServer := engine.NewAliasHandler(accountDb)
	server.RpcRegisterName("AliasesV1", aliasesServer)
	if err := accountDb.CacheAccountingPrefixes(utils.ALIASES_PREFIX); err != nil {
		engine.Logger.Crit(fmt.Sprintf("<Aliases> Could not start, error: %s", err.Error()))
		exitChan <- true
		return
	}
	internalAliaseSChan <- aliasesServer
}

func startUsersServer(internalUserSChan chan engine.UserService, accountDb engine.AccountingStorage, server *engine.Server, exitChan chan bool) {
	userServer, err := engine.NewUserMap(accountDb, cfg.UserServerIndexes)
	if err != nil {
		engine.Logger.Crit(fmt.Sprintf("<UsersService> Could not start, error: %s", err.Error()))
		exitChan <- true
		return
	}
	server.RpcRegisterName("UsersV1", userServer)
	internalUserSChan <- userServer
}

func startRpc(server *engine.Server, internalRaterChan chan *engine.Responder,
	internalCdrSChan chan *engine.CdrServer,
	internalCdrStatSChan chan engine.StatsInterface,
	internalHistorySChan chan history.Scribe,
	internalPubSubSChan chan engine.PublisherSubscriber,
	internalUserSChan chan engine.UserService,
	internalAliaseSChan chan engine.AliasService) {
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
	engine.Logger.Info(*pidFile)
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
	if !*singlecpu {
		runtime.GOMAXPROCS(runtime.NumCPU()) // For now it slows down computing due to CPU management, to be reviewed in future Go releases
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
		engine.Logger.Crit(fmt.Sprintf("Could not parse config: %s exiting!", err))
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
			engine.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		}
		defer ratingDb.Close()
		engine.SetRatingStorage(ratingDb)
		accountDb, err = engine.ConfigureAccountingStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort,
			cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding)
		if err != nil { // Cannot configure getter database, show stopper
			engine.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		}
		defer accountDb.Close()
		engine.SetAccountingStorage(accountDb)
	}
	if cfg.RaterEnabled || cfg.CDRSEnabled || cfg.SchedulerEnabled { // Only connect to storDb if necessary
		logDb, err = engine.ConfigureLogStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort,
			cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding, cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns)
		if err != nil { // Cannot configure logger database, show stopper
			engine.Logger.Crit(fmt.Sprintf("Could not configure logger database: %s exiting!", err))
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
	server := new(engine.Server)

	// Async starts here, will follow cgrates.json start order
	exitChan := make(chan bool)

	// Define internal connections via channels
	internalBalancerChan := make(chan *balancer2go.Balancer, 1)
	internalRaterChan := make(chan *engine.Responder, 1)
	internalSchedulerChan := make(chan *scheduler.Scheduler, 1)
	internalCdrSChan := make(chan *engine.CdrServer, 1)
	internalCdrStatSChan := make(chan engine.StatsInterface, 1)
	internalHistorySChan := make(chan history.Scribe, 1)
	internalPubSubSChan := make(chan engine.PublisherSubscriber, 1)
	internalUserSChan := make(chan engine.UserService, 1)
	internalAliaseSChan := make(chan engine.AliasService, 1)

	// Start balancer service
	if cfg.BalancerEnabled {
		go startBalancer(internalBalancerChan, &stopHandled, exitChan) // Not really needed async here but to cope with uniformity
	}

	// Start rater service
	if cfg.RaterEnabled {
		go startRater(internalRaterChan, internalBalancerChan, internalSchedulerChan, internalCdrStatSChan, internalHistorySChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan,
			server, ratingDb, accountDb, loadDb, cdrDb, logDb, &stopHandled, exitChan)
	}

	// Start Scheduler
	if cfg.SchedulerEnabled {
		go startScheduler(internalSchedulerChan, ratingDb, exitChan)
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
	if cfg.SmFsConfig.Enabled || cfg.SmKamConfig.Enabled || cfg.SmOsipsConfig.Enabled { // Register SessionManagerV1 service
		smRpc = new(v1.SessionManagerV1)
		server.RpcRegister(smRpc)
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
			engine.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	engine.Logger.Info("Stopped all components. CGRateS shutdown!")
}
