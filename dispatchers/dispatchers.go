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
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
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
	if err = dS.connMgr.Call(dS.cfg.DispatcherSCfg().AttributeSConns, nil,
		utils.AttributeSv1ProcessEvent,
		&engine.AttrArgsProcessEvent{
			CGREvent: ev,
			Context:  utils.StringPointer(utils.MetaAuth),
		}, reply); err != nil {
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
		Event: map[string]interface{}{
			utils.APIKey: apiKey,
		},
		APIOpts: map[string]interface{}{utils.Subsys: utils.MetaDispatchers},
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
func (dS *DispatcherService) dispatcherProfileForEvent(tnt string, ev *utils.CGREvent,
	subsys string) (dPrlf *engine.DispatcherProfile, err error) {
	// find out the matching profiles
	anyIdxPrfx := utils.ConcatenatedKey(tnt, utils.MetaAny)
	idxKeyPrfx := anyIdxPrfx
	if subsys != "" {
		idxKeyPrfx = utils.ConcatenatedKey(tnt, subsys)
	}
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	prflIDs, err := engine.MatchingItemIDsForEvent(evNm,
		dS.cfg.DispatcherSCfg().StringIndexedFields,
		dS.cfg.DispatcherSCfg().PrefixIndexedFields,
		dS.cfg.DispatcherSCfg().SuffixIndexedFields,
		dS.dm, utils.CacheDispatcherFilterIndexes, idxKeyPrfx,
		dS.cfg.DispatcherSCfg().IndexedSelects,
		dS.cfg.DispatcherSCfg().NestedFields,
	)
	if err != nil &&
		err != utils.ErrNotFound {
		return nil, err
	}
	if err == utils.ErrNotFound ||
		dS.cfg.DispatcherSCfg().AnySubsystem {
		dPrflAnyIDs, err := engine.MatchingItemIDsForEvent(evNm,
			dS.cfg.DispatcherSCfg().StringIndexedFields,
			dS.cfg.DispatcherSCfg().PrefixIndexedFields,
			dS.cfg.DispatcherSCfg().SuffixIndexedFields,
			dS.dm, utils.CacheDispatcherFilterIndexes, anyIdxPrfx,
			dS.cfg.DispatcherSCfg().IndexedSelects,
			dS.cfg.DispatcherSCfg().NestedFields,
		)
		if prflIDs.Size() == 0 {
			if err != nil { // return the error if no dispatcher matched the needed subsystem
				return nil, err
			}
			prflIDs = dPrflAnyIDs
		} else if err == nil && dPrflAnyIDs.Size() != 0 {
			prflIDs = utils.JoinStringSet(prflIDs, dPrflAnyIDs)
		}
	}
	for prflID := range prflIDs {
		prfl, err := dS.dm.GetDispatcherProfile(tnt, prflID, true, true, utils.NonTransactional)
		if err != nil {
			if err != utils.ErrNotFound {
				return nil, err
			}
			continue
		}
		if !(len(prfl.Subsystems) == 1 && prfl.Subsystems[0] == utils.MetaAny) &&
			!utils.IsSliceMember(prfl.Subsystems, subsys) {
			continue
		}
		if prfl.ActivationInterval != nil && ev.Time != nil &&
			!prfl.ActivationInterval.IsActiveAtTime(*ev.Time) { // not active
			continue
		}
		if pass, err := dS.fltrS.Pass(tnt, prfl.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		if dPrlf == nil || prfl.Weight > dPrlf.Weight {
			dPrlf = prfl
		}
	}
	if dPrlf == nil {
		return nil, utils.ErrNotFound
	}
	return
}

// Dispatch is the method forwarding the request towards the right connection
func (dS *DispatcherService) Dispatch(ev *utils.CGREvent, subsys string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	tnt := ev.Tenant
	if tnt == utils.EmptyString {
		tnt = dS.cfg.GeneralCfg().DefaultTenant
	}
	dPrfl, errDsp := dS.dispatcherProfileForEvent(tnt, ev, subsys)
	if errDsp != nil {
		return utils.NewErrDispatcherS(errDsp)
	}
	tntID := dPrfl.TenantID()
	// get or build the Dispatcher for the config
	var d Dispatcher
	if x, ok := engine.Cache.Get(utils.CacheDispatchers,
		tntID); ok && x != nil {
		d = x.(Dispatcher)
	} else if d, err = newDispatcher(dS.dm, dPrfl); err != nil {
		return utils.NewErrDispatcherS(err)
	}
	if errCh := engine.Cache.Set(utils.CacheDispatchers, tntID, d, nil, true, utils.EmptyString); errCh != nil {
		return utils.NewErrDispatcherS(errCh)
	}
	return d.Dispatch(utils.IfaceAsString(ev.APIOpts[utils.OptsRouteID]), subsys, serviceMethod, args, reply)
}

func (dS *DispatcherService) V1GetProfileForEvent(ev *utils.CGREvent,
	dPfl *engine.DispatcherProfile) (err error) {
	tnt := ev.Tenant
	if tnt == utils.EmptyString {
		tnt = dS.cfg.GeneralCfg().DefaultTenant
	}
	retDPfl, errDpfl := dS.dispatcherProfileForEvent(tnt, ev, utils.IfaceAsString(ev.APIOpts[utils.Subsys]))
	if errDpfl != nil {
		return utils.NewErrDispatcherS(errDpfl)
	}
	*dPfl = *retDPfl
	return
}

/*
// V1Apier is a generic way to cover all APIer methods
func (dS *DispatcherService) V1Apier(apier interface{}, args *utils.MethodParameters, reply *interface{}) (err error) {

	parameters, canCast := args.Parameters.(map[string]interface{})
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

	var realReply interface{}
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
	var realArgs interface{} = reflect.New(realArgsType).Interface()
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

// Call implements rpcclient.ClientConnector interface for internal RPC
func (dS *DispatcherService) Call(serviceMethod string, // all API fuction must be of type: SubsystemMethod
	args interface{}, reply interface{}) error {
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
