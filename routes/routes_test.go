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
package routes

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/utils"
)

var (
	testRoutesPrfs = []*utils.RouteProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile1",
			FilterIDs: []string{"FLTR_RPP_1"},
			Sorting:   utils.MetaWeight,
			Routes: []*utils.Route{
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights:  utils.DynamicWeights{{Weight: 10}},
			Blockers: utils.DynamicBlockers{{Blocker: true}},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile2",
			FilterIDs: []string{"FLTR_SUPP_2"},
			Sorting:   utils.MetaWeight,
			Routes: []*utils.Route{
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
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfilePrefix",
			FilterIDs: []string{"FLTR_SUPP_3"},
			Sorting:   utils.MetaWeight,
			Routes: []*utils.Route{
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
			Event: map[string]any{
				"Route":          "RouteProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]any{
				utils.OptsRoutesProfilesCount: 1,
			},
		},
		{ //matching RouteProfile2
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]any{
				"Route":          "RouteProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]any{
				utils.OptsRoutesProfilesCount: 1,
			},
		},
		{ //matching RouteProfilePrefix
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]any{
				"Route": "RouteProfilePrefix",
			},
			APIOpts: map[string]any{
				utils.OptsRoutesProfilesCount: 1,
			},
		},
		{ //matching
			Tenant: "cgrates.org",
			ID:     "CGR",
			Event: map[string]any{
				"UsageInterval": "1s",
				"PddInterval":   "1s",
			},
			APIOpts: map[string]any{
				utils.OptsRoutesProfilesCount: 1,
			},
		},
	}
)

