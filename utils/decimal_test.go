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
	"strconv"
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

	dec = nil
	if err := dec.UnmarshalBinary([]byte(`10`)); err != nil {
		t.Error(err)
	}

	dec1 := new(Decimal)
	expected2 := []byte(`0`)
	if rcv, err := dec1.MarshalBinary(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected2, rcv) {
		t.Errorf("Expected %+v, received %+v", string(expected2), string(rcv))
	}
}

func TestUnmarshalJSON(t *testing.T) {
	dec1 := new(Decimal)
	expected := NewDecimal(0, 0)
	if err := dec1.UnmarshalJSON([]byte(`0`)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dec1) {
		t.Errorf("Expected %+v, received %+v", expected, dec1)
	}

	dec1 = nil
	if err := dec1.UnmarshalJSON([]byte(`0`)); err != nil {
		t.Error(err)
	}
}

func TestDecimalCalculus(t *testing.T) {
	d1 := NewDecimal(10, 0)
	d2 := NewDecimal(20, 0)
	if d1.Compare(d2) != -1 {
		t.Errorf("%+v should be lower that %+v", d1, d2)
	}

	if rcv := SubstractBig(d2.Big, d1.Big); !reflect.DeepEqual(rcv, d1.Big) {
		t.Errorf("Expected %+v, received %+v", ToJSON(d1.Big), ToJSON(rcv))
	}

	if rcv := MultiplyDecimal(d1, d2); !reflect.DeepEqual(NewDecimal(200, 0), rcv) {
		t.Errorf("Expected %+v, received %+v", ToJSON(NewDecimal(200, 0)), ToJSON(rcv))
	}

	if rcv := SubstractDecimal(d2, d1); !reflect.DeepEqual(d1, rcv) {
		t.Errorf("Expected %+v, received %+v", ToJSON(d1), ToJSON(rcv))
	}
}

func TestMarshalJSON(t *testing.T) {
	dec := new(Decimal)
	if rcv, err := dec.MarshalJSON(); err != nil {
		t.Error(err)
	} else if len(rcv) != 5 {
		t.Error("Expected empty slice", len(rcv))
	}
}

func TestNewDecimalFromUsage(t *testing.T) {
	dec := "12tts"
	expectedErr := "time: unknown unit \"tts\" in duration \"12tts\""
	if _, err := NewDecimalFromUsage(dec); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	dec = "2"
	expected := NewDecimal(2, 0)
	if rcv, err := NewDecimalFromUsage(dec); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	dec = "invalid_decimal_format"
	expectedErr = "strconv.ParseInt: parsing \"invalid_decimal_format\": invalid syntax"
	if _, err := NewDecimalFromUsage(dec); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestDecimalNewDecimalFromString(t *testing.T) {
	str := "123.4"
	received, err := NewDecimalFromString(str)
	if err != nil {
		t.Error(err)
	}
	expected := &Decimal{decimal.New(1234, 1)}
	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestDecimalNewDecimalFromStringErr(t *testing.T) {
	str := "testString"
	_, err := NewDecimalFromString(str)
	expected := "can't convert <" + str + "> to decimal"

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err)
	}
}

func TestDivideBigWithReminder(t *testing.T) {
	x := new(decimal.Big).SetUint64(10)
	y := new(decimal.Big).SetUint64(5)
	qExpected := new(decimal.Big).SetUint64(2)
	rExpected := new(decimal.Big).SetUint64(0)
	qReceived, rReceived := DivideBigWithReminder(x, y)
	if !reflect.DeepEqual(qExpected, qReceived) {
		t.Errorf("Expected divident <+%v> but received <+%v>", qExpected, qReceived)
	} else if !reflect.DeepEqual(rExpected, rReceived) {
		t.Errorf("Expected divident <+%v> but received <+%v>", rExpected, rReceived)
	}
}

func TestNewDecimalFromFloat64(t *testing.T) {
	x := 21.5
	xExp, _ := new(decimal.Big).SetString(strconv.FormatFloat(x, 'f', -1, 64))
	expected := &Decimal{xExp}
	// fmt.Printf("%v of type %T", expected, expected)
	rcv := NewDecimalFromFloat64(x)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected <+%v> but received <+%v>", xExp, rcv)
	}
}

func TestDecimalClone(t *testing.T) {
	d := &Decimal{new(decimal.Big).SetUint64(3)}
	rcv := d.Clone()
	if !reflect.DeepEqual(d, rcv) {
		t.Errorf("Expected <+%v> but received <+%v>", d, rcv)
	}
}
