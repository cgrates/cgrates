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

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewDiameterAgent returns the Diameter Agent
func NewDiameterAgent(cfg *config.CGRConfig) *DiameterAgent {
	return &DiameterAgent{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// DiameterAgent implements Agent interface
type DiameterAgent struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	stopChan chan struct{}

	da *agents.DiameterAgent

	lnet  string
	laddr string

	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (s *DiameterAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CapS,
			utils.ConnManager,
			utils.FilterS,
		},
		registry, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	caps := srvDeps[utils.CapS].(*CapService).Caps()
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()

	s.Lock()
	defer s.Unlock()
	return s.start(fs, cm, caps, shutdown)
}

func (s *DiameterAgent) start(filterS *engine.FilterS, cm *engine.ConnManager, caps *engine.Caps,
	shutdown *utils.SyncedChan) error {
	var err error
	s.da, err = agents.NewDiameterAgent(s.cfg, filterS, cm, caps)
	if err != nil {
		return err
	}
	s.lnet = s.cfg.DiameterAgentCfg().ListenNet
	s.laddr = s.cfg.DiameterAgentCfg().Listen
	s.stopChan = make(chan struct{})
	go func(d *agents.DiameterAgent) {
		if err := d.ListenAndServe(s.stopChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
				utils.DiameterAgent, err))
			shutdown.CloseOnce()
		}
	}(s.da)
	return nil
}

// Reload handles the change of config
func (s *DiameterAgent) Reload(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	s.Lock()
	defer s.Unlock()
	if s.lnet == s.cfg.DiameterAgentCfg().ListenNet &&
		s.laddr == s.cfg.DiameterAgentCfg().Listen {
		return
	}
	close(s.stopChan)

	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CapS,
			utils.ConnManager,
			utils.FilterS,
		},
		registry, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	caps := srvDeps[utils.CapS].(*CapService).Caps()
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	return s.start(fs, cm, caps, shutdown)
}

// Shutdown stops the service
func (s *DiameterAgent) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	s.Lock()
	close(s.stopChan)
	s.da = nil
	s.Unlock()
	return // no shutdown for the momment
}

// ServiceName returns the service name
func (s *DiameterAgent) ServiceName() string {
	return utils.DiameterAgent
}

// ShouldRun returns if the service should be running
func (s *DiameterAgent) ShouldRun() bool {
	return s.cfg.DiameterAgentCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (s *DiameterAgent) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}
