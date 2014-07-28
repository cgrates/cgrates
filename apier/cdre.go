/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"path"
	"strconv"
	"strings"
	"time"
)

// Export Cdrs to file
func (self *ApierV1) ExportCdrsToFile(attr utils.AttrExpFileCdrs, reply *utils.ExportedFileCdrs) error {
	var tStart, tEnd time.Time
	var err error
	engine.Logger.Debug(fmt.Sprintf("ExportCdrsToFile: %+v", attr))
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
	exportTemplate := self.Config.CdreDefaultInstance
	if attr.ExportTemplate != nil { // XML Template defined, can be field names or xml reference
		if strings.HasPrefix(*attr.ExportTemplate, utils.XML_PROFILE_PREFIX) {
			if self.Config.XmlCfgDocument == nil {
				return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, "XmlDocumentNotLoaded")
			}
			expTplStr := *attr.ExportTemplate
			if xmlTemplates := self.Config.XmlCfgDocument.GetCdreCfgs(expTplStr[len(utils.XML_PROFILE_PREFIX):]); xmlTemplates == nil {
				return fmt.Errorf("%s:ExportTemplate", utils.ERR_NOT_FOUND)
			} else {
				exportTemplate = xmlTemplates[expTplStr[len(utils.XML_PROFILE_PREFIX):]].AsCdreConfig()
			}
		} else {
			exportTemplate, _ = config.NewDefaultCdreConfig()
			if contentFlds, err := config.NewCdreCdrFieldsFromIds(exportTemplate.CdrFormat == utils.CDRE_FIXED_WIDTH,
				strings.Split(*attr.ExportTemplate, string(utils.CSV_SEP))...); err != nil {
				return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
			} else {
				exportTemplate.ContentFields = contentFlds
			}
		}
	}
	if exportTemplate == nil {
		return fmt.Errorf("%s:ExportTemplate", utils.ERR_MANDATORY_IE_MISSING)
	}
	cdrFormat := exportTemplate.CdrFormat
	if attr.CdrFormat != nil {
		cdrFormat = strings.ToLower(*attr.CdrFormat)
	}
	if !utils.IsSliceMember(utils.CdreCdrFormats, cdrFormat) {
		return fmt.Errorf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, "CdrFormat")
	}
	fieldSep := exportTemplate.FieldSeparator
	if attr.FieldSeparator != nil {
		fieldSep = *attr.FieldSeparator
	}
	exportDir := exportTemplate.ExportDir
	if attr.ExportDir != nil {
		exportDir = *attr.ExportDir
	}
	exportId := strconv.FormatInt(time.Now().Unix(), 10)
	if attr.ExportId != nil {
		exportId = *attr.ExportId
	}
	fileName := fmt.Sprintf("cdre_%s.%s", exportId, cdrFormat)
	if attr.ExportFileName != nil {
		fileName = *attr.ExportFileName
	}
	filePath := path.Join(exportDir, fileName)
	if cdrFormat == utils.CDRE_DRYRUN {
		filePath = utils.CDRE_DRYRUN
	}
	dataUsageMultiplyFactor := exportTemplate.DataUsageMultiplyFactor
	if attr.DataUsageMultiplyFactor != nil {
		dataUsageMultiplyFactor = *attr.DataUsageMultiplyFactor
	}
	costMultiplyFactor := exportTemplate.CostMultiplyFactor
	if attr.CostMultiplyFactor != nil {
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
	maskDestId := exportTemplate.MaskDestId
	if attr.MaskDestinationId != nil {
		maskDestId = *attr.MaskDestinationId
	}
	maskLen := exportTemplate.MaskLength
	if attr.MaskLength != nil {
		maskLen = *attr.MaskLength
	}
	cdrs, err := self.CdrDb.GetStoredCdrs(attr.CgrIds, attr.MediationRunIds, attr.TORs, attr.CdrHosts, attr.CdrSources, attr.ReqTypes, attr.Directions,
		attr.Tenants, attr.Categories, attr.Accounts, attr.Subjects, attr.DestinationPrefixes, attr.RatedAccounts, attr.RatedSubjects, attr.OrderIdStart, attr.OrderIdEnd,
		tStart, tEnd, attr.SkipErrors, attr.SkipRated, false)
	if err != nil {
		return err
	} else if len(cdrs) == 0 {
		*reply = utils.ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	cdrexp, err := cdre.NewCdrExporter(cdrs, self.LogDb, exportTemplate, cdrFormat, fieldSep, exportId,
		dataUsageMultiplyFactor, costMultiplyFactor, costShiftDigits, roundingDecimals, self.Config.RoundingDecimals, maskDestId, maskLen, self.Config.HttpSkipTlsVerify)
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if cdrexp.TotalExportedCdrs() == 0 {
		*reply = utils.ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	if err := cdrexp.WriteToFile(filePath); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = utils.ExportedFileCdrs{ExportedFilePath: filePath, TotalRecords: len(cdrs), TotalCost: cdrexp.TotalCost(), FirstOrderId: cdrexp.FirstOrderId(), LastOrderId: cdrexp.LastOrderId()}
	if !attr.SuppressCgrIds {
		*reply.ExportedCgrIds = cdrexp.PositiveExports()
		*reply.UnexportedCgrIds = cdrexp.NegativeExports()
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
