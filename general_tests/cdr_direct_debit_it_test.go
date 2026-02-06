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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	eCnt = "*default,processCDR1,*data,*%s,cgrates.org,call,1001,1001,1002,2021-02-02T16:14:50Z,2021-02-02T16:15:00Z,0.000000008,8,\"{\"\"CGRID\"\":\"\"d3f870282f93ede55132845d88d742330806c0ce\"\",\"\"RunID\"\":\"\"*default\"\",\"\"StartTime\"\":\"\"2021-02-02T16:15:00Z\"\",\"\"Usage\"\":8,\"\"Cost\"\":8,\"\"Charges\"\":[{\"\"RatingID\"\":\"\"0e7d75b\"\",\"\"Increments\"\":[{\"\"Usage\"\":1,\"\"Cost\"\":1,\"\"AccountingID\"\":\"\"8931065\"\",\"\"CompressFactor\"\":8}],\"\"CompressFactor\"\":1}],\"\"AccountSummary\"\":{\"\"Tenant\"\":\"\"cgrates.org\"\",\"\"ID\"\":\"\"1001\"\",\"\"BalanceSummaries\"\":[{\"\"UUID\"\":\"\"a6b84f43-18c8-4ae5-8195-271c0b8df6c5\"\",\"\"ID\"\":\"\"balance_monetary\"\",\"\"Type\"\":\"\"*monetary\"\",\"\"Initial\"\":10,\"\"Value\"\":2,\"\"Weight\"\":20,\"\"Disabled\"\":false}],\"\"AllowNegative\"\":false,\"\"Disabled\"\":false},\"\"Rating\"\":{\"\"0e7d75b\"\":{\"\"ConnectFee\"\":0,\"\"RoundingMethod\"\":\"\"*up\"\",\"\"RoundingDecimals\"\":20,\"\"MaxCost\"\":0,\"\"MaxCostStrategy\"\":\"\"\"\",\"\"TimingID\"\":\"\"09f4ea8\"\",\"\"RatesID\"\":\"\"7d7c1f6\"\",\"\"RatingFiltersID\"\":\"\"7a46d1c\"\"}},\"\"Accounting\"\":{\"\"8931065\"\":{\"\"AccountID\"\":\"\"cgrates.org:1001\"\",\"\"BalanceUUID\"\":\"\"a6b84f43-18c8-4ae5-8195-271c0b8df6c5\"\",\"\"RatingID\"\":\"\"\"\",\"\"Units\"\":1,\"\"ExtraChargeID\"\":\"\"\"\"}},\"\"RatingFilters\"\":{\"\"7a46d1c\"\":{\"\"DestinationID\"\":\"\"*any\"\",\"\"DestinationPrefix\"\":\"\"*any\"\",\"\"RatingPlanID\"\":\"\"RP_ANY\"\",\"\"Subject\"\":\"\"*out:cgrates.org:call:1001\"\"}},\"\"Rates\"\":{\"\"7d7c1f6\"\":[{\"\"GroupIntervalStart\"\":0,\"\"Value\"\":1,\"\"RateIncrement\"\":1,\"\"RateUnit\"\":1}]},\"\"Timings\"\":{\"\"09f4ea8\"\":{\"\"Years\"\":[],\"\"Months\"\":[],\"\"MonthDays\"\":[],\"\"WeekDays\"\":[],\"\"StartTime\"\":\"\"00:00:00\"\"}}}\"\n"

	eCnt2Bal = "*default,processCDR1,*data,*%s,cgrates.org,call,1001,1001,1002,2021-02-02T16:14:50Z,2021-02-02T16:15:00Z,0.000000008,8,\"{\"\"CGRID\"\":\"\"d3f870282f93ede55132845d88d742330806c0ce\"\",\"\"RunID\"\":\"\"*default\"\",\"\"StartTime\"\":\"\"2021-02-02T16:15:00Z\"\",\"\"Usage\"\":8,\"\"Cost\"\":8,\"\"Charges\"\":[{\"\"RatingID\"\":\"\"72fce0f\"\",\"\"Increments\"\":[{\"\"Usage\"\":1,\"\"Cost\"\":1,\"\"AccountingID\"\":\"\"7f7aac3\"\",\"\"CompressFactor\"\":5}],\"\"CompressFactor\"\":1},{\"\"RatingID\"\":\"\"72fce0f\"\",\"\"Increments\"\":[{\"\"Usage\"\":1,\"\"Cost\"\":1,\"\"AccountingID\"\":\"\"d71f2ca\"\",\"\"CompressFactor\"\":3}],\"\"CompressFactor\"\":1}],\"\"AccountSummary\"\":{\"\"Tenant\"\":\"\"cgrates.org\"\",\"\"ID\"\":\"\"1001\"\",\"\"BalanceSummaries\"\":[{\"\"UUID\"\":\"\"e006fa7e-82c6-45b8-a5a6-e1efc3ca7fc4\"\",\"\"ID\"\":\"\"balance_monetary\"\",\"\"Type\"\":\"\"*monetary\"\",\"\"Initial\"\":5,\"\"Value\"\":0,\"\"Weight\"\":21,\"\"Disabled\"\":false},{\"\"UUID\"\":\"\"25d78cc4-5200-4cd7-a786-9cb6e0881296\"\",\"\"ID\"\":\"\"balance_monetary2\"\",\"\"Type\"\":\"\"*monetary\"\",\"\"Initial\"\":5,\"\"Value\"\":2,\"\"Weight\"\":20,\"\"Disabled\"\":false}],\"\"AllowNegative\"\":false,\"\"Disabled\"\":false},\"\"Rating\"\":{\"\"72fce0f\"\":{\"\"ConnectFee\"\":0,\"\"RoundingMethod\"\":\"\"*up\"\",\"\"RoundingDecimals\"\":20,\"\"MaxCost\"\":0,\"\"MaxCostStrategy\"\":\"\"\"\",\"\"TimingID\"\":\"\"1201b07\"\",\"\"RatesID\"\":\"\"330aabd\"\",\"\"RatingFiltersID\"\":\"\"32c8a01\"\"}},\"\"Accounting\"\":{\"\"7f7aac3\"\":{\"\"AccountID\"\":\"\"cgrates.org:1001\"\",\"\"BalanceUUID\"\":\"\"e006fa7e-82c6-45b8-a5a6-e1efc3ca7fc4\"\",\"\"RatingID\"\":\"\"\"\",\"\"Units\"\":1,\"\"ExtraChargeID\"\":\"\"\"\"},\"\"d71f2ca\"\":{\"\"AccountID\"\":\"\"cgrates.org:1001\"\",\"\"BalanceUUID\"\":\"\"25d78cc4-5200-4cd7-a786-9cb6e0881296\"\",\"\"RatingID\"\":\"\"\"\",\"\"Units\"\":1,\"\"ExtraChargeID\"\":\"\"\"\"}},\"\"RatingFilters\"\":{\"\"32c8a01\"\":{\"\"DestinationID\"\":\"\"*any\"\",\"\"DestinationPrefix\"\":\"\"*any\"\",\"\"RatingPlanID\"\":\"\"RP_ANY\"\",\"\"Subject\"\":\"\"*out:cgrates.org:call:1001\"\"}},\"\"Rates\"\":{\"\"330aabd\"\":[{\"\"GroupIntervalStart\"\":0,\"\"Value\"\":1,\"\"RateIncrement\"\":1,\"\"RateUnit\"\":1}]},\"\"Timings\"\":{\"\"1201b07\"\":{\"\"Years\"\":[],\"\"Months\"\":[],\"\"MonthDays\"\":[],\"\"WeekDays\"\":[],\"\"StartTime\"\":\"\"00:00:00\"\"}}}\"\n"

	eCntInsBal = "*default,processCDR1,*data,*%s,cgrates.org,call,1001,1001,1002,2021-02-02T16:14:50Z,2021-02-02T16:15:00Z,0.000000008,8,\"{\"\"CGRID\"\":\"\"d3f870282f93ede55132845d88d742330806c0ce\"\",\"\"RunID\"\":\"\"*default\"\",\"\"StartTime\"\":\"\"2021-02-02T16:15:00Z\"\",\"\"Usage\"\":8,\"\"Cost\"\":8,\"\"Charges\"\":[{\"\"RatingID\"\":\"\"81a0b2c\"\",\"\"Increments\"\":[{\"\"Usage\"\":1,\"\"Cost\"\":1,\"\"AccountingID\"\":\"\"38d165f\"\",\"\"CompressFactor\"\":5}],\"\"CompressFactor\"\":1},{\"\"RatingID\"\":\"\"81a0b2c\"\",\"\"Increments\"\":[{\"\"Usage\"\":1,\"\"Cost\"\":1,\"\"AccountingID\"\":\"\"ec4f966\"\",\"\"CompressFactor\"\":3}],\"\"CompressFactor\"\":1}],\"\"AccountSummary\"\":{\"\"Tenant\"\":\"\"cgrates.org\"\",\"\"ID\"\":\"\"1001\"\",\"\"BalanceSummaries\"\":[{\"\"UUID\"\":\"\"740d46eb-1046-440c-9bfa-a11ff2ddf3c2\"\",\"\"ID\"\":\"\"balance_monetary\"\",\"\"Type\"\":\"\"*monetary\"\",\"\"Initial\"\":5,\"\"Value\"\":0,\"\"Weight\"\":20,\"\"Disabled\"\":false},{\"\"UUID\"\":\"\"72da89ad-6ab3-45c8-8b29-eb69bd70ba29\"\",\"\"ID\"\":\"\"*default\"\",\"\"Type\"\":\"\"*monetary\"\",\"\"Initial\"\":0,\"\"Value\"\":-3,\"\"Disabled\"\":false}],\"\"AllowNegative\"\":false,\"\"Disabled\"\":false},\"\"Rating\"\":{\"\"81a0b2c\"\":{\"\"ConnectFee\"\":0,\"\"RoundingMethod\"\":\"\"*up\"\",\"\"RoundingDecimals\"\":20,\"\"MaxCost\"\":0,\"\"MaxCostStrategy\"\":\"\"\"\",\"\"TimingID\"\":\"\"db90f3f\"\",\"\"RatesID\"\":\"\"5bb144b\"\",\"\"RatingFiltersID\"\":\"\"807538c\"\"}},\"\"Accounting\"\":{\"\"38d165f\"\":{\"\"AccountID\"\":\"\"cgrates.org:1001\"\",\"\"BalanceUUID\"\":\"\"740d46eb-1046-440c-9bfa-a11ff2ddf3c2\"\",\"\"RatingID\"\":\"\"\"\",\"\"Units\"\":1,\"\"ExtraChargeID\"\":\"\"\"\"},\"\"ec4f966\"\":{\"\"AccountID\"\":\"\"cgrates.org:1001\"\",\"\"BalanceUUID\"\":\"\"72da89ad-6ab3-45c8-8b29-eb69bd70ba29\"\",\"\"RatingID\"\":\"\"\"\",\"\"Units\"\":1,\"\"ExtraChargeID\"\":\"\"\"\"}},\"\"RatingFilters\"\":{\"\"807538c\"\":{\"\"DestinationID\"\":\"\"*any\"\",\"\"DestinationPrefix\"\":\"\"*any\"\",\"\"RatingPlanID\"\":\"\"RP_ANY\"\",\"\"Subject\"\":\"\"*out:cgrates.org:call:1001\"\"}},\"\"Rates\"\":{\"\"5bb144b\"\":[{\"\"GroupIntervalStart\"\":0,\"\"Value\"\":1,\"\"RateIncrement\"\":1,\"\"RateUnit\"\":1}]},\"\"Timings\"\":{\"\"db90f3f\"\":{\"\"Years\"\":[],\"\"Months\"\":[],\"\"MonthDays\"\":[],\"\"WeekDays\"\":[],\"\"StartTime\"\":\"\"00:00:00\"\"}}}\"\n"
)

