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
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// erEvent is passed from reader to ERs
type erEvent struct {
	cgrEvent *utils.CGREvent
	rdrCfg   *config.EventReaderCfg
	opts     map[string]interface{}
}

// NewERService instantiates the ERService
func NewERService(cfg *config.CGRConfig, filterS *engine.FilterS, connMgr *engine.ConnManager) *ERService {
	return &ERService{
		cfg:       cfg,
		rdrs:      make(map[string]EventReader),
		rdrPaths:  make(map[string]string),
		stopLsn:   make(map[string]chan struct{}),
		rdrEvents: make(chan *erEvent),
		rdrErr:    make(chan error),
		filterS:   filterS,
		connMgr:   connMgr,
	}
}

// ERService is managing the EventReaders
type ERService struct {
	sync.RWMutex
	cfg       *config.CGRConfig
	rdrs      map[string]EventReader   // map[rdrID]EventReader
	rdrPaths  map[string]string        // used for reloads in case of path changes
	stopLsn   map[string]chan struct{} // map[rdrID] chan struct{}
	rdrEvents chan *erEvent            // receive here the events from readers
	rdrErr    chan error               // receive here errors which should stop the app

	filterS *engine.FilterS
	connMgr *engine.ConnManager
}

// ListenAndServe keeps the service alive
func (erS *ERService) ListenAndServe(stopChan, cfgRldChan chan struct{}) (err error) {
	for cfgIdx, rdrCfg := range erS.cfg.ERsCfg().Readers {
		if rdrCfg.Type == utils.MetaNone { // ignore *default reader
			continue
		}
		if err = erS.addReader(rdrCfg.ID, cfgIdx); err != nil {
			utils.Logger.Crit(
				fmt.Sprintf("<%s> adding reader <%s> got error: <%s>",
					utils.ERs, rdrCfg.ID, err.Error()))
			return
		}
	}
	for {
		select {
		case err = <-erS.rdrErr: // got application error
			erS.closeAllRdrs()
			utils.Logger.Crit(
				fmt.Sprintf("<%s> running reader got error: <%s>",
					utils.ERs, err.Error()))
			return
		case <-stopChan:
			erS.closeAllRdrs()
			return
		case erEv := <-erS.rdrEvents:
			if err := erS.processEvent(erEv.cgrEvent, erEv.rdrCfg, erEv.opts); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> reading event: <%s> got error: <%s>",
						utils.ERs, utils.ToIJSON(erEv.cgrEvent), err.Error()))
			}
		case <-cfgRldChan: // handle reload
			cfgIDs := make(map[string]int)
			pathReloaded := make(utils.StringSet)
			// index config IDs
			for i, rdrCfg := range erS.cfg.ERsCfg().Readers {
				cfgIDs[rdrCfg.ID] = i
			}
			erS.Lock()
			// remove the necessary ids
			for id, rdr := range erS.rdrs {
				if cfgIdx, has := cfgIDs[id]; has { // still present
					newCfg := erS.cfg.ERsCfg().Readers[cfgIdx]
					if newCfg.SourcePath == erS.rdrPaths[id] &&
						newCfg.ID == rdr.Config().ID { // make sure the index did not change
						continue
					}
					pathReloaded.Add(id)
				}
				delete(erS.rdrs, id)
				close(erS.stopLsn[id])
				delete(erS.stopLsn, id)
			}
			// add new ids
			for id, rdrIdx := range cfgIDs {
				if _, has := erS.rdrs[id]; has &&
					!pathReloaded.Has(id) {
					continue
				}
				if erS.cfg.ERsCfg().Readers[rdrIdx].Type == utils.MetaNone { // ignore *default reader
					continue
				}
				if err := erS.addReader(id, rdrIdx); err != nil {
					utils.Logger.Crit(
						fmt.Sprintf("<%s> adding reader <%s> got error: <%s>",
							utils.ERs, id, err.Error()))
					erS.rdrErr <- err
				}
			}
			erS.Unlock()
		}
	}
}

// addReader will add a new reader to the service
func (erS *ERService) addReader(rdrID string, cfgIdx int) (err error) {
	erS.stopLsn[rdrID] = make(chan struct{})
	var rdr EventReader
	if rdr, err = NewEventReader(erS.cfg, cfgIdx,
		erS.rdrEvents, erS.rdrErr,
		erS.filterS, erS.stopLsn[rdrID]); err != nil {
		return
	}
	erS.rdrs[rdrID] = rdr
	return rdr.Serve()
}

