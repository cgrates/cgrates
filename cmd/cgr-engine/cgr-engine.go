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
	"os/signal"
	"path"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/analyzers"
	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/cdrc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
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
	cpuProfDir        = flag.String("cpuprof_dir", "", "write cpu profile to files")
	memProfDir        = flag.String("memprof_dir", "", "write memory profile to file")
	memProfInterval   = flag.Duration("memprof_interval", 5*time.Second, "Time betwen memory profile saves")
	memProfNrFiles    = flag.Int("memprof_nrfiles", 1, "Number of memory profile to write")
	scheduledShutdown = flag.String("scheduled_shutdown", "", "shutdown the engine after this duration")
	singlecpu         = flag.Bool("singlecpu", false, "Run on single CPU core")
	syslogger         = flag.String("logger", "", "logger <*syslog|*stdout>")
	nodeID            = flag.String("node_id", "", "The node ID of the engine")
	logLevel          = flag.Int("log_level", -1, "Log level (0-emergency to 7-debug)")

	cfg *config.CGRConfig
)

func startCdrcs(internalCdrSChan, internalRaterChan chan rpcclient.RpcClientConnection,
	exitChan chan bool, filterSChan chan *engine.FilterS) {
	filterS := <-filterSChan
	filterSChan <- filterS
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
			var enabledCfgs []*config.CdrcCfg
			for _, cdrcCfg := range cdrcCfgs { // Take a random config out since they should be the same
				if cdrcCfg.Enabled {
					enabledCfgs = append(enabledCfgs, cdrcCfg)
				}
			}
			if len(enabledCfgs) != 0 {
				go startCdrc(internalCdrSChan, internalRaterChan, enabledCfgs,
					cfg.GeneralCfg().HttpSkipTlsVerify, cdrcChildrenChan,
					exitChan, filterSChan)
			} else {
				utils.Logger.Info("<CDRC> No enabled CDRC clients")
			}
		}
		cdrcInitialized = true // Initialized
	}
}

// Fires up a cdrc instance
func startCdrc(internalCdrSChan, internalRaterChan chan rpcclient.RpcClientConnection, cdrcCfgs []*config.CdrcCfg, httpSkipTlsCheck bool,
	closeChan chan struct{}, exitChan chan bool, filterSChan chan *engine.FilterS) {
	filterS := <-filterSChan
	filterSChan <- filterS
	var cdrcCfg *config.CdrcCfg
	for _, cdrcCfg = range cdrcCfgs { // Take the first config out, does not matter which one
		break
	}
	cdrsConn, err := engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.TlsCfg().ClientKey,
		cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
		cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
		cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
		cdrcCfg.CdrsConns, internalCdrSChan, cfg.GeneralCfg().InternalTtl)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRC> Could not connect to CDRS via RPC: %s", err.Error()))
		exitChan <- true
		return
	}
	cdrc, err := cdrc.NewCdrc(cdrcCfgs, httpSkipTlsCheck, cdrsConn, closeChan,
		cfg.GeneralCfg().DefaultTimezone, cfg.GeneralCfg().RoundingDecimals,
		filterS)
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

func startSessionS(internalSMGChan, internalRaterChan, internalResourceSChan, internalThresholdSChan,
	internalStatSChan, internalSupplierSChan, internalAttrSChan,
	internalCDRSChan, internalChargerSChan chan rpcclient.RpcClientConnection, server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS Session service.")
	var err error
	var ralsConns, resSConns, threshSConns, statSConns, suplSConns, attrSConns, cdrsConn, chargerSConn *rpcclient.RpcClientPool
	if len(cfg.SessionSCfg().ChargerSConns) != 0 {
		chargerSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey, cfg.TlsCfg().ClientCerificate,
			cfg.TlsCfg().CaCertificate, cfg.GeneralCfg().ConnectAttempts,
			cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().ConnectTimeout,
			cfg.GeneralCfg().ReplyTimeout, cfg.SessionSCfg().ChargerSConns,
			internalChargerSChan, cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.SessionS, utils.ChargerS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SessionSCfg().RALsConns) != 0 {
		ralsConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey, cfg.TlsCfg().ClientCerificate,
			cfg.TlsCfg().CaCertificate, cfg.GeneralCfg().ConnectAttempts,
			cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().ConnectTimeout,
			cfg.GeneralCfg().ReplyTimeout, cfg.SessionSCfg().RALsConns,
			internalRaterChan, cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to RALs: %s",
				utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SessionSCfg().ResSConns) != 0 {
		resSConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.SessionSCfg().ResSConns, internalResourceSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to ResourceS: %s",
				utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SessionSCfg().ThreshSConns) != 0 {
		threshSConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.SessionSCfg().ThreshSConns, internalThresholdSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to ThresholdS: %s",
				utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SessionSCfg().StatSConns) != 0 {
		statSConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.SessionSCfg().StatSConns, internalStatSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to StatS: %s",
				utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SessionSCfg().SupplSConns) != 0 {
		suplSConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.SessionSCfg().SupplSConns, internalSupplierSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to SupplierS: %s",
				utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SessionSCfg().AttrSConns) != 0 {
		attrSConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.SessionSCfg().AttrSConns, internalAttrSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to AttributeS: %s",
				utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SessionSCfg().CDRsConns) != 0 {
		cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.SessionSCfg().CDRsConns, internalCDRSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to RALs: %s",
				utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	smgReplConns, err := sessions.NewSessionReplicationConns(cfg.SessionSCfg().SessionReplicationConns,
		cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().ConnectTimeout,
		cfg.GeneralCfg().ReplyTimeout)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to SMGReplicationConnection error: <%s>",
			utils.SessionS, err.Error()))
		exitChan <- true
		return
	}
	sm := sessions.NewSMGeneric(cfg, ralsConns, resSConns, threshSConns,
		statSConns, suplSConns, attrSConns, cdrsConn, chargerSConn,
		smgReplConns, cfg.GeneralCfg().DefaultTimezone)
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
		server.ServeBiJSON(cfg.SessionSCfg().ListenBijson, sm.OnBiJSONConnect, sm.OnBiJSONDisconnect)
		exitChan <- true
	}
}

