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

package ers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// erEvent is passed from reader to ERs
type erEvent struct {
	rawEvent map[string]any
	cgrEvent *utils.CGREvent
	rdrCfg   *config.EventReaderCfg
}

// NewERService instantiates the ERService
func NewERService(cfg *config.CGRConfig, datadb *engine.DataManager, filterS *engine.FilterS, connMgr *engine.ConnManager) (ers *ERService) {
	ers = &ERService{
		cfg:              cfg,
		dataManager:      datadb,
		rdrs:             make(map[string]EventReader),
		rdrPaths:         make(map[string]string),
		stopLsn:          make(map[string]chan struct{}),
		rdrEvents:        make(chan *erEvent),
		partialEvents:    make(chan *erEvent),
		rdrErr:           make(chan error),
		concurrentEvents: make(chan struct{}, cfg.ERsCfg().ConcurrentEvents),
		filterS:          filterS,
		connMgr:          connMgr,
	}
	ers.partialCache = ltcache.NewCache(ltcache.UnlimitedCaching, cfg.ERsCfg().PartialCacheTTL, false, false, []func(itmID string, value any){ers.onEvicted})
	return
}

// ERService is managing the EventReaders
type ERService struct {
	sync.RWMutex
	cfg              *config.CGRConfig
	dataManager      *engine.DataManager
	rdrs             map[string]EventReader   // map[rdrID]EventReader
	rdrPaths         map[string]string        // used for reloads in case of path changes
	stopLsn          map[string]chan struct{} // map[rdrID] chan struct{}
	rdrEvents        chan *erEvent            // receive here the events from readers
	partialEvents    chan *erEvent            // receive here the partial events from readers
	rdrErr           chan error               // receive here errors which should stop the app
	concurrentEvents chan struct{}

	filterS *engine.FilterS
	connMgr *engine.ConnManager

	partialCache *ltcache.Cache
}

// V1RunReaderParams contains required parameters for an ErSv1.RunReader request.
type V1RunReaderParams struct {
	Tenant   string
	ID       string // unique identifier of the request
	ReaderID string
	APIOpts  map[string]any
}

// V1RunReader processes files in the configured directory for the given reader. This function handles files
// based on the reader's type and configuration. Only available for readers that are not processing files
// automatically (RunDelay should equal 0).
//
// Note: This API is not safe to call concurrently for the same reader. Ensure the current files finish being
// processed before calling again.
func (erS *ERService) V1RunReader(ctx *context.Context, params V1RunReaderParams, reply *string) error {
	rdrCfg := erS.cfg.ERsCfg().ReaderCfg(params.ReaderID)
	er, has := erS.rdrs[params.ReaderID]
	if !has || rdrCfg == nil {
		return utils.ErrNotFound
	}
	if rdrCfg.RunDelay != 0 {
		return errors.New("readers with RunDelay different from 0 are not supported")
	}
	switch rdr := er.(type) {
	case *CSVFileER:
		processReaderDir(rdr.sourceDir, utils.CSVSuffix, rdr.processFile)
	case *XMLFileER:
		processReaderDir(rdr.sourceDir, utils.XMLSuffix, rdr.processFile)
	case *FWVFileER:
		processReaderDir(rdr.sourceDir, utils.FWVSuffix, rdr.processFile)
	case *JSONFileER:
		processReaderDir(rdr.sourceDir, utils.JSNSuffix, rdr.processFile)
	default:
		return errors.New("reader type does not yet support manual processing")
	}
	*reply = utils.OK
	return nil
}

