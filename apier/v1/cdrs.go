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
	"strings"
)

func (self *ApierV1) ExportCdrsToFile(attr utils.AttrExpFileCdrs, reply *utils.ExportedFileCdrs) error {
	var tStart, tEnd time.Time
	var err error
	cdrFormat := strings.ToLower(attr.CdrFormat)
	if !utils.IsSliceMember(utils.CdreCdrFormats, cdrFormat) {
		return fmt.Errorf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, "CdrFormat")
 	}
	if len(attr.TimeStart) != 0 {
		if tStart, err = utils.ParseTimeDetectLayout(attr.TimeStart); err != nil {
			return err
		}
	}
	if len(attr.TimeEnd) != 0 {
		if tEnd, err = utils.ParseTimeDetectLayout(attr.TimeEnd); err != nil {
			return err
		}
	}
	cdrs, err := self.CdrDb.GetRatedCdrs(tStart, tEnd)
	if err != nil {
		return err
	}
	var fileName string
	if cdrFormat == utils.CDRE_CSV && len(cdrs) != 0 {
		fileName = path.Join(self.Config.CdreDir, fmt.Sprintf("cdrs_%d.csv", time.Now().Unix()))
		fileOut, err := os.Create(fileName)
		if err != nil {
			return err
		} else {
			defer fileOut.Close()
		}
		csvWriter := cdrexporter.NewCsvCdrWriter(fileOut, self.Config.RoundingDecimals, self.Config.CdreExtraFields)
		for _, cdr := range cdrs {
			if err := csvWriter.Write(cdr); err != nil {
				os.Remove(fileName)
			return err
			}
		}
		csvWriter.Close()
		if attr.RemoveFromDb {
			cgrIds := make([]string, len(cdrs))
			for idx, cdr := range cdrs {
				cgrIds[idx] = cdr.CgrId
			}
			if err := self.CdrDb.RemRatedCdrs(cgrIds); err != nil {
				return err
			}
		}
	}
	*reply = utils.ExportedFileCdrs{fileName, len(cdrs)}
	return nil
}
