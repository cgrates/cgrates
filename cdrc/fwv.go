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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
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

func NewFwvRecordsProcessor(file *os.File) *FwvRecordsProcessor {
	return &FwvRecordsProcessor{file: file}
}

type FwvRecordsProcessor struct {
	file      *os.File
	lineLen   int // Length of the line to read
	offset    int // Index of the last processed byte
	cdrFields [][]*config.CfgCdrField
}

func (self *FwvRecordsProcessor) ProcessNextRecord() ([]*engine.StoredCdr, error) {

	recordStr := ""
	return self.processRecord(recordStr)
}

func (self *FwvRecordsProcessor) processRecord(record string) ([]*engine.StoredCdr, error) {
	return nil, nil
}
