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
		CfgPath:           fs.String(utils.CfgPathCgr, utils.ConfigPath, "Configuration directory path."),
		Version:           fs.Bool(utils.VersionCgr, false, "Prints the application version."),
		PidFile:           fs.String(utils.PidCgr, utils.EmptyString, "Write pid file"),
		HttpPrfPath:       fs.String(utils.HttpPrfPthCgr, utils.EmptyString, "http address used for program profiling"),
		CpuPrfDir:         fs.String(utils.CpuProfDirCgr, utils.EmptyString, "write cpu profile to files"),
		MemPrfDir:         fs.String(utils.MemProfDirCgr, utils.EmptyString, "write memory profile to file"),
		MemPrfInterval:    fs.Duration(utils.MemProfIntervalCgr, 5*time.Second, "Time between memory profile saves"),
		MemPrfNoF:         fs.Int(utils.MemProfNrFilesCgr, 1, "Number of memory profile to write"),
		ScheduledShutDown: fs.String(utils.ScheduledShutdownCgr, utils.EmptyString, "shutdown the engine after this duration"),
		Singlecpu:         fs.Bool(utils.SingleCpuCgr, false, "Run on single CPU core"),
		SysLogger:         fs.String(utils.LoggerCfg, utils.EmptyString, "logger <*syslog|*stdout>"),
		NodeID:            fs.String(utils.NodeIDCfg, utils.EmptyString, "The node ID of the engine"),
		LogLevel:          fs.Int(utils.LogLevelCfg, -1, "Log level (0-emergency to 7-debug)"),
		Preload:           fs.String(utils.PreloadCgr, utils.EmptyString, "LoaderIDs used to load the data before the engine starts"),
		CheckConfig:       fs.Bool(utils.CheckCfgCgr, false, "Verify the config without starting the engine"),
	}
}

type CGREngineFlags struct {
	*flag.FlagSet

	CfgPath           *string
	Version           *bool
	PidFile           *string
	HttpPrfPath       *string
	CpuPrfDir         *string
	MemPrfDir         *string
	MemPrfInterval    *time.Duration
	MemPrfNoF         *int
	ScheduledShutDown *string
	Singlecpu         *bool
	SysLogger         *string
	NodeID            *string
	LogLevel          *int
	Preload           *string
	CheckConfig       *bool
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
	loader *LoaderService) (err error) {
	if !cfg.LoaderCfg().Enabled() {
		err = fmt.Errorf("<%s> not enabled but required by preload mechanism", utils.LoaderS)
		return
	}
	ch := loader.GetRPCChan()
	select {
	case ldrs := <-ch:
		ch <- ldrs
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
	cacheSCh chan *engine.CacheS, connMgr *engine.ConnManager,
	cfg *config.CGRConfig, dm *engine.DataManager) {
	var cacheS *engine.CacheS
	select {
	case cacheS = <-cacheSCh:
	case <-ctx.Done():
		return
	}
	select {
	case <-cacheS.GetPrecacheChannel(utils.CacheFilters):
		iFilterSCh <- engine.NewFilterS(cfg, connMgr, dm)
	case <-ctx.Done():
	}
}

func cgrInitGuardianSv1(iGuardianSCh chan birpc.ClientConnector,
	server *cores.Server, anz *AnalyzerService) {
	// grdSv1 := v1.NewGuardianSv1()
	// if !cfg.DispatcherSCfg().Enabled {
	// server.RpcRegister(grdSv1)
	// }
	// iGuardianSCh <- anz.GetInternalCodec(grdSv1,  utils.GuardianS)
}

func cgrInitServiceManagerV1(iServMngrCh chan birpc.ClientConnector,
	srvMngr *servmanager.ServiceManager, server *cores.Server,
	anz *AnalyzerService) {
	// if !cfg.DispatcherSCfg().Enabled {
	// server.RpcRegister(v1.NewServiceManagerV1(srvMngr))
	// }
	// iServMngrCh <- anz.GetInternalCodec(srvMngr,  utils.ServiceManager)
}

func cgrInitConfigSv1(iConfigCh chan birpc.ClientConnector,
	cfg *config.CGRConfig, server *cores.Server, anz *AnalyzerService) {
	cfgSv1, _ := birpc.NewService(apis.NewConfigSv1(cfg), "", false)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(cfgSv1)
	}
	iConfigCh <- anz.GetInternalCodec(cfgSv1, utils.ConfigSv1)
}

func cgrStartRPC(ctx *context.Context, shtdwnEngine context.CancelFunc,
	cfg *config.CGRConfig, server *cores.Server,
	internalAdminSChan, internalCdrSChan, internalRsChan, internalStatSChan,
	internalAttrSChan, internalChargerSChan, internalThdSChan, internalRouteSChan,
	internalSessionSChan, internalAnalyzerSChan, internalDispatcherSChan,
	internalLoaderSChan, internalCacheSChan, internalEEsChan, internalRateSChan,
	internalActionSChan, internalAccountSChan chan birpc.ClientConnector) {
	if !cfg.DispatcherSCfg().Enabled {
		select { // Any of the rpc methods will unlock listening to rpc requests
		case cdrs := <-internalCdrSChan:
			internalCdrSChan <- cdrs
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
		case analyzerS := <-internalAnalyzerSChan:
			internalAnalyzerSChan <- analyzerS
		case loaderS := <-internalLoaderSChan:
			internalLoaderSChan <- loaderS
		case chS := <-internalCacheSChan: // added in order to start the RPC before precaching is done
			internalCacheSChan <- chS
		case eeS := <-internalEEsChan:
			internalEEsChan <- eeS
		case rateS := <-internalRateSChan:
			internalRateSChan <- rateS
		case actionS := <-internalActionSChan:
			internalActionSChan <- actionS
		case accountS := <-internalAccountSChan:
			internalAccountSChan <- accountS
		case <-ctx.Done():
			return
		}
	} else {
		select {
		case dispatcherS := <-internalDispatcherSChan:
			internalDispatcherSChan <- dispatcherS
		case <-ctx.Done():
			return
		}
	}
	server.StartServer(ctx, shtdwnEngine, cfg)
}

func waitForFilterS(ctx *context.Context, fsCh chan *engine.FilterS) (filterS *engine.FilterS, err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case filterS = <-fsCh:
		fsCh <- filterS
	}
	return
}

func getCacheS(ctx *context.Context, csCh chan *engine.CacheS) (cS *engine.CacheS, err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case cS = <-csCh:
		csCh <- cS
	}
	return
}
