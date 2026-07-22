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
	shdChan *utils.SyncedChan, connMgr *engine.ConnManager, caps *engine.Caps,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &RadiusAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		connMgr:     connMgr,
		caps:        caps,
		srvDep:      srvDep,
	}
}

// RadiusAgent implements Agent interface
type RadiusAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	shdChan     *utils.SyncedChan
	stopChan    chan struct{}

	rad     *agents.RadiusAgent
	connMgr *engine.ConnManager
	caps    *engine.Caps
	srvDep  map[string]*sync.WaitGroup
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

	if rad.rad, err = agents.NewRadiusAgent(rad.cfg, filterS, rad.connMgr, rad.caps); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.RadiusAgent, err.Error()))
		return
	}
	rad.stopChan = make(chan struct{})

	go rad.listenAndServe(rad.rad)

	return
}

func (rad *RadiusAgent) listenAndServe(r *agents.RadiusAgent) (err error) {
	if err = r.ListenAndServe(rad.stopChan); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.RadiusAgent, err.Error()))
		rad.shdChan.CloseOnce()
	}
	return
}

// Reload handles the change of config
func (rad *RadiusAgent) Reload() (err error) {
	rad.Shutdown()
	return rad.Start()
}

// Shutdown stops the service
func (rad *RadiusAgent) Shutdown() (err error) {
	if rad.rad == nil {
		return
	}
	close(rad.stopChan)
	rad.rad.Wait()
	rad.rad.Lock()
	defer rad.rad.Unlock()
	rad.rad = nil
	return // no shutdown for the momment
}

// IsRunning returns if the service is running
func (rad *RadiusAgent) IsRunning() bool {
	rad.RLock()
	defer rad.RUnlock()
	return rad.rad != nil
}

// ServiceName returns the service name
func (rad *RadiusAgent) ServiceName() string {
	return utils.RadiusAgent
}

// ShouldRun returns if the service should be running
func (rad *RadiusAgent) ShouldRun() bool {
	return rad.cfg.RadiusAgentCfg().Enabled
}
