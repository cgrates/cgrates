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
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewServiceManager returns a service manager
func NewServiceManager(cfg *config.CGRConfig, dm *engine.DataManager,
	cacheS *engine.CacheS, cdrStorage engine.CdrStorage,
	loadStorage engine.LoadStorage, filterSChan chan *engine.FilterS,
	server *utils.Server, dispatcherSChan chan rpcclient.RpcClientConnection,
	engineShutdown chan bool) *ServiceManager {
	sm := &ServiceManager{
		cfg:            cfg,
		dm:             dm,
		engineShutdown: engineShutdown,
		cacheS:         cacheS,

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
	sched          *scheduler.Scheduler

	cdrStorage  engine.CdrStorage
	loadStorage engine.LoadStorage
	filterS     chan *engine.FilterS
	server      *utils.Server
	subsystems  map[string]Service

	dispatcherSChan chan rpcclient.RpcClientConnection
}

func (srvMngr *ServiceManager) StartScheduler(waitCache bool) error {
	srvMngr.RLock()
	schedRunning := srvMngr.sched != nil
	srvMngr.RUnlock()
	if schedRunning {
		return utils.NewCGRError(utils.ServiceManager,
			utils.CapitalizedMessage(utils.ServiceAlreadyRunning),
			utils.ServiceAlreadyRunning,
			"the scheduler is already running")
	}
	if waitCache { // Wait for cache to load data before starting
		<-srvMngr.cacheS.GetPrecacheChannel(utils.CacheActionPlans) // wait for ActionPlans to be cached
	}
	utils.Logger.Info("<ServiceManager> Starting CGRateS Scheduler.")
	sched := scheduler.NewScheduler(srvMngr.dm)
	srvMngr.Lock()
	srvMngr.sched = sched
	srvMngr.Unlock()
	go func() {
		sched.Loop()
		srvMngr.Lock()
		srvMngr.sched = nil // if we are after loop, the service is down
		srvMngr.Unlock()
		if srvMngr.cfg.SchedulerCfg().Enabled {
			srvMngr.engineShutdown <- true // shutdown engine since this service should be running
		}
	}()
	return nil
}

func (srvMngr *ServiceManager) StopScheduler() error {
	var sched *scheduler.Scheduler
	srvMngr.Lock()
	if srvMngr.sched != nil {
		sched = srvMngr.sched
		srvMngr.sched = nil // optimize the lock and release here
	}
	srvMngr.Unlock()
	if sched == nil {
		return utils.NewCGRError(utils.ServiceManager,
			utils.CapitalizedMessage(utils.ServiceNotRunning),
			utils.ServiceNotRunning,
			"the scheduler is not running")
	}
	utils.Logger.Info("<ServiceManager> Stoping CGRateS Scheduler.")
	srvMngr.cfg.SchedulerCfg().Enabled = false
	sched.Shutdown()
	return nil
}

func (srvMngr *ServiceManager) GetScheduler() *scheduler.Scheduler {
	srvMngr.RLock()
	defer srvMngr.RUnlock()
	return srvMngr.sched
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
		err = srvMngr.StartScheduler(false)
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
		err = srvMngr.StopScheduler()
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
		running = srvMngr.sched != nil
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

// GetConnection creates a rpcClient to the specified subsystem
func (srvMngr *ServiceManager) GetConnection(subsystem string, conns []*config.RemoteHost) (rpcclient.RpcClientConnection, error) {
	if len(conns) == 0 {
		return nil, nil
	}
	internalChan := srvMngr.subsystems[subsystem].GetIntenternalChan()
	if srvMngr.GetConfig().DispatcherSCfg().Enabled {
		internalChan = srvMngr.dispatcherSChan
	}
	return engine.NewRPCPool(rpcclient.POOL_FIRST,
		srvMngr.cfg.TlsCfg().ClientKey,
		srvMngr.cfg.TlsCfg().ClientCerificate, srvMngr.cfg.TlsCfg().CaCertificate,
		srvMngr.cfg.GeneralCfg().ConnectAttempts, srvMngr.cfg.GeneralCfg().Reconnects,
		srvMngr.cfg.GeneralCfg().ConnectTimeout, srvMngr.cfg.GeneralCfg().ReplyTimeout,
		conns, internalChan, false)
}

// StartServices starts all enabled services
func (srvMngr *ServiceManager) StartServices() (err error) {
	go srvMngr.handleReload()
	if srvMngr.cfg.AttributeSCfg().Enabled {
		go func() {
			if attrS, has := srvMngr.subsystems[utils.AttributeS]; !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.AttributeS))
				srvMngr.engineShutdown <- true
			} else if err = attrS.Start(srvMngr, true); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start %s because: %s", utils.ServiceManager, utils.AttributeS, err))
				srvMngr.engineShutdown <- true
			}
		}()
	}
	if srvMngr.cfg.ChargerSCfg().Enabled {
		go func() {
			if chrS, has := srvMngr.subsystems[utils.ChargerS]; !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.ChargerS))
				srvMngr.engineShutdown <- true
			} else if err = chrS.Start(srvMngr, true); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start %s because: %s", utils.ServiceManager, utils.ChargerS, err))
				srvMngr.engineShutdown <- true
			}
		}()
	}
	if srvMngr.cfg.ThresholdSCfg().Enabled {
		go func() {
			if thrS, has := srvMngr.subsystems[utils.ThresholdS]; !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.ThresholdS))
				srvMngr.engineShutdown <- true
			} else if err = thrS.Start(srvMngr, true); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start %s because: %s", utils.ServiceManager, utils.ThresholdS, err))
				srvMngr.engineShutdown <- true
			}
		}()
	}
	if srvMngr.cfg.StatSCfg().Enabled {
		go func() {
			if stS, has := srvMngr.subsystems[utils.StatS]; !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.StatS))
				srvMngr.engineShutdown <- true
			} else if err = stS.Start(srvMngr, true); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start %s because: %s", utils.ServiceManager, utils.StatS, err))
				srvMngr.engineShutdown <- true
			}
		}()
	}
	if srvMngr.cfg.ResourceSCfg().Enabled {
		go func() {
			if reS, has := srvMngr.subsystems[utils.ResourceS]; !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.ResourceS))
				srvMngr.engineShutdown <- true
			} else if err = reS.Start(srvMngr, true); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start %s because: %s", utils.ServiceManager, utils.ResourceS, err))
				srvMngr.engineShutdown <- true
			}
		}()
	}
	if srvMngr.cfg.SupplierSCfg().Enabled {
		go func() {
			if supS, has := srvMngr.subsystems[utils.SupplierS]; !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.SupplierS))
				srvMngr.engineShutdown <- true
			} else if err = supS.Start(srvMngr, true); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start %s because: %s", utils.ServiceManager, utils.SupplierS, err))
				srvMngr.engineShutdown <- true
			}
		}()
	}
	// startServer()
	return
}

