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

package ees

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (eeS *EeS) V1ProcessEvent(ctx *context.Context, cgrEv *utils.CGREventWithEeIDs, rply *map[string]map[string]any) (err error) {
	eeS.cfg.RLocks(config.EEsJSON)
	defer eeS.cfg.RUnlocks(config.EEsJSON)

	expIDs := utils.NewStringSet(cgrEv.EeIDs)
	lenExpIDs := expIDs.Size()
	cgrDp := utils.MapStorage{
		utils.MetaReq:  cgrEv.Event,
		utils.MetaOpts: cgrEv.APIOpts,
	}

	var wg sync.WaitGroup
	var withErr bool
	var metricMapLock sync.RWMutex
	metricsMap := make(map[string]utils.MapStorage)
	_, hasVerbose := cgrEv.APIOpts[utils.OptsEEsVerbose]
	for cfgIdx, eeCfg := range eeS.cfg.EEsNoLksCfg().Exporters {
		if eeCfg.Type == utils.MetaNone || // ignore *none type exporter
			(lenExpIDs != 0 && !expIDs.Has(eeCfg.ID)) {
			continue
		}

		if len(eeCfg.Filters) != 0 {
			tnt := utils.FirstNonEmpty(cgrEv.Tenant, eeS.cfg.GeneralCfg().DefaultTenant)
			if pass, errPass := eeS.fltrS.Pass(ctx, tnt,
				eeCfg.Filters, cgrDp); errPass != nil {
				return errPass
			} else if !pass {
				continue // does not pass the filters, ignore the exporter
			}
		}

		if eeCfg.Flags.GetBool(utils.MetaAttributes) {
			if err = eeS.attrSProcessEvent(ctx, cgrEv.CGREvent,
				eeCfg.AttributeSIDs, eeCfg.AttributeSCtx); err != nil {
				return
			}
		}

		eeS.eesMux.RLock()
		eeCache, hasCache := eeS.eesChs[eeCfg.Type]
		eeS.eesMux.RUnlock()
		var isCached bool
		var ee EventExporter
		if hasCache {
			var x any
			if x, isCached = eeCache.Get(eeCfg.ID); isCached {
				ee = x.(EventExporter)
			}
		}
		if !isCached {
			if ee, err = NewEventExporter(eeS.cfg.EEsCfg().Exporters[cfgIdx], eeS.cfg, eeS.fltrS, eeS.connMgr); err != nil {
				return
			}
			if hasCache {
				eeCache.Set(eeCfg.ID, ee, nil)
			}
		}

		metricMapLock.Lock()
		metricsMap[ee.Cfg().ID] = utils.MapStorage{} // will return the ID for all processed exporters
		metricMapLock.Unlock()
		ctx := ctx
		if eeCfg.Synchronous {
			wg.Add(1) // wait for synchronous or file ones since these need to be done before continuing
		} else {
			ctx = context.Background() // is async so lose the API context
		}
		// log the message before starting the goroutine, but still execute the exporter
		if hasVerbose && !eeCfg.Synchronous {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> with id <%s>, running verbosed exporter with syncronous false",
					utils.EEs, ee.Cfg().ID))
		}
		go func(evict, sync bool, ee EventExporter) {
			if err := exportEventWithExporter(ctx, ee, eeS.connMgr, cgrEv.CGREvent, evict, eeS.cfg, eeS.fltrS, cgrEv.Tenant); err != nil {
				withErr = true
			}
			if sync {
				if hasVerbose {
					metricMapLock.Lock()
					metricsMap[ee.Cfg().ID] = ee.GetMetrics().ClonedMapStorage()
					metricMapLock.Unlock()
				}
				wg.Done()
			}
		}(!hasCache, eeCfg.Synchronous, ee)
		if eeCfg.Blocker {
			break
		}
	}
	wg.Wait()
	if withErr {
		err = utils.ErrPartiallyExecuted
		return
	}

	*rply = make(map[string]map[string]any)
	metricMapLock.Lock()
	for exporterID, metrics := range metricsMap {
		(*rply)[exporterID] = make(map[string]any)
		for key, val := range metrics {
			switch key {
			case utils.PositiveExports, utils.NegativeExports:
				slsVal, canCast := val.(utils.StringSet)
				if !canCast {
					return fmt.Errorf("cannot cast to map[string]any %+v for positive exports", val)
				}
				(*rply)[exporterID][key] = slsVal.AsSlice()
			default:
				(*rply)[exporterID][key] = val
			}
		}
	}
	metricMapLock.Unlock()
	if len(*rply) == 0 {
		return utils.ErrNotFound
	}
	return
}

type ArchiveEventsArgs struct {
	Tenant  string
	APIOpts map[string]any
	Events  []*utils.EventsWithOpts
}

func exportEventWithExporter(ctx *context.Context, exp EventExporter, connMngr *engine.ConnManager,
	ev *utils.CGREvent, oneTime bool, cfg *config.CGRConfig, filterS *engine.FilterS, tnt string) (err error) {
	defer func() {
		updateEEMetrics(exp.GetMetrics(), ev.ID, ev.Event, err != nil, utils.FirstNonEmpty(exp.Cfg().Timezone,
			cfg.GeneralCfg().DefaultTimezone))
		if oneTime {
			exp.Close()
		}
	}()
	var eEv any

	exp.GetMetrics().Lock()
	exp.GetMetrics().MapStorage[utils.NumberOfEvents] = exp.GetMetrics().MapStorage[utils.NumberOfEvents].(int64) + 1
	exp.GetMetrics().Unlock()
	if len(exp.Cfg().ContentFields()) == 0 {
		if eEv, err = exp.PrepareMap(ev); err != nil {
			return
		}
	} else {
		expNM := utils.NewOrderedNavigableMap()
		err = engine.NewExportRequest(map[string]utils.DataStorage{
			utils.MetaReq:  utils.MapStorage(ev.Event),
			utils.MetaDC:   exp.GetMetrics(),
			utils.MetaOpts: utils.MapStorage(ev.APIOpts),
			utils.MetaCfg:  cfg.GetDataProvider(),
			utils.MetaVars: utils.MapStorage{utils.MetaTenant: ev.Tenant},
		}, utils.FirstNonEmpty(ev.Tenant, cfg.GeneralCfg().DefaultTenant),
			filterS,
			map[string]*utils.OrderedNavigableMap{utils.MetaExp: expNM}).SetFields(ctx, exp.Cfg().ContentFields())
		if eEv, err = exp.PrepareOrderMap(expNM); err != nil {
			return
		}
	}
	extraData := exp.ExtraData(ev)

	return ExportWithAttempts(ctx, exp, eEv, extraData, connMngr, tnt)
}

