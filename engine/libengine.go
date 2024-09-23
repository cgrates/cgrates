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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

// NewRPCPool returns a new pool of connection with the given configuration
func NewRPCPool(ctx *context.Context, dispatchStrategy string, keyPath, certPath, caPath string, connAttempts, reconnects int,
	maxReconnectInterval, connectTimeout, replyTimeout time.Duration, rpcConnCfgs []*config.RemoteHost,
	internalConnChan chan birpc.ClientConnector, lazyConnect bool,
	biRPCClient any, poolID string, connCache *ltcache.Cache) (rpcPool *rpcclient.RPCPool, err error) {
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
			maxReconnectInterval, connectTimeout, replyTimeout, internalConnChan, lazyConnect, biRPCClient,
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
	maxReconnectInterval, connectTimeout, replyTimeout time.Duration, internalConnChan chan birpc.ClientConnector, lazyConnect bool,
	biRPCClient any, poolID, connID string, connCache *ltcache.Cache) (client birpc.ClientConnector, err error) {
	var id string
	if connID != utils.EmptyString {
		id = poolID + utils.ConcatenatedKeySep + connID
		if x, ok := connCache.Get(id); ok && x != nil {
			return x.(birpc.ClientConnector), nil
		}
	}
	if cfg.Address == rpcclient.InternalRPC ||
		cfg.Address == rpcclient.BiRPCInternal {
		client, err = rpcclient.NewRPCClient(ctx, "", "", cfg.TLS,
			utils.FirstNonEmpty(cfg.ClientKey, keyPath),
			utils.FirstNonEmpty(cfg.ClientCertificate, certPath),
			utils.FirstNonEmpty(cfg.CaCertificate, caPath),
			utils.FirstIntNonEmpty(cfg.ConnectAttempts, connAttempts),
			utils.FirstIntNonEmpty(cfg.Reconnects, reconnects),
			utils.FirstDurationNonEmpty(cfg.MaxReconnectInterval, maxReconnectInterval),
			utils.FibDuration,
			utils.FirstDurationNonEmpty(cfg.ConnectTimeout, connectTimeout),
			utils.FirstDurationNonEmpty(cfg.ReplyTimeout, replyTimeout),
			cfg.Address, internalConnChan, lazyConnect, biRPCClient)
	} else {
		client, err = rpcclient.NewRPCClient(ctx, utils.TCP, cfg.Address, cfg.TLS,
			utils.FirstNonEmpty(cfg.ClientKey, keyPath),
			utils.FirstNonEmpty(cfg.ClientCertificate, certPath),
			utils.FirstNonEmpty(cfg.CaCertificate, caPath),
			utils.FirstIntNonEmpty(cfg.ConnectAttempts, connAttempts),
			utils.FirstIntNonEmpty(cfg.Reconnects, reconnects),
			utils.FirstDurationNonEmpty(cfg.MaxReconnectInterval, maxReconnectInterval),
			utils.FibDuration,
			utils.FirstDurationNonEmpty(cfg.ConnectTimeout, connectTimeout),
			utils.FirstDurationNonEmpty(cfg.ReplyTimeout, replyTimeout),
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
func (s RPCClientSet) Call(ctx *context.Context, method string, args any, reply any) error {
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

func NewService(val any) (_ IntService, err error) {
	return NewServiceWithName(val, utils.EmptyString, false)
}
func NewServiceWithPing(val any, name, prefix string) (*birpc.Service, error) {
	srv, err := birpc.NewServiceWithMethodsRename(val, name, true, func(oldFn string) (newFn string) {
		return strings.TrimPrefix(oldFn, prefix)
	})
	if err != nil {
		return nil, err
	}
	srv.Methods["Ping"] = pingM
	return srv, nil
}

func NewServiceWithName(val any, name string, useName bool) (_ IntService, err error) {
	var srv *birpc.Service
	if srv, err = birpc.NewService(val, name, useName); err != nil {
		return
	}
	srv.Methods["Ping"] = pingM
	s := IntService{srv.Name: srv}
	for m, v := range srv.Methods {
		m = strings.TrimPrefix(m, "BiRPC")
		if len(m) < 2 || unicode.ToLower(rune(m[0])) != 'v' {
			continue
		}

		key := srv.Name
		if unicode.IsLower(rune(key[len(key)-1])) {
			key += "V"
		} else {
			key += "v"
		}
		key += string(m[1])
		srv2, has := s[key]
		if !has {
			srv2 = new(birpc.Service)
			*srv2 = *srv
			srv2.Name = key
			srv2.Methods = map[string]*birpc.MethodType{"Ping": pingM}
			s[key] = srv2
		}
		srv2.Methods[m[2:]] = v
	}
	return s, nil
}

func NewDispatcherService(val any) (_ IntService, err error) {
	var srv *birpc.Service
	if srv, err = birpc.NewService(val, utils.EmptyString, false); err != nil {
		return
	}
	srv.Methods["Ping"] = pingM
	s := IntService{srv.Name: srv}
	for m, v := range srv.Methods {
		key := srv.Name
		switch {
		case strings.HasPrefix(m, utils.AccountS):
			m = strings.TrimPrefix(m, utils.AccountS)
			key = utils.AccountS
		case strings.HasPrefix(m, utils.ActionS):
			m = strings.TrimPrefix(m, utils.ActionS)
			key = utils.ActionS
		case strings.HasPrefix(m, utils.AttributeS):
			m = strings.TrimPrefix(m, utils.AttributeS)
			key = utils.AttributeS
		case strings.HasPrefix(m, utils.CacheS):
			m = strings.TrimPrefix(m, utils.CacheS)
			key = utils.CacheS
		case strings.HasPrefix(m, utils.ChargerS):
			m = strings.TrimPrefix(m, utils.ChargerS)
			key = utils.ChargerS
		case strings.HasPrefix(m, utils.ConfigS):
			m = strings.TrimPrefix(m, utils.ConfigS)
			key = utils.ConfigS
		case strings.HasPrefix(m, utils.DispatcherS):
			m = strings.TrimPrefix(m, utils.DispatcherS)
			key = utils.DispatcherS
		case strings.HasPrefix(m, utils.GuardianS):
			m = strings.TrimPrefix(m, utils.GuardianS)
			key = utils.GuardianS
		case strings.HasPrefix(m, utils.RateS):
			m = strings.TrimPrefix(m, utils.RateS)
			key = utils.RateS
		case strings.HasPrefix(m, utils.ReplicatorS):
			m = strings.TrimPrefix(m, utils.ReplicatorS)
			key = utils.ReplicatorS
		case strings.HasPrefix(m, utils.ResourceS):
			m = strings.TrimPrefix(m, utils.ResourceS)
			key = utils.ResourceS
		case strings.HasPrefix(m, utils.RouteS):
			m = strings.TrimPrefix(m, utils.RouteS)
			key = utils.RouteS
		case strings.HasPrefix(m, utils.SessionS):
			m = strings.TrimPrefix(m, utils.SessionS)
			key = utils.SessionS
		case strings.HasPrefix(m, utils.StatS):
			m = strings.TrimPrefix(m, utils.StatS)
			key = utils.StatS
		case strings.HasPrefix(m, utils.ThresholdS):
			m = strings.TrimPrefix(m, utils.ThresholdS)
			key = utils.ThresholdS
		case strings.HasPrefix(m, utils.CDRs):
			m = strings.TrimPrefix(m, utils.CDRs)
			key = utils.CDRs
		case strings.HasPrefix(m, utils.EeS):
			m = strings.TrimPrefix(m, utils.EeS)
			key = utils.EeS
		case strings.HasPrefix(m, utils.CoreS):
			m = strings.TrimPrefix(m, utils.CoreS)
			key = utils.CoreS
		case strings.HasPrefix(m, utils.AnalyzerS):
			m = strings.TrimPrefix(m, utils.AnalyzerS)
			key = utils.AnalyzerS
		case strings.HasPrefix(m, utils.AdminS):
			m = strings.TrimPrefix(m, utils.AdminS)
			key = utils.AdminS
		case strings.HasPrefix(m, utils.LoaderS):
			m = strings.TrimPrefix(m, utils.LoaderS)
			key = utils.LoaderS
		case strings.HasPrefix(m, utils.ServiceManagerS):
			m = strings.TrimPrefix(m, utils.ServiceManagerS)
			key = utils.ServiceManagerS
		}
		if len(m) < 2 || unicode.ToLower(rune(m[0])) != 'v' {
			continue
		}
		if unicode.IsLower(rune(key[len(key)-1])) {
			key += "V"
		} else {
			key += "v"
		}
		key += string(m[1])
		srv2, has := s[key]
		if !has {
			srv2 = new(birpc.Service)
			*srv2 = *srv
			srv2.Name = key
			srv2.Methods = map[string]*birpc.MethodType{"Ping": pingM}
			s[key] = srv2
		}
		srv2.Methods[m[2:]] = v
	}
	return s, nil
}

type IntService map[string]*birpc.Service

func (s IntService) Call(ctx *context.Context, serviceMethod string, args, reply any) error {
	service, has := s[strings.Split(serviceMethod, utils.NestingSep)[0]]
	if !has {
		return errors.New("rpc: can't find service " + serviceMethod)
	}
	return service.Call(ctx, serviceMethod, args, reply)
}

func ping(_ any, _ *context.Context, _ *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

var pingM = &birpc.MethodType{
	Method: reflect.Method{
		Name: "Ping",
		Type: reflect.TypeOf(ping),
		Func: reflect.ValueOf(ping),
	},
	ArgType:   reflect.TypeOf(new(utils.CGREvent)),
	ReplyType: reflect.TypeOf(new(string)),
}
