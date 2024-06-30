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
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// onCacheEvicted is called by ltcache when evicting an item
func onCacheEvicted(_ string, value any) {
	ee := value.(EventExporter)
	ee.Close()
}

// NewEventExporterS instantiates the EventExporterS
func NewEventExporterS(cfg *config.CGRConfig, filterS *engine.FilterS,
	connMgr *engine.ConnManager) (eeS *EventExporterS) {
	eeS = &EventExporterS{
		cfg:     cfg,
		filterS: filterS,
		connMgr: connMgr,
		eesChs:  make(map[string]*ltcache.Cache),
	}
	eeS.setupCache(cfg.EEsNoLksCfg().Cache)
	return
}

// EventExporterS is managing the EventExporters
type EventExporterS struct {
	cfg     *config.CGRConfig
	filterS *engine.FilterS
	connMgr *engine.ConnManager

	eesChs map[string]*ltcache.Cache // map[eeType]*ltcache.Cache
	eesMux sync.RWMutex              // protects the eesChs
}

// ListenAndServe keeps the service alive
func (eeS *EventExporterS) ListenAndServe(stopChan, cfgRld chan struct{}) {
	for {
		select {
		case <-stopChan: // global exit
			return
		case rld := <-cfgRld: // configuration was reloaded, destroy the cache
			cfgRld <- rld
			utils.Logger.Info(fmt.Sprintf("<%s> reloading configuration internals.",
				utils.EEs))
			eeS.setupCache(eeS.cfg.EEsCfg().Cache)
		}
	}
}

// Shutdown is called to shutdown the service
func (eeS *EventExporterS) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.EEs))
	eeS.setupCache(nil) // cleanup exporters
}

// Call implements birpc.ClientConnector interface for internal RPC
func (eeS *EventExporterS) Call(ctx *context.Context, serviceMethod string, args any, reply any) error {
	return utils.RPCCall(eeS, serviceMethod, args, reply)
}

// setupCache deals with cleanup and initialization of the cache of EventExporters
func (eeS *EventExporterS) setupCache(chCfgs map[string]*config.CacheParamCfg) {
	eeS.eesMux.Lock()
	for chID, ch := range eeS.eesChs { // cleanup
		ch.Clear()
		delete(eeS.eesChs, chID)
	}
	for chID, chCfg := range chCfgs { // init
		if chCfg.Limit == 0 { // cache is disabled, will not create
			continue
		}
		eeS.eesChs[chID] = ltcache.NewCache(chCfg.Limit,
			chCfg.TTL, chCfg.StaticTTL, onCacheEvicted)
	}
	eeS.eesMux.Unlock()
}

func (eeS *EventExporterS) attrSProcessEvent(cgrEv *utils.CGREvent, attrIDs []string, ctx string) (*utils.CGREvent, error) {
	var rplyEv engine.AttrSProcessEventReply
	cgrEv.APIOpts[utils.MetaSubsys] = utils.MetaEEs
	cgrEv.APIOpts[utils.OptsAttributesProfileIDs] = attrIDs
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(ctx,
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]), utils.MetaEEs)
	err := eeS.connMgr.Call(context.TODO(),
		eeS.cfg.EEsNoLksCfg().AttributeSConns,
		utils.AttributeSv1ProcessEvent, cgrEv, &rplyEv)

	if err != nil && err.Error() != utils.ErrNotFound.Error() {
		return nil, err
	}
	if len(rplyEv.AlteredFields) != 0 {
		if !slices.ContainsFunc(rplyEv.AlteredFields,
			func(field string) bool {
				return strings.HasPrefix(
					field,
					utils.MetaReq+utils.NestingSep+utils.CostDetails,
				)
			},
		) {
			// CostDetails was not changed, its original value can be used safely.
			if _, has := cgrEv.Event[utils.CostDetails]; has {
				rplyEv.Event[utils.CostDetails] = cgrEv.Event[utils.CostDetails]
			}
			return rplyEv.CGREvent, nil
		}

		// If CostDetails key exists in Event, ensure its value
		// is of type *engine.EventCost.
		if cd, has := rplyEv.Event[utils.CostDetails]; has {
			if _, canCast := cd.(*engine.EventCost); !canCast {
				ec, err := engine.ConvertToEventCost(cd)
				if err != nil {
					return nil, err
				}
				rplyEv.Event[utils.CostDetails] = ec
			}
		}
		return rplyEv.CGREvent, nil
	}
	return cgrEv, nil
}

