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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/utils"
)

var (
	testRoutesPrfs = []*RouteProfile{
		&RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile1",
			FilterIDs: []string{"FLTR_RPP_1"},
			Sorting:   utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
		&RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile2",
			FilterIDs: []string{"FLTR_SUPP_2"},
			Sorting:   utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route2",
					Weights:         utils.DynamicWeights{{Weight: 20}},
					RouteParameters: "param2",
				},
				{
					ID:              "route3",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param3",
				},
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 30}},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 20}},
		},
		&RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "RouteProfilePrefix",
			FilterIDs: []string{"FLTR_SUPP_3"},
			Sorting:   utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
	}
	testRoutesArgs = []*utils.CGREvent{
		{ //matching RouteProfile1
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route":          "RouteProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]interface{}{
				utils.OptsRoutesProfileCount: 1,
			},
		},
		{ //matching RouteProfile2
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route":          "RouteProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]interface{}{
				utils.OptsRoutesProfileCount: 1,
			},
		},
		{ //matching RouteProfilePrefix
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route": "RouteProfilePrefix",
			},
			APIOpts: map[string]interface{}{
				utils.OptsRoutesProfileCount: 1,
			},
		},
		{ //matching
			Tenant: "cgrates.org",
			ID:     "CGR",
			Event: map[string]interface{}{
				"UsageInterval": "1s",
				"PddInterval":   "1s",
			},
			APIOpts: map[string]interface{}{
				utils.OptsRoutesProfileCount: 1,
			},
		},
	}
)

func prepareRoutesData(t *testing.T, dm *DataManager) {
	if err := dm.SetFilter(context.Background(), &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetFilter(context.Background(), &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.PddInterval",
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetFilter(context.Background(), &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Route",
				Values:  []string{"RouteProfilePrefix"},
			},
		},
	}, true); err != nil {
		t.Fatal(err)
	}
	for _, spp := range testRoutesPrfs {
		if err = dm.SetRouteProfile(context.Background(), spp, true); err != nil {
			t.Fatal(err)
		}
	}
	//Test each route profile from cache
	for _, spp := range testRoutesPrfs {
		if tempSpp, err := dm.GetRouteProfile(context.Background(), spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(spp, tempSpp) {
			t.Fatalf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}
}

func TestRoutesCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	prepareRoutesData(t, dmSPP)
}

func TestRoutesmatchingRouteProfilesForEvent(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)

	prepareRoutesData(t, dmSPP)

	for i, spp := range testRoutesPrfs {
		sprf, err := routeService.matchingRouteProfilesForEvent(context.Background(), testRoutesArgs[0].Tenant, testRoutesArgs[i])
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(spp, sprf[0].RouteProfile) {
			t.Errorf("Expecting: %+v, received: %+v", spp, sprf[0].RouteProfile)
		}
	}
}

func TestRoutesSortedForEvent(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	prepareRoutesData(t, dmSPP)

	eFirstRouteProfile := SortedRoutesList{&SortedRoutes{
		ProfileID: "RouteProfile1",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}}
	sprf, err := routeService.sortedRoutesForEvent(context.Background(), testRoutesArgs[0].Tenant, testRoutesArgs[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}

	eFirstRouteProfile = SortedRoutesList{&SortedRoutes{
		ProfileID: "RouteProfile2",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(30.0),
				},
				SortingData: map[string]interface{}{
					utils.Weight: 30.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}}

	sprf, err = routeService.sortedRoutesForEvent(context.Background(), testRoutesArgs[1].Tenant, testRoutesArgs[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}

	eFirstRouteProfile = SortedRoutesList{&SortedRoutes{
		ProfileID: "RouteProfilePrefix",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}}

	sprf, err = routeService.sortedRoutesForEvent(context.Background(), testRoutesArgs[2].Tenant, testRoutesArgs[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}
}

func TestRoutesSortedForEventWithLimit(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	prepareRoutesData(t, dmSPP)

	eFirstRouteProfile := SortedRoutesList{&SortedRoutes{
		ProfileID: "RouteProfile2",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(30.0),
				},
				SortingData: map[string]interface{}{
					utils.Weight: 30.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
		},
	}}
	args := testRoutesArgs[1].Clone()
	args.APIOpts[utils.OptsRoutesLimit] = 2
	sprf, err := routeService.sortedRoutesForEvent(context.Background(), args.Tenant, args)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}
}

func TestRoutesSortedForEventWithOffset(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)

	prepareRoutesData(t, dmSPP)

	eFirstRouteProfile := SortedRoutesList{&SortedRoutes{
		ProfileID: "RouteProfile2",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}}
	args := testRoutesArgs[1].Clone()
	args.APIOpts[utils.OptsRoutesOffset] = 2
	sprf, err := routeService.sortedRoutesForEvent(context.Background(), args.Tenant, args)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}
}

