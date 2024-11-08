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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/fsock"
)

func TestFAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPCClient(new(FSsessions))
}

func TestFsAgentV1DisconnectPeer(t *testing.T) {
	ctx := context.Background()
	args := &utils.DPRArgs{}
	fss := &FSsessions{}
	err := fss.V1DisconnectPeer(ctx, args, nil)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error: %v, got: %v", utils.ErrNotImplemented, err)
	}
}

func TestFsAgentV1AlterSession(t *testing.T) {
	ctx := context.Background()
	cgrEv := utils.CGREvent{}
	fss := &FSsessions{}
	err := fss.V1AlterSession(ctx, &cgrEv, nil)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error: %v, got: %v", utils.ErrNotImplemented, err)
	}
}

func TestFsAgentCreateHandlers(t *testing.T) {
	cfg := &config.FsAgentCfg{
		SubscribePark: true,
	}
	fs := &FSsessions{
		cfg: cfg,
	}
	handlers := fs.createHandlers()
	if _, ok := handlers["CHANNEL_ANSWER"]; !ok {
		t.Error("Expected CHANNEL_ANSWER handler, but not found")
	}
	if _, ok := handlers["CHANNEL_HANGUP_COMPLETE"]; !ok {
		t.Error("Expected CHANNEL_HANGUP_COMPLETE handler, but not found")
	}
	if _, ok := handlers["CHANNEL_PARK"]; !ok {
		t.Error("Expected CHANNEL_PARK handler, but not found")
	}
}

func TestFsAgentV1WarnDisconnect(t *testing.T) {

	cfg := &config.FsAgentCfg{
		LowBalanceAnnFile: "File",
		EventSocketConns: []*config.FsConnCfg{
			{
				Address:              "Address",
				Password:             "Password",
				Reconnects:           3,
				MaxReconnectInterval: 10 * time.Second,
				ReplyTimeout:         5 * time.Second,
				Alias:                "Alias",
			},
		},
	}

	fsa := &FSsessions{
		cfg:   cfg,
		conns: []*fsock.FSock{},
	}
	ctx := &context.Context{}
	args := map[string]interface{}{
		"OriginID": "ID",
		"FsConnID": int64(0),
	}
	var reply string
	fsa.cfg.LowBalanceAnnFile = utils.EmptyString
	err := fsa.V1WarnDisconnect(ctx, args, &reply)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if reply != utils.OK {
		t.Errorf("Expected reply to be 'OK', got '%s'", reply)
	}
	fsa.cfg.LowBalanceAnnFile = "File"
	err = fsa.V1WarnDisconnect(ctx, args, &reply)
	if err == nil {
		t.Errorf("Index out of range[0,0): 0, got %v", err)
	}
	if reply != utils.OK {
		t.Errorf("Expected reply to be 'OK', got '%s'", reply)
	}
	delete(args, "FsConnID")
	err = fsa.V1WarnDisconnect(ctx, args, &reply)
	if err == nil {
		t.Error("Expected an error, got nil")
	}
}

func TestFsAgentReload(t *testing.T) {
	cfg := &config.FsAgentCfg{
		EventSocketConns: []*config.FsConnCfg{
			{
				Address:              "Address",
				Password:             "Password",
				Reconnects:           3,
				MaxReconnectInterval: 10 * time.Second,
				ReplyTimeout:         5 * time.Second,
				Alias:                "Alias",
			},
		},
	}
	fsa := &FSsessions{
		cfg: cfg,
	}
	fsa.Reload()

	if len(fsa.conns) != 1 {
		t.Errorf("Expected fsa.conns length to be 1, got %d", len(fsa.conns))
	}
	if len(fsa.senderPools) != 1 {
		t.Errorf("Expected fsa.senderPools length to be 1, got %d", len(fsa.senderPools))
	}

}

