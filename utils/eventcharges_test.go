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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
)

func TestECNewEventCharges(t *testing.T) {
	expected := &EventCharges{
		Accounting:  make(map[string]*AccountCharge),
		UnitFactors: make(map[string]*UnitFactor),
		Rating:      make(map[string]*RateSInterval),
		Rates:       make(map[string]*IntervalRate),
		Accounts:    make(map[string]*Account),
	}
	received := NewEventCharges()

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestECMergeAbstractsEmpty(t *testing.T) {
	ec1 := &EventCharges{
		Abstracts: &Decimal{decimal.New(1, 1)},
		Concretes: &Decimal{decimal.New(1, 1)},
	}

	ec2 := &EventCharges{
		Abstracts: &Decimal{decimal.New(2, 1)},
		Concretes: &Decimal{decimal.New(2, 1)},
	}

	received := &EventCharges{}
	expected := &EventCharges{
		Abstracts: &Decimal{decimal.New(3, 1)},
		Concretes: &Decimal{decimal.New(3, 1)},
	}
	received.Merge(ec1, ec2)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestECMergeAbstracts(t *testing.T) {
	ec1 := &EventCharges{
		Abstracts: &Decimal{decimal.New(1, 1)},
		Concretes: &Decimal{decimal.New(1, 1)},
	}

	ec2 := &EventCharges{
		Abstracts: &Decimal{decimal.New(2, 1)},
		Concretes: &Decimal{decimal.New(2, 1)},
	}

	received := &EventCharges{
		Abstracts: &Decimal{decimal.New(3, 1)},
		Concretes: &Decimal{decimal.New(3, 1)},
	}
	expected := &EventCharges{
		Abstracts: &Decimal{decimal.New(6, 1)},
		Concretes: &Decimal{decimal.New(6, 1)},
	}

	received.Merge(ec1, ec2)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

/*
func TestEqualsAccountCharge(t *testing.T) {
	accCharge1 := &AccountCharge{
		AccountID:       "AccountID1",
		BalanceID:       "BalanceID1",
		Units:           NewDecimal(20, 0),
		BalanceLimit:    NewDecimal(40, 0),
		UnitFactorID:    "UF1",
		AttributeIDs:    []string{"ID1", "ID2"},
		RatingID:        "RatingID1",
		JoinedChargeIDs: []string{"chID1"},
	}
	accCharge2 := &AccountCharge{
		AccountID:       "AccountID1",
		BalanceID:       "BalanceID1",
		Units:           NewDecimal(20, 0),
		BalanceLimit:    NewDecimal(40, 0),
		UnitFactorID:    "UF1",
		AttributeIDs:    []string{"ID1", "ID2"},
		RatingID:        "RatingID1",
		JoinedChargeIDs: []string{"chID1"},
	}
	if !accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are not equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}

	// not equal for AccountID
	accCharge1.AccountID = "test"
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge1.AccountID = "AccountID1"

	accCharge2.AccountID = "test"
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge2.AccountID = "AccountID1"

	// not equal for BalanceID
	accCharge1.BalanceID = "test"
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge1.BalanceID = "AccountID1"

	accCharge2.BalanceID = "test"
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge2.BalanceID = "AccountID1"

	// not equal for BalanceLimit
	accCharge1.BalanceLimit = NewDecimal(35, 0)
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge1.BalanceLimit = NewDecimal(40, 0)

	accCharge2.BalanceLimit = NewDecimal(35, 0)
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge2.BalanceLimit = NewDecimal(40, 0)

	// not equal for Units
	accCharge1.Units = NewDecimal(35, 0)
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge1.Units = NewDecimal(20, 0)

	accCharge2.Units = NewDecimal(35, 0)
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge2.Units = NewDecimal(20, 0)

	// not equal for AttributeIDs
	accCharge1.AttributeIDs = []string{"ID1"}
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge1.AttributeIDs = []string{"ID1", "ID2"}

	accCharge2.AttributeIDs = []string{"ID1"}
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge2.AttributeIDs = []string{"ID1", "ID2"}

	// not equal for JoinedChargeIDs
	accCharge1.JoinedChargeIDs = []string{}
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge1.JoinedChargeIDs = []string{"chID1"}

	accCharge2.JoinedChargeIDs = []string{}
	if accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
	accCharge2.JoinedChargeIDs = []string{"chID1"}

	//both units and BalanceLimit are nil will be equal
	accCharge1.Units = nil
	accCharge2.Units = nil
	if !accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are not equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}

	accCharge1.BalanceLimit = nil
	accCharge2.BalanceLimit = nil
	if !accCharge1.Equals(accCharge2) {
		t.Errorf("Charge %+v and %+v are not equal", ToJSON(accCharge1), ToJSON(accCharge2))
	}
}
*/

func TestEventChargesEquals(t *testing.T) {
	eEvChgs := &EventCharges{
		Abstracts: NewDecimal(47500, 3),
		Concretes: NewDecimal(515, 2),
		Charges: []*ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*AccountCharge{
			"THIS_GENUUID1": {
				AccountID:    "TestEventChargesEquals",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(8, 1),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR1",
			},
			"GENUUID1": {
				AccountID:       "TestEventChargesEquals",
				BalanceID:       "ABSTRACT2",
				BalanceLimit:    NewDecimal(0, 0),
				RatingID:        "GENUUID_RATING1",
				JoinedChargeIDs: []string{"THIS_GENUUID1"},
			},
			"GENUUID2": {
				AccountID:    "TestEventChargesEquals",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(2, 0),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR2",
				RatingID:     "ID_FOR_RATING",
				AttributeIDs: []string{"ATTR1", "ATTR2"},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"GENUUID_FACTOR1": {
				Factor:    NewDecimal(100, 0),
				FilterIDs: []string{"*string:~*req.Account:1003"},
			},
			"GENUUID_FACTOR2": {
				Factor: NewDecimal(200, 0),
			},
		},
		Rating: map[string]*RateSInterval{
			"GENUUID_RATING1": {
				Increments: []*RateSIncrement{
					{
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
			"GENUUID_RATING2": {
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(4, 2),
						Usage:             NewDecimal(int64(30*time.Second), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(0, 0),
				CompressFactor: 2,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				FixedFee:      NewDecimal(4, 1),
				RecurrentFee:  NewDecimal(24, 1),
			},
		},
		Accounts: map[string]*Account{
			"ACC1": {
				Tenant:    CGRateSorg,
				ID:        "account_1",
				FilterIDs: []string{"*string:~*req.Account:1003"},
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
					},
				},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				Balances: map[string]*Balance{
					"bal1": {
						ID:        "BAL1",
						FilterIDs: []string{"*string:~*req.Account:1003"},
						Weights: []*DynamicWeight{
							{
								Weight: 25,
							},
						},
						Type:  MetaAbstract,
						Units: NewDecimal(int64(30*time.Second), 0),
						UnitFactors: []*UnitFactor{
							{
								Factor:    NewDecimal(100, 0),
								FilterIDs: []string{"*string:~*req.Account:1003"},
							},
							{
								Factor: NewDecimal(200, 0),
							},
						},
						CostIncrements: []*CostIncrement{
							{
								Increment:    NewDecimal(int64(time.Second), 0),
								RecurrentFee: NewDecimal(5, 0),
							},
							{
								FilterIDs:    []string{"*string:~*req.Account:1003"},
								Increment:    NewDecimal(int64(2*time.Second), 0),
								FixedFee:     NewDecimal(1, 0),
								RecurrentFee: NewDecimal(5, 0),
							},
						},
						AttributeIDs: []string{"ATTRIBUTE1"},
					},
					"bal2": {
						ID:        "BAL2",
						FilterIDs: []string{"*string:~*req.Account:1004"},
						Weights: []*DynamicWeight{
							{
								Weight: 25,
							},
						},
						Type:  MetaConcrete,
						Units: NewDecimal(2000, 0),
						UnitFactors: []*UnitFactor{
							{
								Factor: NewDecimal(200, 0),
							},
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"*string:~*req.Account:1004"},
								Increment:    NewDecimal(int64(2*time.Second), 0),
								FixedFee:     NewDecimal(1, 0),
								RecurrentFee: NewDecimal(5, 0),
							},
						},
						AttributeIDs:   []string{"ATTRIBUTE1"},
						RateProfileIDs: []string{"RATE1", "RATE2"},
					},
				},
				ThresholdIDs: []string{},
			},
			"ACC2": {
				Tenant: CGRateSorg,
				ID:     "account_2",
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
				},
				FilterIDs: []string{"*ai:~*req.AnswerTime:2020-10-10T10:00:00Z"},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				ThresholdIDs: []string{},
			},
		},
	}

	expectedEqual := &EventCharges{
		Abstracts: NewDecimal(47500, 3),
		Concretes: NewDecimal(515, 2),
		Charges: []*ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*AccountCharge{
			"THIS_GENUUID1": {
				AccountID:    "TestEventChargesEquals",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(8, 1),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR1",
			},
			"GENUUID1": {
				AccountID:       "TestEventChargesEquals",
				BalanceID:       "ABSTRACT2",
				BalanceLimit:    NewDecimal(0, 0),
				RatingID:        "GENUUID_RATING1",
				JoinedChargeIDs: []string{"THIS_GENUUID1"},
			},
			"GENUUID2": {
				AccountID:    "TestEventChargesEquals",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(2, 0),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR2",
				RatingID:     "ID_FOR_RATING",
				AttributeIDs: []string{"ATTR1", "ATTR2"},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"GENUUID_FACTOR1": {
				Factor:    NewDecimal(100, 0),
				FilterIDs: []string{"*string:~*req.Account:1003"},
			},
			"GENUUID_FACTOR2": {
				Factor: NewDecimal(200, 0),
			},
		},
		Rating: map[string]*RateSInterval{
			"GENUUID_RATING1": {
				Increments: []*RateSIncrement{
					{
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
			"GENUUID_RATING2": {
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(4, 2),
						Usage:             NewDecimal(int64(30*time.Second), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(0, 0),
				CompressFactor: 2,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE1": {
				IntervalStart: NewDecimal(0, 0),
				FixedFee:      NewDecimal(4, 1),
				RecurrentFee:  NewDecimal(24, 1),
			},
		},
		Accounts: map[string]*Account{
			"ACC1": {
				Tenant:    CGRateSorg,
				ID:        "account_1",
				FilterIDs: []string{"*string:~*req.Account:1003"},
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
					},
				},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				Balances: map[string]*Balance{
					"bal1": {
						ID:        "BAL1",
						FilterIDs: []string{"*string:~*req.Account:1003"},
						Weights: []*DynamicWeight{
							{
								Weight: 25,
							},
						},
						Type:  MetaAbstract,
						Units: NewDecimal(int64(30*time.Second), 0),
						UnitFactors: []*UnitFactor{
							{
								Factor:    NewDecimal(100, 0),
								FilterIDs: []string{"*string:~*req.Account:1003"},
							},
							{
								Factor: NewDecimal(200, 0),
							},
						},
						CostIncrements: []*CostIncrement{
							{
								Increment:    NewDecimal(int64(time.Second), 0),
								RecurrentFee: NewDecimal(5, 0),
							},
							{
								FilterIDs:    []string{"*string:~*req.Account:1003"},
								Increment:    NewDecimal(int64(2*time.Second), 0),
								FixedFee:     NewDecimal(1, 0),
								RecurrentFee: NewDecimal(5, 0),
							},
						},
						AttributeIDs: []string{"ATTRIBUTE1"},
					},
					"bal2": {
						ID:        "BAL2",
						FilterIDs: []string{"*string:~*req.Account:1004"},
						Weights: []*DynamicWeight{
							{
								Weight: 25,
							},
						},
						Type:  MetaConcrete,
						Units: NewDecimal(2000, 0),
						UnitFactors: []*UnitFactor{
							{
								Factor: NewDecimal(200, 0),
							},
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"*string:~*req.Account:1004"},
								Increment:    NewDecimal(int64(2*time.Second), 0),
								FixedFee:     NewDecimal(1, 0),
								RecurrentFee: NewDecimal(5, 0),
							},
						},
						AttributeIDs:   []string{"ATTRIBUTE1"},
						RateProfileIDs: []string{"RATE1", "RATE2"},
					},
				},
				ThresholdIDs: []string{},
			},
			"ACC2": {
				Tenant: CGRateSorg,
				ID:     "account_2",
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
				},
				FilterIDs: []string{"*ai:~*req.AnswerTime:2020-10-10T10:00:00Z"},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				ThresholdIDs: []string{},
			},
		},
	}
	if ok := eEvChgs.Equals(expectedEqual); !ok {
		t.Errorf("Expected %+v, received %+v", ToJSON(eEvChgs), ToJSON(expectedEqual))
	}

	eEvChgs.Charges[0].CompressFactor = 2
	expectedEqual.Charges[0].CompressFactor = 3
	if ok := eEvChgs.Equals(expectedEqual); ok {
		t.Errorf("Expected %+v, received %+v", ToJSON(eEvChgs), ToJSON(expectedEqual))
	}

	eEvChgs.Charges[0].CompressFactor = 3

	eEvChgs.Accounts["ACC2"].ID = "id1"
	expectedEqual.Accounts["ACC2"].ID = "id2"

	if ok := eEvChgs.Equals(expectedEqual); ok {
		t.Errorf("Expected %+v, received %+v", ToJSON(eEvChgs), ToJSON(expectedEqual))
	}

	eEvChgs = nil
	if ok := eEvChgs.Equals(expectedEqual); ok {
		t.Errorf("Expected %+v, received %+v", ToJSON(eEvChgs), ToJSON(expectedEqual))
	}

	expectedEqual = nil
	if ok := eEvChgs.Equals(expectedEqual); !ok {
		t.Errorf("Expected %+v, received %+v", ToJSON(eEvChgs), ToJSON(expectedEqual))
	}
}
func TestEventChargerMerge(t *testing.T) {
	eEvChgs := &EventCharges{
		Abstracts: NewDecimal(47500, 3),
		Concretes: NewDecimal(515, 2),
		Charges: []*ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*AccountCharge{
			"THIS_GENUUID1": {
				AccountID:    "TestEventChargesMerge",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(8, 1),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR1",
			},
			"GENUUID1": {
				AccountID:       "TestEventChargesMerge",
				BalanceID:       "ABSTRACT2",
				BalanceLimit:    NewDecimal(0, 0),
				RatingID:        "GENUUID_RATING1",
				JoinedChargeIDs: []string{"THIS_GENUUID1"},
			},
			"GENUUID2": {
				AccountID:    "TestEventChargesMerge",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(2, 0),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR2",
				RatingID:     "ID_FOR_RATING",
				AttributeIDs: []string{"ATTR1", "ATTR2"},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"GENUUID_FACTOR1": {
				Factor:    NewDecimal(100, 0),
				FilterIDs: []string{"*string:~*req.Account:1003"},
			},
			"GENUUID_FACTOR2": {
				Factor: NewDecimal(200, 0),
			},
		},
		Rating: map[string]*RateSInterval{
			"GENUUID_RATING1": {
				Increments: []*RateSIncrement{
					{
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE_2",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
			"GENUUID_RATING2": {
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(4, 2),
						Usage:             NewDecimal(int64(30*time.Second), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE_1",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(0, 0),
				CompressFactor: 2,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE_1": {
				IntervalStart: NewDecimal(0, 0),
				FixedFee:      NewDecimal(4, 1),
				RecurrentFee:  NewDecimal(24, 1),
			},
			"RATE_2": {
				IntervalStart: NewDecimal(12, 1),
				FixedFee:      NewDecimal(1, 0),
				RecurrentFee:  NewDecimal(5, 2),
			},
		},
		Accounts: map[string]*Account{
			"ACC1": {
				Tenant:    CGRateSorg,
				ID:        "account_1",
				FilterIDs: []string{"*string:~*req.Account:1003"},
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
					},
				},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				Balances: map[string]*Balance{
					"bal1": {
						ID:        "BAL1",
						FilterIDs: []string{"*string:~*req.Account:1003"},
						Weights: []*DynamicWeight{
							{
								Weight: 25,
							},
						},
						Type:  MetaAbstract,
						Units: NewDecimal(int64(30*time.Second), 0),
						UnitFactors: []*UnitFactor{
							{
								Factor:    NewDecimal(100, 0),
								FilterIDs: []string{"*string:~*req.Account:1003"},
							},
							{
								Factor: NewDecimal(200, 0),
							},
						},
						CostIncrements: []*CostIncrement{
							{
								Increment:    NewDecimal(int64(time.Second), 0),
								RecurrentFee: NewDecimal(5, 0),
							},
							{
								FilterIDs:    []string{"*string:~*req.Account:1003"},
								Increment:    NewDecimal(int64(2*time.Second), 0),
								FixedFee:     NewDecimal(1, 0),
								RecurrentFee: NewDecimal(5, 0),
							},
						},
						AttributeIDs: []string{"ATTRIBUTE1"},
					},
					"bal2": {
						ID:        "BAL2",
						FilterIDs: []string{"*string:~*req.Account:1004"},
						Weights: []*DynamicWeight{
							{
								Weight: 25,
							},
						},
						Type:  MetaConcrete,
						Units: NewDecimal(2000, 0),
						UnitFactors: []*UnitFactor{
							{
								Factor: NewDecimal(200, 0),
							},
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"*string:~*req.Account:1004"},
								Increment:    NewDecimal(int64(2*time.Second), 0),
								FixedFee:     NewDecimal(1, 0),
								RecurrentFee: NewDecimal(5, 0),
							},
						},
						AttributeIDs:   []string{"ATTRIBUTE1"},
						RateProfileIDs: []string{"RATE1", "RATE2"},
					},
				},
				ThresholdIDs: []string{},
			},
			"ACC2": {
				Tenant: CGRateSorg,
				ID:     "account_2",
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
				},
				FilterIDs: []string{"*ai:~*req.AnswerTime:2020-10-10T10:00:00Z"},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				ThresholdIDs: []string{},
			},
		},
	}

	newEc := &EventCharges{
		Accounting: map[string]*AccountCharge{
			"GENUUID3": {
				AccountID:    "TestEventChargesMerge",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(2, 0),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR2",
				RatingID:     "ID_FOR_RATING",
				AttributeIDs: []string{"ATTR1", "ATTR2"},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"GENUUID_FACTOR3": {
				Factor: NewDecimal(200, 0),
			},
		},
		Rating: map[string]*RateSInterval{
			"GENUUID_RATING3": {
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(4, 2),
						Usage:             NewDecimal(int64(30*time.Second), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE_1",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(0, 0),
				CompressFactor: 2,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE_3": {
				IntervalStart: NewDecimal(0, 0),
				FixedFee:      NewDecimal(4, 1),
				RecurrentFee:  NewDecimal(24, 1),
			},
			"RATE_4": {
				IntervalStart: NewDecimal(12, 1),
				FixedFee:      NewDecimal(1, 0),
				RecurrentFee:  NewDecimal(5, 2),
			},
		},
		Accounts: map[string]*Account{
			"ACC3": {
				Tenant: CGRateSorg,
				ID:     "account_3",
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
				},
				FilterIDs: []string{"*ai:~*req.AnswerTime:2020-10-10T10:00:00Z"},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				ThresholdIDs: []string{},
			},
		},
	}

	expEc := &EventCharges{
		Abstracts: NewDecimal(47500, 3),
		Concretes: NewDecimal(515, 2),
		Charges: []*ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*AccountCharge{
			"THIS_GENUUID1": {
				AccountID:    "TestEventChargesMerge",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(8, 1),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR1",
			},
			"GENUUID1": {
				AccountID:       "TestEventChargesMerge",
				BalanceID:       "ABSTRACT2",
				BalanceLimit:    NewDecimal(0, 0),
				RatingID:        "GENUUID_RATING1",
				JoinedChargeIDs: []string{"THIS_GENUUID1"},
			},
			"GENUUID2": {
				AccountID:    "TestEventChargesMerge",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(2, 0),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR2",
				RatingID:     "ID_FOR_RATING",
				AttributeIDs: []string{"ATTR1", "ATTR2"},
			},
			"GENUUID3": {
				AccountID:    "TestEventChargesMerge",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(2, 0),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR2",
				RatingID:     "ID_FOR_RATING",
				AttributeIDs: []string{"ATTR1", "ATTR2"},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"GENUUID_FACTOR1": {
				Factor:    NewDecimal(100, 0),
				FilterIDs: []string{"*string:~*req.Account:1003"},
			},
			"GENUUID_FACTOR2": {
				Factor: NewDecimal(200, 0),
			},
			"GENUUID_FACTOR3": {
				Factor: NewDecimal(200, 0),
			},
		},
		Rating: map[string]*RateSInterval{
			"GENUUID_RATING1": {
				Increments: []*RateSIncrement{
					{
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE_2",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
			"GENUUID_RATING2": {
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(4, 2),
						Usage:             NewDecimal(int64(30*time.Second), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE_1",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(0, 0),
				CompressFactor: 2,
			},
			"GENUUID_RATING3": {
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(4, 2),
						Usage:             NewDecimal(int64(30*time.Second), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE_1",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(0, 0),
				CompressFactor: 2,
			},
		},
		Rates: map[string]*IntervalRate{
			"RATE_1": {
				IntervalStart: NewDecimal(0, 0),
				FixedFee:      NewDecimal(4, 1),
				RecurrentFee:  NewDecimal(24, 1),
			},
			"RATE_2": {
				IntervalStart: NewDecimal(12, 1),
				FixedFee:      NewDecimal(1, 0),
				RecurrentFee:  NewDecimal(5, 2),
			},

			"RATE_3": {
				IntervalStart: NewDecimal(0, 0),
				FixedFee:      NewDecimal(4, 1),
				RecurrentFee:  NewDecimal(24, 1),
			},
			"RATE_4": {
				IntervalStart: NewDecimal(12, 1),
				FixedFee:      NewDecimal(1, 0),
				RecurrentFee:  NewDecimal(5, 2),
			},
		},
		Accounts: map[string]*Account{
			"ACC1": {
				Tenant:    CGRateSorg,
				ID:        "account_1",
				FilterIDs: []string{"*string:~*req.Account:1003"},
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
					{
						FilterIDs: []string{"*string:~*req.Account:1002"},
					},
				},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				Balances: map[string]*Balance{
					"bal1": {
						ID:        "BAL1",
						FilterIDs: []string{"*string:~*req.Account:1003"},
						Weights: []*DynamicWeight{
							{
								Weight: 25,
							},
						},
						Type:  MetaAbstract,
						Units: NewDecimal(int64(30*time.Second), 0),
						UnitFactors: []*UnitFactor{
							{
								Factor:    NewDecimal(100, 0),
								FilterIDs: []string{"*string:~*req.Account:1003"},
							},
							{
								Factor: NewDecimal(200, 0),
							},
						},
						CostIncrements: []*CostIncrement{
							{
								Increment:    NewDecimal(int64(time.Second), 0),
								RecurrentFee: NewDecimal(5, 0),
							},
							{
								FilterIDs:    []string{"*string:~*req.Account:1003"},
								Increment:    NewDecimal(int64(2*time.Second), 0),
								FixedFee:     NewDecimal(1, 0),
								RecurrentFee: NewDecimal(5, 0),
							},
						},
						AttributeIDs: []string{"ATTRIBUTE1"},
					},
					"bal2": {
						ID:        "BAL2",
						FilterIDs: []string{"*string:~*req.Account:1004"},
						Weights: []*DynamicWeight{
							{
								Weight: 25,
							},
						},
						Type:  MetaConcrete,
						Units: NewDecimal(2000, 0),
						UnitFactors: []*UnitFactor{
							{
								Factor: NewDecimal(200, 0),
							},
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"*string:~*req.Account:1004"},
								Increment:    NewDecimal(int64(2*time.Second), 0),
								FixedFee:     NewDecimal(1, 0),
								RecurrentFee: NewDecimal(5, 0),
							},
						},
						AttributeIDs:   []string{"ATTRIBUTE1"},
						RateProfileIDs: []string{"RATE1", "RATE2"},
					},
				},
				ThresholdIDs: []string{},
			},
			"ACC2": {
				Tenant: CGRateSorg,
				ID:     "account_2",
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
				},
				FilterIDs: []string{"*ai:~*req.AnswerTime:2020-10-10T10:00:00Z"},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				ThresholdIDs: []string{},
			},
			"ACC3": {
				Tenant: CGRateSorg,
				ID:     "account_3",
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
				},
				FilterIDs: []string{"*ai:~*req.AnswerTime:2020-10-10T10:00:00Z"},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				ThresholdIDs: []string{},
			},
		},
	}
	eEvChgs.Merge(newEc)
	if !reflect.DeepEqual(expEc, eEvChgs) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(expEc), ToJSON(eEvChgs))
	}
}

func TestEventChargesAppendChargeEntry(t *testing.T) {
	eC := &EventCharges{
		Accounting: map[string]*AccountCharge{
			"GENUUID3": {
				AccountID:    "TestEventChargesMerge",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(2, 0),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR2",
				RatingID:     "ID_FOR_RATING",
				AttributeIDs: []string{"ATTR1", "ATTR2"},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"GENUUID_FACTOR3": {
				Factor: NewDecimal(200, 0),
			},
		},
		Rating: map[string]*RateSInterval{
			"GENUUID_RATING3": {
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(4, 2),
						Usage:             NewDecimal(int64(30*time.Second), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE_1",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(0, 0),
				CompressFactor: 2,
			},
		},
		Accounts: map[string]*Account{
			"ACC3": {
				Tenant: CGRateSorg,
				ID:     "account_3",
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
				},
				FilterIDs: []string{"*ai:~*req.AnswerTime:2020-10-10T10:00:00Z"},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				ThresholdIDs: []string{},
			},
		},
	}

	cIls := &ChargeEntry{
		ChargingID:     "chrgid1",
		CompressFactor: 1,
	}

	cIls2 := &ChargeEntry{
		ChargingID:     "chrgid2",
		CompressFactor: 2,
	}

	exp := []*ChargeEntry{
		{
			ChargingID:     "chrgid1",
			CompressFactor: 1,
		},
		{
			ChargingID:     "chrgid2",
			CompressFactor: 2,
		},
	}
	eC.appendChargeEntry(cIls, cIls2)
	if !reflect.DeepEqual(exp, eC.Charges) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp[0]), ToJSON(exp))
	}
}

func TestEventChargesAppendChargeEntryNonEmptyCharges(t *testing.T) {
	eC := &EventCharges{
		Charges: []*ChargeEntry{
			{
				ChargingID:     "chrgid3",
				CompressFactor: 3,
			},
			{
				ChargingID:     "chrgid4",
				CompressFactor: 4,
			},
		},
		Accounting: map[string]*AccountCharge{
			"GENUUID3": {
				AccountID:    "TestEventChargesMerge",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(2, 0),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR2",
				RatingID:     "ID_FOR_RATING",
				AttributeIDs: []string{"ATTR1", "ATTR2"},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"GENUUID_FACTOR3": {
				Factor: NewDecimal(200, 0),
			},
		},
		Rating: map[string]*RateSInterval{
			"GENUUID_RATING3": {
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(4, 2),
						Usage:             NewDecimal(int64(30*time.Second), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE_1",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(0, 0),
				CompressFactor: 2,
			},
		},
		Accounts: map[string]*Account{
			"ACC3": {
				Tenant: CGRateSorg,
				ID:     "account_3",
				Weights: []*DynamicWeight{
					{
						Weight: 25,
					},
				},
				FilterIDs: []string{"*ai:~*req.AnswerTime:2020-10-10T10:00:00Z"},
				Opts: map[string]interface{}{
					MetaSubsys: MetaSessionS,
				},
				ThresholdIDs: []string{},
			},
		},
	}

	cIls3 := &ChargeEntry{
		ChargingID:     "chrgid4",
		CompressFactor: 4,
	}

	eC.appendChargeEntry(cIls3)
	if eC.Charges[len(eC.Charges)-1].CompressFactor != 5 {
		fmt.Println(eC.Charges[len(eC.Charges)-1].CompressFactor)
		t.Errorf("Expected 5, received %v", eC.Charges[len(eC.Charges)-1].CompressFactor)
	}
}

func TestUnitFactorID(t *testing.T) {
	ec := &EventCharges{
		UnitFactors: map[string]*UnitFactor{
			"uf1": {
				FilterIDs: []string{"fltr1"},
				Factor:    NewDecimal(int64(2), 0),
			},
		},
	}

	uF := &UnitFactor{
		FilterIDs: []string{"fltr1"},
		Factor:    NewDecimal(int64(2), 0),
	}

	exp := "uf1"
	if rcv := ec.unitFactorID(uF); rcv != exp {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}

	uF.Factor = NewDecimal(int64(3), 0)
	if rcv := ec.unitFactorID(uF); rcv != "" {
		t.Errorf("Expected %v \n but received \n %v", "", rcv)
	}
}

func TestRatingID(t *testing.T) {
	ec := &EventCharges{
		Rating: map[string]*RateSInterval{
			"ri1": {
				IntervalStart:  NewDecimal(int64(2), 0),
				CompressFactor: int64(1),
			},
		},
	}

	rIl := &RateSInterval{
		IntervalStart:  NewDecimal(int64(2), 0),
		CompressFactor: int64(1),
	}

	nIrRef := map[string]*IntervalRate{
		"ir1": {
			IntervalStart: NewDecimal(int64(2), 0),
		},
	}

	exp := "ri1"
	if rcv := ec.ratingID(rIl, nIrRef); rcv != exp {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}

	rIl.IntervalStart = NewDecimal(int64(3), 0)
	exp = ""
	if rcv := ec.ratingID(rIl, nIrRef); rcv != "" {
		t.Errorf("Expected %v \n but received \n %v", "", rcv)
	}
}

func TestAccountChargeID(t *testing.T) {
	ec := &EventCharges{
		Accounting: map[string]*AccountCharge{
			"acc1": {
				AccountID:    "acc_id1",
				BalanceID:    "blncid1",
				Units:        NewDecimal(int64(2), 0),
				BalanceLimit: NewDecimal(int64(3), 0),
				UnitFactorID: "uf1",
				AttributeIDs: []string{"attr_id1"},
				RatingID:     "ri1",
			},
		},
	}

	ac := &AccountCharge{
		AccountID:    "acc_id1",
		BalanceID:    "blncid1",
		Units:        NewDecimal(int64(2), 0),
		BalanceLimit: NewDecimal(int64(3), 0),
		UnitFactorID: "uf1",
		AttributeIDs: []string{"attr_id1"},
		RatingID:     "ri1",
	}

	exp := "acc1"
	if rcv := ec.accountChargeID(ac); rcv != exp {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}

	ac.AccountID = "acc_id2"
	exp = ""
	if rcv := ec.accountChargeID(ac); rcv != "" {
		t.Errorf("Expected %v \n but received \n %v", "", rcv)
	}
}

func TestECChargeEntryCloneEmpty(t *testing.T) {
	ce := &ChargeEntry{}
	if rcv := ce.Clone(); !reflect.DeepEqual(rcv, ce) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(ce), ToJSON(rcv))
	}
}

func TestECChargeEntryClone(t *testing.T) {
	ce := &ChargeEntry{
		ChargingID:     "Charging1",
		CompressFactor: 1,
	}
	if rcv := ce.Clone(); !reflect.DeepEqual(rcv, ce) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(ce), ToJSON(rcv))
	}
}