func TestCDRDebitPostpaid(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7,
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"rals": {
	"enabled": true
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*localhost"],
	"ees_conns": ["*localhost"]
},

"ees": {
	"enabled": true,
	"cache": {
		"*file_csv": {"limit": -1, "ttl": "500ms", "static_ttl": false},
	},
	"exporters": [
	{
        "id": "CSVExporter",
        "type": "*file_csv",
        "export_path": "/tmp/exportedCDRs",
	    "attempts": 1,
        "synchronous": true,
	    "flags": ["*log"],
        "fields":[
			{"tag": "RunID", "path": "*exp.RunID", "type": "*variable", "value": "~*req.RunID"},
            {"tag": "OriginID", "path": "*exp.OriginID", "type": "*variable", "value": "~*req.OriginID", "mandatory": true},
            {"tag": "ToR", "path": "*exp.ToR", "type": "*variable", "value": "~*req.ToR", "mandatory": true},
            {"tag": "RequestType", "path": "*exp.RequestType", "type": "*variable", "value": "~*req.RequestType", "mandatory": true},
            {"tag": "Tenant", "path": "*exp.Tenant", "type": "*variable", "value": "~*req.Tenant", "mandatory": true},
            {"tag": "Category", "path": "*exp.Category", "type": "*variable", "value": "~*req.Category", "mandatory": true},
            {"tag": "Account", "path": "*exp.Account", "type": "*variable", "value": "~*req.Account", "mandatory": true},
            {"tag": "Subject", "path": "*exp.Subject", "type": "*variable", "value": "~*req.Subject", "mandatory": true},
            {"tag": "Destination", "path": "*exp.Destination", "type": "*variable", "value": "~*req.Destination", "mandatory": true},
            {"tag": "SetupTime", "path": "*exp.SetupTime", "type": "*variable", "value": "~*req.SetupTime{*timestring:UTC:2006-01-02T15:04:05Z}" , "mandatory": true},
            {"tag": "AnswerTime", "path": "*exp.AnswerTime", "type": "*variable", "value": "~*req.AnswerTime{*timestring:UTC:2006-01-02T15:04:05Z}", "mandatory": true},
            {"tag": "Usage", "path": "*exp.Usage", "type": "*variable", "value": "~*req.Usage{*duration_seconds}", "mandatory": true},
			{"tag": "Cost", "path": "*exp.Cost", "type": "*variable", "value": "~*req.Cost{*round:4}"},
            {"tag": "CostDetails", "path": "*exp.CostDetails", "type": "*variable", "value": "~*req.CostDetails", "mandatory": true}
		],
	},
	]
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
	"ees_conns": ["*internal"],
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,,,,*unlimited,,10,20,false,false,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	if err := os.MkdirAll("/tmp/exportedCDRs", 0755); err != nil {
		t.Fatal("Error creating folder /tmp/exportedCDRs: ", "/tmp/exportedCDRs", err)
	}
	time.Sleep(100 * time.Millisecond)
	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
		// LogBuffer:  new(bytes.Buffer),
	}
	t.Cleanup(func() {
		// fmt.Println(ng.LogBuffer)
		if err := os.RemoveAll("/tmp/exportedCDRs"); err != nil {
			t.Fatal("Error removing folder /tmp/exportedCDRs: ", err)
		}
	})
	client, _ := ng.Run(t)
	time.Sleep(200 * time.Millisecond)

	t.Run("CheckInitialBalance", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
			t.Fatalf("expected account to have one balance of type *monetary, received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	t.Run("ProcessCDR1", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs, utils.MetaStore},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "*default",
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaData,
						utils.OriginID:     "processCDR1",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaPostpaid,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        8,
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("expected to receive only one CDR: %v", utils.ToJSON(cdrs))
		}
		monetaryBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[0]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *monetary balance current value: %v", err)
		}
		if monetaryBalanceValue != 2. {
			t.Errorf("unexpected balance value: expected %v, received %v", 2., monetaryBalanceValue)
		}
	})

	t.Run("CheckFinalBalance", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
			t.Fatalf("expected account to have one balance of type *monetary , received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 2 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	time.Sleep(1 * time.Second)

	t.Run("TestReadExportedCDRs", func(t *testing.T) {
		var files []string
		err := filepath.Walk("/tmp/exportedCDRs", func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, utils.CSVSuffix) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Error(err)
		}
		if len(files) != 1 {
			t.Errorf("Expected %+v, received: %+v", 1, len(files))
		}
		eCnt := fmt.Sprintf(eCnt, "postpaid")
		if outContent1, err := os.ReadFile(files[0]); err != nil {
			t.Error(err)
		} else if len(eCnt) != len(string(outContent1)) {
			t.Errorf("Expecting: \n<%+v>, \nreceived: \n<%+v>", len(eCnt), len(string(outContent1)))
			t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
		}
	})
}