func startAsteriskAgent(internalSMGChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	utils.Logger.Info("Starting Asterisk agent")
	smgRpcConn := <-internalSMGChan
	internalSMGChan <- smgRpcConn
	birpcClnt := utils.NewBiRPCInternalClient(smgRpcConn.(*sessions.SMGeneric))
	var reply string
	for connIdx := range cfg.AsteriskAgentCfg().AsteriskConns { // Instantiate connections towards asterisk servers
		sma, err := agents.NewAsteriskAgent(cfg, connIdx, birpcClnt)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.AsteriskAgent, err))
			exitChan <- true
			return
		}
		if err := birpcClnt.Call(utils.SessionSv1RegisterInternalBiJSONConn, "", &reply); err != nil { // for session sync
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.AsteriskAgent, err))
		}
		if err = sma.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> runtime error: %s!", utils.AsteriskAgent, err))
		}
	}
	exitChan <- true
}

func startDiameterAgent(internalSMGChan chan rpcclient.RpcClientConnection,
	exitChan chan bool, filterSChan chan *engine.FilterS) {
	var err error
	utils.Logger.Info("Starting CGRateS DiameterAgent service")
	filterS := <-filterSChan
	filterSChan <- filterS
	var smgConn *rpcclient.RpcClientPool
	if len(cfg.DiameterAgentCfg().SessionSConns) != 0 {
		smgConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DiameterAgentCfg().SessionSConns, internalSMGChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DiameterAgent, utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	da, err := agents.NewDiameterAgent(cfg, filterS, smgConn)
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

func startRadiusAgent(internalSMGChan chan rpcclient.RpcClientConnection, exitChan chan bool,
	filterSChan chan *engine.FilterS) {
	filterS := <-filterSChan
	filterSChan <- filterS
	utils.Logger.Info("Starting CGRateS RadiusAgent service")
	var err error
	var smgConn *rpcclient.RpcClientPool
	if len(cfg.RadiusAgentCfg().SessionSConns) != 0 {
		smgConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.RadiusAgentCfg().SessionSConns, internalSMGChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<RadiusAgent> Could not connect to SMG: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	ra, err := agents.NewRadiusAgent(cfg, filterS, smgConn)
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
	birpcClnt := utils.NewBiRPCInternalClient(smgRpcConn.(*sessions.SMGeneric))
	sm := agents.NewFSsessions(cfg.FsAgentCfg(), birpcClnt, cfg.GeneralCfg().DefaultTimezone)
	var reply string
	if err = birpcClnt.Call(utils.SessionSv1RegisterInternalBiJSONConn, "", &reply); err != nil { // for session sync
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
	}
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
	}

	exitChan <- true
}

func startKamAgent(internalSMGChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	var err error
	utils.Logger.Info("Starting Kamailio agent")
	smgRpcConn := <-internalSMGChan
	internalSMGChan <- smgRpcConn
	birpcClnt := utils.NewBiRPCInternalClient(smgRpcConn.(*sessions.SMGeneric))
	ka := agents.NewKamailioAgent(cfg.KamAgentCfg(), birpcClnt,
		utils.FirstNonEmpty(cfg.KamAgentCfg().Timezone, cfg.GeneralCfg().DefaultTimezone))
	var reply string
	if err = birpcClnt.Call(utils.SessionSv1RegisterInternalBiJSONConn, "", &reply); err != nil { // for session sync
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.KamailioAgent, err))
	}
	if err = ka.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, err))
	}
	exitChan <- true
}

func startHTTPAgent(internalSMGChan chan rpcclient.RpcClientConnection,
	exitChan chan bool, server *utils.Server,
	filterSChan chan *engine.FilterS, dfltTenant string) {
	filterS := <-filterSChan
	filterSChan <- filterS
	utils.Logger.Info("Starting HTTP agent")
	var err error
	for _, agntCfg := range cfg.HttpAgentCfg() {
		var sSConn *rpcclient.RpcClientPool
		if len(agntCfg.SessionSConns) != 0 {
			sSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.TlsCfg().ClientKey,
				cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
				cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
				cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
				agntCfg.SessionSConns, internalSMGChan,
				cfg.GeneralCfg().InternalTtl)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<%s> could not connect to %s, error: %s",
					utils.HTTPAgent, utils.SessionS, err.Error()))
				exitChan <- true
				return
			}
		}
		server.RegisterHttpHandler(agntCfg.Url,
			agents.NewHTTPAgent(sSConn, filterS, dfltTenant, agntCfg.RequestPayload,
				agntCfg.ReplyPayload, agntCfg.RequestProcessors))
	}
}

