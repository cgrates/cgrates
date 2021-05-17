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
	timezone string, exitChan chan bool, filterS *engine.FilterS,
	connMgr *engine.ConnManager, cacheConns []string) (ldr *Loader) {
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
		connMgr:       connMgr,
		cacheConns:    cacheConns,
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
	fieldSep      string
	dataTpls      map[string][]*config.FCTemplate      // map[loaderType]*config.FCTemplate
	rdrs          map[string]map[string]*openedCSVFile // map[loaderType]map[fileName]*openedCSVFile for common incremental read
	procRows      int                                  // keep here the last processed row in the file/-s
	bufLoaderData map[string][]LoaderData              // cache of data read, indexed on tenantID
	dm            *engine.DataManager
	timezone      string
	filterS       *engine.FilterS
	connMgr       *engine.ConnManager
	cacheConns    []string
}

func (ldr *Loader) ListenAndServe(exitChan chan struct{}) (err error) {
	utils.Logger.Info(fmt.Sprintf("Starting <%s-%s>", utils.LoaderS, ldr.ldrID))
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return
}

// ProcessFolder will process the content in the folder with locking
func (ldr *Loader) ProcessFolder(caching, loadOption string) (err error) {
	if err = ldr.lockFolder(); err != nil {
		return
	}
	defer ldr.unlockFolder()
	for ldrType := range ldr.rdrs {
		if err = ldr.processFiles(ldrType, caching, loadOption); err != nil {
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

func (ldr *Loader) processFiles(loaderType, caching, loadOption string) (err error) {
	for fName := range ldr.rdrs[loaderType] {
		var rdr *os.File
		if rdr, err = os.Open(path.Join(ldr.tpInDir, fName)); err != nil {
			return err
		}
		csvReader := csv.NewReader(rdr)
		csvReader.Comment = '#'
		csvReader.Comma = rune(ldr.fieldSep[0])
		ldr.rdrs[loaderType][fName] = &openedCSVFile{
			fileName: fName, rdr: rdr, csvRdr: csvReader}
		defer ldr.unreferenceFile(loaderType, fName)
	}
	// based on load option will store or remove the content
	switch loadOption {
	case utils.MetaStore:
		if err = ldr.processContent(loaderType, caching); err != nil {
			return
		}
	case utils.MetaRemove:
		if err = ldr.removeContent(loaderType, caching); err != nil {
			return
		}
	}
	return
}

//processContent will process the contect and will store it into database
func (ldr *Loader) processContent(loaderType, caching string) (err error) {
	// start processing lines
	keepLooping := true // controls looping
	lineNr := 0
	for keepLooping {
		lineNr++
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
				utils.Logger.Warning(
					fmt.Sprintf("<%s> <%s> line: %d, error: %s",
						utils.LoaderS, ldr.ldrID, lineNr, err.Error()))
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
				map[string][]LoaderData{prevTntID: ldr.bufLoaderData[prevTntID]}, caching); err != nil {
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
		map[string][]LoaderData{tntID: ldr.bufLoaderData[tntID]}, caching); err != nil {
		return
	}
	delete(ldr.bufLoaderData, tntID)
	return
}

func (ldr *Loader) storeLoadedData(loaderType string,
	lds map[string][]LoaderData, caching string) (err error) {
	var ids []string
	cacheArgs := utils.InitAttrReloadCache()
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
				// get IDs so we can reload in cache
				ids = append(ids, apf.TenantID())
				if err := ldr.dm.SetAttributeProfile(apf, true); err != nil {
					return err
				}
			}
			cacheArgs.AttributeProfileIDs = ids
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
				// get IDs so we can reload in cache
				ids = append(ids, res.TenantID())
				if err := ldr.dm.SetResourceProfile(res, true); err != nil {
					return err
				}
				if err := ldr.dm.SetResource(
					&engine.Resource{Tenant: res.Tenant,
						ID:     res.ID,
						Usages: make(map[string]*engine.ResourceUsage)}); err != nil {
					return err
				}
				cacheArgs.ResourceProfileIDs = ids
				cacheArgs.ResourceIDs = ids
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
				// get IDs so we can reload in cache
				ids = append(ids, fltrPrf.TenantID())
				if err := ldr.dm.SetFilter(fltrPrf); err != nil {
					return err
				}
				cacheArgs.FilterIDs = ids
			}
		}
	case utils.MetaStats:
		for _, lDataSet := range lds {
			stsModels := make(engine.TpStats, len(lDataSet))
			for i, ld := range lDataSet {
				stsModels[i] = new(engine.TpStat)
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
				// get IDs so we can reload in cache
				ids = append(ids, stsPrf.TenantID())
				if err := ldr.dm.SetStatQueueProfile(stsPrf, true); err != nil {
					return err
				}
				metrics := make(map[string]engine.StatMetric)
				for _, metric := range stsPrf.Metrics {
					stsMetric, err := engine.NewStatMetric(metric.MetricID, stsPrf.MinItems, metric.FilterIDs)
					if err != nil {
						return utils.APIErrorHandler(err)
					}
					metrics[metric.MetricID] = stsMetric
				}
				if err := ldr.dm.SetStatQueue(&engine.StatQueue{Tenant: stsPrf.Tenant, ID: stsPrf.ID, SQMetrics: metrics}); err != nil {
					return err
				}
				cacheArgs.StatsQueueProfileIDs = ids
				cacheArgs.StatsQueueIDs = ids
			}
		}
	case utils.MetaThresholds:
		for _, lDataSet := range lds {
			thModels := make(engine.TpThresholds, len(lDataSet))
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
				// get IDs so we can reload in cache
				ids = append(ids, thPrf.TenantID())
				if err := ldr.dm.SetThresholdProfile(thPrf, true); err != nil {
					return err
				}
				if err := ldr.dm.SetThreshold(&engine.Threshold{Tenant: thPrf.Tenant, ID: thPrf.ID}); err != nil {
					return err
				}
				cacheArgs.ThresholdProfileIDs = ids
				cacheArgs.ThresholdIDs = ids
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
				// get IDs so we can reload in cache
				ids = append(ids, spPrf.TenantID())
				if err := ldr.dm.SetSupplierProfile(spPrf, true); err != nil {
					return err
				}
				cacheArgs.SupplierProfileIDs = ids
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
				// get IDs so we can reload in cache
				ids = append(ids, cpp.TenantID())
				if err := ldr.dm.SetChargerProfile(cpp, true); err != nil {
					return err
				}
				cacheArgs.ChargerProfileIDs = ids
			}
		}
	case utils.MetaDispatchers:
		for _, lDataSet := range lds {
			dispModels := make(engine.TPDispatcherProfiles, len(lDataSet))
			for i, ld := range lDataSet {
				dispModels[i] = new(engine.TPDispatcherProfile)
				if err = utils.UpdateStructWithIfaceMap(dispModels[i], ld); err != nil {
					return
				}
			}
			for _, tpDsp := range dispModels.AsTPDispatcherProfiles() {
				dsp, err := engine.APItoDispatcherProfile(tpDsp, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: DispatcherProfile: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(dsp)))
					continue
				}
				// get IDs so we can reload in cache
				ids = append(ids, dsp.TenantID())
				if err := ldr.dm.SetDispatcherProfile(dsp, true); err != nil {
					return err
				}
				cacheArgs.DispatcherProfileIDs = ids
			}
		}
	case utils.MetaDispatcherHosts:
		for _, lDataSet := range lds {
			dispModels := make(engine.TPDispatcherHosts, len(lDataSet))
			for i, ld := range lDataSet {
				dispModels[i] = new(engine.TPDispatcherHost)
				if err = utils.UpdateStructWithIfaceMap(dispModels[i], ld); err != nil {
					return
				}
			}
			for _, tpDsp := range dispModels.AsTPDispatcherHosts() {
				dsp := engine.APItoDispatcherHost(tpDsp)
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: DispatcherHost: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(dsp)))
					continue
				}
				// get IDs so we can reload in cache
				ids = append(ids, dsp.TenantID())
				if err := ldr.dm.SetDispatcherHost(dsp); err != nil {
					return err
				}
				cacheArgs.DispatcherHostIDs = ids
			}
		}
	}

	if len(ldr.cacheConns) != 0 {
		var tnt string
		if tnt, err = ldr.tenant.ParseValue(""); err != nil {
			return
		}
		var reply string
		switch caching {
		case utils.META_NONE:
			return
		case utils.MetaReload:
			if err = ldr.connMgr.Call(ldr.cacheConns, nil,
				utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithArgDispatcher{
					TenantArg:       utils.TenantArg{Tenant: tnt},
					AttrReloadCache: cacheArgs}, &reply); err != nil {
				return
			}
		case utils.MetaLoad:
			if err = ldr.connMgr.Call(ldr.cacheConns, nil,
				utils.CacheSv1LoadCache, utils.AttrReloadCacheWithArgDispatcher{
					TenantArg:       utils.TenantArg{Tenant: tnt},
					AttrReloadCache: cacheArgs}, &reply); err != nil {
				return
			}
		case utils.MetaRemove:
			if err = ldr.connMgr.Call(ldr.cacheConns, nil,
				utils.CacheSv1FlushCache, utils.AttrReloadCacheWithArgDispatcher{
					TenantArg:       utils.TenantArg{Tenant: tnt},
					AttrReloadCache: cacheArgs}, &reply); err != nil {
				return
			}
		case utils.MetaClear:
			cacheArgs.FlushAll = true
			if err = ldr.connMgr.Call(ldr.cacheConns, nil,
				utils.CacheSv1FlushCache, utils.AttrReloadCacheWithArgDispatcher{
					TenantArg:       utils.TenantArg{Tenant: tnt},
					AttrReloadCache: cacheArgs}, &reply); err != nil {
				return
			}
		}
	}
	return
}

