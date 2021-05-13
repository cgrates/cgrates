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
	"math"
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

func TestCloneAccount(t *testing.T) {
	actPrf := &Account{
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

func TestTenantIDAccount(t *testing.T) {
	actPrf := &Account{
		Tenant: "cgrates.org",
		ID:     "test_ID1",
	}
	exp := "cgrates.org:test_ID1"
	if rcv := actPrf.TenantID(); rcv != exp {
		t.Errorf("Expected %+v, received %+v", exp, rcv)
	}
}

func TestAPIAccountAsAccount(t *testing.T) {
	apiAccPrf := &APIAccount{
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
	expected := &Account{
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
	if rcv, err := apiAccPrf.AsAccount(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expected), ToJSON(rcv))
	}
}

func TestAsAccountError(t *testing.T) {
	apiAccPrf := &APIAccount{
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
	if _, err := apiAccPrf.AsAccount(); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	apiAccPrf.Weights = ";10"
	apiAccPrf.Balances["MonetaryBalance"].Weights = "10"
	expectedErr = "invalid DynamicWeight format for string <10>"
	if _, err := apiAccPrf.AsAccount(); err == nil || err.Error() != expectedErr {
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

func TestAccountBalancesAlteredCompareLength(t *testing.T) {
	actPrf := &Account{
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

func TestAccountBalancesAlteredCheckKeys(t *testing.T) {
	actPrf := &Account{
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

func TestAccountBalancesAlteredCompareValues(t *testing.T) {
	actPrf := &Account{
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

func TestAccountBalancesAlteredFalse(t *testing.T) {
	actPrf := &Account{}

	actBk := AccountBalancesBackup{}

	result := actPrf.BalancesAltered(actBk)
	if result != false {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", false, result)
	}

}

func TestAPRestoreFromBackup(t *testing.T) {
	actPrf := &Account{
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
	actPrf := &Account{
		Balances: map[string]*Balance{
			"testKey": {
				Units: &Decimal{decimal.New(1234, 3)},
			},
		},
	}

	actBk := actPrf.AccountBalancesBackup()
	for key, value := range actBk {
		if actPrf.Balances[key].Units.Big.Cmp(value) != 0 {
			t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", actPrf.Balances[key].Units.Big, value)
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
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestAPApsSort(t *testing.T) {

	apS := AccountsWithWeight{
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
	expected := AccountsWithWeight{
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
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", ToJSON(expected), ToJSON(apS))
	}
}

func TestAPAccount(t *testing.T) {
	apS := AccountsWithWeight{
		{
			Account: &Account{
				Tenant:    "testTenant1",
				ID:        "testID1",
				FilterIDs: []string{"testFID1", "testFID2"},
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, time.April, 12, 0, 0, 0, 0, time.UTC),
					ExpiryTime:     time.Date(2020, time.April, 12, 10, 0, 0, 0, time.UTC),
				},
				Weights: nil,
				Balances: map[string]*Balance{
					"testBalance1": {
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
			Account: &Account{
				Tenant:    "testTenant2",
				ID:        "testID2",
				FilterIDs: []string{"testFID1", "testFID2"},
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, time.April, 12, 0, 0, 0, 0, time.UTC),
					ExpiryTime:     time.Date(2020, time.April, 12, 10, 0, 0, 0, time.UTC),
				},
				Weights: nil,
				Balances: map[string]*Balance{
					"testBalance2": {
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

	expected := make([]*Account, 0)
	for i := range apS {
		expected = append(expected, apS[i].Account)
	}
	received := apS.Accounts()

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", ToJSON(expected), ToJSON(received))
	}
}

func TestAPLockIDs(t *testing.T) {
	apS := AccountsWithWeight{
		{
			Account: &Account{
				Tenant:    "testTenant1",
				ID:        "testID1",
				FilterIDs: []string{"testFID1", "testFID2"},
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, time.April, 12, 0, 0, 0, 0, time.UTC),
					ExpiryTime:     time.Date(2020, time.April, 12, 10, 0, 0, 0, time.UTC),
				},
				Weights: nil,
				Balances: map[string]*Balance{
					"testBalance1": {
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
			Account: &Account{
				Tenant:    "testTenant2",
				ID:        "testID2",
				FilterIDs: []string{"testFID1", "testFID2"},
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, time.April, 12, 0, 0, 0, 0, time.UTC),
					ExpiryTime:     time.Date(2020, time.April, 12, 10, 0, 0, 0, time.UTC),
				},
				Weights: nil,
				Balances: map[string]*Balance{
					"testBalance2": {
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
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestAPBlcsSort(t *testing.T) {

	blncS := BalancesWithWeight{
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
	expected := BalancesWithWeight{
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

	blncS.Sort()
	if !reflect.DeepEqual(blncS, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", ToJSON(expected), ToJSON(blncS))
	}
}

func TestAPBalances(t *testing.T) {
	blncS := BalancesWithWeight{
		{
			Balance: &Balance{
				ID:        "testID1",
				FilterIDs: []string{"testFID1", "testFID2"},
				Type:      MetaAbstract,
				Units:     &Decimal{decimal.New(1234, 3)},
				Weights:   nil,
				UnitFactors: []*UnitFactor{
					{
						Factor: NewDecimal(1, 1),
					},
				},
				Opts: map[string]interface{}{
					MetaBalanceLimit: -1.0,
				},
				CostIncrements: []*CostIncrement{
					{
						Increment:    NewDecimal(int64(time.Duration(time.Second)), 0),
						RecurrentFee: NewDecimal(0, 0),
					},
				},
				AttributeIDs:   []string{"testString1"},
				RateProfileIDs: []string{"testString2"},
			},
			Weight: 23,
		},
		{
			Balance: &Balance{
				ID:        "testID2",
				FilterIDs: []string{"testFID3", "testFID4"},
				Type:      MetaAbstract,
				Units:     &Decimal{decimal.New(1234, 3)},
				Weights:   nil,
				UnitFactors: []*UnitFactor{
					{
						Factor: NewDecimal(1, 1),
					},
				},
				Opts: map[string]interface{}{
					MetaBalanceLimit: -1.0,
				},
				CostIncrements: []*CostIncrement{
					{
						Increment:    NewDecimal(int64(time.Duration(time.Second)), 0),
						RecurrentFee: NewDecimal(0, 0),
					},
				},
				AttributeIDs:   []string{"testString3"},
				RateProfileIDs: []string{"testString4"},
			},
			Weight: 23,
		},
	}

	expected := make([]*Balance, 0)
	for i := range blncS {
		expected = append(expected, blncS[i].Balance)
	}
	received := blncS.Balances()

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", ToJSON(expected), ToJSON(received))
	}
}

func TestEqualsUnitFactor(t *testing.T) {
	uf1 := &UnitFactor{
		FilterIDs: []string{"*string:~*req.Account:1003"},
		Factor:    NewDecimal(10, 0),
	}
	uf2 := &UnitFactor{
		FilterIDs: []string{"*string:~*req.Account:1003"},
		Factor:    NewDecimal(10, 0),
	}
	if !uf1.Equals(uf2) {
		t.Errorf("Unexpected equal result")
	}

	uf1.FilterIDs = []string{"*string:~*req.Account:1004"}
	if uf1.Equals(uf2) {
		t.Errorf("Unexpected equal result")
	}
	uf1.FilterIDs = nil

	if uf1.Equals(uf2) {
		t.Errorf("Unexpected equal result")
	}
	uf1.FilterIDs = []string{"*string:~*req.Account:1003"}

	uf1.Factor = NewDecimal(100, 0)
	if uf1.Equals(uf2) {
		t.Errorf("Unexpected equal result")
	}

	uf1.Factor = nil
	uf2.Factor = nil
	if !uf1.Equals(uf2) {
		t.Errorf("Unexpected equal result")
	}

	uf2.Factor = NewDecimal(10, 0)
	if uf1.Equals(uf2) {
		t.Errorf("Unexpected equal result")
	}
}

func TestAsExtUnitFactor(t *testing.T) {
	uF := &UnitFactor{
		FilterIDs: []string{"test1"},
		Factor:    NewDecimal(123, 2),
	}
	expUf := &ExtUnitFactor{
		FilterIDs: []string{"test1"},
		Factor:    Float64Pointer(1.23),
	}
	if rcv, err := uF.AsExtUnitFactor(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expUf) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expUf), ToJSON(rcv))
	}

	uF = &UnitFactor{
		FilterIDs: []string{},
	}
	expUf = &ExtUnitFactor{
		FilterIDs: []string{},
	}
	if rcv, err := uF.AsExtUnitFactor(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expUf) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expUf), ToJSON(rcv))
	}

	uF.Factor = NewDecimal(int64(math.Inf(1))-1, 0)
	expErr := "cannot convert decimal Factor to float64 "
	if _, err := uF.AsExtUnitFactor(); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+v, received %+v", expErr, err)
	}
}

func TestAsExtAccount(t *testing.T) {
	acc := &Account{
		Tenant:    "tcgrates.org",
		ID:        "TestAsExtAccount",
		FilterIDs: []string{"fltr1", "fltr2"},
		ActivationInterval: &ActivationInterval{
			ActivationTime: time.Date(2021, time.December, 12, 0, 0, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2021, time.May, 12, 10, 0, 0, 0, time.UTC),
		},
		Weights: DynamicWeights{
			{
				FilterIDs: []string{},
				Weight:    20.5,
			},
		},
		Opts: map[string]interface{}{
			Usage: time.Minute,
		},
		Balances: map[string]*Balance{
			"BL1": {
				ID:        "BL1",
				FilterIDs: []string{"fltr3"},
				Weights: DynamicWeights{
					{
						FilterIDs: []string{},
						Weight:    1.5,
					},
				},
				Type:  MetaAbstract,
				Units: &Decimal{decimal.New(300, 2)},
				UnitFactors: []*UnitFactor{
					{
						FilterIDs: []string{"uf_fltr1"},
						Factor:    NewDecimal(13, 4),
					},
				},
				Opts: map[string]interface{}{
					RequestType: Prepaid,
				},
				CostIncrements: []*CostIncrement{
					{
						FilterIDs:    []string{},
						FixedFee:     NewDecimal(int64(time.Minute), 0),
						Increment:    NewDecimal(int64(time.Minute), 0),
						RecurrentFee: NewDecimal(0, 0),
					},
				},
				AttributeIDs:   []string{},
				RateProfileIDs: []string{"RT1"},
			},
		},
		ThresholdIDs: []string{"THRESHOLD1", "THRESHOLD2"},
	}

	expAcc := &ExtAccount{
		Tenant:    "tcgrates.org",
		ID:        "TestAsExtAccount",
		FilterIDs: []string{"fltr1", "fltr2"},
		ActivationInterval: &ActivationInterval{
			ActivationTime: time.Date(2021, time.December, 12, 0, 0, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2021, time.May, 12, 10, 0, 0, 0, time.UTC),
		},
		Weights: DynamicWeights{
			{
				FilterIDs: []string{},
				Weight:    20.5,
			},
		},
		Opts: map[string]interface{}{
			Usage: time.Minute,
		},
		Balances: map[string]*ExtBalance{
			"BL1": {
				ID:        "BL1",
				FilterIDs: []string{"fltr3"},
				Weights: DynamicWeights{
					{
						FilterIDs: []string{},
						Weight:    1.5,
					},
				},
				Type:  MetaAbstract,
				Units: Float64Pointer(3.00),
				UnitFactors: []*ExtUnitFactor{
					{
						FilterIDs: []string{"uf_fltr1"},
						Factor:    Float64Pointer(0.0013),
					},
				},
				Opts: map[string]interface{}{
					RequestType: Prepaid,
				},
				CostIncrements: []*ExtCostIncrement{
					{
						FilterIDs:    []string{},
						FixedFee:     Float64Pointer(float64(time.Minute)),
						Increment:    Float64Pointer(float64(time.Minute)),
						RecurrentFee: Float64Pointer(0),
					},
				},
				AttributeIDs:   []string{},
				RateProfileIDs: []string{"RT1"},
			},
		},
		ThresholdIDs: []string{"THRESHOLD1", "THRESHOLD2"},
	}

	if rcv, err := acc.AsExtAccount(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expAcc) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expAcc), ToJSON(rcv))
	}
}

func TestAsExtAccountCheckErrors(t *testing.T) {
	acc := &Account{
		Tenant:    "tcgrates.org",
		ID:        "TestAsExtAccount",
		FilterIDs: []string{"fltr1", "fltr2"},
		ActivationInterval: &ActivationInterval{
			ActivationTime: time.Date(2021, time.December, 12, 0, 0, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2021, time.May, 12, 10, 0, 0, 0, time.UTC),
		},
		Weights: DynamicWeights{
			{
				FilterIDs: []string{},
				Weight:    20.5,
			},
		},
		Opts: map[string]interface{}{
			Usage: time.Minute,
		},
		Balances: map[string]*Balance{
			"BL1": {
				ID:        "BL1",
				FilterIDs: []string{"fltr3"},
				Weights: DynamicWeights{
					{
						FilterIDs: []string{},
						Weight:    1.5,
					},
				},
				Type:  MetaAbstract,
				Units: NewDecimal(int64(math.Inf(1))-1, 0),
				UnitFactors: []*UnitFactor{
					{
						FilterIDs: []string{"uf_fltr1"},
						Factor:    NewDecimal(13, 4),
					},
				},
				Opts: map[string]interface{}{
					RequestType: Prepaid,
				},
				CostIncrements: []*CostIncrement{
					{
						FilterIDs:    []string{},
						FixedFee:     NewDecimal(int64(time.Minute), 0),
						Increment:    NewDecimal(int64(time.Minute), 0),
						RecurrentFee: NewDecimal(0, 0),
					},
				},
				AttributeIDs:   []string{},
				RateProfileIDs: []string{"RT1"},
			},
		},
		ThresholdIDs: []string{},
	}

	expected := "cannot convert decimal Units to float64 "
	if _, err := acc.AsExtAccount(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, err)
	}
	acc.Balances["BL1"].Units = NewDecimal(0, 0)

	acc.Balances["BL1"].CostIncrements[0].FixedFee = NewDecimal(int64(math.Inf(1))-1, 0)
	expected = "cannot convert decimal FixedFee to float64 "
	if _, err := acc.AsExtAccount(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, err)
	}
	acc.Balances["BL1"].CostIncrements[0].FixedFee = NewDecimal(0, 0)

	acc.Balances["BL1"].CostIncrements[0].Increment = NewDecimal(int64(math.Inf(1))-1, 0)
	expected = "cannot convert decimal Increment to float64 "
	if _, err := acc.AsExtAccount(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, err)
	}
	acc.Balances["BL1"].CostIncrements[0].Increment = NewDecimal(0, 0)

	acc.Balances["BL1"].CostIncrements[0].RecurrentFee = NewDecimal(int64(math.Inf(1))-1, 0)
	expected = "cannot convert decimal RecurrentFee to float64 "
	if _, err := acc.AsExtAccount(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, err)
	}
	acc.Balances["BL1"].CostIncrements[0].RecurrentFee = NewDecimal(0, 0)

	acc.Balances["BL1"].UnitFactors[0].Factor = NewDecimal(int64(math.Inf(1))-1, 0)
	expected = "cannot convert decimal Factor to float64 "
	if _, err := acc.AsExtAccount(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, err)
	}
	acc.Balances["BL1"].UnitFactors[0].Factor = NewDecimal(0, 0)
}

func TestExtAccountEqualsCase1(t *testing.T) {
	eAc := &ExtAccount{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
	}

	extAc := &ExtAccount{
		Tenant: "cgrates.org",
		ID:     "49f2ba",
	}

	if rcv := eAc.Equals(extAc); rcv {
		t.Error("Accounts should not match")
	}
}

func TestExtAccountEqualsCase2(t *testing.T) {
	eAc := &ExtAccount{
		Tenant:    "cgrates.org",
		ID:        "f43a2c",
		FilterIDs: []string{"*string:*req.Account:1001"},
	}

	extAc := &ExtAccount{
		Tenant:    "cgrates.org",
		ID:        "f43a2c",
		FilterIDs: []string{"*string:*req.Account:1003"},
	}

	if rcv := eAc.Equals(extAc); rcv {
		t.Error("Filters should not match")
	}
}

func TestExtAccountEqualsCase3(t *testing.T) {
	eAc := &ExtAccount{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"*string:*req.Account:1003"},
				Weight:    20,
			},
		},
	}

	extAc := &ExtAccount{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"*string:*req.Account:1003"},
				Weight:    10,
			},
		},
	}

	if rcv := eAc.Equals(extAc); rcv {
		t.Error("Weights should not match")
	}
}

func TestExtAccountEqualsCase4(t *testing.T) {
	eAc := &ExtAccount{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Opts: map[string]interface{}{
			"Opt1": "*opt",
		},
	}

	extAc := &ExtAccount{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Opts: map[string]interface{}{
			"Opt1": "*opt2",
		},
	}

	if eAc.Equals(extAc) {
		t.Error("Opts should not match")
	}
}

func TestExtAccountEqualsCase5(t *testing.T) {
	eAc := &ExtAccount{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Balances: map[string]*ExtBalance{
			"*monetary": {
				ID:        "b24d37",
				FilterIDs: []string{},
				Weights: DynamicWeights{
					{
						FilterIDs: []string{"*string:*req.Account:1001"},
						Weight:    10,
					},
				},
				Type:  "*monetary",
				Units: Float64Pointer(3.14),
				UnitFactors: []*ExtUnitFactor{
					{
						FilterIDs: []string{},
						Factor:    Float64Pointer(2.1),
					},
				},
				Opts:           map[string]interface{}{},
				CostIncrements: []*ExtCostIncrement{},
				AttributeIDs:   []string{MetaNone},
				RateProfileIDs: []string{MetaNone},
			},
		},
	}

	extAc := &ExtAccount{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Balances: map[string]*ExtBalance{
			"*monetary": {
				ID:        "b24d37",
				FilterIDs: []string{},
				Weights: DynamicWeights{
					{
						FilterIDs: []string{"*string:*req.Account:1003"},
						Weight:    10,
					},
				},
				Type:  "*monetary",
				Units: Float64Pointer(3.14),
				UnitFactors: []*ExtUnitFactor{
					{
						FilterIDs: []string{},
						Factor:    Float64Pointer(2.1),
					},
				},
				Opts:           map[string]interface{}{},
				CostIncrements: []*ExtCostIncrement{},
				AttributeIDs:   []string{MetaNone},
				RateProfileIDs: []string{MetaNone},
			},
		},
	}

	if rcv := eAc.Equals(extAc); rcv {
		t.Error("Balances should not match")
	}
}

func TestExtAccountEqualsCase6(t *testing.T) {
	eAc := &ExtAccount{
		Tenant:       "cgrates.org",
		ID:           "f43a2c",
		ThresholdIDs: []string{"ACNT_THSD_1003"},
	}

	extAc := &ExtAccount{
		Tenant:       "cgrates.org",
		ID:           "f43a2c",
		ThresholdIDs: []string{"ACNT_THSD_1001"},
	}

	if rcv := eAc.Equals(extAc); rcv {
		t.Error("Thresholds should not match")
	}
}

func TestExtBalanceEqualsCase1(t *testing.T) {
	eBL := &ExtBalance{
		ID: "2f5ba2",
	}

	extBl := &ExtBalance{
		ID: "68d1c5",
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("Balances should not match")
	}
}

func TestExtBalanceEqualsCase2(t *testing.T) {
	eBL := &ExtBalance{
		ID: "2f5ba2",
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"*string:*req.Account:1003"},
				Weight:    10,
			},
		},
	}

	extBl := &ExtBalance{
		ID: "2f5ba2",
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"*string:*req.Account:1003"},
				Weight:    20,
			},
		},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("Balances should not match")
	}
}

