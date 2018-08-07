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
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func fwvValue(cdrLine string, indexStart, width int, padding string) string {
	rawVal := cdrLine[indexStart : indexStart+width]
	switch padding {
	case "left":
		rawVal = strings.TrimLeft(rawVal, " ")
	case "right":
		rawVal = strings.TrimRight(rawVal, " ")
	case "zeroleft":
		rawVal = strings.TrimLeft(rawVal, "0 ")
	case "zeroright":
		rawVal = strings.TrimRight(rawVal, "0 ")
	}
	return rawVal
}

func NewFwvRecordsProcessor(file *os.File, dfltCfg *config.CdrcConfig, cdrcCfgs []*config.CdrcConfig, httpClient *http.Client,
	httpSkipTlsCheck bool, timezone string, filterS *engine.FilterS) *FwvRecordsProcessor {
	return &FwvRecordsProcessor{file: file, cdrcCfgs: cdrcCfgs, dfltCfg: dfltCfg, httpSkipTlsCheck: httpSkipTlsCheck, timezone: timezone, filterS: filterS}
}

type FwvRecordsProcessor struct {
	file               *os.File
	dfltCfg            *config.CdrcConfig // General parameters
	cdrcCfgs           []*config.CdrcConfig
	httpClient         *http.Client
	httpSkipTlsCheck   bool
	timezone           string
	lineLen            int64       // Length of the line in the file
	offset             int64       // Index of the next byte to process
	processedRecordsNr int64       // Number of content records in file
	trailerOffset      int64       // Index where trailer starts, to be used as boundary when reading cdrs
	headerCdr          *engine.CDR // Cache here the general purpose stored CDR
	filterS            *engine.FilterS
}

// Sets the line length based on first line, sets offset back to initial after reading
func (self *FwvRecordsProcessor) setLineLen() error {
	rdr := bufio.NewReader(self.file)
	readBytes, err := rdr.ReadBytes('\n')
	if err != nil {
		return err
	}
	self.lineLen = int64(len(readBytes))
	if _, err := self.file.Seek(0, 0); err != nil {
		return err
	}
	return nil
}

func (self *FwvRecordsProcessor) ProcessedRecordsNr() int64 {
	return self.processedRecordsNr
}

func (self *FwvRecordsProcessor) ProcessNextRecord() ([]*engine.CDR, error) {
	defer func() { self.offset += self.lineLen }() // Schedule increasing the offset once we are out from processing the record
	if self.offset == 0 {                          // First time, set the necessary offsets
		if err := self.setLineLen(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<Cdrc> Row 0, error: cannot set lineLen: %s", err.Error()))
			return nil, io.EOF
		}
		if len(self.dfltCfg.TrailerFields) != 0 {
			if fi, err := self.file.Stat(); err != nil {
				utils.Logger.Err(fmt.Sprintf("<Cdrc> Row 0, error: cannot get file stats: %s", err.Error()))
				return nil, err
			} else {
				self.trailerOffset = fi.Size() - self.lineLen
			}
		}
		if len(self.dfltCfg.HeaderFields) != 0 { // ToDo: Process here the header fields
			if err := self.processHeader(); err != nil {
				utils.Logger.Err(fmt.Sprintf("<Cdrc> Row 0, error reading header: %s", err.Error()))
				return nil, io.EOF
			}
			return nil, nil
		}
	}
	recordCdrs := make([]*engine.CDR, 0) // More CDRs based on the number of filters and field templates
	if self.trailerOffset != 0 && self.offset >= self.trailerOffset {
		if err := self.processTrailer(); err != nil && err != io.EOF {
			utils.Logger.Err(fmt.Sprintf("<Cdrc> Read trailer error: %s ", err.Error()))
		}
		return nil, io.EOF
	}
	buf := make([]byte, self.lineLen)
	nRead, err := self.file.Read(buf)
	if err != nil {
		return nil, err
	} else if nRead != len(buf) {
		utils.Logger.Err(fmt.Sprintf("<Cdrc> Could not read complete line, have instead: %s", string(buf)))
		return nil, io.EOF
	}
	self.processedRecordsNr += 1
	record := string(buf)
	for _, cdrcCfg := range self.cdrcCfgs {
		if passes := self.recordPassesCfgFilter(record, cdrcCfg); !passes {
			continue
		}
		if storedCdr, err := self.recordToStoredCdr(record, cdrcCfg, cdrcCfg.ID); err != nil {
			return nil, fmt.Errorf("Failed converting to StoredCdr, error: %s", err.Error())
		} else {
			recordCdrs = append(recordCdrs, storedCdr)
		}
		if !cdrcCfg.ContinueOnSuccess { // Successfully executed one config, do not continue for next one
			break
		}
	}
	return recordCdrs, nil
}

func (self *FwvRecordsProcessor) recordPassesCfgFilter(record string, cdrcCfg *config.CdrcConfig) bool {
	for _, rsrFilter := range cdrcCfg.CdrFilter {
		if rsrFilter == nil { // Nil filter does not need to match anything
			continue
		}
		if cfgFieldIdx, _ := strconv.Atoi(rsrFilter.Id); len(record) <= cfgFieldIdx {
			fmt.Errorf("Ignoring record: %v - cannot compile filter %+v", record, rsrFilter)
			return false
		} else if _, err := rsrFilter.Parse(record[cfgFieldIdx:]); err != nil {
			return false
		}
	}
	return true
}

