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

// NewThresholdService returns the Threshold Service
func NewThresholdService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalThresholdSChan chan rpcclient.ClientConnector,
	anz *AnalyzerService, srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &ThresholdService{
		connChan:    internalThresholdSChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		server:      server,
		anz:         anz,
		srvDep:      srvDep,
	}
}

// ThresholdService implements Service interface
type ThresholdService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *cores.Server

	thrs     *engine.ThresholdService
	rpc      *v1.ThresholdSv1
	connChan chan rpcclient.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (thrs *ThresholdService) Start() (err error) {
	if thrs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	thrs.srvDep[utils.DataDB].Add(1)

	<-thrs.cacheS.GetPrecacheChannel(utils.CacheThresholdProfiles)
	<-thrs.cacheS.GetPrecacheChannel(utils.CacheThresholds)
	<-thrs.cacheS.GetPrecacheChannel(utils.CacheThresholdFilterIndexes)

	filterS := <-thrs.filterSChan
	thrs.filterSChan <- filterS
	dbchan := thrs.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	thrs.Lock()
	defer thrs.Unlock()
	thrs.thrs = engine.NewThresholdService(datadb, thrs.cfg, filterS)

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ThresholdS))
	thrs.thrs.StartLoop()
	thrs.rpc = v1.NewThresholdSv1(thrs.thrs)
	if !thrs.cfg.DispatcherSCfg().Enabled {
		thrs.server.RpcRegister(thrs.rpc)
	}
	thrs.connChan <- thrs.anz.GetInternalCodec(thrs.rpc, utils.ThresholdS)
	return
}

// Reload handles the change of config
func (thrs *ThresholdService) Reload() (err error) {
	thrs.Lock()
	thrs.thrs.Reload()
	thrs.Unlock()
	return
}

// Shutdown stops the service
func (thrs *ThresholdService) Shutdown() (err error) {
	defer thrs.srvDep[utils.DataDB].Done()
	thrs.Lock()
	defer thrs.Unlock()
	if err = thrs.thrs.Shutdown(); err != nil {
		return
	}
	thrs.thrs = nil
	thrs.rpc = nil
	<-thrs.connChan
	return
}

// IsRunning returns if the service is running
func (thrs *ThresholdService) IsRunning() bool {
	thrs.RLock()
	defer thrs.RUnlock()
	return thrs != nil && thrs.thrs != nil
}

// ServiceName returns the service name
func (thrs *ThresholdService) ServiceName() string {
	return utils.ThresholdS
}

// ShouldRun returns if the service should be running
func (thrs *ThresholdService) ShouldRun() bool {
	return thrs.cfg.ThresholdSCfg().Enabled
}
