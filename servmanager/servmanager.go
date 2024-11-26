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
func NewServiceManager(shdWg *sync.WaitGroup, connMgr *engine.ConnManager,
	cfg *config.CGRConfig, services []Service) (sM *ServiceManager) {
	sM = &ServiceManager{
		cfg:        cfg,
		subsystems: make(map[string]Service),
		shdWg:      shdWg,
		connMgr:    connMgr,
		rldChan:    cfg.GetReloadChan(),
	}
	sM.AddServices(services...)
	return
}

// ServiceManager handles service management ran by the engine
type ServiceManager struct {
	sync.RWMutex // lock access to any shared data
	cfg          *config.CGRConfig
	subsystems   map[string]Service // active subsystems managed by SM

	shdWg   *sync.WaitGroup // list of shutdown items
	connMgr *engine.ConnManager
	rldChan <-chan string // reload signals come over this channelc
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
	var srvName string
	for {
		select {
		case <-ctx.Done():
			srvMngr.ShutdownServices()
			return
		case srvName = <-srvMngr.rldChan:
		}
		if srvName == config.RPCConnsJSON {
			go srvMngr.connMgr.Reload()
		} else {
			go srvMngr.reloadService(srvName, ctx, shtDwn)

		}
		// handle RPC server
	}
}

func (srvMngr *ServiceManager) reloadService(srvName string, ctx *context.Context, shtDwn context.CancelFunc) (err error) {
	srv := srvMngr.GetService(srvName)
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

// ArgsServiceID are passed to Start/Stop/Status RPC methods
type ArgsServiceID struct {
	ServiceID string
	APIOpts   map[string]any
}

// V1StartService starts a service with ID
func (srvMngr *ServiceManager) V1StartService(ctx *context.Context, args *ArgsServiceID, reply *string) (err error) {
	err = toggleService(args.ServiceID, true, srvMngr)
	if err != nil {
		return
	}
	*reply = utils.OK
	return
}

// V1StopService shuts-down a service with ID
func (srvMngr *ServiceManager) V1StopService(ctx *context.Context, args *ArgsServiceID, reply *string) (err error) {
	err = toggleService(args.ServiceID, false, srvMngr)
	if err != nil {
		return
	}
	*reply = utils.OK
	return
}

// V1ServiceStatus  returns the service status
func (srvMngr *ServiceManager) V1ServiceStatus(ctx *context.Context, args *ArgsServiceID, reply *string) error {
	srvMngr.RLock()
	defer srvMngr.RUnlock()

	srv := srvMngr.GetService(args.ServiceID)
	if srv == nil {
		return utils.ErrUnsupportedServiceID
	}
	running := srv.IsRunning()

	if running {
		*reply = utils.RunningCaps
	} else {
		*reply = utils.StoppedCaps
	}
	return nil
}

// GetConfig returns the Configuration
func (srvMngr *ServiceManager) GetConfig() *config.CGRConfig {
	srvMngr.RLock()
	defer srvMngr.RUnlock()
	return srvMngr.cfg
}

func toggleService(serviceID string, status bool, srvMngr *ServiceManager) (err error) {
	srvMngr.Lock()
	defer srvMngr.Unlock()
	switch serviceID {
	case utils.AccountS:
		srvMngr.cfg.AccountSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.ActionS:
		srvMngr.cfg.ActionSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.AdminS:
		srvMngr.cfg.AdminSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.AnalyzerS:
		srvMngr.cfg.AnalyzerSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.AttributeS:
		srvMngr.cfg.AttributeSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.CDRServer:
		srvMngr.cfg.CdrsCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.ChargerS:
		srvMngr.cfg.ChargerSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.DispatcherS:
		srvMngr.cfg.DispatcherSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.EEs:
		srvMngr.cfg.EEsCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.EFs:
		srvMngr.cfg.EFsCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.ERs:
		srvMngr.cfg.ERsCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
		// case utils.LoaderS:
		// 	srvMngr.cfg.LoaderCfg()[0].Enabled = status
		// 	srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.RateS:
		srvMngr.cfg.RateSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.TrendS:
		srvMngr.cfg.TrendSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.RankingS:
		srvMngr.cfg.RankingSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.ResourceS:
		srvMngr.cfg.ResourceSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.RouteS:
		srvMngr.cfg.RouteSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.SessionS:
		srvMngr.cfg.SessionSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.StatS:
		srvMngr.cfg.StatSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.ThresholdS:
		srvMngr.cfg.ThresholdSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.TPeS:
		srvMngr.cfg.TpeSCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.AsteriskAgent:
		srvMngr.cfg.AsteriskAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.DiameterAgent:
		srvMngr.cfg.DiameterAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.DNSAgent:
		srvMngr.cfg.DNSAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.FreeSWITCHAgent:
		srvMngr.cfg.FsAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.KamailioAgent:
		srvMngr.cfg.KamAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.RadiusAgent:
		srvMngr.cfg.RadiusAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	case utils.SIPAgent:
		srvMngr.cfg.SIPAgentCfg().Enabled = status
		srvMngr.cfg.GetReloadChan() <- serviceID
	default:
		err = utils.ErrUnsupportedServiceID
	}
	return
}
