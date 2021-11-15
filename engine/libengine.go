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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

// NewRPCPool returns a new pool of connection with the given configuration
func NewRPCPool(ctx *context.Context, dispatchStrategy string, keyPath, certPath, caPath string, connAttempts, reconnects int,
	connectTimeout, replyTimeout time.Duration, rpcConnCfgs []*config.RemoteHost,
	internalConnChan chan birpc.ClientConnector, lazyConnect bool,
	biRPCClient interface{}, poolID string, connCache *ltcache.Cache) (rpcPool *rpcclient.RPCPool, err error) {
	var rpcClient birpc.ClientConnector
	var atLestOneConnected bool // If one connected we don't longer return errors
	rpcPool = rpcclient.NewRPCPool(dispatchStrategy, replyTimeout)
	for _, rpcConnCfg := range rpcConnCfgs {
		if rpcConnCfg.Address == utils.EmptyString {
			// in case we have only conns with empty addresse
			// mimic an error to signal that the init was not done
			err = rpcclient.ErrDisconnected
			continue
		}
		if rpcClient, err = NewRPCConnection(ctx, rpcConnCfg, keyPath, certPath, caPath, connAttempts, reconnects,
			connectTimeout, replyTimeout, internalConnChan, lazyConnect, biRPCClient,
			poolID, rpcConnCfg.ID, connCache); err == rpcclient.ErrUnsupportedCodec {
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
// connCache is used to cache the connection with ID
func NewRPCConnection(ctx *context.Context, cfg *config.RemoteHost, keyPath, certPath, caPath string, connAttempts, reconnects int,
	connectTimeout, replyTimeout time.Duration, internalConnChan chan birpc.ClientConnector, lazyConnect bool,
	biRPCClient interface{}, poolID, connID string, connCache *ltcache.Cache) (client birpc.ClientConnector, err error) {
	var id string
	if connID != utils.EmptyString {
		id = poolID + utils.ConcatenatedKeySep + connID
		if x, ok := connCache.Get(id); ok && x != nil {
			return x.(birpc.ClientConnector), nil
		}
	}
	if cfg.Address == rpcclient.InternalRPC ||
		cfg.Address == rpcclient.BiRPCInternal {
		client, err = rpcclient.NewRPCClient(ctx, "", "", cfg.TLS, utils.FirstNonEmpty(cfg.ClientKey, keyPath), utils.FirstNonEmpty(cfg.ClientCertificate, certPath), utils.FirstNonEmpty(cfg.CaCertificate, caPath), utils.FirstIntNonEmpty(cfg.ConnectAttempts, connAttempts),
			utils.FirstIntNonEmpty(cfg.Reconnects, reconnects), utils.FirstDurationNonEmpty(cfg.ConnectTimeout, connectTimeout), utils.FirstDurationNonEmpty(cfg.ReplyTimeout, replyTimeout), cfg.Address, internalConnChan, lazyConnect, biRPCClient)
	} else {
		client, err = rpcclient.NewRPCClient(ctx, utils.TCP, cfg.Address, cfg.TLS, utils.FirstNonEmpty(cfg.ClientKey, keyPath), utils.FirstNonEmpty(cfg.ClientCertificate, certPath), utils.FirstNonEmpty(cfg.CaCertificate, caPath),
			utils.FirstIntNonEmpty(cfg.ConnectAttempts, connAttempts),
			utils.FirstIntNonEmpty(cfg.Reconnects, reconnects), utils.FirstDurationNonEmpty(cfg.ConnectTimeout, connectTimeout), utils.FirstDurationNonEmpty(cfg.ReplyTimeout, replyTimeout),
			utils.FirstNonEmpty(cfg.Transport, rpcclient.GOBrpc), nil, lazyConnect, biRPCClient)
	}
	if connID != utils.EmptyString &&
		err == nil {
		connCache.Set(id, client, nil)
	}
	return
}

// RPCClientSet is a RPC ClientConnector for the internal subsystems
type RPCClientSet map[string]chan birpc.ClientConnector

// GetInternalChanel is used when RPCClientSet is passed as internal connection for RPCPool
func (s RPCClientSet) GetInternalChanel() chan birpc.ClientConnector {
	connChan := make(chan birpc.ClientConnector, 1)
	connChan <- s
	return connChan
}

// Call the implementation of the birpc.ClientConnector interface
func (s RPCClientSet) Call(ctx *context.Context, method string, args interface{}, reply interface{}) error {
	methodSplit := strings.Split(method, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	connCh, has := s[methodSplit[0]]
	if !has {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	var conn birpc.ClientConnector
	ctx2, cancel := context.WithTimeout(ctx, config.CgrConfig().GeneralCfg().ConnectTimeout)
	select {
	case conn = <-connCh:
		connCh <- conn
		cancel()
		if conn == nil {
			return rpcclient.ErrDisconnected
		}
	case <-ctx2.Done():
		return ctx2.Err()
	}
	return conn.Call(ctx, method, args, reply)
}

func NewService(val interface{}) (_ IntService, err error) {
	var srv *birpc.Service
	if srv, err = birpc.NewService(val, utils.EmptyString, false); err != nil {
		return
	}
	s := IntService{srv.Name: srv}
	for m, v := range srv.Methods {
		if len(m) < 2 || m[0] != 'V' {
			continue
		}
		key := srv.Name + "v" + string(m[1])
		srv2, has := s[key]
		if !has {
			srv2 = new(birpc.Service)
			*srv2 = *srv
			srv2.Methods = make(map[string]*birpc.MethodType)
			s[key] = srv2
		}
		srv2.Methods[m[2:]] = v

	}
	return s, nil
}

func NewDispatcherService(val interface{}) (_ IntService, err error) {
	var srv *birpc.Service
	if srv, err = birpc.NewService(val, utils.EmptyString, false); err != nil {
		return
	}
	s := IntService{srv.Name: srv}
	for m, v := range srv.Methods {
		key := srv.Name
		switch {
		case strings.HasPrefix(m, utils.AccountS):
			m = strings.TrimRight(m, utils.AccountS)
			key = utils.AccountS
		case strings.HasPrefix(m, utils.ActionS):
			m = strings.TrimRight(m, utils.ActionS)
			key = utils.ActionS
		case strings.HasPrefix(m, utils.AttributeS):
			m = strings.TrimRight(m, utils.AttributeS)
			key = utils.AttributeS
		case strings.HasPrefix(m, utils.CacheS):
			m = strings.TrimRight(m, utils.CacheS)
			key = utils.CacheS
		case strings.HasPrefix(m, utils.ChargerS):
			m = strings.TrimRight(m, utils.ChargerS)
			key = utils.ChargerS
		case strings.HasPrefix(m, utils.ConfigS):
			m = strings.TrimRight(m, utils.ConfigS)
			key = utils.ConfigS
		case strings.HasPrefix(m, utils.DispatcherS):
			m = strings.TrimRight(m, utils.DispatcherS)
			key = utils.DispatcherS
		case strings.HasPrefix(m, utils.GuardianS):
			m = strings.TrimRight(m, utils.GuardianS)
			key = utils.GuardianS
		case strings.HasPrefix(m, utils.RateS):
			m = strings.TrimRight(m, utils.RateS)
			key = utils.RateS
		// case strings.HasPrefix(m, utils.ReplicatorS):
		// 	m = strings.TrimRight(m, utils.ReplicatorS)
		// 	key = utils.ReplicatorS
		case strings.HasPrefix(m, utils.ResourceS):
			m = strings.TrimRight(m, utils.ResourceS)
			key = utils.ResourceS
		case strings.HasPrefix(m, utils.RouteS):
			m = strings.TrimRight(m, utils.RouteS)
			key = utils.RouteS
		case strings.HasPrefix(m, utils.SessionS):
			m = strings.TrimRight(m, utils.SessionS)
			key = utils.SessionS
		case strings.HasPrefix(m, utils.StatS):
			m = strings.TrimRight(m, utils.StatS)
			key = utils.StatS
		case strings.HasPrefix(m, utils.ThresholdS):
			m = strings.TrimRight(m, utils.ThresholdS)
			key = utils.ThresholdS
		case strings.HasPrefix(m, utils.CDRs):
			m = strings.TrimRight(m, utils.CDRs)
			key = utils.CDRs

		case len(m) < 2 || m[0] != 'V':
			continue
		}
		key += "v" + string(m[1])
		srv2, has := s[key]
		if !has {
			srv2 = new(birpc.Service)
			*srv2 = *srv
			srv2.Methods = make(map[string]*birpc.MethodType)
			s[key] = srv2
		}
		srv2.Methods[m[2:]] = v

	}
	return s, nil
}

type IntService map[string]*birpc.Service

func (s IntService) Call(ctx *context.Context, method string, args, reply interface{}) error {
	return s[strings.Split(method, utils.NestingSep)[0]].Call(ctx, method, args, reply)
}
