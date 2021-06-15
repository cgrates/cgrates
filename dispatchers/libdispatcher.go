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
	"math/rand"
	"sort"
	"sync"

	"github.com/cgrates/birpc/context"
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
type Dispatcher interface {
	// Dispatch is used to send the method over the connections given
	Dispatch(dm *engine.DataManager, flts *engine.FilterS,
		ctx *context.Context,
		ev utils.DataProvider, tnt, routeID, subsystem,
		serviceMethod string, args interface{}, reply interface{}) (err error)
}

// newDispatcher constructs instances of Dispatcher
func newDispatcher(pfl *engine.DispatcherProfile) (d Dispatcher, err error) {
	hosts := pfl.Hosts.Clone()
	hosts.Sort() // make sure the connections are sorted
	switch pfl.Strategy {
	case utils.MetaWeight:
		return newSingleDispatcher(hosts, pfl.StrategyParams, pfl.TenantID(), new(noSort))
	case utils.MetaRandom:
		return newSingleDispatcher(hosts, pfl.StrategyParams, pfl.TenantID(), new(randomSort))
	case utils.MetaRoundRobin:
		return newSingleDispatcher(hosts, pfl.StrategyParams, pfl.TenantID(), new(roundRobinSort))
	case rpcclient.PoolBroadcast,
		rpcclient.PoolBroadcastSync,
		rpcclient.PoolBroadcastAsync:
		return &broadcastDispatcher{
			strategy: pfl.Strategy,
			hosts:    hosts,
		}, nil
	default:
		err = fmt.Errorf("unsupported dispatch strategy: <%s>", pfl.Strategy)
	}
	return
}

func getDispatcherHosts(fltrs *engine.FilterS, ev utils.DataProvider, ctx *context.Context, tnt string, hosts engine.DispatcherHostProfiles) (hostIDs engine.DispatcherHostIDs, err error) {
	hostIDs = make(engine.DispatcherHostIDs, 0, len(hosts))
	for _, host := range hosts {
		var pass bool
		if pass, err = fltrs.Pass(ctx, tnt, host.FilterIDs, ev); err != nil {
			return
		}
		if pass {
			hostIDs = append(hostIDs, host.ID)
			if host.Blocker {
				break
			}
		}
	}
	return
}

type hostSorter interface {
	Sort(fltrs *engine.FilterS, ev utils.DataProvider, ctx *context.Context, tnt string, hosts engine.DispatcherHostProfiles) (hostIDs engine.DispatcherHostIDs, err error)
}

type noSort struct{}

func (noSort) Sort(fltrs *engine.FilterS, ev utils.DataProvider, ctx *context.Context, tnt string, hosts engine.DispatcherHostProfiles) (hostIDs engine.DispatcherHostIDs, err error) {
	return getDispatcherHosts(fltrs, ev, ctx, tnt, hosts)
}

type randomSort struct{}

func (randomSort) Sort(fltrs *engine.FilterS, ev utils.DataProvider, ctx *context.Context, tnt string, hosts engine.DispatcherHostProfiles) (hostIDs engine.DispatcherHostIDs, err error) {
	rand.Shuffle(len(hosts), func(i, j int) {
		hosts[i], hosts[j] = hosts[j], hosts[i]
	})
	return getDispatcherHosts(fltrs, ev, ctx, tnt, hosts)
}

type roundRobinSort struct{ nextIDx int }

func (rs *roundRobinSort) Sort(fltrs *engine.FilterS, ev utils.DataProvider, ctx *context.Context, tnt string, hosts engine.DispatcherHostProfiles) (hostIDs engine.DispatcherHostIDs, err error) {
	dh := make(engine.DispatcherHostProfiles, len(hosts))
	idx := rs.nextIDx
	for i := 0; i < len(dh); i++ {
		if idx > len(dh)-1 {
			idx = 0
		}
		dh[i] = hosts[idx]
		idx++
	}
	rs.nextIDx++
	if rs.nextIDx >= len(hosts) {
		rs.nextIDx = 0
	}
	return getDispatcherHosts(fltrs, ev, ctx, tnt, dh)
}

func newSingleDispatcher(hosts engine.DispatcherHostProfiles, params map[string]interface{}, tntID string, sorter hostSorter) (_ Dispatcher, err error) {
	if dflt, has := params[utils.MetaDefaultRatio]; has {
		var ratio int64
		if ratio, err = utils.IfaceAsTInt64(dflt); err != nil {
			return
		}
		return &loadDispatcher{
			tntID:        tntID,
			defaultRatio: ratio,
			sorter:       sorter,
			hosts:        hosts,
		}, nil
	}
	for _, host := range hosts {
		if _, has := host.Params[utils.MetaRatio]; has {
			return &loadDispatcher{
				tntID:        tntID,
				defaultRatio: 1,
				sorter:       sorter,
				hosts:        hosts,
			}, nil
		}
	}
	return &singleResultDispatcher{
		sorter: sorter,
		hosts:  hosts,
	}, nil
}

type singleResultDispatcher struct {
	sorter hostSorter
	hosts  engine.DispatcherHostProfiles
}

