// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestReplicationTaskGobRoundTrip(t *testing.T) {
	// types passed to rpl.replicate() must be gob-registered
	argTypes := []any{
		new(FilterWithAPIOpts),
		new(utils.StatQueueProfileWithAPIOpts),
		new(utils.StatQueueWithAPIOpts),
		new(utils.ThresholdProfileWithAPIOpts),
		new(utils.ThresholdWithAPIOpts),
		new(utils.AccountWithAPIOpts),
		new(utils.ActionProfileWithAPIOpts),
		new(utils.AttributeProfileWithAPIOpts),
		new(utils.ChargerProfileWithAPIOpts),
		new(utils.GetIndexesArg),
		new(utils.IPAllocationsWithAPIOpts),
		new(utils.IPProfileWithAPIOpts),
		new(utils.RankingProfileWithAPIOpts),
		new(utils.RankingWithAPIOpts),
		new(utils.RateProfileWithAPIOpts),
		new(utils.ResourceProfileWithAPIOpts),
		new(utils.ResourceWithAPIOpts),
		new(utils.RouteProfileWithAPIOpts),
		new(utils.SetIndexesArg),
		new(utils.TenantIDWithAPIOpts),
		new(utils.TrendProfileWithAPIOpts),
		new(utils.TrendWithAPIOpts),
	}
	for _, args := range argTypes {
		t.Run(fmt.Sprintf("%T", args), func(t *testing.T) {
			task := &ReplicationTask{
				ConnIDs: []string{"conn1"},
				Method:  "ReplicatorSv1.Test",
				Args:    args,
			}
			var buf bytes.Buffer
			if err := gob.NewEncoder(&buf).Encode(task); err != nil {
				t.Fatalf("encode: %v", err)
			}
			var decoded ReplicationTask
			if err := gob.NewDecoder(&buf).Decode(&decoded); err != nil {
				t.Fatalf("decode: %v", err)
			}
		})
	}
}
