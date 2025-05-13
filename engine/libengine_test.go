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
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
	"golang.org/x/exp/slices"
)

// For the purpose of this test, we don't need our client to establish a connection
// we only want to check if the client loaded with the given config where needed
func TestLibengineNewRPCConnection(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := &config.RemoteHost{
		ID:              "a4f3f",
		Address:         "localhost:6012",
		Transport:       "*json",
		ConnectAttempts: 2,
		Reconnects:      5,
		ConnectTimeout:  2 * time.Minute,
		ReplyTimeout:    3 * time.Minute,
		TLS:             true,
		ClientKey:       "key1",
	}
	expErr := []string{"dial tcp [::1]:6012: connect: connection refused", "dial tcp 127.0.0.1:6012: connect: connection refused"}
	cM := NewConnManager(config.NewDefaultCGRConfig())
	ctx := context.Background()

	_, err := NewRPCConnection(ctx, cfg, cM.cfg.TLSCfg().ClientKey, cM.cfg.TLSCfg().ClientCerificate,
		cM.cfg.TLSCfg().CaCertificate, cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects,
		cM.cfg.GeneralCfg().MaxReconnectInterval, cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
		nil, false, nil, "*localhost", "a4f3f", new(ltcache.Cache))
	if !slices.Contains(expErr, err.Error()) {
		t.Errorf("Unexpected error <%v>", err)
	}
}

func TestLibengineNewRPCConnectionInternal(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := &config.RemoteHost{
		ID:              "a4f3f",
		Address:         rpcclient.InternalRPC,
		Transport:       "",
		ConnectAttempts: 2,
		Reconnects:      5,
		ConnectTimeout:  2 * time.Minute,
		ReplyTimeout:    3 * time.Minute,
		TLS:             true,
		ClientKey:       "key1",
	}
	cM := NewConnManager(config.NewDefaultCGRConfig())
	exp, err := rpcclient.NewRPCClient(context.Background(), "", "", cfg.TLS, cfg.ClientKey, cM.cfg.TLSCfg().ClientCerificate,
		cM.cfg.TLSCfg().ClientCerificate, cfg.ConnectAttempts, cfg.Reconnects, cfg.MaxReconnectInterval, utils.FibDuration,
		cfg.ConnectTimeout, cfg.ReplyTimeout, rpcclient.InternalRPC, cM.rpcInternal["a4f3f"], false, nil)

	// We only want to check if the client loaded with the correct config,
	// therefore connection is not mandatory
	if err != rpcclient.ErrInternallyDisconnected {
		t.Error(err)
	}

	conn, err := NewRPCConnection(context.Background(), cfg, cM.cfg.TLSCfg().ClientKey, cM.cfg.TLSCfg().ClientCerificate, cM.cfg.TLSCfg().CaCertificate,
		cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects, cM.cfg.GeneralCfg().MaxReconnectInterval, cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
		cM.rpcInternal["a4f3f"], false, nil, "*internal", "a4f3f", new(ltcache.Cache))

	if err != rpcclient.ErrInternallyDisconnected {
		t.Error(err)
	}
	if !reflect.DeepEqual(exp, conn) {
		t.Error("Connections don't match")
	}
}

type TestRPCSrvMock struct{} // exported for service

func (TestRPCSrvMock) Do(*context.Context, any, *string) error   { return nil }
func (TestRPCSrvMock) V1Do(*context.Context, any, *string) error { return nil }
func (TestRPCSrvMock) V2Do(*context.Context, any, *string) error { return nil }

type TestRPCSrvMockS struct{} // exported for service

func (TestRPCSrvMockS) V1Do(*context.Context, any, *string) error { return nil }
func (TestRPCSrvMockS) V2Do(*context.Context, any, *string) error { return nil }

func getMethods(s IntService) (methods map[string][]string) {
	methods = map[string][]string{}
	for _, v := range s {
		for m := range v.Methods {
			methods[v.Name] = append(methods[v.Name], m)
		}
	}
	for k := range methods {
		sort.Strings(methods[k])
	}
	return
}

