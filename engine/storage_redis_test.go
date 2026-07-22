//go:build performance

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestNewRedisStorage(t *testing.T) {
	_, err := NewRedisStorage("127.0.0.1:6379", 10, "cgrates", "", "json", 10, 20,
		"", false, 5*time.Second, 0, 0, 0, 0, false, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkRedisScan(b *testing.B) {
	rs, err := NewRedisStorage("127.0.0.1:6379", 10, "cgrates", "", "json", 10, 20,
		"", false, 5*time.Second, 0, 0, 0, 0, false, "", "", "")
	if err != nil {
		b.Fatalf("Failed to create Redis storage: %v", err)
	}
	defer rs.Close()
	chargerProfile := &ChargerProfile{
		ID:           "TestA_CHARGER1",
		Tenant:       "cgrates.org",
		FilterIDs:    []string{"*string:~*req.TestCase:AdminSAPIs"},
		Weight:       10,
		RunID:        "run1",
		AttributeIDs: []string{"ATTR_ TEST1"},
	}
	id := "ChargerP"
	var prfID string
	for i := 0; i <= 20; i++ {
		if i%10 == 0 {
			if (i/10)%2 == 0 {
				prfID = "TestA:"
			} else {
				prfID = "TestB:"
			}
		}
		prfID = prfID[:6] + strconv.Itoa(i) + id
		chargerProfile.ID = prfID
		rs.SetChargerProfileDrv(chargerProfile)
	}
	prfx := []string{"TestA", "TestB", "Test"}
	for _, v := range prfx {
		b.Run(fmt.Sprintf("test case: prefix = %q", v), func(b *testing.B) {
			for b.Loop() {
				rs.GetKeysForPrefix(v, utils.EmptyString)
			}
		})
	}
	rs.Flush("")

}
