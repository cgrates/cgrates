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
	poolID string, connCache *ltcache.Cache) (rpcPool *rpcclient.RPCPool, err error) {
	var rpcClient birpc.ClientConnector
	var atLeastOneConnected bool // If one connected we don't longer return errors
	rpcPool = rpcclient.NewRPCPool(dispatchStrategy, replyTimeout)
	for _, rpcConnCfg := range rpcConnCfgs {
		if rpcConnCfg.Address == utils.EmptyString {
			// in case we have only conns with empty addresse
			// mimic an error to signal that the init was not done
			err = rpcclient.ErrDisconnected
			continue
		}
		if rpcClient, err = NewRPCConnection(ctx, rpcConnCfg, keyPath, certPath, caPath, connAttempts, reconnects,
			maxReconnectInterval, connectTimeout, replyTimeout, internalConnChan, lazyConnect, poolID, rpcConnCfg.ID,
			connCache); err == rpcclient.ErrUnsupportedCodec {
			return nil, fmt.Errorf("Unsupported transport: <%s>", rpcConnCfg.Transport)
		}
		if err == nil {
			atLeastOneConnected = true
		}
		rpcPool.AddClient(rpcClient)
	}
	if atLeastOneConnected {
		err = nil
	}
	return
}

// NewRPCConnection creates a new connection based on the RemoteHost structure
// connCache is used to cache the connection with ID
func NewRPCConnection(ctx *context.Context, cfg *config.RemoteHost, keyPath, certPath, caPath string, connAttempts, reconnects int,
	maxReconnectInterval, connectTimeout, replyTimeout time.Duration, internalConnChan chan birpc.ClientConnector, lazyConnect bool,
	poolID, connID string, connCache *ltcache.Cache) (client birpc.ClientConnector, err error) {
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
			cfg.Address, internalConnChan, lazyConnect, ctx.Client)
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
			utils.FirstNonEmpty(cfg.Transport, rpcclient.GOBrpc),
			nil, lazyConnect, ctx.Client)
	}
	if connID != utils.EmptyString &&
		err == nil {
		connCache.Set(id, client, nil)
	}
	return
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
func (s RPCClientSet) AddInternalRPCClient(name string, connChan chan birpc.ClientConnector) {
	rpc, err := rpcclient.NewRPCClient(context.TODO(), utils.EmptyString, utils.EmptyString, false,
		utils.EmptyString, utils.EmptyString, utils.EmptyString,
		config.CgrConfig().GeneralCfg().ConnectAttempts, config.CgrConfig().GeneralCfg().Reconnects,
		config.CgrConfig().GeneralCfg().MaxReconnectInterval, utils.FibDuration,
		config.CgrConfig().GeneralCfg().ConnectTimeout, config.CgrConfig().GeneralCfg().ReplyTimeout,
		rpcclient.InternalRPC, connChan, true, nil)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Error adding %s to the set: %s", utils.InternalRPCSet, name, err.Error()))
		return
	}
	s[name] = rpc
}

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
	conn, has := s[methodSplit[0]]
	if !has {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	return conn.Call(ctx, method, args, reply)
}

// func (s RPCClientSet) ReconnectInternals(subsystems ...string) (err error) {
// 	for _, subsystem := range subsystems {
// 		if err = s[subsystem].Reconnect(); err != nil {
// 			return
// 		}
// 	}
// }

func NewService(rcvr any) (*birpc.Service, error) {
	return NewServiceWithName(rcvr, utils.EmptyString, false)
}

func NewServiceWithName(rcvr any, name string, useName bool) (*birpc.Service, error) {
	srv, err := birpc.NewService(rcvr, name, useName)
	if err != nil {
		return nil, err
	}
	srv.Methods["Ping"] = pingM
	return srv, nil
}