//removeContent will process the contect and will remove it from database
func (ldr *Loader) removeContent(loaderType, caching string) (err error) {
	// start processing lines
	keepLooping := true // controls looping
	lineNr := 0
	for keepLooping {
		lineNr++
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
				utils.Logger.Warning(
					fmt.Sprintf("<%s> <%s> line: %d, error: %s",
						utils.LoaderS, ldr.ldrID, lineNr, err.Error()))
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
			if err = ldr.removeLoadedData(loaderType, prevTntID, caching); err != nil {
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
	if err = ldr.removeLoadedData(loaderType, tntID, caching); err != nil {
		return
	}
	delete(ldr.bufLoaderData, tntID)
	return
}

//removeLoadedData will remove the data from database
//since we remove we don't need to compose the struct we only need the Tenant and the ID of the profile
func (ldr *Loader) removeLoadedData(loaderType, tntID, caching string) (err error) {
	var ids []string
	cacheArgs := utils.InitAttrReloadCache()
	switch loaderType {
	case utils.MetaAttributes:
		if ldr.dryRun {
			utils.Logger.Info(
				fmt.Sprintf("<%s-%s> DRY_RUN: AttributeProfileID: %s",
					utils.LoaderS, ldr.ldrID, tntID))
		} else {
			tntIDStruct := utils.NewTenantID(tntID)
			// get IDs so we can reload in cache
			ids = append(ids, tntID)
			if err := ldr.dm.RemoveAttributeProfile(tntIDStruct.Tenant, tntIDStruct.ID,
				utils.NonTransactional, true); err != nil {
				return err
			}
			cacheArgs.AttributeProfileIDs = ids
		}
	case utils.MetaResources:
		if ldr.dryRun {
			utils.Logger.Info(
				fmt.Sprintf("<%s-%s> DRY_RUN: ResourceProfileID: %s",
					utils.LoaderS, ldr.ldrID, tntID))

		} else {
			tntIDStruct := utils.NewTenantID(tntID)
			// get IDs so we can reload in cache
			ids = append(ids, tntID)
			if err := ldr.dm.RemoveResourceProfile(tntIDStruct.Tenant,
				tntIDStruct.ID, utils.NonTransactional, true); err != nil {
				return err
			}
			if err := ldr.dm.RemoveResource(tntIDStruct.Tenant, tntIDStruct.ID, utils.NonTransactional); err != nil {
				return err
			}
			cacheArgs.ResourceProfileIDs = ids
			cacheArgs.ResourceIDs = ids
		}

	case utils.MetaFilters:
		if ldr.dryRun {
			utils.Logger.Info(
				fmt.Sprintf("<%s-%s> DRY_RUN: Filter: %s",
					utils.LoaderS, ldr.ldrID, tntID))
		} else {
			tntIDStruct := utils.NewTenantID(tntID)
			// get IDs so we can reload in cache
			ids = append(ids, tntID)
			if err := ldr.dm.RemoveFilter(tntIDStruct.Tenant, tntIDStruct.ID, utils.NonTransactional); err != nil {
				return err
			}
			cacheArgs.FilterIDs = ids
		}
	case utils.MetaStats:
		if ldr.dryRun {
			utils.Logger.Info(
				fmt.Sprintf("<%s-%s> DRY_RUN: StatsQueueProfileID: %s",
					utils.LoaderS, ldr.ldrID, tntID))
		} else {
			tntIDStruct := utils.NewTenantID(tntID)
			// get IDs so we can reload in cache
			ids = append(ids, tntID)
			if err := ldr.dm.RemoveStatQueueProfile(tntIDStruct.Tenant,
				tntIDStruct.ID, utils.NonTransactional, true); err != nil {
				return err
			}
			if err := ldr.dm.RemoveStatQueue(tntIDStruct.Tenant, tntIDStruct.ID, utils.NonTransactional); err != nil {
				return err
			}
			cacheArgs.StatsQueueProfileIDs = ids
			cacheArgs.StatsQueueIDs = ids
		}
	case utils.MetaThresholds:
		if ldr.dryRun {
			utils.Logger.Info(
				fmt.Sprintf("<%s-%s> DRY_RUN: ThresholdProfileID: %s",
					utils.LoaderS, ldr.ldrID, tntID))
		} else {
			tntIDStruct := utils.NewTenantID(tntID)
			// get IDs so we can reload in cache
			ids = append(ids, tntID)
			if err := ldr.dm.RemoveThresholdProfile(tntIDStruct.Tenant,
				tntIDStruct.ID, utils.NonTransactional, true); err != nil {
				return err
			}
			if err := ldr.dm.RemoveThreshold(tntIDStruct.Tenant, tntIDStruct.ID, utils.NonTransactional); err != nil {
				return err
			}
			cacheArgs.ThresholdProfileIDs = ids
			cacheArgs.ThresholdIDs = ids
		}
	case utils.MetaSuppliers:
		if ldr.dryRun {
			utils.Logger.Info(
				fmt.Sprintf("<%s-%s> DRY_RUN: SupplierProfileID: %s",
					utils.LoaderS, ldr.ldrID, tntID))
		} else {
			tntIDStruct := utils.NewTenantID(tntID)
			// get IDs so we can reload in cache
			ids = append(ids, tntID)
			if err := ldr.dm.RemoveSupplierProfile(tntIDStruct.Tenant,
				tntIDStruct.ID, utils.NonTransactional, true); err != nil {
				return err
			}
			cacheArgs.SupplierProfileIDs = ids
		}
	case utils.MetaChargers:
		if ldr.dryRun {
			utils.Logger.Info(
				fmt.Sprintf("<%s-%s> DRY_RUN: ChargerProfileID: %s",
					utils.LoaderS, ldr.ldrID, tntID))
		} else {
			tntIDStruct := utils.NewTenantID(tntID)
			// get IDs so we can reload in cache
			ids = append(ids, tntID)
			if err := ldr.dm.RemoveChargerProfile(tntIDStruct.Tenant,
				tntIDStruct.ID, utils.NonTransactional, true); err != nil {
				return err
			}
			cacheArgs.ChargerProfileIDs = ids
		}
	case utils.MetaDispatchers:
		if ldr.dryRun {
			utils.Logger.Info(
				fmt.Sprintf("<%s-%s> DRY_RUN: DispatcherProfileID: %s",
					utils.LoaderS, ldr.ldrID, tntID))
		} else {
			tntIDStruct := utils.NewTenantID(tntID)
			// get IDs so we can reload in cache
			ids = append(ids, tntID)
			if err := ldr.dm.RemoveDispatcherProfile(tntIDStruct.Tenant,
				tntIDStruct.ID, utils.NonTransactional, true); err != nil {
				return err
			}
			cacheArgs.DispatcherProfileIDs = ids
		}
	case utils.MetaDispatcherHosts:
		if ldr.dryRun {
			utils.Logger.Info(
				fmt.Sprintf("<%s-%s> DRY_RUN: DispatcherHostID: %s",
					utils.LoaderS, ldr.ldrID, tntID))
		} else {
			tntIDStruct := utils.NewTenantID(tntID)
			// get IDs so we can reload in cache
			ids = append(ids, tntID)
			if err := ldr.dm.RemoveDispatcherHost(tntIDStruct.Tenant,
				tntIDStruct.ID, utils.NonTransactional); err != nil {
				return err
			}
			cacheArgs.DispatcherHostIDs = ids
		}
	}

	if len(ldr.cacheConns) != 0 {
		var tnt string
		if tnt, err = ldr.tenant.ParseValue(""); err != nil {
			return
		}
		var reply string
		switch caching {
		case utils.META_NONE:
			return
		case utils.MetaReload:
			if err = ldr.connMgr.Call(ldr.cacheConns, nil,
				utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithArgDispatcher{
					TenantArg:       utils.TenantArg{Tenant: tnt},
					AttrReloadCache: cacheArgs}, &reply); err != nil {
				return
			}
		case utils.MetaLoad:
			if err = ldr.connMgr.Call(ldr.cacheConns, nil,
				utils.CacheSv1LoadCache, utils.AttrReloadCacheWithArgDispatcher{
					TenantArg:       utils.TenantArg{Tenant: tnt},
					AttrReloadCache: cacheArgs}, &reply); err != nil {
				return
			}
		case utils.MetaRemove:
			if err = ldr.connMgr.Call(ldr.cacheConns, nil,
				utils.CacheSv1FlushCache, utils.AttrReloadCacheWithArgDispatcher{
					TenantArg:       utils.TenantArg{Tenant: tnt},
					AttrReloadCache: cacheArgs}, &reply); err != nil {
				return
			}
		case utils.MetaClear:
			cacheArgs.FlushAll = true
			if err = ldr.connMgr.Call(ldr.cacheConns, nil,
				utils.CacheSv1FlushCache, utils.AttrReloadCacheWithArgDispatcher{
					TenantArg:       utils.TenantArg{Tenant: tnt},
					AttrReloadCache: cacheArgs}, &reply); err != nil {
				return
			}
		}
	}
	return
}
