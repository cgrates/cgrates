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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewLoaderService(dm *engine.DataManager, ldrsCfg []*config.LoaderSCfg,
	timezone string, filterS *engine.FilterS) (ldrS *LoaderService) {
	ldrS = &LoaderService{ldrs: make(map[string]*Loader)}
	for _, ldrCfg := range ldrsCfg {
		if !ldrCfg.Enabled {
			continue
		}
		ldrS.ldrs[ldrCfg.Id] = NewLoader(dm, ldrCfg, timezone, filterS)
	}
	return
}

// LoaderS is the Loader service handling independent Loaders
type LoaderService struct {
	ldrs map[string]*Loader
}

// IsEnabled returns true if at least one loader is enabled
func (ldrS *LoaderService) Enabled() bool {
	for _, ldr := range ldrS.ldrs {
		if ldr.enabled {
			return true
		}
	}
	return false
}

func (ldrS *LoaderService) ListenAndServe(exitChan chan bool) (err error) {
	ldrExitChan := make(chan struct{})
	for _, ldr := range ldrS.ldrs {
		go ldr.ListenAndServe(ldrExitChan)
	}
	select { // exit on errors coming from server or any loader
	case e := <-exitChan:
		close(ldrExitChan)
		exitChan <- e // put back for the others listening for shutdown request
	case <-ldrExitChan:
		exitChan <- true
	}
	return
}

type ArgsProcessFolder struct {
	LoaderID  string
	ForceLock bool
}

func (ldrS *LoaderService) V1Load(args *ArgsProcessFolder,
	rply *string) (err error) {
	if args.LoaderID == "" {
		args.LoaderID = utils.META_DEFAULT
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
	if err := ldr.ProcessFolder(); err != nil {
		return utils.NewErrServerError(err)
	}
	*rply = utils.OK
	return
}
