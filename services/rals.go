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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewRalService returns the Ral Service
func NewRalService(cfg *config.CGRConfig, cacheS *engine.CacheS, server *cores.Server,
	internalRALsChan, internalResponderChan chan rpcclient.ClientConnector, shdChan *utils.SyncedChan,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) *RalService {
	resp := NewResponderService(cfg, server, internalResponderChan, shdChan, anz, srvDep)

	return &RalService{
		connChan:  internalRALsChan,
		cfg:       cfg,
		cacheS:    cacheS,
		server:    server,
		responder: resp,
		connMgr:   connMgr,
		anz:       anz,
		srvDep:    srvDep,
	}
}

// RalService implements Service interface
type RalService struct {
	sync.RWMutex
	cfg       *config.CGRConfig
	cacheS    *engine.CacheS
	server    *cores.Server
	rals      *v1.RALsV1
	responder *ResponderService
	connChan  chan rpcclient.ClientConnector
	connMgr   *engine.ConnManager
	anz       *AnalyzerService
	srvDep    map[string]*sync.WaitGroup
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (rals *RalService) Start() (err error) {
	if rals.IsRunning() {
		return utils.ErrServiceAlreadyRunning
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
		if err != utils.ErrServiceAlreadyRunning {
			return
		}
		err = nil
	}

	rals.rals = v1.NewRALsV1()

	if !rals.cfg.DispatcherSCfg().Enabled {
		rals.server.RpcRegister(rals.rals)
	}

	rals.connChan <- rals.anz.GetInternalCodec(rals.rals, utils.RALService)
	return
}

// Reload handles the change of config
func (rals *RalService) Reload() (err error) {
	engine.SetRpSubjectPrefixMatching(rals.cfg.RalsCfg().RpSubjectPrefixMatching)
	if err = rals.responder.Reload(); err != nil {
		return
	}
	return
}

// Shutdown stops the service
func (rals *RalService) Shutdown() (err error) {
	rals.Lock()
	defer rals.Unlock()
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

// GetResponder returns the responder service
func (rals *RalService) GetResponder() *ResponderService {
	return rals.responder
}
