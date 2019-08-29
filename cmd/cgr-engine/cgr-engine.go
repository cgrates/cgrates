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
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	cgrEngineFlags    = flag.NewFlagSet("cgr-engine", flag.ContinueOnError)
	cfgPath           = cgrEngineFlags.String("config_path", utils.CONFIG_PATH, "Configuration directory path.")
	version           = cgrEngineFlags.Bool("version", false, "Prints the application version.")
	pidFile           = cgrEngineFlags.String("pid", "", "Write pid file")
	httpPprofPath     = cgrEngineFlags.String("httprof_path", "", "http address used for program profiling")
	cpuProfDir        = cgrEngineFlags.String("cpuprof_dir", "", "write cpu profile to files")
	memProfDir        = cgrEngineFlags.String("memprof_dir", "", "write memory profile to file")
	memProfInterval   = cgrEngineFlags.Duration("memprof_interval", 5*time.Second, "Time betwen memory profile saves")
	memProfNrFiles    = cgrEngineFlags.Int("memprof_nrfiles", 1, "Number of memory profile to write")
	scheduledShutdown = cgrEngineFlags.String("scheduled_shutdown", "", "shutdown the engine after this duration")
	singlecpu         = cgrEngineFlags.Bool("singlecpu", false, "Run on single CPU core")
	syslogger         = cgrEngineFlags.String("logger", "", "logger <*syslog|*stdout>")
	nodeID            = cgrEngineFlags.String("node_id", "", "The node ID of the engine")
	logLevel          = cgrEngineFlags.Int("log_level", -1, "Log level (0-emergency to 7-debug)")

	cfg *config.CGRConfig
)

func startCdrcs(internalCdrSChan, internalRaterChan, internalDispatcherSChan chan rpcclient.RpcClientConnection,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	cdrcInitialized := false           // Control whether the cdrc was already initialized (so we don't reload in that case)
	var cdrcChildrenChan chan struct{} // Will use it to communicate with the children of one fork
	intCdrSChan := internalCdrSChan
	if cfg.DispatcherSCfg().Enabled {
		intCdrSChan = internalDispatcherSChan
	}
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
				go startCdrc(intCdrSChan, internalRaterChan, enabledCfgs,
					cfg.GeneralCfg().HttpSkipTlsVerify, filterSChan,
					cdrcChildrenChan, exitChan)
			} else {
				utils.Logger.Info("<CDRC> No enabled CDRC clients")
			}
		}
		cdrcInitialized = true // Initialized
	}
}

// Fires up a cdrc instance
func startCdrc(internalCdrSChan, internalRaterChan chan rpcclient.RpcClientConnection, cdrcCfgs []*config.CdrcCfg, httpSkipTlsCheck bool,
	filterSChan chan *engine.FilterS, closeChan chan struct{}, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	var err error
	var cdrsConn rpcclient.RpcClientConnection
	cdrcCfg := cdrcCfgs[0]
	cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.TlsCfg().ClientKey,
		cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
		cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
		cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
		cdrcCfg.CdrsConns, internalCdrSChan, false)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRC> Could not connect to CDRS via RPC: %s", err.Error()))
		exitChan <- true
		return
	}

	cdrc, err := cdrc.NewCdrc(cdrcCfgs, httpSkipTlsCheck, cdrsConn, closeChan,
		cfg.GeneralCfg().DefaultTimezone, filterS)
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
	internalStatSChan, internalSupplierSChan, internalAttrSChan, internalCDRSChan, internalChargerSChan,
	internalDispatcherSChan chan rpcclient.RpcClientConnection, server *utils.Server,
	dm *engine.DataManager, exitChan chan bool) {
	var err error
	var ralsConns, resSConns, threshSConns, statSConns, suplSConns, attrSConns, cdrsConn, chargerSConn rpcclient.RpcClientConnection

	intChargerSChan := internalChargerSChan
	intRaterChan := internalRaterChan
	intResourceSChan := internalResourceSChan
	intThresholdSChan := internalThresholdSChan
	intStatSChan := internalStatSChan
	intSupplierSChan := internalSupplierSChan
	intAttrSChan := internalAttrSChan
	intCDRSChan := internalCDRSChan
	if cfg.DispatcherSCfg().Enabled {
		intChargerSChan = internalDispatcherSChan
		intRaterChan = internalDispatcherSChan
		intResourceSChan = internalDispatcherSChan
		intThresholdSChan = internalDispatcherSChan
		intStatSChan = internalDispatcherSChan
		intSupplierSChan = internalDispatcherSChan
		intAttrSChan = internalDispatcherSChan
		intCDRSChan = internalDispatcherSChan
	}

	if len(cfg.SessionSCfg().ChargerSConns) != 0 {
		chargerSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey, cfg.TlsCfg().ClientCerificate,
			cfg.TlsCfg().CaCertificate, cfg.GeneralCfg().ConnectAttempts,
			cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().ConnectTimeout,
			cfg.GeneralCfg().ReplyTimeout, cfg.SessionSCfg().ChargerSConns,
			intChargerSChan, false)
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
			intRaterChan, false)
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
			cfg.SessionSCfg().ResSConns, intResourceSChan, false)
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
			cfg.SessionSCfg().ThreshSConns, intThresholdSChan, false)
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
			cfg.SessionSCfg().StatSConns, intStatSChan, false)
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
			cfg.SessionSCfg().SupplSConns, intSupplierSChan, false)
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
			cfg.SessionSCfg().AttrSConns, intAttrSChan, false)
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
			cfg.SessionSCfg().CDRsConns, intCDRSChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to RALs: %s",
				utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	sReplConns, err := sessions.NewSReplConns(cfg.SessionSCfg().SessionReplicationConns,
		cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().ConnectTimeout,
		cfg.GeneralCfg().ReplyTimeout)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to SMGReplicationConnection error: <%s>",
			utils.SessionS, err.Error()))
		exitChan <- true
		return
	}
	sm := sessions.NewSessionS(cfg, ralsConns, resSConns, threshSConns,
		statSConns, suplSConns, attrSConns, cdrsConn, chargerSConn,
		sReplConns, dm, cfg.GeneralCfg().DefaultTimezone)
	//start sync session in a separate gorutine
	go func() {
		if err = sm.ListenAndServe(exitChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.SessionS, err))
		}
	}()
	// Pass internal connection via BiRPCClient
	internalSMGChan <- sm
	// Register RPC handler
	smgRpc := v1.NewSMGenericV1(sm)
	server.RpcRegister(smgRpc)

	ssv1 := v1.NewSessionSv1(sm) // methods with multiple options
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(ssv1)
	}
	// Register BiRpc handlers
	if cfg.SessionSCfg().ListenBijson != "" {
		for method, handler := range smgRpc.Handlers() {
			server.BiRPCRegisterName(method, handler)
		}
		for method, handler := range ssv1.Handlers() {
			server.BiRPCRegisterName(method, handler)
		}
		server.ServeBiJSON(cfg.SessionSCfg().ListenBijson, sm.OnBiJSONConnect, sm.OnBiJSONDisconnect)
		exitChan <- true
	}
}

