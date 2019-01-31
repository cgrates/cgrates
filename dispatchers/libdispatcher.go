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
	"sort"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type DispatcherConn struct {
	ID        string
	FilterIDs []string
	Weight    float64                // applied in case of multiple connections need to be ordered
	Params    map[string]interface{} // additional parameters stored for a session
	Blocker   bool                   // no connection after this one
}

// DispatcherProfile is the config for one Dispatcher
type DispatcherProfile struct {
	Tenant             string
	ID                 string
	Subsystems         []string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // activation interval
	Strategy           string
	StrategyParams     map[string]interface{} // ie for distribution, set here the pool weights
	Weight             float64                // used for profile sorting on match
	Connections        []*DispatcherConn      // dispatch to these connections
}

func (dP *DispatcherProfile) TenantID() string {
	return utils.ConcatenatedKey(dP.Tenant, dP.ID)
}

// DispatcherProfiles is a sortable list of Dispatcher profiles
type DispatcherProfiles []*DispatcherProfile

// Sort is part of sort interface, sort based on Weight
func (dps DispatcherProfiles) Sort() {
	sort.Slice(dps, func(i, j int) bool { return dps[i].Weight > dps[j].Weight })
}

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
	switch pfl.Strategy {
	case utils.MetaWeight:
		d = &WeightDispatcher{pfl: pfl}
	default:
		err = fmt.Errorf("unsupported dispatch strategy: <%s>", pfl.Strategy)
	}
	return
}

type WeightDispatcher struct {
	pfl *engine.DispatcherProfile
}

func (wd *WeightDispatcher) SetProfile(pfl *engine.DispatcherProfile) {
	wd.pfl = pfl
	return
}

func (wd *WeightDispatcher) NextConnID() (connID string) {
	return
}
