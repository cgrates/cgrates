/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPaginator(t *testing.T) {
	var in []string

	if rcv, err := Paginate(in, 2, 2, 6); err != nil {
		t.Error(err)
	} else if rcv != nil {
		t.Error("expected nil return")
	}

	in = []string{"FLTR_1", "FLTR_2", "FLTR_3", "FLTR_4", "FLTR_5", "FLTR_6", "FLTR_7",
		"FLTR_8", "FLTR_9", "FLTR_10", "FLTR_11", "FLTR_12", "FLTR_13", "FLTR_14",
		"FLTR_15", "FLTR_16", "FLTR_17", "FLTR_18", "FLTR_19", "FLTR_20"}

	exp := []string{"FLTR_7", "FLTR_8"}
	if rcv, err := Paginate(in, 2, 6, 9); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	exp = []string{"FLTR_19", "FLTR_20"}
	if rcv, err := Paginate(in, 0, 18, 50); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	if rcv, err := Paginate(in, 0, 0, 50); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, in) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", in, rcv)
	}

	experr := `SERVER_ERROR: maximum number of items exceeded`
	if _, err := Paginate(in, 0, 0, 19); err == nil || err.Error() != experr {
		t.Error(err)
	}

	if rcv, err := Paginate(in, 25, 18, 50); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	var expOut []string
	if rcv, err := Paginate(in, 2, 22, 50); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expOut) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expOut, rcv)
	}

	if _, err := Paginate(in, 2, 4, 5); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if _, err := Paginate(in, 0, 18, 19); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

}

type pag struct {
	limit    int
	offset   int
	maxItems int
}

func testName(p pag) string {
	return fmt.Sprintf("limit:<%d>, offset:<%d>, maxItems:<%d>", p.limit, p.offset, p.maxItems)
}