// startERs handles starting of the EventReader Service
func startERs(sSChan, dspSChan chan rpcclient.RpcClientConnection,
	filterSChan chan *engine.FilterS,
	cfgRld chan struct{}, exitChan chan bool) {
	var err error

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ERs))
	filterS := <-filterSChan
	filterSChan <- filterS
	// overwrite the session service channel with dispatcher one
	if cfg.DispatcherSCfg().Enabled {
		sSChan = dspSChan
	}
	var sS rpcclient.RpcClientConnection
	if sS, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
		cfg.TlsCfg().ClientKey,
		cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
		cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
		cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
		cfg.ERsCfg().SessionSConns, sSChan, false); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> failed connecting to <%s>, error: <%s>",
			utils.ERs, utils.SessionS, err.Error()))
		exitChan <- true
		return
	}

	var erS *ers.ERService
	if erS, err = ers.NewERService(cfg, filterS, sS, exitChan); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.ERs, err.Error()))
		exitChan <- true
		return
	}

	if err = erS.ListenAndServe(cfgRld); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.ERs, err.Error()))
	}

	exitChan <- true
}

func startAsteriskAgent(internalSMGChan, internalDispatcherSChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	var err error
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	utils.Logger.Info("Starting Asterisk agent")
	intSMGChan := internalSMGChan
	if cfg.DispatcherSCfg().Enabled {
		intSMGChan = internalDispatcherSChan
	}
	if !cfg.DispatcherSCfg().Enabled && cfg.AsteriskAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		sSIntConn := <-internalSMGChan
		internalSMGChan <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		sS, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.AsteriskAgentCfg().SessionSConns, intSMGChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.AsteriskAgent, utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}

	listenAndServe := func(sma *agents.AsteriskAgent, exitChan chan bool) {
		if err = sma.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> runtime error: %s!", utils.AsteriskAgent, err))
		}
		exitChan <- true
	}
	for connIdx := range cfg.AsteriskAgentCfg().AsteriskConns { // Instantiate connections towards asterisk servers
		sma, err := agents.NewAsteriskAgent(cfg, connIdx, sS)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.AsteriskAgent, err))
			exitChan <- true
			return
		}
		if sSInternal { // bidirectional client backwards connection
			sS.(*utils.BiRPCInternalClient).SetClientConn(sma)
			var rply string
			if err := sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
				utils.EmptyString, &rply); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
					utils.AsteriskAgent, utils.SessionS, err.Error()))
				exitChan <- true
				return
			}
		}
		go listenAndServe(sma, exitChan)
	}
}

func startDiameterAgent(internalSsChan, internalDispatcherSChan chan rpcclient.RpcClientConnection,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	var err error
	utils.Logger.Info("Starting CGRateS DiameterAgent service")
	filterS := <-filterSChan
	filterSChan <- filterS
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	intSsChan := internalSsChan
	if cfg.DispatcherSCfg().Enabled {
		intSsChan = internalDispatcherSChan
	}
	if !cfg.DispatcherSCfg().Enabled && cfg.DiameterAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		sSIntConn := <-internalSsChan
		internalSsChan <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		sS, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DiameterAgentCfg().SessionSConns, intSsChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DiameterAgent, utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}

	da, err := agents.NewDiameterAgent(cfg, filterS, sS)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> error: %s!", err))
		exitChan <- true
		return
	}
	if sSInternal { // bidirectional client backwards connection
		sS.(*utils.BiRPCInternalClient).SetClientConn(da)
		var rply string
		if err := sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
			utils.EmptyString, &rply); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DiameterAgent, utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if err = da.ListenAndServe(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<DiameterAgent> error: %s!", err))
	}
	exitChan <- true
}

func startRadiusAgent(internalSMGChan, internalDispatcherSChan chan rpcclient.RpcClientConnection,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	utils.Logger.Info("Starting CGRateS RadiusAgent service")
	var err error
	var smgConn rpcclient.RpcClientConnection
	intSMGChan := internalSMGChan
	if cfg.DispatcherSCfg().Enabled {
		intSMGChan = internalDispatcherSChan
	}

	smgConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
		cfg.TlsCfg().ClientKey,
		cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
		cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
		cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
		cfg.RadiusAgentCfg().SessionSConns, intSMGChan, false)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to SMG: %s", utils.RadiusAgent, err.Error()))
		exitChan <- true
		return
	}

	ra, err := agents.NewRadiusAgent(cfg, filterS, smgConn)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.RadiusAgent, err.Error()))
		exitChan <- true
		return
	}
	if err = ra.ListenAndServe(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.RadiusAgent, err.Error()))
	}
	exitChan <- true
}

