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
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
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
	Rates              map[string]*Rate

	connFee *utils.Decimal // cached version of the Decimal
	minCost *utils.Decimal
	maxCost *utils.Decimal
}

func (rp *RateProfile) TenantID() string {
	return utils.ConcatenatedKey(rp.Tenant, rp.ID)
}

func (rp *RateProfile) Compile() (err error) {
	rp.connFee = utils.NewDecimalFromFloat64(rp.ConnectFee)
	rp.minCost = utils.NewDecimalFromFloat64(rp.MinCost)
	rp.minCost = utils.NewDecimalFromFloat64(rp.MaxCost)
	for _, rtP := range rp.Rates {
		if err = rtP.Compile(); err != nil {
			return
		}
	}
	return
}

// Route defines rate related information used within a RateProfile
type Rate struct {
	ID              string   // RateID
	FilterIDs       []string // RateFilterIDs
	ActivationTimes string   // ActivationTimes is a cron formatted time interval
	Weight          float64  // RateWeight will decide the winner per interval start
	Blocker         bool     // RateBlocker will make this rate recurrent, deactivating further intervals
	IntervalRates   []*IntervalRate

	aTime cron.Schedule // compiled version of activation time as cron.Schedule interface
}

func (rt *Rate) Compile() (err error) {
	aTime := rt.ActivationTimes
	if aTime == utils.EmptyString {
		aTime = "* * * * *"
	}
	if rt.aTime, err = cron.ParseStandard(aTime); err != nil {
		return
	}
	return
}

// RunTimes returns the set of activation and deactivation times for this rate on the interval between >=sTime and <eTime
// aTimes is in the form of [][]
func (rt *Rate) RunTimes(sTime, eTime time.Time, verbosity int) (aTimes [][]time.Time, err error) {
	origSTime := sTime
	for i := 0; i < 5; i++ { // find out the first activation time in the past enabled for our sTime
		switch i {
		case 0:
			sTime = origSTime.Add(-time.Minute)
		case 1:
			sTime = origSTime.Add(-time.Hour)
		case 2:
			sTime = origSTime.AddDate(0, 0, -1)
		case 3:
			sTime = origSTime.AddDate(0, -1, 0)
		case 4:
			sTime = origSTime.AddDate(-1, 0, 0)
		}
		aTime := rt.aTime.Next(sTime)
		if sTime.Before(aTime) {
			continue
		}
		iTime := rt.aTime.NextInactive(sTime)
		aTimes = append(aTimes, []time.Time{aTime, iTime})
		if iTime.IsZero() || !eTime.After(iTime) { // #TestMe
			return
		}
		if iTime.IsZero() || !eTime.After(iTime) { // #TestMe
			return
		}
		sTime = iTime
		break
	}
	if len(aTimes) == 0 {
		return
	}
	for i := 0; i < verbosity; i++ {
		aTime := rt.aTime.Next(sTime)
		if aTime.IsZero() || !aTime.Before(eTime) { // #TestMe
			return
		}
		iTime := rt.aTime.NextInactive(sTime)
		aTimes = append(aTimes, []time.Time{aTime, iTime})
		if iTime.IsZero() || !eTime.After(iTime) { // #TestMe
			return
		}
		sTime = iTime
	}
	// protect from memory leak
	utils.Logger.Warning(
		fmt.Sprintf(
			"maximum runTime iterations reached for Rate: <%+v>, sTime: <%+v>, eTime: <%+v>",
			rt, sTime, eTime))
	return nil, utils.ErrMaxIterationsReached
}

type IntervalRate struct {
	IntervalStart time.Duration // Starting point when the Rate kicks in
	Unit          time.Duration // RateUnit
	Increment     time.Duration // RateIncrement
	Value         float64       // RateValue

	val *utils.Decimal // cached version of the Decimal
}

// RateProfileWithOpts is used in replicatorV1 for dispatcher
type RateProfileWithOpts struct {
	*RateProfile
	Opts map[string]interface{}
}

// RateSInterval is used by RateS to integrate Rate info for one charging interval
type RateSInterval struct {
	UsageStart time.Duration
	Increments []*RateSIncrement
}

type RateSIncrement struct {
	Rate              *Rate
	IntervalRateIndex int
	Usage             time.Duration
}

// Sort will sort the IntervalRates from each Rate based on IntervalStart
func (rpp *RateProfile) Sort() {
	for _, rate := range rpp.Rates {
		sort.Slice(rate.IntervalRates, func(i, j int) bool {
			return rate.IntervalRates[i].IntervalStart < rate.IntervalRates[j].IntervalStart
		})
	}
}
