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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRadiusAgent returns the Radius Agent
func NewRadiusAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup,
	shtDwn context.CancelFunc) servmanager.Service {
	return &RadiusAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		shtDwn:      shtDwn,
		connMgr:     connMgr,
		srvDep:      srvDep,
	}
}

// RadiusAgent implements Agent interface
type RadiusAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	shtDwn      context.CancelFunc
	stopChan    chan struct{}

	rad     *agents.RadiusAgent
	connMgr *engine.ConnManager
	srvDep  map[string]*sync.WaitGroup

	lnet  string
	lauth string
	lacct string
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

	rad.lnet = rad.cfg.RadiusAgentCfg().ListenNet
	rad.lauth = rad.cfg.RadiusAgentCfg().ListenAuth
	rad.lacct = rad.cfg.RadiusAgentCfg().ListenAcct

	if rad.rad, err = agents.NewRadiusAgent(rad.cfg, filterS, rad.connMgr); err != nil {
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
		rad.shtDwn()
	}
	return
}

// Reload handles the change of config
func (rad *RadiusAgent) Reload() (err error) {
	if rad.lnet == rad.cfg.RadiusAgentCfg().ListenNet &&
		rad.lauth == rad.cfg.RadiusAgentCfg().ListenAuth &&
		rad.lacct == rad.cfg.RadiusAgentCfg().ListenAcct {
		return
	}

	rad.shutdown()
	return rad.Start()
}

// Shutdown stops the service
func (rad *RadiusAgent) Shutdown() (err error) {
	rad.shutdown()
	return // no shutdown for the momment
}

func (rad *RadiusAgent) shutdown() {
	rad.Lock()
	close(rad.stopChan)
	rad.rad = nil
	rad.Unlock()
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