func startDNSAgent(internalSMGChan, internalDispatcherSChan chan rpcclient.RpcClientConnection,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	var err error
	var sS rpcclient.RpcClientConnection
	// var sSInternal bool
	filterS := <-filterSChan
	filterSChan <- filterS
	utils.Logger.Info(fmt.Sprintf("starting %s service", utils.DNSAgent))
	intSMGChan := internalSMGChan
	if cfg.DispatcherSCfg().Enabled {
		intSMGChan = internalDispatcherSChan
	}
	if !cfg.DispatcherSCfg().Enabled && cfg.DNSAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		// sSInternal = true
		sSIntConn := <-internalSMGChan
		internalSMGChan <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		sS, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DNSAgentCfg().SessionSConns, intSMGChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DNSAgent, utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	da, err := agents.NewDNSAgent(cfg, filterS, sS)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
		exitChan <- true
		return
	}
	// if sSInternal { // bidirectional client backwards connection
	// 	sS.(*utils.BiRPCInternalClient).SetClientConn(da)
	// 	var rply string
	// 	if err := sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
	// 		utils.EmptyString, &rply); err != nil {
	// 		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
	// 			utils.DNSAgent, utils.SessionS, err.Error()))
	// 		exitChan <- true
	// 		return
	// 	}
	// }
	if err = da.ListenAndServe(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
	}
	exitChan <- true
}

func startFsAgent(internalSMGChan, internalDispatcherSChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	var err error
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	utils.Logger.Info("Starting FreeSWITCH agent")
	intSMGChan := internalSMGChan
	if cfg.DispatcherSCfg().Enabled {
		intSMGChan = internalDispatcherSChan
	}
	if !cfg.DispatcherSCfg().Enabled && cfg.FsAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		sSIntConn := <-internalSMGChan
		internalSMGChan <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		sS, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.FsAgentCfg().SessionSConns, intSMGChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.FreeSWITCHAgent, utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	sm := agents.NewFSsessions(cfg.FsAgentCfg(), sS, cfg.GeneralCfg().DefaultTimezone)
	if sSInternal { // bidirectional client backwards connection
		sS.(*utils.BiRPCInternalClient).SetClientConn(sm)
		var rply string
		if err := sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
			utils.EmptyString, &rply); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.FreeSWITCHAgent, utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if err = sm.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
	}

	exitChan <- true
}

func startKamAgent(internalSMGChan, internalDispatcherSChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	var err error
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	utils.Logger.Info("Starting Kamailio agent")
	intSMGChan := internalSMGChan
	if cfg.DispatcherSCfg().Enabled {
		intSMGChan = internalDispatcherSChan
	}
	if !cfg.DispatcherSCfg().Enabled && cfg.KamAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		sSIntConn := <-internalSMGChan
		internalSMGChan <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		sS, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.KamAgentCfg().SessionSConns, intSMGChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.KamailioAgent, utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	ka := agents.NewKamailioAgent(cfg.KamAgentCfg(), sS,
		utils.FirstNonEmpty(cfg.KamAgentCfg().Timezone, cfg.GeneralCfg().DefaultTimezone))
	if sSInternal { // bidirectional client backwards connection
		sS.(*utils.BiRPCInternalClient).SetClientConn(ka)
		var rply string
		if err := sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
			utils.EmptyString, &rply); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.KamailioAgent, utils.SessionS, err.Error()))
			exitChan <- true
			return
		}
	}
	if err = ka.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, err))
	}
	exitChan <- true
}

func startHTTPAgent(internalSMGChan, internalDispatcherSChan chan rpcclient.RpcClientConnection,
	server *utils.Server, filterSChan chan *engine.FilterS, dfltTenant string, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	var sS rpcclient.RpcClientConnection
	intSMGChan := internalSMGChan
	if cfg.DispatcherSCfg().Enabled {
		intSMGChan = internalDispatcherSChan
	}
	utils.Logger.Info("Starting HTTP agent")
	var err error
	for _, agntCfg := range cfg.HttpAgentCfg() {
		if len(agntCfg.SessionSConns) != 0 {
			sS, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
				cfg.TlsCfg().ClientKey,
				cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
				cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
				cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
				agntCfg.SessionSConns, intSMGChan, false)
			if err != nil {
				utils.Logger.Crit(fmt.Sprintf("<%s> could not connect to %s, error: %s",
					utils.HTTPAgent, utils.SessionS, err.Error()))
				exitChan <- true
				return
			}
		}
		server.RegisterHttpHandler(agntCfg.Url,
			agents.NewHTTPAgent(sS, filterS, dfltTenant, agntCfg.RequestPayload,
				agntCfg.ReplyPayload, agntCfg.RequestProcessors))
	}
}

func startCDRS(internalCdrSChan, internalRaterChan, internalAttributeSChan, internalThresholdSChan,
	internalStatSChan, internalChargerSChan, internalDispatcherSChan chan rpcclient.RpcClientConnection,
	cdrDb engine.CdrStorage, dm *engine.DataManager, server *utils.Server,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	var err error
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.CDRs))

	var ralConn, attrSConn, thresholdSConn, statsConn, chargerSConn rpcclient.RpcClientConnection

	intChargerSChan := internalChargerSChan
	intRaterChan := internalRaterChan
	intAttributeSChan := internalAttributeSChan
	intThresholdSChan := internalThresholdSChan
	intStatSChan := internalStatSChan
	if cfg.DispatcherSCfg().Enabled {
		intChargerSChan = internalDispatcherSChan
		intRaterChan = internalDispatcherSChan
		intAttributeSChan = internalDispatcherSChan
		intThresholdSChan = internalDispatcherSChan
		intStatSChan = internalDispatcherSChan
	}
	if len(cfg.CdrsCfg().CDRSChargerSConns) != 0 { // Conn pool towards RAL
		chargerSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.CdrsCfg().CDRSChargerSConns, intChargerSChan, false)
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
			cfg.CdrsCfg().CDRSRaterConns, intRaterChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to RAL: %s", err.Error()))
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
			cfg.CdrsCfg().CDRSAttributeSConns, intAttributeSChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
				utils.AttributeS, err.Error()))
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
			cfg.CdrsCfg().CDRSThresholdSConns, intThresholdSChan, false)
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
			cfg.CdrsCfg().CDRSStatSConns, intStatSChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to StatS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	cdrServer := engine.NewCDRServer(cfg, cdrDb, dm,
		ralConn, attrSConn,
		thresholdSConn, statsConn, chargerSConn, filterS)
	utils.Logger.Info("Registering CDRS HTTP Handlers.")
	cdrServer.RegisterHandlersToServer(server)
	utils.Logger.Info("Registering CDRS RPC service.")
	cdrSrv := v1.NewCDRsV1(cdrServer)
	server.RpcRegister(cdrSrv)
	server.RpcRegister(&v2.CDRsV2{CDRsV1: *cdrSrv})
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

