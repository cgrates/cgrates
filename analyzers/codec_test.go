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

package analyzers

import (
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type mockServerCodec struct{}

func (c *mockServerCodec) ReadRequestHeader(r *birpc.Request) (err error) {
	r.Seq = 0
	r.ServiceMethod = utils.CoreSv1Ping
	return
}

func (c *mockServerCodec) ReadRequestBody(_ interface{}) (err error) {
	return
}
func (c *mockServerCodec) WriteResponse(_ *birpc.Response, _ interface{}) error {
	return nil
}
func (c *mockServerCodec) Close() error { return nil }

func TestNewServerCodec(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	cfg.AnalyzerSCfg().TTL = 30 * time.Minute
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}

	codec := NewAnalyzerServerCodec(new(mockServerCodec), anz, utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012")
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

	if err = codec.ReadRequestBody("args"); err != nil {
		t.Fatal(err)
	}
	if err = codec.WriteResponse(&birpc.Response{
		Error: "error",
		Seq:   0,
	}, "reply"); err != nil {
		t.Fatal(err)
	}
	if err = codec.Close(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	runtime.Gosched()
	if cnt, err := anz.db.DocCount(); err != nil {
		t.Fatal(err)
	} else if cnt != 1 {
		t.Errorf("Expected only one document received:%v", cnt)
	}
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}

type mockBiRPCCodec struct{}

func (mockBiRPCCodec) ReadHeader(r *birpc.Request, _ *birpc.Response) error {
	r.Seq = 0
	r.ServiceMethod = utils.CoreSv1Ping
	return nil
}
func (mockBiRPCCodec) ReadRequestBody(interface{}) error                { return nil }
func (mockBiRPCCodec) ReadResponseBody(interface{}) error               { return nil }
func (mockBiRPCCodec) WriteRequest(*birpc.Request, interface{}) error   { return nil }
func (mockBiRPCCodec) WriteResponse(*birpc.Response, interface{}) error { return nil }
func (mockBiRPCCodec) Close() error                                     { return nil }

func TestNewBiRPCCodec(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	cfg.AnalyzerSCfg().TTL = 30 * time.Minute
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}

	codec := NewAnalyzerBiRPCCodec(new(mockBiRPCCodec), anz, rpcclient.BiRPCJSON, "127.0.0.1:5565", "127.0.0.1:2012")
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

	if err = codec.ReadRequestBody("args"); err != nil {
		t.Fatal(err)
	}
	if err = codec.WriteResponse(&birpc.Response{
		Error: "error",
		Seq:   0,
	}, "reply"); err != nil {
		t.Fatal(err)
	}
	if err = codec.Close(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	runtime.Gosched()
	if cnt, err := anz.db.DocCount(); err != nil {
		t.Fatal(err)
	} else if cnt != 1 {
		t.Errorf("Expected only one document received:%v", cnt)
	}
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}

type mockBiRPCCodec2 struct{}

func (mockBiRPCCodec2) ReadHeader(_ *birpc.Request, r *birpc.Response) error {
	r.Seq = 0
	r.Error = "error"
	return nil
}
func (mockBiRPCCodec2) ReadRequestBody(interface{}) error                { return nil }
func (mockBiRPCCodec2) ReadResponseBody(interface{}) error               { return nil }
func (mockBiRPCCodec2) WriteRequest(*birpc.Request, interface{}) error   { return nil }
func (mockBiRPCCodec2) WriteResponse(*birpc.Response, interface{}) error { return nil }
func (mockBiRPCCodec2) Close() error                                     { return nil }

func TestNewBiRPCCodec2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	cfg.AnalyzerSCfg().TTL = 30 * time.Minute
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}

	codec := NewAnalyzerBiRPCCodec(new(mockBiRPCCodec2), anz, rpcclient.BiRPCJSON, "127.0.0.1:5565", "127.0.0.1:2012")
	if err = codec.WriteRequest(&birpc.Request{Seq: 0, ServiceMethod: utils.CoreSv1Ping}, "args"); err != nil {
		t.Fatal(err)
	}
	r := new(birpc.Response)
	expR := &birpc.Response{
		Seq:   0,
		Error: "error",
	}

	if err = codec.ReadHeader(&birpc.Request{}, r); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %v ,received:%v", expR, r)
	}

	if err = codec.ReadResponseBody("args"); err != nil {
		t.Fatal(err)
	}
	if err = codec.Close(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	runtime.Gosched()
	if cnt, err := anz.db.DocCount(); err != nil {
		t.Fatal(err)
	} else if cnt != 1 {
		t.Errorf("Expected only one document received:%v", cnt)
	}
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}
