/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package agents

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestV1WarnDisconnect(t *testing.T) {
	ja := &JanusAgent{}
	err := ja.V1WarnDisconnect(nil, nil, nil)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestV1DisconnectPeerJanus(t *testing.T) {
	ja := &JanusAgent{}
	var ctx context.Context
	var args *utils.DPRArgs
	var msg *string
	err := ja.V1DisconnectPeer(&ctx, args, msg)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestV1AlterSession(t *testing.T) {
	ja := &JanusAgent{}
	var ctx context.Context
	var event utils.CGREvent
	var msg *string
	err := ja.V1AlterSession(&ctx, event, msg)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestV1DisconnectSession(t *testing.T) {
	ja := &JanusAgent{}
	var ctx context.Context
	cgrEv := utils.CGREvent{
		Event: map[string]interface{}{},
	}
	var reply string
	err := ja.V1DisconnectSession(&ctx, cgrEv, &reply)
	if err == nil {
		t.Fatalf("Expected, got %v", err)
	}
	if reply == utils.OK {
		t.Errorf("Expected reply %v, got %v", utils.OK, reply)
	}
}

func TestCORSOptions(t *testing.T) {
	ja := &JanusAgent{}
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("OPTIONS", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	ja.CORSOptions(rr, req)
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "" {
		t.Errorf("Expected Access-Control-Allow-Origin header to be empty, got %v", origin)
	}
	if methods := rr.Header().Get("Access-Control-Allow-Methods"); methods == "POST, GET, OPTIONS, PUT, DELETE" {
		t.Errorf("Expected Access-Control-Allow-Methods header to be 'POST, GET, OPTIONS, PUT, DELETE', got %v", methods)
	}
	if headers := rr.Header().Get("Access-Control-Allow-Headers"); headers == "Accept, Accept-Language, Content-Type" {
		t.Errorf("Expected Access-Control-Allow-Headers header to be 'Accept, Accept-Language, Content-Type', got %v", headers)
	}
}