func TestFsAgentV1GetActiveSessionIDs(t *testing.T) {
	cfg := &config.FsAgentCfg{
		EventSocketConns: []*config.FsConnCfg{
			{
				Address:              "Address",
				Password:             "Password",
				Reconnects:           3,
				MaxReconnectInterval: 10 * time.Second,
				ReplyTimeout:         5 * time.Second,
				Alias:                "Alias",
			},
		},
		ActiveSessionDelimiter: ",",
	}

	fsa := &FSsessions{
		cfg: cfg,
	}

	ctx := context.Background()
	var sessionIDs []*sessions.SessionID
	err := fsa.V1GetActiveSessionIDs(ctx, "", &sessionIDs)
	if err == nil {
		t.Errorf("NO_ACTIVE_SESSION, got %v", err)
	}
	if len(sessionIDs) != 0 {
		t.Errorf("Expected sessionIDs slice to be populated, got empty")
	}

}

func TestFsAgentV1DisconnectSession(t *testing.T) {
	mockEvent := map[string]any{
		utils.OriginID:        "Id",
		utils.DisconnectCause: "Cause",
		FsConnID:              int64(0),
	}
	testReply := ""
	fsa := &FSsessions{
		conns: []*fsock.FSock{},
	}
	err := fsa.V1DisconnectSession(context.Background(), &utils.CGREvent{Event: mockEvent}, &testReply)
	if err == nil {
		t.Errorf("Index out of range[0,0): 0, got %v", err)
	}
	if testReply == utils.OK {
		t.Errorf("Expected reply to be 'OK', got %s", testReply)
	}
}