// startAttributeService fires up the AttributeS
func startAttributeService(internalAttributeSChan chan rpcclient.RpcClientConnection,
	cacheS *engine.CacheS, cfg *config.CGRConfig, dm *engine.DataManager,
	server *utils.Server, filterSChan chan *engine.FilterS, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	<-cacheS.GetPrecacheChannel(utils.CacheAttributeProfiles)
	<-cacheS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes)

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
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(aSv1)
	}
	internalAttributeSChan <- aSv1
}

// startChargerService fires up the ChargerS
func startChargerService(internalChargerSChan, internalAttributeSChan,
	internalDispatcherSChan chan rpcclient.RpcClientConnection,
	cacheS *engine.CacheS, cfg *config.CGRConfig,
	dm *engine.DataManager, server *utils.Server,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	<-cacheS.GetPrecacheChannel(utils.CacheChargerProfiles)
	<-cacheS.GetPrecacheChannel(utils.CacheChargerFilterIndexes)
	var attrSConn rpcclient.RpcClientConnection
	var err error
	intAttributeSChan := internalAttributeSChan
	if cfg.DispatcherSCfg().Enabled {
		intAttributeSChan = internalDispatcherSChan
	}
	if len(cfg.ChargerSCfg().AttributeSConns) != 0 { // AttributeS connection init
		attrSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.ChargerSCfg().AttributeSConns, intAttributeSChan, false)
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
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(cSv1)
	}
	internalChargerSChan <- cSv1
}

func startResourceService(internalRsChan, internalThresholdSChan,
	internalDispatcherSChan chan rpcclient.RpcClientConnection,
	cacheS *engine.CacheS, cfg *config.CGRConfig,
	dm *engine.DataManager, server *utils.Server,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	var err error
	var thdSConn rpcclient.RpcClientConnection
	filterS := <-filterSChan
	filterSChan <- filterS
	intThresholdSChan := internalThresholdSChan
	if cfg.DispatcherSCfg().Enabled {
		intThresholdSChan = internalDispatcherSChan
	}
	if len(cfg.ResourceSCfg().ThresholdSConns) != 0 { // Stats connection init
		thdSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.ResourceSCfg().ThresholdSConns, intThresholdSChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<ResourceS> Could not connect to ThresholdS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	<-cacheS.GetPrecacheChannel(utils.CacheResourceProfiles)
	<-cacheS.GetPrecacheChannel(utils.CacheResources)
	<-cacheS.GetPrecacheChannel(utils.CacheResourceFilterIndexes)

	rS, err := engine.NewResourceService(dm, cfg.ResourceSCfg().StoreInterval,
		thdSConn, filterS, cfg.ResourceSCfg().StringIndexedFields, cfg.ResourceSCfg().PrefixIndexedFields)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<ResourceS> Could not init, error: %s", err.Error()))
		exitChan <- true
		return
	}
	go func() {
		if err := rS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<ResourceS> Could not start, error: %s", err.Error()))

		}
		rS.Shutdown()
		exitChan <- true
		return
	}()
	rsV1 := v1.NewResourceSv1(rS)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(rsV1)
	}
	internalRsChan <- rsV1
}

// startStatService fires up the StatS
func startStatService(internalStatSChan, internalThresholdSChan,
	internalDispatcherSChan chan rpcclient.RpcClientConnection,
	cacheS *engine.CacheS, cfg *config.CGRConfig,
	dm *engine.DataManager, server *utils.Server,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	var err error
	var thdSConn rpcclient.RpcClientConnection
	filterS := <-filterSChan
	filterSChan <- filterS
	intThresholdSChan := internalThresholdSChan
	if cfg.DispatcherSCfg().Enabled {
		intThresholdSChan = internalDispatcherSChan
	}
	if len(cfg.StatSCfg().ThresholdSConns) != 0 { // Stats connection init
		thdSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.StatSCfg().ThresholdSConns, intThresholdSChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<StatS> Could not connect to ThresholdS: %s", err.Error()))
			exitChan <- true
			return
		}
	}
	<-cacheS.GetPrecacheChannel(utils.CacheStatQueueProfiles)
	<-cacheS.GetPrecacheChannel(utils.CacheStatQueues)
	<-cacheS.GetPrecacheChannel(utils.CacheStatFilterIndexes)

	sS, err := engine.NewStatService(dm, cfg.StatSCfg().StoreInterval,
		thdSConn, filterS, cfg.StatSCfg().StringIndexedFields, cfg.StatSCfg().PrefixIndexedFields)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<StatS> Could not init, error: %s", err.Error()))
		exitChan <- true
		return
	}
	go func() {
		if err := sS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<StatS> Error: %s listening for packets", err.Error()))
		}
		sS.Shutdown()
		exitChan <- true
		return
	}()
	stsV1 := v1.NewStatSv1(sS)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(stsV1)
	}
	internalStatSChan <- stsV1
}

