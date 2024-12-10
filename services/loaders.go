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

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewLoaderService returns the Loader Service
func NewLoaderService(cfg *config.CGRConfig,
	filterSChan chan *engine.FilterS,
	connMgr *engine.ConnManager,
	srvIndexer *servmanager.ServiceIndexer) *LoaderService {
	return &LoaderService{
		cfg:         cfg,
		filterSChan: filterSChan,
		connMgr:     connMgr,
		stopChan:    make(chan struct{}),
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// LoaderService implements Service interface
type LoaderService struct {
	sync.RWMutex

	filterSChan chan *engine.FilterS

	ldrs *loaders.LoaderS
	cl   *commonlisteners.CommonListenerS

	stopChan chan struct{}
	connMgr  *engine.ConnManager
	cfg      *config.CGRConfig

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the service start
func (ldrs *LoaderService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if ldrs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	cls := ldrs.srvIndexer.GetService(utils.CommonListenerS).(*CommonListenerService)
	if utils.StructChanTimeout(cls.StateChan(utils.StateServiceUP), ldrs.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.LoaderS, utils.CommonListenerS, utils.StateServiceUP)
	}
	ldrs.cl = cls.CLS()
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, ldrs.filterSChan); err != nil {
		return
	}
	dbs := ldrs.srvIndexer.GetService(utils.DataDB).(*DataDBService)
	if utils.StructChanTimeout(dbs.StateChan(utils.StateServiceUP), ldrs.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.LoaderS, utils.DataDB, utils.StateServiceUP)
	}
	anz := ldrs.srvIndexer.GetService(utils.AnalyzerS).(*AnalyzerService)
	if utils.StructChanTimeout(anz.StateChan(utils.StateServiceUP), ldrs.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.LoaderS, utils.AnalyzerS, utils.StateServiceUP)
	}

	ldrs.Lock()
	defer ldrs.Unlock()

	ldrs.ldrs = loaders.NewLoaderS(ldrs.cfg, dbs.DataManager(), filterS, ldrs.connMgr)

	if !ldrs.ldrs.Enabled() {
		return
	}
	if err = ldrs.ldrs.ListenAndServe(ldrs.stopChan); err != nil {
		return
	}
	srv, _ := engine.NewService(ldrs.ldrs)
	// srv, _ := birpc.NewService(apis.NewLoaderSv1(ldrs.ldrs), "", false)
	if !ldrs.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			ldrs.cl.RpcRegister(s)
		}
	}
	ldrs.intRPCconn = anz.GetInternalCodec(srv, utils.LoaderS)
	close(ldrs.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (ldrs *LoaderService) Reload(ctx *context.Context, _ context.CancelFunc) error {
	filterS, err := waitForFilterS(ctx, ldrs.filterSChan)
	if err != nil {
		return err
	}
	dbs := ldrs.srvIndexer.GetService(utils.DataDB).(*DataDBService)
	if utils.StructChanTimeout(dbs.StateChan(utils.StateServiceUP), ldrs.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.LoaderS, utils.DataDB, utils.StateServiceUP)
	}
	close(ldrs.stopChan)
	ldrs.stopChan = make(chan struct{})

	ldrs.RLock()
	defer ldrs.RUnlock()

	ldrs.ldrs.Reload(dbs.DataManager(), filterS, ldrs.connMgr)
	return ldrs.ldrs.ListenAndServe(ldrs.stopChan)
}

// Shutdown stops the service
func (ldrs *LoaderService) Shutdown() (_ error) {
	ldrs.Lock()
	ldrs.ldrs = nil
	close(ldrs.stopChan)
	ldrs.cl.RpcUnregisterName(utils.LoaderSv1)
	ldrs.Unlock()
	return
}

// IsRunning returns if the service is running
func (ldrs *LoaderService) IsRunning() bool {
	ldrs.RLock()
	defer ldrs.RUnlock()
	return ldrs.ldrs != nil && ldrs.ldrs.Enabled()
}

// ServiceName returns the service name
func (ldrs *LoaderService) ServiceName() string {
	return utils.LoaderS
}

// ShouldRun returns if the service should be running
func (ldrs *LoaderService) ShouldRun() bool {
	return ldrs.cfg.LoaderCfg().Enabled()
}

// GetLoaderS returns the initialized LoaderService
func (ldrs *LoaderService) GetLoaderS() *loaders.LoaderS {
	return ldrs.ldrs
}

// StateChan returns signaling channel of specific state
func (ldrs *LoaderService) StateChan(stateID string) chan struct{} {
	return ldrs.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (ldrs *LoaderService) IntRPCConn() birpc.ClientConnector {
	return ldrs.intRPCconn
}
