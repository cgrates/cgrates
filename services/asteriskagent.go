// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
		cfg: cfg,
	}
}

// AsteriskAgent implements Agent interface
type AsteriskAgent struct {
	mu       sync.RWMutex
	cfg      *config.CGRConfig
	stopChan chan struct{}
	smas     []*agents.AsteriskAgent
}

// Start should handle the sercive start
func (ast *AsteriskAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.CapS,
			utils.FilterS,
		},
		ast.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	caps := srvDeps[utils.CapS].(*CapService).Caps()
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()

	ast.mu.Lock()
	defer ast.mu.Unlock()

	listenAndServe := func(sma *agents.AsteriskAgent, stopChan chan struct{}) {
		if err := sma.ListenAndServe(stopChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> runtime error: %s!", utils.AsteriskAgent, err))
			shutdown.CloseOnce()
		}
	}
	ast.stopChan = make(chan struct{})
	ast.smas = make([]*agents.AsteriskAgent, len(ast.cfg.AsteriskAgentCfg().AsteriskConns))
	for connIdx := range ast.cfg.AsteriskAgentCfg().AsteriskConns { // Instantiate connections towards asterisk servers
		if ast.smas[connIdx], err = agents.NewAsteriskAgent(ast.cfg, connIdx, cm, caps, fs); err != nil {
			return
		}
		go listenAndServe(ast.smas[connIdx], ast.stopChan)
	}
	return
}

// Reload handles the change of config
func (ast *AsteriskAgent) Reload(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	ast.shutdown()
	return ast.Start(shutdown, registry)
}

// Shutdown stops the service
func (ast *AsteriskAgent) Shutdown(_ *servmanager.Registry) (err error) {
	ast.shutdown()
	return
}

func (ast *AsteriskAgent) shutdown() {
	ast.mu.Lock()
	defer ast.mu.Unlock()
	close(ast.stopChan)
	ast.smas = nil
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