// the same test as Postpaid but this time Cost is included in the CDR and RequestType is directdebit
func TestCDRDebitDirect1Balance(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7,
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"rals": {
	"enabled": true
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*localhost"],
	"ees_conns": ["*localhost"]
},

"ees": {
	"enabled": true,
	"cache": {
		"*file_csv": {"limit": -1, "ttl": "500ms", "static_ttl": false},
	},
	"exporters": [
	{
        "id": "CSVExporter",
        "type": "*file_csv",
        "export_path": "/tmp/exportedCDRs",
	    "attempts": 1,
        "synchronous": true,
	    "flags": ["*log"],
        "fields":[
			{"tag": "RunID", "path": "*exp.RunID", "type": "*variable", "value": "~*req.RunID"},
            {"tag": "OriginID", "path": "*exp.OriginID", "type": "*variable", "value": "~*req.OriginID", "mandatory": true},
            {"tag": "ToR", "path": "*exp.ToR", "type": "*variable", "value": "~*req.ToR", "mandatory": true},
            {"tag": "RequestType", "path": "*exp.RequestType", "type": "*variable", "value": "~*req.RequestType", "mandatory": true},
            {"tag": "Tenant", "path": "*exp.Tenant", "type": "*variable", "value": "~*req.Tenant", "mandatory": true},
            {"tag": "Category", "path": "*exp.Category", "type": "*variable", "value": "~*req.Category", "mandatory": true},
            {"tag": "Account", "path": "*exp.Account", "type": "*variable", "value": "~*req.Account", "mandatory": true},
            {"tag": "Subject", "path": "*exp.Subject", "type": "*variable", "value": "~*req.Subject", "mandatory": true},
            {"tag": "Destination", "path": "*exp.Destination", "type": "*variable", "value": "~*req.Destination", "mandatory": true},
            {"tag": "SetupTime", "path": "*exp.SetupTime", "type": "*variable", "value": "~*req.SetupTime{*timestring:UTC:2006-01-02T15:04:05Z}" , "mandatory": true},
            {"tag": "AnswerTime", "path": "*exp.AnswerTime", "type": "*variable", "value": "~*req.AnswerTime{*timestring:UTC:2006-01-02T15:04:05Z}", "mandatory": true},
            {"tag": "Usage", "path": "*exp.Usage", "type": "*variable", "value": "~*req.Usage{*duration_seconds}", "mandatory": true},
			{"tag": "Cost", "path": "*exp.Cost", "type": "*variable", "value": "~*req.Cost{*round:4}"},
            {"tag": "CostDetails", "path": "*exp.CostDetails", "type": "*variable", "value": "~*req.CostDetails", "mandatory": true}
		],
	},
	]
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
	"ees_conns": ["*internal"],
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,,,,*unlimited,,10,20,false,false,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	if err := os.MkdirAll("/tmp/exportedCDRs", 0755); err != nil {
		t.Fatal("Error creating folder /tmp/exportedCDRs: ", "/tmp/exportedCDRs", err)
	}
	time.Sleep(100 * time.Millisecond)
	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
		// LogBuffer:  new(bytes.Buffer),
	}
	t.Cleanup(func() {
		// fmt.Println(ng.LogBuffer)
		if err := os.RemoveAll("/tmp/exportedCDRs"); err != nil {
			t.Fatal("Error removing folder /tmp/exportedCDRs: ", err)
		}
	})
	client, _ := ng.Run(t)
	time.Sleep(200 * time.Millisecond)

	t.Run("CheckInitialBalance", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
			t.Fatalf("expected account to have one balance of type *monetary, received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	t.Run("ProcessCDR1", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs, utils.MetaStore},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "*default",
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaData,
						utils.OriginID:     "processCDR1",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaDirectDebit,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        8,
						utils.Cost:         8,
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("expected to receive only one CDR: %v", utils.ToJSON(cdrs))
		}
		monetaryBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[0]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *monetary balance current value: %v", err)
		}
		if monetaryBalanceValue != 2. {
			t.Errorf("unexpected balance value: expected %v, received %v", 2., monetaryBalanceValue)
		}
	})

	t.Run("CheckFinalBalance", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
			t.Fatalf("expected account to have one balance of type *monetary , received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 2 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	time.Sleep(1 * time.Second)

	t.Run("TestReadExportedCDRs", func(t *testing.T) {
		var files []string
		err := filepath.Walk("/tmp/exportedCDRs", func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, utils.CSVSuffix) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Error(err)
		}
		if len(files) != 1 {
			t.Errorf("Expected %+v, received: %+v", 1, len(files))
		}
		eCnt := fmt.Sprintf(eCnt, "directdebit")
		if outContent1, err := os.ReadFile(files[0]); err != nil {
			t.Error(err)
		} else if len(eCnt) != len(string(outContent1)) {
			t.Errorf("Expecting: \n<%+v>, \nreceived: \n<%+v>", len(eCnt), len(string(outContent1)))
			t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
		}
	})
}

