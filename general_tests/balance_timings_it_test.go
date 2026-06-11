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
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestBalanceTimings(t *testing.T) {
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
	"reply_timeout": "50s"
},

"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080"
},

"rals": {
	"enabled": true,
},

"schedulers": {
	"enabled": true,
	"cdrs_conns": ["*internal"],
},

"cdrs": {
	"enabled": true,
	"chargers_conns":["*internal"],
	"rals_conns": ["*localhost"],
},

"attributes": {
	"enabled": true,
	"apiers_conns": ["*localhost"]
},

"chargers": {
	"enabled": true,
	"attributes_conns": ["*internal"]
},

"sessions": {
	"enabled": true,
	"attributes_conns": ["*internal"],
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
	"chargers_conns": ["*internal"]
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
},

}
`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,AP_PACKAGE_10,,,
cgrates.org,1002,AP_PACKAGE_10,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
AP_PACKAGE_10,ACT_TOPUP_RST_10,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP_RST_10,*topup_reset,,,bal1,*monetary,,*any,,,*unlimited,HALF1,10,10,false,false,10
ACT_TOPUP_RST_10,*topup_reset,,,bal2,*monetary,,*any,,,*unlimited,HALF2,10,10,false,false,99`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0
cgrates.org,Raw,,,*raw,*constant:*req.RequestType:*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_1002_20CNT,DST_1002,RT_20CNT,*up,4,0,`,
		utils.DestinationsCsv: `#Id,Prefix
DST_1002,1002
DST_1001,1001`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_20CNT,0.4,0.2,60s,60s,0s
RT_20CNT,0,0.1,60s,1s,60s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_1001,DR_1002_20CNT,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_1001,`,
		utils.TimingsCsv: `#Tag,Years,Months,MonthDays,WeekDays,Time
HALF1,*any,*any,*any,*any,00:00:00;11:59:59
HALF2,*any,*any,*any,*any,12:00:00;23:59:59`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(50 * time.Millisecond)

	t.Run("TimingIsActiveAt", func(t *testing.T) {
		var reply *bool
		params := &v1.TimeParams{
			TimingID: "HALF1",
			Time:     "2024-09-17T10:00:00Z",
		}
		if err := client.Call(context.Background(), utils.APIerSV1TimingIsActiveAt, params, &reply); err != nil {
			t.Fatal(err)
		} else if !*reply {
			t.Errorf("expected TimingID to be Active")
		}
		params = &v1.TimeParams{
			TimingID: "HALF2",
			Time:     "2024-09-17T10:00:00Z",
		}
		if err := client.Call(context.Background(), utils.APIerSV1TimingIsActiveAt, params, &reply); err != nil {
			t.Fatal(err)
		} else if *reply {
			t.Errorf("expected TimingID to be inactive")
		}
		params = &v1.TimeParams{
			TimingID: "HALF2",
			Time:     "2024-09-17T12:00:00Z",
		}
		if err := client.Call(context.Background(), utils.APIerSV1TimingIsActiveAt, params, &reply); err != nil {
			t.Fatal(err)
		} else if !*reply {
			t.Errorf("expected TimingID to be Active")
		}
	})

	t.Run("GetAccount", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 1 balance of type *monetary, received %v", utils.ToJSON(acnt))
		}

		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "bal1" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal1" && monetaryBalance.Timings[0].ID != "HALF1" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF1", monetaryBalance.Timings)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "bal2" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal2" && monetaryBalance.Timings[0].ID != "HALF2" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF2", monetaryBalance.Timings)
		}
	})

	t.Run("Half2ProcessExternalCDR", func(t *testing.T) {
		var reply string
		args := &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "TestBalanceTimings",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPrepaid,
				AnswerTime:  "2024-08-04T15:00:07Z",
				SetupTime:   "2024-08-04T15:00:00Z",
				Tenant:      "cgrates.org",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "1",
			},
		}
		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, args, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Received: %s", reply)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Half2GetAccount", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 1 balance of type *monetary, received %v", utils.ToJSON(acnt))
		}

		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "bal1" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal1" && monetaryBalance.Timings[0].ID != "HALF1" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF1", monetaryBalance.Timings)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "bal2" || monetaryBalance.Value != 9.4 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal2" && monetaryBalance.Timings[0].ID != "HALF2" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF2", monetaryBalance.Timings)
		}
	})

	t.Run("Half1ProcessExternalCDR", func(t *testing.T) {
		var reply string
		args := &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "TestBalanceTimings2",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPrepaid,
				AnswerTime:  "2024-08-04T11:00:07Z",
				SetupTime:   "2024-08-04T11:00:00Z",
				Tenant:      "cgrates.org",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "1",
			},
		}
		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, args, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Received: %s", reply)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Half1GetAccount", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 1 balance of type *monetary, received %v", utils.ToJSON(acnt))
		}

		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "bal1" || monetaryBalance.Value != 9.4 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal1" && monetaryBalance.Timings[0].ID != "HALF1" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF1", monetaryBalance.Timings)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "bal2" || monetaryBalance.Value != 9.4 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal2" && monetaryBalance.Timings[0].ID != "HALF2" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF2", monetaryBalance.Timings)
		}
	})

}

