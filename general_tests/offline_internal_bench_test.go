//go:build performance

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// benchmark runs get attributes to check how much time it takes to get small or big sample of data for internalDB
func BenchmarkInternalDBGet(b *testing.B) {
	jsonCfg := fmt.Sprintf(`{
    "general": {
        "log_level": 7,
    },
    "data_db": {
        "db_type": "*internal",
    },
	"stor_db": {
        "db_type": "*internal",
    },
    "rals": {
        "enabled": true,
        "thresholds_conns": ["*internal"],
        "max_increments":3000000
    },
    "schedulers": {
        "enabled": true,
        "stats_conns": ["*internal"]
    },
    "attributes": {
        "enabled": true,
        "stats_conns": ["*internal"],
        "resources_conns": ["*internal"],
        "apiers_conns": ["*internal"]
    },
    "chargers": {
        "enabled": true,
        "attributes_conns": ["*internal"]
    },
    "resources": {
        "enabled": true,
        "store_interval": "-1",
        "thresholds_conns": ["*internal"]
    },
    "stats": {
        "enabled": true,
        "store_interval": "-1",
        "thresholds_conns": ["*internal"]
    },
    "thresholds": {
        "enabled": true,
        "store_interval": "-1"
    },
    "routes": {
        "enabled": true,
        "prefix_indexed_fields":["*req.Destination"],
        "stats_conns": ["*internal"],
        "resources_conns": ["*internal"],
        "rals_conns": ["*internal"]
    },
    "sessions": {
        "enabled": true,
        "routes_conns": ["*internal"],
        "resources_conns": ["*internal"],
        "attributes_conns": ["*internal"],
        "rals_conns": ["*internal"],
        "chargers_conns": ["*internal"],
        "backup_interval": "-1"
    },
    "apiers": {
        "enabled": true,
        "scheduler_conns": ["*internal"]
    },
    "filters": {
        "stats_conns": ["*internal"],
        "resources_conns": ["*internal"],
        "apiers_conns": ["*internal"]
    }
}`)
	ng := engine.TestEngine{
		ConfigJSON: jsonCfg,
	}
	client, _ := ng.Run(b)

	time.Sleep(100 * time.Millisecond)

	var result string
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_Sample",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.EventName:AddAccountInfo"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + "Balance",
					Type: utils.MetaVariable,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "~*accounts.1001.BalanceMap.*monetary[0].Value",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	alsPrf.Compile()
	if err := client.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		b.Error(err)
	} else if result != utils.OK {
		b.Error("Unexpected reply returned", result)
	}
	attributes500 := []*engine.Attribute{}
	for i := range 500 {
		attributes500 = append(attributes500, &engine.Attribute{
			Path: utils.MetaReq + utils.NestingSep + "Balance",
			Type: utils.MetaVariable,
			Value: config.RSRParsers{
				&config.RSRParser{
					Rules: fmt.Sprintf("~*accounts.100%v.BalanceMap.*monetary[0].Value", i),
				},
			},
		})
	}

	alsPrf = &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_Sample2",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.EventName:AddAccountInfo"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: attributes500,
			Blocker:    false,
			Weight:     10,
		},
	}
	alsPrf.Compile()
	if err := client.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		b.Error(err)
	} else if result != utils.OK {
		b.Error("Unexpected reply returned", result)
	}

	b.Run("SetSmallDataAttribute", func(b *testing.B) {
		//  benchmark
		b.ResetTimer()
		for b.Loop() {
			var replyAttr *engine.AttributeProfile
			if err := client.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
				utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_Sample"}}, &replyAttr); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("SetBiggerDataAttribute", func(b *testing.B) {
		// benchmark
		b.ResetTimer()
		for b.Loop() {
			var replyAttr *engine.AttributeProfile
			if err := client.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
				utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_Sample2"}}, &replyAttr); err != nil {
				b.Fatal(err)
			}
		}
	})

}
