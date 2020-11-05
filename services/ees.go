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

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewEventExporterService constructs EventExporterService
func NewEventExporterService(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	connMgr *engine.ConnManager, server *cores.Server, exitChan chan bool,
	intConnChan chan rpcclient.ClientConnector,
	anz *AnalyzerService) servmanager.Service {
	return &EventExporterService{
		cfg:         cfg,
		filterSChan: filterSChan,
		connMgr:     connMgr,
		server:      server,
		exitChan:    exitChan,
		intConnChan: intConnChan,
		rldChan:     make(chan struct{}),
		anz:         anz,
	}
}

// EventExporterService is the service structure for EventExporterS
type EventExporterService struct {
	sync.RWMutex

	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	connMgr     *engine.ConnManager
	server      *cores.Server
	exitChan    chan bool
	intConnChan chan rpcclient.ClientConnector
	rldChan     chan struct{}

	eeS *ees.EventExporterS
	rpc *v1.EventExporterSv1
	anz *AnalyzerService
}

// ServiceName returns the service name
func (es *EventExporterService) ServiceName() string {
	return utils.EventExporterS
}

// ShouldRun returns if the service should be running
func (es *EventExporterService) ShouldRun() (should bool) {
	return es.cfg.EEsCfg().Enabled
}

// IsRunning returns if the service is running
func (es *EventExporterService) IsRunning() bool {
	es.RLock()
	defer es.RUnlock()
	return es.eeS != nil
}

// Reload handles the change of config
func (es *EventExporterService) Reload() (err error) {
	es.rldChan <- struct{}{}
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (es *EventExporterService) Shutdown() (err error) {
	es.Lock()

	defer es.Unlock()
	if err = es.eeS.Shutdown(); err != nil {
		return
	}
	es.eeS = nil
	<-es.intConnChan
	return
}

// Start should handle the service start
func (es *EventExporterService) Start() (err error) {
	if es.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	fltrS := <-es.filterSChan
	es.filterSChan <- fltrS

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.EventExporterS))

	es.Lock()
	defer es.Unlock()

	es.eeS, err = ees.NewEventExporterS(es.cfg, fltrS, es.connMgr)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
			utils.EventExporterS, err))
		return
	}
	go func(eeS *ees.EventExporterS, exitChan chan bool, rldChan chan struct{}) {
		if err := eeS.ListenAndServe(exitChan, rldChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.EventExporterS, err.Error()))
			exitChan <- true
		}
	}(es.eeS, es.exitChan, es.rldChan)

	es.rpc = v1.NewEventExporterSv1(es.eeS)
	if !es.cfg.DispatcherSCfg().Enabled {
		es.server.RpcRegister(es.rpc)
	}
	es.intConnChan <- es.anz.GetInternalCodec(es.eeS, utils.EventExporterS)
	return
}
