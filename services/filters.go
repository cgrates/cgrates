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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewFilterService instantiates a new FilterService.
func NewFilterService(cfg *config.CGRConfig, connMgr *engine.ConnManager,
	srvIndexer *servmanager.ServiceIndexer) *FilterService {
	return &FilterService{
		cfg:        cfg,
		connMgr:    connMgr,
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// FilterService implements Service interface.
type FilterService struct {
	mu sync.RWMutex

	fltrS *engine.FilterS

	cfg     *config.CGRConfig
	connMgr *engine.ConnManager

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start handles the service start.
func (s *FilterService) Start(shutdown chan struct{}) error {
	cacheS := s.srvIndexer.GetService(utils.CacheS).(*CacheService)
	if utils.StructChanTimeout(cacheS.StateChan(utils.StateServiceUP), s.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.FilterS, utils.CacheS, utils.StateServiceUP)
	}
	if err := cacheS.WaitToPrecache(shutdown, utils.CacheFilters); err != nil {
		return err
	}
	dbs := s.srvIndexer.GetService(utils.DataDB).(*DataDBService)
	if utils.StructChanTimeout(dbs.StateChan(utils.StateServiceUP), s.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.FilterS, utils.DataDB, utils.StateServiceUP)
	}
	s.fltrS = engine.NewFilterS(s.cfg, s.connMgr, dbs.DataManager())
	close(s.stateDeps.StateChan(utils.StateServiceUP))
	return nil
}

// Reload handles the config changes.
func (s *FilterService) Reload(_ chan struct{}) error {
	return nil
}

// Shutdown stops the service.
func (s *FilterService) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fltrS = nil
	return nil
}

// IsRunning returns whether the service is running or not.
func (s *FilterService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fltrS != nil
}

// ServiceName returns the service name
func (s *FilterService) ServiceName() string {
	return utils.FilterS
}

// ShouldRun returns if the service should be running.
func (s *FilterService) ShouldRun() bool {
	return true
}

// StateChan returns signaling channel of specific state
func (s *FilterService) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (s *FilterService) IntRPCConn() birpc.ClientConnector {
	return s.intRPCconn
}

// FilterS returns the FilterS object.
func (s *FilterService) FilterS() *engine.FilterS {
	return s.fltrS
}
