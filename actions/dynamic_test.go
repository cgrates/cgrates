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

package actions

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type ccMock struct {
	calls map[string]func(ctx *context.Context, args any, reply any) error
}

func (ccM *ccMock) Call(ctx *context.Context, serviceMethod string, args any, reply any) (err error) {
	if call, has := ccM.calls[serviceMethod]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, args, reply)
	}
}

func TestParseParamStringToMap(t *testing.T) {
	testcases := []struct {
		name      string
		paramStr  string
		wantMap   map[string]any
		expectErr bool
		errMsg    string
	}{
		{
			name:     "TenantAndAccount",
			paramStr: "*tenant:cgrates.org&*account:1001",
			wantMap: map[string]any{
				"*tenant":  "cgrates.org",
				"*account": "1001",
			},
			expectErr: false,
		},
		{
			name:     "TenantAccountDestination",
			paramStr: "*tenant:cgrates.org&*account:1001&*destination:1002",
			wantMap: map[string]any{
				"*tenant":      "cgrates.org",
				"*account":     "1001",
				"*destination": "1002",
			},
			expectErr: false,
		},
		{
			name:      "InvalidPairMissingColon",
			paramStr:  "*tenantcgrates.org&*profileID:ID1001",
			wantMap:   nil,
			expectErr: true,
			errMsg:    "invalid key-value pair: *tenantcgrates.org",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := make(map[string]any)
			err := parseParamStringToMap(tc.paramStr, got)

			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				} else if err.Error() != tc.errMsg {
					t.Errorf("expected error <%v>, got <%v>", tc.errMsg, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tc.wantMap) {
				t.Errorf("expected %+v, got %+v", tc.wantMap, got)
			}
		})
	}
}

func TestActDynamicThresholdID(t *testing.T) {
	aL1 := &actDynamicThreshold{
		aCfg: &utils.APAction{
			ID: "TH1",
		},
	}
	if aL1.id() != "TH1" {
		t.Errorf("expected ID <TH1>, got <%s>", aL1.id())
	}

	aL2 := &actDynamicThreshold{
		aCfg: &utils.APAction{
			ID: "",
		},
	}
	if aL2.id() != "" {
		t.Errorf("expected empty ID, got <%s>", aL2.id())
	}
}

func TestActDynamicThresholdCfg(t *testing.T) {
	aCfg1 := &utils.APAction{ID: "TH1"}
	aL1 := &actDynamicThreshold{aCfg: aCfg1}
	if got := aL1.cfg(); got != aCfg1 {
		t.Errorf("Scenario 1: expected cfg <%v>, got <%v>", aCfg1, got)
	}

	aCfg2 := &utils.APAction{}
	aL2 := &actDynamicThreshold{aCfg: aCfg2}
	if got := aL2.cfg(); got != aCfg2 {
		t.Errorf("Scenario 2: expected cfg <%v>, got <%v>", aCfg2, got)
	}

	aL3 := &actDynamicThreshold{aCfg: nil}
	if got := aL3.cfg(); got != nil {
		t.Errorf("Scenario 3: expected cfg <nil>, got <%v>", got)
	}
}

func TestActDynamicThresholdExecuteSort(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().AdminSConns = []string{"admins"}

	connMgr := engine.NewConnManager(cfg)

	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	fltrS := engine.NewFilterS(cfg, connMgr, dm)

	data := utils.MapStorage{
		utils.AccountField: "1001",
	}

	diktats := []*utils.APDiktat{
		{
			ID:        "d1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights:   utils.DynamicWeights{{Weight: 10}},
			Opts: map[string]any{
				utils.MetaTemplate: "cgrates.org;THD1;;1;1;1s;true;AP1;true;ee1;key1:val1",
			},
		},
		{
			ID:        "d2",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights:   utils.DynamicWeights{{Weight: 20}},
			Opts: map[string]any{
				utils.MetaTemplate: "cgrates.org;THD2;;1;1;1s;true;AP2;true;ee2;key2:val2",
			},
		},
	}

	act := &actDynamicThreshold{
		config:  cfg,
		connMgr: connMgr,
		fltrS:   fltrS,
		tnt:     "cgrates.org",
		cgrEv: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "evID",
			Event:  map[string]any{utils.AccountField: "1001"},
		},
		aCfg: &utils.APAction{
			ID:      "ACT_DYNAMIC",
			Diktats: diktats,
		},
	}

	weights := map[string]float64{
		"d1": 10,
		"d2": 20,
	}

	sorted := make([]*utils.APDiktat, len(diktats))
	copy(sorted, diktats)
	slices.SortFunc(sorted, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})

	if sorted[0].ID != "d2" || sorted[1].ID != "d1" {
		t.Errorf("expected d2 to be first and d1 second after sort, got %v", []string{sorted[0].ID, sorted[1].ID})
	}

	ctx := context.Background()
	trgID := "trigger1"
	err := act.execute(ctx, data, trgID)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
}

