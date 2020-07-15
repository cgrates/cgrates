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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewServiceManager returns a service manager
func NewServiceManager(cfg *config.CGRConfig, engineShutdown chan bool) *ServiceManager {
	sm := &ServiceManager{
		cfg:            cfg,
		engineShutdown: engineShutdown,
		subsystems:     make(map[string]Service),
	}
	return sm
}

// ServiceManager handles service management ran by the engine
type ServiceManager struct {
	sync.RWMutex   // lock access to any shared data
	cfg            *config.CGRConfig
	engineShutdown chan bool
	subsystems     map[string]Service
}

func (srvMngr *ServiceManager) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// get method
	method := reflect.ValueOf(srvMngr).MethodByName(parts[0][len(parts[0])-2:] + parts[1]) // Inherit the version in the method
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// construct the params
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}

// ArgStartService are passed to Start/StopService/Status RPC methods
type ArgStartService struct {
	ServiceID string
}

// V1StartService starts a service with ID
func (srvMngr *ServiceManager) V1StartService(args ArgStartService, reply *string) (err error) {
	switch args.ServiceID {
	case utils.MetaScheduler:
		// stop the service using the config
		srvMngr.Lock()
		srvMngr.cfg.SchedulerCfg().Enabled = true
		srvMngr.Unlock()
		srvMngr.cfg.GetReloadChan(config.SCHEDULER_JSN) <- struct{}{}
	default:
		err = errors.New(utils.UnsupportedServiceIDCaps)
	}
	if err != nil {
		return err
	}
	*reply = utils.OK
	return
}

// V1StopService shuts-down a service with ID
func (srvMngr *ServiceManager) V1StopService(args ArgStartService, reply *string) (err error) {
	switch args.ServiceID {
	case utils.MetaScheduler:
		// stop the service using the config
		srvMngr.Lock()
		srvMngr.cfg.SchedulerCfg().Enabled = false
		srvMngr.Unlock()
		srvMngr.cfg.GetReloadChan(config.SCHEDULER_JSN) <- struct{}{}
	default:
		err = errors.New(utils.UnsupportedServiceIDCaps)
	}
	if err != nil {
		return err
	}
	*reply = utils.OK
	return
}

