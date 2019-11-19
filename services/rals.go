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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewRalService returns the Ral Service
func NewRalService(cfg *config.CGRConfig, dm *DataDBService,
	storDB *StorDBService, cacheS *engine.CacheS, filterSChan chan *engine.FilterS, server *utils.Server,
	thsChan, stsChan, cacheSChan, schedChan, attrsChan, dispatcherChan chan rpcclient.RpcClientConnection,
	schedulerService *SchedulerService, exitChan chan bool) *RalService {
	resp := NewResponderService(cfg, server, thsChan, stsChan, dispatcherChan, exitChan)
	apiv1 := NewApierV1Service(cfg, dm, storDB, filterSChan, server, cacheSChan, schedChan, attrsChan, dispatcherChan, schedulerService, resp)
	apiv2 := NewApierV2Service(apiv1, cfg, server)
	return &RalService{
		connChan:  make(chan rpcclient.RpcClientConnection, 1),
		cfg:       cfg,
		cacheS:    cacheS,
		server:    server,
		apiv1:     apiv1,
		apiv2:     apiv2,
		responder: resp,
	}
}

// RalService implements Service interface
type RalService struct {
	sync.RWMutex
	cfg       *config.CGRConfig
	cacheS    *engine.CacheS
	server    *utils.Server
	rals      *v1.RALsV1
	apiv1     *ApierV1Service
	apiv2     *ApierV2Service
	responder *ResponderService
	connChan  chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (rals *RalService) Start() (err error) {
	if rals.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	engine.SetRpSubjectPrefixMatching(rals.cfg.RalsCfg().RpSubjectPrefixMatching)
	rals.Lock()
	defer rals.Unlock()

	<-rals.cacheS.GetPrecacheChannel(utils.CacheDestinations)
	<-rals.cacheS.GetPrecacheChannel(utils.CacheReverseDestinations)
	<-rals.cacheS.GetPrecacheChannel(utils.CacheRatingPlans)
	<-rals.cacheS.GetPrecacheChannel(utils.CacheRatingProfiles)
	<-rals.cacheS.GetPrecacheChannel(utils.CacheActions)
	<-rals.cacheS.GetPrecacheChannel(utils.CacheActionPlans)
	<-rals.cacheS.GetPrecacheChannel(utils.CacheAccountActionPlans)
	<-rals.cacheS.GetPrecacheChannel(utils.CacheActionTriggers)
	<-rals.cacheS.GetPrecacheChannel(utils.CacheSharedGroups)
	<-rals.cacheS.GetPrecacheChannel(utils.CacheTimings)

	if err = rals.responder.Start(); err != nil {
		return
	}

	if err = rals.apiv1.Start(); err != nil {
		return
	}

	if err = rals.apiv2.Start(); err != nil {
		return
	}

	rals.rals = v1.NewRALsV1()

	if !rals.cfg.DispatcherSCfg().Enabled {
		rals.server.RpcRegister(rals.rals)
	}

	utils.RegisterRpcParams(utils.RALsV1, rals.rals)

	rals.connChan <- rals.rals
	return
}

// GetIntenternalChan returns the internal connection chanel
func (rals *RalService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return rals.connChan
}

// Reload handles the change of config
func (rals *RalService) Reload() (err error) {
	engine.SetRpSubjectPrefixMatching(rals.cfg.RalsCfg().RpSubjectPrefixMatching)
	if err = rals.apiv1.Reload(); err != nil {
		return
	}
	if err = rals.apiv2.Reload(); err != nil {
		return
	}
	if err = rals.responder.Reload(); err != nil {
		return
	}
	return
}

// Shutdown stops the service
func (rals *RalService) Shutdown() (err error) {
	rals.Lock()
	defer rals.Unlock()
	if err = rals.apiv1.Shutdown(); err != nil {
		return
	}
	if err = rals.apiv2.Shutdown(); err != nil {
		return
	}
	if err = rals.responder.Shutdown(); err != nil {
		return
	}
	rals.rals = nil
	<-rals.connChan
	return
}

// IsRunning returns if the service is running
func (rals *RalService) IsRunning() bool {
	rals.RLock()
	defer rals.RUnlock()
	return rals != nil && rals.rals != nil
}

// ServiceName returns the service name
func (rals *RalService) ServiceName() string {
	return utils.RALService
}

// ShouldRun returns if the service should be running
func (rals *RalService) ShouldRun() bool {
	return rals.cfg.RalsCfg().Enabled
}

// GetAPIv1 returns the apiv1 service
func (rals *RalService) GetAPIv1() servmanager.Service {
	return rals.apiv1
}

// GetAPIv2 returns the apiv2 service
func (rals *RalService) GetAPIv2() servmanager.Service {
	return rals.apiv2
}

// GetResponder returns the responder service
func (rals *RalService) GetResponder() servmanager.Service {
	return rals.responder
}
