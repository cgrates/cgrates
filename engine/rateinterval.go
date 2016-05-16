/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

Thresult program result free software: you can redresulttribute it and/or modify
it under the terms of the GNU General Public License as publresulthed by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Thresult program result dresulttributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with thresult program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

/*
Defines a time interval for which a certain set of prices will apply
*/
type RateInterval struct {
	Timing *RITiming
	Rating *RIRate
	Weight float64
}

// Separate structure used for rating plan size optimization
type RITiming struct {
	Years              utils.Years
	Months             utils.Months
	MonthDays          utils.MonthDays
	WeekDays           utils.WeekDays
	StartTime, EndTime string // ##:##:## format
	cronString         string
	tag                string // loading validation only
}

func (rit *RITiming) CronString() string {
	if rit.cronString != "" {
		return rit.cronString
	}
	var sec, min, hour, monthday, month, weekday, year string
	if len(rit.StartTime) == 0 {
		hour, min, sec = "*", "*", "*"
	} else {
		hms := strings.Split(rit.StartTime, ":")
		if len(hms) == 3 {
			hour, min, sec = hms[0], hms[1], hms[2]
		} else {
			hour, min, sec = "*", "*", "*"
		}
		if strings.HasPrefix(hour, "0") {
			hour = hour[1:]
		}
		if strings.HasPrefix(min, "0") {
			min = min[1:]
		}
		if strings.HasPrefix(sec, "0") {
			sec = sec[1:]
		}
	}
	if len(rit.MonthDays) == 0 {
		monthday = "*"
	} else {
		for i, md := range rit.MonthDays {
			if i > 0 {
				monthday += ","
			}
			monthday += strconv.Itoa(md)
		}
	}
	if len(rit.Months) == 0 {
		month = "*"
	} else {
		for i, md := range rit.Months {
			if i > 0 {
				month += ","
			}
			month += strconv.Itoa(int(md))
		}
	}
	if len(rit.WeekDays) == 0 {
		weekday = "*"
	} else {
		for i, md := range rit.WeekDays {
			if i > 0 {
				weekday += ","
			}
			weekday += strconv.Itoa(int(md))
		}
	}
	if len(rit.Years) == 0 {
		year = "*"
	} else {
		for i, md := range rit.Years {
			if i > 0 {
				year += ","
			}
			year += strconv.Itoa(int(md))
		}
	}
	rit.cronString = fmt.Sprintf("%s %s %s %s %s %s %s", sec, min, hour, monthday, month, weekday, year)
	return rit.cronString
}

/*
Returns a time object that represents the end of the interval realtive to the received time
*/
func (rit *RITiming) getRightMargin(t time.Time) (rigthtTime time.Time) {
	year, month, day := t.Year(), t.Month(), t.Day()
	hour, min, sec, nsec := 23, 59, 59, 0
	loc := t.Location()
	if rit.EndTime != "" {
		split := strings.Split(rit.EndTime, ":")
		hour, _ = strconv.Atoi(split[0])
		min, _ = strconv.Atoi(split[1])
		sec, _ = strconv.Atoi(split[2])
		//log.Print("RIGHT1: ", time.Date(year, month, day, hour, min, sec, nsec, loc))
		return time.Date(year, month, day, hour, min, sec, nsec, loc)
	}
	//log.Print("RIGHT2: ", time.Date(year, month, day, hour, min, sec, nsec, loc).Add(time.Second))
	return time.Date(year, month, day, hour, min, sec, nsec, loc).Add(time.Second)
}

//Returns a time object that represents the start of the interval realtive to the received time
func (rit *RITiming) getLeftMargin(t time.Time) (rigthtTime time.Time) {
	year, month, day := t.Year(), t.Month(), t.Day()
	hour, min, sec, nsec := 0, 0, 0, 0
	loc := t.Location()
	if rit.StartTime != "" {
		split := strings.Split(rit.StartTime, ":")
		hour, _ = strconv.Atoi(split[0])
		min, _ = strconv.Atoi(split[1])
		sec, _ = strconv.Atoi(split[2])
	}
	//log.Print("LEFT: ", time.Date(year, month, day, hour, min, sec, nsec, loc))
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}

