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
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
)

func TestUnmarshalInvalidSyntax(t *testing.T) {
	x := map[string]*Decimal{
		"*asr": DecimalNaN,
		"*abc": NewDecimal(2, 0),
		"*acd": DecimalNaN,
	}
	var bts []byte
	var err error
	if bts, err = json.Marshal(x); err != nil {

		t.Error(err)
	}

	var reply map[string]*Decimal
	if err = json.Unmarshal(bts, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, x) {
		t.Errorf("Expected %v, received %v", x, reply)
	}
}

func TestConvertDecimalToFloat(t *testing.T) {
	decm := NewDecimal(9999000000000000, 13)
	if conv, ok := decm.Float64(); ok {
		t.Errorf("Can convert decimal: %v to float64: %v", decm, conv)
	}

	decm = NewDecimal(9999, 12)
	if conv, ok := decm.Float64(); !ok {
		t.Errorf("Cannot convert decimal: %v to float64: %v", decm, conv)
	}

	decm = NewDecimal(1000000000000000, 16)
	if conv, ok := decm.Float64(); !ok {
		t.Errorf("Cannot convert decimal: %v to float64: %v", decm, conv)
	}
}

func TestDecimalSum(t *testing.T) {
	dec1 := NewDecimal(495, 1)
	dec2 := NewDecimal(5, 1)
	x := SumBig(dec1.Big, dec2.Big)

	// as decimals, 50 is different from 50.0 (decimal.Big)
	exp := NewDecimal(500, 1)
	if !reflect.DeepEqual(x, exp.Big) {
		t.Errorf("Expected %+v, received %+v", ToJSON(exp.Big), ToJSON(x))
	}

	// same for substract
	diff := SubstractBig(dec1.Big, dec2.Big)
	exp = NewDecimal(490, 1)
	if !reflect.DeepEqual(diff, exp.Big) {
		t.Errorf("Expected %+v, received %+v", ToJSON(exp.Big), ToJSON(x))
	}

	// decimal
	exp2 := NewDecimal(500, 1)
	x2 := SumDecimal(dec1, dec2)
	if !reflect.DeepEqual(x2, exp2) {
		t.Errorf("Expected %+v, received %+v", ToJSON(exp2), ToJSON(x2))
	}

	// same for substract
	exp2 = NewDecimal(490, 1)
	x2 = SubstractDecimal(dec1, dec2)
	if !reflect.DeepEqual(x2, exp2) {
		t.Errorf("Expected %+v, received %+v", ToJSON(exp2), ToJSON(x2))
	}

	// to conclude, 50.0 is different from 50, same difference for 49.0 and 49
	val1, val2 := NewDecimal(50, 0), NewDecimal(500, 1)
	if reflect.DeepEqual(val1, val2) {
		t.Errorf("Expected %+v, received %+v", ToJSON(val1), ToJSON(val2))
	}
	if reflect.DeepEqual(val1.Big, val2.Big) {
		t.Errorf("Expected %+v, received %+v", ToJSON(val1.Big), ToJSON(val2.Big))
	}
}

func TestNewDecimalDivide(t *testing.T) {
	x := decimal.WithContext(DecimalContext).SetUint64(10)
	y := decimal.WithContext(DecimalContext).SetUint64(5)
	expected, _ := decimal.WithContext(DecimalContext).SetUint64(2).Float64()
	received, _ := DivideBig(x, y).Float64()
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}
}

func TestDivideBigNil(t *testing.T) {
	var x, y, expected *decimal.Big
	x = nil
	y = nil
	expected = nil
	received := DivideBig(x, y)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}

	x = decimal.WithContext(DecimalContext).SetUint64(2)
	y = nil
	expected = nil
	received = DivideBig(x, y)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}
}

func TestNewDecimalMultiply(t *testing.T) {
	x := decimal.WithContext(DecimalContext).SetUint64(10)
	y := decimal.WithContext(DecimalContext).SetUint64(5)
	expected, _ := decimal.WithContext(DecimalContext).SetUint64(50).Float64()
	received, _ := MultiplyBig(x, y).Float64()
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}
}

