/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package apier

import (
	"fmt"
	"github.com/cgrates/cgrates/cdrexporter"
	"github.com/cgrates/cgrates/utils"
	"os"
	"path"
	"time"
)

type AttrExpCsvCdrs struct {
	TimeStart string // If provided, will represent the starting of the CDRs interval (>=)
	TimeEnd   string // If provided, will represent the end of the CDRs interval (<)
}

type ExportedCsvCdrs struct {
	ExportedFilePath string // Full path to the newly generated export file
	NumberOfCdrs     int    // Number of CDRs in the export file
}

func (self *ApierV1) ExportCsvCdrs(attr *AttrExpCsvCdrs, reply *ExportedCsvCdrs) error {
	var tStart, tEnd time.Time
	var err error
	if len(attr.TimeStart) != 0 {
		if tStart, err = utils.ParseDate(attr.TimeStart); err != nil {
			return err
		}
	}
	if len(attr.TimeEnd) != 0 {
		if tEnd, err = utils.ParseDate(attr.TimeEnd); err != nil {
			return err
		}
	}
	cdrs, err := self.CdrDb.GetRatedCdrs(tStart, tEnd)
	if err != nil {
		return err
	}
	fileName := path.Join(self.Config.CDRSExportPath, "cgr", "csv", fmt.Sprintf("cdrs_%d.csv", time.Now().Unix()))
	fileOut, err := os.Create(fileName)
	if err != nil {
		return err
	} else {
		defer fileOut.Close()
	}
	csvWriter := cdrexporter.NewCsvCdrWriter(fileOut, self.Config.RoundingDecimals, self.Config.CDRSExportExtraFields)
	for _, cdr := range cdrs {
		if err := csvWriter.Write(cdr.(*utils.RatedCDR)); err != nil {
			os.Remove(fileName)
			return err
		}
	}
	csvWriter.Close()
	*reply = ExportedCsvCdrs{fileName, len(cdrs)}
	return nil
}
