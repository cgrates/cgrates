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
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAttributeService returns the Attribute Service
func NewAttributeService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalChan chan birpc.ClientConnector,
	anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &AttributeService{
		connChan:    internalChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		server:      server,
		anz:         anz,
		srvDep:      srvDep,
	}
}

// AttributeService implements Service interface
type AttributeService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *cores.Server

	attrS    *engine.AttributeService
	rpc      *v1.AttributeSv1           // useful on restart
	connChan chan birpc.ClientConnector // publish the internal Subsystem when available
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the service start
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
	attrS.attrS = engine.NewAttributeService(datadb, filterS, attrS.cfg)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.AttributeS))
	attrS.rpc = v1.NewAttributeSv1(attrS.attrS)
	if !attrS.cfg.DispatcherSCfg().Enabled {
		attrS.server.RpcRegister(attrS.rpc)
	}
	attrS.connChan <- attrS.anz.GetInternalCodec(attrS.rpc, utils.AttributeS)
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
	attrS.attrS.Shutdown()
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
