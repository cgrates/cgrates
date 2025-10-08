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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
				Opts: map[string]any{
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
						Opts: map[string]any{
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
						Opts: map[string]any{
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
				Opts: map[string]any{
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
						Opts: map[string]any{
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
						Opts: map[string]any{
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

func TestEqualsAccounting(t *testing.T) {

	acc1 := &AccountCharge{
		AccountID:       "AccountID1",
		BalanceID:       "BalanceID1",
		Units:           NewDecimal(20, 0),
		BalanceLimit:    NewDecimal(40, 0),
		UnitFactorID:    "UF1",
		AttributeIDs:    []string{"ID1", "ID2"},
		RatingID:        "RatingID1",
		JoinedChargeIDs: []string{"chID1", "chID3"},
	}
	acc2 := &AccountCharge{
		AccountID:       "AccountID2",
		BalanceID:       "BalanceID2",
		Units:           NewDecimal(20, 0),
		BalanceLimit:    NewDecimal(40, 0),
		UnitFactorID:    "UF2",
		AttributeIDs:    []string{"ID3", "ID4"},
		RatingID:        "RatingID2",
		JoinedChargeIDs: []string{"chID2"},
	}

	accM1 := map[string]*AccountCharge{
		"chID1": {
			AccountID:    "AccountID2",
			BalanceID:    "BalanceID2",
			Units:        NewDecimal(20, 0),
			BalanceLimit: NewDecimal(40, 0),
			UnitFactorID: "UF2",
			AttributeIDs: []string{"ID3", "ID4"},
			RatingID:     "RatingID2",
		},
		"chID3": {
			AccountID:    "AccountID3",
			BalanceID:    "BalanceID3",
			Units:        NewDecimal(20, 0),
			BalanceLimit: NewDecimal(40, 0),
			UnitFactorID: "UF2",
			AttributeIDs: []string{"ID3", "ID4"},
			RatingID:     "RatingID3",
		},
		"GENUUID1": {
			AccountID:       "AccountID1",
			BalanceID:       "BalanceID1",
			Units:           NewDecimal(20, 0),
			BalanceLimit:    NewDecimal(40, 0),
			UnitFactorID:    "UF1",
			AttributeIDs:    []string{"ID1", "ID2"},
			RatingID:        "RatingID1",
			JoinedChargeIDs: []string{"chID1", "chID3"},
		},
	}
	accM2 := map[string]*AccountCharge{
		"chID2": {
			AccountID:    "AccountID2",
			BalanceID:    "BalanceID2",
			Units:        NewDecimal(20, 0),
			BalanceLimit: NewDecimal(40, 0),
			UnitFactorID: "UF2",
			AttributeIDs: []string{"ID1", "ID2"},
			RatingID:     "RatingID2",
		},
		"GENUUID2": {
			AccountID:       "AccountID2",
			BalanceID:       "BalanceID2",
			Units:           NewDecimal(20, 0),
			BalanceLimit:    NewDecimal(40, 0),
			UnitFactorID:    "UF2",
			AttributeIDs:    []string{"ID3", "ID4"},
			RatingID:        "RatingID2",
			JoinedChargeIDs: []string{"chID2"},
		}}

	uf1 := map[string]*UnitFactor{
		"UF2": {
			Factor: NewDecimal(200, 0),
		},
	}
	uf2 := map[string]*UnitFactor{
		"UF2": {
			Factor: NewDecimal(200, 0),
		},
	}
	rat1 := map[string]*RateSInterval{
		"RatingID1": {
			Increments: []*RateSIncrement{
				{
					Usage:          NewDecimal(int64(time.Minute), 0),
					CompressFactor: 1,
					RateID:         "IvalRate1",
				},
			},
			IntervalStart:  NewDecimal(int64(time.Second), 0),
			CompressFactor: 1,
		}}
	rat2 := map[string]*RateSInterval{
		"RatingID1": {
			Increments: []*RateSIncrement{
				{
					Usage:          NewDecimal(int64(time.Minute), 0),
					CompressFactor: 1,
					RateID:         "IvalRate1",
				},
			},
			IntervalStart:  NewDecimal(int64(time.Second), 0),
			CompressFactor: 1,
		}}
	rts1 := map[string]*IntervalRate{"IvalRate1": {
		IntervalStart: NewDecimalFromFloat64(1.2),
		FixedFee:      NewDecimalFromFloat64(1.234),
		RecurrentFee:  NewDecimalFromFloat64(0.5),
		Unit:          NewDecimalFromFloat64(7.1),
		Increment:     NewDecimalFromFloat64(-321),
	}}
	rts2 := map[string]*IntervalRate{"IvalRate1": {
		IntervalStart: NewDecimalFromFloat64(1.2),
		FixedFee:      NewDecimalFromFloat64(1.234),
		RecurrentFee:  NewDecimalFromFloat64(0.5),
		Unit:          NewDecimalFromFloat64(7.1),
		Increment:     NewDecimalFromFloat64(-321),
	}}
	//////////////////////////
	acc10 := &AccountCharge{
		AccountID:       "AccountID2",
		BalanceID:       "BalanceID2",
		Units:           NewDecimal(20, 0),
		BalanceLimit:    NewDecimal(40, 0),
		UnitFactorID:    "UF2",
		AttributeIDs:    []string{"ID3", "ID4"},
		RatingID:        "RatingID2",
		JoinedChargeIDs: []string{"chID2"},
	}

	acc20 := &AccountCharge{
		AccountID:       "AccountID2",
		BalanceID:       "BalanceID2",
		Units:           NewDecimal(20, 0),
		BalanceLimit:    NewDecimal(40, 0),
		UnitFactorID:    "UF2",
		AttributeIDs:    []string{"ID3", "ID4"},
		RatingID:        "RatingID2",
		JoinedChargeIDs: []string{"chID3"},
	}

	accM10 := map[string]*AccountCharge{

		"chID2": {
			AccountID:    "AccountID2",
			BalanceID:    "BalanceID2",
			Units:        NewDecimal(20, 0),
			BalanceLimit: NewDecimal(40, 0),
			UnitFactorID: "UF2",
			AttributeIDs: []string{"ID3", "ID4"},
			RatingID:     "RatingID2",
		},
		"GENUUID1": {
			AccountID:       "AccountID1",
			BalanceID:       "BalanceID1",
			Units:           NewDecimal(20, 0),
			BalanceLimit:    NewDecimal(40, 0),
			UnitFactorID:    "UF1",
			AttributeIDs:    []string{"ID1", "ID2"},
			RatingID:        "RatingID1",
			JoinedChargeIDs: []string{"chID2"},
		},
	}

	accM20 := map[string]*AccountCharge{
		"chID3": {
			AccountID:    "AccountID2",
			BalanceID:    "BalanceID5",
			Units:        NewDecimal(20, 0),
			BalanceLimit: NewDecimal(40, 0),
			UnitFactorID: "UF2",
			AttributeIDs: []string{"ID3", "ID4"},
			RatingID:     "RatingID2",
		},

		"GENUUID1": {
			AccountID:       "AccountID1",
			BalanceID:       "BalanceID1",
			Units:           NewDecimal(20, 0),
			BalanceLimit:    NewDecimal(40, 0),
			UnitFactorID:    "UF1",
			AttributeIDs:    []string{"ID1", "ID2"},
			RatingID:        "RatingID1",
			JoinedChargeIDs: []string{"chID3"},
		},
	}

	equalsAccounting(acc1, acc2, accM1, accM2, uf1, uf2, rat1, rat2, rts1, rts2)
	equalsAccounting(acc10, acc20, accM10, accM20, uf1, uf2, rat1, rat2, rts1, rts2)
}

func TestEventChargesFieldAsInterface(t *testing.T) {

	// ToDo: Replace the randomly assigned values with ones resulted from a real charge. For
	// the moment this is enough to test the field retrieval functionality with FieldAsInterface.
	ec := &EventCharges{
		Concretes: &Decimal{decimal.New(152, 1)},
		Abstracts: &Decimal{decimal.New(145, 1)},
		Charges: []*ChargeEntry{
			{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "*rating:rating1",
				CompressFactor: 2,
			},
		},
		Accounting: map[string]*AccountCharge{
			"accounting1": {
				AccountID:       "acc1",
				BalanceID:       "balance1",
				Units:           &Decimal{decimal.New(10, 0)},
				BalanceLimit:    &Decimal{decimal.New(0, 0)},
				UnitFactorID:    "unit_factor1",
				AttributeIDs:    []string{"attr1", "attr2"},
				RatingID:        "rating2",
				JoinedChargeIDs: []string{"joined_charge"},
			},
			"joined_charge": {
				AccountID:       "acc2",
				BalanceID:       "balance2",
				Units:           &Decimal{decimal.New(10, 0)},
				BalanceLimit:    &Decimal{decimal.New(0, 0)},
				UnitFactorID:    "unit_factor2",
				AttributeIDs:    []string{"attr3", "attr4"},
				RatingID:        "rating3",
				JoinedChargeIDs: []string{},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"unit_factor1": {
				FilterIDs: []string{"fltr1", "fltr2"},
				Factor:    &Decimal{decimal.New(2, 0)},
			},
			"unit_factor2": {
				FilterIDs: []string{"fltr3", "fltr4"},
				Factor:    &Decimal{decimal.New(3, 0)},
			},
		},
		Rating: map[string]*RateSInterval{
			"rating1": {
				IntervalStart: &Decimal{decimal.New(4, 0)},
				Increments: []*RateSIncrement{
					{
						IncrementStart:    &Decimal{decimal.New(5, 0)},
						RateIntervalIndex: 1,
						RateID:            "rate1",
						CompressFactor:    1,
						Usage:             &Decimal{decimal.New(6, 0)},
					},
					{
						IncrementStart:    &Decimal{decimal.New(7, 0)},
						RateIntervalIndex: 2,
						RateID:            "rate2",
						CompressFactor:    1,
						Usage:             &Decimal{decimal.New(8, 0)},
					},
				},
				CompressFactor: 3,
			},
			"rating2": {
				IntervalStart: &Decimal{decimal.New(5, 0)},
				Increments: []*RateSIncrement{
					{
						IncrementStart:    &Decimal{decimal.New(9, 0)},
						RateIntervalIndex: 3,
						RateID:            "rate1",
						CompressFactor:    1,
						Usage:             &Decimal{decimal.New(10, 0)},
					},
				},
				CompressFactor: 3,
			},
			"rating3": {
				IntervalStart: &Decimal{},
				Increments: []*RateSIncrement{
					{
						IncrementStart:    &Decimal{decimal.New(11, 0)},
						RateIntervalIndex: 4,
						RateID:            "rate2",
						CompressFactor:    5,
						Usage:             &Decimal{decimal.New(12, 0)},
					},
				},
				CompressFactor: 3,
			},
		},
		Rates: map[string]*IntervalRate{
			"rate1": {
				IntervalStart: &Decimal{decimal.New(1, 0)},
				FixedFee:      &Decimal{decimal.New(2, 0)},
				RecurrentFee:  &Decimal{decimal.New(3, 0)},
				Unit:          &Decimal{decimal.New(4, 0)},
				Increment:     &Decimal{decimal.New(5, 0)},
			},
			"rate2": {
				IntervalStart: &Decimal{decimal.New(6, 0)},
				FixedFee:      &Decimal{decimal.New(7, 0)},
				RecurrentFee:  &Decimal{decimal.New(8, 0)},
				Unit:          &Decimal{decimal.New(9, 0)},
				Increment:     &Decimal{decimal.New(10, 0)},
			},
		},
		Accounts: map[string]*Account{
			"acc1": {
				Tenant:    "cgrates.org",
				FilterIDs: []string{"fltr1"},
				ID:        "acc1",
				Weights: DynamicWeights{
					{
						FilterIDs: []string{"fltr2"},
						Weight:    10,
					},
				},
				Blockers: DynamicBlockers{
					{
						FilterIDs: []string{"fltr3"},
						Blocker:   true,
					},
				},
				Opts: map[string]any{
					"opt1": "value1",
				},
				Balances: map[string]*Balance{
					"balance1": {
						ID:        "balance1",
						FilterIDs: []string{"fltr4"},
						Weights: DynamicWeights{
							{
								FilterIDs: []string{"fltr3"},
								Weight:    20,
							},
						},
						Blockers: DynamicBlockers{
							{
								FilterIDs: []string{"fltr3"},
								Blocker:   true,
							},
						},
						Type:  MetaMonetary,
						Units: &Decimal{decimal.New(1, 0)},
						UnitFactors: []*UnitFactor{
							{
								FilterIDs: []string{"fltr1", "fltr2"},
								Factor:    &Decimal{decimal.New(2, 0)},
							},
						},
						Opts: map[string]any{
							"opt1": "value1",
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"fltr3"},
								Increment:    &Decimal{decimal.New(3, 0)},
								FixedFee:     &Decimal{decimal.New(1, 0)},
								RecurrentFee: &Decimal{decimal.New(2, 0)},
							},
						},
						AttributeIDs:   []string{"attr1"},
						RateProfileIDs: []string{"rate_prf1"},
					},
					"balance2": {
						ID:        "balance2",
						FilterIDs: []string{"fltr3"},
						Weights: DynamicWeights{
							{
								FilterIDs: []string{"fltr5"},
								Weight:    20,
							},
						},
						Blockers: DynamicBlockers{
							{
								FilterIDs: []string{"fltr5"},
								Blocker:   true,
							},
						},
						Type:  MetaVoice,
						Units: &Decimal{decimal.New(5, 0)},
						UnitFactors: []*UnitFactor{
							{
								FilterIDs: []string{"fltr3", "fltr4"},
								Factor:    &Decimal{decimal.New(1, 0)},
							},
						},
						Opts: map[string]any{
							"opts1": "value1",
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"fltr2"},
								Increment:    &Decimal{decimal.New(1, 0)},
								FixedFee:     &Decimal{decimal.New(2, 0)},
								RecurrentFee: &Decimal{decimal.New(3, 0)},
							},
						},
						AttributeIDs:   []string{"attr2"},
						RateProfileIDs: []string{"rate_prf2"},
					},
				},
			},
			"acc2": {
				Tenant:    "cgrates.org",
				FilterIDs: []string{"fltr2"},
				ID:        "acc2",
				Weights: DynamicWeights{
					{
						FilterIDs: []string{"fltr2"},
						Weight:    10,
					},
				},
				Blockers: DynamicBlockers{
					{
						FilterIDs: []string{"fltr3"},
						Blocker:   true,
					},
				},
				Opts: map[string]any{
					"opt2": "value2",
				},
				Balances: map[string]*Balance{
					"balance1": {
						ID:        "balance1",
						FilterIDs: []string{"fltr4"},
						Weights: DynamicWeights{
							{
								FilterIDs: []string{"fltr1"},
								Weight:    20,
							},
						},
						Blockers: DynamicBlockers{
							{
								FilterIDs: []string{"fltr3"},
								Blocker:   false,
							},
						},
						Type:  MetaMonetary,
						Units: &Decimal{decimal.New(11, 0)},
						UnitFactors: []*UnitFactor{
							{
								FilterIDs: []string{"fltr3", "fltr4"},
								Factor:    &Decimal{},
							},
						},
						Opts: map[string]any{
							"opts2": "value2",
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"fltr3"},
								Increment:    &Decimal{decimal.New(12, 0)},
								FixedFee:     &Decimal{decimal.New(13, 0)},
								RecurrentFee: &Decimal{decimal.New(14, 0)},
							},
						},
						AttributeIDs:   []string{"attr3"},
						RateProfileIDs: []string{"rate_prf3"},
					},
				},
			},
		},
	}

	testcases := []struct {
		name   string
		fields []string
		exp    any
	}{
		{
			name:   "Concretes",
			fields: []string{"Concretes"},
			exp:    "15.2",
		},
		{
			name:   "Abstracts",
			fields: []string{"Abstracts"},
			exp:    "14.5",
		},
		{
			name:   "Charges",
			fields: []string{"Charges"},
			exp:    `[{"ChargingID":"*accounting:accounting1","CompressFactor":1},{"ChargingID":"*rating:rating1","CompressFactor":2}]`,
		},
		{
			name:   "Charges[1]",
			fields: []string{"Charges[1]"},
			exp:    `{"ChargingID":"*rating:rating1","CompressFactor":2}`,
		},
		{
			name:   "Charges[1].ChargingID",
			fields: []string{"Charges[1]", "ChargingID"},
			exp:    "*rating:rating1",
		},
		{
			name:   "Charges[1].CompressFactor",
			fields: []string{"Charges[1]", "CompressFactor"},
			exp:    "2",
		},
		{
			name:   "Charges[0].Charging",
			fields: []string{"Charges[0]", "Charging"},
			exp:    `{"AccountID":"acc1","BalanceID":"balance1","Units":10,"BalanceLimit":0,"UnitFactorID":"unit_factor1","AttributeIDs":["attr1","attr2"],"RatingID":"rating2","JoinedChargeIDs":["joined_charge"]}`,
		},
		{
			name:   "Charges[0].Charging.AccountID",
			fields: []string{"Charges[0]", "Charging", "AccountID"},
			exp:    "acc1",
		},
		{
			name:   "Charges[0].Charging.BalanceID",
			fields: []string{"Charges[0]", "Charging", "BalanceID"},
			exp:    "balance1",
		},
		{
			name:   "Charges[0].Charging.Units",
			fields: []string{"Charges[0]", "Charging", "Units"},
			exp:    "10",
		},
		{
			name:   "Charges[0].Charging.BalanceLimit",
			fields: []string{"Charges[0]", "Charging", "BalanceLimit"},
			exp:    "0",
		},
		{
			name:   "Charges[0].Charging.UnitFactorID",
			fields: []string{"Charges[0]", "Charging", "UnitFactorID"},
			exp:    "unit_factor1",
		},
		{
			name:   "Charges[0].Charging.AttributeIDs",
			fields: []string{"Charges[0]", "Charging", "AttributeIDs"},
			exp:    `["attr1","attr2"]`,
		},
		{
			name:   "Charges[0].Charging.AttributeIDs[1]",
			fields: []string{"Charges[0]", "Charging", "AttributeIDs[1]"},
			exp:    "attr2",
		},
		{
			name:   "Charges[0].Charging.RatingID",
			fields: []string{"Charges[0]", "Charging", "RatingID"},
			exp:    "rating2",
		},
		{
			name:   "Charges[0].Charging.JoinedChargeIDs",
			fields: []string{"Charges[0]", "Charging", "JoinedChargeIDs"},
			exp:    `["joined_charge"]`,
		},
		{
			name:   "Charges[0].Charging.JoinedChargeIDs[0]",
			fields: []string{"Charges[0]", "Charging", "JoinedChargeIDs[0]"},
			exp:    "joined_charge",
		},
		{
			name:   "Charges[0].Charging.Account",
			fields: []string{"Charges[0]", "Charging", "Account"},
			exp:    `{"Tenant":"cgrates.org","ID":"acc1","FilterIDs":["fltr1"],"Weights":[{"FilterIDs":["fltr2"],"Weight":10}],"Blockers":[{"FilterIDs":["fltr3"],"Blocker":true}],"Opts":{"opt1":"value1"},"Balances":{"balance1":{"ID":"balance1","FilterIDs":["fltr4"],"Weights":[{"FilterIDs":["fltr3"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr3"],"Blocker":true}],"Type":"*monetary","Units":1,"UnitFactors":[{"FilterIDs":["fltr1","fltr2"],"Factor":2}],"Opts":{"opt1":"value1"},"CostIncrements":[{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}],"AttributeIDs":["attr1"],"RateProfileIDs":["rate_prf1"]},"balance2":{"ID":"balance2","FilterIDs":["fltr3"],"Weights":[{"FilterIDs":["fltr5"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr5"],"Blocker":true}],"Type":"*voice","Units":5,"UnitFactors":[{"FilterIDs":["fltr3","fltr4"],"Factor":1}],"Opts":{"opts1":"value1"},"CostIncrements":[{"FilterIDs":["fltr2"],"Increment":1,"FixedFee":2,"RecurrentFee":3}],"AttributeIDs":["attr2"],"RateProfileIDs":["rate_prf2"]}},"ThresholdIDs":null}`,
		},
		{
			name:   "Charges[0].Charging.Account.Tenant",
			fields: []string{"Charges[0]", "Charging", "Account", "Tenant"},
			exp:    "cgrates.org",
		},
		{
			name:   "Charges[0].Charging.Account.FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Account", "FilterIDs"},
			exp:    `["fltr1"]`,
		},
		{
			name:   "Charges[0].Charging.Account.FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "FilterIDs[0]"},
			exp:    "fltr1",
		},
		{
			name:   "Charges[0].Charging.Account.ID",
			fields: []string{"Charges[0]", "Charging", "Account", "ID"},
			exp:    "acc1",
		},
		{
			name:   "Charges[0].Charging.Account.Weights",
			fields: []string{"Charges[0]", "Charging", "Account", "Weights"},
			exp:    "fltr2;10",
		},
		{
			name:   "Charges[0].Charging.Account.Weights[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Weights[0]"},
			exp:    `{"FilterIDs":["fltr2"],"Weight":10}`,
		},
		{
			name:   "Charges[0].Charging.Account.Weights[0].FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Account", "Weights[0]", "FilterIDs"},
			exp:    `["fltr2"]`,
		},
		{
			name:   "Charges[0].Charging.Account.Weights[0].FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Weights[0]", "FilterIDs[0]"},
			exp:    "fltr2",
		},
		{
			name:   "Charges[0].Charging.Account.Weights[0].Weight",
			fields: []string{"Charges[0]", "Charging", "Account", "Weights[0]", "Weight"},
			exp:    "10",
		},
		{
			name:   "Charges[0].Charging.Account.Blockers",
			fields: []string{"Charges[0]", "Charging", "Account", "Blockers"},
			exp:    "fltr3;true",
		},
		{
			name:   "Charges[0].Charging.Account.Blockers[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Blockers[0]"},
			exp:    `{"FilterIDs":["fltr3"],"Blocker":true}`,
		},
		{
			name:   "Charges[0].Charging.Account.Blockers[0].FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Account", "Blockers[0]", "FilterIDs"},
			exp:    `["fltr3"]`,
		},
		{
			name:   "Charges[0].Charging.Account.Blockers[0].FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Blockers[0]", "FilterIDs[0]"},
			exp:    "fltr3",
		},
		{
			name:   "Charges[0].Charging.Account.Blockers[0].Blocker",
			fields: []string{"Charges[0]", "Charging", "Account", "Blockers[0]", "Blocker"},
			exp:    "true",
		},
		{
			name:   "Charges[0].Charging.Account.Opts",
			fields: []string{"Charges[0]", "Charging", "Account", "Opts"},
			exp:    `{"opt1":"value1"}`,
		},
		{
			name:   "Charges[0].Charging.Account.Opts.opt1",
			fields: []string{"Charges[0]", "Charging", "Account", "Opts", "opt1"},
			exp:    "value1",
		},
		{
			name:   "Charges[0].Charging.Account.Balances",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances"},
			exp:    `{"balance1":{"ID":"balance1","FilterIDs":["fltr4"],"Weights":[{"FilterIDs":["fltr3"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr3"],"Blocker":true}],"Type":"*monetary","Units":1,"UnitFactors":[{"FilterIDs":["fltr1","fltr2"],"Factor":2}],"Opts":{"opt1":"value1"},"CostIncrements":[{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}],"AttributeIDs":["attr1"],"RateProfileIDs":["rate_prf1"]},"balance2":{"ID":"balance2","FilterIDs":["fltr3"],"Weights":[{"FilterIDs":["fltr5"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr5"],"Blocker":true}],"Type":"*voice","Units":5,"UnitFactors":[{"FilterIDs":["fltr3","fltr4"],"Factor":1}],"Opts":{"opts1":"value1"},"CostIncrements":[{"FilterIDs":["fltr2"],"Increment":1,"FixedFee":2,"RecurrentFee":3}],"AttributeIDs":["attr2"],"RateProfileIDs":["rate_prf2"]}}`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1"},
			exp:    `{"ID":"balance1","FilterIDs":["fltr4"],"Weights":[{"FilterIDs":["fltr3"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr3"],"Blocker":true}],"Type":"*monetary","Units":1,"UnitFactors":[{"FilterIDs":["fltr1","fltr2"],"Factor":2}],"Opts":{"opt1":"value1"},"CostIncrements":[{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}],"AttributeIDs":["attr1"],"RateProfileIDs":["rate_prf1"]}`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.ID",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "ID"},
			exp:    "balance1",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "FilterIDs"},
			exp:    `["fltr4"]`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "FilterIDs[0]"},
			exp:    "fltr4",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Weights",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Weights"},
			exp:    "fltr3;20",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Weights[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Weights[0]"},
			exp:    `{"FilterIDs":["fltr3"],"Weight":20}`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Weights[0].FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Weights[0]", "FilterIDs"},
			exp:    `["fltr3"]`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Weights[0].FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Weights[0]", "FilterIDs[0]"},
			exp:    "fltr3",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Weights[0].Weight",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Weights[0]", "Weight"},
			exp:    "20",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Blockers",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Blockers"},
			exp:    "fltr3;true",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Blockers[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Blockers[0]"},
			exp:    `{"FilterIDs":["fltr3"],"Blocker":true}`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Blockers[0].FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Blockers[0]", "FilterIDs"},
			exp:    `["fltr3"]`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Blockers[0].FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Blockers[0]", "FilterIDs[0]"},
			exp:    "fltr3",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Blockers[0].Blocker",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Blockers[0]", "Blocker"},
			exp:    "true",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Type",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Type"},
			exp:    MetaMonetary,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Units",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Units"},
			exp:    "1",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.UnitFactors",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "UnitFactors"},
			exp:    `[{"FilterIDs":["fltr1","fltr2"],"Factor":2}]`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.UnitFactors[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "UnitFactors[0]"},
			exp:    `{"FilterIDs":["fltr1","fltr2"],"Factor":2}`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.UnitFactors[0].FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "UnitFactors[0]", "FilterIDs"},
			exp:    `["fltr1","fltr2"]`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.UnitFactors[0].FilterIDs[1]",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "UnitFactors[0]", "FilterIDs[1]"},
			exp:    "fltr2",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.UnitFactors[0].Factor",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "UnitFactors[0]", "Factor"},
			exp:    "2",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Opts",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Opts"},
			exp:    `{"opt1":"value1"}`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.Opts.opt1",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "Opts", "opt1"},
			exp:    "value1",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.CostIncrements",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "CostIncrements"},
			exp:    `[{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}]`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.CostIncrements[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "CostIncrements[0]"},
			exp:    `{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.CostIncrements[0].FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "CostIncrements[0]", "FilterIDs"},
			exp:    `["fltr3"]`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.CostIncrements[0].FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "CostIncrements[0]", "FilterIDs[0]"},
			exp:    "fltr3",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.CostIncrements[0].Increment",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "CostIncrements[0]", "Increment"},
			exp:    "3",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.CostIncrements[0].FixedFee",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "CostIncrements[0]", "FixedFee"},
			exp:    "1",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.CostIncrements[0].RecurrentFee",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "CostIncrements[0]", "RecurrentFee"},
			exp:    "2",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.AttributeIDs",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "AttributeIDs"},
			exp:    `["attr1"]`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.AttributeIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "AttributeIDs[0]"},
			exp:    "attr1",
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.RateProfileIDs",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "RateProfileIDs"},
			exp:    `["rate_prf1"]`,
		},
		{
			name:   "Charges[0].Charging.Account.Balances.balance1.RateProfileIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Account", "Balances", "balance1", "RateProfileIDs[0]"},
			exp:    "rate_prf1",
		},
		{
			name:   "Charges[0].Charging.Balance",
			fields: []string{"Charges[0]", "Charging", "Balance"},
			exp:    `{"ID":"balance1","FilterIDs":["fltr4"],"Weights":[{"FilterIDs":["fltr3"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr3"],"Blocker":true}],"Type":"*monetary","Units":1,"UnitFactors":[{"FilterIDs":["fltr1","fltr2"],"Factor":2}],"Opts":{"opt1":"value1"},"CostIncrements":[{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}],"AttributeIDs":["attr1"],"RateProfileIDs":["rate_prf1"]}`,
		},
		{
			name:   "Charges[0].Charging.Balance.ID",
			fields: []string{"Charges[0]", "Charging", "Balance", "ID"},
			exp:    "balance1",
		},
		{
			name:   "Charges[0].Charging.Balance.FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Balance", "FilterIDs"},
			exp:    `["fltr4"]`,
		},
		{
			name:   "Charges[0].Charging.Balance.FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Balance", "FilterIDs[0]"},
			exp:    "fltr4",
		},
		{
			name:   "Charges[0].Charging.Balance.Weights",
			fields: []string{"Charges[0]", "Charging", "Balance", "Weights"},
			exp:    "fltr3;20",
		},
		{
			name:   "Charges[0].Charging.Balance.Weights[0]",
			fields: []string{"Charges[0]", "Charging", "Balance", "Weights[0]"},
			exp:    `{"FilterIDs":["fltr3"],"Weight":20}`,
		},
		{
			name:   "Charges[0].Charging.Balance.Weights[0].FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Balance", "Weights[0]", "FilterIDs"},
			exp:    `["fltr3"]`,
		},
		{
			name:   "Charges[0].Charging.Balance.Weights[0].FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Balance", "Weights[0]", "FilterIDs[0]"},
			exp:    "fltr3",
		},
		{
			name:   "Charges[0].Charging.Balance.Weights[0].Weight",
			fields: []string{"Charges[0]", "Charging", "Balance", "Weights[0]", "Weight"},
			exp:    "20",
		},
		{
			name:   "Charges[0].Charging.Balance.Blockers",
			fields: []string{"Charges[0]", "Charging", "Balance", "Blockers"},
			exp:    "fltr3;true",
		},
		{
			name:   "Charges[0].Charging.Balance.Blockers[0]",
			fields: []string{"Charges[0]", "Charging", "Balance", "Blockers[0]"},
			exp:    `{"FilterIDs":["fltr3"],"Blocker":true}`,
		},
		{
			name:   "Charges[0].Charging.Balance.Blockers[0].FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Balance", "Blockers[0]", "FilterIDs"},
			exp:    `["fltr3"]`,
		},
		{
			name:   "Charges[0].Charging.Balance.Blockers[0].FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Balance", "Blockers[0]", "FilterIDs[0]"},
			exp:    "fltr3",
		},
		{
			name:   "Charges[0].Charging.Balance.Blockers[0].Blocker",
			fields: []string{"Charges[0]", "Charging", "Balance", "Blockers[0]", "Blocker"},
			exp:    "true",
		},
		{
			name:   "Charges[0].Charging.Balance.Type",
			fields: []string{"Charges[0]", "Charging", "Balance", "Type"},
			exp:    MetaMonetary,
		},
		{
			name:   "Charges[0].Charging.Balance.Units",
			fields: []string{"Charges[0]", "Charging", "Balance", "Units"},
			exp:    "1",
		},
		{
			name:   "Charges[0].Charging.Balance.UnitFactors",
			fields: []string{"Charges[0]", "Charging", "Balance", "UnitFactors"},
			exp:    `[{"FilterIDs":["fltr1","fltr2"],"Factor":2}]`,
		},
		{
			name:   "Charges[0].Charging.Balance.UnitFactors[0]",
			fields: []string{"Charges[0]", "Charging", "Balance", "UnitFactors[0]"},
			exp:    `{"FilterIDs":["fltr1","fltr2"],"Factor":2}`,
		},
		{
			name:   "Charges[0].Charging.Balance.UnitFactors[0].FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Balance", "UnitFactors[0]", "FilterIDs"},
			exp:    `["fltr1","fltr2"]`,
		},
		{
			name:   "Charges[0].Charging.Balance.UnitFactors[0].FilterIDs[1]",
			fields: []string{"Charges[0]", "Charging", "Balance", "UnitFactors[0]", "FilterIDs[1]"},
			exp:    "fltr2",
		},
		{
			name:   "Charges[0].Charging.Balance.UnitFactors[0].Factor",
			fields: []string{"Charges[0]", "Charging", "Balance", "UnitFactors[0]", "Factor"},
			exp:    "2",
		},
		{
			name:   "Charges[0].Charging.Balance.Opts",
			fields: []string{"Charges[0]", "Charging", "Balance", "Opts"},
			exp:    `{"opt1":"value1"}`,
		},
		{
			name:   "Charges[0].Charging.Balance.Opts.opt1",
			fields: []string{"Charges[0]", "Charging", "Balance", "Opts", "opt1"},
			exp:    "value1",
		},
		{
			name:   "Charges[0].Charging.Balance.CostIncrements",
			fields: []string{"Charges[0]", "Charging", "Balance", "CostIncrements"},
			exp:    `[{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}]`,
		},
		{
			name:   "Charges[0].Charging.Balance.CostIncrements[0]",
			fields: []string{"Charges[0]", "Charging", "Balance", "CostIncrements[0]"},
			exp:    `{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}`,
		},
		{
			name:   "Charges[0].Charging.Balance.CostIncrements[0].FilterIDs",
			fields: []string{"Charges[0]", "Charging", "Balance", "CostIncrements[0]", "FilterIDs"},
			exp:    `["fltr3"]`,
		},
		{
			name:   "Charges[0].Charging.Balance.CostIncrements[0].FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Balance", "CostIncrements[0]", "FilterIDs[0]"},
			exp:    "fltr3",
		},
		{
			name:   "Charges[0].Charging.Balance.CostIncrements[0].Increment",
			fields: []string{"Charges[0]", "Charging", "Balance", "CostIncrements[0]", "Increment"},
			exp:    "3",
		},
		{
			name:   "Charges[0].Charging.Balance.CostIncrements[0].FixedFee",
			fields: []string{"Charges[0]", "Charging", "Balance", "CostIncrements[0]", "FixedFee"},
			exp:    "1",
		},
		{
			name:   "Charges[0].Charging.Balance.CostIncrements[0].RecurrentFee",
			fields: []string{"Charges[0]", "Charging", "Balance", "CostIncrements[0]", "RecurrentFee"},
			exp:    "2",
		},
		{
			name:   "Charges[0].Charging.Balance.AttributeIDs",
			fields: []string{"Charges[0]", "Charging", "Balance", "AttributeIDs"},
			exp:    `["attr1"]`,
		},
		{
			name:   "Charges[0].Charging.Balance.AttributeIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Balance", "AttributeIDs[0]"},
			exp:    "attr1",
		},
		{
			name:   "Charges[0].Charging.Balance.RateProfileIDs",
			fields: []string{"Charges[0]", "Charging", "Balance", "RateProfileIDs"},
			exp:    `["rate_prf1"]`,
		},
		{
			name:   "Charges[0].Charging.Balance.RateProfileIDs[0]",
			fields: []string{"Charges[0]", "Charging", "Balance", "RateProfileIDs[0]"},
			exp:    "rate_prf1",
		},
		{
			name:   "Charges[0].Charging.UnitFactor",
			fields: []string{"Charges[0]", "Charging", "UnitFactor"},
			exp:    `{"FilterIDs":["fltr1","fltr2"],"Factor":2}`,
		},
		{
			name:   "Charges[0].Charging.UnitFactor.FilterIDs",
			fields: []string{"Charges[0]", "Charging", "UnitFactor", "FilterIDs"},
			exp:    `["fltr1","fltr2"]`,
		},
		{
			name:   "Charges[0].Charging.UnitFactor.FilterIDs[0]",
			fields: []string{"Charges[0]", "Charging", "UnitFactor", "FilterIDs[0]"},
			exp:    "fltr1",
		},
		{
			name:   "Charges[0].Charging.UnitFactor.Factor",
			fields: []string{"Charges[0]", "Charging", "UnitFactor", "Factor"},
			exp:    "2",
		},
		{
			name:   "Charges[0].Charging.Rating",
			fields: []string{"Charges[0]", "Charging", "Rating"},
			exp:    `{"IntervalStart":5,"Increments":[{"IncrementStart":9,"RateIntervalIndex":3,"RateID":"rate1","CompressFactor":1,"Usage":10}],"CompressFactor":3}`,
		},
		{
			name:   "Charges[0].Charging.Rating.IntervalStart",
			fields: []string{"Charges[0]", "Charging", "Rating", "IntervalStart"},
			exp:    "5",
		},
		{
			name:   "Charges[0].Charging.Rating.Increments",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments"},
			exp:    `[{"IncrementStart":9,"RateIntervalIndex":3,"RateID":"rate1","CompressFactor":1,"Usage":10}]`,
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0]",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]"},
			exp:    `{"IncrementStart":9,"RateIntervalIndex":3,"RateID":"rate1","CompressFactor":1,"Usage":10}`,
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0].IncrementStart",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]", "IncrementStart"},
			exp:    "9",
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0].RateIntervalIndex",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]", "RateIntervalIndex"},
			exp:    "3",
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0].RateID",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]", "RateID"},
			exp:    "rate1",
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0].CompressFactor",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]", "CompressFactor"},
			exp:    "1",
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0].Usage",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]", "Usage"},
			exp:    "10",
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0].Rate",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]", "Rate"},
			exp:    `{"IntervalStart":1,"FixedFee":2,"RecurrentFee":3,"Unit":4,"Increment":5}`,
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0].Rate.IntervalStart",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]", "Rate", "IntervalStart"},
			exp:    "1",
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0].Rate.FixedFee",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]", "Rate", "FixedFee"},
			exp:    "2",
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0].Rate.RecurrentFee",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]", "Rate", "RecurrentFee"},
			exp:    "3",
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0].Rate.Unit",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]", "Rate", "Unit"},
			exp:    "4",
		},
		{
			name:   "Charges[0].Charging.Rating.Increments[0].Rate.Increment",
			fields: []string{"Charges[0]", "Charging", "Rating", "Increments[0]", "Rate", "Increment"},
			exp:    "5",
		},
		{
			name:   "Charges[0].Charging.Rating.CompressFactor",
			fields: []string{"Charges[0]", "Charging", "Rating", "CompressFactor"},
			exp:    "3",
		},
		{
			name:   "Charges[1].Charging",
			fields: []string{"Charges[1]", "Charging"},
			exp:    `{"IntervalStart":4,"Increments":[{"IncrementStart":5,"RateIntervalIndex":1,"RateID":"rate1","CompressFactor":1,"Usage":6},{"IncrementStart":7,"RateIntervalIndex":2,"RateID":"rate2","CompressFactor":1,"Usage":8}],"CompressFactor":3}`,
		},
		{
			name:   "Charges[1].Charging.IntervalStart",
			fields: []string{"Charges[1]", "Charging", "IntervalStart"},
			exp:    "4",
		},
		{
			name:   "Charges[1].Charging.Increments",
			fields: []string{"Charges[1]", "Charging", "Increments"},
			exp:    `[{"IncrementStart":5,"RateIntervalIndex":1,"RateID":"rate1","CompressFactor":1,"Usage":6},{"IncrementStart":7,"RateIntervalIndex":2,"RateID":"rate2","CompressFactor":1,"Usage":8}]`,
		},
		{
			name:   "Charges[1].Charging.Increments[1]",
			fields: []string{"Charges[1]", "Charging", "Increments[1]"},
			exp:    `{"IncrementStart":7,"RateIntervalIndex":2,"RateID":"rate2","CompressFactor":1,"Usage":8}`,
		},
		{
			name:   "Charges[1].Charging.Increments[1].IncrementStart",
			fields: []string{"Charges[1]", "Charging", "Increments[1]", "IncrementStart"},
			exp:    "7",
		},
		{
			name:   "Charges[1].Charging.Increments[1].RateIntervalIndex",
			fields: []string{"Charges[1]", "Charging", "Increments[1]", "RateIntervalIndex"},
			exp:    "2",
		},
		{
			name:   "Charges[1].Charging.Increments[1].RateID",
			fields: []string{"Charges[1]", "Charging", "Increments[1]", "RateID"},
			exp:    "rate2",
		},
		{
			name:   "Charges[1].Charging.Increments[1].CompressFactor",
			fields: []string{"Charges[1]", "Charging", "Increments[1]", "CompressFactor"},
			exp:    "1",
		},
		{
			name:   "Charges[1].Charging.Increments[1].Usage",
			fields: []string{"Charges[1]", "Charging", "Increments[1]", "Usage"},
			exp:    "8",
		},
		{
			name:   "Charges[1].Charging.Increments[1].Rate",
			fields: []string{"Charges[1]", "Charging", "Increments[1]", "Rate"},
			exp:    `{"IntervalStart":6,"FixedFee":7,"RecurrentFee":8,"Unit":9,"Increment":10}`,
		},
		{
			name:   "Charges[1].Charging.Increments[1].Rate.IntervalStart",
			fields: []string{"Charges[1]", "Charging", "Increments[1]", "Rate", "IntervalStart"},
			exp:    "6",
		},
		{
			name:   "Charges[1].Charging.Increments[1].Rate.FixedFee",
			fields: []string{"Charges[1]", "Charging", "Increments[1]", "Rate", "FixedFee"},
			exp:    "7",
		},
		{
			name:   "Charges[1].Charging.Increments[1].Rate.RecurrentFee",
			fields: []string{"Charges[1]", "Charging", "Increments[1]", "Rate", "RecurrentFee"},
			exp:    "8",
		},
		{
			name:   "Charges[1].Charging.Increments[1].Rate.Unit",
			fields: []string{"Charges[1]", "Charging", "Increments[1]", "Rate", "Unit"},
			exp:    "9",
		},
		{
			name:   "Charges[1].Charging.Increments[1].Rate.Increment",
			fields: []string{"Charges[1]", "Charging", "Increments[1]", "Rate", "Increment"},
			exp:    "10",
		},
		{
			name:   "Charges[1].Charging.CompressFactor",
			fields: []string{"Charges[1]", "Charging", "CompressFactor"},
			exp:    "3",
		},
	}

	for _, tc := range testcases {

		t.Run(tc.name, func(t *testing.T) {
			if val, err := ec.FieldAsString(tc.fields); err != nil {
				t.Error(err)
			} else if tc.exp != val {
				t.Errorf("expected: %s,\nreceived: %s", tc.exp, val)
			}
		})
	}
}
