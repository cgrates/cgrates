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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewEventReaderService returns the EventReader Service
func NewEventReaderService(
	cfg *config.CGRConfig,
	filterSChan chan *engine.FilterS,
	connMgr *engine.ConnManager,
	clSChan chan *commonlisteners.CommonListenerS,
	intConn chan birpc.ClientConnector,
	anzChan chan *AnalyzerService,
	srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &EventReaderService{
		rldChan:     make(chan struct{}, 1),
		cfg:         cfg,
		filterSChan: filterSChan,
		connMgr:     connMgr,
		clSChan:     clSChan,
		intConn:     intConn,
		anzChan:     anzChan,
		srvDep:      srvDep,
		srvIndexer:  srvIndexer,
	}
}

// EventReaderService implements Service interface
type EventReaderService struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	anzChan     chan *AnalyzerService
	filterSChan chan *engine.FilterS

	ers *ers.ERService
	cl  *commonlisteners.CommonListenerS

	rldChan  chan struct{}
	stopChan chan struct{}
	intConn  chan birpc.ClientConnector
	connMgr  *engine.ConnManager
	cfg      *config.CGRConfig
	srvDep   map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the sercive start
func (erS *EventReaderService) Start(ctx *context.Context, shtDwn context.CancelFunc) (err error) {
	if erS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	erS.cl = <-erS.clSChan
	erS.clSChan <- erS.cl
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, erS.filterSChan); err != nil {
		return
	}
	anz := <-erS.anzChan
	erS.anzChan <- anz

	erS.Lock()
	defer erS.Unlock()

	// remake the stop chan
	erS.stopChan = make(chan struct{})

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ERs))

	// build the service
	erS.ers = ers.NewERService(erS.cfg, filterS, erS.connMgr)
	go erS.listenAndServe(erS.ers, erS.stopChan, erS.rldChan, shtDwn)

	srv, err := engine.NewServiceWithPing(erS.ers, utils.ErSv1, utils.V1Prfx)
	if err != nil {
		return err
	}
	if !erS.cfg.DispatcherSCfg().Enabled {
		erS.cl.RpcRegister(srv)
	}
	erS.intRPCconn = anz.GetInternalCodec(srv, utils.ERs)
	erS.intConn <- erS.intRPCconn
	return
}

func (erS *EventReaderService) listenAndServe(ers *ers.ERService, stopChan chan struct{}, rldChan chan struct{}, shtDwn context.CancelFunc) (err error) {
	if err = ers.ListenAndServe(stopChan, rldChan); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%v>", utils.ERs, err))
		shtDwn()
	}
	return
}

// Reload handles the change of config
func (erS *EventReaderService) Reload(*context.Context, context.CancelFunc) (err error) {
	erS.RLock()
	erS.rldChan <- struct{}{}
	erS.RUnlock()
	return
}

// Shutdown stops the service
func (erS *EventReaderService) Shutdown() (err error) {
	erS.Lock()
	defer erS.Unlock()
	close(erS.stopChan)
	erS.ers = nil
	erS.cl.RpcUnregisterName(utils.ErSv1)
	return
}

// IsRunning returns if the service is running
func (erS *EventReaderService) IsRunning() bool {
	erS.RLock()
	defer erS.RUnlock()
	return erS.ers != nil
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

// IntRPCConn returns the internal connection used by RPCClient
func (erS *EventReaderService) IntRPCConn() birpc.ClientConnector {
	return erS.intRPCconn
}
