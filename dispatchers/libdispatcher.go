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
	"encoding/gob"
	"fmt"
	"sort"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func init() {
	gob.Register(new(LoadMetrics))

}

// Dispatcher is responsible for routing requests to pool of connections
// there will be different implementations based on strategy

// Dispatcher is responsible for routing requests to pool of connections
// there will be different implementations based on strategy
type Dispatcher interface {
	// SetProfile is used to update the configuration information within dispatcher
	// to make sure we take decisions based on latest config
	SetProfile(pfl *engine.DispatcherProfile)
	// HostIDs returns the ordered list of host IDs
	HostIDs() (hostIDs engine.DispatcherHostIDs)
	// Dispatch is used to send the method over the connections given
	Dispatch(routeID string, subsystem,
		serviceMethod string, args interface{}, reply interface{}) (err error)
}

type strategyDispatcher interface {
	// dispatch is used to send the method over the connections given
	dispatch(dm *engine.DataManager, routeID string, subsystem, tnt string, hostIDs []string,
		serviceMethod string, args interface{}, reply interface{}) (err error)
}

// newDispatcher constructs instances of Dispatcher
func newDispatcher(dm *engine.DataManager, pfl *engine.DispatcherProfile) (d Dispatcher, err error) {
	pfl.Hosts.Sort() // make sure the connections are sorted
	hosts := pfl.Hosts.Clone()
	switch pfl.Strategy {
	case utils.MetaWeight:
		var strDsp strategyDispatcher
		if strDsp, err = newSingleStrategyDispatcher(hosts, pfl.StrategyParams, pfl.TenantID()); err != nil {
			return
		}
		d = &WeightDispatcher{
			dm:       dm,
			tnt:      pfl.Tenant,
			hosts:    hosts,
			strategy: strDsp,
		}
	case utils.MetaRandom:
		var strDsp strategyDispatcher
		if strDsp, err = newSingleStrategyDispatcher(hosts, pfl.StrategyParams, pfl.TenantID()); err != nil {
			return
		}
		d = &RandomDispatcher{
			dm:       dm,
			tnt:      pfl.Tenant,
			hosts:    hosts,
			strategy: strDsp,
		}
	case utils.MetaRoundRobin:
		var strDsp strategyDispatcher
		if strDsp, err = newSingleStrategyDispatcher(hosts, pfl.StrategyParams, pfl.TenantID()); err != nil {
			return
		}
		d = &RoundRobinDispatcher{
			dm:       dm,
			tnt:      pfl.Tenant,
			hosts:    hosts,
			strategy: strDsp,
		}
	case rpcclient.PoolBroadcast,
		rpcclient.PoolBroadcastSync,
		rpcclient.PoolBroadcastAsync:
		d = &WeightDispatcher{
			dm:       dm,
			tnt:      pfl.Tenant,
			hosts:    hosts,
			strategy: &broadcastStrategyDispatcher{strategy: pfl.Strategy},
		}
	default:
		err = fmt.Errorf("unsupported dispatch strategy: <%s>", pfl.Strategy)
	}
	return
}

// WeightDispatcher selects the next connection based on weight
type WeightDispatcher struct {
	sync.RWMutex
	dm       *engine.DataManager
	tnt      string
	hosts    engine.DispatcherHostProfiles
	strategy strategyDispatcher
}

// SetProfile used to implement Dispatcher interface
func (wd *WeightDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	wd.Lock()
	pfl.Hosts.Sort()
	wd.hosts = pfl.Hosts.Clone() // avoid concurrency on profile
	wd.Unlock()
	return
}

// HostIDs used to implement Dispatcher interface
func (wd *WeightDispatcher) HostIDs() (hostIDs engine.DispatcherHostIDs) {
	wd.RLock()
	hostIDs = wd.hosts.HostIDs()
	wd.RUnlock()
	return
}

// Dispatch used to implement Dispatcher interface
func (wd *WeightDispatcher) Dispatch(routeID string, subsystem,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	return wd.strategy.dispatch(wd.dm, routeID, subsystem, wd.tnt, wd.HostIDs(),
		serviceMethod, args, reply)
}

// RandomDispatcher selects the next connection randomly
// together with RouteID can serve as load-balancer
type RandomDispatcher struct {
	sync.RWMutex
	dm       *engine.DataManager
	tnt      string
	hosts    engine.DispatcherHostProfiles
	strategy strategyDispatcher
}

// SetProfile used to implement Dispatcher interface
func (d *RandomDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	d.Lock()
	d.hosts = pfl.Hosts.Clone()
	d.Unlock()
	return
}

// HostIDs used to implement Dispatcher interface
func (d *RandomDispatcher) HostIDs() (hostIDs engine.DispatcherHostIDs) {
	d.RLock()
	hostIDs = d.hosts.HostIDs()
	d.RUnlock()
	hostIDs.Shuffle() // randomize the connections
	return
}

