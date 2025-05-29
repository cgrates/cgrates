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

package servmanager

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewServiceManager returns a service manager
func NewServiceManager(shdWg *sync.WaitGroup, cfg *config.CGRConfig, registry *ServiceRegistry,
	services []Service) (sM *ServiceManager) {
	sM = &ServiceManager{
		cfg:      cfg,
		registry: registry,
		shdWg:    shdWg,
		rldChan:  cfg.GetReloadChan(),
	}
	sM.registry.Register(services...)
	return
}

// ServiceManager handles service management ran by the engine
type ServiceManager struct {
	sync.RWMutex // lock access to any shared data
	cfg          *config.CGRConfig
	registry     *ServiceRegistry // index here the services for accessing them by their IDs
	shdWg        *sync.WaitGroup  // list of shutdown items
	rldChan      <-chan string    // reload signals come over this channelc
}

// StartServices starts all enabled services
func (m *ServiceManager) StartServices(shutdown *utils.SyncedChan) {
	go m.handleReload(shutdown)
	for _, svc := range m.registry.List() {
		// TODO: verify if service state check is needed. It should
		// be redundant since ServManager manages all services and this
		// runs only at startup
		if svc.ShouldRun() && State(svc) == utils.StateServiceDOWN {
			m.shdWg.Add(1)
			go func() {
				if err := svc.Start(shutdown, m.registry); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> failed to start <%s> service: %v", utils.ServiceManager, svc.ServiceName(), err))
					shutdown.CloseOnce()
				}
				MustSetState(svc, utils.StateServiceUP)
				utils.Logger.Info(fmt.Sprintf("<%s> started <%s> service", utils.ServiceManager, svc.ServiceName()))
			}()
		}
	}
	// startServer()
}

func (m *ServiceManager) handleReload(shutdown *utils.SyncedChan) {
	var serviceID string
	for {
		select {
		case <-shutdown.Done():
			m.ShutdownServices()
			return
		case serviceID = <-m.rldChan:
		}
		go m.reloadService(serviceID, shutdown)
		// handle RPC server
	}
}

func (m *ServiceManager) reloadService(id string, shutdown *utils.SyncedChan) (err error) {
	svc := m.registry.Lookup(id)

	// Consider services in pending states (not up/down) to be up. This assumes
	// Start/Reload/Shutdown functions are handled synchronously.
	isUp := State(svc) != utils.StateServiceDOWN

	if svc.ShouldRun() {
		if isUp {
			if err = svc.Reload(shutdown, m.registry); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> failed to reload <%s> service: %v", utils.ServiceManager, svc.ServiceName(), err))
				shutdown.CloseOnce()
				return // stop if we encounter an error
			}
			utils.Logger.Info(fmt.Sprintf("<%s> reloaded <%s> service", utils.ServiceManager, svc.ServiceName()))
		} else {
			m.shdWg.Add(1)
			if err = svc.Start(shutdown, m.registry); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> failed to start <%s> serivce: %v", utils.ServiceManager, svc.ServiceName(), err))
				shutdown.CloseOnce()
				return // stop if we encounter an error
			}
			MustSetState(svc, utils.StateServiceUP)
			utils.Logger.Info(fmt.Sprintf("<%s> started <%s> service", utils.ServiceManager, svc.ServiceName()))
		}
	} else if isUp {
		if err = svc.Shutdown(m.registry); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed to shut down <%s> service: %v", utils.ServiceManager, svc.ServiceName(), err))
			shutdown.CloseOnce()
		}
		MustSetState(svc, utils.StateServiceDOWN)
		utils.Logger.Info(fmt.Sprintf("<%s> stopped <%s> service", utils.ServiceManager, svc.ServiceName()))
		m.shdWg.Done()
	}
	return
}

// ShutdownServices will stop all services
func (m *ServiceManager) ShutdownServices() {
	for _, svc := range m.registry.List() {
		if State(svc) != utils.StateServiceDOWN {
			go func() {
				defer m.shdWg.Done()
				if err := svc.Shutdown(m.registry); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> failed to shut down <%s> service: %v",
						utils.ServiceManager, svc.ServiceName(), err))
					return
				}
				MustSetState(svc, utils.StateServiceDOWN)
				utils.Logger.Info(fmt.Sprintf("<%s> stopped <%s> service", utils.ServiceManager, svc.ServiceName()))
			}()
		}
	}
}

