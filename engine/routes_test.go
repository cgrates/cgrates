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

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/utils"
)

func TestRoutesSort(t *testing.T) {
	sprs := RouteProfiles{
		&RouteProfile{
			Tenant: "cgrates.org",
			ID:     "RoutePrf1",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting: "",
			Routes: []*Route{
				{
					ID:              "route1",
					Weight:          10.0,
					RouteParameters: "param1",
				},
			},
			Weight: 10,
		},
		&RouteProfile{
			Tenant: "cgrates.org",
			ID:     "RoutePrf2",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting: "",
			Routes: []*Route{
				{
					ID:              "route1",
					Weight:          20.0,
					RouteParameters: "param1",
				},
			},
			Weight: 20.0,
		},
	}
	eRouteProfile := RouteProfiles{
		&RouteProfile{
			Tenant: "cgrates.org",
			ID:     "RoutePrf2",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting: "",
			Routes: []*Route{
				{
					ID:              "route1",
					Weight:          20.0,
					RouteParameters: "param1",
				},
			},
			Weight: 20.0,
		},
		&RouteProfile{
			Tenant: "cgrates.org",
			ID:     "RoutePrf1",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting: "",
			Routes: []*Route{
				{
					ID:              "route1",
					Weight:          10.0,
					RouteParameters: "param1",
				},
			},
			Weight: 10.0,
		},
	}
	sprs.Sort()
	if !reflect.DeepEqual(eRouteProfile, sprs) {
		t.Errorf("Expecting: %+v, received: %+v", eRouteProfile, sprs)
	}
}

var (
	testRoutesPrfs = RouteProfiles{
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile1",
			FilterIDs: []string{"FLTR_RPP_1"},
			Sorting:   utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weight:          10.0,
					RouteParameters: "param1",
				},
			},
			Weight: 10,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile2",
			FilterIDs: []string{"FLTR_SUPP_2"},
			Sorting:   utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route2",
					Weight:          20.0,
					RouteParameters: "param2",
				},
				{
					ID:              "route3",
					Weight:          10.0,
					RouteParameters: "param3",
				},
				{
					ID:              "route1",
					Weight:          30.0,
					RouteParameters: "param1",
				},
			},
			Weight: 20.0,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfilePrefix",
			FilterIDs: []string{"FLTR_SUPP_3"},
			Sorting:   utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weight:          10.0,
					RouteParameters: "param1",
				},
			},
			Weight: 10,
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
			APIOpts: map[string]interface{}{},
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
			APIOpts: map[string]interface{}{},
		},
		{ //matching RouteProfilePrefix
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route": "RouteProfilePrefix",
			},
			APIOpts: map[string]interface{}{},
		},
		{
			Tenant: "cgrates.org",
			ID:     "CGR",
			Event: map[string]interface{}{
				"UsageInterval": "1s",
				"PddInterval":   "1s",
			},
			APIOpts: map[string]interface{}{},
		},
	}
)

func prepareRoutesData(t *testing.T, dm *DataManager) {
	if err := dm.SetFilter(&Filter{
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
	if err := dm.SetFilter(&Filter{
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
	if err := dm.SetFilter(&Filter{
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
		if err = dm.SetRouteProfile(spp, true); err != nil {
			t.Fatal(err)
		}
	}
	//Test each route profile from cache
	for _, spp := range testRoutesPrfs {
		if tempSpp, err := dm.GetRouteProfile(spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(spp, tempSpp) {
			t.Fatalf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}
}

func TestRoutesCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	prepareRoutesData(t, dmSPP)
}

func TestRoutesmatchingRouteProfilesForEvent(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	prepareRoutesData(t, dmSPP)
	for i, spp := range testRoutesPrfs {
		sprf, err := routeService.matchingRouteProfilesForEvent("cgrates.org", testRoutesArgs[i])
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(spp, sprf[0]) {
			t.Errorf("Expecting: %+v, received: %+v", spp, sprf[0])
		}
	}
}

func TestRoutesSortedForEvent(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	prepareRoutesData(t, dmSPP)

	exp := SortedRoutesList{{
		ProfileID: "RouteProfile1",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}}
	sprf, err := routeService.sortedRoutesForEvent("cgrates.org", testRoutesArgs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", exp, sprf)
	}

	exp = SortedRoutesList{{
		ProfileID: "RouteProfile2",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight: 30.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 30.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight: 20.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}}

	sprf, err = routeService.sortedRoutesForEvent("cgrates.org", testRoutesArgs[1])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", exp, sprf)
	}

	exp = SortedRoutesList{{
		ProfileID: "RouteProfilePrefix",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}}

	sprf, err = routeService.sortedRoutesForEvent("cgrates.org", testRoutesArgs[2])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", exp, sprf)
	}
}

func TestRoutesSortedForEventWithLimit(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)

	prepareRoutesData(t, dmSPP)

	eFirstRouteProfile := SortedRoutesList{&SortedRoutes{
		ProfileID: "RouteProfile2",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight: 30.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 30.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight: 20.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
		},
	}}
	testRoutesArgs[1].APIOpts[utils.OptsRoutesLimit] = 2
	delete(testRoutesArgs[1].APIOpts, utils.OptsRoutesOffset)
	sprf, err := routeService.sortedRoutesForEvent("cgrates.org", testRoutesArgs[1])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstRouteProfile, sprf)
	}
}

