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
	"io"
	"os"
)

/*file, _ := os.Open(path.Join("/tmp", "acc_1.log"))
defer file.Close()
fs, _ := file.Stat()
fmt.Printf("FileSize: %d, content size: %d, %q", fs.Size(), len([]byte(fullSuccessfull)), fullSuccessfull)
buf := make([]byte, 109)
_, err := file.ReadAt(buf, fs.Size()-int64(len(buf)))
if err != nil {
	t.Error(err)
}
fmt.Printf("Have read in buffer: <%q>, len: %d", string(buf), len(string(buf)))
*/

func NewFwvRecordsProcessor(file *os.File, cdrcCfgs map[string]*config.CdrcConfig) *FwvRecordsProcessor {
	frp := &FwvRecordsProcessor{file: file, cdrcCfgs: cdrcCfgs}
	for _, frp.dfltCfg = range cdrcCfgs { // Set the first available instance to be used for common parameters
		break
	}
	return frp
}

type FwvRecordsProcessor struct {
	file          *os.File
	cdrcCfgs      map[string]*config.CdrcConfig
	dfltCfg       *config.CdrcConfig // General parameters
	lineLen       int64              // Length of the line in the file
	offset        int64              // Index of the next byte to process
	trailerOffset int64              // Index where trailer starts, to be used as boundary when reading cdrs
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

func (self *FwvRecordsProcessor) ProcessNextRecord() ([]*engine.StoredCdr, error) {
	defer func() { self.offset += self.lineLen }() // Schedule increasing the offset once we are out from processing the record
	if self.offset == 0 {                          // First time, set the necessary offsets
		if err := self.setLineLen(); err != nil {
			engine.Logger.Err(fmt.Sprintf("<Cdrc> Row 0, error: cannot set lineLen: %s", err.Error()))
			return nil, io.EOF
		}
		if len(self.dfltCfg.TrailerFields) != 0 {
			if fi, err := self.file.Stat(); err != nil {
				engine.Logger.Err(fmt.Sprintf("<Cdrc> Row 0, error: cannot get file stats: %s", err.Error()))
				return nil, err
			} else {
				self.trailerOffset = fi.Size() - self.lineLen
			}
		}
		if len(self.dfltCfg.HeaderFields) != 0 { // ToDo: Process here the header fields
			if err := self.processHeader(); err != nil {
				engine.Logger.Err(fmt.Sprintf("<Cdrc> Row 0, error reading header: %s", err.Error()))
				return nil, io.EOF
			}
			return nil, nil
		}
	}
	recordCdrs := make([]*engine.StoredCdr, 0) // More CDRs based on the number of filters and field templates
	if self.trailerOffset != 0 && self.offset >= self.trailerOffset {
		if err := self.processTrailer(); err != nil && err != io.EOF {
			engine.Logger.Err(fmt.Sprintf("<Cdrc> Read trailer error: %s ", err.Error()))
		}
		return nil, io.EOF
	}
	buf := make([]byte, self.lineLen)
	nRead, err := self.file.Read(buf)
	if err != nil {
		return nil, err
	} else if nRead != len(buf) {
		engine.Logger.Err(fmt.Sprintf("<Cdrc> Could not read complete line, have instead: %s", string(buf)))
		return nil, io.EOF
	}
	for cfgKey := range self.cdrcCfgs {
		filterBreak := false
		// ToDo: Field filters
		if filterBreak { // Stop importing cdrc fields profile due to non matching filter
			continue
		}
		if storedCdr, err := self.recordToStoredCdr(string(buf), cfgKey); err != nil {
			return nil, fmt.Errorf("Failed converting to StoredCdr, error: %s", err.Error())
		} else if storedCdr != nil {
			recordCdrs = append(recordCdrs, storedCdr)
		}
	}
	return recordCdrs, nil
}

func (self *FwvRecordsProcessor) recordToStoredCdr(record string, cfgKey string) (*engine.StoredCdr, error) {
	//engine.Logger.Debug(fmt.Sprintf("RecordToStoredCdr: <%q>, cfgKey: %s, offset: %d, trailerOffset: %d, lineLen: %d", record, cfgKey, self.offset, self.trailerOffset, self.lineLen))
	return nil, nil
}

func (self *FwvRecordsProcessor) processHeader() error {
	buf := make([]byte, self.lineLen)
	if nRead, err := self.file.Read(buf); err != nil {
		return err
	} else if nRead != len(buf) {
		return fmt.Errorf("In header, line len: %d, have read: %d", self.lineLen, nRead)
	}
	//engine.Logger.Debug(fmt.Sprintf("Have read header: <%q>", string(buf)))
	return nil
}

func (self *FwvRecordsProcessor) processTrailer() error {
	buf := make([]byte, self.lineLen)
	if nRead, err := self.file.ReadAt(buf, self.trailerOffset); err != nil {
		return err
	} else if nRead != len(buf) {
		return fmt.Errorf("In trailer, line len: %d, have read: %d", self.lineLen, nRead)
	}
	//engine.Logger.Debug(fmt.Sprintf("Have read trailer: <%q>", string(buf)))
	return nil
}
