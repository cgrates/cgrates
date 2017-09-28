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

package engine

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

type ThresholdProfile struct {
	Tenant             string
	ID                 string
	Filters            []*RequestFilter          // Filters for the request
	ActivationInterval *utils.ActivationInterval // Time when this limit becomes active and expires
	MinItems           int                       // number of items agregated for the threshold to match
	Recurrent          bool
	MinSleep           time.Duration
	Blocker            bool // blocker flag to stop processing on filters matched
	Stored             bool
	Weight             float64 // Weight to sort the thresholds
	ActionIDs          []string
}

func (tp *ThresholdProfile) TenantID() string {
	return utils.ConcatenatedKey(tp.Tenant, tp.ID)
}