func startCDRS(internalCdrSChan chan rpcclient.RpcClientConnection,
	cdrDb engine.CdrStorage, dm *engine.DataManager,
	internalRaterChan, internalPubSubSChan, internalAttributeSChan,
	internalUserSChan, internalAliaseSChan,
	internalThresholdSChan, internalStatSChan,
	internalChargerSChan chan rpcclient.RpcClientConnection,
	server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	filterS := <-filterSChan
	filterSChan <- filterS
	var err error
	utils.Logger.Info("Starting CGRateS CDRS service.")
	var ralConn, pubSubConn, usersConn, attrSConn, aliasesConn,
		thresholdSConn, statsConn, chargerSConn *rpcclient.RpcClientPool
	if len(cfg.CdrsCfg().CDRSChargerSConns) != 0 { // Conn pool towards RAL
		chargerSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.CdrsCfg().CDRSChargerSConns, internalChargerSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
				utils.ChargerS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CdrsCfg().CDRSRaterConns) != 0 { // Conn pool towards RAL
		ralConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.CdrsCfg().CDRSRaterConns, internalRaterChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to RAL: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CdrsCfg().CDRSPubSubSConns) != 0 { // Pubsub connection init
		pubSubConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.CdrsCfg().CDRSPubSubSConns, internalPubSubSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to PubSubSystem: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CdrsCfg().CDRSAttributeSConns) != 0 { // Users connection init
		attrSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.CdrsCfg().CDRSAttributeSConns, internalAttributeSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
				utils.AttributeS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CdrsCfg().CDRSUserSConns) != 0 { // Users connection init
		usersConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.CdrsCfg().CDRSUserSConns, internalUserSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to UserS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CdrsCfg().CDRSAliaseSConns) != 0 { // Aliases connection init
		aliasesConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.CdrsCfg().CDRSAliaseSConns, internalAliaseSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to AliaseS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CdrsCfg().CDRSThresholdSConns) != 0 { // Stats connection init
		thresholdSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.CdrsCfg().CDRSThresholdSConns, internalThresholdSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to ThresholdS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.CdrsCfg().CDRSStatSConns) != 0 { // Stats connection init
		statsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.CdrsCfg().CDRSStatSConns, internalStatSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to StatS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	cdrServer, _ := engine.NewCdrServer(cfg, cdrDb, dm, ralConn, pubSubConn,
		attrSConn, usersConn, aliasesConn,
		thresholdSConn, statsConn, chargerSConn, filterS)
	cdrServer.SetTimeToLive(cfg.GeneralCfg().ResponseCacheTTL, nil)
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

func startPubSubServer(internalPubSubSChan chan rpcclient.RpcClientConnection, dm *engine.DataManager, server *utils.Server, exitChan chan bool) {
	pubSubServer, err := engine.NewPubSub(dm, cfg.GeneralCfg().HttpSkipTlsVerify)
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
	/*loadHist, err := dm.DataDB().GetLoadHistory(1, true, utils.NonTransactional)
	if err != nil || len(loadHist) == 0 {
		utils.Logger.Info(fmt.Sprintf("could not get load history: %v (%v)", loadHist, err))
		internalAliaseSChan <- aliasesServer
		return
	}
	*/
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
func startAttributeService(internalAttributeSChan chan rpcclient.RpcClientConnection,
	cacheS *engine.CacheS, cfg *config.CGRConfig, dm *engine.DataManager,
	server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	filterS := <-filterSChan
	filterSChan <- filterS
	<-cacheS.GetPrecacheChannel(utils.CacheAttributeProfiles)

	aS, err := engine.NewAttributeService(dm, filterS,
		cfg.AttributeSCfg().StringIndexedFields,
		cfg.AttributeSCfg().PrefixIndexedFields,
		cfg.AttributeSCfg().ProcessRuns)
	if err != nil {
		utils.Logger.Crit(
			fmt.Sprintf("<%s> Could not init, error: %s",
				utils.AttributeS, err.Error()))
		exitChan <- true
		return
	}
	go func() {
		if err := aS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(
				fmt.Sprintf("<%s> Error: %s listening for packets",
					utils.AttributeS, err.Error()))
		}
		aS.Shutdown()
		exitChan <- true
		return
	}()
	aSv1 := v1.NewAttributeSv1(aS)
	server.RpcRegister(aSv1)
	internalAttributeSChan <- aSv1
}

// startChargerService fires up the ChargerS
func startChargerService(internalChargerSChan chan rpcclient.RpcClientConnection,
	cacheS *engine.CacheS, internalAttributeSChan chan rpcclient.RpcClientConnection,
	cfg *config.CGRConfig, dm *engine.DataManager,
	server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	filterS := <-filterSChan
	filterSChan <- filterS
	<-cacheS.GetPrecacheChannel(utils.CacheChargerProfiles)
	var attrSConn *rpcclient.RpcClientPool
	var err error
	if len(cfg.ChargerSCfg().AttributeSConns) != 0 { // AttributeS connection init
		attrSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.ChargerSCfg().AttributeSConns, internalAttributeSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.ChargerS, utils.AttributeS, err.Error()))
			exitChan <- true
			return
		}
	}
	cS, err := engine.NewChargerService(dm, filterS, attrSConn, cfg)
	if err != nil {
		utils.Logger.Crit(
			fmt.Sprintf("<%s> Could not init, error: %s",
				utils.ChargerS, err.Error()))
		exitChan <- true
		return
	}
	go func() {
		if err := cS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(
				fmt.Sprintf("<%s> Error: %s listening for packets",
					utils.ChargerS, err.Error()))
		}
		cS.Shutdown()
		exitChan <- true
		return
	}()
	cSv1 := v1.NewChargerSv1(cS)
	server.RpcRegister(cSv1)
	internalChargerSChan <- cSv1
}

