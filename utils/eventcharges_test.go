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

/*
func TestECNewEventCharges(t *testing.T) {
	expected := &EventCharges{
		Accounting:  make(map[string]*AccountCharge),
		UnitFactors: make(map[string]*UnitFactor),
		Rating:      make(map[string]*RateSInterval),
	}
	received := NewEventCharges()

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}
*/

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

func TestECAsExtEventChargesEmpty(t *testing.T) {
	ec := &EventCharges{
		Abstracts: nil,
		Concretes: nil,
	}

	expected := &ExtEventCharges{
		Abstracts: nil,
		Concretes: nil,
	}
	received, err := ec.AsExtEventCharges()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}

}

func TestECAsExtEventChargesSuccess(t *testing.T) {
	ec := &EventCharges{
		Abstracts: &Decimal{
			decimal.New(1234, 3),
		},
		Concretes: &Decimal{
			decimal.New(4321, 5),
		},
	}

	expected := &ExtEventCharges{
		Abstracts: Float64Pointer(1.234),
		Concretes: Float64Pointer(0.04321),
	}
	received, err := ec.AsExtEventCharges()
	if err != nil {
		t.Errorf("\nExpected: %v,\nReceived: %v", nil, err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf(
			"\nExpected: <%v>,\nReceived: <%v>",
			ToJSON(expected),
			ToJSON(received),
		)
	}
}

func TestAsExtAccountCharge(t *testing.T) {
	ac := &AccountCharge{
		AccountID:       "ACCID_1",
		BalanceID:       "BALID_1",
		Units:           NewDecimal(123, 4),
		BalanceLimit:    NewDecimal(10, 1),
		UnitFactorID:    "seven",
		AttributeIDs:    []string{"TEST_ID1", "TEST_ID2"},
		RatingID:        "RTID_1",
		JoinedChargeIDs: []string{"TEST_ID2", "TEST_ID2"},
	}
	expAcc := &ExtAccountCharge{
		AccountID:       "ACCID_1",
		BalanceID:       "BALID_1",
		Units:           Float64Pointer(0.0123),
		BalanceLimit:    Float64Pointer(1.0),
		UnitFactorID:    "seven",
		AttributeIDs:    []string{"TEST_ID1", "TEST_ID2"},
		RatingID:        "RTID_1",
		JoinedChargeIDs: []string{"TEST_ID2", "TEST_ID2"},
	}
	if rcv, err := ac.AsExtAccountCharge(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expAcc) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expAcc), ToJSON(rcv))
	}

	ac = &AccountCharge{
		AccountID:       "ACCID_1",
		BalanceID:       "BALID_1",
		Units:           NewDecimal(123, 4),
		UnitFactorID:    "seven",
		JoinedChargeIDs: []string{},
	}
	expAcc = &ExtAccountCharge{
		AccountID:       "ACCID_1",
		BalanceID:       "BALID_1",
		Units:           Float64Pointer(0.0123),
		UnitFactorID:    "seven",
		JoinedChargeIDs: []string{},
	}
	if rcv, err := ac.AsExtAccountCharge(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expAcc) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expAcc), ToJSON(rcv))
	}

	ac.BalanceLimit = NewDecimal(int64(math.Inf(1))-1, 0)
	expErr := "cannot convert decimal BalanceLimit to float64 "
	if _, err := ac.AsExtAccountCharge(); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+v, received %+v", expErr, err)
	}

	ac.Units = NewDecimal(int64(math.Inf(1))-1, 0)
	expErr = "cannot convert decimal Units to float64 "
	if _, err := ac.AsExtAccountCharge(); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+v, received %+v", expErr, err)
	}
}

func TestECAsExtEventChargesErrConvertAbstracts(t *testing.T) {
	v, _ := new(decimal.Big).SetString("900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993")

	ec := &EventCharges{
		Abstracts: &Decimal{v},
		Concretes: &Decimal{decimal.New(1234, 3)},
	}

	expected := "cannot convert decimal Abstracts to float64"
	_, err := ec.AsExtEventCharges()

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: %v,\nReceived: %v", expected, err)
	}
}