func prepareRoutesData(t *testing.T, dm *engine.DataManager) {
	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
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
	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_2",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.PddInterval",
				Values:  []string{time.Second.String()},
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
	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*engine.FilterRule{
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
		if err := dm.SetRouteProfile(context.Background(), spp, true); err != nil {
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	prepareRoutesData(t, dmSPP)
}

func TestRoutesmatchingRouteProfilesForEvent(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltrs, cfg, nil)

	prepareRoutesData(t, dmSPP)

	for i, spp := range testRoutesPrfs {
		sprf, err := routeService.matchingRouteProfilesForEvent(context.Background(), testRoutesArgs[0].Tenant, testRoutesArgs[i])
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(spp, sprf[0]) {
			t.Errorf("Expecting: %+v, received: %+v", spp, sprf[0])
		}
	}
}

func TestRoutesSortedForEvent(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltrs, cfg, nil)
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
				SortingData: map[string]any{
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
				SortingData: map[string]any{
					utils.Weight: 30.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]any{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
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
				SortingData: map[string]any{
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltrs, cfg, nil)
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
				SortingData: map[string]any{
					utils.Weight: 30.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*utils.Decimal{
					utils.Weight: utils.NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]any{
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltrs, cfg, nil)

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
				SortingData: map[string]any{
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltrs, cfg, nil)
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
				SortingData: map[string]any{
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
		APIOpts: map[string]any{
			utils.OptsRoutesMaxCost:      10,
			utils.OptsRoutesIgnoreErrors: true,
		},
	}
	spl := &optsGetRoutes{
		ignoreErrors: true,
		maxCost:      10.0,
		paginator:    &utils.Paginator{},
	}
	sprf, err := newOptsGetRoutes(context.Background(), ev, &engine.FilterS{}, config.CgrConfig().RouteSCfg().Opts)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestRoutesNewOptsGetRoutesFromCfg(t *testing.T) {
	config.CgrConfig().RouteSCfg().Opts.IgnoreErrors = []*config.DynamicBoolOpt{config.NewDynamicBoolOpt(nil, "", true, nil)}
	ev := &utils.CGREvent{
		APIOpts: map[string]any{},
	}
	spl := &optsGetRoutes{
		ignoreErrors: true,
		paginator:    &utils.Paginator{},
	}
	sprf, err := newOptsGetRoutes(context.Background(), ev, &engine.FilterS{}, config.CgrConfig().RouteSCfg().Opts)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestRoutesNewOptsGetRoutesIgnoreErrors(t *testing.T) {
	ev := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.OptsRoutesIgnoreErrors: true,
		},
	}
	spl := &optsGetRoutes{
		ignoreErrors: true,
		paginator:    &utils.Paginator{},
	}
	sprf, err := newOptsGetRoutes(context.Background(), ev, &engine.FilterS{}, config.CgrConfig().RouteSCfg().Opts)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestRoutesMatchWithIndexFalse(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	cfg.RouteSCfg().IndexedSelects = false
	fltrs := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltrs, cfg, nil)
	prepareRoutesData(t, dmSPP)

	for i, spp := range testRoutesPrfs {
		sprf, err := routeService.matchingRouteProfilesForEvent(context.Background(), testRoutesArgs[0].Tenant, testRoutesArgs[i])
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(spp, sprf[0]) {
			t.Errorf("Expecting: %+v, received: %+v", spp, sprf[0])
		}
	}
}

func TestRoutesSortedForEventWithLimitAndOffset2(t *testing.T) {
	engine.Cache.Clear(nil)
	sppTest := []*utils.RouteProfile{
		{
			Tenant:  "cgrates.org",
			ID:      "RouteProfile1",
			Sorting: utils.MetaWeight,
			Routes: []*utils.Route{
				{
					ID:              "route2",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
		{
			Tenant:  "cgrates.org",
			ID:      "RouteProfile2",
			Sorting: utils.MetaWeight,
			Routes: []*utils.Route{
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
		{
			Tenant:  "cgrates.org",
			ID:      "RouteProfilePrefix",
			Sorting: utils.MetaWeight,
			Routes: []*utils.Route{
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 20}},
		},
		{
			Tenant:  "cgrates.org",
			ID:      "RouteProfilePrefix4",
			Sorting: utils.MetaWeight,
			Routes: []*utils.Route{
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
		Event:   map[string]any{},
		APIOpts: map[string]any{utils.OptsRoutesProfilesCount: 3},
	}

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltrs, cfg, nil)

	for _, spp := range sppTest {
		if err := dmSPP.SetRouteProfile(context.Background(), spp, true); err != nil {
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
					SortingData: map[string]any{
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
					SortingData: map[string]any{
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
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltrs, cfg, nil)
	var reply SortedRoutesList
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event:  map[string]any{},
	}
	err := routeService.V1GetRoutes(context.Background(), args, &reply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
}

func TestRoutesV1GetRoutesMsnStructFieldEventError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltrs, cfg, nil)
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
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltrs, cfg, nil)
	var reply SortedRoutesList
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "CGREvent1",
		Event:  map[string]any{},
	}
	err := routeService.V1GetRoutes(context.Background(), args, &reply)
	if err == nil || err.Error() != utils.NotFoundCaps {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.NotFoundCaps, err)
	}
}

func TestRoutesV1GetRoutesNoTenantNotFoundError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltrs, cfg, nil)
	var reply SortedRoutesList
	args := &utils.CGREvent{
		ID:    "CGREvent1",
		Event: map[string]any{},
	}
	err := routeService.V1GetRoutes(context.Background(), args, &reply)
	if err == nil || err.Error() != utils.NotFoundCaps {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.NotFoundCaps, err)
	}
}

func TestRoutesV1GetRoutesAttrConnError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()["testConn"] = config.NewDfltRPCConn()
	cfg.RouteSCfg().AttributeSConns = []string{"testConn"}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dmSPP := engine.NewDataManager(data, cfg, connMng)
	routeService := NewRouteService(dmSPP, nil, cfg, connMng)
	var reply SortedRoutesList
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "CGREvent1",
		Event:  map[string]any{},
	}
	err := routeService.V1GetRoutes(context.Background(), args, &reply)
	if err == nil || err.Error() != "ROUTES_ERROR:%!s(<nil>)" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "ROUTES_ERROR:%!s(<nil>)", err)
	}
}

