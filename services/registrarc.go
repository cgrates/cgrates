// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	return dspS.dspS != nil
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
