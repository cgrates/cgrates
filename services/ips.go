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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewIPService returns the IP Service.
func NewIPService(cfg *config.CGRConfig, dm *DataDBService,
	cache *engine.CacheS, fsChan chan *engine.FilterS,
	server *cores.Server, intIPsChan chan birpc.ClientConnector,
	cm *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &IPService{
		connChan: intIPsChan,
		cfg:      cfg,
		dbs:      dm,
		cache:    cache,
		fsChan:   fsChan,
		server:   server,
		cm:       cm,
		anz:      anz,
		srvDep:   srvDep,
	}
}

// IPService implements Service interface.
type IPService struct {
	mu  sync.RWMutex
	cfg *config.CGRConfig
	cm  *engine.ConnManager

	dbs    *DataDBService
	cache  *engine.CacheS
	fsChan chan *engine.FilterS

	ips      *engine.IPService
	server   *cores.Server
	connChan chan birpc.ClientConnector
	anz      *AnalyzerService

	srvDep map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (s *IPService) Start() error {
	if s.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	s.srvDep[utils.DataDB].Add(1)
	<-s.cache.GetPrecacheChannel(utils.CacheIPProfiles)
	<-s.cache.GetPrecacheChannel(utils.CacheIPAllocations)
	<-s.cache.GetPrecacheChannel(utils.CacheIPFilterIndexes)

	fltrs := <-s.fsChan
	s.fsChan <- fltrs
	dmChan := s.dbs.GetDMChan()
	dm := <-dmChan
	dmChan <- dm

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem",
		utils.CoreS, utils.IPs))
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ips = engine.NewIPService(dm, s.cfg, fltrs, s.cm)
	s.ips.StartLoop()
	srv, err := engine.NewService(v1.NewIPsV1(s.ips))
	if err != nil {
		return err
	}
	if !s.cfg.DispatcherSCfg().Enabled {
		s.server.RpcRegister(srv)
	}
	s.connChan <- s.anz.GetInternalCodec(srv, utils.IPs)
	return nil
}

// Reload handles configuration changes.
func (s *IPService) Reload() error {
	s.mu.Lock()
	s.ips.Reload()
	s.mu.Unlock()
	return nil
}

// Shutdown stops the service.
func (s *IPService) Shutdown() error {
	defer s.srvDep[utils.DataDB].Done()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ips.Shutdown()
	s.ips = nil
	<-s.connChan
	return nil
}

// ServiceName returns the service name.
func (s *IPService) ServiceName() string {
	return utils.IPs
}

// ShouldRun returns if the service should be running.
func (s *IPService) ShouldRun() bool {
	return s.cfg.IPsCfg().Enabled
}

// IsRunning checks whether the service is running.
func (s *IPService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ips != nil
}
