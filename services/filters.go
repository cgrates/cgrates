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
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewFilterService instantiates a new FilterService.
func NewFilterService(cfg *config.CGRConfig) *FilterService {
	return &FilterService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// FilterService implements Service interface.
type FilterService struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	fltrS     *engine.FilterS
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start handles the service start.
func (s *FilterService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.CacheS,
			utils.DataDB,
		},
		registry, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown, utils.CacheFilters); err != nil {
		return err
	}
	dbs := srvDeps[utils.DataDB].(*DataDBService)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.fltrS = engine.NewFilterS(s.cfg, cms.ConnManager(), dbs.DataManager())
	return nil
}

// Reload handles the config changes.
func (s *FilterService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) error {
	return nil
}

// Shutdown stops the service.
func (s *FilterService) Shutdown(_ *servmanager.ServiceRegistry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fltrS = nil
	return nil
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

// FilterS returns the FilterS object.
func (s *FilterService) FilterS() *engine.FilterS {
	return s.fltrS
}