// ListenAndServe keeps the service alive
func (erS *ERService) ListenAndServe(stopChan, cfgRldChan chan struct{}) error {
	for cfgIdx, rdrCfg := range erS.cfg.ERsCfg().Readers {
		if rdrCfg.Type == utils.MetaNone { // ignore *default reader
			continue
		}
		if err := erS.addReader(rdrCfg.ID, cfgIdx); err != nil {
			utils.Logger.Crit(
				fmt.Sprintf("<%s> adding reader <%s> got error: <%s>",
					utils.ERs, rdrCfg.ID, err.Error()))
			return err
		}
	}
	for {
		select {
		case err := <-erS.rdrErr: // got application error
			erS.closeAllRdrs()
			return err
		case <-stopChan:
			erS.closeAllRdrs()
			return nil
		case erEv := <-erS.rdrEvents:
			erS.concurrentEvents <- struct{}{}
			go func() {
				err := erS.processEvent(erEv.cgrEvent, erEv.rdrCfg)
				if err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> reading event: <%s> from reader: <%s> got error: <%v>",
							utils.ERs, utils.ToJSON(erEv.cgrEvent), erEv.rdrCfg.ID, err))
				}
				if err = erS.exportRawEvent(erEv, err != nil); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> exporting event: <%s> from reader: <%s> got error: <%v>",
							utils.ERs, utils.ToJSON(erEv.cgrEvent), erEv.rdrCfg.ID, err))
				}
				<-erS.concurrentEvents
			}()
		case pEv := <-erS.partialEvents:
			err := erS.processPartialEvent(pEv.cgrEvent, pEv.rdrCfg)
			if err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> reading partial event: <%s> from reader: <%s> got error: <%v>",
						utils.ERs, utils.ToJSON(pEv.cgrEvent), pEv.rdrCfg.ID, err))
			}
			if err = erS.exportRawEvent(pEv, err != nil); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> exporting partial event: <%s> from reader: <%s> got error: <%v>",
						utils.ERs, utils.ToJSON(pEv.cgrEvent), pEv.rdrCfg.ID, err))
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
					erS.closeAllRdrs()
					erS.Unlock()
					return err
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
		erS.filterS, erS.stopLsn[rdrID], erS.dataManager); err != nil {
		return
	}
	erS.rdrs[rdrID] = rdr
	return rdr.Serve()
}

// processEvent will be called each time a new event is received from readers
func (erS *ERService) processEvent(cgrEv *utils.CGREvent,
	rdrCfg *config.EventReaderCfg) (err error) {
	startTime := time.Now()
	replyState := utils.OK

	// Defer stats and thresholds processing to ensure it happens even with early returns.
	defer func() {
		endTime := time.Now()
		if rdrCfg.Flags.Has(utils.MetaDryRun) {
			return
		}
		rawStatIDs := rdrCfg.Flags.ParamValue(utils.MetaERsStats)
		rawThIDs := rdrCfg.Flags.ParamValue(utils.MetaERsThresholds)

		// Early return if nothing to process.
		if rawStatIDs == "" && rawThIDs == "" {
			return
		}

		// Clone is needed to prevent data races if requests are sent
		// asynchronously.
		ev := cgrEv.Clone()

		ev.Event[utils.ReplyState] = replyState
		ev.Event[utils.StartTime] = startTime
		ev.Event[utils.EndTime] = endTime
		ev.Event[utils.ProcessingTime] = endTime.Sub(startTime)
		ev.Event[utils.Source] = utils.ERs
		ev.APIOpts[utils.MetaEventType] = utils.ProcessTime

		if rawStatIDs != "" {
			statIDs := strings.Split(rawStatIDs, utils.ANDSep)
			ev.APIOpts[utils.OptsStatsProfileIDs] = statIDs
			var reply []string
			if err := erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().StatSConns,
				utils.StatSv1ProcessEvent, ev, &reply); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> failed to process event in %s: %v",
					utils.ERs, utils.StatS, err))
			}
		}
		if rawThIDs != "" {
			thIDs := strings.Split(rawThIDs, utils.ANDSep)
			ev.APIOpts[utils.OptsThresholdsProfileIDs] = thIDs
			var reply []string
			if err := erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().ThresholdSConns,
				utils.ThresholdSv1ProcessEvent, ev, &reply); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> failed to process event in %s: %v",
					utils.ERs, utils.ThresholdS, err))
			}
		}
	}()
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
		utils.MetaCDRs, utils.MetaEvent, utils.MetaNone, utils.MetaExport} {
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
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> args extraction for reader <%s> failed because <%s>",
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
		if err != nil {
			replyState = utils.ErrReplyStateAuthorize
		}
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
		if err != nil {
			replyState = utils.ErrReplyStateInitiate
		}
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			rdrCfg.Flags.Has(utils.MetaAttributes),
			rdrCfg.Flags.Has(utils.MetaThresholds),
			rdrCfg.Flags.Has(utils.MetaStats),
			rdrCfg.Flags.ParamsSlice(utils.MetaAttributes, utils.MetaIDs),
			rdrCfg.Flags.ParamsSlice(utils.MetaThresholds, utils.MetaIDs),
			rdrCfg.Flags.ParamsSlice(utils.MetaStats, utils.MetaIDs),
			rdrCfg.Flags.Has(utils.MetaAccounts),
			cgrEv, rdrCfg.Flags.Has(utils.MetaFD))
		rply := new(sessions.V1UpdateSessionReply)
		err = erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().SessionSConns, utils.SessionSv1UpdateSession,
			updateArgs, rply)
		if err != nil {
			replyState = utils.ErrReplyStateUpdate
		}
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
		if err != nil {
			replyState = utils.ErrReplyStateTerminate
		}
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
		if err != nil {
			replyState = utils.ErrReplyStateMessage
		}
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if evArgs.Debit {
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
		if err != nil {
			replyState = utils.ErrReplyStateEvent
		}
	case utils.MetaCDRs: // allow CDR processing
	case utils.MetaExport: // allow event exporting
	}
	if err != nil {
		return
	}
	// separate request so we can capture the Terminate/Event also here
	if rdrCfg.Flags.Has(utils.MetaCDRs) &&
		!rdrCfg.Flags.Has(utils.MetaDryRun) {
		rplyCDRs := utils.StringPointer("")
		if err := erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().SessionSConns,
			utils.SessionSv1ProcessCDR, cgrEv, rplyCDRs); err != nil {
			replyState = utils.ErrReplyStateCDRs
			return err
		}
	}
	if rdrCfg.Flags.Has(utils.MetaExport) {
		var reply map[string]map[string]any
		if err := erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().EEsConns,
			utils.EeSv1ProcessEvent,
			&engine.CGREventWithEeIDs{
				EeIDs:    rdrCfg.EEsIDs,
				CGREvent: cgrEv,
			}, &reply); err != nil {
			replyState = utils.ErrReplyStateExport
			return err
		}
	}
	return
}

