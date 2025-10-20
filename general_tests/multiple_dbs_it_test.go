//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package general_tests

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

func TestMultipleDBs(t *testing.T) {
	if err := os.MkdirAll("/tmp/internal_db/db", 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll("/tmp/internal_db"); err != nil {
			t.Error(err)
		}
	})
	ng := engine.TestEngine{
		ConfigPath:       filepath.Join(*utils.DataDir, "conf", "samples", "multiple_dbs"),
		GracefulShutdown: true,
		Encoding:         *utils.Encoding,
	}
	client, cfg := ng.Run(t)
	t.Run("LoadTariffPlans", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.LoaderSv1Run,
			&loaders.ArgsProcessFolder{
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned:", reply)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("CheckChargers", func(t *testing.T) { // stored in redis2
		var chrgrs []*utils.ChargerProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetChargerProfiles,
			&utils.ArgsItemIDs{
				Tenant: "cgrates.org",
			}, &chrgrs); err != nil {
			t.Errorf("AdminSv1GetChargerProfiles failed unexpectedly: %v", err)
		}
		if len(chrgrs) != 3 {
			t.Fatalf("AdminSv1GetChargerProfiles len(chrgrs)=%v, want 3", len(chrgrs))
		}
		sort.Slice(chrgrs, func(i, j int) bool {
			return chrgrs[i].ID > chrgrs[j].ID
		})
		exp := []*utils.ChargerProfile{
			{
				ID:     "SupplierCharges",
				Tenant: "cgrates.org",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				RunID:        "SupplierCharges",
				AttributeIDs: []string{"ATTR_SUPPLIER1"},
			},
			{
				ID:     "Raw",
				Tenant: "cgrates.org",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				RunID:        "raw",
				AttributeIDs: []string{"*constant:*req.RequestType:*none"},
			},
			{
				ID:     "CustomerCharges",
				Tenant: "cgrates.org",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				RunID:        "CustomerCharges",
				AttributeIDs: []string{"*none"},
			},
		}
		if !reflect.DeepEqual(exp, chrgrs) {
			t.Errorf("Expected <%+v>,\n received <%+v>", exp, chrgrs)
		}
	})

	t.Run("CheckChargerFilterIndexes", func(t *testing.T) { // stored in internal
		var replyIdx []string
		expectedIDx := []string{"*none:*any:*any:CustomerCharges", "*none:*any:*any:Raw",
			"*none:*any:*any:SupplierCharges"}
		if err := client.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
			&apis.AttrGetFilterIndexes{
				Tenant:   utils.CGRateSorg,
				ItemType: utils.MetaChargers,
			},
			&replyIdx); err != nil {
			t.Error(err)
		} else {
			sort.Strings(replyIdx)
			sort.Strings(expectedIDx)
			if !reflect.DeepEqual(expectedIDx, replyIdx) {
				t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
			}
		}
	})

	t.Run("CheckAccounts", func(t *testing.T) { // stored in *default (redis in this case)
		var acnts []*utils.Account
		if err := client.Call(context.Background(), utils.AdminSv1GetAccounts,
			&utils.ArgsItemIDs{
				Tenant: "cgrates.org",
			}, &acnts); err != nil {
			t.Errorf("AdminSv2GetAccounts failed unexpectedly: %v", err)
		}
		if len(acnts) != 2 {
			t.Fatalf("AdminSv2GetAccounts len(acnts)=%v, want 2", len(acnts))
		}
		sort.Slice(acnts, func(i, j int) bool {
			return acnts[i].ID > acnts[j].ID
		})
		exp := []*utils.Account{
			{
				Tenant: "cgrates.org",
				ID:     "ACC_PRF_1",
				Opts:   map[string]any{},
				Balances: map[string]*utils.Balance{
					"MonetaryBalance": {
						ID: "MonetaryBalance",
						Weights: utils.DynamicWeights{
							{
								Weight: 10,
							},
						},
						Type:  "*monetary",
						Opts:  map[string]any{},
						Units: utils.NewDecimal(14, 0),
						UnitFactors: []*utils.UnitFactor{
							{
								FilterIDs: []string{"fltr1", "fltr2"},
								Factor:    utils.NewDecimal(100, 0),
							},
							{
								FilterIDs: []string{"fltr3"},
								Factor:    utils.NewDecimal(200, 0),
							},
						},
						CostIncrements: []*utils.CostIncrement{
							{
								FilterIDs:    []string{"fltr1", "fltr2"},
								Increment:    utils.NewDecimal(13, 1),
								FixedFee:     utils.NewDecimal(23, 1),
								RecurrentFee: utils.NewDecimal(33, 1),
							},
						},
						AttributeIDs: []string{"attr1", "attr2"},
					},
				},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				ThresholdIDs: []string{utils.MetaNone},
			},
			{
				Tenant: "cgrates.org",
				ID:     "1001",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Opts: map[string]any{},
				Balances: map[string]*utils.Balance{
					"MonetaryBalance": {
						ID: "MonetaryBalance",
						Weights: utils.DynamicWeights{
							{
								Weight: 10,
							},
						},
						Type:  "*monetary",
						Opts:  map[string]any{},
						Units: utils.NewDecimal(14, 0),
						UnitFactors: []*utils.UnitFactor{
							{
								FilterIDs: []string{"fltr1", "fltr2"},
								Factor:    utils.NewDecimal(100, 0),
							},
							{
								FilterIDs: []string{"fltr3"},
								Factor:    utils.NewDecimal(200, 0),
							},
						},
						CostIncrements: []*utils.CostIncrement{
							{
								FilterIDs:    []string{"fltr1", "fltr2"},
								Increment:    utils.NewDecimal(13, 1),
								FixedFee:     utils.NewDecimal(23, 1),
								RecurrentFee: utils.NewDecimal(33, 1),
							},
						},
						AttributeIDs: []string{"attr1", "attr2"},
					},
					"VoiceBalance": {
						ID: "VoiceBalance",
						Weights: utils.DynamicWeights{
							{
								Weight: 10,
							},
							{
								Weight: 10,
							},
						},
						Blockers: utils.DynamicBlockers{
							{
								FilterIDs: []string{"*string:~*req.Destination:1002"},
								Blocker:   true,
							},
							{
								Blocker: false,
							},
						},
						Type:  "*voice",
						Opts:  map[string]any{},
						Units: utils.NewDecimalFromUsageIgnoreErr("1h"),
					},
				},
				ThresholdIDs: []string{utils.MetaNone},
			},
		}
		if !reflect.DeepEqual(exp, acnts) {
			t.Errorf("Expected <%+v>,\n received <%+v>", exp, acnts)
		}
	})

	t.Run("CheckCdrs", func(t *testing.T) { // stored in mysql
		var cdrs []*utils.CDR
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs, &utils.CDRFilters{Tenant: "cgrates.org"}, &cdrs); err == nil || err.Error() != "retrieving CDRs failed: NOT_FOUND" {
			t.Errorf("Expecting error <%v>, received: <%v>", "retrieving CDRs failed: NOT_FOUND", err)
		}
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestEv1",
			Event: map[string]any{
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestEv1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.Usage:        time.Minute,
			},
			APIOpts: map[string]any{
				utils.MetaRates:    true,
				utils.MetaAccounts: false,
			},
		}
		var rply string
		client.Call(context.Background(), utils.CDRsV1ProcessEvent, ev, &rply)
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs, &utils.CDRFilters{Tenant: "cgrates.org"}, &cdrs); err != nil {
			t.Error(err)
		}
		if len(cdrs) != 1 {
			t.Errorf("unexpected number of cdrs found: %v", len(cdrs))
		}
		exp := &utils.CDR{
			Tenant: utils.CGRateSorg,
			Opts: map[string]any{
				utils.MetaCDRID:    cdrs[0].Opts[utils.MetaCDRID],
				utils.MetaRates:    true,
				utils.MetaAccounts: false,
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.OriginID:     "TestEv1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.Subject:      "1001",
				utils.ToR:          utils.MetaVoice,
				utils.Usage:        6e+10,
			},
			CreatedAt: cdrs[0].CreatedAt,
			UpdatedAt: cdrs[0].UpdatedAt,
		}
		if !reflect.DeepEqual(exp, cdrs[0]) {
			t.Errorf("Expecting <%#v>, \nreceived <%#v>", exp, cdrs[0])
		}
	})
	t.Run("EngineShutdown", func(t *testing.T) {
		if err := engine.KillEngine(100); err != nil {
			t.Error(err)
		}
	})

	t.Run("CountDBFiles", func(t *testing.T) {
		var dirs, files int
		if err := filepath.WalkDir(cfg.DbCfg().Opts.InternalDBDumpPath, func(root string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				dirs++
			} else {
				if !strings.HasPrefix(root, "/tmp/internal_db/db/*charger_filter_indexes") &&
					!strings.HasPrefix(root, "/tmp/internal_db/db/*versions") {
					t.Fatalf("got unexpected folder <%s>", root)
				}
				files++
			}
			return nil
		}); err != nil {
			t.Error(err)
		} else if dirs != 37 {
			t.Errorf("expected <%d> directories, received <%d>", 37, dirs)
		} else if files != 2 {
			t.Errorf("expected <%d> files, received <%d>", 2, files)
		}
	})
}

