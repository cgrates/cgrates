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

package services

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func NewCGREngineFlags() *CGREngineFlags {
	fs := flag.NewFlagSet(utils.CgrEngine, flag.ContinueOnError)
	return &CGREngineFlags{
		FlagSet:           fs,
		Cfgpath:           fs.String(utils.CfgPathCgr, utils.ConfigPath, "Configuration directory path."),
		Version:           fs.Bool(utils.VersionCgr, false, "Prints the application version."),
		Checkconfig:       fs.Bool(utils.CheckCfgCgr, false, "Verify the config without starting the engine"),
		Pidfile:           fs.String(utils.PidCgr, utils.EmptyString, "Write pid file"),
		Httppprofpath:     fs.String(utils.HttpPrfPthCgr, utils.EmptyString, "http address used for program profiling"),
		Cpuprofdir:        fs.String(utils.CpuProfDirCgr, utils.EmptyString, "write cpu profile to files"),
		Memprofdir:        fs.String(utils.MemProfDirCgr, utils.EmptyString, "write memory profile to file"),
		Memprofinterval:   fs.Duration(utils.MemProfIntervalCgr, 5*time.Second, "Time between memory profile saves"),
		Memprofnrfiles:    fs.Int(utils.MemProfNrFilesCgr, 1, "Number of memory profile to write"),
		Scheduledshutdown: fs.String(utils.ScheduledShutdownCgr, utils.EmptyString, "shutdown the engine after this duration"),
		Singlecpu:         fs.Bool(utils.SingleCpuCgr, false, "Run on single CPU core"),
		Syslogger:         fs.String(utils.LoggerCfg, utils.EmptyString, "logger <*syslog|*stdout>"),
		Nodeid:            fs.String(utils.NodeIDCfg, utils.EmptyString, "The node ID of the engine"),
		Loglevel:          fs.Int(utils.LogLevelCfg, -1, "Log level (0-emergency to 7-debug)"),
		Preload:           fs.String(utils.PreloadCgr, utils.EmptyString, "LoaderIDs used to load the data before the engine starts"),
	}
}

type CGREngineFlags struct {
	*flag.FlagSet

	Cfgpath           *string
	Version           *bool
	Checkconfig       *bool
	Pidfile           *string
	Httppprofpath     *string
	Cpuprofdir        *string
	Memprofdir        *string
	Memprofinterval   *time.Duration
	Memprofnrfiles    *int
	Scheduledshutdown *string
	Singlecpu         *bool
	Syslogger         *string
	Nodeid            *string
	Loglevel          *int
	Preload           *string
}

func cgrSingnalHandler(ctx *context.Context, shutdown context.CancelFunc,
	cfg *config.CGRConfig, shdWg *sync.WaitGroup) {
	shutdownSignal := make(chan os.Signal, 1)
	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt,
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	for {
		select {
		case <-ctx.Done():
			shdWg.Done()
			return
		case <-shutdownSignal:
			shutdown()
			shdWg.Done()
			return
		case <-reloadSignal:
			//  do it in it's own goroutine in order to not block the signal handler with the reload functionality
			go func() {
				var reply string
				if err := cfg.V1ReloadConfig(ctx,
					new(config.ReloadArgs), &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("Error reloading configuration: <%s>", err))
				}
			}()
		}
	}
}

func cgrWritePid(pidFile string) (err error) {
	var f *os.File
	if f, err = os.Create(pidFile); err != nil {
		err = fmt.Errorf("could not create pid file: %s", err)
		return
	}
	if _, err = f.WriteString(strconv.Itoa(os.Getpid())); err != nil {
		f.Close()
		err = fmt.Errorf("could not write pid file: %s", err)
		return
	}
	if err = f.Close(); err != nil {
		err = fmt.Errorf("could not close pid file: %s", err)
	}
	return
}

func cgrRunPreload(ctx *context.Context, cfg *config.CGRConfig, loaderIDs string,
	loader *LoaderService, iLoaderSCh chan birpc.ClientConnector) (err error) {
	if !cfg.LoaderCfg().Enabled() {
		err = fmt.Errorf("<%s> not enabled but required by preload mechanism", utils.LoaderS)
		return
	}
	select {
	case ldrs := <-iLoaderSCh:
		iLoaderSCh <- ldrs
	case <-ctx.Done():
		return
	}

	var reply string
	for _, loaderID := range strings.Split(loaderIDs, utils.FieldsSep) {
		if err = loader.GetLoaderS().V1Load(ctx, &loaders.ArgsProcessFolder{
			ForceLock:   true, // force lock will unlock the file in case is locked and return error
			LoaderID:    loaderID,
			StopOnError: true,
		}, &reply); err != nil {
			err = fmt.Errorf("<%s> preload failed on loadID <%s> , err: <%s>", utils.LoaderS, loaderID, err)
			return
		}
	}
	return
}