func TestActDynamicThresholdExecute(t *testing.T) {
	engine.Cache.Clear(nil)
	defer engine.Cache.Clear(nil)

	var tpwo *engine.ThresholdProfileWithAPIOpts
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AdminSv1SetThresholdProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if tpwo, canCast = args.(*engine.ThresholdProfileWithAPIOpts); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				return nil
			},
		},
	}

	testcases := []struct {
		name    string
		diktats []*utils.APDiktat
		expTpwo *engine.ThresholdProfileWithAPIOpts
		wantErr bool
	}{
		{
			name: "SuccessfulRequest",
			expTpwo: &engine.ThresholdProfileWithAPIOpts{
				ThresholdProfile: &engine.ThresholdProfile{
					Tenant:    "cgrates.org",
					ID:        "THD_ACNT_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					MaxHits:   1,
					MinHits:   1,
					MinSleep:  1 * time.Second,
					Blocker:   true,
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					ActionProfileIDs: []string{"ACT_LOG_WARNING"},
					Async:            true,
					EeIDs:            []string{"eeID1", "eeID2"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;THD_ACNT_1001;*string:~*req.Account:1001;&20;1;1;1s;true;ACT_LOG_WARNING;true;eeID1&eeID2;key:value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestWithDyanmicPaths",
			expTpwo: &engine.ThresholdProfileWithAPIOpts{
				ThresholdProfile: &engine.ThresholdProfile{
					Tenant:    "cgrates.org",
					ID:        "THD_ACNT_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					MaxHits:   1,
					MinHits:   1,
					MinSleep:  1 * time.Second,
					Blocker:   true,
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					ActionProfileIDs: []string{"ACT_LOG_WARNING"},
					Async:            true,
					EeIDs:            []string{"eeID1", "eeID2"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;THD_ACNT_<~*req.Account>;*string:~*req.Account:<~*req.Account>;&20;1;1;1s;true;ACT_LOG_WARNING;true;eeID1&eeID2;key:value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestEmptyFields",
			expTpwo: &engine.ThresholdProfileWithAPIOpts{
				ThresholdProfile: &engine.ThresholdProfile{
					Tenant:           "cgrates.org",
					ID:               "THD_ACNT_1001",
					FilterIDs:        nil,
					MaxHits:          0,
					MinHits:          0,
					MinSleep:         0,
					Blocker:          false,
					Weights:          nil,
					ActionProfileIDs: nil,
					Async:            false,
					EeIDs:            nil,
				},
				APIOpts: map[string]any{},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;THD_ACNT_1001;;;;;;;;;;",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "EmptyDiktats",
			diktats: []*utils.APDiktat{},
			wantErr: true,
		},
		{
			name: "InvalidNumberOfParams",
			diktats: []*utils.APDiktat{
				{
					ID: "d4",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;THD_ACNT_1004;1;2",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidMaxHitsParam",
			diktats: []*utils.APDiktat{
				{
					ID: "d5",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;THD_ACNT_1005;*string:~*req.Account:1005;;invalid;1;1s;true;ACT_LOG_WARNING;true;eeID5;key:value",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidMinSleepParam",
			diktats: []*utils.APDiktat{
				{
					ID: "d6",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;THD_ACNT_1006;*string:~*req.Account:1006;&20;1;1;invalid;true;ACT_LOG_WARNING;true;eeID6;key:value",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "WeightParamUnsupportedFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d10",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;THD_ACNT_1010;;w1&w2&w3;1;1;1s;true;ACT_LOG_WARNING;true;eeID10;key:value",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "WeightParamInvalidFloat",
			diktats: []*utils.APDiktat{
				{
					ID: "d11",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;THD_ACNT_1011;;w1&invalid;1;1;1s;true;ACT_LOG_WARNING;true;eeID11;key:value",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "APIOptsInvalidFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d12",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;THD_ACNT_1012;;;;1;1;1s;true;ACT_LOG_WARNING;true;eeID12;keyvalue",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "BlockerFail",
			diktats: []*utils.APDiktat{
				{
					ID: "d14",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;THD_ACNT_1014;*string:~*req.Account:1014;;1;1;1s;notbool;ACT_LOG_WARNING;true;eeID14;key:value",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultCGRConfig()
			cfg.GeneralCfg().DefaultCaching = utils.MetaNone
			cfg.ActionSCfg().AdminSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS)}
			connMgr := engine.NewConnManager(cfg)
			dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
			dm := engine.NewDataManager(dataDB, cfg, connMgr)
			fltrS := engine.NewFilterS(cfg, connMgr, dm)
			rpcInternal := make(chan birpc.ClientConnector, 1)
			rpcInternal <- ccM
			connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), utils.AdminSv1, rpcInternal)

			act := &actDynamicThreshold{
				config:  cfg,
				connMgr: connMgr,
				fltrS:   fltrS,
				tnt:     "cgrates.org",
				cgrEv: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evID",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				},
				aCfg: &utils.APAction{
					ID:      "ACT_DYNAMIC",
					Diktats: tc.diktats,
				},
			}

			t.Cleanup(func() {
				tpwo = nil
			})

			ctx := context.Background()
			data := utils.MapStorage{
				"*req": map[string]any{
					utils.AccountField: "1001",
				},
			}
			trgID := "trigger1"
			err := act.execute(ctx, data, trgID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected an error, but got nil")
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(tpwo, tc.expTpwo) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expTpwo), utils.ToJSON(tpwo))
			}
		})
	}
}

func TestActDynamicStatsID(t *testing.T) {
	actionCfg := &utils.APAction{
		ID: "StatsID",
	}
	aL := &actDynamicStats{
		aCfg: actionCfg,
	}
	got := aL.id()
	if got != "StatsID" {
		t.Errorf("expected %q, got %q", "StatsID", got)
	}
}

func TestActDynamicStatsCfg(t *testing.T) {
	testcases := []struct {
		name     string
		actCfg   *utils.APAction
		expIsNil bool
	}{
		{
			name: "ValidAPAction",
			actCfg: &utils.APAction{
				ID: "StatsProfile!",
			},
			expIsNil: false,
		},
		{
			name:     "NilAPAction",
			actCfg:   nil,
			expIsNil: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			aL := &actDynamicStats{
				config:  config.NewDefaultCGRConfig(),
				connMgr: &engine.ConnManager{},
				aCfg:    tc.actCfg,
				tnt:     "cgrates.org",
				cgrEv:   &utils.CGREvent{},
			}

			got := aL.cfg()

			if tc.expIsNil && got != nil {
				t.Errorf("expected nil, got %+v", got)
			}
			if !tc.expIsNil && got != tc.actCfg {
				t.Errorf("expected pointer %p, got %p", tc.actCfg, got)
			}
		})
	}
}

func TestActDynamicStatsExecute(t *testing.T) {
	engine.Cache.Clear(nil)
	defer engine.Cache.Clear(nil)

	var sqpwo *engine.StatQueueProfileWithAPIOpts
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AdminSv1SetStatQueueProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if sqpwo, canCast = args.(*engine.StatQueueProfileWithAPIOpts); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				return nil
			},
		},
	}

	testcases := []struct {
		name     string
		diktats  []*utils.APDiktat
		expSqpwo *engine.StatQueueProfileWithAPIOpts
		wantErr  bool
	}{
		{
			name: "SuccessfulRequest",
			expSqpwo: &engine.StatQueueProfileWithAPIOpts{
				StatQueueProfile: &engine.StatQueueProfile{
					Tenant:       "cgrates.org",
					ID:           "STATS_1001",
					FilterIDs:    []string{"*string:~*req.Account:1001"},
					QueueLength:  100,
					TTL:          1 * time.Hour,
					MinItems:     1,
					Stored:       true,
					ThresholdIDs: []string{"THD_1", "THD_2"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Weight:    20.0,
						},
					},
					Blockers: utils.DynamicBlockers{
						&utils.DynamicBlocker{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Blocker:   true,
						},
					},
					Metrics: []*engine.MetricWithFilters{
						{
							MetricID:  "ASR",
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Blockers: utils.DynamicBlockers{
								&utils.DynamicBlocker{
									FilterIDs: []string{"*string:~*req.Account:1001"},
									Blocker:   false,
								},
							},
						},
						{
							MetricID:  "PDD",
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Blockers: utils.DynamicBlockers{
								&utils.DynamicBlocker{
									FilterIDs: []string{"*string:~*req.Account:1001"},
									Blocker:   false,
								},
							},
						},
					},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1001;*string:~*req.Account:1001;*string:~*req.Account:1001&20;*string:~*req.Account:1001&true;100;1h;1;true;THD_1&THD_2;ASR&PDD;*string:~*req.Account:1001;*string:~*req.Account:1001&false;key:value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestWithDynamicPaths",
			expSqpwo: &engine.StatQueueProfileWithAPIOpts{
				StatQueueProfile: &engine.StatQueueProfile{
					Tenant:       "cgrates.org",
					ID:           "STATS_1001",
					FilterIDs:    []string{"*string:~*req.Account:1001"},
					QueueLength:  100,
					TTL:          1 * time.Hour,
					MinItems:     1,
					Stored:       true,
					ThresholdIDs: []string{"THD_1", "THD_2"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Weight:    20.0,
						},
					},
					Blockers: utils.DynamicBlockers{
						&utils.DynamicBlocker{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Blocker:   true,
						},
					},
					Metrics: []*engine.MetricWithFilters{
						{
							MetricID:  "ASR",
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Blockers: utils.DynamicBlockers{
								&utils.DynamicBlocker{
									FilterIDs: []string{"*string:~*req.Account:1001"},
									Blocker:   false,
								},
							},
						},
						{
							MetricID:  "PDD",
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Blockers: utils.DynamicBlockers{
								&utils.DynamicBlocker{
									FilterIDs: []string{"*string:~*req.Account:1001"},
									Blocker:   false,
								},
							},
						},
					},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&20;*string:~*req.Account:<~*req.Account>&true;100;1h;1;true;THD_1&THD_2;ASR&PDD;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&false;key:value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestEmptyFields",
			expSqpwo: &engine.StatQueueProfileWithAPIOpts{
				StatQueueProfile: &engine.StatQueueProfile{
					Tenant:       "cgrates.org",
					ID:           "STATS_1001",
					FilterIDs:    nil,
					QueueLength:  0,
					TTL:          0,
					MinItems:     0,
					Stored:       false,
					ThresholdIDs: nil,
					Weights:      nil,
					Blockers:     nil,
					Metrics:      nil,
				},
				APIOpts: map[string]any{},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1001;;;;;;;;;;;;",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "EmptyDiktats",
			diktats: []*utils.APDiktat{},
			wantErr: true,
		},
		{
			name: "InvalidNumberOfParams",
			diktats: []*utils.APDiktat{
				{
					ID: "d4",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1004;param1;param2",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidQueueLengthParam",
			diktats: []*utils.APDiktat{
				{
					ID: "d5",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1005;*string:~*req.Account:1005;;;invalid;1h;1;true;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidTTLParam",
			diktats: []*utils.APDiktat{
				{
					ID: "d6",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1006;*string:~*req.Account:1006;;;100;invalid;1;true;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidMinItemsParam",
			diktats: []*utils.APDiktat{
				{
					ID: "d7",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1007;*string:~*req.Account:1007;;;100;1h;invalid;true;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidStoredParam",
			diktats: []*utils.APDiktat{
				{
					ID: "d8",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1008;*string:~*req.Account:1008;;;100;1h;1;invalid;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "WeightParamUnsupportedFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d9",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1009;;w1&w2&w3;;;;;;;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "WeightParamInvalidFloat",
			diktats: []*utils.APDiktat{
				{
					ID: "d10",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1010;;w1&invalid;;;;;;;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "BlockerParamUnsupportedFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d11",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1011;;;b1&b2&b3;;;;;;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "BlockerParamInvalidBool",
			diktats: []*utils.APDiktat{
				{
					ID: "d12",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1012;;;b1&invalid;;;;;;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "MetricBlockerParamUnsupportedFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d13",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1013;;;;;;;;;;ASR;;mb1&mb2&mb3;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "MetricBlockerParamInvalidBool",
			diktats: []*utils.APDiktat{
				{
					ID: "d14",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1014;;;;;;;;;;ASR;;mb1&invalid;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "APIOptsInvalidFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d15",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1015;;;;;;;;;;;;;keyvalue",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "NoAdminSConns",
			diktats: []*utils.APDiktat{
				{
					ID: "d16",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;STATS_1016;;;;;;;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultCGRConfig()
			cfg.GeneralCfg().DefaultCaching = utils.MetaNone

			if tc.name != "NoAdminSConns" {
				cfg.ActionSCfg().AdminSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS)}
			}

			connMgr := engine.NewConnManager(cfg)
			dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
			dm := engine.NewDataManager(dataDB, cfg, connMgr)
			fltrS := engine.NewFilterS(cfg, connMgr, dm)

			rpcInternal := make(chan birpc.ClientConnector, 1)
			rpcInternal <- ccM
			connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), utils.AdminSv1, rpcInternal)

			act := &actDynamicStats{
				config:  cfg,
				connMgr: connMgr,
				fltrS:   fltrS,
				tnt:     "cgrates.org",
				cgrEv: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evID",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				},
				aCfg: &utils.APAction{
					ID:      "ACT_DYNAMIC_STATS",
					Diktats: tc.diktats,
				},
			}

			t.Cleanup(func() {
				sqpwo = nil
			})

			ctx := context.Background()
			data := utils.MapStorage{
				"*req": map[string]any{
					utils.AccountField: "1001",
				},
			}
			trgID := "trigger1"

			err := act.execute(ctx, data, trgID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected an error, but got nil")
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(sqpwo, tc.expSqpwo) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expSqpwo), utils.ToJSON(sqpwo))
			}
		})
	}
}

func TestActDynamicAttributeId(t *testing.T) {
	testcases := []struct {
		name   string
		aCfg   *utils.APAction
		expect string
	}{
		{
			name: "WithValidID",
			aCfg: &utils.APAction{
				ID:        "AttrA1",
				FilterIDs: []string{"f1", "f2"},
				TTL:       time.Second,
				Type:      "attribute",
				Opts: map[string]any{
					"key1": "val1",
				},
			},
			expect: "AttrA1",
		},
		{
			name: "WithEmptyID",
			aCfg: &utils.APAction{
				ID:        "",
				FilterIDs: nil,
				TTL:       0,
				Type:      "attribute",
				Opts:      nil,
			},
			expect: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			al := &actDynamicAttribute{
				aCfg: tc.aCfg,
			}
			got := al.id()
			if got != tc.expect {
				t.Errorf("expected %q, got %q", tc.expect, got)
			}
		})
	}
}

func TestActDynamicAttributeCfg(t *testing.T) {
	aCfg := &utils.APAction{ID: "AttrA1"}
	al := &actDynamicAttribute{aCfg: aCfg}

	if got := al.cfg(); got != aCfg {
		t.Errorf("expected cfg() to return %+v, got %+v", aCfg, got)
	}
}

func TestActDynamicAttributeExecute(t *testing.T) {
	engine.Cache.Clear(nil)
	defer engine.Cache.Clear(nil)

	var apwo *utils.APIAttributeProfileWithAPIOpts
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AdminSv1SetAttributeProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if apwo, canCast = args.(*utils.APIAttributeProfileWithAPIOpts); !canCast {
					return fmt.Errorf("couldn't cast")
				}
				*reply.(*string) = utils.OK
				return nil
			},
		},
	}

	testcases := []struct {
		name    string
		diktats []*utils.APDiktat
		expApwo *utils.APIAttributeProfileWithAPIOpts
		wantErr bool
	}{
		{
			name: "SuccessfulRequest",
			expApwo: &utils.APIAttributeProfileWithAPIOpts{
				APIAttributeProfile: &utils.APIAttributeProfile{
					Tenant:    "cgrates.org",
					ID:        "ATTR_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Weight:    20.0,
						},
					},
					Blockers: utils.DynamicBlockers{
						&utils.DynamicBlocker{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Blocker:   true,
						},
					},
					Attributes: []*utils.ExternalAttribute{
						{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Blockers: utils.DynamicBlockers{
								&utils.DynamicBlocker{
									FilterIDs: []string{"*string:~*req.Account:1001"},
									Blocker:   false,
								},
							},
							Path:  "*req.Category",
							Type:  "*constant",
							Value: "call",
						},
					},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ATTR_1001;*string:~*req.Account:1001;*string:~*req.Account:1001&20;*string:~*req.Account:1001&true;*string:~*req.Account:1001;*string:~*req.Account:1001&false;*req.Category;*constant;call;key:value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestWithDynamicPaths",
			expApwo: &utils.APIAttributeProfileWithAPIOpts{
				APIAttributeProfile: &utils.APIAttributeProfile{
					Tenant:    "cgrates.org",
					ID:        "ATTR_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Weight:    20.0,
						},
					},
					Blockers: utils.DynamicBlockers{
						&utils.DynamicBlocker{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Blocker:   true,
						},
					},
					Attributes: []*utils.ExternalAttribute{
						{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Blockers: utils.DynamicBlockers{
								&utils.DynamicBlocker{
									FilterIDs: []string{"*string:~*req.Account:1001"},
									Blocker:   false,
								},
							},
							Path:  "*req.Category",
							Type:  "*constant",
							Value: "call",
						},
					},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ATTR_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&20;*string:~*req.Account:<~*req.Account>&true;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&false;*req.Category;*constant;call;key:value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestEmptyFields",
			expApwo: &utils.APIAttributeProfileWithAPIOpts{
				APIAttributeProfile: &utils.APIAttributeProfile{
					Tenant:     "cgrates.org",
					ID:         "ATTR_1001",
					FilterIDs:  nil,
					Weights:    nil,
					Blockers:   nil,
					Attributes: nil,
				},
				APIOpts: map[string]any{},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ATTR_1001;;;;;;;;;",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "WeightParamUnsupportedFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ATTR_1001;;w1&w2&w3;;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "WeightParamInvalidFloat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ATTR_1001;;w1&invalid;;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "BlockerParamUnsupportedFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ATTR_1001;;;b1&b2&b3;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "BlockerParamInvalidBool",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ATTR_1001;;;b1&invalid;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "APIOptsInvalidFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ATTR_1001;;;;;;;;;keyvalue",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "NoAdminSConns",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ATTR_1001;;;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultCGRConfig()
			cfg.GeneralCfg().DefaultCaching = utils.MetaNone

			if tc.name != "NoAdminSConns" {
				cfg.ActionSCfg().AdminSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS)}
			}

			connMgr := engine.NewConnManager(cfg)
			dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
			dm := engine.NewDataManager(dataDB, cfg, connMgr)
			fltrS := engine.NewFilterS(cfg, connMgr, dm)

			rpcInternal := make(chan birpc.ClientConnector, 1)
			rpcInternal <- ccM
			connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), utils.AdminSv1, rpcInternal)

			act := &actDynamicAttribute{
				config:  cfg,
				connMgr: connMgr,
				fltrS:   fltrS,
				tnt:     "cgrates.org",
				cgrEv: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evID",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				},
				aCfg: &utils.APAction{
					ID:      "ACT_DYNAMIC_ATTR",
					Diktats: tc.diktats,
				},
			}

			t.Cleanup(func() {
				apwo = nil
			})

			ctx := context.Background()
			data := utils.MapStorage{
				"*req": map[string]any{
					utils.AccountField: "1001",
				},
			}
			trgID := "trigger1"

			err := act.execute(ctx, data, trgID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected an error, but got nil")
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(apwo, tc.expApwo) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expApwo), utils.ToJSON(apwo))
			}
		})
	}
}

func TestActDynamicResourceId(t *testing.T) {
	testcases := []struct {
		name   string
		aCfg   *utils.APAction
		expect string
	}{
		{
			name: "WithValidID",
			aCfg: &utils.APAction{
				ID:        "ResA1",
				FilterIDs: []string{"f1", "f2"},
				TTL:       time.Second,
				Type:      "resource",
				Opts: map[string]any{
					"Opts": "Opts1",
				},
			},
			expect: "ResA1",
		},
		{
			name: "WithEmptyID",
			aCfg: &utils.APAction{
				ID:        "",
				FilterIDs: nil,
				TTL:       0,
				Type:      "resource",
				Opts:      nil,
			},
			expect: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			al := &actDynamicResource{
				aCfg: tc.aCfg,
			}
			got := al.id()
			if got != tc.expect {
				t.Errorf("expected %q, got %q", tc.expect, got)
			}
		})
	}
}

func TestActDynamicResourceCfg(t *testing.T) {
	testcases := []struct {
		name   string
		aCfg   *utils.APAction
		expect *utils.APAction
	}{
		{
			name: "WithValidCfg",
			aCfg: &utils.APAction{
				ID:        "ResCfg1",
				FilterIDs: []string{"f1"},
				TTL:       time.Second,
				Type:      "resource",
				Opts: map[string]any{
					"Opts1": "val1",
				},
			},
			expect: &utils.APAction{
				ID:        "ResCfg1",
				FilterIDs: []string{"f1"},
				TTL:       time.Second,
				Type:      "resource",
				Opts: map[string]any{
					"Opts1": "val1",
				},
			},
		},
		{
			name:   "WithNilCfg",
			aCfg:   nil,
			expect: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			al := &actDynamicResource{
				aCfg: tc.aCfg,
			}
			got := al.cfg()

			if !reflect.DeepEqual(got, tc.expect) {
				t.Errorf("expected %+v, got %+v", tc.expect, got)
			}
		})
	}
}

func TestActDynamicResourceExecute(t *testing.T) {
	engine.Cache.Clear(nil)
	defer engine.Cache.Clear(nil)

	var rpwo *utils.ResourceProfileWithAPIOpts
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AdminSv1SetResourceProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if rpwo, canCast = args.(*utils.ResourceProfileWithAPIOpts); !canCast {
					return fmt.Errorf("couldn't cast")
				}
				*reply.(*string) = utils.OK
				return nil
			},
		},
	}

	testcases := []struct {
		name    string
		diktats []*utils.APDiktat
		expRpwo *utils.ResourceProfileWithAPIOpts
		wantErr bool
	}{
		{
			name: "SuccessfulRequest",
			expRpwo: &utils.ResourceProfileWithAPIOpts{
				ResourceProfile: &utils.ResourceProfile{
					Tenant:            "cgrates.org",
					ID:                "RES_1001",
					FilterIDs:         []string{"*string:~*req.Account:1001"},
					UsageTTL:          time.Duration(60 * time.Second),
					Limit:             100.0,
					AllocationMessage: "Resource allocated",
					Blocker:           true,
					Stored:            true,
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Weight:    20.0,
						},
					},
					ThresholdIDs: []string{"THD_1001"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_1001;*string:~*req.Account:1001;*string:~*req.Account:1001&20;60s;100;Resource allocated;true;true;THD_1001;key:value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestWithDynamicPaths",
			expRpwo: &utils.ResourceProfileWithAPIOpts{
				ResourceProfile: &utils.ResourceProfile{
					Tenant:            "cgrates.org",
					ID:                "RES_1001",
					FilterIDs:         []string{"*string:~*req.Account:1001"},
					UsageTTL:          time.Duration(60 * time.Second),
					Limit:             100.0,
					AllocationMessage: "Resource allocated",
					Blocker:           true,
					Stored:            true,
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							FilterIDs: []string{"*string:~*req.Account:1001"},
							Weight:    20.0,
						},
					},
					ThresholdIDs: []string{"THD_1001"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_<~*req.Account>;*string:~*req.Account:<~*req.Account>;*string:~*req.Account:<~*req.Account>&20;60s;100;Resource allocated;true;true;THD_<~*req.Account>;key:value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestEmptyFields",
			expRpwo: &utils.ResourceProfileWithAPIOpts{
				ResourceProfile: &utils.ResourceProfile{
					Tenant:            "cgrates.org",
					ID:                "RES_1001",
					FilterIDs:         nil,
					UsageTTL:          0,
					Limit:             0.0,
					AllocationMessage: "",
					Blocker:           false,
					Stored:            false,
					Weights:           nil,
					ThresholdIDs:      nil,
				},
				APIOpts: map[string]any{},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_1001;;;;;;;;;",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "WeightParamUnsupportedFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_1001;;w1&w2&w3;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "WeightParamInvalidFloat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_1001;;w1&invalid;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "UsageTTLInvalidDuration",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_1001;;;invalid;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "LimitInvalidFloat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_1001;;;;invalid;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "BlockerInvalidBool",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_1001;;;;;;invalid;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "StoredInvalidBool",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_1001;;;;;;;invalid;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "APIOptsInvalidFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_1001;;;;;;;;;keyvalue",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidNumberOfParameters",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_1001;not_enough_params",
					},
				},
			},
			wantErr: true,
		},
		{
			name:    "NoDiktatsSpecified",
			diktats: []*utils.APDiktat{},
			wantErr: true,
		},
		{
			name: "NoAdminSConns",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RES_1001;;;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultCGRConfig()
			cfg.GeneralCfg().DefaultCaching = utils.MetaNone

			if tc.name != "NoAdminSConns" {
				cfg.ActionSCfg().AdminSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS)}
			}

			connMgr := engine.NewConnManager(cfg)
			dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
			dm := engine.NewDataManager(dataDB, cfg, connMgr)
			fltrS := engine.NewFilterS(cfg, connMgr, dm)

			rpcInternal := make(chan birpc.ClientConnector, 1)
			rpcInternal <- ccM
			connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), utils.AdminSv1, rpcInternal)

			act := &actDynamicResource{
				config:  cfg,
				connMgr: connMgr,
				fltrS:   fltrS,
				tnt:     "cgrates.org",
				cgrEv: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evID",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				},
				aCfg: &utils.APAction{
					ID:      "ACT_DYNAMIC_RESOURCE",
					Diktats: tc.diktats,
				},
			}

			t.Cleanup(func() {
				rpwo = nil
			})

			ctx := context.Background()
			data := utils.MapStorage{
				"*req": map[string]any{
					utils.AccountField: "1001",
				},
			}
			trgID := "trigger1"

			err := act.execute(ctx, data, trgID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected an error, but got nil")
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(rpwo, tc.expRpwo) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expRpwo), utils.ToJSON(rpwo))
			}
		})
	}
}

