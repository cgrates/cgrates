/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/baningo"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestRatesCostFiltering(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	cgrEv := &utils.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestEv1",
			Event: map[string]any{
				"Account":     "1001",
				"Destination": "1002",
				"OriginID":    "TestEv1",
				"RequestType": "*prepaid",
				"Subject":     "1001",
				"ToR":         "*voice",
				"Usage":       60000000000,
			},
			APIOpts: map[string]any{
				utils.MetaRateSCost: map[string]any{
					utils.Cost: 0.4,
					"CostIntervals": []map[string]any{
						{
							"CompressFactor": 1,
							"Increments": []map[string]any{
								{
									"CompressFactor":    2,
									"RateID":            "Rate1",
									"RateIntervalIndex": 0,
									"Usage":             60000000000,
								},
							},
						},
					},
					"ID": "RT_RETAIL1",
					"Rates": map[string]any{
						"Rate1": map[string]any{
							"Increment":     30000000000,
							"IntervalStart": 0,
							"RecurrentFee":  0.4,
							"Unit":          60000000000,
						},
					},
				},
				utils.MetaRates:      true,
				utils.OptsCDRsExport: true,
				utils.MetaAccounts:   false,
			},
		},
	}
	cgrDP := cgrEv.AsDataProvider()
	if pass, err := filterS.Pass(context.Background(), "cgrates.org", []string{"*gt:~*opts.*rateSCost.Cost:0"}, cgrDP); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expected to pass")
	}
	if pass, err := filterS.Pass(context.Background(), "cgrates.org", []string{"*gt:~*opts.*rateSCost.Cost:0.5"}, cgrDP); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expected to fail")
	}
	if pass, err := filterS.Pass(context.Background(), "cgrates.org", []string{"*string:~*opts.*rateSCost.CostIntervals[0].Increments[0].RateID:Rate1"}, cgrDP); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expected to pass")
	}
	if pass, err := filterS.Pass(context.Background(), "cgrates.org", []string{"*eq:~*opts.*rateSCost.Rates[Rate1].RecurrentFee:0.4"}, cgrDP); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expected to pass")
	}
}

