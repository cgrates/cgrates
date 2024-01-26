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
RT_ANY,0,0.6,60s,1s,0s`,
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
// Previously, the 'Factor' was being populated from the Action's ExtraParameters map, where the key represented the ToR of
// the session being processed. This has now been updated to depend on Category instead of ToR.
//
// The test steps are as follows:
//  1. Create an account with an *sms balance of 10 units with a factor of 0.25 (essentially, this means that for every 1 sms, 4 will
//     be exhausted), and a *monetary balance of 5 units. The RatingPlan used when debiting the *monetary balance will charge 1 unit
//     per second.
//  2. Process an 3 usage (representing 12 sms, when taking into consideration the balance factor) event.
//  3. Ensure that the *sms balance has 2 units left (10 - (2 sms / 0.25 factor)) and that 1 unit was subtracted from the *monetary balance.
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
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,"{""sms"":0.25}",,balance_sms,*sms,,,,,*unlimited,,10,20,false,false,20
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,*any,,,*unlimited,,5,10,false,false,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,sms,1001,2014-01-14T00:00:00Z,RP_ANY,`,
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
		if smsBalance.ID != "balance_sms" || smsBalance.Value != 10 {
			t.Fatalf("received account with unexpected *sms balance: %v", smsBalance)
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "balance_monetary" || monetaryBalance.Value != 5 {
			t.Fatalf("received account with unexpected *monetary balance: %v", monetaryBalance)
		}
	})

	t.Run("ProcessCDRAndCheckBalance", func(t *testing.T) {
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
						utils.OriginID:     "processCDR",
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
				RunIDs: []string{"*default"},
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
		monetaryBalanceValue, err := cdrs[0].CostDetails.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[1]", "Value"})
		if err != nil {
			t.Fatalf("could not retrieve *sms balance current value: %v", err)
		}
		if monetaryBalanceValue != 4. {
			t.Errorf("unexpected balance value: expected %v, received %v", 4., monetaryBalanceValue)
		}
	})
}
