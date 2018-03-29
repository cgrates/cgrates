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
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type openedCSVFile struct {
	fileName string
	rdr      io.ReadCloser // keep reference so we can close it when done
	csvRdr   *csv.Reader
}

func NewLoader(dm *engine.DataManager, cfg *config.LoaderSConfig,
	timezone string) (ldr *Loader) {
	ldr = &Loader{
		enabled:       cfg.Enabled,
		dryRun:        cfg.DryRun,
		ldrID:         cfg.Id,
		tpInDir:       cfg.TpInDir,
		tpOutDir:      cfg.TpOutDir,
		lockFilename:  cfg.LockFileName,
		fieldSep:      cfg.FieldSeparator,
		dataTpls:      make(map[string][]*config.CfgCdrField),
		rdrs:          make(map[string]map[string]*openedCSVFile),
		bufLoaderData: make(map[string][]LoaderData),
		dm:            dm,
		timezone:      timezone,
	}
	for _, ldrData := range cfg.Data {
		ldr.dataTpls[ldrData.Type] = ldrData.Fields
		ldr.rdrs[ldrData.Type] = make(map[string]*openedCSVFile)
		if ldrData.Filename != "" {
			ldr.rdrs[ldrData.Type][ldrData.Filename] = nil
		}
		for _, cfgFld := range ldrData.Fields { // add all possible files to be opened
			for _, cfgFldVal := range cfgFld.Value {
				if idx := strings.Index(cfgFldVal.Id, utils.InInFieldSep); idx != -1 {
					ldr.rdrs[ldrData.Type][cfgFldVal.Id[:idx]] = nil
				}
			}
		}
	}
	return
}

// Loader is one instance loading from a folder
type Loader struct {
	enabled       bool
	dryRun        bool
	ldrID         string
	tpInDir       string
	tpOutDir      string
	lockFilename  string
	cacheSConns   []*config.HaPoolConfig
	fieldSep      string
	dataTpls      map[string][]*config.CfgCdrField     // map[loaderType]*config.CfgCdrField
	rdrs          map[string]map[string]*openedCSVFile // map[loaderType]map[fileName]*openedCSVFile for common incremental read
	procRows      int                                  // keep here the last processed row in the file/-s
	bufLoaderData map[string][]LoaderData              // cache of data read, indexed on tenantID
	dm            *engine.DataManager
	timezone      string
}

func (ldr *Loader) ListenAndServe(exitChan chan struct{}) (err error) {
	utils.Logger.Info(fmt.Sprintf("Starting <%s-%s>", utils.LoaderS, ldr.ldrID))
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return
}

// ProcessFolder will process the content in the folder with locking
func (ldr *Loader) ProcessFolder() (err error) {
	if err = ldr.lockFolder(); err != nil {
		return
	}
	defer ldr.unlockFolder()
	for ldrType := range ldr.rdrs {
		if err = ldr.processFiles(ldrType); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s-%s> loaderType: <%s> cannot open files, err: %s",
				utils.LoaderS, ldr.ldrID, ldrType, err.Error()))
			continue
		}
	}
	return ldr.moveFiles()
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

func (ldr *Loader) isFolderLocked() (locked bool, err error) {
	if _, err = os.Stat(path.Join(ldr.tpInDir,
		ldr.lockFilename)); err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return
}

// unreferenceFile will cleanup an used file by closing and removing from referece map
func (ldr *Loader) unreferenceFile(loaderType, fileName string) (err error) {
	openedCSVFile := ldr.rdrs[loaderType][fileName]
	ldr.rdrs[loaderType][fileName] = nil
	return openedCSVFile.rdr.Close()
}

func (ldr *Loader) moveFiles() (err error) {
	filesInDir, _ := ioutil.ReadDir(ldr.tpInDir)
	for _, file := range filesInDir {
		fName := file.Name()
		if fName == ldr.lockFilename {
			continue
		}
		oldPath := path.Join(ldr.tpInDir, fName)
		newPath := path.Join(ldr.tpOutDir, fName)
		if err = os.Rename(oldPath, newPath); err != nil {
			return
		}
	}
	return
}

