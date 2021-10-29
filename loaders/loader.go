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
	"github.com/cgrates/ltcache"
)

const (
	gprefix = utils.MetaGoogleAPI + utils.ConcatenatedKeySep
)

func removeFromDB(ctx *context.Context, dm *engine.DataManager, lType, tnt, id, ldrID string, dryRun, withIndex, ratesPartial bool, ratesData utils.MapStorage) (_ error) {
	if dryRun {
		var logID string
		switch lType {
		case utils.MetaAttributes:
			logID = "AttributeProfile"
		case utils.MetaResources:
			logID = "ResourceProfile"
		case utils.MetaFilters:
			logID = "Filter"
		case utils.MetaStats:
			logID = "StatsQueueProfile"
		case utils.MetaThresholds:
			logID = "ThresholdProfile"
		case utils.MetaRoutes:
			logID = "RouteProfile"
		case utils.MetaChargers:
			logID = "ChargerProfile"
		case utils.MetaDispatchers:
			logID = "DispatcherProfile"
		case utils.MetaDispatcherHosts:
			logID = "DispatcherHost"
		case utils.MetaRateProfiles:
			logID = "RateProfile"
		case utils.MetaActionProfiles:
			logID = "ActionProfil"
		case utils.MetaAccounts:
			logID = "Account"
		}
		utils.Logger.Info(
			fmt.Sprintf("<%s-%s> DRY_RUN: %sID: %s",
				utils.LoaderS, ldrID, logID, utils.ConcatenatedKey(tnt, id)))
		return
	}
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
			rateIDs, err := RateIDsFromMap(ratesData)
			if err != nil {
				return err
			}
			return dm.RemoveRateProfileRates(ctx, tnt, id, rateIDs, withIndex)
		}
		return dm.RemoveRateProfile(ctx, tnt, id, withIndex)
	case utils.MetaActionProfiles:
		return dm.RemoveActionProfile(ctx, tnt, id, withIndex)
	case utils.MetaAccounts:
		return dm.RemoveAccount(ctx, tnt, id, withIndex)
	}
	return
}