func TestExtBalanceEqualsCase3(t *testing.T) {
	eBL := &ExtBalance{
		ID:        "2f5ba2",
		FilterIDs: []string{"*string:*req.Account:1001"},
	}

	extBl := &ExtBalance{
		ID:        "2f5ba2",
		FilterIDs: []string{"*string:*req.Account:1002"},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("Filters should not match")
	}
}

func TestExtBalanceEqualsCase4(t *testing.T) {
	eBL := &ExtBalance{
		ID: "2f5ba2",
		UnitFactors: []*ExtUnitFactor{
			{
				FilterIDs: []string{"*string:*req.Account:1001"},
				Factor:    Float64Pointer(21.7),
			},
		},
	}

	extBl := &ExtBalance{
		ID: "2f5ba2",
		UnitFactors: []*ExtUnitFactor{
			{
				FilterIDs: []string{"*string:*req.Account:1002"},
				Factor:    Float64Pointer(4.20),
			},
		},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("UnitFactors should not match")
	}
}

func TestExtBalanceEqualsCase5(t *testing.T) {
	eBL := &ExtBalance{
		ID: "2f5ba2",
		Opts: map[string]interface{}{
			"Opt1": "*opt",
		},
	}

	extBl := &ExtBalance{
		ID: "2f5ba2",
		Opts: map[string]interface{}{
			"Opt1": "*opt2",
		},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("Opts should not match")
	}
}

