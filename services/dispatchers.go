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
	"strings"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewDispatcherService returns the Dispatcher Service
func NewDispatcherService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) *DispatcherService {
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
		srvsReload:  make(map[string]chan struct{}),
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
	connChan chan birpc.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup

	srvsReload map[string]chan struct{}
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

	dspS.unregisterAllDispatchedSubsystems() // unregister all rpc services that can be dispatched

	srv, _ := birpc.NewService(apis.NewDispatcherSv1(dspS.dspS), "", false)
	dspS.server.RpcRegister(srv)

	attrsv1, _ := birpc.NewServiceWithMethodsRename(dspS.dspS, utils.AttributeSv1, true, func(oldFn string) (newFn string) {
		if strings.HasPrefix(oldFn, utils.AttributeSv1) {
			return strings.TrimPrefix(oldFn, utils.AttributeSv1)
		}
		return
	})
	dspS.server.RpcRegisterName(utils.AttributeSv1, attrsv1)
	// for the moment we dispable Apier through dispatcher
	// until we figured out a better sollution in case of gob server
	// dspS.server.SetDispatched()
	/*

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

		dspS.server.RpcRegisterName(utils.CacheSv1,
			v1.NewDispatcherCacheSv1(dspS.dspS))

		dspS.server.RpcRegisterName(utils.GuardianSv1,
			v1.NewDispatcherGuardianSv1(dspS.dspS))

		dspS.server.RpcRegisterName(utils.CDRsV1,
			v1.NewDispatcherSCDRsV1(dspS.dspS))

		dspS.server.RpcRegisterName(utils.ConfigSv1,
			v1.NewDispatcherConfigSv1(dspS.dspS))

		dspS.server.RpcRegisterName(utils.CoreSv1,
			v1.NewDispatcherCoreSv1(dspS.dspS))

		dspS.server.RpcRegisterName(utils.ReplicatorSv1,
			v1.NewDispatcherReplicatorSv1(dspS.dspS))

		dspS.server.RpcRegisterName(utils.CDRsV2,
			v2.NewDispatcherSCDRsV2(dspS.dspS))

		dspS.server.RpcRegisterName(utils.RateSv1,
			v1.NewDispatcherRateSv1(dspS.dspS))

		dspS.server.RpcRegisterName(utils.ActionSv1,
			v1.NewDispatcherActionSv1(dspS.dspS))

		dspS.server.RpcRegisterName(utils.AccountSv1,
			v1.NewDispatcherAccountSv1(dspS.dspS))
	*/
	dspS.connChan <- dspS.anz.GetInternalCodec(srv, utils.DispatcherS)

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
	<-dspS.connChan
	dspS.unregisterAllDispatchedSubsystems()
	dspS.sync()
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

func (dspS *DispatcherService) unregisterAllDispatchedSubsystems() {
	dspS.server.RpcUnregisterName(utils.AttributeSv1)
}

func (dspS *DispatcherService) RegisterShutdownChan(subsys string) (c chan struct{}) {
	c = make(chan struct{})
	dspS.Lock()
	dspS.srvsReload[subsys] = c
	dspS.Unlock()
	return
}

func (dspS *DispatcherService) UnregisterShutdownChan(subsys string) {
	dspS.Lock()
	if dspS.srvsReload[subsys] != nil {
		close(dspS.srvsReload[subsys])
	}
	delete(dspS.srvsReload, subsys)
	dspS.Unlock()
}

func (dspS *DispatcherService) sync() {
	for _, c := range dspS.srvsReload {
		c <- struct{}{}
	}
}
