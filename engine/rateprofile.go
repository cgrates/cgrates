/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

// RateProfile represents the configuration of a Rate profile
type RateProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval
	Weight             float64
	ConnectFee         float64
	RoundingMethod     string
	RoundingDecimals   int
	MinCost            float64
	MaxCost            float64
	MaxCostStrategy    string
	Rates              []*Rate

	connFee *utils.Decimal // cached version of the Decimal
	minCost *utils.Decimal
	maxCost *utils.Decimal
}

func (rpp *RateProfile) TenantID() string {
	return utils.ConcatenatedKey(rpp.Tenant, rpp.ID)
}

// Route defines rate related information used within a RateProfile
type Rate struct {
	ID            string        // RateID
	FilterIDs     []string      // RateFilterIDs
	IntervalStart time.Duration // Starting point when the Rate kicks in
	Weight        float64       // RateWeight will decide the winner per interval start
	Value         float64       // RateValue
	Unit          time.Duration // RateUnit
	Increment     time.Duration // RateIncrement
	Blocker       bool          // RateBlocker will make this rate recurrent, deactivating further intervals

	val *utils.Decimal // cached version of the Decimal
}

// RateProfileWithArgDispatcher is used in replicatorV1 for dispatcher
type RateProfileWithArgDispatcher struct {
	*RateProfile
	*utils.ArgDispatcher
}

type TPRateProfile struct {
	TPid               string
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.TPActivationInterval
	Weight             float64
	ConnectFee         float64
	RoundingMethod     string
	RoundingDecimals   int
	MinCost            float64
	MaxCost            float64
	MaxCostStrategy    string
	Rates              []*Rate
}