func TestFilterPassGreaterThan(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaLessThan, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	ev := utils.MapStorage{}
	ev.Set([]string{"ASR"}, 20)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 40)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("equal should not be passing")
	}
	rf, err = NewFilterRule(utils.MetaLessOrEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ASR", []string{"35.5"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 20)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("should not pass")
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~ACD", []string{"1m50s"})
	if err != nil {
		t.Error(err)
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ACD"}, 2*time.Minute)
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not pass")
	}
	// Second
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 20*time.Second)
	rf, err = NewFilterRule("*gte", "~ASR", []string{"10s"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("passing")
	}

	rf, err = NewFilterRule("*gte", "~ASR", []string{"10"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("passing")
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, float64(20*time.Second))
	rf, err = NewFilterRule("*gte", "~ASR", []string{"10s"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("passing")
	}
	//Here converter will be consider part of path and will get error : NOT_FOUND
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 20)
	rf, err = NewFilterRule("*gte", "~ASR{*duration_seconds}", []string{"10"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("passing")
	}
	//Here converter will be consider part of path and will get error : NOT_FOUND
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 20)
	rf, err = NewFilterRule("*gte", "~ASR{*duration_seconds}", []string{"10s"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passGreaterThan(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("passing")
	}
}

func TestFilterPassContains(t *testing.T) {
	rF, err := NewFilterRule(utils.MetaContains, "~Subject", []string{"Any"})
	if err != nil {
		t.Error(err)
	}
	eV := utils.MapStorage{}
	eV.Set([]string{"Subject"}, "SubAnyAcc")

	if pass, err := rF.passContains(eV); err != nil {
		t.Error(err)
	} else if !pass {
		t.Error("not passing")
	}

	eV.Set([]string{"Subject"}, "1001")
	if pass, err := rF.passContains(eV); err != nil {
		t.Error(err)
	} else if pass {
		t.Error("passes")
	}
	rF, err = NewFilterRule(utils.MetaContains, "~Resource", []string{"Prf"})
	if err != nil {
		t.Error(err)
	}

	eV.Set([]string{"Resource"}, "ResPrf1")

	if pass, err := rF.Pass(context.TODO(), eV); err != nil {
		t.Error(err)
	} else if !pass {
		t.Error("not passing")
	}

}

func TestFilterpassEqualTo(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	ev := utils.MapStorage{}
	ev.Set([]string{"ASR"}, 40.0)
	if passes, err := rf.passEqualTo(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, 39)
	if passes, err := rf.passEqualTo(ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("equal should not be passing")
	}
	rf, err = NewFilterRule(utils.MetaNotEqual, "~ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(context.TODO(), ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing", passes)
	}
	ev = utils.MapStorage{}
	ev.Set([]string{"ASR"}, "string1")
	rf, err = NewFilterRule(utils.MetaEqual, "~ASR", []string{"string1"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passEqualTo(ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("not passing")
	}
}

func TestFilterNewRequestFilter(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaString, "~MetaString", []string{"String"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf := &FilterRule{Type: utils.MetaString, Element: "~MetaString", Values: []string{"String"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaEmpty, "~MetaEmpty", []string{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaEmpty, Element: "~MetaEmpty", Values: []string{}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaExists, "~MetaExists", []string{})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaExists, Element: "~MetaExists", Values: []string{}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaPrefix, "~MetaPrefix", []string{"stringPrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaPrefix, Element: "~MetaPrefix", Values: []string{"stringPrefix"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaSuffix, "~MetaSuffix", []string{"stringSuffix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaSuffix, Element: "~MetaSuffix", Values: []string{"stringSuffix"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaLessThan, "~MetaLessThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaLessThan, Element: "~MetaLessThan", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaLessOrEqual, "~MetaLessOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaLessOrEqual, Element: "~MetaLessOrEqual", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaGreaterThan, "~MetaGreaterThan", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaGreaterThan, Element: "~MetaGreaterThan", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
	rf, err = NewFilterRule(utils.MetaGreaterOrEqual, "~MetaGreaterOrEqual", []string{"20"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaGreaterOrEqual, Element: "~MetaGreaterOrEqual", Values: []string{"20"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}

	rf, err = NewFilterRule(utils.MetaRegex, "~MetaRegex", []string{"Regex"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	erf = &FilterRule{Type: utils.MetaRegex, Element: "~MetaRegex", Values: []string{"Regex"}, negative: utils.BoolPointer(false)}
	if err = erf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(erf, rf) {
		t.Errorf("Expecting: %+v, received: %+v", erf, rf)
	}
}

func TestInlineFilterPassFiltersForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	failEvent := map[string]any{
		"Account": "1001",
	}
	passEvent := map[string]any{
		"Account": "1007",
	}
	fEv := utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*string:~*req.Account:1007:error"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*string:~*req.Account:1007"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*string:~*req.Account:1007"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*notstring:~*req.Account:1007"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	failEvent = map[string]any{
		"Account": "2001",
	}
	passEvent = map[string]any{
		"Account": "1007",
	}
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*prefix:~*req.Account:10"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*prefix:~*req.Account:10"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*suffix:~*req.Account:07"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*suffix:~*req.Account:07"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*notsuffix:~*req.Account:07"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	failEvent = map[string]any{
		"Tenant": "anotherTenant.org",
	}
	passEvent = map[string]any{
		"Tenant": "cgrates.org",
	}
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*rsr:~*req.Tenant:~^cgr.*\\.org$"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*rsr:~*req.Tenant:~^cgr.*\\.org$"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*notrsr:~*req.Tenant:~^cgr.*\\.org$"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	failEvent = map[string]any{
		utils.Weight: 10,
	}
	passEvent = map[string]any{
		utils.Weight: 20,
	}
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*gte:~*req.Weight:20"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*gte:~*req.Weight:10"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	failEvent = map[string]any{
		"EmptyString":   "nonEmpty",
		"EmptySlice":    []string{""},
		"EmptyMap":      map[string]string{"": ""},
		"EmptyPtrSlice": &[]string{""},
		"EmptyPtrMap":   &map[string]string{"": ""},
	}
	var testnil *struct{} = nil
	passEvent = map[string]any{
		"EmptyString":   "",
		"EmptySlice":    []string{},
		"EmptyMap":      map[string]string{},
		"EmptyPtr":      testnil,
		"EmptyPtr2":     nil,
		"EmptyPtr3":     &struct{}{},
		"EmptyPtrSlice": &[]string{},
		"EmptyPtrMap":   &map[string]string{},
	}
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	for key := range failEvent {
		if pass, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*empty:~*req." + key + ":"},
			fEv); err != nil {
			t.Error(err)
		} else if pass {
			t.Errorf("For %s expecting: %+v, received: %+v", key, false, pass)
		}
		if pass, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*empty:~*req." + key + ":"},
			pEv); err != nil {
			t.Error(err)
		} else if !pass {
			t.Errorf("For %s expecting: %+v, received: %+v", key, true, pass)
		}
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*exists:~*req.NewKey:"},
		fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("For NewKey expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*notexists:~*req.NewKey:"},
		fEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("For NewKey expecting: %+v, received: %+v", true, pass)
	}

	failEvent = map[string]any{
		"Account": "1001",
	}
	passEvent = map[string]any{
		"Account": "1007",
	}
	fEv = utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, failEvent)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*regex:~*req.Account:^1007:error"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*regex:~*req.Account:\\d{3}7"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*regex:~*req.Account:\\d{3}7"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	//not
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*notregex:~*req.Account:\\d{3}7"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}

	pEv = utils.MapStorage{utils.MetaReq: utils.MapStorage{utils.AccountField: "sip:12345678901234567@abcdefg"}}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*regex:~*req.Account:.{29,}"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*regex:~*req.Account:^.{28}$"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*gte:~*req.Account{*len}:29"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	pEv = utils.MapStorage{utils.MetaReq: utils.MapStorage{utils.AccountField: "[1,2,3]"}}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*eq:~*req.Account{*slice&*len}:3"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

}

func TestPassFiltersForEventWithEmptyFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	passEvent1 := map[string]any{
		utils.Tenant:       "cgrates.org",
		utils.AccountField: "1010",
		utils.Destination:  "+49",
		utils.Weight:       10,
	}
	passEvent2 := map[string]any{
		utils.Tenant:       "itsyscom.com",
		utils.AccountField: "dan",
		utils.Destination:  "+4986517174963",
		utils.Weight:       20,
	}
	pEv1 := utils.MapStorage{}
	pEv1.Set([]string{utils.MetaReq}, passEvent1)
	pEv2 := utils.MapStorage{}
	pEv2.Set([]string{utils.MetaReq}, passEvent2)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{}, pEv1); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "itsyscom.com",
		[]string{}, pEv2); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
	ev := map[string]any{
		"Test": "MultipleCharacter",
	}
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*rsr:~*req.Test:~^\\w{30,}"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	ev = map[string]any{
		"Test": "MultipleCharacter123456789MoreThan30Character",
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*rsr:~*req.Test:~^\\w{30,}"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	ev = map[string]any{
		"Test": map[string]any{
			"Test2": "MultipleCharacter",
		},
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*rsr:~*req.Test.Test2:~^\\w{30,}"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	ev = map[string]any{
		"Test": map[string]any{
			"Test2": "MultipleCharacter123456789MoreThan30Character",
		},
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*rsr:~*req.Test.Test2:~^\\w{30,}"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}

	ev = map[string]any{
		utils.AccountField: "1003",
		utils.Subject:      "1003",
		utils.Destination:  "1002",
		utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
		utils.Usage:        "1m20s",
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, ev)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{
			"*string:~*req.Account:1003",
			"*prefix:~*req.Destination:10",
			"*suffix:~*req.Subject:03",
			"*rsr:~*req.Destination:1002"},
		pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
}

func TestPassFilterMaxCost(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	//check with max usage -1 should fail
	passEvent1 := map[string]any{
		"MaxUsage": -1,
	}
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent1)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*gt:~*req.MaxUsage:0s"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: false , received: %+v", pass)
	}
	//check with max usage 0 should fail
	passEvent2 := map[string]any{
		"MaxUsage": 0,
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent2)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*gt:~*req.MaxUsage:0s"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: false, received: %+v", pass)
	}
	//check with max usage 123 should pass
	passEvent3 := map[string]any{
		"MaxUsage": 123,
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent3)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*gt:~*req.MaxUsage:0"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true, received: %+v", pass)
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*gt:~*req.MaxUsage:0"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true, received: %+v", pass)
	}
}