func TestRoutesSortedForEventWithOffset(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
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
				sortingDataF64: map[string]float64{
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}}
	testRoutesArgs[1].APIOpts[utils.OptsRoutesOffset] = 2
	delete(testRoutesArgs[1].APIOpts, utils.OptsRoutesLimit)
	sprf, err := routeService.sortedRoutesForEvent("cgrates.org", testRoutesArgs[1])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}
}

func TestRoutesSortedForEventWithLimitAndOffset(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
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
				sortingDataF64: map[string]float64{
					utils.Weight: 20.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
		},
	}}
	testRoutesArgs[1].APIOpts[utils.OptsRoutesLimit] = 1
	testRoutesArgs[1].APIOpts[utils.OptsRoutesOffset] = 1
	sprf, err := routeService.sortedRoutesForEvent("cgrates.org", testRoutesArgs[1])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}
}

func TestRoutesAsOptsGetRoutesMaxCost(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)

	prepareRoutesData(t, dmSPP)

	routeService.cgrcfg.RouteSCfg().IndexedSelects = false
	sprf, err := routeService.matchingRouteProfilesForEvent("cgrates.org", testRoutesArgs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testRoutesPrfs[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", testRoutesPrfs[0], sprf[0])
	}

	sprf, err = routeService.matchingRouteProfilesForEvent("cgrates.org", testRoutesArgs[1])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testRoutesPrfs[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", testRoutesPrfs[1], sprf[0])
	}

	sprf, err = routeService.matchingRouteProfilesForEvent("cgrates.org", testRoutesArgs[2])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testRoutesPrfs[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", testRoutesPrfs[2], sprf[0])
	}
}

