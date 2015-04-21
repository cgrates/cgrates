/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"sort"
	"testing"
	"time"
)

func TestLcrQOSSorter(t *testing.T) {
	s := QOSSorter{
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 3,
				"ACD": 3,
			},
			qosSortParams: []string{ASR, ACD},
		},
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 1,
				"ACD": 1,
			},
			qosSortParams: []string{ASR, ACD},
		},
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 2,
				"ACD": 2,
			},
			qosSortParams: []string{ASR, ACD},
		},
	}
	sort.Sort(s)
	if s[0].QOS[ASR] != 3 ||
		s[1].QOS[ASR] != 2 ||
		s[2].QOS[ASR] != 1 {
		t.Error("Lcr qos sort failed: ", s)
	}
}

func TestLcrQOSSorterOACD(t *testing.T) {
	s := QOSSorter{
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 1,
				"ACD": 3,
			},
			qosSortParams: []string{ASR, ACD},
		},
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 1,
				"ACD": 1,
			},
			qosSortParams: []string{ASR, ACD},
		},
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 1,
				"ACD": 2,
			},
			qosSortParams: []string{ASR, ACD},
		},
	}
	sort.Sort(s)
	if s[0].QOS[ACD] != 3 ||
		s[1].QOS[ACD] != 2 ||
		s[2].QOS[ACD] != 1 {
		t.Error("Lcr qos sort failed: ", s)
	}
}

func TestLcrGetQosLimitsAll(t *testing.T) {
	le := &LCREntry{
		StrategyParams: "1.2;2.3;45s;67m",
	}
	minAsr, maxAsr, minAcd, maxAcd := le.GetQOSLimits()
	if minAsr != 1.2 || maxAsr != 2.3 ||
		minAcd != 45*time.Second || maxAcd != 67*time.Minute {
		t.Error("Wrong qos limits parsed: ", minAsr, maxAsr, minAcd, maxAcd)
	}
}

func TestLcrGetQosLimitsSome(t *testing.T) {
	le := &LCREntry{
		StrategyParams: "1.2;;;67m",
	}
	minAsr, maxAsr, minAcd, maxAcd := le.GetQOSLimits()
	if minAsr != 1.2 || maxAsr != -1 ||
		minAcd != -1 || maxAcd != 67*time.Minute {
		t.Error("Wrong qos limits parsed: ", minAsr, maxAsr, minAcd, maxAcd)
	}
}

func TestLcrGetQosLimitsNone(t *testing.T) {
	le := &LCREntry{
		StrategyParams: ";;;",
	}
	minAsr, maxAsr, minAcd, maxAcd := le.GetQOSLimits()
	if minAsr != -1 || maxAsr != -1 ||
		minAcd != -1 || maxAcd != -1 {
		t.Error("Wrong qos limits parsed: ", minAsr, maxAsr, minAcd, maxAcd)
	}
}

func TestLcrGetQosSortParamsNone(t *testing.T) {
	le := &LCREntry{
		Strategy:       LCR_STRATEGY_QOS,
		StrategyParams: "",
	}
	sort := le.GetParams()
	if sort[0] != ASR || sort[1] != ACD {
		t.Error("Wrong qos sort params parsed: ", sort)
	}
}

func TestLcrGetQosSortParamsEmpty(t *testing.T) {
	le := &LCREntry{
		Strategy:       LCR_STRATEGY_QOS,
		StrategyParams: ";",
	}
	sort := le.GetParams()
	if sort[0] != ASR || sort[1] != ACD {
		t.Error("Wrong qos sort params parsed: ", sort)
	}
}

func TestLcrGetQosSortParamsOne(t *testing.T) {
	le := &LCREntry{
		StrategyParams: "ACD",
	}
	sort := le.GetParams()
	if sort[0] != ACD || len(sort) != 1 {
		t.Error("Wrong qos sort params parsed: ", sort)
	}
}

func TestLcrGetQosSortParamsSpace(t *testing.T) {
	le := &LCREntry{
		Strategy:       LCR_STRATEGY_QOS,
		StrategyParams: "; ",
	}
	sort := le.GetParams()
	if sort[0] != ASR || sort[1] != ACD {
		t.Error("Wrong qos sort params parsed: ", sort)
	}
}

func TestLcrGetQosSortParamsFull(t *testing.T) {
	le := &LCREntry{
		StrategyParams: "ACD;ASR",
	}
	sort := le.GetParams()
	if sort[0] != ACD || sort[1] != ASR {
		t.Error("Wrong qos sort params parsed: ", sort)
	}
}

func TestLcrGet(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2015, 04, 06, 17, 40, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 04, 06, 17, 41, 0, 0, time.UTC),
		Tenant:      "cgrates.org",
		Direction:   "*in",
		Category:    "call",
		Destination: "0723098765",
		Account:     "rif",
		Subject:     "rif",
	}
	lcr, err := cd.GetLCR(nil)
	//jsn, _ := json.Marshal(lcr)
	//log.Print("LCR: ", string(jsn))
	if err != nil || lcr == nil {
		t.Errorf("Bad lcr: %+v, %v", lcr, err)
	}
}
