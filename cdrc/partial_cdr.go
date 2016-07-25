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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type PartialCDRRecord struct {
	cdrs            []*engine.CDR         // Number of CDRs
	cacheDumpFields []*config.CfgCdrField // Fields template to use when dumping from cache on disk
}

// Part of sort interface
func (partCDR *PartialCDRRecord) Len() int {
	return len(partCDR.cdrs)
}

// Part of sort interface
func (partCDR *PartialCDRRecord) Less(i, j int) bool {
	return partCDR.cdrs[i].OrderID < partCDR.cdrs[j].OrderID
}

// Part of sort interface
func (partCDR *PartialCDRRecord) Swap(i, j int) {
	partCDR.cdrs[i], partCDR.cdrs[j] = partCDR.cdrs[j], partCDR.cdrs[i]
}

// Orders CDRs and merge them into one final
func (partCDR *PartialCDRRecord) MergeCDRs() *engine.CDR {
	sort.Sort(partCDR)
	if len(partCDR.cdrs) == 0 {
		return nil
	}
	retCdr := partCDR.cdrs[0].Clone()            // Make sure we don't work on original data
	retCdrRVal := reflect.ValueOf(retCdr).Elem() // So we can set it's fields using reflect
	for idx, cdr := range partCDR.cdrs {
		if idx == 0 { // First CDR is not merged
			continue
		}
		cdrRVal := reflect.ValueOf(cdr).Elem()
		for i := 0; i < cdrRVal.NumField(); i++ { // Find out fields which were modified from previous CDR
			fld := cdrRVal.Field(i)
			var updated bool
			switch v := fld.Interface().(type) {
			case string:
				if v != "" {
					updated = true
				}
			case int64:
				if v != 0 {
					updated = true
				}
			case float64:
				if v != 0.0 {
					updated = true
				}
			case bool:
				if v || cdrRVal.Type().Field(i).Name == utils.PartialField { // Partial field is always updated, even if false
					updated = true
				}
			case time.Time:
				nilTime := time.Time{}
				if v != nilTime {
					updated = true
				}
			case time.Duration:
				if v != time.Duration(0) {
					updated = true
				}
			case map[string]string:
				for fldName, fldVal := range v {
					if origVal, hasIt := retCdr.ExtraFields[fldName]; !hasIt || origVal != fldVal {
						retCdr.ExtraFields[fldName] = fldVal
					}
				}
				//default:
				//	fmt.Printf("Unhandled FieldName: %s, Kind: %+v\n", cdrRVal.Type().Field(i).Name, fld.Kind())
			}
			if updated {
				retCdrRVal.Field(i).Set(fld)
			}
		}
	}
	return retCdr
}
