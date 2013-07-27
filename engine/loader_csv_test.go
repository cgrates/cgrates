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

package engine

import (
	//"log"
	"testing"
)

var (
	destinations = `
#Tag,Prefix
GERMANY,49
GERMANY_O2,41
GERMANY_PREMIUM,43
ALL,49
ALL,41
ALL,43
NAT,0256
NAT,0257
NAT,0723
RET,0723
RET,0724
`
	timings = `
WORKDAYS_00,*any,*any,*any,1;2;3;4;5,00:00:00
WORKDAYS_18,*any,*any,*any,1;2;3;4;5,18:00:00
WEEKENDS,*any,*any,*any,6;7,00:00:00
ONE_TIME_RUN,2012,,,,*asap
`
	rates = `
R1,0,0.2,60,1,0,*middle,2,10
R2,0,0.1,60,1,0,*middle,2,10
R3,0,0.05,60,1,0,*middle,2,10
R4,1,1,1,1,0,*up,2,10
R5,0,0.5,1,1,0,*down,2,10
`
	destinationRates = `
RT_STANDARD,GERMANY,R1
RT_STANDARD,GERMANY_O2,R2
RT_STANDARD,GERMANY_PREMIUM,R2
RT_DEFAULT,ALL,R2
RT_STD_WEEKEND,GERMANY,R2
RT_STD_WEEKEND,GERMANY_O2,R3
P1,NAT,R4
P2,NAT,R5
`
	destinationRateTimings = `
STANDARD,RT_STANDARD,WORKDAYS_00,10
STANDARD,RT_STD_WEEKEND,WORKDAYS_18,10
STANDARD,RT_STD_WEEKEND,WEEKENDS,10
PREMIUM,RT_STD_WEEKEND,WEEKENDS,10
DEFAULT,RT_DEFAULT,WORKDAYS_00,10
EVENING,P1,WORKDAYS_00,10
EVENING,P2,WORKDAYS_18,10
EVENING,P2,WEEKENDS,10
`
	ratingProfiles = `
CUSTOMER_1,0,*out,rif:from:tm,2012-01-01T00:00:00Z,PREMIUM,danb
CUSTOMER_1,0,*out,rif:from:tm,2012-02-28T00:00:00Z,STANDARD,danb
CUSTOMER_2,0,*out,danb:87.139.12.167,2012-01-01T00:00:00Z,STANDARD,danb
CUSTOMER_1,0,*out,danb,2012-01-01T00:00:00Z,PREMIUM,
vdf,0,*out,rif,2012-01-01T00:00:00Z,EVENING,
vdf,0,*out,rif,2012-02-28T00:00:00Z,EVENING,
vdf,0,*out,minu,2012-01-01T00:00:00Z,EVENING,
vdf,0,*out,*any,2012-02-28T00:00:00Z,EVENING,
vdf,0,*out,one,2012-02-28T00:00:00Z,STANDARD,
vdf,0,*out,inf,2012-02-28T00:00:00Z,STANDARD,inf
vdf,0,*out,fall,2012-02-28T00:00:00Z,PREMIUM,rif
`
	actions = `
MINI,TOPUP,MINUTES,*out,100,2013-07-19T13:03:22Z,NAT,*absolute,0,10,10
`
	actionTimings = `
MORE_MINUTES,MINI,ONE_TIME_RUN,10
`
	actionTriggers = `
STANDARD_TRIGGER,MINUTES,*out,*min_counter,10,GERMANY_O2,SOME_1,10
STANDARD_TRIGGER,MINUTES,*out,*max_balance,200,GERMANY,SOME_2,10
`
	accountActions = `
vdf,minitsboy,*out,MORE_MINUTES,STANDARD_TRIGGER
`
)

var csvr *CSVReader

func init() {
	csvr = NewStringCSVReader(storageGetter, ',', destinations, timings, rates, destinationRates, destinationRateTimings, ratingProfiles, actions, actionTimings, actionTriggers, accountActions)
	csvr.LoadDestinations()
	csvr.LoadTimings()
	csvr.LoadRates()
	csvr.LoadDestinationRates()
	csvr.LoadDestinationRateTimings()
	csvr.LoadRatingProfiles()
	csvr.LoadActions()
	csvr.LoadActionTimings()
	csvr.LoadActionTriggers()
	csvr.LoadAccountActions()
	csvr.WriteToDatabase(false, false)
}

func TestLoadDestinations(t *testing.T) {
	if len(csvr.destinations) != 6 {
		t.Error("Failed to load destinations: ", csvr.destinations)
	}
}

func TestLoadTimimgs(t *testing.T) {
	if len(csvr.timings) != 4 {
		t.Error("Failed to load timings: ", csvr.timings)
	}
}

func TestLoadRates(t *testing.T) {
	if len(csvr.rates) != 5 {
		t.Error("Failed to load rates: ", csvr.rates)
	}
}

func TestLoadDestinationRates(t *testing.T) {
	if len(csvr.destinationRates) != 5 {
		t.Error("Failed to load rates: ", csvr.rates)
	}
}

func TestLoadDestinationRateTimings(t *testing.T) {
	if len(csvr.activationPeriods) != 4 {
		t.Error("Failed to load rate timings: ", csvr.activationPeriods)
	}
}

func TestLoadRatingProfiles(t *testing.T) {
	if len(csvr.ratingProfiles) != 9 {
		t.Error("Failed to load rating profiles: ", len(csvr.ratingProfiles), csvr.ratingProfiles)
	}
}

func TestLoadActions(t *testing.T) {
	if len(csvr.actions) != 1 {
		t.Error("Failed to load actions: ", csvr.actions)
	}
}

func TestLoadActionTimings(t *testing.T) {
	if len(csvr.actionsTimings) != 1 {
		t.Error("Failed to load action timings: ", csvr.actionsTimings)
	}
}

func TestLoadActionTriggers(t *testing.T) {
	if len(csvr.actionsTriggers) != 1 {
		t.Error("Failed to load action triggers: ", csvr.actionsTriggers)
	}
}

func TestLoadAccountActions(t *testing.T) {
	if len(csvr.accountActions) != 1 {
		t.Error("Failed to load account actions: ", csvr.accountActions)
	}
}
