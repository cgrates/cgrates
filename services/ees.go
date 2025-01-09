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
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewEventExporterService constructs EventExporterService
func NewEventExporterService(cfg *config.CGRConfig,
	connMgr *engine.ConnManager) *EventExporterService {
	return &EventExporterService{
		cfg:       cfg,
		connMgr:   connMgr,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// EventExporterService is the service structure for EventExporterS
type EventExporterService struct {
	mu sync.RWMutex

	eeS *ees.EeS
	cl  *commonlisteners.CommonListenerS

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
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
func (es *EventExporterService) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) error {
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
func (es *EventExporterService) Start(_ chan struct{}, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.FilterS,
			utils.AnalyzerS,
		},
		registry, es.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	es.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	fs := srvDeps[utils.FilterS].(*FilterService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	es.mu.Lock()
	defer es.mu.Unlock()

	es.eeS, err = ees.NewEventExporterS(es.cfg, fs.FilterS(), es.connMgr)
	if err != nil {
		return err
	}

	srv, _ := engine.NewServiceWithPing(es.eeS, utils.EeSv1, utils.V1Prfx)
	// srv, _ := birpc.NewService(es.rpc, "", false)
	if !es.cfg.DispatcherSCfg().Enabled {
		es.cl.RpcRegister(srv)
	}

	es.intRPCconn = anz.GetInternalCodec(srv, utils.EEs)
	return nil
}

// StateChan returns signaling channel of specific state
func (es *EventExporterService) StateChan(stateID string) chan struct{} {
	return es.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (es *EventExporterService) IntRPCConn() birpc.ClientConnector {
	return es.intRPCconn
}
