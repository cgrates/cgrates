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
	"errors"
	"math"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nyaruka/phonenumbers"
)

func TestDataConvertersConvertString(t *testing.T) {
	dcs := &DataConverters{}
	if rcv, err := dcs.ConvertString(EmptyString); err != nil {
		t.Error(err)
	} else if rcv != EmptyString {
		t.Errorf("Expecting: <%+q>, received: <%+q>", EmptyString, rcv)
	}
	if rcv, err := dcs.ConvertString("test"); err != nil {
		t.Error(err)
	} else if rcv != "test" {
		t.Errorf("Expecting: <test>, received: <%+q>", rcv)
	}
}

func TestConvertDurationFormat1(t *testing.T) {
	dcs := &DataConverters{
		&DurationFormatConverter{
			Layout: "15:04:05",
		},
	}
	if rcv, err := dcs.ConvertString("15m"); err != nil {
		t.Error(err)
	} else if rcv != "00:15:00" {
		t.Errorf("Expecting: <%+q>, received: <%+q>", "00:15:00", rcv)
	}
}
func TestConvertDurationFormat2(t *testing.T) {
	dcs := &DataConverters{
		&DurationFormatConverter{
			Layout: "15-04-05.999999999",
		},
	}
	if rcv, err := dcs.ConvertString("20s423ns"); err != nil {
		t.Error(err)
	} else if rcv != "00-00-20.000000423" {
		t.Errorf("Expecting: <%+q>, received: <%+q>", "00-00-20.000000423", rcv)
	}
}

func TestConvertDurationFormatDefault(t *testing.T) {
	dcs := &DataConverters{
		&DurationFormatConverter{},
	}
	if rcv, err := dcs.ConvertString("15m"); err != nil {
		t.Error(err)
	} else if rcv != "00:15:00" {
		t.Errorf("Expecting: <%+q>, received: <%+q>", "00:15:00", rcv)
	}
}

func TestNewDataConverter(t *testing.T) {
	a, err := NewDataConverter(MetaDurationSeconds)
	if err != nil {
		t.Error(err)
	}
	b, err := NewDurationSecondsConverter(EmptyString)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("Error reflect")
	}
	a, err = NewDataConverter(MetaDuration)
	if err != nil {
		t.Error(err)
	}
	b, err = NewDurationConverter(EmptyString)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("Error reflect")
	}
	a, err = NewDataConverter(MetaDurationNanoseconds)
	if err != nil {
		t.Error(err)
	}
	b, err = NewDurationNanosecondsConverter(EmptyString)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("Error reflect")
	}
	a, err = NewDataConverter(MetaRound)
	if err != nil {
		t.Error(err)
	}
	b, err = NewRoundConverter(EmptyString)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("Error reflect")
	}
	a, err = NewDataConverter("*round:07")
	if err != nil {
		t.Error(err)
	}
	b, err = NewRoundConverter("7")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("Error reflect")
	}
	if _, err = NewDataConverter(MetaMultiply); err == nil || err != ErrMandatoryIeMissingNoCaps {
		t.Error(err)
	}
	a, err = NewDataConverter("*multiply:3.3")
	if err != nil {
		t.Error(err)
	}
	b, err = NewMultiplyConverter("3.3")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("Error reflect")
	}
	if _, err = NewDataConverter(MetaDivide); err == nil || err != ErrMandatoryIeMissingNoCaps {
		t.Error(err)
	}
	a, err = NewDataConverter("*divide:3.3")
	if err != nil {
		t.Error(err)
	}
	b, err = NewDivideConverter("3.3")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("Error reflect")
	}
	if _, err = NewDataConverter(MetaLibPhoneNumber); err == nil || err.Error() != "unsupported *libphonenumber converter parameters: <>" {
		t.Error(err)
	}
	a, err = NewDataConverter("*libphonenumber:US")
	if err != nil {
		t.Error(err)
	}
	b, err = NewPhoneNumberConverter("US")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("Error reflect")
	}
	if _, err = NewDataConverter("unsupported"); err == nil || err.Error() != "unsupported converter definition: <unsupported>" {
		t.Error(err)
	}

	hex, err := NewDataConverter(MetaString2Hex)
	if err != nil {
		t.Error(err)
	}
	exp := new(String2HexConverter)
	if !reflect.DeepEqual(hex, exp) {
		t.Errorf("Expected %+v received: %+v", exp, hex)
	}

	tm, err := NewDataConverter(MetaTimeString)
	if err != nil {
		t.Error(err)
	}
	expTime := NewTimeStringConverter(time.RFC3339)
	if !reflect.DeepEqual(tm, expTime) {
		t.Errorf("Expected %+v received: %+v", expTime, tm)
	}

	tm, err = NewDataConverter("*time_string:020106150400")
	if err != nil {
		t.Error(err)
	}
	expTime = NewTimeStringConverter("020106150400")
	if !reflect.DeepEqual(tm, expTime) {
		t.Errorf("Expected %+v received: %+v", expTime, tm)
	}
	expected := &DurationFormatConverter{Layout: "15:04:05"}
	if durFmt, err := NewDataConverter(MetaDurationFormat + ":15:04:05"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(durFmt, expected) {
		t.Errorf("Expected %+v received: %+v", expected, durFmt)
	}

}

func TestNewDataConverterMustCompile(t *testing.T) {
	eOut, _ := NewDataConverter(MetaDurationSeconds)
	if rcv := NewDataConverterMustCompile(MetaDurationSeconds); rcv != eOut {
		t.Errorf("Expecting:  received: %+q", rcv)
	}
}

func TestNewDurationSecondsConverter(t *testing.T) {
	eOut := DurationSecondsConverter{}
	if rcv, err := NewDurationSecondsConverter("test"); err != nil {
		t.Error(err)
	} else if reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
}

func TestDurationSecondsConverterConvert(t *testing.T) {
	mS := &DurationSecondsConverter{}
	if _, err := mS.Convert("string"); err.Error() != "time: invalid duration \"string\"" {
		t.Error(err)
	}
}

