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
	b, err := NewDurationSecondsConverter("")
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

func TestRoundConverterFloat64(t *testing.T) {
	b, err := NewRoundConverter("2")
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
	b, err := NewRoundConverter("2")
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
	b, err := NewRoundConverter("2")
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
	b, err := NewRoundConverter("2")
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
