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

// NewRadiusAgent returns the Radius Agent
func NewRadiusAgent(cfg *config.CGRConfig) *RadiusAgent {
	return &RadiusAgent{
		cfg: cfg,
	}
}

// RadiusAgent implements Agent interface
type RadiusAgent struct {
	mu       sync.RWMutex
	cfg      *config.CGRConfig
	stopChan chan struct{}
	rad      *agents.RadiusAgent
}

// Start should handle the sercive start
func (rad *RadiusAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.CapS,
		},
		rad.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cms := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	caps := srvDeps[utils.CapS].(*CapService).Caps()
	cacheS := srvDeps[utils.CacheS].(*CacheService)

	rad.mu.Lock()
	defer rad.mu.Unlock()

	if rad.rad, err = agents.NewRadiusAgent(rad.cfg, cacheS.CacheS(), fs, cms, caps); err != nil {
		return
	}
	rad.stopChan = make(chan struct{})

	go rad.listenAndServe(rad.rad, shutdown)
	return
}

func (rad *RadiusAgent) listenAndServe(r *agents.RadiusAgent, shutdown *utils.SyncedChan) (err error) {
	if err = r.ListenAndServe(rad.stopChan); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.RadiusAgent, err.Error()))
		shutdown.CloseOnce()
	}
	return
}

// Reload handles the change of config
func (rad *RadiusAgent) Reload(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	rad.Shutdown(registry)
	return rad.Start(shutdown, registry)
}

// Shutdown stops the service
func (rad *RadiusAgent) Shutdown(_ *servmanager.Registry) error {
	if rad.rad == nil {
		return nil
	}
	close(rad.stopChan)
	rad.rad.Wait()
	rad.mu.Lock()
	defer rad.mu.Unlock()
	rad.rad = nil
	return nil
}

// ServiceName returns the service name
func (rad *RadiusAgent) ServiceName() string {
	return utils.RadiusAgent
}

// ShouldRun returns if the service should be running
func (rad *RadiusAgent) ShouldRun() bool {
	return rad.cfg.RadiusAgentCfg().Enabled
}
