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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/registrarc"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCommonListenerService instantiates a new CommonListenerService.
func NewCommonListenerService(cfg *config.CGRConfig, caps *engine.Caps) *CommonListenerService {
	return &CommonListenerService{
		cfg:       cfg,
		caps:      caps,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// CommonListenerService implements Service interface.
type CommonListenerService struct {
	mu sync.RWMutex

	cls *commonlisteners.CommonListenerS

	caps *engine.Caps
	cfg  *config.CGRConfig

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start handles the service start.
func (cl *CommonListenerService) Start(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.cls = commonlisteners.NewCommonListenerS(cl.caps)
	if len(cl.cfg.HTTPCfg().RegistrarSURL) != 0 {
		cl.cls.RegisterHTTPFunc(cl.cfg.HTTPCfg().RegistrarSURL, registrarc.Registrar)
	}
	if cl.cfg.ConfigSCfg().Enabled {
		cl.cls.RegisterHTTPFunc(cl.cfg.ConfigSCfg().URL, config.HandlerConfigS)
	}
	return nil
}

// Reload handles the config changes.
func (cl *CommonListenerService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) error {
	return nil
}

// Shutdown stops the service.
func (cl *CommonListenerService) Shutdown(_ *servmanager.ServiceRegistry) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.cls = nil
	return nil
}

// ServiceName returns the service name
func (cl *CommonListenerService) ServiceName() string {
	return utils.CommonListenerS
}

// ShouldRun returns if the service should be running.
func (cl *CommonListenerService) ShouldRun() bool {
	return true
}

// StateChan returns signaling channel of specific state
func (cl *CommonListenerService) StateChan(stateID string) chan struct{} {
	return cl.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (cl *CommonListenerService) IntRPCConn() birpc.ClientConnector {
	return cl.intRPCconn
}

// CLS returns the CommonListenerS object.
func (cl *CommonListenerService) CLS() *commonlisteners.CommonListenerS {
	return cl.cls
}
