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

package servmanager

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewServiceManager returns a service manager.
func NewServiceManager(shdWg *sync.WaitGroup, cfg *config.CGRConfig, registry *Registry,
	services []Service) *ServiceManager {
	sm := &ServiceManager{
		cfg:      cfg,
		registry: registry,
		shdWg:    shdWg,
		reload:   cfg.GetReloadChan(),
	}
	sm.registry.Register(services...)
	return sm
}

// ServiceManager manages service lifecycle.
type ServiceManager struct {
	mu       sync.RWMutex // protects cfg
	cfg      *config.CGRConfig
	registry *Registry
	shdWg    *sync.WaitGroup
	reload   <-chan string
}

// StartServices starts all enabled services concurrently.
func (m *ServiceManager) StartServices(shutdown *utils.SyncedChan) {
	go m.handleReload(shutdown)
	for _, svc := range m.registry.List() {
		if !svc.ShouldRun() {
			continue
		}
		m.shdWg.Add(1)
		go m.start(svc, shutdown)
	}
}

// start runs svc.Start under the lifecycle lock.
func (m *ServiceManager) start(svc Service, shutdown *utils.SyncedChan) {
	id := svc.ServiceName()
	unlock := m.registry.LockService(id)
	defer unlock()
	if m.registry.State(id) == utils.StateServiceUP {
		m.shdWg.Done()
		return
	}
	if err := svc.Start(shutdown, m.registry); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed to start <%s> service: %v",
			utils.ServiceManager, id, err))
		m.shdWg.Done()
		shutdown.CloseOnce()
		return
	}
	_ = m.registry.SetState(id, utils.StateServiceUP)
	utils.Logger.Info(fmt.Sprintf("<%s> started <%s> service",
		utils.ServiceManager, id))
}

// stop runs svc.Shutdown under the lifecycle lock. Blocks until any
// running Start or Reload finishes.
func (m *ServiceManager) stop(svc Service) {
	id := svc.ServiceName()
	unlock := m.registry.LockService(id)
	defer unlock()
	if m.registry.State(id) != utils.StateServiceUP {
		return
	}
	defer m.shdWg.Done()
	err := svc.Shutdown(m.registry)
	_ = m.registry.SetState(id, utils.StateServiceDOWN)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed to shut down <%s> service: %v",
			utils.ServiceManager, id, err))
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> stopped <%s> service",
		utils.ServiceManager, id))
}

func (m *ServiceManager) handleReload(shutdown *utils.SyncedChan) {
	var serviceID string
	for {
		select {
		case <-shutdown.Done():
			m.ShutdownServices()
			return
		case serviceID = <-m.reload:
		}
		go m.reloadService(serviceID, shutdown)
	}
}

func (m *ServiceManager) reloadService(id string, shutdown *utils.SyncedChan) {
	svc := m.registry.Lookup(id)
	if svc == nil {
		return
	}
	unlock := m.registry.LockService(id)
	defer unlock()
	isUp := m.registry.State(id) == utils.StateServiceUP
	if svc.ShouldRun() {
		if isUp {
			if err := svc.Reload(shutdown, m.registry); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> failed to reload <%s> service: %v",
					utils.ServiceManager, id, err))
				shutdown.CloseOnce()
				return
			}
			utils.Logger.Info(fmt.Sprintf("<%s> reloaded <%s> service",
				utils.ServiceManager, id))
			return
		}
		m.shdWg.Add(1)
		if err := svc.Start(shutdown, m.registry); err != nil {
			m.shdWg.Done()
			utils.Logger.Err(fmt.Sprintf("<%s> failed to start <%s> service: %v",
				utils.ServiceManager, id, err))
			shutdown.CloseOnce()
			return
		}
		_ = m.registry.SetState(id, utils.StateServiceUP)
		utils.Logger.Info(fmt.Sprintf("<%s> started <%s> service",
			utils.ServiceManager, id))
		return
	}
	if !isUp {
		return
	}
	defer m.shdWg.Done()
	err := svc.Shutdown(m.registry)
	_ = m.registry.SetState(id, utils.StateServiceDOWN)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed to shut down <%s> service: %v",
			utils.ServiceManager, id, err))
		shutdown.CloseOnce()
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> stopped <%s> service",
		utils.ServiceManager, id))
}