func TestPassFilterMissingField(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}

	passEvent1 := map[string]any{
		"test": "call",
	}
	pEv := utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent1)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*rsr:~*req.Category:^$"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}

	passEvent2 := map[string]any{
		"Category": "",
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent2)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*rsr:~*req.Category:^$"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: true , received: %+v", pass)
	}

	passEvent3 := map[string]any{
		"Category": "call",
	}
	pEv = utils.MapStorage{}
	pEv.Set([]string{utils.MetaReq}, passEvent3)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*rsr:~*req.Category:^$"}, pEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: false , received: %+v", pass)
	}
}

func TestVerifyPrefixes(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaString, "~*req.Account", []string{"1001"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixes := []string{utils.DynamicDataPrefix + utils.MetaReq}
	if check := verifyPrefixes(rf, prefixes); !check {
		t.Errorf("Expecting: true , received: %+v", check)
	}

	rf, err = NewFilterRule(utils.MetaString, "~*req.Account", []string{"~*req.Field1"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if check := verifyPrefixes(rf, prefixes); !check {
		t.Errorf("Expecting: true , received: %+v", check)
	}

	rf, err = NewFilterRule(utils.MetaString, "~*req.Account", []string{"~*req.Field1", "~*req.Field2"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if check := verifyPrefixes(rf, prefixes); !check {
		t.Errorf("Expecting: true , received: %+v", check)
	}

	rf, err = NewFilterRule(utils.MetaString, "~*vars.Account", []string{"1001"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if check := verifyPrefixes(rf, prefixes); check {
		t.Errorf("Expecting: false , received: %+v", check)
	}

	rf, err = NewFilterRule(utils.MetaString, "~*req.Account", []string{"~*req.Field1", "~*vars.Field2"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if check := verifyPrefixes(rf, prefixes); check {
		t.Errorf("Expecting: false , received: %+v", check)
	}

	rf, err = NewFilterRule(utils.MetaString, "~*req.Account", []string{"~*req.Field1", "~*vars.Field2"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixes = []string{utils.DynamicDataPrefix + utils.MetaReq, utils.DynamicDataPrefix + utils.MetaVars}
	if check := verifyPrefixes(rf, prefixes); !check {
		t.Errorf("Expecting: true , received: %+v", check)
	}
}

func TestPassPartial(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	passEvent := map[string]any{
		"Account": "1007",
	}
	fEv := utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, passEvent)
	prefixes := []string{utils.DynamicDataPrefix + utils.MetaReq}
	if pass, ruleList, err := filterS.LazyPass(context.Background(), "cgrates.org",
		[]string{"*string:~*req.Account:1007"}, fEv, prefixes); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	} else if len(ruleList) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(ruleList))
	}
	// in PartialPass we verify the filters matching the prefixes
	if pass, ruleList, err := filterS.LazyPass(context.Background(), "cgrates.org",
		[]string{"*string:~*req.Account:1007", "*string:~*vars.Field1:Val1"}, fEv, prefixes); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	} else if len(ruleList) != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, len(ruleList))
	}
	if pass, ruleList, err := filterS.LazyPass(context.Background(), "cgrates.org",
		[]string{"*string:~*req.Account:1010", "*string:~*vars.Field1:Val1"}, fEv, prefixes); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	} else if len(ruleList) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(ruleList))
	}
}

func TestNewFilterFromInline(t *testing.T) {
	exp := &Filter{
		Tenant: "cgrates.org",
		ID:     "*string:~*req.Account:~*uhc.<~*opts.*originID;-Account>|1001",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"~*uhc.<~*opts.*originID;-Account>", "1001"},
			},
		},
	}
	if err := exp.Compile(); err != nil {
		t.Fatal(err)
	}
	if rcv, err := NewFilterFromInline("cgrates.org", "*string:~*req.Account:~*uhc.<~*opts.*originID;-Account>|1001"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	if _, err := NewFilterFromInline("cgrates.org", "*string:~*req.Account"); err == nil {
		t.Error("Expected error received nil")
	}

	if _, err := NewFilterFromInline("cgrates.org", "*string:~*req.Account:~*opts.*originID{*|1001"); err == nil {
		t.Error("Expected error received nil")
	}
}

func TestVerifyInlineFilterS(t *testing.T) {
	if err := verifyInlineFilterS([]string{"ATTR", "*string:~*req,Acoount:1001"}); err != nil {
		t.Error(err)
	}
	if err := verifyInlineFilterS([]string{"ATTR", "*string:~*req,Acoount1001"}); err == nil {
		t.Errorf("Expected error received nil")
	}
}

