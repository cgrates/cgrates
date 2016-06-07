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
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
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
		StrategyParams: "1.2;2.3;4;7;45s;67m;16s;17m;8.9;10.11;12.13;14.15;2;3",
	}
	minAsr, maxAsr, minPdd, maxPdd, minAcd, maxAcd, minTcd, maxTcd, minAcc, maxAcc, minTcc, maxTcc, minDdc, maxDdc := le.GetQOSLimits()
	if minAsr != 1.2 || maxAsr != 2.3 ||
		minPdd != 4*time.Second || maxPdd != 7*time.Second ||
		minAcd != 45*time.Second || maxAcd != 67*time.Minute ||
		minTcd != 16*time.Second || maxTcd != 17*time.Minute ||
		minAcc != 8.9 || maxAcc != 10.11 ||
		minTcc != 12.13 || maxTcc != 14.15 ||
		minDdc != 2 || maxDdc != 3 {
		t.Error("Wrong qos limits parsed: ", minAsr, maxAsr, minPdd, maxPdd, minAcd, maxAcd, minTcd, maxTcd, minAcc, maxAcc, minTcc, maxTcc, minDdc, maxDdc)
	}
}

func TestLcrGetQosLimitsSome(t *testing.T) {
	le := &LCREntry{
		StrategyParams: "1.2;;3;;;67m;;30m;1;;3;;;2",
	}
	minAsr, maxAsr, minPdd, maxPdd, minAcd, maxAcd, minTcd, maxTcd, minAcc, maxAcc, minTcc, maxTcc, minDdc, maxDdc := le.GetQOSLimits()
	if minAsr != 1.2 || maxAsr != -1 ||
		minPdd != 3*time.Second || maxPdd != -1 ||
		minAcd != -1 || maxAcd != 67*time.Minute ||
		minTcd != -1 || maxTcd != 30*time.Minute ||
		minAcc != 1 || maxAcc != -1 ||
		minTcc != 3 || maxTcc != -1 ||
		minDdc != -1 || maxDdc != 2 {
		t.Error("Wrong qos limits parsed: ", minAsr, maxAsr, minAcd, maxAcd, minTcd, maxTcd, minTcc, maxTcc, minDdc, maxDdc)
	}
}

func TestLcrGetQosLimitsNone(t *testing.T) {
	le := &LCREntry{
		StrategyParams: ";;;;;;;;;;;",
	}
	minAsr, maxAsr, minPdd, maxPdd, minAcd, maxAcd, minTcd, maxTcd, minAcc, maxAcc, minTcc, maxTcc, minDdc, maxDdc := le.GetQOSLimits()
	if minAsr != -1 || maxAsr != -1 ||
		minPdd != -1 || maxPdd != -1 ||
		minAcd != -1 || maxAcd != -1 ||
		minTcd != -1 || maxTcd != -1 ||
		minAcc != -1 || maxAcc != -1 ||
		minTcc != -1 || maxTcc != -1 ||
		minDdc != -1 || maxDdc != -1 {
		t.Error("Wrong qos limits parsed: ", minAsr, maxAsr, minAcd, maxAcd, minDdc, maxDdc)
	}
}

func TestLcrGetQosSortParamsNone(t *testing.T) {
	le := &LCREntry{
		Strategy:       LCR_STRATEGY_QOS,
		StrategyParams: "",
	}
	sort := le.GetParams()
	if sort[0] != ASR || sort[1] != PDD || sort[2] != ACD || sort[3] != TCD || sort[4] != ACC || sort[5] != TCC {
		t.Error("Wrong qos sort params parsed: ", sort)
	}
}

