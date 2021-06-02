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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
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
	rcv, err := cM.getConn(context.Background(), connID)

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

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: defaultCfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	experr := rpcclient.ErrUnsupportedBiRPC
	exp, err := NewRPCPool(context.Background(), "*first", "", "", "", defaultCfg.GeneralCfg().ConnectAttempts,
		defaultCfg.GeneralCfg().Reconnects, defaultCfg.GeneralCfg().ConnectTimeout,
		defaultCfg.GeneralCfg().ReplyTimeout, nil, cc, true, nil, "", cM.connCache)
	if err != nil {
		t.Fatal(err)
	}
	rcv, err := cM.getConn(context.Background(), connID)

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

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: defaultCfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			"testString": cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	cM.connCache.Set(connID, nil, nil)

	exp, err := NewRPCPool(context.Background(), "*first", defaultCfg.TLSCfg().ClientKey, defaultCfg.TLSCfg().ClientCerificate,
		defaultCfg.TLSCfg().CaCertificate, defaultCfg.GeneralCfg().ConnectAttempts,
		defaultCfg.GeneralCfg().Reconnects, defaultCfg.GeneralCfg().ConnectTimeout,
		defaultCfg.GeneralCfg().ReplyTimeout, defaultCfg.RPCConns()[connID].Conns, cc,
		true, nil, connID, cM.connCache)
	if err != nil {
		t.Fatal(err)
	}
	rcv, err := cM.getConn(context.Background(), connID)

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

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: defaultCfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	experr := fmt.Sprintf("Unsupported transport: <%+s>", "invalid")
	rcv, err := cM.getConnWithConfig(context.Background(), connID, defaultCfg.RPCConns()[connID], cc, true)

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

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: defaultCfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	experr := rpcclient.ErrUnsupportedCodec
	var exp *rpcclient.RPCParallelClientPool
	rcv, err := cM.getConnWithConfig(context.Background(), connID, defaultCfg.RPCConns()[connID], cc, true)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestCMgetConnWithConfigEmptyTransport(t *testing.T) {
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
			Transport: "",
		},
	}

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: defaultCfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	cM.connCache.Set(connID, nil, nil)

	rcv, err := cM.getConnWithConfig(context.Background(), connID, defaultCfg.RPCConns()[connID], cc, true)

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if _, cancast := rcv.(*rpcclient.RPCParallelClientPool); !cancast {
		t.Error("Expected value of type rpcclient.RPCParallelClientPool")
	}
}

func TestCMgetConnWithConfigInternalRPCCodec(t *testing.T) {
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
			Address: rpcclient.InternalRPC,
		},
	}

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: defaultCfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	rcv, err := cM.getConnWithConfig(context.Background(), connID, defaultCfg.RPCConns()[connID], cc, true)

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if _, cancast := rcv.(*rpcclient.RPCParallelClientPool); !cancast {
		t.Error("Expected value of type rpcclient.RPCParallelClientPool")
	}
}

func TestCMgetConnWithConfigInternalBiRPCCodecUnsupported(t *testing.T) {
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
			Address: rpcclient.BiRPCInternal,
		},
	}

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: defaultCfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	experr := rpcclient.ErrUnsupportedCodec
	rcv, err := cM.getConnWithConfig(context.Background(), connID, defaultCfg.RPCConns()[connID], cc, true)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if _, cancast := rcv.(*rpcclient.RPCParallelClientPool); !cancast {
		t.Error("Expected value of type rpcclient.RPCParallelClientPool")
	}
}

func TestCMCallErrgetConn(t *testing.T) {
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
	err := cM.Call(context.Background(), []string{connID}, "", "", "")

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestCMCallWithConnIDsNoSubsHostIDs(t *testing.T) {
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
	subsHostIDs := utils.StringSet{}

	err := cM.CallWithConnIDs([]string{connID}, context.Background(), subsHostIDs, "", "", "")

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestCMCallWithConnIDsNoConnIDs(t *testing.T) {
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
	subsHostIDs := utils.StringSet{
		"key": struct{}{},
	}

	experr := fmt.Sprintf("MANDATORY_IE_MISSING: [%s]", "connIDs")
	err := cM.CallWithConnIDs([]string{}, context.Background(), subsHostIDs, "", "", "")

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestCMCallWithConnIDsNoConns(t *testing.T) {
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
			Address: rpcclient.InternalRPC,
		},
	}

	cM := &ConnManager{
		cfg: defaultCfg,
	}
	subsHostIDs := utils.StringSet{
		"random": struct{}{},
	}

	err := cM.CallWithConnIDs([]string{connID}, context.Background(), subsHostIDs, "", "", "")

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestCMCallWithConnIDsInternallyDCed(t *testing.T) {
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
			Address: rpcclient.InternalRPC,
		},
	}

	cM := &ConnManager{
		cfg:       defaultCfg,
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}
	subsHostIDs := utils.StringSet{
		connID: struct{}{},
	}

	experr := rpcclient.ErrInternallyDisconnected
	err := cM.CallWithConnIDs([]string{connID}, context.Background(), subsHostIDs, "", "", "")

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestCMCallWithConnIDsErrNotNetwork(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	poolID := "poolID"
	connID := "connID"
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.RPCConns()[poolID] = config.NewDfltRPCConn()
	defaultCfg.RPCConns()[poolID].Conns = []*config.RemoteHost{
		{
			ID:      connID,
			Address: "addr",
		},
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			"testMethod": func(ctx *context.Context, args, reply interface{}) error {
				return utils.ErrExists
			},
		},
	}

	cM := &ConnManager{
		cfg:       defaultCfg,
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	cM.connCache.Set(poolID+utils.ConcatenatedKeySep+connID, ccM, nil)

	subsHostIDs := utils.StringSet{
		connID: struct{}{},
	}

	experr := utils.ErrExists
	err := cM.CallWithConnIDs([]string{poolID}, context.Background(), subsHostIDs, "testMethod", "", "")

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestCMReload(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	defaultCfg := config.NewDefaultCGRConfig()

	cM := &ConnManager{
		cfg:       defaultCfg,
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}
	cM.connCache.Set("itmID1", "value of first item", nil)

	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, defaultCfg.CacheCfg(), cM)
	Cache = NewCacheS(defaultCfg, dm, nil)
	Cache.SetWithoutReplicate(utils.CacheRPCConnections, "itmID2",
		"value of 2nd item", nil, true, utils.NonTransactional)

	var exp []string
	cM.Reload()
	rcv1 := cM.connCache.GetItemIDs("itmID1")
	rcv2 := Cache.GetItemIDs(utils.CacheRPCConnections, utils.EmptyString)

	if !reflect.DeepEqual(rcv1, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv1)
	}

	if !reflect.DeepEqual(rcv2, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv2)
	}
}
