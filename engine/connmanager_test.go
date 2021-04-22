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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

func TestCMgetConnNotFound(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	connID := "connID"
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cM := &ConnManager{
		cfg: defaultCfg,
	}

	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, defaultCfg.CacheCfg(), cM)
	Cache = NewCacheS(defaultCfg, dm, nil)
	Cache.SetWithoutReplicate(utils.CacheRPCConnections, connID, nil, nil, true, utils.NonTransactional)

	experr := utils.ErrNotFound
	rcv, err := cM.getConn(connID, nil)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}

}

func TestCMgetConnUnsupportedBiRPC(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	connID := rpcclient.BiRPCInternal + "connID"
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cc := make(chan rpcclient.ClientConnector, 1)

	cM := &ConnManager{
		cfg: defaultCfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	experr := rpcclient.ErrUnsupportedBiRPC
	exp, err := NewRPCPool("*first", "", "", "", defaultCfg.GeneralCfg().ConnectAttempts,
		defaultCfg.GeneralCfg().Reconnects, defaultCfg.GeneralCfg().ConnectTimeout,
		defaultCfg.GeneralCfg().ReplyTimeout, nil, cc, true, nil, "", cM.connCache)
	if err != nil {
		t.Fatal(err)
	}
	rcv, err := cM.getConn(connID, nil)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestCMgetConnNotInternalRPC(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	connID := "connID"
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.RPCConns()[connID] = config.NewDfltRPCConn()
	defaultCfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			ID:      connID,
			Address: utils.MetaInternal,
		},
	}

	cc := make(chan rpcclient.ClientConnector, 1)

	cM := &ConnManager{
		cfg: defaultCfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			"testString": cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	cM.connCache.Set(connID, nil, nil)

	exp, err := NewRPCPool("*first", defaultCfg.TLSCfg().ClientKey, defaultCfg.TLSCfg().ClientCerificate,
		defaultCfg.TLSCfg().CaCertificate, defaultCfg.GeneralCfg().ConnectAttempts,
		defaultCfg.GeneralCfg().Reconnects, defaultCfg.GeneralCfg().ConnectTimeout,
		defaultCfg.GeneralCfg().ReplyTimeout, defaultCfg.RPCConns()[connID].Conns, cc,
		true, nil, connID, cM.connCache)
	if err != nil {
		t.Fatal(err)
	}
	rcv, err := cM.getConn(connID, nil)

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestCMgetConnWithConfigUnsupportedTransport(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	connID := "connID"
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.RPCConns()[connID] = config.NewDfltRPCConn()
	defaultCfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	defaultCfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address:   "invalid",
			Transport: "invalid",
		},
	}

	cc := make(chan rpcclient.ClientConnector, 1)

	cM := &ConnManager{
		cfg: defaultCfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	experr := fmt.Sprintf("Unsupported transport: <%+s>", "invalid")
	rcv, err := cM.getConnWithConfig(connID, defaultCfg.RPCConns()[connID], nil, cc, true)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestCMgetConnWithConfigUnsupportedCodec(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	connID := "connID"
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.RPCConns()[connID] = config.NewDfltRPCConn()
	defaultCfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	defaultCfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address:   "invalid",
			Transport: rpcclient.BiRPCJSON,
		},
	}

	cc := make(chan rpcclient.ClientConnector, 1)

	cM := &ConnManager{
		cfg: defaultCfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	experr := rpcclient.ErrUnsupportedCodec
	var exp *rpcclient.RPCParallelClientPool
	rcv, err := cM.getConnWithConfig(connID, defaultCfg.RPCConns()[connID], nil, cc, true)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

// func TestCMgetConnWithConfigEmptyTransport(t *testing.T) {
// 	tmp := Cache
// 	defer func() {
// 		Cache = tmp
// 	}()

// 	connID := "connID"
// 	defaultCfg := config.NewDefaultCGRConfig()
// 	defaultCfg.RPCConns()[connID] = config.NewDfltRPCConn()
// 	defaultCfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
// 	defaultCfg.RPCConns()[connID].Conns = []*config.RemoteHost{
// 		{
// 			Address:   "invalid",
// 			Transport: "",
// 		},
// 	}

// 	cc := make(chan rpcclient.ClientConnector, 1)

// 	cM := &ConnManager{
// 		cfg: defaultCfg,
// 		rpcInternal: map[string]chan rpcclient.ClientConnector{
// 			connID: cc,
// 		},
// 		connCache: ltcache.NewCache(-1, 0, true, nil),
// 	}

// 	cM.connCache.Set(connID, nil, nil)

// 	exp, err := rpcclient.NewRPCParallelClientPool(utils.TCP, "invalid", false,
// 		defaultCfg.TLSCfg().ClientKey, defaultCfg.TLSCfg().ClientCerificate,
// 		defaultCfg.TLSCfg().CaCertificate, defaultCfg.GeneralCfg().ConnectAttempts,
// 		defaultCfg.GeneralCfg().Reconnects, defaultCfg.GeneralCfg().ConnectTimeout,
// 		defaultCfg.GeneralCfg().ReplyTimeout, rpcclient.GOBrpc, nil,
// 		int64(defaultCfg.GeneralCfg().MaxParallelConns), false, nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rcv, err := cM.getConnWithConfig(connID, defaultCfg.RPCConns()[connID], nil, cc, true)

// 	if err != nil {
// 		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
// 	}

// 	if !reflect.DeepEqual(rcv, exp) {
// 		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
// 	}
// }

// func TestCMgetConnWithConfig2(t *testing.T) {
// 	tmp := Cache
// 	defer func() {
// 		Cache = tmp
// 	}()

// 	connID := "connID"
// 	defaultCfg := config.NewDefaultCGRConfig()
// 	defaultCfg.RPCConns()[connID] = config.NewDfltRPCConn()
// 	defaultCfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
// 	defaultCfg.RPCConns()[connID].Conns = []*config.RemoteHost{
// 		{
// 			ID:        connID,
// 			Address:   rpcclient.InternalRPC,
// 			Transport: rpcclient.BiRPCJSON,
// 		},
// 	}

// 	cc := make(chan rpcclient.ClientConnector, 1)

// 	cM := &ConnManager{
// 		cfg: defaultCfg,
// 		rpcInternal: map[string]chan rpcclient.ClientConnector{
// 			connID: cc,
// 		},
// 		connCache: ltcache.NewCache(-1, 0, true, nil),
// 	}

// 	experr := rpcclient.ErrUnsupportedCodec
// 	var exp *rpcclient.RPCParallelClientPool
// 	rcv, err := cM.getConnWithConfig(connID, defaultCfg.RPCConns()[connID], nil, cc, true)

// 	if err == nil || err != experr {
// 		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
// 	}

// 	if !reflect.DeepEqual(rcv, exp) {
// 		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
// 	}
// }
