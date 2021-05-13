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
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// erEvent is passed from reader to ERs
type erEvent struct {
	cgrEvent *utils.CGREvent
	rdrCfg   *config.EventReaderCfg
}

// NewERService instantiates the ERService
func NewERService(cfg *config.CGRConfig, filterS *engine.FilterS, connMgr *engine.ConnManager) (ers *ERService) {
	ers = &ERService{
		cfg:           cfg,
		rdrs:          make(map[string]EventReader),
		rdrPaths:      make(map[string]string),
		stopLsn:       make(map[string]chan struct{}),
		rdrEvents:     make(chan *erEvent),
		partialEvents: make(chan *erEvent),
		rdrErr:        make(chan error),
		filterS:       filterS,
		connMgr:       connMgr,
	}
	ers.partialCache = ltcache.NewCache(ltcache.UnlimitedCaching, cfg.ERsCfg().PartialCacheTTL, false, ers.onEvicted)
	return
}

// ERService is managing the EventReaders
type ERService struct {
	sync.RWMutex
	cfg           *config.CGRConfig
	rdrs          map[string]EventReader   // map[rdrID]EventReader
	rdrPaths      map[string]string        // used for reloads in case of path changes
	stopLsn       map[string]chan struct{} // map[rdrID] chan struct{}
	rdrEvents     chan *erEvent            // receive here the events from readers
	partialEvents chan *erEvent            // receive here the partial events from readers
	rdrErr        chan error               // receive here errors which should stop the app

	filterS *engine.FilterS
	connMgr *engine.ConnManager

	partialCache *ltcache.Cache
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
			if err := erS.processEvent(erEv.cgrEvent, erEv.rdrCfg); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> reading event: <%s> from reader: <%s> got error: <%s>",
						utils.ERs, utils.ToJSON(erEv.cgrEvent), erEv.rdrCfg.ID, err.Error()))
			}
		case pEv := <-erS.partialEvents:
			if err := erS.processPartialEvent(pEv.cgrEvent, pEv.rdrCfg); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> reading partial event: <%s> from reader: <%s> got error: <%s>",
						utils.ERs, utils.ToJSON(pEv.cgrEvent), pEv.rdrCfg.ID, err.Error()))
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
				if err = erS.addReader(id, rdrIdx); err != nil {
					utils.Logger.Crit(
						fmt.Sprintf("<%s> adding reader <%s> got error: <%s>",
							utils.ERs, id, err.Error()))
					erS.closeAllRdrs()
					erS.Unlock()
					return
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
		erS.rdrEvents, erS.partialEvents, erS.rdrErr,
		erS.filterS, erS.stopLsn[rdrID]); err != nil {
		return
	}
	erS.rdrs[rdrID] = rdr
	return rdr.Serve()
}

// processEvent will be called each time a new event is received from readers
func (erS *ERService) processEvent(cgrEv *utils.CGREvent,
	rdrCfg *config.EventReaderCfg) (err error) {
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
		if cgrArgs, err = utils.GetRoutePaginatorFromOpts(cgrEv.APIOpts); err != nil {
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
			rdrCfg.Flags.ParamValue(utils.MetaRoutesMaxCost),
		)
		rply := new(sessions.V1AuthorizeReply)
		err = erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().SessionSConns, utils.SessionSv1AuthorizeEvent,
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
			cgrEv, rdrCfg.Flags.Has(utils.MetaFD))
		rply := new(sessions.V1InitSessionReply)
		err = erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().SessionSConns, utils.SessionSv1InitiateSession,
			initArgs, rply)
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			rdrCfg.Flags.Has(utils.MetaAttributes),
			rdrCfg.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaAccounts),
			cgrEv, rdrCfg.Flags.Has(utils.MetaFD))
		rply := new(sessions.V1UpdateSessionReply)
		err = erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().SessionSConns, utils.SessionSv1UpdateSession,
			updateArgs, rply)
	case utils.MetaTerminate:
		terminateArgs := sessions.NewV1TerminateSessionArgs(
			rdrCfg.Flags.Has(utils.MetaAccounts),
			rdrCfg.Flags.Has(utils.MetaResources),
			rdrCfg.Flags.Has(utils.MetaThresholds),
			rdrCfg.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaStats),
			rdrCfg.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			cgrEv, rdrCfg.Flags.Has(utils.MetaFD))
		rply := utils.StringPointer("")
		err = erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().SessionSConns, utils.SessionSv1TerminateSession,
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
			rdrCfg.Flags.ParamValue(utils.MetaRoutesMaxCost),
		)
		rply := new(sessions.V1ProcessMessageReply) // need it so rpcclient can clone
		err = erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().SessionSConns, utils.SessionSv1ProcessMessage,
			evArgs, rply)
		// if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
		// cgrEv.Event[utils.Usage] = 0 // avoid further debits
		// } else
		if evArgs.Debit {
			cgrEv.Event[utils.Usage] = rply.MaxUsage // make sure the CDR reflects the debit
		}
	case utils.MetaEvent:
		evArgs := &sessions.V1ProcessEventArgs{
			Flags:     rdrCfg.Flags.SliceFlags(),
			CGREvent:  cgrEv,
			Paginator: cgrArgs,
		}
		rply := new(sessions.V1ProcessEventReply)
		err = erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().SessionSConns, utils.SessionSv1ProcessEvent,
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
		err = erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().SessionSConns, utils.SessionSv1ProcessCDR,
			cgrEv, rplyCDRs)
	}

	return
}

