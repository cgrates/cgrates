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

package loaders

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type openedCSVFile struct {
	fileName string
	fd       *os.File
	csvRdr   *csv.Reader
}

// Loader is one instance loading from a folder
type Loader struct {
	tpInDir      string
	tpOutDir     string
	lockFilename string
	cacheSConns  []*config.HaPoolConfig
	fieldSep     string
	dataTpls     []*config.LoaderSDataType
	rdrs         map[string]map[string]*openedCSVFile // map[loaderType]map[fileName]*openedCSVFile for common incremental read
}

// lockFolder will attempt to lock the folder by creating the lock file
func (ldr *Loader) lockFolder() (err error) {
	_, err = os.OpenFile(path.Join(ldr.tpInDir, ldr.lockFilename),
		os.O_RDONLY|os.O_CREATE, 0644)
	return
}

func (ldr *Loader) unlockFolder() (err error) {
	return os.Remove(path.Join(ldr.tpInDir,
		ldr.lockFilename))
}

// ProcessFolder will process the content in the folder with locking
func (ldr *Loader) ProcessFolder() (err error) {
	if err = ldr.lockFolder(); err != nil {
		return
	}
	defer ldr.unlockFolder()
	for loaderType := range ldr.rdrs {
		switch loaderType {
		case utils.MetaAttributes:
			if err = ldr.processAttributes(); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> loaderType: <%s>, err: %s",
					utils.LoaderS, loaderType, err.Error()))
			}
		default:
			utils.Logger.Warning(fmt.Sprintf("<%s> unsupported loaderType: <%s>",
				utils.LoaderS, loaderType))
		}
	}
	return
}

// unreferenceFile will cleanup an used file by closing and removing from referece map
func (ldr *Loader) unreferenceFile(loaderType, fileName string) (err error) {
	openedCSVFile := ldr.rdrs[loaderType][fileName]
	ldr.rdrs[loaderType][fileName] = nil
	return openedCSVFile.fd.Close()
}

// processAttributes contains the procedure for loading Attributes
func (ldr *Loader) processAttributes() (err error) {
	// open files as csv readers
	for fName := range ldr.rdrs[utils.MetaAttributes] {
		var fd *os.File
		if fd, err = os.Open(path.Join(ldr.tpInDir, fName)); err != nil {
			return err
		}
		ldr.rdrs[utils.MetaAttributes][fName] = &openedCSVFile{
			fileName: fName, fd: fd, csvRdr: csv.NewReader(fd)}
		defer ldr.unreferenceFile(utils.MetaAttributes, fName)
	}
	// start processing lines
	return
}