func setToDB(ctx *context.Context, dm *engine.DataManager, lType, tmz, ldrID string, lDataSet []utils.MapStorage, dryRun, withIndex, ratesPartial bool) (err error) {
	switch lType {
	case utils.MetaAttributes:
		attrModels := make(engine.AttributeMdls, len(lDataSet))
		for i, ld := range lDataSet {
			attrModels[i] = new(engine.AttributeMdl)
			if err = utils.UpdateStructWithIfaceMap(attrModels[i], ld); err != nil {
				return
			}
		}
		for _, tpApf := range attrModels.AsTPAttributes() {
			var apf *engine.AttributeProfile
			if apf, err = engine.APItoAttributeProfile(tpApf, tmz); err != nil {
				return
			}
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: AttributeProfile: %s",
						utils.LoaderS, ldrID, utils.ToJSON(apf)))
				continue
			}
			if err = dm.SetAttributeProfile(ctx, apf, withIndex); err != nil {
				return
			}
		}
	case utils.MetaResources:
		resModels := make(engine.ResourceMdls, len(lDataSet))
		for i, ld := range lDataSet {
			resModels[i] = new(engine.ResourceMdl)
			if err = utils.UpdateStructWithIfaceMap(resModels[i], ld); err != nil {
				return
			}
		}
		for _, tpRes := range resModels.AsTPResources() {
			var res *engine.ResourceProfile
			if res, err = engine.APItoResource(tpRes, tmz); err != nil {
				return
			}
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: ResourceProfile: %s",
						utils.LoaderS, ldrID, utils.ToJSON(res)))
				continue
			}
			if err = dm.SetResourceProfile(ctx, res, withIndex); err != nil {
				return
			}
		}
	case utils.MetaFilters:
		fltrModels := make(engine.FilterMdls, len(lDataSet))
		for i, ld := range lDataSet {
			fltrModels[i] = new(engine.FilterMdl)
			if err = utils.UpdateStructWithIfaceMap(fltrModels[i], ld); err != nil {
				return
			}
		}

		for _, tpFltr := range fltrModels.AsTPFilter() {
			var fltrPrf *engine.Filter
			if fltrPrf, err = engine.APItoFilter(tpFltr, tmz); err != nil {
				return
			}
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: Filter: %s",
						utils.LoaderS, ldrID, utils.ToJSON(fltrPrf)))
				continue
			}
			if err = dm.SetFilter(ctx, fltrPrf, withIndex); err != nil {
				return
			}
		}
	case utils.MetaStats:
		stsModels := make(engine.StatMdls, len(lDataSet))
		for i, ld := range lDataSet {
			stsModels[i] = new(engine.StatMdl)
			if err = utils.UpdateStructWithIfaceMap(stsModels[i], ld); err != nil {
				return
			}
		}
		for _, tpSts := range stsModels.AsTPStats() {
			var stsPrf *engine.StatQueueProfile
			if stsPrf, err = engine.APItoStats(tpSts, tmz); err != nil {
				return
			}
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: StatsQueueProfile: %s",
						utils.LoaderS, ldrID, utils.ToJSON(stsPrf)))
				continue
			}
			if err = dm.SetStatQueueProfile(ctx, stsPrf, withIndex); err != nil {
				return
			}
		}
	case utils.MetaThresholds:
		thModels := make(engine.ThresholdMdls, len(lDataSet))
		for i, ld := range lDataSet {
			thModels[i] = new(engine.ThresholdMdl)
			if err = utils.UpdateStructWithIfaceMap(thModels[i], ld); err != nil {
				return
			}
		}
		for _, tpTh := range thModels.AsTPThreshold() {
			var thPrf *engine.ThresholdProfile
			if thPrf, err = engine.APItoThresholdProfile(tpTh, tmz); err != nil {
				return
			}
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: ThresholdProfile: %s",
						utils.LoaderS, ldrID, utils.ToJSON(thPrf)))
				continue
			}
			if err = dm.SetThresholdProfile(ctx, thPrf, withIndex); err != nil {
				return
			}
		}
	case utils.MetaRoutes:
		sppModels := make(engine.RouteMdls, len(lDataSet))
		for i, ld := range lDataSet {
			sppModels[i] = new(engine.RouteMdl)
			if err = utils.UpdateStructWithIfaceMap(sppModels[i], ld); err != nil {
				return
			}
		}

		for _, tpSpp := range sppModels.AsTPRouteProfile() {
			var spPrf *engine.RouteProfile
			if spPrf, err = engine.APItoRouteProfile(tpSpp, tmz); err != nil {
				return
			}
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: RouteProfile: %s",
						utils.LoaderS, ldrID, utils.ToJSON(spPrf)))
				continue
			}
			if err = dm.SetRouteProfile(ctx, spPrf, withIndex); err != nil {
				return
			}
		}
	case utils.MetaChargers:
		cppModels := make(engine.ChargerMdls, len(lDataSet))
		for i, ld := range lDataSet {
			cppModels[i] = new(engine.ChargerMdl)
			if err = utils.UpdateStructWithIfaceMap(cppModels[i], ld); err != nil {
				return
			}
		}

		for _, tpCPP := range cppModels.AsTPChargers() {
			var cpp *engine.ChargerProfile
			if cpp, err = engine.APItoChargerProfile(tpCPP, tmz); err != nil {
				return
			}
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: ChargerProfile: %s",
						utils.LoaderS, ldrID, utils.ToJSON(cpp)))
				continue
			}
			if err = dm.SetChargerProfile(ctx, cpp, withIndex); err != nil {
				return
			}
		}
	case utils.MetaDispatchers:
		dispModels := make(engine.DispatcherProfileMdls, len(lDataSet))
		for i, ld := range lDataSet {
			dispModels[i] = new(engine.DispatcherProfileMdl)
			if err = utils.UpdateStructWithIfaceMap(dispModels[i], ld); err != nil {
				return
			}
		}
		for _, tpDsp := range dispModels.AsTPDispatcherProfiles() {
			var dsp *engine.DispatcherProfile
			if dsp, err = engine.APItoDispatcherProfile(tpDsp, tmz); err != nil {
				return
			}
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: DispatcherProfile: %s",
						utils.LoaderS, ldrID, utils.ToJSON(dsp)))
				continue
			}
			if err = dm.SetDispatcherProfile(ctx, dsp, withIndex); err != nil {
				return
			}
		}
	case utils.MetaDispatcherHosts:
		dispModels := make(engine.DispatcherHostMdls, len(lDataSet))
		for i, ld := range lDataSet {
			dispModels[i] = new(engine.DispatcherHostMdl)
			if err = utils.UpdateStructWithIfaceMap(dispModels[i], ld); err != nil {
				return
			}
		}
		var tpDsps []*utils.TPDispatcherHost
		if tpDsps, err = dispModels.AsTPDispatcherHosts(); err != nil {
			return
		}
		for _, tpDsp := range tpDsps {
			dsp := engine.APItoDispatcherHost(tpDsp)
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: DispatcherHost: %s",
						utils.LoaderS, ldrID, utils.ToJSON(dsp)))
				continue
			}
			if err = dm.SetDispatcherHost(ctx, dsp); err != nil {
				return
			}
		}
	case utils.MetaRateProfiles:
		rpMdls := make(engine.RateProfileMdls, len(lDataSet))
		for i, ld := range lDataSet {
			rpMdls[i] = new(engine.RateProfileMdl)
			if err = utils.UpdateStructWithIfaceMap(rpMdls[i], ld); err != nil {
				return
			}
		}
		for _, tpRpl := range rpMdls.AsTPRateProfile() {
			var rpl *utils.RateProfile
			if rpl, err = engine.APItoRateProfile(tpRpl, tmz); err != nil {
				return
			}
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: RateProfile: %s",
						utils.LoaderS, ldrID, utils.ToJSON(rpl)))
				continue
			}
			if ratesPartial {
				err = dm.SetRateProfileRates(ctx, rpl, true)
			} else {
				err = dm.SetRateProfile(ctx, rpl, true)
			}
			if err != nil {
				return
			}
		}
	case utils.MetaActionProfiles:
		acpsModels := make(engine.ActionProfileMdls, len(lDataSet))
		for i, ld := range lDataSet {
			acpsModels[i] = new(engine.ActionProfileMdl)
			if err = utils.UpdateStructWithIfaceMap(acpsModels[i], ld); err != nil {
				return
			}
		}

		for _, tpAcp := range acpsModels.AsTPActionProfile() {
			var acp *engine.ActionProfile
			if acp, err = engine.APItoActionProfile(tpAcp, tmz); err != nil {
				return
			}
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: ActionProfile: %s",
						utils.LoaderS, ldrID, utils.ToJSON(acp)))
				continue
			}
			if err = dm.SetActionProfile(ctx, acp, true); err != nil {
				return
			}
		}
	case utils.MetaAccounts:
		acpsModels := make(engine.AccountMdls, len(lDataSet))
		for i, ld := range lDataSet {
			acpsModels[i] = new(engine.AccountMdl)
			if err = utils.UpdateStructWithIfaceMap(acpsModels[i], ld); err != nil {
				return
			}
		}
		var accountTPModels []*utils.TPAccount
		if accountTPModels, err = acpsModels.AsTPAccount(); err != nil {
			return
		}
		for _, tpAcp := range accountTPModels {
			var acp *utils.Account
			if acp, err = engine.APItoAccount(tpAcp, tmz); err != nil {
				return
			}
			if dryRun {
				utils.Logger.Info(
					fmt.Sprintf("<%s-%s> DRY_RUN: Accounts: %s",
						utils.LoaderS, ldrID, utils.ToJSON(acp)))
				continue
			}
			if err = dm.SetAccount(ctx, acp, true); err != nil {
				return
			}
		}
	}
	return
}