func TestRoutesMatchWithIndexFalse(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	prepareRoutesData(t, dmSPP)

	routeService.cgrcfg.RouteSCfg().IndexedSelects = false
	sprf, err := routeService.matchingRouteProfilesForEvent("cgrates.org", testRoutesArgs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testRoutesPrfs[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", testRoutesPrfs[0], sprf[0])
	}

	sprf, err = routeService.matchingRouteProfilesForEvent("cgrates.org", testRoutesArgs[1])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testRoutesPrfs[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", testRoutesPrfs[1], sprf[0])
	}

	sprf, err = routeService.matchingRouteProfilesForEvent("cgrates.org", testRoutesArgs[2])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testRoutesPrfs[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", testRoutesPrfs[2], sprf[0])
	}
}

func TestRoutesSortedForEventWithLimitAndOffset2(t *testing.T) {
	Cache.Clear(nil)
	testRoutesPrfs := RouteProfiles{
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfile1",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route2",
					Weight:          10.0,
					RouteParameters: "param1",
				},
			},
			Weight: 10,
		},
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfile2",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route2",
					Weight:          20.0,
					RouteParameters: "param2",
				},
				{
					ID:              "route3",
					Weight:          10.0,
					RouteParameters: "param3",
				},
				{
					ID:              "route1",
					Weight:          30.0,
					RouteParameters: "param1",
				},
			},
			Weight: 5,
		},
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfilePrefix",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weight:          10.0,
					RouteParameters: "param1",
				},
			},
			Weight: 20,
		},
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfilePrefix4",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weight:          10.0,
					RouteParameters: "param1",
				},
			},
			Weight: 0,
		},
	}
	argsGetRoutes := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event:  map[string]interface{}{},
		APIOpts: map[string]interface{}{
			utils.OptsRoutesProfileCount: 3,
		},
	}

	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)

	for _, spp := range testRoutesPrfs {
		if err = dmSPP.SetRouteProfile(spp, true); err != nil {
			t.Fatal(err)
		}
		if tempSpp, err := dmSPP.GetRouteProfile(spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Fatal(err)
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
					sortingDataF64: map[string]float64{
						utils.Weight: 10.,
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
					sortingDataF64: map[string]float64{
						utils.Weight: 30.,
					},
					SortingData: map[string]interface{}{
						utils.Weight: 30.,
					},
					RouteParameters: "param1",
				},
			},
		},
	}
	argsGetRoutes.APIOpts[utils.OptsRoutesLimit] = 2
	argsGetRoutes.APIOpts[utils.OptsRoutesOffset] = 1
	sprf, err := routeService.sortedRoutesForEvent(argsGetRoutes.Tenant, argsGetRoutes)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}
}
func TestRouteProfileCompileCacheParameters(t *testing.T) {
	rp := &RouteProfile{
		Tenant:    "tnt",
		ID:        "id2",
		FilterIDs: []string{"filter1", "filter2", "filter3"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2022, 12, 1, 8, 0, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),
		},
		Weight:            12,
		Sorting:           "sort",
		SortingParameters: []string{"sort_param:3", "*ratio:1"},
		Routes: []*Route{
			{
				ID:              "id1",
				FilterIDs:       []string{"filter_id1", "filter_id2", "filter_id3", "filter_id4"},
				AccountIDs:      []string{"acc_id1", "acc_id2", "acc_id3", "acc_id3", "acc_id4"},
				RatingPlanIDs:   []string{"rating_id1", "rating_id2", "rating_id3", "rating_id4"},
				ResourceIDs:     []string{"res_id1", "res_id2", "res_id3", "res_id4", "res_id4"},
				StatIDs:         []string{"stats_id1", "stats_id2", "stats_id3", "stats_id3", "stats_id4"},
				Weight:          2.3,
				Blocker:         true,
				RouteParameters: "param",

				lazyCheckRules: []*FilterRule{
					{
						Type:    "*string",
						Element: "elem",
						Values:  []string{"val1", "val2", "val3"},
						rsrValues: config.RSRParsers{
							&config.RSRParser{Rules: "public"},
							{Rules: "private"},
						},
					},
				},
			},
			{
				ID:              "id1",
				FilterIDs:       []string{"filter_id1", "filter_id2", "filter_id3", "filter_id4"},
				AccountIDs:      []string{"acc_id1", "acc_id2", "acc_id3", "acc_id3", "acc_id4"},
				RatingPlanIDs:   []string{"rating_id1", "rating_id2", "rating_id3", "rating_id4"},
				ResourceIDs:     []string{"res_id1", "res_id2", "res_id3", "res_id4", "res_id4"},
				StatIDs:         []string{"stats_id1", "stats_id2", "stats_id3", "stats_id3", "stats_id4"},
				Weight:          2.3,
				Blocker:         true,
				RouteParameters: "param",

				lazyCheckRules: []*FilterRule{
					{
						Type:    "*string",
						Element: "elem",
						Values:  []string{"val1", "val2", "val3"},
						rsrValues: config.RSRParsers{
							&config.RSRParser{Rules: "public"},
							{Rules: "private"},
						},
					},
				},
			},
		},
	}

	if err := rp.compileCacheParameters(); err != nil {
		t.Error(err)
	}
	rp.Sorting = utils.MetaLoad
	if err = rp.compileCacheParameters(); err != nil {
		t.Error(err)
	}

}

func TestRouteServiceCostForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	cfg.RouteSCfg().RALsConns = nil
	routeService := NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: cfg}, cfg, nil)
	ev := &utils.CGREvent{
		Tenant: "tnt",
		ID:     "id",
		Time:   utils.TimePointer(time.Date(2022, 12, 1, 20, 0, 0, 0, time.UTC)),
		Event: map[string]interface{}{
			utils.AccountField: "acc_event",
			utils.Destination:  "desc_event",
			utils.SetupTime:    time.Now(),
		},
	}
	if _, err := routeService.costForEvent(ev, []string{}, []string{}); err == nil {
		t.Error(err)
	} else if _, err := routeService.costForEvent(ev, []string{"acc1", "acc2", "acc3"}, []string{}); err == nil {
		t.Error(err)
	}

}

func TestRouteServiceStatMetrics(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	rpS := NewRouteService(dmSPP, &FilterS{dm: dmSPP, cfg: cfg, connMgr: nil}, cfg, nil)
	if _, err := rpS.statMetrics([]string{"stat1", "stat2"}, "cgrates.org"); err != nil {
		t.Error(err)
	}
}

func TestRouteServiceV1GetRouteProfilesForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	rpS := NewRouteService(dmSPP, &FilterS{dm: dmSPP, cfg: cfg, connMgr: nil}, cfg, nil)
	args := &utils.CGREvent{
		Tenant: "cgrates.orgs",
		ID:     "id",
		Time:   utils.TimePointer(time.Date(2022, 12, 1, 20, 0, 0, 0, time.UTC)),
		Event: map[string]interface{}{
			utils.AccountField: "acc_event",
			utils.Destination:  "desc_event",
			utils.SetupTime:    time.Now(),
		},
	}
	testRoutesPrfs := &RouteProfiles{
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfile1",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route2",
					Weight:          10.0,
					RouteParameters: "param1",
				},
			},
			Weight: 10,
		},
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfile2",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route2",
					Weight:          20.0,
					RouteParameters: "param2",
				},
				{
					ID:              "route3",
					Weight:          10.0,
					RouteParameters: "param3",
				},
				{
					ID:              "route1",
					Weight:          30.0,
					RouteParameters: "param1",
				},
			},
			Weight: 5,
		},
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfilePrefix",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weight:          10.0,
					RouteParameters: "param1",
				},
			},
			Weight: 20,
		},
		&RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "RouteProfilePrefix4",
			Sorting: utils.MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weight:          10.0,
					RouteParameters: "param1",
				},
			},
			Weight: 0,
		},
	}
	if err := rpS.V1GetRouteProfilesForEvent(args, (*[]*RouteProfile)(testRoutesPrfs)); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestRouteServiceV1GetRoutes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg.RouteSCfg().StringIndexedFields = nil
	cfg.RouteSCfg().PrefixIndexedFields = nil
	rpS := NewRouteService(dmSPP, &FilterS{dm: dmSPP, cfg: cfg, connMgr: nil}, cfg, nil)
	args := &utils.CGREvent{
		ID:     "CGREvent1",
		Tenant: "cgrates.orgs",
		Time:   utils.TimePointer(time.Date(2022, 12, 1, 20, 0, 0, 0, time.UTC)),
		Event: map[string]interface{}{
			utils.AccountField: "acc_event",
			utils.Destination:  "desc_event",
			utils.SetupTime:    time.Now(),
		},
	}
	reply := &SortedRoutesList{
		{
			ProfileID: "RouteProfile1",
			Sorting:   utils.MetaWeight,
			Routes: []*SortedRoute{
				{
					RouteID: "route1",
					sortingDataF64: map[string]float64{
						utils.Weight: 10.0,
					},
					SortingData: map[string]interface{}{
						utils.Weight: 10.0,
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
					sortingDataF64: map[string]float64{
						utils.Weight: 10.0,
					},
					SortingData: map[string]interface{}{
						utils.Weight: 10.0,
					},
					RouteParameters: "param1",
				},
			},
		},
	}
	if err := rpS.V1GetRoutes(nil, reply); err == nil {
		t.Error(err)
	} else if err = rpS.V1GetRoutes(args, reply); err == nil {
		t.Error(err)
	}

	if err := rpS.V1GetRoutesList(args, &[]string{}); err == nil {
		t.Error(err)
	}

}
