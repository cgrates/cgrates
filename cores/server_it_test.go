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
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	sessions2 "github.com/cgrates/cgrates/sessions"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"
)

var (
	server *Server

	sTestsServer = []func(t *testing.T){
		testServeGOBPortFail,
		testServeJSON,
		testServeJSONFail,
		testServeJSONFailRpcEnabled,
		testServeGOB,
		testServeHHTPPass,
		testServeHHTPPassUseBasicAuth,
		testServeHHTPEnableHttp,
		testServeHHTPFail,
		testServeHHTPFailEnableRpc,
		testServeBiJSON,
		testServeBiJSONEmptyBiRPCServer,
		testServeBiJSONInvalidPort,
		testServeBiGoB,
		testServeBiGoBEmptyBiRPCServer,
		testServeBiGoBInvalidPort,
		testServeGOBTLS,
		testServeJSONTls,
		testServeCodecTLSErr,
		testLoadTLSConfigErr,
		testServeHTTPTLS,
		testServeHTTPTLSWithBasicAuth,
		testServeHTTPTLSError,
		testServeHTTPTLSHttpNotEnabled,
		testHandleRequest,
		testBiRPCRegisterName,
		testAcceptBiRPC,
		testAcceptBiRPCError,
		testRpcRegisterActions,
		testWebSocket,
	}
)

func TestServerIT(t *testing.T) {
	utils.Logger.SetLogLevel(7)
	for _, test := range sTestsServer {
		log.SetOutput(io.Discard)
		t.Run("Running IT serve tests", test)
	}
}

type mockRegister string

func (x *mockRegister) ForTest(ctx *context.Context, args, reply interface{}) error {
	return nil
}

func (robj *mockRegister) Ping(ctx *context.Context, in string, out *string) error {
	*out = utils.Pong
	return nil
}

type mockListener struct {
	p1   net.Conn
	call bool
}

func (mkL *mockListener) Accept() (net.Conn, error) {
	if !mkL.call {
		mkL.call = true
		return mkL.p1, nil
	}
	return nil, utils.ErrDisconnected
}

func (mkL *mockListener) Close() error   { return mkL.p1.Close() }
func (mkL *mockListener) Addr() net.Addr { return nil }

func testHandleRequest(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = 1
	caps := engine.NewCaps(0, utils.MetaBusy)
	rcv := NewServer(caps)

	rcv.rpcEnabled = true

	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/json_rpc",
		bytes.NewBuffer([]byte("1")))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	rcv.handleRequest(w, req)
	if w.Body.String() != utils.EmptyString {
		t.Errorf("Expected: %q ,received: %q", utils.EmptyString, w.Body.String())
	}

	rcv.StopBiRPC()
}

func testServeJSON(t *testing.T) {
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	buff := new(bytes.Buffer)
	log.SetOutput(buff)

	go server.ServeJSON(":88845", shdChan)
	runtime.Gosched()

	expected := "listen tcp: address 88845: invalid port"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	shdChan.CloseOnce()
	server.StopBiRPC()
}

func testServeJSONFail(t *testing.T) {
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	p1, p2 := net.Pipe()
	l := &mockListener{
		p1: p1,
	}
	go server.accept(l, utils.JSONCaps, newCapsJSONCodec, shdChan)
	runtime.Gosched()
	_, ok := <-shdChan.Done()
	if ok {
		t.Errorf("Expected to be close")
	}
	p2.Close()
	runtime.Gosched()

	shdChan.CloseOnce()
	server.StopBiRPC()
}

func testServeJSONFailRpcEnabled(t *testing.T) {
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()
	server.rpcEnabled = false

	go server.serveCodec(":9999", utils.JSONCaps, newCapsJSONCodec, shdChan)
	runtime.Gosched()

	shdChan.CloseOnce()
	server.StopBiRPC()
}

func testServeGOB(t *testing.T) {
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	go server.ServeGOB(":27697", shdChan)
	runtime.Gosched()

	shdChan.CloseOnce()
	server.StopBiRPC()
}

