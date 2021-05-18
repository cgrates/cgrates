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

package engine

import (
	"fmt"
	"net/http"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

func newMapEventFromReqForm(r *http.Request) (mp MapEvent, err error) {
	if r.Form == nil {
		if err = r.ParseForm(); err != nil {
			return
		}
	}
	mp = MapEvent{utils.Source: r.RemoteAddr}
	for k, vals := range r.Form {
		mp[k] = vals[0] // We only support the first value for now, if more are provided it is considered remote's fault
	}
	return
}

// cgrCdrHandler handles CDRs received over HTTP REST
func (cdrS *CDRServer) cgrCdrHandler(w http.ResponseWriter, r *http.Request) {
	cgrCDR, err := newMapEventFromReqForm(r)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> could not create CDR entry from http: %+v, err <%s>",
				utils.CDRs, r.Form, err.Error()))
		return
	}
	cdr, err := cgrCDR.AsCDR(cdrS.cgrCfg, cdrS.cgrCfg.GeneralCfg().DefaultTenant, cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> could not create CDR entry from rawCDR: %+v, err <%s>",
				utils.CDRs, cgrCDR, err.Error()))
		return
	}
	var ignored string
	if err := cdrS.V1ProcessCDR(&CDRWithAPIOpts{CDR: cdr}, &ignored); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> processing CDR: %s, err: <%s>",
				utils.CDRs, cdr, err.Error()))
	}
}

// fsCdrHandler will handle CDRs received from FreeSWITCH over HTTP-JSON
func (cdrS *CDRServer) fsCdrHandler(w http.ResponseWriter, r *http.Request) {
	fsCdr, err := NewFSCdr(r.Body, cdrS.cgrCfg)
	r.Body.Close()
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Could not create CDR entry: %s", err.Error()))
		return
	}
	cdr, err := fsCdr.AsCDR(cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<CDRS> Could not create AsCDR entry: %s", err.Error()))
		return
	}
	var ignored string
	if err := cdrS.V1ProcessCDR(&CDRWithAPIOpts{CDR: cdr}, &ignored); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> processing CDR: %s, err: <%s>",
				utils.CDRs, cdr, err.Error()))
	}
}

// NewCDRServer is a constructor for CDRServer
func NewCDRServer(cgrCfg *config.CGRConfig, storDBChan chan StorDB, dm *DataManager, filterS *FilterS,
	connMgr *ConnManager) *CDRServer {
	cdrDB := <-storDBChan
	return &CDRServer{
		cgrCfg:     cgrCfg,
		cdrDB:      cdrDB,
		dm:         dm,
		guard:      guardian.Guardian,
		filterS:    filterS,
		connMgr:    connMgr,
		storDBChan: storDBChan,
	}
}

// CDRServer stores and rates CDRs
type CDRServer struct {
	cgrCfg     *config.CGRConfig
	cdrDB      CdrStorage
	dm         *DataManager
	guard      *guardian.GuardianLocker
	filterS    *FilterS
	connMgr    *ConnManager
	storDBChan chan StorDB
}

// ListenAndServe listen for storbd reload
func (cdrS *CDRServer) ListenAndServe(stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		case stordb, ok := <-cdrS.storDBChan:
			if !ok { // the chanel was closed by the shutdown of stordbService
				return
			}
			cdrS.cdrDB = stordb
		}
	}
}

// RegisterHandlersToServer is called by cgr-engine to register HTTP URL handlers
func (cdrS *CDRServer) RegisterHandlersToServer(server utils.Server) {
	server.RegisterHTTPFunc(cdrS.cgrCfg.HTTPCfg().CDRsURL, cdrS.cgrCdrHandler)
	server.RegisterHTTPFunc(cdrS.cgrCfg.HTTPCfg().FreeswitchCDRsURL, cdrS.fsCdrHandler)
}

