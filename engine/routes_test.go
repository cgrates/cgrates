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

var (
	expTimeRoutes = time.Now().Add(time.Duration(20 * time.Minute))
	splService    *RouteService
	dmSPP         *DataManager
	sppTest       = RouteProfiles{
		&RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile1",
			FilterIDs: []string{"FLTR_SUPP_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     expTimeRoutes,
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "supplier1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RatingPlanIDs:   []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weight:          10.0,
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weight: 10,
		},
		&RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile2",
			FilterIDs: []string{"FLTR_SUPP_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     expTimeRoutes,
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "supplier2",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RatingPlanIDs:   []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weight:          20.0,
					RouteParameters: "param2",
				},
				{
					ID:              "supplier3",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RatingPlanIDs:   []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weight:          10.0,
					RouteParameters: "param3",
				},
				{
					ID:              "supplier1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RatingPlanIDs:   []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weight:          30.0,
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weight: 20.0,
		},
		&RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "RouteProfilePrefix",
			FilterIDs: []string{"FLTR_SUPP_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     expTimeRoutes,
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "supplier1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RatingPlanIDs:   []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weight:          10.0,
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weight: 10,
		},
	}
	argsGetRoutes = []*ArgsGetRoutes{
		{ //matching RouteProfile1
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "utils.CGREvent1",
				Event: map[string]interface{}{
					"Route":          "RouteProfile1",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "20.0",
				},
			},
		},
		{ //matching RouteProfile2
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "utils.CGREvent1",
				Event: map[string]interface{}{
					"Route":          "RouteProfile2",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					"PddInterval":    "1s",
					"Weight":         "20.0",
				},
			},
		},
		{ //matching RouteProfilePrefix
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "utils.CGREvent1",
				Event: map[string]interface{}{
					"Route": "RouteProfilePrefix",
				},
			},
		},
		{ //matching
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "CGR",
				Event: map[string]interface{}{
					"UsageInterval": "1s",
					"PddInterval":   "1s",
				},
			},
		},
	}
)

