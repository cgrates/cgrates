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
	"errors"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewPartialFlatstoreRecord(record []string, timezone string) (*PartialFlatstoreRecord, error) {
	if len(record) < 7 {
		return nil, errors.New("MISSING_IE")
	}
	pr := &PartialFlatstoreRecord{Method: record[0], OriginID: record[3] + record[1] + record[2], Values: record}
	var err error
	if pr.Timestamp, err = utils.ParseTimeDetectLayout(record[6], timezone); err != nil {
		return nil, err
	}
	return pr, nil
}

// This is a partial record received from Flatstore, can be INVITE or BYE and it needs to be paired in order to produce duration
type PartialFlatstoreRecord struct {
	Method    string    // INVITE or BYE
	OriginID  string    // Copute here the OriginID
	Timestamp time.Time // Timestamp of the event, as written by db_flastore module
	Values    []string  // Can contain original values or updated via UpdateValues
}

// Pairs INVITE and BYE into final record containing as last element the duration
func pairToRecord(part1, part2 *PartialFlatstoreRecord) ([]string, error) {
	var invite, bye *PartialFlatstoreRecord
	if part1.Method == "INVITE" {
		invite = part1
	} else if part2.Method == "INVITE" {
		invite = part2
	} else {
		return nil, errors.New("MISSING_INVITE")
	}
	if part1.Method == "BYE" {
		bye = part1
	} else if part2.Method == "BYE" {
		bye = part2
	} else {
		return nil, errors.New("MISSING_BYE")
	}
	if len(invite.Values) != len(bye.Values) {
		return nil, errors.New("INCONSISTENT_VALUES_LENGTH")
	}
	record := invite.Values
	for idx := range record {
		switch idx {
		case 0, 1, 2, 3, 6: // Leave these values as they are
		case 4, 5:
			record[idx] = bye.Values[idx] // Update record with status from bye
		default:
			if bye.Values[idx] != "" { // Any value higher than 6 is dynamically inserted, overwrite if non empty
				record[idx] = bye.Values[idx]
			}

		}
	}
	callDur := bye.Timestamp.Sub(invite.Timestamp)
	record = append(record, strconv.FormatFloat(callDur.Seconds(), 'f', -1, 64))
	return record, nil
}

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