func TestRoutesSortedForEventWithLimitAndOffset(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	prepareRoutesData(t, dmSPP)

	eFirstRouteProfile := SortedRoutesList{&SortedRoutes{
		ProfileID: "RouteProfile2",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
		},
	}}
	args := testRoutesArgs[1].Clone()
	args.APIOpts[utils.OptsRoutesLimit] = 1
	args.APIOpts[utils.OptsRoutesOffset] = 1
	sprf, err := routeService.sortedRoutesForEvent(context.Background(), args.Tenant, args)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}
}

func TestRoutesNewOptsGetRoutes(t *testing.T) {
	ev := &utils.CGREvent{
		APIOpts: map[string]interface{}{
			utils.OptsRoutesMaxCost:      10,
			utils.OptsRoutesIgnoreErrors: true,
		},
	}
	spl := &optsGetRoutes{
		ignoreErrors: true,
		maxCost:      10.0,
		paginator:    &utils.Paginator{},
	}
	sprf, err := newOptsGetRoutes(context.Background(), ev, &FilterS{}, config.CgrConfig().RouteSCfg().Opts)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestRoutesNewOptsGetRoutesFromCfg(t *testing.T) {
	config.CgrConfig().RouteSCfg().Opts.IgnoreErrors = []*utils.DynamicBoolOpt{{Value: true}}
	ev := &utils.CGREvent{
		APIOpts: map[string]interface{}{},
	}
	spl := &optsGetRoutes{
		ignoreErrors: true,
		paginator:    &utils.Paginator{},
	}
	sprf, err := newOptsGetRoutes(context.Background(), ev, &FilterS{}, config.CgrConfig().RouteSCfg().Opts)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestRoutesNewOptsGetRoutesIgnoreErrors(t *testing.T) {
	ev := &utils.CGREvent{
		APIOpts: map[string]interface{}{
			utils.OptsRoutesIgnoreErrors: true,
		},
	}
	spl := &optsGetRoutes{
		ignoreErrors: true,
		paginator:    &utils.Paginator{},
	}
	sprf, err := newOptsGetRoutes(context.Background(), ev, &FilterS{}, config.CgrConfig().RouteSCfg().Opts)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestRoutesMatchWithIndexFalse(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	cfg.RouteSCfg().IndexedSelects = false
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	prepareRoutesData(t, dmSPP)

	for i, spp := range testRoutesPrfs {
		sprf, err := routeService.matchingRouteProfilesForEvent(context.Background(), testRoutesArgs[0].Tenant, testRoutesArgs[i])
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(spp, sprf[0].RouteProfile) {
			t.Errorf("Expecting: %+v, received: %+v", spp, sprf[0].RouteProfile)
		}
	}
}

func TestRoutesSortedForEventWithLimitAndOffset2(t *testing.T) {
	Cache.Clear(nil)
	sppTest := []*RouteProfile{
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfile1",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route2",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfile2",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route2",
					Weights:         utils.DynamicWeights{{Weight: 20}},
					RouteParameters: "param2",
				},
				{
					ID:              "route3",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param3",
				},
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 30}},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 5}},
		},
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfilePrefix",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 20}},
		},
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfilePrefix4",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{}},
		},
	}
	args := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "utils.CGREvent1",
		Event:   map[string]interface{}{},
		APIOpts: map[string]interface{}{utils.OptsRoutesProfileCount: 3},
	}

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)

	for _, spp := range sppTest {
		if err = dmSPP.SetRouteProfile(context.Background(), spp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
		if tempSpp, err := dmSPP.GetRouteProfile(context.Background(), spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}

	eFirstRouteProfile := SortedRoutesList{
		{
			ProfileID: "RouteProfile1",
			Sorting:   utils.MetaWeight,
			Routes: []*SortedRoute{
				{
					RouteID: "route2",
					sortingDataDecimal: map[string]*utils.Decimal{
						utils.Weight: utils.NewDecimalFromFloat64(10.),
					},
					SortingData: map[string]interface{}{
						utils.Weight: 10.,
					},
					RouteParameters: "param1",
				},
			},
		},
		{
			ProfileID: "RouteProfile2",
			Sorting:   utils.MetaWeight,
			Routes: []*SortedRoute{
				{
					RouteID: "route1",
					sortingDataDecimal: map[string]*utils.Decimal{
						utils.Weight: utils.NewDecimalFromFloat64(30.),
					},
					SortingData: map[string]interface{}{
						utils.Weight: 30.,
					},
					RouteParameters: "param1",
				},
			},
		},
	}
	args.APIOpts[utils.OptsRoutesLimit] = 2
	args.APIOpts[utils.OptsRoutesOffset] = 1
	sprf, err := routeService.sortedRoutesForEvent(context.Background(), args.Tenant, args)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}
}

