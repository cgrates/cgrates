// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewKamailioAgent returns the Kamailio Agent
func NewKamailioAgent(cfg *config.CGRConfig) *KamailioAgent {
	return &KamailioAgent{
		cfg: cfg,
	}
}

// KamailioAgent implements Agent interface
type KamailioAgent struct {
	sync.RWMutex
	cfg *config.CGRConfig
	kam *agents.KamailioAgent
}

// Start should handle the sercive start
func (kam *KamailioAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{utils.ConnManager, utils.CapS, utils.FilterS},
		kam.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	caps := srvDeps[utils.CapS].(*CapService).Caps()
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()

	kam.Lock()
	defer kam.Unlock()

	kam.kam = agents.NewKamailioAgent(kam.cfg, cm,
		utils.FirstNonEmpty(kam.cfg.KamAgentCfg().Timezone, kam.cfg.GeneralCfg().DefaultTimezone), caps, fs)

	go kam.connect(kam.kam, shutdown)
	return
}

// Reload handles the change of config
func (kam *KamailioAgent) Reload(shutdown *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	kam.Lock()
	defer kam.Unlock()
	if err = kam.kam.Shutdown(); err != nil {
		return
	}
	kam.kam.Reload()
	go kam.connect(kam.kam, shutdown)
	return
}

func (kam *KamailioAgent) connect(k *agents.KamailioAgent, shutdown *utils.SyncedChan) (err error) {
	if err = k.Connect(); err != nil {
		if !strings.Contains(err.Error(), "use of closed network connection") { // if closed by us do not log
			if !strings.Contains(err.Error(), "KamEvapi") {
				utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, err))
			}
			shutdown.CloseOnce()
		}
	}
	return
}

// Shutdown stops the service
func (kam *KamailioAgent) Shutdown(_ *servmanager.Registry) (err error) {
	kam.Lock()
	defer kam.Unlock()
	err = kam.kam.Shutdown()
	kam.kam = nil
	return
}

// ServiceName returns the service name
func (kam *KamailioAgent) ServiceName() string {
	return utils.KamailioAgent
}

// ShouldRun returns if the service should be running
func (kam *KamailioAgent) ShouldRun() bool {
	return kam.cfg.KamAgentCfg().Enabled
}
