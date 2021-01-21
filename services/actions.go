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

	"github.com/cgrates/cgrates/actions"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewActionService returns the Action Service
func NewActionService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	connMgr *engine.ConnManager,
	server *cores.Server, internalChan chan rpcclient.ClientConnector,
	anz *AnalyzerService, srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &ActionService{
		connChan:    internalChan,
		connMgr:     connMgr,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		server:      server,
		anz:         anz,
		srvDep:      srvDep,
		rldChan:     make(chan struct{}, 1),
	}
}

// ActionService implements Service interface
type ActionService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	connMgr     *engine.ConnManager
	server      *cores.Server

	rldChan  chan struct{}
	stopChan chan struct{}

	acts     *actions.ActionS
	rpc      *v1.ActionSv1                  // useful on restart
	connChan chan rpcclient.ClientConnector // publish the internal Subsystem when available
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the service start
func (acts *ActionService) Start() (err error) {
	if acts.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	<-acts.cacheS.GetPrecacheChannel(utils.CacheActionProfiles)
	<-acts.cacheS.GetPrecacheChannel(utils.CacheActionProfilesFilterIndexes)

	filterS := <-acts.filterSChan
	acts.filterSChan <- filterS
	dbchan := acts.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	acts.Lock()
	defer acts.Unlock()
	acts.acts = actions.NewActionS(acts.cfg, filterS, datadb, acts.connMgr)
	acts.stopChan = make(chan struct{})
	go acts.acts.ListenAndServe(acts.stopChan, acts.rldChan)

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ActionS))
	acts.rpc = v1.NewActionSv1(acts.acts)
	if !acts.cfg.DispatcherSCfg().Enabled {
		acts.server.RpcRegister(acts.rpc)
	}
	acts.connChan <- acts.anz.GetInternalCodec(acts.rpc, utils.ActionS)
	return
}

// Reload handles the change of config
func (acts *ActionService) Reload() (err error) {
	acts.rldChan <- struct{}{}
	return // for the moment nothing to reload
}

// Shutdown stops the service
func (acts *ActionService) Shutdown() (err error) {
	acts.Lock()
	defer acts.Unlock()
	close(acts.stopChan)
	if err = acts.acts.Shutdown(); err != nil {
		return
	}
	acts.acts = nil
	acts.rpc = nil
	<-acts.connChan
	return
}

// IsRunning returns if the service is running
func (acts *ActionService) IsRunning() bool {
	acts.RLock()
	defer acts.RUnlock()
	return acts != nil && acts.acts != nil
}

// ServiceName returns the service name
func (acts *ActionService) ServiceName() string {
	return utils.ActionS
}

// ShouldRun returns if the service should be running
func (acts *ActionService) ShouldRun() bool {
	return acts.cfg.ActionSCfg().Enabled
}
