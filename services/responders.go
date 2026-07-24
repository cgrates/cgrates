// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewResponderService returns the Resonder Service
func NewResponderService(cfg *config.CGRConfig, server *utils.Server,
	internalRALsChan chan birpc.ClientConnector,
	exitChan chan bool) *ResponderService {
	return &ResponderService{
		connChan: internalRALsChan,
		cfg:      cfg,
		server:   server,
		exitChan: exitChan,
	}
}

// ResponderService implements Service interface
type ResponderService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	server   *utils.Server
	exitChan chan bool

	resp     *engine.Responder
	connChan chan birpc.ClientConnector
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (resp *ResponderService) Start() (err error) {
	if resp.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	resp.Lock()
	defer resp.Unlock()
	resp.resp = &engine.Responder{
		ExitChan:         resp.exitChan,
		MaxComputedUsage: resp.cfg.RalsCfg().MaxComputedUsage,
	}

	if !resp.cfg.DispatcherSCfg().Enabled {
		resp.server.RpcRegister(resp.resp)
	}

	utils.RegisterRpcParams("", resp.resp)

	resp.connChan <- resp.resp // Rater done
	return
}

// Reload handles the change of config
func (resp *ResponderService) Reload() (err error) {
	resp.Lock()
	resp.resp.SetMaxComputedUsage(resp.cfg.RalsCfg().MaxComputedUsage)
	resp.Unlock()
	return
}

// Shutdown stops the service
func (resp *ResponderService) Shutdown() (err error) {
	resp.Lock()
	resp.resp = nil
	<-resp.connChan
	resp.Unlock()
	return
}

// IsRunning returns if the service is running
func (resp *ResponderService) IsRunning() bool {
	resp.RLock()
	defer resp.RUnlock()
	return resp != nil && resp.resp != nil
}

// ServiceName returns the service name
func (resp *ResponderService) ServiceName() string {
	return utils.ResponderS
}

// GetResponder returns the responder created
func (resp *ResponderService) GetResponder() *engine.Responder {
	resp.RLock()
	defer resp.RUnlock()
	return resp.resp
}

// ShouldRun returns if the service should be running
func (resp *ResponderService) ShouldRun() bool {
	return resp.cfg.RalsCfg().Enabled
}
