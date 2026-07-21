//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func Benchmark10IPsAllocated(b *testing.B) {

	content := `{
		"logger": {
			"level": 7
		},
		"db": {
			"dbConns": {
				"*default": {
					"dbType": "*internal",
							"opts":{
				"internalDBRewriteInterval": "0s",
				"internalDBDumpInterval": "0s"
			}
				}
			},

		},
        "admins": {
	       "enabled": true,
        },
		"ips": {
            "enabled": true,	
			"storeInterval": "-1",
            "stringIndexedFields": ["*req.Account"],
		},
	}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		LogBuffer:  bytes.NewBuffer(nil),
		Encoding:   utils.MetaJSON,
	}
	client, _ := ng.Run(b)

	var reply string
	for i := 1; i <= 10; i++ {
		ipProfile := &utils.IPProfileWithAPIOpts{
			IPProfile: &utils.IPProfile{
				Tenant:    "cgrates.org",
				ID:        fmt.Sprintf("IP_PROF_%d", i),
				FilterIDs: []string{fmt.Sprintf("*string:~*req.Account:IP_PROF_%d", i)},
				TTL:       10 * time.Minute,
				Pools: []*utils.IPPool{
					{
						ID:      "POOL_A",
						Range:   fmt.Sprintf("10.10.10.%d/32", i),
						Message: "Allocated by test",
					},
				},
			},
		}
		if err := client.Call(context.Background(), utils.AdminSv1SetIPProfile, ipProfile, &reply); err != nil {
			b.Fatalf("Failed to set IP profile: %v", err)
		}

	}

	b.Run("IPsAllocateEvent", func(b *testing.B) {
		var prof atomic.Int64
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				i := prof.Add(1) % 10
				allocID := utils.GenUUID()
				allocateIP(b, client, "event1", fmt.Sprintf("IP_PROF_%d", i), allocID)
				checkAllocs(b, client, fmt.Sprintf("IP_PROF_%d", i), allocID)
				releaseIP(b, client, fmt.Sprintf("IP_PROF_%d", i), allocID)

				allocateIP(b, client, "event1", fmt.Sprintf("IP_PROF_%d", i), allocID)
				checkAllocs(b, client, fmt.Sprintf("IP_PROF_%d", i), allocID)
			}
		})
	})
}