func testServeHHTPPass(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	go server.ServeHTTP(
		":6555",
		cfgDflt.HTTPCfg().JsonRPCURL,
		cfgDflt.HTTPCfg().WSURL,
		cfgDflt.HTTPCfg().UseBasicAuth,
		cfgDflt.HTTPCfg().AuthUsers,
		shdChan)

	runtime.Gosched()

	shdChan.CloseOnce()
	server.StopBiRPC()
}

func testServeHHTPPassUseBasicAuth(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	go server.ServeHTTP(
		":56432",
		cfgDflt.HTTPCfg().JsonRPCURL,
		cfgDflt.HTTPCfg().WSURL,
		!cfgDflt.HTTPCfg().UseBasicAuth,
		cfgDflt.HTTPCfg().AuthUsers,
		shdChan)

	runtime.Gosched()

	shdChan.CloseOnce()
	server.StopBiRPC()
}

func testServeHHTPEnableHttp(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	go server.ServeHTTP(
		":45779",
		utils.EmptyString,
		utils.EmptyString,
		!cfgDflt.HTTPCfg().UseBasicAuth,
		cfgDflt.HTTPCfg().AuthUsers,
		shdChan)

	runtime.Gosched()

	shdChan.CloseOnce()
	server.StopBiRPC()
}

func testServeHHTPFail(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()

	go server.ServeHTTP(
		"invalid_port_format",
		cfgDflt.HTTPCfg().JsonRPCURL,
		cfgDflt.HTTPCfg().WSURL,
		cfgDflt.HTTPCfg().UseBasicAuth,
		cfgDflt.HTTPCfg().AuthUsers,
		shdChan)

	runtime.Gosched()

	_, ok := <-shdChan.Done()
	if ok {
		t.Errorf("Expected to be close")
	}
	server.StopBiRPC()
}

func testServeHHTPFailEnableRpc(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	shdChan := utils.NewSyncedChan()
	server.rpcEnabled = false

	go server.ServeHTTP(":1000",
		cfgDflt.HTTPCfg().JsonRPCURL,
		cfgDflt.HTTPCfg().WSURL,
		cfgDflt.HTTPCfg().UseBasicAuth,
		cfgDflt.HTTPCfg().AuthUsers,
		shdChan)

	shdChan.CloseOnce()
	server.StopBiRPC()
}

func testServeBiJSON(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	server.birpcSrv = birpc.NewBirpcServer()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfgDflt.CacheCfg(), nil)

	sessions := sessions2.NewSessionS(cfgDflt, dm, nil)

	go func() {
		if err := server.ServeBiRPC(":3434", "", sessions.OnBiJSONConnect, sessions.OnBiJSONDisconnect); err != nil {
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
		if err := server.ServeBiRPC(":3430", "", sessions.OnBiJSONConnect, sessions.OnBiJSONDisconnect); err == nil || err.Error() != "BiRPCServer should not be nil" {
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
	server.birpcSrv = birpc.NewBirpcServer()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfgDflt.CacheCfg(), nil)

	sessions := sessions2.NewSessionS(cfgDflt, dm, nil)

	expectedErr := "listen tcp: address invalid_port_format: missing port in address"
	if err := server.ServeBiRPC("invalid_port_format", "", sessions.OnBiJSONConnect,
		sessions.OnBiJSONDisconnect); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	server.StopBiRPC()
}

func testServeBiGoB(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	server.birpcSrv = birpc.NewBirpcServer()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfgDflt.CacheCfg(), nil)

	sessions := sessions2.NewSessionS(cfgDflt, dm, nil)

	go func() {
		if err := server.ServeBiRPC("", ":9343", sessions.OnBiJSONConnect, sessions.OnBiJSONDisconnect); err != nil {
			t.Error(err)
		}
	}()
	runtime.Gosched()
}

func testServeBiGoBEmptyBiRPCServer(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfgDflt.CacheCfg(), nil)

	sessions := sessions2.NewSessionS(cfgDflt, dm, nil)

	expectedErr := "BiRPCServer should not be nil"
	go func() {
		if err := server.ServeBiRPC("", ":93430", sessions.OnBiJSONConnect, sessions.OnBiJSONDisconnect); err == nil || err.Error() != "BiRPCServer should not be nil" {
			t.Errorf("Expected %+v, received %+v", expectedErr, err)
		}
	}()

	runtime.Gosched()
}