func (erS *ERService) closeAllRdrs() {
	for _, stopL := range erS.stopLsn {
		close(stopL)
	}
}

type erEvents struct {
	events []*utils.CGREvent
	rdrCfg *config.EventReaderCfg
}

// processPartialEvent process the event as a partial event
func (erS *ERService) processPartialEvent(ev *utils.CGREvent, rdrCfg *config.EventReaderCfg) (err error) {
	// to identify the event the originID and originHost is used to create the CGRID
	orgID, err := ev.FieldAsString(utils.OriginID)
	if err == utils.ErrNotFound { // the field is missing ignore the event
		utils.Logger.Warning(
			fmt.Sprintf("<%s> Missing <OriginID> field for partial event <%s>",
				utils.ERs, utils.ToJSON(ev)))
		return
	}
	orgHost := utils.IfaceAsString(ev.Event[utils.OriginHost])
	cgrID := utils.Sha1(orgID, orgHost)

	evs, has := erS.partialCache.Get(cgrID) // get the existing events from cache
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
	if cgrEv, err = mergePartialEvents(cgrEvs.events, cgrEvs.rdrCfg, erS.filterS, // merge the events
		erS.cfg.GeneralCfg().DefaultTenant,
		erS.cfg.GeneralCfg().DefaultTimezone,
		erS.cfg.GeneralCfg().RSRSep); err != nil {
		return
	}
	if partial := cgrEv.APIOpts[utils.PartialOpt]; !slices.Contains([]string{utils.FalseStr, utils.EmptyString},
		utils.IfaceAsString(partial)) { // if is still partial set it back in cache
		erS.partialCache.Set(cgrID, cgrEvs, nil)
		return
	}

	// complete event
	if len(cgrEvs.events) != 1 { // remove it from cache if there were events in cache
		erS.partialCache.Set(cgrID, nil, nil) // set it with nil in cache to ignore when we expire the item
		erS.partialCache.Remove(cgrID)
	}
	go func() { erS.rdrEvents <- &erEvent{cgrEvent: cgrEv, rdrCfg: rdrCfg} }() // put the event on the complete events chanel( in a goroutine to not block the select from ListenAndServe)
	return
}

