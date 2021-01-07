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

func TestYearsContains(t *testing.T) {
	ys := &Years{2019, 2010, 2020, 2005, 2018, 2007}
	if !ys.Contains(2019) {
		t.Errorf("Expecting true received: false")
	}
	if ys.Contains(1989) {
		t.Errorf("Expecting false received: true")
	}
	ys = &Years{2013, 2014, 2015}
	if !ys.Contains(2013) {
		t.Errorf("Expected: true, received: false")
	} else if ys.Contains(2012) {
		t.Errorf("Expected: false, received: true")
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
	ys1.Parse(in, FIELDS_SEP)
	if !reflect.DeepEqual(ys2, ys1) {
		t.Errorf("Expected: %+v, received: %+v", ys2, ys1)
	}
}

func TestYearsSerialize(t *testing.T) {
	ys := &Years{}
	eOut := MetaAny
	if yString := ys.Serialize(INFIELD_SEP); eOut != yString {
		t.Errorf("Expected: %s, received: %s", eOut, yString)
	}
	ys = &Years{2012}
	eOut = "2012"
	if yString := ys.Serialize(INFIELD_SEP); eOut != yString {
		t.Errorf("Expected: %s, received: %s", eOut, yString)
	}
	ys = &Years{2013, 2014, 2015}
	eOut = "2013;2014;2015"
	if yString := ys.Serialize(INFIELD_SEP); eOut != yString {
		t.Errorf("Expected: %s, received: %s", eOut, yString)
	}
}

func TestYearsEquals(t *testing.T) {
	ys1 := Years{2013, 2014, 2015}
	ys2 := Years{2013, 2014, 2015}
	ys3 := Years{2013, 2014, 2020}
	ys4 := Years{}
	if !ys1.Equals(ys2) {
		t.Errorf("Expected: true, received: true")
	} else if ys1.Equals(ys4) {
		t.Errorf("Expected: false, received: true")
	} else if ys3.Equals(ys2) {
		t.Errorf("Expected: false, received: true")
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

func TestMonthsContains(t *testing.T) {
	m := Months{time.May, time.June, time.July, time.August}
	if !m.Contains(time.May) {
		t.Errorf("Expected: true, received: false")
	} else if m.Contains(time.April) {
		t.Errorf("Expected: false, received: true")
	}
}

func TestMonthsParse(t *testing.T) {
	m1 := Months{}
	m1.Parse(MetaAny, EmptyString)
	eOut := Months{time.May, time.June, time.July, time.August}
	if m1.Parse("5,6,7,8", FIELDS_SEP); !reflect.DeepEqual(eOut, m1) {
		t.Errorf("Expected: %+v, received: %+v", eOut, m1)
	}
}

func TestMonthsSerialize(t *testing.T) {
	mths := &Months{}
	if rcv := mths.Serialize(INFIELD_SEP); !reflect.DeepEqual(MetaAny, rcv) {
		t.Errorf("Expected: %s, received: %s", MetaAny, rcv)
	}
	mths = &Months{time.January}
	if rcv := mths.Serialize(INFIELD_SEP); !reflect.DeepEqual("1", rcv) {
		t.Errorf("Expected: '1', received: %s", rcv)
	}
	mths = &Months{time.January, time.December}
	if rcv := mths.Serialize(INFIELD_SEP); !reflect.DeepEqual("1;12", rcv) {
		t.Errorf("Expected: '1;12', received: %s", rcv)
	}
}

func TestMonthsIsComplete(t *testing.T) {
	months := Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December}
	if !months.IsComplete() {
		t.Error("Error months IsComplete: ", months)
	}
	months = Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November}
	if months.IsComplete() {
		t.Error("Error months IsComplete: ", months)
	}
}

func TestMonthsEquals(t *testing.T) {
	m1 := Months{time.May, time.June, time.July, time.August}
	m2 := Months{time.May, time.June, time.July, time.August}
	m3 := Months{time.May, time.June, time.July, time.September}
	m4 := Months{}
	if !m1.Equals(m2) {
		t.Errorf("Expected: true, received: fasle")
	} else if m1.Equals(m4) {
		t.Errorf("Expected: false, received: true")
	} else if m3.Equals(m2) {
		t.Errorf("Expected: false, received: true")
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

func TestMonthDaysContains(t *testing.T) {
	md := MonthDays{24, 25, 26}
	if !md.Contains(24) {
		t.Errorf("Expected: true, received: false")
	} else if md.Contains(23) {
		t.Errorf("Expected: false, received: true")
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

func TestMonthDaysSerialize(t *testing.T) {
	md := &MonthDays{}
	if rcv := md.Serialize(INFIELD_SEP); !reflect.DeepEqual(MetaAny, rcv) {
		t.Errorf("Expected: %s, received: %s", MetaAny, rcv)
	}

	md = &MonthDays{1}
	if rcv := md.Serialize(INFIELD_SEP); !reflect.DeepEqual("1", rcv) {
		t.Errorf("Expected: '1', received: %s", rcv)
	}

	md = &MonthDays{1, 2, 3, 4, 5}
	if rcv := md.Serialize(INFIELD_SEP); !reflect.DeepEqual("1;2;3;4;5", rcv) {
		t.Errorf("Expected: '1;2;3;4;5', received: %s", rcv)
	}
}

func TestMonthDaysEquals(t *testing.T) {
	md1 := MonthDays{24, 25, 26}
	md2 := MonthDays{24, 25, 26}
	md3 := MonthDays{24, 25, 27}
	md4 := MonthDays{}
	if !md1.Equals(md2) {
		t.Errorf("Expected: true, received: false")
	} else if md1.Equals(md4) {
		t.Errorf("Expected: false, received: true")
	} else if md1.Equals(md3) {
		t.Errorf("Expected: false, received: true")
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

func TestWeekDaysContains(t *testing.T) {
	wds := WeekDays{time.Monday, time.Tuesday}
	if !wds.Contains(time.Monday) {
		t.Errorf("Expected: true, received: false")
	} else if wds.Contains(time.Wednesday) {
		t.Errorf("Expected: false, received: true")
	}
}

func TestWeekDaysParse(t *testing.T) {
	wd := WeekDays{}
	wd.Parse(MetaAny, EmptyString)
	eOut := WeekDays{time.Monday, time.Tuesday, time.Wednesday}
	wd.Parse("1,2,3", FIELDS_SEP)
	if !reflect.DeepEqual(eOut, wd) {
		t.Errorf("Expected: %+v, received: %+v", eOut, wd)
	}
}

func TestWeekDaysSerialize(t *testing.T) {
	wd := &WeekDays{}
	if rcv := wd.Serialize(INFIELD_SEP); !reflect.DeepEqual(MetaAny, rcv) {
		t.Errorf("Expected: %s, received: %s", MetaAny, rcv)
	}

	wd = &WeekDays{time.Monday}
	if rcv := wd.Serialize(INFIELD_SEP); !reflect.DeepEqual("1", rcv) {
		t.Errorf("Expected: '1', received: %s", rcv)
	}

	wd = &WeekDays{time.Monday, time.Saturday, time.Sunday}
	if rcv := wd.Serialize(INFIELD_SEP); !reflect.DeepEqual("1;6;0", rcv) {
		t.Errorf("Expected: '1;6;0', received: %s", rcv)
	}
}

func TestWeekDaysEquals(t *testing.T) {
	wd1 := WeekDays{time.Monday, time.Saturday, time.Sunday}
	wd2 := WeekDays{time.Monday, time.Saturday, time.Sunday}
	wd3 := WeekDays{time.Monday, time.Saturday, time.Tuesday}
	wd4 := WeekDays{time.Monday}
	if !wd1.Equals(wd2) {
		t.Errorf("Expected: true, received: false")
	} else if wd1.Equals(wd3) {
		t.Errorf("Expected: false, received: true")
	} else if wd1.Equals(wd4) {
		t.Errorf("Expected: false, received: true")
	}
}

func TestDaysInMonth(t *testing.T) {
	if rcv := DaysInMonth(2016, 4); rcv != 30 {
		t.Errorf("Expected: %v, received: %v ", 30, rcv)
	}
	if rcv := DaysInMonth(2016, 2); rcv != 29 {
		t.Errorf("Expected: %v, received: %v ", 29, rcv)
	}
	if rcv := DaysInMonth(2016, 1); rcv != 31 {
		t.Errorf("Expected: %v, received: %v ", 31, rcv)
	}
	if rcv := DaysInMonth(2016, 12); rcv != 31 {
		t.Errorf("Expected: %v, received: %v ", 31, rcv)
	}
	if rcv := DaysInMonth(2015, 2); rcv != 28 {
		t.Errorf("Expected: %v, received: %v ", 28, rcv)
	}
}

func TestDaysInYear(t *testing.T) {
	if rcv := DaysInYear(2016); rcv != 366 {
		t.Errorf("Expected: %v, received: %v ", 366, rcv)
	}
	if rcv := DaysInYear(2015); rcv != 365 {
		t.Errorf("Expected: %v, received: %v ", 265, rcv)
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
