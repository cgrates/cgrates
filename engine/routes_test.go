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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/utils"
)

var (
	testRoutesPrfs = []*RouteProfile{
		{
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
			Weights:  utils.DynamicWeights{{Weight: 10}},
			Blockers: utils.DynamicBlockers{{Blocker: true}},
		},
		{
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
		{
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
	sprf, err := newOptsGetRoutes(context.Background(), ev, &FilterS{}, config.CgrConfig().RouteSCfg().Opts)
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
		APIOpts: map[string]any{
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
		{
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
		{
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
		{
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
		{
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
		Event:   map[string]any{},
		APIOpts: map[string]any{utils.OptsRoutesProfilesCount: 3},
	}

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)

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
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
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
		Event:  map[string]any{},
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
		Event: map[string]any{},
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
		Event:  map[string]any{},
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
		Event: map[string]any{},
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
		Event:  map[string]any{},
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
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:   utils.DynamicWeights{{}},
		Blockers: utils.DynamicBlockers{
			{Blocker: false},
		},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}
	if err := rp.Set([]string{}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{"", ""}, "", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{"NotAField"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{"NotAField", "1"}, ":", false, utils.EmptyString); err != utils.ErrWrongPath {
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
	if err := rp.Set([]string{utils.Weights}, ";0", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Blockers}, ";false", false, utils.EmptyString); err != nil {
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
	if err := rp.Set([]string{utils.Routes, utils.Weights}, ";0", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{utils.Routes, utils.Blockers}, ";true", false, utils.EmptyString); err != nil {
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

func TestRouteProfileAsInterface(t *testing.T) {
	rp := RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Blockers:          utils.DynamicBlockers{{Blocker: false}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}
	if _, err := rp.FieldAsInterface(nil); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{"field"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{"field", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.FieldAsInterface([]string{utils.Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := utils.ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ";0"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := ";false"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.SortingParameters}); err != nil {
		t.Fatal(err)
	} else if exp := rp.SortingParameters; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.SortingParameters + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.SortingParameters[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Sorting}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Sorting; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if _, err := rp.FieldAsInterface([]string{utils.Routes + "[4]", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", "", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ";0"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := ";true"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.RouteParameters}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].RouteParameters; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.AccountIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].AccountIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.AccountIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].AccountIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.RateProfileIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].RateProfileIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.RateProfileIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].RateProfileIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.ResourceIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].ResourceIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.ResourceIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].ResourceIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.StatIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].StatIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{utils.Routes + "[0]", utils.StatIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].StatIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := rp.FieldAsString([]string{""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.FieldAsString([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := rp.String(), utils.ToJSON(rp); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := rp.Routes[0].FieldAsString([]string{""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.Routes[0].FieldAsString([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := "RT1"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := rp.Routes[0].String(), utils.ToJSON(rp.Routes[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestRouteProfileMerge(t *testing.T) {
	dp := &RouteProfile{}
	exp := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}
	if dp.Merge(&RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}); !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dp))
	}
}

func TestRouteMerge(t *testing.T) {

	route := &Route{}

	routeV2 := &Route{
		ID:              "RouteId",
		RouteParameters: "RouteParam",
		Weights:         utils.DynamicWeights{{Weight: 10}},
		Blockers:        utils.DynamicBlockers{{Blocker: false}},
		FilterIDs:       []string{"FltrId"},
		AccountIDs:      []string{"AccId"},
		RateProfileIDs:  []string{"RateProfileId"},
		ResourceIDs:     []string{"ResourceId"},
		StatIDs:         []string{"StatId"},
	}
	exp := routeV2

	route.Merge(routeV2)
	if !reflect.DeepEqual(route, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(route))
	}
}

func TestRouteProfileCompileCacheParametersErrParse(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)

	rp := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaLoad,
		SortingParameters: []string{"sort:param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	expErr := `strconv.Atoi: parsing "param": invalid syntax`
	if err := rp.compileCacheParameters(); err.Error() != expErr || err == nil {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestRouteProfileCompileCacheParametersConfigRatio(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)

	rp := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaLoad,
		SortingParameters: []string{"param:1"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	expErr := `strconv.Atoi: parsing "param": invalid syntax`
	if err := rp.compileCacheParameters(); err != nil {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestRouteProfileCompileCacheParametersDefaultRatio(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)

	rp := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaLoad,
		SortingParameters: []string{"*default:1"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	expErr := `strconv.Atoi: parsing "param": invalid syntax`
	if err := rp.compileCacheParameters(); err != nil {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestRouteProfileCompileCacheParametersRouteRatio(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)

	rp := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaLoad,
		SortingParameters: []string{"RT1:1"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}

	expErr := `strconv.Atoi: parsing "param": invalid syntax`
	if err := rp.compileCacheParameters(); err != nil {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}
}

func TestRouteSV1GetRoutesListErr(t *testing.T) {
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	connMng := NewConnManager(cfg)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), connMng)
	fltr := &FilterS{dm: dmSPP, cfg: cfg}

	var reply *[]string
	rpS := NewRouteService(dmSPP, fltr, cfg, connMng)

	if err := rpS.V1GetRoutesList(context.Background(), testRoutesArgs[3], reply); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestRouteSMatchingRouteProfilesForEventGetRouteProfileErr1(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), cM)
	fltrS := NewFilterS(cfg, cM, dm)
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

	if err := Cache.Set(context.Background(), utils.CacheRouteProfiles, "cgrates.org:RouteProfile1", nil, []string{}, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	if _, err := rpS.matchingRouteProfilesForEvent(context.Background(), "cgrates.org", ev[0]); err != utils.ErrNotFound {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotFound, err)
	}

}

func TestRouteSMatchingRouteProfilesForEventGetRouteProfileErr2(t *testing.T) {

	cfgtmp := config.CgrConfig()
	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
		config.SetCgrConfig(cfgtmp)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator)}
	cfg.CacheCfg().Partitions[utils.CacheRouteProfiles].Replicate = true
	config.SetCgrConfig(cfg)

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReplicateSet: func(ctx *context.Context, args, reply any) error { return utils.ErrNotImplemented },
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaReplicator), utils.CacheSv1, cc)

	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), cM)
	fltrS := NewFilterS(cfg, cM, dm)
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

	Cache = NewCacheS(cfg, dm, cM, nil)

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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), cM)
	fltrS := NewFilterS(cfg, cM, dm)
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

	if err := dm.SetFilter(context.Background(), &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*FilterRule{
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), cM)
	fltrS := NewFilterS(cfg, cM, dm)
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
	rprof := []*RouteProfile{
		{
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
			Weights: utils.DynamicWeights{
				{
					FilterIDs: []string{"*stirng:~*req.Account:1001"},
					Weight:    10,
				},
			},
		}}
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), cM)
	fltrS := NewFilterS(cfg, cM, dm)
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
	rprof := []*RouteProfile{
		{
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

	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)

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

	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)

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

	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)

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

	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)

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

	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)

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

	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
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

	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
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

	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
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

	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
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

	dataDB := NewInternalDB(nil, nil, nil)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()["testConn"] = config.NewDfltRPCConn()
	cfg.RouteSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes)}
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*AttrSProcessEventReply) = AttrSProcessEventReply{
					AlteredFields: []*FieldsAltered{{
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

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, cc)

	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), cM)
	fS := NewFilterS(cfg, cM, dm)
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

	rp := &RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*Route{
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.ProfileCount = []*config.DynamicIntPointerOpt{
		config.NewDynamicIntPointerOpt(nil, "cgrates.org", utils.IntPointer(4), nil),
	}
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*Route{
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	cM := NewConnManager(cfg)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), cM)
	fltrS := NewFilterS(cfg, cM, dm)
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

	if err := dm.SetFilter(context.Background(), &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RPP_1",
		Rules: []*FilterRule{
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

	var reply []*RouteProfile

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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	rpS := NewRouteService(dm, &FilterS{
		dm: dm, cfg: cfg}, cfg, nil)

	prepareRoutesData(t, dm)

	exp := []*RouteProfile{
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
			Routes: []*Route{
				{
					ID:              "route1",
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
		},
	}

	var reply []*RouteProfile
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,

		Routes: []*Route{
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*Route{
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,

		Routes: []*Route{
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,

		Routes: []*Route{
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*Route{
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().Opts.Offset = []*config.DynamicIntPointerOpt{
		config.NewDynamicIntPointerOpt(nil, "cgrates.org", utils.IntPointer(10), nil),
	}

	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*Route{
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	fS := NewFilterS(cfg, nil, dm)
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

	rp := &RouteProfile{
		Tenant:  "cgrates.org",
		ID:      "RouteProfile1",
		Sorting: utils.MetaWeight,
		Routes: []*Route{
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

func TestRouteProfileMergeWithRPRoutes(t *testing.T) {
	dp := &RouteProfile{
		Routes: []*Route{
			{
				ID: "RT1",
			},
		},
	}
	exp := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}
	if dp.Merge(&RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           utils.DynamicWeights{{}},
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        utils.DynamicWeights{{}},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}); !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dp))
	}
}
