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
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/cenkalti/rpc2"

	"github.com/cgrates/cgrates/agents"

	"github.com/cgrates/cgrates/analyzers"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewServer(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = 1
	caps := engine.NewCaps(0, utils.MetaBusy)

	expected := &Server{
		httpMux:  http.NewServeMux(),
		httpsMux: http.NewServeMux(),
		caps:     caps,
	}
	rcv := NewServer(caps)
	rcv.stopbiRPCServer = nil
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	cfgDflt.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	analz, err := analyzers.NewAnalyzerService(cfgDflt)
	if err != nil {
		t.Error(err)
	}
	expected.anz = analz
	if rcv.SetAnalyzer(analz); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}

func TestServerRpcRrgister(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = 1
	caps := engine.NewCaps(0, utils.MetaBusy)
	rcv := NewServer(caps)

	cfgDflt.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfgDflt.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	analz, err := analyzers.NewAnalyzerService(cfgDflt)
	if err != nil {
		t.Error(err)
	}
	rcv.SetAnalyzer(analz)

	rcv.RpcRegister(utils.EmptyString)

	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	rcv.StopBiRPC()
}

func TestRegisterHttpFunc(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = 1
	caps := engine.NewCaps(0, utils.MetaBusy)
	rcv := NewServer(caps)

	cfgDflt.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfgDflt.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	analz, err := analyzers.NewAnalyzerService(cfgDflt)
	if err != nil {
		t.Error(err)
	}
	rcv.SetAnalyzer(analz)

	handler := func(http.ResponseWriter, *http.Request) {}

	rcv.RegisterHttpFunc("/home", handler)

	rcv.RpcRegisterName(utils.EmptyString, handler)

	httpAgent := agents.NewHTTPAgent(nil, []string{}, nil, utils.EmptyString, utils.EmptyString, utils.EmptyString, nil)
	rcv.RegisterHttpHandler("invalid_pattern", httpAgent)

	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	rcv.StopBiRPC()
}

func TestBiRPCRegisterName(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = 1
	caps := engine.NewCaps(0, utils.MetaBusy)
	rcv := NewServer(caps)

	cfgDflt.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfgDflt.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	analz, err := analyzers.NewAnalyzerService(cfgDflt)
	if err != nil {
		t.Error(err)
	}
	rcv.SetAnalyzer(analz)

	rcv.birpcSrv = &rpc2.Server{}

	rcv.BiRPCRegister(mockReadWriteCloserErrorNilResponse{})

	rcv.birpcSrv = nil
	rcv.BiRPCRegister(gobServerCodec{})

	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	rcv.StopBiRPC()
}

func TestServeJSONAndGob(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = 1
	caps := engine.NewCaps(0, utils.MetaBusy)
	rcv := NewServer(caps)

	cfgDflt.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfgDflt.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	analz, err := analyzers.NewAnalyzerService(cfgDflt)
	if err != nil {
		t.Error(err)
	}
	rcv.SetAnalyzer(analz)
	rcv.rpcEnabled = true

	shdChan := utils.NewSyncedChan()

	//invalid port format
	rcv.ServeJSON("2015", shdChan)

	rcv.ServeGOB("2015", shdChan)

	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	rcv.StopBiRPC()
}
