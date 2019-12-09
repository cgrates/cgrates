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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewResponderService returns the Resonder Service
func NewResponderService(cfg *config.CGRConfig, server *utils.Server,
	internalRALsChan chan rpcclient.RpcClientConnection,
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
	connChan chan rpcclient.ClientConnector
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (resp *ResponderService) Start() (err error) {
	if resp.IsRunning() {
		return fmt.Errorf("service aleady running")
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

// GetIntenternalChan returns the internal connection chanel
func (resp *ResponderService) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	return resp.connChan
}

// Reload handles the change of config
func (resp *ResponderService) Reload() (err error) {
	resp.Lock()
	resp.resp.MaxComputedUsage = resp.cfg.RalsCfg().MaxComputedUsage // this may cause concurrency problems
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
