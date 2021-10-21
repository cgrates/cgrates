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
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cM := &ConnManager{
		cfg: cfg,
	}

	Cache.Clear(nil)
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
	connID := rpcclient.BiRPCInternal + "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cc := make(chan rpcclient.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	experr := rpcclient.ErrUnsupportedBiRPC
	exp, err := NewRPCPool("*first", "", "", "", cfg.GeneralCfg().ConnectAttempts,
		cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().ConnectTimeout,
		cfg.GeneralCfg().ReplyTimeout, nil, cc, true, nil, "", cM.connCache)
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
	connID := "connID"
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			ID:      connID,
			Address: utils.MetaInternal,
		},
	}

	cc := make(chan rpcclient.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			"testString": cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	cM.connCache.Set(connID, nil, nil)

	exp, err := NewRPCPool("*first", cfg.TLSCfg().ClientKey, cfg.TLSCfg().ClientCerificate,
		cfg.TLSCfg().CaCertificate, cfg.GeneralCfg().ConnectAttempts,
		cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().ConnectTimeout,
		cfg.GeneralCfg().ReplyTimeout, cfg.RPCConns()[connID].Conns, cc,
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
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address:   "invalid",
			Transport: "invalid",
		},
	}

	cc := make(chan rpcclient.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	experr := fmt.Sprintf("Unsupported transport: <%+s>", "invalid")
	rcv, err := cM.getConnWithConfig(connID, cfg.RPCConns()[connID], nil, cc, true)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestCMgetConnWithConfigUnsupportedCodec(t *testing.T) {
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address:   "invalid",
			Transport: rpcclient.BiRPCJSON,
		},
	}

	cc := make(chan rpcclient.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	experr := rpcclient.ErrUnsupportedCodec
	var exp *rpcclient.RPCParallelClientPool
	rcv, err := cM.getConnWithConfig(connID, cfg.RPCConns()[connID], nil, cc, true)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestCMgetConnWithConfigEmptyTransport(t *testing.T) {
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address:   "invalid",
			Transport: "",
		},
	}

	cc := make(chan rpcclient.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	cM.connCache.Set(connID, nil, nil)

	rcv, err := cM.getConnWithConfig(connID, cfg.RPCConns()[connID], nil, cc, true)

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if _, cancast := rcv.(*rpcclient.RPCParallelClientPool); !cancast {
		t.Error("Expected value of type rpcclient.RPCParallelClientPool")
	}
}

type BiRPCConnectorMock struct {
	calls map[string]func(rpcclient.ClientConnector, string, interface{}, interface{}) error
}

func (bRCM *BiRPCConnectorMock) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	return nil
}

func (bRCM *BiRPCConnectorMock) CallBiRPC(cc rpcclient.ClientConnector, method string, args interface{}, reply interface{}) error {
	if call, has := bRCM.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(cc, method, args, reply)
	}
}

func (bRCM *BiRPCConnectorMock) Handlers() map[string]interface{} {
	return nil
}

func TestCMgetConnWithConfigCallBiRPCNilErr(t *testing.T) {
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			ID:        connID,
			Address:   rpcclient.BiRPCInternal,
			Transport: rpcclient.BiRPCJSON,
		},
	}

	cc := make(chan rpcclient.ClientConnector, 1)
	birpc := &BiRPCConnectorMock{
		calls: map[string]func(rpcclient.ClientConnector, string, interface{}, interface{}) error{
			utils.SessionSv1RegisterInternalBiJSONConn: func(cc rpcclient.ClientConnector, s string, i1, i2 interface{}) error {
				return nil
			},
		},
	}
	cc <- birpc

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	exp, err := NewRPCPool("*first", cfg.TLSCfg().ClientKey, cfg.TLSCfg().ClientCerificate,
		cfg.TLSCfg().CaCertificate, cfg.GeneralCfg().ConnectAttempts,
		cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().ConnectTimeout,
		cfg.GeneralCfg().ReplyTimeout, cfg.RPCConns()[connID].Conns, cc,
		false, birpc, connID, cM.connCache)
	if err != nil {
		t.Fatal(err)
	}
	rcv, err := cM.getConnWithConfig(connID, cfg.RPCConns()[connID], birpc, cc, true)

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestCMgetConnWithConfigCallBiRPCErr(t *testing.T) {
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			ID:        connID,
			Address:   rpcclient.BiRPCInternal,
			Transport: rpcclient.BiRPCJSON,
		},
	}

	cc := make(chan rpcclient.ClientConnector, 1)
	birpc := &BiRPCConnectorMock{
		calls: map[string]func(rpcclient.ClientConnector, string, interface{}, interface{}) error{
			"wrong method": func(cc rpcclient.ClientConnector, s string, i1, i2 interface{}) error {
				return nil
			},
		},
	}
	cc <- birpc

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	exp, err := NewRPCPool("*first", cfg.TLSCfg().ClientKey, cfg.TLSCfg().ClientCerificate,
		cfg.TLSCfg().CaCertificate, cfg.GeneralCfg().ConnectAttempts,
		cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().ConnectTimeout,
		cfg.GeneralCfg().ReplyTimeout, cfg.RPCConns()[connID].Conns, cc,
		false, birpc, connID, cM.connCache)
	if err != nil {
		t.Fatal(err)
	}
	experr := rpcclient.ErrUnsupporteServiceMethod
	rcv, err := cM.getConnWithConfig(connID, cfg.RPCConns()[connID], birpc, cc, true)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestCMgetConnWithConfigInternalRPCCodec(t *testing.T) {
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address: rpcclient.InternalRPC,
		},
	}

	cc := make(chan rpcclient.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	rcv, err := cM.getConnWithConfig(connID, cfg.RPCConns()[connID], nil, cc, true)

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if _, cancast := rcv.(*rpcclient.RPCParallelClientPool); !cancast {
		t.Error("Expected value of type rpcclient.RPCParallelClientPool")
	}
}

