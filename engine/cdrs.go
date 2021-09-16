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

// NewCDRServer is a constructor for CDRServer
func NewCDRServer(cfg *config.CGRConfig, storDBChan chan StorDB, dm *DataManager, filterS *FilterS,
	connMgr *ConnManager) *CDRServer {
	cdrDB := <-storDBChan
	return &CDRServer{
		cfg:        cfg,
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
	cfg        *config.CGRConfig
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

// chrgrSProcessEvent forks CGREventWithOpts into multiples based on matching ChargerS profiles
func (cdrS *CDRServer) chrgrSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent) (cgrEvs []*utils.CGREvent, err error) {
	var chrgrs []*ChrgSProcessEventReply
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().ChargerSConns,
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
func (cdrS *CDRServer) attrSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent) (err error) {
	var rplyEv AttrSProcessEventReply
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(map[string]interface{})
	}
	cgrEv.APIOpts[utils.Subsys] = utils.MetaCDRs
	cgrEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
		utils.IfaceAsString(cgrEv.APIOpts[utils.OptsContext]),
		utils.MetaCDRs)

	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().AttributeSConns,
		utils.AttributeSv1ProcessEvent,
		cgrEv, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
		*cgrEv = *rplyEv.CGREvent
	} else if err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // cancel ErrNotFound
	}
	return
}

// rateSProcessEvent will send the event to rateS and return the result
func (cdrS *CDRServer) rateSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent) (rpCost utils.RateProfileCost, err error) {
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().RateSConns,
		utils.RateSv1CostForEvent, cgrEv, &rpCost); err != nil {
		return
	}
	return
}

// thdSProcessEvent will send the event to ThresholdS
func (cdrS *CDRServer) thdSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent) (err error) {
	var tIDs []string
	// we clone the CGREvent so we can add EventType without being propagated
	thArgs := &ThresholdsArgsProcessEvent{
		CGREvent: cgrEv.Clone(),
	}
	if thArgs.APIOpts == nil {
		thArgs.APIOpts = make(map[string]interface{})
	}
	thArgs.APIOpts[utils.MetaEventType] = utils.CDR
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().ThresholdSConns,
		utils.ThresholdSv1ProcessEvent,
		thArgs, &tIDs); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// statSProcessEvent will send the event to StatS
func (cdrS *CDRServer) statSProcessEvent(ctx *context.Context, cgrEv *utils.CGREvent) (err error) {
	var reply []string
	statArgs := &StatsArgsProcessEvent{
		CGREvent: cgrEv.Clone(),
	}
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().StatSConns,
		utils.StatSv1ProcessEvent,
		statArgs, &reply); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// eeSProcessEvent will process the event with the EEs component
func (cdrS *CDRServer) eeSProcessEvent(ctx *context.Context, cgrEv *utils.CGREventWithEeIDs) (err error) {
	var reply map[string]map[string]interface{}
	if err = cdrS.connMgr.Call(ctx, cdrS.cfg.CdrsCfg().EEsConns,
		utils.EeSv1ProcessEvent,
		cgrEv, &reply); err != nil &&
		err.Error() == utils.ErrNotFound.Error() {
		err = nil // NotFound is not considered error
	}
	return
}