func TestDurationNanosecondsConverterConvert(t *testing.T) {
	nS := &DurationNanosecondsConverter{}
	if _, err := nS.Convert("string"); err.Error() != "time: invalid duration \"string\"" {
		t.Error(err)
	}
}
func TestNewRoundConverter(t *testing.T) {
	if _, err := NewRoundConverter("test"); err == nil || err.Error() != "*round converter needs integer as decimals, have: <test>" {
		t.Error(err)
	}
	if _, err := NewRoundConverter(":test"); err == nil || err.Error() != "*round converter needs integer as decimals, have: <>" {
		t.Error(err)
	}
	if _, err := NewRoundConverter("test:"); err == nil || err.Error() != "*round converter needs integer as decimals, have: <test>" {
		t.Error(err)
	}
	eOut := &RoundConverter{
		Method: MetaRoundingMiddle,
	}
	if rcv, err := NewRoundConverter(EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
	eOut = &RoundConverter{
		Decimals: 12,
		Method:   MetaRoundingUp,
	}
	if rcv, err := NewRoundConverter("12:*up"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
	eOut = &RoundConverter{
		Decimals: 12,
		Method:   MetaRoundingDown,
	}
	if rcv, err := NewRoundConverter("12:*down"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
	eOut = &RoundConverter{
		Decimals: 12,
		Method:   MetaRoundingMiddle,
	}
	if rcv, err := NewRoundConverter("12:*middle"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
	if rcv, err := NewRoundConverter("12:*middle:wrong_length"); err == nil || err.Error() != "unsupported *round converter parameters: <12:*middle:wrong_length>" {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, nil) {
		t.Errorf("Expected %+v received: %+v", nil, rcv)
	}

}

func TestRoundConverterConvert(t *testing.T) {
	rnd := &RoundConverter{}
	if rcv, err := rnd.Convert("string_test"); err == nil || err.Error() != `strconv.ParseFloat: parsing "string_test": invalid syntax` {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, nil) {
		t.Errorf("Expected %+v received: %+v", nil, rcv)
	}
	eOut := Round(18, rnd.Decimals, rnd.Method)
	if rcv, err := rnd.Convert(18); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
}

func TestNewMultiplyConverter(t *testing.T) {
	if rcv, err := NewMultiplyConverter(EmptyString); err == nil || err != ErrMandatoryIeMissingNoCaps {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, nil) {
		t.Errorf("Expected %+v received: %+v", nil, rcv)
	}
	if rcv, err := NewMultiplyConverter("string_test"); err == nil || err.Error() != `strconv.ParseFloat: parsing "string_test": invalid syntax` {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, nil) {
		t.Errorf("Expected %+v received: %+v", nil, rcv)
	}
	eOut := &MultiplyConverter{Value: 0.7}
	if rcv, err := NewMultiplyConverter("0.7"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
}

func TestMultiplyConverterConvert(t *testing.T) {
	m := &MultiplyConverter{}
	if rcv, err := m.Convert(EmptyString); err == nil || err.Error() != `strconv.ParseFloat: parsing "": invalid syntax` {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, nil) {
		t.Errorf("Expected %+v received: %+v", nil, rcv)
	}
	if rcv, err := m.Convert("string_test"); err == nil || err.Error() != `strconv.ParseFloat: parsing "string_test": invalid syntax` {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, nil) {
		t.Errorf("Expected %+v received: %+v", nil, rcv)
	}
	if rcv, err := m.Convert(3); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, 0.0) {
		t.Errorf("Expected %+v received: %+v", 0, rcv)
	}
}

func TestNewDivideConverter(t *testing.T) {
	if rcv, err := NewDivideConverter(EmptyString); err == nil || err != ErrMandatoryIeMissingNoCaps {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, nil) {
		t.Errorf("Expected %+v received: %+v", nil, rcv)
	}
	if rcv, err := NewDivideConverter("string_test"); err == nil || err.Error() != `strconv.ParseFloat: parsing "string_test": invalid syntax` {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, nil) {
		t.Errorf("Expected %+v received: %+v", nil, rcv)
	}
	eOut := &DivideConverter{Value: 0.7}
	if rcv, err := NewDivideConverter("0.7"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
}

func TestDivideConverterConvert(t *testing.T) {
	m := DivideConverter{}
	if rcv, err := m.Convert("string_test"); err == nil || err.Error() != `strconv.ParseFloat: parsing "string_test": invalid syntax` {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, nil) {
		t.Errorf("Expected %+v received: %+v", nil, rcv)
	}
	eOut := math.Inf(1)
	if rcv, err := m.Convert("96"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
	eOut = math.Inf(-1)
	if rcv, err := m.Convert("-96"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
	m = DivideConverter{Value: 0.7}
	eOut = 137.14285714285714
	if rcv, err := m.Convert("96"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
}

func TestNewDurationConverter(t *testing.T) {
	nS := &DurationConverter{}
	eOut := time.Duration(0)
	if rcv, err := nS.Convert(EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
	eOut = 7 * time.Nanosecond
	if rcv, err := nS.Convert(7); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %+v received: %+v", eOut, rcv)
	}
}

func TestConvertFloatToSeconds(t *testing.T) {
	b, err := NewDataConverter(MetaDurationSeconds)
	if err != nil {
		t.Error(err.Error())
	}
	a, err := b.Convert(10*time.Second + 300*time.Millisecond)
	if err != nil {
		t.Error(err.Error())
	}
	expVal := 10.3
	if !reflect.DeepEqual(a, expVal) {
		t.Errorf("Expected %+v received: %+v", expVal, a)
	}
}

func TestConvertDurNanoseconds(t *testing.T) {
	d, err := NewDataConverter(MetaDurationNanoseconds)
	if err != nil {
		t.Error(err.Error())
	}
	expVal := int64(102)
	if i, err := d.Convert(102); err != nil {
		t.Error(err.Error())
	} else if expVal != i {
		t.Errorf("expecting: %d, received: %d", expVal, i)
	}
}

func TestConvertDurMinutes(t *testing.T) {
	d, err := NewDataConverter(MetaDurationMinutes)
	if err != nil {
		t.Error(err.Error())
	}
	expVal := 2.5
	dur := 150 * time.Second
	if i, err := d.Convert(dur); err != nil {
		t.Error(err.Error())
	} else if expVal != i {
		t.Errorf("expecting: %f, received: %f", expVal, i)
	}

	expVal = 3
	if i, err := d.Convert("180s"); err != nil {
		t.Error(err.Error())
	} else if expVal != i {
		t.Errorf("expecting: %f, received: %f", expVal, i)
	}

}

func TestRoundConverterFloat64(t *testing.T) {
	b, err := NewDataConverter("*round:2")
	if err != nil {
		t.Error(err.Error())
	}
	expData := &RoundConverter{
		Decimals: 2,
		Method:   MetaRoundingMiddle,
	}
	if !reflect.DeepEqual(b, expData) {
		t.Errorf("Expected %+v received: %+v", expData, b)
	}
	val, err := b.Convert(2.3456)
	if err != nil {
		t.Error(err.Error())
	}
	expV := 2.35
	if !reflect.DeepEqual(expV, val) {
		t.Errorf("Expected %+v received: %+v", expV, val)
	}
}

//testRoundconv string / float / int / time

func TestRoundConverterString(t *testing.T) {
	b, err := NewDataConverter("*round:2")
	if err != nil {
		t.Error(err.Error())
	}
	expData := &RoundConverter{
		Decimals: 2,
		Method:   MetaRoundingMiddle,
	}
	if !reflect.DeepEqual(b, expData) {
		t.Errorf("Expected %+v received: %+v", expData, b)
	}
	val, err := b.Convert("10.4295")
	if err != nil {
		t.Error(err.Error())
	}
	expV := 10.43
	if !reflect.DeepEqual(expV, val) {
		t.Errorf("Expected %+v received: %+v", expV, val)
	}
}

func TestRoundConverterInt64(t *testing.T) {
	b, err := NewDataConverter("*round:2")
	if err != nil {
		t.Error(err.Error())
	}
	expData := &RoundConverter{
		Decimals: 2,
		Method:   MetaRoundingMiddle,
	}
	if !reflect.DeepEqual(b, expData) {
		t.Errorf("Expected %+v received: %+v", expData, b)
	}
	val, err := b.Convert(int64(10))
	if err != nil {
		t.Error(err.Error())
	}
	expV := 10.0
	if !reflect.DeepEqual(expV, val) {
		t.Errorf("Expected %+v received: %+v", expV, val)
	}
}

func TestRoundConverterTime(t *testing.T) {
	b, err := NewDataConverter("*round:2")
	if err != nil {
		t.Error(err.Error())
	}
	expData := &RoundConverter{
		Decimals: 2,
		Method:   MetaRoundingMiddle,
	}
	if !reflect.DeepEqual(b, expData) {
		t.Errorf("Expected %+v received: %+v", expData, b)
	}
	val, err := b.Convert(123 * time.Nanosecond)
	if err != nil {
		t.Error(err.Error())
	}
	expV := 123.0
	if !reflect.DeepEqual(expV, val) {
		t.Errorf("Expected %+v received: %+v", expV, val)
	}
}

func TestMultiplyConverter(t *testing.T) {
	eMpl := &MultiplyConverter{1024.0}
	m, err := NewDataConverter("*multiply:1024.0")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMpl, m) {
		t.Errorf("expecting: %+v, received: %+v", eMpl, m)
	}
	expOut := 2048.0
	if out, err := m.Convert(2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOut, out) {
		t.Errorf("expecting: %+v, received: %+v", expOut, out)
	}
	expOut = 1536.0
	if out, err := m.Convert(1.5); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOut, out) {
		t.Errorf("expecting: %+v, received: %+v", expOut, out)
	}
}

func TestDivideConverter(t *testing.T) {
	eDvd := &DivideConverter{1024.0}
	d, err := NewDataConverter("*divide:1024.0")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDvd, d) {
		t.Errorf("expecting: %+v, received: %+v", eDvd, d)
	}
	expOut := 2.0
	if out, err := d.Convert(2048); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOut, out) {
		t.Errorf("expecting: %+v, received: %+v", expOut, out)
	}
	expOut = 1.5
	if out, err := d.Convert(1536.0); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expOut, out) {
		t.Errorf("expecting: %+v, received: %+v", expOut, out)
	}
	if _, err := eDvd.Convert("strionmg"); err == nil || err.Error() != `strconv.ParseFloat: parsing "strionmg": invalid syntax` {
		t.Error(err)
	}
}

func TestDurationConverter(t *testing.T) {
	d, err := NewDataConverter(MetaDuration)
	if err != nil {
		t.Error(err.Error())
	}
	expVal := 10 * time.Second
	if i, err := d.Convert(10000000000.0); err != nil {
		t.Error(err.Error())
	} else if expVal != i {
		t.Errorf("expecting: %d, received: %d", expVal, i)
	}
	if i, err := d.Convert(10000000000); err != nil {
		t.Error(err.Error())
	} else if expVal != i {
		t.Errorf("expecting: %d, received: %d", expVal, i)
	}
	if i, err := d.Convert(10 * time.Second); err != nil {
		t.Error(err.Error())
	} else if expVal != i {
		t.Errorf("expecting: %d, received: %d", expVal, i)
	}
	if i, err := d.Convert("10s"); err != nil {
		t.Error(err.Error())
	} else if expVal != i {
		t.Errorf("expecting: %d, received: %d", expVal, i)
	}
}

func TestPhoneNumberConverter(t *testing.T) {
	// test for error
	if rcv, err := NewDataConverter("*libphonenumber:US:1:2:error"); err == nil || err.Error() != "unsupported *libphonenumber converter parameters: <US:1:2:error>" {
		t.Error(err)
	} else if !reflect.DeepEqual(nil, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", nil, rcv)
	}
	if rcv, err := NewDataConverter("*libphonenumber:US:X"); err == nil || err.Error() != `strconv.Atoi: parsing "X": invalid syntax` {
		t.Error(err)
	} else if !reflect.DeepEqual(nil, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", nil, rcv)
	}
	eLc := &PhoneNumberConverter{CountryCode: "9786679", Format: phonenumbers.NATIONAL}
	if _, err := eLc.Convert("8976wedf"); err == nil || err.Error() != "invalid country code" {
		t.Errorf("Expecting: 'invalid country code', received: %+v", err)
	}

	// US/National
	eLc = &PhoneNumberConverter{CountryCode: "US", Format: phonenumbers.NATIONAL}
	d, err := NewDataConverter("*libphonenumber:US")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLc, d) {
		t.Errorf("expecting: %+v, received: %+v", eLc, d)
	}
	// simulate an E164 number and Format it into a National number
	phoneNumberConverted, err := d.Convert("+14431234567")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(phoneNumberConverted, "(443) 123-4567") {
		t.Errorf("expecting: %+v, received: %+v", "(443) 123-4567", phoneNumberConverted)
	}
	//US/International
	eLc = &PhoneNumberConverter{CountryCode: "US", Format: phonenumbers.INTERNATIONAL}
	d, err = NewDataConverter("*libphonenumber:US:1")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLc, d) {
		t.Errorf("expecting: %+v, received: %+v", eLc, d)
	}
	// simulate an E164 number and Format it into a National number
	phoneNumberConverted, err = d.Convert("+14431234567")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(phoneNumberConverted, "+1 443-123-4567") {
		t.Errorf("expecting: %+v, received: %+v", "+1 443-123-4567", phoneNumberConverted)
	}
	// DE/International
	eLc = &PhoneNumberConverter{CountryCode: "DE", Format: phonenumbers.INTERNATIONAL}
	d, err = NewDataConverter("*libphonenumber:DE:1")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLc, d) {
		t.Errorf("expecting: %+v, received: %+v", eLc, d)
	}
	phoneNumberConverted, err = d.Convert("6502530000")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(phoneNumberConverted, "+49 6502 530000") {
		t.Errorf("expecting: %+v, received: %+v", "+49 6502 530000", phoneNumberConverted)
	}
	// DE/E164
	eLc = &PhoneNumberConverter{CountryCode: "DE", Format: phonenumbers.E164}
	d, err = NewDataConverter("*libphonenumber:DE:0")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLc, d) {
		t.Errorf("expecting: %+v, received: %+v", eLc, d)
	}
	phoneNumberConverted, err = d.Convert("6502530000")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(phoneNumberConverted, "+496502530000") {
		t.Errorf("expecting: %+v, received: %+v", "+496502530000", phoneNumberConverted)
	}
}

