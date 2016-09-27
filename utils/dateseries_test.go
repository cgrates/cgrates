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
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestDateseriesMonthStoreRestoreJson(t *testing.T) {
	m := Months{5, 6, 7, 8}
	r, _ := json.Marshal(m)
	if string(r) != "[5,6,7,8]" {
		t.Errorf("Error serializing months: %v", string(r))
	}
	o := Months{}
	json.Unmarshal(r, &o)
	if !reflect.DeepEqual(o, m) {
		t.Errorf("Expected %v was  %v", m, o)
	}
}

func TestDateseriesMonthDayStoreRestoreJson(t *testing.T) {
	md := MonthDays{24, 25, 26}
	r, _ := json.Marshal(md)
	if string(r) != "[24,25,26]" {
		t.Errorf("Error serializing month days: %v", string(r))
	}
	o := MonthDays{}
	json.Unmarshal(r, &o)
	if !reflect.DeepEqual(o, md) {
		t.Errorf("Expected %v was  %v", md, o)
	}
}

func TestDateseriesWeekDayStoreRestoreJson(t *testing.T) {
	wd := WeekDays{time.Saturday, time.Sunday}
	r, _ := json.Marshal(wd)
	if string(r) != "[6,0]" {
		t.Errorf("Error serializing week days: %v", string(r))
	}
	o := WeekDays{}
	json.Unmarshal(r, &o)
	if !reflect.DeepEqual(o, wd) {
		t.Errorf("Expected %v was  %v", wd, o)
	}
}

func TestDateseriesYearsSerialize(t *testing.T) {
	ys := &Years{}
	yString := ys.Serialize(";")
	expectString := "*any"
	if expectString != yString {
		t.Errorf("Expected: %s, got: %s", expectString, yString)
	}
	ys2 := &Years{2012}
	yString2 := ys2.Serialize(";")
	expectString2 := "2012"
	if expectString2 != yString2 {
		t.Errorf("Expected: %s, got: %s", expectString2, yString2)
	}
	ys3 := &Years{2013, 2014, 2015}
	yString3 := ys3.Serialize(";")
	expectString3 := "2013;2014;2015"
	if expectString3 != yString3 {
		t.Errorf("Expected: %s, got: %s", expectString3, yString3)
	}
}

func TestDateseriesMonthsSerialize(t *testing.T) {
	mths := &Months{}
	mString := mths.Serialize(";")
	expectString := "*any"
	if expectString != mString {
		t.Errorf("Expected: %s, got: %s", expectString, mString)
	}
	mths2 := &Months{time.January}
	mString2 := mths2.Serialize(";")
	expectString2 := "1"
	if expectString2 != mString2 {
		t.Errorf("Expected: %s, got: %s", expectString2, mString2)
	}
	mths3 := &Months{time.January, time.December}
	mString3 := mths3.Serialize(";")
	expectString3 := "1;12"
	if expectString3 != mString3 {
		t.Errorf("Expected: %s, got: %s", expectString3, mString3)
	}
}

func TestDateseriesMonthDaysSerialize(t *testing.T) {
	mds := &MonthDays{}
	mdsString := mds.Serialize(";")
	expectString := "*any"
	if expectString != mdsString {
		t.Errorf("Expected: %s, got: %s", expectString, mdsString)
	}
	mds2 := &MonthDays{1}
	mdsString2 := mds2.Serialize(";")
	expectString2 := "1"
	if expectString2 != mdsString2 {
		t.Errorf("Expected: %s, got: %s", expectString2, mdsString2)
	}
	mds3 := &MonthDays{1, 2, 3, 4, 5}
	mdsString3 := mds3.Serialize(";")
	expectString3 := "1;2;3;4;5"
	if expectString3 != mdsString3 {
		t.Errorf("Expected: %s, got: %s", expectString3, mdsString3)
	}
}

func TestDateseriesWeekDaysSerialize(t *testing.T) {
	wds := &WeekDays{}
	wdsString := wds.Serialize(";")
	expectString := "*any"
	if expectString != wdsString {
		t.Errorf("Expected: %s, got: %s", expectString, wdsString)
	}
	wds2 := &WeekDays{time.Monday}
	wdsString2 := wds2.Serialize(";")
	expectString2 := "1"
	if expectString2 != wdsString2 {
		t.Errorf("Expected: %s, got: %s", expectString2, wdsString2)
	}
	wds3 := &WeekDays{time.Monday, time.Saturday, time.Sunday}
	wdsString3 := wds3.Serialize(";")
	expectString3 := "1;6;0"
	if expectString3 != wdsString3 {
		t.Errorf("Expected: %s, got: %s", expectString3, wdsString3)
	}
}

func TestDateseriesMonthsIsCompleteNot(t *testing.T) {
	months := Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November}
	if months.IsComplete() {
		t.Error("Error months IsComplete: ", months)
	}
}

func TestDateseriesMonthsIsCompleteYes(t *testing.T) {
	months := Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December}
	if !months.IsComplete() {
		t.Error("Error months IsComplete: ", months)
	}
}

func TestDateseriesDaysInMonth(t *testing.T) {
	if n := DaysInMonth(2016, 4); n != 30 {
		t.Error("error calculating days: ", n)
	}
	if n := DaysInMonth(2016, 2); n != 29 {
		t.Error("error calculating days: ", n)
	}
	if n := DaysInMonth(2016, 1); n != 31 {
		t.Error("error calculating days: ", n)
	}
	if n := DaysInMonth(2016, 12); n != 31 {
		t.Error("error calculating days: ", n)
	}
	if n := DaysInMonth(2015, 2); n != 28 {
		t.Error("error calculating days: ", n)
	}
}

func TestDateseriesDaysInYear(t *testing.T) {
	if n := DaysInYear(2016); n != 366 {
		t.Error("error calculating days: ", n)
	}
	if n := DaysInYear(2015); n != 365 {
		t.Error("error calculating days: ", n)
	}
}