func TestPaginatorMaxItems(t *testing.T) {
	in := []string{"FLTR_1", "FLTR_2", "FLTR_3", "FLTR_4", "FLTR_5"}
	experr := "SERVER_ERROR: maximum number of items exceeded"
	cases := []struct {
		p   pag
		err string
	}{
		{pag{limit: 0, offset: 0, maxItems: 1}, experr},
		{pag{limit: 1, offset: 0, maxItems: 1}, ""},
		{pag{limit: 0, offset: 1, maxItems: 1}, experr},
		{pag{limit: 0, offset: 0, maxItems: 2}, experr},
		{pag{limit: 0, offset: 1, maxItems: 2}, experr},
		{pag{limit: 0, offset: 2, maxItems: 2}, experr},
		{pag{limit: 1, offset: 0, maxItems: 2}, ""},
		{pag{limit: 1, offset: 1, maxItems: 2}, ""},
		{pag{limit: 2, offset: 0, maxItems: 2}, ""},
		{pag{limit: 0, offset: 0, maxItems: 3}, experr},
		{pag{limit: 0, offset: 1, maxItems: 3}, experr},
		{pag{limit: 0, offset: 2, maxItems: 3}, experr},
		{pag{limit: 0, offset: 3, maxItems: 3}, experr},
		{pag{limit: 1, offset: 0, maxItems: 3}, ""},
		{pag{limit: 1, offset: 1, maxItems: 3}, ""},
		{pag{limit: 1, offset: 2, maxItems: 3}, ""},
		{pag{limit: 2, offset: 0, maxItems: 3}, ""},
		{pag{limit: 2, offset: 1, maxItems: 3}, ""},
		{pag{limit: 3, offset: 0, maxItems: 3}, ""},
		{pag{limit: 0, offset: 0, maxItems: 4}, experr},
		{pag{limit: 0, offset: 1, maxItems: 4}, experr},
		{pag{limit: 0, offset: 2, maxItems: 4}, experr},
		{pag{limit: 0, offset: 3, maxItems: 4}, experr},
		{pag{limit: 0, offset: 4, maxItems: 4}, experr},
		{pag{limit: 1, offset: 0, maxItems: 4}, ""},
		{pag{limit: 1, offset: 1, maxItems: 4}, ""},
		{pag{limit: 1, offset: 2, maxItems: 4}, ""},
		{pag{limit: 1, offset: 3, maxItems: 4}, ""},
		{pag{limit: 2, offset: 0, maxItems: 4}, ""},
		{pag{limit: 2, offset: 1, maxItems: 4}, ""},
		{pag{limit: 2, offset: 2, maxItems: 4}, ""},
		{pag{limit: 3, offset: 0, maxItems: 4}, ""},
		{pag{limit: 3, offset: 1, maxItems: 4}, ""},
		{pag{limit: 4, offset: 0, maxItems: 4}, ""},
		{pag{limit: 0, offset: 0, maxItems: 5}, ""},
		{pag{limit: 0, offset: 1, maxItems: 5}, ""},
		{pag{limit: 0, offset: 2, maxItems: 5}, ""},
		{pag{limit: 0, offset: 3, maxItems: 5}, ""},
		{pag{limit: 0, offset: 4, maxItems: 5}, ""},
		{pag{limit: 0, offset: 5, maxItems: 5}, ""},
		{pag{limit: 1, offset: 0, maxItems: 5}, ""},
		{pag{limit: 1, offset: 1, maxItems: 5}, ""},
		{pag{limit: 1, offset: 2, maxItems: 5}, ""},
		{pag{limit: 1, offset: 3, maxItems: 5}, ""},
		{pag{limit: 1, offset: 4, maxItems: 5}, ""},
		{pag{limit: 2, offset: 0, maxItems: 5}, ""},
		{pag{limit: 2, offset: 1, maxItems: 5}, ""},
		{pag{limit: 2, offset: 2, maxItems: 5}, ""},
		{pag{limit: 2, offset: 3, maxItems: 5}, ""},
		{pag{limit: 3, offset: 0, maxItems: 5}, ""},
		{pag{limit: 3, offset: 1, maxItems: 5}, ""},
		{pag{limit: 3, offset: 2, maxItems: 5}, ""},
		{pag{limit: 4, offset: 0, maxItems: 5}, ""},
		{pag{limit: 4, offset: 1, maxItems: 5}, ""},
		{pag{limit: 5, offset: 0, maxItems: 5}, ""},
	}

	for _, c := range cases {
		t.Run(testName(c.p), func(t *testing.T) {
			_, err := Paginate(in, c.p.limit, c.p.offset, c.p.maxItems)
			if err != nil {
				if c.err == "" {
					t.Error("did not expect error")
				}
			} else if c.err != "" {
				t.Errorf("expected error")
			}
		})
	}
}

func TestAPITPDataGetPaginateOpts(t *testing.T) {
	opts := map[string]any{
		PageLimitOpt:    1.3,
		PageOffsetOpt:   4,
		PageMaxItemsOpt: "5",
	}

	if limit, offset, maxItems, err := GetPaginateOpts(opts); err != nil {
		t.Error(err)
	} else if limit != 1 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 1, limit)
	} else if offset != 4 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 4, offset)
	} else if maxItems != 5 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 5, maxItems)
	}

	opts[PageMaxItemsOpt] = false
	experr := `cannot convert field<bool>: false to int`
	if _, _, _, err := GetPaginateOpts(opts); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	opts[PageOffsetOpt] = struct{}{}
	experr = `cannot convert field<struct {}>: {} to int`
	if _, _, _, err := GetPaginateOpts(opts); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	opts[PageLimitOpt] = true
	experr = `cannot convert field<bool>: true to int`
	if _, _, _, err := GetPaginateOpts(opts); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestPaginatorClone(t *testing.T) {
	pgnt := Paginator{
		Limit:    IntPointer(3),
		Offset:   IntPointer(1),
		MaxItems: IntPointer(7),
	}
	if rcv := pgnt.Clone(); !reflect.DeepEqual(pgnt, rcv) {
		t.Errorf("expected %v, received %v", pgnt, rcv)
	}
}