func TestIntServiceNewService(t *testing.T) {
	expErrMsg := `rpc.Register: no service name for type struct {}`
	if _, err := NewService(struct{}{}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	s, err := NewService(new(TestRPCSrvMock))
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 3 {
		t.Errorf("Not all rpc APIs were registerd")
	}
	methods := getMethods(s)
	exp := map[string][]string{
		"TestRPCSrvMock":   {"Do", "Ping", "V1Do", "V2Do"},
		"TestRPCSrvMockV1": {"Do", "Ping"},
		"TestRPCSrvMockV2": {"Do", "Ping"},
	}
	if !reflect.DeepEqual(exp, methods) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(methods))
	}

	s, err = NewService(new(TestRPCSrvMockS))
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 3 {
		t.Errorf("Not all rpc APIs were registerd")
	}
	methods = getMethods(s)
	exp = map[string][]string{
		"TestRPCSrvMockS":   {"Ping", "V1Do", "V2Do"},
		"TestRPCSrvMockSv1": {"Do", "Ping"},
		"TestRPCSrvMockSv2": {"Do", "Ping"},
	}
	if !reflect.DeepEqual(exp, methods) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(methods))
	}

	var rply string
	if err := s.Call(context.Background(), "TestRPCSrvMockSv1.Ping", new(utils.CGREvent), &rply); err != nil {
		t.Fatal(err)
	} else if rply != utils.Pong {
		t.Errorf("Expeceted: %q, received: %q", utils.Pong, rply)
	}

	expErrMsg = `rpc: can't find service TestRPCSrvMockv1.Ping`
	if err := s.Call(context.Background(), "TestRPCSrvMockv1.Ping", new(utils.CGREvent), &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

}

type TestRPCDspMock struct{} // exported for service

func (TestRPCDspMock) AccountSv1Do(*context.Context, any, *string) error       { return nil }
func (TestRPCDspMock) ActionSv1Do(*context.Context, any, *string) error        { return nil }
func (TestRPCDspMock) AttributeSv1Do(*context.Context, any, *string) error     { return nil }
func (TestRPCDspMock) CacheSv1Do(*context.Context, any, *string) error         { return nil }
func (TestRPCDspMock) ChargerSv1Do(*context.Context, any, *string) error       { return nil }
func (TestRPCDspMock) ConfigSv1Do(*context.Context, any, *string) error        { return nil }
func (TestRPCDspMock) GuardianSv1Do(*context.Context, any, *string) error      { return nil }
func (TestRPCDspMock) RateSv1Do(*context.Context, any, *string) error          { return nil }
func (TestRPCDspMock) ReplicatorSv1Do(*context.Context, any, *string) error    { return nil }
func (TestRPCDspMock) ResourceSv1Do(*context.Context, any, *string) error      { return nil }
func (TestRPCDspMock) RouteSv1Do(*context.Context, any, *string) error         { return nil }
func (TestRPCDspMock) SessionSv1Do(*context.Context, any, *string) error       { return nil }
func (TestRPCDspMock) StatSv1Do(*context.Context, any, *string) error          { return nil }
func (TestRPCDspMock) ThresholdSv1Do(*context.Context, any, *string) error     { return nil }
func (TestRPCDspMock) CDRsv1Do(*context.Context, any, *string) error           { return nil }
func (TestRPCDspMock) EeSv1Do(*context.Context, any, *string) error            { return nil }
func (TestRPCDspMock) CoreSv1Do(*context.Context, any, *string) error          { return nil }
func (TestRPCDspMock) AnalyzerSv1Do(*context.Context, any, *string) error      { return nil }
func (TestRPCDspMock) AdminSv1Do(*context.Context, any, *string) error         { return nil }
func (TestRPCDspMock) LoaderSv1Do(*context.Context, any, *string) error        { return nil }
func (TestRPCDspMock) ServiceManagerv1Do(*context.Context, any, *string) error { return nil }

func TestNewRPCPoolUnsupportedTransport(t *testing.T) {
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
	badConf := []*config.RemoteHost{
		{
			Address:   "inexistednt Addr",
			Transport: "unsupported transport",
		},
	}
	experr := "Unsupported transport: <unsupported transport>"
	if _, err := NewRPCPool(context.Background(), utils.MetaFirst, "", "", "", cfg.GeneralCfg().ConnectAttempts,
		cfg.GeneralCfg().Reconnects, cfg.GeneralCfg().MaxReconnectInterval, cfg.GeneralCfg().ConnectTimeout,
		cfg.GeneralCfg().ReplyTimeout, badConf, cc, true, nil, "", cM.connCache); err == nil || err.Error() != experr {
		t.Errorf("Expected error <%v>, received error <%v>", experr, err)
	}

}

