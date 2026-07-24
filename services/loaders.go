// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewLoaderService returns the Loader Service
func NewLoaderService(cfg *config.CGRConfig, dm *DataDBService,
	filterSChan chan *engine.FilterS, server *utils.Server,
	exitChan chan bool, internalLoaderSChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager) servmanager.Service {
	return &LoaderService{
		connChan:    internalLoaderSChan,
		cfg:         cfg,
		dm:          dm,
		filterSChan: filterSChan,
		server:      server,
		exitChan:    exitChan,
		connMgr:     connMgr,
	}
}

// LoaderService implements Service interface
type LoaderService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	filterSChan chan *engine.FilterS
	server      *utils.Server
	exitChan    chan bool

	ldrs     *loaders.LoaderService
	rpc      *v1.LoaderSv1
	connChan chan birpc.ClientConnector
	connMgr  *engine.ConnManager
}

// Start should handle the sercive start
func (ldrs *LoaderService) Start() (err error) {
	if ldrs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-ldrs.filterSChan
	ldrs.filterSChan <- filterS
	dbchan := ldrs.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	ldrs.Lock()
	defer ldrs.Unlock()

	ldrs.ldrs = loaders.NewLoaderService(datadb, ldrs.cfg.LoaderCfg(),
		ldrs.cfg.GeneralCfg().DefaultTimezone, ldrs.exitChan, filterS, ldrs.connMgr)
	if !ldrs.ldrs.Enabled() {
		return
	}
	ldrs.rpc = v1.NewLoaderSv1(ldrs.ldrs)
	if !ldrs.cfg.DispatcherSCfg().Enabled {
		ldrs.server.RpcRegister(ldrs.rpc)
	}
	ldrs.connChan <- ldrs.rpc
	return
}

// Reload handles the change of config
func (ldrs *LoaderService) Reload() (err error) {
	filterS := <-ldrs.filterSChan
	ldrs.filterSChan <- filterS
	dbchan := ldrs.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	ldrs.RLock()

	ldrs.ldrs.Reload(datadb, ldrs.cfg.LoaderCfg(), ldrs.cfg.GeneralCfg().DefaultTimezone,
		ldrs.exitChan, filterS, ldrs.connMgr)
	ldrs.RUnlock()
	return
}

// Shutdown stops the service
func (ldrs *LoaderService) Shutdown() (err error) {
	ldrs.Lock()
	ldrs.ldrs = nil
	ldrs.rpc = nil
	<-ldrs.connChan
	ldrs.Unlock()
	return
}

// IsRunning returns if the service is running
func (ldrs *LoaderService) IsRunning() bool {
	ldrs.RLock()
	defer ldrs.RUnlock()
	return ldrs != nil && ldrs.ldrs != nil && ldrs.ldrs.Enabled()
}

// ServiceName returns the service name
func (ldrs *LoaderService) ServiceName() string {
	return utils.LoaderS
}

// ShouldRun returns if the service should be running
func (ldrs *LoaderService) ShouldRun() bool {
	return ldrs.cfg.LoaderCfg().Enabled()
}
