//go:generate go run ../data/scripts/generate_dispatchers/generator.go
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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
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

func (dS *DispatcherService) authorizeEvent(ctx *context.Context, ev *utils.CGREvent,
	reply *engine.AttrSProcessEventReply) (err error) {
	if err = dS.connMgr.Call(ctx, dS.cfg.DispatcherSCfg().AttributeSConns,
		utils.AttributeSv1ProcessEvent, ev, reply); err != nil {
		if err.Error() == utils.ErrNotFound.Error() {
			err = utils.ErrUnknownApiKey
		}
		return
	}
	return
}

func (dS *DispatcherService) authorize(ctx *context.Context, method, tenant string, apiKey string) (err error) {
	if apiKey == "" {
		return utils.NewErrMandatoryIeMissing(utils.APIKey)
	}
	ev := &utils.CGREvent{
		Tenant: tenant,
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			utils.APIKey: apiKey,
		},
		APIOpts: map[string]any{
			utils.MetaSubsys:  utils.MetaDispatchers,
			utils.OptsContext: utils.MetaAuth,
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err = dS.authorizeEvent(ctx, ev, &rplyEv); err != nil {
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
func (dS *DispatcherService) dispatcherProfilesForEvent(ctx *context.Context, tnt string, ev *utils.CGREvent,
	evNm utils.MapStorage) (dPrlfs engine.DispatcherProfiles, err error) {
	// make sure dispatching is allowed
	var shouldDispatch bool
	if shouldDispatch, err = engine.GetBoolOpts(ctx, tnt, evNm, dS.fltrS, dS.cfg.DispatcherSCfg().Opts.Dispatchers,
		config.DispatchersDispatchersDftOpt, utils.MetaDispatchers); err != nil {
		return
	} else if !shouldDispatch {
		return engine.DispatcherProfiles{
			&engine.DispatcherProfile{Tenant: utils.MetaInternal, ID: utils.MetaInternal}}, nil
	}
	// find out the matching profiles
	var prflIDs utils.StringSet
	if prflIDs, err = engine.MatchingItemIDsForEvent(ctx, evNm,
		dS.cfg.DispatcherSCfg().StringIndexedFields,
		dS.cfg.DispatcherSCfg().PrefixIndexedFields,
		dS.cfg.DispatcherSCfg().SuffixIndexedFields,
		dS.cfg.DispatcherSCfg().ExistsIndexedFields,
		dS.cfg.DispatcherSCfg().NotExistsIndexedFields,
		dS.dm, utils.CacheDispatcherFilterIndexes, tnt,
		dS.cfg.DispatcherSCfg().IndexedSelects,
		dS.cfg.DispatcherSCfg().NestedFields,
	); err != nil {
		return
	}
	for prflID := range prflIDs {
		prfl, err := dS.dm.GetDispatcherProfile(ctx, tnt, prflID, true, true, utils.NonTransactional)
		if err != nil {
			if err != utils.ErrDSPProfileNotFound {
				return nil, err
			}
			continue
		}

		if pass, err := dS.fltrS.Pass(ctx, tnt, prfl.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		dPrlfs = append(dPrlfs, prfl)
	}
	if len(dPrlfs) == 0 {
		err = utils.ErrDSPProfileNotFound
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
func (dS *DispatcherService) Dispatch(ctx *context.Context, ev *utils.CGREvent, subsys string,
	serviceMethod string, args, reply any) (err error) {
	tnt := ev.Tenant
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, serviceMethod, tnt, utils.IfaceAsString(ev.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
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
	if shouldDispatch, err = engine.GetBoolOpts(ctx, tnt, evNm, dS.fltrS, dS.cfg.DispatcherSCfg().Opts.Dispatchers,
		true, utils.MetaDispatchers); err != nil {
		return utils.NewErrDispatcherS(err)
	} else if !shouldDispatch {
		return callDH(ctx,
			newInternalHost(tnt), utils.EmptyString, nil,
			dS.cfg, dS.connMgr.GetDispInternalChan(),
			serviceMethod, args, reply)
	}

	// in case of routeID, route based on previously discovered profile
	var dR *DispatcherRoute
	var dPrfls engine.DispatcherProfiles
	routeID := utils.IfaceAsString(ev.APIOpts[utils.OptsRouteID])
	if routeID != utils.EmptyString { // overwrite routeID with RouteID:Subsystem for subsystem correct routing
		routeID = utils.ConcatenatedKey(routeID, subsys)
		guardID := utils.ConcatenatedKey(utils.DispatcherSv1, utils.OptsRouteID, routeID)
		refID := guardian.Guardian.GuardIDs("",
			dS.cfg.GeneralCfg().LockingTimeout, guardID) // lock the routeID so we can make sure we have time to execute only once before caching
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
		if itmRemote, err = engine.Cache.GetWithRemote(ctx, argsCache); err == nil && itmRemote != nil {
			var canCast bool
			if dR, canCast = itmRemote.(*DispatcherRoute); !canCast {
				err = utils.ErrCastFailed
			} else {
				var d Dispatcher
				if d, err = getDispatcherWithCache(ctx,
					&engine.DispatcherProfile{Tenant: dR.Tenant, ID: dR.ProfileID},
					dS.dm); err == nil {
					for k, v := range dspLoopAPIOpts {
						ev.APIOpts[k] = v // dispatcher loop protection opts
					}
					if err = d.Dispatch(dS.dm, dS.fltrS, dS.cfg,
						ctx, dS.connMgr.GetDispInternalChan(), evNm, tnt, utils.EmptyString, dR,
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
	if dPrfls, err = dS.dispatcherProfilesForEvent(ctx, tnt, ev, evNm); err != nil {
		return utils.NewErrDispatcherS(err)
	} else if len(dPrfls) == 0 { // no profiles matched
		return utils.ErrDSPProfileNotFound
	} else if isInternalDispatcherProfile(dPrfls[0]) { // dispatcherS was disabled
		return callDH(ctx,
			newInternalHost(tnt), utils.EmptyString, nil,
			dS.cfg, dS.connMgr.GetDispInternalChan(),
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
		if d, err = getDispatcherWithCache(ctx, dPrfl, dS.dm); err == nil {
			if err = d.Dispatch(dS.dm, dS.fltrS, dS.cfg,
				ctx, dS.connMgr.GetDispInternalChan(), evNm, tnt, routeID,
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

func (dS *DispatcherService) V1GetProfilesForEvent(ctx *context.Context, ev *utils.CGREvent,
	dPfl *engine.DispatcherProfiles) (err error) {
	tnt := ev.Tenant
	if tnt == utils.EmptyString {
		tnt = dS.cfg.GeneralCfg().DefaultTenant
	}
	retDPfl, errDpfl := dS.dispatcherProfilesForEvent(ctx, tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.MetaSubsys: ev.APIOpts[utils.MetaSubsys],
			utils.MetaMethod: ev.APIOpts[utils.MetaMethod],
		},
	})
	if errDpfl != nil {
		return utils.NewErrDispatcherS(errDpfl)
	}
	*dPfl = retDPfl
	return
}

func (dS *DispatcherService) ping(ctx *context.Context, subsys, method string, args *utils.CGREvent,
	reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, method, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, args, subsys, method, args, reply)
}

func (dS *DispatcherService) DispatcherSv1RemoteStatus(ctx *context.Context, args *utils.TenantWithAPIOpts,
	reply *map[string]any) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.CoreSv1Status, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{
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
		if err = dS.authorize(ctx, utils.CoreSv1Sleep, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{
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
		if err = dS.authorize(ctx, utils.CoreSv1Ping, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, args, utils.MetaCore, utils.CoreSv1Ping, args, reply)
}
