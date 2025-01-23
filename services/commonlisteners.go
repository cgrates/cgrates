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

	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/registrarc"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCommonListenerService instantiates a new CommonListenerService.
func NewCommonListenerService(cfg *config.CGRConfig) *CommonListenerService {
	return &CommonListenerService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// CommonListenerService implements Service interface.
type CommonListenerService struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	cls       *commonlisteners.CommonListenerS
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start handles the service start.
func (s *CommonListenerService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	cs, err := WaitForServiceState(utils.StateServiceUP, utils.CapS, registry,
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cls = commonlisteners.NewCommonListenerS(cs.(*CapService).Caps())
	if len(s.cfg.HTTPCfg().RegistrarSURL) != 0 {
		s.cls.RegisterHTTPFunc(s.cfg.HTTPCfg().RegistrarSURL, registrarc.Registrar)
	}
	if s.cfg.ConfigSCfg().Enabled {
		s.cls.RegisterHTTPFunc(s.cfg.ConfigSCfg().URL, config.HandlerConfigS)
	}
	return nil
}

// Reload handles the config changes.
func (s *CommonListenerService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) error {
	return nil
}

// Shutdown stops the service.
func (s *CommonListenerService) Shutdown(registry *servmanager.ServiceRegistry) error {
	deps := []string{
		utils.AccountS,
		utils.ActionS,
		utils.AdminS,
		utils.AnalyzerS,
		utils.AttributeS,
		utils.CacheS,
		utils.CDRServer,
		utils.ChargerS,
		utils.ConfigS,
		utils.CoreS,
		utils.EEs,
		utils.EFs,
		utils.ERs,
		utils.GuardianS,
		utils.HTTPAgent,
		utils.JanusAgent,
		utils.LoaderS,
		utils.RankingS,
		utils.RateS,
		utils.RegistrarC,
		utils.ResourceS,
		utils.RouteS,
		utils.SessionS,
		utils.StatS,
		utils.ThresholdS,
		utils.TPeS,
		utils.TrendS,
	}
	for _, svcID := range deps {
		if !servmanager.IsServiceInState(registry.Lookup(svcID), utils.StateServiceUP) {
			continue
		}
		_, err := WaitForServiceState(utils.StateServiceDOWN, svcID, registry, s.cfg.GeneralCfg().ConnectTimeout)
		if err != nil {
			return err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cls = nil
	return nil
}

// ServiceName returns the service name
func (s *CommonListenerService) ServiceName() string {
	return utils.CommonListenerS
}

// ShouldRun returns if the service should be running.
func (s *CommonListenerService) ShouldRun() bool {
	return true
}

// StateChan returns signaling channel of specific state
func (s *CommonListenerService) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}

// CLS returns the CommonListenerS object.
func (s *CommonListenerService) CLS() *commonlisteners.CommonListenerS {
	return s.cls
}
