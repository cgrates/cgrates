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
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewRPCPool returns a new pool of connection with the given configuration
func NewRPCPool(dispatchStrategy string, keyPath, certPath, caPath string, connAttempts, reconnects int,
	connectTimeout, replyTimeout time.Duration, rpcConnCfgs []*config.RemoteHost,
	internalConnChan chan rpcclient.ClientConnector, lazyConnect bool,
	biRPCClient rpcclient.BiRPCConector) (rpcPool *rpcclient.RPCPool, err error) {
	var rpcClient rpcclient.ClientConnector
	var atLestOneConnected bool // If one connected we don't longer return errors
	rpcPool = rpcclient.NewRPCPool(dispatchStrategy, replyTimeout)
	for _, rpcConnCfg := range rpcConnCfgs {
		if rpcConnCfg.Address == utils.EmptyString {
			continue
		}
		rpcClient, err = NewRPCConnection(rpcConnCfg, keyPath, certPath, caPath, connAttempts, reconnects,
			connectTimeout, replyTimeout, internalConnChan, lazyConnect, biRPCClient)
		if err == rpcclient.ErrUnsupportedCodec {
			return nil, fmt.Errorf("Unsupported transport: <%s>", rpcConnCfg.Transport)
		}
		if err == nil {
			atLestOneConnected = true
		}
		rpcPool.AddClient(rpcClient)
	}
	if atLestOneConnected {
		err = nil
	}
	return
}

// NewRPCConnection creates a new connection based on the RemoteHost structure
func NewRPCConnection(cfg *config.RemoteHost, keyPath, certPath, caPath string, connAttempts, reconnects int,
	connectTimeout, replyTimeout time.Duration, internalConnChan chan rpcclient.ClientConnector, lazyConnect bool,
	biRPCClient rpcclient.BiRPCConector) (client rpcclient.ClientConnector, err error) {
	if cfg.Address == rpcclient.InternalRPC ||
		cfg.Address == rpcclient.BiRPCInternal {
		return rpcclient.NewRPCClient("", "", cfg.TLS, keyPath, certPath, caPath, connAttempts,
			reconnects, connectTimeout, replyTimeout, cfg.Address, internalConnChan, lazyConnect, biRPCClient)
	}
	return rpcclient.NewRPCClient(utils.TCP, cfg.Address, cfg.TLS, keyPath, certPath, caPath,
		connAttempts, reconnects, connectTimeout, replyTimeout,
		utils.FirstNonEmpty(cfg.Transport, rpcclient.GOBrpc), nil, lazyConnect, biRPCClient)
}

// IntRPC is the global variable that is used to comunicate with all the subsystems internally
var IntRPC RPCClientSet

// NewRPCClientSet initilalizates the map of connections
func NewRPCClientSet() (s RPCClientSet) {
	return make(RPCClientSet)
}

// RPCClientSet is a RPC ClientConnector for the internal subsystems
type RPCClientSet map[string]*rpcclient.RPCClient

// AddInternalRPCClient creates and adds to the set a new rpc client using the provided configuration
func (s RPCClientSet) AddInternalRPCClient(name string, connChan chan rpcclient.ClientConnector) {
	rpc, err := rpcclient.NewRPCClient(utils.EmptyString, utils.EmptyString, false,
		utils.EmptyString, utils.EmptyString, utils.EmptyString,
		config.CgrConfig().GeneralCfg().ConnectAttempts, config.CgrConfig().GeneralCfg().Reconnects,
		config.CgrConfig().GeneralCfg().ConnectTimeout, config.CgrConfig().GeneralCfg().ReplyTimeout,
		rpcclient.InternalRPC, connChan, true, nil)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Error adding %s to the set: %s", utils.InternalRPCSet, name, err.Error()))
		return
	}
	s[name] = rpc
}

// GetInternalChanel is used when RPCClientSet is passed as internal connection for RPCPool
func (s RPCClientSet) GetInternalChanel() chan rpcclient.ClientConnector {
	connChan := make(chan rpcclient.ClientConnector, 1)
	connChan <- s
	return connChan
}

// Call the implementation of the rpcclient.ClientConnector interface
func (s RPCClientSet) Call(method string, args interface{}, reply interface{}) error {
	methodSplit := strings.Split(method, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	conn, has := s[methodSplit[0]]
	if !has {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	return conn.Call(method, args, reply)
}

// func (s RPCClientSet) ReconnectInternals(subsystems ...string) (err error) {
// 	for _, subsystem := range subsystems {
// 		if err = s[subsystem].Reconnect(); err != nil {
// 			return
// 		}
// 	}
// }
