//go:build integration

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
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
	"github.com/cgrates/go-diameter/diam/dict"
	"github.com/cgrates/sipingo"
)

func TestAgentCapsIT(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.InternalDBCfg
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	jsonCfg := `{
"cores": {
	"caps": 2,
	"caps_strategy": "*busy",
	"shutdown_timeout": "5ms"
},
"db": {
	"db_conns": {
		"*default": {
			"db_type": "*internal",
			"opts":{
			"internalDBDumpInterval": "0",	
			"internalDBRewriteInterval": "0"
	}
    	}
	},
},
"sessions":{
	"enabled": true
},
"diameter_agent": {
	"enabled": true,
	"synced_conn_requests": true
},
"http_agent": [
	{
		"id": "caps_test",
		"url": "/caps_test",
		"sessions_conns": ["*internal"],
		"request_payload": "*url",
		"reply_payload": "*xml",
		"request_processors": []
	}
],
"sip_agent": {
	"enabled": true,
	"listen": "127.0.0.1:5099",
	"listen_net": "udp",
	"sessions_conns": ["*internal"],
	"request_processors": []
}
}`

	ng := engine.TestEngine{
		ConfigJSON: jsonCfg,
		DBCfg:      dbCfg,
		Encoding:   *utils.Encoding,
	}
	client, cfg := ng.Run(t)

	time.Sleep(10 * time.Millisecond) // wait for DiameterAgent service to start
	diamClient, err := NewDiameterClient(cfg.DiameterAgentCfg().Listen, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().ListenNet)
	if err != nil {
		t.Fatal(err)
	}

	reqIdx := 0
	sendCCR := func(t *testing.T, replyTimeout time.Duration, wg *sync.WaitGroup, wantResultCode string) {
		t.Helper()
		if wg != nil {
			defer wg.Done()
		}
		reqIdx++
		ccr := diam.NewRequest(diam.CreditControl, 4, nil)
		ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(fmt.Sprintf("session%d", reqIdx)))
		ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
		ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
		ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
		ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
		ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))

		if err := diamClient.SendMessage(ccr); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
			return
		}

		reply := diamClient.ReceivedMessage(replyTimeout)
		if reply == nil {
			t.Error("received empty reply")
			return
		}

		avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
		if err != nil {
			t.Error(err)
			return
		}
		if len(avps) == 0 {
			t.Error("missing AVPs in reply")
			return
		}

		resultCode, err := diamAVPAsString(avps[0])
		if err != nil {
			t.Error(err)
			return
		}
		if resultCode != wantResultCode {
			t.Errorf("Result-Code=%s, want %s", resultCode, wantResultCode)
		}
	}

	// There is currently no traffic. Expecting Result-Code 5012 (DIAMETER_UNABLE_TO_COMPLY),
	// because there are no request processors enabled.
	diamReplyTimeout := 2 * time.Second
	sendCCR(t, diamReplyTimeout, nil, "5012")

	// Caps limit is 2, therefore expecting the same result as in the scenario above.
	doneCh := simulateCapsTraffic(t, client, 1, *cfg.CoreSCfg())
	time.Sleep(time.Millisecond) // ensure traffic requests have been sent
	sendCCR(t, diamReplyTimeout, nil, "5012")
	<-doneCh

	// With caps limit reached, Result-Code 3004 (DIAMETER_TOO_BUSY) is expected.
	doneCh = simulateCapsTraffic(t, client, 2, *cfg.CoreSCfg())
	time.Sleep(time.Millisecond) // ensure traffic requests have been sent
	sendCCR(t, diamReplyTimeout, nil, "3004")
	<-doneCh

	// TODO: Check caps functionality with async diameter requests.

	t.Run("HTTPAgent", func(t *testing.T) {
		httpURL := fmt.Sprintf("http://%s/caps_test", cfg.ListenCfg().HTTPListen)

		// There is currently no traffic. Expecting 200 OK because
		// there are no request processors enabled (empty reply).
		sendHTTPReq(t, httpURL, http.StatusOK)

		// Caps limit is 2, therefore expecting the same result.
		doneCh := simulateCapsTraffic(t, client, 1, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond)
		sendHTTPReq(t, httpURL, http.StatusOK)
		<-doneCh

		// With caps limit reached, 429 Too Many Requests is expected.
		doneCh = simulateCapsTraffic(t, client, 2, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond)
		sendHTTPReq(t, httpURL, http.StatusTooManyRequests)
		<-doneCh
	})

	t.Run("SIPAgent", func(t *testing.T) {
		sipAddr := cfg.SIPAgentCfg().Listen

		// There is currently no traffic. Expecting 500 Internal Server
		// Error because there are no request processors enabled.
		sendSIPReq(t, sipAddr, "SIP/2.0 500 Internal Server Error")

		// Caps limit is 2, therefore expecting the same result.
		doneCh := simulateCapsTraffic(t, client, 1, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond)
		sendSIPReq(t, sipAddr, "SIP/2.0 500 Internal Server Error")
		<-doneCh

		// With caps limit reached, 503 Service Unavailable is expected.
		doneCh = simulateCapsTraffic(t, client, 2, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond)
		sendSIPReq(t, sipAddr, "SIP/2.0 503 Service Unavailable")
		<-doneCh
	})
}

func sendHTTPReq(t *testing.T, url string, wantStatus int) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != wantStatus {
		t.Errorf("HTTP status=%d, want %d", resp.StatusCode, wantStatus)
	}
}

func sendSIPReq(t *testing.T, addr, wantStatus string) {
	t.Helper()
	conn, err := net.Dial("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	invite := "INVITE sip:1002@cgrates.org SIP/2.0\r\n" +
		"Call-ID: caps-test-" + fmt.Sprint(time.Now().UnixNano()) + "\r\n" +
		"CSeq: 1 INVITE\r\n" +
		"From: \"1001\" <sip:1001@cgrates.org>;tag=caps1\r\n" +
		"To: <sip:1002@cgrates.org>\r\n" +
		"Via: SIP/2.0/UDP 127.0.0.1:9999;branch=z9hG4bK-caps-test\r\n" +
		"Max-Forwards: 70\r\n" +
		"Content-Length: 0\r\n\r\n"
	if _, err = conn.Write([]byte(invite)); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, bufferSize)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	received, err := sipingo.NewMessage(string(buf[:n]))
	if err != nil {
		t.Fatal(err)
	}
	if received[requestHeader] != wantStatus {
		t.Errorf("SIP status=%q, want %q", received[requestHeader], wantStatus)
	}
}

func simulateCapsTraffic(t *testing.T, client *birpc.Client, amount int, coresCfg config.CoreSCfg) <-chan struct{} {
	t.Helper()
	var wg sync.WaitGroup
	var reply string
	for i := range amount {
		wg.Add(1)
		go func() {
			t.Helper()
			defer wg.Done()
			if err := client.Call(context.Background(), utils.CoreSv1Sleep,
				&utils.DurationArgs{
					// Use the ShutdownTimeout CoreS setting
					// instead of having to pass an extra
					// variable to the function.
					Duration: coresCfg.ShutdownTimeout,
				}, &reply); err != nil {
				if coresCfg.CapsStrategy == utils.MetaBusy && i >= coresCfg.Caps {
					return // no need to handle errors for this scenario
				}
				t.Errorf("CoreSv1.Sleep unexpected error: %v", err)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}
