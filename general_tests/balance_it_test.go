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

// TestBalanceBlocker tests the usage of the 'Blocker' field in account balances (issue #4163).
//
// Previously, the 'Blocker' field was ignored regardless if it was set to true of false. The test ensures that when 'Blocker'
// is set to true for a certain balance, no additional balances are debited once that balance is exhausted, while also returning
// an error mentioning the lack of funds.
//
// The test steps are as follows:
//  1. Create an account with an *sms balance of 10 units with 'Blocker' set to true, and a *monetary balance of 1 unit.
//  2. Process an 8 usage (representing 8 sms) event and then check whether the *sms balance has 2 units left.
//  3. Process another event, but this time with 'Usage' set to 3. This time an error should occur during processing, due to *sms balance blocking
//     access to subsequent balances required to charge the remaining unit. The error can be found in the CDR's ExtraInfo field.
//  4. Verify that the account's *sms balance is the same as it was in step 2 and that the *monetary balance has not been touched
//     at all.
func TestBalanceBlocker(t *testing.T) {
	switch *dbType {
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
	"enabled": true,
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*internal"]
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
ACT_TOPUP,*topup_reset,,,balance_sms,*sms,,,,,*unlimited,,10,20,true,false,20
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,*any,,,*unlimited,,1,10,false,false,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,sms,1001,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	testEnv := TestEnvironment{
		Name: "TestBalanceBlocker",
		// Encoding:   *encoding,
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _, shutdown, err := testEnv.Setup(t, *waitRater)
	if err != nil {
		t.Fatal(err)
	}

	defer shutdown()

	t.Run("CheckInitialBalance", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 2 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 1 ||
			len(acnt.BalanceMap[utils.MetaSMS]) != 1 {
			t.Fatalf("expected account to have one balance of type *monetary and one of type *sms, received %v", acnt)
		}
		smsBalance := acnt.BalanceMap[utils.MetaSMS][0]
		if smsBalance.ID != "balance_sms" || smsBalance.Value != 10 || !smsBalance.Blocker {
			t.Fatalf("received account with unexpected *sms balance: %v", smsBalance)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 1 {
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
						utils.Category:     "sms",
						utils.ToR:          utils.MetaSMS,
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
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				RunIDs:    []string{"*default"},
				OriginIDs: []string{"processCDR1"},
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("expected to receive only one CDR: %v", utils.ToJSON(cdrs))
		}
		smsBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[0]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *sms balance current value: %v", err)
		}
		if smsBalanceValue != 2. {
			t.Errorf("unexpected balance value: expected %v, received %v", 2., smsBalanceValue)
		}
	})

	t.Run("ProcessCDR2", func(t *testing.T) {
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
						utils.Category:     "sms",
						utils.ToR:          utils.MetaSMS,
						utils.OriginID:     "processCDR2",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaPostpaid,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        3,
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				RunIDs:    []string{"*default"},
				OriginIDs: []string{"processCDR2"},
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("expected to receive only one CDR: %v", utils.ToJSON(cdrs))
		}
		if cdrs[0].ExtraInfo != utils.ErrInsufficientCreditBalanceBlocker.Error() {
			t.Errorf("Unexpected ExtraInfo field value: %v", cdrs[0].ExtraInfo)
		}
	})

	t.Run("CheckFinalBalance", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 2 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 1 ||
			len(acnt.BalanceMap[utils.MetaSMS]) != 1 {
			t.Fatalf("expected account to have one balance of type *monetary and one of type *sms, received %v", acnt)
		}
		smsBalance := acnt.BalanceMap[utils.MetaSMS][0]
		if smsBalance.ID != "balance_sms" || smsBalance.Value != 2 || !smsBalance.Blocker {
			t.Fatalf("received account with unexpected *sms balance: %v", smsBalance)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 1 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})
}

