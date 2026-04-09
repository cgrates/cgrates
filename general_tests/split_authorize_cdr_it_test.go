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

package general_tests

import (
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestBalanceSplitAuthorizeAndCDR(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{
"rals": {
	"enabled": true
},
"cdrs": {
	"enabled": true,
	"rals_conns": ["*internal"]
},
"schedulers": {
	"enabled": true
},
"sessions": {
	"enabled": true,
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
	"chargers_conns": ["*internal"]
},
"chargers": {
	"enabled": true
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
},
"http_agent": [
	{
		"id": "TestAgent",
		"url": "/test_charge",
		"sessions_conns": ["*internal"],
		"request_payload": "*url",
		"reply_payload": "*xml",
		"request_processors": [
			{
				"id": "auth",
				"flags": ["*authorize", "*accounts", "*continue"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.origin_id"},
					{"tag": "OriginHost", "path": "*cgreq.OriginHost", "type": "*constant", "value": "127.0.0.1"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*pseudoprepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*variable", "value": "~*req.category"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*constant", "value": "*now"},
					{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage"}
				],
				"reply_fields": [
					{"tag": "MaxUsage", "path": "*rep.MaxUsage", "type": "*variable", "value": "~*cgrep.MaxUsage"},
					{
						"tag": "ResultCode",
						"path": "*rep.ResultCode",
						"filters": ["*eq:~*cgrep.MaxUsage:0"],
						"type": "*constant",
						"value": "4012",
						"blocker": true
					},
					{
						"tag": "ResultCode",
						"path": "*rep.ResultCode",
						"filters": ["*notempty:~*cgrep.Error:"],
						"type": "*constant",
						"value": "5030",
						"blocker": true
					}
				]
			},
			{
				"id": "cdr",
				"flags": ["*cdrs"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.origin_id"},
					{"tag": "OriginHost", "path": "*cgreq.OriginHost", "type": "*constant", "value": "127.0.0.1"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*pseudoprepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*variable", "value": "~*req.category"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*constant", "value": "*now"},
					{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage"},
					{
						"tag": "UsageDenied",
						"path": "*cgreq.Usage",
						"filters": ["*eq:~*cgrep.MaxUsage:0"],
						"type": "*constant",
						"value": "0"
					}
				],
				"reply_fields": []
			}
		]
	}
]

}`

	tpFiles := map[string]string{
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,DEFAULT,*none,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_DATA,*any,RT_DATA,*up,20,,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_DATA,,0.5,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_DATA,DR_DATA,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,data,1001,2014-01-14T00:00:00Z,RP_DATA,
cgrates.org,data,1002,2014-01-14T00:00:00Z,RP_DATA,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      engine.InternalDBCfg,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)

	var setReply string
	if err := client.Call(context.Background(), utils.APIerSv1SetBalance,
		&utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1001",
			BalanceType: utils.MetaMonetary,
			Balance: map[string]any{
				utils.ID:         "balance_main",
				utils.Value:      5000,
				utils.Weight:     0,
				utils.Categories: "voice;sms",
			},
		}, &setReply); err != nil {
		t.Fatal(err)
	}
	if err := client.Call(context.Background(), utils.APIerSv1SetBalance,
		&utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1002",
			BalanceType: utils.MetaMonetary,
			Balance: map[string]any{
				utils.ID:         "balance_main",
				utils.Value:      5000,
				utils.Weight:     0,
				utils.Categories: "voice;sms",
			},
		}, &setReply); err != nil {
		t.Fatal(err)
	}
	if err := client.Call(context.Background(), utils.APIerSv1SetBalance,
		&utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1002",
			BalanceType: utils.MetaMonetary,
			Balance: map[string]any{
				utils.ID:         "balance_data",
				utils.Value:      100,
				utils.Weight:     20,
				utils.Categories: "data",
			},
		}, &setReply); err != nil {
		t.Fatal(err)
	}

	sendCharge := func(t *testing.T, originID, account, category, destination, usage string) {
		t.Helper()
		resp, err := http.DefaultClient.PostForm("http://127.0.0.1:2080/test_charge",
			url.Values{
				"origin_id":   {originID},
				"account":     {account},
				"category":    {category},
				"destination": {destination},
				"usage":       {usage},
			})
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("HTTP %d: %s", resp.StatusCode, string(body))
		}
	}

	t.Run("Denied", func(t *testing.T) {
		sendCharge(t, "denied_1", "1001", "data", "1002", "2s")
		time.Sleep(20 * time.Millisecond)

		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acnt); err != nil {
			t.Fatal(err)
		}
		for _, bal := range acnt.BalanceMap[utils.MetaMonetary] {
			if bal.ID == utils.MetaDefault {
				t.Errorf("*default balance should not exist, got value=%v", bal.Value)
			}
		}

		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				OriginIDs: []string{"denied_1"},
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("expected 1 CDR, got %d", len(cdrs))
		}
		if cdrs[0].Cost != 0 {
			t.Errorf("denied CDR cost: want 0, got %v", cdrs[0].Cost)
		}
	})

	t.Run("Allowed", func(t *testing.T) {
		sendCharge(t, "allowed_1", "1002", "data", "1002", "2s")
		time.Sleep(20 * time.Millisecond)

		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"}, &acnt); err != nil {
			t.Fatal(err)
		}
		for _, bal := range acnt.BalanceMap[utils.MetaMonetary] {
			switch bal.ID {
			case "balance_data":
				if bal.Value != 99 {
					t.Errorf("balance_data: want 99, got %v", bal.Value)
				}
			case utils.MetaDefault:
				t.Error("*default balance should not exist")
			}
		}

		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				OriginIDs: []string{"allowed_1"},
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("expected 1 CDR, got %d", len(cdrs))
		}
		if cdrs[0].Cost != 1 {
			t.Errorf("allowed CDR cost: want 1, got %v", cdrs[0].Cost)
		}
	})
}