func TestBalanceTimingsSetActions(t *testing.T) {
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
	"reply_timeout": "50s"
},

"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080"
},

"rals": {
	"enabled": true,
},

"schedulers": {
	"enabled": true,
	"cdrs_conns": ["*internal"],
},

"cdrs": {
	"enabled": true,
	"chargers_conns":["*internal"],
	"rals_conns": ["*localhost"],
},

"attributes": {
	"enabled": true,
	"apiers_conns": ["*localhost"]
},

"chargers": {
	"enabled": true,
	"attributes_conns": ["*internal"]
},

"sessions": {
	"enabled": true,
	"attributes_conns": ["*internal"],
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
	"chargers_conns": ["*internal"]
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
},

}
`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,AP_PACKAGE_10,,,
cgrates.org,1002,AP_PACKAGE_10,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
AP_PACKAGE_10,ACT_TOPUP_RST_10,*asap,10`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0
cgrates.org,Raw,,,*raw,*constant:*req.RequestType:*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_1002_20CNT,DST_1002,RT_20CNT,*up,4,0,`,
		utils.DestinationsCsv: `#Id,Prefix
DST_1002,1002
DST_1001,1001`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_20CNT,0.4,0.2,60s,60s,0s
RT_20CNT,0,0.1,60s,1s,60s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_1001,DR_1002_20CNT,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_1001,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)
	time.Sleep(50 * time.Millisecond)

	t.Run("SetTimings", func(t *testing.T) {
		timing := &utils.TPTimingWithAPIOpts{
			TPTiming: &utils.TPTiming{
				ID:        "HALF1",
				StartTime: "00:00:00",
				EndTime:   "11:59:59",
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetTiming, timing, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned", reply)
		}
		timing2 := &utils.TPTimingWithAPIOpts{
			TPTiming: &utils.TPTiming{
				ID:        "HALF2",
				StartTime: "12:00:00",
				EndTime:   "23:59:59",
			},
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetTiming, timing2, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned", reply)
		}
	})

	t.Run("Sv1SetActions", func(t *testing.T) {
		attrs1 := &v1.V1AttrSetActions{
			ActionsId: "ACT_TOPUP_RST_10",
			Actions: []*v1.V1TPAction{
				{
					Identifier:  utils.MetaTopUpReset,
					BalanceId:   "bal1",
					TimingTags:  "HALF1",
					BalanceType: utils.MetaMonetary,
					Units:       10.0,
					ExpiryTime:  utils.MetaUnlimited,
					Weight:      10.0,
				},
				{
					Identifier:  utils.MetaTopUpReset,
					BalanceId:   "bal2",
					TimingTags:  "HALF2",
					BalanceType: utils.MetaMonetary,
					Units:       10.0,
					ExpiryTime:  utils.MetaUnlimited,
					Weight:      99.0,
				},
			},
			Overwrite: true,
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetActions, &attrs1, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Unexpected reply returned: %s", reply)
		}
		// LoadTPFromFolder
		engine.LoadCSVs(t, client, "", tpFiles)
		attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1001", ActionsId: "ACT_TOPUP_RST_10"}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
			t.Errorf("APIerSv1ExecuteAction failed unexpectedly: %v", err)
		}
	})

	t.Run("GetAccount", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 1 balance of type *monetary, received %v", utils.ToJSON(acnt))
		}

		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "bal1" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal1" && monetaryBalance.Timings[0].ID != "HALF1" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF1", monetaryBalance.Timings)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "bal2" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal2" && monetaryBalance.Timings[0].ID != "HALF2" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF2", monetaryBalance.Timings)
		}
	})

	t.Run("Half2ProcessExternalCDR", func(t *testing.T) {
		var reply string
		args := &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "TestBalanceTimings",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPrepaid,
				AnswerTime:  "2024-08-04T15:00:07Z",
				SetupTime:   "2024-08-04T15:00:00Z",
				Tenant:      "cgrates.org",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "1",
			},
		}
		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, args, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Received: %s", reply)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Half2GetAccount", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 1 balance of type *monetary, received %v", utils.ToJSON(acnt))
		}

		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "bal1" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal1" && monetaryBalance.Timings[0].ID != "HALF1" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF1", monetaryBalance.Timings)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "bal2" || monetaryBalance.Value != 9.4 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal2" && monetaryBalance.Timings[0].ID != "HALF2" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF2", monetaryBalance.Timings)
		}
	})

	t.Run("Half1ProcessExternalCDR", func(t *testing.T) {
		var reply string
		args := &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "TestBalanceTimings2",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPrepaid,
				AnswerTime:  "2024-08-04T11:00:07Z",
				SetupTime:   "2024-08-04T11:00:00Z",
				Tenant:      "cgrates.org",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "1",
			},
		}
		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, args, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Received: %s", reply)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Half1GetAccount", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 1 balance of type *monetary, received %v", utils.ToJSON(acnt))
		}

		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "bal1" || monetaryBalance.Value != 9.4 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal1" && monetaryBalance.Timings[0].ID != "HALF1" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF1", monetaryBalance.Timings)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "bal2" || monetaryBalance.Value != 9.4 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal2" && monetaryBalance.Timings[0].ID != "HALF2" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF2", monetaryBalance.Timings)
		}
	})

}

func TestBalanceTimingsV2SetActions(t *testing.T) {
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
	"reply_timeout": "50s"
},

"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080"
},

"rals": {
	"enabled": true,
},

"schedulers": {
	"enabled": true,
	"cdrs_conns": ["*internal"],
},

"cdrs": {
	"enabled": true,
	"chargers_conns":["*internal"],
	"rals_conns": ["*localhost"],
},

"attributes": {
	"enabled": true,
	"apiers_conns": ["*localhost"]
},

"chargers": {
	"enabled": true,
	"attributes_conns": ["*internal"]
},

"sessions": {
	"enabled": true,
	"attributes_conns": ["*internal"],
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
	"chargers_conns": ["*internal"]
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
},

}
`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,AP_PACKAGE_10,,,
cgrates.org,1002,AP_PACKAGE_10,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
AP_PACKAGE_10,ACT_TOPUP_RST_10,*asap,10`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0
cgrates.org,Raw,,,*raw,*constant:*req.RequestType:*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_1002_20CNT,DST_1002,RT_20CNT,*up,4,0,`,
		utils.DestinationsCsv: `#Id,Prefix
DST_1002,1002
DST_1001,1001`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_20CNT,0.4,0.2,60s,60s,0s
RT_20CNT,0,0.1,60s,1s,60s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_1001,DR_1002_20CNT,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_1001,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)
	time.Sleep(50 * time.Millisecond)

	t.Run("SetTimings", func(t *testing.T) {
		timing := &utils.TPTimingWithAPIOpts{
			TPTiming: &utils.TPTiming{
				ID:        "HALF1",
				StartTime: "00:00:00",
				EndTime:   "11:59:59",
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetTiming, timing, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned", reply)
		}
		timing2 := &utils.TPTimingWithAPIOpts{
			TPTiming: &utils.TPTiming{
				ID:        "HALF2",
				StartTime: "12:00:00",
				EndTime:   "23:59:59",
			},
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetTiming, timing2, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned", reply)
		}
	})

	t.Run("Sv2SetActions", func(t *testing.T) {
		attrs1 := &utils.AttrSetActions{
			ActionsId: "ACT_TOPUP_RST_10",
			Actions: []*utils.TPAction{
				{
					Identifier:     utils.MetaTopUpReset,
					BalanceId:      "bal1",
					DestinationIds: utils.MetaAny,
					TimingTags:     "HALF1",
					BalanceType:    utils.MetaMonetary,
					Units:          "10",
					ExpiryTime:     utils.MetaUnlimited,
					Weight:         10.0,
				}, {
					Identifier:     utils.MetaTopUpReset,
					BalanceId:      "bal2",
					DestinationIds: utils.MetaAny,
					TimingTags:     "HALF2",
					BalanceType:    utils.MetaMonetary,
					Units:          "10",
					ExpiryTime:     utils.MetaUnlimited,
					Weight:         99.0,
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv2SetActions, &attrs1, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Unexpected reply returned: %s", reply)
		}
		// LoadTPFromFolder
		engine.LoadCSVs(t, client, "", tpFiles)
		attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1001", ActionsId: "ACT_TOPUP_RST_10"}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
			t.Errorf("APIerSv1ExecuteAction failed unexpectedly: %v", err)
		}
	})

	t.Run("GetAccount", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 1 balance of type *monetary, received %v", utils.ToJSON(acnt))
		}

		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "bal1" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal1" && monetaryBalance.Timings[0].ID != "HALF1" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF1", monetaryBalance.Timings)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "bal2" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal2" && monetaryBalance.Timings[0].ID != "HALF2" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF2", monetaryBalance.Timings)
		}
	})

	t.Run("Half2ProcessExternalCDR", func(t *testing.T) {
		var reply string
		args := &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "TestBalanceTimings",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPrepaid,
				AnswerTime:  "2024-08-04T15:00:07Z",
				SetupTime:   "2024-08-04T15:00:00Z",
				Tenant:      "cgrates.org",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "1",
			},
		}
		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, args, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Received: %s", reply)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Half2GetAccount", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 1 balance of type *monetary, received %v", utils.ToJSON(acnt))
		}

		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "bal1" || monetaryBalance.Value != 10 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal1" && monetaryBalance.Timings[0].ID != "HALF1" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF1", monetaryBalance.Timings)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "bal2" || monetaryBalance.Value != 9.4 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal2" && monetaryBalance.Timings[0].ID != "HALF2" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF2", monetaryBalance.Timings)
		}
	})

	t.Run("Half1ProcessExternalCDR", func(t *testing.T) {
		var reply string
		args := &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "TestBalanceTimings2",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPrepaid,
				AnswerTime:  "2024-08-04T11:00:07Z",
				SetupTime:   "2024-08-04T11:00:00Z",
				Tenant:      "cgrates.org",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "1",
			},
		}
		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, args, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Received: %s", reply)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Half1GetAccount", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Fatalf("expected account to have 1 balance of type *monetary, received %v", utils.ToJSON(acnt))
		}

		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][1]
		if monetaryBalance.ID != "bal1" || monetaryBalance.Value != 9.4 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal1" && monetaryBalance.Timings[0].ID != "HALF1" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF1", monetaryBalance.Timings)
		}
		monetaryBalance = acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "bal2" || monetaryBalance.Value != 9.4 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		} else if monetaryBalance.ID == "bal2" && monetaryBalance.Timings[0].ID != "HALF2" {
			t.Fatalf("expected TimingIDs %v, received: %v", "HALF2", monetaryBalance.Timings)
		}
	})

}

