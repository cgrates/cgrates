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
	"github.com/cgrates/cgrates/registrarc"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRegistrarCService returns the Dispatcher Service
func NewRegistrarCService(cfg *config.CGRConfig, server *cores.Server,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &RegistrarCService{
		cfg:     cfg,
		server:  server,
		connMgr: connMgr,
		anz:     anz,
		srvDep:  srvDep,
	}
}

// RegistrarCService implements Service interface
type RegistrarCService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	server   *cores.Server
	connMgr  *engine.ConnManager
	stopChan chan struct{}
	rldChan  chan struct{}

	dspS   *registrarc.RegistrarCService
	anz    *AnalyzerService
	srvDep map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (dspS *RegistrarCService) Start() (err error) {
	if dspS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	utils.Logger.Info("Starting CGRateS DispatcherH service.")
	dspS.Lock()
	defer dspS.Unlock()

	dspS.stopChan = make(chan struct{})
	dspS.rldChan = make(chan struct{})
	dspS.dspS = registrarc.NewRegistrarCService(dspS.cfg, dspS.connMgr)
	go dspS.dspS.ListenAndServe(dspS.stopChan, dspS.rldChan)

	return
}

// Reload handles the change of config
func (dspS *RegistrarCService) Reload() (err error) {
	dspS.rldChan <- struct{}{}
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (dspS *RegistrarCService) Shutdown() (err error) {
	dspS.Lock()
	close(dspS.stopChan)
	dspS.dspS.Shutdown()
	dspS.dspS = nil
	dspS.Unlock()
	return
}

// IsRunning returns if the service is running
func (dspS *RegistrarCService) IsRunning() bool {
	dspS.RLock()
	defer dspS.RUnlock()
	return dspS != nil && dspS.dspS != nil
}

// ServiceName returns the service name
func (dspS *RegistrarCService) ServiceName() string {
	return utils.RegistrarC
}

// ShouldRun returns if the service should be running
func (dspS *RegistrarCService) ShouldRun() bool {
	return len(dspS.cfg.RegistrarCCfg().RPC.RegistrarSConns) != 0 ||
		len(dspS.cfg.RegistrarCCfg().Dispatchers.RegistrarSConns) != 0
}
