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
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewPartialFlatstoreRecord(record []string, timezone string) (*PartialFlatstoreRecord, error) {
	if len(record) < 7 {
		return nil, errors.New("MISSING_IE")
	}
	pr := &PartialFlatstoreRecord{Method: record[0], AccId: record[3] + record[1] + record[2], Values: record}
	var err error
	if pr.Timestamp, err = utils.ParseTimeDetectLayout(record[6], timezone); err != nil {
		return nil, err
	}
	return pr, nil
}

// This is a partial record received from Flatstore, can be INVITE or BYE and it needs to be paired in order to produce duration
type PartialFlatstoreRecord struct {
	Method    string    // INVITE or BYE
	AccId     string    // Copute here the AccId
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

func NewPartialRecordsCache(ttl time.Duration, cdrOutDir string, csvSep rune) (*PartialRecordsCache, error) {
	return &PartialRecordsCache{ttl: ttl, cdrOutDir: cdrOutDir, csvSep: csvSep,
		partialRecords: make(map[string]map[string]*PartialFlatstoreRecord), guard: engine.NewGuardianLock()}, nil
}

type PartialRecordsCache struct {
	ttl            time.Duration
	cdrOutDir      string
	csvSep         rune
	partialRecords map[string]map[string]*PartialFlatstoreRecord // [FileName"][AccId]*PartialRecord
	guard          *engine.GuardianLock
}

// Dumps the cache into a .unpaired file in the outdir and cleans cache after
func (self *PartialRecordsCache) dumpUnpairedRecords(fileName string) error {
	_, err := self.guard.Guard(func() (interface{}, error) {
		if len(self.partialRecords[fileName]) != 0 { // Only write the file if there are records in the cache
			unpairedFilePath := path.Join(self.cdrOutDir, fileName+UNPAIRED_SUFFIX)
			fileOut, err := os.Create(unpairedFilePath)
			if err != nil {
				engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed creating %s, error: %s", unpairedFilePath, err.Error()))
				return nil, err
			}
			csvWriter := csv.NewWriter(fileOut)
			csvWriter.Comma = self.csvSep
			for _, pr := range self.partialRecords[fileName] {
				if err := csvWriter.Write(pr.Values); err != nil {
					engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed writing unpaired record %v to file: %s, error: %s", pr, unpairedFilePath, err.Error()))
					return nil, err
				}
			}
			csvWriter.Flush()
		}
		delete(self.partialRecords, fileName)
		return nil, nil
	}, 0, fileName)
	return err
}

// Search in cache and return the partial record with accountind id defined, prefFilename is searched at beginning because of better match probability
func (self *PartialRecordsCache) GetPartialRecord(accId, prefFileName string) (string, *PartialFlatstoreRecord) {
	var cachedFilename string
	var cachedPartial *PartialFlatstoreRecord
	checkCachedFNames := []string{prefFileName} // Higher probability to match as firstFileName
	for fName := range self.partialRecords {
		if fName != prefFileName {
			checkCachedFNames = append(checkCachedFNames, fName)
		}
	}
	for _, fName := range checkCachedFNames { // Need to lock them individually
		self.guard.Guard(func() (interface{}, error) {
			var hasPartial bool
			if cachedPartial, hasPartial = self.partialRecords[fName][accId]; hasPartial {
				cachedFilename = fName
			}
			return nil, nil
		}, 0, fName)
		if cachedPartial != nil {
			break
		}
	}
	return cachedFilename, cachedPartial
}

func (self *PartialRecordsCache) CachePartial(fileName string, pr *PartialFlatstoreRecord) {
	self.guard.Guard(func() (interface{}, error) {
		if fileMp, hasFile := self.partialRecords[fileName]; !hasFile {
			self.partialRecords[fileName] = map[string]*PartialFlatstoreRecord{pr.AccId: pr}
			if self.ttl != 0 { // Schedule expiry/dump of the just created entry in cache
				go func() {
					time.Sleep(self.ttl)
					self.dumpUnpairedRecords(fileName)
				}()
			}
		} else if _, hasAccId := fileMp[pr.AccId]; !hasAccId {
			self.partialRecords[fileName][pr.AccId] = pr
		}
		return nil, nil
	}, 0, fileName)
}

func (self *PartialRecordsCache) UncachePartial(fileName string, pr *PartialFlatstoreRecord) {
	self.guard.Guard(func() (interface{}, error) {
		delete(self.partialRecords[fileName], pr.AccId) // Remove the record out of cache
		return nil, nil
	}, 0, fileName)
}

func NewCsvRecordsProcessor(csvReader *csv.Reader, cdrFormat, timezone, fileName, failedCallsPrefix string,
	cdrSourceIds []string, duMultiplyFactors []float64, cdrFilters []utils.RSRFields, cdrFields [][]*config.CfgCdrField,
	httpSkipTlsCheck bool, partialRecordsCache *PartialRecordsCache) *CsvRecordsProcessor {
	return &CsvRecordsProcessor{csvReader: csvReader, cdrFormat: cdrFormat, timezone: timezone, fileName: fileName,
		failedCallsPrefix: failedCallsPrefix, cdrSourceIds: cdrSourceIds,
		duMultiplyFactors: duMultiplyFactors, cdrFilters: cdrFilters, cdrFields: cdrFields,
		httpSkipTlsCheck: httpSkipTlsCheck, partialRecordsCache: partialRecordsCache}

}

type CsvRecordsProcessor struct {
	csvReader           *csv.Reader
	cdrFormat           string
	timezone            string // Timezone for CDRs which are not clearly specifying it
	fileName            string
	failedCallsPrefix   string
	processedRecordsNr  int64    // Number of content records in file
	cdrSourceIds        []string // Should be in sync with cdrFields on indexes
	duMultiplyFactors   []float64
	cdrFilters          []utils.RSRFields       // Should be in sync with cdrFields on indexes
	cdrFields           [][]*config.CfgCdrField // Profiles directly connected with cdrFilters
	httpSkipTlsCheck    bool
	partialRecordsCache *PartialRecordsCache // Shared by cdrc so we can cache for all files in a folder
}

func (self *CsvRecordsProcessor) ProcessedRecordsNr() int64 {
	return self.processedRecordsNr
}

func (self *CsvRecordsProcessor) ProcessNextRecord() ([]*engine.StoredCdr, error) {
	record, err := self.csvReader.Read()
	if err != nil {
		return nil, err
	}
	self.processedRecordsNr += 1
	if utils.IsSliceMember([]string{utils.KAM_FLATSTORE, utils.OSIPS_FLATSTORE}, self.cdrFormat) {
		if record, err = self.processPartialRecord(record); err != nil {
			return nil, err
		} else if record == nil {
			return nil, nil // Due to partial, none returned
		}
	}
	// Record was overwriten with complete information out of cache
	return self.processRecord(record)
}

// Processes a single partial record for flatstore CDRs
func (self *CsvRecordsProcessor) processPartialRecord(record []string) ([]string, error) {
	if strings.HasPrefix(self.fileName, self.failedCallsPrefix) { // Use the first index since they should be the same in all configs
		record = append(record, "0") // Append duration 0 for failed calls flatstore CDR and do not process it further
		return record, nil
	}
	pr, err := NewPartialFlatstoreRecord(record, self.timezone)
	if err != nil {
		return nil, err
	}
	// Retrieve and complete the record from cache
	cachedFilename, cachedPartial := self.partialRecordsCache.GetPartialRecord(pr.AccId, self.fileName)
	if cachedPartial == nil { // Not cached, do it here and stop processing
		self.partialRecordsCache.CachePartial(self.fileName, pr)
		return nil, nil
	}
	pairedRecord, err := pairToRecord(cachedPartial, pr)
	if err != nil {
		return nil, err
	}
	self.partialRecordsCache.UncachePartial(cachedFilename, pr)
	return pairedRecord, nil
}

// Takes the record from a slice and turns it into StoredCdrs, posting them to the cdrServer
func (self *CsvRecordsProcessor) processRecord(record []string) ([]*engine.StoredCdr, error) {
	recordCdrs := make([]*engine.StoredCdr, 0) // More CDRs based on the number of filters and field templates
	for idx := range self.cdrFields {          // cdrFields coming from more templates will produce individual storCdr records
		// Make sure filters are matching
		filterBreak := false
		for _, rsrFilter := range self.cdrFilters[idx] {
			if rsrFilter == nil { // Nil filter does not need to match anything
				continue
			}
			if cfgFieldIdx, _ := strconv.Atoi(rsrFilter.Id); len(record) <= cfgFieldIdx {
				return nil, fmt.Errorf("Ignoring record: %v - cannot compile filter %+v", record, rsrFilter)
			} else if !rsrFilter.FilterPasses(record[cfgFieldIdx]) {
				filterBreak = true
				break
			}
		}
		if filterBreak { // Stop importing cdrc fields profile due to non matching filter
			continue
		}
		if storedCdr, err := self.recordToStoredCdr(record, idx); err != nil {
			return nil, fmt.Errorf("Failed converting to StoredCdr, error: %s", err.Error())
		} else {
			recordCdrs = append(recordCdrs, storedCdr)
		}
	}
	return recordCdrs, nil
}

// Takes the record out of csv and turns it into storedCdr which can be processed by CDRS
func (self *CsvRecordsProcessor) recordToStoredCdr(record []string, cfgIdx int) (*engine.StoredCdr, error) {
	storedCdr := &engine.StoredCdr{CdrHost: "0.0.0.0", CdrSource: self.cdrSourceIds[cfgIdx], ExtraFields: make(map[string]string), Cost: -1}
	var err error
	var lazyHttpFields []*config.CfgCdrField
	for _, cdrFldCfg := range self.cdrFields[cfgIdx] {
		if utils.IsSliceMember([]string{utils.KAM_FLATSTORE, utils.OSIPS_FLATSTORE}, self.cdrFormat) { // Hardcode some values in case of flatstore
			switch cdrFldCfg.CdrFieldId {
			case utils.ACCID:
				cdrFldCfg.Value = utils.ParseRSRFieldsMustCompile("3;1;2", utils.INFIELD_SEP) // in case of flatstore, accounting id is made up out of callid, from_tag and to_tag
			case utils.USAGE:
				cdrFldCfg.Value = utils.ParseRSRFieldsMustCompile(strconv.Itoa(len(record)-1), utils.INFIELD_SEP) // in case of flatstore, last element will be the duration computed by us
			}

		}
		var fieldVal string
		if cdrFldCfg.Type == utils.CDRFIELD {
			for _, cfgFieldRSR := range cdrFldCfg.Value {
				if cfgFieldRSR.IsStatic() {
					fieldVal += cfgFieldRSR.ParseValue("")
				} else { // Dynamic value extracted using index
					if cfgFieldIdx, _ := strconv.Atoi(cfgFieldRSR.Id); len(record) <= cfgFieldIdx {
						return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s", record, cdrFldCfg.Tag)
					} else {
						fieldVal += cfgFieldRSR.ParseValue(record[cfgFieldIdx])
					}
				}
			}
		} else if cdrFldCfg.Type == utils.HTTP_POST {
			lazyHttpFields = append(lazyHttpFields, cdrFldCfg) // Will process later so we can send an estimation of storedCdr to http server
		} else {
			return nil, fmt.Errorf("Unsupported field type: %s", cdrFldCfg.Type)
		}
		if err := populateStoredCdrField(storedCdr, cdrFldCfg.CdrFieldId, fieldVal, self.timezone); err != nil {
			return nil, err
		}
	}
	storedCdr.CgrId = utils.Sha1(storedCdr.AccId, storedCdr.SetupTime.UTC().String())
	if storedCdr.TOR == utils.DATA && self.duMultiplyFactors[cfgIdx] != 0 {
		storedCdr.Usage = time.Duration(float64(storedCdr.Usage.Nanoseconds()) * self.duMultiplyFactors[cfgIdx])
	}
	for _, httpFieldCfg := range lazyHttpFields { // Lazy process the http fields
		var outValByte []byte
		var fieldVal, httpAddr string
		for _, rsrFld := range httpFieldCfg.Value {
			httpAddr += rsrFld.ParseValue("")
		}
		if outValByte, err = utils.HttpJsonPost(httpAddr, self.httpSkipTlsCheck, storedCdr); err != nil && httpFieldCfg.Mandatory {
			return nil, err
		} else {
			fieldVal = string(outValByte)
			if len(fieldVal) == 0 && httpFieldCfg.Mandatory {
				return nil, fmt.Errorf("MandatoryIeMissing: Empty result for http_post field: %s", httpFieldCfg.Tag)
			}
			if err := populateStoredCdrField(storedCdr, httpFieldCfg.CdrFieldId, fieldVal, self.timezone); err != nil {
				return nil, err
			}
		}
	}
	return storedCdr, nil
}
