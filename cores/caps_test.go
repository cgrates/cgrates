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
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
	"syscall"
	"testing"

	"github.com/cenkalti/rpc2"
	jsonrpc2 "github.com/cenkalti/rpc2/jsonrpc"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type mockServerCodec struct{}

func (c *mockServerCodec) ReadRequestHeader(r *rpc.Request) (err error) {
	r.Seq = 0
	r.ServiceMethod = utils.CoreSv1Ping
	return
}

func (c *mockServerCodec) ReadRequestBody(x interface{}) (err error) {
	return utils.ErrNotImplemented
}
func (c *mockServerCodec) WriteResponse(r *rpc.Response, x interface{}) error {
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
	r := new(rpc.Request)
	expR := &rpc.Request{
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

	if err = codec.WriteResponse(&rpc.Response{
		Error:         "error",
		Seq:           0,
		ServiceMethod: utils.CoreSv1Ping,
	}, "reply"); err != nil {
		t.Fatal(err)
	}

	if err = codec.WriteResponse(&rpc.Response{
		Error:         utils.ErrMaxConcurentRPCExceededNoCaps.Error(),
		Seq:           0,
		ServiceMethod: utils.CoreSv1Ping,
	}, "reply"); err != nil {
		t.Fatal(err)
	}
	if err = codec.Close(); err != nil {
		t.Fatal(err)
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
	anz := &analyzers.AnalyzerService{}
	exp := newGobServerCodec(conn)
	if r := newCapsGOBCodec(conn, cr, nil); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
	exp = analyzers.NewAnalyzerServerCodec(newGobServerCodec(conn), anz, utils.MetaGOB, utils.Local, utils.Local)
	if r := newCapsGOBCodec(conn, cr, anz); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
}

func TestNewCapsJSONCodec(t *testing.T) {
	conn := new(mockConn)
	cr := engine.NewCaps(0, utils.MetaBusy)
	anz := &analyzers.AnalyzerService{}
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

func (mockBiRPCCodec) ReadHeader(r *rpc2.Request, _ *rpc2.Response) error {
	r.Seq = 0
	r.Method = utils.CoreSv1Ping
	return nil
}
func (mockBiRPCCodec) ReadRequestBody(interface{}) error               { return utils.ErrNotImplemented }
func (mockBiRPCCodec) ReadResponseBody(interface{}) error              { return nil }
func (mockBiRPCCodec) WriteRequest(*rpc2.Request, interface{}) error   { return nil }
func (mockBiRPCCodec) WriteResponse(*rpc2.Response, interface{}) error { return nil }
func (mockBiRPCCodec) Close() error                                    { return nil }

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
	r := new(rpc2.Request)
	expR := &rpc2.Request{
		Seq:    0,
		Method: utils.CoreSv1Ping,
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

	if err = codec.WriteResponse(&rpc2.Response{
		Error: "error",
		Seq:   0,
	}, "reply"); err != nil {
		t.Fatal(err)
	}

	if err = codec.ReadResponseBody(nil); err != nil {
		t.Fatal(err)
	}

	if err = codec.WriteRequest(&rpc2.Request{
		Seq:    0,
		Method: utils.CoreSv1Ping,
	}, "reply"); err != nil {
		t.Fatal(err)
	}

	if err = codec.WriteResponse(&rpc2.Response{
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
	anz := &analyzers.AnalyzerService{}
	exp := rpc2.NewGobCodec(conn)
	if r := newCapsBiRPCGOBCodec(conn, cr, nil); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
	exp = analyzers.NewAnalyzerBiRPCCodec(rpc2.NewGobCodec(conn), anz, rpcclient.BiRPCGOB, utils.Local, utils.Local)
	if r := newCapsBiRPCGOBCodec(conn, cr, anz); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
}

func TestNewCapsJSONBiRPCCodec(t *testing.T) {
	conn := new(mockConn)
	cr := engine.NewCaps(0, utils.MetaBusy)
	anz := &analyzers.AnalyzerService{}
	exp := jsonrpc2.NewJSONCodec(conn)
	if r := newCapsBiRPCJSONCodec(conn, cr, nil); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
	exp = analyzers.NewAnalyzerBiRPCCodec(jsonrpc2.NewJSONCodec(conn), anz, rpcclient.BiRPCJSON, utils.Local, utils.Local)
	if r := newCapsBiRPCJSONCodec(conn, cr, anz); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
}
