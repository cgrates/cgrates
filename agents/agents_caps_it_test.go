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
	"github.com/cgrates/radigo"
	"github.com/miekg/dns"
)

func TestAgentCapsIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
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
"sessions":{
	"enabled": true
},
"diameter_agent": {
	"enabled": true,
	"synced_conn_requests": true
},
"radius_agent": {
	"enabled": true
},
"dns_agent": {
	"enabled": true,
	"listeners":[
		{
			"address": "127.0.0.1:2053",
			"network": "udp"
		}
	]
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
]
}`

	ng := engine.TestEngine{
		ConfigJSON: jsonCfg,
		DBCfg:      engine.InternalDBCfg,
	}
	conn, cfg := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for services to start

	var i int

	t.Run("DiameterAgent", func(t *testing.T) {
		diamClient, err := NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
			cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
			cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
			cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
		if err != nil {
			t.Fatal(err)
		}

		// There is currently no traffic. Expecting Result-Code 5012 (DIAMETER_UNABLE_TO_COMPLY),
		// because there are no request processors enabled.
		sendCCR(t, diamClient, &i, "5012")

		// Caps limit is 2, therefore expecting the same result as in the scenario above.
		doneCh := simulateCapsTraffic(t, conn, 1, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond) // ensure traffic requests have been sent
		sendCCR(t, diamClient, &i, "5012")
		<-doneCh

		// With caps limit reached, Result-Code 3004 (DIAMETER_TOO_BUSY) is expected.
		doneCh = simulateCapsTraffic(t, conn, 2, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond) // ensure traffic requests have been sent
		sendCCR(t, diamClient, &i, "3004")
		<-doneCh

		// TODO: Check caps functionality with async diameter requests.
	})

	t.Run("RadiusAgent auth", func(t *testing.T) {
		radClient, err := radigo.NewClient(utils.UDP, "127.0.0.1:1812", "CGRateS.org", dictRad, 1, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		// There is currently no traffic. Expecting nil reply because
		// there are no request processors enabled.
		sendRadReq(t, radClient, radigo.AccessRequest, &i, radigo.AccessAccept)
		// Caps limit is 2, therefore expecting the same result as in
		// the scenario above.
		doneCh := simulateCapsTraffic(t, conn, 1, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond) // ensure traffic requests have been sent
		sendRadReq(t, radClient, radigo.AccessRequest, &i, radigo.AccessAccept)
		<-doneCh

		// With caps limit reached, Reply with Code 3 (AccessReject)
		// and ReplyMessage with the caps error is expected.
		doneCh = simulateCapsTraffic(t, conn, 2, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond) // ensure traffic requests have been sent
		sendRadReq(t, radClient, radigo.AccessRequest, &i, radigo.AccessReject)
		<-doneCh
	})

	t.Run("RadiusAgent acct", func(t *testing.T) {
		radClient, err := radigo.NewClient(utils.UDP, "127.0.0.1:1813", "CGRateS.org", dictRad, 1, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		// There is currently no traffic. Expecting nil reply because
		// there are no request processors enabled.
		sendRadReq(t, radClient, radigo.AccountingRequest, &i, 0)

		// Caps limit is 2, therefore expecting the same result as in
		// the scenario above.
		doneCh := simulateCapsTraffic(t, conn, 1, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond) // ensure traffic requests have been sent
		sendRadReq(t, radClient, radigo.AccountingRequest, &i, 0)
		<-doneCh

		// With caps limit reached, Reply with Code 5 (AccountingResponse)
		// and ReplyMessage with the caps error is expected.
		doneCh = simulateCapsTraffic(t, conn, 2, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond) // ensure traffic requests have been sent
		sendRadReq(t, radClient, radigo.AccountingRequest, &i, radigo.AccountingResponse)
		<-doneCh
	})

	t.Run("DNSAgent", func(t *testing.T) {
		client := new(dns.Client)
		dc, err := client.Dial(cfg.DNSAgentCfg().Listeners[0].Address)
		if err != nil {
			t.Fatal(err)
		}

		// There is currently no traffic. Expecting ServerFailure Rcode
		// because there are no request processors enabled.
		writeDNSMsg(t, dc, dns.RcodeServerFailure)

		// Caps limit is 2, therefore expecting the same result as in
		// the scenario above.
		doneCh := simulateCapsTraffic(t, conn, 1, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond) // ensure traffic requests have been sent
		writeDNSMsg(t, dc, dns.RcodeServerFailure)
		<-doneCh

		// With caps limit reached, Refused Rcode is expected.
		doneCh = simulateCapsTraffic(t, conn, 2, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond) // ensure traffic requests have been sent
		writeDNSMsg(t, dc, dns.RcodeRefused)
		<-doneCh

	})

	t.Run("HTTPAgent", func(t *testing.T) {
		httpURL := fmt.Sprintf("http://%s/caps_test", cfg.ListenCfg().HTTPListen)

		// There is currently no traffic. Expecting 200 OK because
		// there are no request processors enabled (empty reply).
		sendHTTPReq(t, httpURL, http.StatusOK)

		// Caps limit is 2, therefore expecting the same result.
		doneCh := simulateCapsTraffic(t, conn, 1, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond)
		sendHTTPReq(t, httpURL, http.StatusOK)
		<-doneCh

		// With caps limit reached, 429 Too Many Requests is expected.
		doneCh = simulateCapsTraffic(t, conn, 2, *cfg.CoreSCfg())
		time.Sleep(time.Millisecond)
		sendHTTPReq(t, httpURL, http.StatusTooManyRequests)
		<-doneCh
	})
}

func sendCCR(t *testing.T, client *DiameterClient, reqIdx *int, wantResultCode string) {
	*reqIdx++
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(fmt.Sprintf("session%d", reqIdx)))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))

	if err := client.SendMessage(ccr); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
		return
	}

	reply := client.ReceivedMessage(2 * time.Second)
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
	}
	if resultCode != wantResultCode {
		t.Errorf("Result-Code=%s, want %s", resultCode, wantResultCode)
	}
}

func sendRadReq(t *testing.T, client *radigo.Client, reqType radigo.PacketCode, reqIdx *int, wantReplyCode radigo.PacketCode) {
	*reqIdx++
	req := client.NewRequest(reqType, uint8(*reqIdx))
	if err := req.AddAVPWithName("User-Name", "1001", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("User-Password", "CGRateSPassword1", ""); err != nil {
		t.Error(err)
	}
	// encode the password as required so we can decode it properly
	req.AVPs[1].RawValue = radigo.EncodeUserPassword([]byte("CGRateSPassword1"), []byte("CGRateS.org"), req.Authenticator[:])
	if err := req.AddAVPWithName("Service-Type", "SIP-Caller-AVPs", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Called-Station-Id", "1002", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Acct-Session-Id", fmt.Sprintf("session%d", reqIdx), ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("NAS-IP-Address", "127.0.0.1", ""); err != nil {
		t.Error(err)
	}
	reply, err := client.SendRequest(req)
	if err != nil && (wantReplyCode == radigo.AccessReject ||
		wantReplyCode == radigo.AccountingResponse) {
		t.Error(err)
	}
	if reply != nil && reply.Code != wantReplyCode {
		t.Errorf("want non-nil negative reply, got: %s", utils.ToJSON(reply))
	}
	if reply != nil && reply.Code == wantReplyCode {
		if len(reply.AVPs) != 1 {
			t.Errorf("reply should have exactly 1 AVP, got: %s", utils.ToJSON(reply))
		}
		got := string(reply.AVPs[0].RawValue)
		want := utils.ErrMaxConcurrentRPCExceededNoCaps.Error()
		if got != want {
			t.Errorf("ReplyMessage=%v, want %v", got, want)
		}
	}
}

func writeDNSMsg(t *testing.T, conn *dns.Conn, wantRcode int) {
	m := new(dns.Msg)
	m.SetQuestion("cgrates.org.", dns.TypeA)
	if err := conn.WriteMsg(m); err != nil {
		t.Error(err)
	}
	if rply, err := conn.ReadMsg(); err != nil {
		t.Error(err)
	} else if rply.Rcode != wantRcode {
		t.Errorf("reply Msg Rcode=%d, want %d", rply.Rcode, wantRcode)
	}
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