func TestActivationIntervalPass(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	passEvent := map[string]any{
		"CustomTime": time.Date(2013, time.July, 1, 0, 0, 0, 0, time.UTC),
	}
	fEv := utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, passEvent)
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*ai:~*req.CustomTime:2013-06-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*ai:~*req.CustomTime:2013-09-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}

	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*ai:~*req.CustomTime:|2013-09-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*ai:~*req.CustomTime:|2013-06-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}

	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*ai:~*req.CustomTime:2013-06-01T00:00:00Z|2013-09-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}

	if pass, err := filterS.Pass(context.TODO(), "cgrates.org",
		[]string{"*ai:~*req.CustomTime:2013-08-01T00:00:00Z|2013-09-01T00:00:00Z"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
}

func TestFilterPassIPNet(t *testing.T) {
	cd := utils.MapStorage{
		"IP":      "192.0.2.0",
		"WrongIP": "192.0.3.",
	}
	rf := &FilterRule{Type: utils.MetaIPNet,
		Element: "~IP", Values: []string{"192.0.2.1/24"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passIPNet(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &FilterRule{Type: utils.MetaIPNet,
		Element: "~IP", Values: []string{"~IP2", "192.0.3.0/30"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passIPNet(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	//not
	rf = &FilterRule{Type: utils.MetaNotIPNet,
		Element: "~IP", Values: []string{"192.0.2.0/24"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.Pass(context.TODO(), cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	rf = &FilterRule{Type: utils.MetaNotIPNet,
		Element: "~IP", Values: []string{"192.0.3.0/24"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.Pass(context.TODO(), cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}

	rf = &FilterRule{Type: utils.MetaIPNet,
		Element: "~IP2", Values: []string{"192.0.2.0/24"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passIPNet(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}

	rf = &FilterRule{Type: utils.MetaIPNet,
		Element: "~IP", Values: []string{"192.0.2.0"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passIPNet(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}

	rf = &FilterRule{Type: utils.MetaIPNet,
		Element: "~WrongIP", Values: []string{"192.0.2.0"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passIPNet(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
	rf = &FilterRule{Type: utils.MetaIPNet,
		Element: "~IP{*duration}", Values: []string{"192.0.2.0/24"}}
	if err := rf.CompileValues(); err != nil {
		t.Fatal(err)
	}
	if _, err := rf.Pass(context.TODO(), cd); err == nil {
		t.Error(err)
	}
}

func TestAPIBan(t *testing.T) {
	var counter int
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		responses := map[string]struct {
			code int
			body []byte
		}{
			"/testKey/check/1.2.3.251": {code: http.StatusOK, body: []byte(`{"ipaddress":["1.2.3.251"], "ID":"987654321"}`)},
			"/testKey/check/1.2.3.254": {code: http.StatusBadRequest, body: []byte(`{"ipaddress":["not blocked"], "ID":"none"}`)},
		}
		if val, has := responses[r.URL.EscapedPath()]; has {
			w.WriteHeader(val.code)
			if val.body != nil {
				w.Write(val.body)
			}
			return
		}
		counter++
		w.WriteHeader(http.StatusOK)
		if counter < 2 {
			_, _ = w.Write([]byte(`{"ipaddress": ["1.2.3.251", "1.2.3.252"], "ID": "100"}`))
		} else {
			_, _ = w.Write([]byte(`{"ID": "none"}`))
			counter = 0
		}
	}))
	defer testServer.Close()
	baningo.RootURL = testServer.URL + "/"

	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"bannedIP":  "1.2.3.251",
			"bannedIP2": "1.2.3.252",
			"IP":        "1.2.3.253",
			"IP2":       "1.2.3.254",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	config.CgrConfig().APIBanCfg().Keys = []string{"testKey"}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*apiban:~*req.IP:*all"}, dp); err != nil {
		t.Fatal(err)
	} else if pass {
		t.Error("Expected not to pass")
	}
	// from cache
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*apiban:~*req.IP:*all"}, dp); err != nil {
		t.Fatal(err)
	} else if pass {
		t.Error("Expected not to pass")
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*apiban:~*req.IP2:*single"}, dp); err != nil {
		t.Fatal(err)
	} else if pass {
		t.Error("Expected not to pass")
	}
	Cache.Clear([]string{utils.MetaAPIBan})
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*apiban:~*req.bannedIP:*single"}, dp); err != nil {
		t.Fatal(err)
	} else if !pass {
		t.Error("Expected to pass")
	}
	if pass, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*apiban:~*req.bannedIP2:*all"}, dp); err != nil {
		t.Fatal(err)
	} else if !pass {
		t.Error("Expected to pass")
	}

	if pass, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*apiban:~*req.notFound:*all"}, dp); err != nil {
		t.Fatal(err)
	} else if pass {
		t.Error("Expected not to pass")
	}
	expErr := "invalid value for apiban filter: <*any>"
	if _, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*apiban:~*req.IP:*any"}, dp); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}
	baningo.RootURL = "http://127.0.0.1:12345/"

	expErr = `Get "http://127.0.0.1:12345/testKey/banned/100": dial tcp 127.0.0.1:12345: connect: connection refused`
	if _, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*apiban:~*req.IP:*all"}, dp); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}
	expErr = `Get "http://127.0.0.1:12345/testKey/check/1.2.3.253": dial tcp 127.0.0.1:12345: connect: connection refused`
	if _, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*apiban:~*req.IP:*single"}, dp); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}

	expErr = `invalid converter value in string: <*>, err: unsupported converter definition: <*>`
	if _, err := filterS.Pass(context.TODO(), "cgrates.org", []string{"*apiban:~*req.<~*req.IP>{*}:*all"}, dp); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %s received: %v", expErr, err)
	}
}

func TestFilterPassRSRFieldsWithMultplieValues(t *testing.T) {
	ev := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"23": "sip:11561561561561568@dan",
		},
	}
	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	flts := NewFilterS(cfg, nil, dm)
	if passes, err := flts.Pass(context.Background(), "cgrate.org", []string{"*rsr:~*req.23:dan|1001"}, ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	if passes, err := flts.Pass(context.Background(), "cgrate.org", []string{"*rsr:~*req.23:dan"}, ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
}

func TestFilterPassCronExpOK(t *testing.T) {
	ev := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AnswerTime: "2021-05-05T12:00:01Z",
		},
	}

	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	fltr := NewFilterS(cfg, nil, dm)

	if passes, err := fltr.Pass(context.Background(), "cgrates.org",
		[]string{"*cronexp:~*req.AnswerTime:0 12 5 5 *"}, ev); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
}

func TestFilterPassCronExpNotActive(t *testing.T) {
	ev := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AnswerTime: "2021-05-05T12:00:01Z",
		},
	}

	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	fltr := NewFilterS(cfg, nil, dm)

	if passes, err := fltr.Pass(context.Background(), "cgrates.org",
		[]string{"*cronexp:~*req.AnswerTime:1 12 5 5 *"}, ev); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("should not be passing")
	}
}

func TestFilterPassCronExpParseErrWrongPath(t *testing.T) {
	ev := utils.MapStorage{
		utils.MetaReq: 13,
	}

	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	fltr := NewFilterS(cfg, nil, dm)
	experr := utils.ErrWrongPath

	if passes, err := fltr.Pass(context.Background(), "cgrates.org",
		[]string{"*cronexp:~*req.AnswerTime:1 12 5 5 *"}, ev); err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	} else if passes {
		t.Errorf("should not be passing")
	}
}

func TestFilterPassCronExpErrNotFound(t *testing.T) {
	ev := utils.MapStorage{}

	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	fltr := NewFilterS(cfg, nil, dm)

	if passes, err := fltr.Pass(context.Background(), "cgrates.org",
		[]string{"*cronexp:~*req.AnswerTime:1 12 5 5 *"}, ev); err != nil {
		t.Errorf("Expected nil, got %+v", err)
	} else if passes {
		t.Error("should not be passing")
	}
}

func TestFilterPassCronExpConvertTimeErr(t *testing.T) {
	ev := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AnswerTime: "invalid time format",
		},
	}

	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	fltr := NewFilterS(cfg, nil, dm)
	experr := "Unsupported time format"

	if passes, err := fltr.Pass(context.Background(), "cgrates.org",
		[]string{"*cronexp:~*req.AnswerTime:1 12 5 5 *"}, ev); err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	} else if passes {
		t.Error("should not be passing")
	}
}

func TestFilterPassCronExpParseExpErr(t *testing.T) {
	fltr, err := NewFilterRule(utils.MetaCronExp, "~*req.AnswerTime", []string{"* * * * * *"})
	if err != nil {
		t.Fatal(err)
	}

	ev := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AnswerTime: "2021-05-05T12:00:01Z",
		},
	}

	if passes, err := fltr.passCronExp(context.Background(), ev); err != nil {
		t.Errorf("Expected nil, got %+v", err)
	} else if passes {
		t.Error("should not be passing")
	}
}