// startThresholdService fires up the ThresholdS
func startThresholdService(internalThresholdSChan chan rpcclient.RpcClientConnection,
	cacheS *engine.CacheS, cfg *config.CGRConfig, dm *engine.DataManager,
	server *utils.Server, filterSChan chan *engine.FilterS, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	<-cacheS.GetPrecacheChannel(utils.CacheThresholdProfiles)
	<-cacheS.GetPrecacheChannel(utils.CacheThresholds)
	<-cacheS.GetPrecacheChannel(utils.CacheThresholdFilterIndexes)

	tS, err := engine.NewThresholdService(dm, cfg.ThresholdSCfg().StringIndexedFields,
		cfg.ThresholdSCfg().PrefixIndexedFields, cfg.ThresholdSCfg().StoreInterval, filterS)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<ThresholdS> Could not init, error: %s", err.Error()))
		exitChan <- true
		return
	}
	go func() {
		if err := tS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<ThresholdS> Error: %s listening for packets", err.Error()))
		}
		tS.Shutdown()
		exitChan <- true
		return
	}()
	tSv1 := v1.NewThresholdSv1(tS)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(tSv1)
	}
	internalThresholdSChan <- tSv1
}

// startSupplierService fires up the SupplierS
func startSupplierService(internalSupplierSChan, internalRsChan, internalStatSChan,
	internalAttrSChan, internalDispatcherSChan chan rpcclient.RpcClientConnection,
	cacheS *engine.CacheS, cfg *config.CGRConfig, dm *engine.DataManager, server *utils.Server,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	var err error
	filterS := <-filterSChan
	filterSChan <- filterS
	var attrSConn, resourceSConn, statSConn rpcclient.RpcClientConnection

	intAttrSChan := internalAttrSChan
	intStatSChan := internalStatSChan
	intRsChan := internalRsChan
	if cfg.DispatcherSCfg().Enabled { // use dispatcher as internal chanel if active
		intAttrSChan = internalDispatcherSChan
		intStatSChan = internalDispatcherSChan
		intRsChan = internalDispatcherSChan
	}
	if len(cfg.SupplierSCfg().AttributeSConns) != 0 {
		attrSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.SupplierSCfg().AttributeSConns, intAttrSChan, false)
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
			cfg.SupplierSCfg().StatSConns, intStatSChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to StatS: %s",
				utils.SupplierS, err.Error()))
			exitChan <- true
			return
		}
	}
	if len(cfg.SupplierSCfg().ResourceSConns) != 0 {
		resourceSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.SupplierSCfg().ResourceSConns, intRsChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to StatS: %s",
				utils.SupplierS, err.Error()))
			exitChan <- true
			return
		}
	}
	<-cacheS.GetPrecacheChannel(utils.CacheSupplierProfiles)
	<-cacheS.GetPrecacheChannel(utils.CacheSupplierFilterIndexes)

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
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(splV1)
	}
	internalSupplierSChan <- splV1
}

// startFilterService fires up the FilterS
func startFilterService(filterSChan chan *engine.FilterS, cacheS *engine.CacheS,
	internalStatSChan, internalResourceSChan, internalRalSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dm *engine.DataManager, exitChan chan bool) {
	<-cacheS.GetPrecacheChannel(utils.CacheFilters)
	filterSChan <- engine.NewFilterS(cfg, internalStatSChan, internalResourceSChan, internalRalSChan, dm)
}

// loaderService will start and register APIs for LoaderService if enabled
func startLoaderS(internalLoaderSChan, cacheSChan chan rpcclient.RpcClientConnection,
	cfg *config.CGRConfig, dm *engine.DataManager, server *utils.Server,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS

	ldrS := loaders.NewLoaderService(dm, cfg.LoaderCfg(),
		cfg.GeneralCfg().DefaultTimezone, exitChan, filterS, cacheSChan)
	if !ldrS.Enabled() {
		return
	}
	go ldrS.ListenAndServe(exitChan)
	ldrSv1 := v1.NewLoaderSv1(ldrS)
	server.RpcRegister(ldrSv1)
	internalLoaderSChan <- ldrSv1
}

// startDispatcherService fires up the DispatcherS
func startDispatcherService(internalDispatcherSChan, internalAttributeSChan chan rpcclient.RpcClientConnection,
	cfg *config.CGRConfig, cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	dm *engine.DataManager, server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS Dispatcher service.")
	fltrS := <-filterSChan
	filterSChan <- fltrS
	<-cacheS.GetPrecacheChannel(utils.CacheDispatcherProfiles)
	<-cacheS.GetPrecacheChannel(utils.CacheDispatcherHosts)
	<-cacheS.GetPrecacheChannel(utils.CacheDispatcherFilterIndexes)

	var err error
	var attrSConn *rpcclient.RpcClientPool
	if len(cfg.DispatcherSCfg().AttributeSConns) != 0 { // AttributeS connection init
		attrSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DispatcherSCfg().AttributeSConns, internalAttributeSChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DispatcherS, utils.AttributeS, err.Error()))
			exitChan <- true
			return
		}
	}
	dspS, err := dispatchers.NewDispatcherService(dm, cfg, fltrS, attrSConn)
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

	// for the moment we dispable Apier through dispatcher
	// until we figured out a better sollution in case of gob server
	// server.SetDispatched()

	server.RpcRegister(v1.NewDispatcherSv1(dspS))

	server.RpcRegisterName(utils.ThresholdSv1,
		v1.NewDispatcherThresholdSv1(dspS))

	server.RpcRegisterName(utils.StatSv1,
		v1.NewDispatcherStatSv1(dspS))

	server.RpcRegisterName(utils.ResourceSv1,
		v1.NewDispatcherResourceSv1(dspS))

	server.RpcRegisterName(utils.SupplierSv1,
		v1.NewDispatcherSupplierSv1(dspS))

	server.RpcRegisterName(utils.AttributeSv1,
		v1.NewDispatcherAttributeSv1(dspS))

	server.RpcRegisterName(utils.SessionSv1,
		v1.NewDispatcherSessionSv1(dspS))

	server.RpcRegisterName(utils.ChargerSv1,
		v1.NewDispatcherChargerSv1(dspS))

	server.RpcRegisterName(utils.Responder,
		v1.NewDispatcherResponder(dspS))

	server.RpcRegisterName(utils.CacheSv1,
		v1.NewDispatcherCacheSv1(dspS))

	server.RpcRegisterName(utils.GuardianSv1,
		v1.NewDispatcherGuardianSv1(dspS))

	server.RpcRegisterName(utils.SchedulerSv1,
		v1.NewDispatcherSchedulerSv1(dspS))

	server.RpcRegisterName(utils.CDRsV1,
		v1.NewDispatcherSCDRsV1(dspS))

	server.RpcRegisterName(utils.ConfigSv1,
		v1.NewDispatcherConfigSv1(dspS))

	server.RpcRegisterName(utils.CoreSv1,
		v1.NewDispatcherCoreSv1(dspS))

	server.RpcRegisterName(utils.RALsV1,
		v1.NewDispatcherRALsV1(dspS))

	internalDispatcherSChan <- dspS
}

