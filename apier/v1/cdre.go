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

package v1

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cgrates/cgrates/cdre"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func (self *ApierV1) ExportCdrsToZipString(attr utils.AttrExpFileCdrs, reply *string) error {
	tmpDir := "/tmp"
	attr.ExportDir = &tmpDir // Enforce exporting to tmp always so we avoid cleanup issues
	efc := utils.ExportedFileCdrs{}
	if err := self.ExportCdrsToFile(attr, &efc); err != nil {
		return err
	} else if efc.TotalRecords == 0 || len(efc.ExportedFilePath) == 0 {
		return errors.New("No CDR records to export")
	}
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)
	// Create a new zip archive.
	w := zip.NewWriter(buf)
	// read generated file
	content, err := ioutil.ReadFile(efc.ExportedFilePath)
	if err != nil {
		return err
	}
	exportFileName := path.Base(efc.ExportedFilePath)
	f, err := w.Create(exportFileName)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	if err != nil {
		return err
	}
	// Write metadata into a separate file with extension .cgr
	medaData, err := json.MarshalIndent(efc, "", "  ")
	if err != nil {
		errors.New("Failed creating metadata content")
	}
	medatadaFileName := exportFileName[:len(path.Ext(exportFileName))] + ".cgr"
	mf, err := w.Create(medatadaFileName)
	if err != nil {
		return err
	}
	_, err = mf.Write(medaData)
	if err != nil {
		return err
	}
	// Make sure to check the error on Close.
	if err := w.Close(); err != nil {
		return err
	}
	if err := os.Remove(efc.ExportedFilePath); err != nil {
		fmt.Errorf("Failed removing exported file at path: %s", efc.ExportedFilePath)
	}
	*reply = base64.StdEncoding.EncodeToString(buf.Bytes())
	return nil
}

// Export Cdrs to file
func (self *ApierV1) ExportCdrsToFile(attr utils.AttrExpFileCdrs, reply *utils.ExportedFileCdrs) error {
	var err error

	cdreReloadStruct := <-self.Config.ConfigReloads[utils.CDRE]                  // Read the content of the channel, locking it
	defer func() { self.Config.ConfigReloads[utils.CDRE] <- cdreReloadStruct }() // Unlock reloads at exit
	exportTemplate := self.Config.CdreProfiles[utils.META_DEFAULT]
	if attr.ExportTemplate != nil && len(*attr.ExportTemplate) != 0 { // Export template prefered, use it
		var hasIt bool
		if exportTemplate, hasIt = self.Config.CdreProfiles[*attr.ExportTemplate]; !hasIt {
			return fmt.Errorf("%s:ExportTemplate", utils.ErrNotFound.Error())
		}
	}
	if exportTemplate == nil {
		return fmt.Errorf("%s:ExportTemplate", utils.ErrMandatoryIeMissing.Error())
	}
	cdrFormat := exportTemplate.CdrFormat
	if attr.CdrFormat != nil && len(*attr.CdrFormat) != 0 {
		cdrFormat = strings.ToLower(*attr.CdrFormat)
	}
	if !utils.IsSliceMember(utils.CdreCdrFormats, cdrFormat) {
		return fmt.Errorf("%s:%s", utils.ErrMandatoryIeMissing.Error(), "CdrFormat")
	}
	fieldSep := exportTemplate.FieldSeparator
	if attr.FieldSeparator != nil && len(*attr.FieldSeparator) != 0 {
		fieldSep, _ = utf8.DecodeRuneInString(*attr.FieldSeparator)
		if fieldSep == utf8.RuneError {
			return fmt.Errorf("%s:FieldSeparator:%s", utils.ErrServerError.Error(), "Invalid")
		}
	}
	exportDir := exportTemplate.ExportFolder
	if attr.ExportDir != nil && len(*attr.ExportDir) != 0 {
		exportDir = *attr.ExportDir
	}
	exportId := strconv.FormatInt(time.Now().Unix(), 10)
	if attr.ExportId != nil && len(*attr.ExportId) != 0 {
		exportId = *attr.ExportId
	}
	fileName := fmt.Sprintf("cdre_%s.%s", exportId, cdrFormat)
	if attr.ExportFileName != nil && len(*attr.ExportFileName) != 0 {
		fileName = *attr.ExportFileName
	}
	filePath := path.Join(exportDir, fileName)
	if cdrFormat == utils.DRYRUN {
		filePath = utils.DRYRUN
	}
	dataUsageMultiplyFactor := exportTemplate.DataUsageMultiplyFactor
	if attr.DataUsageMultiplyFactor != nil && *attr.DataUsageMultiplyFactor != 0.0 {
		dataUsageMultiplyFactor = *attr.DataUsageMultiplyFactor
	}
	smsUsageMultiplyFactor := exportTemplate.SMSUsageMultiplyFactor
	if attr.SmsUsageMultiplyFactor != nil && *attr.SmsUsageMultiplyFactor != 0.0 {
		smsUsageMultiplyFactor = *attr.SmsUsageMultiplyFactor
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
	if attr.MaskDestinationId != nil && len(*attr.MaskDestinationId) != 0 {
		maskDestId = *attr.MaskDestinationId
	}
	maskLen := exportTemplate.MaskLength
	if attr.MaskLength != nil {
		maskLen = *attr.MaskLength
	}
	cdrsFltr, err := attr.AsCDRsFilter(self.Config.DefaultTimezone)
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
	cdrexp, err := cdre.NewCdrExporter(cdrs, self.CdrDb, exportTemplate, cdrFormat, fieldSep, exportId, dataUsageMultiplyFactor, smsUsageMultiplyFactor, genericUsageMultiplyFactor,
		costMultiplyFactor, costShiftDigits, roundingDecimals, self.Config.RoundingDecimals, maskDestId, maskLen, self.Config.HttpSkipTlsVerify, self.Config.DefaultTimezone)
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
	if !attr.SuppressCgrIds {
		reply.ExportedCgrIds = cdrexp.PositiveExports()
		reply.UnexportedCgrIds = cdrexp.NegativeExports()
	}
	return nil
}

// Reloads CDRE configuration out of folder specified
func (apier *ApierV1) ReloadCdreConfig(attrs AttrReloadConfig, reply *string) error {
	if attrs.ConfigDir == "" {
		attrs.ConfigDir = utils.CONFIG_DIR
	}
	newCfg, err := config.NewCGRConfigFromFolder(attrs.ConfigDir)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdreReloadStruct := <-apier.Config.ConfigReloads[utils.CDRE] // Get the CDRE reload channel                     // Read the content of the channel, locking it
	apier.Config.CdreProfiles = newCfg.CdreProfiles
	apier.Config.ConfigReloads[utils.CDRE] <- cdreReloadStruct // Unlock reloads
	utils.Logger.Info("<CDRE> Configuration reloaded")
	*reply = OK
	return nil
}