func TestBalanceTimingsClearTimings(t *testing.T) {
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
            "reply_timeout": "50s"
        },
        
        "listen": {
            "rpc_json": ":2012",
            "rpc_gob": ":2013",
            "http": ":2080"
        },
        
        "rals": {
            "enabled": true,
        },
        
        "schedulers": {
            "enabled": true,
            "cdrs_conns": ["*internal"],
        },
        
        "cdrs": {
            "enabled": true,
            "chargers_conns":["*internal"],
            "rals_conns": ["*localhost"],
        },
        
        "attributes": {
            "enabled": true,
            "apiers_conns": ["*localhost"]
        },
        
        "chargers": {
            "enabled": true,
            "attributes_conns": ["*internal"]
        },
        
        "sessions": {
            "enabled": true,
            "attributes_conns": ["*internal"],
            "rals_conns": ["*internal"],
            "cdrs_conns": ["*internal"],
            "chargers_conns": ["*internal"]
        },
        
        "apiers": {
            "enabled": true,
            "scheduler_conns": ["*internal"]
        },
        
        }
        `

	ng := engine.TestEngine{ConfigJSON: content}
	client, _ := ng.Run(t)
	time.Sleep(50 * time.Millisecond)

	var reply string

	t.Run("SetTiming", func(t *testing.T) {
		timing := &utils.TPTimingWithAPIOpts{
			TPTiming: &utils.TPTiming{
				ID:        "HALF1",
				StartTime: "00:00:00",
				EndTime:   "11:59:59",
			},
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetTiming, timing, &reply); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("SetBalanceWithTiming", func(t *testing.T) {
		args := &utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "testClearTimings",
			BalanceType: utils.MetaMonetary,
			Balance: map[string]any{
				utils.ID:        "balClear",
				utils.TimingIDs: "HALF1",
			},
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetBalance, args, &reply); err != nil {
			t.Fatal(err)
		}
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testClearTimings"}, &acnt); err != nil {
			t.Fatal(err)
		}
		for _, bal := range acnt.BalanceMap[utils.MetaMonetary] {
			if bal.ID == "balClear" && len(bal.Timings) == 0 {
				t.Fatal("expected Timings to be populated after SetBalance with TimingIDs")
			}
		}
	})

	t.Run("ClearTimings", func(t *testing.T) {
		args := &utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "testClearTimings",
			BalanceType: utils.MetaMonetary,
			Balance: map[string]any{
				utils.ID:        "balClear",
				utils.TimingIDs: "",
			},
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetBalance, args, &reply); err != nil {
			t.Fatal(err)
		}
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testClearTimings"}, &acnt); err != nil {
			t.Fatal(err)
		}
		for _, bal := range acnt.BalanceMap[utils.MetaMonetary] {
			if bal.ID == "balClear" {
				if len(bal.Timings) != 0 {
					t.Errorf("expected Timings to be empty after clearing TimingIDs, got: %v", bal.Timings)
				}
				if len(bal.TimingIDs) != 0 {
					t.Errorf("expected TimingIDs to be empty, got: %v", bal.TimingIDs)
				}
			}
		}
	})

	t.Run("BalanceActiveAfterClear", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testClearTimings"}, &acnt); err != nil {
			t.Fatal(err)
		}
		for _, bal := range acnt.BalanceMap[utils.MetaMonetary] {
			if bal.ID == "balClear" {
				pmTime := time.Date(2026, 5, 22, 14, 0, 0, 0, time.UTC)
				if !bal.IsActiveAt(pmTime) {
					t.Error("expected balance to be active at PM after clearing Timings[]")
				}
			}
		}
	})

	t.Run("SetTimingAfterClear", func(t *testing.T) {
		args := &utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "testClearTimings",
			BalanceType: utils.MetaMonetary,
			Balance: map[string]any{
				utils.ID:        "balClear",
				utils.TimingIDs: "HALF1",
			},
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetBalance, args, &reply); err != nil {
			t.Fatal(err)
		}
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testClearTimings"}, &acnt); err != nil {
			t.Fatal(err)
		}
		for _, bal := range acnt.BalanceMap[utils.MetaMonetary] {
			if bal.ID == "balClear" {
				if len(bal.Timings) != 1 {
					t.Errorf("expected 1 timing after re-assign, got %d", len(bal.Timings))
				}
			}
		}
	})
}

func TestBalanceTimingsNegation(t *testing.T) {
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
    "reply_timeout": "50s"
},
 
"listen": {
    "rpc_json": ":2012",
    "rpc_gob": ":2013",
    "http": ":2080"
},
 
"rals": {
    "enabled": true,
},
 
"schedulers": {
    "enabled": true,
    "cdrs_conns": ["*internal"],
},

"attributes" : {
    "enabled": true,
},
 
"cdrs": {
    "enabled": true,
    "chargers_conns":["*internal"],
    "rals_conns": ["*localhost"],
},
 
"chargers": {
    "enabled": true,
    "attributes_conns": ["*internal"]
},
 
"sessions": {
    "enabled": true,
    "rals_conns": ["*internal"],
    "cdrs_conns": ["*internal"],
    "chargers_conns": ["*internal"]
},
 
"apiers": {
    "enabled": true,
    "scheduler_conns": ["*internal"]
},
 
}
`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,apPackage10,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
apPackage10,actTopup,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
actTopup,*topup_reset,,,balNeg,*monetary,,*any,,,*unlimited,,100,10,false,false,10`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0
cgrates.org,Raw,,,*raw,*constant:*req.RequestType:*none,0`,
		utils.DestinationsCsv: `#Id,Prefix
dst1002,1002`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
dr1002,dst1002,rt1Cnt,*up,4,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
rt1Cnt,0,1,60s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
rp1001,dr1002,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,rp1001,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(50 * time.Millisecond)

	t.Run("SetTimings", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetTiming, &utils.TPTimingWithAPIOpts{
			TPTiming: &utils.TPTiming{
				ID:        "half1",
				StartTime: "00:00:00",
				EndTime:   "11:59:59",
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetTiming, &utils.TPTimingWithAPIOpts{
			TPTiming: &utils.TPTiming{
				ID:        "half2",
				StartTime: "12:00:00",
				EndTime:   "23:59:59",
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("SetBalanceWithNegatedTiming", func(t *testing.T) {
		var reply string
		attrs := &utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1001",
			BalanceType: utils.MetaMonetary,
			Balance: map[string]any{
				utils.ID:        "balNeg",
				utils.TimingIDs: "half1;!half2",
				utils.Value:     100.0,
				utils.Weight:    10.0,
			},
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetBalance, attrs, &reply); err != nil {
			t.Fatal(err)
		} else if reply != utils.OK {
			t.Fatalf("unexpected reply: %s", reply)
		}
	})

	t.Run("VerifyTimingIDsMap", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acnt); err != nil {
			t.Fatal(err)
		}
		balances := acnt.BalanceMap[utils.MetaMonetary]
		var balNeg *engine.Balance
		for _, b := range balances {
			if b.ID == "balNeg" {
				balNeg = b
				break
			}
		}
		if balNeg == nil {
			t.Fatal("balNeg not found in account")
		}
		if balNeg.TimingIDs["half1"] != true {
			t.Errorf("expected half1:true in TimingIDs, got %v", balNeg.TimingIDs)
		}
		if balNeg.TimingIDs["half2"] != false {
			t.Errorf("expected half2:false in TimingIDs, got %v", balNeg.TimingIDs)
		}
	})

	t.Run("PMCDRShouldNotDebitBalNeg", func(t *testing.T) {
		var acntBefore engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntBefore); err != nil {
			t.Fatal(err)
		}
		var balBefore float64
		for _, b := range acntBefore.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balNeg" {
				balBefore = b.Value
			}
		}

		var reply string
		args := &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "testPmNegation",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPostpaid,
				SetupTime:   "2024-08-04T15:00:00Z", // PM
				AnswerTime:  "2024-08-04T15:00:00Z",
				Tenant:      "cgrates.org",
				Category:    "call",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "10s",
			},
		}
		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, args, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)

		var acntAfter engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntAfter); err != nil {
			t.Fatal(err)
		}
		var balAfter float64
		for _, b := range acntAfter.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balNeg" {
				balAfter = b.Value
			}
		}

		if balAfter != balBefore {
			t.Errorf("balNeg was debited during PM despite !half2 — before: %v after: %v", balBefore, balAfter)
		}
	})

	t.Run("AMCDRShouldDebitBalNeg", func(t *testing.T) {
		var acntBefore engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntBefore); err != nil {
			t.Fatal(err)
		}
		var balBefore float64
		for _, b := range acntBefore.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balNeg" {
				balBefore = b.Value
			}
		}

		var reply string
		args := &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "testAmNegation",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPostpaid,
				SetupTime:   "2024-08-04T10:00:00Z", // AM
				AnswerTime:  "2024-08-04T10:00:00Z",
				Tenant:      "cgrates.org",
				Category:    "call",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "10s",
			},
		}
		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, args, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)

		var acntAfter engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntAfter); err != nil {
			t.Fatal(err)
		}
		var balAfter float64
		for _, b := range acntAfter.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balNeg" {
				balAfter = b.Value
			}
		}

		if balAfter >= balBefore {
			t.Errorf("balNeg was NOT debited during AM — before: %v after: %v", balBefore, balAfter)
		}
	})
	t.Run("OnlyNegatedTimingShouldNotDebitPM", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetBalance, &utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1001",
			BalanceType: utils.MetaMonetary,
			Balance: map[string]any{
				utils.ID:        "balOnlyNeg",
				utils.TimingIDs: "!half2",
				utils.Value:     100.0,
				utils.Weight:    20,
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}

		var acntBefore engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntBefore); err != nil {
			t.Fatal(err)
		}
		var balBefore float64
		for _, b := range acntBefore.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balOnlyNeg" {
				balBefore = b.Value
			}
		}

		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "testOnlyNegPM",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPostpaid,
				SetupTime:   "2024-08-04T15:00:00Z",
				AnswerTime:  "2024-08-04T15:00:00Z",
				Tenant:      "cgrates.org",
				Category:    "call",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "10s",
			},
		}, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)

		var acntAfter engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntAfter); err != nil {
			t.Fatal(err)
		}
		var balAfter float64
		for _, b := range acntAfter.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balOnlyNeg" {
				balAfter = b.Value
			}
		}
		if balAfter != balBefore {
			t.Errorf("balOnlyNeg was debited during PM , before: %v after: %v", balBefore, balAfter)
		}
	})
	t.Run("OnlyNegatedTimingShouldDebitAM", func(t *testing.T) {
		var acntBefore engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntBefore); err != nil {
			t.Fatal(err)
		}
		var balBefore float64
		for _, b := range acntBefore.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balOnlyNeg" {
				balBefore = b.Value
			}
		}

		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "testOnlyNegAM",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPostpaid,
				SetupTime:   "2024-08-04T10:00:00Z",
				AnswerTime:  "2024-08-04T10:00:00Z",
				Tenant:      "cgrates.org",
				Category:    "call",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "10s",
			},
		}, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)

		var acntAfter engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntAfter); err != nil {
			t.Fatal(err)
		}
		var balAfter float64
		for _, b := range acntAfter.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balOnlyNeg" {
				balAfter = b.Value
			}
		}
		if balAfter >= balBefore {
			t.Errorf("balOnlyNeg was NOT debited during AM, before: %v after: %v", balBefore, balAfter)
		}
	})
	t.Run("BothTimingsNegated", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetAccount, &utils.AttrSetAccount{
			Tenant:  "cgrates.org",
			Account: "1002",
		}, &reply); err != nil {
			t.Fatal(err)
		}

		if err := client.Call(context.Background(), utils.APIerSv1SetBalance, &utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1002",
			BalanceType: utils.MetaMonetary,
			Balance: map[string]any{
				utils.ID:        "balBothNeg",
				utils.TimingIDs: "!half1;!half2",
				utils.Value:     100.0,
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}

		var acntBefore engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"}, &acntBefore); err != nil {
			t.Fatal(err)
		}
		var balBefore float64
		for _, b := range acntBefore.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balBothNeg" {
				balBefore = b.Value
			}
		}

		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "testBothNegAM",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPostpaid,
				SetupTime:   "2024-08-04T10:00:00Z",
				AnswerTime:  "2024-08-04T10:00:00Z",
				Tenant:      "cgrates.org",
				Category:    "call",
				Account:     "1002",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "10s",
			},
		}, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)

		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "testBothNegPM",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPostpaid,
				SetupTime:   "2024-08-04T15:00:00Z",
				AnswerTime:  "2024-08-04T15:00:00Z",
				Tenant:      "cgrates.org",
				Category:    "call",
				Account:     "1002",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "10s",
			},
		}, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)

		var acntAfter engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"}, &acntAfter); err != nil {
			t.Fatal(err)
		}
		var balAfter float64
		for _, b := range acntAfter.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balBothNeg" {
				balAfter = b.Value
			}
		}
		if balAfter != balBefore {
			t.Errorf("balBothNeg was debited despite both timings negated, before: %v after: %v", balBefore, balAfter)
		}
	})
	t.Run("VerifyNegatedTimingStored", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acnt); err != nil {
			t.Fatal(err)
		}
		var balNeg *engine.Balance
		for _, b := range acnt.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balNeg" {
				balNeg = b
				break
			}
		}
		if balNeg == nil {
			t.Fatal("balNeg not found")
		}
		found := false
		for _, tim := range balNeg.Timings {
			if strings.HasPrefix(tim.ID, "!") {
				t.Errorf("Timings entry has unexpected '!' prefix: %s", tim.ID)
			}
			if tim.ID == "half2" {
				found = true
			}
		}
		if !found {
			t.Error("expected half2 in Timings, not found")
		}
		if v, ok := balNeg.TimingIDs["half2"]; !ok || v != false {
			t.Errorf("expected TimingIDs[half2]=false, got %v (exists: %v)", v, ok)
		}
	})

	t.Run("SetActionsWithNegatedTimingTag", func(t *testing.T) {
		attrs := &v1.V1AttrSetActions{
			ActionsId: "actNeagtedTiming",
			Actions: []*v1.V1TPAction{
				{
					Identifier:  utils.MetaTopUpReset,
					BalanceId:   "balNegAction",
					TimingTags:  "!half2",
					BalanceType: utils.MetaMonetary,
					Units:       10.0,
					ExpiryTime:  utils.MetaUnlimited,
					Weight:      10.0,
				},
			},
			Overwrite: true,
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetActions, attrs, &reply); err != nil {
			t.Fatalf("SetActions with negated TimingTag failed: %v", err)
		} else if reply != utils.OK {
			t.Errorf("unexpected reply: %s", reply)
		}
	})
	t.Run("SetActionsV2WithNegatedTimingTag", func(t *testing.T) {
		attrs := &utils.AttrSetActions{
			ActionsId: "actNeagtedTimingV2",
			Overwrite: true,
			Actions: []*utils.TPAction{
				{
					Identifier:  utils.MetaTopUpReset,
					BalanceId:   "balNegActionV2",
					TimingTags:  "!half2",
					BalanceType: utils.MetaMonetary,
					Units:       "10",
					ExpiryTime:  utils.MetaUnlimited,
					Weight:      10.0,
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv2SetActions, attrs, &reply); err != nil {
			t.Fatalf("APIerSv2SetActions with negated TimingTag failed: %v", err)
		} else if reply != utils.OK {
			t.Errorf("unexpected reply: %s", reply)
		}
	})
}

func TestBalanceTimingsOverlapNegation(t *testing.T) {
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
    "reply_timeout": "50s"
},

"listen": {
    "rpc_json": ":2012",
    "rpc_gob": ":2013",
    "http": ":2080"
},

"rals": {
    "enabled": true,
},

"schedulers": {
    "enabled": true,
    "cdrs_conns": ["*internal"],
},

"attributes" : {
    "enabled": true,
},

"cdrs": {
    "enabled": true,
    "chargers_conns":["*internal"],
    "rals_conns": ["*localhost"],
},

"chargers": {
    "enabled": true,
    "attributes_conns": ["*internal"]
},

"sessions": {
    "enabled": true,
    "rals_conns": ["*internal"],
    "cdrs_conns": ["*internal"],
    "chargers_conns": ["*internal"]
},

"apiers": {
    "enabled": true,
    "scheduler_conns": ["*internal"]
},

}
`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,apPackage10,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
apPackage10,actTopup,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
actTopup,*topup_reset,,,balWeekdays,*monetary,,*any,,,*unlimited,,100,10,false,false,10`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0
cgrates.org,Raw,,,*raw,*constant:*req.RequestType:*none,0`,
		utils.DestinationsCsv: `#Id,Prefix