func (erS *ERService) closeAllRdrs() {
	for _, stopL := range erS.stopLsn {
		close(stopL)
	}
}

const (
	partialOpt = "*partial"
)

type erEvents struct {
	events []*utils.CGREvent
	rdrCfg *config.EventReaderCfg
}

func (erS *ERService) processPartialEvent(ev *utils.CGREvent, rdrCfg *config.EventReaderCfg) (err error) {
	orgID, err := ev.FieldAsString(utils.OriginID)
	if err == utils.ErrNotFound {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> Missing <OriginID> field for partial event <%s>",
				utils.ERs, utils.ToJSON(ev)))
		return
	}
	orgHost, err := ev.FieldAsString(utils.OriginHost)
	if err == utils.ErrNotFound {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> Missing <OriginHost> field for partial event <%s>",
				utils.ERs, utils.ToJSON(ev)))
		return
	}
	cgrID := utils.Sha1(orgID, orgHost)

	evs, has := erS.partialCache.Get(cgrID)
	var cgrEvs *erEvents
	if !has || evs == nil {
		cgrEvs = &erEvents{
			events: []*utils.CGREvent{ev},
			rdrCfg: rdrCfg,
		}
	} else {
		cgrEvs = evs.(*erEvents)
		cgrEvs.events = append(cgrEvs.events, ev)
		cgrEvs.rdrCfg = rdrCfg
	}

	var cgrEv *utils.CGREvent
	if cgrEv, err = erS.preparePartialEvents(cgrEvs.events, cgrEvs.rdrCfg); err != nil {
		return
	}
	if partial := cgrEv.APIOpts[partialOpt]; !utils.IsSliceMember([]string{"false", utils.EmptyString}, utils.IfaceAsString(partial)) {
		erS.partialCache.Set(cgrID, cgrEvs, nil)
		return
	}

	// complete CDR
	if len(cgrEvs.events) != 1 {
		erS.partialCache.Set(cgrID, nil, nil) // set it with nil in cache to ignore when we expire the item
		erS.partialCache.Remove(cgrID)
	}
	go func() { erS.rdrEvents <- &erEvent{cgrEvent: cgrEv, rdrCfg: rdrCfg} }()
	return
}

