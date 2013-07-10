/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package rater

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestMonthStoreRestoreJson(t *testing.T) {
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

func TestMonthDayStoreRestoreJson(t *testing.T) {
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

func TestWeekDayStoreRestoreJson(t *testing.T) {
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

func TestYearsSerialize(t *testing.T) {
	ys := &Years{}
	yString := ys.Serialize(";")
	expectString := "*all"
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

func TestMonthsSerialize(t *testing.T) {
	mths := &Months{}
	mString := mths.Serialize(";")
	expectString := "*none"
	if expectString != mString {
		t.Errorf("Expected: %s, got: %s", expectString, mString)
	}
	mths1 := Months(allMonths)
	mString1 := mths1.Serialize(";")
	expectString1 := "*all"
	if expectString1 != mString1 {
		t.Errorf("Expected: %s, got: %s", expectString1, mString1)
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

func TestMonthDaysSerialize(t *testing.T) {
	mds := &MonthDays{}
	mdsString := mds.Serialize(";")
	expectString := "*none"
	if expectString != mdsString {
		t.Errorf("Expected: %s, got: %s", expectString, mdsString)
	}
	mds1 := MonthDays(allMonthDays)
	mdsString1 := mds1.Serialize(";")
	expectString1 := "*all"
	if expectString1 != mdsString1 {
		t.Errorf("Expected: %s, got: %s", expectString1, mdsString1)
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

func TestWeekDaysSerialize(t *testing.T) {
	wds := &WeekDays{}
	wdsString := wds.Serialize(";")
	expectString := "*none"
	if expectString != wdsString {
		t.Errorf("Expected: %s, got: %s", expectString, wdsString)
	}
	wds1 := WeekDays(allWeekDays)
	wdsString1 := wds1.Serialize(";")
	expectString1 := "*all"
	if expectString1 != wdsString1 {
		t.Errorf("Expected: %s, got: %s", expectString1, wdsString1)
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
