/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package utils

import (
	"reflect"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
)

func TestCloneBalance(t *testing.T) {
	expBlc := &Balance{
		ID:        "TEST_ID1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 1.1,
			},
		},
		Type: "*abstract",
		Opts: map[string]interface{}{
			"Destination": 10,
		},
		CostIncrements: []*CostIncrement{
			{
				FilterIDs:    []string{"*string:~*req.Account:1001"},
				Increment:    &Decimal{decimal.New(1, 1)},
				FixedFee:     &Decimal{decimal.New(75, 1)},
				RecurrentFee: &Decimal{decimal.New(20, 1)},
			},
		},
		AttributeIDs:   []string{"attr1", "attr2"},
		RateProfileIDs: []string{"RATE1", "RATE2"},
		UnitFactors: []*UnitFactor{
			{
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Factor:    &Decimal{decimal.New(20, 2)},
			},
		},
		Units: &Decimal{decimal.New(125, 3)},
	}
	if rcv := expBlc.Clone(); !reflect.DeepEqual(rcv, expBlc) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(expBlc), ToJSON(rcv))
	}

	expBlc.Opts = nil
	if rcv := expBlc.Clone(); !reflect.DeepEqual(rcv, expBlc) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(expBlc), ToJSON(rcv))
	}
}

func TestCloneAccountProfile(t *testing.T) {
	actPrf := &AccountProfile{
		Tenant:    "cgrates.org",
		ID:        "Profile_id1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &ActivationInterval{
			ActivationTime: time.Date(2020, 7, 21, 10, 0, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 7, 22, 10, 0, 0, 0, time.UTC),
		},
		Weights: DynamicWeights{
			{
				Weight: 2.4,
			},
		},
		Opts: map[string]interface{}{
			"Destination": 10,
		},
		Balances: map[string]*Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: DynamicWeights{
					{
						Weight: 1.1,
					},
				},
				Type: "*abstract",
				Opts: map[string]interface{}{
					"Destination": 10,
				},
				CostIncrements: []*CostIncrement{
					{
						FilterIDs:    []string{"*string:~*req.Account:1001"},
						Increment:    &Decimal{decimal.New(1, 1)},
						FixedFee:     &Decimal{decimal.New(75, 1)},
						RecurrentFee: &Decimal{decimal.New(20, 1)},
					},
				},
				AttributeIDs: []string{"attr1", "attr2"},
				UnitFactors: []*UnitFactor{
					{
						FilterIDs: []string{"*string:~*req.Account:1001"},
						Factor:    &Decimal{decimal.New(20, 2)},
					},
				},
				Units: &Decimal{decimal.New(125, 3)},
			},
		},
		ThresholdIDs: []string{"*none"},
	}
	if rcv := actPrf.Clone(); !reflect.DeepEqual(rcv, actPrf) {
		t.Errorf("Expected %+v, received %+v", ToJSON(actPrf), ToJSON(rcv))
	}

	actPrf.Opts = nil
	actPrf.ActivationInterval = nil
	if rcv := actPrf.Clone(); !reflect.DeepEqual(rcv, actPrf) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(actPrf), ToJSON(rcv))
	}
}

func TestTenantIDAccountProfile(t *testing.T) {
	actPrf := &AccountProfile{
		Tenant: "cgrates.org",
		ID:     "test_ID1",
	}
	exp := "cgrates.org:test_ID1"
	if rcv := actPrf.TenantID(); rcv != exp {
		t.Errorf("Expected %+v, received %+v", exp, rcv)
	}
}

func TestAccountProfileAsAccountProfile(t *testing.T) {
	apiAccPrf := &APIAccountProfile{
		Tenant: "cgrates.org",
		ID:     "test_ID1",
		Opts:   map[string]interface{}{},
		Balances: map[string]*APIBalance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights:   ";1.2",
				Type:      "*abstract",
				Opts: map[string]interface{}{
					"Destination": 10,
				},
				Units: 0,
			},
		},
		Weights: ";10",
	}
	expected := &AccountProfile{
		Tenant: "cgrates.org",
		ID:     "test_ID1",
		Opts:   map[string]interface{}{},
		Balances: map[string]*Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: DynamicWeights{
					{
						Weight: 1.2,
					},
				},
				Type: "*abstract",
				Opts: map[string]interface{}{
					"Destination": 10,
				},
				Units: NewDecimal(0, 0),
			},
		},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	if rcv, err := apiAccPrf.AsAccountProfile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expected), ToJSON(rcv))
	}
}