func testServeBiGoBInvalidPort(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	server.birpcSrv = birpc.NewBirpcServer()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfgDflt.CacheCfg(), nil)

	sessions := sessions2.NewSessionS(cfgDflt, dm, nil)

	expectedErr := "listen tcp: address invalid_port_format: missing port in address"
	if err := server.ServeBiRPC("", "invalid_port_format", sessions.OnBiJSONConnect,
		sessions.OnBiJSONDisconnect); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	server.StopBiRPC()
}

func testServeGOBTLS(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	shdChan := utils.NewSyncedChan()

	go server.ServeGOBTLS(
		":34476",
		"/usr/share/cgrates/tls/server.crt",
		"/usr/share/cgrates/tls/server.key",
		"/usr/share/cgrates/tls/ca.crt",
		4,
		cfgDflt.TLSCfg().ServerName,
		shdChan,
	)
	runtime.Gosched()

	server.StopBiRPC()
}

func testServeJSONTls(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	shdChan := utils.NewSyncedChan()

	go server.ServeJSONTLS(
		":64779",
		"/usr/share/cgrates/tls/server.crt",
		"/usr/share/cgrates/tls/server.key",
		"/usr/share/cgrates/tls/ca.crt",
		4,
		cfgDflt.TLSCfg().ServerName,
		shdChan,
	)
	runtime.Gosched()
}

func testServeGOBPortFail(t *testing.T) {
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	shdChan := utils.NewSyncedChan()

	buff := new(bytes.Buffer)
	log.SetOutput(buff)

	go server.serveCodecTLS(
		"34776",
		utils.GOBCaps,
		"/usr/share/cgrates/tls/server.crt",
		"/usr/share/cgrates/tls/server.key",
		"/usr/share/cgrates/tls/ca.crt",
		4,
		"Server_name",
		newCapsGOBCodec,
		shdChan,
	)
	runtime.Gosched()
	select {
	case <-time.After(10 * time.Second):
		t.Fatal("timeout")
	case <-shdChan.Done():
	}
	expected := "listen tcp: address 34776: missing port in address when listening"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	log.SetOutput(os.Stderr)
}

func testServeCodecTLSErr(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	shdChan := utils.NewSyncedChan()

	//if rpc is not enabled, won t be able to serve
	server.rpcEnabled = false
	server.serveCodecTLS("13567",
		utils.GOBCaps,
		"/usr/share/cgrates/tls/server.crt",
		"/usr/share/cgrates/tls/server.key",
		"/usr/share/cgrates/tls/ca.crt",
		4,
		cfgDflt.TLSCfg().ServerName,
		newCapsGOBCodec,
		shdChan)

	//unable to load TLS config when there is an inexisting server certificate file
	server.rpcEnabled = true
	server.serveCodecTLS("13567",
		utils.GOBCaps,
		"/usr/share/cgrates/tls/inexisting_cert",
		"/usr/share/cgrates/tls/server.key",
		"/usr/share/cgrates/tls/ca.crt",
		4,
		cfgDflt.TLSCfg().ServerName,
		newCapsGOBCodec,
		shdChan)

	_, ok := <-shdChan.Done()
	if ok {
		t.Errorf("Expected to be close")
	}
}

