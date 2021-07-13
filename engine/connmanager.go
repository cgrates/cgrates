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
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

// NewConnManager returns the Connection Manager
func NewConnManager(cfg *config.CGRConfig, rpcInternal map[string]chan rpcclient.ClientConnector) (cM *ConnManager) {
	cM = &ConnManager{
		cfg:         cfg,
		rpcInternal: rpcInternal,
		connCache:   ltcache.NewCache(-1, 0, true, nil),
	}
	SetConnManager(cM)
	return
}

// ConnManager handle the RPC connections
type ConnManager struct {
	cfg         *config.CGRConfig
	rpcInternal map[string]chan rpcclient.ClientConnector
	connCache   *ltcache.Cache
}

// getConn is used to retrieve a connection from cache
// in case this doesn't exist create it and cache it
func (cM *ConnManager) getConn(connID string, biRPCClient rpcclient.ClientConnector) (conn rpcclient.ClientConnector, err error) {
	//try to get the connection from cache
	if x, ok := Cache.Get(utils.CacheRPCConnections, connID); ok {
		if x == nil {
			return nil, utils.ErrNotFound
		}
		return x.(rpcclient.ClientConnector), nil
	}
	// in case we don't find in cache create the connection and add this in cache
	var intChan chan rpcclient.ClientConnector
	var connCfg *config.RPCConn
	isBiRPCCLient := false
	if internalChan, has := cM.rpcInternal[connID]; has {
		connCfg = cM.cfg.RPCConns()[utils.MetaInternal]
		intChan = internalChan
		isBiRPCCLient = true
	} else {
		connCfg = cM.cfg.RPCConns()[connID]
		for _, rpcConn := range connCfg.Conns {
			if rpcConn.Address == utils.MetaInternal {
				intChan = IntRPC.GetInternalChanel()
				break
			}
		}
	}
	switch {
	case biRPCClient != nil && isBiRPCCLient: // special handling for SessionS BiJSONRPCClient
		var rply string
		sSIntConn := <-intChan
		intChan <- sSIntConn
		conn = utils.NewBiRPCInternalClient(sSIntConn.(utils.BiRPCServer))
		conn.(*utils.BiRPCInternalClient).SetClientConn(biRPCClient)
		if err = conn.Call(utils.SessionSv1RegisterInternalBiJSONConn,
			utils.EmptyString, &rply); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not register biRPCClient, error: <%s>",
				utils.SessionS, err.Error()))
			return
		}
	case connCfg.Strategy == rpcclient.PoolParallel:
		rpcConnCfg := connCfg.Conns[0] // for parallel we need only the first connection
		var conPool *rpcclient.RPCParallelClientPool
		if rpcConnCfg.Address == utils.MetaInternal {
			conPool, err = rpcclient.NewRPCParallelClientPool("", "", rpcConnCfg.TLS,
				cM.cfg.TlsCfg().ClientKey, cM.cfg.TlsCfg().ClientCerificate,
				cM.cfg.TlsCfg().CaCertificate, cM.cfg.GeneralCfg().ConnectAttempts,
				cM.cfg.GeneralCfg().Reconnects, cM.cfg.GeneralCfg().ConnectTimeout,
				cM.cfg.GeneralCfg().ReplyTimeout, rpcclient.InternalRPC, intChan, int64(cM.cfg.GeneralCfg().MaxParallelConns), false)
		} else if utils.SliceHasMember([]string{utils.EmptyString, utils.MetaGOB, utils.MetaJSON}, rpcConnCfg.Transport) {
			codec := rpcclient.GOBrpc
			if rpcConnCfg.Transport != "" {
				codec = rpcConnCfg.Transport
			}
			conPool, err = rpcclient.NewRPCParallelClientPool(utils.TCP, rpcConnCfg.Address, rpcConnCfg.TLS,
				cM.cfg.TlsCfg().ClientKey, cM.cfg.TlsCfg().ClientCerificate,
				cM.cfg.TlsCfg().CaCertificate, cM.cfg.GeneralCfg().ConnectAttempts,
				cM.cfg.GeneralCfg().Reconnects, cM.cfg.GeneralCfg().ConnectTimeout,
				cM.cfg.GeneralCfg().ReplyTimeout, codec, nil, int64(cM.cfg.GeneralCfg().MaxParallelConns), false)
		} else {
			err = fmt.Errorf("Unsupported transport: <%s>", rpcConnCfg.Transport)
		}
		if err != nil {
			return
		}
		conn = conPool
	default:
		var conPool *rpcclient.RPCPool
		if conPool, err = NewRPCPool(connCfg.Strategy,
			cM.cfg.TlsCfg().ClientKey,
			cM.cfg.TlsCfg().ClientCerificate, cM.cfg.TlsCfg().CaCertificate,
			cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects,
			cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
			connCfg.Conns, intChan, false); err != nil {
			return
		}
		conn = conPool
	}

	Cache.Set(utils.CacheRPCConnections, connID, conn, nil,
		true, utils.NonTransactional)
	return
}