// ShutdownServices stops all services in parallel and waits for them.
func (m *ServiceManager) ShutdownServices() {
	var wg sync.WaitGroup
	for _, svc := range m.registry.List() {
		wg.Go(func() {
			m.stop(svc)
		})
	}
	wg.Wait()
}

// Service describes the lifecycle a registered subsystem must implement.
type Service interface {
	Start(*utils.SyncedChan, *Registry) error
	Reload(*utils.SyncedChan, *Registry) error
	Shutdown(*Registry) error
	ShouldRun() bool
	ServiceName() string
}

// ArgsServiceID are passed to Start/Stop/Status RPC methods
type ArgsServiceID struct {
	ServiceID string
	APIOpts   map[string]any
}

// V1StartService starts a service with ID
func (m *ServiceManager) V1StartService(ctx *context.Context, args *ArgsServiceID, reply *string) (err error) {
	err = toggleService(args.ServiceID, true, m)
	if err != nil {
		return
	}
	*reply = utils.OK
	return
}

// V1StopService shuts-down a service with ID
func (m *ServiceManager) V1StopService(ctx *context.Context, args *ArgsServiceID, reply *string) (err error) {
	err = toggleService(args.ServiceID, false, m)
	if err != nil {
		return
	}
	*reply = utils.OK
	return
}

// V1ServiceStatus returns the current state of the specified services.
func (m *ServiceManager) V1ServiceStatus(ctx *context.Context, args *ArgsServiceID, reply *map[string]string) error {
	states := make(map[string]string)
	switch args.ServiceID {
	case utils.MetaAll:
		for _, svc := range m.registry.List() {
			id := svc.ServiceName()
			states[id] = m.registry.State(id)
		}
	default:
		ids := strings.Split(args.ServiceID, utils.FieldsSep)
		for _, id := range ids {
			if m.registry.Lookup(id) == nil {
				return fmt.Errorf("unsupported service ID: %q", id)
			}
			states[id] = m.registry.State(id)
		}
	}
	*reply = states
	return nil
}

// GetConfig returns the Configuration
func (m *ServiceManager) GetConfig() *config.CGRConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

func toggleService(id string, status bool, sm *ServiceManager) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	switch id {
	case utils.AccountS:
		sm.cfg.AccountSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.ActionS:
		sm.cfg.ActionSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.AdminS:
		sm.cfg.AdminSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.AnalyzerS:
		sm.cfg.AnalyzerSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.AttributeS:
		sm.cfg.AttributeSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.CDRServer:
		sm.cfg.CdrsCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.ChargerS:
		sm.cfg.ChargerSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.EEs:
		sm.cfg.EEsCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.EFs:
		sm.cfg.EFsCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.ERs:
		sm.cfg.ERsCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
		// case utils.LoaderS:
		// 	srvMngr.cfg.LoaderCfg()[0].Enabled = status
		// 	srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.RateS:
		sm.cfg.RateSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.TrendS:
		sm.cfg.TrendSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.RankingS:
		sm.cfg.RankingSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.ResourceS:
		sm.cfg.ResourceSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.IPs:
		sm.cfg.IPsCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.RouteS:
		sm.cfg.RouteSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.SessionS:
		sm.cfg.SessionSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.StatS:
		sm.cfg.StatSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.ThresholdS:
		sm.cfg.ThresholdSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.TPeS:
		sm.cfg.TpeSCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.AsteriskAgent:
		sm.cfg.AsteriskAgentCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.DiameterAgent:
		sm.cfg.DiameterAgentCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.DNSAgent:
		sm.cfg.DNSAgentCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.FreeSWITCHAgent:
		sm.cfg.FsAgentCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.KamailioAgent:
		sm.cfg.KamAgentCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.RadiusAgent:
		sm.cfg.RadiusAgentCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	case utils.SIPAgent:
		sm.cfg.SIPAgentCfg().Enabled = status
		sm.cfg.GetReloadChan() <- id
	default:
		return utils.ErrUnsupportedServiceID
	}
	return nil
}