// TestBalanceFactor tests the usage of the 'Factor' field in account balances.
//
// Previously, the 'Factor' was being populated from the Action's ExtraParameters map,
// where the key represented the ToR of the session being processed. This has now been
// updated to depend on Category instead of ToR.
//
// The test steps are as follows:
//  1. Create an account with an *sms balance of 10 units with a factor of 4 (essentially,
//     this means that for every 1 sms, 4 will be exhausted), and a *monetary balance of 5
//     units. The RatingPlan used when debiting the *monetary balance will charge 1 unit per
//     second.
//  2. Process a 4 usage (representing 16 sms, when taking into consideration the balance
//     factor) event.
//  3. Ensure that the *sms balance has 2 units left (10 - (2 sms * 4 factor)) and that 2
//     unit were subtracted from the *monetary balance.
//  4. Try to refund the debit made in the previous step to check whether factor is taken
//     into consideration for refunds as well.
//  5. Do the above steps also for SessionSv1.ProcessCDR with a different usage.
//  6. Initiate a prepaid session (usage 10s), update it twice (usages 5s and 2s), terminate,
//     and process CDR.
//  7. Check to see if balance_voice was debitted 34s ((10s+5s+2s) * voiceFactor, where
//     voiceFactor is 2) and then also check if it applies correctly for refund (similar
//     to step 4).
func TestBalanceFactor(t *testing.T) {
	switch *dbType {
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
	"enabled": true,
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*internal"]
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
},

"sessions": {
	"enabled": true,
	"cdrs_conns": ["*internal"],
	"chargers_conns": ["*internal"],
	"rals_conns": ["*internal"]
},

"chargers": {
	"enabled": true
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,"{""smsFactor"":4}",,balance_sms,*sms,,,,,*unlimited,,10,20,false,false,20
ACT_TOPUP,*topup_reset,"{""voiceFactor"":2}",,balance_voice,*voice,call,,,,*unlimited,,100s,20,false,false,20
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,*any,,,*unlimited,,5,10,false,false,20`,
		utils.ChargersCsv: `#Id,ActionsId,TimingId,Weight
#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,DEFAULT,*none,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_SMS,*any,RT_SMS,*up,20,0,
DR_VOICE,*any,RT_VOICE,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_SMS,0,1,1,1,0
RT_VOICE,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_SMS,DR_SMS,*any,10
RP_VOICE,DR_VOICE,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,sms,1001,2014-01-14T00:00:00Z,RP_SMS,
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_VOICE,`,
	}

	testEnv := TestEnvironment{
		Name: "TestBalanceFactor",
		// Encoding:   *encoding,
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _, shutdown, err := testEnv.Setup(t, *waitRater)
	if err != nil {
		t.Fatal(err)
	}

	defer shutdown()

	t.Run("CheckInitialBalance", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 3 ||
			len(acnt.BalanceMap[utils.MetaMonetary]) != 1 ||
			len(acnt.BalanceMap[utils.MetaSMS]) != 1 ||
			len(acnt.BalanceMap[utils.MetaVoice]) != 1 {
			t.Fatalf("expected account to have one balance of type *monetary, one of type *sms and one of type *voice, received %v", acnt)
		}
		smsBalance := acnt.BalanceMap[utils.MetaSMS][0]
		if smsBalance.ID != "balance_sms" || smsBalance.Value != 10 {
			t.Fatalf("received account with unexpected *sms balance: %v", smsBalance)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 5 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
		voiceBalance := acnt.BalanceMap[utils.MetaVoice][0]
		if voiceBalance.ID != "balance_voice" || voiceBalance.Value != float64(100*time.Second) {
			t.Fatalf("received account with unexpected *voice balance: %v", voiceBalance)
		}
	})

	t.Run("CDRsV1ProcessCDR", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "CDRsV1ProcessCDR",
					Event: map[string]any{
						utils.RunID:           "*default",
						utils.Tenant:          "cgrates.org",
						utils.Category:        "sms",
						utils.ToR:             utils.MetaSMS,
						utils.OriginID:        "processCDR1",
						utils.OriginHost:      "127.0.0.1",
						utils.RequestType:     utils.MetaPostpaid,
						utils.AccountField:    "1001",
						utils.Destination:     "1002",
						utils.SetupTime:       time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:      time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:           4,
						utils.BalanceFactorID: "smsFactor",
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				OriginIDs: []string{"processCDR1"},
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}

		if len(cdrs) != 1 {
			t.Fatalf("expected to receive only one CDR: %v", utils.ToJSON(cdrs))
		}
		smsBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[0]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *sms balance current value: %v", err)
		}
		if smsBalanceValue != 2. {
			t.Errorf("unexpected balance value: expected %v, received %v", 2., smsBalanceValue)
		}
		monetaryBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[2]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *sms balance current value: %v", err)
		}
		if monetaryBalanceValue != 3. {
			t.Errorf("unexpected balance value: expected %v, received %v", 3., monetaryBalanceValue)
		}

		// Attempt refund to check if factor also applies when refunding increments.
		//
		// Initial *sms balance value: 10
		// Initial *monetary balance value: 5
		// CDR Usage: 4
		// Factor: 4
		// Current *sms balance value: 2
		// Current *monetary balance value: 3
		var replyProcessEvent string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags:    []string{utils.MetaRefund, "*store:false"},
				CGREvent: *cdrs[0].AsCGREvent(),
			}, &replyProcessEvent); err != nil {
			t.Fatal(err)
		}
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: "1001",
			}, &acnt); err != nil {
			t.Fatal(err)
		}
		smsBalance := acnt.BalanceMap[utils.MetaSMS][0]
		if smsBalance.ID != "balance_sms" || smsBalance.Value != 10 {
			t.Fatalf("received account with unexpected *sms balance: %v", smsBalance)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 5 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
		smsBalanceFactor, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[0]", "Factors", "smsFactor"})
		if err != nil {
			t.Fatalf("could not retrieve *sms balance factor: %v", err)
		}
		if smsBalanceFactor != 4. {
			t.Errorf("unexpected balance factor: expected %v, received %v", 4., smsBalanceValue)
		}
	})

	t.Run("SessionSv1ProcessCDR", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.SessionSv1ProcessCDR,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "SessionSv1ProcessCDR",
				Event: map[string]any{
					utils.RunID:           "*default",
					utils.Tenant:          "cgrates.org",
					utils.Category:        "sms",
					utils.ToR:             utils.MetaSMS,
					utils.OriginID:        "processCDR2",
					utils.OriginHost:      "127.0.0.1",
					utils.RequestType:     utils.MetaPostpaid,
					utils.AccountField:    "1001",
					utils.Destination:     "1002",
					utils.SetupTime:       time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
					utils.AnswerTime:      time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
					utils.Usage:           6,
					utils.BalanceFactorID: "smsFactor",
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				OriginIDs: []string{"processCDR2"},
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}

		if len(cdrs) != 1 {
			t.Fatalf("expected to receive only one CDR: %v", utils.ToJSON(cdrs))
		}
		smsBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[0]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *sms balance current value: %v", err)
		}
		if smsBalanceValue != 2. {
			t.Errorf("unexpected balance value: expected %v, received %v", 2., smsBalanceValue)
		}
		monetaryBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[2]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *sms balance current value: %v", err)
		}
		if monetaryBalanceValue != 1. {
			t.Errorf("unexpected balance value: expected %v, received %v", 1., monetaryBalanceValue)
		}
	})

	t.Run("PrepaidSession", func(t *testing.T) {
		var replyInit sessions.V1InitSessionReply
		if err := client.Call(context.Background(), utils.SessionSv1InitiateSession,
			&sessions.V1InitSessionArgs{
				InitSession: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "SessionSv1InitiateSession",
					Event: map[string]any{
						utils.OriginID:        "prepaidSession",
						utils.Tenant:          "cgrates.org",
						utils.Category:        "call",
						utils.ToR:             utils.MetaVoice,
						utils.RequestType:     utils.MetaPrepaid,
						utils.AccountField:    "1001",
						utils.Subject:         "1001",
						utils.Destination:     "1002",
						utils.SetupTime:       time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
						utils.AnswerTime:      time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
						utils.Usage:           10 * time.Second,
						utils.BalanceFactorID: "voiceFactor",
					},
					APIOpts: map[string]any{
						utils.OptsDebitInterval: 0,
					},
				},
			}, &replyInit); err != nil {
			t.Error(err)
		}

		var replyUpdate sessions.V1UpdateSessionReply
		if err := client.Call(context.Background(), utils.SessionSv1UpdateSession,
			&sessions.V1UpdateSessionArgs{
				UpdateSession: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "SessionSv1UpdateSession1",
					Event: map[string]any{
						utils.OriginID:        "prepaidSession",
						utils.Tenant:          "cgrates.org",
						utils.Category:        "call",
						utils.ToR:             utils.MetaVoice,
						utils.RequestType:     utils.MetaPrepaid,
						utils.AccountField:    "1001",
						utils.Subject:         "1001",
						utils.Destination:     "1002",
						utils.SetupTime:       time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
						utils.AnswerTime:      time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
						utils.Usage:           5 * time.Second,
						utils.BalanceFactorID: "voiceFactor",
					},
					APIOpts: map[string]any{
						utils.OptsDebitInterval: 0,
					},
				},
			}, &replyUpdate); err != nil {
			t.Error(err)
		}

		if err := client.Call(context.Background(), utils.SessionSv1UpdateSession,
			&sessions.V1UpdateSessionArgs{
				UpdateSession: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "SessionSv1UpdateSession2",
					Event: map[string]any{
						utils.OriginID:        "prepaidSession",
						utils.Tenant:          "cgrates.org",
						utils.Category:        "call",
						utils.ToR:             utils.MetaVoice,
						utils.RequestType:     utils.MetaPrepaid,
						utils.AccountField:    "1001",
						utils.Subject:         "1001",
						utils.Destination:     "1002",
						utils.SetupTime:       time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
						utils.AnswerTime:      time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
						utils.Usage:           2 * time.Second,
						utils.BalanceFactorID: "voiceFactor",
					},
					APIOpts: map[string]any{
						utils.OptsDebitInterval: 0,
					},
				},
			}, &replyUpdate); err != nil {
			t.Error(err)
		}

		var replyTerminate string
		if err := client.Call(context.Background(), utils.SessionSv1TerminateSession,
			&sessions.V1TerminateSessionArgs{
				TerminateSession: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "SessionSv1TerminateSession",
					Event: map[string]any{
						utils.OriginID:     "prepaidSession",
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaVoice,
						utils.RequestType:  utils.MetaPrepaid,
						utils.AccountField: "1001",
						utils.Subject:      "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
					},
					APIOpts: map[string]any{
						utils.OptsDebitInterval: 0,
					},
				},
			}, &replyTerminate); err != nil {
			t.Error(err)
		}

		var replyProcessCDR string
		if err := client.Call(context.Background(), utils.SessionSv1ProcessCDR,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testSesRnd2PrepaidProcessCDR",
				Event: map[string]any{
					utils.OriginID:     "prepaidSession",
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.MetaVoice,
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
					utils.AnswerTime:   time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
					utils.Usage:        0,
				},
				APIOpts: map[string]any{
					utils.OptsDebitInterval: 0,
				},
			}, &replyProcessCDR); err != nil {
			t.Error(err)
		}

		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				OriginIDs: []string{"prepaidSession"},
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}

		if len(cdrs) != 1 {
			t.Fatalf("expected to receive only one CDR: %v", utils.ToJSON(cdrs))
		}
		voiceBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[1]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *voice balance current value: %v", err)
		}
		if voiceBalanceValue != float64(66*time.Second) {
			t.Errorf("unexpected balance value: expected %v, received %v", float64(66*time.Second), voiceBalanceValue)
		}

		// Attempt refund to check if factor also applies when refunding increments.
		//
		// Initial *voice balance value (before ProcessCDR): 100s
		// CDR Usage: 17s
		// Factor: 2
		// Current *voice balance value: 66s
		var replyProcessEvent string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags:    []string{utils.MetaRefund, "*store:false"},
				CGREvent: *cdrs[0].AsCGREvent(),
			}, &replyProcessEvent); err != nil {
			t.Fatal(err)
		}
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: "1001",
			}, &acnt); err != nil {
			t.Fatal(err)
		}
		voiceBalance := acnt.BalanceMap[utils.MetaVoice][0]
		if voiceBalance.ID != "balance_voice" || voiceBalance.Value != float64(100*time.Second) {
			t.Fatalf("received account with unexpected *voice balance: %v", voiceBalance)
		}
	})
}

// TestBalanceCDRLog tests the usage of balance related actions together with a "*cdrlog" action.
//
// The test steps are as follows:
//  1. Create an account with 2 balances of types *sms and *monetary. The topup action for the *monetary one will also include
//     the creation of a CDR.
//  2. Set 3 action bundles with "*topup_reset", "*remove_balance" and "*remove_expired" actions, each coupled with a "*cdrlog" action.
//  3. Retrieve the CDRs and check whether the their fields are set correctly.
func TestBalanceCDRLog(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"cdrs": {
	"enabled": true,
},

"schedulers": {
	"enabled": true,
	"cdrs_conns": ["*localhost"]
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,ACC_TEST,PACKAGE_ACC_TEST,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_ACC_TEST,ACT_TOPUP_MONETARY,*asap,10
PACKAGE_ACC_TEST,ACT_TOPUP_SMS,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_REMOVE_BALANCE_MONETARY,*remove_balance,,,balance_monetary,*monetary,,,,,,,,,,,
ACT_REMOVE_BALANCE_MONETARY,*cdrlog,,,,,,,,,,,,,,,
ACT_REMOVE_EXPIRED_WITH_CATEGORY,*remove_expired,,,,,category2,,,,,,,,,,
ACT_REMOVE_EXPIRED_WITH_CATEGORY,*cdrlog,,,,,,,,,,,,,,,
ACT_REMOVE_EXPIRED,*remove_expired,,,,,,,,,,,,,,,
ACT_REMOVE_EXPIRED,*cdrlog,,,,,,,,,,,,,,,
ACT_TOPUP_MONETARY,*topup_reset,,,balance_monetary,*monetary,,*any,,,*unlimited,,150,20,false,false,20
ACT_TOPUP_MONETARY,*cdrlog,"{""BalanceID"":""~*acnt.BalanceID""}",,,,,,,,,,,,,,
ACT_TOPUP_SMS,*topup_reset,,,balance_sms,*sms,,*any,,,*unlimited,,1000,10,false,false,10`,
	}

	testEnv := TestEnvironment{
		Name: "TestBalanceCDRLog",
		// Encoding:   *encoding,
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _, shutdown, err := testEnv.Setup(t, *waitRater)
	if err != nil {
		t.Fatal(err)
	}

	defer shutdown()

	t.Run("CheckInitialBalances", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: "ACC_TEST",
			}, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 2 || len(acnt.BalanceMap[utils.MetaMonetary]) != 1 || len(acnt.BalanceMap[utils.MetaSMS]) != 1 {
			t.Errorf("unexpected accont received: %v", utils.ToJSON(acnt))
		}
	})

	t.Run("CheckTopupCDR", func(t *testing.T) {
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{}}, &cdrs); err != nil {
			t.Fatal(err)
		}

		if len(cdrs) != 1 ||
			cdrs[0].RunID != utils.MetaTopUpReset ||
			cdrs[0].Source != utils.CDRLog ||
			cdrs[0].ToR != utils.MetaMonetary ||
			cdrs[0].ExtraFields["BalanceID"] != "balance_monetary" ||
			cdrs[0].Cost != 150 {
			t.Errorf("unexpected cdr received: %v", utils.ToJSON(cdrs))
		}
	})

	t.Run("RemoveMonetaryBalance", func(t *testing.T) {
		var reply string
		attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "ACC_TEST", ActionsId: "ACT_REMOVE_BALANCE_MONETARY"}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckRemoveBalanceCDR", func(t *testing.T) {
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				RunIDs: []string{"*remove_balance"},
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}

		if len(cdrs) != 1 ||
			cdrs[0].RunID != utils.MetaRemoveBalance ||
			cdrs[0].Source != utils.CDRLog ||
			cdrs[0].ToR != utils.MetaMonetary ||
			cdrs[0].ExtraFields["BalanceID"] != "balance_monetary" ||
			cdrs[0].Cost != 150 {
			t.Errorf("unexpected cdr received: %v", utils.ToJSON(cdrs))
		}
	})

	t.Run("CheckFinalBalances", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: "ACC_TEST",
			}, &acnt); err != nil {
			t.Error(err)
		}
		if len(acnt.BalanceMap) != 2 || len(acnt.BalanceMap[utils.MetaSMS]) != 1 || len(acnt.BalanceMap[utils.MetaMonetary]) != 0 {
			t.Errorf("unexpected account received: %v", utils.ToJSON(acnt))
		}
	})

	t.Run("RemoveExpiredBalancesNoFilter", func(t *testing.T) {
		expiryTime := time.Now().Add(10 * time.Millisecond)
		attrSetBalance := utils.AttrSetBalances{
			Tenant:  "cgrates.org",
			Account: "ACC_TEST",
			Balances: []*utils.AttrBalance{
				{
					// will not be removed
					BalanceType: utils.MetaMonetary,
					Value:       10,
					Balance: map[string]any{
						utils.ID: "ValidBalanceNotMatching",
					},
				},
				{
					// will be removed
					BalanceType: utils.MetaMonetary,
					Value:       11,
					Balance: map[string]any{
						utils.ID:         "ExpiredBalanceNotMatching1",
						utils.ExpiryTime: expiryTime,
					},
				},
				{
					// will be removed
					BalanceType: utils.MetaMonetary,
					Value:       12,
					Balance: map[string]any{
						utils.ID:         "ExpiredBalanceNotMatching2",
						utils.ExpiryTime: expiryTime,
					},
				},
				{
					// will be removed
					BalanceType: utils.MetaMonetary,
					Value:       13,
					Balance: map[string]any{
						utils.ID:         "ExpiredBalanceNotMatching3",
						utils.ExpiryTime: expiryTime,
						utils.Categories: "category1;category3",
					},
				},
				{
					// will be removed
					BalanceType: utils.MetaMonetary,
					Value:       14,
					Balance: map[string]any{
						utils.ID:         "MatchingExpiredBalance",
						utils.ExpiryTime: expiryTime,
						utils.Categories: "category1;category2",
					},
				},
				{
					// will not be removed
					BalanceType: utils.MetaMonetary,
					Value:       15,
					Balance: map[string]any{
						utils.ID:         "MatchingValidBalance",
						utils.Categories: "category2",
					},
				},
				{
					// will be removed
					BalanceType: utils.MetaSMS,
					Value:       16,
					Balance: map[string]any{
						utils.ID:         "ExpiredSMSBalance",
						utils.ExpiryTime: expiryTime,
					},
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetBalances, &attrSetBalance, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(10 * time.Millisecond)
		attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "ACC_TEST", ActionsId: "ACT_REMOVE_EXPIRED"}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckRemoveExpiredCDRNoFilter", func(t *testing.T) {
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				RunIDs:  []string{"*remove_expired"},
				OrderBy: utils.Cost,
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}

		if len(cdrs) != 5 ||
			cdrs[0].Cost != 11 ||
			cdrs[0].ExtraFields[utils.BalanceID] != "ExpiredBalanceNotMatching1" ||
			cdrs[1].Cost != 12 ||
			cdrs[1].ExtraFields[utils.BalanceID] != "ExpiredBalanceNotMatching2" ||
			cdrs[2].Cost != 13 ||
			cdrs[2].ExtraFields[utils.BalanceID] != "ExpiredBalanceNotMatching3" ||
			cdrs[3].Cost != 14 ||
			cdrs[3].ExtraFields[utils.BalanceID] != "MatchingExpiredBalance" ||
			cdrs[4].Cost != 16 ||
			cdrs[4].ExtraFields[utils.BalanceID] != "ExpiredSMSBalance" {
			t.Errorf("unexpected cdrs received: %v", utils.ToJSON(cdrs))
		}

		assertCommonCDRFields := func(t *testing.T, cdr *engine.CDR, expectedType string) {
			if cdr.RunID != utils.MetaRemoveExpired ||
				cdr.Source != utils.CDRLog ||
				cdr.ToR != expectedType {
				t.Fatalf("unexpected cdrs received: %v", utils.ToJSON(cdrs))
			}
		}
		expType := utils.MetaMonetary
		for i, cdr := range cdrs {
			if i == len(cdrs)-1 {
				expType = utils.MetaSMS
			}
			assertCommonCDRFields(t, cdr, expType)
		}
	})

	t.Run("RemoveExpiredBalancesFiltered", func(t *testing.T) {

		// Remove cdrs from previous test.
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1RemoveCDRs, &utils.RPCCDRsFilter{
			RunIDs: []string{"*remove_expired"},
		}, &reply); err != nil {
			t.Fatal(err)
		}

		expiryTime := time.Now().Add(10 * time.Millisecond)
		attrSetBalance := utils.AttrSetBalances{
			Tenant:  "cgrates.org",
			Account: "ACC_TEST",
			Balances: []*utils.AttrBalance{
				{
					// will not be removed
					BalanceType: utils.MetaMonetary,
					Value:       10,
					Balance: map[string]any{
						utils.ID: "ValidBalanceNotMatching",
					},
				},
				{
					// will not be removed
					BalanceType: utils.MetaMonetary,
					Value:       11,
					Balance: map[string]any{
						utils.ID:         "ExpiredBalanceNotMatching1",
						utils.ExpiryTime: expiryTime,
					},
				},
				{
					// will not be removed
					BalanceType: utils.MetaMonetary,
					Value:       12,
					Balance: map[string]any{
						utils.ID:         "ExpiredBalanceNotMatching2",
						utils.ExpiryTime: expiryTime,
					},
				},
				{
					// will not be removed
					BalanceType: utils.MetaMonetary,
					Value:       13,
					Balance: map[string]any{
						utils.ID:         "ExpiredBalanceNotMatching3",
						utils.ExpiryTime: expiryTime,
						utils.Categories: "category1;category3",
					},
				},
				{
					// will be removed
					BalanceType: utils.MetaMonetary,
					Value:       14,
					Balance: map[string]any{
						utils.ID:         "MatchingExpiredBalance",
						utils.ExpiryTime: expiryTime,
						utils.Categories: "category1;category2",
					},
				},
				{
					// will not be removed
					BalanceType: utils.MetaMonetary,
					Value:       15,
					Balance: map[string]any{
						utils.ID:         "MatchingValidBalance",
						utils.Categories: "category2",
					},
				},
				{
					// will be removed
					BalanceType: utils.MetaSMS,
					Value:       16,
					Balance: map[string]any{
						utils.ID:         "ExpiredSMSBalance",
						utils.ExpiryTime: expiryTime,
						utils.Categories: "category1;category2",
					},
				},
			},
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetBalances, &attrSetBalance, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(10 * time.Millisecond)
		attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "ACC_TEST", ActionsId: "ACT_REMOVE_EXPIRED_WITH_CATEGORY"}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckRemoveExpiredCDRFiltered", func(t *testing.T) {
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				RunIDs:  []string{"*remove_expired"},
				OrderBy: utils.Cost,
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}

		if len(cdrs) != 2 ||
			cdrs[0].Cost != 14 ||
			cdrs[0].ExtraFields[utils.BalanceID] != "MatchingExpiredBalance" ||
			cdrs[0].RunID != utils.MetaRemoveExpired ||
			cdrs[0].Source != utils.CDRLog ||
			cdrs[0].ToR != utils.MetaMonetary ||
			cdrs[1].Cost != 16 ||
			cdrs[1].ExtraFields[utils.BalanceID] != "ExpiredSMSBalance" ||
			cdrs[1].RunID != utils.MetaRemoveExpired ||
			cdrs[1].Source != utils.CDRLog ||
			cdrs[1].ToR != utils.MetaSMS {
			t.Errorf("unexpected cdrs received: %v", utils.ToJSON(cdrs))
		}
	})
}