// onEvicted the function that is called when a element is removed from cache
func (erS *ERService) onEvicted(id string, value any) {
	if value == nil { // is already complete and sent to erS
		return
	}
	eEvs := value.(*erEvents)
	var action string
	if eEvs.rdrCfg.Opts.PartialCacheAction != nil {
		action = *eEvs.rdrCfg.Opts.PartialCacheAction
	}
	switch action {
	case utils.MetaNone: // do nothing with the events
	case utils.MetaPostCDR: // merge the events and post the to erS
		cgrEv, err := mergePartialEvents(eEvs.events, eEvs.rdrCfg, erS.filterS,
			erS.cfg.GeneralCfg().DefaultTenant,
			erS.cfg.GeneralCfg().DefaultTimezone,
			erS.cfg.GeneralCfg().RSRSep)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed posting expired parial events <%s> due error <%s>",
					utils.ERs, utils.ToJSON(eEvs.events), err.Error()))
			return
		}
		erS.rdrEvents <- &erEvent{cgrEvent: cgrEv, rdrCfg: eEvs.rdrCfg}
	case utils.MetaDumpToFile: // apply the cacheDumpFields to the united events and write the record to file
		expPath := eEvs.rdrCfg.ProcessedPath
		if eEvs.rdrCfg.Opts.PartialPath != nil {
			expPath = *eEvs.rdrCfg.Opts.PartialPath
		}
		if expPath == utils.EmptyString { // do not write the partial event to file
			return
		}
		cgrEv, err := mergePartialEvents(eEvs.events, eEvs.rdrCfg, erS.filterS, // merge the partial events
			erS.cfg.GeneralCfg().DefaultTenant,
			erS.cfg.GeneralCfg().DefaultTimezone,
			erS.cfg.GeneralCfg().RSRSep)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed posting expired parial events <%s> due error <%s>",
					utils.ERs, utils.ToJSON(eEvs.events), err.Error()))
			return
		}
		var record []string
		if len(eEvs.rdrCfg.CacheDumpFields) != 0 {
			// convert the event to record
			eeReq := engine.NewExportRequest(map[string]utils.DataStorage{
				utils.MetaReq:  utils.MapStorage(cgrEv.Event),
				utils.MetaOpts: utils.MapStorage(cgrEv.APIOpts),
				utils.MetaCfg:  erS.cfg.GetDataProvider(),
			}, utils.FirstNonEmpty(cgrEv.Tenant, erS.cfg.GeneralCfg().DefaultTenant),
				erS.filterS, map[string]*utils.OrderedNavigableMap{
					utils.MetaExp: utils.NewOrderedNavigableMap(),
				})

			if err = eeReq.SetFields(eEvs.rdrCfg.CacheDumpFields); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> Converting CDR with CGRID: <%s> to record , ignoring due to error: <%s>",
						utils.ERs, id, err.Error()))
				return
			}

			record = eeReq.ExpData[utils.MetaExp].OrderedFieldsAsStrings()
		} else {
			keys := make([]string, 0, len(cgrEv.Event))
			for k := range cgrEv.Event {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			record = make([]string, len(keys))
			for i, k := range keys {
				record[i] = utils.IfaceAsString(cgrEv.Event[k])
			}
		}
		// open the file and write the record
		dumpFilePath := path.Join(expPath, fmt.Sprintf("%s.%d%s",
			id, time.Now().Unix(), utils.TmpSuffix))
		fileOut, err := os.Create(dumpFilePath)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed creating %s, error: %s",
				utils.ERs, dumpFilePath, err.Error()))
			return
		}
		csvWriter := csv.NewWriter(fileOut)
		if eEvs.rdrCfg.Opts.CSV.PartialCSVFieldSeparator != nil {
			csvWriter.Comma = rune((*eEvs.rdrCfg.Opts.CSV.PartialCSVFieldSeparator)[0])
		}

		if err = csvWriter.Write(record); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed writing partial record %v to file: %s, error: %s",
				utils.ERs, record, dumpFilePath, err.Error()))
		}
		csvWriter.Flush()
		fileOut.Close()
	case utils.MetaDumpToJSON: // apply the cacheDumpFields to the united events and write the record to file
		expPath := eEvs.rdrCfg.ProcessedPath
		if eEvs.rdrCfg.Opts.PartialPath != nil {
			expPath = *eEvs.rdrCfg.Opts.PartialPath
		}
		if expPath == utils.EmptyString { // do not write the partial event to file
			return
		}
		cgrEv, err := mergePartialEvents(eEvs.events, eEvs.rdrCfg, erS.filterS, // merge the partial events
			erS.cfg.GeneralCfg().DefaultTenant,
			erS.cfg.GeneralCfg().DefaultTimezone,
			erS.cfg.GeneralCfg().RSRSep)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed posting expired parial events <%s> due error <%s>",
					utils.ERs, utils.ToJSON(eEvs.events), err.Error()))
			return
		}
		var record map[string]any
		if len(eEvs.rdrCfg.CacheDumpFields) != 0 {
			// convert the event to record
			eeReq := engine.NewExportRequest(map[string]utils.DataStorage{
				utils.MetaReq:  utils.MapStorage(cgrEv.Event),
				utils.MetaOpts: utils.MapStorage(cgrEv.APIOpts),
				utils.MetaCfg:  erS.cfg.GetDataProvider(),
			}, utils.FirstNonEmpty(cgrEv.Tenant, erS.cfg.GeneralCfg().DefaultTenant),
				erS.filterS, map[string]*utils.OrderedNavigableMap{
					utils.MetaExp: utils.NewOrderedNavigableMap(),
				})

			if err = eeReq.SetFields(eEvs.rdrCfg.CacheDumpFields); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> Converting CDR with CGRID: <%s> to record , ignoring due to error: <%s>",
						utils.ERs, id, err.Error()))
				return
			}

			record = make(map[string]any)
			for el := eeReq.ExpData[utils.MetaExp].GetFirstElement(); el != nil; el = el.Next() {
				path := el.Value
				nmIt, _ := eeReq.ExpData[utils.MetaExp].Field(path)
				path = path[:len(path)-1] // remove the last index
				record[strings.Join(path, utils.NestingSep)] = nmIt.Data
			}
		} else {
			record = cgrEv.Event
		}
		// open the file and write the record
		dumpFilePath := path.Join(expPath, fmt.Sprintf("%s.%d%s",
			id, time.Now().Unix(), utils.TmpSuffix))
		fileOut, err := os.Create(dumpFilePath)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed creating %s, error: %s",
				utils.ERs, dumpFilePath, err.Error()))
			return
		}

		if err = json.NewEncoder(fileOut).Encode(record); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed writing partial record %v to file: %s, error: %s",
				utils.ERs, record, dumpFilePath, err.Error()))
		}
		fileOut.Close()

	}

}

// exportRawEvent exports the given event. If the processing of the event failed,
// it uses ees_failed_ids; otherwise, it uses ees_success_ids.
func (erS *ERService) exportRawEvent(event *erEvent, processingFailed bool) error {
	var exporterIDs []string
	if processingFailed {
		if len(event.rdrCfg.EEsFailedIDs) == 0 {
			return nil
		}
		exporterIDs = event.rdrCfg.EEsFailedIDs
	} else {
		if len(event.rdrCfg.EEsSuccessIDs) == 0 {
			return nil
		}
		exporterIDs = event.rdrCfg.EEsSuccessIDs
	}
	var reply map[string]map[string]any
	return erS.connMgr.Call(context.TODO(), erS.cfg.ERsCfg().EEsConns, utils.EeSv1ProcessEvent,
		&engine.CGREventWithEeIDs{
			EeIDs: exporterIDs,
			CGREvent: &utils.CGREvent{
				Tenant: erS.cfg.GeneralCfg().DefaultTenant,
				Event:  event.rawEvent,
			},
		}, &reply)
}
