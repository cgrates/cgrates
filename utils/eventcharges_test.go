// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"reflect"
	"testing"
	"time"
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
		Abstracts: NewDecimal(1, 1),
		Concretes: NewDecimal(1, 1),
	}

	ec2 := &EventCharges{
		Abstracts: NewDecimal(2, 1),
		Concretes: NewDecimal(2, 1),
	}

	received := &EventCharges{}
	expected := &EventCharges{
		Abstracts: NewDecimal(3, 1),
		Concretes: NewDecimal(3, 1),
	}
	received.Merge(ec1, ec2)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestECMergeAbstracts(t *testing.T) {
	ec1 := &EventCharges{
		Abstracts: NewDecimal(1, 1),
		Concretes: NewDecimal(1, 1),
	}

	ec2 := &EventCharges{
		Abstracts: NewDecimal(2, 1),
		Concretes: NewDecimal(2, 1),
	}

	received := &EventCharges{
		Abstracts: NewDecimal(3, 1),
		Concretes: NewDecimal(3, 1),
	}
	expected := &EventCharges{
		Abstracts: NewDecimal(6, 1),
		Concretes: NewDecimal(6, 1),
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
						cost:              NewDecimal(4321, 5),
					},
					{
						IncrementStart:    NewDecimalFromFloat64(4.321),
						RateIntervalIndex: 1,
						RateID:            "Rate2",
						CompressFactor:    1,
						Usage:             NewDecimalFromFloat64(-123),
						cost:              NewDecimal(123, 1),
					},
				},
				CompressFactor: 1,
				cost:           NewDecimal(4321, 5),
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
						cost:              NewDecimal(3456, 5),
					},
					{
						IncrementStart:    NewDecimalFromFloat64(4.321),
						RateIntervalIndex: 2,
						RateID:            "Rate2",
						CompressFactor:    2,
						Usage:             NewDecimalFromFloat64(-123),
						cost:              NewDecimal(321, 1),
					},
				},
				CompressFactor: 1,
				cost:           NewDecimal(345, 2),
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
		Concretes: NewDecimal(152, 1),
		Abstracts: NewDecimal(145, 1),
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
				Units:           NewDecimal(10, 0),
				BalanceLimit:    NewDecimal(0, 0),
				UnitFactorID:    "unit_factor1",
				AttributeIDs:    []string{"attr1", "attr2"},
				RatingID:        "rating2",
				JoinedChargeIDs: []string{"joined_charge"},
			},
			"joined_charge": {
				AccountID:       "acc2",
				BalanceID:       "balance2",
				Units:           NewDecimal(10, 0),
				BalanceLimit:    NewDecimal(0, 0),
				UnitFactorID:    "unit_factor2",
				AttributeIDs:    []string{"attr3", "attr4"},
				RatingID:        "rating3",
				JoinedChargeIDs: []string{},
			},
			"accounting2": nil,
		},
		UnitFactors: map[string]*UnitFactor{
			"unit_factor1": {
				FilterIDs: []string{"fltr1", "fltr2"},
				Factor:    NewDecimal(2, 0),
			},
			"unit_factor2": {
				FilterIDs: []string{"fltr3", "fltr4"},
				Factor:    NewDecimal(3, 0),
			},
			"unit_factor3": nil,
		},
		Rating: map[string]*RateSInterval{
			"rating1": {
				IntervalStart: NewDecimal(4, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(5, 0),
						RateIntervalIndex: 1,
						RateID:            "rate1",
						CompressFactor:    1,
						Usage:             NewDecimal(6, 0),
					},
					{
						IncrementStart:    NewDecimal(7, 0),
						RateIntervalIndex: 2,
						RateID:            "rate2",
						CompressFactor:    1,
						Usage:             NewDecimal(8, 0),
					},
				},
				CompressFactor: 3,
			},
			"rating2": {
				IntervalStart: NewDecimal(5, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(9, 0),
						RateIntervalIndex: 3,
						RateID:            "rate1",
						CompressFactor:    1,
						Usage:             NewDecimal(10, 0),
					},
				},
				CompressFactor: 3,
			},
			"rating3": {
				IntervalStart: &Decimal{},
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(11, 0),
						RateIntervalIndex: 4,
						RateID:            "rate2",
						CompressFactor:    5,
						Usage:             NewDecimal(12, 0),
					},
				},
				CompressFactor: 3,
			},
			"rating4": nil,
		},
		Rates: map[string]*IntervalRate{
			"rate1": {
				IntervalStart: NewDecimal(1, 0),
				FixedFee:      NewDecimal(2, 0),
				RecurrentFee:  NewDecimal(3, 0),
				Unit:          NewDecimal(4, 0),
				Increment:     NewDecimal(5, 0),
			},
			"rate2": {
				IntervalStart: NewDecimal(6, 0),
				FixedFee:      NewDecimal(7, 0),
				RecurrentFee:  NewDecimal(8, 0),
				Unit:          NewDecimal(9, 0),
				Increment:     NewDecimal(10, 0),
			},
			"rate3": nil,
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
						Units: NewDecimal(1, 0),
						UnitFactors: []*UnitFactor{
							{
								FilterIDs: []string{"fltr1", "fltr2"},
								Factor:    NewDecimal(2, 0),
							},
						},
						Opts: map[string]any{
							"opt1": "value1",
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"fltr3"},
								Increment:    NewDecimal(3, 0),
								FixedFee:     NewDecimal(1, 0),
								RecurrentFee: NewDecimal(2, 0),
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
						Units: NewDecimal(5, 0),
						UnitFactors: []*UnitFactor{
							{
								FilterIDs: []string{"fltr3", "fltr4"},
								Factor:    NewDecimal(1, 0),
							},
						},
						Opts: map[string]any{
							"opts1": "value1",
						},
						CostIncrements: []*CostIncrement{
							{
								FilterIDs:    []string{"fltr2"},
								Increment:    NewDecimal(1, 0),
								FixedFee:     NewDecimal(2, 0),
								RecurrentFee: NewDecimal(3, 0),
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
						Units: NewDecimal(11, 0),
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
								Increment:    NewDecimal(12, 0),
								FixedFee:     NewDecimal(13, 0),
								RecurrentFee: NewDecimal(14, 0),
							},
						},
						AttributeIDs:   []string{"attr3"},
						RateProfileIDs: []string{"rate_prf3"},
					},
				},
			},
			"acc3": nil,
		},
	}

	testcases := []struct {
		name   string
		fields []string
		exp    any
		expErr string
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
		{
			name:   "error case: NOT_FOUND",
			fields: []string{"Charges[0]", "Charges[1]", "ChargingID"},
			exp:    "",
			expErr: "NOT_FOUND",
		},
		{
			name:   "error case: unsupported field prefix",
			fields: []string{"Charges[1]", "test"},
			exp:    "",
			expErr: "unsupported field prefix: <test>",
		},
		{
			name:   "Accounting",
			fields: []string{"Accounting"},
			exp:    `{"accounting1":{"AccountID":"acc1","BalanceID":"balance1","Units":10,"BalanceLimit":0,"UnitFactorID":"unit_factor1","AttributeIDs":["attr1","attr2"],"RatingID":"rating2","JoinedChargeIDs":["joined_charge"]},"accounting2":null,"joined_charge":{"AccountID":"acc2","BalanceID":"balance2","Units":10,"BalanceLimit":0,"UnitFactorID":"unit_factor2","AttributeIDs":["attr3","attr4"],"RatingID":"rating3","JoinedChargeIDs":[]}}`,
		},
		{
			name:   "Accounting.accounting1",
			fields: []string{"Accounting", "accounting1"},
			exp:    `{"AccountID":"acc1","BalanceID":"balance1","Units":10,"BalanceLimit":0,"UnitFactorID":"unit_factor1","AttributeIDs":["attr1","attr2"],"RatingID":"rating2","JoinedChargeIDs":["joined_charge"]}`,
		},
		{
			name:   "Accounting.accounting1.AccountID",
			fields: []string{"Accounting", "accounting1", "AccountID"},
			exp:    "acc1",
		},
		{
			name:   "Accounting.accounting1.BalanceID",
			fields: []string{"Accounting", "accounting1", "BalanceID"},
			exp:    "balance1",
		},
		{
			name:   "Accounting.accounting1.Units",
			fields: []string{"Accounting", "accounting1", "Units"},
			exp:    "10",
		},
		{
			name:   "Accounting.accounting1.BalanceLimit",
			fields: []string{"Accounting", "accounting1", "BalanceLimit"},
			exp:    "0",
		},
		{
			name:   "Accounting.accounting1.UnitFactorID",
			fields: []string{"Accounting", "accounting1", "UnitFactorID"},
			exp:    "unit_factor1",
		},
		{
			name:   "Accounting.accounting1.AttributeIDs",
			fields: []string{"Accounting", "accounting1", "AttributeIDs"},
			exp:    `["attr1","attr2"]`,
		},
		{
			name:   "Accounting.accounting1.RatingID",
			fields: []string{"Accounting", "accounting1", "RatingID"},
			exp:    "rating2",
		},
		{
			name:   "Accounting.accounting1.JoinedChargeIDs",
			fields: []string{"Accounting", "accounting1", "JoinedChargeIDs"},
			exp:    `["joined_charge"]`,
		},
		{
			name:   "Accounting.joined_charge",
			fields: []string{"Accounting", "joined_charge"},
			exp:    `{"AccountID":"acc2","BalanceID":"balance2","Units":10,"BalanceLimit":0,"UnitFactorID":"unit_factor2","AttributeIDs":["attr3","attr4"],"RatingID":"rating3","JoinedChargeIDs":[]}`,
		},
		{
			name:   "Accounting.joined_charge.AccountID",
			fields: []string{"Accounting", "joined_charge", "AccountID"},
			exp:    "acc2",
		},
		{
			name:   "Accounting.joined_charge.BalanceID",
			fields: []string{"Accounting", "joined_charge", "BalanceID"},
			exp:    "balance2",
		},
		{
			name:   "Accounting.joined_charge.Units",
			fields: []string{"Accounting", "joined_charge", "Units"},
			exp:    "10",
		},
		{
			name:   "Accounting.joined_charge.BalanceLimit",
			fields: []string{"Accounting", "joined_charge", "BalanceLimit"},
			exp:    "0",
		},
		{
			name:   "Accounting.joined_charge.UnitFactorID",
			fields: []string{"Accounting", "joined_charge", "UnitFactorID"},
			exp:    "unit_factor2",
		},
		{
			name:   "Accounting.joined_charge.AttributeIDs",
			fields: []string{"Accounting", "joined_charge", "AttributeIDs"},
			exp:    `["attr3","attr4"]`,
		},
		{
			name:   "Accounting.joined_charge.RatingID",
			fields: []string{"Accounting", "joined_charge", "RatingID"},
			exp:    "rating3",
		},
		{
			name:   "Accounting.joined_charge.JoinedChargeIDs",
			fields: []string{"Accounting", "joined_charge", "JoinedChargeIDs"},
			exp:    `[]`,
		},
		{
			name:   "Accounting.accounting2",
			fields: []string{"Accounting", "accounting2"},
			exp:    "",
		},
		{
			name:   "Accounting.accounting1.JoinedChargeIDs",
			fields: []string{"Accounting", "accounting1", "JoinedChargeIDs"},
			exp:    `["joined_charge"]`,
		},
		// TODO: uncomment after the panic is fixed
		// {
		// 	name:   "Rating",
		// 	fields: []string{"Rating"},
		// 	exp:    "",
		// },
		{
			name:   "Rating",
			fields: []string{"Rating", "rating1"},
			exp:    `{"IntervalStart":4,"Increments":[{"IncrementStart":5,"RateIntervalIndex":1,"RateID":"rate1","CompressFactor":1,"Usage":6},{"IncrementStart":7,"RateIntervalIndex":2,"RateID":"rate2","CompressFactor":1,"Usage":8}],"CompressFactor":3}`,
		},
		{
			name:   "Rating.rating4",
			fields: []string{"Rating", "rating4"},
			exp:    "",
		},
		{
			name:   "Rating.rating2.CompressFactor",
			fields: []string{"Rating", "rating2", "CompressFactor"},
			exp:    "3",
		},
		{
			name:   "Rating.rating1.IntervalStart",
			fields: []string{"Rating", "rating1"},
			exp:    `{"IntervalStart":4,"Increments":[{"IncrementStart":5,"RateIntervalIndex":1,"RateID":"rate1","CompressFactor":1,"Usage":6},{"IncrementStart":7,"RateIntervalIndex":2,"RateID":"rate2","CompressFactor":1,"Usage":8}],"CompressFactor":3}`,
		},
		{
			name:   "Rating.rating1.Increments",
			fields: []string{"Rating", "rating1"},
			exp:    `{"IntervalStart":4,"Increments":[{"IncrementStart":5,"RateIntervalIndex":1,"RateID":"rate1","CompressFactor":1,"Usage":6},{"IncrementStart":7,"RateIntervalIndex":2,"RateID":"rate2","CompressFactor":1,"Usage":8}],"CompressFactor":3}`,
		},
		{
			name:   "Rating.rating1.CompressFactor",
			fields: []string{"Rating", "rating1", "CompressFactor"},
			exp:    "3",
		},
		{
			name:   "Rating.rating1.Increments[0].IncrementStart",
			fields: []string{"Rating", "rating1", "Increments[0]", "IncrementStart"},
			exp:    "5",
		},
		{
			name:   "Rating.rating1.Increments[0].RateIntervalIndex",
			fields: []string{"Rating", "rating1", "Increments[0]", "RateIntervalIndex"},
			exp:    "1",
		},
		{
			name:   "Rating.rating1.Increments[0].RateID",
			fields: []string{"Rating", "rating1", "Increments[0]", "RateID"},
			exp:    "rate1",
		},
		{
			name:   "Rating.rating1.Increments[0].CompressFactor",
			fields: []string{"Rating", "rating1", "Increments[0]", "CompressFactor"},
			exp:    "1",
		},
		{
			name:   "Rating.rating1.Increments[0].Usage",
			fields: []string{"Rating", "rating1", "Increments[0]", "Usage"},
			exp:    "6",
		},
		{
			name:   "Rate",
			fields: []string{"Rate"},
			exp:    `{"rate1":{"IntervalStart":1,"FixedFee":2,"RecurrentFee":3,"Unit":4,"Increment":5},"rate2":{"IntervalStart":6,"FixedFee":7,"RecurrentFee":8,"Unit":9,"Increment":10},"rate3":null}`,
		},
		{
			name:   "Rate.rate1",
			fields: []string{"Rate", "rate1"},
			exp:    `{"IntervalStart":1,"FixedFee":2,"RecurrentFee":3,"Unit":4,"Increment":5}`,
		},
		{
			name:   "Rate.rate1.IntervalStart",
			fields: []string{"Rate", "rate1", "IntervalStart"},
			exp:    "1",
		},
		{
			name:   "Rate.rate1.FixedFee",
			fields: []string{"Rate", "rate1", "FixedFee"},
			exp:    "2",
		},
		{
			name:   "Rate.rate1.RecurrentFee",
			fields: []string{"Rate", "rate1", "RecurrentFee"},
			exp:    "3",
		},
		{
			name:   "Rate.rate1.Unit",
			fields: []string{"Rate", "rate1", "Unit"},
			exp:    "4",
		},
		{
			name:   "Rate.rate1.Increment",
			fields: []string{"Rate", "rate1", "Increment"},
			exp:    "5",
		},
		{
			name:   "Rate.rate2",
			fields: []string{"Rate", "rate2"},
			exp:    `{"IntervalStart":6,"FixedFee":7,"RecurrentFee":8,"Unit":9,"Increment":10}`,
		},
		{
			name:   "Rate.rate2.IntervalStart",
			fields: []string{"Rate", "rate2", "IntervalStart"},
			exp:    "6",
		},
		{
			name:   "Rate.rate2.FixedFee",
			fields: []string{"Rate", "rate2", "FixedFee"},
			exp:    "7",
		},
		{
			name:   "Rate.rate2.RecurrentFee",
			fields: []string{"Rate", "rate2", "RecurrentFee"},
			exp:    "8",
		},
		{
			name:   "Rate.rate2.Unit",
			fields: []string{"Rate", "rate2", "Unit"},
			exp:    "9",
		},
		{
			name:   "Rate.rate2.Increment",
			fields: []string{"Rate", "rate2", "Increment"},
			exp:    "10",
		},
		{
			name:   "Rate.rate3",
			fields: []string{"Rate", "rate3"},
			exp:    "",
		},
		{
			name:   "UnitFactor",
			fields: []string{"UnitFactor"},
			exp:    `{"unit_factor1":{"FilterIDs":["fltr1","fltr2"],"Factor":2},"unit_factor2":{"FilterIDs":["fltr3","fltr4"],"Factor":3},"unit_factor3":null}`,
		},
		{
			name:   "UnitFactor.unit_factor1",
			fields: []string{"UnitFactor", "unit_factor1"},
			exp:    `{"FilterIDs":["fltr1","fltr2"],"Factor":2}`,
		},
		{
			name:   "UnitFactor.unit_factor1.FilterIDs",
			fields: []string{"UnitFactor", "unit_factor1", "FilterIDs"},
			exp:    `["fltr1","fltr2"]`,
		},
		{
			name:   "UnitFactor.unit_factor1.Factor",
			fields: []string{"UnitFactor", "unit_factor1", "Factor"},
			exp:    "2",
		},
		{
			name:   "UnitFactor.unit_factor2",
			fields: []string{"UnitFactor", "unit_factor2"},
			exp:    `{"FilterIDs":["fltr3","fltr4"],"Factor":3}`,
		},
		{
			name:   "UnitFactor.unit_factor2.FilterIDs",
			fields: []string{"UnitFactor", "unit_factor2", "FilterIDs"},
			exp:    `["fltr3","fltr4"]`,
		},
		{
			name:   "UnitFactor.unit_factor2.Factor",
			fields: []string{"UnitFactor", "unit_factor2", "Factor"},
			exp:    "3",
		},
		{
			name:   "UnitFactor.unit_factor3",
			fields: []string{"UnitFactor", "unit_factor3"},
			exp:    "",
		},
		// TODO: uncomment after the panic is fixed
		// {
		// 	name:   "Account",
		// 	fields: []string{"Account"},
		// 	exp:    "",
		// },
		{
			name:   "Account",
			fields: []string{"Account", "acc1"},
			exp:    `{"Tenant":"cgrates.org","ID":"acc1","FilterIDs":["fltr1"],"Weights":[{"FilterIDs":["fltr2"],"Weight":10}],"Blockers":[{"FilterIDs":["fltr3"],"Blocker":true}],"Opts":{"opt1":"value1"},"Balances":{"balance1":{"ID":"balance1","FilterIDs":["fltr4"],"Weights":[{"FilterIDs":["fltr3"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr3"],"Blocker":true}],"Type":"*monetary","Units":1,"UnitFactors":[{"FilterIDs":["fltr1","fltr2"],"Factor":2}],"Opts":{"opt1":"value1"},"CostIncrements":[{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}],"AttributeIDs":["attr1"],"RateProfileIDs":["rate_prf1"]},"balance2":{"ID":"balance2","FilterIDs":["fltr3"],"Weights":[{"FilterIDs":["fltr5"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr5"],"Blocker":true}],"Type":"*voice","Units":5,"UnitFactors":[{"FilterIDs":["fltr3","fltr4"],"Factor":1}],"Opts":{"opts1":"value1"},"CostIncrements":[{"FilterIDs":["fltr2"],"Increment":1,"FixedFee":2,"RecurrentFee":3}],"AttributeIDs":["attr2"],"RateProfileIDs":["rate_prf2"]}},"ThresholdIDs":null}`,
		},
		{
			name:   "Account.acc1",
			fields: []string{"Account", "acc1"},
			exp:    `{"Tenant":"cgrates.org","ID":"acc1","FilterIDs":["fltr1"],"Weights":[{"FilterIDs":["fltr2"],"Weight":10}],"Blockers":[{"FilterIDs":["fltr3"],"Blocker":true}],"Opts":{"opt1":"value1"},"Balances":{"balance1":{"ID":"balance1","FilterIDs":["fltr4"],"Weights":[{"FilterIDs":["fltr3"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr3"],"Blocker":true}],"Type":"*monetary","Units":1,"UnitFactors":[{"FilterIDs":["fltr1","fltr2"],"Factor":2}],"Opts":{"opt1":"value1"},"CostIncrements":[{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}],"AttributeIDs":["attr1"],"RateProfileIDs":["rate_prf1"]},"balance2":{"ID":"balance2","FilterIDs":["fltr3"],"Weights":[{"FilterIDs":["fltr5"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr5"],"Blocker":true}],"Type":"*voice","Units":5,"UnitFactors":[{"FilterIDs":["fltr3","fltr4"],"Factor":1}],"Opts":{"opts1":"value1"},"CostIncrements":[{"FilterIDs":["fltr2"],"Increment":1,"FixedFee":2,"RecurrentFee":3}],"AttributeIDs":["attr2"],"RateProfileIDs":["rate_prf2"]}},"ThresholdIDs":null}`,
		},
		{
			name:   "Account.acc1.Tenant",
			fields: []string{"Account", "acc1", "Tenant"},
			exp:    "cgrates.org",
		},
		{
			name:   "Account.acc1.FilterIDs",
			fields: []string{"Account", "acc1", "FilterIDs"},
			exp:    `["fltr1"]`,
		},
		{
			name:   "Account.acc1.ID",
			fields: []string{"Account", "acc1", "ID"},
			exp:    "acc1",
		},
		{
			name:   "Account.acc1.Weights",
			fields: []string{"Account", "acc1", "Weights"},
			exp:    "fltr2;10",
		},
		{
			name:   "Account.acc1.Blockers",
			fields: []string{"Account", "acc1", "Blockers"},
			exp:    "fltr3;true",
		},
		{
			name:   "Account.acc1.Opts",
			fields: []string{"Account", "acc1", "Opts"},
			exp:    `{"opt1":"value1"}`,
		},
		{
			name:   "Account.acc1.Balances",
			fields: []string{"Account", "acc1", "Balances"},
			exp:    `{"balance1":{"ID":"balance1","FilterIDs":["fltr4"],"Weights":[{"FilterIDs":["fltr3"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr3"],"Blocker":true}],"Type":"*monetary","Units":1,"UnitFactors":[{"FilterIDs":["fltr1","fltr2"],"Factor":2}],"Opts":{"opt1":"value1"},"CostIncrements":[{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}],"AttributeIDs":["attr1"],"RateProfileIDs":["rate_prf1"]},"balance2":{"ID":"balance2","FilterIDs":["fltr3"],"Weights":[{"FilterIDs":["fltr5"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr5"],"Blocker":true}],"Type":"*voice","Units":5,"UnitFactors":[{"FilterIDs":["fltr3","fltr4"],"Factor":1}],"Opts":{"opts1":"value1"},"CostIncrements":[{"FilterIDs":["fltr2"],"Increment":1,"FixedFee":2,"RecurrentFee":3}],"AttributeIDs":["attr2"],"RateProfileIDs":["rate_prf2"]}}`,
		},
		{
			name:   "Account.acc1.Weights[0].FilterIDs",
			fields: []string{"Account", "acc1", "Weights[0]", "FilterIDs"},
			exp:    `["fltr2"]`,
		},
		{
			name:   "Account.acc1.Weights[0].Weight",
			fields: []string{"Account", "acc1", "Weights[0]", "Weight"},
			exp:    "10",
		},
		{
			name:   "Account.acc1.Blockers[0].FilterIDs",
			fields: []string{"Account", "acc1", "Blockers[0]", "FilterIDs"},
			exp:    `["fltr3"]`,
		},
		{
			name:   "Account.acc1.Blockers[0].Weight",
			fields: []string{"Account", "acc1", "Blockers[0]", "Blocker"},
			exp:    "true",
		},
		{
			name:   "Account.acc1.Balances.balance1",
			fields: []string{"Account", "acc1", "Balances", "balance1"},
			exp:    `{"ID":"balance1","FilterIDs":["fltr4"],"Weights":[{"FilterIDs":["fltr3"],"Weight":20}],"Blockers":[{"FilterIDs":["fltr3"],"Blocker":true}],"Type":"*monetary","Units":1,"UnitFactors":[{"FilterIDs":["fltr1","fltr2"],"Factor":2}],"Opts":{"opt1":"value1"},"CostIncrements":[{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}],"AttributeIDs":["attr1"],"RateProfileIDs":["rate_prf1"]}`,
		},
		{
			name:   "Account.acc1.Balances.balance1,ID",
			fields: []string{"Account", "acc1", "Balances", "balance1", "ID"},
			exp:    "balance1",
		},
		{
			name:   "Account.acc1.Balances.balance1.FilterIDs",
			fields: []string{"Account", "acc1", "Balances", "balance1", "FilterIDs"},
			exp:    `["fltr4"]`,
		},
		{
			name:   "Account.acc1.Balances.balance1.Weights",
			fields: []string{"Account", "acc1", "Balances", "balance1", "Weights"},
			exp:    "fltr3;20",
		},
		{
			name:   "Account.acc1.Balances.balance1.Blockers",
			fields: []string{"Account", "acc1", "Balances", "balance1", "Blockers"},
			exp:    "fltr3;true",
		},
		{
			name:   "Account.acc1.Balances.balance1.Type",
			fields: []string{"Account", "acc1", "Balances", "balance1", "Type"},
			exp:    "*monetary",
		},
		{
			name:   "Account.acc1.Balances.balance1.Units",
			fields: []string{"Account", "acc1", "Balances", "balance1", "Units"},
			exp:    "1",
		},
		{
			name:   "Account.acc1.Balances.balance1.UnitFactors",
			fields: []string{"Account", "acc1", "Balances", "balance1", "UnitFactors"},
			exp:    `[{"FilterIDs":["fltr1","fltr2"],"Factor":2}]`,
		},
		{
			name:   "Account.acc1.Balances.balance1.Opts",
			fields: []string{"Account", "acc1", "Balances", "balance1", "Opts"},
			exp:    `{"opt1":"value1"}`,
		},
		{
			name:   "Account.acc1.Balances.balance1.CostIncrements",
			fields: []string{"Account", "acc1", "Balances", "balance1", "CostIncrements"},
			exp:    `[{"FilterIDs":["fltr3"],"Increment":3,"FixedFee":1,"RecurrentFee":2}]`,
		},
		{
			name:   "Account.acc1.Balances.balance1.AttributeIDs",
			fields: []string{"Account", "acc1", "Balances", "balance1", "AttributeIDs"},
			exp:    `["attr1"]`,
		},
		{
			name:   "Account.acc1.Balances.balance1.RateProfileIDs",
			fields: []string{"Account", "acc1", "Balances", "balance1", "RateProfileIDs"},
			exp:    `["rate_prf1"]`,
		},
		// TODO: uncomment after the panic is fixed
		// {
		// 	name:   "Account",
		// 	fields: []string{"Account", "acc2"},
		// 	exp:    "",
		// },
		{
			name:   "Account.acc3",
			fields: []string{"Account", "acc3"},
			exp:    "",
		},
	}

	for _, tc := range testcases {

		t.Run(tc.name, func(t *testing.T) {
			if val, err := ec.FieldAsString(tc.fields); err != nil && err.Error() != tc.expErr {
				t.Errorf("Expected %v \n, recieved %v", tc.expErr, err)
			} else if tc.exp != val {
				t.Errorf("expected: %s,\nreceived: %s", tc.exp, val)
			}
		})
	}
}

