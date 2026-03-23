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
		new(StatQueueProfileWithAPIOpts),
		new(StatQueueWithAPIOpts),
		new(ThresholdProfileWithAPIOpts),
		new(ThresholdWithAPIOpts),
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
