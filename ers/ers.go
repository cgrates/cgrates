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
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/fsnotify/fsnotify"
)

// NewERService instantiates the ERService
func NewERService(cfg *config.CGRConfig, filterS *engine.FilterS,
	sS rpcclient.RpcClientConnection, exitChan chan bool) (erS *ERService, err error) {
	erS = &ERService{
		cfg:      cfg,
		rdrs:     make(map[string][]EventReader),
		stopLsn:  make(map[string]chan struct{}),
		sS:       sS,
		exitChan: exitChan,
	}
	return
}

// ERService is managing the EventReaders
type ERService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	rdrs     map[string][]EventReader // list of readers on specific paths map[path]reader
	stopLsn  map[string]chan struct{} // stops listening on paths
	filterS  *engine.FilterS
	sS       rpcclient.RpcClientConnection // connection towards SessionS
	exitChan chan bool
}

// ListenAndServe loops keeps the service alive
func (erS *ERService) ListenAndServe(cfgRldChan chan struct{}) (err error) {
	var watchDirs []string
	for _, rdrCfg := range erS.cfg.ERsCfg().Readers {
		var rdr EventReader
		if rdr, err = NewEventReader(rdrCfg); err != nil {
			return
		}
		srcPath := rdrCfg.SourcePath
		if strings.HasSuffix(srcPath, utils.Slash) {
			srcPath = strings.TrimSuffix(srcPath, utils.Slash)
		}
		if _, hasPath := erS.rdrs[srcPath]; !hasPath &&
			rdrCfg.Type == utils.MetaFileCSV &&
			rdrCfg.RunDelay == time.Duration(-1) { // set the channel to control listen stop
			erS.stopLsn[srcPath] = make(chan struct{})
			watchDirs = append(watchDirs, srcPath)
		}
		erS.rdrs[srcPath] = append(erS.rdrs[srcPath], rdr)
	}
	go erS.handleReloads(cfgRldChan)
	erS.setDirWatchers(watchDirs)
	e := <-erS.exitChan
	erS.exitChan <- e // put back for the others listening for shutdown request
	return
}

// setDirWatchers sets up directory watchers
func (erS *ERService) setDirWatchers(dirPaths []string) {
	for _, dirPath := range dirPaths {
		go func() {
			if err := erS.watchDir(dirPath); err != nil {
				utils.Logger.Crit(
					fmt.Sprintf("<%s> watching directory <%s> got error: <%s>",
						utils.ERs, dirPath, err.Error()))
				erS.exitChan <- true
			}
		}()
	}
}

// erCfgRef will be used to reference a specific reader
type erCfgRef struct {
	path string
	idx  int
}

// handleReloads will handle the config reloads which are signaled over cfgRldChan
func (erS *ERService) handleReloads(cfgRldChan chan struct{}) {
	for {
		select {
		case <-erS.exitChan:
			return
		case <-cfgRldChan:
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
					inUseIDs[rdr.Config().ID] = &erCfgRef{path: path, idx: i}
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
			// remove the necessary ids
			for id := range remIDs {
				ref := inUseIDs[id]
				rdrSlc := erS.rdrs[ref.path]

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
			var watchDirs []string
			for id := range addIDs {
				rdrCfg := erS.cfg.ERsCfg().Readers[cfgIDs[id]]
				srcPath := rdrCfg.SourcePath
				if strings.HasSuffix(srcPath, utils.Slash) {
					srcPath = strings.TrimSuffix(srcPath, utils.Slash)
				}
				if rdr, err := NewEventReader(rdrCfg); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf(
							"<%s> error reloading config with ID: <%s>, err: <%s>",
							utils.ERs, id, err.Error()))
				} else {
					if _, hasPath := erS.rdrs[srcPath]; !hasPath &&
						rdrCfg.Type == utils.MetaFileCSV &&
						rdrCfg.RunDelay == time.Duration(-1) { // set the channel to control listen stop
						erS.stopLsn[srcPath] = make(chan struct{})
						watchDirs = append(watchDirs, srcPath)
					}
					erS.rdrs[srcPath] = append(erS.rdrs[srcPath], rdr)
				}
			}
			erS.setDirWatchers(watchDirs)
			erS.Unlock()
		}
	}
}

