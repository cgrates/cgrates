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

package history

import (
	"strconv"
	"testing"
)

func TestHistorySet(t *testing.T) {
	rs := records{&Record{"first", "test"}}
	rs.SetOrAdd(&Record{"first", "new value"})
	if len(rs) != 1 || rs[0].Object != "new value" {
		t.Error("error setting new value: ", rs[0])
	}
}

func TestHistoryAdd(t *testing.T) {
	rs := records{&Record{"first", "test"}}
	rs = rs.SetOrAdd(&Record{"second", "new value"})
	if len(rs) != 2 || rs[1].Object != "new value" {
		t.Error("error setting new value: ", rs)
	}
}

func BenchmarkSetOrAdd(b *testing.B) {
	var rs records
	for i := 0; i < 1000; i++ {
		rs = rs.SetOrAdd(&Record{strconv.Itoa(i), strconv.Itoa(i)})
	}
	for i := 0; i < b.N; i++ {
		rs.SetOrAdd(&Record{"400", "test"})
	}
}

func BenchmarkSetOrAddOld(b *testing.B) {
	var rs records
	for i := 0; i < 1000; i++ {
		rs = rs.SetOrAddOld(&Record{strconv.Itoa(i), strconv.Itoa(i)})
	}
	for i := 0; i < b.N; i++ {
		rs.SetOrAddOld(&Record{"400", "test"})
	}
}
