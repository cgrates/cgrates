//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestBalanceFilterWeightRule(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.InternalDBCfg
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbCfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbCfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	config := `{
"apiers": {
	"enabled": true
},
"data_db": {								
	"db_type": "redis",						
	"db_port": 6379, 						
	"db_name": "10", 						
},
"stor_db": {
	"db_password": "CGRateS.org",
},
}`

	tpFiles := map[string]string{
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP_DESC,*topup,,,,*monetary,,,,,+720h,,100,*time_desc,,,10
ACT_TOPUP,*topup,,,bal1,*monetary,,,,,+720h,,100,*time_desc,,,10`,
	}

	ng := engine.TestEngine{
		ConfigJSON: config,
		DBCfg:      dbCfg,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)

	var reply string
	if err := client.Call(context.Background(), utils.APIerSv1SetAccount,
		&utils.AttrSetAccount{Tenant: "cgrates.org", Account: "ACNT_1"}, &reply); err != nil {
		t.Fatal(err)
	}
	if err := client.Call(context.Background(), utils.APIerSv1SetAccount,
		&utils.AttrSetAccount{Tenant: "cgrates.org", Account: "ACNT_2"}, &reply); err != nil {
		t.Fatal(err)
	}
	if err := client.Call(context.Background(), utils.APIerSv2SetActions,
		&utils.AttrSetActions{
			ActionsId: "ACT_TOPUP_ASC",
			Actions: []*utils.TPAction{
				{
					Identifier:    utils.MetaTopUp,
					BalanceType:   utils.MetaVoice,
					ExpiryTime:    "+720h",
					Units:         "60s",
					BalanceWeight: utils.MetaTimeAsc,
					Weight:        10,
				},
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	t.Run("GetActions", func(t *testing.T) {
		var acts []*utils.TPAction
		if err := client.Call(context.Background(), utils.APIerSv1GetActions,
			"ACT_TOPUP_DESC", &acts); err != nil {
			t.Fatal(err)
		}
		if len(acts) != 1 {
			t.Fatal("expected actions to be set")
		}
		if acts[0].BalanceWeight != utils.MetaTimeDesc {
			t.Errorf("expected BalanceWeight %q, received: %s",
				utils.MetaTimeDesc, acts[0].BalanceWeight)
		}

		if err := client.Call(context.Background(), utils.APIerSv1GetActions,
			"ACT_TOPUP_ASC", &acts); err != nil {
			t.Fatal(err)
		}
		if len(acts) != 1 {
			t.Fatal("expected actions to be set")
		}
		if acts[0].BalanceWeight != utils.MetaTimeAsc {
			t.Errorf("expected BalanceWeight %q, received: %s",
				utils.MetaTimeAsc, acts[0].BalanceWeight)
		}
		if err := client.Call(context.Background(), utils.APIerSv1GetActions,
			"ACT_TOPUP", &acts); err != nil {
			t.Fatal(err)
		}
		if len(acts) != 1 {
			t.Fatal("expected actions to be set")
		}
		if acts[0].BalanceWeight != utils.MetaTimeDesc {
			t.Errorf("expected BalanceWeight %q, received: %s",
				utils.MetaTimeAsc, acts[0].BalanceWeight)
		}
	})

	execAct := func(t *testing.T, actionID, accIDs string) {
		t.Helper()
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{
				Tenant:    "cgrates.org",
				Account:   accIDs,
				ActionsId: actionID,
			}, &reply); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("ExecuteActions", func(t *testing.T) {
		execAct(t, "ACT_TOPUP_DESC", "ACNT_1")
		execAct(t, "ACT_TOPUP_ASC", "ACNT_1")
		execAct(t, "ACT_TOPUP", "ACNT_2")

		var acc2 engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_2"}, &acc2); err != nil {
			t.Fatal(err)
		}
		balMon2 := acc2.BalanceMap[utils.MetaMonetary]
		if len(balMon2) != 1 {
			t.Fatal("expected 1 *monetary balance")
		}

		time.Sleep(time.Second)
		execAct(t, "ACT_TOPUP_DESC", "ACNT_1")
		execAct(t, "ACT_TOPUP_ASC", "ACNT_1")
		execAct(t, "ACT_TOPUP", "ACNT_2")
		var acc engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_1"}, &acc); err != nil {
			t.Fatal(err)
		}
		balMon := acc.BalanceMap[utils.MetaMonetary]
		if len(balMon) != 2 {
			t.Fatal("expected 2 *monetary balances")
		}
		if balMon[0].Weight <= balMon[1].Weight {
			t.Errorf("expected older balance from topup to have higher weight for *time_desc  %v, received: %v",
				balMon[0].Weight, balMon[1].Weight)
		}
		balVoice := acc.BalanceMap[utils.MetaVoice]
		if len(balVoice) != 2 {
			t.Fatal("expected 2 *voice balances")
		}
		if balVoice[0].Weight >= balVoice[1].Weight {
			t.Errorf("expected older balance from topup to have smaller weight for *time_asc  %v, received: %v",
				balVoice[0].Weight, balVoice[1].Weight)
		}
		var acc2new engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_2"}, &acc2new); err != nil {
			t.Fatal(err)
		}
		balMon2new := acc2new.BalanceMap[utils.MetaMonetary]
		if len(balMon2new) != 1 {
			t.Fatal("expected 1 *monetary balance")
		}
		if balMon2new[0].Value != 200 {
			t.Errorf("expected value 200, received: %v", balMon2new[0].Value)
		}
		if balMon2[0].Weight != balMon2new[0].Weight {
			t.Errorf("expected topup matched on BalanceID to keep weight %v, received: %v",
				balMon2[0].Weight, balMon2new[0].Weight)
		}
	})
}