func TestCMgetConnWithConfigInternalBiRPCCodecUnsupported(t *testing.T) {
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address: rpcclient.BiRPCInternal,
		},
	}

	cc := make(chan rpcclient.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan rpcclient.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	experr := rpcclient.ErrUnsupportedCodec
	rcv, err := cM.getConnWithConfig(connID, cfg.RPCConns()[connID], nil, cc, true)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if _, cancast := rcv.(*rpcclient.RPCParallelClientPool); !cancast {
		t.Error("Expected value of type rpcclient.RPCParallelClientPool")
	}
}

func TestCMCallErrgetConn(t *testing.T) {
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cM := &ConnManager{
		cfg: cfg,
	}

	Cache.Clear(nil)
	Cache.SetWithoutReplicate(utils.CacheRPCConnections, connID, nil, nil, true, utils.NonTransactional)

	experr := utils.ErrNotFound
	err := cM.Call([]string{connID}, nil, "", "", "")

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestCMCallWithConnIDsNoSubsHostIDs(t *testing.T) {
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cM := &ConnManager{
		cfg: cfg,
	}
	subsHostIDs := utils.StringSet{}

	err := cM.CallWithConnIDs([]string{connID}, subsHostIDs, "", "", "")

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestCMCallWithConnIDsNoConnIDs(t *testing.T) {
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cM := &ConnManager{
		cfg: cfg,
	}
	subsHostIDs := utils.StringSet{
		"key": struct{}{},
	}

	experr := fmt.Sprintf("MANDATORY_IE_MISSING: [%s]", "connIDs")
	err := cM.CallWithConnIDs([]string{}, subsHostIDs, "", "", "")

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestCMCallWithConnIDsNoConns(t *testing.T) {
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			ID:      connID,
			Address: rpcclient.InternalRPC,
		},
	}

	cM := &ConnManager{
		cfg: cfg,
	}
	subsHostIDs := utils.StringSet{
		"random": struct{}{},
	}

	err := cM.CallWithConnIDs([]string{connID}, subsHostIDs, "", "", "")

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestCMCallWithConnIDsInternallyDCed(t *testing.T) {
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			ID:      connID,
			Address: rpcclient.InternalRPC,
		},
	}

	cM := &ConnManager{
		cfg:       cfg,
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}
	subsHostIDs := utils.StringSet{
		connID: struct{}{},
	}

	experr := rpcclient.ErrInternallyDisconnected
	err := cM.CallWithConnIDs([]string{connID}, subsHostIDs, "", "", "")

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestCMCallWithConnIDs2(t *testing.T) {
	poolID := "poolID"
	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[poolID] = config.NewDfltRPCConn()
	cfg.RPCConns()[poolID].Conns = []*config.RemoteHost{
		{
			ID:      connID,
			Address: "addr",
		},
	}

	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			"testMethod": func(args, reply interface{}) error {
				return utils.ErrExists
			},
		},
	}

	cM := &ConnManager{
		cfg:       cfg,
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}

	cM.connCache.Set(poolID+utils.ConcatenatedKeySep+connID, ccM, nil)

	subsHostIDs := utils.StringSet{
		connID: struct{}{},
	}

	experr := utils.ErrExists
	err := cM.CallWithConnIDs([]string{poolID}, subsHostIDs, "testMethod", "", "")

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestCMReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cM := &ConnManager{
		cfg:       cfg,
		connCache: ltcache.NewCache(-1, 0, true, nil),
	}
	cM.connCache.Set("itmID1", "value of first item", nil)

	Cache.Clear(nil)
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

/*
func TestCMDeadLock(t *testing.T) {
	// to not break the next tests reset the values
	tCh := Cache
	tCfg := config.CgrConfig()
	defer func() {
		Cache = tCh
		config.SetCgrConfig(tCfg)
	}()

	cfg := config.NewDefaultCGRConfig()
	// define a dummy replication conn
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheStatQueueProfiles].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	config.SetCgrConfig(cfg)

	Cache = NewCacheS(cfg, nil, nil)

	Cache.SetWithoutReplicate(utils.CacheStatQueueProfiles, "", nil, nil, true, utils.NonTransactional)
	done := make(chan struct{}) // signal

	go func() {
		Cache.Clear(nil)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Deadlock on cache")
	}
}
*/