func TestCDRDebitDirectFail(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7,
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"rals": {
	"enabled": true
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*localhost"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,,,,*unlimited,,10,20,false,false,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}
	time.Sleep(100 * time.Millisecond)
	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
		// LogBuffer:  new(bytes.Buffer),
	}
	// t.Cleanup(func() {
	// 	fmt.Println(ng.LogBuffer)
	// })
	client, _ := ng.Run(t)
	time.Sleep(200 * time.Millisecond)

	t.Run("CheckInitialBalance", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
			t.Fatalf("expected account to have one balance of type *monetary, received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	t.Run("ProcessCDR1", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "*default",
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaData,
						utils.OriginID:     "processCDR1",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaDirectDebit,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        8,
					}, // missing cost field should fail
				},
			}, &reply); err == nil || err.Error() != "MANDATORY_IE_MISSING: [Cost]" {
			t.Fatalf("Expected error <MANDATORY_IE_MISSING: [Cost]>, received <%v>", err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{}, &cdrs); err == nil || err.Error() != "SERVER_ERROR: NOT_FOUND" {
			t.Fatalf("Expected error <SERVER_ERROR: NOT_FOUND>, received <%v>", err)
		}
	})

	t.Run("CheckFinalBalance", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
			t.Fatalf("expected account to have one balance of type *monetary , received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value !=
			10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

}

func TestCDRPostpaidMultipleBalances(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7,
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"rals": {
	"enabled": true
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*localhost"],
	"ees_conns": ["*localhost"]
},

"ees": {
	"enabled": true,
	"cache": {
		"*file_csv": {"limit": -1, "ttl": "500ms", "static_ttl": false},
	},
	"exporters": [
	{
        "id": "CSVExporter",
        "type": "*file_csv",
        "export_path": "/tmp/exportedCDRs",
	    "attempts": 1,
        "synchronous": true,
	    "flags": ["*log"],
        "fields":[
			{"tag": "RunID", "path": "*exp.RunID", "type": "*variable", "value": "~*req.RunID"},
            {"tag": "OriginID", "path": "*exp.OriginID", "type": "*variable", "value": "~*req.OriginID", "mandatory": true},
            {"tag": "ToR", "path": "*exp.ToR", "type": "*variable", "value": "~*req.ToR", "mandatory": true},
            {"tag": "RequestType", "path": "*exp.RequestType", "type": "*variable", "value": "~*req.RequestType", "mandatory": true},
            {"tag": "Tenant", "path": "*exp.Tenant", "type": "*variable", "value": "~*req.Tenant", "mandatory": true},
            {"tag": "Category", "path": "*exp.Category", "type": "*variable", "value": "~*req.Category", "mandatory": true},
            {"tag": "Account", "path": "*exp.Account", "type": "*variable", "value": "~*req.Account", "mandatory": true},
            {"tag": "Subject", "path": "*exp.Subject", "type": "*variable", "value": "~*req.Subject", "mandatory": true},
            {"tag": "Destination", "path": "*exp.Destination", "type": "*variable", "value": "~*req.Destination", "mandatory": true},
            {"tag": "SetupTime", "path": "*exp.SetupTime", "type": "*variable", "value": "~*req.SetupTime{*timestring:UTC:2006-01-02T15:04:05Z}" , "mandatory": true},
            {"tag": "AnswerTime", "path": "*exp.AnswerTime", "type": "*variable", "value": "~*req.AnswerTime{*timestring:UTC:2006-01-02T15:04:05Z}", "mandatory": true},
            {"tag": "Usage", "path": "*exp.Usage", "type": "*variable", "value": "~*req.Usage{*duration_seconds}", "mandatory": true},
			{"tag": "Cost", "path": "*exp.Cost", "type": "*variable", "value": "~*req.Cost{*round:4}"},
            {"tag": "CostDetails", "path": "*exp.CostDetails", "type": "*variable", "value": "~*req.CostDetails", "mandatory": true}
		],
	},
	]
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
	"ees_conns": ["*internal"],
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,,,,*unlimited,,5,21,false,false,21
ACT_TOPUP,*topup_reset,,,balance_monetary2,*monetary,,,,,*unlimited,,5,20,false,false,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	if err := os.MkdirAll("/tmp/exportedCDRs", 0755); err != nil {
		t.Fatal("Error creating folder /tmp/exportedCDRs: ", "/tmp/exportedCDRs", err)
	}
	time.Sleep(100 * time.Millisecond)
	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
		// LogBuffer:  new(bytes.Buffer),
	}
	t.Cleanup(func() {
		// fmt.Println(ng.LogBuffer)
		if err := os.RemoveAll("/tmp/exportedCDRs"); err != nil {
			t.Fatal("Error removing folder /tmp/exportedCDRs: ", err)
		}
	})
	client, _ := ng.Run(t)
	time.Sleep(200 * time.Millisecond)

	t.Run("CheckInitialBalance", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 2 balances of type *monetary, received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 5 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "balance_monetary2" || monetaryBalance.Value != 5 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	t.Run("ProcessCDR1", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs, utils.MetaStore},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "*default",
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaData,
						utils.OriginID:     "processCDR1",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaPostpaid,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        8,
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("expected to receive only one CDR: %v", utils.ToJSON(cdrs))
		}
		monetaryBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[0]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *monetary balance current value: %v", err)
		}
		if monetaryBalanceValue != 0. {
			t.Errorf("unexpected balance value: expected %v, received %v", 0., monetaryBalanceValue)
		}
		monetaryBalanceValue, err = cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[1]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *monetary balance current value: %v", err)
		}
		if monetaryBalanceValue != 2. {
			t.Errorf("unexpected balance value: expected %v, received %v", 2., monetaryBalanceValue)
		}
	})

	t.Run("CheckFinalBalance", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 2 balances of type *monetary , received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 0 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "balance_monetary2" || monetaryBalance.Value != 2 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	time.Sleep(1 * time.Second)

	t.Run("TestReadExportedCDRs", func(t *testing.T) {
		var files []string
		err := filepath.Walk("/tmp/exportedCDRs", func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, utils.CSVSuffix) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Error(err)
		}
		if len(files) != 1 {
			t.Errorf("Expected %+v, received: %+v", 1, len(files))
		}
		eCnt2Bal := fmt.Sprintf(eCnt2Bal, "postpaid")
		if outContent1, err := os.ReadFile(files[0]); err != nil {
			t.Error(err)
		} else if len(eCnt2Bal) != len(string(outContent1)) {
			t.Errorf("Expecting: \n<%+v>, \nreceived: \n<%+v>", len(eCnt), len(string(outContent1)))
			t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
		}
	})
}

