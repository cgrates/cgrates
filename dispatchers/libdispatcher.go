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
	"sync"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// Dispatcher is responsible for routing requests to pool of connections
// there will be different implementations based on strategy
type Dispatcher interface {
	// SetProfile is used to update the configuration information within dispatcher
	// to make sure we take decisions based on latest config
	SetProfile(pfl *engine.DispatcherProfile)
	// ConnIDs returns the ordered list of hosts IDs
	ConnIDs() (conns []string)
	// Dispatch is used to send the method over the connections given
	Dispatch(conns map[string]*rpcclient.RpcClientPool, routeID *string,
		serviceMethod string, args interface{}, reply interface{}) (err error)
}

type StrategyDispatcher interface {
	// Dispatch is used to send the method over the connections given
	Dispatch(connIDs []string, conns map[string]*rpcclient.RpcClientPool, routeID *string,
		serviceMethod string, args interface{}, reply interface{}) (err error)
}

type singleResultStrategyDispatcher struct{}
type brodcastStrategyDispatcher struct{}

// newDispatcher constructs instances of Dispatcher
func newDispatcher(pfl *engine.DispatcherProfile) (d Dispatcher, err error) {
	pfl.Conns.Sort() // make sure the connections are sorted
	switch pfl.Strategy {
	case utils.MetaWeight:
		d = &WeightDispatcher{
			conns:    pfl.Conns.Clone(),
			strategy: new(singleResultStrategyDispatcher),
		}
	case utils.MetaRandom:
		d = &RandomDispatcher{
			conns:    pfl.Conns.Clone(),
			strategy: new(singleResultStrategyDispatcher),
		}
	case utils.MetaRoundRobin:
		d = &RoundRobinDispatcher{
			conns:    pfl.Conns.Clone(),
			strategy: new(singleResultStrategyDispatcher),
		}
	case utils.MetaBroadcast:
		d = &BroadcastDispatcher{
			conns:    pfl.Conns.Clone(),
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
	conns    engine.DispatcherConns
	strategy StrategyDispatcher
}

func (wd *WeightDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	wd.Lock()
	pfl.Conns.Sort()
	wd.conns = pfl.Conns.Clone() // avoid concurrency on profile
	wd.Unlock()
	return
}

func (wd *WeightDispatcher) ConnIDs() (connIDs []string) {
	wd.RLock()
	connIDs = wd.conns.ConnIDs()
	wd.RUnlock()
	return
}

func (wd *WeightDispatcher) Dispatch(conns map[string]*rpcclient.RpcClientPool, routeID *string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	return wd.strategy.Dispatch(wd.ConnIDs(), conns, routeID, serviceMethod, args, reply)
}

// RandomDispatcher selects the next connection randomly
// together with RouteID can serve as load-balancer
type RandomDispatcher struct {
	sync.RWMutex
	conns    engine.DispatcherConns
	strategy StrategyDispatcher
}

func (d *RandomDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	d.Lock()
	d.conns = pfl.Conns.Clone()
	d.Unlock()
	return
}

func (d *RandomDispatcher) ConnIDs() (connIDs []string) {
	d.RLock()
	conns := d.conns.Clone()
	d.RUnlock()
	conns.Shuffle() // randomize the connections
	return conns.ConnIDs()
}

func (d *RandomDispatcher) Dispatch(conns map[string]*rpcclient.RpcClientPool, routeID *string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	return d.strategy.Dispatch(d.ConnIDs(), conns, routeID, serviceMethod, args, reply)
}

// RoundRobinDispatcher selects the next connection in round-robin fashion
type RoundRobinDispatcher struct {
	sync.RWMutex
	conns    engine.DispatcherConns
	connIdx  int // used for the next connection
	strategy StrategyDispatcher
}

func (d *RoundRobinDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	d.Lock()
	d.conns = pfl.Conns.Clone()
	d.Unlock()
	return
}

func (d *RoundRobinDispatcher) ConnIDs() (connIDs []string) {
	d.RLock()
	conns := d.conns.Clone()
	conns.ReorderFromIndex(d.connIdx)
	d.connIdx++
	if d.connIdx >= len(d.conns) {
		d.connIdx = 0
	}
	d.RUnlock()
	return conns.ConnIDs()
}

func (d *RoundRobinDispatcher) Dispatch(conns map[string]*rpcclient.RpcClientPool, routeID *string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	return d.strategy.Dispatch(d.ConnIDs(), conns, routeID, serviceMethod, args, reply)
}

// RoundRobinDispatcher selects the next connection in round-robin fashion
type BroadcastDispatcher struct {
	sync.RWMutex
	conns    engine.DispatcherConns
	strategy StrategyDispatcher
}

func (d *BroadcastDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	d.Lock()
	pfl.Conns.Sort()
	d.conns = pfl.Conns.Clone() // avoid concurrency on profile
	d.Unlock()
	return
}

func (d *BroadcastDispatcher) ConnIDs() (connIDs []string) {
	d.RLock()
	connIDs = d.conns.ConnIDs()
	d.RUnlock()
	return
}

func (d *BroadcastDispatcher) Dispatch(conns map[string]*rpcclient.RpcClientPool, routeID *string,
	serviceMethod string, args interface{}, reply interface{}) (lastErr error) { // no cache needed for this strategy because we need to call all connections
	return d.strategy.Dispatch(d.ConnIDs(), conns, routeID, serviceMethod, args, reply)
}

func (ign *singleResultStrategyDispatcher) Dispatch(connIDs []string, conns map[string]*rpcclient.RpcClientPool, routeID *string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	var connID string
	if routeID != nil &&
		*routeID != "" {
		// use previously discovered route
		if x, ok := engine.Cache.Get(utils.CacheDispatcherRoutes,
			*routeID); ok && x != nil {
			connID = x.(string)
			if err = conns[connID].Call(serviceMethod, args, reply); !utils.IsNetworkError(err) {
				return
			}
		}
	}
	for _, connID = range connIDs {
		conn, has := conns[connID]
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

func (ign *brodcastStrategyDispatcher) Dispatch(connIDs []string, conns map[string]*rpcclient.RpcClientPool, routeID *string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	var hasErrors bool
	for _, connID := range connIDs {
		conn, has := conns[connID]
		if !has {
			err = utils.NewErrDispatcherS(
				fmt.Errorf("no connection with id: <%s>", connID))
			utils.Logger.Err(fmt.Sprintf("<%s> Error at %s strategy for connID %q : %s",
				utils.DispatcherS, utils.MetaBroadcast, connID, err.Error()))
			hasErrors = true
			continue
		}
		if err = conn.Call(serviceMethod, args, reply); utils.IsNetworkError(err) {
			utils.Logger.Err(fmt.Sprintf("<%s> Network Error at %s strategy for connID %q : %s",
				utils.DispatcherS, utils.MetaBroadcast, connID, err.Error()))
			hasErrors = true
		} else if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error at %s strategy for connID %q : %s",
				utils.DispatcherS, utils.MetaBroadcast, connID, err.Error()))
			hasErrors = true
		}
	}
	if hasErrors { // rewrite err if not all call were succesfull
		return utils.ErrPartiallyExecuted
	}
	return
}
