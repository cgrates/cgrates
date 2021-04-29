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
	"os"
	"path"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
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
	timezone string, filterS *engine.FilterS,
	connMgr *engine.ConnManager, cacheConns []string) (ldr *Loader) {
	ldr = &Loader{
		enabled:       cfg.Enabled,
		tenant:        cfg.Tenant,
		dryRun:        cfg.DryRun,
		ldrID:         cfg.ID,
		tpInDir:       cfg.TpInDir,
		tpOutDir:      cfg.TpOutDir,
		lockFilename:  cfg.LockFileName,
		fieldSep:      cfg.FieldSeparator,
		runDelay:      cfg.RunDelay,
		dataTpls:      make(map[string][]*config.FCTemplate),
		flagsTpls:     make(map[string]utils.FlagsWithParams),
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
		ldr.flagsTpls[ldrData.Type] = ldrData.Flags
		ldr.rdrs[ldrData.Type] = make(map[string]*openedCSVFile)
		if ldrData.Filename != "" {
			ldr.rdrs[ldrData.Type][ldrData.Filename] = nil
		}
		for _, cfgFld := range ldrData.Fields { // add all possible files to be opened
			for _, cfgFldVal := range cfgFld.Value {
				rule := cfgFldVal.Rules
				if !strings.HasPrefix(rule, utils.DynamicDataPrefix+utils.MetaFile+utils.FilterValStart) {
					continue
				}
				if idxEnd := strings.Index(rule, utils.FilterValEnd); idxEnd != -1 {
					ldr.rdrs[ldrData.Type][rule[7:idxEnd]] = nil
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
	runDelay      time.Duration
	dataTpls      map[string][]*config.FCTemplate      // map[loaderType]*config.FCTemplate
	flagsTpls     map[string]utils.FlagsWithParams     //map[loaderType]utils.FlagsWithParams
	rdrs          map[string]map[string]*openedCSVFile // map[loaderType]map[fileName]*openedCSVFile for common incremental read
	procRows      int                                  // keep here the last processed row in the file/-s
	bufLoaderData map[string][]LoaderData              // cache of data read, indexed on tenantID
	dm            *engine.DataManager
	timezone      string
	filterS       *engine.FilterS
	connMgr       *engine.ConnManager
	cacheConns    []string
}

func (ldr *Loader) ListenAndServe(stopChan chan struct{}) (err error) {
	utils.Logger.Info(fmt.Sprintf("Starting <%s-%s>", utils.LoaderS, ldr.ldrID))
	return ldr.serve(stopChan)
}

// ProcessFolder will process the content in the folder with locking
func (ldr *Loader) ProcessFolder(caching, loadOption string, stopOnError bool) (err error) {
	if err = ldr.lockFolder(); err != nil {
		return
	}
	defer ldr.unlockFolder()
	for ldrType := range ldr.rdrs {
		if err = ldr.processFiles(ldrType, caching, loadOption); err != nil {
			if stopOnError {
				return
			}
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
	if ldr.tpOutDir == utils.EmptyString {
		return
	}
	filesInDir, _ := os.ReadDir(ldr.tpInDir)
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
		csvReader.Comma = rune(ldr.fieldSep[0])
		csvReader.Comment = '#'
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
	cacheArgs := make(map[string][]string)
	var cacheIDs []string // verify if we need to clear indexe
	switch loaderType {
	case utils.MetaAttributes:
		cacheIDs = []string{utils.CacheAttributeFilterIndexes}
		for _, lDataSet := range lds {
			attrModels := make(engine.AttributeMdls, len(lDataSet))
			for i, ld := range lDataSet {
				attrModels[i] = new(engine.AttributeMdl)
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
				if err := ldr.dm.SetAttributeProfile(context.TODO(), apf, true); err != nil {
					return err
				}
			}
			cacheArgs[utils.AttributeProfileIDs] = ids
		}
	case utils.MetaResources:
		cacheIDs = []string{utils.CacheResourceFilterIndexes}
		for _, lDataSet := range lds {
			resModels := make(engine.ResourceMdls, len(lDataSet))
			for i, ld := range lDataSet {
				resModels[i] = new(engine.ResourceMdl)
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
				var ttl *time.Duration
				if res.UsageTTL > 0 {
					ttl = &res.UsageTTL
				}
				// for non stored we do not save the resource
				if err := ldr.dm.SetResource(
					&engine.Resource{
						Tenant: res.Tenant,
						ID:     res.ID,
						Usages: make(map[string]*engine.ResourceUsage),
					}, ttl, res.Limit, !res.Stored); err != nil {
					return err
				}
				cacheArgs[utils.ResourceProfileIDs] = ids
				cacheArgs[utils.ResourceIDs] = ids
			}
		}
	case utils.MetaFilters:
		for _, lDataSet := range lds {
			fltrModels := make(engine.FilterMdls, len(lDataSet))
			for i, ld := range lDataSet {
				fltrModels[i] = new(engine.FilterMdl)
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
				if err := ldr.dm.SetFilter(fltrPrf, true); err != nil {
					return err
				}
				cacheArgs[utils.FilterIDs] = ids
			}
		}
	case utils.MetaStats:
		cacheIDs = []string{utils.CacheStatFilterIndexes}
		for _, lDataSet := range lds {
			stsModels := make(engine.StatMdls, len(lDataSet))
			for i, ld := range lDataSet {
				stsModels[i] = new(engine.StatMdl)
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
				var sq *engine.StatQueue
				if sq, err = engine.NewStatQueue(stsPrf.Tenant, stsPrf.ID, stsPrf.Metrics,
					stsPrf.MinItems); err != nil {
					return utils.APIErrorHandler(err)
				}
				var ttl *time.Duration
				if stsPrf.TTL > 0 {
					ttl = &stsPrf.TTL
				}

				// for non stored we do not save the metrics
				if err := ldr.dm.SetStatQueue(sq, stsPrf.Metrics,
					stsPrf.MinItems, ttl, stsPrf.QueueLength,
					!stsPrf.Stored); err != nil {
					return err
				}
				cacheArgs[utils.StatsQueueProfileIDs] = ids
				cacheArgs[utils.StatsQueueIDs] = ids
			}
		}
	case utils.MetaThresholds:
		cacheIDs = []string{utils.CacheThresholdFilterIndexes}
		for _, lDataSet := range lds {
			thModels := make(engine.ThresholdMdls, len(lDataSet))
			for i, ld := range lDataSet {
				thModels[i] = new(engine.ThresholdMdl)
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
				if err := ldr.dm.SetThreshold(&engine.Threshold{Tenant: thPrf.Tenant, ID: thPrf.ID}, thPrf.MinSleep, false); err != nil {
					return err
				}
				cacheArgs[utils.ThresholdProfileIDs] = ids
				cacheArgs[utils.ThresholdIDs] = ids
			}
		}
	case utils.MetaRoutes:
		cacheIDs = []string{utils.CacheRouteFilterIndexes}
		for _, lDataSet := range lds {
			sppModels := make(engine.RouteMdls, len(lDataSet))
			for i, ld := range lDataSet {
				sppModels[i] = new(engine.RouteMdl)
				if err = utils.UpdateStructWithIfaceMap(sppModels[i], ld); err != nil {
					return
				}
			}

			for _, tpSpp := range sppModels.AsTPRouteProfile() {
				spPrf, err := engine.APItoRouteProfile(tpSpp, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: RouteProfile: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(spPrf)))
					continue
				}
				// get IDs so we can reload in cache
				ids = append(ids, spPrf.TenantID())
				if err := ldr.dm.SetRouteProfile(spPrf, true); err != nil {
					return err
				}
				cacheArgs[utils.RouteProfileIDs] = ids
			}
		}
	case utils.MetaChargers:
		cacheIDs = []string{utils.CacheChargerFilterIndexes}
		for _, lDataSet := range lds {
			cppModels := make(engine.ChargerMdls, len(lDataSet))
			for i, ld := range lDataSet {
				cppModels[i] = new(engine.ChargerMdl)
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
				cacheArgs[utils.ChargerProfileIDs] = ids
			}
		}
	case utils.MetaDispatchers:
		cacheIDs = []string{utils.CacheDispatcherFilterIndexes}
		for _, lDataSet := range lds {
			dispModels := make(engine.DispatcherProfileMdls, len(lDataSet))
			for i, ld := range lDataSet {
				dispModels[i] = new(engine.DispatcherProfileMdl)
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
				cacheArgs[utils.DispatcherProfileIDs] = ids
			}
		}
	case utils.MetaDispatcherHosts:
		for _, lDataSet := range lds {
			dispModels := make(engine.DispatcherHostMdls, len(lDataSet))
			for i, ld := range lDataSet {
				dispModels[i] = new(engine.DispatcherHostMdl)
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
				cacheArgs[utils.DispatcherHostIDs] = ids
			}
		}
	case utils.MetaRateProfiles:
		cacheIDs = []string{utils.CacheRateProfilesFilterIndexes, utils.CacheRateFilterIndexes}
		for _, lDataSet := range lds {
			rpMdls := make(engine.RateProfileMdls, len(lDataSet))
			for i, ld := range lDataSet {
				rpMdls[i] = new(engine.RateProfileMdl)
				if err = utils.UpdateStructWithIfaceMap(rpMdls[i], ld); err != nil {
					return
				}
			}
			for _, tpRpl := range rpMdls.AsTPRateProfile() {
				rpl, err := engine.APItoRateProfile(tpRpl, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: RateProfile: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(rpl)))
					continue
				}
				// get IDs so we can reload in cache
				ids = append(ids, rpl.TenantID())
				if ldr.flagsTpls[loaderType].GetBool(utils.MetaPartial) {
					if err := ldr.dm.SetRateProfileRates(context.TODO(), rpl, true); err != nil {
						return err
					}
				} else {
					if err := ldr.dm.SetRateProfile(context.TODO(), rpl, true); err != nil {
						return err
					}
				}
				cacheArgs[utils.RateProfileIDs] = ids
			}
		}
	case utils.MetaActionProfiles:
		cacheIDs = []string{utils.CacheActionProfilesFilterIndexes}
		for _, lDataSet := range lds {
			acpsModels := make(engine.ActionProfileMdls, len(lDataSet))
			for i, ld := range lDataSet {
				acpsModels[i] = new(engine.ActionProfileMdl)
				if err = utils.UpdateStructWithIfaceMap(acpsModels[i], ld); err != nil {
					return
				}
			}

			for _, tpAcp := range acpsModels.AsTPActionProfile() {
				acp, err := engine.APItoActionProfile(tpAcp, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: ActionProfile: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(acp)))
					continue
				}
				// get IDs so we can reload in cache
				ids = append(ids, acp.TenantID())
				if err := ldr.dm.SetActionProfile(acp, true); err != nil {
					return err
				}
				cacheArgs[utils.ActionProfileIDs] = ids
			}
		}
	case utils.MetaAccounts:
		cacheIDs = []string{utils.CacheAccountsFilterIndexes}
		for _, lDataSet := range lds {
			acpsModels := make(engine.AccountMdls, len(lDataSet))
			for i, ld := range lDataSet {
				acpsModels[i] = new(engine.AccountMdl)
				if err = utils.UpdateStructWithIfaceMap(acpsModels[i], ld); err != nil {
					return
				}
			}
			accountTPModels, err := acpsModels.AsTPAccount()
			if err != nil {
				return err
			}
			for _, tpAcp := range accountTPModels {
				acp, err := engine.APItoAccount(tpAcp, ldr.timezone)
				if err != nil {
					return err
				}
				if ldr.dryRun {
					utils.Logger.Info(
						fmt.Sprintf("<%s-%s> DRY_RUN: Accounts: %s",
							utils.LoaderS, ldr.ldrID, utils.ToJSON(acp)))
					continue
				}
				// get IDs so we can reload in cache
				ids = append(ids, acp.TenantID())
				if err := ldr.dm.SetAccount(acp, true); err != nil {
					return err
				}
			}
		}
	}

	if len(ldr.cacheConns) != 0 {
		return engine.CallCache(ldr.connMgr, context.TODO(), ldr.cacheConns, caching, cacheArgs, cacheIDs, nil, false)
	}
	return
}