func newLoader(cfg *config.CGRConfig, ldrCfg *config.LoaderSCfg, dm *engine.DataManager,
	timezone string, filterS *engine.FilterS, connMgr *engine.ConnManager, cacheConns []string) *loader {
	return &loader{
		cfg:        cfg,
		ldrCfg:     ldrCfg,
		dm:         dm,
		timezone:   timezone,
		filterS:    filterS,
		connMgr:    connMgr,
		cacheConns: cacheConns,
		dataCache:  ltcache.NewCache(-1, 0, false, nil),
		Locker:     newLocker(ldrCfg.LockFilePath),
	}
}

type loader struct {
	cfg        *config.CGRConfig
	ldrCfg     *config.LoaderSCfg
	dm         *engine.DataManager
	timezone   string
	filterS    *engine.FilterS
	connMgr    *engine.ConnManager
	cacheConns []string

	dataCache *ltcache.Cache
	Locker
}

func (l *loader) process(ctx *context.Context, tntID *utils.TenantID, lDataSet []utils.MapStorage, lType, action, caching string, dryRun, withIndex, partialRates bool) (err error) {
	if lType == "" { // do not set in DB; ToDo: how to determine if is cache or not
		return
	}
	switch action {
	case utils.MetaStore:
		err = setToDB(ctx, l.dm, lType, l.timezone, l.ldrCfg.ID, lDataSet, dryRun, withIndex, partialRates)
	case utils.MetaRemove:
		err = removeFromDB(ctx, l.dm, lType, tntID.Tenant, tntID.ID, l.ldrCfg.ID, dryRun, withIndex, partialRates, lDataSet[0])
	default:
		err = fmt.Errorf("unsupported loader action: <%q>", action)
	}
	if err != nil || dryRun ||
		len(l.cacheConns) == 0 {
		return
	}
	cacheArgs := make(map[string][]string)
	var cacheIDs []string // verify if we need to clear indexe
	tntId := tntID.TenantID()
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
		cacheArgs[utils.CacheRateProfiles] = []string{tntId}
	case utils.MetaAccounts:
		cacheIDs = []string{utils.CacheAccounts, utils.CacheAccountsFilterIndexes}

	}

	return engine.CallCache(l.connMgr, ctx, l.cacheConns, caching, cacheArgs, cacheIDs, nil, false, l.ldrCfg.Tenant)
}

