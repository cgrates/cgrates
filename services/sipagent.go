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

// NewSIPAgent returns the sip Agent
func NewSIPAgent(cfg *config.CGRConfig) *SIPAgent {
	return &SIPAgent{
		cfg: cfg,
	}
}

// SIPAgent implements Agent interface
type SIPAgent struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	sip       *agents.SIPAgent
	oldListen string
}

// Start should handle the sercive start
func (sip *SIPAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.CapS,
		},
		sip.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	caps := srvDeps[utils.CapS].(*CapService).Caps()
	cacheS := srvDeps[utils.CacheS].(*CacheService)

	sip.mu.Lock()
	defer sip.mu.Unlock()
	sip.oldListen = sip.cfg.SIPAgentCfg().Listen
	sip.sip, err = agents.NewSIPAgent(cm, sip.cfg, cacheS.CacheS(), fs, caps)
	if err != nil {
		return
	}
	go sip.listenAndServe(shutdown)
	return
}

func (sip *SIPAgent) listenAndServe(shutdown *utils.SyncedChan) {
	if err := sip.sip.ListenAndServe(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.SIPAgent, err.Error()))
		shutdown.CloseOnce() // stop the engine here
	}
}

// Reload handles the change of config
func (sip *SIPAgent) Reload(shutdown *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	if sip.oldListen == sip.cfg.SIPAgentCfg().Listen {
		return
	}
	sip.mu.Lock()
	sip.sip.Shutdown()
	sip.oldListen = sip.cfg.SIPAgentCfg().Listen
	sip.sip.InitStopChan()
	sip.mu.Unlock()
	go sip.listenAndServe(shutdown)
	return
}

// Shutdown stops the service
func (sip *SIPAgent) Shutdown(_ *servmanager.Registry) (err error) {
	sip.mu.Lock()
	defer sip.mu.Unlock()
	sip.sip.Shutdown()
	sip.sip = nil
	return
}

// ServiceName returns the service name
func (sip *SIPAgent) ServiceName() string {
	return utils.SIPAgent
}

// ShouldRun returns if the service should be running
func (sip *SIPAgent) ShouldRun() bool {
	return sip.cfg.SIPAgentCfg().Enabled
}