// chrgrSProcessEvent forks CGREventWithOpts into multiples based on matching ChargerS profiles
func (cdrS *CDRServer) chrgrSProcessEvent(cgrEv *utils.CGREvent) (cgrEvs []*utils.CGREvent, err error) {
	var chrgrs []*ChrgSProcessEventReply
	if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().ChargerSConns,
		utils.ChargerSv1ProcessEvent,
		cgrEv, &chrgrs); err != nil {
		return
	}
	if len(chrgrs) == 0 {
		return
	}
	cgrEvs = make([]*utils.CGREvent, len(chrgrs))
	for i, cgrPrfl := range chrgrs {
		cgrEvs[i] = cgrPrfl.CGREvent
	}
	return
}

// attrSProcessEvent will send the event to StatS if the connection is configured
func (cdrS *CDRServer) attrSProcessEvent(cgrEv *utils.CGREvent) (err error) {
	var rplyEv AttrSProcessEventReply
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(map[string]interface{})
	}
	cgrEv.APIOpts[utils.Subsys] = utils.MetaCDRs
	var processRuns *int
	if val, has := cgrEv.APIOpts[utils.OptsAttributesProcessRuns]; has {
		if v, err := utils.IfaceAsTInt64(val); err == nil {
			processRuns = utils.IntPointer(int(v))
		}
	}
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
		utils.MetaCDRs)
	attrArgs := &AttrArgsProcessEvent{
		CGREvent:    cgrEv,
		ProcessRuns: processRuns,
	}
	if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().AttributeSConns,
		utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
		*cgrEv = *rplyEv.CGREvent
	} else if err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // cancel ErrNotFound
	}
	return
}

// thdSProcessEvent will send the event to ThresholdS
func (cdrS *CDRServer) thdSProcessEvent(cgrEv *utils.CGREvent) (err error) {
	var tIDs []string
	// we clone the CGREvent so we can add EventType without being propagated
	thArgs := &ThresholdsArgsProcessEvent{
		CGREvent: cgrEv.Clone(),
	}
	if thArgs.APIOpts == nil {
		thArgs.APIOpts = make(map[string]interface{})
	}
	thArgs.APIOpts[utils.MetaEventType] = utils.CDR
	if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().ThresholdSConns,
		utils.ThresholdSv1ProcessEvent,
		thArgs, &tIDs); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// statSProcessEvent will send the event to StatS
func (cdrS *CDRServer) statSProcessEvent(cgrEv *utils.CGREvent) (err error) {
	var reply []string
	statArgs := &StatsArgsProcessEvent{
		CGREvent: cgrEv.Clone(),
	}
	if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().StatSConns,
		utils.StatSv1ProcessEvent,
		statArgs, &reply); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// eeSProcessEvent will process the event with the EEs component