func TestMultipleDBsMongo(t *testing.T) {
	if err := os.MkdirAll("/tmp/internal_db/db", 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll("/tmp/internal_db"); err != nil {
			t.Error(err)
		}
	})
	ng := engine.TestEngine{
		ConfigPath:       filepath.Join(*utils.DataDir, "conf", "samples", "multiple_dbs_mongo"),
		GracefulShutdown: true,
		Encoding:         *utils.Encoding,
	}
	client, cfg := ng.Run(t)
	t.Run("LoadTariffPlans", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.LoaderSv1Run,
			&loaders.ArgsProcessFolder{
				APIOpts: map[string]any{
					utils.MetaCache: utils.MetaNone,
				},
			}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned:", reply)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("CheckChargers", func(t *testing.T) { // stored in mongo
		var chrgrs []*utils.ChargerProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetChargerProfiles,
			&utils.ArgsItemIDs{
				Tenant: "cgrates.org",
			}, &chrgrs); err != nil {
			t.Errorf("AdminSv1GetChargerProfiles failed unexpectedly: %v", err)
		}
		if len(chrgrs) != 3 {
			t.Fatalf("AdminSv1GetChargerProfiles len(chrgrs)=%v, want 3", len(chrgrs))
		}
		sort.Slice(chrgrs, func(i, j int) bool {
			return chrgrs[i].ID > chrgrs[j].ID
		})
		exp := []*utils.ChargerProfile{
			{
				ID:     "SupplierCharges",
				Tenant: "cgrates.org",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				RunID:        "SupplierCharges",
				AttributeIDs: []string{"ATTR_SUPPLIER1"},
			},
			{
				ID:     "Raw",
				Tenant: "cgrates.org",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				RunID:        "raw",
				AttributeIDs: []string{"*constant:*req.RequestType:*none"},
			},
			{
				ID:     "CustomerCharges",
				Tenant: "cgrates.org",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				RunID:        "CustomerCharges",
				AttributeIDs: []string{"*none"},
			},
		}
		if !reflect.DeepEqual(exp, chrgrs) {
			t.Errorf("Expected <%+v>,\n received <%+v>", exp, chrgrs)
		}
	})

	t.Run("CheckChargerFilterIndexes", func(t *testing.T) { // stored in internal
		var replyIdx []string
		expectedIDx := []string{"*none:*any:*any:CustomerCharges", "*none:*any:*any:Raw",
			"*none:*any:*any:SupplierCharges"}
		if err := client.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
			&apis.AttrGetFilterIndexes{
				Tenant:   utils.CGRateSorg,
				ItemType: utils.MetaChargers,
			},
			&replyIdx); err != nil {
			t.Error(err)
		} else {
			sort.Strings(replyIdx)
			sort.Strings(expectedIDx)
			if !reflect.DeepEqual(expectedIDx, replyIdx) {
				t.Errorf("Expected %+v, \nreceived %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
			}
		}
	})

	t.Run("CheckAccounts", func(t *testing.T) { // stored in *default (redis in this case)
		var acnts []*utils.Account
		if err := client.Call(context.Background(), utils.AdminSv1GetAccounts,
			&utils.ArgsItemIDs{
				Tenant: "cgrates.org",
			}, &acnts); err != nil {
			t.Errorf("AdminSv2GetAccounts failed unexpectedly: %v", err)
		}
		if len(acnts) != 2 {
			t.Fatalf("AdminSv2GetAccounts len(acnts)=%v, want 2", len(acnts))
		}
		sort.Slice(acnts, func(i, j int) bool {
			return acnts[i].ID > acnts[j].ID
		})
		exp := []*utils.Account{
			{
				Tenant: "cgrates.org",
				ID:     "ACC_PRF_1",
				Opts:   map[string]any{},
				Balances: map[string]*utils.Balance{
					"MonetaryBalance": {
						ID: "MonetaryBalance",
						Weights: utils.DynamicWeights{
							{
								Weight: 10,
							},
						},
						Type:  "*monetary",
						Opts:  map[string]any{},
						Units: utils.NewDecimal(14, 0),
						UnitFactors: []*utils.UnitFactor{
							{
								FilterIDs: []string{"fltr1", "fltr2"},
								Factor:    utils.NewDecimal(100, 0),
							},
							{
								FilterIDs: []string{"fltr3"},
								Factor:    utils.NewDecimal(200, 0),
							},
						},
						CostIncrements: []*utils.CostIncrement{
							{
								FilterIDs:    []string{"fltr1", "fltr2"},
								Increment:    utils.NewDecimal(13, 1),
								FixedFee:     utils.NewDecimal(23, 1),
								RecurrentFee: utils.NewDecimal(33, 1),
							},
						},
						AttributeIDs: []string{"attr1", "attr2"},
					},
				},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				ThresholdIDs: []string{utils.MetaNone},
			},
			{
				Tenant: "cgrates.org",
				ID:     "1001",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Opts: map[string]any{},
				Balances: map[string]*utils.Balance{
					"MonetaryBalance": {
						ID: "MonetaryBalance",
						Weights: utils.DynamicWeights{
							{
								Weight: 10,
							},
						},
						Type:  "*monetary",
						Opts:  map[string]any{},
						Units: utils.NewDecimal(14, 0),
						UnitFactors: []*utils.UnitFactor{
							{
								FilterIDs: []string{"fltr1", "fltr2"},
								Factor:    utils.NewDecimal(100, 0),
							},
							{
								FilterIDs: []string{"fltr3"},
								Factor:    utils.NewDecimal(200, 0),
							},
						},
						CostIncrements: []*utils.CostIncrement{
							{
								FilterIDs:    []string{"fltr1", "fltr2"},
								Increment:    utils.NewDecimal(13, 1),
								FixedFee:     utils.NewDecimal(23, 1),
								RecurrentFee: utils.NewDecimal(33, 1),
							},
						},
						AttributeIDs: []string{"attr1", "attr2"},
					},
					"VoiceBalance": {
						ID: "VoiceBalance",
						Weights: utils.DynamicWeights{
							{
								Weight: 10,
							},
							{
								Weight: 10,
							},
						},
						Blockers: utils.DynamicBlockers{
							{
								FilterIDs: []string{"*string:~*req.Destination:1002"},
								Blocker:   true,
							},
							{
								Blocker: false,
							},
						},
						Type:  "*voice",
						Opts:  map[string]any{},
						Units: utils.NewDecimalFromUsageIgnoreErr("1h"),
					},
				},
				ThresholdIDs: []string{utils.MetaNone},
			},
		}
		if !reflect.DeepEqual(exp, acnts) {
			t.Errorf("Expected <%+v>,\n received <%+v>", exp, acnts)
		}
	})

	t.Run("CheckCdrs", func(t *testing.T) { // stored in mongo
		var cdrs []*utils.CDR
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs, &utils.CDRFilters{Tenant: "cgrates.org"}, &cdrs); err == nil || err.Error() != "retrieving CDRs failed: NOT_FOUND" {
			t.Errorf("Expecting error <%v>, received: <%v>", "retrieving CDRs failed: NOT_FOUND", err)
		}
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestEv1",
			Event: map[string]any{
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestEv1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.Usage:        time.Minute,
			},
			APIOpts: map[string]any{
				utils.MetaRates:    true,
				utils.MetaAccounts: false,
			},
		}
		var rply string
		client.Call(context.Background(), utils.CDRsV1ProcessEvent, ev, &rply)
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs, &utils.CDRFilters{Tenant: "cgrates.org"}, &cdrs); err != nil {
			t.Error(err)
		}
		if len(cdrs) != 1 {
			t.Errorf("unexpected number of cdrs found: %v", len(cdrs))
		}
		exp := &utils.CDR{
			Tenant: utils.CGRateSorg,
			Opts: map[string]any{
				utils.MetaCDRID:    cdrs[0].Opts[utils.MetaCDRID],
				utils.MetaRates:    true,
				utils.MetaAccounts: false,
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.OriginID:     "TestEv1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.Subject:      "1001",
				utils.ToR:          utils.MetaVoice,
				utils.Usage:        6e+10,
			},
			CreatedAt: cdrs[0].CreatedAt,
			UpdatedAt: cdrs[0].UpdatedAt,
		}
		if !reflect.DeepEqual(exp, cdrs[0]) {
			t.Errorf("Expecting <%#v>, \nreceived <%#v>", exp, cdrs[0])
		}
	})
	t.Run("EngineShutdown", func(t *testing.T) {
		if err := engine.KillEngine(100); err != nil {
			t.Error(err)
		}
	})

	t.Run("CountDBFiles", func(t *testing.T) {
		var dirs, files int
		if err := filepath.WalkDir(cfg.DbCfg().Opts.InternalDBDumpPath, func(root string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				dirs++
			} else {
				if !strings.HasPrefix(root, "/tmp/internal_db/db/*charger_filter_indexes") &&
					!strings.HasPrefix(root, "/tmp/internal_db/db/*versions") {
					t.Fatalf("got unexpected folder <%s>", root)
				}
				files++
			}
			return nil
		}); err != nil {
			t.Error(err)
		} else if dirs != 37 {
			t.Errorf("expected <%d> directories, received <%d>", 37, dirs)
		} else if files != 2 {
			t.Errorf("expected <%d> files, received <%d>", 2, files)
		}
	})
}

func TestMultipleDBsInternalFail(t *testing.T) {
	if err := os.MkdirAll("/tmp/internal_db/db", 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll("/tmp/internal_db"); err != nil {
			t.Error(err)
		}
	})
	cfgJSON := `{
"logger": {
	"level": 7
},

"db": {
	"db_conns": {
		"intrnl": {
			"db_type": "*internal"
		},
	},
	"items":{
		"*charger_profiles": {"limit": -1, "ttl": "", "static_ttl": false, "remote":false, "replicate":false, "dbConn": "intrnl"},
	},
	"opts": {
		"internalDBDumpPath": "/tmp/internal_db/db",
		"internalDBRewriteInterval": "0s"
	}
},

}
`
	expErr := `/001: <db> There can only be 1 internal DB`
	if _, err := engine.StartEngineFromString(cfgJSON, 200, t); err == nil ||
		!strings.Contains(err.Error(), expErr) {
		t.Errorf("expected error <%v>, received <%v>", expErr, err)
	}

}
