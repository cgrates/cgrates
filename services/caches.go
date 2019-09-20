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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewCacheService returns the Cache Service
func NewCacheService() servmanager.Service {
	return &CacheService{
		connChan: make(chan rpcclient.RpcClientConnection, 1),
	}
}

// CacheService implements Service interface
type CacheService struct {
	sync.RWMutex
	chS      *engine.CacheS
	rpc      *v1.CacheSv1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
// inits the CacheS and starts precaching as well as populating internal channel for RPC conns
func (chS *CacheService) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	// safe to not check CacheS should never be stoped and then started again
	// if chS.IsRunning() {
	// return fmt.Errorf("service aleady running")
	// }

	chS.Lock()
	defer chS.Unlock()
	chS.chS = engine.NewCacheS(sp.GetConfig(), sp.GetDM())
	go func() {
		if err := chS.chS.Precache(); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> could not init, error: %s", utils.CacheS, err.Error()))
			sp.GetExitChan() <- true
		}
	}()

	chS.rpc = v1.NewCacheSv1(chS.chS)
	if !sp.GetConfig().DispatcherSCfg().Enabled {
		sp.GetServer().RpcRegister(chS.rpc)
	}
	chS.connChan <- chS.rpc

	// set the cache in ServiceManager
	sp.SetCacheS(chS.chS)
	return
}

// GetIntenternalChan returns the internal connection chanel
func (chS *CacheService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return chS.connChan
}

// Reload handles the change of config
func (chS *CacheService) Reload(sp servmanager.ServiceProvider) (err error) {
	return
}

// Shutdown stops the service
func (chS *CacheService) Shutdown() (err error) {
	return
}

// GetRPCInterface returns the interface to register for server
func (chS *CacheService) GetRPCInterface() interface{} {
	return chS.rpc
}

// IsRunning returns if the service is running
func (chS *CacheService) IsRunning() bool {
	chS.RLock()
	defer chS.RUnlock()
	return chS != nil && chS.chS != nil
}

// ServiceName returns the service name
func (chS *CacheService) ServiceName() string {
	return utils.CacheS
}
