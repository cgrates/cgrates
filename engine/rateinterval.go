/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*
Defines a time interval for which a certain set of prices will apply
*/
type RateInterval struct {
	Years              Years
	Months             Months
	MonthDays          MonthDays
	WeekDays           WeekDays
	StartTime, EndTime string // ##:##:## format
	Weight, ConnectFee float64
	Rates              RateGroups // GroupRateInterval (start time): Rate
	RoundingMethod     string     //ROUNDING_UP, ROUNDING_DOWN, ROUNDING_MIDDLE
	RoundingDecimals   int
}

type Rate struct {
	GroupIntervalStart time.Duration
	Value              float64
	RateIncrement      time.Duration
	RateUnit           time.Duration
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
	// if the received time represents an endtime cosnidere it 24 instead of 0
	hour := t.Hour()
	if endTime && hour == 0 {
		hour = 24
	}
	// check for years
	if len(i.Years) > 0 && !i.Years.Contains(t.Year()) {
		return false
	}
	// check for months
	if len(i.Months) > 0 && !i.Months.Contains(t.Month()) {
		return false
	}
	// check for month days
	if len(i.MonthDays) > 0 && !i.MonthDays.Contains(t.Day()) {
		return false
	}
	// check for weekdays
	if len(i.WeekDays) > 0 && !i.WeekDays.Contains(t.Weekday()) {
		return false
	}
	// check for start hour
	if i.StartTime != "" {
		split := strings.Split(i.StartTime, ":")
		sh, _ := strconv.Atoi(split[0])
		sm, _ := strconv.Atoi(split[1])
		ss, _ := strconv.Atoi(split[2])
		// if the hour result before or result the same hour but the minute result before
		if hour < sh ||
			(hour == sh && t.Minute() < sm) ||
			(hour == sh && t.Minute() == sm && t.Second() < ss) {
			return false
		}
	}
	// check for end hour
	if i.EndTime != "" {
		split := strings.Split(i.EndTime, ":")
		eh, _ := strconv.Atoi(split[0])
		em, _ := strconv.Atoi(split[1])
		es, _ := strconv.Atoi(split[2])
		// if the hour result after or result the same hour but the minute result after
		if hour > eh ||
			(hour == eh && t.Minute() > em) ||
			(hour == eh && t.Minute() == em && t.Second() > es) {
			return false
		}
	}
	return true
}

/*
Returns a time object that represents the end of the interval realtive to the received time
*/
func (i *RateInterval) getRightMargin(t time.Time) (rigthtTime time.Time) {
	year, month, day := t.Year(), t.Month(), t.Day()
	hour, min, sec, nsec := 23, 59, 59, 0
	loc := t.Location()
	if i.EndTime != "" {
		split := strings.Split(i.EndTime, ":")
		hour, _ = strconv.Atoi(split[0])
		min, _ = strconv.Atoi(split[1])
		sec, _ = strconv.Atoi(split[2])
		//log.Print("RIGHT1: ", time.Date(year, month, day, hour, min, sec, nsec, loc))
		return time.Date(year, month, day, hour, min, sec, nsec, loc)
	}
	//log.Print("RIGHT2: ", time.Date(year, month, day, hour, min, sec, nsec, loc).Add(time.Second))
	return time.Date(year, month, day, hour, min, sec, nsec, loc).Add(time.Second)
}

/*
Returns a time object that represents the start of the interval realtive to the received time
*/
func (i *RateInterval) getLeftMargin(t time.Time) (rigthtTime time.Time) {
	year, month, day := t.Year(), t.Month(), t.Day()
	hour, min, sec, nsec := 0, 0, 0, 0
	loc := t.Location()
	if i.StartTime != "" {
		split := strings.Split(i.StartTime, ":")
		hour, _ = strconv.Atoi(split[0])
		min, _ = strconv.Atoi(split[1])
		sec, _ = strconv.Atoi(split[2])
	}
	//log.Print("LEFT: ", time.Date(year, month, day, hour, min, sec, nsec, loc))
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}

func (i *RateInterval) String_DISABLED() string {
	return fmt.Sprintf("%v %v %v %v %v %v", i.Years, i.Months, i.MonthDays, i.WeekDays, i.StartTime, i.EndTime)
}

func (i *RateInterval) Equal(o *RateInterval) bool {
	return reflect.DeepEqual(i.Years, o.Years) &&
		reflect.DeepEqual(i.Months, o.Months) &&
		reflect.DeepEqual(i.MonthDays, o.MonthDays) &&
		reflect.DeepEqual(i.WeekDays, o.WeekDays) &&
		i.StartTime == o.StartTime &&
		i.EndTime == o.EndTime
}

func (i *RateInterval) GetCost(duration, startSecond time.Duration) float64 {
	price, _, rateUnit := i.GetRateParameters(startSecond)
	d := duration.Seconds()
	price /= rateUnit.Seconds()

	return utils.Round(d*price, i.RoundingDecimals, i.RoundingMethod)
}

// Gets the price for a the provided start second
func (i *RateInterval) GetRateParameters(startSecond time.Duration) (price float64, rateIncrement, rateUnit time.Duration) {
	i.Rates.Sort()
	for index, price := range i.Rates {
		if price.GroupIntervalStart <= startSecond && (index == len(i.Rates)-1 ||
			i.Rates[index+1].GroupIntervalStart > startSecond) {
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

// Structure to store intervals according to weight
type RateIntervalList []*RateInterval

func (il RateIntervalList) Len() int {
	return len(il)
}

func (il RateIntervalList) Swap(i, j int) {
	il[i], il[j] = il[j], il[i]
}

func (il RateIntervalList) Less(i, j int) bool {
	return il[i].Weight < il[j].Weight
}

func (il RateIntervalList) Sort() {
	sort.Sort(il)
}
