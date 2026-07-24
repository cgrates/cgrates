// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewKamailioAgent returns the Kamailio Agent
func NewKamailioAgent(cfg *config.CGRConfig,
	exitChan chan bool, connMgr *engine.ConnManager) servmanager.Service {
	return &KamailioAgent{
		cfg:      cfg,
		exitChan: exitChan,
		connMgr:  connMgr,
	}
}

// KamailioAgent implements Agent interface
type KamailioAgent struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	exitChan chan bool

	kam     *agents.KamailioAgent
	connMgr *engine.ConnManager
}

// Start should handle the sercive start
func (kam *KamailioAgent) Start() (err error) {
	if kam.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	kam.Lock()
	defer kam.Unlock()

	kam.kam = agents.NewKamailioAgent(kam.cfg.KamAgentCfg(), kam.connMgr,
		utils.FirstNonEmpty(kam.cfg.KamAgentCfg().Timezone, kam.cfg.GeneralCfg().DefaultTimezone))

	go func() {
		if err = kam.kam.Connect(); err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") { // if closed by us do not log
				return
			}
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, err))
			kam.exitChan <- true
		}
	}()
	return
}

// Reload handles the change of config
func (kam *KamailioAgent) Reload() (err error) {

	if err = kam.Shutdown(); err != nil {
		return
	}
	kam.Lock()
	defer kam.Unlock()
	kam.kam.Reload()
	go func() {
		if err = kam.kam.Connect(); err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") { // if closed by us do not log
				return
			}
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, err))
			kam.exitChan <- true
		}
	}()
	return
}

// Shutdown stops the service
func (kam *KamailioAgent) Shutdown() (err error) {
	kam.Lock()
	defer kam.Unlock()
	if err = kam.kam.Shutdown(); err != nil {
		return
	}
	kam.kam = nil
	return
}

// IsRunning returns if the service is running
func (kam *KamailioAgent) IsRunning() bool {
	kam.RLock()
	defer kam.RUnlock()
	return kam != nil && kam.kam != nil
}

// ServiceName returns the service name
func (kam *KamailioAgent) ServiceName() string {
	return utils.KamailioAgent
}

// ShouldRun returns if the service should be running
func (kam *KamailioAgent) ShouldRun() bool {
	return kam.cfg.KamAgentCfg().Enabled
}