func TestRoutesV1GetRouteProfilesForEventError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dmSPP := engine.NewDataManager(data, cfg, connMng)
	fltr := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltr, cfg, connMng)
	var reply []*utils.RouteProfile
	args := &utils.CGREvent{
		ID:    "CGREvent1",
		Event: map[string]any{},
	}
	err := routeService.V1GetRouteProfilesForEvent(context.Background(), args, &reply)
	if err == nil || err.Error() != utils.NotFoundCaps {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.NotFoundCaps, err)
	}
}

func TestRoutesV1GetRouteProfilesForEventMsnIDError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dmSPP := engine.NewDataManager(data, cfg, connMng)
	fltr := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltr, cfg, connMng)
	var reply []*utils.RouteProfile
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event:  map[string]any{},
	}
	err := routeService.V1GetRouteProfilesForEvent(context.Background(), args, &reply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
}

func TestRoutesV1GetRouteProfilesForEventMsnEventError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dmSPP := engine.NewDataManager(data, cfg, connMng)
	fltr := engine.NewFilterS(cfg, nil, dmSPP)
	routeService := NewRouteService(dmSPP, fltr, cfg, connMng)
	var reply []*utils.RouteProfile
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "CGREvent1",
	}
	err := routeService.V1GetRouteProfilesForEvent(context.Background(), args, &reply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [Event]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [Event]", err)
	}
}

func TestRouteSV1GetRoutesListErr(t *testing.T) {
	engine.Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := engine.NewConnManager(cfg)
	dmSPP := engine.NewDataManager(data, cfg, connMng)
	fltr := engine.NewFilterS(cfg, nil, dmSPP)

	var reply *[]string
	rpS := NewRouteService(dmSPP, fltr, cfg, connMng)

	if err := rpS.V1GetRoutesList(context.Background(), testRoutesArgs[3], reply); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestRouteSMatchingRouteProfilesForEventGetRouteProfileErr1(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg, cM)
	fltrS := engine.NewFilterS(cfg, cM, dm)
	rpS := NewRouteService(dm, fltrS, cfg, cM)

	ev := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]any{
				"Route":          "RouteProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]any{
				utils.OptsRoutesProfilesCount: 1,
			},
		},
	}

	prepareRoutesData(t, dm)

	if err := engine.Cache.Set(context.Background(), utils.CacheRouteProfiles, "cgrates.org:RouteProfile1", nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, err := rpS.matchingRouteProfilesForEvent(context.Background(), "cgrates.org", ev[0]); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

type ccMock struct {
	calls map[string]func(ctx *context.Context, args any, reply any) error
}

func (ccM *ccMock) Call(ctx *context.Context, serviceMethod string, args any, reply any) (err error) {
	if call, has := ccM.calls[serviceMethod]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, args, reply)
	}
}

func TestRouteSMatchingRouteProfilesForEventGetRouteProfileErr2(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheRouteProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)

	dm := engine.NewDataManager(data, cfg, cM)
	fltrS := engine.NewFilterS(cfg, cM, dm)
	rpS := NewRouteService(dm, fltrS, cfg, cM)

	ev := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]any{
				"Route":          "RouteProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]any{
				utils.OptsRoutesProfilesCount: 1,
			},
		},
	}

	engine.Cache = engine.NewCacheS(cfg, dm, cM, nil)

	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
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

	if err := dm.SetRouteProfile(context.Background(), testRoutesPrfs[0], true); err != nil {
		t.Fatal(err)
	}

	if _, err := rpS.matchingRouteProfilesForEvent(context.Background(), "cgrates.org", ev[0]); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}

}

