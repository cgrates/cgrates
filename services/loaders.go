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

	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewLoaderService returns the Loader Service
func NewLoaderService(cfg *config.CGRConfig) *LoaderService {
	return &LoaderService{
		cfg:       cfg,
		stopChan:  make(chan struct{}),
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// LoaderService implements Service interface
type LoaderService struct {
	sync.RWMutex

	ldrs *loaders.LoaderS
	cl   *commonlisteners.CommonListenerS

	stopChan chan struct{}
	cfg      *config.CGRConfig

	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the service start
func (ldrs *LoaderService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
			utils.DataDB,
		},
		registry, ldrs.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	ldrs.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)

	ldrs.Lock()
	defer ldrs.Unlock()

	ldrs.ldrs = loaders.NewLoaderS(ldrs.cfg, dbs.DataManager(), fs.FilterS(), cms.ConnManager())

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
	cms.AddInternalConn(utils.LoaderS, srv)
	return
}

// Reload handles the change of config
func (ldrs *LoaderService) Reload(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.FilterS,
			utils.DataDB,
		},
		registry, ldrs.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	close(ldrs.stopChan)
	ldrs.stopChan = make(chan struct{})

	ldrs.RLock()
	defer ldrs.RUnlock()

	ldrs.ldrs.Reload(dbs.DataManager(), fs.FilterS(), cms.ConnManager())
	return ldrs.ldrs.ListenAndServe(ldrs.stopChan)
}

// Shutdown stops the service
func (ldrs *LoaderService) Shutdown(_ *servmanager.ServiceRegistry) (_ error) {
	ldrs.Lock()
	ldrs.ldrs = nil
	close(ldrs.stopChan)
	ldrs.cl.RpcUnregisterName(utils.LoaderSv1)
	ldrs.Unlock()
	return
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
