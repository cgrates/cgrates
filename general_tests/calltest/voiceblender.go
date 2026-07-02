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
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	voiceblender "github.com/VoiceBlender/voiceblender-go"
)

func startVoiceBlender(t testing.TB, port, rtpPortMin, rtpPortMax int) *voiceblender.Client {
	t.Helper()
	if rtpPortMin > rtpPortMax {
		t.Fatalf("voiceblender: rtp port min %d greater than max %d", rtpPortMin, rtpPortMax)
	}
	path := needBinary(t, "voiceblender")
	httpLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("voiceblender: pick api port: %v", err)
	}
	apiAddr := httpLn.Addr().String()
	_ = httpLn.Close()
	cmd := exec.Command(path)
	cmd.Env = append(os.Environ(),
		"HTTP_ADDR="+apiAddr,
		fmt.Sprintf("SIP_PORT=%d", port),
		"SIP_BIND_IP=127.0.0.1",
		"SIP_LISTEN_IP=127.0.0.1",
		fmt.Sprintf("RTP_PORT_MIN=%d", rtpPortMin),
		fmt.Sprintf("RTP_PORT_MAX=%d", rtpPortMax),
	)
	client := voiceblender.New(voiceblender.WithBaseURL("http://" + apiAddr + "/v1"))
	startCmd(t, "voiceblender", cmd).waitReady(t, 5*time.Second,
		"voiceblender api at "+apiAddr, func() bool {
			return voiceBlenderReady(client)
		})
	return client
}

func voiceBlenderReady(client *voiceblender.Client) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if _, err := client.ListLegs(ctx); err != nil {
		return false
	}
	stream, err := client.Events(ctx)
	if err != nil {
		return false
	}
	_ = stream.Close()
	return true
}

// VoiceBlenderServer starts a daemon and returns its REST client.
type VoiceBlenderServer struct {
	Port       int // SIP listen port
	RTPPortMin int // defaults to 10000
	RTPPortMax int // defaults to 20000
}

func (s VoiceBlenderServer) Start(t testing.TB) *voiceblender.Client {
	t.Helper()
	if s.Port == 0 {
		s.Port = 5090
	}
	if s.RTPPortMin == 0 {
		s.RTPPortMin = 10000
	}
	if s.RTPPortMax == 0 {
		s.RTPPortMax = 20000
	}
	return startVoiceBlender(t, s.Port, s.RTPPortMin, s.RTPPortMax)
}

// VoiceBlenderUAC places calls through a voiceblender daemon.
type VoiceBlenderUAC struct {
	Client  *voiceblender.Client
	Addr    string // proxy or peer address, host:port
	Headers map[string]string
}

func (u VoiceBlenderUAC) Call(t testing.TB, c CallParams) {
	t.Helper()
	if u.Client == nil {
		t.Fatal("voiceblender uac: client not set")
	}
	checkCallParams(t, "voiceblender uac", c)
	toURI := u.toURI(t, c)
	fromUser := c.From
	maxDuration := voiceBlenderMaxDuration(c.HoldTime)
	ctx, cancel := context.WithTimeout(context.Background(), voiceBlenderTimeout(maxDuration))
	defer cancel()

	stream, err := u.Client.Events(ctx)
	if err != nil {
		t.Fatalf("voiceblender event stream: %v", err)
	}
	defer func() { _ = stream.Close() }()

	leg, err := u.Client.CreateLeg(ctx, voiceblender.CreateLegRequest{
		// The create endpoint accepts "sip", not LegTypeSIPOutbound.
		Type:        "sip",
		To:          toURI,
		From:        fromUser,
		Codecs:      []string{"PCMU"},
		MaxDuration: maxDuration,
		Headers:     u.Headers,
	})
	if err != nil {
		t.Fatalf("voiceblender create leg %s->%s: %v", fromUser, toURI, err)
	}
	for {
		ev, err := stream.Next(ctx)
		if err != nil {
			t.Fatalf("voiceblender leg %s: %v", leg.ID, err)
		}
		d, ok := ev.(*voiceblender.LegDisconnectedEvent)
		if !ok || d.LegID != leg.ID {
			continue
		}
		if d.Cdr.DurationAnswered == 0 {
			t.Fatalf("voiceblender leg %s never connected: %s", leg.ID, d.Cdr.Reason)
		}
		return
	}
}

func (u VoiceBlenderUAC) toURI(t testing.TB, c CallParams) string {
	t.Helper()
	if strings.HasPrefix(c.To, "sip:") {
		return c.To
	}
	if u.Addr == "" {
		t.Fatal("voiceblender uac: addr not set")
	}
	return "sip:" + c.To + "@" + u.Addr
}

// VoiceBlenderUAS answers calls on Port until the test ends. Each daemon needs
// its own RTP port range.
type VoiceBlenderUAS struct {
	Port       int
	RTPPortMin int // defaults to 20001
	RTPPortMax int // defaults to 30000
}

func (u VoiceBlenderUAS) Start(t testing.TB) {
	t.Helper()
	if u.Port == 0 {
		t.Fatal("voiceblender uas: port not set")
	}
	if u.RTPPortMin == 0 {
		u.RTPPortMin = 20001
	}
	if u.RTPPortMax == 0 {
		u.RTPPortMax = 30000
	}
	client := startVoiceBlender(t, u.Port, u.RTPPortMin, u.RTPPortMax)
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := client.Events(ctx)
	if err != nil {
		t.Fatalf("voiceblender uas event stream: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			ev, err := stream.Next(ctx)
			if err != nil {
				return
			}
			r, ok := ev.(*voiceblender.LegRingingEvent)
			if !ok || r.LegType != string(voiceblender.LegTypeSIPInbound) {
				continue
			}
			_, _ = client.Leg(r.LegID).Answer(ctx, voiceblender.AnswerLegRequest{})
		}
	}()
	t.Cleanup(func() {
		cancel()
		_ = stream.Close()
		if !waitDone(done, time.Second) {
			t.Errorf("voiceblender uas event loop did not stop")
		}
	})
}

func voiceBlenderMaxDuration(hold time.Duration) int {
	if hold <= 0 {
		return 1
	}
	return int((hold + time.Second - time.Nanosecond) / time.Second)
}

func voiceBlenderTimeout(maxDuration int) time.Duration {
	if maxDuration <= 0 {
		return 60 * time.Second
	}
	return time.Duration(maxDuration)*time.Second + 10*time.Second
}
