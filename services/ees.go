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
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewEventExporterService constructs EventExporterService
func NewEventExporterService(cfg *config.CGRConfig) *EventExporterService {
	return &EventExporterService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// EventExporterService is the service structure for EventExporterS
type EventExporterService struct {
	mu  sync.RWMutex
	cfg *config.CGRConfig

	eeS *ees.EeS
	cl  *commonlisteners.CommonListenerS

	stateDeps *StateDependencies // channel subscriptions for state changes
}

// ServiceName returns the service name
func (es *EventExporterService) ServiceName() string {
	return utils.EEs
}

// ShouldRun returns if the service should be running
func (es *EventExporterService) ShouldRun() (should bool) {
	return es.cfg.EEsCfg().Enabled
}

// Reload handles the change of config
func (es *EventExporterService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) error {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.eeS.ClearExporterCache()
	return es.eeS.SetupExporterCache()
}

// Shutdown stops the service
func (es *EventExporterService) Shutdown(_ *servmanager.ServiceRegistry) error {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.eeS.ClearExporterCache()
	es.eeS = nil
	es.cl.RpcUnregisterName(utils.EeSv1)
	return nil
}

// Start should handle the service start
func (es *EventExporterService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
		},
		registry, es.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	es.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()

	es.mu.Lock()
	defer es.mu.Unlock()

	es.eeS, err = ees.NewEventExporterS(es.cfg, fs, cms.ConnManager())
	if err != nil {
		return err
	}

	srv, _ := engine.NewServiceWithPing(es.eeS, utils.EeSv1, utils.V1Prfx)
	// srv, _ := birpc.NewService(es.rpc, "", false)
	if !es.cfg.DispatcherSCfg().Enabled {
		es.cl.RpcRegister(srv)
	}
	cms.AddInternalConn(utils.EEs, srv)
	return nil
}

// StateChan returns signaling channel of specific state
func (es *EventExporterService) StateChan(stateID string) chan struct{} {
	return es.stateDeps.StateChan(stateID)
}