// processEvent will be called each time a new event is received from readers
func (erS *ERService) processEvent(cgrEv *utils.CGREvent,
	rdrCfg *config.EventReaderCfg, opts map[string]interface{}) (err error) {
	// log the event created if requested by flags
	if rdrCfg.Flags.Has(utils.MetaLog) {
		utils.Logger.Info(
			fmt.Sprintf("<%s> LOG, reader: <%s>, message: %s",
				utils.ERs, rdrCfg.ID, utils.ToIJSON(cgrEv)))
	}
	// find out reqType
	var reqType string
	for _, typ := range []string{
		utils.MetaDryRun, utils.MetaAuthorize,
		utils.MetaInitiate, utils.MetaUpdate,
		utils.MetaTerminate, utils.MetaMessage,
		utils.MetaCDRs, utils.MetaEvent, utils.MetaNone} {
		if rdrCfg.Flags.Has(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}
	var cgrArgs utils.Paginator
	if reqType == utils.MetaAuthorize ||
		reqType == utils.MetaMessage ||
		reqType == utils.MetaEvent {
		if cgrArgs, err = utils.GetRoutePaginatorFromOpts(opts); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> args extraction for reader <%s> failed because <%s>",
				utils.ERs, rdrCfg.ID, err.Error()))
			err = nil // reset the error and continue the processing
		}
	}
	// execute the action based on reqType
	switch reqType {
	default:
		return fmt.Errorf("unsupported reqType: <%s>", reqType)
	case utils.MetaNone: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRYRUN, reader: <%s>, CGREvent: <%s>",
				utils.ERs, rdrCfg.ID, utils.ToJSON(cgrEv)))
	case utils.MetaAuthorize:
		authArgs := sessions.NewV1AuthorizeArgs(
			rdrCfg.Flags.Has(utils.MetaAttributes),
			rdrCfg.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaThresholds),
			rdrCfg.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaStats),
			rdrCfg.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaResources),
			rdrCfg.Flags.Has(utils.MetaAccounts),
			rdrCfg.Flags.Has(utils.MetaRoutes),
			rdrCfg.Flags.Has(utils.MetaRoutesIgnoreErrors),
			rdrCfg.Flags.Has(utils.MetaRoutesEventCost),
			cgrEv, cgrArgs,
			rdrCfg.Flags.Has(utils.MetaFD),
			opts,
		)
		rply := new(sessions.V1AuthorizeReply)
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1AuthorizeEvent,
			authArgs, rply)
	case utils.MetaInitiate:
		initArgs := sessions.NewV1InitSessionArgs(
			rdrCfg.Flags.Has(utils.MetaAttributes),
			rdrCfg.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaThresholds),
			rdrCfg.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaStats),
			rdrCfg.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaResources),
			rdrCfg.Flags.Has(utils.MetaAccounts),
			cgrEv, rdrCfg.Flags.Has(utils.MetaFD),
			opts)
		rply := new(sessions.V1InitSessionReply)
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1InitiateSession,
			initArgs, rply)
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			rdrCfg.Flags.Has(utils.MetaAttributes),
			rdrCfg.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaAccounts),
			cgrEv, rdrCfg.Flags.Has(utils.MetaFD),
			opts)
		rply := new(sessions.V1UpdateSessionReply)
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1UpdateSession,
			updateArgs, rply)
	case utils.MetaTerminate:
		terminateArgs := sessions.NewV1TerminateSessionArgs(
			rdrCfg.Flags.Has(utils.MetaAccounts),
			rdrCfg.Flags.Has(utils.MetaResources),
			rdrCfg.Flags.Has(utils.MetaThresholds),
			rdrCfg.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaStats),
			rdrCfg.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			cgrEv, rdrCfg.Flags.Has(utils.MetaFD),
			opts)
		rply := utils.StringPointer("")
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1TerminateSession,
			terminateArgs, rply)
	case utils.MetaMessage:
		evArgs := sessions.NewV1ProcessMessageArgs(
			rdrCfg.Flags.Has(utils.MetaAttributes),
			rdrCfg.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaThresholds),
			rdrCfg.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaStats),
			rdrCfg.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaResources),
			rdrCfg.Flags.Has(utils.MetaAccounts),
			rdrCfg.Flags.Has(utils.MetaRoutes),
			rdrCfg.Flags.Has(utils.MetaRoutesIgnoreErrors),
			rdrCfg.Flags.Has(utils.MetaRoutesEventCost),
			cgrEv, cgrArgs,
			rdrCfg.Flags.Has(utils.MetaFD),
			opts)
		rply := new(sessions.V1ProcessMessageReply) // need it so rpcclient can clone
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1ProcessMessage,
			evArgs, rply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if evArgs.Debit {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
	case utils.MetaEvent:
		evArgs := &sessions.V1ProcessEventArgs{
			Flags: rdrCfg.Flags.SliceFlags(),
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: cgrEv,
				Opts:     opts,
			},
			Paginator: cgrArgs,
		}
		rply := new(sessions.V1ProcessEventReply)
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1ProcessEvent,
			evArgs, rply)
	case utils.MetaCDRs: // allow CDR processing
	}
	if err != nil {
		return
	}
	// separate request so we can capture the Terminate/Event also here
	if rdrCfg.Flags.Has(utils.MetaCDRs) &&
		!rdrCfg.Flags.Has(utils.MetaDryRun) {
		rplyCDRs := utils.StringPointer("")
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1ProcessCDR,
			&utils.CGREventWithOpts{
				CGREvent: cgrEv,
				Opts:     opts,
			}, rplyCDRs)
	}

	return
}

func (erS *ERService) closeAllRdrs() {
	for _, stopL := range erS.stopLsn {
		close(stopL)
	}
}
