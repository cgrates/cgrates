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
package v2

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Export Cdrs to file
func (self *ApierV2) ExportCdrsToFile(attr utils.AttrExportCdrsToFile, reply *utils.ExportedFileCdrs) error {
	var err error
	cdreReloadStruct := <-self.Config.ConfigReloads[utils.CDRE]                  // Read the content of the channel, locking it
	defer func() { self.Config.ConfigReloads[utils.CDRE] <- cdreReloadStruct }() // Unlock reloads at exit
	exportTemplate := self.Config.CdreProfiles[utils.META_DEFAULT]
	if attr.ExportTemplate != nil && len(*attr.ExportTemplate) != 0 { // Export template prefered, use it
		var hasIt bool
		if exportTemplate, hasIt = self.Config.CdreProfiles[*attr.ExportTemplate]; !hasIt {
			return fmt.Errorf("%s:ExportTemplate", utils.ErrNotFound)
		}
	}
	exportFormat := exportTemplate.ExportFormat
	if attr.CdrFormat != nil && len(*attr.CdrFormat) != 0 {
		exportFormat = strings.ToLower(*attr.CdrFormat)
	}
	if !utils.IsSliceMember(utils.CDRExportFormats, exportFormat) {
		return utils.NewErrMandatoryIeMissing("CdrFormat")
	}
	fieldSep := exportTemplate.FieldSeparator
	if attr.FieldSeparator != nil && len(*attr.FieldSeparator) != 0 {
		fieldSep, _ = utf8.DecodeRuneInString(*attr.FieldSeparator)
		if fieldSep == utf8.RuneError {
			return fmt.Errorf("%s:FieldSeparator:%s", utils.ErrServerError, "Invalid")
		}
	}
	eDir := exportTemplate.ExportPath
	if attr.ExportDirectory != nil && len(*attr.ExportDirectory) != 0 {
		eDir = *attr.ExportDirectory
	}
	exportID := strconv.FormatInt(time.Now().Unix(), 10)
	if attr.ExportID != nil && len(*attr.ExportID) != 0 {
		exportID = *attr.ExportID
	}
	fileName := fmt.Sprintf("cdre_%s.%s", exportID, exportFormat)
	if attr.ExportFileName != nil && len(*attr.ExportFileName) != 0 {
		fileName = *attr.ExportFileName
	}
	filePath := path.Join(eDir, fileName)
	if exportFormat == utils.DRYRUN {
		filePath = utils.DRYRUN
	}
	usageMultiplyFactor := exportTemplate.UsageMultiplyFactor
	if attr.DataUsageMultiplyFactor != nil && *attr.DataUsageMultiplyFactor != 0.0 {
		usageMultiplyFactor[utils.DATA] = *attr.DataUsageMultiplyFactor
	}
	if attr.SMSUsageMultiplyFactor != nil && *attr.SMSUsageMultiplyFactor != 0.0 {
		usageMultiplyFactor[utils.SMS] = *attr.SMSUsageMultiplyFactor
	}
	if attr.MMSUsageMultiplyFactor != nil && *attr.MMSUsageMultiplyFactor != 0.0 {
		usageMultiplyFactor[utils.MMS] = *attr.MMSUsageMultiplyFactor
	}
	if attr.GenericUsageMultiplyFactor != nil && *attr.GenericUsageMultiplyFactor != 0.0 {
		usageMultiplyFactor[utils.GENERIC] = *attr.GenericUsageMultiplyFactor
	}
	costMultiplyFactor := exportTemplate.CostMultiplyFactor
	if attr.CostMultiplyFactor != nil && *attr.CostMultiplyFactor != 0.0 {
		costMultiplyFactor = *attr.CostMultiplyFactor
	}
	cdrsFltr, err := attr.RPCCDRsFilter.AsCDRsFilter(self.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrs, _, err := self.CdrDb.GetCDRs(cdrsFltr, false)
	if err != nil {
		return err
	} else if len(cdrs) == 0 {
		*reply = utils.ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	roundingDecimals := self.Config.RoundingDecimals
	if attr.RoundingDecimals != nil {
		roundingDecimals = *attr.RoundingDecimals
	}
	cdrexp, err := engine.NewCDRExporter(cdrs, exportTemplate, exportFormat, filePath, utils.META_NONE, exportID,
		exportTemplate.Synchronous, exportTemplate.Attempts, fieldSep, usageMultiplyFactor,
		costMultiplyFactor, roundingDecimals, self.Config.HttpSkipTlsVerify, self.HTTPPoster)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := cdrexp.ExportCDRs(); err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrexp.TotalExportedCdrs() == 0 {
		*reply = utils.ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	*reply = utils.ExportedFileCdrs{ExportedFilePath: filePath, TotalRecords: len(cdrs), TotalCost: cdrexp.TotalCost(), FirstOrderId: cdrexp.FirstOrderId(), LastOrderId: cdrexp.LastOrderId()}
	if !attr.Verbose {
		reply.ExportedCgrIds = cdrexp.PositiveExports()
		reply.UnexportedCgrIds = cdrexp.NegativeExports()
	}
	return nil
}