// startAnalyzerService fires up the AnalyzerS
func startAnalyzerService(internalAnalyzerSChan chan rpcclient.RpcClientConnection,
	server *utils.Server, exitChan chan bool) {
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

// initCacheS inits the CacheS and starts precaching as well as populating internal channel for RPC conns
func initCacheS(internalCacheSChan chan rpcclient.RpcClientConnection,
	server *utils.Server, dm *engine.DataManager, exitChan chan bool) (chS *engine.CacheS) {
	chS = engine.NewCacheS(cfg, dm)
	go func() {
		if err := chS.Precache(); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> could not init, error: %s", utils.CacheS, err.Error()))
			exitChan <- true
		}
	}()

	chSv1 := v1.NewCacheSv1(chS)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(chSv1)
	}
	internalCacheSChan <- chS
	return
}

func initGuardianSv1(internalGuardianSChan chan rpcclient.RpcClientConnection, server *utils.Server) {
	grdSv1 := v1.NewGuardianSv1()
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(grdSv1)
	}
	internalGuardianSChan <- grdSv1
}

func initCoreSv1(internalCoreSv1Chan chan rpcclient.RpcClientConnection, server *utils.Server) {
	cSv1 := v1.NewCoreSv1(engine.NewCoreService())
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(cSv1)
	}
	internalCoreSv1Chan <- cSv1
}

func initSchedulerS(internalSchedSChan chan rpcclient.RpcClientConnection,
	srvMngr *servmanager.ServiceManager, server *utils.Server) {
	schdS := servmanager.NewSchedulerS(srvMngr)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(v1.NewSchedulerSv1(schdS))
	}
	internalSchedSChan <- schdS
}

func initServiceManagerV1(internalServiceManagerChan chan rpcclient.RpcClientConnection,
	srvMngr *servmanager.ServiceManager, server *utils.Server) {
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(v1.NewServiceManagerV1(srvMngr))
	}
	internalServiceManagerChan <- srvMngr
}

