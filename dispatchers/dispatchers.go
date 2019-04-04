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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewDispatcherService constructs a DispatcherService
func NewDispatcherService(dm *engine.DataManager,
	cfg *config.CGRConfig, fltrS *engine.FilterS,
	attrS *rpcclient.RpcClientPool) (*DispatcherService, error) {
	if attrS != nil && reflect.ValueOf(attrS).IsNil() {
		attrS = nil
	}
	return &DispatcherService{dm: dm, cfg: cfg,
		fltrS: fltrS, attrS: attrS}, nil
}

// DispatcherService  is the service handling dispatching towards internal components
// designed to handle automatic partitioning and failover
type DispatcherService struct {
	dm    *engine.DataManager
	cfg   *config.CGRConfig
	fltrS *engine.FilterS
	attrS *rpcclient.RpcClientPool // used for API auth
}

// ListenAndServe will initialize the service
func (dS *DispatcherService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info("Starting Dispatcher service")
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Shutdown is called to shutdown the service
func (dS *DispatcherService) Shutdown() error {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.DispatcherS))
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.DispatcherS))
	return nil
}

func (dS *DispatcherService) authorizeEvent(ev *utils.CGREvent,
	reply *engine.AttrSProcessEventReply) (err error) {
	if err = dS.attrS.Call(utils.AttributeSv1ProcessEvent,
		&engine.AttrArgsProcessEvent{
			Context:  utils.StringPointer(utils.MetaAuth),
			CGREvent: *ev}, reply); err != nil {
		if err.Error() == utils.ErrNotFound.Error() {
			err = utils.ErrUnknownApiKey
		}
		return
	}
	return
}

func (dS *DispatcherService) authorize(method, tenant, apiKey string, evTime *time.Time) (err error) {
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
	}
	var rplyEv engine.AttrSProcessEventReply
	if err = dS.authorizeEvent(ev, &rplyEv); err != nil {
		return
	}
	var apiMethods string
	if apiMethods, err = rplyEv.CGREvent.FieldAsString(utils.APIMethods); err != nil {
		return
	}
	if !ParseStringMap(apiMethods).HasKey(method) {
		return utils.ErrUnauthorizedApi
	}
	return
}

// dispatcherForEvent returns a dispatcher instance configured for specific event
// or utils.ErrNotFound if none present
func (dS *DispatcherService) dispatcherProfileForEvent(ev *utils.CGREvent,
	subsys string) (dPrlf *engine.DispatcherProfile, err error) {
	// find out the matching profiles
	anyIdxPrfx := utils.ConcatenatedKey(ev.Tenant, utils.META_ANY)
	idxKeyPrfx := anyIdxPrfx
	if subsys != "" {
		idxKeyPrfx = utils.ConcatenatedKey(ev.Tenant, subsys)
	}
	prflIDs, err := engine.MatchingItemIDsForEvent(ev.Event,
		dS.cfg.DispatcherSCfg().StringIndexedFields,
		dS.cfg.DispatcherSCfg().PrefixIndexedFields,
		dS.dm, utils.CacheDispatcherFilterIndexes,
		idxKeyPrfx, dS.cfg.DispatcherSCfg().IndexedSelects)
	if err != nil {
		// return nil, err
		if err != utils.ErrNotFound {
			return nil, err
		}
		prflIDs, err = engine.MatchingItemIDsForEvent(ev.Event,
			dS.cfg.DispatcherSCfg().StringIndexedFields,
			dS.cfg.DispatcherSCfg().PrefixIndexedFields,
			dS.dm, utils.CacheDispatcherFilterIndexes,
			anyIdxPrfx, dS.cfg.DispatcherSCfg().IndexedSelects)
		if err != nil {
			return nil, err
		}
	}
	for prflID := range prflIDs {
		prfl, err := dS.dm.GetDispatcherProfile(ev.Tenant, prflID, true, true, utils.NonTransactional)
		if err != nil {
			if err != utils.ErrNotFound {
				return nil, err
			}
			continue
		}
		if prfl.ActivationInterval != nil && ev.Time != nil &&
			!prfl.ActivationInterval.IsActiveAtTime(*ev.Time) { // not active
			continue
		}
		if pass, err := dS.fltrS.Pass(ev.Tenant, prfl.FilterIDs,
			config.NewNavigableMap(ev.Event)); err != nil {
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
func (dS *DispatcherService) Dispatch(ev *utils.CGREvent, subsys string, routeID *string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	dPrfl, errDsp := dS.dispatcherProfileForEvent(ev, subsys)
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
	engine.Cache.Set(utils.CacheDispatchers, tntID, d, nil,
		true, utils.EmptyString)
	return d.Dispatch(routeID, serviceMethod, args, reply)
}

func (dS *DispatcherService) V1GetProfileForEvent(ev *DispatcherEvent,
	dPfl *engine.DispatcherProfile) (err error) {
	retDPfl, errDpfl := dS.dispatcherProfileForEvent(&ev.CGREvent, ev.Subsystem)
	if errDpfl != nil {
		return utils.NewErrDispatcherS(errDpfl)
	}
	*dPfl = *retDPfl
	return
}
