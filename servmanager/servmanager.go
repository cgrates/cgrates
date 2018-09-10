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
	"reflect"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewServiceManager(cfg *config.CGRConfig, dm *engine.DataManager,
	engineShutdown chan bool, cacheS *engine.CacheS) *ServiceManager {
	return &ServiceManager{cfg: cfg, dm: dm,
		engineShutdown: engineShutdown, cacheS: cacheS}
}

// ServiceManager handles service management ran by the engine
type ServiceManager struct {
	sync.RWMutex   // lock access to any shared data
	cfg            *config.CGRConfig
	dm             *engine.DataManager
	engineShutdown chan bool
	cacheS         *engine.CacheS
	sched          *scheduler.Scheduler
	rpcChans       map[string]chan rpcclient.RpcClientConnection // services expected to start
	rpcServices    map[string]rpcclient.RpcClientConnection      // services started
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