func TestExtBalanceEqualsCase6(t *testing.T) {
	eBL := &ExtBalance{
		ID: "2f5ba2",
		CostIncrements: []*ExtCostIncrement{
			{
				FilterIDs:    []string{},
				Increment:    Float64Pointer(0.2),
				FixedFee:     Float64Pointer(0.4),
				RecurrentFee: Float64Pointer(1.2),
			},
		},
	}

	extBl := &ExtBalance{
		ID: "2f5ba2",
		CostIncrements: []*ExtCostIncrement{
			{
				FilterIDs:    []string{},
				Increment:    Float64Pointer(0.3),
				FixedFee:     Float64Pointer(0.4),
				RecurrentFee: Float64Pointer(1.2),
			},
		},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("CostIncrements should not match")
	}
}

func TestExtBalanceEqualsCase7(t *testing.T) {
	eBL := &ExtBalance{
		ID:           "2f5ba2",
		AttributeIDs: []string{"ATTR_ID_1001"},
	}

	extBl := &ExtBalance{
		ID:           "2f5ba2",
		AttributeIDs: []string{"ATTR_ID_1003"},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("Attributes should not match")
	}
}

func TestExtBalanceEqualsCase8(t *testing.T) {
	eBL := &ExtBalance{
		ID:             "2f5ba2",
		RateProfileIDs: []string{"RP_1001"},
	}

	extBl := &ExtBalance{
		ID:             "2f5ba2",
		RateProfileIDs: []string{"RP_1002"},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("RateProfiles should not match")
	}
}

