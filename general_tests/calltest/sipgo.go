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

package calltest

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
)

// SipgoUAC places SIP calls through Addr using sipgo's dialog layer.
type SipgoUAC struct {
	Addr    string            // proxy address, host:port
	Headers map[string]string // extra headers appended to the INVITE
}

func (u SipgoUAC) headers() []sip.Header {
	hdrs := make([]sip.Header, 0, len(u.Headers))
	for _, name := range slices.Sorted(maps.Keys(u.Headers)) {
		hdrs = append(hdrs, sip.NewHeader(name, u.Headers[name]))
	}
	return hdrs
}

func (u SipgoUAC) Call(t testing.TB, c CallParams) {
	t.Helper()
	checkCallParams(t, "sipgo uac", c)
	checkAddr(t, "sipgo uac", u.Addr)
	host, portStr, err := net.SplitHostPort(u.Addr)
	if err != nil {
		t.Fatalf("sipgo uac: bad addr %q: %v", u.Addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("sipgo uac: bad addr port %q: %v", portStr, err)
	}
	fromUser := c.From
	ctx, cancel := context.WithTimeout(context.Background(), c.HoldTime+10*time.Second)
	defer cancel()

	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent(fromUser),
		sipgo.WithUserAgentHostname("127.0.0.1"),
	)
	if err != nil {
		t.Fatalf("sipgo uac %s->%s: new ua: %v", fromUser, c.To, err)
	}
	defer func() { _ = ua.Close() }()
	client, err := sipgo.NewClient(ua, sipgo.WithClientHostname("127.0.0.1"))
	if err != nil {
		t.Fatalf("sipgo uac %s->%s: new client: %v", fromUser, c.To, err)
	}
	dialogCli := sipgo.NewDialogClientCache(client, sip.ContactHeader{
		Address: sip.Uri{User: fromUser, Host: "127.0.0.1"},
	})

	dialog, err := dialogCli.Invite(ctx, sip.Uri{User: c.To, Host: host, Port: port}, nil, u.headers()...)
	if err != nil {
		t.Fatalf("sipgo uac %s->%s: invite: %v", fromUser, c.To, err)
	}
	defer func() { _ = dialog.Close() }()
	if err := dialog.WaitAnswer(ctx, sipgo.AnswerOptions{}); err != nil {
		var e *sipgo.ErrDialogResponse
		if errors.As(err, &e) {
			t.Fatalf("sipgo uac %s->%s: SIP %d %s", fromUser, c.To, e.Res.StatusCode, e.Res.Reason)
		}
		t.Fatalf("sipgo uac %s->%s: wait answer: %v", fromUser, c.To, err)
	}
	if err := dialog.Ack(ctx); err != nil {
		t.Fatalf("sipgo uac %s->%s: ack: %v", fromUser, c.To, err)
	}
	select {
	case <-time.After(c.HoldTime):
	case <-ctx.Done():
	}
	if err := dialog.Bye(ctx); err != nil {
		t.Fatalf("sipgo uac %s->%s: bye: %v", fromUser, c.To, err)
	}
}

// SipgoUAS answers calls on Port until the test ends.
type SipgoUAS struct {
	Port int
}

func (u SipgoUAS) Start(t testing.TB) {
	t.Helper()
	if u.Port == 0 {
		t.Fatal("sipgo uas: port not set")
	}
	ua, err := sipgo.NewUA()
	if err != nil {
		t.Fatalf("sipgo uas: new ua: %v", err)
	}
	srv, err := sipgo.NewServer(ua)
	if err != nil {
		t.Fatalf("sipgo uas: new server: %v", err)
	}
	client, err := sipgo.NewClient(ua, sipgo.WithClientHostname("127.0.0.1"))
	if err != nil {
		t.Fatalf("sipgo uas: new client: %v", err)
	}
	dialogSrv := sipgo.NewDialogServerCache(client, sip.ContactHeader{
		Address: sip.Uri{Host: "127.0.0.1", Port: u.Port},
	})
	srv.OnInvite(func(req *sip.Request, tx sip.ServerTransaction) {
		dlg, err := dialogSrv.ReadInvite(req, tx)
		if err != nil {
			return
		}
		_ = dlg.Respond(sip.StatusTrying, "Trying", nil)
		_ = dlg.Respond(sip.StatusOK, "OK", nil)
		go func() {
			<-dlg.Context().Done()
			_ = dlg.Close()
		}()
	})
	srv.OnAck(func(req *sip.Request, tx sip.ServerTransaction) {
		_ = dialogSrv.ReadAck(req, tx)
	})
	srv.OnBye(func(req *sip.Request, tx sip.ServerTransaction) {
		_ = dialogSrv.ReadBye(req, tx)
	})

	conn, err := net.ListenPacket("udp", fmt.Sprintf("127.0.0.1:%d", u.Port))
	if err != nil {
		t.Fatalf("sipgo uas: listen :%d: %v", u.Port, err)
	}
	done := make(chan struct{})
	go func() {
		_ = srv.ServeUDP(conn)
		close(done)
	}()
	t.Cleanup(func() {
		_ = conn.Close()
		if !waitDone(done, time.Second) {
			t.Errorf("sipgo uas: server did not stop")
		}
		_ = ua.Close()
	})
}
