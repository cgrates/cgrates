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
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// const (
// 	gprefix = utils.MetaGoogleAPI + utils.ConcatenatedKeySep
// )

func removeFromDB(ctx *context.Context, dm *engine.DataManager, lType string, withIndex, ratesPartial bool, obj profile) (_ error) {
	tntID := utils.NewTenantID(obj.TenantID())
	tnt, id := tntID.Tenant, tntID.ID
	switch lType {
	case utils.MetaAttributes:
		return dm.RemoveAttributeProfile(ctx, tnt, id, withIndex)
	case utils.MetaResources:
		return dm.RemoveResourceProfile(ctx, tnt, id, withIndex)
	case utils.MetaFilters:
		return dm.RemoveFilter(ctx, tnt, id, withIndex)
	case utils.MetaStats:
		return dm.RemoveStatQueueProfile(ctx, tnt, id, withIndex)
	case utils.MetaThresholds:
		return dm.RemoveThresholdProfile(ctx, tnt, id, withIndex)
	case utils.MetaRoutes:
		return dm.RemoveRouteProfile(ctx, tnt, id, withIndex)
	case utils.MetaChargers:
		return dm.RemoveChargerProfile(ctx, tnt, id, withIndex)
	case utils.MetaDispatchers:
		return dm.RemoveDispatcherProfile(ctx, tnt, id, withIndex)
	case utils.MetaDispatcherHosts:
		return dm.RemoveDispatcherHost(ctx, tnt, id)
	case utils.MetaRateProfiles:
		if ratesPartial {
			rt := obj.(*utils.RateProfile)
			ids := make([]string, 0, len(rt.Rates))
			for k := range rt.Rates {
				ids = append(ids, k)
			}
			return dm.RemoveRateProfileRates(ctx, tnt, id, &ids, withIndex)
		}
		return dm.RemoveRateProfile(ctx, tnt, id, withIndex)
	case utils.MetaActionProfiles:
		return dm.RemoveActionProfile(ctx, tnt, id, withIndex)
	case utils.MetaAccounts:
		return dm.RemoveAccount(ctx, tnt, id, withIndex)
	}
	return
}

func setToDB(ctx *context.Context, dm *engine.DataManager, lType string, data profile, withIndex, ratesPartial bool) (err error) {
	switch lType {
	case utils.MetaAttributes:
		return dm.SetAttributeProfile(ctx, data.(*engine.AttributeProfile), withIndex)
	case utils.MetaResources:
		return dm.SetResourceProfile(ctx, data.(*engine.ResourceProfile), withIndex)
	case utils.MetaFilters:
		fltr := data.(*engine.Filter)
		fltr.Compress()
		if err = fltr.Compile(); err != nil {
			return
		}
		return dm.SetFilter(ctx, fltr, withIndex)
	case utils.MetaStats:
		return dm.SetStatQueueProfile(ctx, data.(*engine.StatQueueProfile), withIndex)
	case utils.MetaThresholds:
		return dm.SetThresholdProfile(ctx, data.(*engine.ThresholdProfile), withIndex)
	case utils.MetaRoutes:
		return dm.SetRouteProfile(ctx, data.(*engine.RouteProfile), withIndex)
	case utils.MetaChargers:
		return dm.SetChargerProfile(ctx, data.(*engine.ChargerProfile), withIndex)
	case utils.MetaDispatchers:
		return dm.SetDispatcherProfile(ctx, data.(*engine.DispatcherProfile), withIndex)
	case utils.MetaDispatcherHosts:
		return dm.SetDispatcherHost(ctx, data.(*engine.DispatcherHost))
	case utils.MetaRateProfiles:
		rpl := data.(*utils.RateProfile)
		if ratesPartial {
			err = dm.SetRateProfile(ctx, rpl, false, true)
		} else {
			err = dm.SetRateProfile(ctx, rpl, true, true)
		}
	case utils.MetaActionProfiles:
		return dm.SetActionProfile(ctx, data.(*engine.ActionProfile), withIndex)
	case utils.MetaAccounts:
		return dm.SetAccount(ctx, data.(*utils.Account), withIndex)
	}
	return
}

