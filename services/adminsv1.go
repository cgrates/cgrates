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

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAPIerSv1Service returns the APIerSv1 Service
func NewAdminSv1Service(cfg *config.CGRConfig) *AdminSv1Service {
	return &AdminSv1Service{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// AdminSv1Service implements Service interface
type AdminSv1Service struct {
	sync.RWMutex
	cfg       *config.CGRConfig
	api       *apis.AdminSv1
	cl        *commonlisteners.CommonListenerS
	stopChan  chan struct{}
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (apiService *AdminSv1Service) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
			utils.DataDB,
			utils.StorDB,
		},
		registry, apiService.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	apiService.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	sdbs := srvDeps[utils.StorDB].(*StorDBService)

	apiService.Lock()
	defer apiService.Unlock()

	apiService.api = apis.NewAdminSv1(apiService.cfg, dbs.DataManager(), cms.ConnManager(), fs.FilterS(), sdbs.DB())

	srv, _ := engine.NewService(apiService.api)
	// srv, _ := birpc.NewService(apiService.api, "", false)

	if !apiService.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			apiService.cl.RpcRegister(s)
		}
		rpl, _ := engine.NewService(apis.NewReplicatorSv1(dbs.DataManager(), apiService.api))
		for _, s := range rpl {
			apiService.cl.RpcRegister(s)
		}
	}
	cms.AddInternalConn(utils.AdminS, srv)
	return
}

// Reload handles the change of config
func (apiService *AdminSv1Service) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return
}

// Shutdown stops the service
func (apiService *AdminSv1Service) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	apiService.Lock()
	// close(apiService.stopChan)
	apiService.api = nil
	apiService.cl.RpcUnregisterName(utils.AdminSv1)
	apiService.Unlock()
	return
}

// ServiceName returns the service name
func (apiService *AdminSv1Service) ServiceName() string {
	return utils.AdminS
}

// ShouldRun returns if the service should be running
func (apiService *AdminSv1Service) ShouldRun() bool {
	return apiService.cfg.AdminSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (apiService *AdminSv1Service) StateChan(stateID string) chan struct{} {
	return apiService.stateDeps.StateChan(stateID)
}