func TestActDynamicTrendId(t *testing.T) {
	tests := []struct {
		name     string
		actionID string
		expected string
	}{
		{
			name:     "With ID",
			actionID: "DynamicTrendID",
			expected: "DynamicTrendID",
		},
		{
			name:     "Empty ID",
			actionID: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actDT := &actDynamicTrend{
				aCfg: &utils.APAction{
					ID: tt.actionID,
				},
			}

			result := actDT.id()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestActDynamicTrendCfg(t *testing.T) {
	apAction := &utils.APAction{ID: "TrendProfile1"}

	a := &actDynamicTrend{
		aCfg: apAction,
	}

	if got := a.cfg(); !reflect.DeepEqual(got, apAction) {
		t.Errorf("cfg() = %v, want %v", got, apAction)
	}
}

func TestActDynamicTrendExecute(t *testing.T) {
	engine.Cache.Clear(nil)
	defer engine.Cache.Clear(nil)

	var tpwo *utils.TrendProfileWithAPIOpts
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AdminSv1SetTrendProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if tpwo, canCast = args.(*utils.TrendProfileWithAPIOpts); !canCast {
					return fmt.Errorf("couldn't cast")
				}
				*reply.(*string) = utils.OK
				return nil
			},
		},
	}

	testcases := []struct {
		name    string
		diktats []*utils.APDiktat
		expTpwo *utils.TrendProfileWithAPIOpts
		wantErr bool
	}{
		{
			name: "SuccessfulRequest",
			expTpwo: &utils.TrendProfileWithAPIOpts{
				TrendProfile: &utils.TrendProfile{
					Tenant:          "cgrates.org",
					ID:              "TREND_1001",
					Schedule:        "* * * * *",
					StatID:          "STAT_1001",
					Metrics:         []string{"*acd", "*tcd"},
					TTL:             60 * time.Second,
					QueueLength:     100,
					MinItems:        10,
					CorrelationType: "*last",
					Tolerance:       0.05,
					Stored:          true,
					ThresholdIDs:    []string{"THD_1001"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;TREND_1001;* * * * *;STAT_1001;*acd&*tcd;60s;100;10;*last;0.05;true;THD_1001;key:value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestWithDynamicPaths",
			expTpwo: &utils.TrendProfileWithAPIOpts{
				TrendProfile: &utils.TrendProfile{
					Tenant:          "cgrates.org",
					ID:              "TREND_1001",
					Schedule:        "* * * * *",
					StatID:          "STAT_1001",
					Metrics:         []string{"*acd", "*tcd"},
					TTL:             60 * time.Second,
					QueueLength:     100,
					MinItems:        10,
					CorrelationType: "*last",
					Tolerance:       0.05,
					Stored:          true,
					ThresholdIDs:    []string{"THD_1001"},
				},
				APIOpts: map[string]any{
					"key": "value",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;TREND_<~*req.Account>;* * * * *;STAT_<~*req.Account>;*acd&*tcd;60s;100;10;*last;0.05;true;THD_<~*req.Account>;key:value",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestEmptyFields",
			expTpwo: &utils.TrendProfileWithAPIOpts{
				TrendProfile: &utils.TrendProfile{
					Tenant:          "cgrates.org",
					ID:              "TREND_1001",
					Schedule:        "",
					StatID:          "",
					Metrics:         nil,
					TTL:             0,
					QueueLength:     0,
					MinItems:        0,
					CorrelationType: "",
					Tolerance:       0.0,
					Stored:          false,
					ThresholdIDs:    nil,
				},
				APIOpts: map[string]any{},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;TREND_1001;;;;;;;;;;;",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "TTLInvalidDuration",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;TREND_1001;;;invalid;;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "QueueLengthInvalidInt",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;TREND_1001;;;;;invalid;;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "MinItemsInvalidInt",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;TREND_1001;;;;;;invalid;;;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "ToleranceInvalidFloat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;TREND_1001;;;;;;;;invalid;;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "StoredInvalidBool",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;TREND_1001;;;;;;;;;invalid;",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidNumberOfParameters",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;TREND_1001;not_enough_params",
					},
				},
			},
			wantErr: true,
		},
		{
			name:    "NoDiktatsSpecified",
			diktats: []*utils.APDiktat{},
			wantErr: true,
		},
		{
			name: "NoAdminSConns",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;TREND_1001;;;;;;;;;;;",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultCGRConfig()
			cfg.GeneralCfg().DefaultCaching = utils.MetaNone

			if tc.name != "NoAdminSConns" {
				cfg.ActionSCfg().AdminSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS)}
			}

			connMgr := engine.NewConnManager(cfg)
			dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
			dm := engine.NewDataManager(dataDB, cfg, connMgr)
			fltrS := engine.NewFilterS(cfg, connMgr, dm)

			rpcInternal := make(chan birpc.ClientConnector, 1)
			rpcInternal <- ccM
			connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), utils.AdminSv1, rpcInternal)

			act := &actDynamicTrend{
				config:  cfg,
				connMgr: connMgr,
				fltrS:   fltrS,
				tnt:     "cgrates.org",
				cgrEv: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evID",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				},
				aCfg: &utils.APAction{
					ID:      "ACT_DYNAMIC_TREND",
					Diktats: tc.diktats,
				},
			}

			t.Cleanup(func() {
				tpwo = nil
			})

			ctx := context.Background()
			data := utils.MapStorage{
				"*req": map[string]any{
					utils.AccountField: "1001",
				},
			}
			trgID := "trigger1"

			err := act.execute(ctx, data, trgID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected an error, but got nil")
				}
			} else if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(tpwo, tc.expTpwo) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expTpwo), utils.ToJSON(tpwo))
			}
		})
	}
}