// Call gets the connection calls the method on it
func (cM *ConnManager) Call(connIDs []string, biRPCClient rpcclient.ClientConnector,
	method string, arg, reply interface{}) (err error) {
	if len(connIDs) == 0 {
		return utils.NewErrMandatoryIeMissing("connIDs")
	}
	var conn rpcclient.ClientConnector
	for _, connID := range connIDs {
		if conn, err = cM.getConn(connID, biRPCClient); err != nil {
			continue
		}
		if err = conn.Call(method, arg, reply); utils.IsNetworkError(err) {
			continue
		} else {
			return
		}
	}
	return
}

func (cM *ConnManager) Reload() {
	Cache.Clear([]string{utils.CacheRPCConnections})
	Cache.Clear([]string{utils.CacheReplicationHosts})
	cM.connCache.Clear()
}

func (cM *ConnManager) CallWithConnIDs(connIDs []string, subsHostIDs utils.StringSet, method string, arg, reply interface{}) (err error) {
	if len(connIDs) == 0 {
		return utils.NewErrMandatoryIeMissing("connIDs")
	}
	// no connection for this id exit here
	if subsHostIDs.Size() == 0 {
		return
	}
	var conn rpcclient.ClientConnector
	for _, connID := range connIDs {
		// recreate the config with only conns that are needed
		connCfg := cM.cfg.RPCConns()[connID]
		newCfg := &config.RPCConn{
			Strategy: connCfg.Strategy,
			PoolSize: connCfg.PoolSize,
			// alloc for all connection in order to not increase the size later
			Conns: make([]*config.RemoteHost, 0, len(connCfg.Conns)),
		}
		for _, conn := range connCfg.Conns {
			if conn.ID != utils.EmptyString &&
				subsHostIDs.Has(conn.ID) {
				newCfg.Conns = append(newCfg.Conns, conn) // the slice will never grow
			}
		}
		if len(newCfg.Conns) == 0 {
			// skip this pool if no connection matches
			continue
		}

		if conn, err = cM.getConnWithConfig(connID, newCfg, nil, nil, false); err != nil {
			continue
		}
		if err = conn.Call(method, arg, reply); !rpcclient.IsNetworkError(err) {
			return
		}
	}
	return
}

func (cM *ConnManager) getConnWithConfig(connID string, connCfg *config.RPCConn,
	biRPCClient rpcclient.ClientConnector, intChan chan rpcclient.ClientConnector,
	isInternalRPC bool) (conn rpcclient.ClientConnector, err error) {
	switch {
	case biRPCClient != nil && isInternalRPC:
		var rply string
		sSIntConn := <-intChan
		intChan <- sSIntConn
		conn = utils.NewBiRPCInternalClient(sSIntConn.(utils.BiRPCServer))
		conn.(*utils.BiRPCInternalClient).SetClientConn(biRPCClient)
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
		case rpcConnCfg.Transport == utils.EmptyString:
			intChan = nil
		case rpcConnCfg.Transport == rpcclient.GOBrpc,
			rpcConnCfg.Transport == rpcclient.JSONrpc:
			codec = rpcConnCfg.Transport
			intChan = nil
		default:
			err = fmt.Errorf("Unsupported transport: <%s>", rpcConnCfg.Transport)
			return
		}
		if conn, err = rpcclient.NewRPCParallelClientPool(utils.TCP, rpcConnCfg.Address, rpcConnCfg.TLS,
			cM.cfg.TlsCfg().ClientKey, cM.cfg.TlsCfg().ClientCerificate,
			cM.cfg.TlsCfg().CaCertificate, cM.cfg.GeneralCfg().ConnectAttempts,
			cM.cfg.GeneralCfg().Reconnects, cM.cfg.GeneralCfg().ConnectTimeout,
			cM.cfg.GeneralCfg().ReplyTimeout, codec, intChan, int64(cM.cfg.GeneralCfg().MaxParallelConns), false); err != nil {
			return
		}
	default:
		if conn, err = NewRPCPool(connCfg.Strategy,
			cM.cfg.TlsCfg().ClientKey,
			cM.cfg.TlsCfg().ClientCerificate, cM.cfg.TlsCfg().CaCertificate,
			cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects,
			cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
			connCfg.Conns, intChan, false); err != nil {
			return
		}
	}
	return
}
