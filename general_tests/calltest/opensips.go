// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package calltest

import (
	"bytes"
	"net"
	"testing"
	"time"
)

// Opensips runs opensips from ConfigFile as a foreground process. Start
// blocks until ReadyAddr answers a SIP OPTIONS and kills the process when the
// test ends. Start cgrates first: the cgrates module connects to it on startup.
type Opensips struct {
	ConfigFile string // -f
	ReadyAddr  string // SIP address polled for readiness, e.g. 127.0.0.1:5060
}

func (o Opensips) Start(t testing.TB) {
	t.Helper()
	p := startProcess(t, "opensips", "-f", o.ConfigFile, "-F")
	p.waitReady(t, 15*time.Second, "opensips at "+o.ReadyAddr, func() bool {
		return sipReachable(o.ReadyAddr)
	})
}

// sipReachable sends an OPTIONS with no user in the RURI; a residential script
// answers 484, enough to prove the SIP stack is up.
func sipReachable(addr string) bool {
	conn, err := net.DialTimeout("udp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	defer func() { _ = conn.Close() }()
	req := "OPTIONS sip:" + addr + " SIP/2.0\r\n" +
		"Via: SIP/2.0/UDP " + conn.LocalAddr().String() + ";branch=z9hG4bKcalltest\r\n" +
		"From: <sip:probe@127.0.0.1>;tag=calltest\r\n" +
		"To: <sip:" + addr + ">\r\n" +
		"Call-ID: calltest-ready\r\n" +
		"CSeq: 1 OPTIONS\r\n" +
		"Max-Forwards: 70\r\n" +
		"Content-Length: 0\r\n\r\n"
	_ = conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	if _, err := conn.Write([]byte(req)); err != nil {
		return false
	}
	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	if err != nil {
		return false
	}
	return bytes.HasPrefix(buf[:n], []byte("SIP/2.0"))
}
