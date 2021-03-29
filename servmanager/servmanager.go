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
func NewServiceManager(cfg *config.CGRConfig, shdChan *utils.SyncedChan, shdWg *sync.WaitGroup, connMgr *engine.ConnManager) *ServiceManager {
	sm := &ServiceManager{
		cfg:        cfg,
		subsystems: make(map[string]Service),
		shdChan:    shdChan,
		shdWg:      shdWg,
		connMgr:    connMgr,
	}
	return sm
}

// ServiceManager handles service management ran by the engine
type ServiceManager struct {
	sync.RWMutex // lock access to any shared data
	cfg          *config.CGRConfig
	subsystems   map[string]Service

	shdChan *utils.SyncedChan
	shdWg   *sync.WaitGroup
	connMgr *engine.ConnManager
}

// Call .
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
		srvMngr.cfg.ActionSCfg().Enabled = true
		srvMngr.Unlock()
		srvMngr.cfg.GetReloadChan(config.ActionSJson) <- struct{}{}
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
		srvMngr.cfg.ActionSCfg().Enabled = false
		srvMngr.Unlock()
		srvMngr.cfg.GetReloadChan(config.ActionSJson) <- struct{}{}
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
			srvMngr.shdWg.Add(1)
			go func(srv Service) {
				if err := srv.Start(); err != nil &&
					err != utils.ErrServiceAlreadyRunning { // in case the service was started in another gorutine
					utils.Logger.Err(fmt.Sprintf("<%s> failed to start %s because: %s", utils.ServiceManager, srv.ServiceName(), err))
					srvMngr.shdChan.CloseOnce()
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
		if _, has := srvMngr.subsystems[srv.ServiceName()]; !has { // do not rewrite the service
			srvMngr.subsystems[srv.ServiceName()] = srv
		}
	}
	srvMngr.Unlock()
}

func (srvMngr *ServiceManager) handleReload() {
	for {
		select {
		case <-srvMngr.shdChan.Done():
			srvMngr.ShutdownServices()
			return
		case <-srvMngr.GetConfig().GetReloadChan(config.ATTRIBUTE_JSN):
			go srvMngr.reloadService(utils.AttributeS)
		case <-srvMngr.GetConfig().GetReloadChan(config.ChargerSCfgJson):
			go srvMngr.reloadService(utils.ChargerS)
		case <-srvMngr.GetConfig().GetReloadChan(config.THRESHOLDS_JSON):
			go srvMngr.reloadService(utils.ThresholdS)
		case <-srvMngr.GetConfig().GetReloadChan(config.STATS_JSON):
			go srvMngr.reloadService(utils.StatS)
		case <-srvMngr.GetConfig().GetReloadChan(config.RESOURCES_JSON):
			go srvMngr.reloadService(utils.ResourceS)
		case <-srvMngr.GetConfig().GetReloadChan(config.RouteSJson):
			go srvMngr.reloadService(utils.RouteS)
		case <-srvMngr.GetConfig().GetReloadChan(config.RALS_JSN):
			go srvMngr.reloadService(utils.RALService)
		case <-srvMngr.GetConfig().GetReloadChan(config.ApierS):
			go func() {
				srvMngr.reloadService(utils.APIerSv1)
				srvMngr.reloadService(utils.APIerSv2)
			}()
		case <-srvMngr.GetConfig().GetReloadChan(config.CDRS_JSN):
			go srvMngr.reloadService(utils.CDRServer)
		case <-srvMngr.GetConfig().GetReloadChan(config.SessionSJson):
			go srvMngr.reloadService(utils.SessionS)
		case <-srvMngr.GetConfig().GetReloadChan(config.ERsJson):
			go srvMngr.reloadService(utils.ERs)
		case <-srvMngr.GetConfig().GetReloadChan(config.DNSAgentJson):
			go srvMngr.reloadService(utils.DNSAgent)
		case <-srvMngr.GetConfig().GetReloadChan(config.FreeSWITCHAgentJSN):
			go srvMngr.reloadService(utils.FreeSWITCHAgent)
		case <-srvMngr.GetConfig().GetReloadChan(config.KamailioAgentJSN):
			go srvMngr.reloadService(utils.KamailioAgent)
		case <-srvMngr.GetConfig().GetReloadChan(config.AsteriskAgentJSN):
			go srvMngr.reloadService(utils.AsteriskAgent)
		case <-srvMngr.GetConfig().GetReloadChan(config.RA_JSN):
			go srvMngr.reloadService(utils.RadiusAgent)
		case <-srvMngr.GetConfig().GetReloadChan(config.DA_JSN):
			go srvMngr.reloadService(utils.DiameterAgent)
		case <-srvMngr.GetConfig().GetReloadChan(config.HttpAgentJson):
			go srvMngr.reloadService(utils.HTTPAgent)
		case <-srvMngr.GetConfig().GetReloadChan(config.LoaderJson):
			go srvMngr.reloadService(utils.LoaderS)
		case <-srvMngr.GetConfig().GetReloadChan(config.AnalyzerCfgJson):
			go srvMngr.reloadService(utils.AnalyzerS)
		case <-srvMngr.GetConfig().GetReloadChan(config.DispatcherSJson):
			go srvMngr.reloadService(utils.DispatcherS)
		case <-srvMngr.GetConfig().GetReloadChan(config.DATADB_JSN):
			go srvMngr.reloadService(utils.DataDB)
		case <-srvMngr.GetConfig().GetReloadChan(config.STORDB_JSN):
			go srvMngr.reloadService(utils.StorDB)
		case <-srvMngr.GetConfig().GetReloadChan(config.EEsJson):
			go srvMngr.reloadService(utils.EventExporterS)
		case <-srvMngr.GetConfig().GetReloadChan(config.RateSJson):
			go srvMngr.reloadService(utils.RateS)
		case <-srvMngr.GetConfig().GetReloadChan(config.RPCConnsJsonName):
			go srvMngr.connMgr.Reload()
		case <-srvMngr.GetConfig().GetReloadChan(config.SIPAgentJson):
			go srvMngr.reloadService(utils.SIPAgent)
		case <-srvMngr.GetConfig().GetReloadChan(config.RegistrarCJson):
			go srvMngr.reloadService(utils.RegistrarC)
		case <-srvMngr.GetConfig().GetReloadChan(config.HTTP_JSN):
			go srvMngr.reloadService(utils.GlobalVarS)
		case <-srvMngr.GetConfig().GetReloadChan(config.AccountSCfgJson):
			go srvMngr.reloadService(utils.AccountS)
		case <-srvMngr.GetConfig().GetReloadChan(config.ActionSJson):
			go srvMngr.reloadService(utils.ActionS)
		case <-srvMngr.GetConfig().GetReloadChan(config.CoreSCfgJson):
			go srvMngr.reloadService(utils.CoreS)
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
				srvMngr.shdChan.CloseOnce()
				return // stop if we encounter an error
			}
		} else {
			srvMngr.shdWg.Add(1)
			if err = srv.Start(); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> failed to start <%s> err <%s>", utils.ServiceManager, srv.ServiceName(), err))
				srvMngr.shdChan.CloseOnce()
				return // stop if we encounter an error
			}
		}
	} else if srv.IsRunning() {
		if err = srv.Shutdown(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed to stop service <%s> err <%s>", utils.ServiceManager, srv.ServiceName(), err))
			srvMngr.shdChan.CloseOnce()
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
