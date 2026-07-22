// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

func TestNewHTTPAgent(t *testing.T) {
	connMgr := &engine.ConnManager{}
	filterS := &engine.FilterS{}
	dfltTenant := "defaultTenant"
	reqPayload := "requestPayload"
	rplyPayload := "responsePayload"
	reqProcessors := []*config.RequestProcessor{
		{},
	}
	sessionConns := []string{"conn1", "conn2"}
	statsConns := []string{"conn1", "conn2"}
	thresholdsConns := []string{"conn1", "conn2"}
	agent := NewHTTPAgent(
		connMgr,
		sessionConns,
		statsConns,
		thresholdsConns,
		filterS,
		dfltTenant,
		reqPayload,
		rplyPayload,
		reqProcessors,
		nil,
	)

	if agent.connMgr != connMgr {
		t.Errorf("Expected connMgr %v, got %v", connMgr, agent.connMgr)
	}
	if agent.filterS != filterS {
		t.Errorf("Expected filterS %v, got %v", filterS, agent.filterS)
	}
	if agent.dfltTenant != dfltTenant {
		t.Errorf("Expected dfltTenant %s, got %s", dfltTenant, agent.dfltTenant)
	}
	if agent.reqPayload != reqPayload {
		t.Errorf("Expected reqPayload %s, got %s", reqPayload, agent.reqPayload)
	}
	if agent.rplyPayload != rplyPayload {
		t.Errorf("Expected rplyPayload %s, got %s", rplyPayload, agent.rplyPayload)
	}
	if len(agent.reqProcessors) != len(reqProcessors) {
		t.Errorf("Expected reqProcessors length %d, got %d", len(reqProcessors), len(agent.reqProcessors))
	}
	for i, processor := range reqProcessors {
		if agent.reqProcessors[i] != processor {
			t.Errorf("Expected reqProcessors[%d] %v, got %v", i, processor, agent.reqProcessors[i])
		}
	}
	if len(agent.sessionConns) != len(sessionConns) {
		t.Errorf("Expected sessionConns length %d, got %d", len(sessionConns), len(agent.sessionConns))
	}
	for i, conn := range sessionConns {
		if agent.sessionConns[i] != conn {
			t.Errorf("Expected sessionConns[%d] %s, got %s", i, conn, agent.sessionConns[i])
		}
	}
	for i, conn := range statsConns {
		if agent.statsConns[i] != conn {
			t.Errorf("Expected statsConns[%d] %s, got %s", i, conn, agent.statsConns[i])
		}
	}
	for i, conn := range thresholdsConns {
		if agent.thresholdsConns[i] != conn {
			t.Errorf("Expected thresholdsConns[%d] %s, got %s", i, conn, agent.thresholdsConns[i])
		}
	}
}

func TestHTTPAgentServeHTTP(t *testing.T) {
	agent := &HTTPAgent{
		caps: engine.NewCaps(0, ""),
	}
	req := httptest.NewRequest("GET", "http://cgrates.org", nil)
	rr := httptest.NewRecorder()
	agent.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}
}