// func NewDispatcherService(val any) (_ IntService, err error) {
// 	var srv *birpc.Service
// 	if srv, err = birpc.NewService(val, utils.EmptyString, false); err != nil {
// 		return
// 	}
// 	srv.Methods["Ping"] = pingM
// 	s := IntService{srv.Name: srv}
// 	for m, v := range srv.Methods {
// 		key := srv.Name
// 		switch {
// 		case strings.HasPrefix(m, utils.AttributeS):
// 			m = strings.TrimPrefix(m, utils.AttributeS)
// 			key = utils.AttributeS
// 		case strings.HasPrefix(m, utils.CacheS):
// 			m = strings.TrimPrefix(m, utils.CacheS)
// 			key = utils.CacheS
// 		case strings.HasPrefix(m, utils.CDRs):
// 			m = strings.TrimPrefix(m, utils.CDRs)
// 			key = utils.CDRs
// 		case strings.HasPrefix(m, utils.ChargerS):
// 			m = strings.TrimPrefix(m, utils.ChargerS)
// 			key = utils.ChargerS
// 		case strings.HasPrefix(m, utils.ConfigS):
// 			m = strings.TrimPrefix(m, utils.ConfigS)
// 			key = utils.ConfigS
// 		case strings.HasPrefix(m, utils.CoreS):
// 			m = strings.TrimPrefix(m, utils.CoreS)
// 			key = utils.CoreS
// 		case strings.HasPrefix(m, utils.DispatcherS):
// 			m = strings.TrimPrefix(m, utils.DispatcherS)
// 			key = utils.DispatcherS
// 		case strings.HasPrefix(m, utils.EeS):
// 			m = strings.TrimPrefix(m, utils.EeS)
// 			key = utils.EeS
// 		case strings.HasPrefix(m, utils.GuardianS):
// 			m = strings.TrimPrefix(m, utils.GuardianS)
// 			key = utils.GuardianS
// 		case strings.HasPrefix(m, utils.RALs):
// 			m = strings.TrimPrefix(m, utils.RALs)
// 			key = utils.RALs
// 		case strings.HasPrefix(m, utils.ReplicatorS):
// 			m = strings.TrimPrefix(m, utils.ReplicatorS)
// 			key = utils.ReplicatorS
// 		case strings.HasPrefix(m, utils.ResourceS):
// 			m = strings.TrimPrefix(m, utils.ResourceS)
// 			key = utils.ResourceS
// 		case strings.HasPrefix(m, utils.Responder):
// 			m = strings.TrimPrefix(m, utils.Responder)
// 			key = utils.Responder
// 		case strings.HasPrefix(m, utils.RouteS):
// 			m = strings.TrimPrefix(m, utils.RouteS)
// 			key = utils.RouteS
// 		case strings.HasPrefix(m, utils.SchedulerS):
// 			m = strings.TrimPrefix(m, utils.SchedulerS)
// 			key = utils.SchedulerS
// 		case strings.HasPrefix(m, utils.SessionS):
// 			m = strings.TrimPrefix(m, utils.SessionS)
// 			key = utils.SessionS
// 		case strings.HasPrefix(m, utils.StatService):
// 			m = strings.TrimPrefix(m, utils.StatService)
// 			key = utils.StatService
// 		case strings.HasPrefix(m, utils.ThresholdS):
// 			m = strings.TrimPrefix(m, utils.ThresholdS)
// 			key = utils.ThresholdS
// 		}
// 		if (len(m) < 2 || unicode.ToLower(rune(m[0])) != 'v') &&
// 			key != utils.Responder {
// 			continue
// 		}
// 		if key != utils.Responder {
// 			if unicode.IsLower(rune(key[len(key)-1])) {
// 				key += "V"
// 			} else {
// 				key += "v"
// 			}
// 			key += string(m[1])
// 			m = m[2:]
// 		}
// 		srv2, has := s[key]
// 		if !has {
// 			srv2 = new(birpc.Service)
// 			*srv2 = *srv
// 			srv2.Name = key
// 			RegisterPingMethod(srv2.Methods)
// 			s[key] = srv2
// 		}
// 		srv2.Methods[m] = v
// 	}
// 	return s, nil
// }

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

func RegisterPingMethod(methodMap map[string]*birpc.MethodType) {
	methodMap["Ping"] = pingM
}

func LoadAllDataDBToCache(dm *DataManager) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	for key, ids := range map[string][]string{
		utils.DestinationPrefix:        {utils.MetaAny},
		utils.ReverseDestinationPrefix: {utils.MetaAny},
		utils.RatingPlanPrefix:         {utils.MetaAny},
		utils.RatingProfilePrefix:      {utils.MetaAny},
		utils.ActionPrefix:             {utils.MetaAny},
		utils.ActionPlanPrefix:         {utils.MetaAny},
		utils.AccountActionPlansPrefix: {utils.MetaAny},
		utils.ActionTriggerPrefix:      {utils.MetaAny},
		utils.SharedGroupPrefix:        {utils.MetaAny},
		utils.ResourceProfilesPrefix:   {utils.MetaAny},
		utils.ResourcesPrefix:          {utils.MetaAny},
		utils.StatQueuePrefix:          {utils.MetaAny},
		utils.StatQueueProfilePrefix:   {utils.MetaAny},
		utils.ThresholdPrefix:          {utils.MetaAny},
		utils.ThresholdProfilePrefix:   {utils.MetaAny},
		utils.FilterPrefix:             {utils.MetaAny},
		utils.RouteProfilePrefix:       {utils.MetaAny},
		utils.AttributeProfilePrefix:   {utils.MetaAny},
		utils.ChargerProfilePrefix:     {utils.MetaAny},
		utils.DispatcherProfilePrefix:  {utils.MetaAny},
		utils.DispatcherHostPrefix:     {utils.MetaAny},
		utils.TimingsPrefix:            {utils.MetaAny},
		utils.AttributeFilterIndexes:   {utils.MetaAny},
		utils.ResourceFilterIndexes:    {utils.MetaAny},
		utils.StatFilterIndexes:        {utils.MetaAny},
		utils.ThresholdFilterIndexes:   {utils.MetaAny},
		utils.RouteFilterIndexes:       {utils.MetaAny},
		utils.ChargerFilterIndexes:     {utils.MetaAny},
		utils.DispatcherFilterIndexes:  {utils.MetaAny},
		utils.FilterIndexPrfx:          {utils.MetaAny},
	} {
		if err = dm.CacheDataFromDB(key, ids, false); err != nil {
			return
		}
	}
	return
}
