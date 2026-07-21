// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package commonlisteners

import (
	"io"
	"log"
	"net/http"
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

	expected := &CommonListenerS{
		httpMux:  http.NewServeMux(),
		httpsMux: http.NewServeMux(),
		caps:     caps,
	}
	rcv := NewCommonListenerS(caps)
	rcv.stopbiRPCServer = nil
	rcv.httpServer = nil
	rcv.httpsServer = nil
	rcv.rpcServer = nil
	rcv.birpcSrv = nil
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	cfgDflt.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	analz, err := analyzers.NewAnalyzerS(cfgDflt)
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
	rcv := NewCommonListenerS(caps)

	cfgDflt.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfgDflt.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	analz, err := analyzers.NewAnalyzerS(cfgDflt)
	if err != nil {
		t.Error(err)
	}
	rcv.SetAnalyzer(analz)

	handler := func(http.ResponseWriter, *http.Request) {}

	rcv.RegisterHTTPFunc("/home", handler)

	rcv.RpcRegisterName(utils.EmptyString, handler)

	if err := os.RemoveAll(cfgDflt.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	rcv.StopBiRPC()
}
