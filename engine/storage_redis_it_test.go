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
		"", false, 5*time.Second, 0, 0, 0, 0, 150*time.Microsecond, 0, false, "", "", "", nil, nil)
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