func TestFsAgentShutdown(t *testing.T) {
	fsa := &FSsessions{
		conns: []*fsock.FSock{},
	}
	err := fsa.Shutdown()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestFsAgentNewFSsessions(t *testing.T) {
	fsAgentConfig := &config.FsAgentCfg{
		EventSocketConns: []*config.FsConnCfg{
			{
				Address:              "Address",
				Password:             "Password",
				Reconnects:           3,
				MaxReconnectInterval: time.Second * 30,
				ReplyTimeout:         time.Second * 10,
			},
		},
	}
	timezone := "UTC"
	connMgr := &engine.ConnManager{}

	fsa, err := NewFSsessions(fsAgentConfig, timezone, connMgr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if fsa == nil {
		t.Fatal("Expected non-nil FSsessions instance, got nil")
	}
	if fsa.cfg != fsAgentConfig {
		t.Errorf("Expected fsa.cfg to be %v, got %v", fsAgentConfig, fsa.cfg)
	}
	if len(fsa.conns) != len(fsAgentConfig.EventSocketConns) {
		t.Errorf("Expected fsa.conns length to be %d, got %d", len(fsAgentConfig.EventSocketConns), len(fsa.conns))
	}
	if len(fsa.senderPools) != len(fsAgentConfig.EventSocketConns) {
		t.Errorf("Expected fsa.senderPools length to be %d, got %d", len(fsAgentConfig.EventSocketConns), len(fsa.senderPools))
	}
	if fsa.timezone != timezone {
		t.Errorf("Expected fsa.timezone to be %s, got %s", timezone, fsa.timezone)
	}
	if fsa.connMgr != connMgr {
		t.Errorf("Expected fsa.connMgr to be %v, got %v", connMgr, fsa.connMgr)
	}
	if fsa.ctx == nil {
		t.Error("Expected fsa.ctx to be initialized, got nil")
	}
}

func TestFSsessionsV1DisconnectPeer(t *testing.T) {
	fsSessions := &FSsessions{

		connMgr: &engine.ConnManager{},
	}
	ctx := context.Background()
	args := &utils.DPRArgs{}
	reply := ""
	err := fsSessions.V1DisconnectPeer(ctx, args, &reply)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestFSsessionsV1AlterSession(t *testing.T) {
	fsSessions := &FSsessions{

		connMgr: &engine.ConnManager{},
	}
	ctx := context.Background()
	event := utils.CGREvent{}
	reply := ""
	err := fsSessions.V1AlterSession(ctx, &event, &reply)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestNewFSsessions(t *testing.T) {
	fsAgentConfig := &config.FsAgentCfg{}
	timezone := "UTC"
	connMgr := &engine.ConnManager{}
	fsSessions, err := NewFSsessions(fsAgentConfig, timezone, connMgr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if fsSessions == nil {
		t.Fatalf("Expected fsSessions to be non-nil")
	}
	if fsSessions.cfg != fsAgentConfig {
		t.Errorf("Expected cfg to be %v, got %v", fsAgentConfig, fsSessions.cfg)
	}
	if len(fsSessions.conns) != len(fsAgentConfig.EventSocketConns) {
		t.Errorf("Expected conns length %d, got %d", len(fsAgentConfig.EventSocketConns), len(fsSessions.conns))
	}
	if len(fsSessions.senderPools) != len(fsAgentConfig.EventSocketConns) {
		t.Errorf("Expected senderPools length %d, got %d", len(fsAgentConfig.EventSocketConns), len(fsSessions.senderPools))
	}
	if fsSessions.timezone != timezone {
		t.Errorf("Expected timezone to be %s, got %s", timezone, fsSessions.timezone)
	}
	if fsSessions.connMgr != connMgr {
		t.Errorf("Expected connMgr to be %v, got %v", connMgr, fsSessions.connMgr)
	}
	if fsSessions.ctx == nil {
		t.Error("Expected ctx to be non-nil")
	}
}

func TestFSsessionsCreateHandlers(t *testing.T) {
	tests := []struct {
		name             string
		subscribePark    bool
		expectedHandlers []string
	}{
		{
			name:             "Without SubscribePark",
			subscribePark:    false,
			expectedHandlers: []string{"CHANNEL_ANSWER", "CHANNEL_HANGUP_COMPLETE"},
		},
		{
			name:             "With SubscribePark",
			subscribePark:    true,
			expectedHandlers: []string{"CHANNEL_ANSWER", "CHANNEL_HANGUP_COMPLETE", "CHANNEL_PARK"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fsSessions := &FSsessions{
				cfg: &config.FsAgentCfg{
					SubscribePark: tt.subscribePark,
				},
			}
			handlers := fsSessions.createHandlers()
			for _, expectedHandler := range tt.expectedHandlers {
				if _, exists := handlers[expectedHandler]; !exists {
					t.Errorf("Expected handler %s to be present, but it was not", expectedHandler)
				}
			}
			for handler := range handlers {
				found := false
				for _, expectedHandler := range tt.expectedHandlers {
					if handler == expectedHandler {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unexpected handler %s found", handler)
				}
			}
		})
	}
}
func TestFSsessionsReload(t *testing.T) {
	eventSocketConns := []string{"conn1", "conn2", "conn3"}
	cfg := &config.FsAgentCfg{}

	fsa := &FSsessions{
		cfg:         cfg,
		conns:       make([]*fsock.FSock, 1),
		senderPools: make([]*fsock.FSockPool, 1),
	}

	fsa.Reload()

	if len(fsa.conns) == len(eventSocketConns) {
		t.Errorf("Expected conns length %d, got %d", len(eventSocketConns), len(fsa.conns))
	}

	if len(fsa.senderPools) == len(eventSocketConns) {
		t.Errorf("Expected senderPools length %d, got %d", len(eventSocketConns), len(fsa.senderPools))
	}

	for i, conn := range fsa.conns {
		if conn != nil {
			t.Errorf("Expected conns[%d] to be nil, got %v", i, conn)
		}
	}
	for i, pool := range fsa.senderPools {
		if pool != nil {
			t.Errorf("Expected senderPools[%d] to be nil, got %v", i, pool)
		}
	}
}

func TestFSsessionsV1WarnDisconnect(t *testing.T) {
	cfg := &config.FsAgentCfg{}
	fsa := FSsessions{
		cfg: cfg,
	}
	ctx := context.Background()
	args := map[string]any{}
	var reply string
	err := fsa.V1WarnDisconnect(ctx, args, &reply)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if reply != utils.OK {
		t.Errorf("Expected reply: %s, but got: %s", utils.OK, reply)
	}

}