func (cdrS *CDRServer) eeSProcessEvent(cgrEv *utils.CGREventWithEeIDs) (err error) {
	var reply map[string]map[string]interface{}
	if err = cdrS.connMgr.Call(context.TODO(), cdrS.cgrCfg.CdrsCfg().EEsConns,
		utils.EeSv1ProcessEvent,
		cgrEv, &reply); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// processEvent processes a CGREvent based on arguments
// in case of partially executed, both error and evs will be returned
func (cdrS *CDRServer) processEvent(ev *utils.CGREvent,
	chrgS, attrS, refund, ralS, store, reRate, export, thdS, stS bool) (evs []*utils.EventWithFlags, err error) {
	if attrS {
		if err = cdrS.attrSProcessEvent(ev); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
					utils.CDRs, err.Error(), utils.ToJSON(ev), utils.AttributeS))
			err = utils.ErrPartiallyExecuted
			return
		}
	}
	var cgrEvs []*utils.CGREvent
	if chrgS {
		if cgrEvs, err = cdrS.chrgrSProcessEvent(ev); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
					utils.CDRs, err.Error(), utils.ToJSON(ev), utils.ChargerS))
			err = utils.ErrPartiallyExecuted
			return
		}
	} else { // ChargerS not requested, charge the original event
		cgrEvs = []*utils.CGREvent{ev}
	}
	// Check if the unique ID was not already processed
	if !refund {
		for _, cgrEv := range cgrEvs {
			me := MapEvent(cgrEv.Event)
			if !me.HasField(utils.CGRID) { // try to compute the CGRID if missing
				me[utils.CGRID] = utils.Sha1(
					me.GetStringIgnoreErrors(utils.OriginID),
					me.GetStringIgnoreErrors(utils.OriginHost),
				)
			}
			uID := utils.ConcatenatedKey(
				me.GetStringIgnoreErrors(utils.CGRID),
				me.GetStringIgnoreErrors(utils.RunID),
			)
			if Cache.HasItem(utils.CacheCDRIDs, uID) && !reRate {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, utils.ErrExists, utils.ToJSON(cgrEv), utils.CacheS))
				return nil, utils.ErrExists
			}
			if errCh := Cache.Set(context.TODO(), utils.CacheCDRIDs, uID, true, nil,
				cacheCommit(utils.NonTransactional), utils.NonTransactional); errCh != nil {
				return nil, errCh
			}
		}
	}
	// Populate CDR list out of events
	cdrs := make([]*CDR, len(cgrEvs))
	if refund || ralS || store || reRate || export {
		for i, cgrEv := range cgrEvs {
			if cdrs[i], err = NewMapEvent(cgrEv.Event).AsCDR(cdrS.cgrCfg,
				cgrEv.Tenant, cdrS.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> converting event %+v to CDR",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv)))
				err = utils.ErrPartiallyExecuted
				return
			}
		}
	}
	procFlgs := make([]utils.StringSet, len(cgrEvs)) // will save the flags for the reply here
	for i := range cgrEvs {
		procFlgs[i] = utils.NewStringSet(nil)
	}

	var partiallyExecuted bool // from here actions are optional and a general error is returned
	if export {
		if len(cdrS.cgrCfg.CdrsCfg().EEsConns) != 0 {
			for _, cgrEv := range cgrEvs {
				evWithOpts := &utils.CGREventWithEeIDs{
					CGREvent: cgrEv,
					EeIDs:    cdrS.cgrCfg.CdrsCfg().OnlineCDRExports,
				}
				if err = cdrS.eeSProcessEvent(evWithOpts); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: <%s> exporting cdr %+v",
							utils.CDRs, err.Error(), utils.ToJSON(evWithOpts)))
					partiallyExecuted = true
				}
			}
		}
	}
	if thdS {
		for _, cgrEv := range cgrEvs {
			if err = cdrS.thdSProcessEvent(cgrEv); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv), utils.ThresholdS))
				partiallyExecuted = true
			}
		}
	}
	if stS {
		for _, cgrEv := range cgrEvs {
			if err = cdrS.statSProcessEvent(cgrEv); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv), utils.StatS))
				partiallyExecuted = true
			}
		}
	}
	if partiallyExecuted {
		err = utils.ErrPartiallyExecuted
	}
	evs = make([]*utils.EventWithFlags, len(cgrEvs))
	for i, cgrEv := range cgrEvs {
		evs[i] = &utils.EventWithFlags{
			Flags: procFlgs[i].AsSlice(),
			Event: cgrEv.Event,
		}
	}
	return
}

