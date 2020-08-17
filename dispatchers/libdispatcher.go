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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
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
		d = &WeightDispatcher{
			dm:       dm,
			tnt:      pfl.Tenant,
			hosts:    hosts,
			strategy: newSingleStrategyDispatcher(hosts, pfl.TenantID()),
		}
	case utils.MetaRandom:
		d = &RandomDispatcher{
			dm:       dm,
			tnt:      pfl.Tenant,
			hosts:    hosts,
			strategy: newSingleStrategyDispatcher(hosts, pfl.TenantID()),
		}
	case utils.MetaRoundRobin:
		d = &RoundRobinDispatcher{
			dm:       dm,
			tnt:      pfl.Tenant,
			hosts:    hosts,
			strategy: newSingleStrategyDispatcher(hosts, pfl.TenantID()),
		}
	case utils.MetaBroadcast:
		d = &WeightDispatcher{
			dm:       dm,
			tnt:      pfl.Tenant,
			hosts:    hosts,
			strategy: new(brodcastStrategyDispatcher),
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

func (wd *WeightDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	wd.Lock()
	pfl.Hosts.Sort()
	wd.hosts = pfl.Hosts.Clone() // avoid concurrency on profile
	wd.Unlock()
	return
}

func (wd *WeightDispatcher) HostIDs() (hostIDs engine.DispatcherHostIDs) {
	wd.RLock()
	hostIDs = wd.hosts.HostIDs()
	wd.RUnlock()
	return
}

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

func (d *RandomDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	d.Lock()
	d.hosts = pfl.Hosts.Clone()
	d.Unlock()
	return
}

func (d *RandomDispatcher) HostIDs() (hostIDs engine.DispatcherHostIDs) {
	d.RLock()
	hostIDs = d.hosts.HostIDs()
	d.RUnlock()
	hostIDs.Shuffle() // randomize the connections
	return
}

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

func (d *RoundRobinDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	d.Lock()
	d.hosts = pfl.Hosts.Clone()
	d.Unlock()
	return
}

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
			if err = dH.Call(serviceMethod, args, reply); !utils.IsNetworkError(err) {
				return
			}
		}
	}
	for _, hostID := range hostIDs {
		if dH, err = dm.GetDispatcherHost(tnt, hostID, true, true, utils.NonTransactional); err != nil {
			err = utils.NewErrDispatcherS(err)
			return
		}
		if err = dH.Call(serviceMethod, args, reply); utils.IsNetworkError(err) {
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
	return
}

type brodcastStrategyDispatcher struct{}

func (*brodcastStrategyDispatcher) dispatch(dm *engine.DataManager, routeID string, subsystem, tnt string, hostIDs []string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	var hasErrors bool
	for _, hostID := range hostIDs {
		var dH *engine.DispatcherHost
		if dH, err = dm.GetDispatcherHost(tnt, hostID, true, true, utils.NonTransactional); err != nil {
			err = utils.NewErrDispatcherS(err)
			return
		}
		if err = dH.Call(serviceMethod, args, reply); utils.IsNetworkError(err) {
			utils.Logger.Err(fmt.Sprintf("<%s> network error: <%s> at %s strategy for hostID %q",
				utils.DispatcherS, err.Error(), utils.MetaBroadcast, hostID))
			hasErrors = true
		} else if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s> at %s strategy for hostID %q",
				utils.DispatcherS, err.Error(), utils.MetaBroadcast, hostID))
			hasErrors = true
		}
	}
	if hasErrors { // rewrite err if not all call were succesfull
		return utils.ErrPartiallyExecuted
	}
	return
}

func newSingleStrategyDispatcher(hosts engine.DispatcherHostProfiles, tntID string) (ls strategyDispatcher) {
	for _, host := range hosts {
		if _, has := host.Params[utils.MetaRatio]; has {
			return &loadStrategyDispatcher{
				tntID: tntID,
				hosts: hosts.Clone(),
			}
		}
	}
	return new(singleResultstrategyDispatcher)
}

type loadStrategyDispatcher struct {
	tntID string
	hosts engine.DispatcherHostProfiles
}

func newLoadMetrics(hosts engine.DispatcherHostProfiles) (*LoadMetrics, error) {
	lM := &LoadMetrics{
		HostsLoad:  make(map[string]int64),
		HostsRatio: make(map[string]int64),
		SumRatio:   0,
	}
	for _, host := range hosts {
		if strRatio, has := host.Params[utils.MetaRatio]; !has {
			lM.HostsRatio[host.ID] = 1
			lM.SumRatio++
		} else if ratio, err := utils.IfaceAsTInt64(strRatio); err != nil {
			return nil, err
		} else {
			lM.HostsRatio[host.ID] = ratio
			lM.SumRatio += ratio
		}
	}
	return lM, nil
}

// LoadMetrics the structure to save the metrix for load strategy
type LoadMetrics struct {
	mutex      sync.RWMutex
	HostsLoad  map[string]int64
	HostsRatio map[string]int64
	SumRatio   int64
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
	} else if lM, err = newLoadMetrics(ld.hosts); err != nil {
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
			if !utils.IsNetworkError(err) {
				return
			}
		}
	}
	for _, hostID := range lM.getHosts(hostIDs) {
		if dH, err = dm.GetDispatcherHost(tnt, hostID, true, true, utils.NonTransactional); err != nil {
			err = utils.NewErrDispatcherS(err)
			return
		}
		lM.incrementLoad(hostID, ld.tntID)
		err = dH.Call(serviceMethod, args, reply)
		lM.decrementLoad(hostID, ld.tntID) // call ended
		if utils.IsNetworkError(err) {
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
	return
}

// used to sort the host IDs based on costs
type hostCosts struct {
	ids   []string
	costs []int64
}

func (hc *hostCosts) Len() int           { return len(hc.ids) }
func (hc *hostCosts) Less(i, j int) bool { return hc.costs[i] < hc.costs[j] }
func (hc *hostCosts) Swap(i, j int) {
	hc.costs[i], hc.costs[j], hc.ids[i], hc.ids[j] = hc.costs[j], hc.costs[i], hc.ids[j], hc.ids[i]
}

func (lM *LoadMetrics) getHosts(hostIDs []string) []string {
	hlp := &hostCosts{
		ids:   hostIDs,
		costs: make([]int64, len(hostIDs)),
	}
	lM.mutex.RLock()
	for i, id := range hostIDs {
		hlp.costs[i] = lM.HostsLoad[id]
		if hlp.costs[i] >= lM.HostsRatio[id] {
			hlp.costs[i] += lM.SumRatio
		}
	}
	lM.mutex.RUnlock()
	sort.Sort(hlp)
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