func TestBalanceEqualsCase1(t *testing.T) {
	eBL := &Balance{
		ID: "2f5ba2",
	}

	extBl := &Balance{
		ID: "68d1c5",
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("Balances should not match")
	}
}

func TestBalanceEqualsCase2(t *testing.T) {
	eBL := &Balance{
		ID:    "2f5ba2",
		Units: NewDecimal(53, 0),
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"*string:*req.Account:1001"},
				Weight:    10,
			},
		},
	}

	extBl := &Balance{
		ID:    "2f5ba2",
		Units: NewDecimal(53, 0),
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"*string:*req.Account:1003"},
				Weight:    20,
			},
		},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("Weights should not match")
	}
}

func TestBalanceEqualsCase3(t *testing.T) {
	eBL := &Balance{
		ID:        "2f5ba2",
		FilterIDs: []string{"*string:*req.Account:1001"},
		Units:     NewDecimal(53, 0),
	}

	extBl := &Balance{
		ID:        "2f5ba2",
		FilterIDs: []string{"*string:*req.Account:1002"},
		Units:     NewDecimal(53, 0),
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("Filters should not match")
	}
}

func TestBalanceEqualsCase4(t *testing.T) {
	eBL := &Balance{
		ID:    "2f5ba2",
		Units: NewDecimal(53, 0),
		UnitFactors: []*UnitFactor{
			{
				FilterIDs: []string{"*string:*req.Account:1001"},
				Factor:    NewDecimal(2, 0),
			},
		},
	}

	extBl := &Balance{
		ID:    "2f5ba2",
		Units: NewDecimal(53, 0),
		UnitFactors: []*UnitFactor{
			{
				FilterIDs: []string{"*string:*req.Account:1002"},
				Factor:    NewDecimal(42, 0),
			},
		},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("UnitFactors should not match")
	}
}

