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
}

// NewERService instantiates the ERService
func NewERService(cfg *config.CGRConfig, filterS *engine.FilterS, stopChan chan struct{}, connMgr *engine.ConnManager) *ERService {
	return &ERService{
		cfg:       cfg,
		rdrs:      make(map[string]EventReader),
		rdrPaths:  make(map[string]string),
		stopLsn:   make(map[string]chan struct{}),
		rdrEvents: make(chan *erEvent),
		rdrErr:    make(chan error),
		filterS:   filterS,
		stopChan:  stopChan,
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

	filterS  *engine.FilterS
	stopChan chan struct{}
	connMgr  *engine.ConnManager
}

// ListenAndServe keeps the service alive
func (erS *ERService) ListenAndServe(cfgRldChan chan struct{}) (err error) {
	for cfgIdx, rdrCfg := range erS.cfg.ERsCfg().Readers {
		if rdrCfg.Type == utils.META_NONE { // ignore *default reader
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
			utils.Logger.Crit(
				fmt.Sprintf("<%s> running reader got error: <%s>",
					utils.ERs, err.Error()))
			return
		case <-erS.stopChan:
			return
		case erEv := <-erS.rdrEvents:
			if err := erS.processEvent(erEv.cgrEvent, erEv.rdrCfg); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> reading event: <%s> got error: <%s>",
						utils.ERs, utils.ToIJSON(erEv.cgrEvent), err.Error()))
			}
		case <-cfgRldChan: // handle reload
			cfgIDs := make(map[string]int)
			pathReloaded := make(map[string]struct{})
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
					pathReloaded[id] = struct{}{}
				}
				delete(erS.rdrs, id)
				close(erS.stopLsn[id])
				delete(erS.stopLsn, id)
			}
			// add new ids
			for id, rdrIdx := range cfgIDs {
				if _, has := erS.rdrs[id]; has {
					if _, has := pathReloaded[id]; !has {
						continue
					}
				}
				if erS.cfg.ERsCfg().Readers[rdrIdx].Type == utils.META_NONE {
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
func (erS *ERService) processEvent(cgrEv *utils.CGREvent, rdrCfg *config.EventReaderCfg) (err error) {
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
	cgrArgs := cgrEv.ExtractArgs(
		rdrCfg.Flags.HasKey(utils.MetaDispatchers),
		reqType == utils.MetaAuth ||
			reqType == utils.MetaMessage ||
			reqType == utils.MetaEvent)
	switch reqType {
	default:
		return fmt.Errorf("unsupported reqType: <%s>", reqType)
	case utils.META_NONE: // do nothing on CGRateS side
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRYRUN, reader: <%s>, CGREvent: <%s>",
				utils.ERs, rdrCfg.ID, utils.ToJSON(cgrEv)))
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
			rdrCfg.Flags.HasKey(utils.MetaFD),
			rdrCfg.Flags.ParamValue(utils.MetaSuppliersMaxCost),
		)
		rply := new(sessions.V1AuthorizeReply)
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1AuthorizeEvent,
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
			cgrEv, cgrArgs.ArgDispatcher,
			rdrCfg.Flags.HasKey(utils.MetaFD))
		rply := new(sessions.V1InitSessionReply)
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1InitiateSession,
			initArgs, rply)
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			rdrCfg.Flags.HasKey(utils.MetaAttributes),
			rdrCfg.Flags.ParamsSlice(utils.MetaAttributes),
			rdrCfg.Flags.HasKey(utils.MetaAccounts),
			cgrEv, cgrArgs.ArgDispatcher,
			rdrCfg.Flags.HasKey(utils.MetaFD))
		rply := new(sessions.V1UpdateSessionReply)
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1UpdateSession,
			updateArgs, rply)
	case utils.MetaTerminate:
		terminateArgs := sessions.NewV1TerminateSessionArgs(
			rdrCfg.Flags.HasKey(utils.MetaAccounts),
			rdrCfg.Flags.HasKey(utils.MetaResources),
			rdrCfg.Flags.HasKey(utils.MetaThresholds),
			rdrCfg.Flags.ParamsSlice(utils.MetaThresholds),
			rdrCfg.Flags.HasKey(utils.MetaStats),
			rdrCfg.Flags.ParamsSlice(utils.MetaStats),
			cgrEv, cgrArgs.ArgDispatcher,
			rdrCfg.Flags.HasKey(utils.MetaFD))
		rply := utils.StringPointer("")
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1TerminateSession,
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
			cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator,
			rdrCfg.Flags.HasKey(utils.MetaFD),
			rdrCfg.Flags.ParamValue(utils.MetaSuppliersMaxCost))
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
			Flags:         rdrCfg.Flags.SliceFlags(),
			CGREvent:      cgrEv,
			ArgDispatcher: cgrArgs.ArgDispatcher,
			Paginator:     *cgrArgs.SupplierPaginator,
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
	if rdrCfg.Flags.HasKey(utils.MetaCDRs) &&
		!rdrCfg.Flags.HasKey(utils.MetaDryRun) {
		rplyCDRs := utils.StringPointer("")
		err = erS.connMgr.Call(erS.cfg.ERsCfg().SessionSConns, nil, utils.SessionSv1ProcessCDR,
			&utils.CGREventWithArgDispatcher{CGREvent: cgrEv,
				ArgDispatcher: cgrArgs.ArgDispatcher}, rplyCDRs)
	}

	return
}