// watchDir sets up a watcher via inotify to be triggered on new files
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
					if err = erS.processPath(filepath.Dir(ev.Name), filepath.Base(ev.Name)); err != nil {
						utils.Logger.Warning(fmt.Sprintf("<%s> processing path <%s>, error: <%s>",
							utils.ERs, ev.Name, err.Error()))
					}
				}()
			}
		case err := <-watcher.Errors:
			utils.Logger.Err(fmt.Sprintf("<%s> inotify error: <%s>", utils.ERs, err.Error()))
		}
	}
}

// processPath will be called each time a new run should be triggered
func (erS *ERService) processPath(itmPath string, itmID string) error {
	rdrs, has := erS.rdrs[itmPath]
	if !has {
		return fmt.Errorf("no reader for path: <%s>", itmPath)
	}
	for _, rdr := range rdrs {
		rdrCfg := rdr.Config()
		if err := rdr.Init(itmPath, itmID); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> init reader <%s>, error: <%s>",
				utils.ERs, rdrCfg.ID, err.Error()))
			continue
		}
		for { // reads until no more events are produced
			cgrEv, err := rdr.Read()
			if err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> processing reader <%s>, error: <%s>",
						utils.ERs, rdrCfg.ID, err.Error()))
				continue
			} else if cgrEv == nil {
				break // no more events
			}
			// log the event created if requested by flags
			if rdrCfg.Flags.HasKey(utils.MetaLog) {
				utils.Logger.Info(
					fmt.Sprintf("<%s> LOG, reader: <%s>, message: %s",
						utils.ERs, rdrCfg.ID, utils.ToIJSON(cgrEv)))
			}
			// find out reqType
			var reqType string
			for _, typ := range []string{
				utils.MetaDryRun, utils.MetaAuth,
				utils.MetaInitiate, utils.MetaUpdate,
				utils.MetaTerminate, utils.MetaMessage,
				utils.MetaCDRs, utils.MetaEvent, utils.META_NONE} {
				if rdrCfg.Flags.HasKey(typ) { // request type is identified through flags
					reqType = typ
					break
				}
			}
			// execute the action based on reqType
			cgrArgs := cgrEv.ConsumeArgs(
				rdrCfg.Flags.HasKey(utils.MetaDispatchers),
				reqType == utils.MetaAuth ||
					reqType == utils.MetaMessage ||
					reqType == utils.MetaEvent)
			switch reqType {
			default:
				utils.Logger.Warning(
					fmt.Sprintf("<%s> processing reader <%s>, unsupported reqType: <%s>",
						utils.ERs, rdrCfg.ID, err.Error()))
				continue
			case utils.META_NONE: // do nothing on CGRateS side
			case utils.MetaDryRun:
				utils.Logger.Info(
					fmt.Sprintf("<%s> DRYRUN, reader: <%s>, CGREvent: <%s>",
						utils.DNSAgent, rdrCfg.ID, utils.ToJSON(cgrEv)))
			case utils.MetaAuth:
				authArgs := sessions.NewV1AuthorizeArgs(
					rdrCfg.Flags.HasKey(utils.MetaAttributes),
					rdrCfg.Flags.ParamsSlice(utils.MetaAttributes),
					rdrCfg.Flags.HasKey(utils.MetaThresholds),
					rdrCfg.Flags.ParamsSlice(utils.MetaThresholds),
					rdrCfg.Flags.HasKey(utils.MetaStats),
					rdrCfg.Flags.ParamsSlice(utils.MetaStats),
					rdrCfg.Flags.HasKey(utils.MetaResources),
					rdrCfg.Flags.HasKey(utils.MetaAccounts),
					rdrCfg.Flags.HasKey(utils.MetaSuppliers),
					rdrCfg.Flags.HasKey(utils.MetaSuppliersIgnoreErrors),
					rdrCfg.Flags.HasKey(utils.MetaSuppliersEventCost),
					cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator,
				)
				rply := new(sessions.V1AuthorizeReply)
				err = erS.sS.Call(utils.SessionSv1AuthorizeEvent,
					authArgs, rply)
			case utils.MetaInitiate:
				initArgs := sessions.NewV1InitSessionArgs(
					rdrCfg.Flags.HasKey(utils.MetaAttributes),
					rdrCfg.Flags.ParamsSlice(utils.MetaAttributes),
					rdrCfg.Flags.HasKey(utils.MetaThresholds),
					rdrCfg.Flags.ParamsSlice(utils.MetaThresholds),
					rdrCfg.Flags.HasKey(utils.MetaStats),
					rdrCfg.Flags.ParamsSlice(utils.MetaStats),
					rdrCfg.Flags.HasKey(utils.MetaResources),
					rdrCfg.Flags.HasKey(utils.MetaAccounts),
					cgrEv, cgrArgs.ArgDispatcher)
				rply := new(sessions.V1InitSessionReply)
				err = erS.sS.Call(utils.SessionSv1InitiateSession,
					initArgs, rply)
			case utils.MetaUpdate:
				updateArgs := sessions.NewV1UpdateSessionArgs(
					rdrCfg.Flags.HasKey(utils.MetaAttributes),
					rdrCfg.Flags.ParamsSlice(utils.MetaAttributes),
					rdrCfg.Flags.HasKey(utils.MetaAccounts),
					cgrEv, cgrArgs.ArgDispatcher)
				rply := new(sessions.V1UpdateSessionReply)
				err = erS.sS.Call(utils.SessionSv1UpdateSession,
					updateArgs, rply)
			case utils.MetaTerminate:
				terminateArgs := sessions.NewV1TerminateSessionArgs(
					rdrCfg.Flags.HasKey(utils.MetaAccounts),
					rdrCfg.Flags.HasKey(utils.MetaResources),
					rdrCfg.Flags.HasKey(utils.MetaThresholds),
					rdrCfg.Flags.ParamsSlice(utils.MetaThresholds),
					rdrCfg.Flags.HasKey(utils.MetaStats),
					rdrCfg.Flags.ParamsSlice(utils.MetaStats),
					cgrEv, cgrArgs.ArgDispatcher)
				rply := utils.StringPointer("")
				err = erS.sS.Call(utils.SessionSv1TerminateSession,
					terminateArgs, rply)
			case utils.MetaMessage:
				evArgs := sessions.NewV1ProcessMessageArgs(
					rdrCfg.Flags.HasKey(utils.MetaAttributes),
					rdrCfg.Flags.ParamsSlice(utils.MetaAttributes),
					rdrCfg.Flags.HasKey(utils.MetaThresholds),
					rdrCfg.Flags.ParamsSlice(utils.MetaThresholds),
					rdrCfg.Flags.HasKey(utils.MetaStats),
					rdrCfg.Flags.ParamsSlice(utils.MetaStats),
					rdrCfg.Flags.HasKey(utils.MetaResources),
					rdrCfg.Flags.HasKey(utils.MetaAccounts),
					rdrCfg.Flags.HasKey(utils.MetaSuppliers),
					rdrCfg.Flags.HasKey(utils.MetaSuppliersIgnoreErrors),
					rdrCfg.Flags.HasKey(utils.MetaSuppliersEventCost),
					cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator)
				rply := new(sessions.V1ProcessMessageReply) // need it so rpcclient can clone
				err = erS.sS.Call(utils.SessionSv1ProcessMessage,
					evArgs, rply)
				if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
					cgrEv.Event[utils.Usage] = 0 // avoid further debits
				} else if rply.MaxUsage != nil {
					cgrEv.Event[utils.Usage] = *rply.MaxUsage // make sure the CDR reflects the debit
				}
			case utils.MetaEvent:
				evArgs := &sessions.V1ProcessEventArgs{
					Flags:         rdrCfg.Flags.SliceFlags(),
					CGREvent:      cgrEv,
					ArgDispatcher: cgrArgs.ArgDispatcher,
					Paginator:     *cgrArgs.SupplierPaginator,
				}
				rply := new(sessions.V1ProcessEventReply)
				err = erS.sS.Call(utils.SessionSv1ProcessEvent,
					evArgs, rply)
			case utils.MetaCDRs: // allow CDR processing
			}
			if err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> reader <%s>, error: <%s> posting event",
						utils.ERs, rdrCfg.ID, err.Error()))
			}
			// separate request so we can capture the Terminate/Event also here
			if rdrCfg.Flags.HasKey(utils.MetaCDRs) &&
				!rdrCfg.Flags.HasKey(utils.MetaDryRun) {
				rplyCDRs := utils.StringPointer("")
				if err = erS.sS.Call(utils.SessionSv1ProcessCDR,
					&utils.CGREventWithArgDispatcher{CGREvent: cgrEv,
						ArgDispatcher: cgrArgs.ArgDispatcher}, &rplyCDRs); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> reader <%s>, error: <%s> posting event",
							utils.ERs, rdrCfg.ID, err.Error()))
				}
			}

		}
		if err := rdr.Close(); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> closing reader <%s>, error: <%s>",
				utils.ERs, rdr.Config().ID, err.Error()))
		}
	}
	return nil
}
