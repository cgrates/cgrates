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

type AttrExportCdrsToFile struct {
	CdrFormat                  *string  // Cdr output file format <utils.CdreCdrFormats>
	FieldSeparator             *string  // Separator used between fields
	ExportID                   *string  // Optional exportid
	ExportDirectory            *string  // If provided it overwrites the configured export directory
	ExportFileName             *string  // If provided the output filename will be set to this
	ExportTemplate             *string  // Exported fields template  <""|fld1,fld2|>
	DataUsageMultiplyFactor    *float64 // Multiply data usage before export (eg: convert from KBytes to Bytes)
	SMSUsageMultiplyFactor     *float64 // Multiply sms usage before export (eg: convert from SMS unit to call duration for some billing systems)
	MMSUsageMultiplyFactor     *float64 // Multiply mms usage before export (eg: convert from MMS unit to call duration for some billing systems)
	GenericUsageMultiplyFactor *float64 // Multiply generic usage before export (eg: convert from GENERIC unit to call duration for some billing systems)
	CostMultiplyFactor         *float64 // Multiply the cost before export, eg: apply VAT
	RoundingDecimals           *int     // force rounding to this value
	Verbose                    bool     // Disable CgrIds reporting in reply/ExportedCgrIds and reply/UnexportedCgrIds
	utils.RPCCDRsFilter                 // Inherit the CDR filter attributes
}

type ExportedFileCdrs struct {
	ExportedFilePath          string            // Full path to the newly generated export file
	TotalRecords              int               // Number of CDRs to be exported
	TotalCost                 float64           // Sum of all costs in exported CDRs
	FirstOrderId, LastOrderId int64             // The order id of the last exported CDR
	ExportedCgrIds            []string          // List of successfuly exported cgrids in the file
	UnexportedCgrIds          map[string]string // Map of errored CDRs, map key is cgrid, value will be the error string
}

// Deprecated, please use ApierV1.ExportCDRs instead
func (self *ApierV2) ExportCdrsToFile(attr AttrExportCdrsToFile, reply *ExportedFileCdrs) (err error) {
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
	var expFormat string
	switch exportFormat {
	case utils.MetaFileFWV:
		expFormat = "fwv"
	case utils.MetaFileCSV:
		expFormat = "csv"
	default:
		expFormat = exportFormat
	}
	fileName := fmt.Sprintf("cdre_%s.%s", exportID, expFormat)
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
	cdrsFltr, err := attr.RPCCDRsFilter.AsCDRsFilter(self.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrs, _, err := self.CdrDb.GetCDRs(cdrsFltr, false)
	if err != nil {
		return err
	} else if len(cdrs) == 0 {
		*reply = ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	roundingDecimals := self.Config.GeneralCfg().RoundingDecimals
	if attr.RoundingDecimals != nil {
		roundingDecimals = *attr.RoundingDecimals
	}
	cdrexp, err := engine.NewCDRExporter(cdrs, exportTemplate, exportFormat,
		filePath, utils.META_NONE, exportID, exportTemplate.Synchronous,
		exportTemplate.Attempts, fieldSep, usageMultiplyFactor, costMultiplyFactor,
		roundingDecimals, self.Config.GeneralCfg().HttpSkipTlsVerify,
		self.HTTPPoster, self.FilterS)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := cdrexp.ExportCDRs(); err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrexp.TotalExportedCdrs() == 0 {
		*reply = ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	*reply = ExportedFileCdrs{ExportedFilePath: filePath, TotalRecords: len(cdrs), TotalCost: cdrexp.TotalCost(), FirstOrderId: cdrexp.FirstOrderId(), LastOrderId: cdrexp.LastOrderId()}
	if !attr.Verbose {
		reply.ExportedCgrIds = cdrexp.PositiveExports()
		reply.UnexportedCgrIds = cdrexp.NegativeExports()
	}
	return nil
}
