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

func TestATExportAndTopup(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7,
	"reply_timeout": "30s"
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*internal"]
},

"rals": {
	"enabled": true
},

"schedulers": {
	"enabled": true,
	"cdrs_conns": ["*internal"]
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
	"ees_conns": ["*internal"]
},

"ees": {
	"enabled": true,
	"exporters": [
		{
			"id": "test_exporter",
			"type": "*virt",
			"attempts": 1,
			"synchronous": true,
			"fields":[
				{"tag": "Balance", "filters": ["*empty:~*uch.ExportCount:"], "path": "*uch.ExportCount", "type": "*sum", "value": "0;1", "blocker": true},
				{"tag": "Balance", "path": "*uch.ExportCount", "type": "*sum", "value": "~*uch.ExportCount;1"}
			]
		}
	]
}

}`

	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,TRIGGER_1001,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP_INITIAL,*asap,`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP_INITIAL,*topup_reset,,,main,*data,,,,,+150ms,,10240,0,,,10
ACT_TOPUP_INITIAL,*topup,,,main,*data,,,,,,,10240,0,,,0
ACT_TOPUP_10GB,*topup,,,main,*data,,,,,,,10240,0,,,
ACT_USAGE_1GB,*export,test_exporter,,,,,,,,,,,,,,30
ACT_USAGE_1GB,*topup,,,main,*data,,,,,,,1024,,,,20
ACT_USAGE_1GB,*cdrlog,"{""ToR"":""*data""}",,,,,,,,,,,,,,10
ACT_USAGE_1GB,*reset_triggers,,,,,,,,,,,,,,,0
ACT_RESET_TRIGGERS,*reset_triggers,,,,,,,,,,,,,,,`,
		utils.ActionTriggersCsv: `#Tag[0],UniqueId[1],ThresholdType[2],ThresholdValue[3],Recurrent[4],MinSleep[5],ExpiryTime[6],ActivationTime[7],BalanceTag[8],BalanceType[9],BalanceCategories[10],BalanceDestinationIds[11],BalanceRatingSubject[12],BalanceSharedGroup[13],BalanceExpiryTime[14],BalanceTimingIds[15],BalanceWeight[16],BalanceBlocker[17],BalanceDisabled[18],ActionsId[19],Weight[20]
TRIGGER_1001,,*min_balance,0,false,0,,,main,*data,,,,,,,,,,ACT_USAGE_1GB,
TRIGGER_1001,,*balance_expired,,true,0,,,main,*data,,,,,,,,,,ACT_TOPUP_INITIAL,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)

	// helper functions
	i := 0
	balanceID := "main"

	// processCDRs processes 'count' CDRs, each debiting 'amount'.
	processCDRs := func(t *testing.T, count int, amount int) {
		t.Helper()
		var reply string
		for range count {
			i++
			if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
				&engine.ArgV1ProcessEvent{
					Flags: []string{utils.MetaRALs},
					CGREvent: utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     fmt.Sprintf("event%d", i),
						Event: map[string]any{
							utils.ToR:          utils.MetaData,
							utils.OriginID:     fmt.Sprintf("processCDR%d", i),
							utils.RequestType:  utils.MetaPostpaid,
							utils.AccountField: "1001",
							utils.Usage:        amount,
						},
					},
				}, &reply); err != nil {
				t.Errorf("CDRsV1ProcessEvent(%d) failed unexpectedly: %v", i, err)
			}
		}
	}

	checkAccountAndCDRs := func(t *testing.T, wantBalVal float64, triggerCount int) time.Time {
		t.Helper()
		var acnts []*engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
			&utils.AttrGetAccounts{
				Tenant: "cgrates.org",
			}, &acnts); err != nil {
			t.Errorf("APIerSv2GetAccounts failed unexpectedly: %v", err)
		}
		if len(acnts) != 1 {
			t.Fatalf("APIerSv2GetAccounts len(acnts)=%v, want 1", len(acnts))
		}
		got, _ := acnts[0].FindBalanceByID(balanceID)
		if got == nil {
			t.Errorf("acnts[0].FindBalanceByID(%q) could not find balance", balanceID)
		} else if got.Value != wantBalVal {
			t.Errorf("acnts[0].FindBalanceByID(%q) returned balance with value %v, want %v", balanceID, got.Value, wantBalVal)
		}
		expiryTime := got.ExpirationDate

		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				RunIDs: []string{utils.MetaTopUp},
			}}, &cdrs); err != nil && triggerCount != 0 {
			t.Errorf("CDRsV1GetCDRs failed unexpectedly: %v", err)
		} else if noCDRs := len(cdrs); noCDRs != triggerCount {
			t.Errorf("CDRsV1GetCDRs *topup cdrs count=%d, want %d", noCDRs, triggerCount)
		}

		var count any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "ExportCount",
			},
		}, &count); err != nil && triggerCount != 0 {
			t.Errorf("CacheSv1GetItem failed unexpectedly: %v", err)
		} else {
			if count == nil {
				count = 0.0
			}
			if count != float64(triggerCount) {
				t.Errorf("CacheSv1GetItem *uch.ExportCount=%v, want %v", count, triggerCount)
			}
		}
		return expiryTime
	}

	executeAction := func(t *testing.T, id string) {
		var reply string
		attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1001", ActionsId: id}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
			t.Errorf("APIerSv1ExecuteAction failed unexpectedly: %v", err)
		}
	}

	t.Run("ProcessCDRs", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond)
		expiryTime := checkAccountAndCDRs(t, 20480, 0) // 20GB balance
		processCDRs(t, 1, 200)                         // -200MB debit, triggered *topup+*export 0 times
		checkAccountAndCDRs(t, 20280, 0)               // 20GB-200MB actual_balance
		processCDRs(t, 1, 20280)                       // (20GB-200MB)-(20GB-200MB) debit, triggered *topup+*export 1 time
		checkAccountAndCDRs(t, 1024, 1)                // 1GB actual_balance
		executeAction(t, "ACT_TOPUP_10GB")             // +10GB topup
		checkAccountAndCDRs(t, 11264, 1)               // 11GB actual_balance
		processCDRs(t, 1, 16384)                       // -16GB debit, triggered *topup+*export 6 times
		checkAccountAndCDRs(t, 1024, 7)                // 1GB actual_balance
		executeAction(t, "ACT_RESET_TRIGGERS")         // execute action triggers, expect nothing to happen since the balance is not expired yet
		checkAccountAndCDRs(t, 1024, 7)                // 1GB actual_balance (unchanged)
		time.Sleep(expiryTime.Sub(time.Now()))         // wait for the balance to expire
		executeAction(t, "ACT_RESET_TRIGGERS")         // execute action triggers, balance has expired and will be reset
		checkAccountAndCDRs(t, 20480, 7)               // 20GB balance
	})
}
