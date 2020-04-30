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
	CdrFormat           *string // Cdr output file format <utils.CdreCdrFormats>
	FieldSeparator      *string // Separator used between fields
	ExportID            *string // Optional exportid
	ExportDirectory     *string // If provided it overwrites the configured export directory
	ExportFileName      *string // If provided the output filename will be set to this
	ExportTemplate      *string // Exported fields template  <""|fld1,fld2|>
	Verbose             bool    // Disable CgrIds reporting in reply/ExportedCgrIds and reply/UnexportedCgrIds
	utils.RPCCDRsFilter         // Inherit the CDR filter attributes
}

// Deprecated, please use APIerSv1.ExportCDRs instead
func (apiv2 *APIerSv2) ExportCdrsToFile(attr AttrExportCdrsToFile, reply *utils.ExportedFileCdrs) (err error) {
	cdreReloadStruct := <-apiv2.Config.ConfigReloads[utils.CDRE]                  // Read the content of the channel, locking it
	defer func() { apiv2.Config.ConfigReloads[utils.CDRE] <- cdreReloadStruct }() // Unlock reloads at exit
	exportTemplate := apiv2.Config.CdreProfiles[utils.MetaDefault]
	if attr.ExportTemplate != nil && len(*attr.ExportTemplate) != 0 { // Export template prefered, use it
		var hasIt bool
		if exportTemplate, hasIt = apiv2.Config.CdreProfiles[*attr.ExportTemplate]; !hasIt {
			return fmt.Errorf("%s:ExportTemplate", utils.ErrNotFound)
		}
	}
	exportFormat := exportTemplate.ExportFormat
	if attr.CdrFormat != nil && len(*attr.CdrFormat) != 0 {
		exportFormat = strings.ToLower(*attr.CdrFormat)
	}
	if !utils.CDRExportFormats.Has(exportFormat) {
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
	cdrsFltr, err := attr.RPCCDRsFilter.AsCDRsFilter(apiv2.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrs, _, err := apiv2.CdrDb.GetCDRs(cdrsFltr, false)
	if err != nil {
		return err
	} else if len(cdrs) == 0 {
		*reply = utils.ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	cdrexp, err := engine.NewCDRExporter(cdrs, exportTemplate, exportFormat,
		filePath, utils.META_NONE, exportID, exportTemplate.Synchronous,
		exportTemplate.Attempts, fieldSep, apiv2.Config.GeneralCfg().HttpSkipTlsVerify,
		apiv2.Config.ApierCfg().AttributeSConns, apiv2.FilterS)
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
	*reply = utils.ExportedFileCdrs{ExportedFilePath: filePath, TotalRecords: len(cdrs),
		TotalCost: cdrexp.TotalCost(), FirstOrderId: cdrexp.FirstOrderID(), LastOrderId: cdrexp.LastOrderID()}
	if attr.Verbose {
		reply.ExportedCgrIds = cdrexp.PositiveExports()
		reply.UnexportedCgrIds = cdrexp.NegativeExports()
	}
	return nil
}
