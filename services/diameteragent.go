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
	shdChan *utils.SyncedChan, connMgr *engine.ConnManager, caps *engine.Caps,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &DiameterAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		connMgr:     connMgr,
		caps:        caps,
		srvDep:      srvDep,
	}
}

// DiameterAgent implements Agent interface
type DiameterAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	shdChan     *utils.SyncedChan
	stopChan    chan struct{}

	da      *agents.DiameterAgent
	connMgr *engine.ConnManager
	caps    *engine.Caps

	srvDep map[string]*sync.WaitGroup
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
	return da.start(filterS, da.caps)
}

func (da *DiameterAgent) start(filterS *engine.FilterS, caps *engine.Caps) error {
	var err error
	da.da, err = agents.NewDiameterAgent(da.cfg, filterS, da.connMgr, caps)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed to initialize agent, error: %s",
			utils.DiameterAgent, err))
		return err
	}
	da.stopChan = make(chan struct{})
	go func(d *agents.DiameterAgent) {
		lnsErr := d.ListenAndServe(da.stopChan)
		if lnsErr != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s",
				utils.DiameterAgent, lnsErr))
			da.shdChan.CloseOnce()
		}
	}(da.da)
	return nil
}

// Reload handles the change of config
func (da *DiameterAgent) Reload() (err error) {
	da.Shutdown()
	return da.Start()
}

// Shutdown stops the service
func (da *DiameterAgent) Shutdown() (err error) {
	da.Lock()
	close(da.stopChan)
	da.da = nil
	da.Unlock()
	return // no shutdown for the momment
}

// IsRunning returns if the service is running
func (da *DiameterAgent) IsRunning() bool {
	da.RLock()
	defer da.RUnlock()
	return da.da != nil
}

// ServiceName returns the service name
func (da *DiameterAgent) ServiceName() string {
	return utils.DiameterAgent
}

// ShouldRun returns if the service should be running
func (da *DiameterAgent) ShouldRun() bool {
	return da.cfg.DiameterAgentCfg().Enabled
}
