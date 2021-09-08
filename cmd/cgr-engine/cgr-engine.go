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
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/loaders"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/services"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	cgrEngineFlags    = flag.NewFlagSet(utils.CgrEngine, flag.ContinueOnError)
	cfgPath           = cgrEngineFlags.String(utils.CfgPathCgr, utils.ConfigPath, "Configuration directory path.")
	version           = cgrEngineFlags.Bool(utils.VersionCgr, false, "Prints the application version.")
	checkConfig       = cgrEngineFlags.Bool(utils.CheckCfgCgr, false, "Verify the config without starting the engine")
	pidFile           = cgrEngineFlags.String(utils.PidCgr, utils.EmptyString, "Write pid file")
	httpPprofPath     = cgrEngineFlags.String(utils.HttpPrfPthCgr, utils.EmptyString, "http address used for program profiling")
	cpuProfDir        = cgrEngineFlags.String(utils.CpuProfDirCgr, utils.EmptyString, "write cpu profile to files")
	memProfDir        = cgrEngineFlags.String(utils.MemProfDirCgr, utils.EmptyString, "write memory profile to file")
	memProfInterval   = cgrEngineFlags.Duration(utils.MemProfIntervalCgr, 5*time.Second, "Time between memory profile saves")
	memProfNrFiles    = cgrEngineFlags.Int(utils.MemProfNrFilesCgr, 1, "Number of memory profile to write")
	scheduledShutdown = cgrEngineFlags.String(utils.ScheduledShutdownCgr, utils.EmptyString, "shutdown the engine after this duration")
	singlecpu         = cgrEngineFlags.Bool(utils.SingleCpuCgr, false, "Run on single CPU core")
	syslogger         = cgrEngineFlags.String(utils.LoggerCfg, utils.EmptyString, "logger <*syslog|*stdout>")
	nodeID            = cgrEngineFlags.String(utils.NodeIDCfg, utils.EmptyString, "The node ID of the engine")
	logLevel          = cgrEngineFlags.Int(utils.LogLevelCfg, -1, "Log level (0-emergency to 7-debug)")
	preload           = cgrEngineFlags.String(utils.PreloadCgr, utils.EmptyString, "LoaderIDs used to load the data before the engine starts")

	cfg *config.CGRConfig
)

// startFilterService fires up the FilterS
func startFilterService(filterSChan chan *engine.FilterS, cacheS *engine.CacheS, connMgr *engine.ConnManager, cfg *config.CGRConfig,
	dm *engine.DataManager) {
	<-cacheS.GetPrecacheChannel(utils.CacheFilters)
	filterSChan <- engine.NewFilterS(cfg, connMgr, dm)
}

// initCacheS inits the CacheS and starts precaching as well as populating internal channel for RPC conns
func initCacheS(internalCacheSChan chan birpc.ClientConnector,
	server *cores.Server, dm *engine.DataManager, shdChan *utils.SyncedChan,
	anz *services.AnalyzerService,
	cpS *engine.CapsStats) (chS *engine.CacheS) {
	chS = engine.NewCacheS(cfg, dm, cpS)
	go func() {
		if err := chS.Precache(context.TODO()); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> could not init, error: %s", utils.CacheS, err.Error()))
			shdChan.CloseOnce()
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
	internalCacheSChan <- rpc
	return
}

func initGuardianSv1(internalGuardianSChan chan birpc.ClientConnector, server *cores.Server,
	anz *services.AnalyzerService) {
	// grdSv1 := v1.NewGuardianSv1()
	// if !cfg.DispatcherSCfg().Enabled {
	// server.RpcRegister(grdSv1)
	// }
	// var rpc birpc.ClientConnector = grdSv1
	// if anz.IsRunning() {
	// rpc = anz.GetAnalyzerS().NewAnalyzerConnector(rpc, utils.MetaInternal, utils.EmptyString, utils.GuardianS)
	// }
	// internalGuardianSChan <- rpc
}

func initServiceManagerV1(internalServiceManagerChan chan birpc.ClientConnector,
	srvMngr *servmanager.ServiceManager, server *cores.Server,
	anz *services.AnalyzerService) {
	// if !cfg.DispatcherSCfg().Enabled {
	// server.RpcRegister(v1.NewServiceManagerV1(srvMngr))
	// }
	// var rpc birpc.ClientConnector = srvMngr
	// if anz.IsRunning() {
	// rpc = anz.GetAnalyzerS().NewAnalyzerConnector(rpc, utils.MetaInternal, utils.EmptyString, utils.ServiceManager)
	// }
	// internalServiceManagerChan <- rpc
}

func initConfigSv1(internalConfigChan chan birpc.ClientConnector,
	server *cores.Server, anz *services.AnalyzerService) {
	cfgSv1, _ := birpc.NewService(apis.NewConfigSv1(cfg), "", false)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(cfgSv1)
	}
	var rpc birpc.ClientConnector = cfgSv1
	if anz.IsRunning() {
		rpc = anz.GetAnalyzerS().NewAnalyzerConnector(rpc, utils.MetaInternal, utils.EmptyString, utils.ConfigSv1)
	}
	internalConfigChan <- rpc
}