// cgrStartFilterService fires up the FilterS
func cgrStartFilterService(ctx *context.Context, iFilterSCh chan *engine.FilterS,
	cacheS *engine.CacheS, connMgr *engine.ConnManager,
	cfg *config.CGRConfig, dm *engine.DataManager) {
	select {
	case <-cacheS.GetPrecacheChannel(utils.CacheFilters):
		iFilterSCh <- engine.NewFilterS(cfg, connMgr, dm)
	case <-ctx.Done():
	}
}

// cgrInitCacheS inits the CacheS and starts precaching as well as populating internal channel for RPC conns
func cgrInitCacheS(ctx *context.Context, shutdown context.CancelFunc,
	iCacheSCh chan birpc.ClientConnector, server *cores.Server,
	cfg *config.CGRConfig, dm *engine.DataManager, anz *AnalyzerService,
	cpS *engine.CapsStats) (chS *engine.CacheS) {
	chS = engine.NewCacheS(cfg, dm, cpS)
	go func() {
		if err := chS.Precache(ctx); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> could not init, error: %s", utils.CacheS, err.Error()))
			shutdown()
		}
	}()

	chSv1, _ := birpc.NewService(apis.NewCacheSv1(chS), "", false)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(chSv1)
	}
	var rpc birpc.ClientConnector = chSv1
	if anz.IsRunning() {
		rpc = anz.GetAnalyzerS().NewAnalyzerConnector(rpc, utils.MetaInternal, utils.EmptyString, utils.CacheS)
	}
	iCacheSCh <- rpc
	return
}

func cgrInitGuardianSv1(iGuardianSCh chan birpc.ClientConnector,
	server *cores.Server, anz *AnalyzerService) {
	// grdSv1 := v1.NewGuardianSv1()
	// if !cfg.DispatcherSCfg().Enabled {
	// server.RpcRegister(grdSv1)
	// }
	// var rpc birpc.ClientConnector = grdSv1
	// if anz.IsRunning() {
	// rpc = anz.GetAnalyzerS().NewAnalyzerConnector(rpc, utils.MetaInternal, utils.EmptyString, utils.GuardianS)
	// }
	// iGuardianSCh <- rpc
}

func cgrInitServiceManagerV1(iServMngrCh chan birpc.ClientConnector,
	srvMngr *servmanager.ServiceManager, server *cores.Server,
	anz *AnalyzerService) {
	// if !cfg.DispatcherSCfg().Enabled {
	// server.RpcRegister(v1.NewServiceManagerV1(srvMngr))
	// }
	// var rpc birpc.ClientConnector = srvMngr
	// if anz.IsRunning() {
	// rpc = anz.GetAnalyzerS().NewAnalyzerConnector(rpc, utils.MetaInternal, utils.EmptyString, utils.ServiceManager)
	// }
	// iServMngrCh <- rpc
}

func cgrInitConfigSv1(iConfigCh chan birpc.ClientConnector,
	cfg *config.CGRConfig, server *cores.Server, anz *AnalyzerService) {
	cfgSv1, _ := birpc.NewService(apis.NewConfigSv1(cfg), "", false)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(cfgSv1)
	}
	var rpc birpc.ClientConnector = cfgSv1
	if anz.IsRunning() {
		rpc = anz.GetAnalyzerS().NewAnalyzerConnector(rpc, utils.MetaInternal, utils.EmptyString, utils.ConfigSv1)
	}
	iConfigCh <- rpc
}