// AddService adds given services
func (srvMngr *ServiceManager) AddService(services ...Service) {
	for _, srv := range services {
		if _, has := srvMngr.subsystems[srv.ServiceName()]; has { // do not rewrite the service
			continue
		}
		srvMngr.subsystems[srv.ServiceName()] = srv
	}
}

func (srvMngr *ServiceManager) handleReload() {
	var err error
	for {
		select {
		case ext := <-srvMngr.engineShutdown:
			srvMngr.engineShutdown <- ext
			return
		case <-srvMngr.cfg.GetReloadChan(config.ATTRIBUTE_JSN):
			attrS, has := srvMngr.subsystems[utils.AttributeS]
			if !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.AttributeS))
				srvMngr.engineShutdown <- true
				return // stop if we encounter an error
			}
			if srvMngr.cfg.AttributeSCfg().Enabled {
				if attrS.IsRunning() {
					if err = attrS.Reload(srvMngr); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to reload <%s>", utils.ServiceManager, utils.AttributeS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				} else {
					if err = attrS.Start(srvMngr, true); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.AttributeS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				}
			} else if attrS.IsRunning() {
				if err = attrS.Shutdown(); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Failed to stop service <%s>", utils.ServiceManager, utils.AttributeS))
					srvMngr.engineShutdown <- true
					return // stop if we encounter an error
				}
			}
		case <-srvMngr.cfg.GetReloadChan(config.ChargerSCfgJson):
			chrS, has := srvMngr.subsystems[utils.ChargerS]
			if !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.ChargerS))
				srvMngr.engineShutdown <- true
				return // stop if we encounter an error
			}
			if srvMngr.cfg.ChargerSCfg().Enabled {
				if chrS.IsRunning() {
					if err = chrS.Reload(srvMngr); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to reload <%s>", utils.ServiceManager, utils.ChargerS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				} else {
					if err = chrS.Start(srvMngr, true); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.ChargerS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				}
			} else if chrS.IsRunning() {
				if err = chrS.Shutdown(); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Failed to stop service <%s>", utils.ServiceManager, utils.ChargerS))
					srvMngr.engineShutdown <- true
					return // stop if we encounter an error
				}
			}
		case <-srvMngr.cfg.GetReloadChan(config.THRESHOLDS_JSON):
			tS, has := srvMngr.subsystems[utils.ThresholdS]
			if !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.ThresholdS))
				srvMngr.engineShutdown <- true
				return // stop if we encounter an error
			}
			if srvMngr.cfg.ThresholdSCfg().Enabled {
				if tS.IsRunning() {
					if err = tS.Reload(srvMngr); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to reload <%s>", utils.ServiceManager, utils.ThresholdS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				} else {
					if err = tS.Start(srvMngr, true); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.ThresholdS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				}
			} else if tS.IsRunning() {
				if err = tS.Shutdown(); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Failed to stop service <%s>", utils.ServiceManager, utils.ThresholdS))
					srvMngr.engineShutdown <- true
					return // stop if we encounter an error
				}
			}
		case <-srvMngr.cfg.GetReloadChan(config.STATS_JSON):
			tS, has := srvMngr.subsystems[utils.StatS]
			if !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.StatS))
				srvMngr.engineShutdown <- true
				return // stop if we encounter an error
			}
			if srvMngr.cfg.StatSCfg().Enabled {
				if tS.IsRunning() {
					if err = tS.Reload(srvMngr); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to reload <%s>", utils.ServiceManager, utils.StatS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				} else {
					if err = tS.Start(srvMngr, true); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.StatS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				}
			} else if tS.IsRunning() {
				if err = tS.Shutdown(); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Failed to stop service <%s>", utils.ServiceManager, utils.StatS))
					srvMngr.engineShutdown <- true
					return // stop if we encounter an error
				}
			}
		case <-srvMngr.cfg.GetReloadChan(config.RESOURCES_JSON):
			tS, has := srvMngr.subsystems[utils.ResourceS]
			if !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.ResourceS))
				srvMngr.engineShutdown <- true
				return // stop if we encounter an error
			}
			if srvMngr.cfg.ResourceSCfg().Enabled {
				if tS.IsRunning() {
					if err = tS.Reload(srvMngr); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to reload <%s>", utils.ServiceManager, utils.ResourceS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				} else {
					if err = tS.Start(srvMngr, true); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.ResourceS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				}
			} else if tS.IsRunning() {
				if err = tS.Shutdown(); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Failed to stop service <%s>", utils.ServiceManager, utils.ResourceS))
					srvMngr.engineShutdown <- true
					return // stop if we encounter an error
				}
			}
		case <-srvMngr.cfg.GetReloadChan(config.SupplierSJson):
			tS, has := srvMngr.subsystems[utils.SupplierS]
			if !has {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.SupplierS))
				srvMngr.engineShutdown <- true
				return // stop if we encounter an error
			}
			if srvMngr.cfg.SupplierSCfg().Enabled {
				if tS.IsRunning() {
					if err = tS.Reload(srvMngr); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to reload <%s>", utils.ServiceManager, utils.SupplierS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				} else {
					if err = tS.Start(srvMngr, true); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> Failed to start <%s>", utils.ServiceManager, utils.SupplierS))
						srvMngr.engineShutdown <- true
						return // stop if we encounter an error
					}
				}
			} else if tS.IsRunning() {
				if err = tS.Shutdown(); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Failed to stop service <%s>", utils.ServiceManager, utils.SupplierS))
					srvMngr.engineShutdown <- true
					return // stop if we encounter an error
				}
			}
		}
		// handle RPC server
	}
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
	// GetConnection creates a rpcClient to the specified subsystem
	GetConnection(subsystem string, cfg []*config.RemoteHost) (rpcclient.RpcClientConnection, error)
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
