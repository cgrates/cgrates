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
	"runtime"
	"testing"
	"time"

	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewCaps(t *testing.T) {
	exp := &Caps{
		strategy: utils.MetaBusy,
		aReqs:    make(chan struct{}, 0),
	}
	cs := NewCaps(0, utils.MetaBusy)

	// only check the strategy
	if !reflect.DeepEqual(exp.strategy, cs.strategy) {
		t.Errorf("Expected: %v ,received: %v", exp, cs)
	}

	if cs.IsLimited() {
		t.Errorf("Expected to not be limited")
	}

	if al := cs.Allocated(); al != 0 {
		t.Errorf("Expected: %v ,received: %v", 0, al)
	}
	if err := cs.Allocate(); err != utils.ErrMaxConcurentRPCExceededNoCaps {
		t.Errorf("Expected: %v ,received: %v", utils.ErrMaxConcurentRPCExceededNoCaps, err)
	}
	cs = NewCaps(1, utils.MetaBusy)
	if err := cs.Allocate(); err != nil {
		t.Error(err)
	}
	cs.Deallocate()
}

func TestCapsStats(t *testing.T) {
	st, err := engine.NewStatAverage(1, utils.MetaDynReq, nil)
	if err != nil {
		t.Error(err)
	}
	exp := &CapsStats{st: st}
	cr := NewCaps(0, utils.MetaBusy)
	exitChan := make(chan bool, 1)
	exitChan <- true
	cs := NewCapsStats(1, cr, exitChan)
	if !reflect.DeepEqual(exp, cs) {
		t.Errorf("Expected: %v ,received: %v", exp, cs)
	}
	<-exitChan
	exitChan = make(chan bool, 1)
	go func() {
		runtime.Gosched()
		time.Sleep(100)
		exitChan <- true
	}()
	cr = NewCaps(10, utils.MetaBusy)
	cr.Allocate()
	cr.Allocate()
	cs.loop(1, exitChan, cr)
	if avg := cs.GetAverage(2); avg <= 0 {
		t.Errorf("Expected at least an event to be processed: %v", avg)
	}
	if pk := cs.GetPeak(); pk != 2 {
		t.Errorf("Expected the peak to be 2 received: %v", pk)
	}
	<-exitChan
}

func TestCapsStatsGetAverage(t *testing.T) {
	st, err := engine.NewStatAverage(1, utils.MetaDynReq, nil)
	if err != nil {
		t.Error(err)
	}
	cs := &CapsStats{st: st}
	cs.addSample("1", 10)
	expAvg := 10.
	if avg := cs.GetAverage(2); avg != expAvg {
		t.Errorf("Expected: %v ,received: %v", expAvg, avg)
	}
	expPk := 10
	if pk := cs.GetPeak(); pk != expPk {
		t.Errorf("Expected: %v ,received:%v", expPk, pk)
	}
	cs.addSample("2", 16)
	expAvg = 13.
	if avg := cs.GetAverage(2); avg != expAvg {
		t.Errorf("Expected: %v ,received: %v", expAvg, avg)
	}
	expPk = 16
	if pk := cs.GetPeak(); pk != expPk {
		t.Errorf("Expected: %v ,received:%v", expPk, pk)
	}
	cs.OnEvict("2", nil)
	expAvg = 10.
	if avg := cs.GetAverage(2); avg != expAvg {
		t.Errorf("Expected: %v ,received: %v", expAvg, avg)
	}
	if pk := cs.GetPeak(); pk != expPk {
		t.Errorf("Expected: %v ,received:%v", expPk, pk)
	}
}

func TestFloatDP(t *testing.T) {
	f := floatDP(10.)
	expStr := "10"
	if s := f.String(); s != expStr {
		t.Errorf("Expected: %v ,received:%v", expStr, s)
	}
	if s, err := f.FieldAsString(nil); err != nil {
		t.Error(err)
	} else if s != expStr {
		t.Errorf("Expected: %v ,received:%v", expStr, s)
	}
	if r := f.RemoteHost(); r != nil {
		t.Errorf("Expected remote host to be nil received:%v", r)
	}
	exp := 10.
	if s, err := f.FieldAsInterface(nil); err != nil {
		t.Error(err)
	} else if s != exp {
		t.Errorf("Expected: %v ,received:%v", exp, s)
	}
}

type mockServerCodec struct{}

func (c *mockServerCodec) ReadRequestHeader(r *rpc.Request) (err error) {
	r.Seq = 0
	r.ServiceMethod = utils.CoreSv1Ping
	return
}

func (c *mockServerCodec) ReadRequestBody(x interface{}) (err error) {
	return
}
func (c *mockServerCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	return nil
}
func (c *mockServerCodec) Close() error { return nil }

func TestNewCapsServerCodec(t *testing.T) {
	mk := new(mockServerCodec)
	cr := NewCaps(0, utils.MetaBusy)
	if r := newCapsServerCodec(mk, cr); !reflect.DeepEqual(mk, r) {
		t.Errorf("Expected: %v ,received:%v", mk, r)
	}
	cr = NewCaps(1, utils.MetaBusy)
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

	if err = codec.ReadRequestBody("args"); err != nil {
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

func (*mockConn) Read(b []byte) (n int, err error)  { return }
func (*mockConn) Write(b []byte) (n int, err error) { return }
func (*mockConn) Close() error                      { return nil }
func (*mockConn) LocalAddr() net.Addr               { return utils.LocalAddr() }
func (*mockConn) RemoteAddr() net.Addr              { return utils.LocalAddr() }

func TestNewCapsGOBCodec(t *testing.T) {
	conn := new(mockConn)
	cr := NewCaps(0, utils.MetaBusy)
	anz := &analyzers.AnalyzerService{}
	exp := newGobServerCodec(conn)
	if r := newCapsGOBCodec(conn, cr, nil); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
	exp = analyzers.NewServerCodec(newGobServerCodec(conn), anz, utils.MetaGOB, utils.Local, utils.Local)
	if r := newCapsGOBCodec(conn, cr, anz); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
}

func TestNewCapsJSONCodec(t *testing.T) {
	conn := new(mockConn)
	cr := NewCaps(0, utils.MetaBusy)
	anz := &analyzers.AnalyzerService{}
	exp := jsonrpc.NewServerCodec(conn)
	if r := newCapsJSONCodec(conn, cr, nil); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
	exp = analyzers.NewServerCodec(jsonrpc.NewServerCodec(conn), anz, utils.MetaJSON, utils.Local, utils.Local)
	if r := newCapsJSONCodec(conn, cr, anz); !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, r)
	}
}
