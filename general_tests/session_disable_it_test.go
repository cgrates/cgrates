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
	"bytes"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionInitiateDisableAccount(t *testing.T) {
	cfgJson := `{
"general": {
	"log_level": 7,
	"reply_timeout": "50s",
},
"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080",
},
"data_db": {								
	"db_type": "redis",						
	"db_port": 6379, 						
	"db_name": "10", 			
},
"stor_db": {
	"db_password": "CGRateS.org",
},
"rals": {
	"enabled": true,
	"thresholds_conns": ["*internal"],
	"max_increments":3000000,
},
"cdrs": {
	"enabled": true,
	"chargers_conns":["*internal"],
	"attributes_conns":["*localhost"],
	"rals_conns":["*localhost"],
	"stats_conns":["*localhost"]
},
"attributes": {
	"enabled": true,
	"stats_conns": ["*localhost"],
	"resources_conns": ["*localhost"],
	"apiers_conns": ["*localhost"]
},
"chargers": {
	"enabled": true,
	"attributes_conns": ["*internal"],
},
"stats": {
	"enabled": true,
	"store_interval": "1s",
	"thresholds_conns": ["*localhost"],
},
"thresholds": {
	"enabled": true,
	"store_interval": "1s",
	"sessions_conns": ["*localhost"],
	"apiers_conns": ["*internal"]
},
"sessions": {
	"enabled": true,
	"attributes_conns": ["*internal"],
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*localhost"],
	"chargers_conns": ["*internal"],
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
},
"schedulers": {
	"enabled": true,
	"cdrs_conns": ["*internal"],
	"stats_conns": ["*localhost"],
},
}
`
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.DBCfg{StorDB: &engine.DBParams{Type: utils.StringPointer("*internal")}}
	case utils.MetaMySQL:
		dbCfg = engine.DBCfg{StorDB: &engine.DBParams{
			Type:     utils.StringPointer("*mysql"),
			Password: utils.StringPointer("CGRateS.org"),
		}}
	case utils.MetaMongo:
		dbCfg = engine.DBCfg{StorDB: &engine.DBParams{
			Type: utils.StringPointer("*mongo"),
			Name: utils.StringPointer("cgrates"),
			Port: utils.IntPointer(27017),
		}}
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatalf("unsupported dbtype %v", *utils.DBType)
	}
	ng := engine.TestEngine{
		ConfigJSON: cfgJson,
		DBCfg:      dbCfg,
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,AP1,DISABLE_TRIGGER,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
AP1,ACT_MON,*asap,10`,
			utils.ActionsCsv: `#ActionsId,Action,ExtraParameters,Filter,BalanceId,BalanceType,Categories,DestinationIds,RatingSubject,SharedGroup,ExpiryTime,TimingIds,Units,BalanceWeight,BalanceBlocker,BalanceDisabled,Weight
ACT_MON,*topup_reset,,,balance1,*monetary,,,,,,,10,,false,,
DISABLE_ACC,*disable_account,,,,,,,,,,,,,,,`,
			utils.ActionTriggersCsv: `#Tag[0],UniqueId[1],ThresholdType[2],ThresholdValue[3],Recurrent[4],MinSleep[5],ExpiryTime[6],ActivationTime[7],BalanceTag[8],BalanceType[9],BalanceCategories[10],BalanceDestinationIds[11],BalanceRatingSubject[12],BalanceSharedGroup[13],BalanceExpiryTime[14],BalanceTimingIds[15],BalanceWeight[16],BalanceBlocker[17],BalanceDisabled[18],ActionsId[19],Weight[20]
DISABLE_TRIGGER,,*max_event_connect,1,false,0,,,,*event_connect,,,,,,,,,,DISABLE_ACC,10`,
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
			utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP,`,
			utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP,DR_RP,*any,10`,
			utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_RP,DST_1002,RT1,*up,4,0,`,
			utils.DestinationsCsv: `#Id,Prefix
DST_1002,1002`,
			utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT1,0.2,0.1,1s,1s,0`,
			utils.AttributesCsv: `#Tenant,ID,Context,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_ACNT,*any,,,,*opts.*accountID,*variable,~*req.Account,false,10`,
			utils.StatsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,STATS_1001,*string:~*req.Account:1001,,,-1,,*sum#1,,true,,,THD_1001`,
			utils.ThresholdsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],Weight[8],ActionIDs[9],Async[10]