func TestECAsExtEventChargesErrConvertConcretes(t *testing.T) {
	v, _ := new(decimal.Big).SetString("900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993900719925474099390071992547409939007199254740993")

	ec := &EventCharges{
		Abstracts: &Decimal{decimal.New(1234, 3)},
		Concretes: &Decimal{v},
	}

	expected := "cannot convert decimal Concretes to float64"
	_, err := ec.AsExtEventCharges()

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: %v,\nReceived: %v", expected, err)
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
/*
func TestCompressEqualsChargingInterval(t *testing.T) {
	chIn1 := &ChargingInterval{
		Increments: []*ChargingIncrement{
			{
				Units:           NewDecimal(10, 0),
				AccountChargeID: "CHARGER1",
				CompressFactor:  1,
			},
		},
		CompressFactor: 2,
	}
	chIn2 := &ChargingInterval{
		Increments: []*ChargingIncrement{
			{
				Units:           NewDecimal(10, 0),
				AccountChargeID: "CHARGER1",
				CompressFactor:  1,
			},
		},
		CompressFactor: 4,
	}

	// compressEquals is not looking for compress factor
	if !chIn1.CompressEquals(chIn2) {
		t.Errorf("Intervals %+v and %+v are not equal", ToJSON(chIn1), ToJSON(chIn2))
	}

	//same thing in ChargingIncrements
	chIn1.Increments[0].CompressFactor = 2
	if !chIn1.CompressEquals(chIn2) {
		t.Errorf("Intervals %+v and %+v are not equal", ToJSON(chIn1), ToJSON(chIn2))
	}

	//not equals for AccountChargeID
	chIn1.Increments[0].AccountChargeID = "Changed_Charger"
	if chIn1.CompressEquals(chIn2) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(chIn1), ToJSON(chIn2))
	}
	chIn1.Increments[0].AccountChargeID = "CHARGER1"

	chIn2.Increments[0].AccountChargeID = "Changed_Charger"
	if chIn1.CompressEquals(chIn2) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(chIn1), ToJSON(chIn2))
	}
	chIn2.Increments[0].AccountChargeID = "CHARGER1"

	//not equals for Units
	chIn1.Increments[0].Units = NewDecimal(30, 0)
	if chIn1.CompressEquals(chIn2) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(chIn1), ToJSON(chIn2))
	}
	chIn1.Increments[0].Units = NewDecimal(10, 0)

	chIn2.Increments[0].Units = NewDecimal(30, 0)
	if chIn1.CompressEquals(chIn2) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(chIn1), ToJSON(chIn2))
	}
	chIn2.Increments[0].Units = NewDecimal(10, 0)

	//not equals by the length of increments
	chIn1 = &ChargingInterval{
		Increments: []*ChargingIncrement{
			{
				Units:           NewDecimal(10, 0),
				AccountChargeID: "CHARGER1",
				CompressFactor:  1,
			},
			{
				Units:           NewDecimal(12, 0),
				AccountChargeID: "CHARGER2",
				CompressFactor:  6,
			},
		},
		CompressFactor: 0,
	}
	if chIn1.CompressEquals(chIn2) {
		t.Errorf("Intervals %+v and %+v are equal", ToJSON(chIn1), ToJSON(chIn2))
	}
}
*/
/*
func TestAsExtEventCharges(t *testing.T) {
	evCh := &EventCharges{
		ChargingIntervals: []*ChargingInterval{
			{
				Increments: []*ChargingIncrement{
					{
						Units: NewDecimal(1, 0),
					},
				},
				CompressFactor: 2,
			},
		},
		Accounts: []*Account{
			{
				Balances: map[string]*Balance{
					"BL1": {
						Units: NewDecimal(300, 0),
					},
				},
			},
		},
		Accounting: map[string]*AccountCharge{
			"first_accounting": {
				BalanceLimit: NewDecimal(2, 0),
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"first_factor": {
				Factor: NewDecimal(10, 0),
			},
		},
		Rating: map[string]*RateSInterval{
			"first_rates_interval": {
				IntervalStart: NewDecimal(int64(time.Minute), 0),
			},
		},
	}

	expEvCh := &ExtEventCharges{
		ChargingIntervals: []*ExtChargingInterval{
			{
				Increments: []*ExtChargingIncrement{
					{
						Units: Float64Pointer(1.0),
					},
				},
				CompressFactor: 2,
			},
		},
		Accounts: []*ExtAccount{
			{
				Balances: map[string]*ExtBalance{
					"BL1": {
						Units: Float64Pointer(300),
					},
				},
			},
		},
		Accounting: map[string]*ExtAccountCharge{
			"first_accounting": {
				BalanceLimit: Float64Pointer(2),
			},
		},
		UnitFactors: map[string]*ExtUnitFactor{
			"first_factor": {
				Factor: Float64Pointer(10),
			},
		},
		Rating: map[string]*ExtRateSInterval{
			"first_rates_interval": {
				IntervalStart: Float64Pointer(float64(time.Minute)),
			},
		},
	}
	if rcv, err := evCh.AsExtEventCharges(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expEvCh) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expEvCh), ToJSON(rcv))
	}
}

func TestAsExtEventChargersCheckErrors(t *testing.T) {
	evCh := &EventCharges{
		ChargingIntervals: []*ChargingInterval{
			{
				Increments: []*ChargingIncrement{
					{
						Units: NewDecimal(int64(math.Inf(1))-1, 0),
					},
				},
				CompressFactor: 2,
			},
		},
		Accounts: []*Account{
			{
				Balances: map[string]*Balance{
					"BL1": {
						Units: NewDecimal(300, 0),
					},
				},
			},
		},
		Accounting: map[string]*AccountCharge{
			"first_accounting": {
				BalanceLimit: NewDecimal(2, 0),
			},
		},
		UnitFactors: map[string]*UnitFactor{
			"first_factor": {
				Factor: NewDecimal(10, 0),
			},
		},
		Rating: map[string]*RateSInterval{
			"first_rates_interval": {
				IntervalStart: NewDecimal(int64(time.Minute), 0),
			},
		},
	}
	expected := "Cannot convert decimal ChargingIncrement into float64 "
	if _, err := evCh.AsExtEventCharges(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	evCh.ChargingIntervals[0].Increments[0].Units = NewDecimal(0, 0)

	evCh.Accounts[0].Balances["BL1"].Units = NewDecimal(int64(math.Inf(1))-1, 0)
	expected = "cannot convert decimal Units to float64 "
	if _, err := evCh.AsExtEventCharges(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	evCh.Accounts[0].Balances["BL1"].Units = NewDecimal(0, 0)

	evCh.Accounting["first_accounting"].BalanceLimit = NewDecimal(int64(math.Inf(1))-1, 0)
	expected = "cannot convert decimal BalanceLimit to float64 "
	if _, err := evCh.AsExtEventCharges(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	evCh.Accounting["first_accounting"].BalanceLimit = NewDecimal(0, 0)

	evCh.UnitFactors["first_factor"].Factor = NewDecimal(int64(math.Inf(1))-1, 0)
	expected = "cannot convert decimal Factor to float64 "
	if _, err := evCh.AsExtEventCharges(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	evCh.UnitFactors["first_factor"].Factor = NewDecimal(0, 0)

	evCh.Rating["first_rates_interval"].IntervalStart = NewDecimal(int64(math.Inf(1))-1, 0)
	expected = "Cannot convert decimal IntervalStart into float64 "
	if _, err := evCh.AsExtEventCharges(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, err)
	}
	evCh.Rating["first_rates_interval"].IntervalStart = NewDecimal(0, 0)
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
			"GENUUID_GHOST1": {
				AccountID:    "TestEventChargesEquals",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(8, 1),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR1",
			},
			"GENUUID3": {
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
						Rate: &Rate{
							ID:              "*costIncrement",
							FilterIDs:       []string{"*string:~*req.Account:1003"},
							ActivationTimes: "* * * * *",
							Blocker:         true,
							IntervalRates: []*IntervalRate{
								{
									IntervalStart: NewDecimal(0, 0),
									Increment:     NewDecimal(int64(time.Second), 0),
									FixedFee:      NewDecimal(0, 0),
									RecurrentFee:  NewDecimal(11, 1),
								},
								{
									IntervalStart: NewDecimal(int64(time.Minute), 0),
									Increment:     NewDecimal(int64(2*time.Second), 0),
									FixedFee:      NewDecimal(1, 0),
									RecurrentFee:  NewDecimal(5, 1),
									Unit:          NewDecimal(8, 1),
								},
							},
						},
						Usage:             NewDecimal(int64(time.Minute), 0),
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
			"GENUUID_RATING2": {
				Increments: []*RateSIncrement{
					{
						IncrementStart: NewDecimal(4, 2),
						Rate: &Rate{
							ID:        "*costIncrement",
							FilterIDs: []string{},
							Weights: []*DynamicWeight{
								{
									FilterIDs: []string{"*string:~*req.Account:1002"},
									Weight:    20,
								},
								{
									Weight: 15,
								},
							},
							IntervalRates: []*IntervalRate{
								{
									FixedFee:     NewDecimal(5, 1),
									RecurrentFee: NewDecimal(2, 1),
								},
							},
						},
						Usage:             NewDecimal(int64(30*time.Second), 0),
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(0, 0),
				CompressFactor: 2,
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
					Subsys: MetaSessionS,
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
						Opts: map[string]interface{}{
							Destinations: "1234",
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
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, 10, 10, 10, 0, 0, 0, time.UTC),
				},
				Opts: map[string]interface{}{
					Subsys: MetaSessionS,
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
			"GENUUID_GHOST1": {
				AccountID:    "TestEventChargesEquals",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(8, 1),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR1",
			},
			"GENUUID3": {
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
						Rate: &Rate{
							ID:              "*costIncrement",
							FilterIDs:       []string{"*string:~*req.Account:1003"},
							ActivationTimes: "* * * * *",
							Blocker:         true,
							IntervalRates: []*IntervalRate{
								{
									IntervalStart: NewDecimal(0, 0),
									Increment:     NewDecimal(int64(time.Second), 0),
									FixedFee:      NewDecimal(0, 0),
									RecurrentFee:  NewDecimal(11, 1),
								},
								{
									IntervalStart: NewDecimal(int64(time.Minute), 0),
									Increment:     NewDecimal(int64(2*time.Second), 0),
									FixedFee:      NewDecimal(1, 0),
									RecurrentFee:  NewDecimal(5, 1),
									Unit:          NewDecimal(8, 1),
								},
							},
						},
						Usage:             NewDecimal(int64(time.Minute), 0),
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
			"GENUUID_RATING2": {
				Increments: []*RateSIncrement{
					{
						IncrementStart: NewDecimal(4, 2),
						Rate: &Rate{
							ID:        "*costIncrement",
							FilterIDs: []string{},
							Weights: []*DynamicWeight{
								{
									FilterIDs: []string{"*string:~*req.Account:1002"},
									Weight:    20,
								},
								{
									Weight: 15,
								},
							},
							IntervalRates: []*IntervalRate{
								{
									FixedFee:     NewDecimal(5, 1),
									RecurrentFee: NewDecimal(2, 1),
								},
							},
						},
						Usage:             NewDecimal(int64(30*time.Second), 0),
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(0, 0),
				CompressFactor: 2,
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
					Subsys: MetaSessionS,
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
						Opts: map[string]interface{}{
							Destinations: "1234",
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
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, 10, 10, 10, 0, 0, 0, time.UTC),
				},
				Opts: map[string]interface{}{
					Subsys: MetaSessionS,
				},
				ThresholdIDs: []string{},
			},
		},
	}
	if ok := eEvChgs.Equals(expectedEqual); !ok {
		t.Errorf("Expected %+v, received %+v", eEvChgs, expectedEqual)
	}

}

func TestEqualsExtEventCharges(t *testing.T) {
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
			"GENUUID_GHOST1": {
				AccountID:    "TestEventChargesEquals",
				BalanceID:    "CONCRETE1",
				Units:        NewDecimal(8, 1),
				BalanceLimit: NewDecimal(200, 0),
				UnitFactorID: "GENUUID_FACTOR1",
			},
			"GENUUID3": {
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
						Rate: &Rate{
							ID:              "*costIncrement",
							FilterIDs:       []string{"*string:~*req.Account:1003"},
							ActivationTimes: "* * * * *",
							Blocker:         true,
							IntervalRates: []*IntervalRate{
								{
									IntervalStart: NewDecimal(0, 0),
									Increment:     NewDecimal(int64(time.Second), 0),
									FixedFee:      NewDecimal(0, 0),
									RecurrentFee:  NewDecimal(11, 1),
								},
								{
									IntervalStart: NewDecimal(int64(time.Minute), 0),
									Increment:     NewDecimal(int64(2*time.Second), 0),
									FixedFee:      NewDecimal(1, 0),
									RecurrentFee:  NewDecimal(5, 1),
									Unit:          NewDecimal(8, 1),
								},
							},
						},
						Usage:             NewDecimal(int64(time.Minute), 0),
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
			"GENUUID_RATING2": {
				Increments: []*RateSIncrement{
					{
						IncrementStart: NewDecimal(4, 2),
						Rate: &Rate{
							ID:        "*costIncrement",
							FilterIDs: []string{},
							Weights: []*DynamicWeight{
								{
									FilterIDs: []string{"*string:~*req.Account:1002"},
									Weight:    20,
								},
								{
									Weight: 15,
								},
							},
							IntervalRates: []*IntervalRate{
								{
									FixedFee:     NewDecimal(5, 1),
									RecurrentFee: NewDecimal(2, 1),
								},
							},
						},
						Usage:             NewDecimal(int64(30*time.Second), 0),
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				IntervalStart:  NewDecimal(0, 0),
				CompressFactor: 2,
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
					Subsys: MetaSessionS,
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
						Opts: map[string]interface{}{
							Destinations: "1234",
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
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, 10, 10, 10, 0, 0, 0, time.UTC),
				},
				Opts: map[string]interface{}{
					Subsys: MetaSessionS,
				},
				ThresholdIDs: []string{},
			},
		},
	}

	// ext equals
	extEvCh := &ExtEventCharges{
		Abstracts: Float64Pointer(47.5),
		Concretes: Float64Pointer(5.15),
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
		Accounting: map[string]*ExtAccountCharge{
			"GENUUID_GHOST1": {
				AccountID:    "TestEventChargesEquals",
				BalanceID:    "CONCRETE1",
				Units:        Float64Pointer(0.8),
				BalanceLimit: Float64Pointer(200),
				UnitFactorID: "GENUUID_FACTOR1",
			},
			"GENUUID3": {
				AccountID:       "TestEventChargesEquals",
				BalanceID:       "ABSTRACT2",
				BalanceLimit:    Float64Pointer(0),
				RatingID:        "GENUUID_RATING1",
				JoinedChargeIDs: []string{"THIS_GENUUID1"},
			},
			"GENUUID2": {
				AccountID:    "TestEventChargesEquals",
				BalanceID:    "CONCRETE1",
				Units:        Float64Pointer(2),
				BalanceLimit: Float64Pointer(200),
				UnitFactorID: "GENUUID_FACTOR2",
				RatingID:     "ID_FOR_RATING",
				AttributeIDs: []string{"ATTR1", "ATTR2"},
			},
		},
		UnitFactors: map[string]*ExtUnitFactor{
			"GENUUID_FACTOR1": {
				Factor:    Float64Pointer(100),
				FilterIDs: []string{"*string:~*req.Account:1003"},
			},
			"GENUUID_FACTOR2": {
				Factor: Float64Pointer(200),
			},
		},
		Rating: map[string]*ExtRateSInterval{
			"GENUUID_RATING1": {
				Increments: []*ExtRateSIncrement{
					{
						Rate: &ExtRate{
							ID:              "*costIncrement",
							FilterIDs:       []string{"*string:~*req.Account:1003"},
							ActivationTimes: "* * * * *",
							Blocker:         true,
							IntervalRates: []*ExtIntervalRate{
								{
									IntervalStart: Float64Pointer(0),
									Increment:     Float64Pointer(float64(time.Second)),
									FixedFee:      Float64Pointer(0),
									RecurrentFee:  Float64Pointer(1.1),
								},
								{
									IntervalStart: Float64Pointer(float64(time.Minute)),
									Increment:     Float64Pointer(float64(2 * time.Second)),
									FixedFee:      Float64Pointer(1),
									RecurrentFee:  Float64Pointer(0.5),
									Unit:          Float64Pointer(0.8),
								},
							},
						},
						Usage:             Float64Pointer(float64(time.Minute)),
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				IntervalStart:  Float64Pointer(float64(time.Second)),
				CompressFactor: 1,
			},
			"GENUUID_RATING2": {
				Increments: []*ExtRateSIncrement{
					{
						IncrementStart: Float64Pointer(0.04),
						Rate: &ExtRate{
							ID:        "*costIncrement",
							FilterIDs: []string{},
							Weights: []*DynamicWeight{
								{
									FilterIDs: []string{"*string:~*req.Account:1002"},
									Weight:    20,
								},
								{
									Weight: 15,
								},
							},
							IntervalRates: []*ExtIntervalRate{
								{
									FixedFee:     Float64Pointer(0.5),
									RecurrentFee: Float64Pointer(0.2),
								},
							},
						},
						Usage:             Float64Pointer(float64(30 * time.Second)),
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				IntervalStart:  Float64Pointer(0),
				CompressFactor: 2,
			},
		},
		Accounts: map[string]*ExtAccount{
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
					Subsys: MetaSessionS,
				},
				Balances: map[string]*ExtBalance{
					"bal1": {
						ID:        "BAL1",
						FilterIDs: []string{"*string:~*req.Account:1003"},
						Weights: []*DynamicWeight{
							{
								Weight: 25,
							},
						},
						Type:  MetaAbstract,
						Units: Float64Pointer(float64(30 * time.Second)),
						UnitFactors: []*ExtUnitFactor{
							{
								Factor:    Float64Pointer(100),
								FilterIDs: []string{"*string:~*req.Account:1003"},
							},
							{
								Factor: Float64Pointer(200),
							},
						},
						CostIncrements: []*ExtCostIncrement{
							{
								Increment:    Float64Pointer(float64(time.Second)),
								RecurrentFee: Float64Pointer(5),
							},
							{
								FilterIDs:    []string{"*string:~*req.Account:1003"},
								Increment:    Float64Pointer(float64(2 * time.Second)),
								FixedFee:     Float64Pointer(1),
								RecurrentFee: Float64Pointer(5),
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
						Units: Float64Pointer(2000),
						UnitFactors: []*ExtUnitFactor{
							{
								Factor: Float64Pointer(200),
							},
						},
						Opts: map[string]interface{}{
							Destinations: "1234",
						},
						CostIncrements: []*ExtCostIncrement{
							{
								FilterIDs:    []string{"*string:~*req.Account:1004"},
								Increment:    Float64Pointer(float64(2 * time.Second)),
								FixedFee:     Float64Pointer(1),
								RecurrentFee: Float64Pointer(5),
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
				ActivationInterval: &ActivationInterval{
					ActivationTime: time.Date(2020, 10, 10, 10, 0, 0, 0, time.UTC),
				},
				Opts: map[string]interface{}{
					Subsys: MetaSessionS,
				},
				ThresholdIDs: []string{},
			},
		},
	}
	rcv, err := eEvChgs.AsExtEventCharges()
	if err != nil {
		t.Error(err)
	}
	if ok := rcv.Equals(extEvCh); ok {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(extEvCh), ToJSON(rcv))
	}
}