// V1ServiceStatus  returns the service status
func (srvMngr *ServiceManager) V1ServiceStatus(args ArgStartService, reply *string) error {
	srvMngr.RLock()
	defer srvMngr.RUnlock()
	var running bool
	switch args.ServiceID {
	case utils.MetaScheduler:
		running = srvMngr.subsystems[utils.SchedulerS].IsRunning()
	default:
		return errors.New(utils.UnsupportedServiceIDCaps)
	}
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

// StartServices starts all enabled services
func (srvMngr *ServiceManager) StartServices() (err error) {
	go srvMngr.handleReload()
	for _, service := range srvMngr.subsystems {
		if service.ShouldRun() && !service.IsRunning() {
			go func(srv Service) {
				if err := srv.Start(); err != nil {
					if err == utils.ErrServiceAlreadyRunning { // in case the service was started in another gorutine
						return
					}
					utils.Logger.Err(fmt.Sprintf("<%s> failed to start %s because: %s", utils.ServiceManager, srv.ServiceName(), err))
					srvMngr.engineShutdown <- true
				}
			}(service)
		}
	}
	// startServer()
	return
}

// AddServices adds given services
func (srvMngr *ServiceManager) AddServices(services ...Service) {
	srvMngr.Lock()
	for _, srv := range services {
		if _, has := srvMngr.subsystems[srv.ServiceName()]; has { // do not rewrite the service
			continue
		}
		srvMngr.subsystems[srv.ServiceName()] = srv
	}
	srvMngr.Unlock()
}

func (srvMngr *ServiceManager) handleReload() {
	var err error
	for {
		select {
		case ext := <-srvMngr.engineShutdown:
			srvMngr.engineShutdown <- ext
			for srviceName, srv := range srvMngr.subsystems { // gracefully stop all running subsystems
				if !srv.IsRunning() {
					continue
				}
				if err := srv.Shutdown(); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Failed to shutdown subsystem <%s> because: %s",
						utils.ServiceManager, srviceName, err))
				}
			}
			return
		case <-srvMngr.GetConfig().GetReloadChan(config.ATTRIBUTE_JSN):
			if err = srvMngr.reloadService(utils.AttributeS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.ChargerSCfgJson):
			if err = srvMngr.reloadService(utils.ChargerS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.THRESHOLDS_JSON):
			if err = srvMngr.reloadService(utils.ThresholdS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.STATS_JSON):
			if err = srvMngr.reloadService(utils.StatS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.RESOURCES_JSON):
			if err = srvMngr.reloadService(utils.ResourceS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.RouteSJson):
			if err = srvMngr.reloadService(utils.RouteS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.SCHEDULER_JSN):
			if err = srvMngr.reloadService(utils.SchedulerS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.RALS_JSN):
			if err = srvMngr.reloadService(utils.RALService); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.ApierS):
			if err = srvMngr.reloadService(utils.APIerSv1); err != nil {
				return
			}
			if err = srvMngr.reloadService(utils.APIerSv2); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.CDRS_JSN):
			if err = srvMngr.reloadService(utils.CDRServer); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.SessionSJson):
			if err = srvMngr.reloadService(utils.SessionS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.ERsJson):
			if err = srvMngr.reloadService(utils.ERs); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.DNSAgentJson):
			if err = srvMngr.reloadService(utils.DNSAgent); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.FreeSWITCHAgentJSN):
			if err = srvMngr.reloadService(utils.FreeSWITCHAgent); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.KamailioAgentJSN):
			if err = srvMngr.reloadService(utils.KamailioAgent); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.AsteriskAgentJSN):
			if err = srvMngr.reloadService(utils.AsteriskAgent); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.RA_JSN):
			if err = srvMngr.reloadService(utils.RadiusAgent); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.DA_JSN):
			if err = srvMngr.reloadService(utils.DiameterAgent); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.HttpAgentJson):
			if err = srvMngr.reloadService(utils.HTTPAgent); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.LoaderJson):
			if err = srvMngr.reloadService(utils.LoaderS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.AnalyzerCfgJson):
			if err = srvMngr.reloadService(utils.AnalyzerS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.DispatcherSJson):
			if err = srvMngr.reloadService(utils.DispatcherS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.DATADB_JSN):
			if err = srvMngr.reloadService(utils.DataDB); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.STORDB_JSN):
			if err = srvMngr.reloadService(utils.StorDB); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.EEsJson):
			if err = srvMngr.reloadService(utils.EventExporterS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.RateSJson):
			if err = srvMngr.reloadService(utils.RateS); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.RPCConnsJsonName):
			engine.Cache.Clear([]string{utils.CacheRPCConnections})
		case <-srvMngr.GetConfig().GetReloadChan(config.SIPAgentJson):
			if err = srvMngr.reloadService(utils.SIPAgent); err != nil {
				return
			}
		}
		// handle RPC server
	}
}

func (srvMngr *ServiceManager) reloadService(srviceName string) (err error) {
	srv := srvMngr.GetService(srviceName)
	if srv.ShouldRun() {
		if srv.IsRunning() {
			if err = srv.Reload(); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> failed to reload <%s> err <%s>", utils.ServiceManager, srv.ServiceName(), err))
				srvMngr.engineShutdown <- true
				return // stop if we encounter an error
			}
		} else {
			if err = srv.Start(); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> failed to start <%s> err <%s>", utils.ServiceManager, srv.ServiceName(), err))
				srvMngr.engineShutdown <- true
				return // stop if we encounter an error
			}
		}
	} else if srv.IsRunning() {
		if err = srv.Shutdown(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed to stop service <%s> err <%s>", utils.ServiceManager, srv.ServiceName(), err))
			srvMngr.engineShutdown <- true
			return // stop if we encounter an error
		}
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

// Service interface that describes what functions should a service implement
type Service interface {
	// Start should handle the sercive start
	Start() error
	// Reload handles the change of config
	Reload() error
	// Shutdown stops the service
	Shutdown() error
	// IsRunning returns if the service is running
	IsRunning() bool
	// ShouldRun returns if the service should be running
	ShouldRun() bool
	// ServiceName returns the service name
	ServiceName() string
}
