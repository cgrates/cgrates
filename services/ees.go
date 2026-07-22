// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewEventExporterService constructs EventExporterService
func NewEventExporterService(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	connMgr *engine.ConnManager, server *cores.Server, intConnChan chan birpc.ClientConnector,
	anz *AnalyzerService, srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &EventExporterService{
		cfg:         cfg,
		filterSChan: filterSChan,
		connMgr:     connMgr,
		server:      server,
		intConnChan: intConnChan,
		anz:         anz,
		srvDep:      srvDep,
	}
}

// EventExporterService is the service structure for EventExporterS
type EventExporterService struct {
	mu sync.RWMutex

	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	connMgr     *engine.ConnManager
	server      *cores.Server
	intConnChan chan birpc.ClientConnector

	eeS    *ees.EventExporterS
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
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.eeS != nil
}

// Reload handles the change of config
func (es *EventExporterService) Reload() error {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.eeS.ClearExporterCache()
	return es.eeS.SetupExporterCache()
}

// Shutdown stops the service
func (es *EventExporterService) Shutdown() error {
	es.mu.Lock()
	defer es.mu.Unlock()
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.EEs))
	es.eeS.ClearExporterCache()
	es.eeS = nil
	<-es.intConnChan
	return nil
}

// Start should handle the service start
func (es *EventExporterService) Start() error {
	if es.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	fltrS := <-es.filterSChan
	es.filterSChan <- fltrS

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.EEs))

	es.mu.Lock()
	defer es.mu.Unlock()

	var err error
	es.eeS, err = ees.NewEventExporterS(es.cfg, fltrS, es.connMgr)
	if err != nil {
		return err
	}

	srv, err := engine.NewService(v1.NewEeSv1(es.eeS))
	if err != nil {
		return err
	}
	if !es.cfg.DispatcherSCfg().Enabled {
		es.server.RpcRegister(srv)
	}
	es.intConnChan <- es.anz.GetInternalCodec(srv, utils.EEs)
	return nil
}