func TestECAccountChargeCloneEmpty(t *testing.T) {
	ac := &AccountCharge{}
	if rcv := ac.Clone(); !reflect.DeepEqual(rcv, ac) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(ac), ToJSON(rcv))
	}
}

func TestECAccountChargeClone(t *testing.T) {
	ac := &AccountCharge{
		AccountID:       "Acc1",
		BalanceID:       "Blnc1",
		Units:           NewDecimalFromFloat64(0.1),
		BalanceLimit:    NewDecimalFromFloat64(1.2),
		UnitFactorID:    "UF1",
		AttributeIDs:    []string{"ATTR1", "ATTR2"},
		RatingID:        "Rating1",
		JoinedChargeIDs: []string{"JC1", "JC2"},
	}
	if rcv := ac.Clone(); !reflect.DeepEqual(rcv, ac) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(ac), ToJSON(rcv))
	}
}

func TestECEventChargesCloneEmpty(t *testing.T) {
	ec := &EventCharges{}
	if rcv := ec.Clone(); !reflect.DeepEqual(rcv, ec) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(ec), ToJSON(rcv))
	}
}

func TestECEventChargesClone(t *testing.T) {
	ec := &EventCharges{
		Abstracts: NewDecimalFromFloat64(0.1),
		Concretes: NewDecimalFromFloat64(2.3),
		Charges: []*ChargeEntry{
			{
				ChargingID:     "Charging1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "Charging2",
				CompressFactor: 2,
			},
		},
		Accounting: map[string]*AccountCharge{
			"Acc1": {
				AccountID:       "Acc1",
				BalanceID:       "Blnc1",
				Units:           NewDecimalFromFloat64(0.1),
				BalanceLimit:    NewDecimalFromFloat64(1.2),
				UnitFactorID:    "UF1",
				AttributeIDs:    []string{"ATTR1", "ATTR2"},
				RatingID:        "Rating1",
				JoinedChargeIDs: []string{"JC1", "JC2"},
			},
			"Acc2": {
				AccountID:       "Acc2",
				BalanceID:       "Blnc2",
				Units:           NewDecimalFromFloat64(0.2),
				BalanceLimit:    NewDecimalFromFloat64(2.3),
				UnitFactorID:    "UF2",
				AttributeIDs:    []string{"ATTR3", "ATTR4"},
				RatingID:        "Rating2",
				JoinedChargeIDs: []string{"JC3", "JC4"},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"UF1": {
				FilterIDs: []string{"FLTR1", "FLTR2"},
				Factor:    NewDecimalFromFloat64(1.234),
			},
			"UF2": {
				FilterIDs: []string{"FLTR3", "FLTR4"},
				Factor:    NewDecimalFromFloat64(432.1),
			},
		},
		Rating: map[string]*RateSInterval{
			"RateSInterval1": {
				IntervalStart: NewDecimalFromFloat64(1.234),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimalFromFloat64(1.234),
						RateIntervalIndex: 1,
						RateID:            "Rate1",
						CompressFactor:    1,
						Usage:             NewDecimalFromFloat64(-321),
						cost:              decimal.New(4321, 5),
					},
					{
						IncrementStart:    NewDecimalFromFloat64(4.321),
						RateIntervalIndex: 1,
						RateID:            "Rate2",
						CompressFactor:    1,
						Usage:             NewDecimalFromFloat64(-123),
						cost:              decimal.New(123, 1),
					},
				},
				CompressFactor: 1,
				cost:           decimal.New(4321, 5),
			},
			"RateSInterval2": {
				IntervalStart: NewDecimalFromFloat64(2.34),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimalFromFloat64(1.234),
						RateIntervalIndex: 2,
						RateID:            "Rate1",
						CompressFactor:    2,
						Usage:             NewDecimalFromFloat64(-321),
						cost:              decimal.New(3456, 5),
					},
					{
						IncrementStart:    NewDecimalFromFloat64(4.321),
						RateIntervalIndex: 2,
						RateID:            "Rate2",
						CompressFactor:    2,
						Usage:             NewDecimalFromFloat64(-123),
						cost:              decimal.New(321, 1),
					},
				},
				CompressFactor: 1,
				cost:           decimal.New(345, 2),
			},
		},
		Rates: map[string]*IntervalRate{
			"IvalRate1": {
				IntervalStart: NewDecimalFromFloat64(1.2),
				FixedFee:      NewDecimalFromFloat64(1.234),
				RecurrentFee:  NewDecimalFromFloat64(0.5),
				Unit:          NewDecimalFromFloat64(7.1),
				Increment:     NewDecimalFromFloat64(-321),
			},
			"IvalRate2": {
				IntervalStart: NewDecimalFromFloat64(123.1),
				FixedFee:      NewDecimalFromFloat64(12.34),
				RecurrentFee:  NewDecimalFromFloat64(0.05),
				Unit:          NewDecimalFromFloat64(5.1),
				Increment:     NewDecimalFromFloat64(-357),
			},
		},
		Accounts: map[string]*Account{
			"Account1": {
				Tenant:    "cgrates.org",
				ID:        "Account1",
				FilterIDs: []string{"FLTR1", "FLTR2"},
				Weights: DynamicWeights{
					{
						Weight: 10,
					},
					{
						FilterIDs: []string{"FLTR3"},
						Weight:    20,
					},
				},
				Opts: map[string]interface{}{
					"optName": "optValue",
				},
				Balances: map[string]*Balance{
					"Blnc1": {
						ID:        "Blnc1",
						FilterIDs: []string{"FLTR1"},
						Weights: DynamicWeights{
							{
								Weight: 10,
							},
							{
								FilterIDs: []string{"FLTR3"},
								Weight:    20,
							},
						},
						Type:  MetaMonetary,
						Units: NewDecimalFromFloat64(0.1),
						UnitFactors: []*UnitFactor{
							{
								FilterIDs: []string{"FLTR1", "FLTR2"},
								Factor:    NewDecimalFromFloat64(1.234),
							},
							{
								FilterIDs: []string{"FLTR3", "FLTR4"},
								Factor:    NewDecimalFromFloat64(123.4),
							},
						},
						Opts: map[string]interface{}{
							"optName": "optValue",
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"FLTR3"},
								Increment:    NewDecimalFromFloat64(1.2),
								FixedFee:     NewDecimalFromFloat64(23.4),
								RecurrentFee: NewDecimalFromFloat64(3.21),
							},
						},
						AttributeIDs:   []string{"ATTR1", "ATTR2"},
						RateProfileIDs: []string{"RatePrf1", "RatePrf2"},
					},
					"Blnc2": {
						ID:        "Blnc2",
						FilterIDs: []string{"FLTR2"},
						Weights: DynamicWeights{
							{
								Weight: 10,
							},
							{
								FilterIDs: []string{"FLTR4"},
								Weight:    20,
							},
						},
						Type:  MetaVoice,
						Units: NewDecimalFromFloat64(0.1),
						UnitFactors: []*UnitFactor{
							{
								FilterIDs: []string{"FLTR3", "FLTR4"},
								Factor:    NewDecimalFromFloat64(1.234),
							},
							{
								FilterIDs: []string{"FLTR2"},
								Factor:    NewDecimalFromFloat64(123.4),
							},
						},
						Opts: map[string]interface{}{
							"optName": "optValue",
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"FLTR3"},
								Increment:    NewDecimalFromFloat64(1.2),
								FixedFee:     NewDecimalFromFloat64(23.4),
								RecurrentFee: NewDecimalFromFloat64(3.21),
							},
						},
						AttributeIDs:   []string{"ATTR1", "ATTR2"},
						RateProfileIDs: []string{"RatePrf1", "RatePrf2"},
					},
				},
			},
			"Account2": {
				Tenant:    "cgrates.org",
				ID:        "Account2",
				FilterIDs: []string{"FLTR3", "FLTR4"},
				Weights: DynamicWeights{
					{
						Weight: 15,
					},
					{
						FilterIDs: []string{"FLTR5"},
						Weight:    25,
					},
				},
				Opts: map[string]interface{}{
					"optName": "optValue",
				},
				Balances: map[string]*Balance{
					"Blnc1": {
						ID:        "Blnc1",
						FilterIDs: []string{"FLTR1"},
						Weights: DynamicWeights{
							{
								Weight: 10,
							},
							{
								FilterIDs: []string{"FLTR3"},
								Weight:    20,
							},
						},
						Type:  MetaMonetary,
						Units: NewDecimalFromFloat64(0.1),
						UnitFactors: []*UnitFactor{
							{
								FilterIDs: []string{"FLTR1", "FLTR2"},
								Factor:    NewDecimalFromFloat64(1.234),
							},
							{
								FilterIDs: []string{"FLTR3", "FLTR4"},
								Factor:    NewDecimalFromFloat64(123.4),
							},
						},
						Opts: map[string]interface{}{
							"optName": "optValue",
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"FLTR3"},
								Increment:    NewDecimalFromFloat64(1.2),
								FixedFee:     NewDecimalFromFloat64(23.4),
								RecurrentFee: NewDecimalFromFloat64(3.21),
							},
						},
						AttributeIDs:   []string{"ATTR1", "ATTR2"},
						RateProfileIDs: []string{"RatePrf1", "RatePrf2"},
					},
					"Blnc2": {
						ID:        "Blnc2",
						FilterIDs: []string{"FLTR2"},
						Weights: DynamicWeights{
							{
								Weight: 10,
							},
							{
								FilterIDs: []string{"FLTR4"},
								Weight:    20,
							},
						},
						Type:  MetaVoice,
						Units: NewDecimalFromFloat64(0.1),
						UnitFactors: []*UnitFactor{
							{
								FilterIDs: []string{"FLTR3", "FLTR4"},
								Factor:    NewDecimalFromFloat64(1.234),
							},
							{
								FilterIDs: []string{"FLTR2"},
								Factor:    NewDecimalFromFloat64(123.4),
							},
						},
						Opts: map[string]interface{}{
							"optName": "optValue",
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"FLTR3"},
								Increment:    NewDecimalFromFloat64(1.2),
								FixedFee:     NewDecimalFromFloat64(23.4),
								RecurrentFee: NewDecimalFromFloat64(3.21),
							},
						},
						AttributeIDs:   []string{"ATTR1", "ATTR2"},
						RateProfileIDs: []string{"RatePrf1", "RatePrf2"},
					},
				},
			},
		},
	}
	if rcv := ec.Clone(); !reflect.DeepEqual(rcv, ec) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(ec), ToJSON(rcv))
	}
}

