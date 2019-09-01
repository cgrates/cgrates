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
	"path/filepath"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/fsnotify/fsnotify"
)

// watchDir sets up a watcher via inotify to be triggered on new files
// sysID is the subsystem ID, f will be triggered on match
func watchDir(dirPath string, f func(itmPath, itmID string) error,
	sysID string, stopWatching chan struct{}) (err error) {
	var watcher *fsnotify.Watcher
	if watcher, err = fsnotify.NewWatcher(); err != nil {
		return
	}
	defer watcher.Close()
	if err = watcher.Add(dirPath); err != nil {
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> monitoring <%s> for file moves.", sysID, dirPath))
	for {
		select {
		case <-stopWatching:
			utils.Logger.Info(fmt.Sprintf("<%s> stop watching path <%s>", sysID, dirPath))
			return
		case ev := <-watcher.Events:
			if ev.Op&fsnotify.Create == fsnotify.Create {
				go func() { //Enable async processing here
					if err = f(filepath.Dir(ev.Name), filepath.Base(ev.Name)); err != nil {
						utils.Logger.Warning(fmt.Sprintf("<%s> processing path <%s>, error: <%s>",
							sysID, ev.Name, err.Error()))
					}
				}()
			}
		case err = <-watcher.Errors:
			return
		}
	}
}

// processReader will process events from reader and publish them to SessionS
func processReader(rdr EventReader, sS rpcclient.RpcClient, rdrExit chan struct{}) (err error) {
	for { // reads until no more events are produced or exit is signaled
		select {
		case <-rdrExit:
			return
		default:
		}
		rdrCfg := rdr.Config()
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
			err = sS.Call(utils.SessionSv1AuthorizeEvent,
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
			err = sS.Call(utils.SessionSv1InitiateSession,
				initArgs, rply)
		case utils.MetaUpdate:
			updateArgs := sessions.NewV1UpdateSessionArgs(
				rdrCfg.Flags.HasKey(utils.MetaAttributes),
				rdrCfg.Flags.ParamsSlice(utils.MetaAttributes),
				rdrCfg.Flags.HasKey(utils.MetaAccounts),
				cgrEv, cgrArgs.ArgDispatcher)
			rply := new(sessions.V1UpdateSessionReply)
			err = sS.Call(utils.SessionSv1UpdateSession,
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
			err = sS.Call(utils.SessionSv1TerminateSession,
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
			err = sS.Call(utils.SessionSv1ProcessMessage,
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
			err = sS.Call(utils.SessionSv1ProcessEvent,
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
			if err = sS.Call(utils.SessionSv1ProcessCDR,
				&utils.CGREventWithArgDispatcher{CGREvent: cgrEv,
					ArgDispatcher: cgrArgs.ArgDispatcher}, &rplyCDRs); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> reader <%s>, error: <%s> posting event",
						utils.ERs, rdrCfg.ID, err.Error()))
			}
		}
	}
	return
}
