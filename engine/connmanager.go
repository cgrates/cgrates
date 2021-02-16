/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewConnManager returns the Connection Manager
func NewConnManager(cfg *config.CGRConfig, rpcInternal map[string]chan rpcclient.ClientConnector) (cM *ConnManager) {
	cM = &ConnManager{cfg: cfg, rpcInternal: rpcInternal}
	SetConnManager(cM)
	return
}

// ConnManager handle the RPC connections
type ConnManager struct {
	cfg         *config.CGRConfig
	rpcInternal map[string]chan rpcclient.ClientConnector
}

// getConn is used to retrieve a connection from cache
// in case this doesn't exist create it and cache it
func (cM *ConnManager) getConn(connID string, biRPCClient rpcclient.BiRPCConector) (conn rpcclient.ClientConnector, err error) {
	//try to get the connection from cache
	if x, ok := Cache.Get(utils.CacheRPCConnections, connID); ok {
		if x == nil {
			return nil, utils.ErrNotFound
		}
		return x.(rpcclient.ClientConnector), nil
	}
	// in case we don't find in cache create the connection and add this in cache
	var intChan chan rpcclient.ClientConnector
	var isInternalRPC bool
	connCfg := cM.cfg.RPCConns()[utils.MetaInternal]
	if intChan, isInternalRPC = cM.rpcInternal[connID]; !isInternalRPC {
		connCfg = cM.cfg.RPCConns()[connID]
		for _, rpcConn := range connCfg.Conns {
			if rpcConn.Address == utils.MetaInternal {
				intChan = IntRPC.GetInternalChanel()
				break
			}
		}
	}
	switch {
	case biRPCClient != nil && isInternalRPC: // special handling for SessionS BiJSONRPCClient
		if conn, err = rpcclient.NewRPCClient(utils.EmptyString, utils.EmptyString, false,
			utils.EmptyString, utils.EmptyString, utils.EmptyString,
			cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects,
			cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
			rpcclient.BiRPCInternal, intChan, false, biRPCClient); err != nil {
			return
		}
		var rply string
		if err = conn.Call(utils.SessionSv1RegisterInternalBiJSONConn,
			utils.EmptyString, &rply); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not register biRPCClient, error: <%s>",
				utils.SessionS, err.Error()))
			return
		}
	case connCfg.Strategy == rpcclient.PoolParallel:
		rpcConnCfg := connCfg.Conns[0] // for parrallel we need only the first connection
		codec := rpcclient.GOBrpc
		switch {
		case rpcConnCfg.Address == rpcclient.InternalRPC:
			codec = rpcclient.InternalRPC
		case rpcConnCfg.Address == rpcclient.BiRPCInternal:
			codec = rpcclient.BiRPCInternal
		case rpcConnCfg.Transport == utils.EmptyString:
			intChan = nil
		case rpcConnCfg.Transport == rpcclient.GOBrpc,
			rpcConnCfg.Transport == rpcclient.JSONrpc,
			rpcConnCfg.Transport == rpcclient.BiRPCGOB,
			rpcConnCfg.Transport == rpcclient.BiRPCJSON:
			codec = rpcConnCfg.Transport
			intChan = nil
		default:
			err = fmt.Errorf("Unsupported transport: <%s>", rpcConnCfg.Transport)
			return
		}
		if conn, err = rpcclient.NewRPCParallelClientPool(utils.TCP, rpcConnCfg.Address, rpcConnCfg.TLS,
			cM.cfg.TLSCfg().ClientKey, cM.cfg.TLSCfg().ClientCerificate,
			cM.cfg.TLSCfg().CaCertificate, cM.cfg.GeneralCfg().ConnectAttempts,
			cM.cfg.GeneralCfg().Reconnects, cM.cfg.GeneralCfg().ConnectTimeout,
			cM.cfg.GeneralCfg().ReplyTimeout, codec, intChan, int64(cM.cfg.GeneralCfg().MaxParallelConns), false, biRPCClient); err != nil {
			return
		}
	default:
		if conn, err = NewRPCPool(connCfg.Strategy,
			cM.cfg.TLSCfg().ClientKey,
			cM.cfg.TLSCfg().ClientCerificate, cM.cfg.TLSCfg().CaCertificate,
			cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects,
			cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
			connCfg.Conns, intChan, false, biRPCClient); err != nil {
			return
		}
	}

	if err = Cache.Set(utils.CacheRPCConnections, connID, conn, nil,
		true, utils.NonTransactional); err != nil {
		return
	}
	return
}

// Call gets the connection calls the method on it
func (cM *ConnManager) Call(connIDs []string, biRPCClient rpcclient.BiRPCConector,
	method string, arg, reply interface{}) (err error) {
	if len(connIDs) == 0 {
		return utils.NewErrMandatoryIeMissing("connIDs")
	}
	var conn rpcclient.ClientConnector
	for _, connID := range connIDs {
		if conn, err = cM.getConn(connID, biRPCClient); err != nil {
			continue
		}
		if err = conn.Call(method, arg, reply); rpcclient.IsNetworkError(err) {
			continue
		} else {
			return
		}
	}
	return
}