func startResourceService(internalRsChan chan rpcclient.RpcClientConnection, cacheS *engine.CacheS,
	internalThresholdSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dm *engine.DataManager, server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	var err error
	var thdSConn *rpcclient.RpcClientPool
	filterS := <-filterSChan
	filterSChan <- filterS
	if len(cfg.ResourceSCfg().ThresholdSConns) != 0 { // Stats connection init
		thdSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.ResourceSCfg().ThresholdSConns, internalThresholdSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<ResourceS> Could not connect to ThresholdS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	<-cacheS.GetPrecacheChannel(utils.CacheResourceProfiles)
	<-cacheS.GetPrecacheChannel(utils.CacheResources)

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
func startStatService(internalStatSChan chan rpcclient.RpcClientConnection, cacheS *engine.CacheS,
	internalThresholdSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dm *engine.DataManager, server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	var err error
	var thdSConn *rpcclient.RpcClientPool
	filterS := <-filterSChan
	filterSChan <- filterS
	if len(cfg.StatSCfg().ThresholdSConns) != 0 { // Stats connection init
		thdSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.StatSCfg().ThresholdSConns, internalThresholdSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<StatS> Could not connect to ThresholdS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	<-cacheS.GetPrecacheChannel(utils.CacheStatQueueProfiles)
	<-cacheS.GetPrecacheChannel(utils.CacheStatQueues)

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
func startThresholdService(internalThresholdSChan chan rpcclient.RpcClientConnection,
	cacheS *engine.CacheS, cfg *config.CGRConfig, dm *engine.DataManager,
	server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	filterS := <-filterSChan
	filterSChan <- filterS
	<-cacheS.GetPrecacheChannel(utils.CacheThresholdProfiles)
	<-cacheS.GetPrecacheChannel(utils.CacheThresholds)

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

// startSupplierService fires up the SupplierS
func startSupplierService(internalSupplierSChan chan rpcclient.RpcClientConnection, cacheS *engine.CacheS,
	internalRsChan, internalStatSChan chan rpcclient.RpcClientConnection,
	cfg *config.CGRConfig, dm *engine.DataManager, server *utils.Server,
	exitChan chan bool, filterSChan chan *engine.FilterS,
	internalAttrSChan chan rpcclient.RpcClientConnection) {
	var err error
	filterS := <-filterSChan
	filterSChan <- filterS
	var attrSConn, resourceSConn, statSConn *rpcclient.RpcClientPool
	if len(cfg.SupplierSCfg().AttributeSConns) != 0 {
		attrSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.SupplierSCfg().AttributeSConns, internalAttrSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.SupplierS, utils.AttributeS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SupplierSCfg().StatSConns) != 0 {
		statSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.SupplierSCfg().StatSConns, internalStatSChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to StatS: %s",
				utils.SupplierS, err.Error()))
			exitChan <- true
			return
		}
	}
	<-cacheS.GetPrecacheChannel(utils.CacheSupplierProfiles)

	splS, err := engine.NewSupplierService(dm, cfg.GeneralCfg().DefaultTimezone,
		filterS, cfg.SupplierSCfg().StringIndexedFields,
		cfg.SupplierSCfg().PrefixIndexedFields, resourceSConn, statSConn, attrSConn)
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
func startFilterService(filterSChan chan *engine.FilterS, cacheS *engine.CacheS,
	internalStatSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dm *engine.DataManager, exitChan chan bool) {
	<-cacheS.GetPrecacheChannel(utils.CacheFilters)
	filterSChan <- engine.NewFilterS(cfg, internalStatSChan, dm)
}

// loaderService will start and register APIs for LoaderService if enabled
func loaderService(cacheS *engine.CacheS, cfg *config.CGRConfig,
	dm *engine.DataManager, server *utils.Server, exitChan chan bool, filterSChan chan *engine.FilterS) {
	filterS := <-filterSChan
	filterSChan <- filterS
	ldrS := loaders.NewLoaderService(dm, cfg.LoaderCfg(),
		cfg.GeneralCfg().DefaultTimezone, filterS)
	if !ldrS.Enabled() {
		return
	}
	go ldrS.ListenAndServe(exitChan)
	server.RpcRegister(v1.NewLoaderSv1(ldrS))
}

// startDispatcherService fires up the DispatcherS
func startDispatcherService(internalDispatcherSChan, internalRaterChan chan rpcclient.RpcClientConnection,
	cacheS *engine.CacheS, dm *engine.DataManager,
	server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS Dispatcher service.")
	var err error
	var ralsConns, resSConns, threshSConns, statSConns, suplSConns, attrSConns, sessionsSConns, chargerSConns *rpcclient.RpcClientPool

	cfg.DispatcherSCfg().DispatchingStrategy = strings.TrimPrefix(cfg.DispatcherSCfg().DispatchingStrategy,
		utils.Meta) // remote * from DispatchingStrategy
	if len(cfg.DispatcherSCfg().RALsConns) != 0 {
		ralsConns, err = engine.NewRPCPool(cfg.DispatcherSCfg().DispatchingStrategy,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DispatcherSCfg().RALsConns, internalRaterChan,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to RALs: %s", utils.DispatcherS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.DispatcherSCfg().ResSConns) != 0 {
		resSConns, err = engine.NewRPCPool(cfg.DispatcherSCfg().DispatchingStrategy,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DispatcherSCfg().ResSConns, nil,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to ResoruceS: %s", utils.DispatcherS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.DispatcherSCfg().ThreshSConns) != 0 {
		threshSConns, err = engine.NewRPCPool(cfg.DispatcherSCfg().DispatchingStrategy,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DispatcherSCfg().ThreshSConns, nil,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to ThresholdS: %s", utils.DispatcherS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.DispatcherSCfg().StatSConns) != 0 {
		statSConns, err = engine.NewRPCPool(cfg.DispatcherSCfg().DispatchingStrategy,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DispatcherSCfg().StatSConns, nil,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to StatQueueS: %s", utils.DispatcherS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.DispatcherSCfg().SupplSConns) != 0 {
		suplSConns, err = engine.NewRPCPool(cfg.DispatcherSCfg().DispatchingStrategy,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DispatcherSCfg().SupplSConns, nil,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to SupplierS: %s", utils.DispatcherS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.DispatcherSCfg().AttrSConns) != 0 {
		attrSConns, err = engine.NewRPCPool(cfg.DispatcherSCfg().DispatchingStrategy,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DispatcherSCfg().AttrSConns, nil,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to AttributeS: %s", utils.DispatcherS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.DispatcherSCfg().SessionSConns) != 0 {
		sessionsSConns, err = engine.NewRPCPool(cfg.DispatcherSCfg().DispatchingStrategy,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DispatcherSCfg().SessionSConns, nil,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to SessionS: %s", utils.DispatcherS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.DispatcherSCfg().ChargerSConns) != 0 {
		chargerSConns, err = engine.NewRPCPool(cfg.DispatcherSCfg().DispatchingStrategy,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DispatcherSCfg().ChargerSConns, nil,
			cfg.GeneralCfg().InternalTtl)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to ChargerS: %s", utils.DispatcherS, err.Error()))
			exitChan <- true
			return
		}
	}
	dspS, err := dispatchers.NewDispatcherService(dm, ralsConns, resSConns,
		threshSConns, statSConns, suplSConns, attrSConns, sessionsSConns, chargerSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.DispatcherS, err.Error()))
		exitChan <- true
		return
	}
	go func() {
		if err := dspS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.DispatcherS, err.Error()))
		}
		dspS.Shutdown()
		exitChan <- true
		return
	}()
	if !cfg.ThresholdSCfg().Enabled && len(cfg.DispatcherSCfg().ThreshSConns) != 0 {
		server.RpcRegisterName(utils.ThresholdSv1,
			v1.NewDispatcherThresholdSv1(dspS))
	}
	if !cfg.StatSCfg().Enabled && len(cfg.DispatcherSCfg().StatSConns) != 0 {
		server.RpcRegisterName(utils.StatSv1,
			v1.NewDispatcherStatSv1(dspS))
	}
	if !cfg.ResourceSCfg().Enabled && len(cfg.DispatcherSCfg().ResSConns) != 0 {
		server.RpcRegisterName(utils.ResourceSv1,
			v1.NewDispatcherResourceSv1(dspS))
	}
	if !cfg.SupplierSCfg().Enabled && len(cfg.DispatcherSCfg().SupplSConns) != 0 {
		server.RpcRegisterName(utils.SupplierSv1,
			v1.NewDispatcherSupplierSv1(dspS))
	}
	if !cfg.AttributeSCfg().Enabled && len(cfg.DispatcherSCfg().AttrSConns) != 0 {
		server.RpcRegisterName(utils.AttributeSv1,
			v1.NewDispatcherAttributeSv1(dspS))
	}
	if !cfg.SessionSCfg().Enabled && len(cfg.DispatcherSCfg().SessionSConns) != 0 {
		server.RpcRegisterName(utils.SessionSv1,
			v1.NewDispatcherSessionSv1(dspS))
	}
	if !cfg.ChargerSCfg().Enabled && len(cfg.DispatcherSCfg().ChargerSConns) != 0 {
		server.RpcRegisterName(utils.ChargerSv1,
			v1.NewDispatcherChargerSv1(dspS))
	}
}

// startAnalyzerService fires up the AnalyzerS
func startAnalyzerService(internalAnalyzerSChan chan rpcclient.RpcClientConnection,
	server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS Analyzer service.")
	var err error
	aS, err := analyzers.NewAnalyzerService()
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.AnalyzerS, err.Error()))
		exitChan <- true
		return
	}
	go func() {
		if err := aS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.AnalyzerS, err.Error()))
		}
		aS.Shutdown()
		exitChan <- true
		return
	}()
	aSv1 := v1.NewAnalyzerSv1(aS)
	server.RpcRegister(aSv1)
	internalAnalyzerSChan <- aSv1
}