func TestLcrGetQosSortParamsEmpty(t *testing.T) {
	le := &LCREntry{
		Strategy:       LCR_STRATEGY_QOS,
		StrategyParams: ";;;;",
	}
	sort := le.GetParams()
	if sort[0] != ASR || sort[1] != PDD || sort[2] != ACD || sort[3] != TCD || sort[4] != ACC || sort[5] != TCC {
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
	if sort[0] != ASR || sort[1] != PDD {
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
	lcr, err := cd.GetLCR(nil, nil, nil)
	if err != nil || lcr == nil {
		t.Errorf("Bad lcr: %+v, %v", lcr, err)
	}
}

func TestLcrGetPrefix(t *testing.T) {
	lcrSubjectPrefixMatching = true
	cd := &CallDescriptor{
		TimeStart:   time.Date(2015, 04, 06, 17, 40, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 04, 06, 17, 41, 0, 0, time.UTC),
		Tenant:      "cgrates.org",
		Direction:   "*in",
		Category:    "call",
		Destination: "0723098765",
		Account:     "rif",
		Subject:     "rifus",
	}
	lcr, err := cd.GetLCR(nil, nil, nil)
	if err != nil || lcr == nil {
		t.Errorf("Bad lcr: %+v, %v", lcr, err)
	}
}

func TestLcrRequestAsCallDescriptor(t *testing.T) {
	sTime := time.Date(2015, 04, 06, 17, 40, 0, 0, time.UTC)
	callDur := time.Duration(1) * time.Minute
	lcrReq := &LcrRequest{Account: "2001", SetupTime: sTime.String()}
	if _, err := lcrReq.AsCallDescriptor(""); err == nil || err != utils.ErrMandatoryIeMissing {
		t.Error("Unexpected error received: %v", err)
	}
	lcrReq = &LcrRequest{Account: "2001", Destination: "2002", SetupTime: sTime.String()}
	eCd := &CallDescriptor{
		Direction:   utils.OUT,
		Tenant:      config.CgrConfig().DefaultTenant,
		Category:    config.CgrConfig().DefaultCategory,
		Account:     lcrReq.Account,
		Subject:     lcrReq.Account,
		Destination: lcrReq.Destination,
		TimeStart:   sTime,
		TimeEnd:     sTime.Add(callDur),
	}
	if cd, err := lcrReq.AsCallDescriptor(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCd, cd) {
		t.Errorf("Expected: %+v, received: %+v", eCd, cd)
	}
}

func TestLCRCostSuppliersSlice(t *testing.T) {
	lcrCost := new(LCRCost)
	if _, err := lcrCost.SuppliersString(); err == nil || err != utils.ErrNotFound {
		t.Errorf("Unexpected error received: %v", err)
	}
	lcrCost = &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_STATIC, StrategyParams: "ivo12;dan12;rif12", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:tenant12:call:ivo12", Cost: 1.8, Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:rif12", Cost: 1.2, Duration: 60 * time.Second},
		},
	}
	eSuppls := []string{"ivo12", "dan12", "rif12"}
	if suppls, err := lcrCost.SuppliersSlice(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSuppls, suppls) {
		t.Errorf("Expecting: %+v, received: %+v", eSuppls, suppls)
	}
}

func TestLCRCostSuppliersString(t *testing.T) {
	lcrCost := new(LCRCost)
	if _, err := lcrCost.SuppliersString(); err == nil || err != utils.ErrNotFound {
		t.Errorf("Unexpected error received: %v", err)
	}
	lcrCost = &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_STATIC, StrategyParams: "ivo12;dan12;rif12", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:tenant12:call:ivo12", Cost: 1.8, Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:rif12", Cost: 1.2, Duration: 60 * time.Second},
		},
	}
	eSupplStr := "ivo12,dan12,rif12"
	if supplStr, err := lcrCost.SuppliersString(); err != nil {
		t.Error(err)
	} else if supplStr != eSupplStr {
		t.Errorf("Expecting: %s, received: %s", eSupplStr, supplStr)
	}
}

func TestLCRCostSuppliersLoad(t *testing.T) {
	setupTime := time.Date(2015, 7, 31, 6, 43, 0, 0, time.UTC)
	lcrCost := &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOAD, StrategyParams: "ivo12:10;dan12:3;*default:7", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:ivo12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  3 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  1 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
				},
			},
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:dan12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
				},
			},
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:rif12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  1 * time.Minute,
						},
					},
				},
			},
		},
	}
	lcrCost.Sort()
	if lcrCost.SupplierCosts[0].Supplier != "*out:tenant12:call:dan12" {
		t.Error("Error soring on load distribution: ", utils.ToIJSON(lcrCost))
	}
}

