// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package commonlisteners

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewServer(t *testing.T) {
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

	analz := new(analyzers.AnalyzerS)
	rcv.SetAnalyzer(analz)
	if got := rcv.anz.Load(); got != analz {
		t.Errorf("Expected %p, received %p", analz, got)
	}
}