func startRpc(server *utils.Server, internalRaterChan,
	internalCdrSChan, internalRsChan, internalStatSChan,
	internalAttrSChan, internalChargerSChan, internalThdSChan, internalSuplSChan,
	internalSMGChan, internalAnalyzerSChan, internalDispatcherSChan,
	internalLoaderSChan, internalRALsv1Chan, internalCacheSChan chan rpcclient.RpcClientConnection,
	exitChan chan bool) {
	if !cfg.DispatcherSCfg().Enabled {
		select { // Any of the rpc methods will unlock listening to rpc requests
		case resp := <-internalRaterChan:
			internalRaterChan <- resp
		case cdrs := <-internalCdrSChan:
			internalCdrSChan <- cdrs
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
		case analyzerS := <-internalAnalyzerSChan:
			internalAnalyzerSChan <- analyzerS
		case loaderS := <-internalLoaderSChan:
			internalLoaderSChan <- loaderS
		case ralS := <-internalRALsv1Chan:
			internalRALsv1Chan <- ralS
		case chS := <-internalCacheSChan: // added in order to start the RPC before precaching is done
			internalCacheSChan <- chS
		}
	} else {
		select {
		case dispatcherS := <-internalDispatcherSChan:
			internalDispatcherSChan <- dispatcherS
		}
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

func schedCDRsConns(internalCDRSChan, internalDispatcherSChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	intChan := internalCDRSChan
	if cfg.DispatcherSCfg().Enabled {
		intChan = internalDispatcherSChan
	}
	cdrsConn, err := engine.NewRPCPool(rpcclient.POOL_FIRST,
		cfg.TlsCfg().ClientKey,
		cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
		cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
		cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
		cfg.SchedulerCfg().CDRsConns, intChan, false)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to CDRServer: %s", utils.SchedulerS, err.Error()))
		exitChan <- true
		return
	}
	engine.SetSchedCdrsConns(cdrsConn)
}

func initConfigSv1(internalConfigChan chan rpcclient.RpcClientConnection,
	server *utils.Server) {
	cfgSv1 := v1.NewConfigSv1(cfg)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(cfgSv1)
	}
	internalConfigChan <- cfgSv1
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

func cpuProfiling(cpuProfDir string, stopChan, doneChan chan struct{}, exitChan chan bool) {
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

func singnalHandler(exitChan chan bool) {
	shutdownSignal := make(chan os.Signal)
	reloadSignal := make(chan os.Signal)
	signal.Notify(shutdownSignal, os.Interrupt,
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	for {
		select {
		case <-shutdownSignal:
			exitChan <- true
		case <-reloadSignal:
			//  do it in it's own gorutine in order to not block the signal handler with the reload functionality
			go func() {
				var reply string
				if err := config.CgrConfig().V1ReloadConfig(
					&config.ConfigReloadWithArgDispatcher{
						Section: utils.EmptyString,
						Path:    config.CgrConfig().ConfigPath, // use the same path
					}, &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("Error reloading configuration: <%s>", err))
				}
			}()
		}
	}
}

func main() {
	if err := cgrEngineFlags.Parse(os.Args[1:]); err != nil {
		return
	}
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
	go singnalHandler(exitChan)

	if *memProfDir != "" {
		go memProfiling(*memProfDir, *memProfInterval, *memProfNrFiles, exitChan)
	}
	cpuProfChanStop := make(chan struct{})
	cpuProfChanDone := make(chan struct{})
	if *cpuProfDir != "" {
		go cpuProfiling(*cpuProfDir, cpuProfChanStop, cpuProfChanDone, exitChan)
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
	cfg, err = config.NewCGRConfigFromPath(*cfgPath)
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

	cfg.LazySanityCheck()

	var loadDb engine.LoadStorage
	var cdrDb engine.CdrStorage
	var dm *engine.DataManager
	if needsDB := cfg.RalsCfg().RALsEnabled || cfg.SchedulerCfg().Enabled || cfg.ChargerSCfg().Enabled ||
		cfg.AttributeSCfg().Enabled || cfg.ResourceSCfg().Enabled || cfg.StatSCfg().Enabled ||
		cfg.ThresholdSCfg().Enabled || cfg.SupplierSCfg().Enabled || cfg.DispatcherSCfg().Enabled; needsDB ||
		cfg.SessionSCfg().Enabled { // Some services can run without db, ie:  CDRC
		dm, err = engine.ConfigureDataStorage(cfg.DataDbCfg().DataDbType,
			cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort,
			cfg.DataDbCfg().DataDbName, cfg.DataDbCfg().DataDbUser,
			cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
			cfg.CacheCfg(), cfg.DataDbCfg().DataDbSentinelName)
		if needsDB && err != nil { // Cannot configure getter database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		} else if cfg.SessionSCfg().Enabled && err != nil {
			utils.Logger.Warning(fmt.Sprintf("Could not configure dataDb: %s.Some SessionS APIs will not work", err))
		} else {
			defer dm.DataDB().Close()
			engine.SetDataStorage(dm)
			if err := engine.CheckVersions(dm.DataDB()); err != nil {
				fmt.Println(err.Error())
				return
			}
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

	// Rpc/http server
	server := utils.NewServer()

	if *httpPprofPath != "" {
		go server.RegisterProfiler(*httpPprofPath)
	}
	// Async starts here, will follow cgrates.json start order

	// Define internal connections via channels
	filterSChan := make(chan *engine.FilterS, 1)
	internalDispatcherSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalRaterChan := make(chan rpcclient.RpcClientConnection, 1)
	internalCdrSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalSMGChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAttributeSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalChargerSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalRsChan := make(chan rpcclient.RpcClientConnection, 1)
	internalStatSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalThresholdSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalSupplierSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAnalyzerSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalCacheSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalSchedSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalGuardianSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalLoaderSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalApierV1Chan := make(chan rpcclient.RpcClientConnection, 1)
	internalApierV2Chan := make(chan rpcclient.RpcClientConnection, 1)
	internalServeManagerChan := make(chan rpcclient.RpcClientConnection, 1)
	internalConfigChan := make(chan rpcclient.RpcClientConnection, 1)
	internalCoreSv1Chan := make(chan rpcclient.RpcClientConnection, 1)
	internalRALsv1Chan := make(chan rpcclient.RpcClientConnection, 1)

	// init internalRPCSet
	engine.IntRPC = engine.NewRPCClientSet()
	if cfg.DispatcherSCfg().Enabled {
		engine.IntRPC.AddInternalRPCClient(utils.AnalyzerSv1, internalAnalyzerSChan)
		engine.IntRPC.AddInternalRPCClient(utils.ApierV1, internalApierV1Chan)
		engine.IntRPC.AddInternalRPCClient(utils.ApierV2, internalApierV2Chan)
		engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1, internalAttributeSChan)
		engine.IntRPC.AddInternalRPCClient(utils.CacheSv1, internalCacheSChan) // server or from apier
		engine.IntRPC.AddInternalRPCClient(utils.CDRsV1, internalCdrSChan)
		engine.IntRPC.AddInternalRPCClient(utils.CDRsV2, internalCdrSChan)
		engine.IntRPC.AddInternalRPCClient(utils.ChargerSv1, internalChargerSChan)
		engine.IntRPC.AddInternalRPCClient(utils.GuardianSv1, internalGuardianSChan)
		engine.IntRPC.AddInternalRPCClient(utils.LoaderSv1, internalLoaderSChan)
		engine.IntRPC.AddInternalRPCClient(utils.ResourceSv1, internalRsChan)
		engine.IntRPC.AddInternalRPCClient(utils.Responder, internalRaterChan)
		engine.IntRPC.AddInternalRPCClient(utils.SchedulerSv1, internalSchedSChan) // server or from apier
		engine.IntRPC.AddInternalRPCClient(utils.SessionSv1, internalSMGChan)      // server or from apier
		engine.IntRPC.AddInternalRPCClient(utils.StatSv1, internalStatSChan)
		engine.IntRPC.AddInternalRPCClient(utils.SupplierSv1, internalSupplierSChan)
		engine.IntRPC.AddInternalRPCClient(utils.ThresholdSv1, internalThresholdSChan)
		engine.IntRPC.AddInternalRPCClient(utils.ServiceManagerV1, internalServeManagerChan)
		engine.IntRPC.AddInternalRPCClient(utils.ConfigSv1, internalConfigChan)
		engine.IntRPC.AddInternalRPCClient(utils.CoreSv1, internalCoreSv1Chan)
		engine.IntRPC.AddInternalRPCClient(utils.RALsV1, internalRALsv1Chan)
	}

	// init CacheS
	cacheS := initCacheS(internalCacheSChan, server, dm, exitChan)

	// init GuardianSv1
	initGuardianSv1(internalGuardianSChan, server)

	// init CoreSv1
	initCoreSv1(internalCoreSv1Chan, server)

	// Start ServiceManager
	srvManager := servmanager.NewServiceManager(cfg, dm, cacheS, exitChan)
	initServiceManagerV1(internalServeManagerChan, srvManager, server)

	// init SchedulerS
	initSchedulerS(internalSchedSChan, srvManager, server)

	initConfigSv1(internalConfigChan, server)

	// Start Scheduler
	if cfg.SchedulerCfg().Enabled {
		go srvManager.StartScheduler(true)
	}

	// Start RALs
	if cfg.RalsCfg().RALsEnabled {
		go startRater(internalRaterChan, internalApierV1Chan, internalApierV2Chan,
			internalThresholdSChan, internalStatSChan, internalCacheSChan, internalSchedSChan,
			internalAttributeSChan, internalDispatcherSChan, internalRALsv1Chan,
			srvManager, server, dm, loadDb, cdrDb, cacheS, filterSChan, exitChan)
	}

	// Start CDR Server
	if cfg.CdrsCfg().CDRSEnabled {
		go startCDRS(internalCdrSChan, internalRaterChan, internalAttributeSChan,
			internalThresholdSChan, internalStatSChan, internalChargerSChan,
			internalDispatcherSChan, cdrDb, dm, server, filterSChan, exitChan)
	}

	// Create connection to CDR Server and share it in engine(used for *cdrlog action)
	if len(cfg.SchedulerCfg().CDRsConns) != 0 {
		go schedCDRsConns(internalCdrSChan, internalDispatcherSChan, exitChan)
	}

	// Start CDRC components if necessary
	go startCdrcs(internalCdrSChan, internalRaterChan, internalDispatcherSChan, filterSChan, exitChan)

	// Start SM-Generic
	if cfg.SessionSCfg().Enabled {
		go startSessionS(internalSMGChan, internalRaterChan, internalRsChan,
			internalThresholdSChan, internalStatSChan, internalSupplierSChan,
			internalAttributeSChan, internalCdrSChan, internalChargerSChan,
			internalDispatcherSChan, server, dm, exitChan)
	}

	if cfg.ERsCfg().Enabled {
		go startERs(internalSMGChan, internalDispatcherSChan,
			filterSChan, cfg.GetReloadChan(config.ERsJson), exitChan)
	}
	// Start FreeSWITCHAgent
	if cfg.FsAgentCfg().Enabled {
		go startFsAgent(internalSMGChan, internalDispatcherSChan, exitChan)
	}

	// Start SM-Kamailio
	if cfg.KamAgentCfg().Enabled {
		go startKamAgent(internalSMGChan, internalDispatcherSChan, exitChan)
	}

	if cfg.AsteriskAgentCfg().Enabled {
		go startAsteriskAgent(internalSMGChan, internalDispatcherSChan, exitChan)
	}

	if cfg.DiameterAgentCfg().Enabled {
		go startDiameterAgent(internalSMGChan, internalDispatcherSChan, filterSChan, exitChan)
	}

	if cfg.RadiusAgentCfg().Enabled {
		go startRadiusAgent(internalSMGChan, internalDispatcherSChan, filterSChan, exitChan)
	}

	if cfg.DNSAgentCfg().Enabled {
		go startDNSAgent(internalSMGChan, internalDispatcherSChan, filterSChan, exitChan)
	}

	if len(cfg.HttpAgentCfg()) != 0 {
		go startHTTPAgent(internalSMGChan, internalDispatcherSChan, server, filterSChan,
			cfg.GeneralCfg().DefaultTenant, exitChan)
	}

	// Start FilterS
	go startFilterService(filterSChan, cacheS, internalStatSChan, internalRsChan, internalRaterChan, cfg, dm, exitChan)

	if cfg.AttributeSCfg().Enabled {
		go startAttributeService(internalAttributeSChan, cacheS,
			cfg, dm, server, filterSChan, exitChan)
	}
	if cfg.ChargerSCfg().Enabled {
		go startChargerService(internalChargerSChan, internalAttributeSChan,
			internalDispatcherSChan, cacheS, cfg, dm, server,
			filterSChan, exitChan)
	}

	// Start RL service
	if cfg.ResourceSCfg().Enabled {
		go startResourceService(internalRsChan, internalThresholdSChan,
			internalDispatcherSChan, cacheS, cfg, dm, server,
			filterSChan, exitChan)
	}

	if cfg.StatSCfg().Enabled {
		go startStatService(internalStatSChan, internalThresholdSChan,
			internalDispatcherSChan, cacheS, cfg, dm, server,
			filterSChan, exitChan)
	}

	if cfg.ThresholdSCfg().Enabled {
		go startThresholdService(internalThresholdSChan, cacheS,
			cfg, dm, server, filterSChan, exitChan)
	}

	if cfg.SupplierSCfg().Enabled {
		go startSupplierService(internalSupplierSChan, internalRsChan,
			internalStatSChan, internalAttributeSChan, internalDispatcherSChan,
			cacheS, cfg, dm, server, filterSChan, exitChan)
	}
	if cfg.DispatcherSCfg().Enabled {
		go startDispatcherService(internalDispatcherSChan,
			internalAttributeSChan, cfg, cacheS, filterSChan,
			dm, server, exitChan)
	}

	if cfg.AnalyzerSCfg().Enabled {
		go startAnalyzerService(internalAnalyzerSChan, server, exitChan)
	}

	go startLoaderS(internalLoaderSChan, internalCacheSChan, cfg, dm, server, filterSChan, exitChan)

	// Serve rpc connections
	go startRpc(server, internalRaterChan, internalCdrSChan,
		internalRsChan, internalStatSChan,
		internalAttributeSChan, internalChargerSChan, internalThresholdSChan,
		internalSupplierSChan, internalSMGChan, internalAnalyzerSChan,
		internalDispatcherSChan, internalLoaderSChan, internalRALsv1Chan, internalCacheSChan, exitChan)
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
