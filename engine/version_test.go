// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestVersionCompare(t *testing.T) {
	current := Versions{utils.AccountsStr: 1, utils.Actions: 1}
	stored := Versions{utils.AccountsStr: 0, utils.Actions: 1}

	if subsys := stored.Compare(current); subsys != utils.AccountsStr {
		t.Errorf("expected mismatch on %s, received %q", utils.AccountsStr, subsys)
	}

	stored[utils.AccountsStr] = 1
	if subsys := stored.Compare(current); subsys != "" {
		t.Errorf("expected no mismatch, received %q", subsys)
	}
}

func TestCurrentDBVersions(t *testing.T) {
	expVersDataDB := Versions{
		utils.Stats: 1, utils.AccountsStr: 1, utils.Actions: 1,
		utils.Thresholds: 1, utils.Routes: 1, utils.Attributes: 1,
		utils.RQF: 1, utils.ResourceStr: 1,
		utils.Subscribers: 1,
		utils.Chargers:    1,
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

func TestCurrentAllDBVersions(t *testing.T) {
	expected := Versions{
		utils.Stats:          1,
		utils.AccountsStr:    1,
		utils.Actions:        1,
		utils.Thresholds:     1,
		utils.Routes:         1,
		utils.Attributes:     1,
		utils.RQF:            1,
		utils.ResourceStr:    1,
		utils.Subscribers:    1,
		utils.Chargers:       1,
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
