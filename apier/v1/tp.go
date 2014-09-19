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

package v1

// Tariff plan related APIs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrGetTPIds struct {
}

// Queries tarrif plan identities gathered from all tables.
func (self *ApierV1) GetTPIds(attrs AttrGetTPIds, reply *[]string) error {
	if ids, err := self.StorDb.GetTPIds(); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

type AttrImportTPZipFile struct {
	TPid string
	File []byte
}

func (self *ApierV1) ImportTPZipFile(attrs AttrImportTPZipFile, reply *string) error {
	tmpDir, err := ioutil.TempDir("/tmp", "cgr_")
	if err != nil {
		*reply = "ERROR: creating temp directory!"
		return err
	}
	zipFile := filepath.Join(tmpDir, "/file.zip")
	if err = ioutil.WriteFile(zipFile, attrs.File, os.ModePerm); err != nil {
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
			csvImporter := engine.TPCSVImporter{attrs.TPid, self.StorDb, path, ',', false, ""}
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
	*reply = "OK"
	return nil
}
