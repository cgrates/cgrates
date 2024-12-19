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
	"github.com/cgrates/cgrates/registrarc"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRegistrarCService returns the Dispatcher Service
func NewRegistrarCService(cfg *config.CGRConfig) *RegistrarCService {
	return &RegistrarCService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// RegistrarCService implements Service interface
type RegistrarCService struct {
	mu  sync.Mutex
	cfg *config.CGRConfig

	dspS *registrarc.RegistrarCService

	stopChan chan struct{}
	rldChan  chan struct{}

	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (dspS *RegistrarCService) Start(_ chan struct{}, registry *servmanager.ServiceRegistry) (err error) {
	cms, err := WaitForServiceState(utils.StateServiceUP, utils.ConnManager, registry, dspS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}

	dspS.stopChan = make(chan struct{})
	dspS.rldChan = make(chan struct{})
	dspS.dspS = registrarc.NewRegistrarCService(dspS.cfg, cms.(*ConnManagerService).ConnManager())
	go dspS.dspS.ListenAndServe(dspS.stopChan, dspS.rldChan)
	close(dspS.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (dspS *RegistrarCService) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) (err error) {
	dspS.rldChan <- struct{}{}
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (dspS *RegistrarCService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	close(dspS.stopChan)
	dspS.dspS.Shutdown()
	dspS.dspS = nil
	close(dspS.StateChan(utils.StateServiceDOWN))
	return
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

// StateChan returns signaling channel of specific state
func (dspS *RegistrarCService) StateChan(stateID string) chan struct{} {
	return dspS.stateDeps.StateChan(stateID)
}

// Lock implements the sync.Locker interface
func (s *RegistrarCService) Lock() {
	s.mu.Lock()
}

// Unlock implements the sync.Locker interface
func (s *RegistrarCService) Unlock() {
	s.mu.Unlock()
}
