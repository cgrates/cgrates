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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestVersionCompare(t *testing.T) {
	x := Versions{utils.AccountsStr: 2, utils.Actions: 2,
		utils.Attributes: 2, utils.Chargers: 2,
		utils.CostDetails: 2}
	y := Versions{utils.AccountsStr: 1, utils.Actions: 2,
		utils.Attributes: 2, utils.Chargers: 2,
		utils.CostDetails: 2}

	message1 := y.Compare(x, utils.MetaMongo)
	if message1 != "cgr-migrator -exec=*accounts" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -exec=*accounts", message1)
	}
	message7 := y.Compare(x, utils.MetaRedis)
	if message7 != "cgr-migrator -exec=*accounts" {
		t.Errorf("Error failed to compare to curent version expected: %s received: %s", "cgr-migrator -exec=*accounts", message7)
	}

	y[utils.AccountsStr] = 2
	message8 := y.Compare(x, utils.MetaRedis)
	if message8 != utils.EmptyString {
		t.Errorf("Expected %+v, received %+v", utils.EmptyString, message8)
	}
}

func TestCurrentDBVersions(t *testing.T) {
	expVersDataDB := Versions{
		utils.Stats: 4, utils.AccountsStr: 3, utils.Actions: 2,
		utils.Thresholds: 4, utils.Routes: 2, utils.Attributes: 7,
		utils.RQF: 5, utils.ResourceStr: 1,
		utils.Subscribers: 1,
		utils.Chargers:    2,
		utils.LoadIDsVrs:  1, utils.RateProfiles: 1,
		utils.ActionProfiles: 1,
	}

	if vrs := CurrentDBVersions(utils.MetaMongo); !reflect.DeepEqual(expVersDataDB, vrs) {
		t.Errorf("Expectred %+v, received %+v", expVersDataDB, vrs)
	}

	if vrs := CurrentDBVersions(utils.MetaRedis); !reflect.DeepEqual(expVersDataDB, vrs) {
		t.Errorf("Expectred %+v, received %+v", expVersDataDB, vrs)
	}

	if vrs := CurrentDBVersions("NOT_A_DB"); vrs != nil {
		t.Error(vrs)
	}

}

func TestCurrentStorDBVersions(t *testing.T) {
	expected := Versions{
		utils.CostDetails:      2,
		utils.SessionSCosts:    3,
		utils.CDRs:             2,
		utils.TpFilters:        1,
		utils.TpThresholds:     1,
		utils.TpRoutes:         1,
		utils.TpStats:          1,
		utils.TpResources:      1,
		utils.TpResource:       1,
		utils.TpChargers:       1,
		utils.TpRateProfiles:   1,
		utils.TpActionProfiles: 1,
	}

	actual := CurrentStorDBVersions()

	if len(actual) != len(expected) {
		t.Fatalf("Expected %d versions, got %d", len(expected), len(actual))
	}

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists {
			t.Errorf("Expected version for %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("For %s, expected %d, got %d", key, expectedValue, actualValue)
		}
	}
}

func TestCurrentAllDBVersions(t *testing.T) {
	expected := Versions{
		utils.Stats:          4,
		utils.AccountsStr:    3,
		utils.Actions:        2,
		utils.Thresholds:     4,
		utils.Routes:         2,
		utils.Attributes:     7,
		utils.RQF:            5,
		utils.ResourceStr:    1,
		utils.Subscribers:    1,
		utils.Chargers:       2,
		utils.LoadIDsVrs:     1,
		utils.RateProfiles:   1,
		utils.ActionProfiles: 1,
	}

	actual := CurrentAllDBVersions()

	if len(actual) != len(expected) {
		t.Fatalf("Expected %d versions, got %d", len(expected), len(actual))
	}

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists {
			t.Errorf("Expected version for %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("For %s, expected %d, got %d", key, expectedValue, actualValue)
		}
	}
}
