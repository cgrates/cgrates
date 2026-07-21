//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestThresholdProcessEventWithAttributes(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.InternalDBCfg
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	content := `{
"general": {
	"log_level": 7
},
"thresholds": {
	"enabled": true,
	"storeInterval": "-1",
	"indexedSelects": false,
	"conns": {
		"*attributes": [{"connIDs": ["*localhost"]}],
		"*actions": [{"connIDs": ["*localhost"]}]
	}
},
"attributes": {
	"enabled": true,
	"indexedSelects": false
},
"actions": {
	"enabled": true,
	"indexedSelects": false,
	"conns": {
		"*ees": [{"connIDs": ["*localhost"]}]
	}
},
"ees": {
	"enabled": true,
	"exporters": [{
		"id": "virtual_ees",
		"type": "*virt",
		"fields": [
			{"tag": "Field1", "path": "*uch.Field1", "type": "*variable", "value": "~*req.Field1"}
		]
	}]
},
"admins": {
	"enabled": true
}}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      dbCfg,
		Encoding:   *utils.Encoding,
	}
	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	t.Run("SetActionProfile", func(t *testing.T) {
		actPrf := &utils.ActionProfileWithAPIOpts{
			ActionProfile: &utils.ActionProfile{
				Tenant:    "cgrates.org",
				ID:        "ACT_1",
				FilterIDs: []string{"*string:~*req.Field1:Value1"},
				Weights:   utils.DynamicWeights{{Weight: 10}},
				Schedule:  utils.MetaASAP,
				Actions: []*utils.APAction{
					{
						ID:   "EXPORT",
						Type: utils.MetaExport,
						Opts: map[string]any{
							utils.MetaExporterIDs: "virtual_ees",
						},
					},
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetActionProfile, actPrf, &reply); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("SetAttributeProfile", func(t *testing.T) {
		attrPrf := &utils.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &utils.APIAttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_1",
				FilterIDs: []string{"*string:~*opts.*context:*thresholds"},
				Weights:   utils.DynamicWeights{{Weight: 10}},
				Attributes: []*utils.ExternalAttribute{
					{
						Path:  "*req.Field1",
						Type:  utils.MetaConstant,
						Value: "Value1",
					},
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetAttributeProfile, attrPrf, &reply); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("SetThresholdProfile", func(t *testing.T) {
		tPrfl := &utils.ThresholdProfileWithAPIOpts{
			ThresholdProfile: &utils.ThresholdProfile{
				Tenant:           "cgrates.org",
				ID:               "THR_ATTR_TEST",
				FilterIDs:        []string{"*string:~*req.Account:1001"},
				MaxHits:          -1,
				MinHits:          1,
				Weights:          utils.DynamicWeights{{Weight: 10}},
				ActionProfileIDs: []string{"ACT_1"},
				AttributeIDs:     []string{"ATTR_1"},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetThresholdProfile, tPrfl, &reply); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ProcessEvent", func(t *testing.T) {
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event1",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
			},
			APIOpts: map[string]any{},
		}
		var thIDs []string
		if err := client.Call(context.Background(), utils.ThresholdSv1ProcessEvent, ev, &thIDs); err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if !slices.Contains(thIDs, "THR_ATTR_TEST") {
			t.Errorf("expected THR_ATTR_TEST in reply, got %v", thIDs)
		}
		var thr utils.Threshold
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THR_ATTR_TEST"},
			}, &thr); err != nil {
			t.Error(err)
		} else if thr.Hits != 1 {
			t.Errorf("expected 1 hit, got %d", thr.Hits)
		}
	})

	t.Run("VerifyActionExecution", func(t *testing.T) {
		var field1 any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem,
			&utils.ArgsGetCacheItemWithAPIOpts{
				Tenant: "cgrates.org",
				ArgsGetCacheItem: utils.ArgsGetCacheItem{
					CacheID: utils.CacheUCH,
					ItemID:  "Field1",
				},
			}, &field1); err != nil {
			t.Error(err)
		} else if field1 != "Value1" {
			t.Errorf("expected Field1=Value1 in UCH, got %v", field1)
		}
	})
}
