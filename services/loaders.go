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

	"github.com/cgrates/cgrates/apis"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

// NewLoaderService returns the Loader Service
func NewLoaderService(cfg *config.CGRConfig, dm *DataDBService,
	filterSChan chan *engine.FilterS, server *cores.Server,
	internalLoaderSChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) *LoaderService {
	return &LoaderService{
		connChan:    internalLoaderSChan,
		cfg:         cfg,
		dm:          dm,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
		stopChan:    make(chan struct{}),
		anz:         anz,
		srvDep:      srvDep,
	}
}

// LoaderService implements Service interface
type LoaderService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	filterSChan chan *engine.FilterS
	server      *cores.Server
	stopChan    chan struct{}

	ldrs     *loaders.LoaderService
	rpc      *apis.LoaderSv1
	connChan chan birpc.ClientConnector
	connMgr  *engine.ConnManager
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the service start
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
		ldrs.cfg.GeneralCfg().DefaultTimezone, filterS, ldrs.connMgr)

	if !ldrs.ldrs.Enabled() {
		return
	}
	if err = ldrs.ldrs.ListenAndServe(ldrs.stopChan); err != nil {
		return
	}
	ldrs.rpc = apis.NewLoaderSv1(ldrs.ldrs)
	srv, _ := birpc.NewService(ldrs.rpc, "", false)
	if !ldrs.cfg.DispatcherSCfg().Enabled {
		ldrs.server.RpcRegister(srv)
	}
	ldrs.connChan <- ldrs.anz.GetInternalCodec(srv, utils.LoaderS)
	return
}

// Reload handles the change of config
func (ldrs *LoaderService) Reload() (err error) {
	filterS := <-ldrs.filterSChan
	ldrs.filterSChan <- filterS
	dbchan := ldrs.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb
	close(ldrs.stopChan)
	ldrs.stopChan = make(chan struct{})

	ldrs.RLock()

	ldrs.ldrs.Reload(datadb, ldrs.cfg.LoaderCfg(), ldrs.cfg.GeneralCfg().DefaultTimezone,
		filterS, ldrs.connMgr)
	if err = ldrs.ldrs.ListenAndServe(ldrs.stopChan); err != nil {
		return
	}
	ldrs.RUnlock()
	return
}

// Shutdown stops the service
func (ldrs *LoaderService) Shutdown() (err error) {
	ldrs.Lock()
	ldrs.ldrs = nil
	ldrs.rpc = nil
	close(ldrs.stopChan)
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

// GetLoaderS returns the initialized LoaderService
func (ldrs *LoaderService) GetLoaderS() *loaders.LoaderService {
	return ldrs.ldrs
}