func startRpc(server *utils.Server, internalRaterChan,
	internalCdrSChan, internalPubSubSChan, internalUserSChan,
	internalAliaseSChan, internalRsChan, internalStatSChan,
	internalAttrSChan, internalChargerSChan, internalThdSChan, internalSuplSChan,
	internalSMGChan, internalDispatcherSChan, internalAnalyzerSChan chan rpcclient.RpcClientConnection,
	exitChan chan bool) {
	select { // Any of the rpc methods will unlock listening to rpc requests
	case resp := <-internalRaterChan:
		internalRaterChan <- resp
	case cdrs := <-internalCdrSChan:
		internalCdrSChan <- cdrs
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
	case attrS := <-internalAttrSChan:
		internalAttrSChan <- attrS
	case chrgS := <-internalChargerSChan:
		internalChargerSChan <- chrgS
	case thS := <-internalThdSChan:
		internalThdSChan <- thS
	case splS := <-internalSuplSChan:
		internalSuplSChan <- splS
	case dispatcherS := <-internalDispatcherSChan:
		internalDispatcherSChan <- dispatcherS
	case analyzerS := <-internalAnalyzerSChan:
		internalAnalyzerSChan <- analyzerS
	}

	go server.ServeJSON(cfg.ListenCfg().RPCJSONListen)
	go server.ServeGOB(cfg.ListenCfg().RPCGOBListen)
	go server.ServeHTTP(
		cfg.ListenCfg().HTTPListen,
		cfg.HTTPCfg().HTTPJsonRPCURL,
		cfg.HTTPCfg().HTTPWSURL,
		cfg.HTTPCfg().HTTPUseBasicAuth,
		cfg.HTTPCfg().HTTPAuthUsers,
		exitChan,
	)
	if cfg.ListenCfg().RPCGOBTLSListen != "" {
		if cfg.TlsCfg().ServerCerificate == "" || cfg.TlsCfg().ServerKey == "" {
			utils.Logger.Warning("WARNING: missing TLS certificate/key file!")
		} else {
			go server.ServeGOBTLS(
				cfg.ListenCfg().RPCGOBTLSListen,
				cfg.TlsCfg().ServerCerificate,
				cfg.TlsCfg().ServerKey,
				cfg.TlsCfg().CaCertificate,
				cfg.TlsCfg().ServerPolicy,
				cfg.TlsCfg().ServerName,
			)
		}
	}
	if cfg.ListenCfg().RPCJSONTLSListen != "" {
		if cfg.TlsCfg().ServerCerificate == "" || cfg.TlsCfg().ServerKey == "" {
			utils.Logger.Warning("WARNING: missing TLS certificate/key file!")
		} else {
			go server.ServeJSONTLS(
				cfg.ListenCfg().RPCJSONTLSListen,
				cfg.TlsCfg().ServerCerificate,
				cfg.TlsCfg().ServerKey,
				cfg.TlsCfg().CaCertificate,
				cfg.TlsCfg().ServerPolicy,
				cfg.TlsCfg().ServerName,
			)
		}
	}
	if cfg.ListenCfg().HTTPTLSListen != "" {
		if cfg.TlsCfg().ServerCerificate == "" || cfg.TlsCfg().ServerKey == "" {
			utils.Logger.Warning("WARNING: missing TLS certificate/key file!")
		} else {
			go server.ServeHTTPTLS(
				cfg.ListenCfg().HTTPTLSListen,
				cfg.TlsCfg().ServerCerificate,
				cfg.TlsCfg().ServerKey,
				cfg.TlsCfg().CaCertificate,
				cfg.TlsCfg().ServerPolicy,
				cfg.TlsCfg().ServerName,
				cfg.HTTPCfg().HTTPJsonRPCURL,
				cfg.HTTPCfg().HTTPWSURL,
				cfg.HTTPCfg().HTTPUseBasicAuth,
				cfg.HTTPCfg().HTTPAuthUsers,
			)
		}
	}
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
	sylogger := cfg.GeneralCfg().Logger
	if *syslogger != "" { // Modify the log level if provided by command arguments
		sylogger = *syslogger
	}
	err := utils.Newlogger(sylogger, cfg.GeneralCfg().NodeID)
	if err != nil {
		return err
	}
	return nil
}