func dryRun(ctx *context.Context, lType, ldrID string, obj profile) (err error) {
	var msg string
	switch lType {
	case utils.MetaAttributes:
		msg = "<%s-%s> DRY_RUN: AttributeProfile: %s"
	case utils.MetaResources:
		msg = "<%s-%s> DRY_RUN: ResourceProfile: %s"
	case utils.MetaFilters:
		fltr := obj.(*engine.Filter)
		fltr.Compress()
		if err = fltr.Compile(); err != nil {
			return
		}
		msg = "<%s-%s> DRY_RUN: Filter: %s"
	case utils.MetaStats:
		msg = "<%s-%s> DRY_RUN: StatsQueueProfile: %s"
	case utils.MetaThresholds:
		msg = "<%s-%s> DRY_RUN: ThresholdProfile: %s"
	case utils.MetaRoutes:
		msg = "<%s-%s> DRY_RUN: RouteProfile: %s"
	case utils.MetaChargers:
		msg = "<%s-%s> DRY_RUN: ChargerProfile: %s"
	case utils.MetaDispatchers:
		msg = "<%s-%s> DRY_RUN: DispatcherProfile: %s"
	case utils.MetaDispatcherHosts:
		msg = "<%s-%s> DRY_RUN: DispatcherHost: %s"
	case utils.MetaRateProfiles:
		msg = "<%s-%s> DRY_RUN: RateProfile: %s"
	case utils.MetaActionProfiles:
		msg = "<%s-%s> DRY_RUN: ActionProfile: %s"
	case utils.MetaAccounts:
		msg = "<%s-%s> DRY_RUN: Accounts: %s"
	}
	utils.Logger.Info(fmt.Sprintf(msg,
		utils.LoaderS, ldrID, utils.ToJSON(obj)))
	return
}

func newLoader(cfg *config.CGRConfig, ldrCfg *config.LoaderSCfg, dm *engine.DataManager, dataCache map[string]*ltcache.Cache,
	filterS *engine.FilterS, connMgr *engine.ConnManager, cacheConns []string) *loader {
	return &loader{
		cfg:        cfg,
		ldrCfg:     ldrCfg,
		dm:         dm,
		filterS:    filterS,
		connMgr:    connMgr,
		cacheConns: cacheConns,
		dataCache:  dataCache,
		Locker:     newLocker(ldrCfg.GetLockFilePath(), ldrCfg.ID),
	}
}

type loader struct {
	cfg        *config.CGRConfig
	ldrCfg     *config.LoaderSCfg
	dm         *engine.DataManager
	filterS    *engine.FilterS
	connMgr    *engine.ConnManager
	cacheConns []string

	dataCache map[string]*ltcache.Cache
	Locker
}

func (l *loader) process(ctx *context.Context, obj profile, lType, action string, opts map[string]any, withIndex, partialRates bool) (err error) {
	switch action {
	case utils.MetaParse:
		return
	case utils.MetaDryRun:
		return dryRun(ctx, lType, l.ldrCfg.ID, obj)
	case utils.MetaStore:
		err = setToDB(ctx, l.dm, lType, obj, withIndex, partialRates)
	case utils.MetaRemove:
		err = removeFromDB(ctx, l.dm, lType, withIndex, partialRates, obj)
	default:
		return fmt.Errorf("unsupported loader action: <%q>", action)
	}
	if err != nil || len(l.cacheConns) == 0 {
		return
	}
	cacheArgs := make(map[string][]string)
	var cacheIDs []string // verify if we need to clear indexe
	tntId := obj.TenantID()
	switch lType {
	case utils.MetaAttributes:
		cacheIDs = []string{utils.CacheAttributeFilterIndexes}
		cacheArgs[utils.CacheAttributeProfiles] = []string{tntId}
	case utils.MetaResources:
		cacheIDs = []string{utils.CacheResourceFilterIndexes}
		cacheArgs[utils.CacheResourceProfiles] = []string{tntId}
		cacheArgs[utils.CacheResources] = []string{tntId}
	case utils.MetaFilters:
		cacheArgs[utils.CacheFilters] = []string{tntId}
	case utils.MetaStats:
		cacheIDs = []string{utils.CacheStatFilterIndexes}
		cacheArgs[utils.CacheStatQueueProfiles] = []string{tntId}
		cacheArgs[utils.CacheStatQueues] = []string{tntId}
	case utils.MetaThresholds:
		cacheIDs = []string{utils.CacheThresholdFilterIndexes}
		cacheArgs[utils.CacheThresholdProfiles] = []string{tntId}
		cacheArgs[utils.CacheThresholds] = []string{tntId}
	case utils.MetaRoutes:
		cacheIDs = []string{utils.CacheRouteFilterIndexes}
		cacheArgs[utils.CacheRouteProfiles] = []string{tntId}
	case utils.MetaChargers:
		cacheIDs = []string{utils.CacheChargerFilterIndexes}
		cacheArgs[utils.CacheChargerProfiles] = []string{tntId}
	case utils.MetaDispatchers:
		cacheIDs = []string{utils.CacheDispatcherFilterIndexes}
		cacheArgs[utils.CacheDispatcherProfiles] = []string{tntId}
	case utils.MetaDispatcherHosts:
		cacheArgs[utils.CacheDispatcherHosts] = []string{tntId}
	case utils.MetaRateProfiles:
		cacheIDs = []string{utils.CacheRateProfilesFilterIndexes, utils.CacheRateFilterIndexes}
		cacheArgs[utils.CacheRateProfiles] = []string{tntId}
	case utils.MetaActionProfiles:
		cacheIDs = []string{utils.CacheActionProfiles, utils.CacheActionProfilesFilterIndexes}
		cacheArgs[utils.CacheActionProfiles] = []string{tntId}
	case utils.MetaAccounts:
		cacheIDs = []string{utils.CacheAccounts, utils.CacheAccountsFilterIndexes}
	}

	// delay if needed before cache reload
	if l.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<%v> Delaying cache reload for %v", utils.LoaderS, l.cfg.GeneralCfg().CachingDelay))
		time.Sleep(l.cfg.GeneralCfg().CachingDelay)
	}

	return engine.CallCache(l.connMgr, ctx, l.cacheConns,
		utils.FirstNonEmpty(utils.IfaceAsString(opts[utils.MetaCache]), l.ldrCfg.Opts.Cache, l.cfg.GeneralCfg().DefaultCaching),
		cacheArgs, cacheIDs, opts, false, l.ldrCfg.Tenant)
}

