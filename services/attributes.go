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

// NewAttributeService returns the Attribute Service
func NewAttributeService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *utils.Server, internalChan chan birpc.ClientConnector) servmanager.Service {
	return &AttributeService{
		connChan:    internalChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		server:      server,
	}
}

// AttributeService implements Service interface
type AttributeService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *utils.Server

	attrS    *engine.AttributeService
	rpc      *v1.AttributeSv1
	connChan chan birpc.ClientConnector
}

// Start should handle the sercive start
func (attrS *AttributeService) Start() (err error) {
	if attrS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	<-attrS.cacheS.GetPrecacheChannel(utils.CacheAttributeProfiles)
	<-attrS.cacheS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes)

	filterS := <-attrS.filterSChan
	attrS.filterSChan <- filterS
	dbchan := attrS.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	attrS.Lock()
	defer attrS.Unlock()
	attrS.attrS, err = engine.NewAttributeService(datadb, filterS, attrS.cfg)
	if err != nil {
		utils.Logger.Crit(
			fmt.Sprintf("<%s> Could not init, error: %s",
				utils.AttributeS, err.Error()))
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.AttributeS))
	attrS.rpc = v1.NewAttributeSv1(attrS.attrS)
	if !attrS.cfg.DispatcherSCfg().Enabled {
		attrS.server.RpcRegister(attrS.rpc)
	}
	attrS.connChan <- attrS.rpc
	return
}

// Reload handles the change of config
func (attrS *AttributeService) Reload() (err error) {
	return // for the moment nothing to reload
}

// Shutdown stops the service
func (attrS *AttributeService) Shutdown() (err error) {
	attrS.Lock()
	defer attrS.Unlock()
	if err = attrS.attrS.Shutdown(); err != nil {
		return
	}
	attrS.attrS = nil
	attrS.rpc = nil
	<-attrS.connChan
	return
}

// IsRunning returns if the service is running
func (attrS *AttributeService) IsRunning() bool {
	attrS.RLock()
	defer attrS.RUnlock()
	return attrS != nil && attrS.attrS != nil
}

// ServiceName returns the service name
func (attrS *AttributeService) ServiceName() string {
	return utils.AttributeS
}

// ShouldRun returns if the service should be running
func (attrS *AttributeService) ShouldRun() bool {
	return attrS.cfg.AttributeSCfg().Enabled
}