func TestEventChargesFieldAsInterfaceErrors(t *testing.T) {
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
		Accounting:  nil,
		UnitFactors: nil,
		Rating:      nil,
		Rates:       nil,
		Accounts:    nil,
	}

	tests := []struct {
		name   string
		ec     *EventCharges
		fields []string
		expErr string
	}{
		{
			name:   "Abstracts",
			ec:     ec,
			fields: []string{"Abstracts", "test"},
			expErr: "NOT_FOUND",
		},
		{
			name:   "Concretes",
			ec:     ec,
			fields: []string{"Concretes", "test"},
			expErr: "NOT_FOUND",
		},
		{
			name:   "Charges",
			ec:     ec,
			fields: []string{"Charges", "test"},
			expErr: "NOT_FOUND",
		},
		{
			name:   "Charges index out of range",
			ec:     ec,
			fields: []string{"Charges[2]"},
			expErr: "NOT_FOUND",
		},
		{
			name:   "unsupported field prefix",
			ec:     ec,
			fields: []string{"Charges[invalid]"},
			expErr: "unsupported field prefix: <Charges[invalid]>",
		},
		{
			name:   "nil Accounting",
			ec:     ec,
			fields: []string{"Accounting", "accounting1"},
			expErr: "NOT_FOUND",
		},
		{
			name:   "unsupported field: Accounting",
			ec:     ec,
			fields: []string{"Accounting[0]"},
			expErr: "unsupported field prefix: <Accounting>",
		},
		{
			name:   "nil UnitFactors",
			ec:     ec,
			fields: []string{"UnitFactor", "unit_factor1"},
			expErr: "NOT_FOUND",
		},
		{
			name:   "unsupported field: UnitFactor",
			ec:     ec,
			fields: []string{"UnitFactor[0]"},
			expErr: "unsupported field prefix: <UnitFactor>",
		},
		{
			name:   "nil Rating",
			ec:     ec,
			fields: []string{"Rating", "rating1"},
			expErr: "NOT_FOUND",
		},
		{
			name:   "unsupported field: Rating",
			ec:     ec,
			fields: []string{"Rating[0]"},
			expErr: "unsupported field prefix: <Rating>",
		},
		{
			name:   "nil Rates",
			ec:     ec,
			fields: []string{"Rate", "rate1"},
			expErr: "NOT_FOUND",
		},
		{
			name:   "unsupported field: Rates",
			ec:     ec,
			fields: []string{"Rates[0]"},
			expErr: "unsupported field prefix: <Rates>",
		},
		{
			name:   "nil Accounts",
			ec:     ec,
			fields: []string{"Account", "acc1"},
			expErr: "NOT_FOUND",
		},
		{
			name:   "unsupported field: Account",
			ec:     ec,
			fields: []string{"Account[0]"},
			expErr: "unsupported field prefix: <Account>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := tt.ec.FieldAsInterface(tt.fields)
			if err == nil || err.Error() != tt.expErr {
				t.Errorf("Expected error: %s, received: %s", tt.expErr, err)
			}

			if !reflect.DeepEqual(rcv, nil) {
				t.Errorf("Expected nil, got %v", rcv)
			}
		})
	}
}