func TestCDRDebitDirectMultipleBalances(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7,
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"rals": {
	"enabled": true
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*localhost"],
	"ees_conns": ["*localhost"]
},

"ees": {
	"enabled": true,
	"cache": {
		"*file_csv": {"limit": -1, "ttl": "500ms", "static_ttl": false},
	},
	"exporters": [
	{
        "id": "CSVExporter",
        "type": "*file_csv",
        "export_path": "/tmp/exportedCDRs",
	    "attempts": 1,
        "synchronous": true,
	    "flags": ["*log"],
        "fields":[
			{"tag": "RunID", "path": "*exp.RunID", "type": "*variable", "value": "~*req.RunID"},
            {"tag": "OriginID", "path": "*exp.OriginID", "type": "*variable", "value": "~*req.OriginID", "mandatory": true},
            {"tag": "ToR", "path": "*exp.ToR", "type": "*variable", "value": "~*req.ToR", "mandatory": true},
            {"tag": "RequestType", "path": "*exp.RequestType", "type": "*variable", "value": "~*req.RequestType", "mandatory": true},
            {"tag": "Tenant", "path": "*exp.Tenant", "type": "*variable", "value": "~*req.Tenant", "mandatory": true},
            {"tag": "Category", "path": "*exp.Category", "type": "*variable", "value": "~*req.Category", "mandatory": true},
            {"tag": "Account", "path": "*exp.Account", "type": "*variable", "value": "~*req.Account", "mandatory": true},
            {"tag": "Subject", "path": "*exp.Subject", "type": "*variable", "value": "~*req.Subject", "mandatory": true},
            {"tag": "Destination", "path": "*exp.Destination", "type": "*variable", "value": "~*req.Destination", "mandatory": true},
            {"tag": "SetupTime", "path": "*exp.SetupTime", "type": "*variable", "value": "~*req.SetupTime{*timestring:UTC:2006-01-02T15:04:05Z}" , "mandatory": true},
            {"tag": "AnswerTime", "path": "*exp.AnswerTime", "type": "*variable", "value": "~*req.AnswerTime{*timestring:UTC:2006-01-02T15:04:05Z}", "mandatory": true},
            {"tag": "Usage", "path": "*exp.Usage", "type": "*variable", "value": "~*req.Usage{*duration_seconds}", "mandatory": true},
			{"tag": "Cost", "path": "*exp.Cost", "type": "*variable", "value": "~*req.Cost{*round:4}"},
            {"tag": "CostDetails", "path": "*exp.CostDetails", "type": "*variable", "value": "~*req.CostDetails", "mandatory": true}
		],
	},
	]
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
	"ees_conns": ["*internal"],
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,,,,*unlimited,,5,21,false,false,21
ACT_TOPUP,*topup_reset,,,balance_monetary2,*monetary,,,,,*unlimited,,5,20,false,false,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	if err := os.MkdirAll("/tmp/exportedCDRs", 0755); err != nil {
		t.Fatal("Error creating folder /tmp/exportedCDRs: ", "/tmp/exportedCDRs", err)
	}
	time.Sleep(100 * time.Millisecond)
	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
		// LogBuffer:  new(bytes.Buffer),
	}
	t.Cleanup(func() {
		// fmt.Println(ng.LogBuffer)
		if err := os.RemoveAll("/tmp/exportedCDRs"); err != nil {
			t.Fatal("Error removing folder /tmp/exportedCDRs: ", err)
		}
	})
	client, _ := ng.Run(t)
	time.Sleep(200 * time.Millisecond)

	t.Run("CheckInitialBalance", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 2 balances of type *monetary, received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 5 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "balance_monetary2" || monetaryBalance.Value != 5 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	t.Run("ProcessCDR1", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs, utils.MetaStore},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "*default",
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaData,
						utils.OriginID:     "processCDR1",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaDirectDebit,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        8,
						utils.Cost:         8,
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("expected to receive only one CDR: %v", utils.ToJSON(cdrs))
		}
		monetaryBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[0]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *monetary balance current value: %v", err)
		}
		if monetaryBalanceValue != 0. {
			t.Errorf("unexpected balance value: expected %v, received %v", 0., monetaryBalanceValue)
		}
		monetaryBalanceValue, err = cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[1]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *monetary balance current value: %v", err)
		}
		if monetaryBalanceValue != 2. {
			t.Errorf("unexpected balance value: expected %v, received %v", 2., monetaryBalanceValue)
		}
	})

	t.Run("CheckFinalBalance", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 2 balances of type *monetary , received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 0 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "balance_monetary2" || monetaryBalance.Value != 2 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	time.Sleep(1 * time.Second)

	t.Run("TestReadExportedCDRs", func(t *testing.T) {
		var files []string
		err := filepath.Walk("/tmp/exportedCDRs", func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, utils.CSVSuffix) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Error(err)
		}
		if len(files) != 1 {
			t.Errorf("Expected %+v, received: %+v", 1, len(files))
		}
		eCnt2Bal := fmt.Sprintf(eCnt2Bal, "directdebit")
		if outContent1, err := os.ReadFile(files[0]); err != nil {
			t.Error(err)
		} else if len(eCnt2Bal) != len(string(outContent1)) {
			t.Errorf("Expecting: \n<%+v>, \nreceived: \n<%+v>", len(eCnt2Bal), len(string(outContent1)))
			t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt2Bal, string(outContent1))
		}
	})
}

