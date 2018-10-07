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
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

func NewUnpairedRecordsCache(ttl time.Duration, cdrOutDir string, csvSep rune) (*UnpairedRecordsCache, error) {
	return &UnpairedRecordsCache{ttl: ttl, cdrOutDir: cdrOutDir, csvSep: csvSep,
		partialRecords: make(map[string]map[string]*UnpairedRecord), guard: guardian.Guardian}, nil
}

type UnpairedRecordsCache struct {
	ttl            time.Duration
	cdrOutDir      string
	csvSep         rune
	partialRecords map[string]map[string]*UnpairedRecord // [FileName"][OriginID]*PartialRecord
	guard          *guardian.GuardianLocker
}

// Dumps the cache into a .unpaired file in the outdir and cleans cache after
func (self *UnpairedRecordsCache) dumpUnpairedRecords(fileName string) error {
	_, err := self.guard.Guard(func() (interface{}, error) {
		if len(self.partialRecords[fileName]) != 0 { // Only write the file if there are records in the cache
			unpairedFilePath := path.Join(self.cdrOutDir, fileName+UNPAIRED_SUFFIX)
			fileOut, err := os.Create(unpairedFilePath)
			if err != nil {
				utils.Logger.Err(fmt.Sprintf("<Cdrc> Failed creating %s, error: %s", unpairedFilePath, err.Error()))
				return nil, err
			}
			csvWriter := csv.NewWriter(fileOut)
			csvWriter.Comma = self.csvSep
			for _, pr := range self.partialRecords[fileName] {
				if err := csvWriter.Write(pr.Values); err != nil {
					utils.Logger.Err(fmt.Sprintf("<Cdrc> Failed writing unpaired record %v to file: %s, error: %s", pr, unpairedFilePath, err.Error()))
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
func (self *UnpairedRecordsCache) GetPartialRecord(OriginID, prefFileName string) (string, *UnpairedRecord) {
	var cachedFilename string
	var cachedPartial *UnpairedRecord
	checkCachedFNames := []string{prefFileName} // Higher probability to match as firstFileName
	for fName := range self.partialRecords {
		if fName != prefFileName {
			checkCachedFNames = append(checkCachedFNames, fName)
		}
	}
	for _, fName := range checkCachedFNames { // Need to lock them individually
		self.guard.Guard(func() (interface{}, error) {
			var hasPartial bool
			if cachedPartial, hasPartial = self.partialRecords[fName][OriginID]; hasPartial {
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

func (self *UnpairedRecordsCache) CachePartial(fileName string, pr *UnpairedRecord) {
	self.guard.Guard(func() (interface{}, error) {
		if fileMp, hasFile := self.partialRecords[fileName]; !hasFile {
			self.partialRecords[fileName] = map[string]*UnpairedRecord{pr.OriginID: pr}
			if self.ttl != 0 { // Schedule expiry/dump of the just created entry in cache
				go func() {
					time.Sleep(self.ttl)
					self.dumpUnpairedRecords(fileName)
				}()
			}
		} else if _, hasOriginID := fileMp[pr.OriginID]; !hasOriginID {
			self.partialRecords[fileName][pr.OriginID] = pr
		}
		return nil, nil
	}, 0, fileName)
}

func (self *UnpairedRecordsCache) UncachePartial(fileName string, pr *UnpairedRecord) {
	self.guard.Guard(func() (interface{}, error) {
		delete(self.partialRecords[fileName], pr.OriginID) // Remove the record out of cache
		return nil, nil
	}, 0, fileName)
}

func NewUnpairedRecord(record []string, timezone string) (*UnpairedRecord, error) {
	if len(record) < 7 {
		return nil, errors.New("MISSING_IE")
	}
	pr := &UnpairedRecord{Method: record[0], OriginID: record[3] + record[1] + record[2], Values: record}
	var err error
	if pr.Timestamp, err = utils.ParseTimeDetectLayout(record[6], timezone); err != nil {
		return nil, err
	}
	return pr, nil
}

// This is a partial record received from Flatstore, can be INVITE or BYE and it needs to be paired in order to produce duration
type UnpairedRecord struct {
	Method    string    // INVITE or BYE
	OriginID  string    // Copute here the OriginID
	Timestamp time.Time // Timestamp of the event, as written by db_flastore module
	Values    []string  // Can contain original values or updated via UpdateValues
}

// Pairs INVITE and BYE into final record containing as last element the duration
func pairToRecord(part1, part2 *UnpairedRecord) ([]string, error) {
	var invite, bye *UnpairedRecord
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