func testLoadTLSConfigErr(t *testing.T) {
	flPath := "/tmp/testLoadTLSConfigErr1"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(flPath, "file.txt"))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`
TEST
`))
	file.Close()

	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	expectedErr := "Cannot append certificate authority"
	if _, err := loadTLSConfig(
		"/usr/share/cgrates/tls/server.crt",
		"/usr/share/cgrates/tls/server.key",
		path.Join(flPath, "file.txt"),
		0,
		utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	expectedErr = "open /tmp/testLoadTLSConfigErr1/file1.txt: no such file or directory"
	if _, err := loadTLSConfig(
		"/usr/share/cgrates/tls/server.crt",
		"/usr/share/cgrates/tls/server.key",
		path.Join(flPath, "file1.txt"),
		0,
		utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
	if err := os.Remove(path.Join(flPath, "file.txt")); err != nil {
		t.Error(err)
	} else if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testServeHTTPTLS(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	shdChan := utils.NewSyncedChan()

	//cannot serve HHTPTls when rpc is not enabled
	server.rpcEnabled = false
	server.ServeHTTPTLS(
		"17789",
		"/usr/share/cgrates/tls/server.crt",
		"/usr/share/cgrates/tls/server.key",
		"/usr/share/cgrates/tls/ca.crt",
		cfgDflt.TLSCfg().ServerPolicy,
		cfgDflt.TLSCfg().ServerName,
		cfgDflt.HTTPCfg().JsonRPCURL,
		cfgDflt.HTTPCfg().WSURL,
		cfgDflt.HTTPCfg().UseBasicAuth,
		cfgDflt.HTTPCfg().AuthUsers,
		shdChan)

	//Invalid port address
	server.rpcEnabled = true
	go server.ServeHTTPTLS(
		"17789",
		"/usr/share/cgrates/tls/server.crt",
		"/usr/share/cgrates/tls/server.key",
		"/usr/share/cgrates/tls/ca.crt",
		cfgDflt.TLSCfg().ServerPolicy,
		cfgDflt.TLSCfg().ServerName,
		cfgDflt.HTTPCfg().JsonRPCURL,
		cfgDflt.HTTPCfg().WSURL,
		cfgDflt.HTTPCfg().UseBasicAuth,
		cfgDflt.HTTPCfg().AuthUsers,
		shdChan)
	runtime.Gosched()

	_, ok := <-shdChan.Done()
	if ok {
		t.Errorf("Expected to be close")
	}
}

func testServeHTTPTLSWithBasicAuth(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	shdChan := utils.NewSyncedChan()

	//Invalid port address
	server.rpcEnabled = true
	go server.ServeHTTPTLS(
		"57235",
		"/usr/share/cgrates/tls/server.crt",
		"/usr/share/cgrates/tls/server.key",
		"/usr/share/cgrates/tls/ca.crt",
		cfgDflt.TLSCfg().ServerPolicy,
		cfgDflt.TLSCfg().ServerName,
		cfgDflt.HTTPCfg().JsonRPCURL,
		cfgDflt.HTTPCfg().WSURL,
		!cfgDflt.HTTPCfg().UseBasicAuth,
		cfgDflt.HTTPCfg().AuthUsers,
		shdChan)
	runtime.Gosched()

	_, ok := <-shdChan.Done()
	if ok {
		t.Errorf("Expected to be close")
	}
}

func testServeHTTPTLSError(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	shdChan := utils.NewSyncedChan()

	//Invalid port address
	go server.ServeHTTPTLS(
		"57235",
		"/usr/share/cgrates/tls/inexisting_file",
		"/usr/share/cgrates/tls/server.key",
		"/usr/share/cgrates/tls/ca.crt",
		cfgDflt.TLSCfg().ServerPolicy,
		cfgDflt.TLSCfg().ServerName,
		cfgDflt.HTTPCfg().JsonRPCURL,
		cfgDflt.HTTPCfg().WSURL,
		!cfgDflt.HTTPCfg().UseBasicAuth,
		cfgDflt.HTTPCfg().AuthUsers,
		shdChan)
	runtime.Gosched()

	_, ok := <-shdChan.Done()
	if ok {
		t.Errorf("Expected to be close")
	}
}

func testServeHTTPTLSHttpNotEnabled(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	shdChan := utils.NewSyncedChan()

	server.httpEnabled = false
	go server.ServeHTTPTLS(
		"17789",
		"/usr/share/cgrates/tls/server.crt",
		"/usr/share/cgrates/tls/server.key",
		"/usr/share/cgrates/tls/ca.crt",
		cfgDflt.TLSCfg().ServerPolicy,
		cfgDflt.TLSCfg().ServerName,
		utils.EmptyString,
		utils.EmptyString,
		cfgDflt.HTTPCfg().UseBasicAuth,
		cfgDflt.HTTPCfg().AuthUsers,
		shdChan)

	shdChan.CloseOnce()
}

func testBiRPCRegisterName(t *testing.T) {
	caps := engine.NewCaps(0, utils.MetaBusy)
	server := NewServer(caps)

	handler := struct{}{}
	go server.BiRPCRegisterName(utils.APIerSv1Ping, handler)
	runtime.Gosched()

	server.StopBiRPC()
}

func testAcceptBiRPC(t *testing.T) {
	caps := engine.NewCaps(0, utils.MetaBusy)
	server := NewServer(caps)
	server.RpcRegister(new(mockRegister))
	server.birpcSrv = birpc.NewBirpcServer()

	p1, p2 := net.Pipe()
	l := &mockListener{
		p1: p1,
	}
	go server.acceptBiRPC(server.birpcSrv, l, utils.JSONCaps, jsonrpc.NewJSONBirpcCodec)
	rpc := jsonrpc.NewClient(p2)
	var reply string
	expected := "birpc: can't find method AttributeSv1.Ping"
	if err := rpc.Call(context.TODO(), utils.AttributeSv1Ping, utils.CGREvent{}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	p2.Close()
	runtime.Gosched()
}

type mockListenError struct {
	*mockListener
}

func (mK *mockListenError) Accept() (net.Conn, error) {
	return nil, errors.New("use of closed network connection")
}

func testAcceptBiRPCError(t *testing.T) {
	caps := engine.NewCaps(0, utils.MetaBusy)
	server := NewServer(caps)
	server.RpcRegister(new(mockRegister))
	server.birpcSrv = birpc.NewBirpcServer()

	//it will contain "use of closed network connection"
	l := new(mockListenError)
	go server.acceptBiRPC(server.birpcSrv, l, utils.JSONCaps, jsonrpc.NewJSONBirpcCodec)
	runtime.Gosched()
}

func testRpcRegisterActions(t *testing.T) {
	caps := engine.NewCaps(0, utils.MetaBusy)
	server := NewServer(caps)

	r, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/json_rpc",
		bytes.NewBuffer([]byte("1")))
	if err != nil {
		t.Fatal(err)
	}
	rmtIP, _ := utils.GetRemoteIP(r)
	rmtAddr, _ := net.ResolveIPAddr(utils.EmptyString, rmtIP)

	rpcReq := newRPCRequest(r.Body, rmtAddr, server.caps, nil)
	rpcReq.remoteAddr = utils.NewNetAddr("network", "127.0.0.1:2012")

	if n, err := rpcReq.Write([]byte(`TEST`)); err != nil {
		t.Error(err)
	} else if n != 4 {
		t.Errorf("Expected 4, received %+v", n)
	}

	if rcv := rpcReq.LocalAddr(); !reflect.DeepEqual(rcv, utils.LocalAddr()) {
		t.Errorf("Received %+v, expected %+v", utils.ToJSON(rcv), utils.ToJSON(utils.LocalAddr()))
	}

	exp := utils.NewNetAddr("network", "127.0.0.1:2012")
	if rcv := rpcReq.RemoteAddr(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Received %+v, expected %+v", utils.ToJSON(rcv), utils.ToJSON(exp))
	}
}

func testWebSocket(t *testing.T) {
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegisterName("mockRegister", new(mockRegister))

	s := httptest.NewServer(websocket.Handler(server.handleWebSocket))
	config, err := websocket.NewConfig(fmt.Sprintf("ws://%s", s.Listener.Addr().String()), "http://localhost")
	if err != nil {
		t.Fatal(err)
	}

	c1, err := net.Dial(utils.TCP, s.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	conn1, err := websocket.NewClient(config, c1)
	if err != nil {
		t.Fatal(err)
	}

	rpc := jsonrpc.NewClient(conn1)
	var reply string
	err = rpc.Call(context.TODO(), "mockRegister.Ping", "", &reply)
	if err != nil {
		t.Fatal(err)
	}
	if reply != utils.Pong {
		t.Errorf("Expected Pong, receive %+s", reply)
	}

	conn1.Close()

	s.Close()
}
