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
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	internalDispatcher = &engine.DispatcherProfile{Tenant: utils.MetaInternal, ID: utils.MetaInternal}
)

func init() {
	gob.Register(new(LoadMetrics))
	gob.Register(new(DispatcherRoute))
	//gob.RegisterName("dispatchers.DispatcherRoute", DispatcherRoute{})
}

// isInternalDispatcherProfile compares the profile to the internal one
func isInternalDispatcherProfile(d *engine.DispatcherProfile) bool {
	return d.Tenant == internalDispatcher.Tenant && d.ID == internalDispatcher.ID
}

// DispatcherRoute is bounded to a routeID
type DispatcherRoute struct {
	Tenant, ProfileID, HostID string
}

// getDispatcherWithCache
func getDispatcherWithCache(ctx *context.Context, dPrfl *engine.DispatcherProfile, dm *engine.DataManager) (d Dispatcher, err error) {
	tntID := dPrfl.TenantID()
	if x, ok := engine.Cache.Get(utils.CacheDispatchers,
		tntID); ok && x != nil {
		d = x.(Dispatcher)
		return
	}
	if dPrfl.Hosts == nil { // dispatcher profile was not retrieved but built artificially above, try retrieving
		if dPrfl, err = dm.GetDispatcherProfile(ctx, dPrfl.Tenant, dPrfl.ID,
			true, true, utils.NonTransactional); err != nil {
			return
		}
	}
	if d, err = newDispatcher(dPrfl); err != nil {
		return
	} else if err = engine.Cache.Set(ctx, utils.CacheDispatchers, tntID, d, // cache the built Dispatcher
		nil, true, utils.EmptyString); err != nil {
		return
	}
	return
}