func schedCDRsConns(internalCDRSChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	var err error
	var cdrsConn *rpcclient.RpcClientPool
	cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
		cfg.TlsCfg().ClientKey,
		cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
		cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
		cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
		cfg.SchedulerCfg().CDRsConns, internalCDRSChan,
		cfg.GeneralCfg().InternalTtl)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to CDRServer: %s", utils.SchedulerS, err.Error()))
		exitChan <- true
		return
	}
	engine.SetSchedCdrsConns(cdrsConn)
}

func memProfFile(memProfPath string) bool {
	f, err := os.Create(memProfPath)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<memProfile>could not create memory profile file: %s", err))
		return false
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<memProfile>could not write memory profile: %s", err))
		f.Close()
		return false
	}
	f.Close()
	return true
}

func memProfiling(memProfDir string, interval time.Duration, nrFiles int, exitChan chan bool) {
	for i := 1; ; i++ {
		time.Sleep(interval)
		memPath := path.Join(memProfDir, fmt.Sprintf("mem%v.prof", i))
		if !memProfFile(memPath) {
			exitChan <- true
		}
		if i%nrFiles == 0 {
			i = 0 // reset the counting
		}
	}
}

func cpuProfiling(cpuProfDir string, exitChan chan bool, stopChan, doneChan chan struct{}) {
	cpuPath := path.Join(cpuProfDir, "cpu.prof")
	f, err := os.Create(cpuPath)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<cpuProfiling>could not create cpu profile file: %s", err))
		exitChan <- true
		return
	}
	pprof.StartCPUProfile(f)
	<-stopChan
	pprof.StopCPUProfile()
	f.Close()
	doneChan <- struct{}{}
}