func (l *loader) processData(ctx *context.Context, csv *CSVFile, tmpls []*config.FCTemplate, lType, action string, opts map[string]any, withIndex, partialRates bool) (err error) {
	newPrf := newProfileFunc(lType)
	obj := newPrf()
	var prevTntID string
	for lineNr := 1; ; lineNr++ {
		var record []string
		if record, err = csv.Read(); err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			utils.Logger.Warning(
				fmt.Sprintf("<%s> <%s> reading file<%s> on line: %d, error: %s",
					utils.LoaderS, l.ldrCfg.ID, csv.Path(), lineNr, err))
			return
		}

		tmp := newPrf()
		if err = newRecord(config.NewSliceDP(record, nil), tmp, l.ldrCfg.Tenant, l.cfg, l.dataCache[lType]).
			SetFields(ctx, tmpls, l.filterS, l.cfg.GeneralCfg().RoundingDecimals, l.cfg.GeneralCfg().DefaultTimezone, l.cfg.GeneralCfg().RSRSep); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> <%s> file<%s> line: %d, error: %s",
					utils.LoaderS, l.ldrCfg.ID, csv.Path(), lineNr, err))
			return
		}

		if tntID := tmp.TenantID(); prevTntID != tntID {
			if len(prevTntID) != 0 {
				if err = l.process(ctx, obj, lType, action, opts, withIndex, partialRates); err != nil {
					return
				}
			}
			prevTntID = tntID
			obj = newPrf()
		}
		obj.Merge(tmp)
	}
	if len(prevTntID) != 0 {
		err = l.process(ctx, obj, lType, action, opts, withIndex, partialRates)
	}
	return
}

func (l *loader) processFile(ctx *context.Context, cfg *config.LoaderDataType, inPath, outPath, action string, opts map[string]any, withIndex bool, prv CSVProvider) (err error) {
	var csv *CSVFile
	if csv, err = NewCSVReader(prv, inPath, cfg.Filename, rune(l.ldrCfg.FieldSeparator[0]), 0); err != nil {
		return
	}
	defer csv.Close()
	if err = l.processData(ctx, csv, cfg.Fields, cfg.Type, action, opts,
		withIndex, cfg.Flags.GetBool(utils.PartialRatesOpt)); err != nil || // encounterd error
		outPath == utils.EmptyString || // or no moving
		prv.Type() != utils.MetaFileCSV { // or the type can not be moved(e.g. url)
		return
	}
	return os.Rename(path.Join(inPath, cfg.Filename), path.Join(outPath, cfg.Filename))
}

func (l *loader) getCfg(fileName string) (cfg *config.LoaderDataType) {
	for _, cfg = range l.ldrCfg.Data {
		if cfg.Filename == fileName {
			return
		}
	}
	return nil
}