// Converts a record (header or normal) to CDR
func (self *FwvRecordsProcessor) recordToStoredCdr(record string, cdrcCfg *config.CdrcConfig, cfgKey string) (*engine.CDR, error) {
	var err error
	var lazyHttpFields []*config.CfgCdrField
	var cfgFields []*config.CfgCdrField
	var duMultiplyFactor float64
	var storedCdr *engine.CDR
	if self.headerCdr != nil { // Clone the header CDR so we can use it as base to future processing (inherit fields defined there)
		storedCdr = self.headerCdr.Clone()
	} else {
		storedCdr = &engine.CDR{OriginHost: "0.0.0.0", ExtraFields: make(map[string]string), Cost: -1}
	}
	if cfgKey == "*header" {
		cfgFields = cdrcCfg.HeaderFields
		storedCdr.Source = cdrcCfg.CdrSourceId
		duMultiplyFactor = cdrcCfg.DataUsageMultiplyFactor
	} else {
		cfgFields = cdrcCfg.ContentFields
		storedCdr.Source = cdrcCfg.CdrSourceId
		duMultiplyFactor = cdrcCfg.DataUsageMultiplyFactor
	}
	for _, cdrFldCfg := range cfgFields {
		var fieldVal string
		switch cdrFldCfg.Type {
		case utils.META_COMPOSED:
			for _, cfgFieldRSR := range cdrFldCfg.Value {
				if cfgFieldRSR.IsStatic() {
					if parsed, err := cfgFieldRSR.Parse(""); err != nil {
						return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s, err: %s",
							record, cdrFldCfg.Tag, err.Error())
					} else {
						fieldVal += parsed
					}
				} else { // Dynamic value extracted using index
					cfgFieldIdx, _ := strconv.Atoi(cfgFieldRSR.Id)
					if len(record) <= cfgFieldIdx {
						return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s",
							record, cdrFldCfg.Tag)
					}
					if parsed, err := cfgFieldRSR.Parse(fwvValue(record, cfgFieldIdx,
						cdrFldCfg.Width, cdrFldCfg.Padding)); err != nil {
						return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s, err: %s",
							record, cdrFldCfg.Tag, err.Error())
					} else {
						fieldVal += parsed
					}
				}
			}
		case utils.META_HTTP_POST:
			lazyHttpFields = append(lazyHttpFields, cdrFldCfg) // Will process later so we can send an estimation of storedCdr to http server
		default:
			//return nil, fmt.Errorf("Unsupported field type: %s", cdrFldCfg.Type)
			continue // Don't do anything for unsupported fields
		}
		if err := storedCdr.ParseFieldValue(cdrFldCfg.FieldId, fieldVal, self.timezone); err != nil {
			return nil, err
		}
	}
	if storedCdr.CGRID == "" && storedCdr.OriginID != "" && cfgKey != "*header" {
		storedCdr.CGRID = utils.Sha1(storedCdr.OriginID, storedCdr.SetupTime.UTC().String())
	}
	if storedCdr.ToR == utils.DATA && duMultiplyFactor != 0 {
		storedCdr.Usage = time.Duration(float64(storedCdr.Usage.Nanoseconds()) * duMultiplyFactor)
	}
	for _, httpFieldCfg := range lazyHttpFields { // Lazy process the http fields
		var outValByte []byte
		var fieldVal, httpAddr string
		for _, rsrFld := range httpFieldCfg.Value {
			if parsed, err := rsrFld.Parse(""); err != nil {
				return nil, fmt.Errorf("Ignoring record: %v - cannot extract http address, err: %s",
					record, err.Error())
			} else {
				httpAddr += parsed
			}
		}
		var jsn []byte
		jsn, err = json.Marshal(storedCdr)
		if err != nil {
			return nil, err
		}
		if outValByte, err = engine.HttpJsonPost(httpAddr, self.httpSkipTlsCheck, jsn); err != nil && httpFieldCfg.Mandatory {
			return nil, err
		} else {
			fieldVal = string(outValByte)
			if len(fieldVal) == 0 && httpFieldCfg.Mandatory {
				return nil, fmt.Errorf("MandatoryIeMissing: Empty result for http_post field: %s", httpFieldCfg.Tag)
			}
			if err := storedCdr.ParseFieldValue(httpFieldCfg.FieldId, fieldVal, self.timezone); err != nil {
				return nil, err
			}
		}
	}
	return storedCdr, nil
}

func (self *FwvRecordsProcessor) processHeader() error {
	buf := make([]byte, self.lineLen)
	if nRead, err := self.file.Read(buf); err != nil {
		return err
	} else if nRead != len(buf) {
		return fmt.Errorf("In header, line len: %d, have read: %d", self.lineLen, nRead)
	}
	var err error
	if self.headerCdr, err = self.recordToStoredCdr(string(buf), self.dfltCfg, "*header"); err != nil {
		return err
	}
	return nil
}

func (self *FwvRecordsProcessor) processTrailer() error {
	buf := make([]byte, self.lineLen)
	if nRead, err := self.file.ReadAt(buf, self.trailerOffset); err != nil {
		return err
	} else if nRead != len(buf) {
		return fmt.Errorf("In trailer, line len: %d, have read: %d", self.lineLen, nRead)
	}
	return nil
}
