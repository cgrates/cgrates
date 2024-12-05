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
	"fmt"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerService returns the Analyzer Service
func NewAnalyzerService(cfg *config.CGRConfig, clSChan chan *commonlisteners.CommonListenerS,
	filterSChan chan *engine.FilterS,
	anzChan chan *AnalyzerService,
	srvIndexer *servmanager.ServiceIndexer) *AnalyzerService {
	return &AnalyzerService{
		cfg:         cfg,
		clSChan:     clSChan,
		filterSChan: filterSChan,
		anzChan:     anzChan,
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// AnalyzerService implements Service interface
type AnalyzerService struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	filterSChan chan *engine.FilterS
	anzChan     chan *AnalyzerService

	anz *analyzers.AnalyzerS
	cl  *commonlisteners.CommonListenerS

	cancelFunc context.CancelFunc
	cfg        *config.CGRConfig

	intRPCconn birpc.ClientConnector       // share the API object implementing API calls for internal
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes

}

// Start should handle the sercive start
func (anz *AnalyzerService) Start(ctx *context.Context, shtDwn context.CancelFunc) (err error) {
	if anz.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	anz.cl = <-anz.clSChan
	anz.clSChan <- anz.cl

	anz.Lock()
	defer anz.Unlock()
	if anz.anz, err = analyzers.NewAnalyzerS(anz.cfg); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.AnalyzerS, err.Error()))
		return
	}
	anz.anzChan <- anz
	anzCtx, cancel := context.WithCancel(ctx)
	anz.cancelFunc = cancel
	go func(a *analyzers.AnalyzerS) {
		if err := a.ListenAndServe(anzCtx); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.AnalyzerS, err.Error()))
			shtDwn()
		}
	}(anz.anz)
	anz.cl.SetAnalyzer(anz.anz)
	go anz.start(ctx)
	close(anz.stateDeps.StateChan(utils.StateServiceUP))
	return
}

func (anz *AnalyzerService) start(ctx *context.Context) {
	fS, err := waitForFilterS(ctx, anz.filterSChan)
	if err != nil {
		return
	}

	if !anz.IsRunning() {
		return
	}
	anz.Lock()
	anz.anz.SetFilterS(fS)

	srv, _ := engine.NewService(anz.anz)
	// srv, _ := birpc.NewService(apis.NewAnalyzerSv1(anz.anz), "", false)
	if !anz.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			anz.cl.RpcRegister(s)
		}
	}
	anz.Unlock()
}

// Reload handles the change of config
func (anz *AnalyzerService) Reload(*context.Context, context.CancelFunc) (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (anz *AnalyzerService) Shutdown() (err error) {
	anz.Lock()
	anz.cancelFunc()
	anz.cl.SetAnalyzer(nil)

	// Close the channel before making it nil to prevent stale goroutines
	// in case there are other services waiting on AnalyzerS.
	close(anz.anzChan)

	anz.anzChan = nil
	anz.anz.Shutdown()
	anz.anz = nil
	anz.Unlock()
	anz.cl.RpcUnregisterName(utils.AnalyzerSv1)
	return
}

// IsRunning returns if the service is running
func (anz *AnalyzerService) IsRunning() bool {
	anz.RLock()
	defer anz.RUnlock()
	return anz.anz != nil
}

// ServiceName returns the service name
func (anz *AnalyzerService) ServiceName() string {
	return utils.AnalyzerS
}

// ShouldRun returns if the service should be running
func (anz *AnalyzerService) ShouldRun() bool {
	return anz.cfg.AnalyzerSCfg().Enabled
}

// GetInternalCodec returns the connection wrapped in analyzer connector
func (anz *AnalyzerService) GetInternalCodec(c birpc.ClientConnector, to string) birpc.ClientConnector {
	if !anz.IsRunning() {
		return c
	}
	return anz.anz.NewAnalyzerConnector(c, utils.MetaInternal, utils.EmptyString, to)
}

// StateChan returns signaling channel of specific state
func (anz *AnalyzerService) StateChan(stateID string) chan struct{} {
	return anz.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (anz *AnalyzerService) IntRPCConn() birpc.ClientConnector {
	return anz.intRPCconn
}
