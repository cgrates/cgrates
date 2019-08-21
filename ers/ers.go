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
	"path"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/fsnotify/fsnotify"
)

// NewERService instantiates the ERService
func NewERService(cfg *config.CGRConfig,
	filterS *engine.FilterS,
	sS rpcclient.RpcClientConnection,
	cfgRld chan struct{}) (erS *ERService, err error) {
	erS = &ERService{
		cfg:     cfg,
		rdrs:    make(map[string][]EventReader),
		stopLsn: make(map[string]chan struct{}),
		cfgRld:  cfgRld,
		sS:      sS,
	}
	return
}

// ERService is managing the EventReaders
type ERService struct {
	sync.RWMutex
	cfg     *config.CGRConfig
	filterS *engine.FilterS
	rdrs    map[string][]EventReader      // list of readers on specific paths map[path]reader
	stopLsn map[string]chan struct{}      // stops listening on paths
	cfgRld  chan struct{}                 // signal the need of config reloading - chan path / *any
	sS      rpcclient.RpcClientConnection // connection towards SessionS

}

// ListenAndServe loops keeps the service alive
func (erS *ERService) ListenAndServe(exitChan chan bool) (err error) {
	for _, rdrCfg := range erS.cfg.ERsCfg().Readers {
		var rdr EventReader
		if rdr, err = NewEventReader(rdrCfg); err != nil {
			return
		}
		if _, hasPath := erS.rdrs[rdrCfg.SourcePath]; !hasPath &&
			rdrCfg.Type == utils.MetaFileCSV &&
			rdrCfg.RunDelay == time.Duration(-1) { // set the channel to control listen stop
			erS.stopLsn[rdrCfg.SourcePath] = make(chan struct{})
			if err = erS.watchDir(rdrCfg.SourcePath); err != nil {
				return
			}
		}
		erS.rdrs[rdrCfg.SourcePath] = append(erS.rdrs[rdrCfg.SourcePath], rdr)

	}
	go erS.handleReloads()
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return
}

// erCfgRef will be used to reference a specific reader
type erCfgRef struct {
	path string
	idx  int
}

func (erS *ERService) handleReloads() {
	for {
		<-erS.cfgRld
		cfgIDs := make(map[string]int)         // IDs which are configured in EventReader profiles as map[id]cfgIdx
		inUseIDs := make(map[string]*erCfgRef) // IDs which are running in ERService map[id]rdrIdx
		addIDs := make(map[string]struct{})    // IDs which need to be added to ERService
		remIDs := make(map[string]struct{})    // IDs which need to be removed from ERService
		// index config IDs
		for i, rdrCfg := range erS.cfg.ERsCfg().Readers {
			cfgIDs[rdrCfg.ID] = i
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
			if len(rdrSlc) == 0 { // no more
				delete(erS.rdrs, ref.path)
				if chn, has := erS.stopLsn[ref.path]; has {
					close(chn)
				}
			}
		}
		// add new ids:
		for id := range addIDs {
			rdrCfg := erS.cfg.ERsCfg().Readers[cfgIDs[id]]
			if rdr, err := NewEventReader(rdrCfg); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> error reloading config with ID: <%s>, err: <%s>",
						utils.ERs, id, err.Error()))
			} else {
				if _, hasPath := erS.rdrs[rdrCfg.SourcePath]; !hasPath &&
					rdrCfg.Type == utils.MetaFileCSV &&
					rdrCfg.RunDelay == time.Duration(-1) { // set the channel to control listen stop
					erS.stopLsn[rdrCfg.SourcePath] = make(chan struct{})
					if err := erS.watchDir(rdrCfg.SourcePath); err != nil {
						utils.Logger.Warning(
							fmt.Sprintf(
								"<%s> error scheduling dir watch for config: <%s>, err: <%s>",
								utils.ERs, id, err.Error()))
					}
				}
				erS.rdrs[rdrCfg.SourcePath] = append(erS.rdrs[rdrCfg.SourcePath], rdr)
			}
		}
		erS.Unlock()
	}
}

// trackFiles
func (erS *ERService) watchDir(dirPath string) (err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()
	err = watcher.Add(dirPath)
	if err != nil {
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> monitoring <%s> for file moves.", utils.ERs, dirPath))
	stopLsnChan := erS.stopLsn[dirPath]
	for {
		select {
		case <-stopLsnChan: // stop listening
			utils.Logger.Info(fmt.Sprintf("<%s> stop listening on path <%s>", utils.ERs, dirPath))
			return
		case ev := <-watcher.Events:
			if ev.Op&fsnotify.Create == fsnotify.Create &&
				path.Ext(ev.Name) == utils.CSVSuffix {
				go func() { //Enable async processing here
					if err = erS.processPath(ev.Name); err != nil {
						utils.Logger.Err(fmt.Sprintf("<%s> processing path <%s>, error: <%s>",
							utils.ERs, ev.Name, err.Error()))
					}
				}()
			}
		case err := <-watcher.Errors:
			utils.Logger.Err(fmt.Sprintf("<%s> inotify error: <%s>", utils.ERs, err.Error()))
		}
	}
}

func (erS *ERService) processPath(path string) (err error) {
	return
}
