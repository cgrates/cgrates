// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRalService returns the Ral Service
func NewRalService(cfg *config.CGRConfig, cacheS *engine.CacheS, server *utils.Server,
	internalRALsChan, internalResponderChan chan birpc.ClientConnector, exitChan chan bool,
	connMgr *engine.ConnManager) *RalService {
	resp := NewResponderService(cfg, server, internalResponderChan, exitChan)

	return &RalService{
		connChan:  internalRALsChan,
		cfg:       cfg,
		cacheS:    cacheS,
		server:    server,
		responder: resp,
		connMgr:   connMgr,
	}
}

// RalService implements Service interface
type RalService struct {
	sync.RWMutex
	cfg       *config.CGRConfig
	cacheS    *engine.CacheS
	server    *utils.Server
	rals      *v1.RALsV1
	responder *ResponderService
	connChan  chan birpc.ClientConnector
	connMgr   *engine.ConnManager
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
func (rals *RalService) GetResponder() servmanager.Service {
	return rals.responder
}

// GetResponder returns the responder service
func (rals *RalService) GetResponderService() *ResponderService {
	return rals.responder
}
