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
		dbCfg = engine.PostgresDBCfg
	default:
		t.Fatal("Unknown Database type")
	}

	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "tutinternal"),
		DBCfg:      dbCfg,
		Encoding:   *utils.Encoding,
		// LogBuffer:  &bytes.Buffer{},
	}
	// t.Cleanup(func() {
	// 	t.Log(ng.LogBuffer)
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
								ID:        "SetVoiceID",
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Opts: map[string]any{
									"*balancePath":  "*balance.VOICE.ID",
									"*balanceValue": "testBalanceIDMonetary",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 14,
									},
								},
							},
							{
								ID: "SetMonetaryType",
								Opts: map[string]any{
									"*balancePath":  "*balance.MONETARY.Type",
									"*balanceValue": utils.MetaConcrete,
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 13,
									},
								},
							},
							{
								ID: "SetMonetaryUnits",
								Opts: map[string]any{
									"*balancePath":  "*balance.MONETARY.Units",
									"*balanceValue": "1048576",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 12,
									},
								},
							},
							{
								ID: "SetMonetaryWeights",
								Opts: map[string]any{
									"*balancePath":  "*balance.MONETARY.Weights",
									"*balanceValue": "`;2`",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 11,
									},
								},
							},
							{
								ID: "SetMonetaryCostIncrements",
								Opts: map[string]any{
									"*balancePath":  "*balance.MONETARY.CostIncrements",
									"*balanceValue": "`*string:~*req.ToR:*data;1024;0;0.01`",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 9,
									},
								},
								Blockers: utils.DynamicBlockers{
									{
										Blocker: true,
									},
								},
							},
							{
								ID:        "SetVoiceIDNotFoundFilter",
								FilterIDs: []string{"*string:~*req.Account:1003"},
								Opts: map[string]any{
									"*balancePath":  "*balance.VOICE.ID",
									"*balanceValue": "testBalanceIDMonetaryNOTFOUND",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 10,
									},
								},
							},
							{
								ID: "SetVoiceIDBlocked",
								Opts: map[string]any{
									"*balancePath":  "*balance.VOICE.ID",
									"*balanceValue": "testBalanceIDMonetaryBLOCKED",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 8,
									},
								},
							},
						},
					},
					{
						ID:   "SET_ADD_BAL",
						Type: utils.MetaAddBalance,
						Diktats: []*utils.APDiktat{
							{
								ID: "AddVoiceID",
								Opts: map[string]any{
									"*balancePath":  "*balance.VOICE.ID",
									"*balanceValue": "testBalanceID",
								},
							},
							{
								ID: "AddVoiceType",
								Opts: map[string]any{
									"*balancePath":  "*balance.VOICE.Type",
									"*balanceValue": utils.MetaAbstract,
								},
							},
							{
								ID: "AddVoiceFilterIDs",
								Opts: map[string]any{
									"*balancePath":  "*balance.VOICE.FilterIDs",
									"*balanceValue": "`*string:~*req.ToR:*voice`",
								},
							},
							{
								ID: "AddVoiceUnits",
								Opts: map[string]any{
									"*balancePath":  "*balance.VOICE.Units",
									"*balanceValue": strconv.FormatInt((time.Hour).Nanoseconds(), 10),
								},
							},
							{
								ID: "AddVoiceWeights",
								Opts: map[string]any{
									"*balancePath":  "*balance.VOICE.Weights",
									"*balanceValue": "`;2`",
								},
							},
							{
								ID: "AddVoiceCostIncrements",
								Opts: map[string]any{
									"*balancePath":  "*balance.VOICE.CostIncrements",
									"*balanceValue": "`*string:~*req.ToR:*voice;1000000000;0;0.01`",
								},
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
				Targets: map[string]utils.StringSet{
					utils.MetaThresholds: {"someID": {}},
					utils.MetaStats:      {"someID": {}},
					utils.MetaAttributes: {"someID": {}},
					utils.MetaResources:  {"someID": {}},
					utils.MetaTrends:     {"someID": {}},
					utils.MetaRankings:   {"someID": {}},
					utils.MetaFilters:    {"someID": {}},
				},
				Actions: []*utils.APAction{
					{
						ID:   "Dynamic_Threshold_ID",
						Type: utils.MetaDynamicThreshold,
						Diktats: []*utils.APDiktat{
							{
								ID:        "CreateDynamicThreshold1002",
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_THD_<~*req.Account>;*string:~*req.Account:1002;*string:~*req.Account:1002&10;1;1;1s;false;ACT_LOG_WARNING;true;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 50,
									},
								},
							},
							{
								ID:        "CreateDynamicThreshold1002NotFoundFilter",
								FilterIDs: []string{"*string:~*req.Account:1003"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_THD_2_<~*req.Account>;*string:~*req.Account:1002;*string:~*req.Account:1002&10;1;1;1s;false;ACT_LOG_WARNING;true;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 90,
									},
								},
							},
							{
								ID: "CreateDynamicThreshold1002Blocker",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_THD_3_<~*req.Account>;*string:~*req.Account:1002;*string:~*req.Account:1002&10;1;1;1s;false;ACT_LOG_WARNING;true;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 20,
									},
								},
								Blockers: utils.DynamicBlockers{
									{
										Blocker: true,
									},
								},
							},
							{
								ID: "CreateDynamicThreshold1002Blocked",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_THD_4_<~*req.Account>;*string:~*req.Account:1002;*string:~*req.Account:1002&10;1;1;1s;false;ACT_LOG_WARNING;true;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 10,
									},
								},
							},
						},
					},
					{
						ID:   "Dynamic_Stats_ID",
						Type: utils.MetaDynamicStats,
						Diktats: []*utils.APDiktat{
							{
								ID:        "CreateDynamicStat1002",
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_STAT_<~*req.Account>;*string:~*req.Account:1002;*string:~*req.Account:1002&30;*string:~*req.Account:1002&true;100;-1;0;false;*none;*tcc&*tcd;*string:~*req.Account:1002;*string:~*req.Account:1002&true;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 50,
									},
								},
							},
							{
								ID:        "CreateDynamicStat1002NotFoundFilter",
								FilterIDs: []string{"*string:~*req.Account:1003"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_STAT_2_<~*req.Account>;*string:~*req.Account:1002;*string:~*req.Account:1002&30;*string:~*req.Account:1002&true;100;-1;0;false;*none;*tcc&*tcd;*string:~*req.Account:1002;*string:~*req.Account:1002&true;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 90,
									},
								},
							},
							{
								ID: "CreateDynamicStat1002Blocker",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_STAT_3_<~*req.Account>;*string:~*req.Account:1002;*string:~*req.Account:1002&30;*string:~*req.Account:1002&true;100;-1;0;false;*none;*tcc&*tcd;*string:~*req.Account:1002;*string:~*req.Account:1002&true;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 20,
									},
								},
								Blockers: utils.DynamicBlockers{
									{
										Blocker: true,
									},
								},
							},
							{
								ID: "CreateDynamicStat10022Blocked",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_STAT_4_<~*req.Account>;*string:~*req.Account:1002;*string:~*req.Account:1002&30;*string:~*req.Account:1002&true;100;-1;0;false;*none;*tcc&*tcd;*string:~*req.Account:1002;*string:~*req.Account:1002&true;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 10,
									},
								},
							},
						},
					},
					{
						ID:   "Dynamic_Attribute_ID",
						Type: utils.MetaDynamicAttribute,
						Diktats: []*utils.APDiktat{
							{
								ID:        "CreateDynamicAttribute1002",
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_ATTR_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&30;*string:~*req.Account:<~*req.Account>&true;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&true;*req.Subject;*constant;SUPPLIER1;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 50,
									},
								},
							},
							{
								ID:        "CreateDynamicAttribute1002",
								FilterIDs: []string{"*string:~*req.Account:1003NotFoundFilter"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_ATTR_2_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&30;*string:~*req.Account:<~*req.Account>&true;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&true;*req.Subject;*constant;SUPPLIER1;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 90,
									},
								},
							},
							{
								ID: "CreateDynamicAttribute1002Blockers",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_ATTR_3_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&30;*string:~*req.Account:<~*req.Account>&true;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&true;*req.Subject;*constant;SUPPLIER1;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 20,
									},
								},
								Blockers: utils.DynamicBlockers{
									{
										Blocker: true,
									},
								},
							},
							{
								ID: "CreateDynamicAttribute1002Blocked",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_ATTR_4_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&30;*string:~*req.Account:<~*req.Account>&true;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&true;*req.Subject;*constant;SUPPLIER1;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 10,
									},
								},
							},
						},
					},
					{
						ID:   "Dynamic_Resource_ID",
						Type: utils.MetaDynamicResource,
						Diktats: []*utils.APDiktat{
							{
								ID:        "CreateDynamicResource1002",
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_RES_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&30;5s;5;alloc_msg;true;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 50,
									},
								},
							},
							{
								ID:        "CreateDynamicResource1002NotFoundFilter",
								FilterIDs: []string{"*string:~*req.Account:1003"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_RES_2_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&30;5s;5;alloc_msg;true;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 90,
									},
								},
							},
							{
								ID: "CreateDynamicResource1002Blocker",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_RES_3_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&30;5s;5;alloc_msg;true;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 20,
									},
								},
								Blockers: utils.DynamicBlockers{
									{
										Blocker: true,
									},
								},
							},
							{
								ID: "CreateDynamicResource1002Blocked",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_RES_4_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&30;5s;5;alloc_msg;true;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 10,
									},
								},
							},
						},
					},
					{
						ID:   "Dynamic_Trend_ID",
						Type: utils.MetaDynamicTrend,
						Diktats: []*utils.APDiktat{
							{
								ID:        "CreateDynamicTrend1002",
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_TRND_<~*req.Account>;@every 1s;Stats1_1;*acc&*tcc;-1;-1;1;*last;1;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 50,
									},
								},
							},
							{
								ID:        "CreateDynamicTrend1002NotFoundFilter",
								FilterIDs: []string{"*string:~*req.Account:1003"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_TRND_2_<~*req.Account>;@every 1s;Stats1_1;*acc&*tcc;-1;-1;1;*last;1;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 90,
									},
								},
							},
							{
								ID: "CreateDynamicTrend1002Blocker",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_TRND_3_<~*req.Account>;@every 1s;Stats1_1;*acc&*tcc;-1;-1;1;*last;1;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 20,
									},
								},
								Blockers: utils.DynamicBlockers{
									{
										Blocker: true,
									},
								},
							},
							{
								ID: "CreateDynamicTrend1002Blocked",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_TRND_4_<~*req.Account>;@every 1s;Stats1_1;*acc&*tcc;-1;-1;1;*last;1;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 10,
									},
								},
							},
						},
					},
					{
						ID:   "Dynamic_Ranking_ID",
						Type: utils.MetaDynamicRanking,
						Diktats: []*utils.APDiktat{
							{
								ID:        "CreateDynamicRanking1002",
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_RNK_<~*req.Account>;@every 1s;Stats1&Stats2;*acc&*tcc;*asc;*acc&*pdd;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 50,
									},
								},
							},
							{
								ID:        "CreateDynamicRanking1002NotFoundFilter",
								FilterIDs: []string{"*string:~*req.Account:1003"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_RNK_2_<~*req.Account>;@every 1s;Stats1&Stats2;*acc&*tcc;*asc;*acc&*pdd;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 90,
									},
								},
							},
							{
								ID: "CreateDynamicRanking1002Blocker",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_RNK_3_<~*req.Account>;@every 1s;Stats1&Stats2;*acc&*tcc;*asc;*acc&*pdd;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 20,
									},
								},
								Blockers: utils.DynamicBlockers{
									{
										Blocker: true,
									},
								},
							},
							{
								ID: "CreateDynamicRanking1002Blocked",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_RNK_4_<~*req.Account>;@every 1s;Stats1&Stats2;*acc&*tcc;*asc;*acc&*pdd;true;THID1&THID2;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 10,
									},
								},
							},
						},
					},
					{
						ID:   "Dynamic_Filter_ID",
						Type: utils.MetaDynamicFilter,
						Diktats: []*utils.APDiktat{
							{
								ID:        "CreateDynamicFilter1002",
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_FLTR_<~*req.Account>;*string;~*req.Account;1003&1002;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 50,
									},
								},
							},
							{
								ID:        "CreateDynamicFilter1002NotFoundFilter",
								FilterIDs: []string{"*string:~*req.Account:1003"},
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_FLTR_2_<~*req.Account>;*string;~*req.Account;1003&1002;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 90,
									},
								},
							},
							{
								ID: "CreateDynamicFilter1002Blocker",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_FLTR_3_<~*req.Account>;*string;~*req.Account;1003&1002;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 20,
									},
								},
								Blockers: utils.DynamicBlockers{
									{
										Blocker: true,
									},
								},
							},
							{
								ID: "CreateDynamicFilter1002Blocked",
								Opts: map[string]any{
									"*template": "*tenant;DYNAMICLY_FLTR_4_<~*req.Account>;*string;~*req.Account;1003&1002;~*opts",
								},
								Weights: utils.DynamicWeights{
									{
										Weight: 10,
									},
								},
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
	t.Run("GetDynamicThresholdProfile", func(t *testing.T) {
		var thrsholds []*engine.ThresholdProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetThresholdProfiles,
			&utils.ArgsItemIDs{
				Tenant: utils.CGRateSorg,
			}, &thrsholds); err != nil {
			t.Errorf("AdminSv1GetThresholdProfiles failed unexpectedly: %v", err)
		}
		if len(thrsholds) != 3 {
			t.Fatalf("AdminSv1GetThresholdProfiles len(thrsholds)=%v, want 3", len(thrsholds))
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
				ID:        "DYNAMICLY_THD_3_1002",
				FilterIDs: []string{"*string:~*req.Account:1002"},
				MaxHits:   1,
				MinHits:   1,
				MinSleep:  time.Second,
				Blocker:   false,
				Weights: utils.DynamicWeights{
					&utils.DynamicWeight{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Weight:    10,
					},
				},
				ActionProfileIDs: []string{"ACT_LOG_WARNING"},
				Async:            true,
			},
			{
				Tenant:    utils.CGRateSorg,
				ID:        "DYNAMICLY_THD_1002",
				FilterIDs: []string{"*string:~*req.Account:1002"},
				MaxHits:   1,
				MinHits:   1,
				MinSleep:  time.Second,
				Blocker:   false,
				Weights: utils.DynamicWeights{
					&utils.DynamicWeight{
						FilterIDs: []string{"*string:~*req.Account:1002"},
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

	t.Run("GetDynamicStatQueueProfile", func(t *testing.T) {
		exp := []*engine.StatQueueProfile{
			{
				Tenant:    utils.CGRateSorg,
				ID:        "DYNAMICLY_STAT_3_1002",
				FilterIDs: []string{"*string:~*req.Account:1002"},
				Weights: utils.DynamicWeights{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Weight:    30,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Blocker:   true,
					},
				},
				QueueLength:  100,
				TTL:          -1,
				MinItems:     0,
				Stored:       false,
				ThresholdIDs: []string{utils.MetaNone},
				Metrics: []*engine.MetricWithFilters{
					{
						MetricID:  utils.MetaTCC,
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Blockers: utils.DynamicBlockers{
							{
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Blocker:   true,
							},
						},
					},
					{
						MetricID:  utils.MetaTCD,
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Blockers: utils.DynamicBlockers{
							{
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Blocker:   true,
							},
						},
					},
				},
			},
			{
				Tenant:    utils.CGRateSorg,
				ID:        "DYNAMICLY_STAT_1002",
				FilterIDs: []string{"*string:~*req.Account:1002"},
				Weights: utils.DynamicWeights{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Weight:    30,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Blocker:   true,
					},
				},
				QueueLength:  100,
				TTL:          -1,
				MinItems:     0,
				Stored:       false,
				ThresholdIDs: []string{utils.MetaNone},
				Metrics: []*engine.MetricWithFilters{
					{
						MetricID:  utils.MetaTCC,
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Blockers: utils.DynamicBlockers{
							{
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Blocker:   true,
							},
						},
					},
					{
						MetricID:  utils.MetaTCD,
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Blockers: utils.DynamicBlockers{
							{
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Blocker:   true,
							},
						},
					},
				},
			},
		}

		var rply []*engine.StatQueueProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetStatQueueProfiles, &utils.ArgsItemIDs{
			Tenant: utils.CGRateSorg,
		}, &rply); err != nil {
			t.Error(err)
		} else if len(rply) != 2 {
			t.Fatalf("AdminSv1GetStatQueueProfiles len(rply)=%v, want 2", len(rply))
		}
		sort.Slice(rply, func(i, j int) bool {
			return rply[i].ID > rply[j].ID
		})

		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected <%v> \n received <%v>", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	})

	t.Run("GetDynamicAttributeProfile", func(t *testing.T) {
		var attrs []*utils.APIAttributeProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetAttributeProfiles,
			&utils.ArgsItemIDs{
				Tenant: utils.CGRateSorg,
			}, &attrs); err != nil {
			t.Errorf("AdminSv1GetAttributeProfiles failed unexpectedly: %v", err)
		}
		if len(attrs) != 2 {
			t.Fatalf("AdminSv1GetAttributeProfiles len(attrs)=%v, want 2", len(attrs))
		}
		sort.Slice(attrs, func(i, j int) bool {
			return attrs[i].ID > attrs[j].ID
		})
		exp := []*utils.APIAttributeProfile{
			{
				Tenant:    utils.CGRateSorg,
				ID:        "DYNAMICLY_ATTR_3_1002",
				FilterIDs: []string{"*string:~*req.Account:1002"},
				Weights: utils.DynamicWeights{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Weight:    30,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Blocker:   true,
					},
				},
				Attributes: []*utils.ExternalAttribute{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Blockers: utils.DynamicBlockers{
							{
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Blocker:   true,
							},
						},
						Path:  "*req.Subject",
						Type:  "*constant",
						Value: "SUPPLIER1",
					},
				},
			},
			{
				Tenant:    utils.CGRateSorg,
				ID:        "DYNAMICLY_ATTR_1002",
				FilterIDs: []string{"*string:~*req.Account:1002"},
				Weights: utils.DynamicWeights{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Weight:    30,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Blocker:   true,
					},
				},
				Attributes: []*utils.ExternalAttribute{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Blockers: utils.DynamicBlockers{
							{
								FilterIDs: []string{"*string:~*req.Account:1002"},
								Blocker:   true,
							},
						},
						Path:  "*req.Subject",
						Type:  "*constant",
						Value: "SUPPLIER1",
					},
				},
			},
		}

		if !reflect.DeepEqual(exp, attrs) {
			t.Errorf("Expected <%v>\nReceived <%v>", utils.ToJSON(exp), utils.ToJSON(attrs))
		}
	})

	t.Run("GetDynamicResourceProfile", func(t *testing.T) {
		var rsc []*utils.ResourceProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetResourceProfiles,
			&utils.ArgsItemIDs{
				Tenant: utils.CGRateSorg,
			}, &rsc); err != nil {
			t.Errorf("AdminSv1GetResourceProfiles failed unexpectedly: %v", err)
		}
		if len(rsc) != 2 {
			t.Fatalf("AdminSv1GetResourceProfiles len(rsc)=%v, want 2", len(rsc))
		}
		sort.Slice(rsc, func(i, j int) bool {
			return rsc[i].ID > rsc[j].ID
		})
		exp := []*utils.ResourceProfile{
			{
				Tenant:            "cgrates.org",
				ID:                "DYNAMICLY_RES_3_1002",
				FilterIDs:         []string{"*string:~*req.Account:1002"},
				UsageTTL:          5 * time.Second,
				Limit:             5,
				AllocationMessage: "alloc_msg",
				Blocker:           true,
				Stored:            true,
				Weights: utils.DynamicWeights{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Weight:    30,
					},
				},
				ThresholdIDs: []string{"THID1", "THID2"},
			},
			{
				Tenant:            "cgrates.org",
				ID:                "DYNAMICLY_RES_1002",
				FilterIDs:         []string{"*string:~*req.Account:1002"},
				UsageTTL:          5 * time.Second,
				Limit:             5,
				AllocationMessage: "alloc_msg",
				Blocker:           true,
				Stored:            true,
				Weights: utils.DynamicWeights{
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
						Weight:    30,
					},
				},
				ThresholdIDs: []string{"THID1", "THID2"},
			},
		}

		if !reflect.DeepEqual(exp, rsc) {
			t.Errorf("Expected <%v>\nReceived <%v>", utils.ToJSON(exp), utils.ToJSON(rsc))
		}
	})

	t.Run("GetDynamicTrendProfile", func(t *testing.T) {
		var rcv []*utils.TrendProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetTrendProfiles,
			&utils.ArgsItemIDs{
				Tenant: utils.CGRateSorg,
			}, &rcv); err != nil {
			t.Errorf("AdminSv1GetTrendProfiles failed unexpectedly: %v", err)
		}
		if len(rcv) != 2 {
			t.Fatalf("AdminSv1GetTrendProfiles len(rcv)=%v, want 2", len(rcv))
		}
		sort.Slice(rcv, func(i, j int) bool {
			return rcv[i].ID > rcv[j].ID
		})
		exp := []*utils.TrendProfile{
			{
				Tenant:          "cgrates.org",
				ID:              "DYNAMICLY_TRND_3_1002",
				Schedule:        "@every 1s",
				StatID:          "Stats1_1",
				Metrics:         []string{"*acc", "*tcc"},
				TTL:             -1,
				QueueLength:     -1,
				MinItems:        1,
				CorrelationType: "*last",
				Tolerance:       1,
				Stored:          true,
				ThresholdIDs:    []string{"THID1", "THID2"},
			},
			{

				Tenant:          "cgrates.org",
				ID:              "DYNAMICLY_TRND_1002",
				Schedule:        "@every 1s",
				StatID:          "Stats1_1",
				Metrics:         []string{"*acc", "*tcc"},
				TTL:             -1,
				QueueLength:     -1,
				MinItems:        1,
				CorrelationType: "*last",
				Tolerance:       1,
				Stored:          true,
				ThresholdIDs:    []string{"THID1", "THID2"},
			},
		}

		if !reflect.DeepEqual(exp, rcv) {
			t.Errorf("Expected <%v>\nReceived <%v>", utils.ToJSON(exp), utils.ToJSON(rcv))
		}
	})

	t.Run("GetDynamicRankingProfile", func(t *testing.T) {
		var rcv []*utils.RankingProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetRankingProfiles,
			&utils.ArgsItemIDs{
				Tenant: utils.CGRateSorg,
			}, &rcv); err != nil {
			t.Errorf("AdminSv1GetRankingProfiles failed unexpectedly: %v", err)
		}
		if len(rcv) != 2 {
			t.Fatalf("AdminSv1GetRankingProfiles len(rcv)=%v, want 2", len(rcv))
		}
		sort.Slice(rcv, func(i, j int) bool {
			return rcv[i].ID > rcv[j].ID
		})
		exp := []*utils.RankingProfile{
			{
				Tenant:            "cgrates.org",
				ID:                "DYNAMICLY_RNK_3_1002",
				Schedule:          "@every 1s",
				StatIDs:           []string{"Stats1", "Stats2"},
				MetricIDs:         []string{"*acc", "*tcc"},
				Sorting:           "*asc",
				SortingParameters: []string{"*acc", "*pdd"},
				Stored:            true,
				ThresholdIDs:      []string{"THID1", "THID2"},
			},
			{

				Tenant:            "cgrates.org",
				ID:                "DYNAMICLY_RNK_1002",
				Schedule:          "@every 1s",
				StatIDs:           []string{"Stats1", "Stats2"},
				MetricIDs:         []string{"*acc", "*tcc"},
				Sorting:           "*asc",
				SortingParameters: []string{"*acc", "*pdd"},
				Stored:            true,
				ThresholdIDs:      []string{"THID1", "THID2"},
			},
		}

		if !reflect.DeepEqual(exp, rcv) {
			t.Errorf("Expected <%v>\nReceived <%v>", utils.ToJSON(exp), utils.ToJSON(rcv))
		}
	})

	t.Run("GetDynamicFilter", func(t *testing.T) {
		var rcv []*engine.Filter
		if err := client.Call(context.Background(), utils.AdminSv1GetFilters,
			&utils.ArgsItemIDs{
				Tenant: utils.CGRateSorg,
			}, &rcv); err != nil {
			t.Errorf("AdminSv1GetFilters failed unexpectedly: %v", err)
		}
		if len(rcv) != 2 {
			t.Fatalf("AdminSv1GetFilters len(rcv)=%v, want 2", len(rcv))
		}
		sort.Slice(rcv, func(i, j int) bool {
			return rcv[i].ID > rcv[j].ID
		})
		exp := []*engine.Filter{
			{
				Tenant: "cgrates.org",
				ID:     "DYNAMICLY_FLTR_3_1002",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"1003", "1002"},
					},
				},
			},
			{

				Tenant: "cgrates.org",
				ID:     "DYNAMICLY_FLTR_1002",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"1003", "1002"},
					},
				},
			},
		}

		if !reflect.DeepEqual(exp, rcv) {
			t.Errorf("Expected <%v>\nReceived <%v>", utils.ToJSON(exp), utils.ToJSON(rcv))
		}
	})
}