func TestFiltersPassGreaterThanErrIncomparable(t *testing.T) {
	fltr, err := NewFilterRule(utils.MetaGreaterThan, "~*req.Usage", []string{"10"})
	fltr.rsrElement.Rules = "rules"
	if err != nil {
		t.Fatal(err)
	}
	ev := utils.MapStorage{
		utils.MetaReq: map[string]any{
			utils.Usage: nil,
		},
	}

	if passes, err := fltr.passCronExp(context.Background(), ev); err != nil {
		t.Errorf("Expected nil, got %+v", err)
	} else if passes {
		t.Error("should not be passing")
	}
}

func TestFilterPassCronExpParseDPErr(t *testing.T) {
	fltr, err := NewFilterRule(utils.MetaCronExp, "~*req.AnswerTime", []string{"~* * * * *"})
	if err != nil {
		t.Fatal(err)
	}

	ev := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AnswerTime: "2021-05-05T12:00:01Z",
		},
	}

	if passes, err := fltr.passCronExp(context.Background(), ev); err != nil {
		t.Errorf("Expected nil, got %+v", err)
	} else if passes {
		t.Error("should not be passing")
	}
}
func TestFilterSet(t *testing.T) {
	fltr := Filter{}
	exp := Filter{
		Tenant: "cgrates.org",
		ID:     "ID",
		Rules: []*FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1001", "1002"},
		}},
	}
	if err := fltr.Set([]string{}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := fltr.Set([]string{"NotAField"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := fltr.Set([]string{"NotAField", "1"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := fltr.Set([]string{utils.Tenant}, "cgrates.org", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := fltr.Set([]string{utils.ID}, "ID", false, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if err := fltr.Set([]string{utils.Rules, utils.Type}, utils.MetaString, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := fltr.Set([]string{utils.Rules, utils.Element}, "~*req.Account", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := fltr.Set([]string{utils.Rules, utils.Values}, "1001", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := fltr.Set([]string{utils.Rules, utils.Type}, utils.MetaString, true, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := fltr.Set([]string{utils.Rules, utils.Element}, "~*req.Account", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := fltr.Set([]string{utils.Rules, utils.Values}, "1002", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := fltr.Set([]string{utils.Rules, "Wrong"}, "1002", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}

	fltr.Compress()

	if !reflect.DeepEqual(exp, fltr) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(fltr))
	}
}

func TestFilterAsInterface(t *testing.T) {
	fltr := Filter{
		Tenant: "cgrates.org",
		ID:     "ID",
		Rules: []*FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1001", "1002"},
		}},
	}
	if _, err := fltr.FieldAsInterface(nil); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := fltr.FieldAsInterface([]string{"field"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := fltr.FieldAsInterface([]string{"field", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := fltr.FieldAsInterface([]string{utils.Rules + "[4]", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	expErrMsg := `strconv.Atoi: parsing "a": invalid syntax`
	if _, err := fltr.FieldAsInterface([]string{utils.Rules + "[a]", ""}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := fltr.FieldAsInterface([]string{utils.Rules + "[0]", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := fltr.FieldAsInterface([]string{utils.Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := fltr.FieldAsInterface([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := utils.ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := fltr.FieldAsInterface([]string{utils.Rules + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := fltr.Rules[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if val, err := fltr.FieldAsInterface([]string{utils.Rules + "[0]", utils.Values}); err != nil {
		t.Fatal(err)
	} else if exp := fltr.Rules[0].Values; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := fltr.FieldAsInterface([]string{utils.Rules + "[0]", utils.Values + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := fltr.Rules[0].Values[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if val, err := fltr.FieldAsInterface([]string{utils.Rules + "[0]", utils.Element}); err != nil {
		t.Fatal(err)
	} else if exp := fltr.Rules[0].Element; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if val, err := fltr.FieldAsInterface([]string{utils.Rules + "[0]", utils.Type}); err != nil {
		t.Fatal(err)
	} else if exp := fltr.Rules[0].Type; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := fltr.FieldAsString([]string{""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := fltr.FieldAsString([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := fltr.String(), utils.ToJSON(fltr); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := fltr.Rules[0].FieldAsString([]string{}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := fltr.Rules[0].FieldAsString([]string{utils.Type}); err != nil {
		t.Fatal(err)
	} else if exp := utils.MetaString; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := fltr.Rules[0].String(), utils.ToJSON(fltr.Rules[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

}

func TestFilterMerge(t *testing.T) {
	dp := &Filter{}
	exp := &Filter{
		Tenant: "cgrates.org",
		ID:     "ID",
		Rules: []*FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1001", "1002"},
		}},
	}
	if dp.Merge(&Filter{
		Tenant: "cgrates.org",
		ID:     "ID",
		Rules: []*FilterRule{{
			Type:    utils.MetaString,
			Element: "~*req.Account",
			Values:  []string{"1001", "1002"},
		}},
	}); !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dp))
	}
}

func TestFiltersFilterRuleIsValid(t *testing.T) {
	fltr := &FilterRule{
		Type:    utils.EmptyString,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNever,
		Element: "",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaExists,
		Element: "",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaExists,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaExists,
		Element: "~*req.Element",
		Values:  []string{"value1"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotExists,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaEmpty,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotEmpty,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaString,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotString,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaPrefix,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotPrefix,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaSuffix,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotSuffix,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaCronExp,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotCronExp,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaRSR,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotRSR,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaLessThan,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaLessOrEqual,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaGreaterThan,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaGreaterOrEqual,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaEqual,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotEqual,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaIPNet,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotIPNet,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaAPIBan,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotAPIBan,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaActivationInterval,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotActivationInterval,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaRegex,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotRegex,
		Element: "~*req.Element",
		Values:  []string{},
	}
	if isValid := fltr.IsValid(); isValid != false {
		t.Error("filter should not be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaString,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotString,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaPrefix,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotPrefix,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaSuffix,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotSuffix,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaCronExp,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotCronExp,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaRSR,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotRSR,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaLessThan,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaLessOrEqual,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaGreaterThan,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaGreaterOrEqual,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaEqual,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotEqual,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaIPNet,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotIPNet,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaAPIBan,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotAPIBan,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaActivationInterval,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotActivationInterval,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaRegex,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}

	fltr = &FilterRule{
		Type:    utils.MetaNotRegex,
		Element: "~*req.Element",
		Values:  []string{"value1", "value2"},
	}
	if isValid := fltr.IsValid(); isValid != true {
		t.Error("filter should be valid")
	}
}

func TestPassPartialErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmFilterPass := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := FilterS{
		cfg: cfg,
		dm:  dmFilterPass,
	}
	passEvent := map[string]any{
		"Account": "1007",
	}
	fEv := utils.MapStorage{}
	fEv.Set([]string{utils.MetaReq}, passEvent)
	prefixes := []string{utils.DynamicDataPrefix + utils.MetaReq}
	expErr := "NOT_FOUND:bad fltr"
	if _, _, err := filterS.LazyPass(context.Background(), "cgrates.org",
		[]string{"bad fltr"}, fEv, prefixes); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v> ", expErr, err.Error())
	}

}

func TestNewFilterRuleErrSupportedFltrType(t *testing.T) {
	expErr := "Unsupported filter Type: unsupported"
	if _, err := NewFilterRule("unsupported", "", []string{}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v> ", expErr, err.Error())

	}
}

func TestNewFilterRuleErrNoFldName(t *testing.T) {
	expErr := "Element is mandatory for Type: *cronexp"
	if _, err := NewFilterRule(utils.MetaCronExp, "", []string{}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v> ", expErr, err.Error())

	}
}
func TestNewFilterRuleErrNoVals(t *testing.T) {
	expErr := "Values is mandatory for Type: *cronexp"
	if _, err := NewFilterRule(utils.MetaCronExp, "~*req.AnswerTime", []string{}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v> ", expErr, err.Error())

	}
}

func TestPassNever(t *testing.T) {
	fltr := &FilterRule{}
	dDP := utils.MapStorage{}
	if ok, err := fltr.passNever(dDP); ok != false && err != nil {
		t.Errorf("Expected filter to pass never, unexpected error <%v>", err)
	}
}

func TestFilterRulePassRegexParseErrNotFound(t *testing.T) {

	rsrBadParse := config.NewRSRParserMustCompile("~*opts.<~*opts.*originID;~*req.RunID;-Cost>")

	fltr := &FilterRule{

		rsrElement: rsrBadParse,
	}

	dDP := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{},
		utils.MetaOpts: utils.MapStorage{
			utils.MetaOriginID: "originIDUniq",
		},
	}
	if ok, err := fltr.passRegex(dDP); ok != false && err != nil {
		t.Errorf("Expected error <%v>, Received error <%v>. Ok <%v>", nil, err, ok)
	}
}

func TestFilterRulePassRegexParseErr(t *testing.T) {

	rsrBadParse, err := config.NewRSRParser("~*opts.*originID<~*opts.Converter>")
	if err != nil {
		t.Fatal(err)
	}
	fltr := &FilterRule{
		Type:       utils.EmptyString,
		Element:    "~*req.Element",
		Values:     []string{"value1", "value2"},
		rsrElement: rsrBadParse,
	}

	dDP := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{},
		utils.MetaOpts: utils.MapStorage{
			"Converter":        "{*",
			utils.MetaOriginID: "originIDUniq",
		},
	}
	expErr := `invalid converter terminator in rule: <~*opts.*originID{*>`
	if ok, err := fltr.passRegex(dDP); ok != false || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>. Ok <%v>", expErr, err, ok)
	}
}
func TestCheckFilterErrValFuncElement(t *testing.T) {

	fltr := &Filter{
		Tenant: utils.CGRateSorg,
		ID:     "FLTR_CP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~missing path",
				Values:  []string{"ChargerProfile1"},
			},
		},
	}
	expErr := `Path is missing  for filter <{"Tenant":"cgrates.org","ID":"FLTR_CP_1","Rules":[{"Type":"*string","Element":"~missing path","Values":["ChargerProfile1"]}]}>`
	if err := CheckFilter(fltr); err.Error() != expErr {
		t.Error(err)
	}
}

func TestCheckFilterErrValFuncValues(t *testing.T) {

	fltr := &Filter{
		Tenant: utils.CGRateSorg,
		ID:     "FLTR_CP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Charger",
				Values:  []string{"~missing path"},
			},
		},
	}
	expErr := `Path is missing  for filter <{"Tenant":"cgrates.org","ID":"FLTR_CP_1","Rules":[{"Type":"*string","Element":"~*req.Charger","Values":["~missing path"]}]}>`
	if err := CheckFilter(fltr); err.Error() != expErr {
		t.Error(err)
	}
}

func TestPassActivationIntervalParseTimeElementErr(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaEqual, "~*req.ASR", []string{"40"})
	if err != nil {
		t.Error(err)
	}
	ev := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"ASR": 39,
		},
	}

	errExp := "Unsupported time format"
	if ok, err := rf.passActivationInterval(ev); ok {
		t.Errorf("Expected false, Received <%v>", ok)
	} else if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

}

func TestPassActivationIntervalParseTimeEndErr(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaEqual, "~*req.AnswerTime", []string{"40", "50"})
	if err != nil {
		t.Error(err)
	}
	ev := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AnswerTime: "2019-04-04T11:45:26.371Z",
		},
	}

	errExp := "Unsupported time format"
	if ok, err := rf.passActivationInterval(ev); ok {
		t.Errorf("Expected false, Received <%v>", ok)
	} else if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

}

func TestPassActivationIntervalParseTimeStartErr(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaEqual, "~*req.AnswerTime", []string{"40", "2020-04-04T11:45:26.371Z"})
	if err != nil {
		t.Error(err)
	}
	ev := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AnswerTime: "2019-04-04T11:45:26.371Z",
		},
	}

	errExp := "Unsupported time format"
	if ok, err := rf.passActivationInterval(ev); ok {
		t.Errorf("Expected false, Received <%v>", ok)
	} else if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

}

func TestPassActivationIntervalParseTimeStart2Err(t *testing.T) {
	rf, err := NewFilterRule(utils.MetaEqual, "~*req.AnswerTime", []string{"55"})
	if err != nil {
		t.Error(err)
	}
	ev := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AnswerTime: "2019-04-04T11:45:26.371Z",
		},
	}

	errExp := "Unsupported time format"
	if ok, err := rf.passActivationInterval(ev); ok {
		t.Errorf("Expected false, Received <%v>", ok)
	} else if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

}

func TestFilterRuleCompileValuesRSRParseErr(t *testing.T) {

	fltr := &FilterRule{
		Type:   utils.MetaRSR,
		Values: []string{"~(^_^", "^rule2$"},
	}

	if err := fltr.CompileValues(); err == nil || err.Error() != "error parsing regexp: missing closing ): `(^_^`" {
		t.Errorf("Expected <error parsing regexp: missing closing ): `(^_^`> ,received: <%+v>", err)
	}

}

func TestFilterRuleCompileValuesNeverParseErr(t *testing.T) {

	fltr, err := NewFilterRule(utils.MetaNever, utils.Accounts, []string{"val1"})
	if err != nil {
		t.Error(err)
	}

	cFiltr := fltr
	if err := fltr.CompileValues(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cFiltr, fltr) {
		t.Errorf("Expected \n<%v>, \n received \n<%v>", cFiltr, fltr)
	}

}

func TestFilterRuleCompileValuesActivationIntervalParseErr(t *testing.T) {

	fltr := &FilterRule{
		Type:   utils.MetaActivationInterval,
		Values: []string{"a{*"},
	}

	if err := fltr.CompileValues(); err == nil || err.Error() != "invalid converter terminator in rule: <a{*>" {
		t.Errorf("Expected <invalid converter terminator in rule: <a{*>> ,received: <%+v>", err)
	}

}

func TestFilterRuleCompileValuesNewRSRParsElementErr(t *testing.T) {

	fltr := &FilterRule{
		Type:    utils.MetaExists,
		Element: "a{*",
	}

	if err := fltr.CompileValues(); err == nil || err.Error() != "invalid converter terminator in rule: <a{*>" {
		t.Errorf("Expected <invalid converter terminator in rule: <a{*>> ,received: <%+v>", err)
	}

}

func TestFRPassMetaNever(t *testing.T) {

	fltr := &FilterRule{
		Type: utils.MetaNever,
	}

	if pass, err := fltr.Pass(context.Background(), utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if pass != false {
		t.Errorf("Expected to never pass")
	}

}

func TestFRPassStringParseDataProviderErr(t *testing.T) {
	rsrParse := &config.RSRParser{
		Rules: "~*opts.<~*opts.*originID;~*req.RunID;-Cost>",
	}
	if err := rsrParse.Compile(); err != nil {
		t.Error(err)
	}

	valParse := config.RSRParsers{
		&config.RSRParser{
			Rules: "~*opts.*originID;~*req.RunID;-Cost",
		},
	}
	if err := valParse.Compile(); err != nil {
		t.Error(err)
	}
	fltr := &FilterRule{
		Type:       utils.MetaString,
		Element:    "~*req.Charger",
		Values:     []string{"ChargerProfile1"},
		rsrElement: rsrParse,
		rsrValues:  valParse,
	}

	data := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{

			utils.RunID: utils.MetaDefault,
		},
		utils.MetaOpts: utils.MapStorage{
			utils.MetaOriginID:  "Uniq",
			"Uniq*default-Cost": 10,
		},
	}

	if pass, err := fltr.passString(data); err != nil {
		t.Error(err)
	} else if pass != false {
		t.Errorf("Expected to never pass")
	}

}
func TestHttpInlineFilter(t *testing.T) {

	dP := &utils.MapStorage{
		utils.MetaReq: map[string]any{
			"Attribute":        "AttributeProfile1",
			"CGRID":            "CGRATES_ID1",
			utils.AccountField: "1002",
			utils.AnswerTime:   time.Date(2013, 12, 30, 14, 59, 31, 0, time.UTC),
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if len(r.URL.Query()) != 0 {
			queryVal := r.URL.Query()
			reply := queryVal.Has("~*req.Account") && queryVal.Get("~*req.Account") == "1002"
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, reply)
			return
		}

		var data map[string]any
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, has := data["*req"]
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, has)

	}))
	defer srv.Close()
	url := "*http#" + "[" + srv.URL + "]"

	exp := &Filter{
		Tenant: "cgrates.org",
		ID:     url + ":" + "~*req.Account:",
		Rules: []*FilterRule{
			{
				Type:    url,
				Element: "~*req.Account",
			},
		},
	}
	if err := exp.Compile(); err != nil {
		t.Fatal(err)
	}
	if fl, err := NewFilterFromInline("cgrates.org", url+":"+"~*req.Account:"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, fl) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(exp), utils.ToJSON(fl))
	}
	if pass, err := exp.Rules[0].Pass(context.Background(), dP); err != nil {
		t.Error(err)
	} else if !pass {
		t.Error("should had passed")
	}

	exp2 := &Filter{
		Tenant: "cgrates.org",
		ID:     url + ":" + "*any:",
		Rules: []*FilterRule{
			{
				Type:    url,
				Element: "*any",
			},
		},
	}

	if fl2, err := NewFilterFromInline("cgrates.org", url+":"+"*any:"); err != nil {
		t.Error(err)
	} else if fl2.Rules[0].Type != exp2.Rules[0].Type {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(exp2), utils.ToJSON(fl2))
	} else if pass, err := fl2.Rules[0].Pass(context.Background(), dP); err != nil {
		t.Error(err)
	} else if !pass {
		t.Error("should had passed")
	}
}

func TestFilterTrends(t *testing.T) {
	tmpConn := connMgr
	defer func() {
		connMgr = tmpConn
	}()
	clientConn := make(chan context.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.TrendSv1GetTrendSummary: func(ctx *context.Context, args, reply any) error {
				argGetTrend, ok := args.(*utils.TenantIDWithAPIOpts)
				if !ok {
					return fmt.Errorf("wrong args")
				}
				if argGetTrend.ID == "Trend1" && argGetTrend.Tenant == "cgrates.org" {
					now := time.Now()
					now2 := now.Add(time.Second)
					tr := Trend{
						Tenant:   "cgrates.org",
						ID:       "Trend1",
						RunTimes: []time.Time{now, now2},
						Metrics: map[time.Time]map[string]*MetricWithTrend{
							now: {
								"*acc": {ID: "*acc", Value: 45, TrendGrowth: -1.0, TrendLabel: utils.NotAvailable},
								"*acd": {ID: "*acd", Value: 50, TrendGrowth: -1.0, TrendLabel: utils.NotAvailable},
							},
							now2: {
								"*acc": {ID: "*acc", Value: 42, TrendGrowth: -3.0, TrendLabel: utils.MetaNegative},
								"*acd": {ID: "*acd", Value: 52, TrendGrowth: 2.0, TrendLabel: utils.MetaPositive},
							},
						},
					}
					trS := tr.asTrendSummary()
					*reply.(*TrendSummary) = *trS
					return nil
				}
				return utils.ErrNotFound
			},
		},
	}
	now3 := time.Now().Add(-time.Second * 3).Format(time.RFC3339)
	connMgr = NewConnManager(config.NewDefaultCGRConfig())

	connMgr.rpcInternal = map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTrends): clientConn,
	}
	testCases := []struct {
		name       string
		filter     string
		shouldPass bool
	}{
		{
			name:       "TrendPassACCLabel",
			filter:     "*string:~*trends.Trend1.Metrics.*acc.TrendLabel:*negative",
			shouldPass: true,
		},
		{

			name:       "TrendPassID",
			filter:     "*string:~*trends.Trend1.ID:Trend1",
			shouldPass: true,
		},
		{
			name:       "TrendPassACDLabel",
			filter:     "*string:~*trends.Trend1.Metrics.*acd.TrendLabel:*positive",
			shouldPass: true,
		},
		{
			name:       "TrendFailACCGrowth",
			filter:     "*gt:~*trends.Trend1.Metrics.*acc.TrendGrowth:1.0",
			shouldPass: false,
		},
		{
			name:       "TrendPassACDGrowth",
			filter:     "*gte:~*trends.Trend1.Metrics.*acd.TrendGrowth:2",
			shouldPass: true,
		},
		{
			name:       "TrendPassTime",
			filter:     fmt.Sprintf("*ai:~*trends.Trend1.Time:%s", now3),
			shouldPass: true,
		},
	}
	initDP := utils.MapStorage{}
	dp := newDynamicDP(context.Background(), nil, nil, nil, []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTrends)}, nil, "cgrates.org", initDP)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fl, err := NewFilterFromInline("cgrates.org", tc.filter)
			if err != nil {
				t.Fatal(err)
			}
			rule := fl.Rules[0]
			pass, err := rule.Pass(context.Background(), dp)
			if err != nil {
				t.Fatalf("rule.Pass() unexpected error: %v", err)
			}
			if pass != tc.shouldPass {
				t.Errorf("rule.Pass() expected: %t ,got: %t", tc.shouldPass, pass)
			}
		})
	}
}