cgrates.org,THD_1001,*string:~*req.StatID:STATS_1001,,-1,5,0,false,,DISABLE_ACC,false`,
		},

		LogBuffer: bytes.NewBuffer(nil),
	}
	//defer t.Log(ng.LogBuffer)
	client, _ := ng.Run(t)
	t.Run("AuthorizeEvent", func(t *testing.T) {
		var accRepl engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &accRepl); err != nil {
			t.Error(err)
		}
		if accRepl.Disabled {
			t.Errorf("account should not be disabled")
		}
		var reply sessions.V1AuthorizeReply
		if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent, &sessions.V1AuthorizeArgs{GetAttributes: true,
			GetMaxUsage: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]any{"Account": "1001",
					"Destination": "1002",
					"OriginHost":  "127.0.0.1:8448",
					"RequestType": "*prepaid",
					"SetupTime":   "1747212851",
					"Source":      "KamailioAgent"}}}, &reply); err != nil {
			t.Error(err)
		}

		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &accRepl); err != nil {
			t.Error(err)
		} else if *reply.MaxUsage != (98 * time.Second) {
			t.Errorf("expected to get usage even though account got disabled")
		}
		if accRepl.Disabled {
			t.Errorf("expected account to be not be disabled")
		}

	})
	t.Run("DisableAccountFromThresholds", func(t *testing.T) {
		var replyStr string
		rmReq := v1.AttrRemoveAccountActionTriggers{Tenant: "cgrates.org", Account: "1001", GroupID: "DISABLE_TRIGGER"}
		if err := client.Call(context.Background(), utils.APIerSv1RemoveAccountActionTriggers, rmReq, &replyStr); err != nil {
			t.Error("Got error on APIerSv1.RemoveActionTiming: ", err.Error())
		} else if replyStr != utils.OK {
			t.Error("Unexpected answer received", replyStr)
		}

		for range 5 {
			argsEv := &engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs, utils.MetaAttributes},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     utils.GenUUID(),
					Event: map[string]any{
						utils.RunID:        "run_1",
						utils.CGRID:        utils.GenUUID(),
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.OriginID:     "origin1",
						utils.OriginHost:   "OriginHost1",
						utils.RequestType:  utils.MetaPrepaid,
						utils.Subject:      "1001",
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        10 * time.Second,
					},
				},
			}
			if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &replyStr); err != nil {
				t.Error("Unexpected error: ", err.Error())
			} else if replyStr != utils.OK {
				t.Error("Unexpected reply received: ", replyStr)
			}
		}
		var accRepl engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &accRepl); err != nil {
			t.Error(err)
		}
		if !accRepl.Disabled {
			t.Errorf("expected account to be disabled")
		}
		t.Run("AuthorizeEventWithAccountDisabled", func(t *testing.T) {
			var reply string
			if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent, &sessions.V1AuthorizeArgs{GetAttributes: true,
				GetMaxUsage: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     utils.GenUUID(),
					Time:   utils.TimePointer(time.Now()),
					Event: map[string]any{"Account": "1001",
						"Destination": "1002",
						"OriginHost":  "127.0.0.1:8448",
						"RequestType": "*prepaid",
						"SetupTime":   "1747212851",
						"Source":      "KamailioAgent"}}}, &reply); err == nil || err.Error() != utils.NewErrRALs(utils.ErrAccountDisabled).Error() {
				t.Errorf("expected: ACCOUNT_DISABLED error, got: %v", err)
			}
		})

	})
}
