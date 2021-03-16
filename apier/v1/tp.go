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

// Tariff plan related APIs

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrGetTPIds struct {
}

// Queries tarrif plan identities gathered from all tables.
func (apierSv1 *APIerSv1) GetTPIds(attrs *AttrGetTPIds, reply *[]string) error {
	if ids, err := apierSv1.StorDb.GetTpIds(utils.EmptyString); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

type AttrImportTPZipFile struct {
	TPid string
	File []byte
}

func (apierSv1 *APIerSv1) ImportTPZipFile(attrs *AttrImportTPZipFile, reply *string) error {
	tmpDir, err := os.MkdirTemp("/tmp", "cgr_")
	if err != nil {
		*reply = "ERROR: creating temp directory!"
		return err
	}
	zipFile := filepath.Join(tmpDir, "/file.zip")
	if err = os.WriteFile(zipFile, attrs.File, os.ModePerm); err != nil {
		*reply = "ERROR: writing zip file!"
		return err
	}
	if err = utils.Unzip(zipFile, tmpDir); err != nil {
		*reply = "ERROR: unziping file!"
		return err
	}
	csvfilesFound := false
	if err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		csvFiles, err := filepath.Glob(filepath.Join(path, "*csv"))
		if csvFiles != nil {
			if attrs.TPid == "" {
				*reply = "ERROR: missing TPid!"
				return err
			}
			csvImporter := engine.TPCSVImporter{
				TPid:     attrs.TPid,
				StorDb:   apierSv1.StorDb,
				DirPath:  path,
				Sep:      utils.CSVSep,
				Verbose:  false,
				ImportId: "",
			}
			if errImport := csvImporter.Run(); errImport != nil {
				return errImport
			}
			csvfilesFound = true
		}
		return err
	}); err != nil || !csvfilesFound {
		*reply = "ERROR: finding csv files!"
		return err
	}
	os.RemoveAll(tmpDir)
	*reply = utils.OK
	return nil
}

type AttrRemTp struct {
	TPid string
}

func (apierSv1 *APIerSv1) RemTP(attrs *AttrRemTp, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing(utils.TPid)
	}
	if err := apierSv1.StorDb.RemTpData("", attrs.TPid, nil); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}

func (apierSv1 *APIerSv1) ExportTPToFolder(attrs *utils.AttrDirExportTP, exported *utils.ExportedTPStats) error {
	if attrs.TPid == nil || *attrs.TPid == "" {
		return utils.NewErrMandatoryIeMissing(utils.TPid)
	}
	dir := apierSv1.Config.GeneralCfg().TpExportPath
	if attrs.ExportPath != nil {
		dir = *attrs.ExportPath
	}
	fileFormat := utils.CSV
	if attrs.FileFormat != nil {
		fileFormat = *attrs.FileFormat
	}
	sep := ","
	if attrs.FieldSeparator != nil {
		sep = *attrs.FieldSeparator
	}
	compress := false
	if attrs.Compress != nil {
		compress = *attrs.Compress
	}
	tpExporter, err := engine.NewTPExporter(apierSv1.StorDb, *attrs.TPid, dir, fileFormat, sep, compress)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := tpExporter.Run(); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*exported = *tpExporter.ExportStats()
	}

	return nil
}

func (apierSv1 *APIerSv1) ExportTPToZipString(attrs *utils.AttrDirExportTP, reply *string) error {
	if attrs.TPid == nil || *attrs.TPid == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.TPid)
	}
	dir := utils.EmptyString
	fileFormat := utils.CSV
	if attrs.FileFormat != nil {
		fileFormat = *attrs.FileFormat
	}
	sep := ","
	if attrs.FieldSeparator != nil {
		sep = *attrs.FieldSeparator
	}
	tpExporter, err := engine.NewTPExporter(apierSv1.StorDb, *attrs.TPid, dir, fileFormat, sep, true)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := tpExporter.Run(); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = base64.StdEncoding.EncodeToString(tpExporter.GetCacheBuffer().Bytes())
	return nil
}
