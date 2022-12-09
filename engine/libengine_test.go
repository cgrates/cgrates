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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
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
	expectedErr := "dial tcp [::1]:6012: connect: connection refused"
	cM := NewConnManager(config.NewDefaultCGRConfig())
	ctx := context.Background()
	exp, err := rpcclient.NewRPCClient(ctx, utils.TCP, cfg.Address, cfg.TLS, cfg.ClientKey,
		cM.cfg.TLSCfg().ClientCerificate, cM.cfg.TLSCfg().CaCertificate, cfg.ConnectAttempts, cfg.Reconnects,
		cfg.MaxReconnectInterval, utils.FibDuration, cfg.ConnectTimeout, cfg.ReplyTimeout, cfg.Transport, nil, false, nil)

	if err.Error() != expectedErr {
		t.Errorf("Expected %v \n but received \n %v", expectedErr, err)
	}

	conn, err := NewRPCConnection(ctx, cfg, cM.cfg.TLSCfg().ClientKey, cM.cfg.TLSCfg().ClientCerificate,
		cM.cfg.TLSCfg().CaCertificate, cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects,
		cM.cfg.GeneralCfg().MaxReconnectInterval, cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
		nil, false, nil, "*localhost", "a4f3f", new(ltcache.Cache))
	if err.Error() != expectedErr {
		t.Errorf("Expected %v \n but received \n %v", expectedErr, err)
	}
	if !reflect.DeepEqual(exp, conn) {
		//t.Errorf("Expected %v \n but received \n %v", exp, conn)
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

func (TestRPCSrvMock) Do(*context.Context, interface{}, *string) error   { return nil }
func (TestRPCSrvMock) V1Do(*context.Context, interface{}, *string) error { return nil }
func (TestRPCSrvMock) V2Do(*context.Context, interface{}, *string) error { return nil }

type TestRPCSrvMockS struct{} // exported for service

func (TestRPCSrvMockS) V1Do(*context.Context, interface{}, *string) error { return nil }
func (TestRPCSrvMockS) V2Do(*context.Context, interface{}, *string) error { return nil }

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

func (TestRPCDspMock) AccountSv1Do(*context.Context, interface{}, *string) error    { return nil }
func (TestRPCDspMock) ActionSv1Do(*context.Context, interface{}, *string) error     { return nil }
func (TestRPCDspMock) AttributeSv1Do(*context.Context, interface{}, *string) error  { return nil }
func (TestRPCDspMock) CacheSv1Do(*context.Context, interface{}, *string) error      { return nil }
func (TestRPCDspMock) ChargerSv1Do(*context.Context, interface{}, *string) error    { return nil }
func (TestRPCDspMock) ConfigSv1Do(*context.Context, interface{}, *string) error     { return nil }
func (TestRPCDspMock) DispatcherSv1Do(*context.Context, interface{}, *string) error { return nil }
func (TestRPCDspMock) GuardianSv1Do(*context.Context, interface{}, *string) error   { return nil }
func (TestRPCDspMock) RateSv1Do(*context.Context, interface{}, *string) error       { return nil }
func (TestRPCDspMock) ReplicatorSv1Do(*context.Context, interface{}, *string) error { return nil }
func (TestRPCDspMock) ResourceSv1Do(*context.Context, interface{}, *string) error   { return nil }
func (TestRPCDspMock) RouteSv1Do(*context.Context, interface{}, *string) error      { return nil }
func (TestRPCDspMock) SessionSv1Do(*context.Context, interface{}, *string) error    { return nil }
func (TestRPCDspMock) StatSv1Do(*context.Context, interface{}, *string) error       { return nil }
func (TestRPCDspMock) ThresholdSv1Do(*context.Context, interface{}, *string) error  { return nil }
func (TestRPCDspMock) CDRsv1Do(*context.Context, interface{}, *string) error        { return nil }
func (TestRPCDspMock) EeSv1Do(*context.Context, interface{}, *string) error         { return nil }
func (TestRPCDspMock) CoreSv1Do(*context.Context, interface{}, *string) error       { return nil }
func (TestRPCDspMock) AnalyzerSv1Do(*context.Context, interface{}, *string) error   { return nil }
func (TestRPCDspMock) AdminSv1Do(*context.Context, interface{}, *string) error      { return nil }
func (TestRPCDspMock) LoaderSv1Do(*context.Context, interface{}, *string) error     { return nil }

func TestIntServiceNewDispatcherService(t *testing.T) {
	expErrMsg := `rpc.Register: no service name for type struct {}`
	if _, err := NewDispatcherService(struct{}{}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	s, err := NewDispatcherService(new(TestRPCDspMock))
	if err != nil {
		t.Fatal(err)
	}
	methods := getMethods(s)
	exp := map[string][]string{
		"AccountSv1":     {"Do", "Ping"},
		"ActionSv1":      {"Do", "Ping"},
		"AttributeSv1":   {"Do", "Ping"},
		"CDRsV1":         {"Do", "Ping"},
		"CacheSv1":       {"Do", "Ping"},
		"ChargerSv1":     {"Do", "Ping"},
		"ConfigSv1":      {"Do", "Ping"},
		"DispatcherSv1":  {"Do", "Ping"},
		"GuardianSv1":    {"Do", "Ping"},
		"RateSv1":        {"Do", "Ping"},
		"ResourceSv1":    {"Do", "Ping"},
		"RouteSv1":       {"Do", "Ping"},
		"SessionSv1":     {"Do", "Ping"},
		"StatSv1":        {"Do", "Ping"},
		"TestRPCDspMock": {"AccountSv1Do", "ActionSv1Do", "AdminSv1Do", "AnalyzerSv1Do", "AttributeSv1Do", "CDRsv1Do", "CacheSv1Do", "ChargerSv1Do", "ConfigSv1Do", "CoreSv1Do", "DispatcherSv1Do", "EeSv1Do", "GuardianSv1Do", "LoaderSv1Do", "Ping", "RateSv1Do", "ReplicatorSv1Do", "ResourceSv1Do", "RouteSv1Do", "SessionSv1Do", "StatSv1Do", "ThresholdSv1Do"},
		"ThresholdSv1":   {"Do", "Ping"},
		"ReplicatorSv1":  {"Do", "Ping"},

		"EeSv1":       {"Do", "Ping"},
		"CoreSv1":     {"Do", "Ping"},
		"AnalyzerSv1": {"Do", "Ping"},
		"AdminSv1":    {"Do", "Ping"},
		"LoaderSv1":   {"Do", "Ping"},
	}
	if !reflect.DeepEqual(exp, methods) {
		t.Errorf("Expeceted: %v, \nreceived: \n%v", utils.ToJSON(exp), utils.ToJSON(methods))
	}
}
