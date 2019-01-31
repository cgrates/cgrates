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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewDispatcherService initializes a DispatcherService
func NewDispatcherService(dm *engine.DataManager,
	cfg *config.CGRConfig) (*DispatcherService, error) {
	return &DispatcherService{dm: dm, cfg: cfg}, nil
}

// DispatcherService  is the service handling dispatching towards internal components
// designed to handle automatic partitioning and failover
type DispatcherService struct {
	dm                  *engine.DataManager
	cfg                 *config.CGRConfig
	filterS             *engine.FilterS
	stringIndexedFields *[]string
	prefixIndexedFields *[]string
	conns               map[string]*rpcclient.RpcClientPool // available connections, accessed based on connID
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
	idxKeyPrfx := utils.ConcatenatedKey(ev.Tenant, utils.META_ANY)
	if subsys != "" {
		idxKeyPrfx = utils.ConcatenatedKey(ev.Tenant, subsys)
	}
	matchingPrfls := make(map[string]*engine.DispatcherProfile)
	prflIDs, err := engine.MatchingItemIDsForEvent(ev.Event, dS.stringIndexedFields, dS.prefixIndexedFields,
		dS.dm, utils.CacheDispatcherFilterIndexes, idxKeyPrfx, dS.cfg.FilterSCfg().IndexedSelects)
	if err != nil {
		return nil, err
	}
	for prflID := range prflIDs {
		prfl, err := dS.dm.GetDispatcherProfile(ev.Tenant, prflID, true, true, utils.NonTransactional)
		if err != nil {
			if err != utils.ErrNotFound {
				return nil, err
			}
			anyIdxPrfx := utils.ConcatenatedKey(ev.Tenant, utils.META_ANY)
			if idxKeyPrfx == anyIdxPrfx {
				continue // already checked *any
			}
			// check *any as subsystem
			if prfl, err = dS.dm.GetDispatcherProfile(ev.Tenant, prflID, true, true, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					continue
				}
				return nil, err
			}

		}
		if prfl.ActivationInterval != nil && ev.Time != nil &&
			!prfl.ActivationInterval.IsActiveAtTime(*ev.Time) { // not active
			continue
		}
		if pass, err := dS.filterS.Pass(ev.Tenant, prfl.FilterIDs,
			config.NewNavigableMap(ev.Event)); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		matchingPrfls[prflID] = prfl
	}
	if len(matchingPrfls) == 0 {
		return nil, utils.ErrNotFound
	}
	// All good, convert from Map to Slice so we can sort
	prfls := make(engine.DispatcherProfiles, len(matchingPrfls))
	i := 0
	for _, prfl := range matchingPrfls {
		prfls[i] = prfl
		i++
	}
	prfls.Sort()
	matchedPrlf := prfls[0] // only use the first profile
	tntID := matchedPrlf.TenantID()
	// get or build the Dispatcher for the config
	if x, ok := engine.Cache.Get(utils.CacheDispatchers,
		tntID); ok && x != nil {
		d = x.(Dispatcher)
		d.SetProfile(matchedPrlf)
		return
	}
	if d, err = newDispatcher(matchedPrlf); err != nil {
		return
	}
	engine.Cache.Set(utils.CacheDispatchers, tntID, d, nil,
		true, utils.EmptyString)
	return
}

// Dispatch is the method forwarding the request towards the right
func (dS *DispatcherService) Dispatch(ev *utils.CGREvent, subsys string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	d, errDsp := dS.dispatcherForEvent(ev, subsys)
	if errDsp != nil {
		return utils.NewErrDispatcherS(errDsp)
	}
	for i := 0; i < d.MaxConns(); i++ {
		connID := d.NextConnID()
		conn, has := dS.conns[connID]
		if !has {
			utils.NewErrDispatcherS(
				fmt.Errorf("no connection with id: <%s>", connID))
		}
		if err = conn.Call(serviceMethod, args, reply); !utils.IsNetworkError(err) {
			break
		}
	}
	return
}