// func TestRPCClientSetCallOK(t *testing.T) {
// 	tmp := Cache
// 	defer func() {
// 		Cache = tmp
// 	}()
// 	Cache.Clear(nil)

// 	connID := "connID"
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
// 	cfg.RPCConns()[connID].Strategy = rpcclient.PoolParallel
// 	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
// 		{
// 			Address:   rpcclient.InternalRPC, //might need integration
// 			Transport: rpcclient.GOBrpc,
// 		},
// 	}

// 	cc := make(chan birpc.ClientConnector, 1)

// 	cM := &ConnManager{
// 		cfg: cfg,
// 		rpcInternal: map[string]chan birpc.ClientConnector{
// 			connID: cc,
// 		},
// 		connCache: ltcache.NewCache(-1, 0, true, false, nil),
// 	}

// 	cM.connCache.Set(connID, nil, nil)

// 	rpcConnCfg := cfg.RPCConns()[connID].Conns[0]
// 	codec := rpcclient.GOBrpc

// 	conn, err := rpcclient.NewRPCParallelClientPool(context.Background(), utils.TCP, rpcConnCfg.Address, rpcConnCfg.TLS,
// 		utils.FirstNonEmpty(rpcConnCfg.ClientKey, cM.cfg.TLSCfg().ClientKey), utils.FirstNonEmpty(rpcConnCfg.ClientCertificate, cM.cfg.TLSCfg().ClientCerificate),
// 		utils.FirstNonEmpty(rpcConnCfg.CaCertificate, cM.cfg.TLSCfg().CaCertificate), utils.FirstIntNonEmpty(rpcConnCfg.ConnectAttempts, cM.cfg.GeneralCfg().ConnectAttempts),
// 		utils.FirstIntNonEmpty(rpcConnCfg.Reconnects, cM.cfg.GeneralCfg().Reconnects), utils.FirstDurationNonEmpty(rpcConnCfg.MaxReconnectInterval, cM.cfg.GeneralCfg().MaxReconnectInterval),
// 		utils.FibDuration, utils.FirstDurationNonEmpty(rpcConnCfg.ConnectTimeout, cM.cfg.GeneralCfg().ConnectTimeout), utils.FirstDurationNonEmpty(rpcConnCfg.ReplyTimeout, cM.cfg.GeneralCfg().ReplyTimeout),
// 		codec, nil, int64(cM.cfg.GeneralCfg().MaxParallelConns), false, context.Background().Client)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	connChan := make(chan birpc.ClientConnector, 1)
// 	connChan <- conn
// 	s := &RPCClientSet{
// 		utils.SessionSv1: connChan,
// 	}

// 	var rply string
// 	if err := s.Call(context.Background(), utils.SessionSv1RegisterInternalBiJSONConn, connID, &rply); err != nil {
// 		t.Error(err)
// 	}

// }

func TestRPCClientSetCallErrCtxTimeOut(t *testing.T) {
	tmp := Cache
	cfgtmp := config.CgrConfig()

	defer func() {
		Cache = tmp
		config.SetCgrConfig(cfgtmp)

	}()
	Cache.Clear(nil)

	connChan := make(chan birpc.ClientConnector, 1)
	s := &RPCClientSet{
		"test": connChan,
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().ConnectTimeout = 1 * time.Millisecond
	config.SetCgrConfig(cfg)

	var args any
	var reply any
	expErr := "context deadline exceeded"
	if err := s.Call(context.Background(), "test.bad", args, reply); err == nil || expErr != err.Error() {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestRPCClientSetCallErrBadMethod(t *testing.T) {

	connChan := make(chan birpc.ClientConnector, 1)
	s := &RPCClientSet{
		"test": connChan,
	}

	var args any
	var reply any
	expErr := rpcclient.ErrUnsupporteServiceMethod
	if err := s.Call(context.Background(), "bad method", args, reply); err == nil || expErr != err {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestRPCClientSetCallErr2BadMethod(t *testing.T) {

	connChan := make(chan birpc.ClientConnector, 1)
	s := &RPCClientSet{
		"test": connChan,
	}

	var args any
	var reply any
	expErr := rpcclient.ErrUnsupporteServiceMethod
	if err := s.Call(context.Background(), "bad.method", args, reply); err == nil || expErr != err {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}