// V1ProcessEvent will be called each time a new event is received from readers
// rply -> map[string]map[string]any
func (eeS *EventExporterS) V1ProcessEvent(ctx *context.Context, cgrEv *engine.CGREventWithEeIDs, rply *map[string]map[string]any) (err error) {
	eeS.cfg.RLocks(config.EEsJson)
	defer eeS.cfg.RUnlocks(config.EEsJson)

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

		if cgrEv.APIOpts == nil {
			cgrEv.APIOpts = make(map[string]any)
		}
		cgrEv.APIOpts[utils.MetaExporterID] = eeCfg.ID

		if len(eeCfg.Filters) != 0 {
			tnt := utils.FirstNonEmpty(cgrEv.Tenant, eeS.cfg.GeneralCfg().DefaultTenant)
			if pass, errPass := eeS.filterS.Pass(tnt,
				eeCfg.Filters, cgrDp); errPass != nil {
				return errPass
			} else if !pass {
				continue // does not pass the filters, ignore the exporter
			}
		}

		exportEvent := cgrEv.CGREvent
		if eeCfg.Flags.GetBool(utils.MetaAttributes) {
			if exportEvent, err = eeS.attrSProcessEvent(
				cgrEv.CGREvent,
				eeCfg.AttributeSIDs,
				utils.FirstNonEmpty(
					eeCfg.AttributeSCtx,
					utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
					utils.MetaEEs)); err != nil {
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
			if ee, err = NewEventExporter(eeS.cfg.EEsCfg().Exporters[cfgIdx], eeS.cfg, eeS.filterS, eeS.connMgr); err != nil {
				return
			}
			if hasCache {
				eeS.eesMux.Lock()
				if _, has := eeCache.Get(eeCfg.ID); !has {
					eeCache.Set(eeCfg.ID, ee, nil)
				} else {
					// Another exporter instance with the same ID has been cached in
					// the meantime. Mark this instance to be closed after the export.
					hasCache = false
				}
				eeS.eesMux.Unlock()
			}
		}

		metricMapLock.Lock()
		metricsMap[ee.Cfg().ID] = utils.MapStorage{} // will return the ID for all processed exporters
		metricMapLock.Unlock()

		if eeCfg.Synchronous {
			wg.Add(1) // wait for sync to complete before returning
		}

		// log the message before starting the gorutine, but still execute the exporter
		if hasVerbose && !eeCfg.Synchronous {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> with id <%s>, running verbosed exporter with synchronous false",
					utils.EEs, ee.Cfg().ID))
		}
		go func(evict, sync bool, ee EventExporter) {
			if err := exportEventWithExporter(ee, exportEvent, evict, eeS.cfg, eeS.filterS); err != nil {
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

func exportEventWithExporter(exp EventExporter, ev *utils.CGREvent, oneTime bool, cfg *config.CGRConfig, filterS *engine.FilterS) (err error) {
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
		dsMap := map[string]utils.DataStorage{
			utils.MetaReq:  utils.MapStorage(ev.Event),
			utils.MetaDC:   exp.GetMetrics(),
			utils.MetaOpts: utils.MapStorage(ev.APIOpts),
			utils.MetaCfg:  cfg.GetDataProvider(),
			utils.MetaVars: utils.MapStorage{utils.MetaTenant: ev.Tenant},
		}

		var canCast bool
		dsMap[utils.MetaEC], canCast = ev.Event[utils.CostDetails].(*engine.EventCost)
		if !canCast {
			dsMap[utils.MetaEC] = engine.NewBareEventCost()
		}

		err = engine.NewExportRequest(dsMap,
			utils.FirstNonEmpty(ev.Tenant, cfg.GeneralCfg().DefaultTenant),
			filterS, map[string]*utils.OrderedNavigableMap{
				utils.MetaExp: expNM,
			}).SetFields(exp.Cfg().ContentFields())
		if eEv, err = exp.PrepareOrderMap(expNM); err != nil {
			return
		}
	}
	key := utils.ConcatenatedKey(utils.FirstNonEmpty(engine.MapEvent(ev.Event).GetStringIgnoreErrors(utils.CGRID), utils.GenUUID()),
		utils.FirstNonEmpty(engine.MapEvent(ev.Event).GetStringIgnoreErrors(utils.RunID), utils.MetaDefault))

	return ExportWithAttempts(exp, eEv, key)
}

func ExportWithAttempts(exp EventExporter, eEv any, key string) (err error) {
	if exp.Cfg().FailedPostsDir != utils.MetaNone {
		defer func() {
			if err != nil {
				AddFailedPost(exp.Cfg().FailedPostsDir, exp.Cfg().ExportPath,
					exp.Cfg().Type, eEv, exp.Cfg().Opts)
			}
		}()
	}
	fib := utils.FibDuration(time.Second, 0)

	for i := 0; i < exp.Cfg().Attempts; i++ {
		if err = exp.Connect(); err == nil {
			break
		}
		if i+1 < exp.Cfg().Attempts {
			time.Sleep(fib())
		}
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> Exporter <%s> could not connect because err: <%s>",
				utils.EEs, exp.Cfg().ID, err.Error()))
		return
	}

	if exp.Cfg().Flags.GetBool(utils.MetaLog) {
		var evLog string
		switch c := eEv.(type) {
		case []byte:
			evLog = string(c)
		case string:
			evLog = c
		case *HTTPPosterRequest:
			evByt, cancast := c.Body.([]byte)
			if cancast {
				evLog = string(evByt)
				break
			}
			evLog = utils.ToJSON(c.Body)
		default:
			evLog = utils.ToJSON(c)
		}
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, exporter <%s>, message: %s",
				utils.EEs, exp.Cfg().ID, evLog))
	}

	for i := 0; i < exp.Cfg().Attempts; i++ {
		if err = exp.ExportEvent(eEv, key); err == nil ||
			err == utils.ErrDisconnected { // special error in case the exporter was closed
			break
		}
		if i+1 < exp.Cfg().Attempts {
			time.Sleep(fib())
		}
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> Exporter <%s> could not export because err: <%s>",
				utils.EEs, exp.Cfg().ID, err.Error()))
	}
	return
}