func TestRouteSMatchingRouteProfilesForEventPassErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg, cM)
	fltrS := engine.NewFilterS(cfg, cM, dm)
	rpS := NewRouteService(dm, fltrS, cfg, cM)

	ev := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]any{
				"Route":          "RouteProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]any{
				utils.OptsRoutesProfilesCount: 1,
			},
		},
	}

	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*engine.FilterRule{
			{
				Type:    "bad input",
				Element: "bad input",
				Values:  []string{"bad input"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
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

	if err := dm.SetRouteProfile(context.Background(), testRoutesPrfs[0], true); err != nil {
		t.Fatal(err)
	}

	expErr := "NOT_IMPLEMENTED:bad input"
	if _, err := rpS.matchingRouteProfilesForEvent(context.Background(), "cgrates.org", ev[0]); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}

}

func TestRouteSMatchingRPSForEventWeightFromDynamicsErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg, cM)
	fltrS := engine.NewFilterS(cfg, cM, dm)
	rpS := NewRouteService(dm, fltrS, cfg, cM)

	ev := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]any{
				"Route":          "RouteProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]any{
				utils.OptsRoutesProfilesCount: 1,
			},
		},
	}
	rprof := []*utils.RouteProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile1",
			FilterIDs: []string{"FLTR_RPP_1"},
			Sorting:   utils.MetaWeight,
			Routes: []*utils.Route{
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{
				{
					FilterIDs: []string{"*stirng:~*req.Account:1001"},
					Weight:    10,
				},
			},
		}}
	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
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
	if err := dm.SetRouteProfile(context.Background(), rprof[0], true); err != nil {
		t.Fatal(err)
	}

	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err := rpS.matchingRouteProfilesForEvent(context.Background(), "cgrates.org", ev[0])
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestRouteSMatchingRPSForEventBlockerFromDynamicsErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg, cM)
	fltrS := engine.NewFilterS(cfg, cM, dm)
	rpS := NewRouteService(dm, fltrS, cfg, cM)

	ev := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]any{
				"Route":          "RouteProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]any{
				utils.OptsRoutesProfilesCount: 1,
			},
		},
	}
	rprof := []*utils.RouteProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile1",
			FilterIDs: []string{"FLTR_RPP_1"},
			Sorting:   utils.MetaWeight,
			Routes: []*utils.Route{
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					FilterIDs: []string{"*stirng:~*req.Account:1001"},
					Blocker:   false,
				},
			},
		}}
	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
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
	if err := dm.SetRouteProfile(context.Background(), rprof[0], true); err != nil {
		t.Fatal(err)
	}

	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err := rpS.matchingRouteProfilesForEvent(context.Background(), "cgrates.org", ev[0])
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestNewOptsGetRoutesGetBoolOptsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.IgnoreErrors = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt([]string{"*string.invalid:filter"}, "cgrates.org", false, nil),
	}

	dataDB := engine.NewInternalDB(nil, nil, nil)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}

	expErr := `inline parse error for string: <*string.invalid:filter>`
	_, err := newOptsGetRoutes(context.Background(), ev, fS, cfg.RouteSCfg().Opts)
	if err == nil ||
		err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestNewOptsGetRoutesGetIntPointerOptsLimitErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.Limit = []*config.DynamicIntPointerOpt{
		config.NewDynamicIntPointerOpt([]string{"*string.invalid:filter"}, "cgrates.org", utils.IntPointer(4), nil),
	}

	dataDB := engine.NewInternalDB(nil, nil, nil)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}

	expErr := `inline parse error for string: <*string.invalid:filter>`
	_, err := newOptsGetRoutes(context.Background(), ev, fS, cfg.RouteSCfg().Opts)
	if err == nil ||
		err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestNewOptsGetRoutesGetIntPointerOptsOffsetErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.Offset = []*config.DynamicIntPointerOpt{
		config.NewDynamicIntPointerOpt([]string{"*string.invalid:filter"}, "cgrates.org", utils.IntPointer(4), nil),
	}

	dataDB := engine.NewInternalDB(nil, nil, nil)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}

	expErr := `inline parse error for string: <*string.invalid:filter>`
	_, err := newOptsGetRoutes(context.Background(), ev, fS, cfg.RouteSCfg().Opts)
	if err == nil ||
		err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestNewOptsGetRoutesGetIntPointerOptsMaxItemsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.MaxItems = []*config.DynamicIntPointerOpt{
		config.NewDynamicIntPointerOpt([]string{"*string.invalid:filter"}, "cgrates.org", utils.IntPointer(4), nil),
	}

	dataDB := engine.NewInternalDB(nil, nil, nil)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}

	expErr := `inline parse error for string: <*string.invalid:filter>`
	_, err := newOptsGetRoutes(context.Background(), ev, fS, cfg.RouteSCfg().Opts)
	if err == nil ||
		err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestNewOptsGetRoutesGetInterfaceOptsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.MaxCost = []*config.DynamicInterfaceOpt{
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     2,
		},
	}

	dataDB := engine.NewInternalDB(nil, nil, nil)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]any{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]any{},
	}

	expErr := `inline parse error for string: <*string.invalid:filter>`
	_, err := newOptsGetRoutes(context.Background(), ev, fS, cfg.RouteSCfg().Opts)
	if err == nil ||
		err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestSortedRoutesForEventsortedRoutesForProfileErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.ProfileCount = []*config.DynamicIntPointerOpt{
		config.NewDynamicIntPointerOpt([]string{"*string.invalid:filter"}, "cgrates.org", utils.IntPointer(4), nil),
	}

	dataDB := engine.NewInternalDB(nil, nil, nil)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := &RouteS{
		dm:      dm,
		fltrS:   fS,
		cfg:     cfg,
		connMgr: nil,
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 1,
		},
	}

	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
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

	rp := &utils.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "RouteProfile1",
		FilterIDs: []string{"FLTR_RPP_1"},
		Sorting:   utils.MetaWeight,
		Routes: []*utils.Route{
			{
				ID:              "route1",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	if err := dm.SetRouteProfile(context.Background(), rp, true); err != nil {
		t.Fatal(err)
	}

	expErr := `unsupported sorting strategy: *weight`
	_, err := routeService.sortedRoutesForEvent(context.Background(), ev.Tenant, ev)
	if err == nil ||
		err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestSortedRoutesForEventGetIntPointerOptsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.ProfileCount = []*config.DynamicIntPointerOpt{
		config.NewDynamicIntPointerOpt([]string{"*string.invalid:filter"}, "cgrates.org", utils.IntPointer(4), nil),
	}

	dataDB := engine.NewInternalDB(nil, nil, nil)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := &RouteS{
		dm:      dm,
		fltrS:   fS,
		cfg:     cfg,
		connMgr: nil,
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{},
	}

	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
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

	rp := &utils.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "RouteProfile1",
		FilterIDs: []string{"FLTR_RPP_1"},
		Sorting:   utils.MetaWeight,
		Routes: []*utils.Route{
			{
				ID:              "route1",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	if err := dm.SetRouteProfile(context.Background(), rp, true); err != nil {
		t.Fatal(err)
	}

	expErr := `inline parse error for string: <*string.invalid:filter>`
	_, err := routeService.sortedRoutesForEvent(context.Background(), ev.Tenant, ev)
	if err == nil ||
		err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestSortedRoutesForEventNewOptsGetRoutesErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.IgnoreErrors = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt([]string{"*string.invalid:filter"}, "cgrates.org", false, nil),
	}

	dataDB := engine.NewInternalDB(nil, nil, nil)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := &RouteS{
		dm:      dm,
		fltrS:   fS,
		cfg:     cfg,
		connMgr: nil,
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{},
	}

	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
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

	rp := &utils.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "RouteProfile1",
		FilterIDs: []string{"FLTR_RPP_1"},
		Sorting:   utils.MetaWeight,
		Routes: []*utils.Route{
			{
				ID:              "route1",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	if err := dm.SetRouteProfile(context.Background(), rp, true); err != nil {
		t.Fatal(err)
	}

	expErr := `inline parse error for string: <*string.invalid:filter>`
	_, err := routeService.sortedRoutesForEvent(context.Background(), ev.Tenant, ev)
	if err == nil ||
		err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestSortedRoutesForEventExceedMaxItemsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.MaxItems = []*config.DynamicIntPointerOpt{
		config.NewDynamicIntPointerOpt(nil, "cgrates.org", utils.IntPointer(1), nil),
	}
	cfg.RouteSCfg().Opts.Limit = []*config.DynamicIntPointerOpt{
		config.NewDynamicIntPointerOpt(nil, "cgrates.org", utils.IntPointer(2), nil),
	}

	dataDB := engine.NewInternalDB(nil, nil, nil)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := &RouteS{
		dm:      dm,
		fltrS:   fS,
		cfg:     cfg,
		connMgr: nil,
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{},
	}

	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
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

	rp := &utils.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "RouteProfile1",
		FilterIDs: []string{"FLTR_RPP_1"},
		Sorting:   utils.MetaWeight,
		Routes: []*utils.Route{
			{
				ID:              "route1",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	if err := dm.SetRouteProfile(context.Background(), rp, true); err != nil {
		t.Fatal(err)
	}

	expErr := `SERVER_ERROR: maximum number of items exceeded`
	_, err := routeService.sortedRoutesForEvent(context.Background(), ev.Tenant, ev)
	if err == nil ||
		err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestRouteSV1GetRoutesGetStringOptsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.Context = []*config.DynamicStringOpt{
		config.NewDynamicStringOpt([]string{"*string.invalid:filter"}, "cgrates.org", "value2", nil),
	}
	cfg.RouteSCfg().AttributeSConns = []string{"testConn"}

	dataDB := engine.NewInternalDB(nil, nil, nil)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := &RouteS{
		dm:      dm,
		fltrS:   fS,
		cfg:     cfg,
		connMgr: nil,
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{},
	}

	var reply *SortedRoutesList
	expErr := `inline parse error for string: <*string.invalid:filter>`
	err := routeService.V1GetRoutes(context.Background(), ev, reply)
	if err == nil ||
		err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestRoutesV1GetRoutesCallWithAlteredFields(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()["testConn"] = config.NewDfltRPCConn()
	cfg.RouteSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes)}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*engine.AttrSProcessEventReply) = engine.AttrSProcessEventReply{
					AlteredFields: []*engine.FieldsAltered{{
						Fields: []string{utils.AccountField},
					}},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "utils.CGREvent1",
						Event: map[string]any{
							"Route":          "RouteProfile1",
							utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
							"UsageInterval":  "1s",
							"PddInterval":    "1s",
							utils.Weight:     "20.0",
						},
						APIOpts: map[string]any{
							utils.OptsRoutesProfilesCount: 1,
						},
					},
				}
				return nil
			},
		},
	}

	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, cc)

	dm := engine.NewDataManager(data, cfg, cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	routeService := NewRouteService(dm, fS, cfg, cM)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 1,
		},
	}

	rp := &utils.RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*utils.Route{
			{
				ID:              "route1",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	if err := dm.SetRouteProfile(context.Background(), rp, true); err != nil {
		t.Fatal(err)
	}

	var reply SortedRoutesList

	exp := SortedRoutesList{
		{
			ProfileID: "RouteProfile1",
			Sorting:   utils.MetaWeight,
			Routes: []*SortedRoute{
				{
					RouteID:         "route1",
					RouteParameters: "param1",
					SortingData: map[string]any{
						"Weight": 10,
					},
				},
			},
		},
	}

	err := routeService.V1GetRoutes(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(reply)) {
		t.Errorf("Expected error <%+v>, received error <%+v>", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestRoutesV1GetRoutesSortedRoutesForEventErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.ProfileCount = []*config.DynamicIntPointerOpt{
		config.NewDynamicIntPointerOpt(nil, "cgrates.org", utils.IntPointer(4), nil),
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := &RouteS{
		dm:      dm,
		fltrS:   fS,
		cfg:     cfg,
		connMgr: nil,
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 1,
		},
	}

	rp := &utils.RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*utils.Route{
			{
				ID:              "route1",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	if err := dm.SetRouteProfile(context.Background(), rp, true); err != nil {
		t.Fatal(err)
	}

	var reply SortedRoutesList

	expErr := `SERVER_ERROR: unsupported sorting strategy: *weight`
	err := routeService.V1GetRoutes(context.Background(), args, &reply)
	if err == nil ||
		err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}
}

func TestV1GetRouteProfilesForEventMatchingRouteProfErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(data, cfg, cM)
	fltrS := engine.NewFilterS(cfg, cM, dm)
	rpS := NewRouteService(dm, fltrS, cfg, cM)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 1,
		},
	}

	if err := dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*engine.FilterRule{
			{
				Type:    "bad input",
				Element: "bad input",
				Values:  []string{"bad input"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
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

	if err := dm.SetRouteProfile(context.Background(), testRoutesPrfs[0], true); err != nil {
		t.Fatal(err)
	}

	var reply []*utils.RouteProfile

	expErr := "SERVER_ERROR: NOT_IMPLEMENTED:bad input"
	err := rpS.V1GetRouteProfilesForEvent(context.Background(), args, &reply)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}
}

func TestV1GetRouteProfilesForEventOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rpS := NewRouteService(dm, fltrs, cfg, nil)

	prepareRoutesData(t, dm)

	exp := []*utils.RouteProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile1",
			FilterIDs: []string{"FLTR_RPP_1"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			Sorting: utils.MetaWeight,
			Routes: []*utils.Route{
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
		},
	}

	var reply []*utils.RouteProfile
	err := rpS.V1GetRouteProfilesForEvent(context.Background(), testRoutesArgs[0], &reply)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(utils.ToJSON(reply), utils.ToJSON(exp)) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}

}

func TestRoutessortedRoutesForProfileLazyPassErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := &RouteS{
		dm:      dm,
		fltrS:   fS,
		cfg:     cfg,
		connMgr: nil,
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 1,
		},
	}

	rp := &utils.RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,

		Routes: []*utils.Route{
			{
				ID:              "route1",
				FilterIDs:       []string{"bad fltr"},
				Weights:         utils.DynamicWeights{{Weight: 10}},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	pag := utils.Paginator{}
	extraOpts := &optsGetRoutes{}
	expErr := "NOT_FOUND:bad fltr"
	if _, err := routeService.sortedRoutesForProfile(context.Background(), "cgrates.org", rp, args, pag, extraOpts); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v> ", expErr, err.Error())

	}
}