func shutdownSingnalHandler(exitChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-c
	exitChan <- true
}

func main() {
	flag.Parse()
	vers := utils.GetCGRVersion()
	if *version {
		fmt.Println(vers)
		return
	}
	if *pidFile != "" {
		writePid()
	}
	if *singlecpu {
		runtime.GOMAXPROCS(1) // Having multiple cpus may slow down computing due to CPU management, to be reviewed in future Go releases
	}

	exitChan := make(chan bool)
	go shutdownSingnalHandler(exitChan)

	if *memProfDir != "" {
		go memProfiling(*memProfDir, *memProfInterval, *memProfNrFiles, exitChan)
	}
	cpuProfChanStop := make(chan struct{})
	cpuProfChanDone := make(chan struct{})
	if *cpuProfDir != "" {
		go cpuProfiling(*cpuProfDir, exitChan, cpuProfChanStop, cpuProfChanDone)
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
	utils.Logger.Debug(fmt.Sprintf("Starting CGRateS with version <%s>", vers))
	var err error
	// Init config
	cfg, err = config.NewCGRConfigFromFolder(*cfgDir)
	if err != nil {
		log.Fatalf("Could not parse config: <%s>", err.Error())
		return
	}
	if *nodeID != "" {
		cfg.GeneralCfg().NodeID = *nodeID
	}
	config.SetCgrConfig(cfg) // Share the config object

	// init syslog
	if err = initLogger(cfg); err != nil {
		log.Fatalf("Could not initialize syslog connection, err: <%s>", err.Error())
		return
	}
	lgLevel := cfg.GeneralCfg().LogLevel
	if *logLevel != -1 { // Modify the log level if provided by command arguments
		lgLevel = *logLevel
	}
	utils.Logger.SetLogLevel(lgLevel)

	var loadDb engine.LoadStorage
	var cdrDb engine.CdrStorage
	var dm *engine.DataManager
	if cfg.RalsCfg().RALsEnabled || cfg.PubSubServerEnabled ||
		cfg.AliasesServerEnabled || cfg.UserServerEnabled || cfg.SchedulerCfg().Enabled ||
		cfg.AttributeSCfg().Enabled || cfg.ResourceSCfg().Enabled || cfg.StatSCfg().Enabled ||
		cfg.ThresholdSCfg().Enabled || cfg.SupplierSCfg().Enabled { // Some services can run without db, ie: SessionS or CDRC
		dm, err = engine.ConfigureDataStorage(cfg.DataDbCfg().DataDbType,
			cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort,
			cfg.DataDbCfg().DataDbName, cfg.DataDbCfg().DataDbUser,
			cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
			cfg.CacheCfg(), cfg.DataDbCfg().DataDbSentinelName)
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
	if cfg.RalsCfg().RALsEnabled || cfg.CdrsCfg().CDRSEnabled {
		storDb, err := engine.ConfigureStorStorage(cfg.StorDbCfg().StorDBType,
			cfg.StorDbCfg().StorDBHost, cfg.StorDbCfg().StorDBPort,
			cfg.StorDbCfg().StorDBName, cfg.StorDbCfg().StorDBUser,
			cfg.StorDbCfg().StorDBPass, cfg.GeneralCfg().DBDataEncoding,
			cfg.StorDbCfg().StorDBMaxOpenConns, cfg.StorDbCfg().StorDBMaxIdleConns,
			cfg.StorDbCfg().StorDBConnMaxLifetime, cfg.StorDbCfg().StorDBCDRSIndexes)
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
	engine.SetRoundingDecimals(cfg.GeneralCfg().RoundingDecimals)
	engine.SetRpSubjectPrefixMatching(cfg.RalsCfg().RpSubjectPrefixMatching)
	engine.SetLcrSubjectPrefixMatching(cfg.RalsCfg().LcrSubjectPrefixMatching)
	stopHandled := false

	// Rpc/http server
	server := new(utils.Server)

	// init cache
	cacheS := engine.NewCacheS(cfg, dm)
	server.RpcRegister(v1.NewCacheSv1(cacheS)) // before pre-caching so we can check status via API
	go func() {
		if err := cacheS.Precache(); err != nil {
			errCGR := err.(*utils.CGRError)
			errCGR.ActivateLongError()
			utils.Logger.Crit(fmt.Sprintf("<%s> error: %s on precache",
				utils.CacheS, err.Error()))
			exitChan <- true
		}
	}()

	// Async starts here, will follow cgrates.json start order

	// Define internal connections via channels
	internalRaterChan := make(chan rpcclient.RpcClientConnection, 1)
	internalCdrSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalPubSubSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalUserSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAliaseSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalSMGChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAttributeSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalChargerSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalRsChan := make(chan rpcclient.RpcClientConnection, 1)
	internalStatSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalThresholdSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalSupplierSChan := make(chan rpcclient.RpcClientConnection, 1)
	filterSChan := make(chan *engine.FilterS, 1)
	internalDispatcherSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAnalyzerSChan := make(chan rpcclient.RpcClientConnection, 1)

	// Start ServiceManager
	srvManager := servmanager.NewServiceManager(cfg, dm, exitChan, cacheS)

	// Start rater service
	if cfg.RalsCfg().RALsEnabled {
		go startRater(internalRaterChan, cacheS, internalThresholdSChan,
			internalStatSChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan,
			srvManager, server, dm, loadDb, cdrDb, &stopHandled, exitChan, filterSChan)
	}

	// Start Scheduler
	if cfg.SchedulerCfg().Enabled {
		go srvManager.StartScheduler(true)
	}

	// Start CDR Server
	if cfg.CdrsCfg().CDRSEnabled {
		go startCDRS(internalCdrSChan, cdrDb, dm,
			internalRaterChan, internalPubSubSChan, internalAttributeSChan,
			internalUserSChan, internalAliaseSChan,
			internalThresholdSChan, internalStatSChan, internalChargerSChan,
			server, exitChan, filterSChan)
	}

	// Create connection to CDR Server and share it in engine(used for *cdrlog action)
	if len(cfg.SchedulerCfg().CDRsConns) != 0 {
		go schedCDRsConns(internalCdrSChan, exitChan)
	}

	// Start CDRC components if necessary
	go startCdrcs(internalCdrSChan, internalRaterChan, exitChan, filterSChan)

	// Start SM-Generic
	if cfg.SessionSCfg().Enabled {
		go startSessionS(internalSMGChan, internalRaterChan,
			internalRsChan, internalThresholdSChan,
			internalStatSChan, internalSupplierSChan, internalAttributeSChan,
			internalCdrSChan, internalChargerSChan, server, exitChan)
	}
	// Start FreeSWITCHAgent
	if cfg.FsAgentCfg().Enabled {
		go startFsAgent(internalSMGChan, exitChan)
	}

	// Start SM-Kamailio
	if cfg.KamAgentCfg().Enabled {
		go startKamAgent(internalSMGChan, exitChan)
	}

	if cfg.AsteriskAgentCfg().Enabled {
		go startAsteriskAgent(internalSMGChan, exitChan)
	}

	if cfg.DiameterAgentCfg().Enabled {
		go startDiameterAgent(internalSMGChan, exitChan, filterSChan)
	}

	if cfg.RadiusAgentCfg().Enabled {
		go startRadiusAgent(internalSMGChan, exitChan, filterSChan)
	}

	if len(cfg.HttpAgentCfg()) != 0 {
		go startHTTPAgent(internalSMGChan, exitChan, server, filterSChan,
			cfg.GeneralCfg().DefaultTenant)
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
	go startFilterService(filterSChan, cacheS, internalStatSChan, cfg, dm, exitChan)

	if cfg.AttributeSCfg().Enabled {
		go startAttributeService(internalAttributeSChan, cacheS,
			cfg, dm, server, exitChan, filterSChan)
	}
	if cfg.ChargerSCfg().Enabled {
		go startChargerService(internalChargerSChan, cacheS,
			internalAttributeSChan, cfg, dm, server, exitChan, filterSChan)
	}

	// Start RL service
	if cfg.ResourceSCfg().Enabled {
		go startResourceService(internalRsChan, cacheS,
			internalThresholdSChan, cfg, dm, server, exitChan, filterSChan)
	}

	if cfg.StatSCfg().Enabled {
		go startStatService(internalStatSChan, cacheS,
			internalThresholdSChan, cfg, dm, server, exitChan, filterSChan)
	}

	if cfg.ThresholdSCfg().Enabled {
		go startThresholdService(internalThresholdSChan, cacheS,
			cfg, dm, server, exitChan, filterSChan)
	}

	if cfg.SupplierSCfg().Enabled {
		go startSupplierService(internalSupplierSChan, cacheS,
			internalRsChan, internalStatSChan,
			cfg, dm, server, exitChan, filterSChan, internalAttributeSChan)
	}
	if cfg.DispatcherSCfg().Enabled {
		go startDispatcherService(internalDispatcherSChan,
			internalRaterChan, cacheS, dm, server, exitChan)
	}

	if cfg.AnalyzerSCfg().Enabled {
		go startAnalyzerService(internalAnalyzerSChan, server, exitChan)
	}

	go loaderService(cacheS, cfg, dm, server, exitChan, filterSChan)

	// Serve rpc connections
	go startRpc(server, internalRaterChan, internalCdrSChan,
		internalPubSubSChan, internalUserSChan, internalAliaseSChan, internalRsChan,
		internalStatSChan,
		internalAttributeSChan, internalChargerSChan, internalThresholdSChan,
		internalSupplierSChan,
		internalSMGChan, internalDispatcherSChan, internalAnalyzerSChan, exitChan)
	<-exitChan

	if *cpuProfDir != "" { // wait to end cpuProfiling
		cpuProfChanStop <- struct{}{}
		<-cpuProfChanDone
	}
	if *memProfDir != "" { // write last memory profiling
		memProfFile(path.Join(*memProfDir, "mem_final.prof"))
	}
	if *pidFile != "" {
		if err := os.Remove(*pidFile); err != nil {
			utils.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	utils.Logger.Info("Stopped all components. CGRateS shutdown!")
}
