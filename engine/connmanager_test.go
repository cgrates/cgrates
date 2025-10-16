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
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cM := &ConnManager{
		cfg: cfg,
	}

	db, _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := NewDataManager(dbCM, cfg, cM)
	Cache = NewCacheS(cfg, dm, nil, nil)
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
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, false, nil),
	}

	experr := rpcclient.ErrUnsupportedBiRPC
	exp, err := NewRPCPool(context.Background(), utils.MetaFirst, "", "", "", cfg.GeneralCfg().ConnectAttempts,
		cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().MaxReconnectInterval, cfg.GeneralCfg().ConnectTimeout,
		cfg.GeneralCfg().ReplyTimeout, nil, cc, true, nil, "", cM.connCache)
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
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			ID:      connID,
			Address: utils.MetaInternal,
		},
	}

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			"testString": cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, false, nil),
	}

	cM.connCache.Set(connID, nil, nil)

	exp, err := NewRPCPool(context.Background(), utils.MetaFirst, cfg.TLSCfg().ClientKey, cfg.TLSCfg().ClientCerificate,
		cfg.TLSCfg().CaCertificate, cfg.GeneralCfg().ConnectAttempts,
		cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().MaxReconnectInterval,
		cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
		cfg.RPCConns()[connID].Conns, cc, true, nil, connID, cM.connCache)
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
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address:   "invalid",
			Transport: "invalid",
		},
	}

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, false, nil),
	}

	experr := fmt.Sprintf("Unsupported transport: <%+s>", "invalid")
	rcv, err := cM.getConnWithConfig(context.Background(), connID, cfg.RPCConns()[connID], cc, true)

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
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address:   "invalid",
			Transport: rpcclient.BiRPCJSON,
		},
	}

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, false, nil),
	}

	experr := rpcclient.ErrUnsupportedCodec
	var exp *rpcclient.RPCParallelClientPool
	rcv, err := cM.getConnWithConfig(context.Background(), connID, cfg.RPCConns()[connID], cc, true)

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
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address:   "invalid",
			Transport: "",
		},
	}

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, false, nil),
	}

	cM.connCache.Set(connID, nil, nil)

	rcv, err := cM.getConnWithConfig(context.Background(), connID, cfg.RPCConns()[connID], cc, true)

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
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address: rpcclient.InternalRPC,
		},
	}

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, false, nil),
	}

	rcv, err := cM.getConnWithConfig(context.Background(), connID, cfg.RPCConns()[connID], cc, true)

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
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			Address: rpcclient.BiRPCInternal,
		},
	}

	cc := make(chan birpc.ClientConnector, 1)

	cM := &ConnManager{
		cfg: cfg,
		rpcInternal: map[string]chan birpc.ClientConnector{
			connID: cc,
		},
		connCache: ltcache.NewCache(-1, 0, true, false, nil),
	}

	experr := rpcclient.ErrUnsupportedCodec
	rcv, err := cM.getConnWithConfig(context.Background(), connID, cfg.RPCConns()[connID], cc, true)

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
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cM := &ConnManager{
		cfg: cfg,
	}

	db, _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := NewDataManager(dbCM, cfg, cM)
	Cache = NewCacheS(cfg, dm, nil, nil)
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
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cM := &ConnManager{
		cfg: cfg,
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
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()

	cM := &ConnManager{
		cfg: cfg,
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
		connCache: ltcache.NewCache(-1, 0, true, false, nil),
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

type ccMock struct {
	calls map[string]func(ctx *context.Context, args any, reply any) error
}

func (ccM *ccMock) Call(ctx *context.Context, serviceMethod string, args any, reply any) (err error) {
	if call, has := ccM.calls[serviceMethod]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, args, reply)
	}
}

func TestCMCallWithConnIDsErrNotNetwork(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

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
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			"testMethod": func(ctx *context.Context, args, reply any) error {
				return utils.ErrExists
			},
		},
	}

	cM := &ConnManager{
		cfg:       cfg,
		connCache: ltcache.NewCache(-1, 0, true, false, nil),
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

	cfg := config.NewDefaultCGRConfig()

	cM := &ConnManager{
		cfg:       cfg,
		connCache: ltcache.NewCache(-1, 0, true, false, nil),
	}
	cM.connCache.Set("itmID1", "value of first item", nil)

	db, _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := NewDataManager(dbCM, cfg, cM)
	Cache = NewCacheS(cfg, dm, nil, nil)
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

func TestCMGetInternalChan(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)
	cM.dispIntCh = cM.rpcInternal

	exp := make(chan context.ClientConnector, 1)
	rcv := cM.GetInternalChan()
	rcvType := reflect.TypeOf(rcv)
	expType := reflect.TypeOf(exp)
	if rcvType != expType {
		t.Errorf("Unexpected return type, expected %+v, received %+v", rcvType, expType)
	}

}

func TestCMGetDispInternalChan(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)
	cM.dispIntCh = cM.rpcInternal

	exp := make(chan context.ClientConnector, 1)
	rcv := cM.GetDispInternalChan()
	rcvType := reflect.TypeOf(rcv)
	expType := reflect.TypeOf(exp)
	if rcvType != expType {
		t.Errorf("Unexpected return type, expected %+v, received %+v", rcvType, expType)
	}

}

