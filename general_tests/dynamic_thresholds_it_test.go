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
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestDynThdIT(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.InternalDBCfg
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbCfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	// buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "tutinternal"),
		DBCfg:      dbCfg,
		Encoding:   *utils.Encoding,
		// LogBuffer:  buf,
	}
	// t.Cleanup(func() {
	// 	t.Log(buf)
	// })
	client, _ := ng.Run(t)

	t.Run("SetBalance", func(t *testing.T) {
		actPrf := &utils.ActionProfileWithAPIOpts{
			ActionProfile: &utils.ActionProfile{
				Tenant:   utils.CGRateSorg,
				ID:       "1002",
				Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1002": {}}},
				Schedule: utils.MetaASAP,
				Actions: []*utils.APAction{
					{
						ID:   "SET_NEW_BAL",
						Type: utils.MetaSetBalance,
						Diktats: []*utils.APDiktat{
							{
								Path:  "*balance.VOICE.ID",
								Value: "testBalanceIDMonetary",
							},
							{
								Path:  "*balance.MONETARY.Type",
								Value: utils.MetaConcrete,
							},
							{
								Path:  "*balance.MONETARY.Units",
								Value: "1048576",
							},
							{
								Path:  "*balance.MONETARY.Weights",
								Value: "`;2`",
							},
							{
								Path:  "*balance.MONETARY.CostIncrements",
								Value: "`*string:~*req.ToR:*data;1024;0;0.01`",
							},
						},
					},
					{
						ID:   "SET_ADD_BAL",
						Type: utils.MetaAddBalance,
						Diktats: []*utils.APDiktat{
							{
								Path:  "*balance.VOICE.ID",
								Value: "testBalanceID",
							},
							{
								Path:  "*balance.VOICE.Type",
								Value: utils.MetaAbstract,
							},
							{
								Path:  "*balance.VOICE.FilterIDs",
								Value: "`*string:~*req.ToR:*voice`",
							},
							{
								Path:  "*balance.VOICE.Units",
								Value: strconv.FormatInt((time.Hour).Nanoseconds(), 10),
							},
							{
								Path:  "*balance.VOICE.Weights",
								Value: "`;2`",
							},
							{
								Path:  "*balance.VOICE.CostIncrements",
								Value: "`*string:~*req.ToR:*voice;1000000000;0;0.01`",
							},
						},
					},
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetActionProfile, actPrf, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned", reply)
		}
		time.Sleep(100 * time.Millisecond)
		var reply1 string
		if err := client.Call(context.Background(), utils.ActionSv1ExecuteActions, &utils.CGREvent{
			Tenant: utils.CGRateSorg,
			Event: map[string]any{
				"Account": 1002,
			},
		}, &reply1); err != nil {
			t.Error(err)
		} else if reply1 != utils.OK {
			t.Error("Unexpected reply returned", reply1)
		}
		time.Sleep(100 * time.Millisecond)

		var reply2 *[]*utils.Account
		args := &utils.ArgsItemIDs{}
		if err := client.Call(context.Background(), utils.AdminSv1GetAccounts,
			args, &reply2); err != nil {
			t.Error(err)
		}
	})

	t.Run("SetAction", func(t *testing.T) {
		actPrf := &utils.ActionProfileWithAPIOpts{
			ActionProfile: &utils.ActionProfile{
				Tenant: utils.CGRateSorg,
				ID:     "DYNAMIC_THRESHOLD_ACTION",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Targets: map[string]utils.StringSet{utils.MetaThresholds: {"someID": {}}},
				Actions: []*utils.APAction{
					{
						Type: utils.MetaDynamicThreshold,
						Diktats: []*utils.APDiktat{
							{
								Path:  "ExtraParameters",
								Value: "cgrates.org;THD_ACNT_1001;*string:~*req.Account:1002;&10;1;1;1s;false;ACT_LOG_WARNING;true;",
							},
						},
					},
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetActionProfile, actPrf, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned", reply)
		}
		var result *utils.ActionProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetActionProfile, &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: actPrf.Tenant, ID: actPrf.ID}}, &result); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(actPrf.ActionProfile, result) {
			t.Errorf("Expecting : %+v, received: %+v", actPrf.ActionProfile, result)
		}
	})

	t.Run("SetThresholdProfile", func(t *testing.T) {
		thPrf1 := &engine.ThresholdProfileWithAPIOpts{
			ThresholdProfile: &engine.ThresholdProfile{
				Tenant:           utils.CGRateSorg,
				ID:               "THD_ACNT_1002",
				FilterIDs:        []string{"*string:~*opts.*acntProfileIDs:1002"},
				MaxHits:          1,
				ActionProfileIDs: []string{"DYNAMIC_THRESHOLD_ACTION"},
				Async:            true,
			},
		}

		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
			thPrf1, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned:", reply)
		}

	})

	t.Run("GetThresholdProfile", func(t *testing.T) {
		var rplyTh engine.Threshold
		var rplyThPrf engine.ThresholdProfile
		expTh := engine.Threshold{
			Tenant: utils.CGRateSorg,
			ID:     "THD_ACNT_1002",
		}
		expThPrf := engine.ThresholdProfile{
			Tenant:           utils.CGRateSorg,
			ID:               "THD_ACNT_1002",
			FilterIDs:        []string{"*string:~*opts.*acntProfileIDs:1002"},
			MaxHits:          1,
			ActionProfileIDs: []string{"DYNAMIC_THRESHOLD_ACTION"},
			Async:            true,
		}

		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: utils.CGRateSorg,
					ID:     "THD_ACNT_1002",
				},
			}, &rplyTh); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(rplyTh, expTh) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expTh), utils.ToJSON(rplyTh))
		}
		if err := client.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
			utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "THD_ACNT_1002",
			}, &rplyThPrf); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(rplyThPrf, expThPrf) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expThPrf), utils.ToJSON(rplyThPrf))
		}
	})
	t.Run("ThresholdProcessEvent", func(t *testing.T) {
		time.Sleep(50 * time.Millisecond)
		tEv := &utils.CGREvent{
			Tenant: utils.CGRateSorg,
			ID:     "event1",
			Event: map[string]any{
				utils.AccountField: "1002",
			},
			APIOpts: map[string]any{
				utils.MetaUsage:              5 * time.Second,
				utils.OptsAccountsProfileIDs: "1002",
			},
		}
		var ids []string
		if err := client.Call(context.Background(), utils.ThresholdSv1ProcessEvent, tEv, &ids); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(ids, []string{"THD_ACNT_1002"}) {
			t.Error("Unexpected reply returned", ids)
		}
		time.Sleep(100 * time.Millisecond) //wait for async
	})
	t.Run("GetThresholdProfile", func(t *testing.T) {
		var thrsholds []*engine.ThresholdProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetThresholdProfiles,
			&utils.ArgsItemIDs{
				Tenant: utils.CGRateSorg,
			}, &thrsholds); err != nil {
			t.Errorf("AdminSv1GetThresholdProfiles failed unexpectedly: %v", err)
		}
		if len(thrsholds) != 2 {
			t.Fatalf("AdminSv1GetThresholdProfiles len(thrsholds)=%v, want 2", len(thrsholds))
		}
		sort.Slice(thrsholds, func(i, j int) bool {
			return thrsholds[i].ID > thrsholds[j].ID
		})
		exp := []*engine.ThresholdProfile{
			{
				Tenant:           utils.CGRateSorg,
				ID:               "THD_ACNT_1002",
				FilterIDs:        []string{"*string:~*opts.*acntProfileIDs:1002"},
				MaxHits:          1,
				MinHits:          0,
				MinSleep:         0,
				Blocker:          false,
				Weights:          nil,
				ActionProfileIDs: []string{"DYNAMIC_THRESHOLD_ACTION"},
				Async:            true,
			},
			{
				Tenant:    utils.CGRateSorg,
				ID:        "THD_ACNT_1001",
				FilterIDs: []string{"*string:~*req.Account:1002"},
				MaxHits:   1,
				MinHits:   1,
				MinSleep:  time.Second,
				Blocker:   false,
				Weights: utils.DynamicWeights{
					&utils.DynamicWeight{
						FilterIDs: nil,
						Weight:    10,
					},
				},
				ActionProfileIDs: []string{"ACT_LOG_WARNING"},
				Async:            true,
			},
		}
		if !reflect.DeepEqual(thrsholds, exp) {
			t.Errorf("Expected <%v> \n received <%v>", utils.ToJSON(exp), utils.ToJSON(thrsholds))
		}
	})

}
