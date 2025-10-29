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
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestArgEEsUnmarshalJSON(t *testing.T) {
	var testEC = &EventCost{
		CGRID:     "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID:     utils.MetaDefault,
		StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
		AccountSummary: &AccountSummary{
			Tenant: "cgrates.org",
			ID:     "dan",
			BalanceSummaries: []*BalanceSummary{
				{
					UUID:     "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
					ID:       "BALANCE_1",
					Type:     utils.MetaMonetary,
					Value:    50,
					Initial:  60,
					Disabled: false,
				},
			},
			AllowNegative: false,
			Disabled:      false,
		},
	}
	cdBytes, err := json.Marshal(testEC)
	if err != nil {
		t.Fatal(err)
	}

	cgrEvWithIDs := CGREventWithEeIDs{
		EeIDs: []string{"id1", "id2"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
		},
	}

	t.Run("UnmarshalFromMap", func(t *testing.T) {
		var cdMap map[string]any
		if err = json.Unmarshal(cdBytes, &cdMap); err != nil {
			t.Fatal(err)
		}

		cgrEvWithIDs.Event[utils.CostDetails] = cdMap

		cgrEvBytes, err := json.Marshal(cgrEvWithIDs)
		if err != nil {
			t.Fatal(err)
		}

		var rcvCGREv CGREventWithEeIDs
		if err = json.Unmarshal(cgrEvBytes, &rcvCGREv); err != nil {
			t.Fatal(err)
		}

		expectedType := "*engine.EventCost"
		if cdType := fmt.Sprintf("%T", rcvCGREv.Event[utils.CostDetails]); cdType != expectedType {
			t.Fatalf("expected type to be %v, received %v", expectedType, cdType)
		}

		cgrEvWithIDs.Event[utils.CostDetails] = testEC
		if !reflect.DeepEqual(rcvCGREv, cgrEvWithIDs) {
			t.Errorf("expected: %v,\nreceived: %v",
				utils.ToJSON(cgrEvWithIDs), utils.ToJSON(rcvCGREv))
		}
	})
	t.Run("UnmarshalFromString", func(t *testing.T) {
		cdStringBytes, err := json.Marshal(string(cdBytes))
		if err != nil {
			t.Fatal(err)
		}
		var cdString string
		if err = json.Unmarshal(cdStringBytes, &cdString); err != nil {
			t.Fatal(err)
		}

		cgrEvWithIDs.Event[utils.CostDetails] = cdString

		cgrEvBytes, err := json.Marshal(cgrEvWithIDs)
		if err != nil {
			t.Fatal(err)
		}

		var rcvCGREv CGREventWithEeIDs
		if err = json.Unmarshal(cgrEvBytes, &rcvCGREv); err != nil {
			t.Fatal(err)
		}

		expectedType := "*engine.EventCost"
		if cdType := fmt.Sprintf("%T", rcvCGREv.Event[utils.CostDetails]); cdType != expectedType {
			t.Fatalf("expected type to be %v, received %v", expectedType, cdType)
		}

		cgrEvWithIDs.Event[utils.CostDetails] = testEC
		if !reflect.DeepEqual(rcvCGREv, cgrEvWithIDs) {
			t.Errorf("expected: %v,\nreceived: %v",
				utils.ToJSON(cgrEvWithIDs), utils.ToJSON(rcvCGREv))
		}
	})
}

func TestEngineCGREventWithEeIDsSetCloneable(t *testing.T) {
	attr := &CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{}}
	attr.SetCloneable(true)
	if !attr.clnb {
		t.Error("Expected attribute.clnb to be true after calling SetCloneable(true)")
	}
	attr.SetCloneable(false)
	if attr.clnb {
		t.Error("Expected attribute.clnb to be false after calling SetCloneable(false)")
	}
}
