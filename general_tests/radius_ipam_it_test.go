//go:build integration
// +build integration

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

package general_tests

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

func TestRadiusIPAM(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	var testRadiusDict = `
ATTRIBUTE	Framed-Pool		88		string

VALUE	Service-Type		Framed				2
VALUE	Framed-Protocol		PPP					1
VALUE	Framed-Protocol		GPRS-PDP-Context	7
VALUE	NAS-Port-Type		Virtual				5
VALUE	Acct-Status-Type	Start				1
VALUE	Acct-Status-Type	Stop				2
VALUE	Acct-Status-Type	Alive				3
VALUE	Acct-Authentic		RADIUS				1
VALUE	Acct-Terminate-Cause	User-Request	1
`

	dictDir := t.TempDir()
	dictPath := filepath.Join(dictDir, "dictionary.test")
	if err := os.WriteFile(dictPath, []byte(testRadiusDict), 0644); err != nil {
		t.Fatal(err)
	}

	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "radius_ipam"),
		ConfigJSON: fmt.Sprintf(`{
"radius_agent": {
	"client_dictionaries": {
		"*default": [
			%q
		]
	}
}
}`, dictDir+"/"),
		DBCfg:     engine.InternalDBCfg,
		LogBuffer: &bytes.Buffer{},
	}
	t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, cfg := ng.Run(t)

	var replySetCharger string
	if err := client.Call(context.Background(), utils.APIerSv1SetChargerProfile,
		&engine.ChargerProfileWithAPIOpts{
			ChargerProfile: &engine.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "DEFAULT",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{utils.MetaNone},
			},
		}, &replySetCharger); err != nil {
		t.Fatal(err)
	}

	ippID := "IMSI_123456789012345"
	var replySet string
	if err := client.Call(context.Background(), utils.APIerSv1SetIPProfile,
		&engine.IPProfileWithAPIOpts{
			IPProfile: &engine.IPProfile{
				Tenant:    "cgrates.org",
				ID:        ippID,
				FilterIDs: []string{"*string:~*req.IMSI:123456789012345"},
				TTL:       -1,
				Pools: []*engine.IPPool{
					{
						ID:       "DEFAULT",
						Type:     "*ipv4",
						Range:    "10.100.0.1/32", // Single IP to ensure rejection scenario
						Strategy: "*ascending",
						Message:  "Default IP pool",
					},
				},
			},
		}, &replySet); err != nil {
		t.Error(err)
	}

	imsi := "123456789012345"
	msisdn := "987654321098765"
	apn := "internet.test.apn"

	nasID := "test-nas-server-1"
	nasIP := "192.168.1.10"
	poolName := "test-pool-primary"

	passwd := "CGRateSPassword1"
	currentTimestamp := fmt.Sprintf("%d", time.Now().Unix())

	authSessionID := "auth-session-12345-67890"
	acctSessionID := "acct-session-abcdef-123456"

	proxyAuth := "4829"
	proxyAuthReject := "4830"
	proxyAcctStart := "4831"
	proxyAcctAlive := "4832"
	proxyAcctStop := "4833"

	// Step 1: Access-Request (should not allocate)
	dictRad := radigo.RFC2865Dictionary()
	dictRad.ParseFromReader(strings.NewReader(testRadiusDict))
	secret := cfg.RadiusAgentCfg().ClientSecrets[utils.MetaDefault]
	net := cfg.RadiusAgentCfg().Listeners[0].Network
	authAddr := cfg.RadiusAgentCfg().Listeners[0].AuthAddr
	clientAuth, err := radigo.NewClient(net, authAddr, secret, dictRad, 1, nil, utils.Logger)
	if err != nil {
		t.Fatal(err)
	}

	reply := sendRadReq(t, clientAuth, radigo.AccessRequest, 1,
		map[string]string{
			"User-Name":          imsi,
			"Service-Type":       "Framed",
			"Framed-Protocol":    "GPRS-PDP-Context",
			"Called-Station-Id":  apn,
			"Calling-Station-Id": msisdn,
			"NAS-Identifier":     nasID,
			"Acct-Session-Id":    authSessionID,
			"Framed-Pool":        poolName,
			"User-Password":      passwd,
			"Event-Timestamp":    currentTimestamp,
			"NAS-IP-Address":     nasIP,
			"Proxy-State":        proxyAuth,
		}, radigo.AccessAccept,
	)
	checkAllocs(t, client, ippID)
	checkActiveSessions(t, client, 0)

	// retrieve allocatedIP (to be used in Accounting-Request Start)
	var allocatedIP string
	for _, avp := range reply.AVPs {
		if avp.Number == 8 { // Framed-IP-Address
			if len(avp.RawValue) == 4 {
				allocatedIP = fmt.Sprintf("%d.%d.%d.%d",
					avp.RawValue[0], avp.RawValue[1],
					avp.RawValue[2], avp.RawValue[3])
				break
			}
		}
	}
	if allocatedIP != "10.100.0.1" {
		t.Errorf("expected IP from DEFAULT pool (10.100.0.1), got %s", allocatedIP)
	}

	// Step 2: Accounting-Request Start (should allocate)
	acctAddr := cfg.RadiusAgentCfg().Listeners[0].AcctAddr
	clientAcct, err := radigo.NewClient(net, acctAddr, secret, dictRad, 1, nil, utils.Logger)
	if err != nil {
		t.Fatal(err)
	}

	sendRadReq(t, clientAcct, radigo.AccountingRequest, 2,
		map[string]string{
			"User-Name":          imsi,
			"Acct-Status-Type":   "Start",
			"NAS-Identifier":     nasID,
			"Called-Station-Id":  apn,
			"Framed-Protocol":    "GPRS-PDP-Context",
			"Service-Type":       "Framed",
			"NAS-Port-Type":      "Virtual",
			"Calling-Station-Id": msisdn,
			"Acct-Authentic":     "RADIUS",
			"Acct-Delay-Time":    "0",
			"Acct-Session-Id":    acctSessionID,
			"Framed-IP-Address":  allocatedIP,
			"NAS-IP-Address":     nasIP,
			"Event-Timestamp":    currentTimestamp,
			"Proxy-State":        proxyAcctStart,
		}, radigo.AccountingResponse,
	)
	checkAllocs(t, client, ippID, acctSessionID)
	checkActiveSessions(t, client, 1)

	// Step 2.5: Send another Access-Request after IP allocation
	// This should receive Access-Reject since IP is already allocated.
	rejectSessID := "reject-session-98765"
	rejectTimestamp := fmt.Sprintf("%d", time.Now().Unix())

	rejectReply := sendRadReq(t, clientAuth, radigo.AccessRequest, 5,
		map[string]string{
			"User-Name":          imsi,
			"Service-Type":       "Framed",
			"Framed-Protocol":    "GPRS-PDP-Context",
			"Called-Station-Id":  apn,
			"Calling-Station-Id": msisdn,
			"NAS-Identifier":     nasID,
			"Acct-Session-Id":    rejectSessID,
			"Framed-Pool":        poolName,
			"User-Password":      passwd,
			"Event-Timestamp":    rejectTimestamp,
			"NAS-IP-Address":     nasIP,
			"Proxy-State":        proxyAuthReject,
		}, radigo.AccessReject)

	// Verify Access-Reject response contains error message and no IP address.
	var hasReplyMessage, hasFramedIP bool
	var replyMessage string
	for _, avp := range rejectReply.AVPs {
		if avp.Number == 18 { // Reply-Message
			hasReplyMessage = true
			replyMessage = string(avp.RawValue)
		}
		if avp.Number == 8 { // Framed-IP-Address
			hasFramedIP = true
		}
	}
	if !hasReplyMessage || replyMessage == "" {
		t.Errorf("Access-Reject should contain Reply-Message with error, got: %q", replyMessage)
	}
	if hasFramedIP {
		t.Error("Access-Reject should not contain Framed-IP-Address")
	}
	if !strings.Contains(replyMessage, "IP_UNAUTHORIZED") {
		t.Errorf("Reply-Message should contain IP_UNAUTHORIZED error, got: %q", replyMessage)
	}

	// Step 3: Accounting-Request Alive (should maintain allocation)
	time.Sleep(100 * time.Millisecond)
	aliveTimestamp := fmt.Sprintf("%d", time.Now().Unix())
	sendRadReq(t, clientAcct, radigo.AccountingRequest, 3,
		map[string]string{
			"User-Name":          imsi,
			"Acct-Status-Type":   "Alive",
			"Service-Type":       "Framed",
			"Acct-Session-Id":    acctSessionID,
			"Framed-Protocol":    "GPRS-PDP-Context",
			"Called-Station-Id":  apn,
			"Calling-Station-Id": msisdn,
			"NAS-Identifier":     nasID,
			"Acct-Input-Octets":  "1234567",
			"Acct-Output-Octets": "7654321",
			"NAS-IP-Address":     nasIP,
			"Event-Timestamp":    aliveTimestamp,
			"Proxy-State":        proxyAcctAlive,
		}, radigo.AccountingResponse,
	)
	checkAllocs(t, client, ippID, acctSessionID)

	// Step 4: Accounting-Request Stop (should release)
	time.Sleep(100 * time.Millisecond)
	stopTimestamp := fmt.Sprintf("%d", time.Now().Unix())
	sendRadReq(t, clientAcct, radigo.AccountingRequest, 4,
		map[string]string{
			"User-Name":            imsi,
			"Acct-Status-Type":     "Stop",
			"Service-Type":         "Framed",
			"Acct-Session-Id":      acctSessionID,
			"Framed-Protocol":      "GPRS-PDP-Context",
			"Called-Station-Id":    apn,
			"Calling-Station-Id":   msisdn,
			"NAS-Identifier":       nasID,
			"Acct-Input-Octets":    "9876543",
			"Acct-Output-Octets":   "1234567",
			"Acct-Terminate-Cause": "User-Request",
			"NAS-IP-Address":       nasIP,
			"Event-Timestamp":      stopTimestamp,
			"Proxy-State":          proxyAcctStop,
		}, radigo.AccountingResponse,
	)
	time.Sleep(time.Second)
	checkAllocs(t, client, ippID)
	checkActiveSessions(t, client, 0)
	checkCDR(t, client, imsi)
}

