/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package services

import (
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/tpes"
	"github.com/cgrates/cgrates/utils"
)

// NewTPeService is the constructor for the TpeService
func NewTPeService(cfg *config.CGRConfig) *TPeService {
	return &TPeService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// TypeService implements Service interface
type TPeService struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	tpes      *tpes.TPeS
	srv       *birpc.Service
	stopChan  chan struct{}
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the service start
func (ts *TPeService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.DataDB,
		},
		registry, ts.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	dbs := srvDeps[utils.DataDB].(*DataDBService).DataManager()

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.tpes = tpes.NewTPeS(ts.cfg, dbs, cm)
	ts.stopChan = make(chan struct{})
	ts.srv, _ = birpc.NewService(apis.NewTPeSv1(ts.tpes), utils.EmptyString, false)
	cl.RpcRegister(ts.srv)
	return
}

// Reload handles the change of config
func (ts *TPeService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return
}

// Shutdown stops the service
func (ts *TPeService) Shutdown(registry *servmanager.ServiceRegistry) (err error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.srv = nil
	close(ts.stopChan)
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.TPeSv1)
	return
}

// ServiceName returns the service name
func (ts *TPeService) ServiceName() string {
	return utils.TPeS
}

// ShouldRun returns if the service should be running
func (ts *TPeService) ShouldRun() bool {
	return ts.cfg.TpeSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (ts *TPeService) StateChan(stateID string) chan struct{} {
	return ts.stateDeps.StateChan(stateID)
}