func TestActDynamicRankingId(t *testing.T) {
	tests := []struct {
		name string
		dr   actDynamicRanking
		want string
	}{
		{
			name: "WithValidID",
			dr:   actDynamicRanking{aCfg: &utils.APAction{ID: "RankingProfile1"}},
			want: "RankingProfile1",
		},
		{
			name: "WithEmptyID",
			dr:   actDynamicRanking{aCfg: &utils.APAction{ID: ""}},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dr.id()
			if got != tt.want {
				t.Errorf("id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActDynamicRankingCfg(t *testing.T) {
	tests := []struct {
		name string
		al   actDynamicRanking
		want *utils.APAction
	}{
		{
			name: "WithValidCfg",
			al:   actDynamicRanking{aCfg: &utils.APAction{ID: "RankingProfile1"}},
			want: &utils.APAction{ID: "RankingProfile1"},
		},
		{
			name: "WithNilCfg",
			al:   actDynamicRanking{aCfg: nil},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.al.cfg()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cfg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActDynamicRankingExecute(t *testing.T) {
	engine.Cache.Clear(nil)
	defer engine.Cache.Clear(nil)

	var rpwo *utils.RankingProfileWithAPIOpts
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AdminSv1SetRankingProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if rpwo, canCast = args.(*utils.RankingProfileWithAPIOpts); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				*(reply.(*string)) = utils.OK
				return nil
			},
		},
	}

	testcases := []struct {
		name    string
		diktats []*utils.APDiktat
		expRpwo *utils.RankingProfileWithAPIOpts
		wantErr bool
	}{
		{
			name: "SuccessfulRequest",
			expRpwo: &utils.RankingProfileWithAPIOpts{
				RankingProfile: &utils.RankingProfile{
					Tenant:            "cgrates.org",
					ID:                "RANK_ACNT_1001",
					Schedule:          "* * * * *",
					Sorting:           "*desc",
					SortingParameters: []string{"*tcc", "*tcd"},
					Stored:            true,
					StatIDs:           []string{"STAT_1", "STAT_2"},
					MetricIDs:         []string{"*tcc", "*tcd"},
					ThresholdIDs:      []string{"THD_1", "THD_2"},
				},
				APIOpts: map[string]any{
					"key1": "value1",
					"key2": "value2",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RANK_ACNT_1001;* * * * *;STAT_1&STAT_2;*tcc&*tcd;*desc;*tcc&*tcd;true;THD_1&THD_2;key1:value1&key2:value2",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestWithDynamicPaths",
			expRpwo: &utils.RankingProfileWithAPIOpts{
				RankingProfile: &utils.RankingProfile{
					Tenant:            "cgrates.org",
					ID:                "RANK_ACNT_1001",
					Schedule:          "0 0 * * *",
					Sorting:           "*asc",
					SortingParameters: []string{"*acc"},
					Stored:            false,
					StatIDs:           []string{"STAT_1001"},
					MetricIDs:         []string{"*acc"},
					ThresholdIDs:      []string{"THD_1001"},
				},
				APIOpts: map[string]any{
					"account": "1001",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 15.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RANK_ACNT_<~*req.Account>;0 0 * * *;STAT_<~*req.Account>;*acc;*asc;*acc;false;THD_<~*req.Account>;account:<~*req.Account>",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestEmptyFields",
			expRpwo: &utils.RankingProfileWithAPIOpts{
				RankingProfile: &utils.RankingProfile{
					Tenant:            "cgrates.org",
					ID:                "RANK_EMPTY",
					Schedule:          "",
					Sorting:           "",
					SortingParameters: nil,
					Stored:            false,
					StatIDs:           nil,
					MetricIDs:         nil,
					ThresholdIDs:      nil,
				},
				APIOpts: map[string]any{},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 10.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RANK_EMPTY;;;;;;;;",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "EmptyDiktats",
			diktats: []*utils.APDiktat{},
			wantErr: true,
		},
		{
			name: "InvalidNumberOfParams",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RANK_INVALID;only;four;params",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidStoredParam",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RANK_INVALID;* * * * *;STAT_1;*tcc;*desc;*tcc;invalid_bool;THD_1;key:value",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "APIOptsInvalidFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;RANK_INVALID;* * * * *;STAT_1;*tcc;*desc;*tcc;true;THD_1;invalidformat",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultCGRConfig()
			cfg.GeneralCfg().DefaultCaching = utils.MetaNone
			cfg.ActionSCfg().AdminSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS)}
			connMgr := engine.NewConnManager(cfg)
			dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
			dm := engine.NewDataManager(dataDB, cfg, connMgr)
			fltrS := engine.NewFilterS(cfg, connMgr, dm)
			rpcInternal := make(chan birpc.ClientConnector, 1)
			rpcInternal <- ccM
			connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), utils.AdminSv1, rpcInternal)

			act := &actDynamicRanking{
				config:  cfg,
				connMgr: connMgr,
				fltrS:   fltrS,
				tnt:     "cgrates.org",
				cgrEv: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evID",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				},
				aCfg: &utils.APAction{
					ID:      "ACT_DYNAMIC_RANKING",
					Diktats: tc.diktats,
				},
			}

			t.Cleanup(func() {
				rpwo = nil
			})

			ctx := context.Background()
			data := utils.MapStorage{
				"*req": map[string]any{
					utils.AccountField: "1001",
				},
			}
			trgID := "trigger1"
			err := act.execute(ctx, data, trgID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected an error, but got nil")
				}
			} else if err != nil {
				t.Error(err)
			} else if tc.expRpwo != nil && !reflect.DeepEqual(rpwo, tc.expRpwo) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expRpwo), utils.ToJSON(rpwo))
			}
		})
	}
}

