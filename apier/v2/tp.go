/*
Real-time Charging System for Telecom & ISP environments
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
	"encoding/base64"
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (self *ApierV2) RemTP(tpid string, reply *string) error {
	if err := self.StorDb.RemTPData("", tpid); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}

func (self *ApierV2) ExportTPToFolder(attrs utils.AttrDirExportTP, exported *utils.ExportedTPStats) error {
	if len(*attrs.TPid) == 0 {
		return fmt.Errorf("%s:TPid", utils.ERR_MANDATORY_IE_MISSING)
	}
	dir := self.Config.TpExportPath
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
	tpExporter, err := engine.NewTPExporter(self.StorDb, *attrs.TPid, dir, fileFormat, sep, compress)
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if err := tpExporter.Run(); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*exported = *tpExporter.ExportStats()
	}

	return nil
}

func (self *ApierV2) ExportTPToZipString(attrs utils.AttrDirExportTP, reply *string) error {
	if len(*attrs.TPid) == 0 {
		return fmt.Errorf("%s:TPid", utils.ERR_MANDATORY_IE_MISSING)
	}
	dir := ""
	fileFormat := utils.CSV
	if attrs.FileFormat != nil {
		fileFormat = *attrs.FileFormat
	}
	sep := ","
	if attrs.FieldSeparator != nil {
		sep = *attrs.FieldSeparator
	}
	tpExporter, err := engine.NewTPExporter(self.StorDb, *attrs.TPid, dir, fileFormat, sep, true)
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if err := tpExporter.Run(); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = base64.StdEncoding.EncodeToString(tpExporter.GetCacheBuffer().Bytes())
	return nil
}
