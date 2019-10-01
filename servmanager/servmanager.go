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
func NewServiceManager(cfg *config.CGRConfig, dm *engine.DataManager,
	cdrStorage engine.CdrStorage,
	loadStorage engine.LoadStorage, filterSChan chan *engine.FilterS,
	server *utils.Server, dispatcherSChan chan rpcclient.RpcClientConnection,
	engineShutdown chan bool) *ServiceManager {
	sm := &ServiceManager{
		cfg:            cfg,
		dm:             dm,
		engineShutdown: engineShutdown,

		cdrStorage:      cdrStorage,
		loadStorage:     loadStorage,
		filterS:         filterSChan,
		server:          server,
		subsystems:      make(map[string]Service),
		dispatcherSChan: dispatcherSChan,
	}
	return sm
}

// ServiceManager handles service management ran by the engine
type ServiceManager struct {
	sync.RWMutex   // lock access to any shared data
	cfg            *config.CGRConfig
	dm             *engine.DataManager
	engineShutdown chan bool
	cacheS         *engine.CacheS
	cdrStorage     engine.CdrStorage
	loadStorage    engine.LoadStorage
	filterS        chan *engine.FilterS
	server         *utils.Server
	subsystems     map[string]Service

	dispatcherSChan chan rpcclient.RpcClientConnection
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

// ArgShutdownService are passed to Start/StopService/Status RPC methods
type ArgStartService struct {
	ServiceID string
}

// ShutdownService shuts-down a service with ID
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

// ShutdownService shuts-down a service with ID
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

// ShutdownService shuts-down a service with ID
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

// GetDM returns the DataManager
func (srvMngr *ServiceManager) GetDM() *engine.DataManager {
	srvMngr.RLock()
	defer srvMngr.RUnlock()
	return srvMngr.dm
}

// GetCDRStorage returns the CdrStorage
func (srvMngr *ServiceManager) GetCDRStorage() engine.CdrStorage {
	srvMngr.RLock()
	defer srvMngr.RUnlock()
	return srvMngr.cdrStorage
}

// GetLoadStorage returns the LoadStorage
func (srvMngr *ServiceManager) GetLoadStorage() engine.LoadStorage {
	srvMngr.RLock()
	defer srvMngr.RUnlock()
	return srvMngr.loadStorage
}

// GetConfig returns the Configuration
func (srvMngr *ServiceManager) GetConfig() *config.CGRConfig {
	srvMngr.RLock()
	defer srvMngr.RUnlock()
	return srvMngr.cfg
}

// GetCacheS returns the CacheS
func (srvMngr *ServiceManager) GetCacheS() *engine.CacheS {
	srvMngr.RLock()
	defer srvMngr.RUnlock()
	return srvMngr.cacheS
}

// GetFilterS returns the FilterS
func (srvMngr *ServiceManager) GetFilterS() (fS *engine.FilterS) {
	srvMngr.RLock()
	defer srvMngr.RUnlock()
	fS = <-srvMngr.filterS
	srvMngr.filterS <- fS
	return
}

// GetServer returns the Server
func (srvMngr *ServiceManager) GetServer() *utils.Server {
	srvMngr.RLock()
	defer srvMngr.RUnlock()
	return srvMngr.server
}

// GetExitChan returns the exit chanel
func (srvMngr *ServiceManager) GetExitChan() chan bool {
	return srvMngr.engineShutdown
}

// NewConnection creates a rpcClient to the specified subsystem
func (srvMngr *ServiceManager) NewConnection(subsystem string, conns []*config.RemoteHost) (rpcclient.RpcClientConnection, error) {
	if len(conns) == 0 {
		return nil, nil
	}
	service, has := srvMngr.GetService(subsystem)
	if !has { // used to not cause panics because of services that are not already migrated
		return nil, errors.New(utils.UnsupportedServiceIDCaps)
	}
	internalChan := service.GetIntenternalChan()
	if srvMngr.GetConfig().DispatcherSCfg().Enabled {
		internalChan = srvMngr.dispatcherSChan
	}
	return engine.NewRPCPool(rpcclient.POOL_FIRST,
		srvMngr.GetConfig().TlsCfg().ClientKey,
		srvMngr.GetConfig().TlsCfg().ClientCerificate, srvMngr.GetConfig().TlsCfg().CaCertificate,
		srvMngr.GetConfig().GeneralCfg().ConnectAttempts, srvMngr.GetConfig().GeneralCfg().Reconnects,
		srvMngr.GetConfig().GeneralCfg().ConnectTimeout, srvMngr.GetConfig().GeneralCfg().ReplyTimeout,
		conns, internalChan, false)
}

// StartServices starts all enabled services
func (srvMngr *ServiceManager) StartServices() (err error) {
	// start the cacheS
	if srvMngr.GetCacheS() == nil {
		chS, has := srvMngr.GetService(utils.CacheS)
		if !has {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
				utils.ServiceManager, utils.CacheS))
			return
		}
		chS.Start(srvMngr, true)
	}

	go srvMngr.handleReload()
	if srvMngr.GetConfig().AttributeSCfg().Enabled {
		go srvMngr.startService(utils.AttributeS)
	}
	if srvMngr.GetConfig().ChargerSCfg().Enabled {
		go srvMngr.startService(utils.ChargerS)
	}
	if srvMngr.GetConfig().ThresholdSCfg().Enabled {
		go srvMngr.startService(utils.ThresholdS)
	}
	if srvMngr.GetConfig().StatSCfg().Enabled {
		go srvMngr.startService(utils.StatS)
	}
	if srvMngr.GetConfig().ResourceSCfg().Enabled {
		go srvMngr.startService(utils.ResourceS)
	}
	if srvMngr.GetConfig().SupplierSCfg().Enabled {
		go srvMngr.startService(utils.SupplierS)
	}
	if srvMngr.GetConfig().SchedulerCfg().Enabled {
		go srvMngr.startService(utils.SchedulerS)
	}
	if srvMngr.GetConfig().CdrsCfg().Enabled {
		go srvMngr.startService(utils.CDRServer)
	}
	if srvMngr.GetConfig().RalsCfg().Enabled {
		go srvMngr.startService(utils.RALService)
	}
	if srvMngr.GetConfig().SessionSCfg().Enabled {
		go srvMngr.startService(utils.SessionS)
	}
	if srvMngr.GetConfig().ERsCfg().Enabled {
		go srvMngr.startService(utils.ERs)
	}
	if srvMngr.GetConfig().DNSAgentCfg().Enabled {
		go srvMngr.startService(utils.DNSAgent)
	}
	if srvMngr.GetConfig().FsAgentCfg().Enabled {
		go srvMngr.startService(utils.FreeSWITCHAgent)
	}
	if srvMngr.GetConfig().KamAgentCfg().Enabled {
		go srvMngr.startService(utils.KamailioAgent)
	}
	if srvMngr.GetConfig().AsteriskAgentCfg().Enabled {
		go srvMngr.startService(utils.AsteriskAgent)
	}
	if srvMngr.GetConfig().RadiusAgentCfg().Enabled {
		go srvMngr.startService(utils.RadiusAgent)
	}
	// startServer()
	return
}

