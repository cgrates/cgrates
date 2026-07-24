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

// NewDiameterAgent returns the Diameter Agent
func NewDiameterAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	exitChan chan bool, connMgr *engine.ConnManager) servmanager.Service {
	return &DiameterAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		exitChan:    exitChan,
		connMgr:     connMgr,
	}
}

// DiameterAgent implements Agent interface
type DiameterAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	exitChan    chan bool

	da      *agents.DiameterAgent
	connMgr *engine.ConnManager
}

// Start should handle the sercive start
func (da *DiameterAgent) Start() (err error) {
	if da.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-da.filterSChan
	da.filterSChan <- filterS

	da.Lock()
	defer da.Unlock()

	da.da, err = agents.NewDiameterAgent(da.cfg, filterS, da.connMgr)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
			utils.DiameterAgent, err))
		return
	}

	go func() {
		if err = da.da.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
				utils.DiameterAgent, err))
		}
		da.exitChan <- true
	}()
	return
}

// Reload handles the change of config
func (da *DiameterAgent) Reload() (err error) {
	return
}

// Shutdown stops the service
func (da *DiameterAgent) Shutdown() (err error) {
	return // no shutdown for the moment
}

// IsRunning returns if the service is running
func (da *DiameterAgent) IsRunning() bool {
	da.RLock()
	defer da.RUnlock()
	return da != nil && da.da != nil
}

// ServiceName returns the service name
func (da *DiameterAgent) ServiceName() string {
	return utils.DiameterAgent
}

// ShouldRun returns if the service should be running
func (da *DiameterAgent) ShouldRun() bool {
	return da.cfg.DiameterAgentCfg().Enabled
}
