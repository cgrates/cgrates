/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ees

import (
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// onCacheEvicted is called by ltcache when evicting an item
func onCacheEvicted(_ string, value any) {
	ee := value.(EventExporter)
	ee.GetMetrics().StopCron()
	ee.Close()
}

// NewEventExporterS initializes a new EventExporterS.
func NewEventExporterS(cfg *config.CGRConfig, fltrS *engine.FilterS,
	connMgr *engine.ConnManager) (*EeS, error) {
	eeS := &EeS{
		cfg:     cfg,
		fltrS:   fltrS,
		connMgr: connMgr,
	}
	if err := eeS.SetupExporterCache(); err != nil {
		return nil, fmt.Errorf("failed to set up exporter cache: %v", err)
	}
	return eeS, nil
}

// EeS is managing the EventExporters
type EeS struct {
	cfg     *config.CGRConfig
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager

	exporterCache map[string]*ltcache.Cache // map[eeType]*ltcache.Cache
	mu            sync.RWMutex              // protects exporterCache
}

// ClearExporterCache clears the cache of EventExporters.
func (eeS *EeS) ClearExporterCache() {
	eeS.mu.Lock()
	defer eeS.mu.Unlock()
	for chID, ch := range eeS.exporterCache {
		ch.Clear()
		delete(eeS.exporterCache, chID)
	}
}

// SetupExporterCache initializes the cache for EventExporters.
func (eeS *EeS) SetupExporterCache() error {
	expCache := make(map[string]*ltcache.Cache)
	eesCfg := eeS.cfg.EEsNoLksCfg()

	// Initialize cache.
	for chID, chCfg := range eesCfg.Cache {
		if chCfg.Limit == 0 {
			continue // skip if caching is disabled
		}

		expCache[chID] = ltcache.NewCache(chCfg.Limit, chCfg.TTL, chCfg.StaticTTL, false,
			[]func(itmID string, value any){
				onCacheEvicted,
			})

		// Precache exporters if required.
		if chCfg.Precache {
			for _, expCfg := range eesCfg.Exporters {
				if expCfg.Type == chID {
					ee, err := NewEventExporter(expCfg, eeS.cfg, eeS.fltrS, eeS.connMgr)
					if err != nil {
						return fmt.Errorf("precache: failed to init EventExporter %q: %v", expCfg.ID, err)
					}
					expCache[chID].Set(expCfg.ID, ee, nil)
				}
			}
		}
	}
	eeS.exporterCache = expCache
	return nil
}

func (eeS *EeS) attrSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent, attrIDs []string, attributeSCtx string) (err error) {
	var rplyEv attributes.AttrSProcessEventReply
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
