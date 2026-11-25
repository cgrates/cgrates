//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"cmp"
	"fmt"
	"slices"
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
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	config := `{
"apiers": {
	"enabled": true,
	"attributes_conns": ["*internal"],
},
"attributes": {
	"enabled": true
},

}`

	tpFiles := map[string]string{
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP_DESC,*topup,,,,*monetary,,,,,+720h,,100,*constant:*req.BalanceWeight:*none{*time_desc},,,10
ACT_TOPUP_ASC,*topup,,,,*monetary,,,,,+720h,,100,*constant:*req.BalanceWeight:*none{*time_asc},,,10
ACT_DATA_BONUS,*topup,,,,*data,,,,,+720h,,1024,*sum:*req.BalanceWeight:*none{*time_desc}&1000000,,,20
ACT_DATA_BONUS,*topup,,,,*voice,,,,,+720h,,60m,*difference:*req.BalanceWeight:*none{*time_desc}&1000000,,,10
ACT_TOPUP_ID,*topup,,,BAL_DESC,*monetary,,,,,+720h,,100,*constant:*req.BalanceWeight:*none{*time_desc},,,10
ACT_TOPUP_BalVal,*topup,,,,*monetary,,,,,+720h,,500,*sum:*req.BalanceWeight:~*req.BalanceValue&10,,,10
ACT_TOPUP_INVALID,*topup,,,,*monetary,,,,,+720h,,100,*summ:*req.BalanceWeight:1&2,,,10
ACT_TOPUP_NO_WEIGHT,*topup,,,,*monetary,,,,,+720h,,100,*sum:*req.Weight:1&2,,,10
ACT_DATA_BONUS1,*topup,,,Bal1,*data,,,,,+720h,,1024,*sum:*req.BalanceWeight:*none{*time_desc}&1000000000,,,10
ACT_DATA_BONUS2,*topup,,,Bal2,*data,,,,,+720h,,1024,*difference:*req.BalanceWeight:*none{*time_desc}&1000000000,,,10
ACT_TOPUP_PRF,*topup,,,,*monetary,,,,,+720h,,100,ATTR_BALANCE_WEIGHT,,,10
ACT_TOPUP_PRF_TIME,*topup,,,,*monetary,,,,,+720h,,100,ATTR_TIME_WEIGHT,,,10
ACT_TOPUP_EMPTY_VAL,*topup,,,,*monetary,,,,,+720h,,100,*constant:*req.BalanceWeight:{*time_desc},,,10`,
		utils.AttributesCsv: `#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_BALANCE_WEIGHT,,*string:~*req.ActionID:ACT_TOPUP_PRF,,,*req.BalanceWeight,*sum,~*req.BalanceValue;100,false,10
cgrates.org,ATTR_TIME_WEIGHT,,*string:~*req.ActionID:ACT_TOPUP_PRF_TIME,,,*req.BalanceWeight,*sum,*none{*time_desc};1000,false,10`,
	}

	ng := engine.TestEngine{
		ConfigJSON: config,
		DBCfg:      dbCfg,
		TpFiles:    tpFiles,
	}

	client, _ := ng.Run(t)

	var reply string
	for i := 1; i <= 11; i++ {
		if err := client.Call(context.Background(), utils.APIerSv1SetAccount,
			&utils.AttrSetAccount{Tenant: "cgrates.org", Account: fmt.Sprintf("ACNT_%d", i)}, &reply); err != nil {
			t.Fatal(err)
		}
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
		if acts[0].BalanceWeight != "*constant:*req.BalanceWeight:*none{*time_desc}" {
			t.Errorf("expected BalanceWeight %q, received: %s",
				"*constant:*req.BalanceWeight:*none{*time_desc}", acts[0].BalanceWeight)
		}

		if err := client.Call(context.Background(), utils.APIerSv1GetActions,
			"ACT_TOPUP_ASC", &acts); err != nil {
			t.Fatal(err)
		}
		if len(acts) != 1 {
			t.Fatal("expected actions to be set")
		}
		if acts[0].BalanceWeight != "*constant:*req.BalanceWeight:*none{*time_asc}" {
			t.Errorf("expected BalanceWeight %q, received: %s",
				"*constant:*req.BalanceWeight:*none{*time_asc}", acts[0].BalanceWeight)
		}
		if err := client.Call(context.Background(), utils.APIerSv1GetActions,
			"ACT_DATA_BONUS", &acts); err != nil {
			t.Fatal(err)
		}
		if len(acts) != 2 {
			t.Fatal("expected 2 actions to be set")
		}
		slices.SortFunc(acts, func(a, b *utils.TPAction) int {
			return cmp.Compare(a.Weight, b.Weight)
		})
		if acts[0].BalanceWeight != "*difference:*req.BalanceWeight:*none{*time_desc}&1000000" {
			t.Error("balance weight didn't set correctly")
		}
		if acts[1].BalanceWeight != "*sum:*req.BalanceWeight:*none{*time_desc}&1000000" {
			t.Error("balance weight didn't set correctly")
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
		execAct(t, "ACT_TOPUP_ASC", "ACNT_2")
		execAct(t, "ACT_DATA_BONUS", "ACNT_3")
		time.Sleep(time.Second)
		execAct(t, "ACT_TOPUP_DESC", "ACNT_1")
		execAct(t, "ACT_TOPUP_ASC", "ACNT_2")
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
			t.Errorf("expected older balance from topup to have bigger weight for *time_desc  %v, received: %v",
				balMon[0].Weight, balMon[1].Weight)
		}

		var acc2 engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_2"}, &acc2); err != nil {
			t.Fatal(err)
		}
		balMon2 := acc2.BalanceMap[utils.MetaMonetary]
		if len(balMon2) != 2 {
			t.Fatal("expected 2 *monetary balances")
		}
		if balMon2[0].Weight >= balMon2[1].Weight {
			t.Errorf("expected older balance from topup to have smaller weight for *time_asc  %v, received: %v",
				balMon2[0].Weight, balMon2[1].Weight)
		}

		var acc3 engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_3"}, &acc3); err != nil {
			t.Fatal(err)
		}
		balVoice := acc3.BalanceMap[utils.MetaVoice]
		if len(balVoice) != 1 {
			t.Fatal("expected 1 *voice balances")
		}
		balData := acc3.BalanceMap[utils.MetaData]
		if len(balData) != 1 {
			t.Fatal("expected 1 *data balances")
		}
		if balData[0].Weight <= balVoice[0].Weight {
			t.Errorf("expected data balance to have higher weight %v, received: %v",
				balData[0].Weight, balVoice[0].Weight)
		}
	})

	t.Run("DynamicWeightFilterWithBalanceID", func(t *testing.T) {
		execAct(t, "ACT_TOPUP_ID", "ACNT_4")
		var acc engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_4"}, &acc); err != nil {
			t.Fatal(err)
		}
		balMon := acc.BalanceMap[utils.MetaMonetary]
		if len(balMon) != 1 {
			t.Fatal("expected 1 *monetary balance")
		}
		firstWeight := balMon[0].Weight
		time.Sleep(time.Second)
		execAct(t, "ACT_TOPUP_ID", "ACNT_4")

		var accAfter engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_4"}, &accAfter); err != nil {
			t.Fatal(err)
		}
		balMonAfter := accAfter.BalanceMap[utils.MetaMonetary]
		if len(balMonAfter) != 1 {
			t.Fatal("expected topup with same BalanceID to update the existing balance")
		}
		if balMonAfter[0].Value != 200 {
			t.Errorf("expected balance value 200, received: %v", balMonAfter[0].Value)
		}
		if balMonAfter[0].Weight != firstWeight {
			t.Errorf("expected weight from first topup %v, received: %v",
				firstWeight, balMonAfter[0].Weight)
		}
	})

	t.Run("WeightFromBalanceValue", func(t *testing.T) {
		execAct(t, "ACT_TOPUP_BalVal", "ACNT_5")
		var acc engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_5"}, &acc); err != nil {
			t.Fatal(err)
		}
		balMon := acc.BalanceMap[utils.MetaMonetary]
		if len(balMon) != 1 {
			t.Fatal("expected 1 *monetary balance")
		}
		if balMon[0].Weight != 510 {
			t.Errorf("expected weight %v, received: %v",
				510, balMon[0].Weight)
		}
	})

	t.Run("InvalidAttributeRule", func(t *testing.T) {
		execAct(t, "ACT_TOPUP_INVALID", "ACNT_6")
		var acc engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_6"}, &acc); err != nil {
			t.Fatal(err)
		}
		balMon := acc.BalanceMap[utils.MetaMonetary]
		if len(balMon) != 1 {
			t.Fatal("expected a balance when invalid attribute rule")
		}
		if balMon[0].Weight != 0 {
			t.Errorf("expected weight 0, received weight %v",
				balMon[0].Weight)
		}
	})

	t.Run("RuleWithoutWeight", func(t *testing.T) {
		execAct(t, "ACT_TOPUP_NO_WEIGHT", "ACNT_7")
		var acc engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_7"}, &acc); err != nil {
			t.Fatal(err)
		}
		balMon := acc.BalanceMap[utils.MetaMonetary]
		if len(balMon) != 1 {
			t.Fatal("expected balance when no *req.BalanceWeight")
		}
		if balMon[0].Weight != 0 {
			t.Errorf("expected weight 0, received weight %v",
				balMon[0].Weight)
		}
	})

	t.Run("BonusBalances", func(t *testing.T) {
		execAct(t, "ACT_DATA_BONUS1", "ACNT_8")
		execAct(t, "ACT_DATA_BONUS2", "ACNT_8")
		var acc engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_8"}, &acc); err != nil {
			t.Fatal(err)
		}
		balData := acc.BalanceMap[utils.MetaData]
		if len(balData) != 2 {
			t.Fatal("expected 2 *data balances")
		}

		if balData[0].Weight <= balData[1].Weight {
			t.Errorf("expected balance using *sum larger than balance with diff %v, received: %v",
				balData[0].Weight, balData[1].Weight)
		}

	})

	t.Run("WeightFromAttributeProfile", func(t *testing.T) {
		execAct(t, "ACT_TOPUP_PRF", "ACNT_9")
		var acc engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_9"}, &acc); err != nil {
			t.Fatal(err)
		}
		balMon := acc.BalanceMap[utils.MetaMonetary]
		if len(balMon) != 1 {
			t.Fatal("expected 1 *monetary balance")
		}
		if balMon[0].Weight != 200 {
			t.Errorf("expected weight %v, received: %v",
				5100, balMon[0].Weight)
		}
	})

	t.Run("TimeWeightFromAttributeProfile", func(t *testing.T) {
		weightNow := float64(time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC).Unix()-time.Now().Unix()) + 1000
		execAct(t, "ACT_TOPUP_PRF_TIME", "ACNT_10")
		var acc engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_10"}, &acc); err != nil {
			t.Fatal(err)
		}
		balMon := acc.BalanceMap[utils.MetaMonetary]
		if len(balMon) != 1 {
			t.Fatal("expected 1 *monetary balance")
		}
		if balMon[0].Weight > weightNow {
			t.Errorf("expected weight close to %v, received: %v",
				weightNow, balMon[0].Weight)
		}
	})

	t.Run("TimeConverterWithoutNoneVal", func(t *testing.T) {
		execAct(t, "ACT_TOPUP_EMPTY_VAL", "ACNT_11")
		var acc engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "ACNT_11"}, &acc); err != nil {
			t.Fatal(err)
		}
		balMon := acc.BalanceMap[utils.MetaMonetary]
		if len(balMon) != 1 {
			t.Fatal("expected 1 *monetary balance")
		}
		if balMon[0].Weight != 0 {
			t.Errorf("expected weight 0, received weight %v",
				balMon[0].Weight)
		}
	})

}
