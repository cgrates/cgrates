// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package commonlisteners

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"syscall"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type mockServerCodec struct{}

func (c *mockServerCodec) ReadRequestHeader(r *birpc.Request) (err error) {
	r.Seq = 0
	r.ServiceMethod = utils.CoreSv1Ping
	return
}

func (c *mockServerCodec) ReadRequestBody(x any) (err error) {
	return utils.ErrNotImplemented
}
func (c *mockServerCodec) WriteResponse(r *birpc.Response, x any) error {
	return nil
}
func (c *mockServerCodec) Close() error { return nil }

func TestNewCapsServerCodec(t *testing.T) {
	mk := new(mockServerCodec)
	cr := engine.NewCaps(0, utils.MetaBusy)
	if r := newCapsServerCodec(mk, cr); !reflect.DeepEqual(mk, r) {
		t.Errorf("Expected: %v ,received:%v", mk, r)
	}
	cr = engine.NewCaps(1, utils.MetaBusy)
	exp := &capsServerCodec{
		sc:   mk,
		caps: cr,
	}
	codec := newCapsServerCodec(mk, cr)
	if !reflect.DeepEqual(exp, codec) {
		t.Errorf("Expected: %v ,received:%v", exp, codec)
	}
	var err error
	r := new(birpc.Request)
	expR := &birpc.Request{
		Seq:           0,
		ServiceMethod: utils.CoreSv1Ping,
	}
	if err = codec.ReadRequestHeader(r); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %v ,received:%v", expR, r)
	}

	if err = codec.ReadRequestBody("args"); err == nil || err != utils.ErrNotImplemented {
		t.Fatal(err)
	}

	if err = codec.ReadRequestBody("args"); err != utils.ErrMaxConcurentRPCExceededNoCaps {
		t.Errorf("Expected error: %v ,received: %v ", utils.ErrMaxConcurentRPCExceededNoCaps, err)
	}

	if err = codec.WriteResponse(&birpc.Response{
		Error: "error",
		Seq:   0,
	}, "reply"); err != nil {
		t.Fatal(err)
	}

	if err = codec.WriteResponse(&birpc.Response{
		Error: utils.ErrMaxConcurentRPCExceededNoCaps.Error(),
		Seq:   0,
	}, "reply"); err != nil {
		t.Fatal(err)
	}
	if err = codec.Close(); err != nil {
		t.Fatal(err)
	}
}

type capsCompletionService struct{}

func (*capsCompletionService) CapsError(_ context.Context, _ string, _ *string) error {
	return utils.ErrMaxConcurentRPCExceededNoCaps
}

func TestCapsHTTPRequestCompletion(t *testing.T) {
	for _, tc := range []struct {
		name      string
		request   string
		full      bool
		wantError string
	}{
		{
			name:      "unknown method while full",
			request:   `{"method":"Missing.Call","params":["x"],"id":1}`,
			full:      true,
			wantError: "can't find service",
		},
		{
			name:      "handler caps error",
			request:   `{"method":"Test.CapsError","params":["x"],"id":1}`,
			wantError: utils.ErrMaxConcurentRPCExceeded.Error(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			caps := engine.NewCaps(1, utils.MetaBusy)
			server := NewCommonListenerS(caps)
			server.RpcRegisterName("Test", new(capsCompletionService))
			wantAllocated := 0
			if tc.full {
				if err := caps.Allocate(); err != nil {
					t.Fatal(err)
				}
				wantAllocated = 1
			}

			request := httptest.NewRequest(http.MethodPost, "/jsonrpc", strings.NewReader(tc.request))
			response := httptest.NewRecorder()
			server.handleRequest(response, request)
			var reply struct{ Error string }
			if err := json.Unmarshal(response.Body.Bytes(), &reply); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(reply.Error, tc.wantError) {
				t.Errorf("expected error containing %q, got %q", tc.wantError, reply.Error)
			}
			if allocated := caps.Allocated(); allocated != wantAllocated {
				t.Errorf("expected %d allocated caps, got %d", wantAllocated, allocated)
			}
		})
	}
}

func TestCapsBiRPCRequestCompletion(t *testing.T) {
	caps := engine.NewCaps(1, utils.MetaBusy)
	server := birpc.NewBirpcServer()
	if err := server.RegisterName("Test", new(capsCompletionService)); err != nil {
		t.Fatal(err)
	}
	serverConn, clientConn := net.Pipe()
	done := make(chan struct{})
	go func() {
		server.ServeCodec(newCapsBiRPCJSONCodec(serverConn, caps, nil))
		close(done)
	}()
	t.Cleanup(func() {
		_ = clientConn.Close()
		<-done
	})
	if _, err := io.WriteString(clientConn, `{"method":"Test.CapsError","params":["x"],"id":1}`); err != nil {
		t.Fatal(err)
	}
	var reply struct{ Error string }
	if err := json.NewDecoder(clientConn).Decode(&reply); err != nil {
		t.Fatal(err)
	}
	_ = clientConn.Close()
	<-done
	if reply.Error != utils.ErrMaxConcurentRPCExceeded.Error() {
		t.Errorf("unexpected error: %q", reply.Error)
	}
	if allocated := caps.Allocated(); allocated != 0 {
		t.Errorf("expected no allocated caps, got %d", allocated)
	}
}

type mockConn struct{}

