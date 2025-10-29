/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewRPCPool returns a new pool of connection with the given configuration
func NewRPCPool(dispatchStrategy string, keyPath, certPath, caPath string, connAttempts, reconnects int,
	connectTimeout, replyTimeout time.Duration, rpcConnCfgs []*config.RemoteHost,
	internalConnChan chan birpc.ClientConnector, lazyConnect bool) (*rpcclient.RPCPool, error) {
	var rpcClient *rpcclient.RPCClient
	var err error
	rpcPool := rpcclient.NewRPCPool(dispatchStrategy, replyTimeout)
	atLestOneConnected := false // If one connected we don't longer return errors
	for _, rpcConnCfg := range rpcConnCfgs {
		if rpcConnCfg.Address == utils.MetaInternal {
			rpcClient, err = rpcclient.NewRPCClient(context.TODO(), "", "", rpcConnCfg.TLS, keyPath, certPath, caPath,
				connAttempts, reconnects, 0, utils.FibDuration, connectTimeout, replyTimeout, rpcclient.InternalRPC,
				internalConnChan, lazyConnect, nil)
		} else if utils.SliceHasMember([]string{utils.EmptyString, utils.MetaGOB, utils.MetaJSON}, rpcConnCfg.Transport) {
			codec := rpcclient.GOBrpc
			if rpcConnCfg.Transport != "" {
				codec = rpcConnCfg.Transport
			}
			rpcClient, err = rpcclient.NewRPCClient(context.TODO(), utils.TCP, rpcConnCfg.Address, rpcConnCfg.TLS,
				keyPath, certPath, caPath, connAttempts, reconnects, 0, utils.FibDuration,
				connectTimeout, replyTimeout, codec, nil, lazyConnect, nil)
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
func (s *RPCClientSet) AddInternalRPCClient(name string, connChan chan birpc.ClientConnector) {
	rpc, err := rpcclient.NewRPCClient(context.TODO(), utils.EmptyString, utils.EmptyString, false,
		utils.EmptyString, utils.EmptyString, utils.EmptyString, config.CgrConfig().GeneralCfg().ConnectAttempts,
		config.CgrConfig().GeneralCfg().Reconnects, 0, utils.FibDuration, config.CgrConfig().GeneralCfg().ConnectTimeout,
		config.CgrConfig().GeneralCfg().ReplyTimeout, rpcclient.InternalRPC, connChan, true, nil)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Error adding %s to the set: %s", utils.InternalRPCSet, name, err.Error()))
		return
	}
	s.set[name] = rpc
}

// GetInternalChanel is used when RPCClientSet is passed as internal connection for RPCPool
func (s *RPCClientSet) GetInternalChanel() chan birpc.ClientConnector {
	connChan := make(chan birpc.ClientConnector, 1)
	connChan <- s
	return connChan
}

// Call the implementation of the birpc.ClientConnector interface
func (s *RPCClientSet) Call(ctx *context.Context, method string, args any, reply any) error {
	methodSplit := strings.Split(method, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	conn, has := s.set[methodSplit[0]]
	if !has {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	return conn.Call(context.TODO(), method, args, reply)
}

func NewBiRPCService(val any) (srv *birpc.Service, err error) {
	var initialSrv *birpc.Service
	if initialSrv, err = birpc.NewService(val, "", false); err != nil {
		return
	}
	srv = new(birpc.Service)
	*srv = *initialSrv
	srv.Methods = make(map[string]*birpc.MethodType)
	for mName, mType := range initialSrv.Methods {
		mName = strings.TrimPrefix(mName, "BiRPC")
		if len(mName) < 2 || unicode.ToLower(rune(mName[0])) != 'v' {
			continue
		}
		srv.Methods[mName[2:]] = mType
	}
	return
}
