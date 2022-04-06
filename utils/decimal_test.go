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
	fmt.Printf("dec2 is NaN: %v\n", dec2.IsNaN(0))

}
