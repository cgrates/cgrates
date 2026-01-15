/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type BiRPCThresholdSFuncs struct {
	thresholdSOnBiJSONConnect    func(c birpc.ClientConnector) //store OnBiJSONConnect
	thresholdSOnBiJSONDisconnect func(c birpc.ClientConnector) //store OnBiJSONDisconnect
}

// Returns a new BiRPCThresholdSFuncs struct
func NewThresholdSBiJSONFuncs(onConn, onDisconn func(c birpc.ClientConnector)) *BiRPCThresholdSFuncs {
	return &BiRPCThresholdSFuncs{
		thresholdSOnBiJSONConnect:    onConn,
		thresholdSOnBiJSONDisconnect: onDisconn,
	}
}

// GetThresholdSOnBiJSONFuncs returns ThresholdSOnBiJSONConnect and ThresholdSOnBiJSONDisconnect (1 time use)
func (thrs *ThresholdService) GetThresholdSOnBiJSONFuncs() (onConn, onDisconn func(c birpc.ClientConnector)) {
	<-thrs.biRPCFuncsDone // Make sure funcs are initialized in the structure before getting them
	return thrs.biRPCFuncs.thresholdSOnBiJSONConnect, thrs.biRPCFuncs.thresholdSOnBiJSONDisconnect
}

// NewThresholdService returns the Threshold Service
func NewThresholdService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalThresholdSChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) *ThresholdService {
	return &ThresholdService{
		connChan:       internalThresholdSChan,
		cfg:            cfg,
		dm:             dm,
		cacheS:         cacheS,
		filterSChan:    filterSChan,
		server:         server,
		connMgr:        connMgr,
		anz:            anz,
		srvDep:         srvDep,
		biRPCFuncsDone: make(chan struct{}),
	}
}

// ThresholdService implements Service interface
type ThresholdService struct {
	sync.RWMutex
	cfg            *config.CGRConfig
	dm             *DataDBService
	cacheS         *engine.CacheS
	filterSChan    chan *engine.FilterS
	server         *cores.Server
	birpcEnabled   bool
	biRPCFuncs     *BiRPCThresholdSFuncs
	biRPCFuncsDone chan struct{} // marks when biRPCFuncs are initialized
	connMgr        *engine.ConnManager

	thrs     *engine.ThresholdService
	connChan chan birpc.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (thrs *ThresholdService) Start() error {
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
	thrs.thrs = engine.NewThresholdService(datadb, thrs.cfg, filterS, thrs.connMgr)

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ThresholdS))
	thrs.thrs.StartLoop()
	srv, err := engine.NewService(v1.NewThresholdSv1(thrs.thrs))
	if err != nil {
		return err
	}
	if !thrs.cfg.DispatcherSCfg().Enabled {
		thrs.server.RpcRegister(srv)
	}
	thrs.connChan <- thrs.anz.GetInternalCodec(srv, utils.ThresholdS)
	// Register BiRpc handlers
	if thrs.cfg.ListenCfg().BiJSONListen != "" {
		thrs.birpcEnabled = true
		thrs.server.BiRPCRegisterName(utils.ThresholdSv1, srv)
		thrs.biRPCFuncs = NewThresholdSBiJSONFuncs(thrs.thrs.OnBiJSONConnect, thrs.thrs.OnBiJSONDisconnect)
		thrs.biRPCFuncsDone <- struct{}{}
	}
	return nil
}

func (thrs *ThresholdService) DisableBiRPC() {
	if thrs == nil {
		return
	}
	thrs.Lock()
	thrs.birpcEnabled = false
	thrs.Unlock()
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
	thrs.thrs.Shutdown()
	if thrs.birpcEnabled {
		thrs.server.StopBiRPC()
		thrs.birpcEnabled = false
	}
	thrs.thrs = nil
	<-thrs.connChan
	return
}

// IsRunning returns if the service is running
func (thrs *ThresholdService) IsRunning() bool {
	thrs.RLock()
	defer thrs.RUnlock()
	return thrs.thrs != nil
}

// ServiceName returns the service name
func (thrs *ThresholdService) ServiceName() string {
	return utils.ThresholdS
}

// ShouldRun returns if the service should be running
func (thrs *ThresholdService) ShouldRun() bool {
	return thrs.cfg.ThresholdSCfg().Enabled
}
