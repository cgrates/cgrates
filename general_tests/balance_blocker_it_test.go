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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestBlockerTrueAuthorize(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_voice,*voice,,*any,,,*unlimited,,10s,10,true,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	testNg := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := testNg.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("Authorize", func(t *testing.T) {
		args := &sessions.V1AuthorizeArgs{
			GetMaxUsage: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockTrueAuth",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestBalBlockTrue",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        "30s",
				},
			},
		}
		var rplyFirst sessions.V1AuthorizeReply
		if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rplyFirst); err != nil {
			t.Fatal(err)
		}
		if *rplyFirst.MaxUsage != 10*time.Second {
			t.Errorf("Expected usage <%+v>, Received <%+v>", 10*time.Second, rplyFirst.MaxUsage)
		}
	})
}

func TestBlockerTrueInterimUpdate(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_voice,*voice,,*any,,,*unlimited,,10s,10,true,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("InterimUpdate30s", func(t *testing.T) {
		updateArgs := &sessions.V1UpdateSessionArgs{
			UpdateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockTrueUpdate",
				Event: map[string]any{
					utils.OriginID:     "TestBalBlockTrue",
					utils.ToR:          utils.MetaVoice,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.RequestType:  utils.MetaPrepaid,
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        "1s",
				},
			},
		}
		var acnt *engine.Account
		attrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		expAccVal := 10000000000.0
		expUsage := 1 * time.Second
		for i := 1; i != 30; i++ {
			var updateRpl *sessions.V1UpdateSessionReply
			if err := client.Call(context.Background(), utils.SessionSv1UpdateSession,
				updateArgs, &updateRpl); err != nil {
				t.Error(err)
			}
			if *updateRpl.MaxUsage != expUsage {
				t.Errorf("Expected usage <%+v>, Received <%+v>", time.Duration(i)*time.Second, updateRpl.MaxUsage)
			}
			if i == 10 {
				expUsage = 0
			}
			if expAccVal != 0 {
				expAccVal = expAccVal - 1000000000
			}
			if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
				t.Error(err)
			} else if len(acnt.BalanceMap) != 1 {
				t.Errorf("Unexpected number of balances in the account: <%v>", utils.ToJSON(acnt))
			} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != expAccVal {
				t.Errorf("Expected value %+v, received <%v>",
					expAccVal, utils.ToJSON(acnt.BalanceMap[utils.MetaVoice]))
			}
		}
	})
}

func TestBlockerTrueProcessEvent(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_voice,*voice,,*any,,,*unlimited,,10s,10,true,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("ProcessEvent", func(t *testing.T) {
		args := &sessions.V1ProcessEventArgs{
			Flags: []string{utils.MetaCDRs},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockTrueProcEv",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestBalBlockTrue",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        30 * time.Second,
				},
			},
		}
		expRply := &sessions.V1ProcessEventReply{}
		var reply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			args, &reply); err != nil {
			t.Error(err)
		} else if utils.ToJSON(expRply) != utils.ToJSON(reply) {
			t.Errorf("Expected <%+v>, \nreceived <%+v>", expRply, reply)
		}
		var acnt *engine.Account
		attrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Error(err)
		} else if len(acnt.BalanceMap) != 1 {
			t.Errorf("Unexpected number of balances in the account: <%v>", utils.ToJSON(acnt))
		} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != 10000000000 {
			// it shouldnt take from the balance since its trying to take more than the balance has
			t.Errorf("Expected value %+v, received <%v>",
				10000000000, utils.ToJSON(acnt.BalanceMap[utils.MetaVoice]))
		}

		var cdrCnt int64
		req := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{}}
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRsCount, req, &cdrCnt); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if cdrCnt != 1 {
			t.Error("Unexpected number of CDRs returned: ", cdrCnt)
		}

		var cdrs []*engine.CDR
		argsCdr := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}}}
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, argsCdr, &cdrs); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if len(cdrs) != 1 {
			t.Error("Unexpected number of CDRs returned: ", len(cdrs))
		} else {
			if cdrs[0].Cost != -1.0 {
				t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
			}
			if cdrs[0].ExtraInfo != utils.ErrInsufficientCreditBalanceBlocker.Error() {
				t.Errorf("Expected error <%v>, received <%v>", utils.ErrInsufficientCreditBalanceBlocker, cdrs[0].ExtraInfo)
			}
		}

	})
}

