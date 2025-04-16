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
package utils

import (
	"reflect"
	"testing"
	"time"
)

var (
	testRoutesPrfs = []*RouteProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile1",
			FilterIDs: []string{"FLTR_RPP_1"},
			Sorting:   MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weights:         DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights:  DynamicWeights{{Weight: 10}},
			Blockers: DynamicBlockers{{Blocker: true}},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfile2",
			FilterIDs: []string{"FLTR_SUPP_2"},
			Sorting:   MetaWeight,
			Routes: []*Route{
				{
					ID:              "route2",
					Weights:         DynamicWeights{{Weight: 20}},
					RouteParameters: "param2",
				},
				{
					ID:              "route3",
					Weights:         DynamicWeights{{Weight: 10}},
					RouteParameters: "param3",
				},
				{
					ID:              "route1",
					Weights:         DynamicWeights{{Weight: 30}},
					RouteParameters: "param1",
				},
			},
			Weights: DynamicWeights{{Weight: 20}},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "RouteProfilePrefix",
			FilterIDs: []string{"FLTR_SUPP_3"},
			Sorting:   MetaWeight,
			Routes: []*Route{
				{
					ID:              "route1",
					Weights:         DynamicWeights{{Weight: 10}},
					RouteParameters: "param1",
				},
			},
			Weights: DynamicWeights{{Weight: 10}},
		},
	}
	testRoutesArgs = []*CGREvent{
		{ //matching RouteProfile1
			Tenant: "cgrates.org",
			ID:     "CGREvent1",
			Event: map[string]any{
				"Route":         "RouteProfile1",
				AnswerTime:      time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval": "1s",
				"PddInterval":   "1s",
				Weight:          "20.0",
			},
			APIOpts: map[string]any{
				OptsRoutesProfilesCount: 1,
			},
		},
		{ //matching RouteProfile2
			Tenant: "cgrates.org",
			ID:     "CGREvent1",
			Event: map[string]any{
				"Route":         "RouteProfile2",
				AnswerTime:      time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval": "1s",
				"PddInterval":   "1s",
				Weight:          "20.0",
			},
			APIOpts: map[string]any{
				OptsRoutesProfilesCount: 1,
			},
		},
		{ //matching RouteProfilePrefix
			Tenant: "cgrates.org",
			ID:     "CGREvent1",
			Event: map[string]any{
				"Route": "RouteProfilePrefix",
			},
			APIOpts: map[string]any{
				OptsRoutesProfilesCount: 1,
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
				OptsRoutesProfilesCount: 1,
			},
		},
	}
)

func TestRouteProfileSet(t *testing.T) {
	rp := RouteProfile{}
	exp := RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:   DynamicWeights{{}},
		Blockers: DynamicBlockers{
			{Blocker: false},
		},
		Sorting:           MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        DynamicWeights{{}},
			Blockers: DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}
	if err := rp.Set([]string{}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{"", ""}, "", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{"NotAField"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{"NotAField", "1"}, ":", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := rp.Set([]string{Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Weights}, ";0", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Blockers}, ";false", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Sorting}, MetaQOS, false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{SortingParameters}, "param", false); err != nil {
		t.Error(err)
	}

	if err := rp.Set([]string{Routes, ID}, "RT1", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Routes, FilterIDs}, "fltr1", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Routes, AccountIDs}, "acc1", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Routes, RateProfileIDs}, "rp1", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Routes, ResourceIDs}, "res1", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Routes, StatIDs}, "stat1", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Routes, Weights}, ";0", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Routes, Blockers}, ";true", false); err != nil {
		t.Error(err)
	}
	if err := rp.Set([]string{Routes, RouteParameters}, "params", false); err != nil {
		t.Error(err)
	}

	if err := rp.Set([]string{SortingParameters, "wrong"}, "param", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := rp.Set([]string{Routes, "wrong"}, "param", false); err != ErrWrongPath {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, rp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(rp))
	}
}

