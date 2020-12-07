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
	"github.com/cgrates/cgrates/dispatcherh"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewDispatcherHostsService returns the Dispatcher Service
func NewDispatcherHostsService(cfg *config.CGRConfig, server *cores.Server,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &DispatcherHostsService{
		cfg:     cfg,
		server:  server,
		connMgr: connMgr,
		anz:     anz,
		srvDep:  srvDep,
	}
}

// DispatcherHostsService implements Service interface
type DispatcherHostsService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	server   *cores.Server
	connMgr  *engine.ConnManager
	stopChan chan struct{}

	dspS   *dispatcherh.DispatcherHostsService
	anz    *AnalyzerService
	srvDep map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (dspS *DispatcherHostsService) Start() (err error) {
	if dspS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	utils.Logger.Info("Starting CGRateS DispatcherH service.")
	dspS.Lock()
	defer dspS.Unlock()

	dspS.stopChan = make(chan struct{})
	dspS.dspS = dispatcherh.NewDispatcherHService(dspS.cfg, dspS.connMgr)
	go dspS.dspS.ListenAndServe(dspS.stopChan)

	return
}

// Reload handles the change of config
func (dspS *DispatcherHostsService) Reload() (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (dspS *DispatcherHostsService) Shutdown() (err error) {
	dspS.Lock()
	close(dspS.stopChan)
	dspS.dspS.Shutdown()
	dspS.dspS = nil
	dspS.Unlock()
	return
}

// IsRunning returns if the service is running
func (dspS *DispatcherHostsService) IsRunning() bool {
	dspS.RLock()
	defer dspS.RUnlock()
	return dspS != nil && dspS.dspS != nil
}

// ServiceName returns the service name
func (dspS *DispatcherHostsService) ServiceName() string {
	return utils.DispatcherH
}

// ShouldRun returns if the service should be running
func (dspS *DispatcherHostsService) ShouldRun() bool {
	return dspS.cfg.DispatcherHCfg().Enabled
}
