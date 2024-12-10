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
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewDispatcherService returns the Dispatcher Service
func NewDispatcherService(cfg *config.CGRConfig, dm *DataDBService,
	filterSChan chan *engine.FilterS,
	clSChan chan *commonlisteners.CommonListenerS,
	connMgr *engine.ConnManager,
	srvIndexer *servmanager.ServiceIndexer) *DispatcherService {
	return &DispatcherService{
		cfg:         cfg,
		dm:          dm,
		filterSChan: filterSChan,
		clSChan:     clSChan,
		connMgr:     connMgr,
		srvsReload:  make(map[string]chan struct{}),
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// DispatcherService implements Service interface
type DispatcherService struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	dm          *DataDBService
	filterSChan chan *engine.FilterS

	dspS *dispatchers.DispatcherService
	cl   *commonlisteners.CommonListenerS

	connMgr    *engine.ConnManager
	cfg        *config.CGRConfig
	srvsReload map[string]chan struct{}

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the sercive start
func (dspS *DispatcherService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if dspS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	utils.Logger.Info("Starting CGRateS DispatcherS service.")
	dspS.cl = <-dspS.clSChan
	dspS.clSChan <- dspS.cl
	cacheS := dspS.srvIndexer.GetService(utils.CacheS).(*CacheService)
	if utils.StructChanTimeout(cacheS.StateChan(utils.StateServiceUP), dspS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.DispatcherS, utils.CacheS, utils.StateServiceUP)
	}
	if err = cacheS.WaitToPrecache(ctx,
		utils.CacheDispatcherProfiles,
		utils.CacheDispatcherHosts,
		utils.CacheDispatcherFilterIndexes); err != nil {
		return
	}
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, dspS.filterSChan); err != nil {
		return
	}
	var datadb *engine.DataManager
	if datadb, err = dspS.dm.WaitForDM(ctx); err != nil {
		return
	}
	anz := dspS.srvIndexer.GetService(utils.AnalyzerS).(*AnalyzerService)
	if utils.StructChanTimeout(anz.StateChan(utils.StateServiceUP), dspS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.DispatcherS, utils.AnalyzerS, utils.StateServiceUP)
	}

	dspS.Lock()
	defer dspS.Unlock()

	dspS.dspS = dispatchers.NewDispatcherService(datadb, dspS.cfg, filterS, dspS.connMgr)

	dspS.unregisterAllDispatchedSubsystems() // unregister all rpc services that can be dispatched

	srv, _ := engine.NewDispatcherService(dspS.dspS)
	// srv, _ := birpc.NewService(apis.NewDispatcherSv1(dspS.dspS), "", false)
	for _, s := range srv {
		dspS.cl.RpcRegister(s)
	}
	dspS.connMgr.EnableDispatcher(srv)
	// for the moment we dispable Apier through dispatcher
	// until we figured out a better sollution in case of gob server
	// dspS.server.SetDispatched()
	dspS.intRPCconn = anz.GetInternalCodec(srv, utils.DispatcherS)
	close(dspS.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (dspS *DispatcherService) Reload(*context.Context, context.CancelFunc) (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (dspS *DispatcherService) Shutdown() (err error) {
	dspS.Lock()
	defer dspS.Unlock()
	dspS.dspS.Shutdown()
	dspS.dspS = nil
	dspS.cl.RpcUnregisterName(utils.DispatcherSv1)
	dspS.cl.RpcUnregisterName(utils.AttributeSv1)

	dspS.unregisterAllDispatchedSubsystems()
	dspS.connMgr.DisableDispatcher()
	dspS.sync()
	return
}

// IsRunning returns if the service is running
func (dspS *DispatcherService) IsRunning() bool {
	dspS.RLock()
	defer dspS.RUnlock()
	return dspS.dspS != nil
}

// ServiceName returns the service name
func (dspS *DispatcherService) ServiceName() string {
	return utils.DispatcherS
}

// ShouldRun returns if the service should be running
func (dspS *DispatcherService) ShouldRun() bool {
	return dspS.cfg.DispatcherSCfg().Enabled
}

func (dspS *DispatcherService) unregisterAllDispatchedSubsystems() {
	dspS.cl.RpcUnregisterName(utils.AttributeSv1)
}

func (dspS *DispatcherService) RegisterShutdownChan(subsys string) (c chan struct{}) {
	c = make(chan struct{})
	dspS.Lock()
	dspS.srvsReload[subsys] = c
	dspS.Unlock()
	return
}

func (dspS *DispatcherService) UnregisterShutdownChan(subsys string) {
	dspS.Lock()
	if dspS.srvsReload[subsys] != nil {
		close(dspS.srvsReload[subsys])
	}
	delete(dspS.srvsReload, subsys)
	dspS.Unlock()
}

func (dspS *DispatcherService) sync() {
	for _, c := range dspS.srvsReload {
		c <- struct{}{}
	}
}

// StateChan returns signaling channel of specific state
func (dspS *DispatcherService) StateChan(stateID string) chan struct{} {
	return dspS.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (dspS *DispatcherService) IntRPCConn() birpc.ClientConnector {
	return dspS.intRPCconn
}
