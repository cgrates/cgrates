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
	"testing"

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
cgrates.org,ACC_SRC,PACKAGE_ACC_SRC,,,
cgrates.org,ACC_DEST,PACKAGE_ACC_DEST,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_ACC_SRC,ACT_TOPUP_SRC,*asap,10
PACKAGE_ACC_DEST,ACT_TOPUP_DEST,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP_SRC,*topup_reset,,,balance_src,*monetary,,*any,,,*unlimited,,10,20,false,false,20
ACT_TOPUP_DEST,*topup_reset,,,balance_dest,*monetary,,*any,,,*unlimited,,10,10,false,false,10
ACT_TRANSFER,*transfer_balance,"{""DestinationAccountID"":""cgrates.org:ACC_DEST"",""DestinationBalanceID"":""balance_dest""}",,balance_src,*monetary,,,,,*unlimited,,4,,,,`,
	}

	testEnv := TestEnvironment{
		Name:       "TestTransferBalance",
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _, shutdown, err := testEnv.Setup(t, *waitRater)
	if err != nil {
		t.Fatal(err)
	}

	defer shutdown()

	t.Run("CheckInitialBalances", func(t *testing.T) {
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
		}, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckBalancesAfterTransferBalanceAPI", func(t *testing.T) {
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
}
