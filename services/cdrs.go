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
	"runtime"
	"sync"

	"github.com/cgrates/cgrates/cdrs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCDRServer returns the CDR Server
func NewCDRServer(cfg *config.CGRConfig) *CDRService {
	return &CDRService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// CDRService implements Service interface
type CDRService struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	cdrS      *cdrs.CDRServer
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (cs *CDRService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
			utils.DataDB,
			utils.StorDB,
		},
		registry, cs.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	sdbs := srvDeps[utils.StorDB].(*StorDBService).DB()

	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.cdrS = cdrs.NewCDRServer(cs.cfg, dbs.DataManager(), fs, cms.ConnManager(), sdbs)
	runtime.Gosched()
	srv, err := engine.NewServiceWithPing(cs.cdrS, utils.CDRsV1, utils.V1Prfx)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.CDRServer, srv)
	return
}

// Reload handles the change of config
func (cs *CDRService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return
}

// Shutdown stops the service
func (cs *CDRService) Shutdown(registry *servmanager.ServiceRegistry) (err error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.cdrS = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.CDRsV1)
	return
}

// ServiceName returns the service name
func (cs *CDRService) ServiceName() string {
	return utils.CDRServer
}

// ShouldRun returns if the service should be running
func (cs *CDRService) ShouldRun() bool {
	return cs.cfg.CdrsCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (cs *CDRService) StateChan(stateID string) chan struct{} {
	return cs.stateDeps.StateChan(stateID)
}