func TestBlockerFalseAuthorize(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_voice,*voice,,*any,,,*unlimited,,10s,10,false,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("Authorize", func(t *testing.T) {
		args := &sessions.V1AuthorizeArgs{
			GetMaxUsage: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockFalseAuth",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestBalBlockFalse",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        "30s",
				},
			},
		}
		var rplyFirst sessions.V1AuthorizeReply
		if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rplyFirst); err != nil {
			t.Fatal(err)
		}
		if *rplyFirst.MaxUsage != 10*time.Second {
			t.Errorf("Expected usage <%+v>, Received <%+v>", 10*time.Second, rplyFirst.MaxUsage)
		}
	})
}

func TestBlockerFalseInterimUpdate(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_voice,*voice,,*any,,,*unlimited,,10s,10,false,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("InterimUpdate30s", func(t *testing.T) {
		updateArgs := &sessions.V1UpdateSessionArgs{
			UpdateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockFalseUpdate",
				Event: map[string]any{
					utils.OriginID:     "TestBalBlockFalse",
					utils.ToR:          utils.MetaVoice,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.RequestType:  utils.MetaPrepaid,
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        "1s",
				},
			},
		}
		var acnt *engine.Account
		attrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		expAccVal := 10000000000.0
		expUsage := 1 * time.Second
		for i := 1; i != 30; i++ {
			var updateRpl *sessions.V1UpdateSessionReply
			if err := client.Call(context.Background(), utils.SessionSv1UpdateSession,
				updateArgs, &updateRpl); err != nil {
				t.Error(err)
			}
			if *updateRpl.MaxUsage != expUsage {
				t.Errorf("Expected usage <%+v>, Received <%+v>", time.Duration(i)*time.Second, updateRpl.MaxUsage)
			}
			if i == 10 { // After 10 runs, maxUsage should be 0
				expUsage = 0
			}
			if expAccVal != 0 { // After 10 runs, balance should be 0
				expAccVal = expAccVal - 1000000000
			}
			if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
				t.Error(err)
			} else if len(acnt.BalanceMap) != 1 {
				t.Errorf("Unexpected number of balances in the account: <%v>", utils.ToJSON(acnt))
			} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != expAccVal {
				t.Errorf("Expected value %+v, received <%v>",
					expAccVal, utils.ToJSON(acnt.BalanceMap[utils.MetaVoice]))
			}
		}
	})
}

func TestBlockerFalseProcessEvent(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_voice,*voice,,*any,,,*unlimited,,10s,10,false,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("ProcessEvent", func(t *testing.T) {
		args := &sessions.V1ProcessEventArgs{
			Flags: []string{utils.MetaCDRs},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockFalseProcEv",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestBalBlockFalse",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        30 * time.Second,
				},
			},
		}
		expRply := &sessions.V1ProcessEventReply{}
		var reply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			args, &reply); err != nil {
			t.Error(err)
		} else if utils.ToJSON(expRply) != utils.ToJSON(reply) {
			t.Errorf("Expected <%+v>, \nreceived <%+v>", expRply, reply)
		}
		var acnt *engine.Account
		attrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Error(err)
		} else if len(acnt.BalanceMap) != 2 {
			t.Errorf("Unexpected number of balances in the account: <%v>", utils.ToJSON(acnt))
		} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 0 {
			t.Errorf("Expected value %+v, received <%v>",
				0, utils.ToJSON(acnt.BalanceMap[utils.MetaVoice]))
		} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != -20 {
			t.Errorf("Expected value %+v, received <%v>",
				-20, utils.ToJSON(acnt.BalanceMap[utils.MetaMonetary]))
		}

		var cdrCnt int64
		req := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{}}
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRsCount, req, &cdrCnt); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if cdrCnt != 1 {
			t.Error("Unexpected number of CDRs returned: ", cdrCnt)
		}

		var cdrs []*engine.CDR
		argsCdr := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}}}
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, argsCdr, &cdrs); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if len(cdrs) != 1 {
			t.Error("Unexpected number of CDRs returned: ", len(cdrs))
		} else {
			if cdrs[0].Cost != 20 { // cost should be equal to 20 seconds
				t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
			}
			if cdrs[0].ExtraInfo != "" {
				t.Errorf("Expected no error, received <%v>", cdrs[0].ExtraInfo)
			}
		}

	})
}

// Testing monetary to test ConnectFee
func TestBlockerTrueMonetaryAuthorize(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,*any,,,*unlimited,,10,10,true,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,1,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("Authorize", func(t *testing.T) {
		args := &sessions.V1AuthorizeArgs{
			GetMaxUsage: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockTrueAuth",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestBalBlockTrue",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        30,
				},
			},
		}
		var rplyFirst sessions.V1AuthorizeReply
		if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rplyFirst); err != nil {
			t.Fatal(err)
		}
		if *rplyFirst.MaxUsage != 9*time.Nanosecond {
			t.Errorf("Expected usage <%+v>, Received <%+v>", 9*time.Nanosecond, rplyFirst.MaxUsage)
		}
	})
}

