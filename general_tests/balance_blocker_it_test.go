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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestBalanceBlocker tests the usage of the 'Blocker' field in account balances (issue #4163).
//
// Previously, the 'Blocker' field was ignored regardless if it was set to true of false. The test ensures that when 'Blocker'
// is set to true for a certain balance, no additional balances are debited once that balance is exhausted.
//
// The test steps are as follows:
// 	1. Create an account with a single *monetary balance of 1 unit with 'Blocker' set to true.
// 	2. Process a 2-minute usage CDR charging 0.6 units per minute (total charge of 1.2 units).
// 	3. Verify that the account is left with a single empty balance, as the 'Blocker' should prevent overcharging.
//
// Before the fix, the account would have two balances post processing:
// one empty and another with -0.2 units (with id *default).

func TestBalanceBlocker(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

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
		fmt.Println(utils.ToJSON(acnt))
	})

	t.Run("ProcessCDR", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "run_1",
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
						utils.Usage:        11,
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				RunIDs: []string{"run_1"},
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}

		fmt.Println(utils.ToJSON(cdrs))
	})

	t.Run("CheckFinalBalance", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		fmt.Println(utils.ToJSON(acnt))
	})
}
