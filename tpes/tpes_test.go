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

package tpes

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewTPeS(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpExporterTypes.Add("not_valid")
	// utils.Logger, err = utils.NewLogger(utils.MetaStdLog, utils.EmptyString, 6)
	// if err != nil {
	// 	t.Error(err)
	// }
	// // utils.Logger.SetLogLevel(7)
	// buff := new(bytes.Buffer)
	// log.SetOutput(buff)
	_ = NewTPeS(cfg, dm, connMng)
	tpExporterTypes.Remove("not_valid")
	// expected := "<not_valid>"
	// if rcv := buff.String(); !strings.Contains(rcv, expected) {
	// 	t.Errorf("Expected %v, received %v", expected, rcv)
	// }
}

func TestGetTariffPlansKeys(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)

	//Attributes
	attr := &utils.AttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_TEST",
		FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.AccountField,
				Type:  utils.MetaConstant,
				Value: nil,
			},
			{
				Path:  "*tenant",
				Type:  utils.MetaConstant,
				Value: nil,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	dm.SetAttributeProfile(context.Background(), attr, false)
	rcv, _ := getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaAttributes)
	exp := []string{"TEST_ATTRIBUTES_TEST"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Actions
	act := &utils.ActionProfile{
		Tenant: "cgrates.org",
		ID:     "SET_BAL",
		FilterIDs: []string{
			"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Schedule: utils.MetaASAP,
		Actions: []*utils.APAction{
			{
				ID:   "SET_BAL",
				Type: utils.MetaSetBalance,
				Diktats: []*utils.APDiktat{
					{
						Path:  "MONETARY",
						Value: "10",
					}},
			},
		},
	}
	dm.SetActionProfile(context.Background(), act, false)
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaActions)
	exp = []string{"SET_BAL"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Accounts
	acc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "Account_simple",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						Weight: 12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": "10",
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	dm.SetAccount(context.Background(), acc, false)
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaAccounts)
	exp = []string{"Account_simple"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Chargers
	chgr := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Chargers1",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	dm.SetChargerProfile(context.Background(), chgr, false)
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaChargers)
	exp = []string{"Chargers1"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Filters
	fltr := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_prf",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1004", "6774", "22312"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*opts.Subsystems",
				Values:  []string{"*attributes"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destinations",
				Values:  []string{"+0775", "+442"},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.NumberOfEvents",
			},
		},
	}
	dm.SetFilter(context.Background(), fltr, false)
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaFilters)
	exp = []string{"fltr_for_prf"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Rates
	rt := &utils.RateProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_RATE_TEST",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rt, false, true); err != nil {
		t.Error(err)
	}
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaRates)
	exp = []string{"TEST_RATE_TEST"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	// Resources
	rsc := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResGroup1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		Limit:             10,
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			}},
		ThresholdIDs: []string{utils.MetaNone},
	}
	dm.SetResourceProfile(context.Background(), rsc, false)
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaResources)
	exp = []string{"ResGroup1"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Routes
	rte := &utils.RouteProfile{
		ID:     "ROUTE_2003",
		Tenant: "cgrates.org",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*utils.Route{
			{
				ID: "route1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			},
		},
	}
	dm.SetRouteProfile(context.Background(), rte, false)
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaRoutes)
	exp = []string{"ROUTE_2003"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Stats
	stq := &engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ_2",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		QueueLength: 14,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaASR,
			},
			{
				MetricID: utils.MetaTCD,
			},
			{
				MetricID: utils.MetaPDD,
			},
			{
				MetricID: utils.MetaTCC,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	dm.SetStatQueueProfile(context.Background(), stq, false)
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaStats)
	exp = []string{"SQ_2"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Thresholds
	thd := &engine.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_2",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}
	dm.SetThresholdProfile(context.Background(), thd, false)
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaThresholds)
	exp = []string{"THD_2"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Unsupported
	_, err := getTariffPlansKeys(context.Background(), dm, "cgrates.org", "not_valid")
	errExpect := "Unsuported exporter type"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err.Error())
	}
}

func TestV1ExportTariffPlan(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	attr := &utils.AttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_TEST",
		FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.AccountField,
				Type:  utils.MetaConstant,
				Value: nil,
			},
			{
				Path:  "*tenant",
				Type:  utils.MetaConstant,
				Value: nil,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	dm.SetAttributeProfile(context.Background(), attr, false)
	tpE := NewTPeS(cfg, dm, connMng)
	var reply []byte
	args := &ArgsExportTP{
		Tenant: utils.EmptyString,
		ExportItems: map[string][]string{
			utils.MetaAttributes: {"TEST_ATTRIBUTES_TEST"},
		},
	}
	err := tpE.V1ExportTariffPlan(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
}

func TestV1ExportTariffPlanZeroExp(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpE := NewTPeS(cfg, dm, connMng)
	var reply []byte
	args := &ArgsExportTP{
		Tenant:      utils.EmptyString,
		ExportItems: map[string][]string{},
	}
	err := tpE.V1ExportTariffPlan(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
}

func TestV1ExportTariffPlanZeroIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpE := NewTPeS(cfg, dm, connMng)
	var reply []byte
	args := &ArgsExportTP{
		Tenant: utils.EmptyString,
		ExportItems: map[string][]string{
			utils.MetaAttributes: {},
		},
	}
	err := tpE.V1ExportTariffPlan(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
}

func TestV1ExportTariffPlanInvalidExpType(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpE := NewTPeS(cfg, dm, connMng)
	var reply []byte
	args := &ArgsExportTP{
		Tenant: utils.EmptyString,
		ExportItems: map[string][]string{
			"not_valid": {},
		},
	}
	err := tpE.V1ExportTariffPlan(context.Background(), args, &reply)
	errExp := "UNSUPPORTED_TPEXPORTER_TYPE:not_valid"
	if err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err.Error())
	}

}
