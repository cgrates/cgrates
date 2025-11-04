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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
	"github.com/cgrates/go-diameter/diam/dict"
)

func TestDiamPrometheus(t *testing.T) {
	t.Skip("test by looking at the log output")
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigJSON: `{
"admins": {
	"enabled": true
},
"db": {
	"db_conns": {
		"*default": {
			"db_type": "*internal",
				"opts":{
				"internalDBRewriteInterval": "0s",
				"internalDBDumpInterval": "0s"
			}
    	}
	}
},
"sessions": {
	"enabled": true,
	"cdrs_conns": ["*internal"]
},
"cdrs": {
	"enabled": true,
	"store_cdrs": false
},
"prometheus_agent": {
	"enabled": true,
	"stats_conns": ["*localhost"],
	"stat_queue_ids": ["SQ_1","SQ_2"]
},
"stats": {
	"enabled": true,
	"store_interval": "-1"
},
"thresholds": {
	"enabled": true,
	"store_interval": "-1"
},
"rpc_conns": {
	"async": {
		"strategy": "*async",
		"conns": [
			{
				"address": "*internal"
			}
		]
	}
},
"diameter_agent": {
	"enabled": true,
	"sessions_conns": ["*birpc_internal"],
	"stats_conns": ["*localhost"],
	"thresholds_conns": ["*localhost"],
	"request_processors": [{
		"id": "message",
		"filters": [
			"*string:~*vars.*cmd:CCR",
			"*prefix:~*req.Service-Context-Id:message",
			"*string:~*req.CC-Request-Type:4"
		],
		"flags": [
			"*cdrs",
			"*daStats:SQ_1&SQ_2",
			// "*daThresholds:TH_1&TH_2",
		],
		"request_fields": [{
				"tag": "ToR",
				"path": "*cgreq.ToR",
				"type": "*constant",
				"value": "*sms"
			},
			{
				"tag": "OriginID",
				"path": "*cgreq.OriginID",
				"type": "*variable",
				"value": "~*req.Session-Id",
				"mandatory": true
			},
			{
				"tag": "Category",
				"path": "*cgreq.Category",
				"type": "*constant",
				"value": "sms"
			},
			{
				"tag": "RequestType",
				"path": "*cgreq.RequestType",
				"type": "*constant",
				"value": "*prepaid"
			},
			{
				"tag": "Account",
				"path": "*cgreq.Account",
				"type": "*variable",
				"mandatory": true,
				"value": "~*req.Subscription-Id.Subscription-Id-Data[~Subscription-Id-Type(0)]"
			},
			{
				"tag": "Destination",
				"path": "*cgreq.Destination",
				"type": "*variable",
				"mandatory": true,
				"value": "~*req.Service-Information.SMS-Information.Recipient-Address.Address-Data"
			},
			{
				"tag": "SetupTime",
				"path": "*cgreq.SetupTime",
				"type": "*variable",
				"value": "~*req.Event-Timestamp",
				"mandatory": true
			},
			{
				"tag": "AnswerTime",
				"path": "*cgreq.AnswerTime",
				"type": "*variable",
				"value": "~*req.Event-Timestamp",
				"mandatory": true
			},
			{
				"tag": "Usage",
				"path": "*cgreq.Usage",
				"type": "*variable",
				"value": "~*req.Requested-Service-Unit.CC-Time",
				"mandatory": true
			}
		],
		"reply_fields": [{
				"tag": "CCATemplate",
				"type": "*template",
				"value": "*cca"
			},
			{
				"tag": "ResultCode",
				"path": "*rep.Result-Code",
				"filters": ["*notempty:~*cgrep.Error:"],
				"type": "*constant",
				"value": "5030",
				"blocker": true
			}
		]
	}]
}
}`,
		DBCfg:            engine.InternalDBCfg,
		LogBuffer:        &bytes.Buffer{},
		GracefulShutdown: true,
		Encoding:         *utils.Encoding,
	}
	t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, cfg := ng.Run(t)

	var reply string

	if err := client.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
		engine.StatQueueProfileWithAPIOpts{
			StatQueueProfile: &engine.StatQueueProfile{
				Tenant:      "cgrates.org",
				ID:          "SQ_1",
				FilterIDs:   []string{"*string:~*req.Category:sms"},
				QueueLength: -1,
				TTL:         5 * time.Second,
				Metrics: []*engine.MetricWithFilters{
					{
						MetricID: "*average#~*req.ProcessingTime",
					},
					{
						MetricID: "*sum#~*req.ProcessingTime",
					},
					{
						MetricID: "*highest#~*req.ProcessingTime",
					},
					{
						MetricID: "*lowest#~*req.ProcessingTime",
					},
					{
						MetricID: "*distinct#~*req.ProcessingTime",
					},
				},
				Stored:   true,
				MinItems: 1,
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
		engine.StatQueueProfileWithAPIOpts{
			StatQueueProfile: &engine.StatQueueProfile{
				Tenant:      "cgrates.org",
				ID:          "SQ_2",
				FilterIDs:   []string{"*string:~*req.Category:sms"},
				QueueLength: -1,
				TTL:         10 * time.Second,
				Metrics: []*engine.MetricWithFilters{
					{
						MetricID: "*average#~*req.ProcessingTime",
					},
					{
						MetricID: "*sum#~*req.ProcessingTime",
					},
					{
						MetricID: "*highest#~*req.ProcessingTime",
					},
					{
						MetricID: "*lowest#~*req.ProcessingTime",
					},
					{
						MetricID: "*distinct#~*req.ProcessingTime",
					},
				},
				Stored:   true,
				MinItems: 1,
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)
	diamClient, err := NewDiameterClient(cfg.DiameterAgentCfg().Listen, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().ListenNet)
	if err != nil {
		t.Fatal(err)
	}

	for range 10 {
		time.Sleep(time.Second)
		sendDiamCCR(t, diamClient, 5*time.Second, "2001")
		scrapePromURL(t)
	}
}

func scrapePromURL(t *testing.T) {
	t.Helper()
	url := "http://localhost:2080/prometheus"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	bodyString := string(body)
	fmt.Println(bodyString)
}

// sendDiamCCR sends a CCR and verifies the expected result code, returning success status
func sendDiamCCR(tb testing.TB, client *DiameterClient, replyTimeout time.Duration, wantResultCode string) bool {
	tb.Helper()
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(utils.UUIDSha1Prefix()))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("message@DiamItCCRSMS"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(4))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 5, 11, 43, 10, 0, time.UTC)))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Enumerated(0))
	ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(1))}})
	ccr.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{ //
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("22509")), // Calling-Vlr-Number
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("4002")),  // Called-Party-NP
				},
			}),
			diam.NewAVP(2000, avp.Mbit, 10415, &diam.GroupedAVP{ // SMS-Information
				AVP: []*diam.AVP{
					diam.NewAVP(886, avp.Mbit, 10415, &diam.GroupedAVP{ // Originator-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1001")), // Address-Data
						}}),
					diam.NewAVP(1201, avp.Mbit, 10415, &diam.GroupedAVP{ // Recipient-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1003")), // Address-Data
						}}),
				},
			}),
		}})

	if err := client.SendMessage(ccr); err != nil {
		tb.Errorf("failed to send diameter message: %v", err)
		return false
	}

	reply := client.ReceivedMessage(replyTimeout)
	if reply == nil {
		tb.Error("received empty reply")
		return false
	}

	avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		tb.Error(err)
		return false
	}
	if len(avps) == 0 {
		tb.Error("missing AVPs in reply")
		return false
	}

	resultCode, err := diamAVPAsString(avps[0])
	if err != nil {
		tb.Error(err)
		return false
	}
	if resultCode != wantResultCode {
		tb.Errorf("Result-Code=%s, want %s", resultCode, wantResultCode)
		return false
	}
	return true
}
