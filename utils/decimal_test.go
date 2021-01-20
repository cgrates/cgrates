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

func TestNewDecimalDivide(t *testing.T) {
	x := new(decimal.Big).SetUint64(10)
	y := new(decimal.Big).SetUint64(5)
	expected, _ := new(decimal.Big).SetUint64(2).Float64()
	received, _ := DivideBig(x, y).Float64()
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}
}

func TestNewDecimalMultiply(t *testing.T) {
	x := new(decimal.Big).SetUint64(10)
	y := new(decimal.Big).SetUint64(5)
	expected, _ := new(decimal.Big).SetUint64(50).Float64()
	received, _ := MultiplyBig(x, y).Float64()
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}
}

func TestNewDecimalAdd(t *testing.T) {
	x := new(decimal.Big).SetUint64(10)
	y := new(decimal.Big).SetUint64(5)
	expected, _ := new(decimal.Big).SetUint64(15).Float64()
	received, _ := SumBig(x, y).Float64()
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}
}

func TestNewDecimalFromUnit(t *testing.T) {
	if val, err := NewDecimalFromUsage("1ns"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, NewDecimal(1, 0)) {
		t.Errorf("Expected %+v, received %+v", NewDecimal(1, 0), val)
	}
}

func TestUnmarshalMarshalBinary(t *testing.T) {
	dec := &Decimal{}
	expected := NewDecimal(10, 0)
	if err := dec.UnmarshalBinary([]byte(`10`)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.Big, dec.Big) {
		t.Errorf("Expected %T, received %T", expected, dec.Big)
	}

	expected2 := []byte(`10`)
	if rcv, err := dec.MarshalBinary(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected2, rcv) {
		t.Errorf("Expected %+v, received %+v", expected2, rcv)
	}
}
