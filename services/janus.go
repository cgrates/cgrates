// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewJanusAgent returns the Janus Agent
func NewJanusAgent(cfg *config.CGRConfig) *JanusAgent {
	return &JanusAgent{
		cfg: cfg,
	}
}

// JanusAgent implements Service interface
type JanusAgent struct {
	mu  sync.RWMutex
	cfg *config.CGRConfig
	jA  *agents.JanusAgent

	// we can realy stop the JanusAgent so keep a flag
	// if we registerd the jandlers
	started bool
}

// Start should jandle the sercive start
func (ja *JanusAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
			utils.CapS,
		},
		ja.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService)
	caps := srvDeps[utils.CapS].(*CapService).Caps()

	ja.mu.Lock()
	defer ja.mu.Unlock()
	if ja.started {
		return utils.ErrServiceAlreadyRunning
	}
	ja.jA, err = agents.NewJanusAgent(ja.cfg, cms.ConnManager(), fs.FilterS(), caps)
	if err != nil {
		return
	}
	if err = ja.jA.Connect(); err != nil {
		return
	}

	cl.RegisterHttpHandler("POST "+ja.cfg.JanusAgentCfg().URL, http.HandlerFunc(ja.jA.CreateSession))
	cl.RegisterHttpHandler("OPTIONS "+ja.cfg.JanusAgentCfg().URL, http.HandlerFunc(ja.jA.CORSOptions))
	cl.RegisterHttpHandler(fmt.Sprintf("OPTIONS %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.SessionKeepalive))
	cl.RegisterHttpHandler(fmt.Sprintf("OPTIONS %s/{sessionID}/", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.CORSOptions))
	cl.RegisterHttpHandler(fmt.Sprintf("GET %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.PollSession))
	cl.RegisterHttpHandler(fmt.Sprintf("POST %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.AttachPlugin))
	cl.RegisterHttpHandler(fmt.Sprintf("POST %s/{sessionID}/{handleID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.HandlePlugin))

	ja.started = true
	return
}

// Reload jandles the change of config
func (ja *JanusAgent) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return // no reload
}

// Shutdown stops the service
func (ja *JanusAgent) Shutdown(_ *servmanager.Registry) (err error) {
	ja.mu.Lock()
	err = ja.jA.Shutdown()
	ja.started = false
	ja.mu.Unlock()
	return // no shutdown for the momment
}

// ServiceName returns the service name
func (ja *JanusAgent) ServiceName() string {
	return utils.JanusAgent
}

// ShouldRun returns if the service should be running
func (ja *JanusAgent) ShouldRun() bool {
	return ja.cfg.JanusAgentCfg().Enabled
}
