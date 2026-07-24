// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewThresholdService returns the Threshold Service
func NewThresholdService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *utils.Server, internalThresholdSChan chan birpc.ClientConnector) servmanager.Service {
	return &ThresholdService{
		connChan:    internalThresholdSChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		server:      server,
	}
}

// ThresholdService implements Service interface
type ThresholdService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *utils.Server

	thrs     *engine.ThresholdService
	rpc      *v1.ThresholdSv1
	connChan chan birpc.ClientConnector
}

// Start should handle the sercive start
func (thrs *ThresholdService) Start() (err error) {
	if thrs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

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
	thrs.thrs, err = engine.NewThresholdService(datadb, thrs.cfg, filterS)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.ThresholdS, err.Error()))
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ThresholdS))
	thrs.thrs.StartLoop()
	thrs.rpc = v1.NewThresholdSv1(thrs.thrs)
	if !thrs.cfg.DispatcherSCfg().Enabled {
		thrs.server.RpcRegister(thrs.rpc)
	}
	thrs.connChan <- thrs.rpc
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
