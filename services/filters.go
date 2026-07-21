// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
		cfg: cfg,
	}
}

// FilterService implements Service interface.
type FilterService struct {
	mu    sync.RWMutex
	cfg   *config.CGRConfig
	fltrS *engine.FilterS
}

// Start handles the service start.
func (s *FilterService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.CacheS,
			utils.DB,
		},
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown, utils.CacheFilters); err != nil {
		return err
	}
	dbs := srvDeps[utils.DB].(*DBService)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.fltrS = engine.NewFilterS(s.cfg, cms.ConnManager(), dbs.DataManager())
	s.fltrS.SetCache(cacheS.CacheS())
	return nil
}

// Reload handles the config changes.
func (s *FilterService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	return nil
}

// Shutdown stops the service.
func (s *FilterService) Shutdown(_ *servmanager.Registry) error {
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

// FilterS returns the FilterS object.
func (s *FilterService) FilterS() *engine.FilterS {
	return s.fltrS
}
