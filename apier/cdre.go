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
	"github.com/cgrates/cgrates/cdre"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// Export Cdrs to file
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
	fileName := attr.ExportFileName
	exportId := attr.ExportId
	if len(exportId) == 0 {
		exportId = strconv.FormatInt(time.Now().Unix(), 10)
	}
	roundDecimals := attr.RoundingDecimals
	if roundDecimals == 0 {
		roundDecimals = self.Config.RoundingDecimals
	}
	cdrs, err := self.CdrDb.GetStoredCdrs(attr.CgrIds, attr.MediationRunId, attr.CdrHost, attr.CdrSource, attr.ReqType, attr.Direction,
		attr.Tenant, attr.Tor, attr.Account, attr.Subject, attr.DestinationPrefix, tStart, tEnd, attr.SkipErrors, attr.SkipRated)
	if err != nil {
		return err
	} else if len(cdrs) == 0 {
		*reply = utils.ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	switch cdrFormat {
	case utils.CDRE_DRYRUN:
		exportedIds := make([]string, len(cdrs))
		for idxCdr, cdr := range cdrs {
			exportedIds[idxCdr] = cdr.CgrId
		}
		*reply = utils.ExportedFileCdrs{ExportedFilePath: utils.CDRE_DRYRUN, TotalRecords: len(cdrs), ExportedCgrIds: exportedIds}
	case utils.CDRE_CSV:
		if len(fileName) == 0 {
			fileName = fmt.Sprintf("cdre_%s.csv", exportId)
		}
		exportedFields := self.Config.CdreExportedFields
		if len(attr.ExportTemplate) != 0 {
			if exportedFields, err = config.ParseRSRFields(attr.ExportTemplate); err != nil {
				return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
			}
		}
		if len(exportedFields) == 0 {
			return fmt.Errorf("%s:ExportTemplate", utils.ERR_MANDATORY_IE_MISSING)
		}
		filePath := path.Join(self.Config.CdreDir, utils.CDRE_CSV, fileName)
		fileOut, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer fileOut.Close()
		csvWriter := cdre.NewCsvCdrWriter(fileOut, roundDecimals, exportedFields)
		exportedIds := make([]string, 0)
		unexportedIds := make(map[string]string)
		for _, cdr := range cdrs {
			if err := csvWriter.WriteCdr(cdr); err != nil {
				unexportedIds[cdr.CgrId] = err.Error()
			} else {
				exportedIds = append(exportedIds, cdr.CgrId)
			}
		}
		csvWriter.Close()
		*reply = utils.ExportedFileCdrs{ExportedFilePath: filePath, TotalRecords: len(cdrs), ExportedCgrIds: exportedIds, UnexportedCgrIds: unexportedIds}
	case utils.CDRE_FIXED_WIDTH:
		if len(fileName) == 0 {
			fileName = fmt.Sprintf("cdre_%s.fwv", exportId)
		}
		exportTemplate := self.Config.CdreFWXmlTemplate
		if len(attr.ExportTemplate) != 0 && self.Config.XmlCfgDocument != nil {
			if xmlTemplate, err := self.Config.XmlCfgDocument.GetCdreFWCfg(attr.ExportTemplate[len(utils.XML_PROFILE_PREFIX):]); err != nil {
				return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
			} else if xmlTemplate != nil {
				exportTemplate = xmlTemplate
			}
		}
		if exportTemplate == nil {
			return fmt.Errorf("%s:ExportTemplate", utils.ERR_MANDATORY_IE_MISSING)
		}
		filePath := path.Join(self.Config.CdreDir, utils.CDRE_FIXED_WIDTH, fileName)
		fileOut, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer fileOut.Close()
		fww, _ := cdre.NewFWCdrWriter(self.LogDb, fileOut, exportTemplate, exportId, roundDecimals)
		exportedIds := make([]string, 0)
		unexportedIds := make(map[string]string)
		for _, cdr := range cdrs {
			if err := fww.WriteCdr(cdr); err != nil {
				unexportedIds[cdr.CgrId] = err.Error()
			} else {
				exportedIds = append(exportedIds, cdr.CgrId)
			}
		}
		fww.Close()
		*reply = utils.ExportedFileCdrs{ExportedFilePath: filePath, TotalRecords: len(cdrs), ExportedCgrIds: exportedIds, UnexportedCgrIds: unexportedIds}
	}
	return nil
}

// Remove Cdrs out of CDR storage
func (self *ApierV1) RemCdrs(attrs utils.AttrRemCdrs, reply *string) error {
	if len(attrs.CgrIds) == 0 {
		return fmt.Errorf("%s:CgrIds", utils.ERR_MANDATORY_IE_MISSING)
	}
	if err := self.CdrDb.RemStoredCdrs(attrs.CgrIds); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}