func (l *loader) processData(ctx *context.Context, csv CSVReader, tmpls []*config.FCTemplate, lType, action, caching string, dryRun, withIndex, partialRates bool) (err error) {
	var prevTntID *utils.TenantID
	var lData []utils.MapStorage
	for lineNr := 1; ; lineNr++ {
		var record []string
		if record, err = csv.Read(); err != nil {
			if err == io.EOF {
				break
			}
			utils.Logger.Warning(
				fmt.Sprintf("<%s> <%s> reading file<%s> on line: %d, error: %s",
					utils.LoaderS, l.ldrCfg.ID, csv.Path(), lineNr, err))
			return
		}
		var data utils.MapStorage
		if data, err = newRecord(ctx, config.NewSliceDP(record, nil), tmpls, l.ldrCfg.Tenant, l.filterS, l.cfg, l.dataCache); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> <%s> file<%s> line: %d, error: %s",
					utils.LoaderS, l.ldrCfg.ID, csv.Path(), lineNr, err))
			return
		}
		tntID := TenantIDFromMap(data)
		if !prevTntID.Equal(tntID) {
			if err = l.process(ctx, prevTntID, lData, lType, action, caching, dryRun, withIndex, partialRates); err != nil {
				return
			}
			prevTntID = tntID
			lData = make([]utils.MapStorage, 0, 1)
		}
		lData = append(lData, data)
	}
	return l.process(ctx, prevTntID, lData, lType, action, caching, dryRun, withIndex, partialRates)
}

func (l *loader) processFile(ctx *context.Context, cfg *config.LoaderDataType, inPath, outPath, action, caching string, dryRun, withIndex bool) (err error) {
	csvType := utils.MetaFileCSV
	switch {
	case strings.HasPrefix(inPath, gprefix):
		csvType = utils.MetaGoogleAPI
		inPath = strings.TrimPrefix(inPath, gprefix)
	case utils.IsURL(inPath):
		csvType = utils.MetaUrl
	}
	var csv CSVReader
	if csv, err = NewCSVReader(csvType, inPath, cfg.Filename, rune(l.ldrCfg.FieldSeparator[0]), 0); err != nil {
		return
	}
	defer csv.Close()
	if err = l.processData(ctx, csv, cfg.Fields, cfg.Type, action, caching,
		dryRun, withIndex, cfg.Flags.GetBool(utils.MetaPartial)); err != nil || // encounterd error
		outPath == utils.EmptyString || // or no moving
		csvType != utils.MetaFileCSV { // or the type can not be moved(e.g. url)
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
	return
}

func (l *loader) processIFile(_, fileName string) (err error) {
	cfg := l.getCfg(fileName)
	if cfg == nil {
		return
	}

	if err = l.Lock(); err != nil {
		return
	}
	defer l.Unlock()
	return l.processFile(context.Background(), cfg, l.ldrCfg.TpInDir, l.ldrCfg.TpOutDir, l.ldrCfg.Action, l.ldrCfg.Caching, l.ldrCfg.DryRun, l.ldrCfg.WithIndex)
}

func (l *loader) processFolder(ctx *context.Context, action, caching string, dryRun, withIndex, stopOnError bool) (err error) {
	if err = l.Lock(); err != nil {
		return
	}
	defer l.Unlock()
	proces := func(i int) (err error) {
		cfg := l.ldrCfg.Data[i]
		if err = l.processFile(ctx, cfg, l.ldrCfg.TpInDir, l.ldrCfg.TpOutDir, action, caching, dryRun, withIndex); err != nil && !stopOnError {
			utils.Logger.Warning(fmt.Sprintf("<%s-%s> loaderType: <%s> cannot open files, err: %s",
				utils.LoaderS, l.ldrCfg.ID, cfg.Type, err))
			err = nil
		}
		return
	}
	switch l.ldrCfg.Action {
	case utils.MetaStore:
		for i := range l.ldrCfg.Data {
			if err = proces(i); err != nil {
				return
			}
		}
	case utils.MetaRemove:
		for i := len(l.ldrCfg.Data) - 1; i >= 0; i-- {
			if err = proces(i); err != nil {
				return
			}
		}
	}
	return
}

func (l *loader) handleFolder(stopChan chan struct{}) {
	for {
		go l.processFolder(context.Background(), l.ldrCfg.Action, l.ldrCfg.Caching, l.ldrCfg.DryRun, l.ldrCfg.WithIndex, false)
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
