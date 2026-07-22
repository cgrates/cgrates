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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewEventReaderService returns the EventReader Service
func NewEventReaderService(
	cfg *config.CGRConfig,
	dm *DataDBService,
	filterSChan chan *engine.FilterS,
	shdChan *utils.SyncedChan,
	connMgr *engine.ConnManager,
	server *cores.Server,
	intConn chan birpc.ClientConnector,
	anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup,
) servmanager.Service {
	return &EventReaderService{
		rldChan:     make(chan struct{}, 1),
		cfg:         cfg,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		dm:          dm,
		connMgr:     connMgr,
		server:      server,
		intConn:     intConn,
		anz:         anz,
		srvDep:      srvDep,
	}
}

// EventReaderService implements Service interface
type EventReaderService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	shdChan     *utils.SyncedChan

	ers      *ers.ERService
	rldChan  chan struct{}
	stopChan chan struct{}
	dm       *DataDBService
	connMgr  *engine.ConnManager
	server   *cores.Server
	intConn  chan birpc.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (erS *EventReaderService) Start() (err error) {
	if erS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	erS.Lock()
	defer erS.Unlock()

	filterS := <-erS.filterSChan
	erS.filterSChan <- filterS
	dbchan := erS.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	// remake the stop chan
	erS.stopChan = make(chan struct{})

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ERs))

	// build the service
	erS.ers = ers.NewERService(erS.cfg, datadb, filterS, erS.connMgr)
	go erS.listenAndServe(erS.ers, erS.stopChan, erS.rldChan)

	// Register ERsV1 methods.
	srv, err := engine.NewService(v1.NewErSv1(erS.ers))
	if err != nil {
		return err
	}
	if !erS.cfg.DispatcherSCfg().Enabled {
		erS.server.RpcRegister(srv)
	}

	erS.intConn <- erS.anz.GetInternalCodec(srv, utils.ERs)
	return
}

func (erS *EventReaderService) listenAndServe(ers *ers.ERService, stopChan chan struct{}, rldChan chan struct{}) (err error) {
	if err = ers.ListenAndServe(stopChan, rldChan); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%v>", utils.ERs, err))
		erS.shdChan.CloseOnce()
	}
	return
}

// Reload handles the change of config
func (erS *EventReaderService) Reload() (err error) {
	erS.RLock()
	erS.rldChan <- struct{}{}
	erS.RUnlock()
	return
}

// Shutdown stops the service
func (erS *EventReaderService) Shutdown() (err error) {
	erS.Lock()
	close(erS.stopChan)
	erS.ers = nil
	erS.Unlock()
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