func TestHexConvertor(t *testing.T) {
	hx := new(IP2HexConverter)
	val := "127.0.0.1"
	expected := "0x7f000001"
	if rpl, err := hx.Convert(val); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("expecting: %+v, received: %+v", expected, rpl)
	}
	val2 := net.ParseIP("127.0.0.1")
	if rpl, err := hx.Convert(val2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("expecting: %+v, received: %+v", expected, rpl)
	}

	val3 := []byte("127.0.0.1")
	if rpl, err := hx.Convert(val3); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("expecting: %+v, received: %+v", expected, rpl)
	}

	val = ""
	expected = ""
	if rpl, err := hx.Convert(val); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("expecting: %+v, received: %+v", expected, rpl)
	}

	val = "62.87.114.244"
	expected = "0x3e5772f4"
	if rpl, err := hx.Convert(val); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("expecting: %+v, received: %+v", expected, rpl)
	}
}

type testMockConverter struct{}

// Convert function to implement DataConverter
func (*testMockConverter) Convert(any) (any, error) { return nil, ErrNotFound }
func TestDataConvertersConvertString2(t *testing.T) {
	hex, err := NewDataConverter(MetaIP2Hex)
	if err != nil {
		t.Error(err)
	}

	host, err := NewDataConverter(MetaSIPURIHost)
	if err != nil {
		t.Error(err)
	}
	user, err := NewDataConverter(MetaSIPURIUser)
	if err != nil {
		t.Error(err)
	}
	method, err := NewDataConverter(MetaSIPURIMethod)
	if err != nil {
		t.Error(err)
	}

	dc := DataConverters{new(testMockConverter), hex, host, user, method}
	if _, err := dc.ConvertString(""); err != ErrNotFound {
		t.Errorf("Expected error %s ,received %v", ErrNotFound, err)
	}
}

