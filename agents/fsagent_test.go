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
	"github.com/cgrates/fsock"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
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
	err := fss.V1AlterSession(ctx, cgrEv, nil)
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
	err := fsa.V1DisconnectSession(context.Background(), utils.CGREvent{Event: mockEvent}, &testReply)
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

func TestFsAgentCall(t *testing.T) {
	ctx := context.Background()
	serviceMethod := "Method"
	args := "Args"
	reply := new(interface{})
	fsa := &FSsessions{}
	err := fsa.Call(ctx, serviceMethod, args, reply)
	if err == nil {
		t.Errorf("UNSUPPORTED_SERVICE_METHOD, got %v", err)
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