dst1002,1002`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
dr1002,dst1002,rt1Cnt,*up,4,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
rt1Cnt,0,1,60s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
rp1001,dr1002,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,rp1001,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(50 * time.Millisecond)

	t.Run("SetTimings", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetTiming, &utils.TPTimingWithAPIOpts{
			TPTiming: &utils.TPTiming{
				ID:       "alldays",
				WeekDays: utils.WeekDays{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday},

				StartTime: "00:00:00",
				EndTime:   "23:59:59",
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetTiming, &utils.TPTimingWithAPIOpts{
			TPTiming: &utils.TPTiming{
				ID:        "weekend",
				WeekDays:  utils.WeekDays{time.Saturday, time.Sunday},
				StartTime: "00:00:00",
				EndTime:   "23:59:59",
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("SetBalanceAlldaysNotWeekend", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetBalance, &utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1001",
			BalanceType: utils.MetaMonetary,
			Balance: map[string]any{
				utils.ID:        "balWeekdays",
				utils.TimingIDs: "alldays;!weekend",
				utils.Value:     100.0,
				utils.Weight:    10.0,
			},
		}, &reply); err != nil {
			t.Fatal(err)
		} else if reply != utils.OK {
			t.Fatalf("unexpected reply: %s", reply)
		}
	})

	t.Run("WeekendCDRShouldNotDebitBalWeekdays", func(t *testing.T) {
		var acntBefore engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntBefore); err != nil {
			t.Fatal(err)
		}
		var balBefore float64
		for _, b := range acntBefore.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balWeekdays" {
				balBefore = b.Value
			}
		}

		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "testWeekendNoDebit",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPostpaid,
				SetupTime:   "2024-08-04T10:00:00Z", // sunday
				AnswerTime:  "2024-08-04T10:00:00Z",
				Tenant:      "cgrates.org",
				Category:    "call",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "10s",
			},
		}, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)

		var acntAfter engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntAfter); err != nil {
			t.Fatal(err)
		}
		var balAfter float64
		for _, b := range acntAfter.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balWeekdays" {
				balAfter = b.Value
			}
		}
		if balAfter != balBefore {
			t.Errorf("balWeekdays debited during weekend, before: %v after: %v", balBefore, balAfter)
		}
	})

	t.Run("WeekdayCDRShouldDebitBalWeekdays", func(t *testing.T) {
		var acntBefore engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntBefore); err != nil {
			t.Fatal(err)
		}
		var balBefore float64
		for _, b := range acntBefore.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balWeekdays" {
				balBefore = b.Value
			}
		}

		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessExternalCDR, &engine.ExternalCDRWithAPIOpts{
			ExternalCDR: &engine.ExternalCDR{
				OriginID:    "testWeekdayDebit",
				ToR:         utils.MetaVoice,
				RequestType: utils.MetaPostpaid,
				SetupTime:   "2024-08-05T10:00:00Z", // monday
				AnswerTime:  "2024-08-05T10:00:00Z",
				Tenant:      "cgrates.org",
				Category:    "call",
				Account:     "1001",
				Subject:     "1001",
				Destination: "1002",
				Usage:       "10s",
			},
		}, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)

		var acntAfter engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acntAfter); err != nil {
			t.Fatal(err)
		}
		var balAfter float64
		for _, b := range acntAfter.BalanceMap[utils.MetaMonetary] {
			if b.ID == "balWeekdays" {
				balAfter = b.Value
			}
		}
		if balAfter >= balBefore {
			t.Errorf("balWeekdays not debited during week, before: %v after: %v", balBefore, balAfter)
		}
	})
}
