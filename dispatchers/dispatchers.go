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

package dispatchers

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewDispatcherService constructs a DispatcherService
func NewDispatcherService(dm *engine.DataManager,
	cfg *config.CGRConfig, fltrS *engine.FilterS,
	connMgr *engine.ConnManager) *DispatcherService {
	return &DispatcherService{
		dm:      dm,
		cfg:     cfg,
		fltrS:   fltrS,
		connMgr: connMgr,
	}
}

// DispatcherService  is the service handling dispatching towards internal components
// designed to handle automatic partitioning and failover
type DispatcherService struct {
	dm      *engine.DataManager
	cfg     *config.CGRConfig
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager
}

// Shutdown is called to shutdown the service
func (dS *DispatcherService) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.DispatcherS))
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.DispatcherS))
}

func (dS *DispatcherService) authorizeEvent(ev *utils.CGREvent,
	reply *engine.AttrSProcessEventReply) (err error) {
	ev.APIOpts[utils.OptsContext] = utils.MetaAuth
	if err = dS.connMgr.Call(context.TODO(), dS.cfg.DispatcherSCfg().AttributeSConns,
		utils.AttributeSv1ProcessEvent, ev, reply); err != nil {
		if err.Error() == utils.ErrNotFound.Error() {
			err = utils.ErrUnknownApiKey
		}
		return
	}
	return
}