// Dispatcher is responsible for routing requests to pool of connections
// there will be different implementations based on strategy
type Dispatcher interface {
	// Dispatch is used to send the method over the connections given
	Dispatch(dm *engine.DataManager, flts *engine.FilterS, cfg *config.CGRConfig,
		ctx *context.Context, iPRCCh chan birpc.ClientConnector,
		ev utils.DataProvider, tnt, routeID string, dR *DispatcherRoute,
		serviceMethod string, args any, reply any) (err error)
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

// getDispatcherHosts returns a list of host IDs matching the event with filters
func getDispatcherHosts(fltrs *engine.FilterS, ev utils.DataProvider,
	ctx *context.Context, tnt string, hosts engine.DispatcherHostProfiles) (hostIDs engine.DispatcherHostIDs, err error) {
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

// hostSorted is the sorting interface used by singleDispatcher
type hostSorter interface {
	Sort(fltrs *engine.FilterS, ev utils.DataProvider,
		ctx *context.Context, tnt string, hosts engine.DispatcherHostProfiles) (hostIDs engine.DispatcherHostIDs, err error)
}

// noSort will just return the matching hosts for the event.
type noSort struct{}

func (noSort) Sort(fltrs *engine.FilterS, ev utils.DataProvider,
	ctx *context.Context, tnt string, hosts engine.DispatcherHostProfiles) (hostIDs engine.DispatcherHostIDs, err error) {
	return getDispatcherHosts(fltrs, ev, ctx, tnt, hosts)
}

// randomSort will randomize the matching hosts for the event
type randomSort struct{}

func (randomSort) Sort(fltrs *engine.FilterS, ev utils.DataProvider,
	ctx *context.Context, tnt string, hosts engine.DispatcherHostProfiles) (hostIDs engine.DispatcherHostIDs, err error) {
	rand.Shuffle(len(hosts), func(i, j int) {
		hosts[i], hosts[j] = hosts[j], hosts[i]
	})
	return getDispatcherHosts(fltrs, ev, ctx, tnt, hosts)
}

// roundRoinSort will sort the matching hosts for the event in a round-robin fashion via nextIDx
// which will be increased on each Sort iteration
type roundRobinSort struct{ nextIDx int }

func (rs *roundRobinSort) Sort(fltrs *engine.FilterS, ev utils.DataProvider,
	ctx *context.Context, tnt string, hosts engine.DispatcherHostProfiles) (hostIDs engine.DispatcherHostIDs, err error) {
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

// newSingleDispatcher is the constructor for singleDispatcher struct.
func newSingleDispatcher(hosts engine.DispatcherHostProfiles, params map[string]any,
	tntID string, sorter hostSorter) (_ Dispatcher, err error) {
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

// singleResultDispatcher routes the event to a single host
// implements Dispatcher interface
type singleResultDispatcher struct {
	sorter hostSorter
	hosts  engine.DispatcherHostProfiles
}

func (sd *singleResultDispatcher) Dispatch(dm *engine.DataManager, flts *engine.FilterS, cfg *config.CGRConfig,
	ctx *context.Context, iPRCCh chan birpc.ClientConnector,
	ev utils.DataProvider, tnt, routeID string, dR *DispatcherRoute,
	serviceMethod string, args any, reply any) (err error) {
	if dR != nil && dR.HostID != utils.EmptyString { // route to previously discovered route
		return callDHwithID(ctx, tnt, dR.HostID, routeID, dR, dm,
			cfg, iPRCCh, serviceMethod, args, reply)
	}
	var hostIDs []string
	if hostIDs, err = sd.sorter.Sort(flts, ev, ctx, tnt, sd.hosts); err != nil {
		return
	} else if len(hostIDs) == 0 { // in case we do not match any host
		return utils.ErrDSPHostNotFound
	}
	for _, hostID := range hostIDs {
		var dRh *DispatcherRoute
		if routeID != utils.EmptyString {
			dRh = &DispatcherRoute{
				Tenant:    dR.Tenant,
				ProfileID: dR.ProfileID,
				HostID:    hostID,
			}
		}
		if err = callDHwithID(ctx, tnt, hostID, routeID, dRh, dm,
			cfg, iPRCCh, serviceMethod, args, reply); err != utils.ErrDSPHostNotFound &&
			!rpcclient.ShouldFailover(err) { // successful dispatch with normal errors
			return
		}
		if err != nil {
			// not found or network errors will continue with standard dispatching
			utils.Logger.Warning(fmt.Sprintf("<%s> error <%s> dispatching to host with identity <%s>",
				utils.DispatcherS, err.Error(), hostID))
		}
	}
	return
}

// broadcastDispatcher routes the event to multiple hosts in a pool
// implements the Dispatcher interface
type broadcastDispatcher struct {
	strategy string
	hosts    engine.DispatcherHostProfiles
}

func (b *broadcastDispatcher) Dispatch(dm *engine.DataManager, flts *engine.FilterS, cfg *config.CGRConfig,
	ctx *context.Context, iPRCCh chan birpc.ClientConnector,
	ev utils.DataProvider, tnt, routeID string, dR *DispatcherRoute,
	serviceMethod string, args any, reply any) (err error) {
	var hostIDs []string
	if hostIDs, err = getDispatcherHosts(flts, ev, ctx, tnt, b.hosts); err != nil {
		return
	}
	var hasHosts bool
	pool := rpcclient.NewRPCPool(b.strategy, config.CgrConfig().GeneralCfg().ReplyTimeout)
	for _, hostID := range hostIDs {
		var dH *engine.DispatcherHost
		if dH, err = dm.GetDispatcherHost(ctx, tnt, hostID, true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrDSPHostNotFound {
				utils.Logger.Warning(fmt.Sprintf("<%s> could not find host with ID %q",
					utils.DispatcherS, hostID))
				err = nil
				continue
			}
			return utils.NewErrDispatcherS(err)
		}
		hasHosts = true
		var dRh *DispatcherRoute
		if routeID != utils.EmptyString {
			dRh = &DispatcherRoute{
				Tenant:    dR.Tenant,
				ProfileID: dR.ProfileID,
				HostID:    hostID,
			}
		}
		pool.AddClient(&lazyDH{
			dh:      dH,
			cfg:     cfg,
			iPRCCh:  iPRCCh,
			routeID: routeID,
			dR:      dRh,
		})
	}
	if !hasHosts { // in case we do not match any host
		return utils.ErrDSPHostNotFound
	}
	return pool.Call(ctx, serviceMethod, args, reply)
}

type loadDispatcher struct {
	tntID        string
	defaultRatio int64
	sorter       hostSorter
	hosts        engine.DispatcherHostProfiles
}

func (ld *loadDispatcher) Dispatch(dm *engine.DataManager, flts *engine.FilterS, cfg *config.CGRConfig,
	ctx *context.Context, iPRCCh chan birpc.ClientConnector,
	ev utils.DataProvider, tnt, routeID string, dR *DispatcherRoute,
	serviceMethod string, args any, reply any) (err error) {

	var lM *LoadMetrics
	if x, ok := engine.Cache.Get(utils.CacheDispatcherLoads, ld.tntID); ok && x != nil {
		var canCast bool
		if lM, canCast = x.(*LoadMetrics); !canCast {
			return fmt.Errorf("cannot cast %+v to *LoadMetrics", x)
		}
	} else if lM, err = newLoadMetrics(ld.hosts, ld.defaultRatio); err != nil {
		return
	}
	if dR != nil { // route to previously discovered route
		lM.incrementLoad(ctx, dR.HostID, ld.tntID)
		err = callDHwithID(ctx, tnt, dR.HostID, routeID, dR, dm,
			cfg, iPRCCh, serviceMethod, args, reply)
		lM.decrementLoad(ctx, dR.HostID, ld.tntID)                             // call ended
		if err != utils.ErrDSPHostNotFound && !rpcclient.ShouldFailover(err) { // successful dispatch with normal errors
			return
		}
		// not found or network errors will continue with standard dispatching
		utils.Logger.Warning(fmt.Sprintf("<%s> error <%s> dispatching to host with id <%q>",
			utils.DispatcherS, err.Error(), dR.HostID))
	}
	var hostIDs []string
	if hostIDs, err = ld.sorter.Sort(flts, ev, ctx, tnt, lM.getHosts(ld.hosts)); err != nil {
		return
	} else if len(hostIDs) == 0 { // in case we do not match any host
		return utils.ErrDSPHostNotFound
	}
	for _, hostID := range hostIDs {
		var dRh *DispatcherRoute
		if routeID != utils.EmptyString {
			dRh = &DispatcherRoute{
				Tenant:    dR.Tenant,
				ProfileID: dR.ProfileID,
				HostID:    hostID,
			}
		}
		lM.incrementLoad(ctx, hostID, ld.tntID)
		err = callDHwithID(ctx, tnt, hostID, routeID, dRh, dm,
			cfg, iPRCCh, serviceMethod, args, reply)
		lM.decrementLoad(ctx, hostID, ld.tntID)                                // call ended
		if err != utils.ErrDSPHostNotFound && !rpcclient.ShouldFailover(err) { // successful dispatch with normal errors
			return
		}
		if err != nil {
			// not found or network errors will continue with standard dispatching
			utils.Logger.Warning(fmt.Sprintf("<%s> error <%s> dispatching to host with id <%q>",
				utils.DispatcherS, err.Error(), hostID))
		}
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

// callDHwithID is a wrapper on callDH using ID of the host which the other cannot do due to lazyDH
// if routeID provided, will also cache once the call is successful
func callDHwithID(ctx *context.Context, tnt, hostID, routeID string, dR *DispatcherRoute,
	dm *engine.DataManager, cfg *config.CGRConfig, iPRCCh chan birpc.ClientConnector,
	serviceMethod string, args, reply any) (err error) {
	var dH *engine.DispatcherHost
	if dH, err = dm.GetDispatcherHost(ctx, tnt, hostID, true, true, utils.NonTransactional); err != nil {
		return
	}
	if err = callDH(ctx, dH, routeID, dR, cfg, iPRCCh, serviceMethod, args, reply); err != nil {
		return
	}
	return
}

func callDH(ctx *context.Context,
	dh *engine.DispatcherHost, routeID string, dR *DispatcherRoute,
	cfg *config.CGRConfig, iPRCCh chan birpc.ClientConnector,
	method string, args, reply any) (err error) {
	if routeID != utils.EmptyString { // cache the discovered route before dispatching
		argsCache := &utils.ArgCacheReplicateSet{
			Tenant: dh.Tenant,
			APIOpts: map[string]any{
				utils.MetaSubsys: utils.MetaDispatchers,
				utils.MetaNodeID: cfg.GeneralCfg().NodeID,
			},
			CacheID:  utils.CacheDispatcherRoutes,
			ItemID:   routeID,
			Value:    dR,
			GroupIDs: []string{utils.ConcatenatedKey(utils.CacheDispatcherProfiles, dR.Tenant, dR.ProfileID)},
		}
		if err = engine.Cache.SetWithReplicate(ctx, argsCache); err != nil {
			if !rpcclient.ShouldFailover(err) {
				return
			}
			// did not dispatch properly, fail-back to standard dispatching
			utils.Logger.Warning(fmt.Sprintf("<%s> ignoring cache network error <%s> setting route dR %+v",
				utils.DispatcherS, err.Error(), dR))
		}
	}
	var conn birpc.ClientConnector
	if conn, err = dh.GetConn(ctx, cfg, iPRCCh); err != nil {
		return
	}
	if err = conn.Call(ctx, method, args, reply); err != nil {
		return
	}
	return
}

// lazyDH is created for the broadcast strategy so we can make sure host exists during setup phase
type lazyDH struct {
	dh      *engine.DispatcherHost
	cfg     *config.CGRConfig
	iPRCCh  chan birpc.ClientConnector
	routeID string
	dR      *DispatcherRoute
}

func (l *lazyDH) Call(ctx *context.Context, method string, args, reply any) (err error) {
	return callDH(ctx, l.dh, l.routeID, l.dR, l.cfg, l.iPRCCh, method, args, reply)
}

// newInternalHost returns an internal host as needed for internal dispatching
func newInternalHost(tnt string) *engine.DispatcherHost {
	return &engine.DispatcherHost{
		Tenant: tnt,
		RemoteHost: &config.RemoteHost{
			ID:              utils.MetaInternal,
			Address:         utils.MetaInternal,
			ConnectAttempts: 1,
			Reconnects:      1,
			ConnectTimeout:  time.Second,
			ReplyTimeout:    time.Second,
		},
	}
}
