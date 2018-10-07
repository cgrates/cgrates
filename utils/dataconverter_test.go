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
)

func TestNewDataConverter(t *testing.T) {
	a, err := NewDataConverter(MetaDurationSeconds)
	if err != nil {
		t.Error(err.Error())
	}
	b, err := NewDurationSecondsConverter("")
	if err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("Error reflect")
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
}