func (dS *DispatcherService) authorize(method, tenant string, apiKey string, evTime *time.Time) (err error) {
	if apiKey == "" {
		return utils.NewErrMandatoryIeMissing(utils.APIKey)
	}
	ev := &utils.CGREvent{
		Tenant: tenant,
		ID:     utils.UUIDSha1Prefix(),
		Time:   evTime,
		Event: map[string]any{
			utils.APIKey: apiKey,
		},
		APIOpts: map[string]any{utils.MetaSubsys: utils.MetaDispatchers},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err = dS.authorizeEvent(ev, &rplyEv); err != nil {
		return
	}
	var apiMethods string
	if apiMethods, err = rplyEv.CGREvent.FieldAsString(utils.APIMethods); err != nil {
		return
	}
	if !ParseStringSet(apiMethods).Has(method) {
		return utils.ErrUnauthorizedApi
	}
	return
}

// dispatcherForEvent returns a dispatcher instance configured for specific event
// or utils.ErrNotFound if none present
func (dS *DispatcherService) dispatcherProfilesForEvent(tnt string, ev *utils.CGREvent,
	evNm utils.MapStorage, subsys string) (dPrlfs engine.DispatcherProfiles, err error) {
	// make sure dispatching is allowed
	var shouldDispatch bool
	if shouldDispatch, err = utils.GetBoolOpts(ev, true, utils.MetaDispatchers); err != nil {
		return
	} else {
		var subsys string
		if subsys, err = evNm.FieldAsString([]string{utils.MetaOpts, utils.MetaSubsys}); err != nil &&
			err != utils.ErrNotFound {
			return
		}
		if !shouldDispatch || (dS.cfg.DispatcherSCfg().PreventLoop &&
			subsys == utils.MetaDispatchers) {
			return engine.DispatcherProfiles{
				&engine.DispatcherProfile{Tenant: utils.MetaInternal, ID: utils.MetaInternal}}, nil
		}
	}
	// find out the matching profiles
	anyIdxPrfx := utils.ConcatenatedKey(tnt, utils.MetaAny)
	idxKeyPrfx := anyIdxPrfx
	if subsys != "" {
		idxKeyPrfx = utils.ConcatenatedKey(tnt, subsys)
	}
	var prflIDs utils.StringSet
	if prflIDs, err = engine.MatchingItemIDsForEvent(evNm,
		dS.cfg.DispatcherSCfg().StringIndexedFields,
		dS.cfg.DispatcherSCfg().PrefixIndexedFields,
		dS.cfg.DispatcherSCfg().SuffixIndexedFields,
		dS.cfg.DispatcherSCfg().ExistsIndexedFields,
		dS.dm, utils.CacheDispatcherFilterIndexes, idxKeyPrfx,
		dS.cfg.DispatcherSCfg().IndexedSelects,
		dS.cfg.DispatcherSCfg().NestedFields,
	); err != nil &&
		err != utils.ErrNotFound {
		return
	}
	if err == utils.ErrNotFound ||
		dS.cfg.DispatcherSCfg().AnySubsystem {
		var dPrflAnyIDs utils.StringSet
		if dPrflAnyIDs, err = engine.MatchingItemIDsForEvent(evNm,
			dS.cfg.DispatcherSCfg().StringIndexedFields,
			dS.cfg.DispatcherSCfg().PrefixIndexedFields,
			dS.cfg.DispatcherSCfg().SuffixIndexedFields,
			dS.cfg.DispatcherSCfg().ExistsIndexedFields,
			dS.dm, utils.CacheDispatcherFilterIndexes, anyIdxPrfx,
			dS.cfg.DispatcherSCfg().IndexedSelects,
			dS.cfg.DispatcherSCfg().NestedFields,
		); prflIDs.Size() == 0 {
			if err != nil { // return the error if no dispatcher matched the needed subsystem
				return
			}
			prflIDs = dPrflAnyIDs
		} else if err == nil && dPrflAnyIDs.Size() != 0 {
			prflIDs = utils.JoinStringSet(prflIDs, dPrflAnyIDs)
		}
		err = nil // make sure we ignore the error from *any subsystem matching
	}
	dPrlfs = make(engine.DispatcherProfiles, 0, len(prflIDs))
	for prflID := range prflIDs {
		prfl, err := dS.dm.GetDispatcherProfile(tnt, prflID, true, true, utils.NonTransactional)
		if err != nil {
			if err != utils.ErrDSPProfileNotFound {
				return nil, err
			}
			continue
		}

		if ((len(prfl.Subsystems) != 1 || prfl.Subsystems[0] != utils.MetaAny) &&
			!slices.Contains(prfl.Subsystems, subsys)) ||
			(prfl.ActivationInterval != nil && ev.Time != nil &&
				!prfl.ActivationInterval.IsActiveAtTime(*ev.Time)) { // not active
			continue
		}
		if pass, err := dS.fltrS.Pass(tnt, prfl.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		dPrlfs = append(dPrlfs, prfl)
	}
	if len(dPrlfs) == 0 {
		err = utils.ErrNotFound
		return
	}
	prfCount := len(dPrlfs) // if the option is not present return for all profiles
	if prfCountOpt, err := ev.OptAsInt64(utils.OptsDispatchersProfilesCount); err != nil {
		if err != utils.ErrNotFound { // is an conversion error
			return nil, err
		}
	} else if prfCount > int(prfCountOpt) { // it has the option and is smaller that the current number of profiles
		prfCount = int(prfCountOpt)
	}
	dPrlfs.Sort()
	dPrlfs = dPrlfs[:prfCount]
	return
}

// Dispatch is the method forwarding the request towards the right connection
func (dS *DispatcherService) Dispatch(ev *utils.CGREvent, subsys string,
	serviceMethod string, args any, reply any) (err error) {
	tnt := ev.Tenant
	if tnt == utils.EmptyString {
		tnt = dS.cfg.GeneralCfg().DefaultTenant
	}
	if ev.APIOpts == nil {
		ev.APIOpts = make(map[string]any)
	}
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.MetaSubsys: subsys,
			utils.MetaMethod: serviceMethod,
		},
	}
	dspLoopAPIOpts := map[string]any{
		utils.MetaSubsys: utils.MetaDispatchers,
		utils.MetaNodeID: dS.cfg.GeneralCfg().NodeID,
	}
	// avoid further processing if the request is internal
	var shouldDispatch bool
	if shouldDispatch, err = utils.GetBoolOpts(ev, true, utils.MetaDispatchers); err != nil {
		return utils.NewErrDispatcherS(err)
	} else {
		var subsys string
		if subsys, err = evNm.FieldAsString([]string{utils.MetaOpts, utils.MetaSubsys}); err != nil &&
			err != utils.ErrNotFound {
			return
		}
		if !shouldDispatch || (dS.cfg.DispatcherSCfg().PreventLoop &&
			subsys == utils.MetaDispatchers) {
			return callDH(newInternalHost(tnt), utils.EmptyString, nil,
				serviceMethod, args, reply)
		}
	}
	// in case of routeID, route based on previously discovered profile
	var dR *DispatcherRoute
	var dPrfls engine.DispatcherProfiles
	routeID := utils.IfaceAsString(ev.APIOpts[utils.OptsRouteID])
	if routeID != utils.EmptyString { // overwrite routeID with RouteID:Subsystem for subsystem correct routing
		routeID = utils.ConcatenatedKey(routeID, subsys)
		guardID := utils.ConcatenatedKey(utils.DispatcherSv1, utils.OptsRouteID, routeID)
		refID := guardian.Guardian.GuardIDs("", dS.cfg.GeneralCfg().LockingTimeout,
			guardID) // lock the routeID so we can make sure we have time to execute only once before caching
		defer guardian.Guardian.UnguardIDs(refID)
		// use previously discovered route
		argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant:  ev.Tenant,
			APIOpts: dspLoopAPIOpts,
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheDispatcherRoutes,
				ItemID:  routeID,
			}}
		var itmRemote any
		if itmRemote, err = engine.Cache.GetWithRemote(argsCache); err == nil && itmRemote != nil {
			var canCast bool
			if dR, canCast = itmRemote.(*DispatcherRoute); !canCast {
				err = utils.ErrCastFailed
			} else {
				var d Dispatcher
				if d, err = getDispatcherWithCache(
					&engine.DispatcherProfile{Tenant: dR.Tenant, ID: dR.ProfileID},
					dS.dm); err == nil {
					for k, v := range dspLoopAPIOpts {
						ev.APIOpts[k] = v // dispatcher loop protection opts
					}
					if err = d.Dispatch(dS.dm, dS.fltrS, evNm, tnt, utils.EmptyString, dR,
						serviceMethod, args, reply); !rpcclient.ShouldFailover(err) {
						return // dispatch success or specific error coming from upstream
					}
				}
			}
		}
		if err != nil {
			// did not dispatch properly, fail-back to standard dispatching
			utils.Logger.Warning(fmt.Sprintf("<%s> error <%s> using cached routing for dR %+v, continuing with normal dispatching",
				utils.DispatcherS, err.Error(), dR))
		}
	}
	if dPrfls, err = dS.dispatcherProfilesForEvent(tnt, ev, evNm, subsys); err != nil {
		return utils.NewErrDispatcherS(err)
	} else if len(dPrfls) == 0 { // no profiles matched
		return utils.NewErrDispatcherS(utils.ErrPrefixNotFound("PROFILE"))
	} else if isInternalDispatcherProfile(dPrfls[0]) { // dispatcherS was disabled
		return callDH(newInternalHost(tnt), utils.EmptyString, nil,
			serviceMethod, args, reply)
	}
	if ev.APIOpts == nil {
		ev.APIOpts = make(map[string]any)
	}
	ev.APIOpts[utils.MetaSubsys] = utils.MetaDispatchers // inject into args
	ev.APIOpts[utils.MetaNodeID] = dS.cfg.GeneralCfg().NodeID
	for _, dPrfl := range dPrfls {
		// get or build the Dispatcher for the config
		var d Dispatcher
		if d, err = getDispatcherWithCache(dPrfl, dS.dm); err == nil {
			if err = d.Dispatch(dS.dm, dS.fltrS, evNm, tnt, routeID,
				&DispatcherRoute{
					Tenant:    dPrfl.Tenant,
					ProfileID: dPrfl.ID,
				},
				serviceMethod, args, reply); !rpcclient.ShouldFailover(err) {
				return
			}
		}
		utils.Logger.Warning(fmt.Sprintf("<%s> error <%s> dispatching with the profile: <%+v>",
			utils.DispatcherS, err.Error(), dPrfl))
	}
	return // return the last error
}

