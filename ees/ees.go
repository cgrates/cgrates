/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// onCacheEvicted is called by ltcache when evicting an item
func onCacheEvicted(itmID string, value interface{}) {
	ee := value.(EventExporter)
	ee.OnEvicted(itmID, value)
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
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s>",
		utils.CoreS, utils.EventExporterS))
	for {
		select {
		case <-stopChan: // global exit
			return
		case rld := <-cfgRld: // configuration was reloaded, destroy the cache
			cfgRld <- rld
			utils.Logger.Info(fmt.Sprintf("<%s> reloading configuration internals.",
				utils.EventExporterS))
			eeS.setupCache(eeS.cfg.EEsCfg().Cache)
		}
	}
}

// Shutdown is called to shutdown the service
func (eeS *EventExporterS) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.EventExporterS))
	eeS.setupCache(nil) // cleanup exporters
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (eeS *EventExporterS) Call(serviceMethod string, args interface{}, reply interface{}) error {
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

func (eeS *EventExporterS) attrSProcessEvent(cgrEv *utils.CGREvent, attrIDs []string, ctx string) (err error) {
	var rplyEv engine.AttrSProcessEventReply
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(map[string]interface{})
	}
	cgrEv.APIOpts[utils.Subsys] = utils.MetaEEs
	var processRuns *int
	if val, has := cgrEv.APIOpts[utils.OptsAttributesProcessRuns]; has {
		if v, err := utils.IfaceAsTInt64(val); err == nil {
			processRuns = utils.IntPointer(int(v))
		}
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		AttributeIDs: attrIDs,
		Context:      utils.StringPointer(ctx),
		CGREvent:     cgrEv,
		ProcessRuns:  processRuns,
	}
	if err = eeS.connMgr.Call(
		eeS.cfg.EEsNoLksCfg().AttributeSConns, nil,
		utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
		cgrEv = rplyEv.CGREvent
	} else if err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // cancel ErrNotFound
	}
	return
}