func (*mockConn) Read(b []byte) (n int, err error)  { return 0, syscall.EINVAL }
func (*mockConn) Write(b []byte) (n int, err error) { return }
func (*mockConn) Close() error                      { return nil }
func (*mockConn) LocalAddr() net.Addr               { return utils.LocalAddr() }
func (*mockConn) RemoteAddr() net.Addr              { return utils.LocalAddr() }

func TestNewCapsGOBCodec(t *testing.T) {
	conn := new(mockConn)
	cr := engine.NewCaps(0, utils.MetaBusy)
	anz := &analyzers.AnalyzerS{}
	exp := birpc.NewServerCodec(conn)
	if r := newCapsGOBCodec(conn, cr, nil); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
	exp = analyzers.NewAnalyzerServerCodec(birpc.NewServerCodec(conn), anz, utils.MetaGOB, utils.Local, utils.Local)
	if r := newCapsGOBCodec(conn, cr, anz); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
}

func TestNewCapsJSONCodec(t *testing.T) {
	conn := new(mockConn)
	cr := engine.NewCaps(0, utils.MetaBusy)
	anz := &analyzers.AnalyzerS{}
	exp := jsonrpc.NewServerCodec(conn)
	if r := newCapsJSONCodec(conn, cr, nil); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
	exp = analyzers.NewAnalyzerServerCodec(jsonrpc.NewServerCodec(conn), anz, utils.MetaJSON, utils.Local, utils.Local)
	if r := newCapsJSONCodec(conn, cr, anz); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
}

type mockBiRPCCodec struct{}

func (mockBiRPCCodec) ReadHeader(r *birpc.Request, _ *birpc.Response) error {
	r.Seq = 0
	r.ServiceMethod = utils.CoreSv1Ping
	return nil
}
func (mockBiRPCCodec) ReadRequestBody(any) error                { return utils.ErrNotImplemented }
func (mockBiRPCCodec) ReadResponseBody(any) error               { return nil }
func (mockBiRPCCodec) WriteRequest(*birpc.Request, any) error   { return nil }
func (mockBiRPCCodec) WriteResponse(*birpc.Response, any) error { return nil }
func (mockBiRPCCodec) Close() error                             { return nil }

func TestNewCapsBiRPCCodec(t *testing.T) {
	mk := new(mockBiRPCCodec)
	cr := engine.NewCaps(0, utils.MetaBusy)
	if r := newCapsBiRPCCodec(mk, cr); !reflect.DeepEqual(mk, r) {
		t.Errorf("Expected: %v ,received:%v", mk, r)
	}
	cr = engine.NewCaps(1, utils.MetaBusy)
	exp := &capsBiRPCCodec{
		sc:   mk,
		caps: cr,
	}
	codec := newCapsBiRPCCodec(mk, cr)
	if !reflect.DeepEqual(exp, codec) {
		t.Errorf("Expected: %v ,received:%v", exp, codec)
	}
	var err error
	r := new(birpc.Request)
	expR := &birpc.Request{
		Seq:           0,
		ServiceMethod: utils.CoreSv1Ping,
	}
	if err = codec.ReadHeader(r, nil); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %v ,received:%v", expR, r)
	}

	if err = codec.ReadRequestBody("args"); err == nil || err != utils.ErrNotImplemented {
		t.Fatal(err)
	}

	if err = codec.ReadRequestBody("args"); err != utils.ErrMaxConcurentRPCExceededNoCaps {
		t.Errorf("Expected error: %v ,received: %v ", utils.ErrMaxConcurentRPCExceededNoCaps, err)
	}

	if err = codec.WriteResponse(&birpc.Response{
		Error: "error",
		Seq:   0,
	}, "reply"); err != nil {
		t.Fatal(err)
	}

	if err = codec.ReadResponseBody(nil); err != nil {
		t.Fatal(err)
	}

	if err = codec.WriteRequest(&birpc.Request{
		Seq:           0,
		ServiceMethod: utils.CoreSv1Ping,
	}, "reply"); err != nil {
		t.Fatal(err)
	}

	if err = codec.WriteResponse(&birpc.Response{
		Error: utils.ErrMaxConcurentRPCExceededNoCaps.Error(),
		Seq:   0,
	}, "reply"); err != nil {
		t.Fatal(err)
	}
	if err = codec.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewCapsGOBBiRPCCodec(t *testing.T) {
	conn := new(mockConn)
	cr := engine.NewCaps(0, utils.MetaBusy)
	anz := &analyzers.AnalyzerS{}
	exp := birpc.NewGobBirpcCodec(conn)
	if r := newCapsBiRPCGOBCodec(conn, cr, nil); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
	exp = analyzers.NewAnalyzerBiRPCCodec(birpc.NewGobBirpcCodec(conn), anz, rpcclient.BiRPCGOB, utils.Local, utils.Local)
	if r := newCapsBiRPCGOBCodec(conn, cr, anz); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
}

func TestNewCapsJSONBiRPCCodec(t *testing.T) {
	conn := new(mockConn)
	cr := engine.NewCaps(0, utils.MetaBusy)
	anz := &analyzers.AnalyzerS{}
	exp := jsonrpc.NewJSONBirpcCodec(conn)
	if r := newCapsBiRPCJSONCodec(conn, cr, nil); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
	exp = analyzers.NewAnalyzerBiRPCCodec(jsonrpc.NewJSONBirpcCodec(conn), anz, rpcclient.BiRPCJSON, utils.Local, utils.Local)
	if r := newCapsBiRPCJSONCodec(conn, cr, anz); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
}
