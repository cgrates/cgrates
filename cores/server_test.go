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
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"

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
	rcv.httpServer = nil
	rcv.httpsServer = nil
	rcv.rpcServer = nil
	rcv.birpcSrv = nil
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

func TestRegisterHTTPFunc(t *testing.T) {
	log.SetOutput(io.Discard)
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

	rcv.RegisterHTTPFunc("/home", handler)

	rcv.RpcRegisterName(utils.EmptyString, handler)

	httpAgent := agents.NewHTTPAgent(nil, []string{}, nil, utils.EmptyString, utils.EmptyString, utils.EmptyString, nil)
	rcv.RegisterHttpHandler("invalid_pattern", httpAgent)

	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	rcv.StopBiRPC()
}

func TestRegisterProfiler(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = 1
	caps := engine.NewCaps(0, utils.MetaBusy)
	rcv := NewServer(caps)

	registerProfiler("test_prefix", rcv.httpMux)

	rcv.RegisterProfiler("/test_prefix")

	rcv.StopBiRPC()
}