func TestRoutesV1GetRoutesMsnStructFieldIDError(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	var reply SortedRoutesList
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event:  map[string]interface{}{},
	}
	err := routeService.V1GetRoutes(context.Background(), args, &reply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
}

func TestRoutesV1GetRoutesMsnStructFieldEventError(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	var reply SortedRoutesList
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "CGREvent1",
	}
	err := routeService.V1GetRoutes(context.Background(), args, &reply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [Event]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [Event]", err)
	}
}

func TestRoutesV1GetRoutesNotFoundError(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	var reply SortedRoutesList
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "CGREvent1",
		Event:  map[string]interface{}{},
	}
	err := routeService.V1GetRoutes(context.Background(), args, &reply)
	if err == nil || err.Error() != utils.NotFoundCaps {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.NotFoundCaps, err)
	}
}

func TestRoutesV1GetRoutesNoTenantNotFoundError(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	var reply SortedRoutesList
	args := &utils.CGREvent{
		ID:    "CGREvent1",
		Event: map[string]interface{}{},
	}
	err := routeService.V1GetRoutes(context.Background(), args, &reply)
	if err == nil || err.Error() != utils.NotFoundCaps {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.NotFoundCaps, err)
	}
}

func TestRoutesV1GetRoutesAttrConnError(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()["testConn"] = config.NewDfltRPCConn()
	cfg.RouteSCfg().AttributeSConns = []string{"testConn"}
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := NewConnManager(cfg)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), connMng)
	routeService := NewRouteService(dmSPP, nil, cfg, connMng)
	var reply SortedRoutesList
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "CGREvent1",
		Event:  map[string]interface{}{},
	}
	err := routeService.V1GetRoutes(context.Background(), args, &reply)
	if err == nil || err.Error() != "ROUTES_ERROR:%!s(<nil>)" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "ROUTES_ERROR:%!s(<nil>)", err)
	}
}

func TestRoutesV1GetRouteProfilesForEventError(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := NewConnManager(cfg)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), connMng)
	fltr := &FilterS{dm: dmSPP, cfg: cfg}
	routeService := NewRouteService(dmSPP, fltr, cfg, connMng)
	var reply []*RouteProfile
	args := &utils.CGREvent{
		ID:    "CGREvent1",
		Event: map[string]interface{}{},
	}
	err := routeService.V1GetRouteProfilesForEvent(context.Background(), args, &reply)
	if err == nil || err.Error() != utils.NotFoundCaps {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.NotFoundCaps, err)
	}
}

func TestRoutesV1GetRouteProfilesForEventMsnIDError(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := NewConnManager(cfg)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), connMng)
	fltr := &FilterS{dm: dmSPP, cfg: cfg}
	routeService := NewRouteService(dmSPP, fltr, cfg, connMng)
	var reply []*RouteProfile
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event:  map[string]interface{}{},
	}
	err := routeService.V1GetRouteProfilesForEvent(context.Background(), args, &reply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
}

func TestRoutesV1GetRouteProfilesForEventMsnEventError(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := NewConnManager(cfg)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), connMng)
	fltr := &FilterS{dm: dmSPP, cfg: cfg}
	routeService := NewRouteService(dmSPP, fltr, cfg, connMng)
	var reply []*RouteProfile
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "CGREvent1",
	}
	err := routeService.V1GetRouteProfilesForEvent(context.Background(), args, &reply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [Event]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [Event]", err)
	}
}

func TestRouteProfileSet(t *testing.T) {
	rp := RouteProfile{}
	exp := RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:              "RT1",
			FilterIDs:       []string{"fltr1"},
			AccountIDs:      []string{"acc1"},
			RateProfileIDs:  []string{"rp1"},
			ResourceIDs:     []string{"res1"},
			StatIDs:         []string{"stat1"},
			Weights:         utils.DynamicWeights{{}},
			Blocker:         true,
			RouteParameters: "params",
		}},
	}
	if err := rp.Set([]string{}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{"NotAField"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{"NotAField", "1"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := rp.Set([]string{utils.Tenant}, "cgrates.org", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.ID}, "ID", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.FilterIDs}, "fltr1;*string:~*req.Account:1001", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Weights}, "", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Sorting}, utils.MetaQOS, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.SortingParameters}, "param", false, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if err := rp.Set([]string{utils.Routes, utils.ID}, "RT1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Routes, utils.FilterIDs}, "fltr1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Routes, utils.AccountIDs}, "acc1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Routes, utils.RateProfileIDs}, "rp1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Routes, utils.ResourceIDs}, "res1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Routes, utils.StatIDs}, "stat1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Routes, utils.Weights}, "", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Routes, utils.Blocker}, "true", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Routes, utils.RouteParameters}, "params", false, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if err := rp.Set([]string{utils.SortingParameters, "wrong"}, "param", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Routes, "wrong"}, "param", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, rp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rp))
	}
}
