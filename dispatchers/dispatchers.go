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
	if err = dS.connMgr.Call(context.TODO(), dS.cfg.DispatcherSCfg().AttributeSConns,
		utils.AttributeSv1ProcessEvent,
		&engine.AttrArgsProcessEvent{
			CGREvent: ev,
		}, reply); err != nil {
		if err.Error() == utils.ErrNotFound.Error() {
			err = utils.ErrUnknownApiKey
		}
		return
	}
	return
}

func (dS *DispatcherService) authorize(method, tenant string, apiKey string) (err error) {
	if apiKey == "" {
		return utils.NewErrMandatoryIeMissing(utils.APIKey)
	}
	ev := &utils.CGREvent{
		Tenant: tenant,
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]interface{}{
			utils.APIKey: apiKey,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys:      utils.MetaDispatchers,
			utils.OptsContext: utils.MetaAuth,
		},
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
func (dS *DispatcherService) dispatcherProfilesForEvent(ctx *context.Context, tnt string, ev *utils.CGREvent,
	evNm utils.MapStorage) (dPrlfs engine.DispatcherProfiles, err error) {
	// find out the matching profiles
	var prflIDs utils.StringSet
	if prflIDs, err = engine.MatchingItemIDsForEvent(ctx, evNm,
		dS.cfg.DispatcherSCfg().StringIndexedFields,
		dS.cfg.DispatcherSCfg().PrefixIndexedFields,
		dS.cfg.DispatcherSCfg().SuffixIndexedFields,
		dS.dm, utils.CacheDispatcherFilterIndexes, tnt,
		dS.cfg.DispatcherSCfg().IndexedSelects,
		dS.cfg.DispatcherSCfg().NestedFields,
	); err != nil {
		return
	}
	for prflID := range prflIDs {
		prfl, err := dS.dm.GetDispatcherProfile(ctx, tnt, prflID, true, true, utils.NonTransactional)
		if err != nil {
			if err != utils.ErrNotFound {
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
func (dS *DispatcherService) Dispatch(ctx *context.Context, ev *utils.CGREvent, subsys string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	tnt := ev.Tenant
	if tnt == utils.EmptyString {
		tnt = dS.cfg.GeneralCfg().DefaultTenant
	}
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.Subsys:     subsys,
			utils.MetaMethod: serviceMethod,
		},
	}
	var dPrfls engine.DispatcherProfiles
	if dPrfls, err = dS.dispatcherProfilesForEvent(ctx, tnt, ev, evNm); err != nil {
		return utils.NewErrDispatcherS(err)
	}
	for _, dPrfl := range dPrfls {
		tntID := dPrfl.TenantID()
		// get or build the Dispatcher for the config
		var d Dispatcher
		if x, ok := engine.Cache.Get(utils.CacheDispatchers,
			tntID); ok && x != nil {
			d = x.(Dispatcher)
		} else if d, err = newDispatcher(dPrfl); err != nil {
			return utils.NewErrDispatcherS(err)
		}
		if err = engine.Cache.Set(ctx, utils.CacheDispatchers, tntID, d, nil, true, utils.EmptyString); err != nil {
			return utils.NewErrDispatcherS(err)
		}
		if err = d.Dispatch(dS.dm, dS.fltrS, ctx, evNm, tnt, utils.IfaceAsString(ev.APIOpts[utils.OptsRouteID]),
			subsys, serviceMethod, args, reply); !rpcclient.IsNetworkError(err) {
			return
		}
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
			utils.Subsys:     ev.APIOpts[utils.Subsys],
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
		if err = dS.authorize(method, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, args, subsys, method, args, reply)
}
