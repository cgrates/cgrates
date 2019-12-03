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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewConnManagerService(cfg *config.CGRConfig, intConns map[string]chan rpcclient.RpcClientConnection) *ConnManagerService {
	fmt.Println("Enter in NewConnManagerService")
	fmt.Println(intConns)
	return &ConnManagerService{
		cfg:     cfg,
		connMgr: engine.NewConnManager(cfg, intConns),
	}
}

type ConnManagerService struct {
	sync.RWMutex
	cfg     *config.CGRConfig
	connMgr *engine.ConnManager
}

// Start should handle the sercive start
func (cM *ConnManagerService) Start() (err error) {
	return
}

// GetIntenternalChan returns the internal connection chanel
func (cM *ConnManagerService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return nil
}

// Reload handles the change of config
func (cM *ConnManagerService) Reload() (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (cM *ConnManagerService) Shutdown() (err error) {
	return
}

// IsRunning returns if the service is running
func (cM *ConnManagerService) IsRunning() bool {
	return true
}

// ServiceName returns the service name
func (cM *ConnManagerService) ServiceName() string {
	return utils.RPCConnS
}

// ShouldRun returns if the service should be running
func (cM *ConnManagerService) ShouldRun() bool {
	return true
}

func (cM *ConnManagerService) GetConnMgr() *engine.ConnManager {
	return cM.connMgr
}
