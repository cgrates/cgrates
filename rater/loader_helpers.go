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
	"github.com/cgrates/cgrates/utils"
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
	Tag                                                           string
	ConnectFee, Price, PricedUnits, RateIncrements, GroupInterval float64
	RoundingMethod                                                string
	RoundingDecimals                                              int
	Weight                                                        float64
}

func NewRate(tag, connectFee, price, pricedUnits, rateIncrements, groupInterval, roundingMethod, roundingDecimals, weight string) (r *Rate, err error) {
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
	gi, err := strconv.ParseFloat(groupInterval, 64)
	if err != nil {
		log.Printf("Error parsing group interval from: %v", price)
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
	rd, err := strconv.Atoi(roundingDecimals)
	if err != nil {
		log.Printf("Error parsing rounding decimals: %s", roundingDecimals)
		return
	}

	r = &Rate{
		Tag:              tag,
		ConnectFee:       cf,
		Price:            p,
		GroupInterval:    gi,
		PricedUnits:      pu,
		RateIncrements:   ri,
		Weight:           wght,
		RoundingMethod:   roundingMethod,
		RoundingDecimals: rd,
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
		Prices:         PriceGroups{&Price{dr.Rate.GroupInterval, dr.Rate.Price}},
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

type TPCSVRowValidator struct {
	FileName      string // File name
	Rule      *regexp.Regexp // Regexp rule
	ErrMessage string // Error message
}

var TPCSVRowValidators = []*TPCSVRowValidator{
			&TPCSVRowValidator{utils.DESTINATIONS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\d+.?\d*){1}$`),
				"Tag[0-9A-Za-z_],Prefix[0-9]"},
			&TPCSVRowValidator{utils.TIMINGS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\*all\s*,\s*|(?:\d{1,4};?)+\s*,\s*|\s*,\s*){4}(?:\d{2}:\d{2}:\d{2}|\*asap){1}$`),
				"Tag[0-9A-Za-z_],Years[0-9;]|*all|<empty>,Months[0-9;]|*all|<empty>,MonthDays[0-9;]|*all|<empty>,WeekDays[0-9;]|*all|<empty>,Time[0-9:]|*asap(00:00:00)"},
			&TPCSVRowValidator{utils.RATES_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){2}(?:\d+.?\d*,?){4}$`),
				"Tag[0-9A-Za-z_],ConnectFee[0-9.],Price[0-9.],PricedUnits[0-9.],RateIncrement[0-9.]"},
			&TPCSVRowValidator{utils.DESTINATION_RATES_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){2}(?:\d+.?\d*,?){4}$`),
				"Tag[0-9A-Za-z_],DestinationsTag[0-9A-Za-z_],RateTag[0-9A-Za-z_]"},
			&TPCSVRowValidator{utils.DESTRATE_TIMINGS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:\d+.?\d*){1}$`),
				"Tag[0-9A-Za-z_],DestinationRatesTag[0-9A-Za-z_],TimingProfile[0-9A-Za-z_],Weight[0-9.]"},
			&TPCSVRowValidator{utils.RATE_PROFILES_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\d+\s*,\s*){1}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\*all\s*,\s*|[\w:\.]+\s*,\s*){1}(?:\w*\s*,\s*){1}(?:\w+\s*,\s*){1}(?:\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z){1}$`),
				"Tenant[0-9A-Za-z_],TOR[0-9],Direction OUT|IN,Subject[0-9A-Za-z_:.]|*all,RatesFallbackSubject[0-9A-Za-z_]|<empty>,RatesTimingTag[0-9A-Za-z_],ActivationTime[[0-9T:X]] (2012-01-01T00:00:00Z)"},
			&TPCSVRowValidator{utils.ACTIONS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\d+\s*,\s*){1}(?:\w+\s*,\s*|\*all\s*,\s*){1}(?:ABSOLUTE\s*,\s*|PERCENT\s*,\s*|\s*,\s*){1}(?:\d*\.?\d*\s*,?\s*){3}$`),
				"Tag[0-9A-Za-z_],Action[0-9A-Za-z_],BalanceTag[0-9A-Za-z_],Direction OUT|IN,Units[0-9],DestinationTag[0-9A-Za-z_]|*all,PriceType ABSOLUT|PERCENT,PriceValue[0-9.],MinutesWeight[0-9.],Weight[0-9.]"},
			&TPCSVRowValidator{utils.ACTION_TIMINGS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:\d+\.?\d*){1}`),
				"Tag[0-9A-Za-z_],ActionsTag[0-9A-Za-z_],TimingTag[0-9A-Za-z_],Weight[0-9.]"},
			&TPCSVRowValidator{utils.ACTION_TRIGGERS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:MONETARY\s*,\s*|SMS\s*,\s*|MINUTES\s*,\s*|INTERNET\s*,\s*|INTERNET_TIME\s*,\s*){1}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\d+\.?\d*\s*,\s*){1}(?:\w+\s*,\s*|\*all\s*,\s*){1}(?:\w+\s*,\s*){1}(?:\d+\.?\d*){1}$`),
				"Tag[0-9A-Za-z_],BalanceTag MONETARY|SMS|MINUTES|INTERNET|INTERNET_TIME,Direction OUT|IN,ThresholdValue[0-9.],DestinationTag[0-9A-Za-z_]|*all,ActionsTag[0-9A-Za-z_],Weight[0-9.]"},
			&TPCSVRowValidator{utils.ACCOUNT_ACTIONS_CSV,
				regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:[\w:.]+\s*,\s*){1}(?:OUT\s*,\s*|IN\s*,\s*){1}(?:\w+\s*,?\s*){2}$`),
				"Tenant[0-9A-Za-z_],Account[0-9A-Za-z_:.],Direction OUT|IN,ActionTimingsTag[0-9A-Za-z_],ActionTriggersTag[0-9A-Za-z_]"},
		}
