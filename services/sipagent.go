// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

// NewSIPAgent returns the sip Agent
func NewSIPAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	shdChan *utils.SyncedChan, connMgr *engine.ConnManager, caps *engine.Caps,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &SIPAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		connMgr:     connMgr,
		caps:        caps,
		srvDep:      srvDep,
	}
}

// SIPAgent implements Agent interface
type SIPAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	shdChan     *utils.SyncedChan

	sip     *agents.SIPAgent
	connMgr *engine.ConnManager
	caps    *engine.Caps
	srvDep  map[string]*sync.WaitGroup

	oldListen string
}

// Start should handle the sercive start
func (sip *SIPAgent) Start() (err error) {
	if sip.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-sip.filterSChan
	sip.filterSChan <- filterS

	sip.Lock()
	defer sip.Unlock()
	sip.oldListen = sip.cfg.SIPAgentCfg().Listen
	sip.sip, err = agents.NewSIPAgent(sip.connMgr, sip.cfg, filterS, sip.caps)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
			utils.SIPAgent, err))
		return
	}
	go func() {
		if err = sip.sip.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.SIPAgent, err.Error()))
			sip.shdChan.CloseOnce() // stop the engine here
		}
	}()
	return
}

// Reload handles the change of config
func (sip *SIPAgent) Reload() (err error) {
	if sip.oldListen == sip.cfg.SIPAgentCfg().Listen {
		return
	}
	sip.Lock()
	sip.sip.Shutdown()
	sip.oldListen = sip.cfg.SIPAgentCfg().Listen
	sip.sip.InitStopChan()
	sip.Unlock()
	go func() {
		if err := sip.sip.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.SIPAgent, err.Error()))
			sip.shdChan.CloseOnce() // stop the engine here
		}
	}()
	return
}

// Shutdown stops the service
func (sip *SIPAgent) Shutdown() (err error) {
	sip.Lock()
	defer sip.Unlock()
	sip.sip.Shutdown()
	sip.sip = nil
	return
}

// IsRunning returns if the service is running
func (sip *SIPAgent) IsRunning() bool {
	sip.RLock()
	defer sip.RUnlock()
	return sip.sip != nil
}

// ServiceName returns the service name
func (sip *SIPAgent) ServiceName() string {
	return utils.SIPAgent
}

// ShouldRun returns if the service should be running
func (sip *SIPAgent) ShouldRun() bool {
	return sip.cfg.SIPAgentCfg().Enabled
}