func TestBalanceEqualsCase5(t *testing.T) {
	eBL := &Balance{
		ID:    "2f5ba2",
		Units: NewDecimal(53, 0),
		Opts: map[string]interface{}{
			"Opt1": "*opt",
		},
	}

	extBl := &Balance{
		ID:    "2f5ba2",
		Units: NewDecimal(53, 0),
		Opts: map[string]interface{}{
			"Opt1": "*opt2",
		},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("Opts should not match")
	}
}

func TestBalanceEqualsCase6(t *testing.T) {
	eBL := &Balance{
		ID:    "2f5ba2",
		Units: NewDecimal(53, 0),
		CostIncrements: []*CostIncrement{
			{
				FilterIDs:    []string{},
				Increment:    NewDecimal(1, 0),
				FixedFee:     NewDecimal(3, 0),
				RecurrentFee: NewDecimal(7, 0),
			},
		},
	}

	extBl := &Balance{
		ID:    "2f5ba2",
		Units: NewDecimal(53, 0),
		CostIncrements: []*CostIncrement{
			{
				FilterIDs:    []string{},
				Increment:    NewDecimal(1, 0),
				FixedFee:     NewDecimal(2, 0),
				RecurrentFee: NewDecimal(10, 0),
			},
		},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("CostIncrements should not match")
	}
}