func TestActDynamicFilterId(t *testing.T) {
	tests := []struct {
		name string
		df   actDynamicFilter
		want string
	}{
		{
			name: "WithValidID",
			df:   actDynamicFilter{aCfg: &utils.APAction{ID: "FilterProfile1"}},
			want: "FilterProfile1",
		},
		{
			name: "WithEmptyID",
			df:   actDynamicFilter{aCfg: &utils.APAction{ID: ""}},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.df.id()
			if got != tt.want {
				t.Errorf("id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActDynamicFilterCfg(t *testing.T) {
	tests := []struct {
		name string
		df   actDynamicFilter
		want *utils.APAction
	}{
		{
			name: "WithValidCfg",
			df:   actDynamicFilter{aCfg: &utils.APAction{ID: "FilterProfile1"}},
			want: &utils.APAction{ID: "FilterProfile1"},
		},
		{
			name: "WithNilCfg",
			df:   actDynamicFilter{aCfg: nil},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.df.cfg()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cfg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActDynamicFilterExecute(t *testing.T) {
	engine.Cache.Clear(nil)
	defer engine.Cache.Clear(nil)

	var fwo *engine.FilterWithAPIOpts
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AdminSv1SetFilter: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if fwo, canCast = args.(*engine.FilterWithAPIOpts); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				*(reply.(*string)) = utils.OK
				return nil
			},
		},
	}

	testcases := []struct {
		name    string
		diktats []*utils.APDiktat
		expFwo  *engine.FilterWithAPIOpts
		wantErr bool
	}{
		{
			name: "SuccessfulRequest",
			expFwo: &engine.FilterWithAPIOpts{
				Filter: &engine.Filter{
					Tenant: "cgrates.org",
					ID:     "FLTR_ACNT_1001",
					Rules: []*engine.FilterRule{{
						Type:    "*string",
						Element: "~*req.Account",
						Values:  []string{"1001", "1002"},
					}},
				},
				APIOpts: map[string]any{
					utils.MetaSubsys: "*sessions",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;FLTR_ACNT_1001;*string;~*req.Account;1001&1002;*subsys:*sessions",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestWithDynamicPaths",
			expFwo: &engine.FilterWithAPIOpts{
				Filter: &engine.Filter{
					Tenant: "cgrates.org",
					ID:     "FLTR_ACNT_1001",
					Rules: []*engine.FilterRule{{
						Type:    "*string",
						Element: "~*req.Account",
						Values:  []string{"1001"},
					}},
				},
				APIOpts: map[string]any{
					"account": "1001",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 15.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;FLTR_ACNT_<~*req.Account>;*string;~*req.Account;<~*req.Account>;account:<~*req.Account>",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestEmptyFields",
			expFwo: &engine.FilterWithAPIOpts{
				Filter: &engine.Filter{
					Tenant: "cgrates.org",
					ID:     "FLTR_EMPTY",
					Rules: []*engine.FilterRule{{
						Type:    "*string",
						Element: "~*req.Field",
						Values:  nil,
					}},
				},
				APIOpts: map[string]any{},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 10.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;FLTR_EMPTY;*string;~*req.Field;;",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "EmptyDiktats",
			diktats: []*utils.APDiktat{},
			wantErr: true,
		},
		{
			name: "InvalidNumberOfParams",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;FLTR_INVALID;only;four;params",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "APIOptsInvalidFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;FLTR_INVALID;*string;~*req.Field;value;invalidformat",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultCGRConfig()
			cfg.GeneralCfg().DefaultCaching = utils.MetaNone
			cfg.ActionSCfg().AdminSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS)}
			connMgr := engine.NewConnManager(cfg)
			dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
			dm := engine.NewDataManager(dataDB, cfg, connMgr)
			fltrS := engine.NewFilterS(cfg, connMgr, dm)
			rpcInternal := make(chan birpc.ClientConnector, 1)
			rpcInternal <- ccM
			connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), utils.AdminSv1, rpcInternal)

			act := &actDynamicFilter{
				config:  cfg,
				connMgr: connMgr,
				fltrS:   fltrS,
				tnt:     "cgrates.org",
				cgrEv: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evID",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				},
				aCfg: &utils.APAction{
					ID:      "ACT_DYNAMIC_FILTER",
					Diktats: tc.diktats,
				},
			}

			t.Cleanup(func() {
				fwo = nil
			})

			ctx := context.Background()
			data := utils.MapStorage{
				"*req": map[string]any{
					utils.AccountField: "1001",
				},
			}
			trgID := "trigger1"
			err := act.execute(ctx, data, trgID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected an error, but got nil")
				}
			} else if err != nil {
				t.Error(err)
			} else if tc.expFwo != nil && !reflect.DeepEqual(fwo, tc.expFwo) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expFwo), utils.ToJSON(fwo))
			}
		})
	}
}

