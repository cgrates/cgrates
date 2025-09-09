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
	"reflect"
	"slices"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

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
