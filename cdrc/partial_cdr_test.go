/*
Real-time Charging System for Telecom & ISP environments
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

package cdrc

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestPartialCDRRecordSort(t *testing.T) {
	cdrsRaw := []*engine.CDR{&engine.CDR{OrderID: 3}, &engine.CDR{OrderID: 1}, &engine.CDR{OrderID: 2}}
	pCdr := &PartialCDRRecord{cdrs: cdrsRaw}
	sort.Sort(pCdr)
	cdrsO := []*engine.CDR{&engine.CDR{OrderID: 1}, &engine.CDR{OrderID: 2}, &engine.CDR{OrderID: 3}}
	if !reflect.DeepEqual(cdrsO, pCdr.cdrs) {
		t.Errorf("Expecting: %+v, received: %+v", cdrsO, pCdr.cdrs)
	}
}

func TestPartialCDRRecordMergeCDRs(t *testing.T) {
	cdr1 := &engine.CDR{OrderID: 1, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: "TestPartialCDRRecordMergeCDRs", RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), Partial: true,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	cdr2 := &engine.CDR{OrderID: 3, Partial: false,
		ExtraFields: map[string]string{"disconnect_direction": "upstream"},
		Usage:       time.Duration(62 * time.Second),
	}
	cdr3 := &engine.CDR{OrderID: 2, Partial: true,
		ExtraFields: map[string]string{"field_extr1": "val_extr11"},
		AnswerTime:  time.Date(2013, 11, 7, 8, 43, 0, 0, time.UTC),
		Usage:       time.Duration(30 * time.Second),
	}
	pCdr := &PartialCDRRecord{cdrs: []*engine.CDR{cdr1, cdr2, cdr3}}
	eCDR := &engine.CDR{OrderID: 3, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: "TestPartialCDRRecordMergeCDRs", RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 43, 0, 0, time.UTC), Partial: false,
		Usage:       time.Duration(62 * time.Second),
		ExtraFields: map[string]string{"field_extr1": "val_extr11", "fieldextr2": "valextr2", "disconnect_direction": "upstream"},
	}
	if mCdr := pCdr.MergeCDRs(); !reflect.DeepEqual(eCDR, mCdr) {
		t.Errorf("Expecting: %+v, received: %+v", eCDR, mCdr)
	}
}