// Dispatch used to implement Dispatcher interface
func (d *RandomDispatcher) Dispatch(routeID string, subsystem,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	return d.strategy.dispatch(d.dm, routeID, subsystem, d.tnt, d.HostIDs(),
		serviceMethod, args, reply)
}

// RoundRobinDispatcher selects the next connection in round-robin fashion
type RoundRobinDispatcher struct {
	sync.RWMutex
	dm       *engine.DataManager
	tnt      string
	hosts    engine.DispatcherHostProfiles
	hostIdx  int // used for the next connection
	strategy strategyDispatcher
}

// SetProfile used to implement Dispatcher interface
func (d *RoundRobinDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	d.Lock()
	d.hosts = pfl.Hosts.Clone()
	d.Unlock()
	return
}

// HostIDs used to implement Dispatcher interface
func (d *RoundRobinDispatcher) HostIDs() (hostIDs engine.DispatcherHostIDs) {
	d.RLock()
	hostIDs = d.hosts.HostIDs()
	hostIDs.ReorderFromIndex(d.hostIdx)
	d.hostIdx++
	if d.hostIdx >= len(d.hosts) {
		d.hostIdx = 0
	}
	d.RUnlock()
	return
}

// Dispatch used to implement Dispatcher interface
func (d *RoundRobinDispatcher) Dispatch(routeID string, subsystem,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	return d.strategy.dispatch(d.dm, routeID, subsystem, d.tnt, d.HostIDs(),
		serviceMethod, args, reply)
}

type singleResultstrategyDispatcher struct{}

func (*singleResultstrategyDispatcher) dispatch(dm *engine.DataManager, routeID string, subsystem, tnt string,
	hostIDs []string, serviceMethod string, args interface{}, reply interface{}) (err error) {
	var dH *engine.DispatcherHost
	if routeID != utils.EmptyString {
		// overwrite routeID with RouteID:Subsystem
		routeID = utils.ConcatenatedKey(routeID, subsystem)
		// use previously discovered route
		if x, ok := engine.Cache.Get(utils.CacheDispatcherRoutes,
			routeID); ok && x != nil {
			dH = x.(*engine.DispatcherHost)
			if err = dH.Call(serviceMethod, args, reply); !rpcclient.IsNetworkError(err) {
				return
			}
		}
	}
	var called bool
	for _, hostID := range hostIDs {
		if dH, err = dm.GetDispatcherHost(tnt, hostID, true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				utils.Logger.Warning(fmt.Sprintf("<%s> could not find host with ID %q",
					utils.DispatcherS, hostID))
				err = nil
				continue
			}
			err = utils.NewErrDispatcherS(err)
			return
		}
		called = true
		if err = dH.Call(serviceMethod, args, reply); rpcclient.IsNetworkError(err) {
			continue
		}
		if routeID != utils.EmptyString { // cache the discovered route
			if err = engine.Cache.Set(utils.CacheDispatcherRoutes, routeID, dH,
				nil, true, utils.EmptyString); err != nil {
				return
			}
		}
		break
	}
	if !called { // in case we do not match any host
		err = utils.ErrHostNotFound
		return
	}
	return
}

type broadcastStrategyDispatcher struct {
	strategy string
}

func (b *broadcastStrategyDispatcher) dispatch(dm *engine.DataManager, routeID string, subsystem, tnt string, hostIDs []string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	var hasHosts bool
	pool := rpcclient.NewRPCPool(b.strategy, config.CgrConfig().GeneralCfg().ReplyTimeout)
	for _, hostID := range hostIDs {
		var dH *engine.DispatcherHost
		if dH, err = dm.GetDispatcherHost(tnt, hostID, true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				utils.Logger.Warning(fmt.Sprintf("<%s> could not find host with ID %q",
					utils.DispatcherS, hostID))
				err = nil
				continue
			}
			return utils.NewErrDispatcherS(err)
		}
		hasHosts = true
		pool.AddClient(dH)
	}
	if !hasHosts { // in case we do not match any host
		return utils.ErrHostNotFound
	}
	return pool.Call(serviceMethod, args, reply)
}

func newSingleStrategyDispatcher(hosts engine.DispatcherHostProfiles, params map[string]interface{}, tntID string) (ls strategyDispatcher, err error) {
	if dflt, has := params[utils.MetaDefaultRatio]; has {
		var ratio int64
		if ratio, err = utils.IfaceAsTInt64(dflt); err != nil {
			return nil, err
		}
		return &loadStrategyDispatcher{
			tntID:        tntID,
			hosts:        hosts.Clone(),
			defaultRatio: ratio,
		}, nil
	}
	for _, host := range hosts {
		if _, has := host.Params[utils.MetaRatio]; has {
			return &loadStrategyDispatcher{
				tntID:        tntID,
				hosts:        hosts.Clone(),
				defaultRatio: 1,
			}, nil
		}
	}
	return new(singleResultstrategyDispatcher), nil
}

type loadStrategyDispatcher struct {
	tntID        string
	hosts        engine.DispatcherHostProfiles
	defaultRatio int64
}