// AddService adds given services
func (srvMngr *ServiceManager) AddService(services ...Service) {
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
			for srviceName, srv := range srvMngr.subsystems { // gracefully stop all running subsystems
				if !srv.IsRunning() {
					continue
				}
				if err := srv.Shutdown(); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Failed to shutdown subsystem <%s> because: %s",
						utils.ServiceManager, srviceName, err))
				}
			}
			srvMngr.engineShutdown <- ext
			return
		case <-srvMngr.GetConfig().GetReloadChan(config.ATTRIBUTE_JSN):
			if err = srvMngr.reloadService(utils.AttributeS, srvMngr.GetConfig().AttributeSCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.ChargerSCfgJson):
			if err = srvMngr.reloadService(utils.ChargerS, srvMngr.GetConfig().ChargerSCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.THRESHOLDS_JSON):
			if err = srvMngr.reloadService(utils.ThresholdS, srvMngr.GetConfig().ThresholdSCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.STATS_JSON):
			if err = srvMngr.reloadService(utils.StatS, srvMngr.GetConfig().StatSCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.RESOURCES_JSON):
			if err = srvMngr.reloadService(utils.ResourceS, srvMngr.GetConfig().ResourceSCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.SupplierSJson):
			if err = srvMngr.reloadService(utils.SupplierS, srvMngr.GetConfig().SupplierSCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.SCHEDULER_JSN):
			if err = srvMngr.reloadService(utils.SchedulerS, srvMngr.GetConfig().SchedulerCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.CDRS_JSN):
			if err = srvMngr.reloadService(utils.CDRServer, srvMngr.GetConfig().CdrsCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.RALS_JSN):
			if err = srvMngr.reloadService(utils.RALService, srvMngr.GetConfig().RalsCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.SessionSJson):
			if err = srvMngr.reloadService(utils.SessionS, srvMngr.GetConfig().SessionSCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.ERsJson):
			if err = srvMngr.reloadService(utils.ERs, srvMngr.GetConfig().ERsCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.DNSAgentJson):
			if err = srvMngr.reloadService(utils.DNSAgent, srvMngr.GetConfig().DNSAgentCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.FreeSWITCHAgentJSN):
			if err = srvMngr.reloadService(utils.FreeSWITCHAgent, srvMngr.GetConfig().FsAgentCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.KamailioAgentJSN):
			if err = srvMngr.reloadService(utils.KamailioAgent, srvMngr.GetConfig().KamAgentCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.AsteriskAgentJSN):
			if err = srvMngr.reloadService(utils.AsteriskAgent, srvMngr.GetConfig().AsteriskAgentCfg().Enabled); err != nil {
				return
			}
		case <-srvMngr.GetConfig().GetReloadChan(config.RA_JSN):
			if err = srvMngr.reloadService(utils.RadiusAgent, srvMngr.GetConfig().RadiusAgentCfg().Enabled); err != nil {
				return
			}
		}
		// handle RPC server
	}
}

