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

package ers

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewERService instantiates the ERService
func NewERService(cfg *config.CGRConfig, cdrS rpcclient.RpcClientConnection,
	cfgRld chan struct{}) (erS *ERService, err error) {
	erS = &ERService{
		rdrs:   make(map[string][]EventReader),
		cfgRld: cfgRld,
	}
	for _, rdrCfg := range cfg.ERsCfg().Readers {
		if rdr, err := NewEventReader(rdrCfg); err != nil {
			return nil, err
		} else {
			erS.rdrs[rdrCfg.SourcePath] = append(erS.rdrs[rdrCfg.SourcePath], rdr)
		}
	}
	return
}

// ERService is managing the EventReaders
type ERService struct {
	sync.RWMutex
	cfg    *config.CGRConfig
	rdrs   map[string][]EventReader // list of readers on specific paths map[path]reader
	cfgRld chan struct{}            // signal the need of config reloading - chan path / *any
	sS     rpcclient.RpcClientConnection
}

// ListenAndServe loops keeps the service alive
func (erS *ERService) ListenAndServe(exitChan chan bool) error {
	go erS.handleReloads() // start backup loop
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// erCfgRef will be used to reference a specific reader
type erCfgRef struct {
	path string
	idx  int
}

func (erS *ERService) handleReloads() {
	for {
		<-erS.cfgRld
		cfgIDs := make(map[string]*erCfgRef)   // IDs which are configured in EventReader profiles
		inUseIDs := make(map[string]*erCfgRef) // IDs which are running in ERService indexed on path
		addIDs := make(map[string]struct{})    // IDs which need to be added to ERService
		remIDs := make(map[string]struct{})    // IDs which need to be removed from ERService
		// index config IDs
		for i, rdrCfg := range erS.cfg.ERsCfg().Readers {
			cfgIDs[rdrCfg.ID] = &erCfgRef{path: rdrCfg.SourcePath, idx: i}
		}
		erS.Lock()
		// index in use IDs
		for path, rdrs := range erS.rdrs {
			for i, rdr := range rdrs {
				inUseIDs[rdr.ID()] = &erCfgRef{path: path, idx: i}
			}
		}
		// find out removed ids
		for id := range inUseIDs {
			if _, has := cfgIDs[id]; !has {
				remIDs[id] = struct{}{}
			}
		}
		// find out added ids
		for id := range cfgIDs {
			if _, has := inUseIDs[id]; !has {
				addIDs[id] = struct{}{}
			}
		}
		for id := range remIDs {
			ref := inUseIDs[id]
			rdrSlc := erS.rdrs[ref.path]
			// remove the ids
			copy(rdrSlc[ref.idx:], rdrSlc[ref.idx+1:])
			rdrSlc[len(rdrSlc)-1] = nil // so it can be garbage collected
			rdrSlc = rdrSlc[:len(rdrSlc)-1]
		}
		// add new ids:
		for id := range addIDs {
			cfgRef := cfgIDs[id]
			if newRdr, err := NewEventReader(erS.cfg.ERsCfg().Readers[cfgRef.idx]); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> error reloading config with ID: <%s>, err: <%s>",
						utils.ERs, id, err.Error()))
			} else {
				erS.rdrs[cfgRef.path] = append(erS.rdrs[cfgRef.path], newRdr)
			}

		}
		erS.Unlock()
	}
}