// V1ArchiveEventsInReply should archive the events sent with existing exporters. The zipped content should be returned back as a reply.
func (eeS *EeS) V1ArchiveEventsInReply(ctx *context.Context, args *ArchiveEventsArgs, reply *[]byte) (err error) {
	if args.Tenant == utils.EmptyString {
		args.Tenant = eeS.cfg.GeneralCfg().DefaultTenant
	}
	expID, has := args.APIOpts[utils.MetaExporterID]
	if !has {
		return fmt.Errorf("ExporterID is missing from argument's options")
	}
	// check if there are any exporters that match our expID
	var eesCfg *config.EventExporterCfg
	for _, exporter := range eeS.cfg.EEsCfg().Exporters {
		if exporter.ID == expID {
			eesCfg = exporter
			break
		}
	}
	if eesCfg == nil {
		return fmt.Errorf("exporter config with ID: %s is missing", expID)
	}
	// also mandatory to be synchronous
	if !eesCfg.Synchronous {
		return fmt.Errorf("exporter with ID: %s is not synchronous", expID)
	}
	// also mandatory to be type of *buffer
	if eesCfg.ExportPath != utils.MetaBuffer {
		return fmt.Errorf("exporter with ID: %s has an invalid ExportPath for archiving", expID)
	}
	var dc *utils.SafeMapStorage
	if dc, err = newEEMetrics(utils.FirstNonEmpty(
		eesCfg.Timezone,
		eeS.cfg.GeneralCfg().DefaultTimezone)); err != nil {
		return
	}
	var ee EventExporter

	buff := new(bytes.Buffer)
	zBuff := zip.NewWriter(buff)
	var wrtr io.Writer
	// create the file where will be stored in zip
	if wrtr, err = zBuff.CreateHeader(&zip.FileHeader{
		Method:   zip.Deflate, // to be compressed
		Name:     "events.csv",
		Modified: time.Now(),
	}); err != nil {
		return err
	}
	switch eesCfg.Type {
	case utils.MetaFileCSV:
		ee, err = NewFileCSVee(eesCfg, eeS.cfg, eeS.fltrS, dc, &buffer{wrtr})
	case utils.MetaFileFWV:
		ee, err = NewFileFWVee(eesCfg, eeS.cfg, eeS.fltrS, dc, &buffer{wrtr})
	default:
		err = fmt.Errorf("unsupported exporter type: %s>", eesCfg.Type)
	}
	if err != nil {
		return err
	}
	// we will build the dataPrvider in order to match the filters
	cgrDp := utils.MapStorage{
		utils.MetaOpts: args.APIOpts,
	}
	// check if APIOpts have to ignore the filters
	var ignoreFltr bool
	if val, has := args.APIOpts[utils.MetaProfileIgnoreFilters]; has {
		ignoreFltr, err = utils.IfaceAsBool(val)
		if err != nil {
			return err
		}
	}
	tnt := utils.FirstNonEmpty(args.Tenant, eeS.cfg.GeneralCfg().DefaultTenant)
	var exported bool
	for _, event := range args.Events {
		if len(eesCfg.Filters) != 0 && !ignoreFltr {
			cgrDp[utils.MetaReq] = event.Event
			if pass, errPass := eeS.fltrS.Pass(ctx, tnt,
				eesCfg.Filters, cgrDp); errPass != nil {
				return errPass
			} else if !pass {
				continue // does not pass the filters, ignore the exporter
			}
		}
		// in case of event's Opts got another *exporterID that is different from the initial Opts, will skip that Event and will continue the iterations
		if newExpID, ok := event.Opts[utils.MetaExporterID]; ok && newExpID != expID {
			continue
		}
		cgrEv := &utils.CGREvent{
			ID:      utils.UUIDSha1Prefix(),
			Tenant:  tnt,
			Event:   event.Event,
			APIOpts: make(map[string]any),
		}
		// here we will join the APIOpts from the initial args and Opts from every CDR(EventsWithOPts)
		for key, val := range args.APIOpts {
			if _, ok := event.Opts[key]; ok {
				val = event.Opts[key]
			}
			event.Opts[key] = val
		}
		cgrEv.APIOpts = event.Opts

		// exported will be true if there will be at least one exporter archived
		exported = true
		if err = exportEventWithExporter(ctx, ee, eeS.connMgr, cgrEv, false, eeS.cfg, eeS.fltrS, cgrEv.Tenant); err != nil {
			return err
		}
	}
	// most probably beacause of not matching filters
	if !exported {
		return utils.NewErrServerError(fmt.Errorf("NO EXPORTS"))
	}
	if err = ee.Close(); err != nil {
		return err
	}
	if err = zBuff.Close(); err != nil {
		return err
	}
	*reply = buff.Bytes()
	return
}
