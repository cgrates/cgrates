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

// NewFreeswitchAgent returns the Freeswitch Agent
func NewFreeswitchAgent(cfg *config.CGRConfig) *FreeswitchAgent {
	return &FreeswitchAgent{
		cfg: cfg,
	}
}

// FreeswitchAgent implements Agent interface
type FreeswitchAgent struct {
	sync.RWMutex
	cfg *config.CGRConfig
	fS  *agents.FSsessions
}

// Start should handle the sercive start
func (fS *FreeswitchAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.CapS,
		},
		fS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService)
	caps := srvDeps[utils.CapS].(*CapService).Caps()
	cacheS := srvDeps[utils.CacheS].(*CacheService)

	fS.Lock()
	defer fS.Unlock()

	fS.fS, err = agents.NewFSsessions(fs.cfg, cacheS.CacheS(), fs.FilterS(), fS.cfg.GeneralCfg().DefaultTimezone, cms.ConnManager(), caps)
	if err != nil {
		return
	}
	go fS.connect(shutdown)
	return
}

// Reload handles the change of config
func (fS *FreeswitchAgent) Reload(shutdown *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	fS.Lock()
	defer fS.Unlock()
	if err = fS.fS.Shutdown(); err != nil {
		return
	}
	fS.fS.Reload()
	go fS.connect(shutdown)
	return
}

func (fS *FreeswitchAgent) connect(shutdown *utils.SyncedChan) {
	if err := fS.fS.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
		shutdown.CloseOnce() // stop the engine here
	}
}

// Shutdown stops the service
func (fS *FreeswitchAgent) Shutdown(_ *servmanager.Registry) (err error) {
	fS.Lock()
	defer fS.Unlock()
	err = fS.fS.Shutdown()
	fS.fS = nil
	return
}

// ServiceName returns the service name
func (fS *FreeswitchAgent) ServiceName() string {
	return utils.FreeSWITCHAgent
}

// ShouldRun returns if the service should be running
func (fS *FreeswitchAgent) ShouldRun() bool {
	return fS.cfg.FsAgentCfg().Enabled
}
