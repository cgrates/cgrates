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

// NewERService instantiates the EventExporterS
func NewEventExporterS(cfg *config.CGRConfig, filterS *engine.FilterS,
	connMgr *engine.ConnManager) *EventExporterS {
	return &EventExporterS{
		cfg:     cfg,
		filterS: filterS,
		connMgr: connMgr,
		eesChs:  make(map[string]*ltcache.Cache),
	}
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
func (eeS *EventExporterS) ListenAndServe(exitChan chan bool, cfgRld chan struct{}) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s>",
		utils.CoreS, utils.EventExporterS))
	for {
		select {
		case e := <-exitChan: // global exit
			eeS.Shutdown()
			exitChan <- e // put back for the others listening for shutdown request
			break
		case rld := <-cfgRld: // configuration was reloaded, destroy the cache
			cfgRld <- rld
			utils.Logger.Info(fmt.Sprintf("<%s> reloading configuration internals.",
				utils.EventExporterS))
			eeS.setupCache(eeS.cfg.EEsCfg().Cache)
		}
	}
	return
}

// Shutdown is called to shutdown the service
func (eeS *EventExporterS) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.EventExporterS))
	eeS.setupCache(nil) // cleanup exporters
	return
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

func (eeS *EventExporterS) attrSProcessEvent(cgrEv *utils.CGREventWithOpts, attrIDs []string, ctx string) (err error) {
	var rplyEv engine.AttrSProcessEventReply
	attrArgs := &engine.AttrArgsProcessEvent{
		AttributeIDs:  attrIDs,
		Context:       utils.StringPointer(ctx),
		CGREvent:      cgrEv.CGREvent,
		ArgDispatcher: cgrEv.ArgDispatcher,
	}
	if err = eeS.connMgr.Call(
		eeS.cfg.EEsNoLksCfg().AttributeSConns, nil,
		utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
		cgrEv.CGREvent = rplyEv.CGREvent
		cgrEv.Opts = rplyEv.Opts
	} else if err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // cancel ErrNotFound
	}
	return
}

// ProcessEvent will be called each time a new event is received from readers
func (eeS *EventExporterS) V1ProcessEvent(cgrEv *utils.CGREventWithOpts, rply *string) (err error) {
	eeS.cfg.RLocks(config.EEsJson)
	defer eeS.cfg.RUnlocks(config.EEsJson)

	var wg sync.WaitGroup
	var withErr bool
	for cfgIdx, eeCfg := range eeS.cfg.EEsNoLksCfg().Exporters {

		if len(eeCfg.Filters) != 0 {
			cgrDp := config.NewNavigableMap(map[string]interface{}{
				utils.MetaReq: cgrEv.Event,
			})
			tnt := cgrEv.Tenant
			if eeTnt, errTnt := eeCfg.Tenant.ParseEvent(cgrEv.Event); errTnt == nil && eeTnt != utils.EmptyString {
				tnt = eeTnt
			}
			if pass, errPass := eeS.filterS.Pass(tnt,
				eeCfg.Filters, cgrDp); errPass != nil || !pass {
				continue // does not pass the filters, ignore the exporter
			}
		}

		if eeCfg.Flags.GetBool(utils.MetaAttributes) {
			if err = eeS.attrSProcessEvent(
				cgrEv,
				eeCfg.AttributeSIDs,
				utils.FirstNonEmpty(
					eeCfg.AttributeSCtx,
					utils.IfaceAsString(cgrEv.Opts[utils.Context]),
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
			if x, isCached := eeCache.Get(eeCfg.ID); isCached {
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
		go func(evict, sync bool) {
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
				wg.Done()
			}
		}(!hasCache, eeCfg.Synchronous)
	}
	wg.Wait()
	if withErr {
		err = utils.ErrPartiallyExecuted
	}

	return
}
