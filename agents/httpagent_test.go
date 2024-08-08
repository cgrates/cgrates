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
	agent := NewHTTPAgent(
		connMgr,
		sessionConns,
		filterS,
		dfltTenant,
		reqPayload,
		rplyPayload,
		reqProcessors,
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
}

func TestHTTPAgentServeHTTP(t *testing.T) {
	agent := &HTTPAgent{}
	req := httptest.NewRequest("GET", "http://cgrates.org", nil)
	rr := httptest.NewRecorder()
	agent.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}
}