// Service interface that describes what functions should a service implement
type Service interface {
	// Start should handle the service start
	Start(*utils.SyncedChan, *ServiceRegistry) error
	// Reload handles the change of config
	Reload(*utils.SyncedChan, *ServiceRegistry) error
	// Shutdown stops the service
	Shutdown(*ServiceRegistry) error
	// ShouldRun returns if the service should be running
	ShouldRun() bool
	// ServiceName returns the service name
	ServiceName() string
	// StateChan returns the channel for specific state subscription
	StateChan(stateID string) chan struct{}
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
	m.RLock()
	defer m.RUnlock()
	states := make(map[string]string)
	switch args.ServiceID {
	case utils.MetaAll:
		for _, svc := range m.registry.List() {
			states[svc.ServiceName()] = State(svc)
		}
	default:
		ids := strings.Split(args.ServiceID, utils.FieldsSep)
		for _, id := range ids {
			svc := m.registry.Lookup(id)
			if svc == nil {
				return fmt.Errorf("unsupported service ID: %q", id)
			}
			states[id] = State(svc)
		}
	}
	*reply = states
	return nil
}

// GetConfig returns the Configuration
func (m *ServiceManager) GetConfig() *config.CGRConfig {
	m.RLock()
	defer m.RUnlock()
	return m.cfg
}

func toggleService(id string, status bool, srvMngr *ServiceManager) (err error) {
	srvMngr.Lock()
	defer srvMngr.Unlock()
	switch id {
	case utils.AccountS:
		srvMngr.cfg.AccountSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.ActionS:
		srvMngr.cfg.ActionSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.AdminS:
		srvMngr.cfg.AdminSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.AnalyzerS:
		srvMngr.cfg.AnalyzerSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.AttributeS:
		srvMngr.cfg.AttributeSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.CDRServer:
		srvMngr.cfg.CdrsCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.ChargerS:
		srvMngr.cfg.ChargerSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.EEs:
		srvMngr.cfg.EEsCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.EFs:
		srvMngr.cfg.EFsCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.ERs:
		srvMngr.cfg.ERsCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
		// case utils.LoaderS:
		// 	srvMngr.cfg.LoaderCfg()[0].Enabled = status
		// 	srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.RateS:
		srvMngr.cfg.RateSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.TrendS:
		srvMngr.cfg.TrendSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.RankingS:
		srvMngr.cfg.RankingSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.ResourceS:
		srvMngr.cfg.ResourceSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.IPs:
		srvMngr.cfg.IPsCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.RouteS:
		srvMngr.cfg.RouteSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.SessionS:
		srvMngr.cfg.SessionSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.StatS:
		srvMngr.cfg.StatSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.ThresholdS:
		srvMngr.cfg.ThresholdSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.TPeS:
		srvMngr.cfg.TpeSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.AsteriskAgent:
		srvMngr.cfg.AsteriskAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.DiameterAgent:
		srvMngr.cfg.DiameterAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.DNSAgent:
		srvMngr.cfg.DNSAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.FreeSWITCHAgent:
		srvMngr.cfg.FsAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.KamailioAgent:
		srvMngr.cfg.KamAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.RadiusAgent:
		srvMngr.cfg.RadiusAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	case utils.SIPAgent:
		srvMngr.cfg.SIPAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- id
	default:
		err = utils.ErrUnsupportedServiceID
	}
	return
}

// MustSetState changes a service's state, panicking if it fails.
func MustSetState(svc Service, state string) {
	if err := SetState(svc, state); err != nil {
		panic(err)
	}
}

// SetServiceState moves the state signal to a new valid state.
// Returns an error if the state is invalid or if no current state was found.
func SetState(svc Service, state string) error {
	if !slices.Contains([]string{utils.StateServiceUP, utils.StateServiceDOWN}, state) {
		return fmt.Errorf("invalid service state: %q", state)
	}
	select {
	case <-svc.StateChan(utils.StateServiceUP):
	case <-svc.StateChan(utils.StateServiceDOWN):
	default:
		return fmt.Errorf("service %q in undefined state", svc.ServiceName())
	}
	svc.StateChan(state) <- struct{}{}
	return nil
}

// State returns the current state of a service by checking which channel holds
// the state signal. Returns empty string if no valid state is found.
func State(svc Service) string {
	select {
	case <-svc.StateChan(utils.StateServiceUP):
		svc.StateChan(utils.StateServiceUP) <- struct{}{}
		return utils.StateServiceUP
	case <-svc.StateChan(utils.StateServiceDOWN):
		svc.StateChan(utils.StateServiceDOWN) <- struct{}{}
		return utils.StateServiceDOWN
	default:
		return ""
	}
}