/*
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
*/

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

func singnalHandler(shdWg *sync.WaitGroup, shdChan *utils.SyncedChan) {
	shutdownSignal := make(chan os.Signal, 1)
	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt,
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	for {
		select {
		case <-shdChan.Done():
			shdWg.Done()
			return
		case <-shutdownSignal:
			shdChan.CloseOnce()
			shdWg.Done()
			return
		case <-reloadSignal:
			//  do it in it's own goroutine in order to not block the signal handler with the reload functionality
			go func() {
				var reply string
				if err := config.CgrConfig().V1ReloadConfig(context.Background(),
					new(config.ReloadArgs), &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("Error reloading configuration: <%s>", err))
				}
			}()
		}
	}
}

func runPreload(loader *services.LoaderService, internalLoaderSChan chan birpc.ClientConnector,
	shdChan *utils.SyncedChan) {
	if !cfg.LoaderCfg().Enabled() {
		utils.Logger.Err(fmt.Sprintf("<%s> not enabled but required by preload mechanism", utils.LoaderS))
		shdChan.CloseOnce()
		return
	}
	ldrs := <-internalLoaderSChan
	internalLoaderSChan <- ldrs

	var reply string
	for _, loaderID := range strings.Split(*preload, utils.FieldsSep) {
		if err := loader.GetLoaderS().V1Load(context.Background(), &loaders.ArgsProcessFolder{
			ForceLock:   true, // force lock will unlock the file in case is locked and return error
			LoaderID:    loaderID,
			StopOnError: true,
		}, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> preload failed on loadID <%s> , err: <%s>", utils.LoaderS, loaderID, err.Error()))
			shdChan.CloseOnce()
			return
		}
	}
}

func main() {
	if err := cgrEngineFlags.Parse(os.Args[1:]); err != nil {
		return
	}

	srvManager.StartServices()
	// Start FilterS
	go startFilterService(filterSChan, cacheS, connManager,
		cfg, dmService.GetDM())

	initServiceManagerV1(internalServeManagerChan, srvManager, server, anz)

	initConfigSv1(internalConfigChan, server, anz)

	if *preload != utils.EmptyString {
		runPreload(ldrs, internalLoaderSChan, shdChan)
	}

	// Serve rpc connections
	// go startRPC(server, internalAdminSChan, internalCDRServerChan,
	// 	internalResourceSChan, internalStatSChan,
	// 	internalAttributeSChan, internalChargerSChan, internalThresholdSChan,
	// 	internalRouteSChan, internalSessionSChan, internalAnalyzerSChan,
	// 	internalDispatcherSChan, internalLoaderSChan,
	// 	internalCacheSChan, internalEEsChan, internalRateSChan, internalActionSChan,
	// 	internalAccountSChan, shdChan)

	<-shdChan.Done()
	shtdDone := make(chan struct{})
	go func() {
		shdWg.Wait()
		close(shtdDone)
	}()
	select {
	case <-shtdDone:
	case <-time.After(cfg.CoreSCfg().ShutdownTimeout):
		utils.Logger.Err(fmt.Sprintf("<%s> Failed to shutdown all subsystems in the given time",
			utils.ServiceManager))
	}

	if *memProfDir != utils.EmptyString { // write last memory profiling
		cores.MemProfFile(path.Join(*memProfDir, utils.MemProfFileCgr))
	}
	if *pidFile != utils.EmptyString {
		if err := os.Remove(*pidFile); err != nil {
			utils.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	utils.Logger.Info("<CoreS> stopped all components. CGRateS shutdown!")
}
