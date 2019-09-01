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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewERService instantiates the ERService
func NewERService(cfg *config.CGRConfig, filterS *engine.FilterS,
	sS rpcclient.RpcClientConnection, exitChan chan bool) (erS *ERService, err error) {
	erS = &ERService{
		cfg:      cfg,
		rdrs:     make(map[string]EventReader),
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
	rdrs     map[string]EventReader   // map[rdrID]EventReader
	stopLsn  map[string]chan struct{} // map[rdrID] chan struct{}
	filterS  *engine.FilterS
	sS       rpcclient.RpcClientConnection // connection towards SessionS
	exitChan chan bool
}

// ListenAndServe keeps the service alive
func (erS *ERService) ListenAndServe(cfgRldChan chan struct{}) (err error) {
	for _, rdrCfg := range erS.cfg.ERsCfg().Readers {
		if err = erS.addReader(rdrCfg); err != nil {
			utils.Logger.Crit(
				fmt.Sprintf("<%s> adding reader <%s> got error: <%s>",
					utils.ERs, rdrCfg.ID, err.Error()))
			return
		}
	}
	go erS.handleReloads(cfgRldChan)
	e := <-erS.exitChan
	erS.exitChan <- e // put back for the others listening for shutdown request
	return
}

// addReader will add a new reader to the service
func (erS *ERService) addReader(rdrCfg *config.EventReaderCfg) (err error) {
	var rdr EventReader
	if rdr, err = NewEventReader(rdrCfg); err != nil {
		return
	}
	erS.rdrs[rdrCfg.ID] = rdr
	erS.stopLsn[rdrCfg.ID] = make(chan struct{})
	return rdr.Subscribe()
}

// setDirWatchers sets up directory watchers
/*func (erS *ERService) setDirWatchers(dirPaths []string) {
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
*/

// handleReloads will handle the config reloads which are signaled over cfgRldChan
func (erS *ERService) handleReloads(cfgRldChan chan struct{}) {
	for {
		select {
		case <-erS.exitChan:
			return
		case <-cfgRldChan:
			cfgIDs := make(map[string]*config.EventReaderCfg)
			// index config IDs
			for _, rdrCfg := range erS.cfg.ERsCfg().Readers {
				cfgIDs[rdrCfg.ID] = rdrCfg
			}
			erS.Lock()
			// remove the necessary ids
			for id := range erS.rdrs {
				if _, has := cfgIDs[id]; has { // still present
					continue
				}
				delete(erS.rdrs, id)
				close(erS.stopLsn[id])
				delete(erS.stopLsn, id)
			}
			// add new ids
			for id, rdrCfg := range cfgIDs {
				if _, has := erS.rdrs[id]; has {
					continue
				}
				if err := erS.addReader(rdrCfg); err != nil {
					utils.Logger.Crit(
						fmt.Sprintf("<%s> adding reader <%s> got error: <%s>",
							utils.ERs, rdrCfg.ID, err.Error()))
					erS.exitChan <- true
				}
			}
			erS.Unlock()
		}
	}
}

/*
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
*/