func TestBalanceEqualsCase7(t *testing.T) {
	eBL := &Balance{
		ID:           "2f5ba2",
		Units:        NewDecimal(53, 0),
		AttributeIDs: []string{"ATTR_ID_1001"},
	}

	extBl := &Balance{
		ID:           "2f5ba2",
		Units:        NewDecimal(53, 0),
		AttributeIDs: []string{"ATTR_ID_1003"},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("Attributes should not match")
	}
}

func TestBalanceEqualsCase8(t *testing.T) {
	eBL := &Balance{
		ID:             "2f5ba2",
		Units:          NewDecimal(53, 0),
		RateProfileIDs: []string{"RP_1001"},
	}

	extBl := &Balance{
		ID:             "2f5ba2",
		Units:          NewDecimal(53, 0),
		RateProfileIDs: []string{"RP_1002"},
	}

	if rcv := eBL.Equals(extBl); rcv {
		t.Error("RateProfiles should not match")
	}
}

func TestExtCostIncrementCase1(t *testing.T) {
	eCi := &ExtCostIncrement{
		FilterIDs:    []string{},
		Increment:    Float64Pointer(0.2),
		FixedFee:     Float64Pointer(0.4),
		RecurrentFee: Float64Pointer(1.2),
	}

	extCi := &ExtCostIncrement{
		FilterIDs:    []string{"*string:*req.Account:1002"},
		Increment:    Float64Pointer(0.2),
		FixedFee:     Float64Pointer(0.4),
		RecurrentFee: Float64Pointer(1.2),
	}

	if rcv := eCi.Equals(extCi); rcv {
		t.Error("RateProfiles should not match")
	}
}

