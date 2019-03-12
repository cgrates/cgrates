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
	attrS *rpcclient.RpcClientPool,
	conns map[string]*rpcclient.RpcClientPool) (*DispatcherService, error) {
	if attrS != nil && reflect.ValueOf(attrS).IsNil() {
		attrS = nil
	}
	return &DispatcherService{dm: dm, cfg: cfg,
		fltrS: fltrS, attrS: attrS, conns: conns}, nil
}

// DispatcherService  is the service handling dispatching towards internal components
// designed to handle automatic partitioning and failover
type DispatcherService struct {
	dm    *engine.DataManager
	cfg   *config.CGRConfig
	fltrS *engine.FilterS
	attrS *rpcclient.RpcClientPool            // used for API auth
	conns map[string]*rpcclient.RpcClientPool // available connections, accessed based on connID
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

// dispatcherForEvent returns a dispatcher instance configured for specific event
// or utils.ErrNotFound if none present
func (dS *DispatcherService) dispatcherForEvent(ev *utils.CGREvent,
	subsys string) (d Dispatcher, err error) {
	// find out the matching profiles
	anyIdxPrfx := utils.ConcatenatedKey(ev.Tenant, utils.META_ANY)
	idxKeyPrfx := anyIdxPrfx
	if subsys != "" {
		idxKeyPrfx = utils.ConcatenatedKey(ev.Tenant, subsys)
	}
	var matchedPrlf *engine.DispatcherProfile
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
		if matchedPrlf == nil || prfl.Weight > matchedPrlf.Weight {
			matchedPrlf = prfl
		}
	}
	if matchedPrlf == nil {
		return nil, utils.ErrNotFound
	}
	tntID := matchedPrlf.TenantID()
	// get or build the Dispatcher for the config
	if x, ok := engine.Cache.Get(utils.CacheDispatchers,
		tntID); ok && x != nil {
		d = x.(Dispatcher)
		return
	}
	if d, err = newDispatcher(matchedPrlf); err != nil {
		return
	}
	engine.Cache.Set(utils.CacheDispatchers, tntID, d, nil,
		true, utils.EmptyString)
	return
}

// Dispatch is the method forwarding the request towards the right connection
func (dS *DispatcherService) Dispatch(ev *utils.CGREvent, subsys string, routeID *string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	d, errDsp := dS.dispatcherForEvent(ev, subsys)
	if errDsp != nil {
		return utils.NewErrDispatcherS(errDsp)
	}
	var connID string
	if routeID != nil &&
		*routeID != "" {
		// use previously discovered route
		if x, ok := engine.Cache.Get(utils.CacheDispatcherRoutes,
			*routeID); ok && x != nil {
			connID = x.(string)
			if err = dS.conns[connID].Call(serviceMethod, args, reply); !utils.IsNetworkError(err) {
				return
			}
		}
	}
	for _, connID = range d.ConnIDs() {
		conn, has := dS.conns[connID]
		if !has {
			err = utils.NewErrDispatcherS(
				fmt.Errorf("no connection with id: <%s>", connID))
			continue
		}
		if err = conn.Call(serviceMethod, args, reply); utils.IsNetworkError(err) {
			continue
		}
		if routeID != nil &&
			*routeID != "" { // cache the discovered route
			engine.Cache.Set(utils.CacheDispatcherRoutes, *routeID, connID,
				nil, true, utils.EmptyString)
		}
		break
	}
	return
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