//removeContent will process the content and will remove it from database
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
			if err = ldr.removeLoadedData(loaderType,
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
	if err = ldr.removeLoadedData(loaderType,
		map[string][]LoaderData{tntID: ldr.bufLoaderData[tntID]}, caching); err != nil {
		return
	}
	delete(ldr.bufLoaderData, tntID)
	return
}

//removeLoadedData will remove the data from database
//since we remove we don't need to compose the struct we only need the Tenant and the ID of the profile
func (ldr *Loader) removeLoadedData(loaderType string, lds map[string][]LoaderData, caching string) (err error) {
	var ids []string
	cacheArgs := make(map[string][]string)
	var cacheIDs []string // verify if we need to clear indexe
	switch loaderType {
	case utils.MetaAttributes:
		cacheIDs = []string{utils.CacheAttributeFilterIndexes}
		for tntID := range lds {
			if ldr.dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: AttributeProfileID: %s",
						utils.LoaderS, ldr.ldrID, tntID))
			} else {
				tntIDStruct := utils.NewTenantID(tntID)
				// get IDs so we can reload in cache
				ids = append(ids, tntID)
				if err := ldr.dm.RemoveAttributeProfile(context.TODO(), tntIDStruct.Tenant, tntIDStruct.ID,
					utils.NonTransactional, true); err != nil {
					return err
				}
				cacheArgs[utils.AttributeProfileIDs] = ids
			}
		}

	case utils.MetaResources:
		cacheIDs = []string{utils.CacheResourceFilterIndexes}
		for tntID := range lds {
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
				cacheArgs[utils.ResourceProfileIDs] = ids
				cacheArgs[utils.ResourceIDs] = ids
			}
		}
	case utils.MetaFilters:
		for tntID := range lds {
			if ldr.dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: Filter: %s",
						utils.LoaderS, ldr.ldrID, tntID))
			} else {
				tntIDStruct := utils.NewTenantID(tntID)
				// get IDs so we can reload in cache
				ids = append(ids, tntID)
				if err := ldr.dm.RemoveFilter(tntIDStruct.Tenant, tntIDStruct.ID,
					utils.NonTransactional, true); err != nil {
					return err
				}
				cacheArgs[utils.FilterIDs] = ids
			}
		}
	case utils.MetaStats:
		cacheIDs = []string{utils.CacheStatFilterIndexes}
		for tntID := range lds {
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
				cacheArgs[utils.StatsQueueProfileIDs] = ids
				cacheArgs[utils.StatsQueueIDs] = ids
			}
		}
	case utils.MetaThresholds:
		cacheIDs = []string{utils.CacheThresholdFilterIndexes}
		for tntID := range lds {
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
				cacheArgs[utils.ThresholdProfileIDs] = ids
				cacheArgs[utils.ThresholdIDs] = ids
			}
		}
	case utils.MetaRoutes:
		cacheIDs = []string{utils.CacheRouteFilterIndexes}
		for tntID := range lds {
			if ldr.dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: RouteProfileID: %s",
						utils.LoaderS, ldr.ldrID, tntID))
			} else {
				tntIDStruct := utils.NewTenantID(tntID)
				// get IDs so we can reload in cache
				ids = append(ids, tntID)
				if err := ldr.dm.RemoveRouteProfile(tntIDStruct.Tenant,
					tntIDStruct.ID, utils.NonTransactional, true); err != nil {
					return err
				}
				cacheArgs[utils.RouteProfileIDs] = ids
			}
		}
	case utils.MetaChargers:
		cacheIDs = []string{utils.CacheChargerFilterIndexes}
		for tntID := range lds {
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
				cacheArgs[utils.ChargerProfileIDs] = ids
			}
		}
	case utils.MetaDispatchers:
		cacheIDs = []string{utils.CacheDispatcherFilterIndexes}
		for tntID := range lds {
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
				cacheArgs[utils.DispatcherProfileIDs] = ids
			}
		}
	case utils.MetaDispatcherHosts:
		for tntID := range lds {
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
				cacheArgs[utils.DispatcherHostIDs] = ids
			}
		}
	case utils.MetaRateProfiles:
		cacheIDs = []string{utils.CacheRateProfilesFilterIndexes, utils.CacheRateFilterIndexes}
		for tntID, ldData := range lds {
			if ldr.dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: RateProfileIDs: %s",
						utils.LoaderS, ldr.ldrID, tntID))
			} else {
				tntIDStruct := utils.NewTenantID(tntID)
				// get IDs so we can reload in cache
				ids = append(ids, tntID)
				if ldr.flagsTpls[loaderType].GetBool(utils.MetaPartial) {
					rateIDs, err := ldData[0].GetRateIDs()
					if err != nil {
						return err
					}
					if err := ldr.dm.RemoveRateProfileRates(context.TODO(), tntIDStruct.Tenant,
						tntIDStruct.ID, rateIDs, true); err != nil {
						return err
					}
				} else {
					if err := ldr.dm.RemoveRateProfile(context.TODO(), tntIDStruct.Tenant,
						tntIDStruct.ID, utils.NonTransactional, true); err != nil {
						return err
					}
				}

				cacheArgs[utils.RateProfileIDs] = ids
			}
		}
	case utils.MetaActionProfiles:
		cacheIDs = []string{utils.CacheActionProfiles, utils.CacheActionProfilesFilterIndexes}
		for tntID := range lds {
			if ldr.dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: ActionProfileID: %s",
						utils.LoaderS, ldr.ldrID, tntID))
			} else {
				tntIDStruct := utils.NewTenantID(tntID)
				// get IDs so we can reload in cache
				ids = append(ids, tntID)
				if err := ldr.dm.RemoveActionProfile(tntIDStruct.Tenant,
					tntIDStruct.ID, utils.NonTransactional, true); err != nil {
					return err
				}
				cacheArgs[utils.ActionProfileIDs] = ids
			}
		}
	case utils.MetaAccounts:
		cacheIDs = []string{utils.CacheAccounts, utils.CacheAccountsFilterIndexes}
		for tntID := range lds {
			if ldr.dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: AccountIDs: %s",
						utils.LoaderS, ldr.ldrID, tntID))
			} else {
				tntIDStruct := utils.NewTenantID(tntID)
				// get IDs so we can reload in cache
				ids = append(ids, tntID)
				if err := ldr.dm.RemoveAccount(tntIDStruct.Tenant,
					tntIDStruct.ID, utils.NonTransactional, true); err != nil {
					return err
				}
			}
		}
	}

	if len(ldr.cacheConns) != 0 {
		return engine.CallCache(ldr.connMgr, context.TODO(), ldr.cacheConns, caching, cacheArgs, cacheIDs, nil, false)
	}
	return
}