func TestActDynamicRouteId(t *testing.T) {
	tests := []struct {
		name string
		aCfg *utils.APAction
		want string
	}{
		{
			name: "WithValidID",
			aCfg: &utils.APAction{ID: "RouteID1"},
			want: "RouteID1",
		},
		{
			name: "WithEmptyID",
			aCfg: &utils.APAction{ID: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aL := &actDynamicRoute{aCfg: tt.aCfg}
			if got := aL.id(); got != tt.want {
				t.Errorf("id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActDynamicRouteCfg(t *testing.T) {
	tests := []struct {
		name string
		dr   actDynamicRoute
		want *utils.APAction
	}{
		{
			name: "WithValidCfg",
			dr:   actDynamicRoute{aCfg: &utils.APAction{ID: "RouteProfile1"}},
			want: &utils.APAction{ID: "RouteProfile1"},
		},
		{
			name: "WithNilCfg",
			dr:   actDynamicRoute{aCfg: nil},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dr.cfg()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cfg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActDynamicRouteExecute(t *testing.T) {
	engine.Cache.Clear(nil)
	defer engine.Cache.Clear(nil)

	var rpo *utils.RouteProfileWithAPIOpts
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AdminSv1SetRouteProfile: func(ctx *context.Context, args, reply any) error {
				var canCast bool
				if rpo, canCast = args.(*utils.RouteProfileWithAPIOpts); !canCast {
					return fmt.Errorf("couldnt cast")
				}
				*(reply.(*string)) = utils.OK
				return nil
			},
		},
	}

	testcases := []struct {
		name    string
		diktats []*utils.APDiktat
		expRpo  *utils.RouteProfileWithAPIOpts
		wantErr bool
	}{
		{
			name: "SuccessfulRequest",
			expRpo: &utils.RouteProfileWithAPIOpts{
				RouteProfile: &utils.RouteProfile{
					Tenant:    "cgrates.org",
					ID:        "ROUTE_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							FilterIDs: []string{"*weight"},
							Weight:    10.0,
						},
					},
					Blockers: utils.DynamicBlockers{
						&utils.DynamicBlocker{
							FilterIDs: []string{"*blocker"},
							Blocker:   false,
						},
					},
					Sorting:           "*weight",
					SortingParameters: []string{"*weight", "*cost"},
					Routes: []*utils.Route{
						{
							ID:             "RT_1001",
							FilterIDs:      []string{"*string:~*req.Route:1001"},
							AccountIDs:     []string{"1001", "1002"},
							RateProfileIDs: []string{"RP_1001"},
							ResourceIDs:    []string{"RES_1001"},
							StatIDs:        []string{"STAT_1001"},
							Weights: utils.DynamicWeights{
								&utils.DynamicWeight{
									FilterIDs: []string{"*route_weight"},
									Weight:    20.0,
								},
							},
							Blockers: utils.DynamicBlockers{
								&utils.DynamicBlocker{
									FilterIDs: []string{"*route_blocker"},
									Blocker:   false,
								},
							},
							RouteParameters: "param1=value1",
						},
					},
				},
				APIOpts: map[string]any{
					utils.MetaSubsys: "*sessions",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ROUTE_1001;*string:~*req.Account:1001;*weight&10.0;*blocker&false;*weight;*weight&*cost;RT_1001;*string:~*req.Route:1001;1001&1002;RP_1001;RES_1001;STAT_1001;*route_weight&20.0;*route_blocker&false;param1=value1;*subsys:*sessions",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestWithDynamicPaths",
			expRpo: &utils.RouteProfileWithAPIOpts{
				RouteProfile: &utils.RouteProfile{
					Tenant:  "cgrates.org",
					ID:      "ROUTE_1001",
					Sorting: "*weight",
					Routes: []*utils.Route{
						{
							ID:              "RT_1001",
							AccountIDs:      []string{"1001"},
							RouteParameters: "1001",
						},
					},
				},
				APIOpts: map[string]any{
					"account": "1001",
				},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 15.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ROUTE_<~*req.Account>;;;;*weight;;RT_<~*req.Account>;;1001;;;;;;<~*req.Account>;account:<~*req.Account>",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessfulRequestEmptyFields",
			expRpo: &utils.RouteProfileWithAPIOpts{
				RouteProfile: &utils.RouteProfile{
					Tenant:  "cgrates.org",
					ID:      "ROUTE_EMPTY",
					Sorting: "",
				},
				APIOpts: map[string]any{},
			},
			diktats: []*utils.APDiktat{
				{
					ID:        "d1",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 10.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ROUTE_EMPTY;;;;;;;;;;;;;;;",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "EmptyDiktats",
			diktats: []*utils.APDiktat{},
			wantErr: true,
		},
		{
			name: "InvalidNumberOfParams",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ROUTE_INVALID;only;four;params",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "APIOptsInvalidFormat",
			diktats: []*utils.APDiktat{
				{
					ID: "d1",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{Weight: 20.0},
					},
					Opts: map[string]any{
						utils.MetaTemplate: "cgrates.org;ROUTE_INVALID;;;;;;;;;;;;;;;;invalidformat",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewDefaultCGRConfig()
			cfg.GeneralCfg().DefaultCaching = utils.MetaNone
			cfg.ActionSCfg().AdminSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS)}
			connMgr := engine.NewConnManager(cfg)
			dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
			dm := engine.NewDataManager(dataDB, cfg, connMgr)
			fltrS := engine.NewFilterS(cfg, connMgr, dm)
			rpcInternal := make(chan birpc.ClientConnector, 1)
			rpcInternal <- ccM
			connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), utils.AdminSv1, rpcInternal)

			act := &actDynamicRoute{
				config:  cfg,
				connMgr: connMgr,
				dm:      dm,
				fltrS:   fltrS,
				tnt:     "cgrates.org",
				cgrEv: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evID",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				},
				aCfg: &utils.APAction{
					ID:      "ACT_DYNAMIC_ROUTE",
					Diktats: tc.diktats,
				},
			}

			t.Cleanup(func() {
				rpo = nil
			})

			ctx := context.Background()
			data := utils.MapStorage{
				"*req": map[string]any{
					utils.AccountField: "1001",
				},
			}
			trgID := "trigger1"
			err := act.execute(ctx, data, trgID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected an error, but got nil")
				}
			} else if err != nil {
				t.Error(err)
			} else if tc.expRpo != nil && !reflect.DeepEqual(rpo, tc.expRpo) {
				t.Errorf("Expected <%v>\nReceived\n<%v>", utils.ToJSON(tc.expRpo), utils.ToJSON(rpo))
			}
		})
	}
}

func TestActDynamicRateId(t *testing.T) {
	tests := []struct {
		name string
		aL   actDynamicRate
		want string
	}{
		{
			name: "WithValidID",
			aL:   actDynamicRate{aCfg: &utils.APAction{ID: "RateProfile1"}},
			want: "RateProfile1",
		},
		{
			name: "WithEmptyID",
			aL:   actDynamicRate{aCfg: &utils.APAction{ID: ""}},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.aL.id(); got != tt.want {
				t.Errorf("actDynamicRate.id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActDynamicRateCfg(t *testing.T) {
	tests := []struct {
		name string
		aL   actDynamicRate
		want *utils.APAction
	}{
		{
			name: "WithValidCfg",
			aL:   actDynamicRate{aCfg: &utils.APAction{ID: "RateProfile1"}},
			want: &utils.APAction{ID: "RateProfile1"},
		},
		{
			name: "WithEmptyCfg",
			aL:   actDynamicRate{aCfg: &utils.APAction{}},
			want: &utils.APAction{},
		},
		{
			name: "WithNilCfg",
			aL:   actDynamicRate{aCfg: nil},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.aL.cfg(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("actDynamicRate.cfg() = %v, want %v", got, tt.want)
			}
		})
	}
}
