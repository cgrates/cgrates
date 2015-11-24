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
	"bufio"
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

func NewFwvRecordsProcessor(file *os.File, dfltCfg *config.CdrcConfig, cdrcCfgs map[string]*config.CdrcConfig, httpClient *http.Client, httpSkipTlsCheck bool, timezone string) *FwvRecordsProcessor {
	return &FwvRecordsProcessor{file: file, cdrcCfgs: cdrcCfgs, dfltCfg: dfltCfg, httpSkipTlsCheck: httpSkipTlsCheck, timezone: timezone}
}

type FwvRecordsProcessor struct {
	file               *os.File
	dfltCfg            *config.CdrcConfig // General parameters
	cdrcCfgs           map[string]*config.CdrcConfig
	httpClient         *http.Client
	httpSkipTlsCheck   bool
	timezone           string
	lineLen            int64             // Length of the line in the file
	offset             int64             // Index of the next byte to process
	processedRecordsNr int64             // Number of content records in file
	trailerOffset      int64             // Index where trailer starts, to be used as boundary when reading cdrs
	headerCdr          *engine.StoredCdr // Cache here the general purpose stored CDR
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

func (self *FwvRecordsProcessor) ProcessNextRecord() ([]*engine.StoredCdr, error) {
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
	recordCdrs := make([]*engine.StoredCdr, 0) // More CDRs based on the number of filters and field templates
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
	for cfgKey, cdrcCfg := range self.cdrcCfgs {
		if passes := self.recordPassesCfgFilter(record, cfgKey); !passes {
			continue
		}
		if storedCdr, err := self.recordToStoredCdr(record, cfgKey); err != nil {
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

func (self *FwvRecordsProcessor) recordPassesCfgFilter(record, configKey string) bool {
	filterPasses := true
	for _, rsrFilter := range self.cdrcCfgs[configKey].CdrFilter {
		if rsrFilter == nil { // Nil filter does not need to match anything
			continue
		}
		if cfgFieldIdx, _ := strconv.Atoi(rsrFilter.Id); len(record) <= cfgFieldIdx {
			fmt.Errorf("Ignoring record: %v - cannot compile filter %+v", record, rsrFilter)
			return false
		} else if !rsrFilter.FilterPasses(record[cfgFieldIdx:]) {
			filterPasses = false
			break
		}
	}
	return filterPasses
}

// Converts a record (header or normal) to StoredCdr
func (self *FwvRecordsProcessor) recordToStoredCdr(record string, cfgKey string) (*engine.StoredCdr, error) {
	var err error
	var lazyHttpFields []*config.CfgCdrField
	var cfgFields []*config.CfgCdrField
	var duMultiplyFactor float64
	var storedCdr *engine.StoredCdr
	if self.headerCdr != nil { // Clone the header CDR so we can use it as base to future processing (inherit fields defined there)
		storedCdr = self.headerCdr.Clone()
	} else {
		storedCdr = &engine.StoredCdr{CdrHost: "0.0.0.0", ExtraFields: make(map[string]string), Cost: -1}
	}
	if cfgKey == "*header" {
		cfgFields = self.dfltCfg.HeaderFields
		storedCdr.CdrSource = self.dfltCfg.CdrSourceId
		duMultiplyFactor = self.dfltCfg.DataUsageMultiplyFactor
	} else {
		cfgFields = self.cdrcCfgs[cfgKey].ContentFields
		storedCdr.CdrSource = self.cdrcCfgs[cfgKey].CdrSourceId
		duMultiplyFactor = self.cdrcCfgs[cfgKey].DataUsageMultiplyFactor
	}
	for _, cdrFldCfg := range cfgFields {
		var fieldVal string
		switch cdrFldCfg.Type {
		case utils.CDRFIELD:
			for _, cfgFieldRSR := range cdrFldCfg.Value {
				if cfgFieldRSR.IsStatic() {
					fieldVal += cfgFieldRSR.ParseValue("")
				} else { // Dynamic value extracted using index
					if cfgFieldIdx, _ := strconv.Atoi(cfgFieldRSR.Id); len(record) <= cfgFieldIdx {
						return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s", record, cdrFldCfg.Tag)
					} else {
						fieldVal += cfgFieldRSR.ParseValue(fwvValue(record, cfgFieldIdx, cdrFldCfg.Width, cdrFldCfg.Padding))
					}
				}
			}
		case utils.HTTP_POST:
			lazyHttpFields = append(lazyHttpFields, cdrFldCfg) // Will process later so we can send an estimation of storedCdr to http server
		default:
			//return nil, fmt.Errorf("Unsupported field type: %s", cdrFldCfg.Type)
			continue // Don't do anything for unsupported fields
		}
		if err := storedCdr.ParseFieldValue(cdrFldCfg.FieldId, fieldVal, self.timezone); err != nil {
			return nil, err
		}
	}
	if storedCdr.CgrId == "" && storedCdr.AccId != "" && cfgKey != "*header" {
		storedCdr.CgrId = utils.Sha1(storedCdr.AccId, storedCdr.SetupTime.UTC().String())
	}
	if storedCdr.TOR == utils.DATA && duMultiplyFactor != 0 {
		storedCdr.Usage = time.Duration(float64(storedCdr.Usage.Nanoseconds()) * duMultiplyFactor)
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
	if self.headerCdr, err = self.recordToStoredCdr(string(buf), "*header"); err != nil {
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
