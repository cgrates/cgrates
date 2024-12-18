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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAsteriskAgent returns the Asterisk Agent
func NewAsteriskAgent(cfg *config.CGRConfig) *AsteriskAgent {
	return &AsteriskAgent{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// AsteriskAgent implements Agent interface
type AsteriskAgent struct {
	sync.RWMutex
	cfg       *config.CGRConfig
	stopChan  chan struct{}
	smas      []*agents.AsteriskAgent
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (ast *AsteriskAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	cms, err := WaitForServiceState(utils.StateServiceUP, utils.ConnManager, registry, ast.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}

	ast.Lock()
	defer ast.Unlock()

	listenAndServe := func(sma *agents.AsteriskAgent, stopChan chan struct{}) {
		if err := sma.ListenAndServe(stopChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> runtime error: %s!", utils.AsteriskAgent, err))
			shutdown.CloseOnce()
		}
	}
	ast.stopChan = make(chan struct{})
	ast.smas = make([]*agents.AsteriskAgent, len(ast.cfg.AsteriskAgentCfg().AsteriskConns))
	for connIdx := range ast.cfg.AsteriskAgentCfg().AsteriskConns { // Instantiate connections towards asterisk servers
		ast.smas[connIdx] = agents.NewAsteriskAgent(ast.cfg, connIdx, cms.(*ConnManagerService).ConnManager())
		go listenAndServe(ast.smas[connIdx], ast.stopChan)
	}
	return
}

// Reload handles the change of config
func (ast *AsteriskAgent) Reload(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	ast.shutdown()
	return ast.Start(shutdown, registry)
}

// Shutdown stops the service
func (ast *AsteriskAgent) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	ast.shutdown()
	return
}

func (ast *AsteriskAgent) shutdown() {
	ast.Lock()
	close(ast.stopChan)
	ast.smas = nil
	ast.Unlock()
	return // no shutdown for the momment
}

// ServiceName returns the service name
func (ast *AsteriskAgent) ServiceName() string {
	return utils.AsteriskAgent
}

// ShouldRun returns if the service should be running
func (ast *AsteriskAgent) ShouldRun() bool {
	return ast.cfg.AsteriskAgentCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (ast *AsteriskAgent) StateChan(stateID string) chan struct{} {
	return ast.stateDeps.StateChan(stateID)
}
