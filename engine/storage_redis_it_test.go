//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// go test -bench RedisGetKeysForPrefix -run=^# -count 3 -benchtime=10s
func BenchmarkRedisGetKeysForPrefix(b *testing.B) {
	rs, _ := NewRedisStorage("127.0.0.1:6379", 10, utils.CGRateSLwr, "", "json", 10, 20,
		"", false, 5*time.Second, 0, 0, 0, 0, 150*time.Microsecond, 0, false, "", "", "", 1000, nil, nil)
	chargerProfile := &utils.ChargerProfile{
		ID:        "TestA_CHARGER1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		RunID:        "run1",
		AttributeIDs: []string{"ATTR_ TEST1"},
	}
	id := "ChargerP"
	var prfID string
	for i := 0; i <= 10000; i++ {
		if i%1000 == 0 {
			if i/1000%2 == 0 {
				prfID = "TestA:"
			} else {
				prfID = "TestB:"
			}
		}
		prfID = prfID[:6] + strconv.Itoa(i) + id
		chargerProfile.ID = prfID
		rs.SetChargerProfileDrv(context.Background(), chargerProfile)
	}
	for i := 0; i < b.N; i++ {
		rs.GetKeysForPrefix(context.Background(), "TestA", "")
	}
	prfx := []string{"TestA", "TestB", "Test"}
	for _, v := range prfx {
		b.Run(fmt.Sprintf("test case: prefix = %q", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				rs.GetKeysForPrefix(context.Background(), v, "")
			}
		})
	}
	rs.Flush("")

}