// V1ProcessEvent will be called each time a new event is received from readers
// rply -> map[string]map[string]interface{}
func (eeS *EventExporterS) V1ProcessEvent(cgrEv *utils.CGREventWithEeIDs, rply *map[string]map[string]interface{}) (err error) {
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

		if len(eeCfg.Filters) != 0 {
			tnt := utils.FirstNonEmpty(cgrEv.Tenant, eeS.cfg.GeneralCfg().DefaultTenant)
			if eeTnt, errTnt := eeCfg.Tenant.ParseDataProvider(cgrDp); errTnt == nil && eeTnt != utils.EmptyString {
				tnt = eeTnt
			}
			if pass, errPass := eeS.filterS.Pass(tnt,
				eeCfg.Filters, cgrDp); errPass != nil {
				return errPass
			} else if !pass {
				continue // does not pass the filters, ignore the exporter
			}
		}

		if eeCfg.Flags.GetBool(utils.MetaAttributes) {
			if err = eeS.attrSProcessEvent(
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
			var x interface{}
			if x, isCached = eeCache.Get(eeCfg.ID); isCached {
				ee = x.(EventExporter)
			}
		}
		if !isCached {
			if ee, err = NewEventExporter(eeS.cfg, cfgIdx, eeS.filterS); err != nil {
				return
			}
			if hasCache {
				eeCache.Set(eeCfg.ID, ee, nil)
			}
		}
		if eeCfg.Synchronous {
			wg.Add(1) // wait for synchronous or file ones since these need to be done before continuing
		}
		metricMapLock.Lock()
		metricsMap[ee.ID()] = utils.MapStorage{}
		metricMapLock.Unlock()
		// log the message before starting the gorutine, but still execute the exporter
		if hasVerbose && !eeCfg.Synchronous {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> with id <%s>, running verbosed exporter with syncronous false",
					utils.EventExporterS, ee.ID()))
		}
		go func(evict, sync bool, ee EventExporter) {
			if err := ee.ExportEvent(cgrEv.CGREvent); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> with id <%s>, error: <%s>",
						utils.EventExporterS, ee.ID(), err.Error()))
				withErr = true
			}
			if evict {
				ee.OnEvicted("", nil) // so we can close ie the file
			}
			if sync {
				if hasVerbose {
					metricMapLock.Lock()
					metricsMap[ee.ID()] = ee.GetMetrics()
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

	*rply = make(map[string]map[string]interface{})
	metricMapLock.Lock()
	for exporterID, metrics := range metricsMap {
		(*rply)[exporterID] = make(map[string]interface{})
		for key, val := range metrics {
			switch key {
			case utils.PositiveExports, utils.NegativeExports:
				slsVal, canCast := val.(utils.StringSet)
				if !canCast {
					return fmt.Errorf("cannot cast to map[string]interface{} %+v for positive exports", val)
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

func newEEMetrics(location string) (utils.MapStorage, error) {
	tNow := time.Now()
	loc, err := time.LoadLocation(location)
	if err != nil {
		return nil, err
	}
	return utils.MapStorage{
		utils.NumberOfEvents:  int64(0),
		utils.PositiveExports: utils.StringSet{},
		utils.NegativeExports: utils.StringSet{},
		utils.TimeNow: time.Date(tNow.Year(), tNow.Month(), tNow.Day(),
			tNow.Hour(), tNow.Minute(), tNow.Second(), tNow.Nanosecond(), loc),
	}, nil
}

func updateEEMetrics(dc utils.MapStorage, ev engine.MapEvent, timezone string) {
	if aTime, err := ev.GetTime(utils.AnswerTime, timezone); err == nil {
		if _, has := dc[utils.FirstEventATime]; !has {
			dc[utils.FirstEventATime] = time.Time{}
		}
		if _, has := dc[utils.LastEventATime]; !has {
			dc[utils.LastEventATime] = time.Time{}
		}
		if dc[utils.FirstEventATime].(time.Time).IsZero() ||
			aTime.Before(dc[utils.FirstEventATime].(time.Time)) {
			dc[utils.FirstEventATime] = aTime
		}
		if aTime.After(dc[utils.LastEventATime].(time.Time)) {
			dc[utils.LastEventATime] = aTime
		}
	}
	if oID, err := ev.GetTInt64(utils.OrderID); err == nil {
		if _, has := dc[utils.FirstExpOrderID]; !has {
			dc[utils.FirstExpOrderID] = int64(0)
		}
		if _, has := dc[utils.LastExpOrderID]; !has {
			dc[utils.LastExpOrderID] = int64(0)
		}
		if dc[utils.FirstExpOrderID].(int64) == 0 ||
			dc[utils.FirstExpOrderID].(int64) > oID {
			dc[utils.FirstExpOrderID] = oID
		}
		if dc[utils.LastExpOrderID].(int64) < oID {
			dc[utils.LastExpOrderID] = oID
		}
	}
	if cost, err := ev.GetFloat64(utils.Cost); err == nil {
		if _, has := dc[utils.TotalCost]; !has {
			dc[utils.TotalCost] = float64(0.0)
		}
		dc[utils.TotalCost] = dc[utils.TotalCost].(float64) + cost
	}
	if tor, err := ev.GetString(utils.ToR); err == nil {
		if usage, err := ev.GetDuration(utils.Usage); err == nil {
			switch tor {
			case utils.MetaVoice:
				if _, has := dc[utils.TotalDuration]; !has {
					dc[utils.TotalDuration] = time.Duration(0)
				}
				dc[utils.TotalDuration] = dc[utils.TotalDuration].(time.Duration) + usage
			case utils.MetaSMS:
				if _, has := dc[utils.TotalSMSUsage]; !has {
					dc[utils.TotalSMSUsage] = time.Duration(0)
				}
				dc[utils.TotalSMSUsage] = dc[utils.TotalSMSUsage].(time.Duration) + usage
			case utils.MetaMMS:
				if _, has := dc[utils.TotalMMSUsage]; !has {
					dc[utils.TotalMMSUsage] = time.Duration(0)
				}
				dc[utils.TotalMMSUsage] = dc[utils.TotalMMSUsage].(time.Duration) + usage
			case utils.MetaGeneric:
				if _, has := dc[utils.TotalGenericUsage]; !has {
					dc[utils.TotalGenericUsage] = time.Duration(0)
				}
				dc[utils.TotalGenericUsage] = dc[utils.TotalGenericUsage].(time.Duration) + usage
			case utils.MetaData:
				if _, has := dc[utils.TotalDataUsage]; !has {
					dc[utils.TotalDataUsage] = time.Duration(0)
				}
				dc[utils.TotalDataUsage] = dc[utils.TotalDataUsage].(time.Duration) + usage
			}
		}
	}
}
