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

func TestRoutesCache(t *testing.T) {
	var expTimeRoutes = time.Now().Add(20 * time.Minute)
	expTimeStr := expTimeRoutes.Format(time.RFC3339)
	var dmSPP *DataManager
	sppTest := []*RouteProfile{
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfile1",
			FilterIDs:         []string{"FLTR_RPP_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfile2",
			FilterIDs:         []string{"FLTR_SUPP_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route2",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 20}},
					RouteParameters: "param2",
				},
				{
					ID:              "route3",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param3",
				},
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 30}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 20}},
		},
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfilePrefix",
			FilterIDs:         []string{"FLTR_SUPP_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
	}

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSPP = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.RouteSCfg().StringIndexedFields = nil
	defaultCfg.RouteSCfg().PrefixIndexedFields = nil
	fltrSupp1 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp1, true)
	fltrSupp2 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp2, true)
	fltrSupp3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Route",
				Values:  []string{"RouteProfilePrefix"},
			},
		},
	}
	dmSPP.SetFilter(context.Background(), fltrSupp3, true)

	for _, spp := range sppTest {
		if err = dmSPP.SetRouteProfile(context.Background(), spp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each route profile from cache
	for _, spp := range sppTest {
		if tempSpp, err := dmSPP.GetRouteProfile(context.Background(), spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}
}

func TestRoutesmatchingRouteProfilesForEvent(t *testing.T) {
	var expTimeRoutes = time.Now().Add(20 * time.Minute)
	expTimeStr := expTimeRoutes.Format(time.RFC3339)
	var routeService *RouteService
	var dmSPP *DataManager
	var sppTest = RouteProfilesWithWeight{
		&RouteProfileWithWeight{
			RouteProfile: &RouteProfile{
				Tenant:            "cgrates.org",
				ID:                "RouteProfile1",
				FilterIDs:         []string{"FLTR_RPP_1", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*Route{
					{
						ID:              "route1",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 10}},
						Blocker:         false,
						RouteParameters: "param1",
					},
				},
				Weights: utils.DynamicWeights{{Weight: 10}},
			},
			Weight: 10,
		},
		&RouteProfileWithWeight{
			RouteProfile: &RouteProfile{
				Tenant:            "cgrates.org",
				ID:                "RouteProfile2",
				FilterIDs:         []string{"FLTR_SUPP_2", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*Route{
					{
						ID:              "route2",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 20}},
						RouteParameters: "param2",
					},
					{
						ID:              "route3",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 10}},
						RouteParameters: "param3",
					},
					{
						ID:              "route1",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 30}},
						Blocker:         false,
						RouteParameters: "param1",
					},
				},
				Weights: utils.DynamicWeights{{Weight: 20}},
			},
			Weight: 20,
		},
		&RouteProfileWithWeight{
			RouteProfile: &RouteProfile{
				Tenant:            "cgrates.org",
				ID:                "RouteProfilePrefix",
				FilterIDs:         []string{"FLTR_SUPP_3", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*Route{
					{
						ID:              "route1",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 10}},
						Blocker:         false,
						RouteParameters: "param1",
					},
				},
				Weights: utils.DynamicWeights{{Weight: 10}},
			},
			Weight: 10,
		},
	}
	args := []*utils.CGREvent{
		//matching RouteProfile1
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route":          "RouteProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route":          "RouteProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route": "RouteProfilePrefix",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "CGR",
			Event: map[string]interface{}{
				"UsageInterval": "1s",
				"PddInterval":   "1s",
			},
		},
	}

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSPP = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.RouteSCfg().StringIndexedFields = nil
	defaultCfg.RouteSCfg().PrefixIndexedFields = nil
	routeService = NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: defaultCfg}, defaultCfg, nil)

	fltrSupp1 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp1, true)
	fltrSupp2 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp2, true)
	fltrSupp3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Route",
				Values:  []string{"RouteProfilePrefix"},
			},
		},
	}
	dmSPP.SetFilter(context.Background(), fltrSupp3, true)
	for _, spp := range sppTest {
		if err = dmSPP.SetRouteProfile(context.Background(), spp.RouteProfile, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each route profile from cache
	for _, spp := range sppTest {
		if tempSpp, err := dmSPP.GetRouteProfile(context.Background(), spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp.RouteProfile, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}
	sprf, err := routeService.matchingRouteProfilesForEvent(context.Background(), args[0].Tenant, args[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[0], sprf[0])
	}

	sprf, err = routeService.matchingRouteProfilesForEvent(context.Background(), args[1].Tenant, args[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[1], sprf[0])
	}

	sprf, err = routeService.matchingRouteProfilesForEvent(context.Background(), args[2].Tenant, args[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[2], sprf[0])
	}
}

func TestRoutesSortedForEvent(t *testing.T) {
	var expTimeRoutes = time.Now().Add(20 * time.Minute)
	expTimeStr := expTimeRoutes.Format(time.RFC3339)
	var routeService *RouteService
	var dmSPP *DataManager
	var sppTest = []*RouteProfile{
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfile1",
			FilterIDs:         []string{"FLTR_RPP_1", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfile2",
			FilterIDs:         []string{"FLTR_SUPP_2", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route2",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 20}},
					RouteParameters: "param2",
				},
				{
					ID:              "route3",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param3",
				},
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 30}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 20}},
		},
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfilePrefix",
			FilterIDs:         []string{"FLTR_SUPP_3", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
	}
	args := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route":          "RouteProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route":          "RouteProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route": "RouteProfilePrefix",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "CGR",
			Event: map[string]interface{}{
				"UsageInterval": "1s",
				"PddInterval":   "1s",
			},
		},
	}

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSPP = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.RouteSCfg().StringIndexedFields = nil
	defaultCfg.RouteSCfg().PrefixIndexedFields = nil
	routeService = NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: defaultCfg}, defaultCfg, nil)

	fltrSupp1 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp1, true)
	fltrSupp2 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp2, true)
	fltrSupp3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Route",
				Values:  []string{"RouteProfilePrefix"},
			},
		},
	}
	dmSPP.SetFilter(context.Background(), fltrSupp3, true)
	for _, spp := range sppTest {
		if err = dmSPP.SetRouteProfile(context.Background(), spp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each route profile from cache
	for _, spp := range sppTest {
		if tempSpp, err := dmSPP.GetRouteProfile(context.Background(), spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}
	eFirstRouteProfile := SortedRoutesList{&SortedRoutes{
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
	sprf, err := routeService.sortedRoutesForEvent(context.Background(), args[0].Tenant, args[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstRouteProfile, sprf)
	}

	eFirstRouteProfile = SortedRoutesList{&SortedRoutes{
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

	sprf, err = routeService.sortedRoutesForEvent(context.Background(), args[1].Tenant, args[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstRouteProfile, sprf)
	}

	eFirstRouteProfile = SortedRoutesList{&SortedRoutes{
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

	sprf, err = routeService.sortedRoutesForEvent(context.Background(), args[2].Tenant, args[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstRouteProfile, sprf)
	}
}

func TestRoutesSortedForEventWithLimit(t *testing.T) {
	expTimeRoutes := time.Now().Add(20 * time.Minute)
	expTimeStr := expTimeRoutes.Format(time.RFC3339)
	var routeService *RouteService
	sppTest := []*RouteProfile{
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfile1",
			FilterIDs:         []string{"FLTR_RPP_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfile2",
			FilterIDs:         []string{"FLTR_SUPP_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route2",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 20}},
					RouteParameters: "param2",
				},
				{
					ID:              "route3",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param3",
				},
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 30}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 20}},
		},
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfilePrefix",
			FilterIDs:         []string{"FLTR_SUPP_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
	}
	args := []*utils.CGREvent{
		{
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
		{
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
		{
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

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.RouteSCfg().StringIndexedFields = nil
	defaultCfg.RouteSCfg().PrefixIndexedFields = nil
	routeService = NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: defaultCfg}, defaultCfg, nil)

	fltrSupp1 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp1, true)
	fltrSupp2 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp2, true)
	fltrSupp3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Route",
				Values:  []string{"RouteProfilePrefix"},
			},
		},
	}
	dmSPP.SetFilter(context.Background(), fltrSupp3, true)

	for _, spp := range sppTest {
		if err = dmSPP.SetRouteProfile(context.Background(), spp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each route profile from cache
	for _, spp := range sppTest {
		if tempSpp, err := dmSPP.GetRouteProfile(context.Background(), spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}

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
	args[1].APIOpts[utils.OptsRoutesLimit] = 2
	sprf, err := routeService.sortedRoutesForEvent(context.Background(), args[1].Tenant, args[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstRouteProfile, sprf)
	}
}

func TestRoutesSortedForEventWithOffset(t *testing.T) {
	expTimeRoutes := time.Now().Add(20 * time.Minute)
	expTimeStr := expTimeRoutes.Format(time.RFC3339)
	var routeService *RouteService
	sppTest := []*RouteProfile{
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfile1",
			FilterIDs:         []string{"FLTR_RPP_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfile2",
			FilterIDs:         []string{"FLTR_SUPP_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route2",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 20}},
					RouteParameters: "param2",
				},
				{
					ID:              "route3",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param3",
				},
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 30}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 20}},
		},
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfilePrefix",
			FilterIDs:         []string{"FLTR_SUPP_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
	}
	args := []*utils.CGREvent{
		{
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
		{
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
		{
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

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.RouteSCfg().StringIndexedFields = nil
	defaultCfg.RouteSCfg().PrefixIndexedFields = nil
	routeService = NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: defaultCfg}, defaultCfg, nil)

	fltrSupp1 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp1, true)
	fltrSupp2 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp2, true)
	fltrSupp3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Route",
				Values:  []string{"RouteProfilePrefix"},
			},
		},
	}
	dmSPP.SetFilter(context.Background(), fltrSupp3, true)

	for _, spp := range sppTest {
		if err = dmSPP.SetRouteProfile(context.Background(), spp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each route profile from cache
	for _, spp := range sppTest {
		if tempSpp, err := dmSPP.GetRouteProfile(context.Background(), spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}

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
	args[1].APIOpts[utils.OptsRoutesOffset] = 2
	sprf, err := routeService.sortedRoutesForEvent(context.Background(), args[1].Tenant, args[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}
}

func TestRoutesSortedForEventWithLimitAndOffset(t *testing.T) {
	expTimeRoutes := time.Now().Add(20 * time.Minute)
	expTimeStr := expTimeRoutes.Format(time.RFC3339)
	var routeService *RouteService
	sppTest := []*RouteProfile{
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfile1",
			FilterIDs:         []string{"FLTR_RPP_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfile2",
			FilterIDs:         []string{"FLTR_SUPP_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route2",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 20}},
					RouteParameters: "param2",
				},
				{
					ID:              "route3",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					RouteParameters: "param3",
				},
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 30}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 20}},
		},
		&RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "RouteProfilePrefix",
			FilterIDs:         []string{"FLTR_SUPP_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RateProfileIDs:  []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weights:         utils.DynamicWeights{{Weight: 10}},
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{{Weight: 10}},
		},
	}
	args := []*utils.CGREvent{
		{
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
		{
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
		{
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

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.RouteSCfg().StringIndexedFields = nil
	defaultCfg.RouteSCfg().PrefixIndexedFields = nil
	routeService = NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: defaultCfg}, defaultCfg, nil)

	fltrSupp1 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp1, true)
	fltrSupp2 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp2, true)
	fltrSupp3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Route",
				Values:  []string{"RouteProfilePrefix"},
			},
		},
	}
	dmSPP.SetFilter(context.Background(), fltrSupp3, true)

	for _, spp := range sppTest {
		if err = dmSPP.SetRouteProfile(context.Background(), spp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each route profile from cache
	for _, spp := range sppTest {
		if tempSpp, err := dmSPP.GetRouteProfile(context.Background(), spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}

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
	args[1].APIOpts[utils.OptsRoutesLimit] = 1
	args[1].APIOpts[utils.OptsRoutesOffset] = 1
	sprf, err := routeService.sortedRoutesForEvent(context.Background(), args[1].Tenant, args[1])
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
	config.CgrConfig().RouteSCfg().Opts.IgnoreErrors = []*utils.DynamicBoolOpt{
		{
			Value: true,
		},
	}
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

func TestRoutesNewOptsGetRoutesMaxCost(t *testing.T) {
	expTimeRoutes := time.Now().Add(20 * time.Minute)
	expTimeStr := expTimeRoutes.Format(time.RFC3339)
	var routeService *RouteService
	sppTest := RouteProfilesWithWeight{
		&RouteProfileWithWeight{
			RouteProfile: &RouteProfile{
				Tenant:            "cgrates.org",
				ID:                "RouteProfile1",
				FilterIDs:         []string{"FLTR_RPP_1", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*Route{
					{
						ID:              "route1",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 10}},
						Blocker:         false,
						RouteParameters: "param1",
					},
				},
				Weights: utils.DynamicWeights{{Weight: 10}},
			},
			Weight: 10,
		},
		&RouteProfileWithWeight{
			RouteProfile: &RouteProfile{
				Tenant:            "cgrates.org",
				ID:                "RouteProfile2",
				FilterIDs:         []string{"FLTR_SUPP_2", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*Route{
					{
						ID:              "route2",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 20}},
						RouteParameters: "param2",
					},
					{
						ID:              "route3",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 10}},
						RouteParameters: "param3",
					},
					{
						ID:              "route1",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 30}},
						Blocker:         false,
						RouteParameters: "param1",
					},
				},
				Weights: utils.DynamicWeights{{Weight: 20}},
			},
			Weight: 20,
		},
		&RouteProfileWithWeight{
			RouteProfile: &RouteProfile{
				Tenant:            "cgrates.org",
				ID:                "RouteProfilePrefix",
				FilterIDs:         []string{"FLTR_SUPP_3", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*Route{
					{
						ID:              "route1",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 10}},
						Blocker:         false,
						RouteParameters: "param1",
					},
				},
				Weights: utils.DynamicWeights{{Weight: 10}},
			},
			Weight: 10,
		},
	}
	args := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route":          "RouteProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route":          "RouteProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route": "RouteProfilePrefix",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "CGR",
			Event: map[string]interface{}{
				"UsageInterval": "1s",
				"PddInterval":   "1s",
			},
		},
	}

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.RouteSCfg().StringIndexedFields = nil
	defaultCfg.RouteSCfg().PrefixIndexedFields = nil
	routeService = NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: defaultCfg}, defaultCfg, nil)

	fltrSupp1 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp1, true)
	fltrSupp2 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp2, true)
	fltrSupp3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Route",
				Values:  []string{"RouteProfilePrefix"},
			},
		},
	}
	dmSPP.SetFilter(context.Background(), fltrSupp3, true)

	for _, spp := range sppTest {
		if err = dmSPP.SetRouteProfile(context.Background(), spp.RouteProfile, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each route profile from cache
	for _, spp := range sppTest {
		if tempSpp, err := dmSPP.GetRouteProfile(context.Background(), spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp.RouteProfile, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}

	routeService.cfg.RouteSCfg().IndexedSelects = false
	sprf, err := routeService.matchingRouteProfilesForEvent(context.Background(), args[0].Tenant, args[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[0], sprf[0])
	}

	sprf, err = routeService.matchingRouteProfilesForEvent(context.Background(), args[1].Tenant, args[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[1], sprf[0])
	}

	sprf, err = routeService.matchingRouteProfilesForEvent(context.Background(), args[2].Tenant, args[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[2], sprf[0])
	}
}

func TestRoutesMatchWithIndexFalse(t *testing.T) {
	expTimeRoutes := time.Now().Add(20 * time.Minute)
	expTimeStr := expTimeRoutes.Format(time.RFC3339)
	var routeService *RouteService
	sppTest := RouteProfilesWithWeight{
		&RouteProfileWithWeight{
			RouteProfile: &RouteProfile{
				Tenant:            "cgrates.org",
				ID:                "RouteProfile1",
				FilterIDs:         []string{"FLTR_RPP_1", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*Route{
					{
						ID:              "route1",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 10}},
						Blocker:         false,
						RouteParameters: "param1",
					},
				},
				Weights: utils.DynamicWeights{{Weight: 10}},
			},
			Weight: 10,
		},
		&RouteProfileWithWeight{
			RouteProfile: &RouteProfile{
				Tenant:            "cgrates.org",
				ID:                "RouteProfile2",
				FilterIDs:         []string{"FLTR_SUPP_2", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*Route{
					{
						ID:              "route2",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 20}},
						RouteParameters: "param2",
					},
					{
						ID:              "route3",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 10}},
						RouteParameters: "param3",
					},
					{
						ID:              "route1",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 30}},
						Blocker:         false,
						RouteParameters: "param1",
					},
				},
				Weights: utils.DynamicWeights{{Weight: 20}},
			},
			Weight: 20,
		},
		&RouteProfileWithWeight{
			RouteProfile: &RouteProfile{
				Tenant:            "cgrates.org",
				ID:                "RouteProfilePrefix",
				FilterIDs:         []string{"FLTR_SUPP_3", "*ai:*now:2014-07-14T14:25:00Z|" + expTimeStr},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*Route{
					{
						ID:              "route1",
						FilterIDs:       []string{},
						AccountIDs:      []string{},
						RateProfileIDs:  []string{},
						ResourceIDs:     []string{},
						StatIDs:         []string{},
						Weights:         utils.DynamicWeights{{Weight: 10}},
						Blocker:         false,
						RouteParameters: "param1",
					},
				},
				Weights: utils.DynamicWeights{{Weight: 10}},
			},
			Weight: 10,
		},
	}
	args := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route":          "RouteProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route":          "RouteProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "utils.CGREvent1",
			Event: map[string]interface{}{
				"Route": "RouteProfilePrefix",
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "CGR",
			Event: map[string]interface{}{
				"UsageInterval": "1s",
				"PddInterval":   "1s",
			},
		},
	}

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.RouteSCfg().StringIndexedFields = nil
	defaultCfg.RouteSCfg().PrefixIndexedFields = nil
	routeService = NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: defaultCfg}, defaultCfg, nil)

	fltrSupp1 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp1, true)
	fltrSupp2 := &Filter{
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
	}
	dmSPP.SetFilter(context.Background(), fltrSupp2, true)
	fltrSupp3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Route",
				Values:  []string{"RouteProfilePrefix"},
			},
		},
	}
	dmSPP.SetFilter(context.Background(), fltrSupp3, true)

	for _, spp := range sppTest {
		if err = dmSPP.SetRouteProfile(context.Background(), spp.RouteProfile, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each route profile from cache
	for _, spp := range sppTest {
		if tempSpp, err := dmSPP.GetRouteProfile(context.Background(), spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp.RouteProfile, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}

	routeService.cfg.RouteSCfg().IndexedSelects = false
	sprf, err := routeService.matchingRouteProfilesForEvent(context.Background(), args[0].Tenant, args[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[0], sprf[0])
	}

	sprf, err = routeService.matchingRouteProfilesForEvent(context.Background(), args[1].Tenant, args[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[1], sprf[0])
	}

	sprf, err = routeService.matchingRouteProfilesForEvent(context.Background(), args[2].Tenant, args[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[2], sprf[0])
	}
}

func TestRoutesSortedForEventWithLimitAndOffset2(t *testing.T) {
	var routeService *RouteService
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

	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmSPP := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.RouteSCfg().StringIndexedFields = nil
	defaultCfg.RouteSCfg().PrefixIndexedFields = nil
	routeService = NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: defaultCfg}, defaultCfg, nil)

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
