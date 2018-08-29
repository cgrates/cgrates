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

package cdrc

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"reflect"
	"sort"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

const (
	PartialRecordsSuffix = "partial"
)

func NewPartialRecordsCache(ttl time.Duration, expiryAction string, cdrOutDir string, csvSep rune,
	roundDecimals int, timezone string, httpSkipTlsCheck bool,
	cdrs rpcclient.RpcClientConnection, filterS *engine.FilterS) (*PartialRecordsCache, error) {
	return &PartialRecordsCache{ttl: ttl, expiryAction: expiryAction, cdrOutDir: cdrOutDir,
		csvSep: csvSep, roundDecimals: roundDecimals, timezone: timezone,
		httpSkipTlsCheck: httpSkipTlsCheck, cdrs: cdrs,
		partialRecords: make(map[string]*PartialCDRRecord),
		dumpTimers:     make(map[string]*time.Timer),
		guard:          guardian.Guardian, filterS: filterS}, nil
}

type PartialRecordsCache struct {
	ttl              time.Duration
	expiryAction     string
	cdrOutDir        string
	csvSep           rune
	roundDecimals    int
	timezone         string
	httpSkipTlsCheck bool
	cdrs             rpcclient.RpcClientConnection
	partialRecords   map[string]*PartialCDRRecord // [OriginID]*PartialRecord
	dumpTimers       map[string]*time.Timer       // [OriginID]*time.Timer which can be canceled or reset
	guard            *guardian.GuardianLocker
	filterS          *engine.FilterS
}

// Dumps the cache into a .unpaired file in the outdir and cleans cache after
func (prc *PartialRecordsCache) dumpPartialRecords(originID string) {
	_, err := prc.guard.Guard(func() (interface{}, error) {
		if prc.partialRecords[originID].Len() != 0 { // Only write the file if there are records in the cache
			dumpFilePath := path.Join(prc.cdrOutDir, fmt.Sprintf("%s.%s.%d", originID, PartialRecordsSuffix, time.Now().Unix()))
			fileOut, err := os.Create(dumpFilePath)
			if err != nil {
				utils.Logger.Err(fmt.Sprintf("<Cdrc> Failed creating %s, error: %s", dumpFilePath, err.Error()))
				return nil, err
			}
			csvWriter := csv.NewWriter(fileOut)
			csvWriter.Comma = prc.csvSep
			for _, cdr := range prc.partialRecords[originID].cdrs {
				expRec, err := cdr.AsExportRecord(prc.partialRecords[originID].cacheDumpFields,
					prc.httpSkipTlsCheck, nil, prc.roundDecimals, prc.filterS)
				if err != nil {
					return nil, err
				}
				if err := csvWriter.Write(expRec); err != nil {
					utils.Logger.Err(fmt.Sprintf("<Cdrc> Failed writing partial CDR %v to file: %s, error: %s", cdr, dumpFilePath, err.Error()))
					return nil, err
				}
			}
			csvWriter.Flush()
		}
		delete(prc.partialRecords, originID)
		return nil, nil
	}, 0, originID)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRC> Failed dumping CDR with originID: %s, error: %s", originID, err.Error()))
	}
}

// Called when record expires in cache, will send the CDR merged (forcing it's completion) to the CDRS
func (prc *PartialRecordsCache) postCDR(originID string) {
	_, err := prc.guard.Guard(func() (interface{}, error) {
		if prc.partialRecords[originID].Len() != 0 { // Only write the file if there are records in the cache
			cdr := prc.partialRecords[originID].MergeCDRs()
			cdr.Partial = false // force completion
			var reply string
			if err := prc.cdrs.Call(utils.CdrsV2ProcessCDR, cdr.AsCGREvent(), &reply); err != nil {
				utils.Logger.Err(fmt.Sprintf("<Cdrc> Failed sending CDR  %+v from partial cache, error: %s", cdr, err.Error()))
			} else if reply != utils.OK {
				utils.Logger.Err(fmt.Sprintf("<Cdrc> Received unexpected reply for CDR, %+v, reply: %s", cdr, reply))
			}
		}
		delete(prc.partialRecords, originID)
		return nil, nil
	}, 0, originID)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRC> Failed posting from cache CDR with originID: %s, error: %s", originID, err.Error()))
	}
}