func TestRoutesSort(t *testing.T) {
	sprs := RouteProfiles{
		&RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile1",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting:           "",
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "supplier1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RatingPlanIDs:   []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weight:          10.0,
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weight: 10,
		},
		&RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile2",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting:           "",
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "supplier1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RatingPlanIDs:   []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weight:          20.0,
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weight: 20.0,
		},
	}
	eRouteProfile := RouteProfiles{
		&RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile2",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting:           "",
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "supplier1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RatingPlanIDs:   []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weight:          20.0,
					Blocker:         false,
					RouteParameters: "param1",
				},
			},
			Weight: 20.0,
		},
		&RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "supplierprofile1",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Sorting:           "",
			SortingParameters: []string{},
			Routes: []*Route{
				{
					ID:              "supplier1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RatingPlanIDs:   []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{},
					Weight:          10.0,
					Blocker:         false,
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

func TestRoutesPopulateRouteService(t *testing.T) {
	defaultCfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, defaultCfg.DataDbCfg().Items)
	dmSPP = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.RouteSCfg().StringIndexedFields = nil
	defaultCfg.RouteSCfg().PrefixIndexedFields = nil
	splService, err = NewRouteService(dmSPP, &FilterS{
		dm: dmSPP, cfg: defaultCfg}, defaultCfg, nil)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestRoutesAddFilters(t *testing.T) {
	fltrSupp1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_SUPP_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Route",
				Values:  []string{"RouteProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmSPP.SetFilter(fltrSupp1)
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
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmSPP.SetFilter(fltrSupp2)
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
	dmSPP.SetFilter(fltrSupp3)
}

func TestRoutesCache(t *testing.T) {
	for _, spp := range sppTest {
		if err = dmSPP.SetRouteProfile(spp, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each supplier profile from cache
	for _, spp := range sppTest {
		if tempSpp, err := dmSPP.GetRouteProfile(spp.Tenant,
			spp.ID, true, true, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(spp, tempSpp) {
			t.Errorf("Expecting: %+v, received: %+v", spp, tempSpp)
		}
	}
}

func TestRoutesmatchingRouteProfilesForEvent(t *testing.T) {
	sprf, err := splService.matchingRouteProfilesForEvent(argsGetRoutes[0].CGREvent, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[0], sprf[0])
	}

	sprf, err = splService.matchingRouteProfilesForEvent(argsGetRoutes[1].CGREvent, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[1], sprf[0])
	}

	sprf, err = splService.matchingRouteProfilesForEvent(argsGetRoutes[2].CGREvent, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[2], sprf[0])
	}
}

func TestRoutesSortedForEvent(t *testing.T) {
	eFirstRouteProfile := &SortedRoutes{
		ProfileID: "RouteProfile1",
		Sorting:   utils.MetaWeight,
		Count:     1,
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}
	sprf, err := splService.sortedRoutesForEvent(argsGetRoutes[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstRouteProfile, sprf)
	}

	eFirstRouteProfile = &SortedRoutes{
		ProfileID: "RouteProfile2",
		Sorting:   utils.MetaWeight,
		Count:     3,
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 30.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}

	sprf, err = splService.sortedRoutesForEvent(argsGetRoutes[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstRouteProfile, sprf)
	}

	eFirstRouteProfile = &SortedRoutes{
		ProfileID: "RouteProfilePrefix",
		Sorting:   utils.MetaWeight,
		Count:     1,
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}

	sprf, err = splService.sortedRoutesForEvent(argsGetRoutes[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstRouteProfile, sprf)
	}
}

func TestRoutesSortedForEventWithLimit(t *testing.T) {
	eFirstRouteProfile := &SortedRoutes{
		ProfileID: "RouteProfile2",
		Sorting:   utils.MetaWeight,
		Count:     2,
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 30.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				RouteParameters: "param2",
			},
		},
	}
	argsGetRoutes[1].Paginator = utils.Paginator{
		Limit: utils.IntPointer(2),
	}
	sprf, err := splService.sortedRoutesForEvent(argsGetRoutes[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v, received: %+v", eFirstRouteProfile, sprf)
	}
}

func TestRoutesSortedForEventWithOffset(t *testing.T) {
	eFirstRouteProfile := &SortedRoutes{
		ProfileID: "RouteProfile2",
		Sorting:   utils.MetaWeight,
		Count:     1,
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}
	argsGetRoutes[1].Paginator = utils.Paginator{
		Offset: utils.IntPointer(2),
	}
	sprf, err := splService.sortedRoutesForEvent(argsGetRoutes[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}
}

func TestRoutesSortedForEventWithLimitAndOffset(t *testing.T) {
	eFirstRouteProfile := &SortedRoutes{
		ProfileID: "RouteProfile2",
		Sorting:   utils.MetaWeight,
		Count:     1,
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				RouteParameters: "param2",
			},
		},
	}
	argsGetRoutes[1].Paginator = utils.Paginator{
		Limit:  utils.IntPointer(1),
		Offset: utils.IntPointer(1),
	}
	sprf, err := splService.sortedRoutesForEvent(argsGetRoutes[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eFirstRouteProfile, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", utils.ToJSON(eFirstRouteProfile), utils.ToJSON(sprf))
	}
}

func TestRoutesAsOptsGetRoutes(t *testing.T) {
	s := &ArgsGetRoutes{
		IgnoreErrors: true,
		MaxCost:      "10.0",
	}
	spl := &optsGetRoutes{
		ignoreErrors: true,
		maxCost:      10.0,
	}
	sprf, err := s.asOptsGetRoutes()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestRoutesAsOptsGetRoutesIgnoreErrors(t *testing.T) {
	s := &ArgsGetRoutes{
		IgnoreErrors: true,
	}
	spl := &optsGetRoutes{
		ignoreErrors: true,
	}
	sprf, err := s.asOptsGetRoutes()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestRoutesAsOptsGetRoutesMaxCost(t *testing.T) {
	s := &ArgsGetRoutes{
		MaxCost: "10.0",
	}
	spl := &optsGetRoutes{
		maxCost: 10.0,
	}
	sprf, err := s.asOptsGetRoutes()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(spl, sprf) {
		t.Errorf("Expecting: %+v,received: %+v", spl, sprf)
	}
}

func TestRoutesMatchWithIndexFalse(t *testing.T) {
	splService.cgrcfg.RouteSCfg().IndexedSelects = false
	sprf, err := splService.matchingRouteProfilesForEvent(argsGetRoutes[0].CGREvent, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[0], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[0], sprf[0])
	}

	sprf, err = splService.matchingRouteProfilesForEvent(argsGetRoutes[1].CGREvent, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[1], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[1], sprf[0])
	}

	sprf, err = splService.matchingRouteProfilesForEvent(argsGetRoutes[2].CGREvent, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(sppTest[2], sprf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", sppTest[2], sprf[0])
	}
}
