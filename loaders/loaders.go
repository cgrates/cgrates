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
	"errors"
	"fmt"
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func NewLoaderService(cfg *config.CGRConfig, dm *engine.DataManager,
	filterS *engine.FilterS,
	connMgr *engine.ConnManager) (ldrS *LoaderS) {
	ldrS = &LoaderS{cfg: cfg, cache: make(map[string]*ltcache.Cache)}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		ldrS.cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ldrS.createLoaders(dm, filterS, connMgr)
	return
}

// LoaderS is the Loader service handling independent Loaders
type LoaderS struct {
	sync.RWMutex
	cfg   *config.CGRConfig
	cache map[string]*ltcache.Cache
	ldrs  map[string]*loader
}

// Enabled returns true if at least one loader is enabled
func (ldrS *LoaderS) Enabled() bool {
	return len(ldrS.ldrs) != 0
}

func (ldrS *LoaderS) ListenAndServe(stopChan chan struct{}) (err error) {
	for _, ldr := range ldrS.ldrs {
		if err = ldr.ListenAndServe(stopChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s-%s> error: <%s>", utils.LoaderS, ldr.ldrCfg.ID, err.Error()))
			return
		}
	}
	return
}

type ArgsProcessFolder struct {
	LoaderID string
	APIOpts  map[string]interface{}
}

func (ldrS *LoaderS) V1Run(ctx *context.Context, args *ArgsProcessFolder,
	rply *string) (err error) {
	ldrS.RLock()
	defer ldrS.RUnlock()

	if args.LoaderID == utils.EmptyString {
		args.LoaderID = utils.MetaDefault
	}
	ldr, has := ldrS.ldrs[args.LoaderID]
	if !has {
		return fmt.Errorf("UNKNOWN_LOADER: %s", args.LoaderID)
	}
	var locked bool
	if locked, err = ldr.Locked(); err != nil {
		return utils.NewErrServerError(err)
	} else if locked {
		fl := ldr.ldrCfg.Opts.ForceLock
		if val, has := args.APIOpts[utils.MetaForceLock]; has {
			if fl, err = utils.IfaceAsBool(val); err != nil {
				return
			}
		}
		if !fl {
			return errors.New("ANOTHER_LOADER_RUNNING")
		}
		if err := ldr.Unlock(); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	wI := ldr.ldrCfg.Opts.WithIndex
	if val, has := args.APIOpts[utils.MetaWithIndex]; has {
		if wI, err = utils.IfaceAsBool(val); err != nil {
			return
		}
	}

	soE := ldr.ldrCfg.Opts.StopOnError
	if val, has := args.APIOpts[utils.MetaStopOnError]; has {
		if soE, err = utils.IfaceAsBool(val); err != nil {
			return
		}
	}
	if err := ldr.processFolder(context.Background(), utils.FirstNonEmpty(utils.IfaceAsString(args.APIOpts[utils.MetaCache]), ldr.ldrCfg.Opts.Cache, ldrS.cfg.GeneralCfg().DefaultCaching),
		wI, soE); err != nil {
		return utils.NewErrServerError(err)
	}
	*rply = utils.OK
	return
}

// Reload recreates the loaders map thread safe
func (ldrS *LoaderS) Reload(dm *engine.DataManager,
	filterS *engine.FilterS, connMgr *engine.ConnManager) {
	ldrS.Lock()
	ldrS.createLoaders(dm, filterS, connMgr)
	ldrS.Unlock()
}

// Reload recreates the loaders map thread safe
func (ldrS *LoaderS) createLoaders(dm *engine.DataManager,
	filterS *engine.FilterS, connMgr *engine.ConnManager) {
	ldrS.ldrs = make(map[string]*loader)
	for _, ldrCfg := range ldrS.cfg.LoaderCfg() {
		if ldrCfg.Enabled {
			ldrS.ldrs[ldrCfg.ID] = newLoader(ldrS.cfg, ldrCfg, dm, ldrS.cache, filterS, connMgr, ldrCfg.CacheSConns)
		}
	}
}