func TestExtCostIncrementCase2(t *testing.T) {
	eCi := &ExtCostIncrement{
		FilterIDs:    []string{"*string:*req.Account:1001"},
		Increment:    Float64Pointer(0.2),
		FixedFee:     Float64Pointer(0.4),
		RecurrentFee: Float64Pointer(1.2),
	}

	extCi := &ExtCostIncrement{
		FilterIDs:    []string{"*string:*req.Account:1002"},
		Increment:    Float64Pointer(0.2),
		FixedFee:     Float64Pointer(0.4),
		RecurrentFee: Float64Pointer(1.2),
	}

	if rcv := eCi.Equals(extCi); rcv {
		t.Error("RateProfiles should not match")
	}
}

func TestCostIncrementCase1(t *testing.T) {
	eCi := &CostIncrement{
		FilterIDs:    []string{},
		Increment:    NewDecimal(1, 0),
		FixedFee:     NewDecimal(3, 0),
		RecurrentFee: NewDecimal(7, 0),
	}

	extCi := &CostIncrement{
		FilterIDs:    []string{"*string:*req.Account:1002"},
		Increment:    NewDecimal(1, 0),
		FixedFee:     NewDecimal(3, 0),
		RecurrentFee: NewDecimal(7, 0),
	}

	if rcv := eCi.Equals(extCi); rcv {
		t.Error("RateProfiles should not match")
	}
}

func TestCostIncrementCase2(t *testing.T) {
	eCi := &CostIncrement{
		FilterIDs:    []string{"*string:*req.Account:1001"},
		Increment:    NewDecimal(1, 0),
		FixedFee:     NewDecimal(3, 0),
		RecurrentFee: NewDecimal(7, 0),
	}

	extCi := &CostIncrement{
		FilterIDs:    []string{"*string:*req.Account:1002"},
		Increment:    NewDecimal(1, 0),
		FixedFee:     NewDecimal(3, 0),
		RecurrentFee: NewDecimal(7, 0),
	}

	if rcv := eCi.Equals(extCi); rcv {
		t.Error("RateProfiles should not match")
	}
}

func TestExtUnitFactorCase1(t *testing.T) {
	eUf := &ExtUnitFactor{
		FilterIDs: []string{},
		Factor:    Float64Pointer(2.1),
	}

	extUf := &ExtUnitFactor{
		FilterIDs: []string{"*string:*req.Account:1002"},
		Factor:    Float64Pointer(2.1),
	}

	if rcv := eUf.Equals(extUf); rcv {
		t.Error("RateProfiles should not match")
	}
}