// Testing monetary to test ConnectFee
func TestBlockerTrueMonetaryInterimUpdate(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,*any,,,*unlimited,,10,10,true,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,1,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("InterimUpdate30s", func(t *testing.T) {
		updateArgs := &sessions.V1UpdateSessionArgs{
			UpdateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockTrueUpdate",
				Event: map[string]any{
					utils.OriginID:     "TestBalBlockTrue",
					utils.ToR:          utils.MetaVoice,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.RequestType:  utils.MetaPrepaid,
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        1,
				},
			},
		}
		var acnt *engine.Account
		attrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		expAccVal := 9.0 // Expecting 1 unit to be already taken because of connectFee
		expUsage := 1 * time.Nanosecond
		for i := 1; i != 30; i++ {
			var updateRpl *sessions.V1UpdateSessionReply
			if err := client.Call(context.Background(), utils.SessionSv1UpdateSession,
				updateArgs, &updateRpl); err != nil {
				t.Error(err)
			}
			if *updateRpl.MaxUsage != expUsage {
				t.Errorf("Expected usage <%+v>, Received <%+v>", time.Duration(i)*time.Nanosecond, updateRpl.MaxUsage)
			}
			if i == 9 {
				expUsage = 0
			}
			if expAccVal != 0 {
				expAccVal = expAccVal - 1
			}
			if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
				t.Error(err)
			} else if len(acnt.BalanceMap) != 1 {
				t.Errorf("Unexpected number of balances in the account: <%v>", utils.ToJSON(acnt))
			} else if rply := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rply != expAccVal {
				t.Errorf("Expected value %+v, received <%v>",
					expAccVal, utils.ToJSON(acnt.BalanceMap[utils.MetaMonetary]))
			}
		}
	})
}

// Testing monetary to test ConnectFee
func TestBlockerTrueMonetaryProcessEvent(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,*any,,,*unlimited,,10,10,true,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,1,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("ProcessEvent", func(t *testing.T) {
		args := &sessions.V1ProcessEventArgs{
			Flags: []string{utils.MetaCDRs},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockTrueProcEv",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestBalBlockTrue",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        30,
				},
			},
		}
		expRply := &sessions.V1ProcessEventReply{}
		var reply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			args, &reply); err != nil {
			t.Error(err)
		} else if utils.ToJSON(expRply) != utils.ToJSON(reply) {
			t.Errorf("Expected <%+v>, \nreceived <%+v>", expRply, reply)
		}
		var acnt *engine.Account
		attrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Error(err)
		} else if len(acnt.BalanceMap) != 1 {
			t.Errorf("Unexpected number of balances in the account: <%v>", utils.ToJSON(acnt))
		} else if rply := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rply != 10 {
			// it shouldnt take from the balance since its trying to take more than the balance has
			t.Errorf("Expected value %+v, received <%v>",
				10, utils.ToJSON(acnt.BalanceMap[utils.MetaMonetary]))
		}

		var cdrCnt int64
		req := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{}}
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRsCount, req, &cdrCnt); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if cdrCnt != 1 {
			t.Error("Unexpected number of CDRs returned: ", cdrCnt)
		}

		var cdrs []*engine.CDR
		argsCdr := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}}}
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, argsCdr, &cdrs); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if len(cdrs) != 1 {
			t.Error("Unexpected number of CDRs returned: ", len(cdrs))
		} else {
			if cdrs[0].Cost != -1.0 {
				t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
			}
			if cdrs[0].ExtraInfo != utils.ErrInsufficientCreditBalanceBlocker.Error() {
				t.Errorf("Expected error <%v>, received <%v>", utils.ErrInsufficientCreditBalanceBlocker, cdrs[0].ExtraInfo)
			}
		}

	})
}

// Testing monetary to test ConnectFee
func TestBlockerFalseMonetaryAuthorize(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,*any,,,*unlimited,,10,10,false,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,1,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("Authorize", func(t *testing.T) {
		args := &sessions.V1AuthorizeArgs{
			GetMaxUsage: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockFalseAuth",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestBalBlockFalse",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        30,
				},
			},
		}
		var rplyFirst sessions.V1AuthorizeReply
		if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rplyFirst); err != nil {
			t.Fatal(err)
		}
		if *rplyFirst.MaxUsage != 9*time.Nanosecond {
			t.Errorf("Expected usage <%+v>, Received <%+v>", 9*time.Nanosecond, rplyFirst.MaxUsage)
		}
	})
}