// processEvent processes a CGREvent based on arguments
// in case of partially executed, both error and evs will be returned
func (cdrS *CDRServer) processEvent(ctx *context.Context, ev *utils.CGREvent,
	chrgS, attrS, rateS, eeS, thdS, stS bool) (evs []*utils.EventWithFlags, err error) {
	if attrS {
		if err = cdrS.attrSProcessEvent(ctx, ev); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
					utils.CDRs, err.Error(), utils.ToJSON(ev), utils.AttributeS))
			err = utils.ErrPartiallyExecuted
			return
		}
	}
	var cgrEvs []*utils.CGREvent
	if chrgS {
		if cgrEvs, err = cdrS.chrgrSProcessEvent(ctx, ev); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
					utils.CDRs, err.Error(), utils.ToJSON(ev), utils.ChargerS))
			err = utils.ErrPartiallyExecuted
			return
		}
	} else { // ChargerS not requested, charge the original event
		cgrEvs = []*utils.CGREvent{ev}
	}

	var partiallyExecuted bool // from here actions are optional and a general error is returned

	if rateS {
		for _, cgrEv := range cgrEvs {
			if rtsEvCost, err := cdrS.rateSProcessEvent(ctx, ev); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(ev), utils.RateS))
				partiallyExecuted = true
			} else {
				cgrEv.Event[utils.MetaRateSCost] = rtsEvCost
			}
		}
	}

	if eeS {
		if len(cdrS.cfg.CdrsCfg().EEsConns) != 0 {
			for _, cgrEv := range cgrEvs {
				evWithOpts := &utils.CGREventWithEeIDs{
					CGREvent: cgrEv,
					EeIDs:    cdrS.cfg.CdrsCfg().OnlineCDRExports,
				}
				if err = cdrS.eeSProcessEvent(ctx, evWithOpts); err != nil {
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
			if err = cdrS.thdSProcessEvent(ctx, cgrEv); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: <%s> processing event %+v with %s",
						utils.CDRs, err.Error(), utils.ToJSON(cgrEv), utils.ThresholdS))
				partiallyExecuted = true
			}
		}
	}
	if stS {
		for _, cgrEv := range cgrEvs {
			if err = cdrS.statSProcessEvent(ctx, cgrEv); err != nil {
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
	return
}

// ArgV1ProcessEvent is the CGREvent with proccesing Flags
type ArgV1ProcessEvent struct {
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
	return &ArgV1ProcessEvent{
		CGREvent: *attr.CGREvent.Clone(),
	}
}

// V1ProcessEvent will process the CGREvent
func (cdrS *CDRServer) V1ProcessEvent(ctx *context.Context, arg *ArgV1ProcessEvent, reply *string) (err error) {
	if arg.CGREvent.ID == utils.EmptyString {
		arg.CGREvent.ID = utils.GenUUID()
	}
	if arg.CGREvent.Tenant == utils.EmptyString {
		arg.CGREvent.Tenant = cdrS.cfg.GeneralCfg().DefaultTenant
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
		defer Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	// processing options
	attrS := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsAttributeS, len(cdrS.cfg.CdrsCfg().AttributeSConns) != 0)
	export := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsExport, len(cdrS.cfg.CdrsCfg().OnlineCDRExports) != 0 ||
		len(cdrS.cfg.CdrsCfg().EEsConns) != 0)
	thdS := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsThresholdS, len(cdrS.cfg.CdrsCfg().ThresholdSConns) != 0)
	stS := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsStatS, len(cdrS.cfg.CdrsCfg().ThresholdSConns) != 0)
	chrgS := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsChargerS, len(cdrS.cfg.CdrsCfg().ThresholdSConns) != 0)
	rateS := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsRateS, len(cdrS.cfg.CdrsCfg().RateSConns) != 0)

	// end of processing options

	if _, err = cdrS.processEvent(ctx, &arg.CGREvent,
		chrgS, attrS, rateS, export, thdS, stS); err != nil {
		return
	}
	*reply = utils.OK
	return nil
}

// V1ProcessEventWithGet has the same logic with V1ProcessEvent except it adds the proccessed events to the reply
func (cdrS *CDRServer) V1ProcessEventWithGet(ctx *context.Context, arg *ArgV1ProcessEvent, evs *[]*utils.EventWithFlags) (err error) {
	if arg.ID == "" {
		arg.ID = utils.GenUUID()
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
				*evs = *cachedResp.Result.(*[]*utils.EventWithFlags)
			}
			return cachedResp.Error
		}
		defer Cache.Set(ctx, utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: evs, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	// processing options
	attrS := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsAttributeS, len(cdrS.cfg.CdrsCfg().AttributeSConns) != 0)
	export := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsExport, len(cdrS.cfg.CdrsCfg().OnlineCDRExports) != 0 ||
		len(cdrS.cfg.CdrsCfg().EEsConns) != 0)
	thdS := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsThresholdS, len(cdrS.cfg.CdrsCfg().ThresholdSConns) != 0)
	stS := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsStatS, len(cdrS.cfg.CdrsCfg().ThresholdSConns) != 0)
	chrgS := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsChargerS, len(cdrS.cfg.CdrsCfg().ThresholdSConns) != 0)
	rateS := utils.OptAsBoolOrDef(arg.APIOpts, utils.OptsCDRsRateS, len(cdrS.cfg.CdrsCfg().RateSConns) != 0)
	// end of processing options

	var procEvs []*utils.EventWithFlags
	if procEvs, err = cdrS.processEvent(ctx, &arg.CGREvent,
		chrgS, attrS, rateS, export, thdS, stS); err != nil {
		return
	}
	*evs = procEvs
	return nil
}
