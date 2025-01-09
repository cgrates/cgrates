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
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewDiameterAgent returns the Diameter Agent
func NewDiameterAgent(cfg *config.CGRConfig,
	connMgr *engine.ConnManager, caps *engine.Caps) *DiameterAgent {
	return &DiameterAgent{
		cfg:       cfg,
		connMgr:   connMgr,
		caps:      caps,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// DiameterAgent implements Agent interface
type DiameterAgent struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	stopChan chan struct{}

	da      *agents.DiameterAgent
	connMgr *engine.ConnManager
	caps    *engine.Caps

	lnet  string
	laddr string

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start should handle the sercive start
func (da *DiameterAgent) Start(shutdown chan struct{}, registry *servmanager.ServiceRegistry) error {
	fs, err := waitForServiceState(utils.StateServiceUP, utils.FilterS, registry,
		da.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}

	da.Lock()
	defer da.Unlock()
	return da.start(fs.(*FilterService).FilterS(), da.caps, shutdown)
}

func (da *DiameterAgent) start(filterS *engine.FilterS, caps *engine.Caps, shutdown chan struct{}) error {
	var err error
	da.da, err = agents.NewDiameterAgent(da.cfg, filterS, da.connMgr, caps)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed to initialize agent: %v",
			utils.DiameterAgent, err))
		return err
	}
	da.lnet = da.cfg.DiameterAgentCfg().ListenNet
	da.laddr = da.cfg.DiameterAgentCfg().Listen
	da.stopChan = make(chan struct{})
	go func(d *agents.DiameterAgent) {
		if err := d.ListenAndServe(da.stopChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
				utils.DiameterAgent, err))
			close(shutdown)
		}
	}(da.da)
	return nil
}

// Reload handles the change of config
func (da *DiameterAgent) Reload(shutdown chan struct{}, registry *servmanager.ServiceRegistry) (err error) {
	da.Lock()
	defer da.Unlock()
	if da.lnet == da.cfg.DiameterAgentCfg().ListenNet &&
		da.laddr == da.cfg.DiameterAgentCfg().Listen {
		return
	}
	close(da.stopChan)

	fs, err := waitForServiceState(utils.StateServiceUP, utils.FilterS, registry,
		da.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	return da.start(fs.(*FilterService).FilterS(), da.caps, shutdown)
}

// Shutdown stops the service
func (da *DiameterAgent) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	da.Lock()
	close(da.stopChan)
	da.da = nil
	da.Unlock()
	return // no shutdown for the momment
}

// ServiceName returns the service name
func (da *DiameterAgent) ServiceName() string {
	return utils.DiameterAgent
}

// ShouldRun returns if the service should be running
func (da *DiameterAgent) ShouldRun() bool {
	return da.cfg.DiameterAgentCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (da *DiameterAgent) StateChan(stateID string) chan struct{} {
	return da.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (da *DiameterAgent) IntRPCConn() birpc.ClientConnector {
	return da.intRPCconn
}
