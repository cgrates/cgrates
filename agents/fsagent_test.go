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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/fsock"
)

func TestFAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPClient(new(FSsessions))
}

func TestFSsessionsReload(t *testing.T) {
	cfg := &config.FsAgentCfg{}
	sm := &FSsessions{
		cfg:         cfg,
		conns:       []*fsock.FSock{},
		senderPools: []*fsock.FSockPool{},
	}
	sm.Reload()
	if len(sm.conns) != len(cfg.EventSocketConns) {
		t.Errorf("Expected conns length %d, but got %d", len(cfg.EventSocketConns), len(sm.conns))
	}
	if len(sm.senderPools) != len(cfg.EventSocketConns) {
		t.Errorf("Expected senderPools length %d, but got %d", len(cfg.EventSocketConns), len(sm.senderPools))
	}
	for i, conn := range sm.conns {
		if conn != nil {
			t.Errorf("Expected conns[%d] to be nil, but got %v", i, conn)
		}
	}
	for i, pool := range sm.senderPools {
		if pool != nil {
			t.Errorf("Expected senderPools[%d] to be nil, but got %v", i, pool)
		}
	}
}

func TestNewFSsessions(t *testing.T) {
	cfg := &config.FsAgentCfg{}
	timezone := "UTC"
	connMgr := &engine.ConnManager{}
	fsa := NewFSsessions(cfg, timezone, connMgr)
	if fsa.cfg != cfg {
		t.Errorf("Expected cfg to be %v, but got %v", cfg, fsa.cfg)
	}
	if fsa.timezone != timezone {
		t.Errorf("Expected timezone to be %s, but got %s", timezone, fsa.timezone)
	}
	if fsa.connMgr != connMgr {
		t.Errorf("Expected connMgr to be %v, but got %v", connMgr, fsa.connMgr)
	}
	if len(fsa.conns) != len(cfg.EventSocketConns) {
		t.Errorf("Expected conns length %d, but got %d", len(cfg.EventSocketConns), len(fsa.conns))
	}
	for i, conn := range fsa.conns {
		if conn != nil {
			t.Errorf("Expected conns[%d] to be nil, but got %v", i, conn)
		}
	}
	if len(fsa.senderPools) != len(cfg.EventSocketConns) {
		t.Errorf("Expected senderPools length %d, but got %d", len(cfg.EventSocketConns), len(fsa.senderPools))
	}
	for i, pool := range fsa.senderPools {
		if pool != nil {
			t.Errorf("Expected senderPools[%d] to be nil, but got %v", i, pool)
		}
	}
}

func TestFSsessionsCreateHandlers(t *testing.T) {
	cfgNoPark := &config.FsAgentCfg{
		SubscribePark: false,
	}
	smNoPark := &FSsessions{
		cfg: cfgNoPark,
	}
	handlersNoPark := smNoPark.createHandlers()
	if len(handlersNoPark["CHANNEL_ANSWER"]) != 1 {
		t.Errorf("Expected 1 handler for CHANNEL_ANSWER, but got %d", len(handlersNoPark["CHANNEL_ANSWER"]))
	}
	if len(handlersNoPark["CHANNEL_HANGUP_COMPLETE"]) != 1 {
		t.Errorf("Expected 1 handler for CHANNEL_HANGUP_COMPLETE, but got %d", len(handlersNoPark["CHANNEL_HANGUP_COMPLETE"]))
	}
	if _, ok := handlersNoPark["CHANNEL_PARK"]; ok {
		t.Errorf("CHANNEL_PARK handler should not be present when SubscribePark is false")
	}
	cfgWithPark := &config.FsAgentCfg{
		SubscribePark: true,
	}
	smWithPark := &FSsessions{
		cfg: cfgWithPark,
	}
	handlersWithPark := smWithPark.createHandlers()
	if len(handlersWithPark["CHANNEL_ANSWER"]) != 1 {
		t.Errorf("Expected 1 handler for CHANNEL_ANSWER, but got %d", len(handlersWithPark["CHANNEL_ANSWER"]))
	}
	if len(handlersWithPark["CHANNEL_HANGUP_COMPLETE"]) != 1 {
		t.Errorf("Expected 1 handler for CHANNEL_HANGUP_COMPLETE, but got %d", len(handlersWithPark["CHANNEL_HANGUP_COMPLETE"]))
	}
	if len(handlersWithPark["CHANNEL_PARK"]) != 1 {
		t.Errorf("Expected 1 handler for CHANNEL_PARK when SubscribePark is true, but got %d", len(handlersWithPark["CHANNEL_PARK"]))
	}
}

func TestFSsessionsV1GetActiveSessionIDsErrorHandling(t *testing.T) {
	sm := &FSsessions{}
	var sessionIDs []*sessions.SessionID
	err := sm.V1GetActiveSessionIDs("", &sessionIDs)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(sessionIDs) != 0 {
		t.Errorf("Expected no session IDs, but got %d", len(sessionIDs))
	}
}