// V1ProcessCDR processes a CDR
func (cdrS *CDRServer) V1ProcessCDR(cdr *CDRWithAPIOpts, reply *string) (err error) {
	if cdr.CGRID == utils.EmptyString { // Populate CGRID if not present
		cdr.ComputeCGRID()
	}
	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessCDR, cdr.CGRID, cdr.RunID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer Cache.Set(context.TODO(), utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if cdr.RequestType == utils.EmptyString {
		cdr.RequestType = cdrS.cgrCfg.GeneralCfg().DefaultReqType
	}
	if cdr.Tenant == utils.EmptyString {
		cdr.Tenant = cdrS.cgrCfg.GeneralCfg().DefaultTenant
	}
	if cdr.Category == utils.EmptyString {
		cdr.Category = cdrS.cgrCfg.GeneralCfg().DefaultCategory
	}
	if cdr.Subject == utils.EmptyString { // Use account information as rating subject if missing
		cdr.Subject = cdr.Account
	}
	if cdr.RunID == utils.EmptyString {
		cdr.RunID = utils.MetaDefault
	}
	cgrEv := cdr.AsCGREvent()
	cgrEv.APIOpts = cdr.APIOpts

	if _, err = cdrS.processEvent(cgrEv,
		len(cdrS.cgrCfg.CdrsCfg().ChargerSConns) != 0 && !cdr.PreRated,
		len(cdrS.cgrCfg.CdrsCfg().AttributeSConns) != 0,
		false,
		!cdr.PreRated, // rate the CDR if is not PreRated
		cdrS.cgrCfg.CdrsCfg().StoreCdrs,
		false, // no rerate
		len(cdrS.cgrCfg.CdrsCfg().OnlineCDRExports) != 0 || len(cdrS.cgrCfg.CdrsCfg().EEsConns) != 0,
		len(cdrS.cgrCfg.CdrsCfg().ThresholdSConns) != 0,
		len(cdrS.cgrCfg.CdrsCfg().StatSConns) != 0); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// ArgV1ProcessEvent is the CGREvent with proccesing Flags
type ArgV1ProcessEvent struct {
	Flags []string
	utils.CGREvent
	clnb bool //rpcclonable
}

// SetCloneable sets if the args should be clonned on internal connections
func (attr *ArgV1ProcessEvent) SetCloneable(rpcCloneable bool) {
	attr.clnb = rpcCloneable
}

// RPCClone implements rpcclient.RPCCloner interface
func (attr *ArgV1ProcessEvent) RPCClone() (interface{}, error) {
	if !attr.clnb {
		return attr, nil
	}
	return attr.Clone(), nil
}

// Clone creates a clone of the object
func (attr *ArgV1ProcessEvent) Clone() *ArgV1ProcessEvent {
	var flags []string
	if attr.Flags != nil {
		flags = make([]string, len(attr.Flags))
		for i, id := range attr.Flags {
			flags[i] = id
		}
	}
	return &ArgV1ProcessEvent{
		Flags:    flags,
		CGREvent: *attr.CGREvent.Clone(),
	}
}

// V1ProcessEvent will process the CGREvent
func (cdrS *CDRServer) V1ProcessEvent(arg *ArgV1ProcessEvent, reply *string) (err error) {
	if arg.CGREvent.ID == utils.EmptyString {
		arg.CGREvent.ID = utils.GenUUID()
	}
	if arg.CGREvent.Tenant == utils.EmptyString {
		arg.CGREvent.Tenant = cdrS.cgrCfg.GeneralCfg().DefaultTenant
	}
	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessEvent, arg.CGREvent.ID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer Cache.Set(context.TODO(), utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	// processing options
	flgs := utils.FlagsWithParamsFromSlice(arg.Flags)
	attrS := len(cdrS.cgrCfg.CdrsCfg().AttributeSConns) != 0
	if flgs.Has(utils.MetaAttributes) {
		attrS = flgs.GetBool(utils.MetaAttributes)
	}
	store := cdrS.cgrCfg.CdrsCfg().StoreCdrs
	if flgs.Has(utils.MetaStore) {
		store = flgs.GetBool(utils.MetaStore)
	}
	export := len(cdrS.cgrCfg.CdrsCfg().OnlineCDRExports) != 0 || len(cdrS.cgrCfg.CdrsCfg().EEsConns) != 0
	if flgs.Has(utils.MetaExport) {
		export = flgs.GetBool(utils.MetaExport)
	}
	thdS := len(cdrS.cgrCfg.CdrsCfg().ThresholdSConns) != 0
	if flgs.Has(utils.MetaThresholds) {
		thdS = flgs.GetBool(utils.MetaThresholds)
	}
	stS := len(cdrS.cgrCfg.CdrsCfg().StatSConns) != 0
	if flgs.Has(utils.MetaStats) {
		stS = flgs.GetBool(utils.MetaStats)
	}
	chrgS := len(cdrS.cgrCfg.CdrsCfg().ChargerSConns) != 0 // activate charging for the Event
	if flgs.Has(utils.MetaChargers) {
		chrgS = flgs.GetBool(utils.MetaChargers)
	}
	var ralS bool // activate single rating for the CDR
	// if flgs.Has(utils.MetaRALs) {
	// 	ralS = flgs.GetBool(utils.MetaRALs)
	// }
	var reRate bool
	// if flgs.Has(utils.MetaRerate) {
	// 	if reRate = flgs.GetBool(utils.MetaRerate); reRate {
	// 		ralS = true
	// 	}
	// }
	var refund bool
	if flgs.Has(utils.MetaRefund) {
		refund = flgs.GetBool(utils.MetaRefund)
	}
	// end of processing options

	if _, err = cdrS.processEvent(&arg.CGREvent, chrgS, attrS, refund,
		ralS, store, reRate, export, thdS, stS); err != nil {
		return
	}
	*reply = utils.OK
	return nil
}

// V2ProcessEvent has the same logic with V1ProcessEvent except it adds the proccessed events to the reply
func (cdrS *CDRServer) V2ProcessEvent(arg *ArgV1ProcessEvent, evs *[]*utils.EventWithFlags) (err error) {
	if arg.ID == "" {
		arg.ID = utils.GenUUID()
	}
	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.CDRsV2ProcessEvent, arg.CGREvent.ID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*evs = *cachedResp.Result.(*[]*utils.EventWithFlags)
			}
			return cachedResp.Error
		}
		defer Cache.Set(context.TODO(), utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: evs, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	// processing options
	flgs := utils.FlagsWithParamsFromSlice(arg.Flags)
	attrS := len(cdrS.cgrCfg.CdrsCfg().AttributeSConns) != 0
	if flgs.Has(utils.MetaAttributes) {
		attrS = flgs.GetBool(utils.MetaAttributes)
	}
	store := cdrS.cgrCfg.CdrsCfg().StoreCdrs
	if flgs.Has(utils.MetaStore) {
		store = flgs.GetBool(utils.MetaStore)
	}
	export := len(cdrS.cgrCfg.CdrsCfg().OnlineCDRExports) != 0 || len(cdrS.cgrCfg.CdrsCfg().EEsConns) != 0
	if flgs.Has(utils.MetaExport) {
		export = flgs.GetBool(utils.MetaExport)
	}
	thdS := len(cdrS.cgrCfg.CdrsCfg().ThresholdSConns) != 0
	if flgs.Has(utils.MetaThresholds) {
		thdS = flgs.GetBool(utils.MetaThresholds)
	}
	stS := len(cdrS.cgrCfg.CdrsCfg().StatSConns) != 0
	if flgs.Has(utils.MetaStats) {
		stS = flgs.GetBool(utils.MetaStats)
	}
	chrgS := len(cdrS.cgrCfg.CdrsCfg().ChargerSConns) != 0 // activate charging for the Event
	if flgs.Has(utils.MetaChargers) {
		chrgS = flgs.GetBool(utils.MetaChargers)
	}
	var ralS bool // activate single rating for the CDR
	// if flgs.Has(utils.MetaRALs) {
	// 	ralS = flgs.GetBool(utils.MetaRALs)
	// }
	var reRate bool
	// if flgs.Has(utils.MetaRerate) {
	// reRate = flgs.GetBool(utils.MetaRerate)
	// if reRate {
	// ralS = true
	// }
	// }
	var refund bool
	if flgs.Has(utils.MetaRefund) {
		refund = flgs.GetBool(utils.MetaRefund)
	}
	// end of processing options

	var procEvs []*utils.EventWithFlags
	if procEvs, err = cdrS.processEvent(&arg.CGREvent, chrgS, attrS, refund,
		ralS, store, reRate, export, thdS, stS); err != nil {
		return
	}
	*evs = procEvs
	return nil
}

// ArgRateCDRs a cdr with extra flags
type ArgRateCDRs struct {
	Flags []string
	utils.RPCCDRsFilter
	Tenant  string
	APIOpts map[string]interface{}
}

// V1RateCDRs is used for re-/rate CDRs which are already stored within StorDB
// FixMe: add RPC caching
func (cdrS *CDRServer) V1RateCDRs(arg *ArgRateCDRs, reply *string) (err error) {
	var cdrFltr *utils.CDRsFilter
	if cdrFltr, err = arg.RPCCDRsFilter.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
		return utils.NewErrServerError(err)
	}
	var cdrs []*CDR
	if cdrs, _, err = cdrS.cdrDB.GetCDRs(cdrFltr, false); err != nil {
		return
	}
	flgs := utils.FlagsWithParamsFromSlice(arg.Flags)
	store := cdrS.cgrCfg.CdrsCfg().StoreCdrs
	if flgs.Has(utils.MetaStore) {
		store = flgs.GetBool(utils.MetaStore)
	}
	export := len(cdrS.cgrCfg.CdrsCfg().OnlineCDRExports) != 0 || len(cdrS.cgrCfg.CdrsCfg().EEsConns) != 0
	if flgs.Has(utils.MetaExport) {
		export = flgs.GetBool(utils.MetaExport)
	}
	thdS := len(cdrS.cgrCfg.CdrsCfg().ThresholdSConns) != 0
	if flgs.Has(utils.MetaThresholds) {
		thdS = flgs.GetBool(utils.MetaThresholds)
	}
	statS := len(cdrS.cgrCfg.CdrsCfg().StatSConns) != 0
	if flgs.Has(utils.MetaStats) {
		statS = flgs.GetBool(utils.MetaStats)
	}
	chrgS := len(cdrS.cgrCfg.CdrsCfg().ChargerSConns) != 0
	if flgs.Has(utils.MetaChargers) {
		chrgS = flgs.GetBool(utils.MetaChargers)
	}
	attrS := len(cdrS.cgrCfg.CdrsCfg().AttributeSConns) != 0
	if flgs.Has(utils.MetaAttributes) {
		attrS = flgs.GetBool(utils.MetaAttributes)
	}

	if chrgS && len(cdrS.cgrCfg.CdrsCfg().ChargerSConns) == 0 {
		return utils.NewErrNotConnected(utils.ChargerS)
	}
	for _, cdr := range cdrs {
		cdr.Cost = -1 // the cost will be recalculated
		cgrEv := cdr.AsCGREvent()
		cgrEv.APIOpts = arg.APIOpts
		if _, err = cdrS.processEvent(cgrEv, chrgS, attrS, false,
			true, store, true, export, thdS, statS); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	*reply = utils.OK
	return
}

// V1ProcessExternalCDR is used to process external CDRs
func (cdrS *CDRServer) V1ProcessExternalCDR(eCDR *ExternalCDRWithAPIOpts, reply *string) error {
	cdr, err := NewCDRFromExternalCDR(eCDR.ExternalCDR,
		cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		return err
	}
	return cdrS.V1ProcessCDR(&CDRWithAPIOpts{
		CDR:     cdr,
		APIOpts: eCDR.APIOpts,
	}, reply)
}

// V1GetCDRs returns CDRs from DB
func (cdrS *CDRServer) V1GetCDRs(args utils.RPCCDRsFilterWithAPIOpts, cdrs *[]*CDR) error {
	cdrsFltr, err := args.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		if err.Error() != utils.NotFoundCaps {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	qryCDRs, _, err := cdrS.cdrDB.GetCDRs(cdrsFltr, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*cdrs = qryCDRs
	return nil
}

// V1CountCDRs counts CDRs from DB
func (cdrS *CDRServer) V1CountCDRs(args *utils.RPCCDRsFilterWithAPIOpts, cnt *int64) error {
	cdrsFltr, err := args.AsCDRsFilter(cdrS.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		if err.Error() != utils.NotFoundCaps {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	cdrsFltr.Count = true
	_, qryCnt, err := cdrS.cdrDB.GetCDRs(cdrsFltr, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*cnt = qryCnt
	return nil
}
