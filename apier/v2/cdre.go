/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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

	"github.com/cgrates/cgrates/cdre"
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
	cdrFormat := exportTemplate.CdrFormat
	if attr.CdrFormat != nil && len(*attr.CdrFormat) != 0 {
		cdrFormat = strings.ToLower(*attr.CdrFormat)
	}
	if !utils.IsSliceMember(utils.CdreCdrFormats, cdrFormat) {
		return utils.NewErrMandatoryIeMissing("CdrFormat")
	}
	fieldSep := exportTemplate.FieldSeparator
	if attr.FieldSeparator != nil && len(*attr.FieldSeparator) != 0 {
		fieldSep, _ = utf8.DecodeRuneInString(*attr.FieldSeparator)
		if fieldSep == utf8.RuneError {
			return fmt.Errorf("%s:FieldSeparator:%s", utils.ErrServerError, "Invalid")
		}
	}
	ExportFolder := exportTemplate.ExportFolder
	if attr.ExportFolder != nil && len(*attr.ExportFolder) != 0 {
		ExportFolder = *attr.ExportFolder
	}
	ExportID := strconv.FormatInt(time.Now().Unix(), 10)
	if attr.ExportID != nil && len(*attr.ExportID) != 0 {
		ExportID = *attr.ExportID
	}
	fileName := fmt.Sprintf("cdre_%s.%s", ExportID, cdrFormat)
	if attr.ExportFileName != nil && len(*attr.ExportFileName) != 0 {
		fileName = *attr.ExportFileName
	}
	filePath := path.Join(ExportFolder, fileName)
	if cdrFormat == utils.DRYRUN {
		filePath = utils.DRYRUN
	}
	dataUsageMultiplyFactor := exportTemplate.DataUsageMultiplyFactor
	if attr.DataUsageMultiplyFactor != nil && *attr.DataUsageMultiplyFactor != 0.0 {
		dataUsageMultiplyFactor = *attr.DataUsageMultiplyFactor
	}
	SMSUsageMultiplyFactor := exportTemplate.SMSUsageMultiplyFactor
	if attr.SMSUsageMultiplyFactor != nil && *attr.SMSUsageMultiplyFactor != 0.0 {
		SMSUsageMultiplyFactor = *attr.SMSUsageMultiplyFactor
	}
	MMSUsageMultiplyFactor := exportTemplate.MMSUsageMultiplyFactor
	if attr.MMSUsageMultiplyFactor != nil && *attr.MMSUsageMultiplyFactor != 0.0 {
		MMSUsageMultiplyFactor = *attr.MMSUsageMultiplyFactor
	}
	genericUsageMultiplyFactor := exportTemplate.GenericUsageMultiplyFactor
	if attr.GenericUsageMultiplyFactor != nil && *attr.GenericUsageMultiplyFactor != 0.0 {
		genericUsageMultiplyFactor = *attr.GenericUsageMultiplyFactor
	}
	costMultiplyFactor := exportTemplate.CostMultiplyFactor
	if attr.CostMultiplyFactor != nil && *attr.CostMultiplyFactor != 0.0 {
		costMultiplyFactor = *attr.CostMultiplyFactor
	}
	costShiftDigits := exportTemplate.CostShiftDigits
	if attr.CostShiftDigits != nil {
		costShiftDigits = *attr.CostShiftDigits
	}
	roundingDecimals := exportTemplate.CostRoundingDecimals
	if attr.RoundDecimals != nil {
		roundingDecimals = *attr.RoundDecimals
	}
	maskDestId := exportTemplate.MaskDestinationID
	if attr.MaskDestinationID != nil && len(*attr.MaskDestinationID) != 0 {
		maskDestId = *attr.MaskDestinationID
	}
	maskLen := exportTemplate.MaskLength
	if attr.MaskLength != nil {
		maskLen = *attr.MaskLength
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
	cdrexp, err := cdre.NewCdrExporter(cdrs, self.CdrDb, exportTemplate, cdrFormat, fieldSep, ExportID, dataUsageMultiplyFactor, SMSUsageMultiplyFactor, MMSUsageMultiplyFactor, genericUsageMultiplyFactor, costMultiplyFactor, costShiftDigits, roundingDecimals, self.Config.RoundingDecimals, maskDestId, maskLen, self.Config.HttpSkipTlsVerify, self.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrexp.TotalExportedCdrs() == 0 {
		*reply = utils.ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	if err := cdrexp.WriteToFile(filePath); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.ExportedFileCdrs{ExportedFilePath: filePath, TotalRecords: len(cdrs), TotalCost: cdrexp.TotalCost(), FirstOrderId: cdrexp.FirstOrderId(), LastOrderId: cdrexp.LastOrderId()}
	if !attr.Verbose {
		reply.ExportedCgrIds = cdrexp.PositiveExports()
		reply.UnexportedCgrIds = cdrexp.NegativeExports()
	}
	return nil
}