func TestRoutessortedRoutesForProfileLazyPassFalse(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := &RouteS{
		dm:      dm,
		fltrS:   fS,
		cfg:     cfg,
		connMgr: nil,
		sorter: RouteSortDispatcher{
			utils.MetaWeight: NewWeightSorter(cfg),
		},
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 1,
		},
	}

	rp := &utils.RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*utils.Route{
			{
				ID:              "route1",
				FilterIDs:       []string{"*string:~*req.Account:1010", "*string:~*vars.Field1:Val1"},
				Weights:         utils.DynamicWeights{{Weight: 10}},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	pag := utils.Paginator{}
	extraOpts := &optsGetRoutes{}
	exp := SortedRoutes{
		ProfileID: "RouteProfile1",
		Sorting:   "*weight",
		Routes:    []*SortedRoute{},
	}
	if rcv, err := routeService.sortedRoutesForProfile(context.Background(), "cgrates.org", rp, args, pag, extraOpts); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestRoutessortedRoutesForProfileWeightFromDynamicsErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := &RouteS{
		dm:      dm,
		fltrS:   fS,
		cfg:     cfg,
		connMgr: nil,
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 1,
		},
	}

	rp := &utils.RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,

		Routes: []*utils.Route{
			{
				ID: "route1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: []string{"*stirng:~*req.Account:1001"},
						Weight:    10,
					},
				},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	pag := utils.Paginator{}
	extraOpts := &optsGetRoutes{}
	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err := routeService.sortedRoutesForProfile(context.Background(), "cgrates.org", rp, args, pag, extraOpts)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}
}

