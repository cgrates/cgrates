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
		t.Errorf("\nReceived: <%+v>, \nExpected: <%+v>", received, expected)
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
		t.Errorf("\nReceived: <%v>, \nExpected: <%v>", received, expected)
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
		t.Errorf("\nReceived: <%v>, \nExpected: <%v>", received, expected)
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
		t.Errorf("\nReceived: <%v>, \nExpected: <%v>", received, expected)
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
		t.Errorf("\nReceived: %v,\nExpected: %v", err, nil)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf(
			"\nReceived: <%v>,\nExpected: <%v>",
			ToJSON(received),
			ToJSON(expected),
		)
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
		t.Errorf("\nReceived: %v,\nExpected: %v", err, expected)
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
		t.Errorf("\nReceived: %v,\nExpected: %v", err, expected)
	}
}
