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
	shdChan *utils.SyncedChan, connMgr *engine.ConnManager, caps *engine.Caps,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &KamailioAgent{
		cfg:     cfg,
		shdChan: shdChan,
		connMgr: connMgr,
		caps:    caps,
		srvDep:  srvDep,
	}
}

// KamailioAgent implements Agent interface
type KamailioAgent struct {
	sync.RWMutex
	cfg     *config.CGRConfig
	shdChan *utils.SyncedChan

	kam     *agents.KamailioAgent
	connMgr *engine.ConnManager
	caps    *engine.Caps
	srvDep  map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (kam *KamailioAgent) Start() error {
	if kam.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	kam.Lock()
	defer kam.Unlock()

	var err error
	kam.kam, err = agents.NewKamailioAgent(kam.cfg.KamAgentCfg(), kam.connMgr,
		utils.FirstNonEmpty(kam.cfg.KamAgentCfg().Timezone, kam.cfg.GeneralCfg().DefaultTimezone), kam.caps)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed to initialize agent, error: %s", utils.KamailioAgent, err))
		return err
	}

	go func(k *agents.KamailioAgent) {
		if connErr := k.Connect(); connErr != nil &&
			!strings.Contains(connErr.Error(), "use of closed network connection") { // if closed by us do not log
			if !strings.Contains(connErr.Error(), "KamEvapi") {
				utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, connErr))
			}
			kam.shdChan.CloseOnce()
		}
	}(kam.kam)
	return nil
}

// Reload handles the change of config
func (kam *KamailioAgent) Reload() (err error) {
	kam.Lock()
	defer kam.Unlock()
	if err = kam.kam.Shutdown(); err != nil {
		return
	}
	kam.kam.Reload()
	go kam.reload(kam.kam)
	return
}

func (kam *KamailioAgent) reload(k *agents.KamailioAgent) (err error) {
	if err = k.Connect(); err != nil {
		if strings.Contains(err.Error(), "use of closed network connection") { // if closed by us do not log
			return
		}
		if !strings.Contains(err.Error(), "KamEvapi") {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, err))
		}
		kam.shdChan.CloseOnce()
	}
	return
}

// Shutdown stops the service
func (kam *KamailioAgent) Shutdown() (err error) {
	kam.Lock()
	defer kam.Unlock()
	err = kam.kam.Shutdown()
	kam.kam = nil
	return
}

// IsRunning returns if the service is running
func (kam *KamailioAgent) IsRunning() bool {
	kam.RLock()
	defer kam.RUnlock()
	return kam.kam != nil
}

// ServiceName returns the service name
func (kam *KamailioAgent) ServiceName() string {
	return utils.KamailioAgent
}

// ShouldRun returns if the service should be running
func (kam *KamailioAgent) ShouldRun() bool {
	return kam.cfg.KamAgentCfg().Enabled
}
