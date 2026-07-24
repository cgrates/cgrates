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

// NewRadiusAgent returns the Radius Agent
func NewRadiusAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	exitChan chan bool, connMgr *engine.ConnManager) servmanager.Service {
	return &RadiusAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		exitChan:    exitChan,
		connMgr:     connMgr,
	}
}

// RadiusAgent implements Agent interface
type RadiusAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	stopChan    chan struct{}
	exitChan    chan bool

	rad     *agents.RadiusAgent
	connMgr *engine.ConnManager
}

// Start should handle the sercive start
func (rad *RadiusAgent) Start() (err error) {
	if rad.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-rad.filterSChan
	rad.filterSChan <- filterS

	rad.Lock()
	defer rad.Unlock()

	if rad.rad, err = agents.NewRadiusAgent(rad.cfg, filterS, rad.connMgr); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.RadiusAgent, err.Error()))
		return
	}
	rad.stopChan = make(chan struct{})
	go func() {
		if err = rad.rad.ListenAndServe(rad.stopChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.RadiusAgent, err.Error()))
		}
		rad.exitChan <- true
	}()
	return
}

// Reload handles the change of config
func (rad *RadiusAgent) Reload() (err error) {
	return
}

// Shutdown stops the service
func (rad *RadiusAgent) Shutdown() (err error) {
	rad.Lock()
	close(rad.stopChan)
	rad.rad = nil
	rad.Unlock()
	return
}

// IsRunning returns if the service is running
func (rad *RadiusAgent) IsRunning() bool {
	rad.RLock()
	defer rad.RUnlock()
	return rad != nil && rad.rad != nil
}

// ServiceName returns the service name
func (rad *RadiusAgent) ServiceName() string {
	return utils.RadiusAgent
}

// ShouldRun returns if the service should be running
func (rad *RadiusAgent) ShouldRun() bool {
	return rad.cfg.RadiusAgentCfg().Enabled
}
