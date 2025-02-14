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
		Opts: map[string]any{
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
		FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2020-07-21T10:00:00Z|2020-07-22T10:00:00Z"},
		Weights: DynamicWeights{
			{
				Weight: 2.4,
			},
		},
		Opts: map[string]any{
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
				Opts: map[string]any{
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
				FilterIDs: []string{"testFID1", "testFID2", "*ai:~*req.AnswerTime:2020-04-12T00:00:00Z|2020-04-12T10:00:00Z"},
				Weights:   nil,
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
				FilterIDs: []string{"testFID1", "testFID2", "*ai:~*req.AnswerTime:2020-04-12T00:00:00Z|2020-04-12T10:00:00Z"},
				Weights:   nil,
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
				FilterIDs: []string{"testFID1", "testFID2", "*ai:~*req.AnswerTime:2020-04-12T00:00:00Z|2020-04-12T10:00:00Z"},
				Weights:   nil,
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
				FilterIDs: []string{"testFID1", "testFID2", "*ai:~*req.AnswerTime:2020-04-12T00:00:00Z|2020-04-12T10:00:00Z"},
				Weights:   nil,
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
		Opts: map[string]any{
			"Opt1": "*opt",
		},
	}

	extBl := &Balance{
		ID:    "2f5ba2",
		Units: NewDecimal(53, 0),
		Opts: map[string]any{
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
		Opts: map[string]any{
			"Opt1": "*opt",
		},
	}

	extAc := &Account{
		Tenant: "cgrates.org",
		ID:     "f43a2c",
		Opts: map[string]any{
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
				Opts:           map[string]any{},
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
				Opts:           map[string]any{},
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

func TestAccountClone(t *testing.T) {
	aI := &ActivationInterval{
		ActivationTime: time.Date(2021, 10, 7, 13, 0, 0, 0, time.Local),
		ExpiryTime:     time.Date(2021, 10, 7, 18, 0, 0, 0, time.Local),
	}

	expAI := &ActivationInterval{
		ActivationTime: time.Date(2021, 10, 7, 13, 0, 0, 0, time.Local),
		ExpiryTime:     time.Date(2021, 10, 7, 18, 0, 0, 0, time.Local),
	}

	rcv := aI.Clone()
	if !reflect.DeepEqual(rcv, expAI) {
		t.Errorf("Expected %v \n but received \n %v", expAI, rcv)
	}

	aI = nil
	if err := aI.Clone(); err != nil {
		t.Error(err)
	}
}

func TestAccountSet(t *testing.T) {
	acc := Account{Balances: map[string]*Balance{}}
	exp := Account{
		Tenant:       "cgrates.org",
		ID:           "ID",
		FilterIDs:    []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:      DynamicWeights{{}},
		Blockers:     DynamicBlockers{{}},
		ThresholdIDs: []string{"TH1"},
		Opts: map[string]any{
			"bal":  "val",
			"bal2": "val2",
			"bal3": "val2",
			"bal4": "val2",
			"bal5": MapStorage{"bal6": "val3"},
		},
		Balances: map[string]*Balance{
			"bal1": {
				ID:   "bal1",
				Type: MetaConcrete,
				Opts: map[string]any{
					"bal7":  "val3",
					"bal8":  MapStorage{"bal9": "val3"},
					"bal10": "val3",
				},
				Units:          NewDecimal(10, 0),
				FilterIDs:      []string{"*string:~*req.Account:1001"},
				AttributeIDs:   []string{"Attr1", "Attr2"},
				RateProfileIDs: []string{"Attr1", "Attr2"},
				Weights:        DynamicWeights{{Weight: 10}},
				Blockers:       DynamicBlockers{{Blocker: false}},
				UnitFactors: []*UnitFactor{
					{FilterIDs: []string{"fltr1"}, Factor: NewDecimal(10, 0)},
					{FilterIDs: []string{"fltr1"}, Factor: NewDecimal(101, 0)},
				},
				CostIncrements: []*CostIncrement{
					{FilterIDs: []string{"fltr1"}, Increment: NewDecimal(10, 0), FixedFee: NewDecimal(10, 0), RecurrentFee: NewDecimal(10, 0)},
					{FilterIDs: []string{"fltr1"}, Increment: NewDecimal(101, 0), FixedFee: NewDecimal(101, 0), RecurrentFee: NewDecimal(101, 0)},
				},
			},
		},
	}
	if err := acc.Set([]string{}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := acc.Set([]string{"NotAField"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := acc.Set([]string{"NotAField", "1"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	expErr := `malformed map pair: <"bal">`
	if err := acc.Set([]string{Opts}, "bal", false); err == nil || err.Error() != expErr {
		t.Error(err)
	}
	if err := acc.Set([]string{Opts}, "bal:val;bal2:val2", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Opts, "bal3"}, "val2", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Opts + "[bal4]"}, "val2", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Opts + "[bal5]", "bal6"}, "val3", false); err != nil {
		t.Error(err)
	}

	if err := acc.Set([]string{Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{ThresholdIDs}, "TH1", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Weights}, ";0", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Blockers}, ";0", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances + "[bal1]", ID}, "bal1", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", Type}, MetaConcrete, false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", Opts}, "", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", Opts + "bal7]"}, "val3", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", Opts + "bal7]", ""}, "val3", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", Opts + "[bal7]"}, "val3", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", Opts + "[bal8]", "bal9"}, "val3", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", Opts, "bal10"}, "val3", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", FilterIDs}, "*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", AttributeIDs}, "Attr1;Attr2", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", RateProfileIDs}, "Attr1;Attr2", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", Units}, "10", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", Blockers}, ";false", false); err != nil {
		t.Error(err)
	}

	expErr = `invalid key: <1> for BalanceUnitFactors`
	if err := acc.Set([]string{Balances, "bal1", UnitFactors}, "1", false); err == nil || err.Error() != expErr {
		t.Error(err)
	}
	expErr = `can't convert <a> to decimal`
	if err := acc.Set([]string{Balances, "bal1", UnitFactors}, "a;a", false); err == nil || err.Error() != expErr {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", UnitFactors}, "fltr1;10", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", UnitFactors, "Wrong"}, "fltr1;10", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", UnitFactors, Factor}, "101", true); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", UnitFactors, FilterIDs}, "fltr1", false); err != nil {
		t.Error(err)
	}

	expErr = `invalid key: <1> for BalanceCostIncrements`
	if err := acc.Set([]string{Balances, "bal1", CostIncrements}, "1", false); err == nil || err.Error() != expErr {
		t.Error(err)
	}
	expErr = `can't convert <a> to decimal`
	if err := acc.Set([]string{Balances, "bal1", CostIncrements}, "fltr1;10;a;10", false); err == nil || err.Error() != expErr {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", CostIncrements}, "fltr1;a;10;10", false); err == nil || err.Error() != expErr {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", CostIncrements}, "fltr1;10;10;a", false); err == nil || err.Error() != expErr {
		t.Error(err)
	}

	if err := acc.Set([]string{Balances, "bal1", CostIncrements}, "fltr1;10;10;10", false); err != nil {
		t.Error(err)
	}

	if err := acc.Set([]string{Balances, "bal1", CostIncrements, FixedFee}, "101", true); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", CostIncrements, RecurrentFee}, "101", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", CostIncrements, Increment}, "101", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", CostIncrements, FilterIDs}, "fltr1", false); err != nil {
		t.Error(err)
	}
	if err := acc.Set([]string{Balances, "bal1", CostIncrements, "Wrong"}, "fltr1", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := acc.Balances["bal1"].Set(nil, "fltr1", false); err != ErrWrongPath {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, acc) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(acc))
	}
}
func TestAccountFieldAsInterface(t *testing.T) {
	acc := Account{
		Tenant:       "cgrates.org",
		ID:           "ID",
		FilterIDs:    []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:      DynamicWeights{{}},
		Blockers:     DynamicBlockers{{}},
		ThresholdIDs: []string{"TH1"},
		Opts: map[string]any{
			"bal":  "val",
			"bal2": "val2",
			"bal3": "val2",
			"bal4": "val2",
			"bal5": MapStorage{"bal6": "val3"},
		},
		Balances: map[string]*Balance{
			"bal1": {
				ID:   "bal1",
				Type: MetaConcrete,
				Opts: map[string]any{
					"bal7":  "val3",
					"bal8":  MapStorage{"bal9": "val3"},
					"bal10": "val3",
				},
				Units:          NewDecimal(10, 0),
				FilterIDs:      []string{"*string:~*req.Account:1001"},
				AttributeIDs:   []string{"Attr1", "Attr2"},
				RateProfileIDs: []string{"Attr1", "Attr2"},
				Weights:        DynamicWeights{{Weight: 10}},
				Blockers:       DynamicBlockers{{Blocker: false}},
				UnitFactors: []*UnitFactor{
					{FilterIDs: []string{"fltr1"}, Factor: NewDecimal(10, 0)},
					{FilterIDs: []string{"fltr1"}, Factor: NewDecimal(101, 0)},
				},
				CostIncrements: []*CostIncrement{
					{FilterIDs: []string{"fltr1"}, Increment: NewDecimal(10, 0), FixedFee: NewDecimal(10, 0), RecurrentFee: NewDecimal(10, 0)},
					{FilterIDs: []string{"fltr1"}, Increment: NewDecimal(101, 0), FixedFee: NewDecimal(101, 0), RecurrentFee: NewDecimal(101, 0)},
				},
			},
		},
	}
	if _, err := acc.FieldAsInterface(nil); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{"field"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{"field", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Opts + "[f]"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Opts + "[f]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := acc.FieldAsInterface([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ";0"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if val, err := acc.FieldAsInterface([]string{Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := ";false"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if val, err := acc.FieldAsInterface([]string{Opts}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Opts; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := acc.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{ThresholdIDs}); err != nil {
		t.Fatal(err)
	} else if exp := acc.ThresholdIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := acc.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{ThresholdIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := acc.ThresholdIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]"}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances, "bal1"}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	expErrMsg := `strconv.Atoi: parsing "a": invalid syntax`
	if _, err := acc.FieldAsInterface([]string{FilterIDs + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := acc.FieldAsInterface([]string{ThresholdIDs + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if _, err := acc.FieldAsInterface([]string{Balances + "[]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", "", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", Opts}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].Opts; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", Type}); err != nil {
		t.Fatal(err)
	} else if exp := MetaConcrete; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", Units}); err != nil {
		t.Fatal(err)
	} else if exp := NewDecimal(10, 0); exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ";10"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := ";false"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", AttributeIDs}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].AttributeIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", RateProfileIDs}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].RateProfileIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", AttributeIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].AttributeIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances, "bal1", RateProfileIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].RateProfileIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances, "bal1", ID}); err != nil {
		t.Fatal(err)
	} else if exp := "bal1"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", UnitFactors}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].UnitFactors; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].CostIncrements; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", UnitFactors + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].UnitFactors[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].CostIncrements[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", Opts + "[0]"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", Opts + "[0]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", FilterIDs + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", AttributeIDs + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", RateProfileIDs + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", UnitFactors + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", UnitFactors + "[a]", ""}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements + "[a]", ""}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", UnitFactors, ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements, ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", UnitFactors + "[4]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements + "[4]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", UnitFactors + "[0]", "", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements + "[0]", "", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", UnitFactors + "[0]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements + "[0]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}

	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", UnitFactors + "[0]", FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].UnitFactors[0].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", UnitFactors + "[0]", FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].UnitFactors[0].FilterIDs[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", UnitFactors + "[0]", Factor}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].UnitFactors[0].Factor; exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements + "[0]", FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].CostIncrements[0].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements + "[0]", FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].CostIncrements[0].FilterIDs[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements + "[0]", Increment}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].CostIncrements[0].Increment; exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements + "[0]", FixedFee}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].CostIncrements[0].FixedFee; exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := acc.FieldAsInterface([]string{Balances + "[bal1]", CostIncrements + "[0]", RecurrentFee}); err != nil {
		t.Fatal(err)
	} else if exp := acc.Balances["bal1"].CostIncrements[0].RecurrentFee; exp.Cmp(val.(*Decimal).Big) != 0 {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := acc.FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := acc.FieldAsString([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := acc.String(), ToJSON(acc); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := acc.Balances["bal1"].FieldAsString([]string{}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := acc.Balances["bal1"].FieldAsString([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := "bal1"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := acc.Balances["bal1"].String(), ToJSON(acc.Balances["bal1"]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := acc.Balances["bal1"].UnitFactors[0].FieldAsString([]string{}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := acc.Balances["bal1"].UnitFactors[0].FieldAsString([]string{FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := "fltr1"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := acc.Balances["bal1"].UnitFactors[0].String(), ToJSON(acc.Balances["bal1"].UnitFactors[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := acc.Balances["bal1"].CostIncrements[0].FieldAsString([]string{}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := acc.Balances["bal1"].CostIncrements[0].FieldAsString([]string{FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := "fltr1"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := acc.Balances["bal1"].CostIncrements[0].String(), ToJSON(acc.Balances["bal1"].CostIncrements[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
}

func TestAccountMerge(t *testing.T) {
	acc := &Account{
		Opts: make(map[string]any),
		Balances: map[string]*Balance{
			"bal1": {
				Type: MetaConcrete,
				Opts: make(map[string]any),
			},
			"bal3": {},
		},
	}
	exp := &Account{
		Tenant:       "cgrates.org",
		ID:           "ID",
		FilterIDs:    []string{"fltr1"},
		Weights:      DynamicWeights{{}},
		Opts:         map[string]any{"opt1": "val"},
		ThresholdIDs: []string{"TH1"},
		Balances: map[string]*Balance{
			"bal1": {
				ID:             "bal1",
				Type:           MetaConcrete,
				FilterIDs:      []string{"fltr1"},
				Weights:        DynamicWeights{{}},
				Units:          DecimalNaN,
				Opts:           map[string]any{"opt1": "val"},
				AttributeIDs:   []string{"ATTR1"},
				RateProfileIDs: []string{"RT1"},
				UnitFactors:    []*UnitFactor{{}},
				CostIncrements: []*CostIncrement{{}},
			},
			"bal2": {},
			"bal3": {Type: MetaConcrete},
		},
	}
	if acc.Merge(&Account{
		Tenant:       "cgrates.org",
		ID:           "ID",
		FilterIDs:    []string{"fltr1"},
		Weights:      DynamicWeights{{}},
		Opts:         map[string]any{"opt1": "val"},
		ThresholdIDs: []string{"TH1"},
		Balances: map[string]*Balance{
			"bal3": {Type: MetaConcrete},
			"bal2": {},
			"bal1": {
				ID:             "bal1",
				FilterIDs:      []string{"fltr1"},
				Weights:        DynamicWeights{{}},
				Units:          DecimalNaN,
				Opts:           map[string]any{"opt1": "val"},
				AttributeIDs:   []string{"ATTR1"},
				RateProfileIDs: []string{"RT1"},
				UnitFactors:    []*UnitFactor{{}},
				CostIncrements: []*CostIncrement{{}},
			},
		},
	}); !reflect.DeepEqual(exp, acc) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(acc))
	}

}

func TestAccountUnitFactorCloneEmpty(t *testing.T) {
	uF := &UnitFactor{}
	if rcv := uF.Clone(); !reflect.DeepEqual(rcv, uF) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(uF), ToJSON(rcv))
	}
}

func TestAccountUnitFactorClone(t *testing.T) {
	uF := &UnitFactor{
		FilterIDs: []string{"FLTR1", "FLTR2"},
		Factor:    NewDecimalFromFloat64(1.234),
	}
	if rcv := uF.Clone(); !reflect.DeepEqual(rcv, uF) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(uF), ToJSON(rcv))
	}
}

func TestAccountObjectID(t *testing.T) {
	apWws := &AccountsWithWeight{
		{
			Account: &Account{
				ID: "ID",
			},
		},
	}

	exp := &Account{
		ID: "ID",
	}

	if val := apWws.Account("ID"); !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %v ,received %v", exp, val)
	}

}
