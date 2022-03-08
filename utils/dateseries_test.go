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

func TestYearsSort(t *testing.T) {
	ys := &Years{}
	ys.Sort()
	if !reflect.DeepEqual(ys, &Years{}) {
		t.Errorf("Expecting %+v received: %+v", &Years{}, ys)
	}
	ys = &Years{2019, 2010, 2020, 2005, 2018, 2007}
	ysSorted := &Years{2005, 2007, 2010, 2018, 2019, 2020}
	ys.Sort()
	if !reflect.DeepEqual(ys, ysSorted) {
		t.Errorf("Expecting %+v received: %+v", ysSorted, ys)
	}
}

func TestYearsLen(t *testing.T) {
	ys := &Years{}
	if rcv := ys.Len(); rcv != 0 {
		t.Errorf("Expecting %+v received: %+v", 0, rcv)
	}
	ys = &Years{2019, 2010, 2020, 2005, 2018, 2007}
	if rcv := ys.Len(); rcv != 6 {
		t.Errorf("Expecting %+v received: %+v", 6, rcv)
	}
}

func TestYearsSwap(t *testing.T) {
	ys := &Years{2019, 2010, 2020, 2005, 2018, 2007}
	ys.Swap(0, 1)
	ysSwapped := &Years{2010, 2019, 2020, 2005, 2018, 2007}
	if !reflect.DeepEqual(ys, ysSwapped) {
		t.Errorf("Expecting %+v received: %+v", ysSwapped, ys)
	}
}

func TestYearsLess(t *testing.T) {
	ys := &Years{2019, 2010, 2020, 2005, 2018, 2007}
	if ys.Less(0, 1) {
		t.Errorf("Expecting false received: true")
	}
	if !ys.Less(1, 2) {
		t.Errorf("Expecting true received: false")
	}
}

func TestYearsParse(t *testing.T) {
	ys1 := Years{}
	ys1.Parse(MetaAny, EmptyString)
	ys2 := Years{2013, 2014, 2015}
	in := "2013,2014,2015"
	if reflect.DeepEqual(ys2, ys1) != false {
		t.Errorf("Expected: %+v, received: %+v", Years{}, ys1)
	}
	ys1.Parse(in, FieldsSep)
	if !reflect.DeepEqual(ys2, ys1) {
		t.Errorf("Expected: %+v, received: %+v", ys2, ys1)
	}
}

func TestMonthsSort(t *testing.T) {
	m := &Months{}
	m.Sort()
	if !reflect.DeepEqual(m, &Months{}) {
		t.Errorf("Expecting %+v received: %+v", &Months{}, m)
	}
	m = &Months{time.November, time.July, time.April, time.December, time.October, time.August}
	mSorted := &Months{time.April, time.July, time.August, time.October, time.November, time.December}
	m.Sort()
	if !reflect.DeepEqual(m, mSorted) {
		t.Errorf("Expecting %+v received: %+v", mSorted, m)
	}
}

func TestMonthsLen(t *testing.T) {
	m := &Months{}
	if rcv := m.Len(); rcv != 0 {
		t.Errorf("Expecting %+v received: %+v", 0, rcv)
	}
	m = &Months{time.November, time.July, time.April, time.December, time.October, time.August}
	if rcv := m.Len(); rcv != 6 {
		t.Errorf("Expecting %+v received: %+v", 6, rcv)
	}
}

func TestMonthsSwap(t *testing.T) {
	m := &Months{time.November, time.July, time.April, time.December, time.October, time.August}
	m.Swap(0, 1)
	mSwapped := &Months{time.July, time.November, time.April, time.December, time.October, time.August}
	if !reflect.DeepEqual(mSwapped, m) {
		t.Errorf("Expecting %+v received: %+v", mSwapped, m)
	}
}

func TestMonthsLess(t *testing.T) {
	m := &Months{time.November, time.July, time.December, time.April, time.October, time.August}
	if !m.Less(1, 2) {
		t.Errorf("Expecting true received: false")
	}
	if m.Less(0, 1) {
		t.Errorf("Expecting false received: true")
	}
}

func TestMonthsParse(t *testing.T) {
	m1 := Months{}
	m1.Parse(MetaAny, EmptyString)
	eOut := Months{time.May, time.June, time.July, time.August}
	if m1.Parse("5,6,7,8", FieldsSep); !reflect.DeepEqual(eOut, m1) {
		t.Errorf("Expected: %+v, received: %+v", eOut, m1)
	}
}

func TestMonthDaysSort(t *testing.T) {
	md := &MonthDays{}
	md.Sort()
	if !reflect.DeepEqual(md, &MonthDays{}) {
		t.Errorf("Expecting %+v received: %+v", &MonthDays{}, md)
	}
	md = &MonthDays{7, 3, 5}
	mdSorted := &MonthDays{3, 5, 7}
	md.Sort()
	if !reflect.DeepEqual(md, mdSorted) {
		t.Errorf("Expecting %+v received: %+v", mdSorted, md)
	}
}

func TestMonthDaysLen(t *testing.T) {
	ys := &MonthDays{}
	if rcv := ys.Len(); rcv != 0 {
		t.Errorf("Expecting %+v received: %+v", 0, rcv)
	}
	ys = &MonthDays{3, 5, 7, 21, 25, 18}
	if rcv := ys.Len(); rcv != 6 {
		t.Errorf("Expecting %+v received: %+v", 6, rcv)
	}
}

func TestMonthDaysSwap(t *testing.T) {
	ys := &MonthDays{3, 5, 7, 21, 25, 18}
	ys.Swap(0, 1)
	ysSwapped := &MonthDays{5, 3, 7, 21, 25, 18}
	if !reflect.DeepEqual(ys, ysSwapped) {
		t.Errorf("Expecting %+v received: %+v", ysSwapped, ys)
	}
}