func (ldr *Loader) serve(stopChan chan struct{}) (err error) {
	switch ldr.runDelay {
	case time.Duration(0): // 0 disables the automatic read, maybe done per API
		return
	case time.Duration(-1):
		return utils.WatchDir(ldr.tpInDir, ldr.processFile,
			utils.LoaderS+"-"+ldr.ldrID, stopChan)
	default:
		go ldr.handleFolder(stopChan)
	}
	return
}

func (ldr *Loader) handleFolder(stopChan chan struct{}) {
	for {
		go ldr.ProcessFolder(config.CgrConfig().GeneralCfg().DefaultCaching, utils.MetaStore, false)
		timer := time.NewTimer(ldr.runDelay)
		select {
		case <-stopChan:
			utils.Logger.Info(
				fmt.Sprintf("<%s-%s> stop monitoring path <%s>",
					utils.LoaderS, ldr.ldrID, ldr.tpInDir))
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (ldr *Loader) processFile(_, itmID string) (err error) {
	loaderType := ldr.getLdrType(itmID)
	if len(loaderType) == 0 {
		return
	}
	if err = ldr.lockFolder(); err != nil {
		return
	}
	defer ldr.unlockFolder()
	if ldr.rdrs[loaderType][itmID] != nil {
		ldr.unreferenceFile(loaderType, itmID)
	}
	var rdr *os.File
	if rdr, err = os.Open(path.Join(ldr.tpInDir, itmID)); err != nil {
		return
	}
	csvReader := csv.NewReader(rdr)
	csvReader.Comma = rune(ldr.fieldSep[0])
	csvReader.Comment = '#'
	ldr.rdrs[loaderType][itmID] = &openedCSVFile{
		fileName: itmID, rdr: rdr, csvRdr: csvReader}
	if !ldr.allFilesPresent(loaderType) {
		return
	}
	for fName := range ldr.rdrs[loaderType] {
		defer ldr.unreferenceFile(loaderType, fName)
	}

	err = ldr.processContent(loaderType, config.CgrConfig().GeneralCfg().DefaultCaching)

	if ldr.tpOutDir == utils.EmptyString {
		return
	}
	for fName := range ldr.rdrs[loaderType] {
		oldPath := path.Join(ldr.tpInDir, fName)
		newPath := path.Join(ldr.tpOutDir, fName)
		if nerr := os.Rename(oldPath, newPath); nerr != nil {
			return nerr
		}
	}
	return
}

func (ldr *Loader) allFilesPresent(ldrType string) bool {
	for _, rdr := range ldr.rdrs[ldrType] {
		if rdr == nil {
			return false
		}
	}
	return true
}

// getLdrType returns loaderType for the given fileName
func (ldr *Loader) getLdrType(fName string) (ldrType string) {
	for ldr, rdrs := range ldr.rdrs {
		if _, has := rdrs[fName]; has {
			return ldr
		}
	}
	return
}