func TestLCRCostSuppliersLoadAllRounded(t *testing.T) {
	setupTime := time.Date(2015, 7, 31, 6, 43, 0, 0, time.UTC)
	lcrCost := &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOAD, StrategyParams: "ivo12:3;dan12:5;*default:2", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:ivo12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  3 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  1 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
				},
			},
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:dan12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime.Add(200 * time.Minute)}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
				},
			},
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:rif12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime.Add(400 * time.Minute)}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  1 * time.Minute,
						},
					},
				},
			},
		},
	}
	lcrCost.Sort()
	if lcrCost.SupplierCosts[0].Supplier != "*out:tenant12:call:ivo12" ||
		lcrCost.SupplierCosts[1].Supplier != "*out:tenant12:call:dan12" ||
		lcrCost.SupplierCosts[2].Supplier != "*out:tenant12:call:rif12" {
		t.Error("Error soring on load distribution: ", utils.ToIJSON(lcrCost))
	}
}

func TestLCRCostSuppliersLoadAllOver(t *testing.T) {
	setupTime := time.Date(2015, 7, 31, 6, 43, 0, 0, time.UTC)
	lcrCost := &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOAD, StrategyParams: "ivo12:2;dan12:4;*default:2", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:ivo12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  3 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  1 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
				},
			},
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:dan12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime.Add(200 * time.Minute)}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
				},
			},
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:rif12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime.Add(400 * time.Minute)}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  1 * time.Minute,
						},
					},
				},
			},
		},
	}
	lcrCost.Sort()
	if lcrCost.SupplierCosts[0].Supplier != "*out:tenant12:call:ivo12" ||
		lcrCost.SupplierCosts[1].Supplier != "*out:tenant12:call:dan12" ||
		lcrCost.SupplierCosts[2].Supplier != "*out:tenant12:call:rif12" {
		t.Error("Error soring on load distribution: ", utils.ToIJSON(lcrCost))
	}
}

func TestLCRCostSuppliersLoadAllOverMisingDefault(t *testing.T) {
	setupTime := time.Date(2015, 7, 31, 6, 43, 0, 0, time.UTC)
	lcrCost := &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOAD, StrategyParams: "ivo12:2;dan12:4", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:ivo12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  3 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  1 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
				},
			},
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:dan12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime.Add(200 * time.Minute)}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
				},
			},
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:rif12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime.Add(400 * time.Minute)}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  1 * time.Minute,
						},
					},
				},
			},
		},
	}
	lcrCost.Sort()
	if len(lcrCost.SupplierCosts) != 2 ||
		lcrCost.SupplierCosts[0].Supplier != "*out:tenant12:call:ivo12" ||
		lcrCost.SupplierCosts[1].Supplier != "*out:tenant12:call:dan12" {
		t.Error("Error soring on load distribution: ", utils.ToIJSON(lcrCost))
	}
}

func TestLCRCostSuppliersLoadAllOverMisingParams(t *testing.T) {
	setupTime := time.Date(2015, 7, 31, 6, 43, 0, 0, time.UTC)
	lcrCost := &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOAD, StrategyParams: "", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:ivo12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  3 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  1 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
				},
			},
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:dan12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime.Add(60 * time.Minute)}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
				},
			},
			&LCRSupplierCost{
				Supplier: "*out:tenant12:call:rif12",
				supplierQueues: []*StatsQueue{
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  7 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{}, &QCdr{SetupTime: setupTime.Add(200 * time.Minute)}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  10 * time.Minute,
						},
					},
					&StatsQueue{
						Cdrs: []*QCdr{&QCdr{}, &QCdr{SetupTime: setupTime}},
						conf: &CdrStats{
							QueueLength: 0,
							TimeWindow:  1 * time.Minute,
						},
					},
				},
			},
		},
	}
	lcrCost.Sort()
	if len(lcrCost.SupplierCosts) != 3 {
		t.Error("Error soring on load distribution: ", utils.ToIJSON(lcrCost))
	}
}
