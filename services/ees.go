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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewEventExporterService constructs EventExporterService
func NewEventExporterService(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	connMgr *engine.ConnManager, server *cores.Server,
	intConnChan chan birpc.ClientConnector, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &EventExporterService{
		cfg:         cfg,
		filterSChan: filterSChan,
		connMgr:     connMgr,
		server:      server,
		intConnChan: intConnChan,
		rldChan:     make(chan struct{}),
		anz:         anz,
		srvDep:      srvDep,
	}
}

// EventExporterService is the service structure for EventExporterS
type EventExporterService struct {
	sync.RWMutex

	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	connMgr     *engine.ConnManager
	server      *cores.Server
	intConnChan chan birpc.ClientConnector
	rldChan     chan struct{}
	stopChan    chan struct{}

	eeS    *ees.EventExporterS
	rpc    *apis.EeSv1
	anz    *AnalyzerService
	srvDep map[string]*sync.WaitGroup
}

// ServiceName returns the service name
func (es *EventExporterService) ServiceName() string {
	return utils.EEs
}

// ShouldRun returns if the service should be running
func (es *EventExporterService) ShouldRun() (should bool) {
	return es.cfg.EEsCfg().Enabled
}

// IsRunning returns if the service is running
func (es *EventExporterService) IsRunning() bool {
	es.RLock()
	defer es.RUnlock()
	return es != nil && es.eeS != nil
}

// Reload handles the change of config
func (es *EventExporterService) Reload(*context.Context, context.CancelFunc) (err error) {
	es.rldChan <- struct{}{}
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (es *EventExporterService) Shutdown() (err error) {
	es.Lock()
	defer es.Unlock()
	close(es.stopChan)
	es.eeS.Shutdown()
	es.eeS = nil
	<-es.intConnChan
	es.server.RpcUnregisterName(utils.EeSv1)
	return
}

// Start should handle the service start
func (es *EventExporterService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if es.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, es.filterSChan); err != nil {
		return
	}

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.EEs))

	es.Lock()
	defer es.Unlock()

	es.eeS = ees.NewEventExporterS(es.cfg, filterS, es.connMgr)
	es.stopChan = make(chan struct{})
	go es.eeS.ListenAndServe(es.stopChan, es.rldChan)

	es.rpc = apis.NewEeSv1(es.eeS)
	srv, _ := birpc.NewService(es.rpc, "", false)
	if !es.cfg.DispatcherSCfg().Enabled {
		es.server.RpcRegister(srv)
	}
	es.intConnChan <- es.anz.GetInternalCodec(srv, utils.EEs)
	return
}
