// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"reflect"
	"regexp"
	"testing"
)

func TestProcessReSearchReplaceNil(t *testing.T) {
	rsr := &ReSearchReplace{SearchRegexp: nil, ReplaceTemplate: "0$1@$2"}
	source := "<sip:+4986517174963@127.0.0.1;transport=tcp>"
	expectOut := ""
	if outStr := rsr.Process(source); outStr != expectOut {
		t.Error("Unexpected output from SearchReplace: ", outStr)
	}
}

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

func TestProcessReSearchReplace5(t *testing.T) {
	rsr := &ReSearchReplace{SearchRegexp: regexp.MustCompile(`^(.*)_`), ReplaceTemplate: "$1"}
	if outStr := rsr.Process("TEST_EVENT"); outStr != "TEST" {
		t.Error("Unexpected output from SearchReplace: ", outStr)
	}
}

func TestProcessReSearchReplace6(t *testing.T) {
	rsr := &ReSearchReplace{SearchRegexp: regexp.MustCompile(`(.*)`), ReplaceTemplate: "${1}_suffix"}
	if outStr := rsr.Process("call"); outStr != "call_suffix" {
		t.Error("Unexpected output from SearchReplace: ", outStr)
	}
}

func TestReSearchReplaceClone(t *testing.T) {
	rsr := &ReSearchReplace{
		SearchRegexp:    regexp.MustCompile(`(\d+)`),
		ReplaceTemplate: EmptyString,
	}
	rcv := rsr.Clone()
	if !reflect.DeepEqual(rsr, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", ToJSON(rsr), ToJSON(rcv))
	}
	*rcv.SearchRegexp = regexp.Regexp{}
	if reflect.DeepEqual(rsr.SearchRegexp, rcv.SearchRegexp) {
		t.Errorf("Expected clone to not modify the cloned")
	}

	rsr = nil
	rcv = rsr.Clone()
	if !reflect.DeepEqual(rsr, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", ToJSON(rsr), ToJSON(rcv))
	}
}