func TestExtUnitFactorCase2(t *testing.T) {
	eUf := &ExtUnitFactor{
		FilterIDs: []string{"*string:*req.Account:1001"},
		Factor:    Float64Pointer(2.1),
	}

	extUf := &ExtUnitFactor{
		FilterIDs: []string{"*string:*req.Account:1002"},
		Factor:    Float64Pointer(2.1),
	}

	if rcv := eUf.Equals(extUf); rcv {
		t.Error("RateProfiles should not match")
	}
}

func TestAccountEqualsCase1(t *testing.T) {
	eAc := &Account{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
	}

	extAc := &Account{
		Tenant: "cgrates.org",
		ID:     "49f2ba",
	}

	if rcv := eAc.Equals(extAc); rcv {
		t.Error("Accounts should not match")
	}
}

func TestAccountEqualsCase2(t *testing.T) {
	eAc := &Account{
		Tenant:    "cgrates.org",
		ID:        "f43a2c",
		FilterIDs: []string{"*string:*req.Account:1001"},
	}

	extAc := &Account{
		Tenant:    "cgrates.org",
		ID:        "f43a2c",
		FilterIDs: []string{"*string:*req.Account:1003"},
	}

	if rcv := eAc.Equals(extAc); rcv {
		t.Error("Filters should not match")
	}
}

func TestAccountEqualsCase3(t *testing.T) {
	eAc := &Account{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"*string:*req.Account:1003"},
				Weight:    20,
			},
		},
	}

	extAc := &Account{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"*string:*req.Account:1003"},
				Weight:    10,
			},
		},
	}

	if rcv := eAc.Equals(extAc); rcv {
		t.Error("Weights should not match")
	}
}

func TestAccountEqualsCase4(t *testing.T) {
	eAc := &Account{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Opts: map[string]interface{}{
			"Opt1": "*opt",
		},
	}

	extAc := &Account{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Opts: map[string]interface{}{
			"Opt1": "*opt2",
		},
	}

	if eAc.Equals(extAc) {
		t.Error("Opts should not match")
	}
}

func TestAccountEqualsCase5(t *testing.T) {
	eAc := &Account{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Balances: map[string]*Balance{
			"*monetary": {
				ID:        "b24d37",
				FilterIDs: []string{},
				Weights: DynamicWeights{
					{
						FilterIDs: []string{"*string:*req.Account:1001"},
						Weight:    10,
					},
				},
				Type:  "*monetary",
				Units: NewDecimal(42, 1),
				UnitFactors: []*UnitFactor{
					{
						FilterIDs: []string{},
						Factor:    NewDecimal(2, 1),
					},
				},
				Opts:           map[string]interface{}{},
				CostIncrements: []*CostIncrement{},
				AttributeIDs:   []string{MetaNone},
				RateProfileIDs: []string{MetaNone},
			},
		},
	}

	extAc := &Account{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Balances: map[string]*Balance{
			"*monetary": {
				ID:        "b24d37",
				FilterIDs: []string{},
				Weights: DynamicWeights{
					{
						FilterIDs: []string{"*string:*req.Account:1001"},
						Weight:    10,
					},
				},
				Type:  "*monetary",
				Units: NewDecimal(65, 1),
				UnitFactors: []*UnitFactor{
					{
						FilterIDs: []string{},
						Factor:    NewDecimal(3, 1),
					},
				},
				Opts:           map[string]interface{}{},
				CostIncrements: []*CostIncrement{},
				AttributeIDs:   []string{MetaNone},
				RateProfileIDs: []string{MetaNone},
			},
		},
	}

	if rcv := eAc.Equals(extAc); rcv {
		t.Error("Balances should not match")
	}
}

func TestAccountEqualsCase6(t *testing.T) {
	eAc := &Account{
		Tenant:       "cgrates.org",
		ID:           "f43a2c",
		ThresholdIDs: []string{"ACNT_THSD_1003"},
	}

	extAc := &Account{
		Tenant:       "cgrates.org",
		ID:           "f43a2c",
		ThresholdIDs: []string{"ACNT_THSD_1001"},
	}

	if rcv := eAc.Equals(extAc); rcv {
		t.Error("Thresholds should not match")
	}
}
