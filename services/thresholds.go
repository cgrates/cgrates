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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewThresholdService returns the Threshold Service
func NewThresholdService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *CacheService, filterSChan chan *engine.FilterS,
	connMgr *engine.ConnManager,
	server *cores.Server, internalThresholdSChan chan birpc.ClientConnector,
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
		connMgr:     connMgr,
	}
}

// ThresholdService implements Service interface
type ThresholdService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *CacheService
	filterSChan chan *engine.FilterS
	server      *cores.Server
	connMgr     *engine.ConnManager

	thrs     *engine.ThresholdService
	connChan chan birpc.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (thrs *ThresholdService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if thrs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	thrs.srvDep[utils.DataDB].Add(1)
	if err = thrs.cacheS.WaitToPrecache(ctx,
		utils.CacheThresholdProfiles,
		utils.CacheThresholds,
		utils.CacheThresholdFilterIndexes); err != nil {
		return
	}

	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, thrs.filterSChan); err != nil {
		return
	}

	var datadb *engine.DataManager
	if datadb, err = thrs.dm.WaitForDM(ctx); err != nil {
		return
	}

	thrs.Lock()
	defer thrs.Unlock()
	thrs.thrs = engine.NewThresholdService(datadb, thrs.cfg, filterS, thrs.connMgr)

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ThresholdS))
	thrs.thrs.StartLoop(ctx)
	srv, _ := birpc.NewService(apis.NewThresholdSv1(thrs.thrs), "", false)
	if !thrs.cfg.DispatcherSCfg().Enabled {
		thrs.server.RpcRegister(srv)
	}
	thrs.connChan <- thrs.anz.GetInternalCodec(srv, utils.ThresholdS)
	return
}

// Reload handles the change of config
func (thrs *ThresholdService) Reload(ctx *context.Context, _ context.CancelFunc) (_ error) {
	thrs.Lock()
	thrs.thrs.Reload(ctx)
	thrs.Unlock()
	return
}

// Shutdown stops the service
func (thrs *ThresholdService) Shutdown() (_ error) {
	defer thrs.srvDep[utils.DataDB].Done()
	thrs.Lock()
	defer thrs.Unlock()
	thrs.thrs.Shutdown(context.TODO())
	thrs.thrs = nil
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