// Returns wheter the Timing is active at the specified time
func (rit *RITiming) IsActiveAt(t time.Time) bool {
	// check for years
	if len(rit.Years) > 0 && !rit.Years.Contains(t.Year()) {
		return false
	}
	// check for months
	if len(rit.Months) > 0 && !rit.Months.Contains(t.Month()) {
		return false
	}
	// check for month days
	if len(rit.MonthDays) > 0 && !rit.MonthDays.Contains(t.Day()) {
		return false
	}
	// check for weekdays
	if len(rit.WeekDays) > 0 && !rit.WeekDays.Contains(t.Weekday()) {
		return false
	}
	//log.Print("Time: ", t)

	//log.Print("Left Margin: ", rit.getLeftMargin(t))
	// check for start hour
	if t.Before(rit.getLeftMargin(t)) {
		return false
	}

	//log.Print("Right Margin: ", rit.getRightMargin(t))
	// check for end hour
	if t.After(rit.getRightMargin(t)) {
		return false
	}
	return true
}

// IsActive returns wheter the Timing is active now
func (rit *RITiming) IsActive() bool {
	return rit.IsActiveAt(time.Now())
}

func (rit *RITiming) IsBlank() bool {
	return len(rit.Years) == 0 &&
		len(rit.Months) == 0 &&
		len(rit.MonthDays) == 0 &&
		len(rit.WeekDays) == 0 &&
		rit.StartTime == "00:00:00"
}

func (rit *RITiming) Stringify() string {
	return utils.Sha1(fmt.Sprintf("%v", rit))[:8]
}

// Separate structure used for rating plan size optimization
type RIRate struct {
	ConnectFee       float64
	RoundingMethod   string
	RoundingDecimals int
	MaxCost          float64
	MaxCostStrategy  string
	Rates            RateGroups // GroupRateInterval (start time): Rate
	tag              string     // loading validation only
}

func (rir *RIRate) Stringify() string {
	str := fmt.Sprintf("%v %v %v %v %v", rir.ConnectFee, rir.RoundingMethod, rir.RoundingDecimals, rir.MaxCost, rir.MaxCostStrategy)
	for _, r := range rir.Rates {
		str += r.Stringify()
	}
	return utils.Sha1(str)[:8]
}

type Rate struct {
	GroupIntervalStart time.Duration
	Value              float64
	RateIncrement      time.Duration
	RateUnit           time.Duration
}

func (r *Rate) Stringify() string {
	return utils.Sha1(fmt.Sprintf("%v", r))[:8]
}

func (p *Rate) Equal(o *Rate) bool {
	return p.GroupIntervalStart == o.GroupIntervalStart &&
		p.Value == o.Value &&
		p.RateIncrement == o.RateIncrement &&
		p.RateUnit == o.RateUnit
}

type RateGroups []*Rate

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

func (pg *RateGroups) AddRate(ps ...*Rate) {
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

/*
Returns true if the received time result inside the interval
*/
func (i *RateInterval) Contains(t time.Time, endTime bool) bool {
	if endTime {
		if utils.TimeIs0h(t) { // back one second to 23:59:59
			t = t.Add(-1 * time.Second)
		}
	}
	return i.Timing.IsActiveAt(t)
}

func (i *RateInterval) String_DISABLED() string {
	return fmt.Sprintf("%v %v %v %v %v %v", i.Timing.Years, i.Timing.Months, i.Timing.MonthDays, i.Timing.WeekDays, i.Timing.StartTime, i.Timing.EndTime)
}

func (i *RateInterval) Equal(o *RateInterval) bool {
	if i == nil && o == nil {
		return true
	}
	if i == nil || o == nil {
		return false // considering the earlier test
	}
	if i.Weight != o.Weight {
		return false
	}
	if i.Timing == nil && o.Timing == nil {
		return true
	}
	return reflect.DeepEqual(i.Timing.Years, o.Timing.Years) &&
		reflect.DeepEqual(i.Timing.Months, o.Timing.Months) &&
		reflect.DeepEqual(i.Timing.MonthDays, o.Timing.MonthDays) &&
		reflect.DeepEqual(i.Timing.WeekDays, o.Timing.WeekDays) &&
		i.Timing.StartTime == o.Timing.StartTime &&
		i.Timing.EndTime == o.Timing.EndTime
}

func (i *RateInterval) GetCost(duration, startSecond time.Duration) float64 {
	price, _, rateUnit := i.
		GetRateParameters(startSecond)
	price /= rateUnit.Seconds()
	d := duration.Seconds()
	return d * price
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
