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
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewRPCPool returns a new pool of connection with the given configuration
func NewRPCPool(dispatchStrategy string, keyPath, certPath, caPath string, connAttempts, reconnects int,
	connectTimeout, replyTimeout time.Duration, rpcConnCfgs []*config.RemoteHost,
	internalConnChan chan rpcclient.ClientConnector, lazyConnect bool) (*rpcclient.RPCPool, error) {
	var rpcClient *rpcclient.RPCClient
	var err error
	rpcPool := rpcclient.NewRPCPool(dispatchStrategy, replyTimeout)
	atLestOneConnected := false // If one connected we don't longer return errors
	for _, rpcConnCfg := range rpcConnCfgs {
		if rpcConnCfg.Address == utils.MetaInternal {
			rpcClient, err = rpcclient.NewRPCClient("", "", rpcConnCfg.TLS, keyPath, certPath, caPath, connAttempts,
				reconnects, connectTimeout, replyTimeout, rpcclient.InternalRPC, internalConnChan, lazyConnect)
		} else if utils.SliceHasMember([]string{utils.EmptyString, utils.MetaGOB, utils.MetaJSON}, rpcConnCfg.Transport) {
			codec := rpcclient.GOBrpc
			if rpcConnCfg.Transport != "" {
				codec = rpcConnCfg.Transport
			}
			rpcClient, err = rpcclient.NewRPCClient(utils.TCP, rpcConnCfg.Address, rpcConnCfg.TLS, keyPath, certPath, caPath,
				connAttempts, reconnects, connectTimeout, replyTimeout, codec, nil, lazyConnect)
		} else {
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
	return rpcPool, err
}

// IntRPC is the global variable that is used to comunicate with all the subsystems internally
var IntRPC *RPCClientSet

// NewRPCClientSet initilalizates the map of connections
func NewRPCClientSet() (s *RPCClientSet) {
	return &RPCClientSet{set: make(map[string]*rpcclient.RPCClient)}
}

// RPCClientSet is a RPC ClientConnector for the internal subsystems
type RPCClientSet struct {
	set map[string]*rpcclient.RPCClient
}

// AddInternalRPCClient creates and adds to the set a new rpc client using the provided configuration
func (s *RPCClientSet) AddInternalRPCClient(name string, connChan chan rpcclient.ClientConnector) {
	rpc, err := rpcclient.NewRPCClient(utils.EmptyString, utils.EmptyString, false,
		utils.EmptyString, utils.EmptyString, utils.EmptyString,
		config.CgrConfig().GeneralCfg().ConnectAttempts, config.CgrConfig().GeneralCfg().Reconnects,
		config.CgrConfig().GeneralCfg().ConnectTimeout, config.CgrConfig().GeneralCfg().ReplyTimeout,
		rpcclient.InternalRPC, connChan, true)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Error adding %s to the set: %s", utils.InternalRPCSet, name, err.Error()))
		return
	}
	s.set[name] = rpc
}

// GetInternalChanel is used when RPCClientSet is passed as internal connection for RPCPool
func (s *RPCClientSet) GetInternalChanel() chan rpcclient.ClientConnector {
	connChan := make(chan rpcclient.ClientConnector, 1)
	connChan <- s
	return connChan
}

// Call the implementation of the rpcclient.ClientConnector interface
func (s *RPCClientSet) Call(method string, args interface{}, reply interface{}) error {
	methodSplit := strings.Split(method, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	conn, has := s.set[methodSplit[0]]
	if !has {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	return conn.Call(method, args, reply)
}
