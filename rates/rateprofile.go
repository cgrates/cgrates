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

package rates

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
	ConnectFee         float64 // #ToDo: replace here with decimal.Big
	RoundingMethod     string
	RoundingDecimals   int
	MinCost            float64 // #ToDo: replace here with decimal.Big
	MaxCost            float64 // #ToDo: replace here with decimal.Big
	MaxCostStrategy    string
	Rates              []*Rate
}

// Route defines rate related information used within a RateProfile
type Rate struct {
	ID        string        // RateID
	FilterIDs []string      // RateFilterIDs
	Weight    float64       // RateWeight
	Value     float64       // RateValue, #ToDo: replace here with decimal.Big
	Unit      time.Duration // RateUnit
	Increment time.Duration // RateIncrement
	Blocker   bool          // RateBlocker will make this rate recurrent

}
