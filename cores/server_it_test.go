//go:build integration
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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"golang.org/x/net/websocket"
)

var (
	server *Server

	sTestsServer = []func(t *testing.T){
		testServeJSON,
		testServeHHTPFail,
		testServeBiJSONInvalidPort,
		testServeBiGoBInvalidPort,
		testLoadTLSConfigErr,
		testHandleRequest,
		testRpcRegisterActions,
		testWebSocket,
	}
)

func TestServerIT(t *testing.T) {
	utils.Logger.SetLogLevel(7)
	for _, test := range sTestsServer {
		log.SetOutput(io.Discard)
		t.Run("TestServerIT", test)
	}
}

type mockRegister struct{}

func (*mockRegister) ForTest(ctx *context.Context, args, reply interface{}) error {
	return nil
}

func (*mockRegister) Ping(ctx *context.Context, in string, out *string) error {
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

func (mkL *mockListener) Close() error { return mkL.p1.Close() }
func (*mockListener) Addr() net.Addr   { return nil }

func testHandleRequest(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = 1
	caps := engine.NewCaps(0, utils.MetaBusy)
	rcv := NewServer(caps)

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
}

func testServeJSON(t *testing.T) {
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	buff := new(bytes.Buffer)
	log.SetOutput(buff)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer server.Stop()
	go server.ServeJSON(ctx, cancel, ":88845")
	runtime.Gosched()

	expected := "listen tcp: address 88845: invalid port"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

}

func testServeHHTPFail(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	var closed bool
	ch := make(chan struct{})
	go server.ServeHTTP(func() {
		closed = true
		close(ch)
	},
		"invalid_port_format",
		cfgDflt.HTTPCfg().JsonRPCURL,
		cfgDflt.HTTPCfg().WSURL,
		cfgDflt.HTTPCfg().UseBasicAuth,
		cfgDflt.HTTPCfg().AuthUsers,
	)

	runtime.Gosched()

	select {
	case <-ch:
		if !closed {
			t.Errorf("Expected to be close")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Timeout")
	}
	server.Stop()
}

func testServeBiJSONInvalidPort(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfgDflt.CacheCfg(), nil)

	ss := sessions.NewSessionS(cfgDflt, dm, nil, nil)

	expectedErr := "listen tcp: address invalid_port_format: missing port in address"
	if err := server.ServeBiRPC("invalid_port_format", "", ss.OnBiJSONConnect,
		ss.OnBiJSONDisconnect); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	server.StopBiRPC()
}

func testServeBiGoBInvalidPort(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(100, utils.MetaBusy)
	server = NewServer(caps)
	server.RpcRegister(new(mockRegister))
	server.birpcSrv = birpc.NewBirpcServer()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfgDflt.CacheCfg(), nil)

	ss := sessions.NewSessionS(cfgDflt, dm, nil, nil)

	expectedErr := "listen tcp: address invalid_port_format: missing port in address"
	if err := server.ServeBiRPC("", "invalid_port_format", ss.OnBiJSONConnect,
		ss.OnBiJSONDisconnect); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	server.StopBiRPC()
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

type mockListenError mockListener

func (*mockListenError) Accept() (net.Conn, error) {
	return nil, errors.New("use of closed network connection")
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

	rpcReq := newRPCRequest(birpc.DefaultServer, r.Body, rmtAddr, server.caps, nil)
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
