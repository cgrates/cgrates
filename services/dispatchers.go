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

	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewDispatcherService returns the Dispatcher Service
func NewDispatcherService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalChan chan rpcclient.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &DispatcherService{
		connChan:    internalChan,
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

// DispatcherService implements Service interface
type DispatcherService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *cores.Server
	connMgr     *engine.ConnManager

	dspS     *dispatchers.DispatcherService
	rpc      *v1.DispatcherSv1
	connChan chan rpcclient.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (dspS *DispatcherService) Start() (err error) {
	if dspS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	utils.Logger.Info("Starting CGRateS DispatcherS service.")
	fltrS := <-dspS.filterSChan
	dspS.filterSChan <- fltrS
	<-dspS.cacheS.GetPrecacheChannel(utils.CacheDispatcherProfiles)
	<-dspS.cacheS.GetPrecacheChannel(utils.CacheDispatcherHosts)
	<-dspS.cacheS.GetPrecacheChannel(utils.CacheDispatcherFilterIndexes)
	dbchan := dspS.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	dspS.Lock()
	defer dspS.Unlock()

	dspS.dspS = dispatchers.NewDispatcherService(datadb, dspS.cfg, fltrS, dspS.connMgr)

	// for the moment we dispable Apier through dispatcher
	// until we figured out a better sollution in case of gob server
	// dspS.server.SetDispatched()

	dspS.server.RpcRegister(v1.NewDispatcherSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.ThresholdSv1,
		v1.NewDispatcherThresholdSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.StatSv1,
		v1.NewDispatcherStatSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.ResourceSv1,
		v1.NewDispatcherResourceSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.RouteSv1,
		v1.NewDispatcherRouteSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.AttributeSv1,
		v1.NewDispatcherAttributeSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.SessionSv1,
		v1.NewDispatcherSessionSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.ChargerSv1,
		v1.NewDispatcherChargerSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.Responder,
		v1.NewDispatcherResponder(dspS.dspS))

	dspS.server.RpcRegisterName(utils.CacheSv1,
		v1.NewDispatcherCacheSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.GuardianSv1,
		v1.NewDispatcherGuardianSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.SchedulerSv1,
		v1.NewDispatcherSchedulerSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.CDRsV1,
		v1.NewDispatcherSCDRsV1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.ConfigSv1,
		v1.NewDispatcherConfigSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.CoreSv1,
		v1.NewDispatcherCoreSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.RALsV1,
		v1.NewDispatcherRALsV1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.ReplicatorSv1,
		v1.NewDispatcherReplicatorSv1(dspS.dspS))

	dspS.server.RpcRegisterName(utils.CDRsV2,
		v2.NewDispatcherSCDRsV2(dspS.dspS))

	dspS.connChan <- dspS.anz.GetInternalCodec(dspS.dspS, utils.DispatcherS)

	return
}

// Reload handles the change of config
func (dspS *DispatcherService) Reload() (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (dspS *DispatcherService) Shutdown() (err error) {
	dspS.Lock()
	defer dspS.Unlock()
	dspS.dspS.Shutdown()
	dspS.dspS = nil
	dspS.rpc = nil
	<-dspS.connChan
	return
}

// IsRunning returns if the service is running
func (dspS *DispatcherService) IsRunning() bool {
	dspS.RLock()
	defer dspS.RUnlock()
	return dspS != nil && dspS.dspS != nil
}

// ServiceName returns the service name
func (dspS *DispatcherService) ServiceName() string {
	return utils.DispatcherS
}

// ShouldRun returns if the service should be running
func (dspS *DispatcherService) ShouldRun() bool {
	return dspS.cfg.DispatcherSCfg().Enabled
}
