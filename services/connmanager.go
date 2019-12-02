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
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewConnManager returns the Connection Manager
func NewConnManager(cfg *config.CGRConfig) (cM *ConnManager) {
	return &ConnManager{cfg: cfg}
}

type ConnManager struct {
	sync.RWMutex
	cfg *config.CGRConfig
}

func (cM *ConnManager) GetConn(connID string,
	internalChan chan rpcclient.RpcClientConnection) (connPool *rpcclient.RpcClientPool, err error) {
	//try to get the connection from cache
	if x, ok := engine.Cache.Get(utils.CacheRPCConnections, connID); ok {
		if x == nil {
			return nil, utils.ErrNotFound
		}
		return x.(*rpcclient.RpcClientPool), nil
	}
	// in case we don't found in cache create the connection and add this in cache
	connCfg := cM.cfg.RPCConns()[connID]
	connPool, err = engine.NewRPCPool(connCfg.Strategy,
		cM.cfg.TlsCfg().ClientKey,
		cM.cfg.TlsCfg().ClientCerificate, cM.cfg.TlsCfg().CaCertificate,
		cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects,
		cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
		connCfg.Conns, internalChan, false)
	if err != nil {
		return
	}
	engine.Cache.Set(utils.CacheRPCConnections, connID, connPool, nil,
		true, utils.NonTransactional)
	return
}

// Start should handle the sercive start
func (cM *ConnManager) Start() (err error) {
	return
}

// GetIntenternalChan returns the internal connection chanel
func (cM *ConnManager) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return nil
}

// Reload handles the change of config
func (cM *ConnManager) Reload() (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (cM *ConnManager) Shutdown() (err error) {
	return
}

// IsRunning returns if the service is running
func (cM *ConnManager) IsRunning() bool {
	return true
}

// ServiceName returns the service name
func (cM *ConnManager) ServiceName() string {
	return utils.RPCConnS
}

// ShouldRun returns if the service should be running
func (cM *ConnManager) ShouldRun() bool {
	return true
}
