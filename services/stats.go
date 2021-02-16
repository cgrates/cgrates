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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewStatService returns the Stat Service
func NewStatService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalStatSChan chan rpcclient.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &StatService{
		connChan:    internalStatSChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
		anz:         anz,
		srvDep:      srvDep,
	}
}

// StatService implements Service interface
type StatService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *cores.Server
	connMgr     *engine.ConnManager

	sts      *engine.StatService
	rpc      *v1.StatSv1
	connChan chan rpcclient.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (sts *StatService) Start() (err error) {
	if sts.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	sts.srvDep[utils.DataDB].Add(1)

	<-sts.cacheS.GetPrecacheChannel(utils.CacheStatQueueProfiles)
	<-sts.cacheS.GetPrecacheChannel(utils.CacheStatQueues)
	<-sts.cacheS.GetPrecacheChannel(utils.CacheStatFilterIndexes)

	filterS := <-sts.filterSChan
	sts.filterSChan <- filterS
	dbchan := sts.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	sts.Lock()
	defer sts.Unlock()
	sts.sts = engine.NewStatService(datadb, sts.cfg, filterS, sts.connMgr)

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem",
		utils.CoreS, utils.StatS))
	sts.sts.StartLoop()
	sts.rpc = v1.NewStatSv1(sts.sts)
	if !sts.cfg.DispatcherSCfg().Enabled {
		sts.server.RpcRegister(sts.rpc)
	}
	sts.connChan <- sts.anz.GetInternalCodec(sts.rpc, utils.StatS)
	return
}

// Reload handles the change of config
func (sts *StatService) Reload() (err error) {
	sts.Lock()
	sts.sts.Reload()
	sts.Unlock()
	return
}

// Shutdown stops the service
func (sts *StatService) Shutdown() (err error) {
	defer sts.srvDep[utils.DataDB].Done()
	sts.Lock()
	defer sts.Unlock()
	if err = sts.sts.Shutdown(); err != nil {
		return
	}
	sts.sts = nil
	sts.rpc = nil
	<-sts.connChan
	return
}

// IsRunning returns if the service is running
func (sts *StatService) IsRunning() bool {
	sts.RLock()
	defer sts.RUnlock()
	return sts != nil && sts.sts != nil
}

// ServiceName returns the service name
func (sts *StatService) ServiceName() string {
	return utils.StatS
}

// ShouldRun returns if the service should be running
func (sts *StatService) ShouldRun() bool {
	return sts.cfg.StatSCfg().Enabled
}