func (sd *singleResultDispatcher) Dispatch(dm *engine.DataManager, flts *engine.FilterS,
	ctx *context.Context,
	ev utils.DataProvider, tnt, routeID, subsystem string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	var dH *engine.DispatcherHost
	if routeID != utils.EmptyString {
		// overwrite routeID with RouteID:Subsystem
		routeID = utils.ConcatenatedKey(routeID, subsystem)
		// use previously discovered route
		if x, ok := engine.Cache.Get(utils.CacheDispatcherRoutes,
			routeID); ok && x != nil {
			dH = x.(*engine.DispatcherHost)
			if err = dH.Call(ctx, serviceMethod, args, reply); !rpcclient.IsNetworkError(err) {
				return
			}
		}
	}
	var hostIDs []string
	if hostIDs, err = sd.sorter.Sort(flts, ev, ctx, tnt, sd.hosts); err != nil {
		return
	}
	var called bool
	for _, hostID := range hostIDs {
		if dH, err = dm.GetDispatcherHost(ctx, tnt, hostID, true, true, utils.NonTransactional); err != nil {
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
		if err = dH.Call(ctx, serviceMethod, args, reply); rpcclient.IsNetworkError(err) {
			continue
		}
		if routeID != utils.EmptyString { // cache the discovered route
			if err = engine.Cache.Set(ctx, utils.CacheDispatcherRoutes, routeID, dH,
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

type broadcastDispatcher struct {
	strategy string
	hosts    engine.DispatcherHostProfiles
}

func (b *broadcastDispatcher) Dispatch(dm *engine.DataManager, flts *engine.FilterS,
	ctx *context.Context,
	ev utils.DataProvider, tnt, routeID, subsystem string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	var hostIDs []string
	if hostIDs, err = getDispatcherHosts(flts, ev, ctx, tnt, b.hosts); err != nil {
		return
	}
	var hasHosts bool
	pool := rpcclient.NewRPCPool(b.strategy, config.CgrConfig().GeneralCfg().ReplyTimeout)
	for _, hostID := range hostIDs {
		var dH *engine.DispatcherHost
		if dH, err = dm.GetDispatcherHost(ctx, tnt, hostID, true, true, utils.NonTransactional); err != nil {
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
	return pool.Call(ctx, serviceMethod, args, reply)
}

type loadDispatcher struct {
	tntID        string
	defaultRatio int64
	sorter       hostSorter
	hosts        engine.DispatcherHostProfiles
}

func (ld *loadDispatcher) Dispatch(dm *engine.DataManager, flts *engine.FilterS,
	ctx *context.Context,
	ev utils.DataProvider, tnt, routeID, subsystem string,
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
			lM.incrementLoad(ctx, dH.ID, ld.tntID)
			err = dH.Call(ctx, serviceMethod, args, reply)
			lM.decrementLoad(ctx, dH.ID, ld.tntID) // call ended
			if !rpcclient.IsNetworkError(err) {
				return
			}
		}
	}
	var hostIDs []string
	if hostIDs, err = ld.sorter.Sort(flts, ev, ctx, tnt, lM.getHosts(ld.hosts)); err != nil {
		return
	}
	var called bool
	for _, hostID := range hostIDs {
		if dH, err = dm.GetDispatcherHost(ctx, tnt, hostID, true, true, utils.NonTransactional); err != nil {
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
		lM.incrementLoad(ctx, hostID, ld.tntID)
		err = dH.Call(ctx, serviceMethod, args, reply)
		lM.decrementLoad(ctx, hostID, ld.tntID) // call ended
		if rpcclient.IsNetworkError(err) {
			continue
		}
		if routeID != utils.EmptyString { // cache the discovered route
			if err = engine.Cache.Set(ctx, utils.CacheDispatcherRoutes, routeID, dH,
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

// used to sort the host IDs based on costs
type hostCosts struct {
	hosts engine.DispatcherHostProfiles
	load  []int64
}

func (hc *hostCosts) Len() int           { return len(hc.hosts) }
func (hc *hostCosts) Less(i, j int) bool { return hc.load[i] < hc.load[j] }
func (hc *hostCosts) Swap(i, j int) {
	hc.load[i], hc.load[j] = hc.load[j], hc.load[i]
	hc.hosts[i], hc.hosts[j] = hc.hosts[j], hc.hosts[i]
}

func (lM *LoadMetrics) getHosts(hosts engine.DispatcherHostProfiles) engine.DispatcherHostProfiles {
	hlp := &hostCosts{
		hosts: make(engine.DispatcherHostProfiles, 0, len(hosts)),
		load:  make([]int64, 0, len(hosts)),
	}
	lM.mutex.RLock()

	for _, host := range hosts {
		switch {
		case lM.HostsRatio[host.ID] < 0:
			hlp.load = append(hlp.load, 0)
		case lM.HostsRatio[host.ID] == 0:
			continue
		default:
			hlp.load = append(hlp.load, lM.HostsLoad[host.ID]/lM.HostsRatio[host.ID])
		}
		hlp.hosts = append(hlp.hosts, host)
	}
	lM.mutex.RUnlock()
	sort.Stable(hlp)
	return hlp.hosts
}

func (lM *LoadMetrics) incrementLoad(ctx *context.Context, hostID, tntID string) {
	lM.mutex.Lock()
	lM.HostsLoad[hostID]++
	engine.Cache.ReplicateSet(ctx, utils.CacheDispatcherLoads, tntID, lM)
	lM.mutex.Unlock()
}

func (lM *LoadMetrics) decrementLoad(ctx *context.Context, hostID, tntID string) {
	lM.mutex.Lock()
	lM.HostsLoad[hostID]--
	engine.Cache.ReplicateSet(ctx, utils.CacheDispatcherLoads, tntID, lM)
	lM.mutex.Unlock()
}
