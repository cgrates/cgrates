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
	value.(EventExporter).Close()
}

// NewEventExporterS instantiates the EventExporterS
func NewEventExporterS(cfg *config.CGRConfig, filterS *engine.FilterS,
	connMgr *engine.ConnManager) (eeS *EeS) {
	eeS = &EeS{
		cfg:     cfg,
		fltrS:   filterS,
		connMgr: connMgr,
		eesChs:  make(map[string]*ltcache.Cache),
	}
	eeS.setupCache(cfg.EEsNoLksCfg().Cache)
	return
}

// EeS is managing the EventExporters
type EeS struct {
	cfg     *config.CGRConfig
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager

	eesChs map[string]*ltcache.Cache // map[eeType]*ltcache.Cache
	eesMux sync.RWMutex              // protects the eesChs
}

// ListenAndServe keeps the service alive
func (eeS *EeS) ListenAndServe(stopChan, cfgRld chan struct{}) {
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
func (eeS *EeS) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.EEs))
	eeS.setupCache(nil) // cleanup exporters
}

// setupCache deals with cleanup and initialization of the cache of EventExporters
func (eeS *EeS) setupCache(chCfgs map[string]*config.CacheParamCfg) {
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

func (eeS *EeS) attrSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent, attrIDs []string, attributeSCtx string) (err error) {
	var rplyEv engine.AttrSProcessEventReply
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(map[string]any)
	}
	cgrEv.APIOpts[utils.MetaSubsys] = utils.MetaEEs
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		attributeSCtx,
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
		utils.MetaEEs)
	cgrEv.APIOpts[utils.OptsAttributesProfileIDs] = attrIDs

	if err = eeS.connMgr.Call(ctx,
		eeS.cfg.EEsNoLksCfg().AttributeSConns,
		utils.AttributeSv1ProcessEvent,
		cgrEv, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
	} else if err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // cancel ErrNotFound
	}
	return
}

func ExportWithAttempts(ctx *context.Context, exp EventExporter, eEv any, key any,
	connMngr *engine.ConnManager, tnt string) (err error) {
	if exp.Cfg().FailedPostsDir != utils.MetaNone {
		defer func() {
			if err != nil {
				args := &utils.ArgsFailedPosts{
					Tenant:    tnt,
					Path:      exp.Cfg().ExportPath,
					Event:     eEv,
					FailedDir: exp.Cfg().FailedPostsDir,
					Module:    utils.EEs,
					APIOpts:   exp.Cfg().Opts.AsMapInterface(),
				}
				var reply string
				if err = connMngr.Call(ctx, exp.Cfg().EFsConns,
					utils.EfSv1ProcessEvent, args, &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> Exporter <%s> could not be written with <%s> service because err: <%s>",
							utils.EEs, exp.Cfg().ID, utils.EFs, err.Error()))
				}
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
	for i := 0; i < exp.Cfg().Attempts; i++ {
		if err = exp.ExportEvent(ctx, eEv, key); err == nil ||
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