func sendRadReq(t *testing.T, client *radigo.Client, code radigo.PacketCode, id uint8, avps map[string]string, expectedCode radigo.PacketCode) *radigo.Packet {
	t.Helper()
	req := client.NewRequest(code, id)

	for attr, val := range avps {
		if err := req.AddAVPWithName(attr, val, ""); err != nil {
			t.Fatal(err)
		}
		if code == radigo.AccessRequest && attr == "User-Password" {
			secret := []byte("CGRateS.org")
			for i := len(req.AVPs) - 1; i >= 0; i-- {
				if req.AVPs[i].Name == "User-Password" {
					req.AVPs[i].RawValue = radigo.EncodeUserPassword([]byte(val), secret, req.Authenticator[:])
					break
				}
			}
		}
	}

	replyPacket, err := client.SendRequest(req)
	if err != nil {
		t.Fatal(err)
	}

	if replyPacket.Code != expectedCode {
		t.Errorf("expected reply code %s, got %s for request %s: %+v", expectedCode.String(), replyPacket.Code.String(), code.String(), utils.ToJSON(replyPacket))
	}

	return replyPacket
}

func checkAllocs(tb testing.TB, client *birpc.Client, id string, wantAllocs ...string) {
	tb.Helper()
	var allocs engine.IPAllocations
	if err := client.Call(context.Background(), utils.IPsV1GetIPAllocations,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     id,
			},
		}, &allocs); err != nil {
		tb.Fatalf("Failed to get IP allocations for %s: %v", id, err)
	}
	if len(allocs.Allocations) != len(wantAllocs) {
		tb.Errorf("%s unexpected result: %s", utils.IPsV1GetIPAllocations, utils.ToJSON(allocs))
	}

	for _, allocID := range wantAllocs {
		if _, exists := allocs.Allocations[allocID]; !exists {
			tb.Errorf("%s unexpected result: %s", utils.IPsV1GetIPAllocations, utils.ToJSON(allocs))
			return
		}
	}
}

