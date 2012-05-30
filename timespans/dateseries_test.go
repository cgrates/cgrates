/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	"testing"
	"reflect"
	"time"
)

func TestMonthStoreRestore(t *testing.T) {
	m := Months{5, 6, 7, 8}
	r := m.store()
	if r != "5,6,7,8," {
		t.Errorf("Error serializing months: %v", r)
	}
	o := Months{}
	o.restore(r)
	if !reflect.DeepEqual(o, m) {
		t.Errorf("Expected %v was  %v", m, o)
	}
}

func TestMonthDayStoreRestore(t *testing.T) {
	md := MonthDays{24, 25, 26}
	r := md.store()
	if r != "24,25,26," {
		t.Errorf("Error serializing month days: %v", r)
	}
	o := MonthDays{}
	o.restore(r)
	if !reflect.DeepEqual(o, md) {
		t.Errorf("Expected %v was  %v", md, o)
	}
}

func TestWeekDayStoreRestore(t *testing.T) {
	wd := WeekDays{time.Saturday, time.Sunday}
	r := wd.store()
	if r != "6,0," {
		t.Errorf("Error serializing week days: %v", r)
	}
	o := WeekDays{}
	o.restore(r)
	if !reflect.DeepEqual(o, wd) {
		t.Errorf("Expected %v was  %v", wd, o)
	}
}
