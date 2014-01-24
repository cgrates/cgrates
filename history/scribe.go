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
	"encoding/json"
	"io"
	"reflect"
	"sort"
	"strings"
)

type Scribe interface {
	Record(Record, *int) error
}

type Record interface {
	GetId() string
}

type records []Record

var (
	recordsMap  = make(map[string]records)
	filenameMap = make(map[reflect.Type]string)
)

func (rs records) Len() int {
	return len(rs)
}

func (rs records) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (rs records) Less(i, j int) bool {
	return rs[i].GetId() < rs[j].GetId()
}

func (rs records) Sort() {
	sort.Sort(rs)
}

func (rs records) SetOrAdd(rec Record) records {
	//rs.Sort()
	n := len(rs)
	i := sort.Search(n, func(i int) bool { return rs[i].GetId() >= rec.GetId() })
	if i < n && rs[i].GetId() == rec.GetId() {
		rs[i] = rec
	} else {
		// i is the index where it would be inserted.
		rs = append(rs, nil)
		copy(rs[i+1:], rs[i:])
		rs[i] = rec
	}
	return rs
}

/*func (rs records) SetOrAdd(rec Record) records {
	found := false
	for _, r := range rs {
		if r.GetId() == rec.GetId() {
			found = true
			r.Object = rec.Object
			return rs
		}
	}
	if !found {
		rs = append(rs, rec)
	}
	return rs
}*/

func format(b io.Writer, recs records) error {
	recs.Sort()
	b.Write([]byte("["))
	for i, r := range recs {
		src, err := json.Marshal(r)
		if err != nil {
			return err
		}
		b.Write(src)
		if i < len(recs)-1 {
			b.Write([]byte(",\n"))
		}
	}
	b.Write([]byte("]"))
	return nil
}

func GetRFN(rec Record) string {
	if fn, ok := filenameMap[reflect.TypeOf(rec)]; ok {
		return fn
	} else {
		typ := reflect.TypeOf(rec)
		typeSegments := strings.Split(typ.String(), ".")
		fn = strings.ToLower(typeSegments[len(typeSegments)-1]) + "s.json"
		filenameMap[typ] = fn
		recordsMap[fn] = make(records, 0)
		return fn
	}
}

var RegisterRecordFilename = GetRFN // will create a key in filename and records map
