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

	"github.com/ericlagergren/decimal"
)

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

func TestAsExtChargingInterval(t *testing.T) {
	chrgInt := &ChargingInterval{
		Increments: []*ChargingIncrement{
			{
				Units:           NewDecimal(123, 3),
				AccountChargeID: "ID1",
				CompressFactor:  40,
			},
			{
				Units:           NewDecimal(15238, 3),
				AccountChargeID: "ID2",
			},
		},
		CompressFactor: 9,
	}
	expChInt := &ExtChargingInterval{
		Increments: []*ExtChargingIncrement{
			{
				Units:           Float64Pointer(0.123),
				AccountChargeID: "ID1",
				CompressFactor:  40,
			},
			{
				Units:           Float64Pointer(15.238),
				AccountChargeID: "ID2",
				CompressFactor:  0,
			},
		},
		CompressFactor: 9,
	}
	if rcv, err := chrgInt.AsExtChargingInterval(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expChInt) {
		t.Errorf("Expected %+v \n, received %+v", ToJSON(expChInt), ToJSON(rcv))
	}

	chrgInt.Increments[0].Units = NewDecimal(int64(math.Inf(1))-1, 0)
	expected := "Cannot convert decimal ChargingIncrement "
	if _, err := chrgInt.AsExtChargingInterval(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
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

func TestECAsExtEventChargesFailAbstracts(t *testing.T) {
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

func TestECAsExtEventChargesFailConcretes(t *testing.T) {
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