func (srvMngr *ServiceManager) reloadService(srviceName string, shouldRun bool) (err error) {
	srv, has := srvMngr.GetService(srviceName)
	if !has { // this should not happen (check the added services)
		utils.Logger.Err(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
			utils.ServiceManager, srviceName))
		srvMngr.engineShutdown <- true
		return
	}
	if shouldRun {
		if srv.IsRunning() {
			if err = srv.Reload(srvMngr); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to reload <%s>", utils.ServiceManager, srv.ServiceName()))
				srvMngr.engineShutdown <- true
				return // stop if we encounter an error
			}
		} else {
			if err = srv.Start(srvMngr, true); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, srv.ServiceName()))
				srvMngr.engineShutdown <- true
				return // stop if we encounter an error
			}
		}
	} else if srv.IsRunning() {
		if err = srv.Shutdown(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed to stop service <%s>", utils.ServiceManager, srv.ServiceName()))
			srvMngr.engineShutdown <- true
			return // stop if we encounter an error
		}
	}
	return
}

func (srvMngr *ServiceManager) startService(srviceName string) {
	srv, has := srvMngr.GetService(srviceName)
	if !has { // this should not happen (check the added services)
		utils.Logger.Err(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
			utils.ServiceManager, srviceName))
		srvMngr.engineShutdown <- true
		return
	}
	if err := srv.Start(srvMngr, true); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed to start %s because: %s", utils.ServiceManager, srviceName, err))
		srvMngr.engineShutdown <- true
	}
}

// GetService returns the named service
func (srvMngr ServiceManager) GetService(subsystem string) (srv Service, has bool) {
	srvMngr.RLock()
	srv, has = srvMngr.subsystems[subsystem]
	srvMngr.RUnlock()
	return
}

// SetCacheS sets the cacheS
func (srvMngr *ServiceManager) SetCacheS(chS *engine.CacheS) {
	srvMngr.Lock()
	srvMngr.cacheS = chS
	srvMngr.Unlock()
}

// ServiceProvider should implement this to provide information for service
type ServiceProvider interface {
	// GetDM returns the DataManager
	GetDM() *engine.DataManager
	// GetCDRStorage returns the CdrStorage
	GetCDRStorage() engine.CdrStorage
	// GetLoadStorage returns the LoadStorage
	GetLoadStorage() engine.LoadStorage
	// GetConfig returns the Configuration
	GetConfig() *config.CGRConfig
	// GetCacheS returns the CacheS
	GetCacheS() *engine.CacheS
	// GetFilterS returns the FilterS
	GetFilterS() *engine.FilterS
	// GetServer returns the Server
	GetServer() *utils.Server
	// GetExitChan returns the exit chanel
	GetExitChan() chan bool
	// NewConnection creates a rpcClient to the specified subsystem
	NewConnection(subsystem string, cfg []*config.RemoteHost) (rpcclient.RpcClientConnection, error)
	// GetService returns the named service
	GetService(subsystem string) (Service, bool)
	// AddService adds the given serices
	AddService(services ...Service)
	// SetCacheS sets the cacheS
	// Called when starting Cache Service
	SetCacheS(chS *engine.CacheS)
}

// Service interface that describes what functions should a service implement
type Service interface {
	// Start should handle the sercive start
	Start(sp ServiceProvider, waitCache bool) error
	// Reload handles the change of config
	Reload(sp ServiceProvider) error
	// GetRPCInterface returns the interface to register for server
	GetRPCInterface() interface{}
	// GetIntenternalChan returns the internal connection chanel
	GetIntenternalChan() chan rpcclient.RpcClientConnection
	// IsRunning returns if the service is running
	IsRunning() bool
	// ServiceName returns the service name
	ServiceName() string
	// Shutdown stops the service
	Shutdown() error
}