func TestAsAccountProfileError(t *testing.T) {
	apiAccPrf := &APIAccountProfile{
		Tenant: "cgrates.org",
		ID:     "test_ID1",
		Opts:   map[string]interface{}{},
		Balances: map[string]*APIBalance{
			"MonetaryBalance": {
				Weights: ";10",
			},
		},
		Weights: "10",
	}
	expectedErr := "invalid DynamicWeight format for string <10>"
	if _, err := apiAccPrf.AsAccountProfile(); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	apiAccPrf.Weights = ";10"
	apiAccPrf.Balances["MonetaryBalance"].Weights = "10"
	expectedErr = "invalid DynamicWeight format for string <10>"
	if _, err := apiAccPrf.AsAccountProfile(); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestAPIBalanceAsBalance(t *testing.T) {
	blc := &APIBalance{
		ID: "VoiceBalance",
		CostIncrements: []*APICostIncrement{
			{
				FilterIDs:    []string{"*string:~*req.Account:1001"},
				Increment:    Float64Pointer(1),
				FixedFee:     Float64Pointer(10),
				RecurrentFee: Float64Pointer(35),
			},
		},
		Weights: ";10",
		UnitFactors: []*APIUnitFactor{
			{
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Factor:    20,
			},
		},
	}
	expected := &Balance{
		ID: "VoiceBalance",
		CostIncrements: []*CostIncrement{
			{
				FilterIDs:    []string{"*string:~*req.Account:1001"},
				Increment:    NewDecimal(1, 0),
				FixedFee:     NewDecimal(10, 0),
				RecurrentFee: NewDecimal(35, 0),
			},
		},
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		UnitFactors: []*UnitFactor{
			{
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Factor:    NewDecimal(20, 0),
			},
		},
		Units: NewDecimal(0, 0),
	}
	if rcv, err := blc.AsBalance(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(expected), ToJSON(rcv))
	}

}

func TestAccountProfileBalancesAlteredCompareLength(t *testing.T) {
	actPrf := &AccountProfile{
		Balances: map[string]*Balance{
			"testString":  {},
			"testString2": {},
		},
	}

	actBk := map[string]*decimal.Big{
		"testString": {},
	}

	result := actPrf.BalancesAltered(actBk)
	if result != true {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", true, result)
	}

}

func TestAccountProfileBalancesAlteredCheckKeys(t *testing.T) {
	actPrf := &AccountProfile{
		Balances: map[string]*Balance{
			"testString": {},
		},
	}

	actBk := map[string]*decimal.Big{
		"testString2": {},
	}

	result := actPrf.BalancesAltered(actBk)
	if result != true {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", true, result)
	}

}

func TestAccountProfileBalancesAlteredCompareValues(t *testing.T) {
	actPrf := &AccountProfile{
		Balances: map[string]*Balance{
			"testString": {
				Units: &Decimal{decimal.New(1, 1)},
			},
		},
	}

	actBk := map[string]*decimal.Big{
		"testString": {},
	}

	result := actPrf.BalancesAltered(actBk)
	if result != true {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", true, result)
	}

}

func TestAccountProfileBalancesAlteredFalse(t *testing.T) {
	actPrf := &AccountProfile{}

	actBk := AccountBalancesBackup{}

	result := actPrf.BalancesAltered(actBk)
	if result != false {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", false, result)
	}

}

func TestAPRestoreFromBackup(t *testing.T) {
	actPrf := &AccountProfile{
		Balances: map[string]*Balance{
			"testString": {
				Units: &Decimal{},
			},
		},
	}

	actBk := AccountBalancesBackup{
		"testString": decimal.New(1, 1),
	}

	actPrf.RestoreFromBackup(actBk)
	for key, value := range actBk {
		if actPrf.Balances[key].Units.Big != value {
			t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", actPrf.Balances[key].Units.Big, value)
		}
	}
}

func TestAPAccountBalancesBackup(t *testing.T) {
	actPrf := &AccountProfile{
		Balances: map[string]*Balance{
			"testKey": {
				Units: &Decimal{decimal.New(1234, 3)},
			},
		},
	}

	actBk := actPrf.AccountBalancesBackup()
	for key, value := range actBk {
		if actPrf.Balances[key].Units.Big.Cmp(value) != 0 {
			t.Errorf("\ngot: <%+v>, \nwant: <%+v>", value, actPrf.Balances[key].Units.Big)
		}
	}
}

