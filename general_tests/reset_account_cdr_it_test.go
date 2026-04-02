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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestResetAccountCDR(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	ng := engine.TestEngine{
		ConfigJSON: `{
"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"attributes":{
	"enabled": true,
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
	}
`,
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1002,PACKAGE_1002,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1002,ACT_TOPUP,*asap,10`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,,*monetary,,*any,,,*unlimited,,100,10,true,false,20`,
			utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
			utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1,1,0`,
			utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
			utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2024-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,1002,2024-01-14T00:00:00Z,RP_ANY,`,
		},
	}
	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	var reply string
	if err := client.Call(context.Background(), utils.APIerSv2SetActions,
		&utils.AttrSetActions{
			ActionsId: "ACT_RESET_CDR",
			Actions: []*utils.TPAction{
				{Identifier: utils.MetaCDRAccount, ExpiryTime: utils.MetaUnlimited, Weight: 20},
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	t.Run("RestoreBalanceMatchByID", func(t *testing.T) {

		if err := client.Call(context.Background(), utils.APIerSv1SetBalances,
			&utils.AttrSetBalances{
				Tenant:  "cgrates.org",
				Account: "1001",
				Balances: []*utils.AttrBalance{
					{BalanceType: utils.MetaMonetary, Value: 100, Balance: map[string]any{utils.ID: "Balance1"}},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}

		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     utils.GenUUID(),
					Event: map[string]any{
						utils.ToR:          utils.MetaMonetary,
						utils.Category:     "call",
						utils.OriginID:     "processCDR",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaPostpaid,
						utils.AccountField: "1001",
						utils.Subject:      "1001",
						utils.Destination:  "1002",
						utils.AnswerTime:   time.Date(2025, 2, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        30,
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}

		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs,
			&utils.RPCCDRsFilterWithAPIOpts{
				RPCCDRsFilter: &utils.RPCCDRsFilter{
					Accounts: []string{"1001"},
				},
			}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) == 0 || cdrs[0].CostDetails == nil || cdrs[0].CostDetails.AccountSummary == nil {
			t.Fatal("expected CDR for account")
		}

		var cdrBalance1UUID string
		for _, bs := range cdrs[0].CostDetails.AccountSummary.BalanceSummaries {
			cdrBalance1UUID = bs.UUID
		}

		// resetting balance along with its UUID
		if err := client.Call(context.Background(), utils.APIerSv1RemoveBalances,
			&utils.AttrSetBalance{
				Tenant:      "cgrates.org",
				Account:     "1001",
				BalanceType: utils.MetaMonetary,
				Balance:     map[string]any{utils.ID: "Balance1"},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetBalances,
			&utils.AttrSetBalances{
				Tenant:  "cgrates.org",
				Account: "1001",
				Balances: []*utils.AttrBalance{
					{BalanceType: utils.MetaMonetary, Value: 100, Balance: map[string]any{utils.ID: "Balance1"}},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}

		// execute newly reset Balance  to be updated from the one in CDRs
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{
				Tenant:    "cgrates.org",
				Account:   "1001",
				ActionsId: "ACT_RESET_CDR",
			}, &reply); err != nil {
			t.Fatal(err)
		}

		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acnt); err != nil {
			t.Fatal(err)
		}

		var restoredBalance1 *engine.Balance
		for _, b := range acnt.BalanceMap[utils.MetaMonetary] {
			if b.ID == "Balance1" {
				restoredBalance1 = b
			}
		}
		// check if newly balance got updated by matched only by ID
		if restoredBalance1.Uuid == cdrBalance1UUID || restoredBalance1.Value != 70 {
			t.Errorf("Balance1 didn't got reset by CDR balance %v", restoredBalance1.Value)
		}
	})

	t.Run("RestoreBalanceWithoutIDMatchByUUID", func(t *testing.T) {

		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"}, &acnt); err != nil {
			t.Fatal(err)
		}
		noIDBalUUID := acnt.BalanceMap[utils.MetaMonetary][0].Uuid
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     utils.GenUUID(),
					Event: map[string]any{
						utils.ToR:          utils.MetaMonetary,
						utils.Category:     "call",
						utils.OriginID:     "processCDR2",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaPostpaid,
						utils.AccountField: "1002",
						utils.Subject:      "1002",
						utils.Destination:  "1003",
						utils.AnswerTime:   time.Date(2025, 2, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        40,
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}

		// reset the existing balance value
		if err := client.Call(context.Background(), utils.APIerSv1SetBalances,
			&utils.AttrSetBalances{
				Tenant:  "cgrates.org",
				Account: "1002",
				Balances: []*utils.AttrBalance{
					{BalanceType: utils.MetaMonetary, Value: 100, Balance: map[string]any{utils.UUID: noIDBalUUID}},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}

		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{
				Tenant:    "cgrates.org",
				Account:   "1002",
				ActionsId: "ACT_RESET_CDR",
			}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"}, &acnt); err != nil {
			t.Fatal(err)
		}

		var restoredBalanceNoID *engine.Balance
		for _, b := range acnt.BalanceMap[utils.MetaMonetary] {
			restoredBalanceNoID = b
		}
		// check if balance got updated matching with UUID
		if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 || restoredBalanceNoID.Value != 60 || restoredBalanceNoID.Uuid != noIDBalUUID {
			t.Errorf("Balance didn't got reset by CDR balance %v", restoredBalanceNoID.Value)
		}
	})

}