func (erS *ERService) preparePartialEvents(cgrEvs []*utils.CGREvent, cfg *config.EventReaderCfg) (cgrEv *utils.CGREvent, err error) {
	cgrEv = cgrEvs[0]
	if len(cgrEvs) != 1 {
		ordFld := utils.IfaceAsString(cfg.Opts[utils.PartialOrderFieldOpt])
		if ordFld == utils.EmptyString {
			return nil, utils.NewErrMandatoryIeMissing(utils.PartialOrderFieldOpt)
		}
		fields := make([]interface{}, len(cgrEvs))

		var ordPath config.RSRParsers
		if ordPath, err = config.NewRSRParsers(ordFld, erS.cfg.GeneralCfg().RSRSep); err != nil {
			return nil, err
		}

		for i, ev := range cgrEvs {
			if fields[i], err = ordPath.ParseDataProviderWithInterfaces(ev.AsDataProvider()); err != nil {
				return
			}
			if fldStr, castStr := fields[i].(string); castStr { // attempt converting string since deserialization fails here (ie: time.Time fields)
				fields[i] = utils.StringToInterface(fldStr)
			}
		}
		//sort CGREvents based on partialOrderFieldOpt
		sort.Slice(cgrEvs, func(i, j int) bool {
			gt, serr := utils.GreaterThan(fields[i], fields[j], true)
			if serr != nil {
				err = serr
			}
			return gt
		})
		if err != nil {
			return
		}

		// compose the CGREvent from slice
		cgrEv = &utils.CGREvent{
			Tenant:  cgrEvs[0].Tenant,
			ID:      utils.UUIDSha1Prefix(),
			Time:    utils.TimePointer(time.Now()),
			Event:   make(map[string]interface{}),
			APIOpts: make(map[string]interface{}),
		}
		for _, ev := range cgrEvs {
			for key, value := range ev.Event {
				cgrEv.Event[key] = value
			}
			for key, val := range ev.APIOpts {
				cgrEv.APIOpts[key] = val
			}
		}
	}
	if len(cfg.PartialCommitFields) != 0 {
		agReq := agents.NewAgentRequest(
			utils.MapStorage(cgrEv.Event), nil,
			nil, nil, cgrEv.APIOpts, cfg.Tenant,
			erS.cfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(cfg.Timezone,
				erS.cfg.GeneralCfg().DefaultTimezone),
			erS.filterS, nil) // create an AgentRequest
		if err = agReq.SetFields(cfg.PartialCommitFields); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> processing partial event: <%s>, ignoring due to error: <%s>",
					utils.ERs, utils.ToJSON(cgrEv), err.Error()))
			return
		}
		cgrEv = utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	}

	return
}

func (erS *ERService) onEvicted(id string, value interface{}) {
	if value == nil {
		return
	}
	eEvs := value.(*erEvents)
	action := erS.cfg.ERsCfg().PartialCacheAction
	if cAct, has := eEvs.rdrCfg.Opts[utils.PartialCacheAction]; has {
		action = utils.IfaceAsString(cAct)
	}
	switch action {
	case utils.MetaNone:
	case utils.MetaPostCDR:
		cgrEv, err := erS.preparePartialEvents(eEvs.events, eEvs.rdrCfg)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed posting expired parial events <%s> due error <%s>",
					utils.ERs, utils.ToJSON(eEvs.events), err.Error()))
			return
		}
		erS.rdrEvents <- &erEvent{cgrEvent: cgrEv, rdrCfg: eEvs.rdrCfg}
	case utils.MetaDumpToFile:
		tmz := utils.FirstNonEmpty(eEvs.rdrCfg.Timezone, erS.cfg.GeneralCfg().DefaultTimezone)
		expPath := erS.cfg.ERsCfg().PartialPath
		if path, has := eEvs.rdrCfg.Opts[utils.PartialPathOpt]; has {
			expPath = utils.IfaceAsString(path)
		}
		if expPath == utils.EmptyString { // do not send the partial events to any file
			return
		}
		dumpFilePath := path.Join(expPath, fmt.Sprintf("%s.%d%s",
			id, time.Now().Unix(), utils.TmpSuffix))
		fileOut, err := os.Create(dumpFilePath)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed creating %s, error: %s",
				utils.ERs, dumpFilePath, err.Error()))
			return
		}
		defer fileOut.Close()
		csvWriter := csv.NewWriter(fileOut)
		if fldSep, has := eEvs.rdrCfg.Opts[utils.PartialCSVFieldSepartor]; has {
			csvWriter.Comma = rune(utils.IfaceAsString(fldSep)[0])
		}
		for _, ev := range eEvs.events {
			oNm := map[string]*utils.OrderedNavigableMap{
				utils.MetaExp: utils.NewOrderedNavigableMap(),
			}
			eeReq := engine.NewEventRequest(utils.MapStorage(ev.Event), utils.MapStorage{}, ev.APIOpts,
				eEvs.rdrCfg.Tenant, erS.cfg.GeneralCfg().DefaultTenant,
				tmz, erS.filterS, oNm)

			if err = eeReq.SetFields(eEvs.rdrCfg.CacheDumpFields); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> Converting CDR with CGRID: <%s> to record , ignoring due to error: <%s>",
						utils.ERs, id, err.Error()))
				return
			}

			record := eeReq.OrdNavMP[utils.MetaExp].OrderedFieldsAsStrings()
			if err = csvWriter.Write(record); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed writing partial record %v to file: %s, error: %s",
					utils.ERs, record, dumpFilePath, err.Error()))
				return
			}
		}
		csvWriter.Flush()
	}

}
