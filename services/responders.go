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
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewResponderService returns the Resonder Service
func NewResponderService(cfg *config.CGRConfig, server *cores.Server,
	internalRALsChan chan rpcclient.ClientConnector,
	shdChan *utils.SyncedChan, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) *ResponderService {
	return &ResponderService{
		connChan:  internalRALsChan,
		cfg:       cfg,
		server:    server,
		shdChan:   shdChan,
		anz:       anz,
		srvDep:    srvDep,
		syncChans: make(map[string]chan *engine.Responder),
	}
}

// ResponderService implements Service interface
// this service is manged by the RALs as a component
type ResponderService struct {
	sync.RWMutex
	cfg     *config.CGRConfig
	server  *cores.Server
	shdChan *utils.SyncedChan

	resp      *engine.Responder
	connChan  chan rpcclient.ClientConnector
	anz       *AnalyzerService
	srvDep    map[string]*sync.WaitGroup
	syncChans map[string]chan *engine.Responder
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
		ShdChan:          resp.shdChan,
		MaxComputedUsage: resp.cfg.RalsCfg().MaxComputedUsage,
	}

	if !resp.cfg.DispatcherSCfg().Enabled {
		resp.server.RpcRegister(resp.resp)
	}

	resp.connChan <- resp.anz.GetInternalCodec(resp.resp, utils.ResponderS) // Rater done
	resp.sync()
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
	for _, c := range resp.syncChans {
		c <- nil // just tell the services that responder is nil
	}
	resp.Unlock()
	return
}

// IsRunning returns if the service is running
func (resp *ResponderService) IsRunning() bool {
	resp.RLock()
	defer resp.RUnlock()
	return resp.isRunning()
}

func (resp *ResponderService) isRunning() bool {
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

// RegisterSyncChan used by dependent subsystems to register a chanel to reload only the responder(thread safe)
func (resp *ResponderService) RegisterSyncChan(srv string, c chan *engine.Responder) {
	resp.Lock()
	resp.syncChans[srv] = c
	if resp.isRunning() {
		c <- resp.resp
	}
	resp.Unlock()
}

// UnregisterSyncChan used by dependent subsystems to unregister a chanel
func (resp *ResponderService) UnregisterSyncChan(srv string) {
	resp.Lock()
	c, has := resp.syncChans[srv]
	if has {
		close(c)
		delete(resp.syncChans, srv)
	}
	resp.Unlock()
}

// sync sends the responder over syncChansv (not thrad safe)
func (resp *ResponderService) sync() {
	if resp.isRunning() {
		for _, c := range resp.syncChans {
			c <- resp.resp
		}
	}
}
