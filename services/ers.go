/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package services

import (
	"fmt"
	"sync"

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
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	ers       *ers.ERService
	rldChan   chan struct{}
	stopChan  chan struct{}
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (erS *EventReaderService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
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
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService)

	erS.mu.Lock()
	defer erS.mu.Unlock()

	// remake the stop chan
	erS.stopChan = make(chan struct{})

	// build the service
	erS.ers = ers.NewERService(erS.cfg, fs.FilterS(), cms.ConnManager())
	go erS.listenAndServe(erS.ers, erS.stopChan, erS.rldChan, shutdown)

	srv, err := engine.NewServiceWithPing(erS.ers, utils.ErSv1, utils.V1Prfx)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.ERs, srv)
	return
}

func (erS *EventReaderService) listenAndServe(ers *ers.ERService, stopChan, rldChan chan struct{}, shutdown *utils.SyncedChan) (err error) {
	if err = ers.ListenAndServe(stopChan, rldChan); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%v>", utils.ERs, err))
		shutdown.CloseOnce()
	}
	return
}

// Reload handles the change of config
func (erS *EventReaderService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	erS.mu.RLock()
	erS.rldChan <- struct{}{}
	erS.mu.RUnlock()
	return
}

// Shutdown stops the service
func (erS *EventReaderService) Shutdown(registry *servmanager.ServiceRegistry) (err error) {
	erS.mu.Lock()
	defer erS.mu.Unlock()
	close(erS.stopChan)
	erS.ers = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.ErSv1)
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