func TestMonthDaysLess(t *testing.T) {
	ys := &MonthDays{3, 7, 5, 21, 25, 18}
	if !ys.Less(2, 3) {
		t.Errorf("Expecting true received: false")
	}
	if ys.Less(1, 2) {
		t.Errorf("Expecting false received: true")
	}

}

func TestMonthDaysParse(t *testing.T) {
	md1 := MonthDays{}
	md1.Parse(MetaAny, EmptyString)

	eOut := MonthDays{24, 25, 26}
	md1.Parse("24,25,26", ",")
	if !reflect.DeepEqual(eOut, md1) {
		t.Errorf("Expected: %+v, received: %+v", eOut, md1)
	}
}

func TestWeekdaysSort(t *testing.T) {
	wd := &WeekDays{}
	wd.Sort()
	if !reflect.DeepEqual(wd, &WeekDays{}) {
		t.Errorf("Expecting %+v received: %+v", &WeekDays{}, wd)
	}
	wd = &WeekDays{time.Thursday, time.Sunday, time.Monday, time.Friday}
	wdSorted := &WeekDays{time.Sunday, time.Monday, time.Thursday, time.Friday}
	wd.Sort()
	if !reflect.DeepEqual(wd, wdSorted) {
		t.Errorf("Expecting %+v received: %+v", wdSorted, wd)
	}
}

func TestWeekDaysLen(t *testing.T) {
	wd := &WeekDays{}
	if rcv := wd.Len(); rcv != 0 {
		t.Errorf("Expecting %+v received: %+v", 0, rcv)
	}
	wd = &WeekDays{time.Thursday, time.Sunday, time.Monday, time.Friday}
	if rcv := wd.Len(); rcv != 4 {
		t.Errorf("Expecting %+v received: %+v", 4, rcv)
	}
}

func TestWeekDaysSwap(t *testing.T) {
	wd := &WeekDays{time.Thursday, time.Sunday, time.Monday, time.Friday}
	wd.Swap(0, 1)
	ysSwapped := &WeekDays{time.Sunday, time.Thursday, time.Monday, time.Friday}
	if !reflect.DeepEqual(wd, ysSwapped) {
		t.Errorf("Expecting %+v received: %+v", ysSwapped, wd)
	}
}

func TestWeekDaysLess(t *testing.T) {
	ys := &WeekDays{time.Thursday, time.Sunday, time.Monday, time.Friday}
	if ys.Less(0, 1) {
		t.Errorf("Expecting false received: true")
	}
	if !ys.Less(1, 2) {
		t.Errorf("Expecting true received: false")
	}
}

func TestWeekDaysParse(t *testing.T) {
	wd := WeekDays{}
	wd.Parse(MetaAny, EmptyString)
	eOut := WeekDays{time.Monday, time.Tuesday, time.Wednesday}
	wd.Parse("1,2,3", FieldsSep)
	if !reflect.DeepEqual(eOut, wd) {
		t.Errorf("Expected: %+v, received: %+v", eOut, wd)
	}
}

func TestLocalAddr(t *testing.T) {
	eOut := &NetAddr{network: Local, ip: Local}
	if rcv := LocalAddr(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
}

func TestNewNetAddr(t *testing.T) {
	eOut := &NetAddr{}
	if rcv := NewNetAddr(EmptyString, EmptyString); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
	eOut = &NetAddr{network: "network"}
	if rcv := NewNetAddr("network", EmptyString); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
	eOut = &NetAddr{ip: "127.0.0.1", port: 2012}
	if rcv := NewNetAddr(EmptyString, "127.0.0.1:2012"); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
	eOut = &NetAddr{network: "network", ip: "127.0.0.1", port: 2012}
	if rcv := NewNetAddr("network", "127.0.0.1:2012"); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
}

func TestNetAddrNetwork(t *testing.T) {
	lc := NetAddr{network: "network", ip: "127.0.0.1", port: 2012}
	if rcv := lc.Network(); !reflect.DeepEqual(rcv, lc.network) {
		t.Errorf("Expected: %+v, received: %+v ", lc.network, rcv)
	}
}

func TestNetAddrString(t *testing.T) {
	lc := NetAddr{network: "network", ip: "127.0.0.1", port: 2012}
	if rcv := lc.String(); !reflect.DeepEqual(rcv, lc.ip) {
		t.Errorf("Expected: %+v, received: %+v ", lc.ip, rcv)
	}
}

func TestNetAddrPort(t *testing.T) {
	lc := NetAddr{network: "network", ip: "127.0.0.1", port: 2012}
	if rcv := lc.Port(); !reflect.DeepEqual(rcv, lc.port) {
		t.Errorf("Expected: %+v, received: %+v ", lc.port, rcv)
	}
}

func TestNetAddrHost(t *testing.T) {
	lc := NetAddr{network: "network", ip: "127.0.0.1", port: 2012}
	if rcv := lc.Host(); !reflect.DeepEqual(rcv, "127.0.0.1:2012") {
		t.Errorf("Expected: '127.0.0.1:2012', received: %+v ", rcv)
	}
	lc = NetAddr{network: "network", ip: Local, port: 2012}
	if rcv := lc.Host(); !reflect.DeepEqual(rcv, Local) {
		t.Errorf("Expected: %+v, received: %+v ", Local, rcv)
	}
}

func TestMonthStoreRestoreJson(t *testing.T) {
	m := Months{time.May, time.June, time.July, time.August}
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
