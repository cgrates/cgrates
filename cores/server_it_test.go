// +build integration
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

package cores

import (
	"net"
	"runtime"
	"testing"

	sessions2 "github.com/cgrates/cgrates/sessions"

	"github.com/cenkalti/rpc2"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"
)

var (
	server *Server

	sTestsServer = []func(t *testing.T){
		testServeJSONPass,
		testServeJSONFail,
		testServeJSONFailRpcEnabled,
		testServeGOBPass,
		testServeHHTPPass,
		testServeHHTPPassUseBasicAuth,
		testServeHHTPEnableHttp,
		testServeHHTPFail,
		testServeHHTPFailEnableRpc,
		testServeBiJSON,
		testServeBiJSONEmptyBiRPCServer,
		testServeBiJSONInvalidPort,
	}
)

func TestServerIT(t *testing.T) {
	for _, test := range sTestsServer {
		t.Run("Running IT serve tests", test)
	}
}

type mockRegister struct{}

func (robj *mockRegister) Ping(in string, out *string) error {
	*out = utils.Pong
	return nil
}

type mockListener struct{}

func (mkL *mockListener) Accept() (net.Conn, error) { return nil, utils.ErrDisconnected }
func (mkL *mockListener) Close() error              { return nil }
func (mkL *mockListener) Addr() net.Addr            { return nil }

func testServeJSONPass(t *testing.T) {
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	go server.ServeJSON(":2016", shdChan)
	runtime.Gosched()

	shdChan.CloseOnce()
}

func testServeJSONFail(t *testing.T) {
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	l := &mockListener{}
	go server.accept(l, utils.JSONCaps, newCapsJSONCodec, shdChan)
	runtime.Gosched()
	_, ok := <-shdChan.Done()
	if ok {
		t.Errorf("Expected to be close")
	}
}

func testServeJSONFailRpcEnabled(t *testing.T) {
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()
	server.rpcEnabled = false

	go server.serveCodec(":9999", utils.JSONCaps, newCapsJSONCodec, shdChan)

	shdChan.CloseOnce()
}

func testServeGOBPass(t *testing.T) {
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	go server.ServeGOB(":2019", shdChan)
	runtime.Gosched()

	shdChan.CloseOnce()
}

func testServeHHTPPass(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	go server.ServeHTTP(
		":2021",
		cfgDflt.HTTPCfg().HTTPJsonRPCURL,
		cfgDflt.HTTPCfg().HTTPWSURL,
		cfgDflt.HTTPCfg().HTTPUseBasicAuth,
		cfgDflt.HTTPCfg().HTTPAuthUsers,
		shdChan)

	runtime.Gosched()

	shdChan.CloseOnce()
}

func testServeHHTPPassUseBasicAuth(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	go server.ServeHTTP(
		":2075",
		cfgDflt.HTTPCfg().HTTPJsonRPCURL,
		cfgDflt.HTTPCfg().HTTPWSURL,
		!cfgDflt.HTTPCfg().HTTPUseBasicAuth,
		cfgDflt.HTTPCfg().HTTPAuthUsers,
		shdChan)

	runtime.Gosched()

	shdChan.CloseOnce()
}

func testServeHHTPEnableHttp(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	go server.ServeHTTP(
		":2077",
		utils.EmptyString,
		utils.EmptyString,
		!cfgDflt.HTTPCfg().HTTPUseBasicAuth,
		cfgDflt.HTTPCfg().HTTPAuthUsers,
		shdChan)

	runtime.Gosched()

	shdChan.CloseOnce()
}

func testServeHHTPFail(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	go server.ServeHTTP(
		"invalid_portt_format",
		cfgDflt.HTTPCfg().HTTPJsonRPCURL,
		cfgDflt.HTTPCfg().HTTPWSURL,
		cfgDflt.HTTPCfg().HTTPUseBasicAuth,
		cfgDflt.HTTPCfg().HTTPAuthUsers,
		shdChan)

	runtime.Gosched()

	_, ok := <-shdChan.Done()
	if ok {
		t.Errorf("Expected to be close")
	}
}

func testServeHHTPFailEnableRpc(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()
	server.rpcEnabled = false

	go server.ServeHTTP(":1000",
		cfgDflt.HTTPCfg().HTTPJsonRPCURL,
		cfgDflt.HTTPCfg().HTTPWSURL,
		cfgDflt.HTTPCfg().HTTPUseBasicAuth,
		cfgDflt.HTTPCfg().HTTPAuthUsers,
		shdChan)

	shdChan.CloseOnce()
}

func testServeBiJSON(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	server.birpcSrv = rpc2.NewServer()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfgDflt.CacheCfg(), nil)

	sessions := sessions2.NewSessionS(cfgDflt, dm, nil)

	go func() {
		if err := server.ServeBiJSON(":3434", sessions.OnBiJSONConnect, sessions.OnBiJSONDisconnect); err != nil {
			t.Error(err)
		}
	}()

	runtime.Gosched()
}

func testServeBiJSONEmptyBiRPCServer(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfgDflt.CacheCfg(), nil)

	sessions := sessions2.NewSessionS(cfgDflt, dm, nil)

	expectedErr := "BiRPCServer should not be nil"
	go func() {

		if err := server.ServeBiJSON(":3430", sessions.OnBiJSONConnect, sessions.OnBiJSONDisconnect); err == nil || err.Error() != "BiRPCServer should not be nil" {
			t.Errorf("Expected %+v, received %+v", expectedErr, err)
		}
	}()

	runtime.Gosched()
}

func testServeBiJSONInvalidPort(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	server.birpcSrv = rpc2.NewServer()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfgDflt.CacheCfg(), nil)

	sessions := sessions2.NewSessionS(cfgDflt, dm, nil)

	expectedErr := "listen tcp: address invalid_port_format: missing port in address"
	if err := server.ServeBiJSON("invalid_port_format", sessions.OnBiJSONConnect,
		sessions.OnBiJSONDisconnect); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

/*
func testServeGOBTLS(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(randomObj))

	shdChan := utils.NewSyncedChan()

	go server.ServeGOBTLS(
		":1256",
		cfgDflt.TLSCfg().ServerCerificate,
		cfgDflt.TLSCfg().ServerKey,
		cfgDflt.TLSCfg().CaCertificate,
		cfgDflt.TLSCfg().ServerPolicy,
		cfgDflt.TLSCfg().ServerName,
		shdChan,
	)


	shdChan.CloseOnce()
}

*/