func allocateIP(tb testing.TB, client *birpc.Client, eventID, id, allocID string) {
	tb.Helper()
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     eventID,
		Event: map[string]any{
			utils.AccountField: id,
			utils.AnswerTime:   utils.TimePointer(time.Now()),
			utils.Usage:        10,
			utils.Tenant:       "cgrates.org",
		},
		APIOpts: map[string]any{
			utils.OptsIPsAllocationID: allocID,
		},
	}
	var reply engine.AllocatedIP
	if err := client.Call(context.Background(), utils.IPsV1AllocateIP, args, &reply); err != nil {
		tb.Fatalf("Failed to allocate IP for profile %s with allocation ID %s: %v", id, allocID, err)
	}
}

func releaseIP(tb testing.TB, client *birpc.Client, id, allocID string) {
	tb.Helper()
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.GenUUID(),
		Event: map[string]any{
			utils.AccountField: id,
			utils.AnswerTime:   utils.TimePointer(time.Now()),
			utils.Usage:        10,
			utils.Tenant:       "cgrates.org",
		},
		APIOpts: map[string]any{
			utils.OptsIPsAllocationID: allocID,
		},
	}
	if err := client.Call(context.Background(), utils.IPsV1ReleaseIP, args, nil); err != nil {
		tb.Errorf("Error releasing IPProfile %s: %v", id, err)
	}
}

func checkCDR(tb testing.TB, client *birpc.Client, acnt string) {
	tb.Helper()
	var cdrs []*engine.ExternalCDR
	if err := client.Call(context.Background(), utils.APIerSv1GetCDRs,
		&utils.AttrGetCdrs{
			Accounts: []string{acnt},
		}, &cdrs); err != nil {
		tb.Fatal(err)
	}
	if len(cdrs) != 1 {
		tb.Fatalf("%s received %d cdrs, want exactly one", utils.APIerSv1GetCDRs, len(cdrs))
	}
	tb.Logf("CDR contents: %s", utils.ToIJSON(cdrs[0]))
}

func checkActiveSessions(tb testing.TB, client *birpc.Client, wantCount int) {
	tb.Helper()
	var sessions []*sessions.ExternalSession
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		&utils.SessionFilter{}, &sessions); err != nil {
		if wantCount == 0 && err.Error() == utils.ErrNotFound.Error() {
			tb.Logf("no active sessions found (expected)")
			return
		}
		tb.Fatalf("failed to get active sessions: %v", err)
	}
	if len(sessions) != wantCount {
		tb.Fatalf("%s received %d sessions, want exactly %d",
			utils.SessionSv1GetActiveSessions, len(sessions), wantCount)
	}
	tb.Logf("%s reply: %s", utils.SessionSv1GetActiveSessions, utils.ToIJSON(sessions))
}
