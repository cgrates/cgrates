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

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewDispatcherService returns the Dispatcher Service
func NewDispatcherService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalChan chan birpc.ClientConnector,
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
	connChan chan birpc.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (dspS *DispatcherService) Start() error {
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

	dspS.server.RpcUnregisterName(utils.AttributeSv1)

	srv, err := newDispatcherServiceMap(dspS.dspS)
	if err != nil {
		return err
	}
	for _, s := range srv {
		dspS.server.RpcRegister(s)
	}
	dspS.connChan <- dspS.anz.GetInternalCodec(srv, utils.DispatcherS)

	return nil
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
	return dspS.dspS != nil
}

// ServiceName returns the service name
func (dspS *DispatcherService) ServiceName() string {
	return utils.DispatcherS
}

// ShouldRun returns if the service should be running
func (dspS *DispatcherService) ShouldRun() bool {
	return dspS.cfg.DispatcherSCfg().Enabled
}

func newDispatcherServiceMap(val *dispatchers.DispatcherService) (engine.IntService, error) {
	srvMap := make(engine.IntService)

	srv, err := birpc.NewService(v1.NewDispatcherAttributeSv1(val),
		utils.AttributeSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherCacheSv1(val),
		utils.CacheSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherSCDRsV1(val),
		utils.CDRsV1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v2.NewDispatcherSCDRsV2(val),
		utils.CDRsV2, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherChargerSv1(val),
		utils.ChargerSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherConfigSv1(val),
		utils.ConfigSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherCoreSv1(val),
		utils.CoreSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherSv1(val),
		utils.DispatcherSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherEeSv1(val),
		utils.EeSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherErSv1(val),
		utils.ErSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherGuardianSv1(val),
		utils.GuardianSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherRALsV1(val),
		utils.RALsV1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherReplicatorSv1(val),
		utils.ReplicatorSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherResourceSv1(val),
		utils.ResourceSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherThresholdSv1(val),
		utils.ThresholdSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherRankingSv1(val), utils.RankingSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherTrendSv1(val), utils.TrendSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherResponder(val),
		utils.Responder, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherRouteSv1(val),
		utils.RouteSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherSchedulerSv1(val),
		utils.SchedulerSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherSessionSv1(val),
		utils.SessionSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	srv, err = birpc.NewService(v1.NewDispatcherStatSv1(val),
		utils.StatSv1, true)
	if err != nil {
		return nil, err
	}
	srvMap[srv.Name] = srv

	return srvMap, nil
}
