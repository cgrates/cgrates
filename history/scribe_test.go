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
package history

import (
	"strconv"
	"testing"
)

func TestHistorySet(t *testing.T) {
	rs := records{&Record{Id: "first"}}
	second := &Record{Id: "first"}
	rs.Modify(second)
	if len(rs) != 1 || rs[0] != second {
		t.Error("error setting new value: ", rs[0])
	}
}

func TestHistoryAdd(t *testing.T) {
	rs := records{&Record{Id: "first"}}
	second := &Record{Id: "second"}
	rs = rs.Modify(second)
	if len(rs) != 2 || rs[1] != second {
		t.Error("error setting new value: ", rs)
	}
}

func TestHistoryRemove(t *testing.T) {
	rs := records{&Record{Id: "first"}, &Record{Id: "second"}}
	rs = rs.Modify(&Record{Id: "first", Deleted: true})
	if len(rs) != 1 || rs[0].Id != "second" {
		t.Error("error deleting record: ", rs)
	}
}

func BenchmarkModify(b *testing.B) {
	var rs records
	for i := 0; i < 1000; i++ {
		rs = rs.Modify(&Record{Id: strconv.Itoa(i)})
	}
	for i := 0; i < b.N; i++ {
		rs.Modify(&Record{Id: "400"})
	}
}