func TestRoutessortedRoutesForProfileBlockerFromDynamicsErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := &RouteS{
		dm:      dm,
		fltrS:   fS,
		cfg:     cfg,
		connMgr: nil,
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 1,
		},
	}

	rp := &utils.RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,

		Routes: []*utils.Route{
			{
				ID: "route1",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						FilterIDs: []string{"*stirng:~*req.Account:1001"},
						Blocker:   false,
					},
				},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	pag := utils.Paginator{}
	extraOpts := &optsGetRoutes{}
	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err := routeService.sortedRoutesForProfile(context.Background(), "cgrates.org", rp, args, pag, extraOpts)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}
}

func TestRoutessortedRoutesForProfileSortHasBlocker(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := NewRouteService(dm, fS, cfg, nil)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 1,
		},
	}

	rp := &utils.RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*utils.Route{
			{
				ID:              "route1",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				Blockers:        utils.DynamicBlockers{{Blocker: true}},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	pag := utils.Paginator{}
	extraOpts := &optsGetRoutes{}
	exp := SortedRoutes{
		ProfileID: "RouteProfile1",
		Sorting:   "*weight",
		Routes: []*SortedRoute{
			{
				RouteID:         "route1",
				RouteParameters: "param1",
				SortingData:     map[string]any{utils.Blocker: true, utils.Weight: 10}},
		},
	}

	if rcv, err := routeService.sortedRoutesForProfile(context.Background(), "cgrates.org", rp, args, pag, extraOpts); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expecting: \n%+v,\n received: \n%+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestRoutessortedRoutesForEventNoSortedRoutesErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.Offset = []*config.DynamicIntPointerOpt{
		config.NewDynamicIntPointerOpt(nil, "cgrates.org", utils.IntPointer(10), nil),
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := NewRouteService(dm, fS, cfg, nil)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 1,
		},
	}

	rp := &utils.RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*utils.Route{
			{
				ID:              "route1",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				Blockers:        utils.DynamicBlockers{{Blocker: true}},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	if err := dm.SetRouteProfile(context.Background(), rp, true); err != nil {
		t.Fatal(err)
	}

	_, err := routeService.sortedRoutesForEvent(context.Background(), "cgrates.org", args)
	if err != utils.ErrNotFound {
		t.Errorf("Expected error <%+v>, received error <%+v>", utils.ErrNotFound, err)
	}
}

func TestRouteSV1GetRoutesListOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fS := engine.NewFilterS(cfg, nil, dm)
	routeService := NewRouteService(dm, fS, cfg, nil)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Route":          "RouteProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfilesCount: 1,
		},
	}

	rp := &utils.RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*utils.Route{
			{
				ID:              "route1",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				RouteParameters: "param1",
			},
		},
		Weights:  utils.DynamicWeights{{Weight: 10}},
		Blockers: utils.DynamicBlockers{{Blocker: true}},
	}

	if err := dm.SetRouteProfile(context.Background(), rp, true); err != nil {
		t.Fatal(err)
	}

	exp := []string{"route1:param1"}

	var reply []string
	err := routeService.V1GetRoutesList(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(reply)) {
		t.Errorf("Expecting: \n%+v,\n received: \n%+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}
