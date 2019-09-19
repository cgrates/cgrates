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

package services

import (
	"fmt"
	"sync"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewSchedulerService returns the Scheduler Service
func NewSchedulerService() *SchedulerService {
	return &SchedulerService{
		connChan: make(chan rpcclient.RpcClientConnection, 1),
	}
}

// SchedulerService implements Service interface
type SchedulerService struct {
	sync.RWMutex
	schS     *scheduler.Scheduler
	rpc      *v1.SchedulerSv1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
func (schS *SchedulerService) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if schS.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	schS.Lock()
	if waitCache { // Wait for cache to load data before starting
		<-sp.GetCacheS().GetPrecacheChannel(utils.CacheActionPlans) // wait for ActionPlans to be cached
	}
	utils.Logger.Info("<ServiceManager> Starting CGRateS Scheduler.")
	schS.schS = scheduler.NewScheduler(sp.GetDM())
	go schS.schS.Loop()

	schS.rpc = v1.NewSchedulerSv1(sp.GetConfig())
	if !sp.GetConfig().DispatcherSCfg().Enabled {
		sp.GetServer().RpcRegister(schS.rpc)
	}
	schS.Unlock()
	schS.connChan <- schS.rpc

	// Create connection to CDR Server and share it in engine(used for *cdrlog action)
	cdrsConn, err := sp.GetConnection(utils.CDRServer, sp.GetConfig().SchedulerCfg().CDRsConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to CDRServer: %s", utils.SchedulerS, err.Error()))
		return
	}

	// ToDo: this should be send to scheduler
	engine.SetSchedCdrsConns(cdrsConn)

	return
}

// GetIntenternalChan returns the internal connection chanel
func (schS *SchedulerService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return schS.connChan
}

// Reload handles the change of config
func (schS *SchedulerService) Reload(sp servmanager.ServiceProvider) (err error) {
	schS.RLock()
	schS.schS.Reload()
	defer schS.RUnlock()
	return
}

// Shutdown stops the service
func (schS *SchedulerService) Shutdown() (err error) {
	schS.schS.Shutdown()
	schS.Lock()
	schS.schS = nil
	schS.rpc = nil
	schS.Unlock()
	<-schS.connChan
	return
}

// GetRPCInterface returns the interface to register for server
func (schS *SchedulerService) GetRPCInterface() interface{} {
	schS.RLock()
	defer schS.RUnlock()
	return schS.rpc
}

// IsRunning returns if the service is running
func (schS *SchedulerService) IsRunning() bool {
	schS.RLock()
	defer schS.RUnlock()
	return schS != nil && schS.schS != nil
}

// ServiceName returns the service name
func (schS *SchedulerService) ServiceName() string {
	return utils.SchedulerS
}

// GetScheduler returns the Scheduler
func (schS *SchedulerService) GetScheduler() *scheduler.Scheduler {
	schS.RLock()
	defer schS.RUnlock()
	return schS.schS
}
