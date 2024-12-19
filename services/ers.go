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

	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewEventReaderService returns the EventReader Service
func NewEventReaderService(cfg *config.CGRConfig) *EventReaderService {
	return &EventReaderService{
		rldChan:   make(chan struct{}, 1),
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// EventReaderService implements Service interface
type EventReaderService struct {
	mu  sync.Mutex
	cfg *config.CGRConfig

	ers *ers.ERService
	cl  *commonlisteners.CommonListenerS

	rldChan  chan struct{}
	stopChan chan struct{}

	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (erS *EventReaderService) Start(shutdown chan struct{}, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
		},
		registry, erS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	erS.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService)

	// remake the stop chan
	erS.stopChan = make(chan struct{})

	// build the service
	erS.ers = ers.NewERService(erS.cfg, fs.FilterS(), cms.ConnManager())
	go erS.listenAndServe(erS.ers, erS.stopChan, erS.rldChan, shutdown)

	srv, err := engine.NewServiceWithPing(erS.ers, utils.ErSv1, utils.V1Prfx)
	if err != nil {
		return err
	}
	if !erS.cfg.DispatcherSCfg().Enabled {
		erS.cl.RpcRegister(srv)
	}
	cms.AddInternalConn(utils.ERs, srv)
	close(erS.stateDeps.StateChan(utils.StateServiceUP))
	return
}

func (erS *EventReaderService) listenAndServe(ers *ers.ERService, stopChan, rldChan, shutdown chan struct{}) (err error) {
	if err = ers.ListenAndServe(stopChan, rldChan); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%v>", utils.ERs, err))
		close(shutdown)
	}
	return
}

// Reload handles the change of config
func (erS *EventReaderService) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) (err error) {
	erS.rldChan <- struct{}{}
	return
}

// Shutdown stops the service
func (erS *EventReaderService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	close(erS.stopChan)
	erS.ers = nil
	erS.cl.RpcUnregisterName(utils.ErSv1)
	close(erS.StateChan(utils.StateServiceDOWN))
	return
}

// ServiceName returns the service name
func (erS *EventReaderService) ServiceName() string {
	return utils.ERs
}

// ShouldRun returns if the service should be running
func (erS *EventReaderService) ShouldRun() bool {
	return erS.cfg.ERsCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (erS *EventReaderService) StateChan(stateID string) chan struct{} {
	return erS.stateDeps.StateChan(stateID)
}

// Lock implements the sync.Locker interface
func (s *EventReaderService) Lock() {
	s.mu.Lock()
}

// Unlock implements the sync.Locker interface
func (s *EventReaderService) Unlock() {
	s.mu.Unlock()
}
