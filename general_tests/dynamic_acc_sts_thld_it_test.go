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
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestDynamicAccountWithStatsAndThreshold(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "dynamic_account_threshold"),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "testit"),
	}
	client, _ := ng.Run(t)

	t.Run("SetInitiativeThresholdProfile", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		thdPrf := &engine.ThresholdProfileWithAPIOpts{
			ThresholdProfile: &engine.ThresholdProfile{
				Tenant:    "cgrates.org",
				ID:        "THD_DYNAMIC_STATS_AND_THRESHOLD_INIT",
				FilterIDs: []string{"*exists:~*opts.*accountID:"},
				ActionIDs: []string{"ACT_DYN_THRESHOLD_AND_STATS_CREATION"},
				MinHits:   1,
				MaxHits:   -1,
				Weight:    1, // keep in mind weight should be lower than the dynamicaly created thresholds so that we dont retrigger this threshold for already created accounts
				Async:     true,
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetThresholdProfile, thdPrf, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned", reply)
		}
	})

	t.Run("SetAccountBlockAction", func(t *testing.T) {
		attrs1 := &utils.AttrSetActions{
			ActionsId: "ACT_BLOCK_ACC",
			Actions: []*utils.TPAction{
				{
					Identifier: utils.MetaDisableAccount,
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv2SetActions, &attrs1, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Unexpected reply returned: %s", reply)
		}
	})

	t.Run("SetDynamicThresholdAndStatsAction", func(t *testing.T) {
		attrs1 := &utils.AttrSetActions{
			ActionsId: "ACT_DYN_THRESHOLD_AND_STATS_CREATION",
			Actions: []*utils.TPAction{
				{
					// dynamic threshold for already created dynamic accounts, needed so we can ignore matching thresholds the events (which dont come from stats) where an account has already been dynamicaly created by  the initial threhold THD_DYNAMIC_STATS_AND_THRESHOLD_INIT. The threshold itself is only used for blocking
					Identifier: utils.MetaDynamicThreshold,
					// get tenant and accountID from event, threshold triggers when the event account matches the already dynamicaly created account. If it matches the filter, it will block other thresholds from matching with the event. Make sure dynamic thresholds weight is higher than the initiative threshold THD_DYNAMIC_STATS_AND_THRESHOLD_INIT
					ExtraParameters: "*tenant;THD_BLOCKER_ACNT_<~*req.Account>;*string:~*opts.*accountID:<~*req.Account>;*now;-1;1;;true;3;;true;",
				},
				{
					Identifier: utils.MetaDynamicThreshold,
					// get tenant and accountID from event, threshold triggers when sum of statID hits 100, after triggers the action, the threshold will be disabled for 24 hours, make sure dynamic thresholds weight is higher than the initiative threshold THD_DYNAMIC_STATS_AND_THRESHOLD_INIT and blocker threshold THD_BLOCKER_ACNT_<~*req.Account>
					ExtraParameters: "*tenant;THD_ACNT_<~*req.Account>;*string:~*req.StatID:Stat_<~*req.Account>&*string:~*req.*sum#1:100;*now;-1;1;24h;true;4;ACT_BLOCK_ACC;true;",
				},
				{
					Identifier: utils.MetaDynamicStats,
					// get tenant and accountID from event, stat triggers when an event contains account with dynamicaly created accountID and also has a *accountID field in APIOpts, each encounter that matches the filters will raise the *sum number and call the thresholdIDs specified. when the ttl is reached, *sum will go down also
					ExtraParameters: "*tenant;Stat_<~*req.Account>;*string:~*req.Account:<~*req.Account>&*exists:~*opts.*accountID:;*now;;5s;;*sum#1;;true;false;10;THD_ACNT_<~*req.Account>&THD_BLOCKER_ACNT_<~*req.Account>;",
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv2SetActions, &attrs1, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Unexpected reply returned: %s", reply)
		}
	})

	t.Run("SetActionPlanOfDynaPrepaidAccounts", func(t *testing.T) {
		var reply string
		atms1 := &v1.AttrSetActionPlan{
			Id: "DYNA_ACC",
			ActionPlan: []*v1.AttrActionPlan{
				{
					ActionsId: "TOPUP_RST_DATA_100",
					Time:      "00:00:00",
					TimingID:  utils.MetaMonthlyEstimated,
					Weight:    20,
				},
			},
			Overwrite: false,
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetActionPlan, &atms1, &reply); err != nil {
			t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
		} else if reply != utils.OK {
			t.Errorf("Unexpected reply returned: %s", reply)
		}
	})

	t.Run("Make100AuthCalls", func(t *testing.T) {
		args1 := &sessions.V1AuthorizeArgs{
			GetMaxUsage:       true,
			ProcessThresholds: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.OriginID:     "sessDynaprepaid",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "sessDynaprepaid",
					utils.ToR:          utils.MetaData,
					utils.RequestType:  utils.MetaDynaprepaid,
					utils.AccountField: "CreatedAccount",
					utils.Destination:  "+1234567",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        1024,
				},
				APIOpts: map[string]any{"*accountID": "CreatedAccount"}, // account has to be in apiopts for stats to push it to threhsoldsv1ProcessEvent so that it knows which account to disable
			},
		}
		var rply1 sessions.V1AuthorizeReply
		if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
			args1, &rply1); err != nil {
			t.Error(err)
			return
		} else if *rply1.MaxUsage != 1024*time.Nanosecond {
			t.Errorf("Expected <%+v>, received <%+v>", 1024*time.Nanosecond, *rply1.MaxUsage)
		}
		for i := range 100 {
			strI := strconv.Itoa(i)
			args1 := &sessions.V1AuthorizeArgs{
				GetMaxUsage:  true,
				ProcessStats: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					Event: map[string]any{
						utils.OriginID:     "sessPrepaid" + strI,
						utils.OriginHost:   "192.168.1.1",
						utils.Source:       "sessPrepaid",
						utils.ToR:          utils.MetaData,
						utils.RequestType:  utils.MetaPrepaid,
						utils.AccountField: "CreatedAccount",
						utils.Destination:  "+1234567",
						utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
						utils.Usage:        1024,
					},
					APIOpts: map[string]any{"*accountID": "CreatedAccount"}, // account has to be in apiopts for stats to push it to threhsoldsv1ProcessEvent so that it knows which account to disable
				},
			}
			var rply1 sessions.V1AuthorizeReply
			if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
				args1, &rply1); err != nil {
				t.Error(err)
				return
			} else if *rply1.MaxUsage != 1024*time.Nanosecond {
				t.Errorf("Expected <%+v>, received <%+v>", 1024*time.Nanosecond, *rply1.MaxUsage)
			}
		}
	})

	t.Run("CheckAccountBlocked", func(t *testing.T) {
		// wait for account to be disabled async
		time.Sleep(10 * time.Millisecond)
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "CreatedAccount"}, &acnt); err != nil {
			t.Error(err)
		}
		expAcc := &engine.Account{
			ID: "cgrates.org:CreatedAccount",
			BalanceMap: map[string]engine.Balances{
				utils.MetaData: {
					&engine.Balance{
						Uuid:           acnt.BalanceMap[utils.MetaData][0].Uuid,
						ID:             "",
						Categories:     utils.StringMap{},
						SharedGroups:   utils.StringMap{},
						TimingIDs:      utils.StringMap{},
						Value:          4096,
						ExpirationDate: acnt.BalanceMap[utils.MetaData][0].ExpirationDate,
						Weight:         10,
						DestinationIDs: utils.StringMap{},
					},
				},
			},
			UpdateTime: acnt.UpdateTime,
			Disabled:   true,
		}
		if !reflect.DeepEqual(utils.ToJSON(expAcc), utils.ToJSON(acnt)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(expAcc), utils.ToJSON(acnt))
		}
	})

}