func TestSIPURIConverter(t *testing.T) {
	host := new(SIPURIHostConverter)
	val := "INVITE sip:1002@192.168.58.203 SIP/2.0"
	expected := "192.168.58.203"
	if rply, err := host.Convert(val); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected %q, received: %q", rply, expected)
	}

	method := new(SIPURIMethodConverter)
	expected = "INVITE"
	if rply, err := method.Convert(val); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected %q, received: %q", rply, expected)
	}

	user := new(SIPURIUserConverter)
	expected = "1002"
	if rply, err := user.Convert(val); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected %q, received: %q", rply, expected)
	}

}

func TestNewDataConverterMustCompile2(t *testing.T) {
	defer func() {
		expectedMessage := "parsing: <*multiply>, error: mandatory information missing"
		if r := recover(); r != expectedMessage {
			t.Errorf("Expected %q, received: %q", expectedMessage, r)
		}
	}()
	NewDataConverterMustCompile(MetaMultiply)
}

func TestNewTimeStringConverter(t *testing.T) {
	//empty
	eOut := &TimeStringConverter{Layout: EmptyString}
	if rcv := NewTimeStringConverter(EmptyString); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}

	//default
	eOut = &TimeStringConverter{Layout: time.RFC3339}
	var rcv DataConverter
	if rcv = NewTimeStringConverter(time.RFC3339); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	exp := "2015-07-07T14:52:08Z"
	if rcv, err := rcv.Convert("1436280728"); err != nil {
		t.Error(err)
	} else if rcv.(string) != exp {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
	exp = "2013-07-30T19:33:10Z"
	if rcv, err := rcv.Convert("1375212790"); err != nil {
		t.Error(err)
	} else if rcv.(string) != exp {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}

	//other
	eOut = &TimeStringConverter{"020106150400"}
	if rcv = NewTimeStringConverter("020106150400"); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	exp = "070715145200"
	if rcv, err := rcv.Convert("1436280728"); err != nil {
		t.Error(err)
	} else if rcv.(string) != exp {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
	exp = "290720175900"
	if rcv, err := rcv.Convert("2020-07-29T17:59:59Z"); err != nil {
		t.Error(err)
	} else if rcv.(string) != exp {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}

	//wrong cases
	eOut = &TimeStringConverter{"not really a good time"}
	if rcv = NewTimeStringConverter("not really a good time"); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	exp = "not really a good time"
	if rcv, err := rcv.Convert(EmptyString); err != nil {
		t.Error(err)
	} else if rcv.(string) != exp {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
	if rcv, err := rcv.Convert("1375212790"); err != nil {
		t.Error(err)
	} else if rcv.(string) != exp {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
	if _, err := rcv.Convert("137521s2790"); err == nil {
		t.Errorf("Expected error received: %v:", err)
	}
}

func TestStringHexConvertor(t *testing.T) {
	hx := new(String2HexConverter)
	val := "127.0.0.1"
	expected := "0x3132372e302e302e31"
	if rpl, err := hx.Convert(val); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("expecting: %+v, received: %+v", expected, rpl)
	}

	val3 := []byte("127.0.0.1")
	if rpl, err := hx.Convert(val3); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("expecting: %+v, received: %+v", expected, rpl)
	}

	val = ""
	expected = ""
	if rpl, err := hx.Convert(val); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("expecting: %+v, received: %+v", expected, rpl)
	}

	val3 = []byte{0x94, 0x71, 0x02, 0x31, 0x01, 0x59}
	expected = "0x947102310159"
	if rpl, err := hx.Convert(val3); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("expecting: %+v, received: %+v", expected, rpl)
	}

	val3 = []byte{0x88, 0x90, 0xa6}
	expected = "0x8890a6"
	if rpl, err := hx.Convert(val3); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("expecting: %+v, received: %+v", expected, rpl)
	}
}

