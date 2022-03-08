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
func NewServiceManager(shdWg *sync.WaitGroup, connMgr *engine.ConnManager, rldChan <-chan string) *ServiceManager {
	return &ServiceManager{
		subsystems: make(map[string]Service),
		shdWg:      shdWg,
		connMgr:    connMgr,
		rldChan:    rldChan,
	}
}

// ServiceManager handles service management ran by the engine
type ServiceManager struct {
	sync.RWMutex // lock access to any shared data
	subsystems   map[string]Service

	shdWg   *sync.WaitGroup
	connMgr *engine.ConnManager
	rldChan <-chan string
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
