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
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewServiceManager returns a service manager
func NewServiceManager(cfg *config.CGRConfig, shdWg *sync.WaitGroup, connMgr *engine.ConnManager) *ServiceManager {
	return &ServiceManager{
		cfg:        cfg,
		subsystems: make(map[string]Service),
		shdWg:      shdWg,
		connMgr:    connMgr,
	}
}

// ServiceManager handles service management ran by the engine
type ServiceManager struct {
	sync.RWMutex // lock access to any shared data
	cfg          *config.CGRConfig
	subsystems   map[string]Service

	shdWg   *sync.WaitGroup
	connMgr *engine.ConnManager
}

// GetConfig returns the Configuration
func (srvMngr *ServiceManager) GetConfig() *config.CGRConfig {
	srvMngr.RLock()
	defer srvMngr.RUnlock()
	return srvMngr.cfg
}

// StartServices starts all enabled services
func (srvMngr *ServiceManager) StartServices(ctx *context.Context, shtDwn context.CancelFunc) {
	go srvMngr.handleReload(ctx, shtDwn)
	for _, service := range srvMngr.subsystems {
		if service.ShouldRun() && !service.IsRunning() {
			srvMngr.shdWg.Add(1)
			go func(srv Service) {
				if err := srv.Start(ctx, shtDwn); err != nil &&
					err != utils.ErrServiceAlreadyRunning { // in case the service was started in another gorutine
					utils.Logger.Err(fmt.Sprintf("<%s> failed to start %s because: %s", utils.ServiceManager, srv.ServiceName(), err))
					shtDwn()
				}
			}(service)
		}
	}
	// startServer()
}

// AddServices adds given services
func (srvMngr *ServiceManager) AddServices(services ...Service) {
	srvMngr.Lock()
	for _, srv := range services {
		if _, has := srvMngr.subsystems[srv.ServiceName()]; !has { // do not rewrite the service
			srvMngr.subsystems[srv.ServiceName()] = srv
		}
	}
	srvMngr.Unlock()
}

func (srvMngr *ServiceManager) handleReload(ctx *context.Context, shtDwn context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			srvMngr.ShutdownServices()
			return
		case <-srvMngr.GetConfig().GetReloadChan(config.AttributeSJSON):
			go srvMngr.reloadService(utils.AttributeS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.ChargerSJSON):
			go srvMngr.reloadService(utils.ChargerS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.ThresholdSJSON):
			go srvMngr.reloadService(utils.ThresholdS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.StatSJSON):
			go srvMngr.reloadService(utils.StatS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.ResourceSJSON):
			go srvMngr.reloadService(utils.ResourceS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.RouteSJSON):
			go srvMngr.reloadService(utils.RouteS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.AdminSJSON):
			go srvMngr.reloadService(utils.AdminS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.CDRsJSON):
			go srvMngr.reloadService(utils.CDRServer, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.SessionSJSON):
			go srvMngr.reloadService(utils.SessionS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.ERsJSON):
			go srvMngr.reloadService(utils.ERs, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.DNSAgentJSON):
			go srvMngr.reloadService(utils.DNSAgent, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.FreeSWITCHAgentJSON):
			go srvMngr.reloadService(utils.FreeSWITCHAgent, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.KamailioAgentJSON):
			go srvMngr.reloadService(utils.KamailioAgent, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.AsteriskAgentJSON):
			go srvMngr.reloadService(utils.AsteriskAgent, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.RadiusAgentJSON):
			go srvMngr.reloadService(utils.RadiusAgent, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.DiameterAgentJSON):
			go srvMngr.reloadService(utils.DiameterAgent, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.HTTPAgentJSON):
			go srvMngr.reloadService(utils.HTTPAgent, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.LoaderSJSON):
			go srvMngr.reloadService(utils.LoaderS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.AnalyzerSJSON):
			go srvMngr.reloadService(utils.AnalyzerS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.DispatcherSJSON):
			go srvMngr.reloadService(utils.DispatcherS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.DataDBJSON):
			go srvMngr.reloadService(utils.DataDB, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.StorDBJSON):
			go srvMngr.reloadService(utils.StorDB, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.EEsJSON):
			go srvMngr.reloadService(utils.EEs, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.RateSJSON):
			go srvMngr.reloadService(utils.RateS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.RPCConnsJSON):
			go srvMngr.connMgr.Reload()
		case <-srvMngr.GetConfig().GetReloadChan(config.SIPAgentJSON):
			go srvMngr.reloadService(utils.SIPAgent, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.RegistrarCJSON):
			go srvMngr.reloadService(utils.RegistrarC, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.HTTPJSON):
			go srvMngr.reloadService(utils.GlobalVarS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.AccountSJSON):
			go srvMngr.reloadService(utils.AccountS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.ActionSJSON):
			go srvMngr.reloadService(utils.ActionS, ctx, shtDwn)
		case <-srvMngr.GetConfig().GetReloadChan(config.CoreSJSON):
			go srvMngr.reloadService(utils.CoreS, ctx, shtDwn)
		}
		// handle RPC server
	}
}

func (srvMngr *ServiceManager) reloadService(srviceName string, ctx *context.Context, shtDwn context.CancelFunc) (err error) {
	srv := srvMngr.GetService(srviceName)
	if srv.ShouldRun() {
		if srv.IsRunning() {
			if err = srv.Reload(ctx, shtDwn); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> failed to reload <%s> err <%s>", utils.ServiceManager, srv.ServiceName(), err))
				shtDwn()
				return // stop if we encounter an error
			}
		} else {
			srvMngr.shdWg.Add(1)
			if err = srv.Start(ctx, shtDwn); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> failed to start <%s> err <%s>", utils.ServiceManager, srv.ServiceName(), err))
				shtDwn()
				return // stop if we encounter an error
			}
		}
	} else if srv.IsRunning() {
		if err = srv.Shutdown(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed to stop service <%s> err <%s>", utils.ServiceManager, srv.ServiceName(), err))
			shtDwn()
		}
		srvMngr.shdWg.Done()
	}
	return
}

// GetService returns the named service
func (srvMngr *ServiceManager) GetService(subsystem string) (srv Service) {
	srvMngr.RLock()
	srv = srvMngr.subsystems[subsystem]
	srvMngr.RUnlock()
	return
}

// ShutdownServices will stop all services
func (srvMngr *ServiceManager) ShutdownServices() {
	for _, srv := range srvMngr.subsystems { // gracefully stop all running subsystems
		if srv.IsRunning() {
			go func(srv Service) {
				if err := srv.Shutdown(); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Failed to shutdown subsystem <%s> because: %s",
						utils.ServiceManager, srv.ServiceName(), err))
				}
				srvMngr.shdWg.Done()
			}(srv)
		}
	}
}

// Service interface that describes what functions should a service implement
type Service interface {
	// Start should handle the service start
	Start(*context.Context, context.CancelFunc) error
	// Reload handles the change of config
	Reload(*context.Context, context.CancelFunc) error
	// Shutdown stops the service
	Shutdown() error
	// IsRunning returns if the service is running
	IsRunning() bool
	// ShouldRun returns if the service should be running
	ShouldRun() bool
	// ServiceName returns the service name
	ServiceName() string
}
