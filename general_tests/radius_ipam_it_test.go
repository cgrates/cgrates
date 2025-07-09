//go:build integration
// +build integration

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

package general_tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
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
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
		// LogBuffer:        &bytes.Buffer{},
	}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, cfg := ng.Run(t)

	var replySet string
	if err := client.Call(context.Background(), utils.AdminSv1SetIPProfile,
		&utils.IPProfileWithAPIOpts{
			IPProfile: &utils.IPProfile{
				Tenant:    "cgrates.org",
				ID:        "IPsAPI",
				FilterIDs: []string{"*string:~*req.Account:123456789012345"},
				Weights: utils.DynamicWeights{
					{
						Weight: 15,
					},
				},
				TTL:    -1,
				Stored: false,

				// Pool selection logic:
				// POOL_A (10.100.0.1): weight 50, blocks (APN=internet.test.apn)
				// POOL_B (10.100.0.2): weight 30, gets removed
				// POOL_C (10.100.0.3): weight 100 should win
				Pools: []*utils.IPPool{
					{
						ID:        "POOL_A",
						FilterIDs: []string{},
						Type:      "*ipv4",
						Range:     "10.100.0.1/32",
						Strategy:  "*ascending",
						Message:   "Pool A message",
						Weights: utils.DynamicWeights{
							{
								FilterIDs: []string{},
								Weight:    50,
							},
						},
						Blockers: utils.DynamicBlockers{
							{
								FilterIDs: []string{"*string:~*req.APN:internet.test.apn"},
								Blocker:   true,
							},
						},
					},
					{
						ID:        "POOL_B",
						FilterIDs: []string{},
						Type:      "*ipv4",
						Range:     "10.100.0.2/32",
						Strategy:  "*ascending",
						Message:   "Pool B message",
						Weights: utils.DynamicWeights{
							{
								FilterIDs: []string{},
								Weight:    30,
							},
						},
					},
					{
						ID:        "POOL_C",
						FilterIDs: []string{},
						Type:      "*ipv4",
						Range:     "10.100.0.3/32",
						Strategy:  "*ascending",
						Message:   "Pool C message",
						Weights: utils.DynamicWeights{
							{
								FilterIDs: []string{"*string:~*req.APN:internet.test.apn"},
								Weight:    100,
							},
							{
								FilterIDs: []string{},
								Weight:    10,
							},
						},
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
	proxyAcctStart := "4830"
	proxyAcctAlive := "4831"
	proxyAcctStop := "4832"

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
		},
	)
	checkAllocs(t, client, "IPsAPI")

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
	if allocatedIP != "10.100.0.3" {
		t.Errorf("expected IP from POOL_C (10.100.0.3), got %s", allocatedIP)
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
			"NAS-Identifier":     "test-nas-server-2", // Different NAS for accounting
			"Called-Station-Id":  apn,
			"Framed-Protocol":    "GPRS-PDP-Context",
			"Service-Type":       "Framed",
			"NAS-Port-Type":      "Virtual",
			"Calling-Station-Id": msisdn,
			"Acct-Authentic":     "RADIUS",
			"Acct-Delay-Time":    "0",
			"Acct-Session-Id":    acctSessionID,
			"Framed-IP-Address":  allocatedIP,
			"NAS-IP-Address":     "192.168.1.11", // Different NAS IP for accounting
			"Event-Timestamp":    currentTimestamp,
			"Proxy-State":        proxyAcctStart,
		},
	)
	checkAllocs(t, client, "IPsAPI", acctSessionID)

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
			"NAS-Identifier":     "test-nas-server-3",
			"Acct-Input-Octets":  "1234567",
			"Acct-Output-Octets": "7654321",
			"NAS-IP-Address":     "192.168.1.12",
			"Event-Timestamp":    aliveTimestamp,
			"Proxy-State":        proxyAcctAlive,
		},
	)
	checkAllocs(t, client, "IPsAPI", acctSessionID)

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
			"NAS-Identifier":       "test-nas-server-3",
			"Acct-Input-Octets":    "9876543",
			"Acct-Output-Octets":   "1234567",
			"Acct-Terminate-Cause": "User-Request",
			"NAS-IP-Address":       "192.168.1.12",
			"Event-Timestamp":      stopTimestamp,
			"Proxy-State":          proxyAcctStop,
		},
	)
	checkAllocs(t, client, "IPsAPI")
	checkCDR(t, client, imsi)
}

func sendRadReq(t *testing.T, client *radigo.Client, code radigo.PacketCode, id uint8, avps map[string]string) *radigo.Packet {
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

	var failed bool
	switch code {
	case radigo.AccessRequest:
		failed = replyPacket.Code != radigo.AccessAccept
	case radigo.AccountingRequest:
		failed = replyPacket.Code != radigo.AccountingResponse
	}
	if failed {
		t.Errorf("unexpected reply received to %s: %+v", code.String(), utils.ToJSON(replyPacket))
	}

	return replyPacket
}

func checkAllocs(t *testing.T, client *birpc.Client, id string, wantAllocs ...string) {
	t.Helper()
	var allocs utils.IPAllocations
	if err := client.Call(context.Background(), utils.IPsV1GetIPAllocations,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     id,
			},
		}, &allocs); err != nil {
		t.Error(err)
	}
	if len(allocs.Allocations) != len(wantAllocs) {
		t.Errorf("%s unexpected result: %s", utils.IPsV1GetIPAllocations, utils.ToJSON(allocs))
	}

	for _, allocID := range wantAllocs {
		if _, exists := allocs.Allocations[allocID]; !exists {
			t.Errorf("%s unexpected result: %s", utils.IPsV1GetIPAllocations, utils.ToJSON(allocs))
			return
		}
	}
}

func checkCDR(t *testing.T, client *birpc.Client, acnt string) {
	t.Helper()
	var cdrs []*utils.CDR
	if err := client.Call(context.Background(), utils.AdminSv1GetCDRs,
		&utils.CDRFilters{
			FilterIDs: []string{
				fmt.Sprintf("*string:~*req.Account:%s", acnt),
			},
		}, &cdrs); err != nil {
		t.Fatal(err)
	}
	if len(cdrs) != 1 {
		t.Fatalf("%s received %d cdrs, want exactly one", utils.AdminSv1GetCDRs, len(cdrs))
	}
	t.Logf("CDR contents: %s", utils.ToIJSON(cdrs[0]))
}
