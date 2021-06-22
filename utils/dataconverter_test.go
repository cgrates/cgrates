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
	"net"
	"reflect"
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
func (*testMockConverter) Convert(interface{}) (interface{}, error) { return nil, ErrNotFound }
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

	expected := 0
	if rcv, err := cnv.Convert("[]"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}
