// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