func TestEqualsAccountCharges(t *testing.T) {
	var ac *AccountCharge
	var nAc *AccountCharge
	if rcv := ac.equals(nAc); rcv != true {
		t.Errorf("Expected <true>, Recevied <%v>", rcv)
	}

	ac = &AccountCharge{
		AttributeIDs: []string{"test", "range"},
	}
	nAc = &AccountCharge{
		AttributeIDs: []string{"test2", "range"},
	}

	if rcv := ac.equals(nAc); rcv != false {
		t.Error(rcv)
	}
}

func TestSyncIDsEventCharges(t *testing.T) {
	eEvChgs := &EventCharges{
		Charges: []*ChargeEntry{
			{
				ChargingID: "GENUUID3",
			},
		},
		Accounting: map[string]*AccountCharge{
			"THIS_GENUUID1": {
				AccountID: "TestEventChargesEquals",
			},
			"GENUUID3": {
				AccountID:       "TestEventChargesMerge",
				BalanceID:       "CONCRETE1",
				Units:           NewDecimal(8, 1),
				BalanceLimit:    NewDecimal(200, 0),
				UnitFactorID:    "GENUUID_FACTOR3",
				RatingID:        "GENUUID_RATING1",
				JoinedChargeIDs: []string{"THIS_GENUUID1"},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"GENUUID_FACTOR3": {
				Factor:    NewDecimal(100, 0),
				FilterIDs: []string{"*string:~*req.Account:1003"},
			},
		},
		Rating: map[string]*RateSInterval{
			"GENUUID_RATING1": {
				Increments: []*RateSIncrement{
					{
						Usage:          NewDecimal(int64(time.Minute), 0),
						CompressFactor: 1,
					},
				},
				IntervalStart:  NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
		},
	}

	newEc := &EventCharges{
		Charges: []*ChargeEntry{
			{
				ChargingID:     "GENUUID2",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*AccountCharge{
			"THIS_GENUUID2": {
				AccountID:    "TestEventChargesEquals",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(8, 1),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR1",
			},
			"GENUUID2": {
				AccountID:       "TestEventChargesMerge",
				BalanceID:       "CONCRETE1",
				Units:           NewDecimal(8, 1),
				BalanceLimit:    NewDecimal(200, 0),
				UnitFactorID:    "GENUUID_FACTOR2",
				RatingID:        "GENUUID_RATING2",
				JoinedChargeIDs: []string{"THIS_GENUUID2"},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"GENUUID_FACTOR2": {
				Factor:    NewDecimal(100, 0),
				FilterIDs: []string{"*string:~*req.Account:1003"},
			},
		},
		Rating: map[string]*RateSInterval{
			"GENUUID_RATING2": {
				Increments: []*RateSIncrement{
					{
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateIntervalIndex: 0,
						RateID:            "RATE_2",
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
		},
	}

	expEc := &EventCharges{
		Charges: []*ChargeEntry{
			{
				ChargingID: "GENUUID3",
			},
		},
		Accounting: map[string]*AccountCharge{
			"THIS_GENUUID1": {
				AccountID: "TestEventChargesEquals",
			},
			"GENUUID3": {
				AccountID:       "TestEventChargesMerge",
				BalanceID:       "CONCRETE1",
				Units:           NewDecimal(8, 1),
				BalanceLimit:    NewDecimal(200, 0),
				UnitFactorID:    "GENUUID_FACTOR3",
				RatingID:        "GENUUID_RATING1",
				JoinedChargeIDs: []string{"THIS_GENUUID1"},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"GENUUID_FACTOR3": {
				Factor:    NewDecimal(100, 0),
				FilterIDs: []string{"*string:~*req.Account:1003"},
			},
		},
		Rating: map[string]*RateSInterval{
			"GENUUID_RATING1": {
				Increments: []*RateSIncrement{
					{
						Usage:          NewDecimal(int64(time.Minute), 0),
						CompressFactor: 1,
					},
				},
				IntervalStart:  NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
		},
	}
	eEvChgs.SyncIDs(newEc)
	if !reflect.DeepEqual(expEc, eEvChgs) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(expEc), ToJSON(eEvChgs))
	}
}
