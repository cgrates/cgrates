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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Dispatcher is responsible for routing requests to pool of connections
// there will be different implementations based on strategy
type Dispatcher interface {
	// SetConfig is used to update the configuration information within dispatcher
	// to make sure we take decisions based on latest config
	SetProfile(pfl *engine.DispatcherProfile)
	// GetConnID returns an ordered list of connection IDs for the event
	NextConnID() (connID string)
}

// newDispatcher constructs instances of Dispatcher
func newDispatcher(pfl *engine.DispatcherProfile) (d Dispatcher, err error) {
	pfl.Conns.Sort() // make sure the connections are sorted
	switch pfl.Strategy {
	case utils.MetaWeight:
		d = &WeightDispatcher{pfl: pfl}
	default:
		err = fmt.Errorf("unsupported dispatch strategy: <%s>", pfl.Strategy)
	}
	return
}

// WeightDispatcher selects the next connection based on weight
type WeightDispatcher struct {
	pfl         *engine.DispatcherProfile
	nextConnIdx int // last returned connection index
}

func (wd *WeightDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	pfl.Conns.Sort()
	wd.pfl = pfl
	return
}

func (wd *WeightDispatcher) NextConnID() (connID string) {
	connID = wd.pfl.Conns[wd.nextConnIdx].ID
	wd.nextConnIdx++
	if wd.nextConnIdx > len(wd.pfl.Conns)-1 {
		wd.nextConnIdx = 0 // start from beginning
	}
	return
}