func (ldr *Loader) processFiles(loaderType string) (err error) {
	for fName := range ldr.rdrs[loaderType] {
		var rdr *os.File
		if rdr, err = os.Open(path.Join(ldr.tpInDir, fName)); err != nil {
			return err
		}
		csvReader := csv.NewReader(rdr)
		csvReader.Comment = '#'
		ldr.rdrs[loaderType][fName] = &openedCSVFile{
			fileName: fName, rdr: rdr, csvRdr: csvReader}
		defer ldr.unreferenceFile(loaderType, fName)
		if err = ldr.processContent(loaderType); err != nil {
			return
		}
	}
	return
}

func (ldr *Loader) processContent(loaderType string) (err error) {
	// start processing lines
	keepLooping := true // controls looping
	lineNr := 0
	for keepLooping {
		lineNr += 1
		var hasErrors bool
		lData := make(LoaderData) // one row
		for fName, rdr := range ldr.rdrs[loaderType] {
			var record []string
			if record, err = rdr.csvRdr.Read(); err != nil {
				if err == io.EOF {
					keepLooping = false
					break
				}
				hasErrors = true
				utils.Logger.Warning(
					fmt.Sprintf("<%s> <%s> reading line: %d, error: %s",
						utils.LoaderS, ldr.ldrID, lineNr, err.Error()))
			}
			if hasErrors { // if any of the readers will give errors, we ignore the line
				continue
			}
			if err := lData.UpdateFromCSV(fName, record,
				ldr.dataTpls[utils.MetaAttributes]); err != nil {
				fmt.Sprintf("<%s> <%s> line: %d, error: %s",
					utils.LoaderS, ldr.ldrID, lineNr, err.Error())
				hasErrors = true
				continue
			}
			// Record from map
			// update dataDB
		}
		if len(lData) == 0 { // no data, could be the last line in file
			continue
		}
		tntID := lData.TenantID()
		if _, has := ldr.bufLoaderData[tntID]; !has &&
			len(ldr.bufLoaderData) == 1 { // process previous records before going futher
			var prevTntID string
			for prevTntID = range ldr.bufLoaderData {
				break // have stolen the existing key in buffer
			}
			if err = ldr.storeLoadedData(loaderType,
				map[string][]LoaderData{prevTntID: ldr.bufLoaderData[prevTntID]}); err != nil {
				return
			}
			delete(ldr.bufLoaderData, prevTntID)
		}
		ldr.bufLoaderData[tntID] = append(ldr.bufLoaderData[tntID], lData)
	}
	// proceed with last element in bufLoaderData
	var tntID string
	for tntID = range ldr.bufLoaderData {
		break // get the first tenantID
	}
	if err = ldr.storeLoadedData(loaderType,
		map[string][]LoaderData{tntID: ldr.bufLoaderData[tntID]}); err != nil {
		return
	}
	delete(ldr.bufLoaderData, tntID)
	return
}

func (ldr *Loader) storeLoadedData(loaderType string,
	lds map[string][]LoaderData) (err error) {
	switch loaderType {
	case utils.MetaAttributes:
		for _, lDataSet := range lds {
			attrModels := make(engine.TPAttributes, len(lDataSet))
			for i, ld := range lDataSet {
				attrModels[i] = new(engine.TPAttribute)
				if err = utils.UpdateStructWithIfaceMap(attrModels[i], ld); err != nil {
					return
				}
			}
			for _, tpApf := range attrModels.AsTPAttributes() {
				apf, err := engine.APItoAttributeProfile(tpApf, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: AttributeProfile: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(apf)))
					continue
				}
				if err := ldr.dm.SetAttributeProfile(apf, true); err != nil {
					return err
				}
			}
		}
	}
	return
}