func TestMultiplyBigNil(t *testing.T) {
	var x, y, expected *decimal.Big
	x = nil
	y = nil
	expected = nil
	received := MultiplyBig(x, y)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}

	x = decimal.WithContext(DecimalContext).SetUint64(2)
	y = nil
	expected = nil
	received = MultiplyBig(x, y)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}
}

func TestNewDecimalAdd(t *testing.T) {
	x := decimal.WithContext(DecimalContext).SetUint64(10)
	y := decimal.WithContext(DecimalContext).SetUint64(5)
	expected, _ := decimal.WithContext(DecimalContext).SetUint64(15).Float64()
	received, _ := SumBig(x, y).Float64()
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}
}

func TestSumBigNil(t *testing.T) {
	var x, y, expected *decimal.Big
	x = nil
	y = decimal.WithContext(DecimalContext).SetUint64(2)
	expected = decimal.WithContext(DecimalContext).SetUint64(2)
	received := SumBig(x, y)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}

	x = decimal.WithContext(DecimalContext).SetUint64(3)
	y = nil
	expected = decimal.WithContext(DecimalContext).SetUint64(3)
	received = SumBig(x, y)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", expected, received)
	}
}

func TestSubstractBigNil(t *testing.T) {
	var x, y, expected *decimal.Big
	x = decimal.WithContext(DecimalContext).SetUint64(10)
	y = nil
	expected = decimal.WithContext(DecimalContext).SetUint64(10)
	received := SubstractBig(x, y)
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

func TestDecimalCompareNaN(t *testing.T) {
	dec1 := NewDecimal(2, 0)
	if comp := dec1.Compare(DecimalNaN); comp <= 0 {
		t.Errorf("%v is higher than %v, means comp is: %v", dec1, DecimalNaN, comp)
	}
	if comp := DecimalNaN.Compare(dec1); comp >= 0 {
		t.Errorf("%v is lower than %v, means comp is: %v", DecimalNaN, dec1, comp)
	}

	dec1 = DecimalNaN
	if comp := dec1.Compare(DecimalNaN); comp != 0 {
		t.Errorf("%v is equal to %v, means comp is: %v", dec1, DecimalNaN, comp)
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

	var dec2 Decimal
	if err := dec2.UnmarshalJSON([]byte(`0`)); err != nil {
		t.Error(err)
	}

	var decnil *Decimal
	if err := decnil.UnmarshalJSON([]byte(`0`)); err != nil {
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
	expectedErr = "can't convert <invalid_decimal_format> to decimal"
	if _, err := NewDecimalFromUsage(dec); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	dec = "1000"
	expected = NewDecimal(1000, 0)
	if rcv, err := NewDecimalFromUsage(dec); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	dec = "12m"
	expected = NewDecimal(int64(12*time.Minute), 0)
	if rcv, err := NewDecimalFromUsage(dec); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	dec = "12m10s"
	expected = NewDecimal(int64(12*time.Minute+10*time.Second), 0)
	if rcv, err := NewDecimalFromUsage(dec); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	dec = "12.44"
	expected = NewDecimal(1244, 2)
	if rcv, err := NewDecimalFromUsage(dec); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	dec = "12.44.5"
	expectedErr = "can't convert <12.44.5> to decimal"
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
	x := decimal.WithContext(DecimalContext).SetUint64(10)
	y := decimal.WithContext(DecimalContext).SetUint64(5)
	qExpected := decimal.WithContext(DecimalContext).SetUint64(2)
	rExpected := decimal.WithContext(DecimalContext).SetUint64(0)
	qReceived, rReceived := DivideBigWithReminder(x, y)
	if !reflect.DeepEqual(qExpected, qReceived) {
		t.Errorf("Expected divident <+%v> but received <+%v>", qExpected, qReceived)
	} else if !reflect.DeepEqual(rExpected, rReceived) {
		t.Errorf("Expected divident <+%v> but received <+%v>", rExpected, rReceived)
	}

	x = nil
	qReceived, rReceived = DivideBigWithReminder(x, y)
	qExpected = nil
	rExpected = nil
	if !reflect.DeepEqual(qExpected, qReceived) {
		t.Errorf("Expected divident <+%v> but received <+%v>", qExpected, qReceived)
	} else if !reflect.DeepEqual(rExpected, rReceived) {
		t.Errorf("Expected divident <+%v> but received <+%v>", rExpected, rReceived)
	}

	x = decimal.WithContext(DecimalContext).SetUint64(10)
	y = nil
	qReceived, rReceived = DivideBigWithReminder(x, y)
	qExpected = nil
	rExpected = nil
	if !reflect.DeepEqual(qExpected, qReceived) {
		t.Errorf("Expected divident <+%v> but received <+%v>", qExpected, qReceived)
	} else if !reflect.DeepEqual(rExpected, rReceived) {
		t.Errorf("Expected divident <+%v> but received <+%v>", rExpected, rReceived)
	}
}

func TestDecimalNewDecimalFromUsageEmptyStringCase(t *testing.T) {
	exp := NewDecimal(0, 0)
	if rcv, err := NewDecimalFromUsage(EmptyString); err != nil {
		t.Error(err)
	} else if rcv.Compare(exp) != 0 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestDecimalFloat64(t *testing.T) {
	d := NewDecimal(123, 2)
	exp := 1.23
	if rcv, ok := d.Float64(); !ok {
		t.Error("expected ok to be true")
	} else if rcv != exp {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestDecimalDuration(t *testing.T) {
	d := NewDecimal(int64(3), 0)
	rcv, ok := d.Duration()
	if !ok {
		t.Error("Cannot convert")
	} else if rcv != time.Nanosecond*3 {
		t.Errorf("Expected 3ns")
	}

	d = NewDecimal(999999999999999999, 5555555555555555555)
	rcv, ok = d.Duration()
	if ok {
		t.Errorf("Cannot convert %v", rcv)
	}
}

func TestMarshalUnmarshalNA(t *testing.T) {
	mrsh, err := DecimalNaN.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	var dec2 Decimal
	if err := dec2.UnmarshalJSON(mrsh); err != nil {
		t.Error(err)
	}
	if dec2.Compare(DecimalNaN) != 0 {
		t.Errorf("%v and %v are different", dec2, DecimalNaN)
	}
}

func TestNewRoundingMode(t *testing.T) {
	var tests = []struct {
		rnd string
		exp decimal.RoundingMode
	}{
		{"*toNearestEven", 0},
		{"*toNearestAway", 1},
		{"*toZero", 2},
		{"*awayFromZero", 3},
		{"*toNegativeInf", 4},
		{"*toPositiveInf", 5},
		{"*toNearestTowardZero", 6},
	}

	unsupp := "unsupported"

	for _, e := range tests {
		if rcv, err := NewRoundingMode(e.rnd); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(rcv, e.exp) {
			t.Errorf("expected: <%v>, received: <%v>", e.exp, rcv)
		}

	}
	expErr := "usupoorted rounding: <\"unsupported\">"
	if _, err := NewRoundingMode(unsupp); err == nil || err.Error() != expErr {
		t.Errorf("expected: <%+v>, received: <%+v>", expErr, err)
	}

}

func TestNewRoundingModeToString(t *testing.T) {
	var tests = []struct {
		rnd decimal.RoundingMode
		exp string
	}{
		{0, "*toNearestEven"},
		{1, "*toNearestAway"},
		{2, "*toZero"},
		{3, "*awayFromZero"},
		{4, "*toNegativeInf"},
		{5, "*toPositiveInf"},
		{6, "*toNearestTowardZero"},
		{7, ""},
	}

	for _, e := range tests {
		if rcv := RoundingModeToString(e.rnd); !reflect.DeepEqual(rcv, e.exp) {
			t.Errorf("expected: <%v>, received: <%v>", e.exp, rcv)
		}
	}

}

func TestDivideDecimal(t *testing.T) {
	var x *Decimal = NewDecimal(12, 0)
	var y *Decimal = NewDecimal(2, 0)
	exp := NewDecimal(6, 0)
	if rcv := DivideDecimal(x, y); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, received: <%+v>", exp, rcv)
	}
}

func TestNewDecimalFromStringIgnoreError(t *testing.T) {
	if rcv := NewDecimalFromStringIgnoreError("teststring"); rcv != nil {
		t.Errorf("Expected <nil>, received <%+v>", rcv)
	}
}