func TestTruncateSimpleAbstracts(t *testing.T) {
	eC := &EventCharges{
		Abstracts: NewDecimal(300000, 0),
		Charges: []*ChargeEntry{
			{
				ChargingID:     "97aa08e",
				CompressFactor: 1,
			},
			{
				ChargingID:     "43e77a7",
				CompressFactor: 1,
			},
			{
				ChargingID:     "97aa08e",
				CompressFactor: 3,
			},
			{
				ChargingID:     "f894244",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*AccountCharge{
			"43e77a7": {
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			"97aa08e": {
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(50000, 0),
				UnitFactorID: "UF2",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			"f894244": {
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(60000, 0),
				UnitFactorID: "UF3",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"UF1": {
				Factor: NewDecimal(100, 0),
			},
			"UF2": {
				Factor: NewDecimal(100, 0),
			},
			"UF3": {
				Factor: NewDecimal(1, 9),
			},
		},
		Rating: map[string]*RateSInterval{
			"877a74e": {
				Increments: []*RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "3365d99",
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"3365d99": {
				RecurrentFee: NewDecimal(0, 0),
			},
		},
		Accounts: map[string]*Account{
			"2343000000000123": {
				Tenant:    CGRateSorg,
				ID:        "2343000000000123",
				FilterIDs: []string{"*string:~*req.IMSI:2343000000000123"},
				Balances: map[string]*Balance{
					"DATA1": {
						ID: "DATA1",
						Weights: []*DynamicWeight{
							{
								Weight: 5,
							},
						},
						Type:  MetaAbstract,
						Units: NewDecimal(700*1000, 0),
						CostIncrements: []*CostIncrement{
							{
								Increment: NewDecimal(1, 0),
							},
						},
					},
				},
			},
		},
	}
	// Truncate all
	testEC := eC.Clone()
	if rstEc, err := testEC.Truncate(NewDecimal(0, 0)); err != nil {
		t.Error(err)
	} else if len(testEC.Charges) != 0 {
		t.Errorf("Initial eC: %+v\n", testEC)
	} else if len(rstEc.Charges) != 4 {
		t.Errorf("Rest eC: %+v\n", rstEc)
	}
	// Truncate nothing
	testEC = eC.Clone()
	if rstEc, err := testEC.Truncate(NewDecimal(300*1000, 0)); err != nil {
		t.Error(err)
	} else if len(testEC.Charges) != 4 {
		t.Errorf("Initial eC: %+v\n", testEC)
	} else if rstEc != nil {
		t.Errorf("Rest eC: %+v\n", rstEc)
	}
	// Truncate first unit
	testEC = eC.Clone()
	atIdx := NewDecimal(1, 0)
	if rstEc, err := testEC.Truncate(atIdx); err != nil {
		t.Error(err)
	} else if testEC.Abstracts.Compare(atIdx) != 0 ||
		len(testEC.Charges) != 1 ||
		testEC.Charges[0].CompressFactor != 1 ||
		testEC.Accounting[testEC.Charges[0].ChargingID].Units.Compare(atIdx) != 0 {
		t.Errorf("Initial eC: %+v, atIndex: %s\n", testEC, atIdx)
	} else if rstEc.Abstracts.Compare(SubstractDecimal(eC.Abstracts, atIdx)) != 0 ||
		len(rstEc.Charges) != 4 ||
		rstEc.Charges[0].CompressFactor != 1 ||
		rstEc.Accounting[rstEc.Charges[0].ChargingID].Units.Compare(SubstractDecimal(eC.Accounting[eC.Charges[0].ChargingID].Units, atIdx)) != 0 {
		t.Errorf("Rest eC: %+v, atIndex: %s\n", rstEc, atIdx)
	}
	// Truncate first compress factor
	testEC = eC.Clone()
	atIdx = NewDecimal(50*1000, 0)
	if rstEc, err := testEC.Truncate(atIdx); err != nil {
		t.Error(err)
	} else if testEC.Abstracts.Compare(atIdx) != 0 ||
		len(testEC.Charges) != 1 ||
		testEC.Charges[0].CompressFactor != 1 ||
		testEC.Accounting[testEC.Charges[0].ChargingID].Units.Compare(atIdx) != 0 {
		t.Errorf("Initial eC: %+v, atIndex: %s\n", testEC, atIdx)
	} else if rstEc.Abstracts.Compare(SubstractDecimal(eC.Abstracts, atIdx)) != 0 ||
		len(rstEc.Charges) != 3 ||
		rstEc.Charges[0].CompressFactor != 1 ||
		rstEc.Accounting[rstEc.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[1].ChargingID].Units) != 0 {
		t.Errorf("Rest eC: %+v, atIndex: %s\n", rstEc, atIdx)
	}
	// Truncate second charge
	testEC = eC.Clone()
	atIdx = NewDecimal(89*1000, 0)
	if rstEc, err := testEC.Truncate(atIdx); err != nil {
		t.Error(err)
	} else if testEC.Abstracts.Compare(atIdx) != 0 ||
		len(testEC.Charges) != 2 ||
		testEC.Charges[0].CompressFactor != 1 ||
		testEC.Accounting[testEC.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[0].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[1].ChargingID].Units.Compare(NewDecimal(39*1000, 0)) != 0 {
		t.Errorf("Initial eC: %+v, atIndex: %s\n", testEC, atIdx)
	} else if rstEc.Abstracts.Compare(SubstractDecimal(eC.Abstracts, atIdx)) != 0 ||
		len(rstEc.Charges) != 3 ||
		rstEc.Charges[0].CompressFactor != 1 ||
		rstEc.Accounting[rstEc.Charges[0].ChargingID].Units.Compare(NewDecimal(1000, 0)) != 0 ||
		rstEc.Accounting[rstEc.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 ||
		rstEc.Accounting[rstEc.Charges[2].ChargingID].Units.Compare(eC.Accounting[eC.Charges[3].ChargingID].Units) != 0 {
		t.Errorf("Rest eC: %+v, atIndex: %s\n", rstEc, atIdx)
	}
	// Truncate second charge
	testEC = eC.Clone()
	atIdx = NewDecimal(90*1000, 0)
	if rstEc, err := testEC.Truncate(atIdx); err != nil {
		t.Error(err)
	} else if testEC.Abstracts.Compare(atIdx) != 0 ||
		len(testEC.Charges) != 2 ||
		testEC.Charges[0].CompressFactor != 1 ||
		testEC.Accounting[testEC.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[0].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[1].ChargingID].Units) != 0 {
		t.Errorf("Initial eC: %+v, atIndex: %s\n", testEC, atIdx)
	} else if rstEc.Abstracts.Compare(SubstractDecimal(eC.Abstracts, atIdx)) != 0 ||
		len(rstEc.Charges) != 2 ||
		rstEc.Charges[0].CompressFactor != 3 ||
		rstEc.Accounting[rstEc.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 ||
		rstEc.Accounting[rstEc.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[3].ChargingID].Units) != 0 {
		t.Errorf("Rest eC: %+v, atIndex: %s\n", rstEc, atIdx)
	}
	// Truncate part of third charge with compress factor
	testEC = eC.Clone()
	atIdx = NewDecimal(91*1000, 0)
	if rstEc, err := testEC.Truncate(atIdx); err != nil {
		t.Error(err)
	} else if testEC.Abstracts.Compare(atIdx) != 0 ||
		len(testEC.Charges) != 3 ||
		testEC.Charges[2].CompressFactor != 1 ||
		testEC.Accounting[testEC.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[0].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[1].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[2].ChargingID].Units.Compare(NewDecimal(1000, 0)) != 0 {
		t.Errorf("Initial eC: %+v, atIndex: %s\n", testEC, atIdx)
	} else if rstEc.Abstracts.Compare(SubstractDecimal(eC.Abstracts, atIdx)) != 0 ||
		len(rstEc.Charges) != 3 ||
		rstEc.Charges[0].CompressFactor != 1 ||
		rstEc.Charges[1].CompressFactor != 2 ||
		rstEc.Accounting[rstEc.Charges[0].ChargingID].Units.Compare(NewDecimal(49*1000, 0)) != 0 ||
		rstEc.Accounting[rstEc.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 ||
		rstEc.Accounting[rstEc.Charges[2].ChargingID].Units.Compare(eC.Accounting[eC.Charges[3].ChargingID].Units) != 0 {
		t.Errorf("Rest eC: %+v, atIndex: %s\n", rstEc, atIdx)
	}
	// Truncate full of third charge with compress factor
	testEC = eC.Clone()
	atIdx = NewDecimal(140*1000, 0)
	if rstEc, err := testEC.Truncate(atIdx); err != nil {
		t.Error(err)
	} else if testEC.Abstracts.Compare(atIdx) != 0 ||
		len(testEC.Charges) != 3 ||
		testEC.Charges[2].CompressFactor != 1 ||
		testEC.Accounting[testEC.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[0].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[1].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[2].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 {
		t.Errorf("Initial eC: %+v, atIndex: %s\n", testEC, atIdx)
	} else if rstEc.Abstracts.Compare(SubstractDecimal(eC.Abstracts, atIdx)) != 0 ||
		len(rstEc.Charges) != 2 ||
		rstEc.Charges[0].CompressFactor != 2 ||
		rstEc.Accounting[rstEc.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 ||
		rstEc.Accounting[rstEc.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[3].ChargingID].Units) != 0 {
		t.Errorf("Rest eC: %+v, atIndex: %s\n", rstEc, atIdx)
	}
	// Truncate full of third charge with compress factor
	testEC = eC.Clone()
	atIdx = NewDecimal(141*1000, 0)
	if rstEc, err := testEC.Truncate(atIdx); err != nil {
		t.Error(err)
	} else if testEC.Abstracts.Compare(atIdx) != 0 ||
		len(testEC.Charges) != 4 ||
		testEC.Charges[0].CompressFactor != 1 ||
		testEC.Charges[1].CompressFactor != 1 ||
		testEC.Accounting[testEC.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[0].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[1].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[2].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 {
		t.Errorf("Initial eC: %+v, atIndex: %s\n", testEC, atIdx)
	} else if rstEc.Abstracts.Compare(SubstractDecimal(eC.Abstracts, atIdx)) != 0 ||
		len(rstEc.Charges) != 3 ||
		rstEc.Charges[1].CompressFactor != 1 ||
		rstEc.Accounting[rstEc.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 ||
		rstEc.Accounting[rstEc.Charges[2].ChargingID].Units.Compare(eC.Accounting[eC.Charges[3].ChargingID].Units) != 0 {
		t.Errorf("Rest eC: %+v, atIndex: %s\n", rstEc, atIdx)
	}
	// Truncate second compress factor of third charge
	testEC = eC.Clone()
	atIdx = NewDecimal(190*1000, 0)
	if rstEc, err := testEC.Truncate(atIdx); err != nil {
		t.Error(err)
	} else if testEC.Abstracts.Compare(atIdx) != 0 ||
		len(testEC.Charges) != 3 ||
		testEC.Charges[0].CompressFactor != 1 ||
		testEC.Charges[2].CompressFactor != 2 ||
		testEC.Accounting[testEC.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[0].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[1].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[2].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 {
		t.Errorf("Initial eC: %+v, atIndex: %s\n", testEC, atIdx)
	} else if rstEc.Abstracts.Compare(SubstractDecimal(eC.Abstracts, atIdx)) != 0 ||
		len(rstEc.Charges) != 2 ||
		rstEc.Charges[0].CompressFactor != 1 ||
		rstEc.Charges[1].CompressFactor != 1 ||
		rstEc.Accounting[rstEc.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 ||
		rstEc.Accounting[rstEc.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[3].ChargingID].Units) != 0 {
		t.Errorf("Rest eC: %+v, atIndex: %s\n", rstEc, atIdx)
	}
	// Truncate second compress factor of third charge
	testEC = eC.Clone()
	atIdx = NewDecimal(191*1000, 0)
	if rstEc, err := testEC.Truncate(atIdx); err != nil {
		t.Error(err)
	} else if testEC.Abstracts.Compare(atIdx) != 0 ||
		len(testEC.Charges) != 4 ||
		testEC.Charges[0].CompressFactor != 1 ||
		testEC.Charges[1].CompressFactor != 1 ||
		testEC.Charges[2].CompressFactor != 2 ||
		testEC.Charges[3].CompressFactor != 1 ||
		testEC.Accounting[testEC.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[0].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[1].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[2].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[3].ChargingID].Units.Compare(NewDecimal(1000, 0)) != 0 {
		t.Errorf("Initial eC: %+v, atIndex: %s\n", testEC, atIdx)
	} else if rstEc.Abstracts.Compare(SubstractDecimal(eC.Abstracts, atIdx)) != 0 ||
		len(rstEc.Charges) != 2 ||
		rstEc.Charges[0].CompressFactor != 1 ||
		rstEc.Charges[1].CompressFactor != 1 ||
		rstEc.Accounting[rstEc.Charges[0].ChargingID].Units.Compare(NewDecimal(49*1000, 0)) != 0 ||
		rstEc.Accounting[rstEc.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[3].ChargingID].Units) != 0 {
		t.Errorf("Rest eC: %+v, atIndex: %s\n", rstEc, atIdx)
	}
	// Truncate to only last charge in full
	testEC = eC.Clone()
	atIdx = NewDecimal(240*1000, 0)
	if rstEc, err := testEC.Truncate(atIdx); err != nil {
		t.Error(err)
	} else if testEC.Abstracts.Compare(atIdx) != 0 ||
		len(testEC.Charges) != 3 ||
		testEC.Charges[0].CompressFactor != 1 ||
		testEC.Charges[1].CompressFactor != 1 ||
		testEC.Charges[2].CompressFactor != 3 ||
		testEC.Accounting[testEC.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[0].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[1].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[2].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 {
		t.Errorf("Initial eC: %+v, atIndex: %s\n", testEC, atIdx)
	} else if rstEc.Abstracts.Compare(SubstractDecimal(eC.Abstracts, atIdx)) != 0 ||
		len(rstEc.Charges) != 1 ||
		rstEc.Charges[0].CompressFactor != 1 ||
		rstEc.Accounting[rstEc.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[3].ChargingID].Units) != 0 {
		t.Errorf("Rest eC: %+v, atIndex: %s\n", rstEc, atIdx)
	}
	// Truncate inside last charge
	testEC = eC.Clone()
	atIdx = NewDecimal(241*1000, 0)
	if rstEc, err := testEC.Truncate(atIdx); err != nil {
		t.Error(err)
	} else if testEC.Abstracts.Compare(atIdx) != 0 ||
		len(testEC.Charges) != 4 ||
		testEC.Charges[0].CompressFactor != 1 ||
		testEC.Charges[1].CompressFactor != 1 ||
		testEC.Charges[2].CompressFactor != 3 ||
		testEC.Charges[3].CompressFactor != 1 ||
		testEC.Accounting[testEC.Charges[0].ChargingID].Units.Compare(eC.Accounting[eC.Charges[0].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[1].ChargingID].Units.Compare(eC.Accounting[eC.Charges[1].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[2].ChargingID].Units.Compare(eC.Accounting[eC.Charges[2].ChargingID].Units) != 0 ||
		testEC.Accounting[testEC.Charges[3].ChargingID].Units.Compare(NewDecimal(1000, 0)) != 0 {
		t.Errorf("Initial eC: %+v, atIndex: %s\n", testEC, atIdx)
	} else if rstEc.Abstracts.Compare(SubstractDecimal(eC.Abstracts, atIdx)) != 0 ||
		len(rstEc.Charges) != 1 ||
		rstEc.Charges[0].CompressFactor != 1 ||
		rstEc.Accounting[rstEc.Charges[0].ChargingID].Units.Compare(NewDecimal(59000, 0)) != 0 {
		t.Errorf("Rest eC: %+v, atIndex: %s\n", rstEc, atIdx)
	}
}

func TestEventChargesAbstractConcretes(t *testing.T) {
	tests := []struct {
		name         string
		ec           *EventCharges
		expAbstracts *Decimal
		expConcretes *Decimal
	}{
		{
			name: "Balance Type: MetaAbstract",
			ec: &EventCharges{
				Abstracts: NewDecimal(30000, 0),
				Concretes: NewDecimal(30000, 0),
				Charges: []*ChargeEntry{
					{
						ChargingID:     "f894244",
						CompressFactor: 1,
					},
					{
						ChargingID:     "f894245",
						CompressFactor: 4,
					},
				},
				Accounting: map[string]*AccountCharge{
					"f894244": {
						AccountID:    "2343000000000123",
						BalanceID:    "DATA1",
						Units:        NewDecimal(30000, 0),
						BalanceLimit: NewDecimal(0, 0),
						RatingID:     "877a74e",
					},
					"f894245": {
						AccountID:    "2343000000000123",
						BalanceID:    "DATA2",
						Units:        NewDecimal(20000, 0),
						BalanceLimit: NewDecimal(0, 0),
						RatingID:     "877a74e",
					},
				},
				Rating: map[string]*RateSInterval{
					"877a74e": {
						Increments: []*RateSIncrement{
							{
								RateIntervalIndex: 0,
								RateID:            "3365d99",
								CompressFactor:    1,
							},
						},
						CompressFactor: 1,
					},
				},
				Rates: map[string]*IntervalRate{
					"3365d99": {
						RecurrentFee: NewDecimal(0, 0),
					},
				},
				Accounts: map[string]*Account{
					"2343000000000123": {
						Tenant:    CGRateSorg,
						ID:        "2343000000000123",
						FilterIDs: []string{"*string:~*req.IMSI:2343000000000123"},
						Balances: map[string]*Balance{
							"DATA1": {
								ID: "DATA1",
								Weights: []*DynamicWeight{
									{
										Weight: 4,
									},
								},
								Type:  MetaAbstract,
								Units: NewDecimal(300*1000, 0),
								CostIncrements: []*CostIncrement{
									{
										Increment: NewDecimal(1, 0),
									},
								},
							},
							"DATA2": {
								ID: "DATA2",
								Weights: []*DynamicWeight{
									{
										Weight: 4,
									},
								},
								Type:  MetaAbstract,
								Units: NewDecimal(500*1000, 0),
								CostIncrements: []*CostIncrement{
									{
										Increment: NewDecimal(1, 0),
									},
								},
							},
						},
					},
				},
			},
			expAbstracts: NewDecimal(110000, 0),
			expConcretes: nil,
		},
		{
			name: "Balance Type: MetaConcrete",
			ec: &EventCharges{
				Charges: []*ChargeEntry{
					{
						ChargingID:     "97aa08e",
						CompressFactor: 1,
					},
					{
						ChargingID:     "97aa08e",
						CompressFactor: 3,
					},
				},
				Accounting: map[string]*AccountCharge{
					"97aa08e": {
						AccountID:    "2343000000000123",
						BalanceID:    "DATA3",
						Units:        NewDecimal(40000, 0),
						BalanceLimit: NewDecimal(0, 0),
						RatingID:     "877a74e",
					},
				},
				Rating: map[string]*RateSInterval{
					"877a74e": {
						Increments: []*RateSIncrement{
							{
								RateIntervalIndex: 0,
								RateID:            "3365d99",
								CompressFactor:    1,
							},
						},
						CompressFactor: 1,
					},
				},
				Rates: map[string]*IntervalRate{
					"3365d99": {
						RecurrentFee: NewDecimal(0, 0),
					},
				},
				Accounts: map[string]*Account{
					"2343000000000123": {
						Tenant:    CGRateSorg,
						ID:        "2343000000000123",
						FilterIDs: []string{"*string:~*req.IMSI:2343000000000123"},
						Balances: map[string]*Balance{
							"DATA3": {
								ID: "DATA3",
								Weights: []*DynamicWeight{
									{
										Weight: 5,
									},
								},
								Type:  MetaConcrete,
								Units: NewDecimal(50*1000, 0),
								CostIncrements: []*CostIncrement{
									{
										Increment: NewDecimal(1, 0),
									},
								},
							},
						},
					},
				},
			},
			expAbstracts: nil,
			expConcretes: NewDecimal(160000, 0),
		},
		{
			name: "Account with both balance types: MetaConcrete and MetaAbstract",
			ec: &EventCharges{
				Charges: []*ChargeEntry{
					{
						ChargingID:     "43e77a7",
						CompressFactor: 1,
					},
					{
						ChargingID:     "f894244",
						CompressFactor: 1,
					},
				},
				Accounting: map[string]*AccountCharge{
					"43e77a7": {
						AccountID:    "2343000000000123",
						BalanceID:    "DATA1",
						Units:        NewDecimal(50000, 0),
						BalanceLimit: NewDecimal(0, 0),
						RatingID:     "877a74e",
					},
					"f894244": {
						AccountID:    "2343000000000123",
						BalanceID:    "DATA2",
						Units:        NewDecimal(30000, 0),
						BalanceLimit: NewDecimal(0, 0),
						RatingID:     "877a74e",
					},
				},
				Rating: map[string]*RateSInterval{
					"877a74e": {
						Increments: []*RateSIncrement{
							{
								RateIntervalIndex: 0,
								RateID:            "3365d99",
								CompressFactor:    1,
							},
						},
						CompressFactor: 1,
					},
				},
				Rates: map[string]*IntervalRate{
					"3365d99": {
						RecurrentFee: NewDecimal(0, 0),
					},
				},
				Accounts: map[string]*Account{
					"2343000000000123": {
						Tenant:    CGRateSorg,
						ID:        "2343000000000123",
						FilterIDs: []string{"*string:~*req.IMSI:2343000000000123"},
						Balances: map[string]*Balance{
							"DATA1": {
								ID: "DATA1",
								Weights: []*DynamicWeight{
									{
										Weight: 5,
									},
								},
								Type:  MetaConcrete,
								Units: NewDecimal(50*1000, 0),
								CostIncrements: []*CostIncrement{
									{
										Increment: NewDecimal(1, 0),
									},
								},
							},
							"DATA2": {
								ID: "DATA2",
								Weights: []*DynamicWeight{
									{
										Weight: 4,
									},
								},
								Type:  MetaAbstract,
								Units: NewDecimal(300*1000, 0),
								CostIncrements: []*CostIncrement{
									{
										Increment: NewDecimal(1, 0),
									},
								},
							},
						},
					},
				},
			},
			expAbstracts: NewDecimal(30000, 0),
			expConcretes: NewDecimal(50000, 0),
		},
		{
			name: "0 Units",
			ec: &EventCharges{
				Charges: []*ChargeEntry{
					{
						ChargingID:     "43e77a7",
						CompressFactor: 1,
					},
				},
				Accounting: map[string]*AccountCharge{
					"43e77a7": {
						AccountID:    "2343000000000123",
						BalanceID:    "DATA1",
						Units:        NewDecimal(0, 0),
						BalanceLimit: NewDecimal(0, 0),
						RatingID:     "877a74e",
					},
				},
				Rating: map[string]*RateSInterval{
					"877a74e": {
						Increments: []*RateSIncrement{
							{
								RateIntervalIndex: 0,
								RateID:            "3365d99",
								CompressFactor:    1,
							},
						},
						CompressFactor: 1,
					},
				},
				Rates: map[string]*IntervalRate{
					"3365d99": {
						RecurrentFee: NewDecimal(0, 0),
					},
				},
				Accounts: map[string]*Account{
					"2343000000000123": {
						Tenant:    CGRateSorg,
						ID:        "2343000000000123",
						FilterIDs: []string{"*string:~*req.IMSI:2343000000000123"},
						Balances: map[string]*Balance{
							"DATA1": {
								ID: "DATA1",
								Weights: []*DynamicWeight{
									{
										Weight: 5,
									},
								},
								Type:  MetaConcrete,
								Units: NewDecimal(50*1000, 0),
								CostIncrements: []*CostIncrement{
									{
										Increment: NewDecimal(1, 0),
									},
								},
							},
						},
					},
				},
			},
			expAbstracts: nil,
			expConcretes: NewDecimal(0, 0),
		},
		{
			name: "Balance type different from  MetaConcrete or MetaAbstract",
			ec: &EventCharges{
				Charges: []*ChargeEntry{
					{
						ChargingID:     "43e77a7",
						CompressFactor: 1,
					},
				},
				Accounting: map[string]*AccountCharge{
					"43e77a7": {
						AccountID:    "2343000000000123",
						BalanceID:    "DATA1",
						Units:        NewDecimal(1000, 0),
						BalanceLimit: NewDecimal(0, 0),
						RatingID:     "877a74e",
					},
				},
				Rating: map[string]*RateSInterval{
					"877a74e": {
						Increments: []*RateSIncrement{
							{
								RateIntervalIndex: 0,
								RateID:            "3365d99",
								CompressFactor:    1,
							},
						},
						CompressFactor: 1,
					},
				},
				Rates: map[string]*IntervalRate{
					"3365d99": {
						RecurrentFee: NewDecimal(0, 0),
					},
				},
				Accounts: map[string]*Account{
					"2343000000000123": {
						Tenant:    CGRateSorg,
						ID:        "2343000000000123",
						FilterIDs: []string{"*string:~*req.IMSI:2343000000000123"},
						Balances: map[string]*Balance{
							"DATA1": {
								ID: "DATA1",
								Weights: []*DynamicWeight{
									{
										Weight: 5,
									},
								},
								Type:  MetaMonetary,
								Units: NewDecimal(50*1000, 0),
								CostIncrements: []*CostIncrement{
									{
										Increment: NewDecimal(1, 0),
									},
								},
							},
						},
					},
				},
			},
			expAbstracts: nil,
			expConcretes: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ec.abstractConcretes()
			if !reflect.DeepEqual(tt.expAbstracts, tt.ec.Abstracts) {
				t.Errorf("Expected Abstracts: %v, \nrecieved %v", tt.expAbstracts, tt.ec.Abstracts)
			}

			if !reflect.DeepEqual(tt.expConcretes, tt.ec.Concretes) {
				t.Errorf("Expected Concretes: %v, \nrecieved %v", tt.expConcretes, tt.ec.Concretes)
			}
		})
	}
}

func TestEventChargesCleanup(t *testing.T) {
	eC := &EventCharges{
		Abstracts: NewDecimal(300000, 0),
		Charges: []*ChargeEntry{
			{
				ChargingID:     "97aa08e",
				CompressFactor: 1,
			},
			{
				ChargingID:     "43e77a7",
				CompressFactor: 1,
			},
			{
				ChargingID:     "97aa08e",
				CompressFactor: 3,
			},
			{
				ChargingID:     "f894244",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*AccountCharge{
			"43e77a7": {
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			"97aa08e": {
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(50000, 0),
				UnitFactorID: "UF2",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			"f894244": {
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(60000, 0),
				UnitFactorID: "UF3",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"UF1": {
				Factor: NewDecimal(100, 0),
			},
			"UF2": {
				Factor: NewDecimal(100, 0),
			},
			"UF3": {
				Factor: NewDecimal(1, 9),
			},
		},
		Rating: map[string]*RateSInterval{
			"877a74e": {
				Increments: []*RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "3365d99",
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"3365d99": {
				RecurrentFee: NewDecimal(0, 0),
			},
		},
		Accounts: map[string]*Account{
			"2343000000000123": {
				Tenant:    CGRateSorg,
				ID:        "2343000000000123",
				FilterIDs: []string{"*string:~*req.IMSI:2343000000000123"},
				Balances: map[string]*Balance{
					"DATA1": {
						ID: "DATA1",
						Weights: []*DynamicWeight{
							{
								Weight: 5,
							},
						},
						Type:  MetaAbstract,
						Units: NewDecimal(700*1000, 0),
						CostIncrements: []*CostIncrement{
							{
								Increment: NewDecimal(1, 0),
							},
						},
					},
				},
			},
		},
	}

	evntC := eC.Clone()
	eC.Cleanup()
	if !reflect.DeepEqual(evntC, eC) {
		t.Errorf("Expected %v, \nrecieved %v", evntC, eC)
	}

	eC.Charges = []*ChargeEntry{}
	exp := &EventCharges{
		Charges:     []*ChargeEntry{},
		Accounting:  map[string]*AccountCharge{},
		UnitFactors: map[string]*UnitFactor{},
		Rating:      map[string]*RateSInterval{},
		Rates:       map[string]*IntervalRate{},
		Accounts:    map[string]*Account{},
	}
	eC.Cleanup()
	if !reflect.DeepEqual(exp, eC) {
		t.Errorf("Expected %v, \nrecieved %v", exp, eC)
	}
}

func TestGetChargesForPath(t *testing.T) {
	ec := &EventCharges{
		Accounting: map[string]*AccountCharge{
			"accounting1": {
				AccountID:       "acc1",
				BalanceID:       "balance1",
				Units:           NewDecimal(10, 0),
				BalanceLimit:    NewDecimal(0, 0),
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
		Rating: map[string]*RateSInterval{
			"rating1": {
				IntervalStart: NewDecimal(4, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(5, 0),
						RateIntervalIndex: 1,
						RateID:            "rate1",
						CompressFactor:    1,
						Usage:             NewDecimal(6, 0),
					},
				},
				CompressFactor: 3,
			},
		},
	}

	tests := []struct {
		name    string
		fldPath []string
		chr     *ChargeEntry
		exp     any
		expErr  string
	}{
		{
			name:    "nil ChargeEntry",
			fldPath: []string{"ChargingID"},
			chr:     nil,
			exp:     nil,
			expErr:  "NOT_FOUND",
		},
		{
			name:    "empty field",
			fldPath: []string{},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
		},
		{
			name:    "ChargingID",
			fldPath: []string{"ChargingID"},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: "*accounting:accounting1",
		},
		{
			name:    "CompressFactor",
			fldPath: []string{"CompressFactor"},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: 1,
		},
		{
			name:    "ChargingID without separator",
			fldPath: []string{"Charging"},
			chr: &ChargeEntry{
				ChargingID:     "accounting1",
				CompressFactor: 1,
			},
			expErr: "expected ChargingID format '*accounting:*' or '*rating:*', got 'accounting1'",
		},
		{
			name:    "error case: unsupported field",
			fldPath: []string{"accounting1"},
			chr: &ChargeEntry{
				ChargingID:     "accounting1",
				CompressFactor: 1,
			},
			expErr: "unsupported field prefix: <accounting1>",
		},
		{
			name:    "Accounting",
			fldPath: []string{"Charging"},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: ec.Accounting["accounting1"],
		},
		{
			name:    "Accounting.AccountID",
			fldPath: []string{"Charging", "AccountID"},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: "acc1",
		},
		{
			name:    "Accounting.BalanceID",
			fldPath: []string{"Charging", "BalanceID"},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: "balance1",
		},
		{
			name:    "Accounting.Units",
			fldPath: []string{"Charging", "Units"},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: NewDecimal(10, 0),
		},
		{
			name:    "Accounting.BalanceLimit",
			fldPath: []string{"Charging", "BalanceLimit"},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: NewDecimal(0, 0),
		},
		{
			name:    "Accounting.UnitFactorID",
			fldPath: []string{"Charging", "UnitFactorID"},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: "unit_factor1",
		},
		{
			name:    "Accounting.AttributeIDs",
			fldPath: []string{"Charging", "AttributeIDs"},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: []string{"attr1", "attr2"},
		},
		{
			name:    "Accounting.RatingID",
			fldPath: []string{"Charging", "RatingID"},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: "rating2",
		},
		{
			name:    "Accounting.JoinedChargeIDs",
			fldPath: []string{"Charging", "JoinedChargeIDs"},
			chr: &ChargeEntry{
				ChargingID:     "*accounting:accounting1",
				CompressFactor: 1,
			},
			exp: []string{"joined_charge"},
		},
		{
			name:    "Rating",
			fldPath: []string{"Charging"},
			chr: &ChargeEntry{
				ChargingID:     "*rating:rating1",
				CompressFactor: 1,
			},
			exp: ec.Rating["rating1"],
		},
		{
			name:    "Rating.IntervalStart",
			fldPath: []string{"Charging", "IntervalStart"},
			chr: &ChargeEntry{
				ChargingID:     "*rating:rating1",
				CompressFactor: 1,
			},
			exp: NewDecimal(4, 0),
		},
		{
			name:    "Rating.Increments",
			fldPath: []string{"Charging", "Increments[0]"},
			chr: &ChargeEntry{
				ChargingID:     "*rating:rating1",
				CompressFactor: 1,
			},
			exp: &RateSIncrement{
				IncrementStart:    NewDecimal(5, 0),
				RateIntervalIndex: 1,
				RateID:            "rate1",
				CompressFactor:    1,
				Usage:             NewDecimal(6, 0),
			},
		},
		{
			name:    "Rating.Increments.IncrementStart",
			fldPath: []string{"Charging", "Increments[0]", "IncrementStart"},
			chr: &ChargeEntry{
				ChargingID:     "*rating:rating1",
				CompressFactor: 1,
			},
			exp: NewDecimal(5, 0),
		},
		{
			name:    "Rating.Increments.RateIntervalIndex",
			fldPath: []string{"Charging", "Increments[0]", "RateIntervalIndex"},
			chr: &ChargeEntry{
				ChargingID:     "*rating:rating1",
				CompressFactor: 1,
			},
			exp: 1,
		},
		{
			name:    "Rating.Increments.RateID",
			fldPath: []string{"Charging", "Increments[0]", "RateID"},
			chr: &ChargeEntry{
				ChargingID:     "*rating:rating1",
				CompressFactor: 1,
			},
			exp: "rate1",
		},
		{
			name:    "Rating.Increments.CompressFactor",
			fldPath: []string{"Charging", "Increments[0]", "CompressFactor"},
			chr: &ChargeEntry{
				ChargingID:     "*rating:rating1",
				CompressFactor: 1,
			},
			exp: int64(1),
		},
		{
			name:    "Rating.Increments.Usage",
			fldPath: []string{"Charging", "Increments[0]", "Usage"},
			chr: &ChargeEntry{
				ChargingID:     "*rating:rating1",
				CompressFactor: 1,
			},
			exp: NewDecimal(6, 0),
		},
		{
			name:    "Nil Increments",
			fldPath: []string{"Charging", "Increments[1]"},
			chr: &ChargeEntry{
				ChargingID:     "*rating:rating1",
				CompressFactor: 1,
			},
			exp:    nil,
			expErr: "NOT_FOUND",
		},
		{
			name:    "Rating.CompressFactor",
			fldPath: []string{"Charging", "CompressFactor"},
			chr: &ChargeEntry{
				ChargingID:     "*rating:rating1",
				CompressFactor: 1,
			},
			exp: int64(3),
		},
		{
			name:    "unsupported charging type",
			fldPath: []string{"Charging"},
			chr: &ChargeEntry{
				ChargingID:     "*unsupported:id1",
				CompressFactor: 1,
			},
			expErr: "unsupported field prefix: <Charging>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := ec.getChargesForPath(tt.fldPath, tt.chr)
			if err != nil && err.Error() != tt.expErr {
				t.Errorf("Expected <%v>, received <%v>", tt.expErr, err)
			}

			if !reflect.DeepEqual(tt.exp, val) {
				t.Errorf("Expected: <%#v>, \nreceived: <%#v>", tt.exp, val)
			}
		})
	}
}

func TestEventChargesGetAccountingForPath(t *testing.T) {
	ec := &EventCharges{
		Accounting: map[string]*AccountCharge{
			"joined_charge": {
				AccountID:       "2343000000000456",
				BalanceID:       "DATA2",
				Units:           &Decimal{decimal.New(10, 0)},
				BalanceLimit:    &Decimal{decimal.New(0, 0)},
				UnitFactorID:    "UF2",
				AttributeIDs:    []string{"attr3", "attr4"},
				RatingID:        "rating3",
				JoinedChargeIDs: []string{},
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"UF1": {
				Factor: NewDecimal(100, 0),
			},
			"UF2": {
				Factor: NewDecimal(100, 0),
			},
		},
		Rating: map[string]*RateSInterval{
			"877a74e": {
				Increments: []*RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "3365d99",
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*IntervalRate{
			"3365d99": {
				RecurrentFee: NewDecimal(0, 0),
			},
		},
		Accounts: map[string]*Account{
			"2343000000000123": {
				Tenant:    CGRateSorg,
				ID:        "2343000000000123",
				FilterIDs: []string{"*string:~*req.IMSI:2343000000000123"},
				Balances: map[string]*Balance{
					"DATA1": {
						ID: "DATA1",
						Weights: []*DynamicWeight{
							{
								Weight: 5,
							},
						},
						Type:  MetaAbstract,
						Units: NewDecimal(700*1000, 0),
						CostIncrements: []*CostIncrement{
							{
								Increment: NewDecimal(1, 0),
							},
						},
					},
				},
			},
			"2343000000000456": {
				Tenant:    CGRateSorg,
				ID:        "2343000000000456",
				FilterIDs: []string{"*string:~*req.IMSI:2343000000000456"},
				Balances: map[string]*Balance{
					"DATA2": {
						ID: "DATA2",
						Weights: []*DynamicWeight{
							{
								Weight: 5,
							},
						},
						Type:  MetaAbstract,
						Units: NewDecimal(700*1000, 0),
						CostIncrements: []*CostIncrement{
							{
								Increment: NewDecimal(1, 0),
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		fldPath   []string
		accCharge *AccountCharge
		want      any
		wantErr   string
	}{
		{
			name:    "Account",
			fldPath: []string{"Account"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: &Account{
				Tenant:    CGRateSorg,
				ID:        "2343000000000123",
				FilterIDs: []string{"*string:~*req.IMSI:2343000000000123"},
				Balances: map[string]*Balance{
					"DATA1": {
						ID: "DATA1",
						Weights: []*DynamicWeight{
							{
								Weight: 5,
							},
						},
						Type:  MetaAbstract,
						Units: NewDecimal(700*1000, 0),
						CostIncrements: []*CostIncrement{
							{
								Increment: NewDecimal(1, 0),
							},
						},
					},
				},
			},
		},
		{
			name:    "Account.Balances.DATA1",
			fldPath: []string{"Account", "Balances", "DATA1"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: &Balance{
				ID: "DATA1",
				Weights: []*DynamicWeight{
					{
						Weight: 5,
					},
				},
				Type:  MetaAbstract,
				Units: NewDecimal(700*1000, 0),
				CostIncrements: []*CostIncrement{
					{
						Increment: NewDecimal(1, 0),
					},
				},
			},
		},
		{
			name:    "AccountID",
			fldPath: []string{"AccountID"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: "2343000000000123",
		},
		{
			name:    "BalanceID",
			fldPath: []string{"BalanceID"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: "DATA1",
		},
		{
			name:    "Units",
			fldPath: []string{"Units"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: NewDecimal(40000, 0),
		},
		{
			name:    "UnitFactorID",
			fldPath: []string{"UnitFactorID"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: "UF1",
		},
		{
			name:    "BalanceLimit",
			fldPath: []string{"BalanceLimit"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: NewDecimal(0, 0),
		},
		{
			name:    "RatingID",
			fldPath: []string{"RatingID"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: "877a74e",
		},
		{
			name:    "UnitFactor",
			fldPath: []string{"UnitFactor"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: &UnitFactor{
				Factor: NewDecimal(100, 0),
			},
		},
		{
			name:    "UnitFactor.Factor",
			fldPath: []string{"UnitFactor", "Factor"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: NewDecimal(100, 0),
		},
		{
			name:    "Nil UnitFactorID",
			fldPath: []string{"UnitFactor"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want:    nil,
			wantErr: "NOT_FOUND",
		},
		{
			name:    "JoinedChargeID not found",
			fldPath: []string{"JoinedChargeIDs[1]"},
			accCharge: &AccountCharge{
				AccountID:       "2343000000000123",
				BalanceID:       "DATA1",
				UnitFactorID:    "UF1",
				Units:           NewDecimal(40000, 0),
				BalanceLimit:    NewDecimal(0, 0),
				RatingID:        "877a74e",
				JoinedChargeIDs: []string{"joined_charge"},
			},
			want:    nil,
			wantErr: "NOT_FOUND",
		},
		{
			name:    "JoinedCharge[0]",
			fldPath: []string{"JoinedCharge[0]"},
			accCharge: &AccountCharge{
				AccountID:       "2343000000000123",
				BalanceID:       "DATA1",
				UnitFactorID:    "UF1",
				Units:           NewDecimal(40000, 0),
				BalanceLimit:    NewDecimal(0, 0),
				RatingID:        "877a74e",
				JoinedChargeIDs: []string{"joined_charge"},
			},
			want: &AccountCharge{
				AccountID:       "2343000000000456",
				BalanceID:       "DATA2",
				Units:           &Decimal{decimal.New(10, 0)},
				BalanceLimit:    &Decimal{decimal.New(0, 0)},
				UnitFactorID:    "UF2",
				AttributeIDs:    []string{"attr3", "attr4"},
				RatingID:        "rating3",
				JoinedChargeIDs: []string{},
			},
		},
		{
			name:    "Rating",
			fldPath: []string{"Rating"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: &RateSInterval{
				Increments: []*RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "3365d99",
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		{
			name:    "Balance",
			fldPath: []string{"Balance"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: &Balance{
				ID: "DATA1",
				Weights: []*DynamicWeight{
					{
						Weight: 5,
					},
				},
				Type:  MetaAbstract,
				Units: NewDecimal(700*1000, 0),
				CostIncrements: []*CostIncrement{
					{
						Increment: NewDecimal(1, 0),
					},
				},
			},
		},
		{
			name:    "Balance.ID",
			fldPath: []string{"Balance", "ID"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want: "DATA1",
		},
		{
			name:      "Nil accCharge",
			fldPath:   []string{"RatingID"},
			accCharge: nil,
			want:      nil,
			wantErr:   "NOT_FOUND",
		},
		{
			name:    "Nil AccountID",
			fldPath: []string{"Account"},
			accCharge: &AccountCharge{
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want:    nil,
			wantErr: "NOT_FOUND",
		},
		{
			name:    "Nil BalanceID",
			fldPath: []string{"Balance"},
			accCharge: &AccountCharge{
				AccountID:    "2343000000000123",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want:    nil,
			wantErr: "NOT_FOUND",
		},
		{
			name:    "Nil AccountID for Balance",
			fldPath: []string{"Balance"},
			accCharge: &AccountCharge{
				BalanceID:    "DATA1",
				Units:        NewDecimal(40000, 0),
				UnitFactorID: "UF1",
				BalanceLimit: NewDecimal(0, 0),
				RatingID:     "877a74e",
			},
			want:    nil,
			wantErr: "NOT_FOUND",
		},
		{
			name:    "error case: invalid index",
			fldPath: []string{"JoinedCharge"},
			accCharge: &AccountCharge{
				AccountID:       "2343000000000123",
				BalanceID:       "DATA1",
				UnitFactorID:    "UF1",
				Units:           NewDecimal(40000, 0),
				BalanceLimit:    NewDecimal(0, 0),
				RatingID:        "877a74e",
				JoinedChargeIDs: []string{},
			},
			want:    nil,
			wantErr: "invalid index for 'JoinedCharge' field",
		},
		{
			name:    "error case: unsupported field",
			fldPath: []string{"Accounting"},
			accCharge: &AccountCharge{
				AccountID:       "2343000000000456",
				BalanceID:       "DATA2",
				UnitFactorID:    "UF2",
				Units:           NewDecimal(40000, 0),
				BalanceLimit:    NewDecimal(0, 0),
				RatingID:        "877a74e",
				JoinedChargeIDs: []string{"joined_charge"},
			},
			want:    nil,
			wantErr: "unsupported field prefix: <Accounting>",
		},
		{
			name:    "Empty JoinedChargeIDs",
			fldPath: []string{"JoinedCharge[0]", "AccountID"},
			accCharge: &AccountCharge{
				AccountID:       "2343000000000123",
				BalanceID:       "DATA1",
				UnitFactorID:    "UF1",
				Units:           NewDecimal(40000, 0),
				BalanceLimit:    NewDecimal(0, 0),
				RatingID:        "877a74e",
				JoinedChargeIDs: []string{},
			},
			want:    nil,
			wantErr: "NOT_FOUND",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := ec.getAccountingForPath(tt.fldPath, tt.accCharge)
			if err != nil && err.Error() != tt.wantErr {
				t.Errorf("Expected <%v>, received <%v>", tt.wantErr, err)
			}
			if !reflect.DeepEqual(rcv, tt.want) {
				t.Errorf("Expected <%+v>, received <%+v>", tt.want, rcv)
			}
		})
	}
}

func TestEventChargesGetRatingForPath(t *testing.T) {
	ec := &EventCharges{
		Accounting: map[string]*AccountCharge{
			"accounting1": {
				AccountID:       "acc1",
				BalanceID:       "balance1",
				Units:           NewDecimal(10, 0),
				BalanceLimit:    NewDecimal(0, 0),
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
			"UF1": {
				Factor: NewDecimal(100, 0),
			},
			"UF2": {
				Factor: NewDecimal(100, 0),
			},
		},
		Rates: map[string]*IntervalRate{
			"3365d99": {
				RecurrentFee: NewDecimal(0, 0),
			},
			"3365d88": nil,
		},
		Accounts: map[string]*Account{
			"2343000000000123": {
				Tenant:    CGRateSorg,
				ID:        "2343000000000123",
				FilterIDs: []string{"*string:~*req.IMSI:2343000000000123"},
				Balances: map[string]*Balance{
					"DATA1": {
						ID: "DATA1",
						Weights: []*DynamicWeight{
							{
								Weight: 5,
							},
						},
						Type:  MetaAbstract,
						Units: NewDecimal(700*1000, 0),
						CostIncrements: []*CostIncrement{
							{
								Increment: NewDecimal(1, 0),
							},
						},
					},
				},
			},
			"2343000000000456": {
				Tenant:    CGRateSorg,
				ID:        "2343000000000456",
				FilterIDs: []string{"*string:~*req.IMSI:2343000000000456"},
				Balances: map[string]*Balance{
					"DATA2": {
						ID: "DATA2",
						Weights: []*DynamicWeight{
							{
								Weight: 5,
							},
						},
						Type:  MetaAbstract,
						Units: NewDecimal(700*1000, 0),
						CostIncrements: []*CostIncrement{
							{
								Increment: NewDecimal(1, 0),
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name       string
		fldPath    []string
		rtInterval *RateSInterval
		want       any
		wantErr    string
	}{
		{
			name:       "Nil RateSInterval",
			fldPath:    []string{"RateSInterval"},
			rtInterval: nil,
			want:       nil,
			wantErr:    "NOT_FOUND",
		},
		{
			name:    "fldPath is empty",
			fldPath: []string{},
			rtInterval: &RateSInterval{
				IntervalStart: NewDecimal(0, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(0, 0),
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateID:            "3365d99",
						RateIntervalIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
			want: &RateSInterval{
				IntervalStart: NewDecimal(0, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(0, 0),
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateID:            "3365d99",
						RateIntervalIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		{
			name:    "Increments",
			fldPath: []string{"Increments"},
			rtInterval: &RateSInterval{
				IntervalStart: NewDecimal(0, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(0, 0),
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateID:            "3365d99",
						RateIntervalIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
			want: []*RateSIncrement{
				{
					IncrementStart:    NewDecimal(0, 0),
					Usage:             NewDecimal(int64(time.Minute), 0),
					RateID:            "3365d99",
					RateIntervalIndex: 0,
					CompressFactor:    1,
				},
			},
		},
		{
			name:    "Increments[0]",
			fldPath: []string{"Increments[0]"},
			rtInterval: &RateSInterval{
				IntervalStart: NewDecimal(0, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(0, 0),
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateID:            "3365d99",
						RateIntervalIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
			want: &RateSIncrement{
				IncrementStart:    NewDecimal(0, 0),
				Usage:             NewDecimal(int64(time.Minute), 0),
				RateID:            "3365d99",
				RateIntervalIndex: 0,
				CompressFactor:    1,
			},
		},
		{
			name:    "Increments[1]",
			fldPath: []string{"Increments[1]"},
			rtInterval: &RateSInterval{
				IntervalStart: NewDecimal(0, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(0, 0),
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateID:            "3365d99",
						RateIntervalIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
			want:    nil,
			wantErr: "NOT_FOUND",
		},
		{
			name:    "Increments[0].RateID",
			fldPath: []string{"Increments[0]", "RateID"},
			rtInterval: &RateSInterval{
				IntervalStart: NewDecimal(0, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(0, 0),
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateID:            "3365d99",
						RateIntervalIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
			want: "3365d99",
		},
		{
			name:    "Increments[0].Rate",
			fldPath: []string{"Increments[0]", "Rate"},
			rtInterval: &RateSInterval{
				IntervalStart: NewDecimal(0, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(0, 0),
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateID:            "3365d99",
						RateIntervalIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
			want: &IntervalRate{
				RecurrentFee: NewDecimal(0, 0),
			},
		},
		{
			name:    "Rate not found",
			fldPath: []string{"Increments[0]", "Rate"},
			rtInterval: &RateSInterval{
				IntervalStart: NewDecimal(0, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(0, 0),
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateID:            "3365d88",
						RateIntervalIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
			want:    nil,
			wantErr: "NOT_FOUND",
		},
		{
			name:    "Increments[0].Rate.RecurrentFee",
			fldPath: []string{"Increments[0]", "Rate", "RecurrentFee"},
			rtInterval: &RateSInterval{
				IntervalStart: NewDecimal(0, 0),
				Increments: []*RateSIncrement{
					{
						IncrementStart:    NewDecimal(0, 0),
						Usage:             NewDecimal(int64(time.Minute), 0),
						RateID:            "3365d99",
						RateIntervalIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
			want: NewDecimal(0, 0),
		},
		{
			name:    "Increments[0].Rate with nil Increment",
			fldPath: []string{"Increments[0]", "Rate"},
			rtInterval: &RateSInterval{
				IntervalStart: NewDecimal(0, 0),
				Increments: []*RateSIncrement{
					nil,
				},
				CompressFactor: 1,
			},
			want:    nil,
			wantErr: "NOT_FOUND",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := ec.getRatingForPath(tt.fldPath, tt.rtInterval)
			if err != nil && err.Error() != tt.wantErr {
				t.Errorf("Expected <%v>, received <%v>", tt.wantErr, err)
			}
			if !reflect.DeepEqual(rcv, tt.want) {
				t.Errorf("Expected <%+v>, received <%+v>", tt.want, rcv)
			}
		})
	}
}
