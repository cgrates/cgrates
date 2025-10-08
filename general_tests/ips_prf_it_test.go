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
		"general": {
			"log_level": 7
		},
		"data_db": {
			"db_type": "*internal"
		},
		"stor_db": {
			"db_type": "*internal"
		},
        "admins": {
	       "enabled": true,
        },
		"ips": {
            "enabled": true,	
			"store_interval": "-1",
            "string_indexed_fields": ["*req.Account"],
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