func (dS *DispatcherService) DispatcherSv1GetProfilesForEvent(ctx *context.Context, ev *utils.CGREvent,
	dPfl *engine.DispatcherProfiles) (err error) {
	tnt := ev.Tenant
	if tnt == utils.EmptyString {
		tnt = dS.cfg.GeneralCfg().DefaultTenant
	}
	retDPfl, errDpfl := dS.dispatcherProfilesForEvent(tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.MetaMethod: ev.APIOpts[utils.MetaMethod],
		},
	}, utils.IfaceAsString(ev.APIOpts[utils.MetaSubsys]))
	if errDpfl != nil {
		return utils.NewErrDispatcherS(errDpfl)
	}
	*dPfl = retDPfl
	return
}

/*
// V1Apier is a generic way to cover all APIer methods
func (dS *DispatcherService) V1Apier(ctx *context.Context,apier any, args *utils.MethodParameters, reply *any) (err error) {

	parameters, canCast := args.Parameters.(map[string]any)
	if !canCast {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}

	var argD *utils.ArgDispatcher
	//check if we have APIKey in event and in case it has add it in ArgDispatcher
	apiKeyIface, hasAPIKey := parameters[utils.APIKey]
	if hasAPIKey && apiKeyIface != nil {
		argD = &utils.ArgDispatcher{
			APIKey: utils.StringPointer(apiKeyIface.(string)),
		}
	}
	//check if we have RouteID in event and in case it has add it in ArgDispatcher
	routeIDIface, hasRouteID := parameters[utils.RouteID]
	if hasRouteID && routeIDIface != nil {
		if !hasAPIKey || apiKeyIface == nil { //in case we don't have APIKey, but we have RouteID we need to initialize the struct
			argD = &utils.ArgDispatcher{
				RouteID: utils.StringPointer(routeIDIface.(string)),
			}
		} else {
			argD.RouteID = utils.StringPointer(routeIDIface.(string))
		}
	}

	tenant := utils.FirstNonEmpty(utils.IfaceAsString(parameters[utils.Tenant]), config.CgrConfig().GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if argD == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(args.Method,
			tenant,
			argD.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	// split the method
	methodSplit := strings.Split(args.Method, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	method := reflect.ValueOf(apier).MethodByName(methodSplit[1])
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// take the arguments (args + reply)
	methodType := method.Type()
	if methodType.NumIn() != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// convert type of reply to the right one based on method
	realReplyType := methodType.In(1)

	var realReply any
	if realReplyType.Kind() == reflect.Ptr {
		trply := reflect.New(realReplyType.Elem()).Elem().Interface()
		realReply = &trply
	} else {
		realReply = reflect.New(realReplyType).Elem().Interface()
	}
	//convert parameters so we can unmarshal the informations into to right struct
	argsByte, err := json.Marshal(parameters)
	if err != nil {
		return err
	}
	// find the type for arg
	realArgsType := methodType.In(0)
	// create the arg with the right type for method
	var realArgs any = reflect.New(realArgsType).Interface()
	// populate realArgs with data
	if err := json.Unmarshal(argsByte, &realArgs); err != nil {
		return err
	}
	if realArgsType.Kind() != reflect.Ptr {
		realArgs = reflect.ValueOf(realArgs).Elem().Interface()
	}

	var routeID *string
	if argD != nil {
		routeID = argD.RouteID
	}
	if err := dS.Dispatch(&utils.CGREvent{Tenant: tenant, Event: parameters}, utils.MetaApier, routeID,
		args.Method, realArgs, realReply); err != nil {
		return err
	}
	*reply = realReply
	return nil

}
*/

// Call implements birpc.ClientConnector interface for internal RPC
func (dS *DispatcherService) Call(ctx *context.Context, serviceMethod string, // all API fuction must be of type: SubsystemMethod
	args any, reply any) error {
	methodSplit := strings.Split(serviceMethod, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	method := reflect.ValueOf(dS).MethodByName(methodSplit[0] + methodSplit[1])
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}

func (dS *DispatcherService) DispatcherSv1RemoteStatus(ctx *context.Context, args *cores.V1StatusParams,
	reply *map[string]any) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1Status, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaCore, utils.CoreSv1Status, args, reply)
}

func (dS *DispatcherService) DispatcherSv1RemoteSleep(ctx *context.Context, args *utils.DurationArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1Sleep, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaCore, utils.CoreSv1Sleep, args, reply)
}

func (dS *DispatcherService) DispatcherSv1RemotePing(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1Ping, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaCore, utils.CoreSv1Ping, args, reply)
}