func TestCDRPostpaidInsufficientBalances(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7,
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"rals": {
	"enabled": true
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*localhost"],
	"ees_conns": ["*localhost"]
},

"ees": {
	"enabled": true,
	"cache": {
		"*file_csv": {"limit": -1, "ttl": "500ms", "static_ttl": false},
	},
	"exporters": [
	{
        "id": "CSVExporter",
        "type": "*file_csv",
        "export_path": "/tmp/exportedCDRs",
	    "attempts": 1,
        "synchronous": true,
	    "flags": ["*log"],
        "fields":[
			{"tag": "RunID", "path": "*exp.RunID", "type": "*variable", "value": "~*req.RunID"},
            {"tag": "OriginID", "path": "*exp.OriginID", "type": "*variable", "value": "~*req.OriginID", "mandatory": true},
            {"tag": "ToR", "path": "*exp.ToR", "type": "*variable", "value": "~*req.ToR", "mandatory": true},
            {"tag": "RequestType", "path": "*exp.RequestType", "type": "*variable", "value": "~*req.RequestType", "mandatory": true},
            {"tag": "Tenant", "path": "*exp.Tenant", "type": "*variable", "value": "~*req.Tenant", "mandatory": true},
            {"tag": "Category", "path": "*exp.Category", "type": "*variable", "value": "~*req.Category", "mandatory": true},
            {"tag": "Account", "path": "*exp.Account", "type": "*variable", "value": "~*req.Account", "mandatory": true},
            {"tag": "Subject", "path": "*exp.Subject", "type": "*variable", "value": "~*req.Subject", "mandatory": true},
            {"tag": "Destination", "path": "*exp.Destination", "type": "*variable", "value": "~*req.Destination", "mandatory": true},
            {"tag": "SetupTime", "path": "*exp.SetupTime", "type": "*variable", "value": "~*req.SetupTime{*timestring:UTC:2006-01-02T15:04:05Z}" , "mandatory": true},
            {"tag": "AnswerTime", "path": "*exp.AnswerTime", "type": "*variable", "value": "~*req.AnswerTime{*timestring:UTC:2006-01-02T15:04:05Z}", "mandatory": true},
            {"tag": "Usage", "path": "*exp.Usage", "type": "*variable", "value": "~*req.Usage{*duration_seconds}", "mandatory": true},
			{"tag": "Cost", "path": "*exp.Cost", "type": "*variable", "value": "~*req.Cost{*round:4}"},
            {"tag": "CostDetails", "path": "*exp.CostDetails", "type": "*variable", "value": "~*req.CostDetails", "mandatory": true}
		],
	},
	]
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
	"ees_conns": ["*internal"],
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,,,,*unlimited,,5,20,false,false,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	if err := os.MkdirAll("/tmp/exportedCDRs", 0755); err != nil {
		t.Fatal("Error creating folder /tmp/exportedCDRs: ", "/tmp/exportedCDRs", err)
	}
	time.Sleep(100 * time.Millisecond)
	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
		// LogBuffer:  new(bytes.Buffer),
	}
	t.Cleanup(func() {
		// fmt.Println(ng.LogBuffer)
		if err := os.RemoveAll("/tmp/exportedCDRs"); err != nil {
			t.Fatal("Error removing folder /tmp/exportedCDRs: ", err)
		}
	})
	client, _ := ng.Run(t)
	time.Sleep(200 * time.Millisecond)

	t.Run("CheckInitialBalance", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
			t.Fatalf("expected account to have 1 balances of type *monetary, received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 5 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	t.Run("ProcessCDR1", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs, utils.MetaStore},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "*default",
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaData,
						utils.OriginID:     "processCDR1",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaPostpaid,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        8,
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("expected to receive only one CDR: %v", utils.ToJSON(cdrs))
		}
		monetaryBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[0]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *monetary balance current value: %v", err)
		}
		if monetaryBalanceValue != 0. {
			t.Errorf("unexpected balance value: expected %v, received %v", 0., monetaryBalanceValue)
		}
		monetaryBalanceValue, err = cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[1]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *monetary balance current value: %v", err)
		}
		if monetaryBalanceValue != -3. {
			t.Errorf("unexpected balance value: expected %v, received %v", -3, monetaryBalanceValue)
		}
	})

	t.Run("CheckFinalBalance", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 2 balances of type *monetary , received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 0 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != utils.MetaDefault || monetaryBalance.Value != -3 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	time.Sleep(1 * time.Second)

	t.Run("TestReadExportedCDRs", func(t *testing.T) {
		var files []string
		err := filepath.Walk("/tmp/exportedCDRs", func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, utils.CSVSuffix) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Error(err)
		}
		if len(files) != 1 {
			t.Errorf("Expected %+v, received: %+v", 1, len(files))
		}
		eCntInsBal := fmt.Sprintf(eCntInsBal, "postpaid")
		if outContent1, err := os.ReadFile(files[0]); err != nil {
			t.Error(err)
		} else if len(eCntInsBal) != len(string(outContent1)) {
			t.Errorf("Expecting: \n<%+v>, \nreceived: \n<%+v>", len(eCnt), len(string(outContent1)))
			t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
		}
	})
}

