// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/registrarc"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRegistrarCService returns the Dispatcher Service
func NewRegistrarCService(cfg *config.CGRConfig) *RegistrarCService {
	return &RegistrarCService{
		cfg: cfg,
	}
}

// RegistrarCService implements Service interface
type RegistrarCService struct {
	mu       sync.RWMutex
	cfg      *config.CGRConfig
	dspS     *registrarc.RegistrarCService
	stopChan chan struct{}
	rldChan  chan struct{}
}

// Start should handle the sercive start
func (dspS *RegistrarCService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	dspS.mu.Lock()
	defer dspS.mu.Unlock()

	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.CacheS,
		},
		dspS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)

	dspS.stopChan = make(chan struct{})
	dspS.rldChan = make(chan struct{})
	dspS.dspS = registrarc.NewRegistrarCService(dspS.cfg, cms.ConnManager())
	go dspS.dspS.ListenAndServe(dspS.stopChan, dspS.rldChan)
	return
}

// Reload handles the change of config
func (dspS *RegistrarCService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	dspS.rldChan <- struct{}{}
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (dspS *RegistrarCService) Shutdown(_ *servmanager.Registry) (err error) {
	dspS.mu.Lock()
	defer dspS.mu.Unlock()
	close(dspS.stopChan)
	dspS.dspS.Shutdown()
	dspS.dspS = nil
	return
}

// ServiceName returns the service name
func (dspS *RegistrarCService) ServiceName() string {
	return utils.RegistrarC
}

// ShouldRun returns if the service should be running
func (dspS *RegistrarCService) ShouldRun() bool {
	return len(dspS.cfg.RegistrarCCfg().RPC.RegistrarSConns) != 0
}
