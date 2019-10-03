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
	internalChan := srvMngr.GetService(subsystem).GetIntenternalChan()
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
	/*
		if srvMngr.GetCacheS() == nil {
			go srvMngr.startService(utils.CacheS)
			go srvMngr.startService(utils.GuardianS)
		}
	*/

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
	} /*
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
		if srvMngr.GetConfig().DiameterAgentCfg().Enabled {
			go srvMngr.startService(utils.DiameterAgent)
		}
		if len(srvMngr.GetConfig().HttpAgentCfg()) != 0 {
			go srvMngr.startService(utils.HTTPAgent)
		}
	*/
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
			} /*
				case <-srvMngr.GetConfig().GetReloadChan(config.SupplierSJson):
					if err = srvMngr.reloadService(utils.SupplierS); err != nil {
						return
					}
				case <-srvMngr.GetConfig().GetReloadChan(config.SCHEDULER_JSN):
					if err = srvMngr.reloadService(utils.SchedulerS); err != nil {
						return
					}
				case <-srvMngr.GetConfig().GetReloadChan(config.CDRS_JSN):
					if err = srvMngr.reloadService(utils.CDRServer); err != nil {
						return
					}
				case <-srvMngr.GetConfig().GetReloadChan(config.RALS_JSN):
					if err = srvMngr.reloadService(utils.RALService); err != nil {
						return
					}
				case <-srvMngr.GetConfig().GetReloadChan(config.Apier):
					if err = srvMngr.reloadService(utils.ApierV1); err != nil {
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
			*/
		}
		// handle RPC server
	}
}

func (srvMngr *ServiceManager) reloadService(srviceName string) (err error) {
	srv := srvMngr.GetService(srviceName)

	if srv.ShouldRun() {
		if srv.IsRunning() {
			if err = srv.Reload(); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to reload <%s>", utils.ServiceManager, srv.ServiceName()))
				srvMngr.engineShutdown <- true
				return // stop if we encounter an error
			}
		} else {
			if err = srv.Start(); err != nil {
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
	srv := srvMngr.GetService(srviceName)
	if err := srv.Start(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed to start %s because: %s", utils.ServiceManager, srviceName, err))
		srvMngr.engineShutdown <- true
	}
}

// GetService returns the named service
func (srvMngr ServiceManager) GetService(subsystem string) (srv Service) {
	var has bool
	srvMngr.RLock()
	srv, has = srvMngr.subsystems[subsystem]
	srvMngr.RUnlock()
	if !has { // this should not happen (check the added services)
		panic(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
			utils.ServiceManager, subsystem)) // because this is not dinamic this should not happen
	}
	return
}

// SetCacheS sets the cacheS
func (srvMngr *ServiceManager) SetCacheS(chS *engine.CacheS) {
	srvMngr.Lock()
	srvMngr.cacheS = chS
	srvMngr.Unlock()
}

// Service interface that describes what functions should a service implement
type Service interface {
	// Start should handle the sercive start
	Start() error
	// Reload handles the change of config
	Reload() error
	// Shutdown stops the service
	Shutdown() error
	// GetIntenternalChan returns the internal connection chanel
	GetIntenternalChan() chan rpcclient.RpcClientConnection
	// IsRunning returns if the service is running
	IsRunning() bool
	// ShouldRun returns if the service should be running
	ShouldRun() bool
	// ServiceName returns the service name
	ServiceName() string
}
