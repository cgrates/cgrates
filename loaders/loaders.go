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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewLoaderService(dm *engine.DataManager, ldrsCfg []*config.LoaderSCfg,
	timezone string, filterS *engine.FilterS,
	connMgr *engine.ConnManager) (ldrS *LoaderService) {
	ldrS = &LoaderService{ldrs: make(map[string]*Loader)}
	for _, ldrCfg := range ldrsCfg {
		if ldrCfg.Enabled {
			ldrS.ldrs[ldrCfg.ID] = NewLoader(dm, ldrCfg, timezone, filterS, connMgr, ldrCfg.CacheSConns)
		}
	}
	return
}

// LoaderService is the Loader service handling independent Loaders
type LoaderService struct {
	sync.RWMutex
	ldrs map[string]*Loader
}

// Enabled returns true if at least one loader is enabled
func (ldrS *LoaderService) Enabled() bool {
	return len(ldrS.ldrs) != 0
}

func (ldrS *LoaderService) ListenAndServe(exitChan chan struct{}) (err error) {
	for _, ldr := range ldrS.ldrs {
		if err = ldr.ListenAndServe(exitChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s-%s> error: <%s>", utils.LoaderS, ldr.ldrID, err.Error()))
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

func (ldrS *LoaderService) V1Load(args *ArgsProcessFolder,
	rply *string) (err error) {
	ldrS.RLock()
	defer ldrS.RUnlock()
	if args.LoaderID == "" {
		args.LoaderID = utils.MetaDefault
	}
	ldr, has := ldrS.ldrs[args.LoaderID]
	if !has {
		return fmt.Errorf("UNKNOWN_LOADER: %s", args.LoaderID)
	}
	if locked, err := ldr.isFolderLocked(); err != nil {
		return utils.NewErrServerError(err)
	} else if locked {
		if !args.ForceLock {
			return errors.New("ANOTHER_LOADER_RUNNING")
		}
		if err := ldr.unlockFolder(); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	//verify If Caching is present in arguments
	caching := config.CgrConfig().GeneralCfg().DefaultCaching
	if args.Caching != nil {
		caching = *args.Caching
	}
	if err := ldr.ProcessFolder(caching, utils.MetaStore, args.StopOnError); err != nil {
		return utils.NewErrServerError(err)
	}
	*rply = utils.OK
	return
}

func (ldrS *LoaderService) V1Remove(args *ArgsProcessFolder,
	rply *string) (err error) {
	ldrS.RLock()
	defer ldrS.RUnlock()
	if args.LoaderID == "" {
		args.LoaderID = utils.MetaDefault
	}
	ldr, has := ldrS.ldrs[args.LoaderID]
	if !has {
		return fmt.Errorf("UNKNOWN_LOADER: %s", args.LoaderID)
	}
	if locked, err := ldr.isFolderLocked(); err != nil {
		return utils.NewErrServerError(err)
	} else if locked {
		if args.ForceLock {
			if err := ldr.unlockFolder(); err != nil {
				return utils.NewErrServerError(err)
			}
		}
		return errors.New("ANOTHER_LOADER_RUNNING")
	}
	//verify If Caching is present in arguments
	caching := config.CgrConfig().GeneralCfg().DefaultCaching
	if args.Caching != nil {
		caching = *args.Caching
	}
	if err := ldr.ProcessFolder(caching, utils.MetaRemove, args.StopOnError); err != nil {
		return utils.NewErrServerError(err)
	}
	*rply = utils.OK
	return
}

// Reload recreates the loaders map thread safe
func (ldrS *LoaderService) Reload(dm *engine.DataManager, ldrsCfg []*config.LoaderSCfg,
	timezone string, filterS *engine.FilterS, connMgr *engine.ConnManager) {
	ldrS.Lock()
	ldrS.ldrs = make(map[string]*Loader)
	for _, ldrCfg := range ldrsCfg {
		if ldrCfg.Enabled {
			ldrS.ldrs[ldrCfg.ID] = NewLoader(dm, ldrCfg, timezone, filterS, connMgr, ldrCfg.CacheSConns)
		}
	}
	ldrS.Unlock()
}
