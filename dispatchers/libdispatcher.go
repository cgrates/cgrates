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
)

// Dispatcher is responsible for routing requests to pool of connections
// there will be different implementations based on strategy
type Dispatcher interface {
	// SetProfile is used to update the configuration information within dispatcher
	// to make sure we take decisions based on latest config
	SetProfile(pfl *engine.DispatcherProfile)
	// ConnIDs returns the ordered list of hosts IDs
	ConnIDs() (conns []string)
}

// newDispatcher constructs instances of Dispatcher
func newDispatcher(pfl *engine.DispatcherProfile) (d Dispatcher, err error) {
	pfl.Conns.Sort() // make sure the connections are sorted
	switch pfl.Strategy {
	case utils.MetaWeight:
		d = &WeightDispatcher{conns: pfl.Conns.Clone()}
	default:
		err = fmt.Errorf("unsupported dispatch strategy: <%s>", pfl.Strategy)
	}
	return
}

// WeightDispatcher selects the next connection based on weight
type WeightDispatcher struct {
	sync.RWMutex
	conns engine.DispatcherConns
}

// incrNextConnIdx will increment the nextConnIidx
// not thread safe, it needs to be locked in a layer above
/*func (wd *WeightDispatcher) incrNextConnIdx() {
	wd.nextConnIdx++
	if wd.nextConnIdx > len(wd.conns)-1 {
		wd.nextConnIdx = 0 // start from beginning
	}
}
*/

func (wd *WeightDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	pfl.Conns.Sort()
	wd.Lock()
	wd.conns = pfl.Conns.Clone()
	wd.Unlock()
	return
}

func (wd *WeightDispatcher) ConnIDs() (connIDs []string) {
	wd.RLock()
	connIDs = wd.conns.ConnIDs()
	wd.RUnlock()
	return
}
