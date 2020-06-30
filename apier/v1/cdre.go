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

package v1

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (apierSv1 *APIerSv1) ExportCdrsToZipString(attr *utils.AttrExpFileCdrs, reply *string) error {
	tmpDir := "/tmp"
	attr.ExportDir = &tmpDir // Enforce exporting to tmp always so we avoid cleanup issues
	efc := utils.ExportedFileCdrs{}
	if err := apierSv1.ExportCdrsToFile(attr, &efc); err != nil {
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

// Deprecated by AttrExportCDRsToFile
func (apierSv1 *APIerSv1) ExportCdrsToFile(attr *utils.AttrExpFileCdrs, reply *utils.ExportedFileCdrs) (err error) {
	exportTemplate := apierSv1.Config.CdreProfiles[utils.MetaDefault]
	if attr.ExportTemplate != nil && len(*attr.ExportTemplate) != 0 { // Export template prefered, use it
		var hasIt bool
		if exportTemplate, hasIt = apierSv1.Config.CdreProfiles[*attr.ExportTemplate]; !hasIt {
			return fmt.Errorf("%s:ExportTemplate", utils.ErrNotFound.Error())
		}
	}
	if exportTemplate == nil {
		return fmt.Errorf("%s:ExportTemplate", utils.ErrMandatoryIeMissing.Error())
	}
	exportFormat := exportTemplate.ExportFormat
	if attr.CdrFormat != nil && len(*attr.CdrFormat) != 0 {
		exportFormat = strings.ToLower(*attr.CdrFormat)
	}
	if !utils.CDRExportFormats.Has(exportFormat) {
		return fmt.Errorf("%s:%s", utils.ErrMandatoryIeMissing.Error(), "CdrFormat")
	}
	fieldSep := exportTemplate.FieldSeparator
	if attr.FieldSeparator != nil && len(*attr.FieldSeparator) != 0 {
		fieldSep, _ = utf8.DecodeRuneInString(*attr.FieldSeparator)
		if fieldSep == utf8.RuneError {
			return fmt.Errorf("%s:FieldSeparator:%s", utils.ErrServerError.Error(), "Invalid")
		}
	}
	exportPath := exportTemplate.ExportPath
	if attr.ExportDir != nil && len(*attr.ExportDir) != 0 {
		exportPath = *attr.ExportDir
	}
	exportID := strconv.FormatInt(time.Now().Unix(), 10)
	if attr.ExportId != nil && len(*attr.ExportId) != 0 {
		exportID = *attr.ExportId
	}
	fileName := fmt.Sprintf("cdre_%s.%s", exportID, exportFormat)
	if attr.ExportFileName != nil && len(*attr.ExportFileName) != 0 {
		fileName = *attr.ExportFileName
	}
	filePath := path.Join(exportPath, fileName)
	if exportFormat == utils.DRYRUN {
		filePath = utils.DRYRUN
	}
	cdrsFltr, err := attr.AsCDRsFilter(apierSv1.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrs, _, err := apierSv1.CdrDb.GetCDRs(cdrsFltr, false)
	if err != nil {
		return err
	} else if len(cdrs) == 0 {
		*reply = utils.ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	cdrexp, err := engine.NewCDRExporter(cdrs, exportTemplate, exportFormat,
		filePath, utils.META_NONE, exportID, exportTemplate.Synchronous,
		exportTemplate.Attempts, fieldSep,
		apierSv1.Config.GeneralCfg().HttpSkipTlsVerify,
		apierSv1.Config.ApierCfg().AttributeSConns, apierSv1.FilterS)
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
	*reply = utils.ExportedFileCdrs{ExportedFilePath: filePath,
		TotalRecords: len(cdrs), TotalCost: cdrexp.TotalCost(),
		FirstOrderId: cdrexp.FirstOrderID(), LastOrderId: cdrexp.LastOrderID()}
	if !attr.SuppressCgrIds {
		reply.ExportedCgrIds = cdrexp.PositiveExports()
		reply.UnexportedCgrIds = cdrexp.NegativeExports()
	}
	return nil
}

type ConfigPathArg struct {
	ConfigPath string
}

// Reloads CDRE configuration out of folder specified
func (apierSv1 *APIerSv1) ReloadCdreConfig(attrs *ConfigPathArg, reply *string) error {
	if attrs.ConfigPath == utils.EmptyString {
		attrs.ConfigPath = utils.CONFIG_PATH
	}
	newCfg, err := config.NewCGRConfigFromPath(attrs.ConfigPath)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdreReloadStruct := <-apierSv1.Config.ConfigReloads[utils.CDRE] // Get the CDRE reload channel                     // Read the content of the channel, locking it
	apierSv1.Config.CdreProfiles = newCfg.CdreProfiles
	apierSv1.Config.ConfigReloads[utils.CDRE] <- cdreReloadStruct // Unlock reloads
	utils.Logger.Info("<CDRE> Configuration reloaded")
	*reply = utils.OK
	return nil
}

// ArgExportCDRs are the arguments passed to ExportCDRs method
type ArgExportCDRs struct {
	ExportArgs          map[string]interface{}
	Verbose             bool // Disable CgrIds reporting in reply/ExportedCgrIds and reply/UnexportedCgrIds
	utils.RPCCDRsFilter      // Inherit the CDR filter attributes
}

// RplExportedCDRs contain the reply of the ExportCDRs API
type RplExportedCDRs struct {
	ExportedPath              string            // Full path to the newly generated export file
	TotalRecords              int               // Number of CDRs to be exported
	TotalCost                 float64           // Sum of all costs in exported CDRs
	FirstOrderID, LastOrderID int64             // The order id of the last exported CDR
	ExportedCGRIDs            []string          // List of successfuly exported cgrids in the file
	UnexportedCGRIDs          map[string]string // Map of errored CDRs, map key is cgrid, value will be the error string
}

// ExportCDRs exports CDRs on a path (file or remote)
func (apierSv1 *APIerSv1) ExportCDRs(arg *ArgExportCDRs, reply *RplExportedCDRs) (err error) {
	cdreReloadStruct := <-apierSv1.Config.ConfigReloads[utils.CDRE]                  // Read the content of the channel, locking it
	defer func() { apierSv1.Config.ConfigReloads[utils.CDRE] <- cdreReloadStruct }() // Unlock reloads at exit
	exportTemplate := apierSv1.Config.CdreProfiles[utils.MetaDefault]
	if expTemplate, has := arg.ExportArgs[utils.ExportTemplate]; has {
		var hasIt bool
		if exportTemplate, hasIt = apierSv1.Config.CdreProfiles[utils.IfaceAsString(expTemplate)]; !hasIt {
			return fmt.Errorf("%s:ExportTemplate", utils.ErrNotFound)
		}
	}
	exportFormat := exportTemplate.ExportFormat
	if expformat, has := arg.ExportArgs[utils.ExportFormat]; has {
		exportFormat = strings.ToLower(utils.IfaceAsString(expformat))
	}
	if !utils.CDRExportFormats.Has(exportFormat) {
		return utils.NewErrMandatoryIeMissing("CdrFormat")
	}
	synchronous := exportTemplate.Synchronous
	if sync, has := arg.ExportArgs[utils.Synchronous]; has {
		if synchronous, err = utils.IfaceAsBool(sync); err != nil {
			return
		}
	}
	attempts := exportTemplate.Attempts
	if ate, has := arg.ExportArgs[utils.Attempts]; has {
		var atte int64
		if atte, err = utils.IfaceAsTInt64(ate); err != nil {
			return
		}
		attempts = int(atte)
	}
	fieldSep := exportTemplate.FieldSeparator
	if fieldS, has := arg.ExportArgs[utils.FieldSeparator]; has {
		fieldSep, _ = utf8.DecodeRuneInString(utils.IfaceAsString(fieldS))
		if fieldSep == utf8.RuneError {
			return fmt.Errorf("%s:FieldSeparator:%s", utils.ErrServerError, "Invalid")
		}
	}
	eDir := exportTemplate.ExportPath
	if expPath, has := arg.ExportArgs[utils.ExportPath]; has {
		eDir = utils.IfaceAsString(expPath)
	}
	exportID := strconv.FormatInt(time.Now().Unix(), 10)
	if expID, has := arg.ExportArgs[utils.ExportID]; has {
		exportID = utils.IfaceAsString(expID)
	}
	var expFormat string
	switch exportFormat {
	case utils.MetaFileFWV:
		expFormat = utils.FWV
	case utils.MetaFileCSV:
		expFormat = utils.CSV
	default:
		expFormat = exportFormat
	}
	fileName := fmt.Sprintf("cdre_%s.%s", exportID, expFormat)
	if expFn, has := arg.ExportArgs[utils.ExportFileName]; has {
		fileName = utils.IfaceAsString(expFn)
	}
	var filePath string
	switch exportFormat {
	case utils.MetaFileFWV, utils.MetaFileCSV:
		filePath = path.Join(eDir, fileName)
	case utils.DRYRUN:
		filePath = utils.DRYRUN
	default:
		u, _ := url.Parse(eDir)
		u.Path = path.Join(u.Path, fileName)
		filePath = u.String()
	}
	cdrsFltr, err := arg.RPCCDRsFilter.AsCDRsFilter(apierSv1.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrs, _, err := apierSv1.CdrDb.GetCDRs(cdrsFltr, false)
	if err != nil {
		return err
	} else if len(cdrs) == 0 {
		return
	}
	cdrexp, err := engine.NewCDRExporter(cdrs, exportTemplate, exportFormat,
		filePath, utils.META_NONE, exportID,
		synchronous, attempts, fieldSep,
		apierSv1.Config.GeneralCfg().HttpSkipTlsVerify,
		apierSv1.Config.ApierCfg().AttributeSConns, apierSv1.FilterS)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := cdrexp.ExportCDRs(); err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrexp.TotalExportedCdrs() == 0 {
		return
	}
	*reply = RplExportedCDRs{ExportedPath: filePath, TotalRecords: len(cdrs), TotalCost: cdrexp.TotalCost(),
		FirstOrderID: cdrexp.FirstOrderID(), LastOrderID: cdrexp.LastOrderID()}
	if arg.Verbose {
		reply.ExportedCGRIDs = cdrexp.PositiveExports()
		reply.UnexportedCGRIDs = cdrexp.NegativeExports()
	}
	return nil
}
