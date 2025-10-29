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

package migrator

import (
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestConvertSupplierToRoute(t *testing.T) {
	supplierProfile := &SupplierProfile{
		Tenant:             "cgrates.org",
		ID:                 "ProfileID",
		FilterIDs:          []string{"filter1", "filter2"},
		ActivationInterval: &utils.ActivationInterval{},
		Sorting:            "weight",
		SortingParameters:  []string{"param1", "param2"},
		Suppliers: []*Supplier{
			{
				ID:                 "supplier1",
				FilterIDs:          []string{"filterA"},
				AccountIDs:         []string{"account1"},
				RatingPlanIDs:      []string{"rating1"},
				ResourceIDs:        []string{"resource1"},
				StatIDs:            []string{"stat1"},
				Weight:             10.0,
				Blocker:            false,
				SupplierParameters: "param1",
			},
			{
				ID:                 "supplier2",
				FilterIDs:          []string{"filterB"},
				AccountIDs:         []string{"account2"},
				RatingPlanIDs:      []string{"rating2"},
				ResourceIDs:        []string{"resource2"},
				StatIDs:            []string{"stat2"},
				Weight:             20.0,
				Blocker:            true,
				SupplierParameters: "param2",
			},
		},
		Weight: 15.0,
	}

	expectedRoute := &engine.RouteProfile{
		Tenant:             "cgrates.org",
		ID:                 "ProfileID",
		FilterIDs:          []string{"filter1", "filter2"},
		ActivationInterval: &utils.ActivationInterval{},
		Sorting:            "weight",
		SortingParameters:  []string{"param1", "param2"},
		Weight:             15.0,
		Routes: []*engine.Route{
			{
				ID:              "supplier1",
				FilterIDs:       []string{"filterA"},
				AccountIDs:      []string{"account1"},
				RatingPlanIDs:   []string{"rating1"},
				ResourceIDs:     []string{"resource1"},
				StatIDs:         []string{"stat1"},
				Weight:          10.0,
				Blocker:         false,
				RouteParameters: "param1",
			},
			{
				ID:              "supplier2",
				FilterIDs:       []string{"filterB"},
				AccountIDs:      []string{"account2"},
				RatingPlanIDs:   []string{"rating2"},
				ResourceIDs:     []string{"resource2"},
				StatIDs:         []string{"stat2"},
				Weight:          20.0,
				Blocker:         true,
				RouteParameters: "param2",
			},
		},
	}

	result := convertSupplierToRoute(supplierProfile)

	if result.Tenant != expectedRoute.Tenant {
		t.Errorf("expected Tenant %s, got %s", expectedRoute.Tenant, result.Tenant)
	}
	if result.ID != expectedRoute.ID {
		t.Errorf("expected ID %s, got %s", expectedRoute.ID, result.ID)
	}
	if len(result.FilterIDs) != len(expectedRoute.FilterIDs) {
		t.Errorf("expected FilterIDs length %d, got %d", len(expectedRoute.FilterIDs), len(result.FilterIDs))
	}
	for i, filterID := range expectedRoute.FilterIDs {
		if result.FilterIDs[i] != filterID {
			t.Errorf("expected FilterID %s, got %s", filterID, result.FilterIDs[i])
		}
	}
}
