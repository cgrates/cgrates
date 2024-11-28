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
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/tpes"
	"github.com/cgrates/cgrates/utils"
)

// NewTPeService is the constructor for the TpeService
func NewTPeService(cfg *config.CGRConfig, connMgr *engine.ConnManager, dm *DataDBService,
	clSChan chan *commonlisteners.CommonListenerS, srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &TPeService{
		cfg:        cfg,
		srvDep:     srvDep,
		dm:         dm,
		connMgr:    connMgr,
		clSChan:    clSChan,
		srvIndexer: srvIndexer,
	}
}

// TypeService implements Service interface
type TPeService struct {
	sync.RWMutex

	clSChan chan *commonlisteners.CommonListenerS
	dm      *DataDBService

	tpes *tpes.TPeS
	cl   *commonlisteners.CommonListenerS
	srv  *birpc.Service

	stopChan chan struct{}
	connMgr  *engine.ConnManager
	cfg      *config.CGRConfig
	srvDep   map[string]*sync.WaitGroup

	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the service start
func (ts *TPeService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	ts.cl = <-ts.clSChan
	ts.clSChan <- ts.cl
	var datadb *engine.DataManager
	if datadb, err = ts.dm.WaitForDM(ctx); err != nil {
		return
	}

	ts.tpes = tpes.NewTPeS(ts.cfg, datadb, ts.connMgr)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.TPeS))
	ts.stopChan = make(chan struct{})
	ts.srv, _ = birpc.NewService(apis.NewTPeSv1(ts.tpes), utils.EmptyString, false)
	ts.cl.RpcRegister(ts.srv)
	return
}

// Reload handles the change of config
func (ts *TPeService) Reload(*context.Context, context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (ts *TPeService) Shutdown() (err error) {
	ts.srv = nil
	close(ts.stopChan)
	utils.Logger.Info(fmt.Sprintf("<%s> stopped <%s> subsystem", utils.CoreS, utils.TPeS))
	return
}

// IsRunning returns if the service is running
func (ts *TPeService) IsRunning() bool {
	ts.Lock()
	defer ts.Unlock()
	return ts.tpes != nil
}

// ServiceName returns the service name
func (ts *TPeService) ServiceName() string {
	return utils.TPeS
}

// ShouldRun returns if the service should be running
func (ts *TPeService) ShouldRun() bool {
	return ts.cfg.TpeSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (ts *TPeService) StateChan(stateID string) chan struct{} {
	return ts.stateDeps.StateChan(stateID)
}
