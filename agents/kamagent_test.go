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
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/kamevapi"
)

func TestKAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPCClient(new(KamailioAgent))
}

func TestKamailioAgentV1WarnDisconnect(t *testing.T) {
	agent := KamailioAgent{}
	ctx := context.Background()
	args := make(map[string]any)
	var reply string
	err := agent.V1WarnDisconnect(ctx, args, &reply)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
}

func TestKamailioAgentV1DisconnectPeer(t *testing.T) {
	agent := KamailioAgent{}
	ctx := context.Background()
	dprArgs := &utils.DPRArgs{}
	var reply string

	err := agent.V1DisconnectPeer(ctx, dprArgs, &reply)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
}

func TestKamailioAgentV1AlterSession(t *testing.T) {
	agent := KamailioAgent{}
	ctx := context.Background()
	cgrEvent := utils.CGREvent{}
	var reply string
	err := agent.V1AlterSession(ctx, cgrEvent, &reply)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
}

func TestKamailioAgentReload(t *testing.T) {
	cfg := config.KamAgentCfg{
		EvapiConns: []*config.KamConnCfg{
			{},
			{},
			{},
		},
	}
	ka := &KamailioAgent{
		cfg: &cfg,
	}
	ka.Reload()
	if len(ka.conns) != len(cfg.EvapiConns) {
		t.Errorf("Expected conns length %d, but got %d", len(cfg.EvapiConns), len(ka.conns))
	}
	for i, conn := range ka.conns {
		if conn != nil {
			t.Errorf("Expected ka.conns[%d] to be nil, but got  value", i)
		}
	}
}

func TestKamailioAgentReloadResizesChannel(t *testing.T) {
	cfg := &config.KamAgentCfg{
		EvapiConns: []*config.KamConnCfg{{}, {}, {}},
	}
	ka := &KamailioAgent{
		cfg:     cfg,
		replyCh: make(chan []*sessions.SessionID, 1),
	}
	ka.Reload()
	if cap(ka.replyCh) != len(cfg.EvapiConns) {
		t.Errorf("expected replyCh cap %d after reload, got %d", len(cfg.EvapiConns), cap(ka.replyCh))
	}
}

func TestKamailioAgentOnDlgListDelivery(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ka := &KamailioAgent{replyCh: make(chan []*sessions.SessionID, 1)}
		evData := []byte(`{"Jsonrpl_body":{"Result":[]}}`)

		ka.onDlgList(evData, 0)
		select {
		case <-ka.replyCh:
		default:
			t.Fatal("expected the reply on the channel")
		}

		ka.replyCh <- nil // fill the buffer
		done := make(chan struct{})
		go func() {
			ka.onDlgList(evData, 0)
			close(done)
		}()
		synctest.Wait()
		select {
		case <-done:
		default:
			t.Fatal("onDlgList blocked on a full reply channel")
		}
	})
}

func TestKamailioAgentV1GetActiveSessionIDs(t *testing.T) {
	t.Run("collects replies from the dialog list", func(t *testing.T) {
		addr := startMockKamailio(t, buildDlgList("call-1", "call-2"), 0)
		ka := dialMockKamailio(t, addr, time.Second)

		var sIDs []*sessions.SessionID
		if err := ka.V1GetActiveSessionIDs(context.Background(), "", &sIDs); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sIDs) != 2 {
			t.Fatalf("expected 2 session IDs, got %d", len(sIDs))
		}
		if sIDs[0].OriginID != "call-1;tag" {
			t.Errorf("expected OriginID call-1;tag, got %q", sIDs[0].OriginID)
		}
		if sIDs[0].OriginHost == "" {
			t.Error("expected OriginHost set from the connection")
		}
	})

	t.Run("times out when no reply comes", func(t *testing.T) {
		addr := startMockKamailio(t, nil, -1)
		ka := dialMockKamailio(t, addr, 50*time.Millisecond)

		var sIDs []*sessions.SessionID
		if err := ka.V1GetActiveSessionIDs(context.Background(), "", &sIDs); err == nil {
			t.Fatal("expected timeout error, got nil")
		}
	})

	t.Run("returns ErrNoActiveSession when the dialog list is empty", func(t *testing.T) {
		addr := startMockKamailio(t, buildDlgList(), 0)
		ka := dialMockKamailio(t, addr, time.Second)

		var sIDs []*sessions.SessionID
		if err := ka.V1GetActiveSessionIDs(context.Background(), "", &sIDs); err != utils.ErrNoActiveSession {
			t.Fatalf("expected ErrNoActiveSession, got %v", err)
		}
	})

	t.Run("drops a stale reply left over from a previous sync", func(t *testing.T) {
		addr := startMockKamailio(t, buildDlgList("fresh"), 0)
		ka := dialMockKamailio(t, addr, time.Second)
		ka.replyCh <- []*sessions.SessionID{{OriginID: "stale"}}

		var sIDs []*sessions.SessionID
		if err := ka.V1GetActiveSessionIDs(context.Background(), "", &sIDs); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sIDs) != 1 || sIDs[0].OriginID != "fresh;tag" {
			t.Fatalf("expected only the fresh reply, got %+v", sIDs)
		}
	})
}

func dialMockKamailio(tb testing.TB, addr string, replyTimeout time.Duration) *KamailioAgent {
	tb.Helper()
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().ReplyTimeout = replyTimeout
	config.SetCgrConfig(cfg)
	ka := &KamailioAgent{
		conns:   make([]*kamevapi.KamEvapi, 1),
		replyCh: make(chan []*sessions.SessionID, 1),
	}
	conn, err := kamevapi.NewKamEvapi(addr, 0, 0, 0, utils.FibDuration,
		map[*regexp.Regexp][]func([]byte, int){kamDlgListRegexp: {ka.onDlgList}}, utils.Logger)
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { conn.Disconnect() })
	ka.conns[0] = conn
	return ka
}

// buildDlgList builds a CGR_DLG_LIST reply carrying one dialog per originID.
func buildDlgList(originIDs ...string) []byte {
	var b strings.Builder
	b.WriteString(`{"Event":"CGR_DLG_LIST","Jsonrpl_body":{"result":[`)
	for i, id := range originIDs {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"call-id":%q,"caller":{"tag":"tag"}}`, id)
	}
	b.WriteString(`]}}`)
	return []byte(b.String())
}

// startMockKamailio replies to each request with response after delay.
// A negative delay never replies, to hit the timeout path.
func startMockKamailio(tb testing.TB, response []byte, delay time.Duration) string {
	tb.Helper()
	netstring := fmt.Appendf(nil, "%d:%s,", len(response), response)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go serveMockKamailio(conn, netstring, delay)
		}
	}()
	return ln.Addr().String()
}

func serveMockKamailio(conn net.Conn, response []byte, delay time.Duration) {
	defer conn.Close()
	rd := bufio.NewReaderSize(conn, 8192)
	for {
		lenStr, err := rd.ReadString(':')
		if err != nil {
			return
		}
		n, err := strconv.Atoi(lenStr[:len(lenStr)-1])
		if err != nil {
			return
		}
		if _, err := io.CopyN(io.Discard, rd, int64(n+1)); err != nil { // payload + trailing comma
			return
		}
		if delay < 0 {
			continue // never reply, let the agent time out
		}
		if delay > 0 {
			time.Sleep(delay)
		}
		if _, err := conn.Write(response); err != nil {
			return
		}
	}
}
