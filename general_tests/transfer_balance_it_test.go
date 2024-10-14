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
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestTransferBalance tests the implementation of the "*transfer_balance" action.
//
// The test steps are as follows:
// 	1. Create two accounts with a single *monetary balance of 10 units.
// 	2. Set a "*transfer_balance" action that takes 4 units from the source balance and adds them to the destination balance.
// 	3. Execute that action using the method "APIerSv1.ExecuteAction".
// 	4. Check the balances; the source balance should now have 6 units while the destination balance should have 14.

func TestTransferBalance(t *testing.T) {
	switch *utils.DBType {
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
	"cdrs_conns": ["*internal"]
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,ACC_SRC,PACKAGE_ACC_SRC,,,
cgrates.org,ACC_DEST,PACKAGE_ACC_DEST,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_ACC_SRC,ACT_TOPUP_SRC,*asap,10
PACKAGE_ACC_DEST,ACT_TOPUP_DEST,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP_SRC,*topup_reset,,,balance_src,*monetary,,*any,,,*unlimited,,10,20,false,false,20
ACT_TOPUP_DEST,*topup_reset,,,balance_dest,*monetary,,*any,,,*unlimited,,10,10,false,false,10
ACT_TRANSFER,*transfer_balance,"{""DestinationAccountID"":""cgrates.org:ACC_DEST"",""DestinationBalanceID"":""balance_dest""}",,balance_src,,,,,,*unlimited,,4,,,,
ACT_TRANSFER,*cdrlog,,,,,,,,,,,,,,,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)

	t.Run("CheckInitialBalances", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		var acnts []*engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
			&utils.AttrGetAccounts{
				Tenant: "cgrates.org",
			}, &acnts); err != nil {
			t.Error(err)
		}
		if len(acnts) != 2 {
			t.Fatal("expecting 2 accounts to be retrieved")
		}
		sort.Slice(acnts, func(i, j int) bool {
			return acnts[i].ID > acnts[j].ID
		})
		if len(acnts[0].BalanceMap) != 1 || len(acnts[0].BalanceMap[utils.MetaMonetary]) != 1 {
			t.Fatalf("expected account to have only one balance of type *monetary, received %v", acnts[0])
		}
		balance := acnts[0].BalanceMap[utils.MetaMonetary][0]
		if balance.ID != "balance_src" || balance.Value != 10 {
			t.Fatalf("received account with unexpected balance: %v", balance)
		}
		if len(acnts[1].BalanceMap) != 1 || len(acnts[1].BalanceMap[utils.MetaMonetary]) != 1 {
			t.Fatalf("expected account to have only one balance of type *monetary, received %v", acnts[1])
		}
		balance = acnts[1].BalanceMap[utils.MetaMonetary][0]
		if balance.ID != "balance_dest" || balance.Value != 10 {
			t.Fatalf("received account with unexpected balance: %v", balance)
		}
	})

	t.Run("TransferBalance", func(t *testing.T) {
		var reply string
		attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "ACC_SRC", ActionsId: "ACT_TRANSFER"}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckBalancesAfterActionExecute", func(t *testing.T) {
		var acnts []*engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
			&utils.AttrGetAccounts{
				Tenant: "cgrates.org",
			}, &acnts); err != nil {
			t.Error(err)
		}
		if len(acnts) != 2 {
			t.Fatal("expecting 2 accounts to be retrieved")
		}
		sort.Slice(acnts, func(i, j int) bool {
			return acnts[i].ID > acnts[j].ID
		})
		if len(acnts[0].BalanceMap) != 1 || len(acnts[0].BalanceMap[utils.MetaMonetary]) != 1 {
			t.Errorf("expected account to have only one balance of type *monetary, received %v", acnts[0])
		}
		balance := acnts[0].BalanceMap[utils.MetaMonetary][0]
		if balance.ID != "balance_src" || balance.Value != 6 {
			t.Errorf("received account with unexpected balance: %v", balance)
		}
		if len(acnts[1].BalanceMap) != 1 || len(acnts[1].BalanceMap[utils.MetaMonetary]) != 1 {
			t.Errorf("expected account to have only one balance of type *monetary, received %v", acnts[1])
		}
		balance = acnts[1].BalanceMap[utils.MetaMonetary][0]
		if balance.ID != "balance_dest" || balance.Value != 14 {
			t.Errorf("received account with unexpected balance: %v", balance)
		}
	})

	t.Run("TransferBalanceByAPI", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1TransferBalance, utils.AttrTransferBalance{
			Tenant:               "cgrates.org",
			SourceAccountID:      "ACC_SRC",
			SourceBalanceID:      "balance_src",
			DestinationAccountID: "ACC_DEST",
			DestinationBalanceID: "balance_dest",
			Units:                2,
			Cdrlog:               true,
		}, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckBalancesAfterTransferBalanceAPI1", func(t *testing.T) {
		var acnts []*engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
			&utils.AttrGetAccounts{
				Tenant: "cgrates.org",
			}, &acnts); err != nil {
			t.Error(err)
		}
		if len(acnts) != 2 {
			t.Fatal("expecting 2 accounts to be retrieved")
		}
		sort.Slice(acnts, func(i, j int) bool {
			return acnts[i].ID > acnts[j].ID
		})
		if len(acnts[0].BalanceMap) != 1 || len(acnts[0].BalanceMap[utils.MetaMonetary]) != 1 {
			t.Errorf("expected account to have only one balance of type *monetary, received %v", acnts[0])
		}
		balance := acnts[0].BalanceMap[utils.MetaMonetary][0]
		if balance.ID != "balance_src" || balance.Value != 4 {
			t.Errorf("received account with unexpected balance: %v", balance)
		}
		if len(acnts[1].BalanceMap) != 1 || len(acnts[1].BalanceMap[utils.MetaMonetary]) != 1 {
			t.Errorf("expected account to have only one balance of type *monetary, received %v", acnts[1])
		}
		balance = acnts[1].BalanceMap[utils.MetaMonetary][0]
		if balance.ID != "balance_dest" || balance.Value != 16 {
			t.Errorf("received account with unexpected balance: %v", balance)
		}
	})

	t.Run("TransferBalanceByAPINonexistentDestBalance", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1TransferBalance, utils.AttrTransferBalance{
			Tenant:               "cgrates.org",
			SourceAccountID:      "ACC_SRC",
			SourceBalanceID:      "balance_src",
			DestinationAccountID: "ACC_DEST",
			DestinationBalanceID: "nonexistent_balance",
			Units:                3,
			Cdrlog:               true,
		}, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckBalancesAfterTransferBalanceAPI2", func(t *testing.T) {
		var acnts []*engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
			&utils.AttrGetAccounts{
				Tenant: "cgrates.org",
			}, &acnts); err != nil {
			t.Error(err)
		}
		if len(acnts) != 2 {
			t.Fatal("expecting 2 accounts to be retrieved")
		}
		sort.Slice(acnts, func(i, j int) bool {
			return acnts[i].ID > acnts[j].ID
		})
		if len(acnts[0].BalanceMap) != 1 || len(acnts[0].BalanceMap[utils.MetaMonetary]) != 1 {
			t.Errorf("expected account to have only one balance of type *monetary, received %v", acnts[0])
		}
		balance := acnts[0].BalanceMap[utils.MetaMonetary][0]
		if balance.ID != "balance_src" || balance.Value != 1 {
			t.Errorf("received account with unexpected balance: %v", balance)
		}
		if len(acnts[1].BalanceMap) != 1 || len(acnts[1].BalanceMap[utils.MetaMonetary]) != 2 {
			t.Errorf("expected account to have only one balance of type *monetary, received %v", acnts[1])
		}
		balance = acnts[1].BalanceMap[utils.MetaMonetary][1]
		if balance.ID != "nonexistent_balance" || balance.Value != 3 {
			t.Errorf("received account with unexpected balance: %v", balance)
		}
	})

	t.Run("CheckTransferBalanceCDRs", func(t *testing.T) {
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				OrderBy: utils.Cost,
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}

		if len(cdrs) != 3 {
			t.Errorf("expected to receive 3 cdrs: %v", utils.ToJSON(cdrs))
		}
		if cdrs[0].Account != "ACC_SRC" ||
			cdrs[0].Destination != "ACC_DEST" ||
			cdrs[0].RunID != utils.MetaTransferBalance ||
			cdrs[0].Source != utils.CDRLog ||
			cdrs[0].ToR != utils.MetaVoice ||
			!strings.Contains(cdrs[0].ExtraFields[utils.DestinationBalanceSummary], "ID\":\"balance_dest\",\"Type\":\"*monetary\",\"Initial\":14,\"Value\":16,\"Weight\":10,\"Disabled\":false") ||
			!strings.Contains(cdrs[0].ExtraFields[utils.SourceBalanceSummary], "\"ID\":\"balance_src\",\"Type\":\"*monetary\",\"Initial\":6,\"Value\":4,\"Weight\":20,\"Disabled\":false") {
			t.Errorf("unexpected cdr received: %v", utils.ToJSON(cdrs[0]))
		}
		if cdrs[0].Cost != 2 {
			t.Errorf("cost expected to be %v, received %v", 2, cdrs[0].Cost)
		}
		if cdrs[1].Account != "ACC_SRC" ||
			cdrs[1].Destination != "ACC_DEST" ||
			cdrs[1].RunID != utils.MetaTransferBalance ||
			cdrs[1].Source != utils.CDRLog ||
			cdrs[1].ToR != utils.MetaVoice ||
			!strings.Contains(cdrs[1].ExtraFields[utils.DestinationBalanceSummary], "\"ID\":\"nonexistent_balance\",\"Type\":\"*monetary\",\"Initial\":0,\"Value\":3,\"Disabled\":false") ||
			!strings.Contains(cdrs[1].ExtraFields[utils.SourceBalanceSummary], "\"ID\":\"balance_src\",\"Type\":\"*monetary\",\"Initial\":4,\"Value\":1,\"Weight\":20,\"Disabled\":false") {
			t.Errorf("unexpected cdr received: %v", utils.ToJSON(cdrs[1]))
		}
		if cdrs[1].Cost != 3 {
			t.Errorf("cost expected to be %v, received %v", 2, cdrs[1].Cost)
		}
		if cdrs[2].Account != "ACC_SRC" ||
			cdrs[2].Destination != "ACC_DEST" ||
			cdrs[2].RunID != utils.MetaTransferBalance ||
			cdrs[2].Source != utils.CDRLog ||
			cdrs[2].ToR != utils.MetaVoice ||
			!strings.Contains(cdrs[2].ExtraFields[utils.DestinationBalanceSummary], "\"ID\":\"balance_dest\",\"Type\":\"*monetary\",\"Initial\":10,\"Value\":14,\"Weight\":10,\"Disabled\":false") ||
			!strings.Contains(cdrs[2].ExtraFields[utils.SourceBalanceSummary], "\"ID\":\"balance_src\",\"Type\":\"*monetary\",\"Initial\":10,\"Value\":6,\"Weight\":20,\"Disabled\":false") {
			t.Errorf("unexpected cdr received: %v", utils.ToJSON(cdrs[2]))
		}
		if cdrs[2].Cost != 4 {
			t.Errorf("cost expected to be %v, received %v", 2, cdrs[2].Cost)
		}
	})
}
