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

func TestNewDecimalFromFloat64(t *testing.T) {
	expected := &Decimal{new(decimal.Big).SetFloat64(1.25)}
	received := NewDecimalFromFloat64(1.25)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestNewDecimal(t *testing.T) {
	expected := &Decimal{new(decimal.Big)}
	received := NewDecimal()
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestDecimalFloat64(t *testing.T) {
	expected := 3.2795784983858396
	received := NewDecimalFromFloat64(3.2795784983858396).Float64()
	if expected != received {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestDecimalFloat64Negative(t *testing.T) {
	expected := -3.2795784983858396
	received := NewDecimalFromFloat64(-3.2795784983858396).Float64()
	if expected != received {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestDecimalMarshalUnmarshalJSON(t *testing.T) {
	a := NewDecimal()
	received, err := NewDecimalFromFloat64(3.27).MarshalJSON()
	if err != nil {
		t.Errorf("Expecting: nil, received: %+v", received)
	}
	if err := a.UnmarshalJSON(received); err != nil {
		t.Error(err)
	}
	rcv := a.Float64()
	expected := 3.27
	if expected != rcv {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, rcv)
	}
}

func TestNewDecimalFromUint64(t *testing.T) {
	expected := &Decimal{new(decimal.Big).SetUint64(18446744073709551615)}
	received := NewDecimalFromUint64(18446744073709551615)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}
}

func TestNewDecimalDivide(t *testing.T) {
	a := NewDecimal()
	x := NewDecimalFromUint64(10)
	y := NewDecimalFromUint64(5)
	expected := NewDecimalFromUint64(2)
	received := a.Divide(x, y)
	if !reflect.DeepEqual(expected.Float64(), received.Float64()) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected.Float64(), received.Float64())
	}
}

func TestNewDecimalMultiply(t *testing.T) {
	a := NewDecimal()
	x := NewDecimalFromUint64(10)
	y := NewDecimalFromUint64(5)
	expected := NewDecimalFromUint64(50)
	received := a.Multiply(x, y)
	if !reflect.DeepEqual(expected.Float64(), received.Float64()) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected.Float64(), received.Float64())
	}
}

func TestNewDecimalAdd(t *testing.T) {
	a := NewDecimal()
	x := NewDecimalFromUint64(10)
	y := NewDecimalFromUint64(5)
	expected := NewDecimalFromUint64(15)
	received := a.Add(x, y)
	if !reflect.DeepEqual(expected.Float64(), received.Float64()) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected.Float64(), received.Float64())
	}
}

func TestNewDecimalCompareEqual(t *testing.T) {
	a := NewDecimalFromUint64(5)
	y := NewDecimalFromUint64(5)
	received := a.Compare(y)
	if !reflect.DeepEqual(0, received) {
		t.Errorf("Expecting: <0>, received: <%+v>", received)
	}
}

func TestNewDecimalCompareGreaterThan(t *testing.T) {
	a := NewDecimalFromUint64(5)
	y := NewDecimalFromUint64(4)
	received := a.Compare(y)
	if !reflect.DeepEqual(1, received) {
		t.Errorf("Expecting: <1>, received: <%+v>", received)
	}
}

func TestNewDecimalCompareSmallerThan(t *testing.T) {
	a := NewDecimalFromUint64(4)
	y := NewDecimalFromUint64(5)
	received := a.Compare(y)
	if !reflect.DeepEqual(-1, received) {
		t.Errorf("Expecting: <-1>, received: <%+v>", received)
	}
}
