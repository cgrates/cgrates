/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package rater

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

type TPLoader interface {
	LoadDestinations() error
	LoadRates() error
	LoadDestinationRates() error
	LoadTimings() error
	LoadDestinationRateTimings() error
	LoadRatingProfiles() error
	LoadActions() error
	LoadActionTimings() error
	LoadActionTriggers() error
	LoadAccountActions() error
	WriteToDatabase(bool, bool) error
}

type Rate struct {
	Tag                                                    string
	ConnectFee, Price, PricedUnits, RateIncrements, Weight float64
}

func NewRate(tag, connectFee, price, pricedUnits, rateIncrements, weight string) (r *Rate, err error) {
	cf, err := strconv.ParseFloat(connectFee, 64)
	if err != nil {
		log.Printf("Error parsing connect fee from: %v", connectFee)
		return
	}
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		log.Printf("Error parsing price from: %v", price)
		return
	}
	pu, err := strconv.ParseFloat(pricedUnits, 64)
	if err != nil {
		log.Printf("Error parsing priced units from: %v", pricedUnits)
		return
	}
	ri, err := strconv.ParseFloat(rateIncrements, 64)
	if err != nil {
		log.Printf("Error parsing rates increments from: %v", rateIncrements)
		return
	}
	wght, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		log.Printf("Error parsing rates increments from: %s", weight)
		return
	}
	r = &Rate{
		Tag:            tag,
		ConnectFee:     cf,
		Price:          p,
		PricedUnits:    pu,
		RateIncrements: ri,
		Weight:         wght,
	}
	return
}

type DestinationRate struct {
	Tag             string
	DestinationsTag string
	RateTag         string
	Rate            *Rate
}

type Timing struct {
	Id        string
	Years     Years
	Months    Months
	MonthDays MonthDays
	WeekDays  WeekDays
	StartTime string
}

func NewTiming(timingInfo ...string) (rt *Timing) {
	rt = &Timing{}
	rt.Id = timingInfo[0]
	rt.Years.Parse(timingInfo[1], ";")
	rt.Months.Parse(timingInfo[2], ";")
	rt.MonthDays.Parse(timingInfo[3], ";")
	rt.WeekDays.Parse(timingInfo[4], ";")
	rt.StartTime = timingInfo[5]
	return
}

type DestinationRateTiming struct {
	Tag                 string
	DestinationRatesTag string
	Weight              float64
	TimingsTag          string // intermediary used when loading from db
	timing              *Timing
}

func NewDestinationRateTiming(destinationRatesTag string, timing *Timing, weight string) (rt *DestinationRateTiming) {
	w, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		log.Printf("Error parsing weight unit from: %v", weight)
		return
	}
	rt = &DestinationRateTiming{
		DestinationRatesTag: destinationRatesTag,
		Weight:              w,
		timing:              timing,
	}
	return
}

func (rt *DestinationRateTiming) GetInterval(dr *DestinationRate) (i *Interval) {
	i = &Interval{
		Years:          rt.timing.Years,
		Months:         rt.timing.Months,
		MonthDays:      rt.timing.MonthDays,
		WeekDays:       rt.timing.WeekDays,
		StartTime:      rt.timing.StartTime,
		Weight:         rt.Weight,
		ConnectFee:     dr.Rate.ConnectFee,
		Price:          dr.Rate.Price,
		PricedUnits:    dr.Rate.PricedUnits,
		RateIncrements: dr.Rate.RateIncrements,
	}
	return
}

type AccountAction struct {
	Tenant, Account, Direction, ActionTimingsTag, ActionTriggersTag string
}

func ValidateCSVData(fn string, re *regexp.Regexp) (err error) {
	fin, err := os.Open(fn)
	if err != nil {
		// do not return the error, the file might be not needed
		return nil
	}
	defer fin.Close()
	r := bufio.NewReader(fin)
	line_number := 1
	for {
		line, truncated, err := r.ReadLine()
		if err != nil {
			break
		}
		if truncated {
			return errors.New("line too long")
		}
		// skip the header line
		if line_number > 1 {
			if !re.Match(line) {
				return errors.New(fmt.Sprintf("%s: error on line %d: %s", fn, line_number, line))
			}
		}
		line_number++
	}
	return
}