func TestFilterRanking(t *testing.T) {
	tmpConn := connMgr
	defer func() {
		connMgr = tmpConn
	}()
	clientConn := make(chan context.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.RankingSv1GetRankingSummary: func(ctx *context.Context, args any, reply any) error {
				argTntID, ok := args.(*utils.TenantIDWithAPIOpts)
				if !ok {
					return fmt.Errorf("wrong args")
				}
				if argTntID.ID == "Ranking1" && argTntID.Tenant == "cgrates.org" {
					rn := Ranking{
						Tenant:        "cgrates.org",
						ID:            "Ranking1",
						LastUpdate:    time.Now(),
						SortedStatIDs: []string{"Stat5", "Stat6", "Stat7", "Stat4", "Stat3", "Stat1", "Stat2"},
					}
					rnS := rn.asRankingSummary()
					*reply.(*RankingSummary) = *rnS
					return nil
				}
				return utils.ErrNotFound
			},
		}}

	connMgr = NewConnManager(config.NewDefaultCGRConfig())
	connMgr.rpcInternal = map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRankings): clientConn,
	}
	now := time.Now().Add(-2 * time.Second).Format(time.RFC3339)
	testCases := []struct {
		name   string
		filter string
		pass   bool
	}{
		{name: "RankingPassID", filter: "*string:~*rankings.Ranking1.ID:Ranking1", pass: true},
		{name: "RankingPassLastUpdate", filter: fmt.Sprintf("*ai:~*rankings.Ranking1.LastUpdate:%s", now), pass: true},
		{name: "RankingPassFirstSortedStatIDs", filter: "*string:~*rankings.Ranking1.SortedStatIDs[0]:Stat5", pass: true},
		{name: "RankingPassLastSortedStatIDs", filter: "*string:~*rankings.Ranking1.SortedStatIDs[6]:Stat2", pass: true},
		{name: "RankingFailSortedStatIDsIdx", filter: "*string:~*rankings.Ranking1.SortedStatIDs[1]:Stat4", pass: false},
	}

	initDP := utils.MapStorage{}
	dp := newDynamicDP(context.Background(), nil, nil, nil, nil, []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRankings)}, "cgrates.org", initDP)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fl, err := NewFilterFromInline(dp.tenant, tc.filter)
			if err != nil {
				t.Fatal(err)
			}
			rule := fl.Rules[0]
			pass, err := rule.Pass(context.Background(), dp)
			if err != nil {
				t.Fatalf("rule.Pass() unexpected error: %v", err)
			}
			if pass != tc.pass {
				t.Errorf("rule.Pass() expected: %t ,got: %t", tc.pass, pass)
			}
		})
	}
}