func TestRouteProfileAsInterface(t *testing.T) {
	rp := RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           DynamicWeights{{}},
		Blockers:          DynamicBlockers{{Blocker: false}},
		Sorting:           MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        DynamicWeights{{}},
			Blockers: DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}
	if _, err := rp.FieldAsInterface(nil); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{"field"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{"field", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.FieldAsInterface([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ";0"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := ";false"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{SortingParameters}); err != nil {
		t.Fatal(err)
	} else if exp := rp.SortingParameters; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{SortingParameters + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.SortingParameters[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Sorting}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Sorting; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if _, err := rp.FieldAsInterface([]string{Routes + "[4]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{Routes + "[0]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{Routes + "[0]", "", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", ID}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ";0"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := ";true"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", RouteParameters}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].RouteParameters; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", AccountIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].AccountIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", AccountIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].AccountIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", RateProfileIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].RateProfileIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", RateProfileIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].RateProfileIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", ResourceIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].ResourceIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", ResourceIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].ResourceIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", StatIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].StatIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Routes + "[0]", StatIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Routes[0].StatIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := rp.FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.FieldAsString([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := rp.String(), ToJSON(rp); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := rp.Routes[0].FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.Routes[0].FieldAsString([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := "RT1"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := rp.Routes[0].String(), ToJSON(rp.Routes[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
}

func TestRouteProfileMerge(t *testing.T) {
	dp := &RouteProfile{}
	exp := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           DynamicWeights{{}},
		Sorting:           MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        DynamicWeights{{}},
			Blockers: DynamicBlockers{
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
		Weights:           DynamicWeights{{}},
		Sorting:           MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        DynamicWeights{{}},
			Blockers: DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}); !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(dp))
	}
}

func TestRouteMerge(t *testing.T) {

	route := &Route{}

	routeV2 := &Route{
		ID:              "RouteId",
		RouteParameters: "RouteParam",
		Weights:         DynamicWeights{{Weight: 10}},
		Blockers:        DynamicBlockers{{Blocker: false}},
		FilterIDs:       []string{"FltrId"},
		AccountIDs:      []string{"AccId"},
		RateProfileIDs:  []string{"RateProfileId"},
		ResourceIDs:     []string{"ResourceId"},
		StatIDs:         []string{"StatId"},
	}
	exp := routeV2

	route.Merge(routeV2)
	if !reflect.DeepEqual(route, exp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(route))
	}
}

func TestRouteProfileCompileCacheParametersErrParse(t *testing.T) {
	rp := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           DynamicWeights{{}},
		Sorting:           MetaLoad,
		SortingParameters: []string{"sort:param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        DynamicWeights{{}},
			Blockers: DynamicBlockers{
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
	rp := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           DynamicWeights{{}},
		Sorting:           MetaLoad,
		SortingParameters: []string{"param:1"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        DynamicWeights{{}},
			Blockers: DynamicBlockers{
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
	rp := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           DynamicWeights{{}},
		Sorting:           MetaLoad,
		SortingParameters: []string{"*default:1"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        DynamicWeights{{}},
			Blockers: DynamicBlockers{
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
	rp := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ID",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:           DynamicWeights{{}},
		Sorting:           MetaLoad,
		SortingParameters: []string{"RT1:1"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        DynamicWeights{{}},
			Blockers: DynamicBlockers{
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
		Weights:           DynamicWeights{{}},
		Sorting:           MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        DynamicWeights{{}},
			Blockers: DynamicBlockers{
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
		Weights:           DynamicWeights{{}},
		Sorting:           MetaQOS,
		SortingParameters: []string{"param"},
		Routes: []*Route{{
			ID:             "RT1",
			FilterIDs:      []string{"fltr1"},
			AccountIDs:     []string{"acc1"},
			RateProfileIDs: []string{"rp1"},
			ResourceIDs:    []string{"res1"},
			StatIDs:        []string{"stat1"},
			Weights:        DynamicWeights{{}},
			Blockers: DynamicBlockers{
				{
					Blocker: true,
				},
			},
			RouteParameters: "params",
		}},
	}); !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(dp))
	}
}
