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