func TestUnixTimeConverter(t *testing.T) {
	exp := new(UnixTimeConverter)
	cnv, err := NewDataConverter(MetaUnixTime)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp, cnv) {
		t.Errorf("Expecting: %+v, received: %+v", exp, cnv)
	}

	expected := int64(1436280728)
	if rcv, err := cnv.Convert("2015-07-07T14:52:08Z"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if _, err := cnv.Convert("NotAValidTime"); err == nil {
		t.Errorf("Expected error received %v", err)
	}
}

func TestRandomConverter(t *testing.T) {
	exp := new(RandomConverter)
	if cnv, err := NewRandomConverter(EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, cnv) {
		t.Errorf("Expecting: %+v, received: %+v", exp, cnv)
	}

	if rcv, err := exp.Convert(nil); err != nil {
		t.Error(err)
	} else if rcv == 0 {
		t.Errorf("Expecting different than 0, received: %+v", rcv)
	}
	exp.begin = 10
	if rcv, err := exp.Convert(nil); err != nil {
		t.Error(err)
	} else if rcv.(int) < 10 {
		t.Errorf("Expecting bigger than 10, received: %+v", rcv)
	}
	exp.end = 20
	if rcv, err := exp.Convert(nil); err != nil {
		t.Error(err)
	} else if rcv.(int) < 10 || rcv.(int) > 20 {
		t.Errorf("Expecting bigger than 10 and smaller than 20, received: %+v", rcv)
	}
}

func TestDCNewDataConverterRandomPrefixEmpty(t *testing.T) {

	a, err := NewDataConverter(MetaRandom)
	if err != nil {
		t.Error(err)
	}
	b, err := NewRandomConverter(EmptyString)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("Error reflect")
	}

}

func TestDCNewDataConverterRandomPrefix(t *testing.T) {
	params := "*random:1:2"
	a, err := NewDataConverter(params)
	if err != nil {
		t.Error(err)
	}

	b, err := NewRandomConverter(params[len(MetaRandom)+1:])
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("Error reflect")
	}
}

func TestDCNewRandomConverterCase2Begin(t *testing.T) {
	params := "test:15"

	_, err := NewRandomConverter(params)
	expected := "strconv.Atoi: parsing \"test\": invalid syntax"

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err.Error())
	}
}

func TestDCNewRandomConverterCase2End(t *testing.T) {
	params := "15:test"

	_, err := NewRandomConverter(params)
	expected := "strconv.Atoi: parsing \"test\": invalid syntax"

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err.Error())
	}
}

func TestDCNewRandomConverterCase1Begin(t *testing.T) {
	params := "test"

	_, err := NewRandomConverter(params)
	expected := "strconv.Atoi: parsing \"test\": invalid syntax"

	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, err.Error())
	}
}

func TestDCrCConvert(t *testing.T) {
	randConv := &RandomConverter{
		begin: 0,
		end:   2,
	}

	received, err := randConv.Convert(randConv.begin)
	if err != nil {
		t.Error(err)
	}
	receivedAsInt, err := IfaceAsInt64(received)
	if err != nil {
		t.Error(err)
	}
	if receivedAsInt != 0 && receivedAsInt != 1 {
		t.Errorf("\nExpected 0 or 1, \nReceived: <%+v>", received)
	}
}
func TestLenTimeConverter(t *testing.T) {
	exp := new(LengthConverter)
	cnv, err := NewDataConverter(MetaLen)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp, cnv) {
		t.Errorf("Expecting: %+v, received: %+v", exp, cnv)
	}

	expected := 20
	if rcv, err := cnv.Convert("2015-07-07T14:52:08Z"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}

func TestLenTimeConverter2(t *testing.T) {
	exp := new(LengthConverter)
	cnv, err := NewDataConverter(MetaLen)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp, cnv) {
		t.Errorf("Expecting: %+v, received: %+v", exp, cnv)
	}

	expected := 7
	if rcv, err := cnv.Convert("[slice]"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}

