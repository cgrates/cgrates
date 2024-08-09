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
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

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
	rcv.stopBiRPCServer = nil
	rcv.rpcSrv = nil
	rcv.birpcSrv = nil
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("\nExpected %+v,\nreceived %+v", expected, rcv)
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

func TestRegisterHttpFunc(t *testing.T) {
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

	rcv.RegisterHttpFunc("/home", handler)

	rcv.RpcRegisterName(utils.EmptyString, handler)

	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	rcv.StopBiRPC()
}

func TestHandleRequestCORSHeaders(t *testing.T) {
	caps := engine.NewCaps(0, utils.MetaBusy)
	rcv := NewServer(caps)

	rcv.rpcEnabled = true

	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/jsonrpc",
		bytes.NewBuffer([]byte("1")))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://origin.com")

	w := httptest.NewRecorder()

	rcv.handleRequest(w, req)

	if origin := req.Header.Get("Origin"); origin != "" {
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Errorf("Expected <%v>, got <%v>", "http://origin.com", got)
		}
	}

	expectedMethods := "POST, GET, OPTIONS, PUT, DELETE"
	if got := w.Header().Get("Access-Control-Allow-Methods"); got != expectedMethods {
		t.Errorf("Expected <%v>; got <%v>", expectedMethods, got)
	}

	expectedHeaders := "Accept, Accept-Language, Content-Type"
	if got := w.Header().Get("Access-Control-Allow-Headers"); got != expectedHeaders {
		t.Errorf("Expected <%v>; got <%v>", expectedHeaders, got)
	}
}