// func TestCMEnableDispatcher(t *testing.T) {
// 	tmp := Cache
// 	defer func() {
// 		Cache = tmp
// 	}()
// 	Cache.Clear(nil)
// 	cfg := config.NewDefaultCGRConfig()
// 	cM := NewConnManager(cfg)
// 	data , _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
// 	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: data}, cfg.DbCfg())
// dm := NewDataManager(dbCM, cfg.CacheCfg(), nil)
// 	fltrs := NewFilterS(cfg, nil, dm)
// 	Cache = NewCacheS(cfg, dm, nil, nil)
// 	var storDB StorDB
// 	storDBChan := make(chan StorDB, 1)
// 	storDBChan <- storDB
// 	newCDRSrv := NewCDRServer(cfg, dm, fltrs, nil, storDBChan)

// 	srvcNames := []string{utils.AccountS, utils.ActionS, utils.AttributeS,
// 		utils.CacheS, utils.ChargerS, utils.ConfigS, utils.DispatcherS,
// 		utils.GuardianS, utils.RateS, utils.ResourceS, utils.RouteS,
// 		utils.SessionS, utils.StatS, utils.ThresholdS, utils.CDRs,
// 		utils.ReplicatorS, utils.EeS, utils.CoreS, utils.AnalyzerS,
// 		utils.AdminS, utils.LoaderS, utils.ServiceManager}

// 	for _, name := range srvcNames {

// 		newSrvcWName, err := NewServiceWithName(newCDRSrv, name, true)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		cM.EnableDispatcher(newSrvcWName)

// 	}
// }

// func TestCMDisableDispatcher(t *testing.T) {
// 	tmp := Cache
// 	defer func() {
// 		Cache = tmp
// 	}()
// 	Cache.Clear(nil)
// 	cfg := config.NewDefaultCGRConfig()

// 	cM := &ConnManager{
// 		cfg:       cfg,
// 		connCache: ltcache.NewCache(-1, 0, true, false, nil),
// 	}
// 	cM.connCache.Set("itmID1", "value of first item", nil)
// 	data , _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
// 	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: data}, cfg.DbCfg())
// dm := NewDataManager(dbCM, cfg.CacheCfg(), nil)
// 	fltrs := NewFilterS(cfg, nil, dm)
// 	var storDB StorDB
// 	storDBChan := make(chan StorDB, 1)
// 	storDBChan <- storDB
// 	newCDRSrv := NewCDRServer(cfg, dm, fltrs, nil, storDBChan)
// 	newSrvcWName, err := NewServiceWithName(newCDRSrv, utils.AccountS, true)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	cM.EnableDispatcher(newSrvcWName)

// 	Cache = NewCacheS(cfg, dm, cM, nil)
// 	Cache.SetWithoutReplicate(utils.CacheRPCConnections, "itmID2",
// 		"value of 2nd item", nil, true, utils.NonTransactional)

// 	var exp []string

// 	cM.DisableDispatcher()
// 	rcv1 := cM.connCache.GetItemIDs("itmID1")
// 	rcv2 := Cache.GetItemIDs(utils.CacheRPCConnections, utils.EmptyString)

// 	if !reflect.DeepEqual(rcv1, exp) {
// 		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv1)
// 	} else if !reflect.DeepEqual(rcv2, exp) {
// 		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv2)
// 	} else if cM.disp != nil || cM.dispIntCh != nil {
// 		t.Errorf("\nexpected nil cM.disp and cM.dispIntCh, \nreceived cM.disp: <%+v>, \n cM.dispIntCh: <%+v>", cM.disp, cM.dispIntCh)
// 	}

// }

// func TestCMgetInternalConnChanFromDisp(t *testing.T) {
// 	tmp := Cache
// 	defer func() {
// 		Cache = tmp
// 	}()
// 	Cache.Clear(nil)
// 	cfg := config.NewDefaultCGRConfig()

// 	cM := &ConnManager{
// 		cfg:       cfg,
// 		connCache: ltcache.NewCache(-1, 0, true, false, nil),
// 	}
// 	cM.connCache.Set("itmID1", "value of first item", nil)
// 	data , _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
// 	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: data}, cfg.DbCfg())
// dm := NewDataManager(dbCM, cfg.CacheCfg(), nil)
// 	fltrs := NewFilterS(cfg, nil, dm)
// 	var storDB StorDB
// 	storDBChan := make(chan StorDB, 1)
// 	storDBChan <- storDB
// 	newCDRSrv := NewCDRServer(cfg, dm, fltrs, nil, storDBChan)
// 	newSrvcWName, err := NewServiceWithName(newCDRSrv, utils.AccountS, true)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	cM.EnableDispatcher(newSrvcWName)

// 	if rcv, ok := cM.getInternalConnChan(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)); !ok {
// 		t.Errorf("Unexpected error getting internalConnChan, Received <%+v>", rcv)
// 	}

// }
