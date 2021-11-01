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
)

func NewLoaderService(cfg *config.CGRConfig, dm *engine.DataManager,
	timezone string, filterS *engine.FilterS,
	connMgr *engine.ConnManager) (ldrS *LoaderService) {
	ldrS = &LoaderService{cfg: cfg}
	ldrS.createLoaders(dm, timezone, filterS, connMgr)
	return
}

// LoaderService is the Loader service handling independent Loaders
type LoaderService struct {
	sync.RWMutex
	cfg  *config.CGRConfig
	ldrs map[string]*loader
}

// Enabled returns true if at least one loader is enabled
func (ldrS *LoaderService) Enabled() bool {
	return len(ldrS.ldrs) != 0
}

func (ldrS *LoaderService) ListenAndServe(stopChan chan struct{}) (err error) {
	for _, ldr := range ldrS.ldrs {
		if err = ldr.ListenAndServe(stopChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s-%s> error: <%s>", utils.LoaderS, ldr.ldrCfg.ID, err.Error()))
			return
		}
	}
	return
}

type ArgsProcessFolder struct {
	LoaderID    string
	ForceLock   bool
	Caching     *string
	StopOnError bool
}

func (ldrS *LoaderService) V1Load(ctx *context.Context, args *ArgsProcessFolder,
	rply *string) (err error) {
	return ldrS.process(ctx, args, utils.MetaStore, rply)
}

func (ldrS *LoaderService) V1Remove(ctx *context.Context, args *ArgsProcessFolder,
	rply *string) (err error) {
	return ldrS.process(ctx, args, utils.MetaRemove, rply)
}

// Reload recreates the loaders map thread safe
func (ldrS *LoaderService) Reload(dm *engine.DataManager,
	timezone string, filterS *engine.FilterS, connMgr *engine.ConnManager) {
	ldrS.Lock()
	ldrS.createLoaders(dm, timezone, filterS, connMgr)
	ldrS.Unlock()
}

// Reload recreates the loaders map thread safe
func (ldrS *LoaderService) createLoaders(dm *engine.DataManager,
	timezone string, filterS *engine.FilterS, connMgr *engine.ConnManager) {
	ldrS.ldrs = make(map[string]*loader)
	for _, ldrCfg := range ldrS.cfg.LoaderCfg() {
		if ldrCfg.Enabled {
			ldrS.ldrs[ldrCfg.ID] = newLoader(ldrS.cfg, ldrCfg, dm, timezone, filterS, connMgr, ldrCfg.CacheSConns)
		}
	}
}

func (ldrS *LoaderService) process(ctx *context.Context, args *ArgsProcessFolder, action string,
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
	if locked, err := ldr.Locked(); err != nil {
		return utils.NewErrServerError(err)
	} else if locked {
		if !args.ForceLock {
			return errors.New("ANOTHER_LOADER_RUNNING")
		}
		if err := ldr.Unlock(); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	//verify If Caching is present in arguments
	caching := utils.FirstNonEmpty(ldr.ldrCfg.Caching, ldrS.cfg.GeneralCfg().DefaultCaching)
	if args.Caching != nil {
		caching = *args.Caching
	}

	if err := ldr.processFolder(context.Background(), action, caching, false, ldr.ldrCfg.WithIndex, args.StopOnError); err != nil {
		return utils.NewErrServerError(err)
	}
	*rply = utils.OK
	return
}