// Testing monetary to test ConnectFee
func TestBlockerFalseMonetaryInterimUpdate(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,*any,,,*unlimited,,10,10,false,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,1,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("InterimUpdate30s", func(t *testing.T) {
		updateArgs := &sessions.V1UpdateSessionArgs{
			UpdateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockFalseUpdate",
				Event: map[string]any{
					utils.OriginID:     "TestBalBlockFalse",
					utils.ToR:          utils.MetaVoice,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.RequestType:  utils.MetaPrepaid,
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        1,
				},
			},
		}
		var acnt *engine.Account
		attrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		expAccVal := 9.0 // Expecting 1 unit to be already taken because of connectFee
		expUsage := 1 * time.Nanosecond
		for i := 1; i != 30; i++ {
			var updateRpl *sessions.V1UpdateSessionReply
			if err := client.Call(context.Background(), utils.SessionSv1UpdateSession,
				updateArgs, &updateRpl); err != nil {
				t.Error(err)
			}
			if *updateRpl.MaxUsage != expUsage {
				t.Errorf("Expected usage <%+v>, Received <%+v>", time.Duration(i)*time.Nanosecond, updateRpl.MaxUsage)
			}
			if i == 9 { // After 9 runs, maxUsage should be 0
				expUsage = 0
			}
			if expAccVal != 0 { // After 10 runs, balance should be 0
				expAccVal = expAccVal - 1
			}
			if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
				t.Error(err)
			} else if len(acnt.BalanceMap) != 1 {
				t.Errorf("Unexpected number of balances in the account: <%v>", utils.ToJSON(acnt))
			} else if rply := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rply != expAccVal {
				t.Errorf("Expected value %+v, received <%v>",
					expAccVal, utils.ToJSON(acnt.BalanceMap[utils.MetaMonetary]))
			}
		}
	})
}

// Testing monetary to test ConnectFee
func TestBlockerFalseMonetaryProcessEvent(t *testing.T) {
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
	"rals_conns": ["*internal"],
	"session_cost_retries": 0,
},

"chargers": {
	"enabled": true,
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],	
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,*any,,,*unlimited,,10,10,false,false,20`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,1,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("ProcessEvent", func(t *testing.T) {
		args := &sessions.V1ProcessEventArgs{
			Flags: []string{utils.MetaCDRs},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBalBlockFalseProcEv",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestBalBlockFalse",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2024, time.August, 9, 9, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.August, 9, 9, 0, 1, 0, time.UTC),
					utils.Usage:        30,
				},
			},
		}
		expRply := &sessions.V1ProcessEventReply{}
		var reply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			args, &reply); err != nil {
			t.Error(err)
		} else if utils.ToJSON(expRply) != utils.ToJSON(reply) {
			t.Errorf("Expected <%+v>, \nreceived <%+v>", expRply, reply)
		}
		var acnt *engine.Account
		attrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Error(err)
		} else if len(acnt.BalanceMap) != 1 {
			t.Errorf("Unexpected number of balances in the account: <%v>", utils.ToJSON(acnt))
		} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 2 {
			t.Errorf("Unexpected number of monetary balances in the account: <%v>", utils.ToJSON(acnt))
		} else if acnt.BalanceMap[utils.MetaMonetary][0].GetValue() != 0 {
			t.Errorf("Expected value %+v, received <%v>",
				0, utils.ToJSON(acnt.BalanceMap[utils.MetaMonetary][0]))
		} else if acnt.BalanceMap[utils.MetaMonetary][1].GetValue() != -21 {
			t.Errorf("Expected value %+v, received <%v>",
				-21, utils.ToJSON(acnt.BalanceMap[utils.MetaMonetary][1]))
		}

		var cdrCnt int64
		req := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{}}
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRsCount, req, &cdrCnt); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if cdrCnt != 1 {
			t.Error("Unexpected number of CDRs returned: ", cdrCnt)
		}

		var cdrs []*engine.CDR
		argsCdr := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}}}
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, argsCdr, &cdrs); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if len(cdrs) != 1 {
			t.Error("Unexpected number of CDRs returned: ", len(cdrs))
		} else {
			if cdrs[0].Cost != 31 { // since there is no voice balance this time, the cost includes the precreated monetary balances used cost
				t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
			}
			if cdrs[0].ExtraInfo != "" {
				t.Errorf("Expected no error, received <%v>", cdrs[0].ExtraInfo)
			}
		}

	})
}