func newLoadMetrics(hosts engine.DispatcherHostProfiles, dfltRatio int64) (*LoadMetrics, error) {
	lM := &LoadMetrics{
		HostsLoad:  make(map[string]int64),
		HostsRatio: make(map[string]int64),
	}
	for _, host := range hosts {
		if strRatio, has := host.Params[utils.MetaRatio]; !has {
			lM.HostsRatio[host.ID] = dfltRatio
		} else if ratio, err := utils.IfaceAsTInt64(strRatio); err != nil {
			return nil, err
		} else {
			lM.HostsRatio[host.ID] = ratio
		}
	}
	return lM, nil
}

// LoadMetrics the structure to save the metrix for load strategy
type LoadMetrics struct {
	mutex      sync.RWMutex
	HostsLoad  map[string]int64
	HostsRatio map[string]int64
}

func (ld *loadStrategyDispatcher) dispatch(dm *engine.DataManager, routeID string, subsystem, tnt string, hostIDs []string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	var dH *engine.DispatcherHost
	var lM *LoadMetrics
	if x, ok := engine.Cache.Get(utils.CacheDispatcherLoads, ld.tntID); ok && x != nil {
		var canCast bool
		if lM, canCast = x.(*LoadMetrics); !canCast {
			return fmt.Errorf("cannot cast %+v to *LoadMetrics", x)
		}
	} else if lM, err = newLoadMetrics(ld.hosts, ld.defaultRatio); err != nil {
		return
	}

	if routeID != utils.EmptyString {
		// overwrite routeID with RouteID:Subsystem
		routeID = utils.ConcatenatedKey(routeID, subsystem)
		// use previously discovered route
		if x, ok := engine.Cache.Get(utils.CacheDispatcherRoutes,
			routeID); ok && x != nil {
			dH = x.(*engine.DispatcherHost)
			lM.incrementLoad(dH.ID, ld.tntID)
			err = dH.Call(serviceMethod, args, reply)
			lM.decrementLoad(dH.ID, ld.tntID) // call ended
			if !rpcclient.IsNetworkError(err) {
				return
			}
		}
	}
	var called bool
	for _, hostID := range lM.getHosts(hostIDs) {
		if dH, err = dm.GetDispatcherHost(tnt, hostID, true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				utils.Logger.Warning(fmt.Sprintf("<%s> could not find host with ID %q",
					utils.DispatcherS, hostID))
				err = nil
				continue
			}
			err = utils.NewErrDispatcherS(err)
			return
		}
		called = true
		lM.incrementLoad(hostID, ld.tntID)
		err = dH.Call(serviceMethod, args, reply)
		lM.decrementLoad(hostID, ld.tntID) // call ended
		if rpcclient.IsNetworkError(err) {
			continue
		}
		if routeID != utils.EmptyString { // cache the discovered route
			if err = engine.Cache.Set(utils.CacheDispatcherRoutes, routeID, dH,
				nil, true, utils.EmptyString); err != nil {
				return
			}
		}
		break
	}
	if !called { // in case we do not match any host
		err = utils.ErrHostNotFound
		return
	}
	return
}

// used to sort the host IDs based on costs
type hostCosts struct {
	ids      []string
	multiple []int64
}

func (hc *hostCosts) Len() int           { return len(hc.ids) }
func (hc *hostCosts) Less(i, j int) bool { return hc.multiple[i] < hc.multiple[j] }
func (hc *hostCosts) Swap(i, j int) {
	hc.multiple[i], hc.multiple[j] = hc.multiple[j], hc.multiple[i]
	hc.ids[i], hc.ids[j] = hc.ids[j], hc.ids[i]
}

func (lM *LoadMetrics) getHosts(hostIDs []string) []string {
	hlp := &hostCosts{
		ids:      make([]string, 0, len(hostIDs)),
		multiple: make([]int64, 0, len(hostIDs)),
	}
	lM.mutex.RLock()

	for _, id := range hostIDs {
		switch {
		case lM.HostsRatio[id] < 0:
			hlp.multiple = append(hlp.multiple, 0)
		case lM.HostsRatio[id] == 0:
			continue
		default:
			hlp.multiple = append(hlp.multiple, lM.HostsLoad[id]/lM.HostsRatio[id])
		}
		hlp.ids = append(hlp.ids, id)
	}
	lM.mutex.RUnlock()
	sort.Stable(hlp)
	return hlp.ids
}

func (lM *LoadMetrics) incrementLoad(hostID, tntID string) {
	lM.mutex.Lock()
	lM.HostsLoad[hostID]++
	engine.Cache.ReplicateSet(utils.CacheDispatcherLoads, tntID, lM)
	lM.mutex.Unlock()
}

func (lM *LoadMetrics) decrementLoad(hostID, tntID string) {
	lM.mutex.Lock()
	lM.HostsLoad[hostID]--
	engine.Cache.ReplicateSet(utils.CacheDispatcherLoads, tntID, lM)
	lM.mutex.Unlock()
}
