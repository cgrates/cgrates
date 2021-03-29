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
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type RGRate struct {
	GroupIntervalStart time.Duration
	Value              float64
	RateIncrement      time.Duration
	RateUnit           time.Duration
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (r *RGRate) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if r == nil || len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.GroupIntervalStart:
		return r.GroupIntervalStart, nil
	case utils.Value:
		return r.Value, nil
	case utils.RateIncrement:
		return r.RateIncrement, nil
	case utils.RateUnit:
		return r.RateUnit, nil
	}
}

func (r *RGRate) Stringify() string {
	return utils.Sha1(fmt.Sprintf("%v", r))[:8]
}

func (p *RGRate) Equal(o *RGRate) bool {
	return p.GroupIntervalStart == o.GroupIntervalStart &&
		p.Value == o.Value &&
		p.RateIncrement == o.RateIncrement &&
		p.RateUnit == o.RateUnit
}

type RateGroups []*RGRate

func (pg RateGroups) Len() int {
	return len(pg)
}

func (pg RateGroups) Swap(i, j int) {
	pg[i], pg[j] = pg[j], pg[i]
}

func (pg RateGroups) Less(i, j int) bool {
	return pg[i].GroupIntervalStart < pg[j].GroupIntervalStart
}

func (pg RateGroups) Sort() {
	sort.Sort(pg)
}

func (pg RateGroups) Equal(og RateGroups) bool {
	if len(pg) != len(og) {
		return false
	}
	for i := 0; i < len(pg); i++ {
		if !pg[i].Equal(og[i]) {
			return false
		}
	}
	return true
}

func (pg *RateGroups) AddRate(ps ...*RGRate) {
	for _, p := range ps {
		found := false
		for _, op := range *pg {
			if op.Equal(p) {
				found = true
				break
			}
		}
		if !found {
			*pg = append(*pg, p)
		}
	}
}

func (pg RateGroups) Equals(oRG RateGroups) bool {
	if len(pg) != len(oRG) {
		return false
	}
	for i := range pg {
		if !pg[i].Equal(oRG[i]) {
			return false
		}
	}
	return true
}

func (pg RateGroups) Clone() (cln RateGroups) {
	cln = make(RateGroups, len(pg))
	for i, rt := range pg {
		cln[i] = new(RGRate)
		*cln[i] = *rt
	}
	return
}

func (i *RateInterval) GetCost(duration, startSecond time.Duration) float64 {
	price, _, rateUnit := i.GetRateParameters(startSecond)
	price /= float64(rateUnit.Nanoseconds())
	d := float64(duration.Nanoseconds())
	return utils.Round(d*price, globalRoundingDecimals, utils.MetaRoundingMiddle)
}

// Gets the price for a the provided start second
func (i *RateInterval) GetRateParameters(startSecond time.Duration) (rate float64, rateIncrement, rateUnit time.Duration) {
	if i.Rating == nil {
		return -1, -1, -1
	}
	i.Rating.Rates.Sort()
	for index, price := range i.Rating.Rates {
		if price.GroupIntervalStart <= startSecond && (index == len(i.Rating.Rates)-1 ||
			i.Rating.Rates[index+1].GroupIntervalStart > startSecond) {
			if price.RateIncrement == 0 {
				price.RateIncrement = 1 * time.Second
			}
			if price.RateUnit == 0 {
				price.RateUnit = 1 * time.Second
			}
			return price.Value, price.RateIncrement, price.RateUnit
		}
	}
	return -1, -1, -1
}

func (ri *RateInterval) GetMaxCost() (float64, string) {
	if ri.Rating == nil {
		return 0.0, ""
	}
	return ri.Rating.MaxCost, ri.Rating.MaxCostStrategy
}

// Structure to store intervals according to weight
type RateIntervalList []*RateInterval

func (rl RateIntervalList) GetWeight() float64 {
	// all reates should have the same weight
	// just in case get the max
	var maxWeight float64
	for _, r := range rl {
		if r.Weight > maxWeight {
			maxWeight = r.Weight
		}
	}
	return maxWeight
}

// Structure to store intervals according to weight
type RateIntervalTimeSorter struct {
	referenceTime time.Time
	ris           []*RateInterval
}

func (il *RateIntervalTimeSorter) Len() int {
	return len(il.ris)
}

func (il *RateIntervalTimeSorter) Swap(i, j int) {
	il.ris[i], il.ris[j] = il.ris[j], il.ris[i]
}

// we need higher weights earlyer in the list
func (il *RateIntervalTimeSorter) Less(j, i int) bool {
	if il.ris[i].Weight < il.ris[j].Weight {
		return il.ris[i].Weight < il.ris[j].Weight
	}
	t1 := il.ris[i].Timing.getLeftMargin(il.referenceTime)
	t2 := il.ris[j].Timing.getLeftMargin(il.referenceTime)
	return t1.After(t2)
}

func (il *RateIntervalTimeSorter) Sort() []*RateInterval {
	sort.Sort(il)
	return il.ris
}

// Clone clones RateInterval
func (i *RateInterval) Clone() (cln *RateInterval) {
	if i == nil {
		return
	}
	cln = &RateInterval{
		Timing: i.Timing.Clone(),
		Rating: i.Rating.Clone(),
		Weight: i.Weight,
	}
	return
}

// Clone clones RITiming
func (rit *RITiming) Clone() (cln *RITiming) {
	if rit == nil {
		return
	}
	cln = &RITiming{
		ID:        rit.ID,
		StartTime: rit.StartTime,
		EndTime:   rit.EndTime,
	}
	if len(rit.Years) != 0 {
		cln.Years = make(utils.Years, len(rit.Years))
		for i, year := range rit.Years {
			cln.Years[i] = year
		}
	}
	if len(rit.Months) != 0 {
		cln.Months = make(utils.Months, len(rit.Months))
		for i, month := range rit.Months {
			cln.Months[i] = month
		}
	}
	if len(rit.MonthDays) != 0 {
		cln.MonthDays = make(utils.MonthDays, len(rit.MonthDays))
		for i, monthDay := range rit.MonthDays {
			cln.MonthDays[i] = monthDay
		}
	}
	if len(rit.WeekDays) != 0 {
		cln.WeekDays = make(utils.WeekDays, len(rit.WeekDays))
		for i, weekDay := range rit.WeekDays {
			cln.WeekDays[i] = weekDay
		}
	}
	return
}

// Clone clones RIRate
func (rit *RIRate) Clone() (cln *RIRate) {
	if rit == nil {
		return
	}
	cln = &RIRate{
		ConnectFee:       rit.ConnectFee,
		RoundingMethod:   rit.RoundingMethod,
		RoundingDecimals: rit.RoundingDecimals,
		MaxCost:          rit.MaxCost,
		MaxCostStrategy:  rit.MaxCostStrategy,
	}
	if rit.Rates != nil {
		cln.Rates = make([]*RGRate, len(rit.Rates))
		for i, rate := range rit.Rates {
			cln.Rates[i] = rate.Clone()
		}
	}
	return cln
}

// Clone clones Rates
func (r *RGRate) Clone() (cln *RGRate) {
	if r == nil {
		return
	}
	cln = &RGRate{
		GroupIntervalStart: r.GroupIntervalStart,
		Value:              r.Value,
		RateIncrement:      r.RateIncrement,
		RateUnit:           r.RateUnit,
	}
	return
}