// Called to cache a partial record.
// If exists in cache, CDRs will be updated
// Locking should be handled at higher layer
func (prc *PartialRecordsCache) cachePartialCDR(pCDR *PartialCDRRecord) (*PartialCDRRecord, error) {
	originID := pCDR.cdrs[0].OriginID
	if tmr, hasIt := prc.dumpTimers[originID]; hasIt { // Update existing timer
		tmr.Reset(prc.ttl)
	} else {
		switch prc.expiryAction {
		case utils.MetaDumpToFile:
			prc.dumpTimers[originID] = time.AfterFunc(prc.ttl, func() { prc.dumpPartialRecords(originID) }) // Schedule dumping of the partial CDR
		case utils.MetaPostCDR:
			prc.dumpTimers[originID] = time.AfterFunc(prc.ttl, func() { prc.postCDR(originID) }) // Schedule dumping of the partial CDR
		default:
			return nil, fmt.Errorf("Unsupported PartialCacheExpiryAction: %s", prc.expiryAction)
		}

	}
	if _, hasIt := prc.partialRecords[originID]; !hasIt {
		prc.partialRecords[originID] = pCDR
	} else { // Exists, update it's records
		prc.partialRecords[originID].cdrs = append(prc.partialRecords[originID].cdrs, pCDR.cdrs...)
	}
	return prc.partialRecords[originID], nil
}

// Called to uncache partialCDR and remove automatic dumping of the cached records
func (prc *PartialRecordsCache) uncachePartialCDR(pCDR *PartialCDRRecord) {
	originID := pCDR.cdrs[0].OriginID
	if tmr, hasIt := prc.dumpTimers[originID]; hasIt {
		tmr.Stop()
	}
	delete(prc.partialRecords, originID)
}

// Returns PartialCDR only if merge was possible
func (prc *PartialRecordsCache) MergePartialCDRRecord(pCDR *PartialCDRRecord) (*engine.CDR, error) {
	if pCDR.Len() == 0 || pCDR.cdrs[0].OriginID == "" { // Sanity check
		return nil, nil
	}
	originID := pCDR.cdrs[0].OriginID
	pCDRIf, err := prc.guard.Guard(func() (interface{}, error) {
		if _, hasIt := prc.partialRecords[originID]; !hasIt && pCDR.Len() == 1 && !pCDR.cdrs[0].Partial {
			return pCDR.cdrs[0], nil // Special case when not a partial CDR and not having cached CDRs on same OriginID
		}
		cachedPartialCDR, err := prc.cachePartialCDR(pCDR)
		if err != nil {
			return nil, err
		}
		var final bool
		for _, cdr := range pCDR.cdrs {
			if !cdr.Partial {
				final = true
				break
			}
		}
		if !final {
			return nil, nil
		}
		prc.uncachePartialCDR(cachedPartialCDR)
		return cachedPartialCDR.MergeCDRs(), nil
	}, 0, originID)
	if pCDRIf == nil {
		return nil, err
	}
	return pCDRIf.(*engine.CDR), err
}

func NewPartialCDRRecord(cdr *engine.CDR, cacheDumpFlds []*config.FCTemplate) *PartialCDRRecord {
	return &PartialCDRRecord{cdrs: []*engine.CDR{cdr}, cacheDumpFields: cacheDumpFlds}
}

// PartialCDRRecord is a record which can be updated later
// different from PartialFlatstoreRecordsCache which is incomplete (eg: need to calculate duration out of 2 records)
type PartialCDRRecord struct {
	cdrs            []*engine.CDR        // Number of CDRs
	cacheDumpFields []*config.FCTemplate // Fields template to use when dumping from cache on disk
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
				if v || cdrRVal.Type().Field(i).Name == utils.Partial { // Partial field is always updated, even if false
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
			}
			if updated {
				retCdrRVal.Field(i).Set(fld)
			}
		}
	}
	return retCdr
}
