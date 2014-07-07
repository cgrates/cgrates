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

package utils

import (
	"regexp"
	"testing"
)

func TestProcessReSearchReplace(t *testing.T) {
	rsr := &ReSearchReplace{SearchRegexp: regexp.MustCompile(`sip:\+49(\d+)@(\d*\.\d*\.\d*\.\d*)`), ReplaceTemplate: "0$1@$2"}
	source := "<sip:+4986517174963@127.0.0.1;transport=tcp>"
	expectOut := "086517174963@127.0.0.1"
	if outStr := rsr.Process(source); outStr != expectOut {
		t.Error("Unexpected output from SearchReplace: ", outStr)
	}
}

func TestProcessReSearchReplace2(t *testing.T) {
	rsr := &ReSearchReplace{SearchRegexp: regexp.MustCompile(`(\d+)`), ReplaceTemplate: "+$1"}
	source := "4986517174963"
	expectOut := "+4986517174963"
	if outStr := rsr.Process(source); outStr != expectOut {
		t.Error("Unexpected output from SearchReplace: ", outStr)
	}
}

func TestProcessReSearchReplace3(t *testing.T) { //"MatchedDestId":"CST_31800_DE080"
	rsr := &ReSearchReplace{SearchRegexp: regexp.MustCompile(`"MatchedDestId":".+_(\w{5})"`), ReplaceTemplate: "$1"}
	source := `[{"TimeStart":"2014-04-15T22:17:57+02:00","TimeEnd":"2014-04-15T22:18:01+02:00","Cost":0,"RateInterval":{"Timing":{"Years":[],"Months":[],"MonthDays":[],"WeekDays":[],"StartTime":"00:00:00","EndTime":""},"Rating":{"ConnectFee":0,"Rates":[{"GroupIntervalStart":0,"Value":0,"RateIncrement":1000000000,"RateUnit":60000000000}],"RoundingMethod":"*middle","RoundingDecimals":4},"Weight":10},"CallDuration":4000000000,"Increments":null,"MatchedSubject":"*out:sip.test.cgrates.org:call:*any","MatchedPrefix":"+49800","MatchedDestId":"CST_31800_DE080"}]`
	expectOut := "DE080"
	if outStr := rsr.Process(source); outStr != expectOut {
		t.Error("Unexpected output from SearchReplace: ", outStr)
	}
}

func TestProcessReSearchReplace4(t *testing.T) {
	rsr := &ReSearchReplace{SearchRegexp: regexp.MustCompile(`^\+49(\d+)`), ReplaceTemplate: "0$1"}
	if outStr := rsr.Process("+4986517174963"); outStr != "086517174963" {
		t.Error("Unexpected output from SearchReplace: ", outStr)
	}
	if outStr := rsr.Process("+186517174963"); outStr != "+186517174963" {
		t.Error("Unexpected output from SearchReplace: ", outStr)
	}
}