func TestAPNewDefaultBalance(t *testing.T) {

	const torFltr = "*string:~*req.ToR:"
	id := "testID"

	expected := &Balance{
		ID:    id,
		Type:  MetaConcrete,
		Units: NewDecimal(0, 0),
		CostIncrements: []*CostIncrement{
			{
				FilterIDs:    []string{torFltr + MetaVoice},
				Increment:    NewDecimal(int64(time.Second), 0),
				RecurrentFee: NewDecimal(0, 0),
			},
			{
				FilterIDs:    []string{torFltr + MetaData},
				Increment:    NewDecimal(1024*1024, 0),
				RecurrentFee: NewDecimal(0, 0),
			},
			{
				FilterIDs:    []string{torFltr + MetaSMS},
				Increment:    NewDecimal(1, 0),
				RecurrentFee: NewDecimal(0, 0),
			},
		},
	}

	received := NewDefaultBalance(id)

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nReceived: <%+v>,\nExpected: <%+v>", received, expected)
	}
}

func TestAPSort(t *testing.T) {

	apS := AccountProfilesWithWeight{
		{
			Weight: 2,
		},
		{
			Weight: 1,
		},
		{
			Weight: 3,
		},
	}
	expected := AccountProfilesWithWeight{
		{
			Weight: 3,
		},
		{
			Weight: 2,
		},
		{
			Weight: 1,
		},
	}

	apS.Sort()
	if !reflect.DeepEqual(apS, expected) {
		t.Errorf("\nReceived: <%+v>, \nExpected: <%+v>", ToJSON(apS), ToJSON(expected))
	}
}

func TestAPAccountProfiles(t *testing.T) {

	apS := AccountProfilesWithWeight{
		{
			AccountProfile: &AccountProfile{
				Tenant:    "testTenant1",
				ID:        "testID1",
				FilterIDs: []string{"testFID1", "testFID2"},
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, time.April, 12, 0, 0, 0, 0, time.UTC),
					ExpiryTime:     time.Date(2020, time.April, 12, 10, 0, 0, 0, time.UTC),
				},
				Weights: nil,
				Balances: map[string]*Balance{
					"testBalance1": &Balance{
						ID:    "testBalance1",
						Type:  MetaAbstract,
						Units: &Decimal{decimal.New(0, 0)},
					},
				},
			},
			Weight: 23,
			LockID: "testString1",
		},
		{
			AccountProfile: &AccountProfile{
				Tenant:    "testTenant2",
				ID:        "testID2",
				FilterIDs: []string{"testFID1", "testFID2"},
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, time.April, 12, 0, 0, 0, 0, time.UTC),
					ExpiryTime:     time.Date(2020, time.April, 12, 10, 0, 0, 0, time.UTC),
				},
				Weights: nil,
				Balances: map[string]*Balance{
					"testBalance2": &Balance{
						ID:    "testBalance2",
						Type:  MetaAbstract,
						Units: &Decimal{decimal.New(0, 0)},
					},
				},
			},
			Weight: 15,
			LockID: "testString2",
		},
	}

	expected := make([]*AccountProfile, 0)
	for i := range apS {
		expected = append(expected, apS[i].AccountProfile)
	}
	received := apS.AccountProfiles()

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nReceived: <%+v>,\nExpected: <%+v>", ToJSON(received), ToJSON(expected))
	}
}

func TestAPLockIDs(t *testing.T) {
	apS := AccountProfilesWithWeight{
		{
			AccountProfile: &AccountProfile{
				Tenant:    "testTenant1",
				ID:        "testID1",
				FilterIDs: []string{"testFID1", "testFID2"},
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, time.April, 12, 0, 0, 0, 0, time.UTC),
					ExpiryTime:     time.Date(2020, time.April, 12, 10, 0, 0, 0, time.UTC),
				},
				Weights: nil,
				Balances: map[string]*Balance{
					"testBalance1": &Balance{
						ID:    "testBalance1",
						Type:  MetaAbstract,
						Units: &Decimal{decimal.New(0, 0)},
					},
				},
			},
			Weight: 23,
			LockID: "testString1",
		},
		{
			AccountProfile: &AccountProfile{
				Tenant:    "testTenant2",
				ID:        "testID2",
				FilterIDs: []string{"testFID1", "testFID2"},
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, time.April, 12, 0, 0, 0, 0, time.UTC),
					ExpiryTime:     time.Date(2020, time.April, 12, 10, 0, 0, 0, time.UTC),
				},
				Weights: nil,
				Balances: map[string]*Balance{
					"testBalance2": &Balance{
						ID:    "testBalance2",
						Type:  MetaAbstract,
						Units: &Decimal{decimal.New(0, 0)},
					},
				},
			},
			Weight: 15,
			LockID: "testString2",
		},
	}

	expected := make([]string, 0)
	for i := range apS {
		expected = append(expected, apS[i].LockID)
	}
	received := apS.LockIDs()

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nReceived: <%v>, \nExpected: <%+v>", received, expected)
	}
}