func TestCDRDebitDirectInsufficientBalance(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7,
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"rals": {
	"enabled": true
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*localhost"],
	"ees_conns": ["*localhost"]
},

"ees": {
	"enabled": true,
	"cache": {
		"*file_csv": {"limit": -1, "ttl": "500ms", "static_ttl": false},
	},
	"exporters": [
	{
        "id": "CSVExporter",
        "type": "*file_csv",
        "export_path": "/tmp/exportedCDRs",
	    "attempts": 1,
        "synchronous": true,
	    "flags": ["*log"],
        "fields":[
			{"tag": "RunID", "path": "*exp.RunID", "type": "*variable", "value": "~*req.RunID"},
            {"tag": "OriginID", "path": "*exp.OriginID", "type": "*variable", "value": "~*req.OriginID", "mandatory": true},
            {"tag": "ToR", "path": "*exp.ToR", "type": "*variable", "value": "~*req.ToR", "mandatory": true},
            {"tag": "RequestType", "path": "*exp.RequestType", "type": "*variable", "value": "~*req.RequestType", "mandatory": true},
            {"tag": "Tenant", "path": "*exp.Tenant", "type": "*variable", "value": "~*req.Tenant", "mandatory": true},
            {"tag": "Category", "path": "*exp.Category", "type": "*variable", "value": "~*req.Category", "mandatory": true},
            {"tag": "Account", "path": "*exp.Account", "type": "*variable", "value": "~*req.Account", "mandatory": true},
            {"tag": "Subject", "path": "*exp.Subject", "type": "*variable", "value": "~*req.Subject", "mandatory": true},
            {"tag": "Destination", "path": "*exp.Destination", "type": "*variable", "value": "~*req.Destination", "mandatory": true},
            {"tag": "SetupTime", "path": "*exp.SetupTime", "type": "*variable", "value": "~*req.SetupTime{*timestring:UTC:2006-01-02T15:04:05Z}" , "mandatory": true},
            {"tag": "AnswerTime", "path": "*exp.AnswerTime", "type": "*variable", "value": "~*req.AnswerTime{*timestring:UTC:2006-01-02T15:04:05Z}", "mandatory": true},
            {"tag": "Usage", "path": "*exp.Usage", "type": "*variable", "value": "~*req.Usage{*duration_seconds}", "mandatory": true},
			{"tag": "Cost", "path": "*exp.Cost", "type": "*variable", "value": "~*req.Cost{*round:4}"},
            {"tag": "CostDetails", "path": "*exp.CostDetails", "type": "*variable", "value": "~*req.CostDetails", "mandatory": true}
		],
	},
	]
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
	"ees_conns": ["*internal"],
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,,,,*unlimited,,5,20,false,false,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	if err := os.MkdirAll("/tmp/exportedCDRs", 0755); err != nil {
		t.Fatal("Error creating folder /tmp/exportedCDRs: ", "/tmp/exportedCDRs", err)
	}
	time.Sleep(100 * time.Millisecond)
	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
		// LogBuffer:  new(bytes.Buffer),
	}
	t.Cleanup(func() {
		// fmt.Println(ng.LogBuffer)
		if err := os.RemoveAll("/tmp/exportedCDRs"); err != nil {
			t.Fatal("Error removing folder /tmp/exportedCDRs: ", err)
		}
	})
	client, _ := ng.Run(t)
	time.Sleep(200 * time.Millisecond)

	t.Run("CheckInitialBalance", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
			t.Fatalf("expected account to have 1 balances of type *monetary, received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 5 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	t.Run("ProcessCDR1", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs, utils.MetaStore},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "*default",
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaData,
						utils.OriginID:     "processCDR1",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaDirectDebit,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        8,
						utils.Cost:         8,
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("expected to receive only one CDR: %v", utils.ToJSON(cdrs))
		}
		monetaryBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[0]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *monetary balance current value: %v", err)
		}
		if monetaryBalanceValue != 0. {
			t.Errorf("unexpected balance value: expected %v, received %v", 0., monetaryBalanceValue)
		}
		monetaryBalanceValue, err = cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[1]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *monetary balance current value: %v", err)
		}
		if monetaryBalanceValue != -3. {
			t.Errorf("unexpected balance value: expected %v, received %v", -3, monetaryBalanceValue)
		}
	})

	t.Run("CheckFinalBalance", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 2 balances of type *monetary , received %v", acnt)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 0 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != utils.MetaDefault || monetaryBalance.Value != -3 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	time.Sleep(1 * time.Second)

	t.Run("TestReadExportedCDRs", func(t *testing.T) {
		var files []string
		err := filepath.Walk("/tmp/exportedCDRs", func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, utils.CSVSuffix) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Error(err)
		}
		if len(files) != 1 {
			t.Errorf("Expected %+v, received: %+v", 1, len(files))
		}
		eCntInsBal := fmt.Sprintf(eCntInsBal, "directdebit")
		if outContent1, err := os.ReadFile(files[0]); err != nil {
			t.Error(err)
		} else if len(eCntInsBal) != len(string(outContent1)) {
			t.Errorf("Expecting: \n<%+v>, \nreceived: \n<%+v>", len(eCntInsBal), len(string(outContent1)))
			t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCntInsBal, string(outContent1))
		}
	})
}