func TestLenTimeConverter3(t *testing.T) {
	exp := new(LengthConverter)
	cnv, err := NewDataConverter(MetaLen)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp, cnv) {
		t.Errorf("Expecting: %+v, received: %+v", exp, cnv)
	}

	expected := 2
	if rcv, err := cnv.Convert([]int{0, 0}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert("[]"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	expected = 0
	if rcv, err := cnv.Convert([]string{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]any{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]bool{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]int{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]int8{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]int16{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]int32{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]int64{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]uint{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]uint8{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]uint16{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]uint32{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]uint64{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]uintptr{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]float32{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]float64{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]complex64{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert([]complex128{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	if rcv, err := cnv.Convert(nil); err != nil {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}

func TestFloat64Converter(t *testing.T) {
	exp := new(Float64Converter)
	cnv, err := NewDataConverter(MetaFloat64)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp, cnv) {
		t.Errorf("Expecting: %+v, received: %+v", exp, cnv)
	}

	expected := 21.7
	if rcv, err := cnv.Convert("21.7"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}

	expected2 := "strconv.ParseFloat: parsing \"invalid_input\": invalid syntax"
	if _, err := cnv.Convert("invalid_input"); err == nil {
		t.Error("Expected error")
	} else if !reflect.DeepEqual(expected2, err.Error()) {
		t.Errorf("Expecting: %+v, received: %+v", expected2, err.Error())
	}
}

func TestSliceConverter(t *testing.T) {
	exp := new(SliceConverter)
	cnv, err := NewDataConverter(MetaSlice)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp, cnv) {
		t.Errorf("Expecting: %+v, received: %+v", exp, cnv)
	}
	expected := []string{"A", "B"}
	if rcv, err := cnv.Convert([]string{"A", "B"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	expected2 := []any{"A", "B"}
	if rcv, err := cnv.Convert(`["A","B"]`); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected2, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}

	if _, err := cnv.Convert(`test`); err != nil {
		t.Error(err)
	}
}

func TestE164FromNAPTRConverter(t *testing.T) {
	exp := new(e164Converter)
	cnv, err := NewDataConverter(E164Converter)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp, cnv) {
		t.Errorf("Expecting: %+v, received: %+v", exp, cnv)
	}
	if _, err := cnv.Convert("8.7.6.5.4.3.2.1"); err == nil {
		t.Error("Error")
	}
	if e164, err := cnv.Convert("8.7.6.5.4.3.2.1.0.1.6.e164.arpa."); err != nil {
		t.Error(err)
	} else if e164 != "61012345678" {
		t.Errorf("received: <%s>", e164)
	}
}

func TestDomainNameFromNAPTRConverter(t *testing.T) {
	exp := new(e164DomainConverter)
	cnv, err := NewDataConverter(E164DomainConverter)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp, cnv) {
		t.Errorf("Expecting: %+v, received: %+v", exp, cnv)
	}

	if dName, err := cnv.Convert("8.7.6.5.4.3.2.1.0.1.6.e164.arpa."); err != nil {
		t.Fatal(err)
	} else if dName != "e164.arpa" {
		t.Errorf("received: <%s>", dName)
	}
	if dName, err := cnv.Convert("8.7.6.5.4.3.2.1.0.1.6.e164.itsyscom.com."); err != nil {
		t.Fatal(err)
	} else if dName != "e164.itsyscom.com" {
		t.Errorf("received: <%s>", dName)
	}
	if dName, err := cnv.Convert("8.7.6.5.4.3.2.1.0.1.6.itsyscom.com."); err != nil {
		t.Fatal(err)
	} else if dName != "8.7.6.5.4.3.2.1.0.1.6.itsyscom.com" {
		t.Errorf("received: <%s>", dName)
	}
}

type structWithFuncField struct {
	ID       string
	Function func(int) bool
}

func TestDataConverterConvertJSONErrUnsupportedType(t *testing.T) {
	dc, err := NewDataConverter(MetaJSON)
	if err != nil {
		t.Error(err)
	}

	obj := structWithFuncField{
		ID: "testStruct",
		Function: func(i int) bool {
			return i != 0
		},
	}

	experr := `json: unsupported type: func(int) bool`
	if rcv, err := dc.Convert(obj); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	} else if rcv != EmptyString {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", EmptyString, rcv)
	}
}

func TestDataConverterConvertJSONOK(t *testing.T) {
	dc, err := NewDataConverter(MetaJSON)
	if err != nil {
		t.Error(err)
	}

	obj := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestCGREv",
		Event: map[string]any{
			AccountField: "1001",
		},
		APIOpts: map[string]any{
			"opt": "value",
		},
	}

	exp := ToJSON(obj)
	if rcv, err := dc.Convert(obj); err != nil {
		t.Error(err)
	} else if rcv.(string) != exp {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestStripConverter(t *testing.T) {
	tests := []struct {
		name           string
		params         string
		input          string
		expected       string
		constructorErr bool
		convertErr     bool
	}{

		{
			name:           "Strip 5 leading characters",
			params:         "*strip:*prefix:5",
			input:          "12345TEST12345",
			expected:       "TEST12345",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip 5 trailing characters",
			params:         "*strip:*suffix:5",
			input:          "12345TEST12345",
			expected:       "12345TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip 5 characters from both sides",
			params:         "*strip:*both:5",
			input:          "12345TEST12345",
			expected:       "TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip all trailing nils",
			params:         "*strip:*suffix:*nil",
			input:          "TEST\u0000\u0000\u0000\u0000",
			expected:       "TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip all leading nils",
			params:         "*strip:*prefix:*nil",
			input:          "\u0000\u0000\u0000\u0000TEST",
			expected:       "TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip all nils from both sides",
			params:         "*strip:*both:*nil",
			input:          "\u0000\u0000TEST\u0000\u0000",
			expected:       "TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip 2 trailing nils",
			params:         "*strip:*suffix:*nil:2",
			input:          "TEST\u0000\u0000\u0000\u0000",
			expected:       "TEST\u0000\u0000",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip 2 leading nils",
			params:         "*strip:*prefix:*nil:2",
			input:          "\u0000\u0000\u0000\u0000TEST",
			expected:       "\u0000\u0000TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip 1 nil from both sides",
			params:         "*strip:*both:*nil:1",
			input:          "\u0000\u0000TEST\u0000\u0000",
			expected:       "\u0000TEST\u0000",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip all trailing spaces",
			params:         "*strip:*suffix:*space",
			input:          "TEST    ",
			expected:       "TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip all leading spaces",
			params:         "*strip:*prefix:*space",
			input:          "    TEST",
			expected:       "TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip all spaces from both sides",
			params:         "*strip:*both:*space",
			input:          "  TEST  ",
			expected:       "TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip 2 trailing spaces",
			params:         "*strip:*suffix:*space:2",
			input:          "TEST    ",
			expected:       "TEST  ",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip 2 leading spaces",
			params:         "*strip:*prefix:*space:2",
			input:          "    TEST",
			expected:       "  TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip 1 space from both sides",
			params:         "*strip:*both:*space:1",
			input:          "  TEST  ",
			expected:       " TEST ",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip all trailing 'abcd' char groups",
			params:         "*strip:*suffix:*char:abcd",
			input:          "TESTabcdabcdabcdabcd",
			expected:       "TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip all leading 'abcd' char groups",
			params:         "*strip:*prefix:*char:abcd",
			input:          "abcdabcdabcdabcdTEST",
			expected:       "TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip all 'abcd' char groups from both sides",
			params:         "*strip:*both:*char:abcd",
			input:          "abcdabcdTESTabcdabcd",
			expected:       "TEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip 2 trailing 'abcd' char groups",
			params:         "*strip:*suffix:*char:abcd:2",
			input:          "TESTabcdabcdabcdabcd",
			expected:       "TESTabcdabcd",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip 2 leading 'abcd' char groups",
			params:         "*strip:*prefix:*char:abcd:2",
			input:          "abcdabcdabcdabcdTEST",
			expected:       "abcdabcdTEST",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip 1 'abcd' char group from both sides",
			params:         "*strip:*both:*char:abcd:1",
			input:          "abcdabcdTESTabcdabcd",
			expected:       "abcdTESTabcd",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Empty third parameter",
			params:         "*strip:*prefix:",
			input:          "TEST",
			expected:       "strip converter: substr parameter cannot be empty",
			constructorErr: true,
			convertErr:     false,
		},
		{
			name:           "Invalid side parameter",
			params:         "*strip:*invalid:*nil",
			input:          "TEST",
			expected:       "strip converter: invalid side parameter",
			constructorErr: false,
			convertErr:     true,
		},
		{
			name:           "Invalid nr. of params *char",
			params:         "*strip:*prefix:*char:*nil:abc:3",
			input:          "TEST",
			expected:       "strip converter: invalid number of parameters (should have 3, 4 or 5)",
			constructorErr: true,
			convertErr:     false,
		},
		{
			name:           "Invalid amount parameter",
			params:         "*strip:*prefix:*char:0:three",
			input:          "000TEST",
			expected:       "strip converter: invalid amount parameter (strconv.Atoi: parsing \"three\": invalid syntax)",
			constructorErr: true,
			convertErr:     false,
		},
		{
			name:           "Strip a prefix longer than the value",
			params:         "*strip:*prefix:5",
			input:          "TEST",
			expected:       "",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip a suffix longer than the value",
			params:         "*strip:*suffix:5",
			input:          "TEST",
			expected:       "",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "Strip from both ends an amount of characters longer than the value",
			params:         "*strip:*both:3",
			input:          "TEST",
			expected:       "",
			constructorErr: false,
			convertErr:     false,
		},
		{
			name:           "*char missing substring case 1",
			params:         "*strip:*prefix:*char",
			input:          "12345TEST",
			expected:       "strip converter: usage of *char implies the need of 4 or 5 non-empty params",
			constructorErr: true,
			convertErr:     false,
		},
		{
			name:           "*char missing substring case 2",
			params:         "*strip:*prefix:*char::2",
			input:          "12345TEST",
			expected:       "strip converter: usage of *char implies the need of 4 or 5 non-empty params",
			constructorErr: true,
			convertErr:     false,
		},
		{
			name:           "*char missing substring case 3",
			params:         "*strip:*prefix:*char:",
			input:          "12345TEST",
			expected:       "strip converter: usage of *char implies the need of 4 or 5 non-empty params",
			constructorErr: true,
			convertErr:     false,
		},
		{
			name:           "*nil/*space too many parameters",
			params:         "*strip:*prefix:*nil:5:12345",
			input:          "12345TEST",
			expected:       "strip converter: cannot have 5 params in *nil/*space case",
			constructorErr: true,
			convertErr:     false,
		},
		{
			name:           "third param numeric with too many params case 1",
			params:         "*strip:*prefix:1:1",
			input:          "12345TEST",
			expected:       "strip converter: just the amount specified, cannot have more than 3 params",
			constructorErr: true,
			convertErr:     false,
		},
		{
			name:           "third param numeric with too many params case 2",
			params:         "*strip:*prefix:1:12345:1",
			input:          "12345TEST",
			expected:       "strip converter: just the amount specified, cannot have more than 3 params",
			constructorErr: true,
			convertErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc, err := NewDataConverter(tt.params)
			if (err != nil) != tt.constructorErr {
				t.Errorf("NewStripConverter() error = %v, constructorErr %v", err, tt.constructorErr)
				return
			}
			if tt.constructorErr {
				if err.Error() != tt.expected {
					t.Errorf("expected error message: %v, received: %v", tt.expected, err.Error())
				}
				return
			}
			rcv, err := sc.Convert(tt.input)
			if (err != nil) != tt.convertErr {
				t.Errorf("Convert() error = %v, convertErr %v", err, tt.convertErr)
				return
			}
			if tt.convertErr {
				if err.Error() != tt.expected {
					t.Errorf("expected error message: %s, received: %s", tt.expected, err.Error())
				}
				return
			}
			if rcv != tt.expected {
				t.Errorf("expected: %q, received: %q", tt.expected, rcv)
			}
		})
	}
}

func TestURLDecodeConverter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		{
			name:     "Decode string with escaped $ character",
			input:    "123%24123",
			expected: "123$123",
			isError:  false,
		},
		{
			name:     "Wrong escaped character",
			input:    "a%2Fdestination%user%26password%2Cid",
			expected: "invalid URL escape \"%us\"",
			isError:  true,
		},
		{
			name:     "Decode string with multiple escaped characters",
			input:    "query=%40special%23characters%24",
			expected: "query=@special#characters$",
			isError:  false,
		},
		{
			name:     "Decode url with escaped chars",
			input:    "https://www.example.com/search?q=hello%20world&query=%40special%23characters%24&path=%2Fa%20b%2Fc%3Fd%3D1%26e%3D2&data=%E6%97%A5%E6%9C%AC%E8%AA%9E",
			expected: "https://www.example.com/search?q=hello world&query=@special#characters$&path=/a b/c?d=1&e=2&data=日本語",
			isError:  false,
		},
	}
	conv, err := NewDataConverter(URLDecConverter)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := conv.Convert(tt.input)
			if (err != nil) != tt.isError {
				t.Errorf("Convert() error =%v,expected err %v", err, tt.isError)
			}
			if tt.isError {
				if err.Error() != tt.expected {
					t.Errorf("expected error message: %s, received: %s", tt.expected, err.Error())
				}
				return
			}
			if rcv != tt.expected {
				t.Errorf("expected: %q, received: %q", tt.expected, rcv)
			}
		})
	}
}

