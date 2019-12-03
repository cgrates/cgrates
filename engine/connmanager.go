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

package engine

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewConnManager returns the Connection Manager
func NewConnManager(cfg *config.CGRConfig, rpcInternal map[string]chan rpcclient.RpcClientConnection) (cM *ConnManager) {
	return &ConnManager{cfg: cfg, rpcInternal: rpcInternal}
}

type ConnManager struct {
	cfg         *config.CGRConfig
	rpcInternal map[string]chan rpcclient.RpcClientConnection
}

func (cM *ConnManager) getConn(connID string) (connPool *rpcclient.RpcClientPool, err error) {
	//try to get the connection from cache
	if x, ok := Cache.Get(utils.CacheRPCConnections, connID); ok {
		if x == nil {
			return nil, utils.ErrNotFound
		}
		return x.(*rpcclient.RpcClientPool), nil
	}
	// in case we don't found in cache create the connection and add this in cache
	var intChan chan rpcclient.RpcClientConnection
	var connCfg *config.RPCConn
	if internalChan, has := cM.rpcInternal[connID]; has {
		connCfg = cM.cfg.RPCConns()[utils.MetaInternal]
		intChan = internalChan
	} else {
		connCfg = cM.cfg.RPCConns()[connID]
	}
	if connPool, err = NewRPCPool(connCfg.Strategy,
		cM.cfg.TlsCfg().ClientKey,
		cM.cfg.TlsCfg().ClientCerificate, cM.cfg.TlsCfg().CaCertificate,
		cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects,
		cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
		connCfg.Conns, intChan, false); err != nil {
		return
	}
	Cache.Set(utils.CacheRPCConnections, connID, connPool, nil,
		true, utils.NonTransactional)
	return
}

func (cM *ConnManager) Call(connIDs []string, method string, arg, reply interface{}) (err error) {
	var conn *rpcclient.RpcClientPool
	for _, connID := range connIDs {
		if conn, err = cM.getConn(connID); err != nil {
			continue
		}
		return conn.Call(method, arg, reply)
	}
	return
}
