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

func NewLoader(dm *engine.DataManager, cfg *config.LoaderSCfg,
	timezone string, filterS *engine.FilterS) (ldr *Loader) {
	ldr = &Loader{
		enabled:       cfg.Enabled,
		tenant:        cfg.Tenant,
		dryRun:        cfg.DryRun,
		ldrID:         cfg.Id,
		tpInDir:       cfg.TpInDir,
		tpOutDir:      cfg.TpOutDir,
		lockFilename:  cfg.LockFileName,
		fieldSep:      cfg.FieldSeparator,
		dataTpls:      make(map[string][]*config.FCTemplate),
		rdrs:          make(map[string]map[string]*openedCSVFile),
		bufLoaderData: make(map[string][]LoaderData),
		dm:            dm,
		timezone:      timezone,
		filterS:       filterS,
	}
	for _, ldrData := range cfg.Data {
		ldr.dataTpls[ldrData.Type] = ldrData.Fields
		ldr.rdrs[ldrData.Type] = make(map[string]*openedCSVFile)
		if ldrData.Filename != "" {
			ldr.rdrs[ldrData.Type][ldrData.Filename] = nil
		}
		for _, cfgFld := range ldrData.Fields { // add all possible files to be opened
			for _, cfgFldVal := range cfgFld.Value {
				if idx := strings.Index(cfgFldVal.Rules, utils.InInFieldSep); idx != -1 {
					ldr.rdrs[ldrData.Type][cfgFldVal.Rules[:idx]] = nil
				}
			}
		}
	}
	return
}

// Loader is one instance loading from a folder
type Loader struct {
	enabled       bool
	tenant        config.RSRParsers
	dryRun        bool
	ldrID         string
	tpInDir       string
	tpOutDir      string
	lockFilename  string
	cacheSConns   []*config.HaPoolConfig
	fieldSep      string
	dataTpls      map[string][]*config.FCTemplate      // map[loaderType]*config.FCTemplate
	rdrs          map[string]map[string]*openedCSVFile // map[loaderType]map[fileName]*openedCSVFile for common incremental read
	procRows      int                                  // keep here the last processed row in the file/-s
	bufLoaderData map[string][]LoaderData              // cache of data read, indexed on tenantID
	dm            *engine.DataManager
	timezone      string
	filterS       *engine.FilterS
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
				ldr.dataTpls[loaderType], ldr.tenant, ldr.filterS); err != nil {
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
	case utils.MetaResources:
		for _, lDataSet := range lds {
			resModels := make(engine.TpResources, len(lDataSet))
			for i, ld := range lDataSet {
				resModels[i] = new(engine.TpResource)
				if err = utils.UpdateStructWithIfaceMap(resModels[i], ld); err != nil {
					return
				}
			}

			for _, tpRes := range resModels.AsTPResources() {
				res, err := engine.APItoResource(tpRes, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: ResourceProfile: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(res)))
					continue
				}
				if err := ldr.dm.SetResourceProfile(res, true); err != nil {
					return err
				}
				if err := ldr.dm.SetResource(
					&engine.Resource{Tenant: res.Tenant,
						ID:     res.ID,
						Usages: make(map[string]*engine.ResourceUsage)}); err != nil {
					return err
				}
			}
		}
	case utils.MetaFilters:
		for _, lDataSet := range lds {
			fltrModels := make(engine.TpFilterS, len(lDataSet))
			for i, ld := range lDataSet {
				fltrModels[i] = new(engine.TpFilter)
				if err = utils.UpdateStructWithIfaceMap(fltrModels[i], ld); err != nil {
					return
				}
			}

			for _, tpFltr := range fltrModels.AsTPFilter() {
				fltrPrf, err := engine.APItoFilter(tpFltr, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: Filter: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(fltrPrf)))
					continue
				}
				if err := ldr.dm.SetFilter(fltrPrf); err != nil {
					return err
				}
			}
		}
	case utils.MetaStats:
		for _, lDataSet := range lds {
			stsModels := make(engine.TpStatsS, len(lDataSet))
			for i, ld := range lDataSet {
				stsModels[i] = new(engine.TpStats)
				if err = utils.UpdateStructWithIfaceMap(stsModels[i], ld); err != nil {
					return
				}
			}
			for _, tpSts := range stsModels.AsTPStats() {
				stsPrf, err := engine.APItoStats(tpSts, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: StatsQueueProfile: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(stsPrf)))
					continue
				}
				if err := ldr.dm.SetStatQueueProfile(stsPrf, true); err != nil {
					return err
				}
				metrics := make(map[string]engine.StatMetric)
				for _, metricwithparam := range stsPrf.Metrics {
					if metric, err := engine.NewStatMetric(metricwithparam.MetricID, stsPrf.MinItems, metricwithparam.Parameters); err != nil {
						return utils.APIErrorHandler(err)
					} else {
						metrics[metricwithparam.MetricID] = metric
					}
				}
				if err := ldr.dm.SetStatQueue(&engine.StatQueue{Tenant: stsPrf.Tenant, ID: stsPrf.ID, SQMetrics: metrics}); err != nil {
					return err
				}
			}
		}
	case utils.MetaThresholds:
		for _, lDataSet := range lds {
			thModels := make(engine.TpThresholdS, len(lDataSet))
			for i, ld := range lDataSet {
				thModels[i] = new(engine.TpThreshold)
				if err = utils.UpdateStructWithIfaceMap(thModels[i], ld); err != nil {
					return
				}
			}

			for _, tpTh := range thModels.AsTPThreshold() {
				thPrf, err := engine.APItoThresholdProfile(tpTh, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: ThresholdProfile: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(thPrf)))
					continue
				}
				if err := ldr.dm.SetThresholdProfile(thPrf, true); err != nil {
					return err
				}
				if err := ldr.dm.SetThreshold(&engine.Threshold{Tenant: thPrf.Tenant, ID: thPrf.ID}); err != nil {
					return err
				}
			}
		}
	case utils.MetaSuppliers:
		for _, lDataSet := range lds {
			sppModels := make(engine.TpSuppliers, len(lDataSet))
			for i, ld := range lDataSet {
				sppModels[i] = new(engine.TpSupplier)
				if err = utils.UpdateStructWithIfaceMap(sppModels[i], ld); err != nil {
					return
				}
			}

			for _, tpSpp := range sppModels.AsTPSuppliers() {
				spPrf, err := engine.APItoSupplierProfile(tpSpp, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: SupplierProfile: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(spPrf)))
					continue
				}
				if err := ldr.dm.SetSupplierProfile(spPrf, true); err != nil {
					return err
				}
			}
		}
	case utils.MetaChargers:
		for _, lDataSet := range lds {
			cppModels := make(engine.TPChargers, len(lDataSet))
			for i, ld := range lDataSet {
				cppModels[i] = new(engine.TPCharger)
				if err = utils.UpdateStructWithIfaceMap(cppModels[i], ld); err != nil {
					return
				}
			}

			for _, tpCPP := range cppModels.AsTPChargers() {
				cpp, err := engine.APItoChargerProfile(tpCPP, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: ChargerProfile: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(cpp)))
					continue
				}
				if err := ldr.dm.SetChargerProfile(cpp, true); err != nil {
					return err
				}
			}
		}
	}
	return
}