func startRPC(server *cores.Server, internalAdminSChan,
	internalCdrSChan, internalRsChan, internalStatSChan,
	internalAttrSChan, internalChargerSChan, internalThdSChan, internalRouteSChan,
	internalSessionSChan, internalAnalyzerSChan, internalDispatcherSChan,
	internalLoaderSChan, internalCacheSChan,
	internalEEsChan, internalRateSChan, internalActionSChan,
	internalAccountSChan chan birpc.ClientConnector,
	shdChan *utils.SyncedChan) {
	if !cfg.DispatcherSCfg().Enabled {
		select { // Any of the rpc methods will unlock listening to rpc requests
		// case cdrs := <-internalCdrSChan:
		// 	internalCdrSChan <- cdrs
		case smg := <-internalSessionSChan:
			internalSessionSChan <- smg
		case rls := <-internalRsChan:
			internalRsChan <- rls
		case statS := <-internalStatSChan:
			internalStatSChan <- statS
		case admS := <-internalAdminSChan:
			internalAdminSChan <- admS
		case attrS := <-internalAttrSChan:
			internalAttrSChan <- attrS
		case chrgS := <-internalChargerSChan:
			internalChargerSChan <- chrgS
		case thS := <-internalThdSChan:
			internalThdSChan <- thS
		case rtS := <-internalRouteSChan:
			internalRouteSChan <- rtS
		// case analyzerS := <-internalAnalyzerSChan:
		// 	internalAnalyzerSChan <- analyzerS
		case loaderS := <-internalLoaderSChan:
			internalLoaderSChan <- loaderS
		case chS := <-internalCacheSChan: // added in order to start the RPC before precaching is done
			internalCacheSChan <- chS
		// case eeS := <-internalEEsChan:
		// 	internalEEsChan <- eeS
		case rateS := <-internalRateSChan:
			internalRateSChan <- rateS
		case actionS := <-internalActionSChan:
			internalActionSChan <- actionS
		case accountS := <-internalAccountSChan:
			internalAccountSChan <- accountS
		case <-shdChan.Done():
			return
		}
	} else {
		select {
		case dispatcherS := <-internalDispatcherSChan:
			internalDispatcherSChan <- dispatcherS
		case <-shdChan.Done():
			return
		}
	}

	go server.ServeJSON(cfg.ListenCfg().RPCJSONListen, shdChan)
	go server.ServeGOB(cfg.ListenCfg().RPCGOBListen, shdChan)
	go server.ServeHTTP(
		cfg.ListenCfg().HTTPListen,
		cfg.HTTPCfg().JsonRPCURL,
		cfg.HTTPCfg().WSURL,
		cfg.HTTPCfg().UseBasicAuth,
		cfg.HTTPCfg().AuthUsers,
		shdChan,
	)
	if (len(cfg.ListenCfg().RPCGOBTLSListen) != 0 ||
		len(cfg.ListenCfg().RPCJSONTLSListen) != 0 ||
		len(cfg.ListenCfg().HTTPTLSListen) != 0) &&
		(len(cfg.TLSCfg().ServerCerificate) == 0 ||
			len(cfg.TLSCfg().ServerKey) == 0) {
		utils.Logger.Warning("WARNING: missing TLS certificate/key file!")
		return
	}
	if cfg.ListenCfg().RPCGOBTLSListen != utils.EmptyString {
		go server.ServeGOBTLS(
			cfg.ListenCfg().RPCGOBTLSListen,
			cfg.TLSCfg().ServerCerificate,
			cfg.TLSCfg().ServerKey,
			cfg.TLSCfg().CaCertificate,
			cfg.TLSCfg().ServerPolicy,
			cfg.TLSCfg().ServerName,
			shdChan,
		)
	}
	if cfg.ListenCfg().RPCJSONTLSListen != utils.EmptyString {
		go server.ServeJSONTLS(
			cfg.ListenCfg().RPCJSONTLSListen,
			cfg.TLSCfg().ServerCerificate,
			cfg.TLSCfg().ServerKey,
			cfg.TLSCfg().CaCertificate,
			cfg.TLSCfg().ServerPolicy,
			cfg.TLSCfg().ServerName,
			shdChan,
		)
	}
	if cfg.ListenCfg().HTTPTLSListen != utils.EmptyString {
		go server.ServeHTTPTLS(
			cfg.ListenCfg().HTTPTLSListen,
			cfg.TLSCfg().ServerCerificate,
			cfg.TLSCfg().ServerKey,
			cfg.TLSCfg().CaCertificate,
			cfg.TLSCfg().ServerPolicy,
			cfg.TLSCfg().ServerName,
			cfg.HTTPCfg().JsonRPCURL,
			cfg.HTTPCfg().WSURL,
			cfg.HTTPCfg().UseBasicAuth,
			cfg.HTTPCfg().AuthUsers,
			shdChan,
		)
	}
}
