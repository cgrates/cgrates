/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package main

import (
	"testing"
)

var (
	dest = `
Tag,Prefix
NAT,0256
NAT,0257
NAT,0723
RET,0723
RET,0724
`
	rts = `
P1,NAT,0,1,1
P2,NAT,0,0.5,1
`
	ts = `
WORKDAYS_00,*all,*all,1;2;3;4;5,00:00:00
WORKDAYS_18,*all,*all,1;2;3;4;5,18:00:00
WEEKENDS,*all,*all,6;7,00:00:00
ONE_TIME_RUN,*none,*none,*none,*now
`
	rtts = `
EVENING,P1,WORKDAYS_00,10
EVENING,P2,WORKDAYS_18,10
EVENING,P2,WEEKENDS,10
`
	rp = `
vdf,0,OUT,rif,,EVENING,2012-01-01T00:00:00Z
vdf,0,OUT,rif,,EVENING,2012-02-28T00:00:00Z
vdf,0,OUT,minu,,EVENING,2012-01-01T00:00:00Z
vdf,0,OUT,minu,,EVENING,2012-02-28T00:00:00Z
`
	a = `
MINI,TOPUP,MINUTES,OUT,100,NAT,ABSOLUTE,0,10,10
`
	atms = `
MORE_MINUTES,MINI,ONE_TIME_RUN,10
`
	atrs = `
STANDARD_TRIGGER,MINUTES,OUT,10,GERMANY_O2,SOME_1,10
STANDARD_TRIGGER,MINUTES,OUT,200,GERMANY,SOME_2,10
`
	accs = `
vdf,minitsboy,OUT,MORE_MINUTES,STANDARD_TRIGGER
`
)

func TestDestinations(t *testing.T) {
	csvr := &CSVReader{openStringCSVReader}
	csvr.loadDestinations(dest)
	if len(destinations) != 2 {
		t.Error("Failed to load destinations: ", destinations)
	}
}

func TestRates(t *testing.T) {
	csvr := &CSVReader{openStringCSVReader}
	csvr.loadRates(rts)
	if len(rates) != 2 {
		t.Error("Failed to load rates: ", rates)
	}
}

func TestTimimgs(t *testing.T) {
	csvr := &CSVReader{openStringCSVReader}
	csvr.loadTimings(ts)
	if len(timings) != 4 {
		t.Error("Failed to load timings: ", timings)
	}
}

func TestRateTimings(t *testing.T) {
	csvr := &CSVReader{openStringCSVReader}
	csvr.loadRateTimings(rtts)
	if len(activationPeriods) != 1 {
		t.Error("Failed to load rate timings: ", activationPeriods)
	}
}

func TestRatingProfiles(t *testing.T) {
	csvr := &CSVReader{openStringCSVReader}
	csvr.loadRatingProfiles(rp)
	if len(ratingProfiles) != 4 {
		t.Error("Failed to load rating profiles: ", ratingProfiles)
	}
}

func TestActions(t *testing.T) {
	csvr := &CSVReader{openStringCSVReader}
	csvr.loadActions(a)
	if len(actions) != 1 {
		t.Error("Failed to load actions: ", actions)
	}
}

func TestActionTimings(t *testing.T) {
	csvr := &CSVReader{openStringCSVReader}
	csvr.loadActionTimings(atms)
	if len(actionsTimings) != 1 {
		t.Error("Failed to load action timings: ", actionsTimings)
	}
}

func TestActionTriggers(t *testing.T) {
	csvr := &CSVReader{openStringCSVReader}
	csvr.loadActionTriggers(atrs)
	if len(actionsTriggers) != 1 {
		t.Error("Failed to load action triggers: ", actionsTriggers)
	}
}

func TestAccountActions(t *testing.T) {
	csvr := &CSVReader{openStringCSVReader}
	csvr.loadAccountActions(accs)
	if len(accountActions) != 1 {
		t.Error("Failed to load account actions: ", accountActions)
	}
}