func (l *loader) processIFile(fileName string) (err error) {
	cfg := l.getCfg(fileName)
	if cfg == nil {
		if pathIn := path.Join(l.ldrCfg.TpInDir, fileName); !l.IsLockFile(pathIn) && len(l.ldrCfg.TpOutDir) != 0 {
			err = os.Rename(pathIn, path.Join(l.ldrCfg.TpOutDir, fileName))
		}
		return
	}

	if err = l.Lock(); err != nil {
		return
	}
	defer l.Unlock()
	return l.processFile(context.Background(), cfg, l.ldrCfg.TpInDir, l.ldrCfg.TpOutDir, l.ldrCfg.Action, nil, l.ldrCfg.Opts.WithIndex, fileProvider{})
}

func (l *loader) processFolder(ctx *context.Context, opts map[string]any, withIndex, stopOnError bool) (err error) {
	if err = l.Lock(); err != nil {
		return
	}
	defer l.Unlock()
	var csvType CSVProvider = fileProvider{}
	switch {
	// case strings.HasPrefix(inPath, gprefix): // uncomment this after *gapi is implemented
	// 	csvType = utils.MetaGoogleAPI
	// 	inPath = strings.TrimPrefix(inPath, gprefix)
	case utils.IsURL(l.ldrCfg.TpInDir):
		csvType = urlProvider{}
	}
	for _, cfg := range l.ldrCfg.Data {
		if err = l.processFile(ctx, cfg, l.ldrCfg.TpInDir, l.ldrCfg.TpOutDir, l.ldrCfg.Action, opts, withIndex, csvType); err != nil {
			if !stopOnError {
				utils.Logger.Warning(fmt.Sprintf("<%s-%s> loaderType: <%s> cannot open files, err: %s",
					utils.LoaderS, l.ldrCfg.ID, cfg.Type, err))
				err = nil
				continue
			}
			return
		}
	}
	if len(l.ldrCfg.TpOutDir) != 0 {
		err = l.moveUnprocessedFiles()
	}
	return
}

func (l *loader) moveUnprocessedFiles() (err error) {
	var fs []os.DirEntry
	if fs, err = os.ReadDir(l.ldrCfg.TpInDir); err != nil {
		return
	}
	for _, f := range fs {
		if pathIn := path.Join(l.ldrCfg.TpInDir, f.Name()); !l.IsLockFile(pathIn) {
			if err = os.Rename(pathIn, path.Join(l.ldrCfg.TpOutDir, f.Name())); err != nil {
				return
			}
		}
	}
	return
}

func (l *loader) handleFolder(stopChan chan struct{}) {
	for {
		go l.processFolder(context.Background(), nil, l.ldrCfg.Opts.WithIndex, false)
		timer := time.NewTimer(l.ldrCfg.RunDelay)
		select {
		case <-stopChan:
			utils.Logger.Info(
				fmt.Sprintf("<%s-%s> stop monitoring path <%s>",
					utils.LoaderS, l.ldrCfg.ID, l.ldrCfg.TpInDir))
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (l *loader) ListenAndServe(stopChan chan struct{}) (err error) {
	utils.Logger.Info(fmt.Sprintf("Starting <%s-%s>", utils.LoaderS, l.ldrCfg.ID))
	switch l.ldrCfg.RunDelay {
	case 0: // 0 disables the automatic read, maybe done per API
	case -1:
		return utils.WatchDir(l.ldrCfg.TpInDir, l.processIFile,
			utils.LoaderS+"-"+l.ldrCfg.ID, stopChan)
	default:
		go l.handleFolder(stopChan)
	}
	return
}

func (l *loader) processZip(ctx *context.Context, opts map[string]any, withIndex, stopOnError bool, zipR *zip.Reader) (err error) {
	if err = l.Lock(); err != nil {
		return
	}
	defer l.Unlock()
	ziP := zipProvider{zipR}
	for _, cfg := range l.ldrCfg.Data {
		if err = l.processFile(ctx, cfg, l.ldrCfg.TpInDir, l.ldrCfg.TpOutDir, l.ldrCfg.Action, opts, withIndex, ziP); err != nil {
			if !stopOnError {
				utils.Logger.Warning(fmt.Sprintf("<%s-%s> loaderType: <%s> cannot open files, err: %s",
					utils.LoaderS, l.ldrCfg.ID, cfg.Type, err))
				err = nil
				continue
			}
			return
		}
	}
	return
}
