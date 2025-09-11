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