func TestURLEncodeConverter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "Encode a string with special character", input: "123$123", expected: "123%24123"},
		{name: "Encode special characters in path,query and fragment", input: "https://www.example.com/search日?data=日本語&path=/a b/c?d=1&e=2&q=hello world&query=@special#日本characters$", expected: "https://www.example.com/search%E6%97%A5?data=%E6%97%A5%E6%9C%AC%E8%AA%9E&e=2&path=%2Fa+b%2Fc%3Fd%3D1&q=hello+world&query=%40special#%E6%97%A5%E6%9C%ACcharacters$"},
		{name: "Encode a string with multiple special character", input: "foo☺@$'()*,baz;?&=#+!", expected: "foo%E2%98%BA%40%24%27%28%29%2A%2Cbaz%3B"},
	}
	conv, err := NewDataConverter(URLEncConverter)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := conv.Convert(tt.input)
			if err != nil {
				t.Error(err)
				return
			}
			if tt.expected != rcv {
				t.Errorf("expected %q,received %q", tt.expected, rcv)
			}
		})
	}
}

func TestStripConverterConvert(t *testing.T) {
	tests := []struct {
		name       string
		converter  StripConverter
		input      any
		wantOutput any
		wantErr    error
	}{
		{
			name: "Valid prefix strip",
			converter: StripConverter{
				amount: 3,
				side:   MetaPrefix,
				substr: EmptyString,
			},
			input:      "DatSms",
			wantOutput: "Sms",
			wantErr:    nil,
		},
		{
			name: "Valid suffix strip",
			converter: StripConverter{
				amount: 3,
				side:   MetaSuffix,
				substr: EmptyString,
			},
			input:      "abcdef",
			wantOutput: "abc",
			wantErr:    nil,
		},
		{
			name: "Invalid input type",
			converter: StripConverter{
				amount: 3,
				side:   MetaPrefix,
			},
			input:      123,
			wantOutput: nil,
			wantErr:    ErrCastFailed,
		},
		{
			name: "No strip with non-positive amount",
			converter: StripConverter{
				amount: 0,
				side:   MetaPrefix,
			},
			input:      "cgrates",
			wantOutput: "cgrates",
			wantErr:    nil,
		},
		{
			name: "Trim prefix using substring",
			converter: StripConverter{
				amount: -1,
				side:   MetaPrefix,
				substr: "data",
			},
			input:      "dataTariff",
			wantOutput: "Tariff",
			wantErr:    nil,
		},
		{
			name: "Trim both sides using substring",
			converter: StripConverter{
				amount: -1,
				side:   MetaBoth,
				substr: "a",
			},
			input:      "aaaabcdefaaa",
			wantOutput: "bcdef",
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.converter.Convert(tt.input)
			if got != tt.wantOutput {
				t.Errorf("Convert() = %v, want %v", got, tt.wantOutput)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Convert() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
func TestGigaWordsConverter(t *testing.T) {
	converter := GigawordsConverter{}
	multiplier := int64(4294967296)
	testCases := []struct {
		name          string
		input         any
		expectedValue int64
		expectError   bool
		errorContains string
	}{
		{
			name:          "Input Zero (int)",
			input:         int(0),
			expectedValue: 0,
			expectError:   false,
		},
		{
			name:          "Input Zero (int32)",
			input:         int32(0),
			expectedValue: 0,
			expectError:   false,
		},
		{
			name:          "Input One (int)",
			input:         int(1),
			expectedValue: multiplier,
			expectError:   false,
		},
		{
			name:          "Input Two (int64)",
			input:         int64(2),
			expectedValue: 2 * multiplier,
			expectError:   false,
		},
		{
			name:          "Input Three (string)",
			input:         "3",
			expectedValue: 3 * multiplier,
			expectError:   false,
		},
		{
			name:          "Input Nil",
			input:         nil,
			expectedValue: 0,
			expectError:   true,
			errorContains: "cannot convert",
		},
		{
			name:          "Input Invalid String",
			input:         "abc",
			expectedValue: 0,
			expectError:   true,
			errorContains: "strconv.ParseInt: parsing \"abc\": invalid syntax",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := converter.Convert(tc.input)
			if tc.expectError {
				if err == nil || !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error '%s', but got '%v'", tc.errorContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}

			outputInt64, ok := output.(int64)
			if !ok {
				t.Fatalf("Expected output type int64, but got %T (%v)", output, output)
			}
			if outputInt64 != tc.expectedValue {
				t.Errorf("Expected output %d, but got %d", tc.expectedValue, outputInt64)
			}
		})
	}
}

func TestDurationMinutesConverter(t *testing.T) {
	converter := DurationMinutesConverter{}
	testCases := []struct {
		name          string
		input         any
		expectedValue float64
		expectError   bool
		errorContains string
	}{
		{
			name:          "Input Zero Duration",
			input:         time.Duration(0),
			expectedValue: 0.0,
			expectError:   false,
		},
		{
			name:          "Input 90 Seconds",
			input:         90 * time.Second,
			expectedValue: 1.5,
			expectError:   false,
		},
		{
			name:          "Input 2 Hour",
			input:         2 * time.Hour,
			expectedValue: 120,
			expectError:   false,
		},
		{
			name:          "Input Zero Duration",
			input:         "0s",
			expectedValue: 0.0,
			expectError:   false,
		},
		{
			name:          "Input 90 Seconds",
			input:         "90s",
			expectedValue: 1.5,
			expectError:   false,
		},
		{
			name:          "Input 1 Minute 30 Seconds",
			input:         "1m30s",
			expectedValue: 1.5,
			expectError:   false,
		},
		{
			name:          "Input 2 Hours",
			input:         "2h",
			expectedValue: 120,
			expectError:   false,
		},
		{
			name:          "Input Invalid String",
			input:         "abc",
			expectError:   true,
			errorContains: "invalid duration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := converter.Convert(tc.input)
			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected an error containing '%s', but got nil", tc.errorContains)
				}
				if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error message to contain '%s', but got: %v", tc.errorContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}
			outputFloat64, ok := output.(float64)
			if !ok {
				t.Fatalf("Expected output type float64, but got %T (%v)", output, output)
			}
			if outputFloat64 != tc.expectedValue {
				t.Errorf("Expected output %.2f, but got %.2f", tc.expectedValue, outputFloat64)
			}
		})
	}
}

func TestLocalTimeDurationConverter(t *testing.T) {
	testCases := []struct {
		name        string
		input       any
		params      string
		expectValue string
		expectedErr error
	}{
		{name: "Convert to CEST timezone", input: "2025-05-07T14:25:08Z", params: "*localtime:Europe/Berlin", expectValue: "2025-05-07 16:25:08"},
		{name: "Convert to UTC timezone", input: "2025-05-07T16:25:08+02:00", params: "*localtime:UTC", expectValue: "2025-05-07 14:25:08"},
		{name: "Convert to UTC+01:00 timezone", input: "2025-05-07T14:25:08Z", params: "*localtime:Europe/Dublin", expectValue: "2025-05-07 15:25:08"},
		{name: "Convert to UTC+03:00 timezone", input: time.Date(2025, 5, 5, 15, 5, 0, 0, time.UTC), params: "*localtime:Europe/Istanbul", expectValue: "2025-05-05 18:05:00"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			converter, err := NewDataConverter(tc.params)
			if err != nil {
				t.Fatal(err)
			}
			val, err := converter.Convert(tc.input)
			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("Expected an error containing '%s', but got nil", tc.expectedErr.Error())
				}
				if !strings.Contains(err.Error(), tc.expectedErr.Error()) {
					t.Errorf("Expected error message to contain '%s', but got: %v", tc.expectedErr.Error(), err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}
			if tc.expectValue != val {
				t.Errorf("Expected output %s, but got %s", tc.expectValue, val)
			}
		})
	}
}
