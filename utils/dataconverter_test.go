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
		t.Error(err.Error())
	}
	b, err := NewDurationSecondsConverter(EmptyString)
	if err != nil {
		t.Error(err.Error())
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
	if a, err = NewDataConverter(MetaMultiply); err == nil || err != ErrMandatoryIeMissingNoCaps {
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
	if a, err = NewDataConverter(MetaDivide); err == nil || err != ErrMandatoryIeMissingNoCaps {
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
	if a, err = NewDataConverter(MetaLibPhoneNumber); err == nil || err.Error() != "unsupported *libphonenumber converter parameters: <>" {
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
	if _, err := NewDataConverter("unsupported"); err == nil || err.Error() != "unsupported converter definition: <unsupported>" {
	}

}

func TestNewDataConverterMustCompile(t *testing.T) {
	eOut, _ := NewDataConverter(MetaDurationSeconds)
	if rcv := NewDataConverterMustCompile(MetaDurationSeconds); rcv != eOut {
		t.Errorf("Expecting:  received: %+q", rcv)
	}
}

func TestConvertFloatToSeconds(t *testing.T) {
	b, err := NewDataConverter(MetaDurationSeconds)
	if err != nil {
		t.Error(err.Error())
	}
	a, err := b.Convert(time.Duration(10*time.Second + 300*time.Millisecond))
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
	if i, err := d.Convert(time.Duration(102)); err != nil {
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
		Method:   ROUNDING_MIDDLE,
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
		Method:   ROUNDING_MIDDLE,
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
		Method:   ROUNDING_MIDDLE,
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
		Method:   ROUNDING_MIDDLE,
	}
	if !reflect.DeepEqual(b, expData) {
		t.Errorf("Expected %+v received: %+v", expData, b)
	}
	val, err := b.Convert(time.Duration(123 * time.Nanosecond))
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
	if out, err := m.Convert(time.Duration(2)); err != nil {
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
	if out, err := d.Convert(time.Duration(2048)); err != nil {
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
	expVal := time.Duration(10 * time.Second)
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
	if i, err := d.Convert(time.Duration(10 * time.Second)); err != nil {
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
	if _, err := NewDataConverter("*libphonenumber:US:1:2:error"); err == nil ||
		err.Error() != "unsupported *libphonenumber converter parameters: <US:1:2:error>" {
		t.Error(err)
	}

	eLc := &PhoneNumberConverter{CountryCode: "US", Format: phonenumbers.NATIONAL}
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
}

func TestPhoneNumberConverter2(t *testing.T) {
	eLc := &PhoneNumberConverter{CountryCode: "US", Format: phonenumbers.INTERNATIONAL}
	d, err := NewDataConverter("*libphonenumber:US:1")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLc, d) {
		t.Errorf("expecting: %+v, received: %+v", eLc, d)
	}
	// simulate an E164 number and Format it into a National number
	phoneNumberConverted, err := d.Convert("+14431234567")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(phoneNumberConverted, "+1 443-123-4567") {
		t.Errorf("expecting: %+v, received: %+v", "+1 443-123-4567", phoneNumberConverted)
	}
}

func TestPhoneNumberConverter3(t *testing.T) {
	eLc := &PhoneNumberConverter{CountryCode: "DE", Format: phonenumbers.INTERNATIONAL}
	d, err := NewDataConverter("*libphonenumber:DE:1")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLc, d) {
		t.Errorf("expecting: %+v, received: %+v", eLc, d)
	}
	phoneNumberConverted, err := d.Convert("6502530000")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(phoneNumberConverted, "+49 6502 530000") {
		t.Errorf("expecting: %+v, received: %+v", "+49 6502 530000", phoneNumberConverted)
	}

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
