// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"reflect"
	"sort"
	"strconv"
	"testing"
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

func TestSortedRoutesListDigest(t *testing.T) {
	sSpls := SortedRoutesList{{
		ProfileID: "ROUTE_WEIGHT_2",
		Sorting:   MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
			},
			{
				RouteID: "route2",
			},
			{
				RouteID: "route5",
			},
			{
				RouteID: "route2",
			},
			{
				RouteID: "route3",
			},
			{
				RouteID: "route0",
			},
			{
				RouteID: "route1",
			},
		},
	}}
	// digest will join unique RouteID from SortedRoutes
	expected := "route1,route2,route5,route3,route0"
	if rcv := sSpls.Digest(); rcv != expected {
		t.Errorf("Expected %s received %s", expected, rcv)
	}
}

func TestLibSuppliersSortCost(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.05),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.05,
					Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}
	sSpls.SortLeastCost()
	eOrderedSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.05),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.05,
					Weight: 10.0,
				},
				RouteParameters: "param3",
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			ToJSON(eOrderedSpls), ToJSON(sSpls))
	}
}

func TestLibRoutesSortWeight(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]any{
					Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(10.5),
				},
				SortingData: map[string]any{
					Weight: 10.5,
				},
				RouteParameters: "param3",
			},
		},
	}
	sSpls.SortWeight()
	eOrderedSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]any{
					Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(10.5),
				},
				SortingData: map[string]any{
					Weight: 10.5,
				},
				RouteParameters: "param3",
			},
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Weight: 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			ToJSON(eOrderedSpls), ToJSON(sSpls))
	}
}

func TestSortedRoutesDigest(t *testing.T) {
	eSpls := SortedRoutes{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]any{
					Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Weight: 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}
	exp := "route2:param2,route1:param1"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedRoutesDigest2(t *testing.T) {
	eSpls := SortedRoutes{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(30.0),
				},
				SortingData: map[string]any{
					Weight: 30.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]any{
					Weight: 20.0,
				},
				RouteParameters: "param2",
			},
		},
	}
	exp := "route1:param1,route2:param2"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedRoutesDigest3(t *testing.T) {
	eSpls := SortedRoutes{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   MetaWeight,
		Routes:    []*SortedRoute{},
	}
	exp := ""
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestLibRoutesSortHighestCost(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(15.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 15.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.2),
					Weight: NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]any{
					Cost:   0.2,
					Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.05),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.05,
					Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}
	sSpls.SortHighestCost()
	eOrderedSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.2),
					Weight: NewDecimalFromFloat64(20.0),
				},
				SortingData: map[string]any{
					Cost:   0.2,
					Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(15.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 15.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.05),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.05,
					Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			ToJSON(eOrderedSpls), ToJSON(sSpls))
	}
}

// sort based on *acd and *tcd
func TestLibRoutesSortQOS(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				//the average value for route1 for *acd is 0.5 , *tcd  1.1
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Cost:    NewDecimalFromFloat64(0.5),
					Weight:  NewDecimalFromFloat64(10.0),
					MetaACD: NewDecimalFromFloat64(0.5),
					MetaTCD: NewDecimalFromFloat64(1.1),
				},
				SortingData: map[string]any{
					Cost:    0.5,
					Weight:  10.0,
					MetaACD: 0.5,
					MetaTCD: 1.1,
				},
			},
			{
				//the average value for route2 for *acd is 0.5 , *tcd 4.1
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Cost:    NewDecimalFromFloat64(0.1),
					Weight:  NewDecimalFromFloat64(15.0),
					MetaACD: NewDecimalFromFloat64(0.5),
					MetaTCD: NewDecimalFromFloat64(4.1),
				},
				SortingData: map[string]any{
					Cost:    0.1,
					Weight:  15.0,
					MetaACD: 0.5,
					MetaTCD: 4.1,
				},
			},
			{
				//the average value for route3 for *acd is 0.4 , *tcd 5.1
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Cost:    NewDecimalFromFloat64(1.1),
					Weight:  NewDecimalFromFloat64(17.8),
					MetaACD: NewDecimalFromFloat64(0.4),
					MetaTCD: NewDecimalFromFloat64(5.1),
				},
				SortingData: map[string]any{
					Cost:    1.1,
					Weight:  17.8,
					MetaACD: 0.4,
					MetaTCD: 5.1,
				},
			},
		},
	}

	//sort base on *acd and *tcd
	sSpls.SortQOS([]string{MetaACD, MetaTCD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route2", "route1", "route3"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

// sort based on *acd and *tcd
func TestLibRoutesSortQOS2(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				//the average value for route1 for *acd is 0.5 , *tcd  1.1
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(10.0),
					MetaACD: NewDecimalFromFloat64(0.5),
					MetaTCD: NewDecimalFromFloat64(1.1),
				},
				SortingData: map[string]any{
					Weight:  10.0,
					MetaACD: 0.5,
					MetaTCD: 1.1,
				},
			},
			{
				//the worst value for route1 for *acd is 0.5 , *tcd  1.1
				//route1 and route2 have the same value for *acd and *tcd
				//will be sorted based on weight
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(17.0),
					MetaACD: NewDecimalFromFloat64(0.5),
					MetaTCD: NewDecimalFromFloat64(1.1),
				},
				SortingData: map[string]any{
					Weight:  17.0,
					MetaACD: 0.5,
					MetaTCD: 1.1,
				},
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Cost:    NewDecimalFromFloat64(0.5),
					Weight:  NewDecimalFromFloat64(10.0),
					MetaACD: NewDecimalFromFloat64(0.7),
					MetaTCD: NewDecimalFromFloat64(1.1),
				},
				SortingData: map[string]any{
					Cost:    0.5,
					Weight:  10.0,
					MetaACD: 0.7,
					MetaTCD: 1.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{MetaACD, MetaTCD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route3", "route2", "route1"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

// sort based on *pdd
func TestLibRoutesSortQOS3(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				//the worst value for route1 for *pdd is 0.7 , *tcd  1.1
				//route1 and route3 have the same value for *pdd
				//will be sorted based on weight
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(15.0),
					MetaPDD: NewDecimalFromFloat64(0.7),
					MetaTCD: NewDecimalFromFloat64(1.1),
				},
				SortingData: map[string]any{
					Weight:  15.0,
					MetaPDD: 0.7,
					MetaTCD: 1.1,
				},
			},
			{
				//the worst value for route2 for *pdd is 1.2, *tcd  1.1
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(10.0),
					MetaPDD: NewDecimalFromFloat64(1.2),
					MetaTCD: NewDecimalFromFloat64(1.1),
				},
				SortingData: map[string]any{
					Weight:  10.0,
					MetaPDD: 1.2,
					MetaTCD: 1.1,
				},
			},
			{
				//the worst value for route3 for *pdd is 0.7, *tcd  10.1
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(10.0),
					MetaPDD: NewDecimalFromFloat64(0.7),
					MetaTCD: NewDecimalFromFloat64(10.1),
				},
				SortingData: map[string]any{
					Weight:  10.0,
					MetaPDD: 0.7,
					MetaTCD: 10.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{MetaPDD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route1", "route3", "route2"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS4(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					MetaACD: NewDecimalFromFloat64(0.2),
					MetaTCD: NewDecimalFromFloat64(15.0),
					MetaASR: NewDecimalFromFloat64(1.2),
				},
				SortingData: map[string]any{
					MetaACD: 0.2,
					MetaTCD: 15.0,
					MetaASR: 1.2,
				},
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					MetaACD: NewDecimalFromFloat64(0.2),
					MetaTCD: NewDecimalFromFloat64(20.0),
					MetaASR: NewDecimalFromFloat64(-1.0),
				},
				SortingData: map[string]any{
					MetaACD: 0.2,
					MetaTCD: 20.0,
					MetaASR: -1.0,
				},
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					MetaACD: NewDecimalFromFloat64(0.1),
					MetaTCD: NewDecimalFromFloat64(10.0),
					MetaASR: NewDecimalFromFloat64(1.2),
				},
				SortingData: map[string]any{
					MetaACD: 0.1,
					MetaTCD: 10.0,
					MetaASR: 1.2,
				},
			},
		},
	}
	sSpls.SortQOS([]string{MetaASR, MetaACD, MetaTCD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route1", "route3", "route2"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS5(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					MetaACD: NewDecimalFromFloat64(0.2),
					MetaTCD: NewDecimalFromFloat64(15.0),
					MetaASR: NewDecimalFromFloat64(-1.0),
					MetaTCC: NewDecimalFromFloat64(10.1),
				},
				SortingData: map[string]any{
					MetaACD: 0.2,
					MetaTCD: 15.0,
					MetaASR: -1.0,
					MetaTCC: 10.1,
				},
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					MetaACD: NewDecimalFromFloat64(0.2),
					MetaTCD: NewDecimalFromFloat64(20.0),
					MetaASR: NewDecimalFromFloat64(1.2),
					MetaTCC: NewDecimalFromFloat64(10.1),
				},
				SortingData: map[string]any{
					MetaACD: 0.2,
					MetaTCD: 20.0,
					MetaASR: 1.2,
					MetaTCC: 10.1,
				},
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					MetaACD: NewDecimalFromFloat64(0.1),
					MetaTCD: NewDecimalFromFloat64(10.0),
					MetaASR: NewDecimalFromFloat64(1.2),
					MetaTCC: NewDecimalFromFloat64(10.1),
				},
				SortingData: map[string]any{
					MetaACD: 0.1,
					MetaTCD: 10.0,
					MetaASR: 1.2,
					MetaTCC: 10.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{MetaTCC, MetaASR, MetaACD, MetaTCD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route2", "route3", "route1"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS6(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(15.0),
					MetaACD: NewDecimalFromFloat64(0.2),
				},
				SortingData: map[string]any{
					Weight:  15.0,
					MetaACD: 0.2,
				},
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(25.0),
					MetaACD: NewDecimalFromFloat64(0.2),
				},
				SortingData: map[string]any{
					Weight:  25.0,
					MetaACD: 0.2,
				},
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(20.0),
					MetaACD: NewDecimalFromFloat64(0.1),
				},
				SortingData: map[string]any{
					Weight:  20.0,
					MetaACD: 0.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{MetaACD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route2", "route1", "route3"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS7(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(15.0),
					MetaACD: NewDecimalFromFloat64(-1.0),
				},
				SortingData: map[string]any{
					Weight:  15.0,
					MetaACD: -1.0,
				},
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(25.0),
					MetaACD: NewDecimalFromFloat64(-1.0),
				},
				SortingData: map[string]any{
					Weight:  25.0,
					MetaACD: -1.0,
				},
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(20.0),
					MetaACD: NewDecimalFromFloat64(-1.0),
				},
				SortingData: map[string]any{
					Weight:  20.0,
					MetaACD: -1.0,
				},
			},
		},
	}
	sSpls.SortQOS([]string{MetaACD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route2", "route3", "route1"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS8(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(15.0),
					MetaACD: NewDecimalFromFloat64(-1.0),
				},
				SortingData: map[string]any{
					Weight:  15.0,
					MetaACD: -1.0,
				},
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(25.0),
					MetaACD: NewDecimalFromFloat64(-1.0),
				},
				SortingData: map[string]any{
					Weight:  25.0,
					MetaACD: -1.0,
				},
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Weight:  NewDecimalFromFloat64(20.0),
					MetaACD: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Weight:  20.0,
					MetaACD: 10.0,
				},
			},
		},
	}
	sSpls.SortQOS([]string{MetaACD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route3", "route2", "route1"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortLoadDistribution(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(25.0),
					Ratio:  NewDecimalFromFloat64(4.0),
					Load:   NewDecimalFromFloat64(3.0),
				},
				SortingData: map[string]any{
					Weight: 25.0,
					Ratio:  4.0,
					Load:   3.0,
				},
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(15.0),
					Ratio:  NewDecimalFromFloat64(10.0),
					Load:   NewDecimalFromFloat64(5.0),
				},
				SortingData: map[string]any{
					Weight: 15.0,
					Ratio:  10.0,
					Load:   5.0,
				},
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Weight: NewDecimalFromFloat64(25.0),
					Ratio:  NewDecimalFromFloat64(1.0),
					Load:   NewDecimalFromFloat64(1.0),
				},
				SortingData: map[string]any{
					Weight: 25.0,
					Ratio:  1.0,
					Load:   1.0,
				},
			},
		},
	}
	sSpls.SortLoadDistribution()
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route2", "route1", "route3"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID
	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func BenchmarkRouteSortCost(b *testing.B) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 10.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		sSpls.SortLeastCost()
	}
}

func TestRouteIDsGetIDs(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 10.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}
	expected := []string{"route1", "route2", "route3"}
	sort.Strings(expected)
	rcv := sSpls.RouteIDs()
	sort.Strings(rcv)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}

func TestSortHighestCost(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(11.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 11.0,
				},
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					Cost:   NewDecimalFromFloat64(0.1),
					Weight: NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Cost:   0.1,
					Weight: 10.0,
				},
			},
		},
	}
	sSpls.SortHighestCost()
	ex := sSpls
	if !reflect.DeepEqual(ex, sSpls) {
		t.Errorf("Expected %+v, received %+v", ex, sSpls)
	}
}

func TestSortLoadDistribution(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "ROUTE1",
				sortingDataDecimal: map[string]*Decimal{
					Ratio:  NewDecimalFromFloat64(6.0),
					Load:   NewDecimalFromFloat64(10.0),
					Weight: NewDecimalFromFloat64(15.5),
				},
				SortingData: map[string]any{
					Ratio:  6.0,
					Load:   10.0,
					Weight: 15.5,
				},
			},
			{
				RouteID: "ROUTE2",
				sortingDataDecimal: map[string]*Decimal{
					Ratio:  NewDecimalFromFloat64(6.0),
					Load:   NewDecimalFromFloat64(10.0),
					Weight: NewDecimalFromFloat64(14.5),
				},
				SortingData: map[string]any{
					Ratio:  6.0,
					Load:   10.0,
					Weight: 14.5,
				},
			},
		},
	}
	sSpls.SortLoadDistribution()
	expSSPls := sSpls
	if !reflect.DeepEqual(sSpls, expSSPls) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expSSPls), ToJSON(sSpls))
	}
}

func TestSortedRouteAsNavigableMap(t *testing.T) {
	sSpls := &SortedRoute{
		RouteID:         "ROUTE1",
		RouteParameters: "SORTING_PARAMETER",
		sortingDataDecimal: map[string]*Decimal{
			Ratio:  NewDecimalFromFloat64(6.0),
			Load:   NewDecimalFromFloat64(10.0),
			Weight: NewDecimalFromFloat64(15.5),
		},
		SortingData: map[string]any{
			Ratio:  6.0,
			Load:   10.0,
			Weight: 15.5,
		},
	}
	expNavMap := &DataNode{
		Type: NMMapType,
		Map: map[string]*DataNode{
			RouteID:         NewLeafNode("ROUTE1"),
			RouteParameters: NewLeafNode("SORTING_PARAMETER"),
			SortingData: {
				Type: NMMapType,
				Map: map[string]*DataNode{
					Ratio:  NewLeafNode(6.0),
					Load:   NewLeafNode(10.0),
					Weight: NewLeafNode(15.5),
				},
			},
		},
	}
	if rcv := sSpls.AsNavigableMap(); !reflect.DeepEqual(rcv, expNavMap) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expNavMap), ToJSON(rcv))
	}
}

func TestSortedRoutesAsNavigableMap(t *testing.T) {
	sSpls := &SortedRoutes{
		ProfileID: "TEST_ID1",
		Sorting:   MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID:         "ROUTE1",
				RouteParameters: "SORTING_PARAMETER",
				sortingDataDecimal: map[string]*Decimal{
					Ratio:  NewDecimalFromFloat64(6.0),
					Load:   NewDecimalFromFloat64(10.0),
					Weight: NewDecimalFromFloat64(15.5),
				},
				SortingData: map[string]any{
					Ratio:  6.0,
					Load:   10.0,
					Weight: 15.5,
				},
			},
			{
				RouteID:         "ROUTE2",
				RouteParameters: "SORTING_PARAMETER_SECOND",
				sortingDataDecimal: map[string]*Decimal{
					Ratio: NewDecimalFromFloat64(7.0),
					Load:  NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					Ratio: 7.0,
					Load:  10.0,
				},
			},
		},
	}

	expNavMap := &DataNode{
		Type: NMMapType,
		Map: map[string]*DataNode{
			ProfileID: NewLeafNode("TEST_ID1"),
			Sorting:   NewLeafNode(MetaWeight),
			CapRoutes: {
				Type: NMSliceType,
				Slice: []*DataNode{
					{
						Type: NMMapType,
						Map: map[string]*DataNode{
							RouteID:         NewLeafNode("ROUTE1"),
							RouteParameters: NewLeafNode("SORTING_PARAMETER"),
							SortingData: {
								Type: NMMapType,
								Map: map[string]*DataNode{
									Ratio:  NewLeafNode(6.0),
									Load:   NewLeafNode(10.0),
									Weight: NewLeafNode(15.5),
								},
							},
						},
					},
					{
						Type: NMMapType,
						Map: map[string]*DataNode{
							RouteID:         NewLeafNode("ROUTE2"),
							RouteParameters: NewLeafNode("SORTING_PARAMETER_SECOND"),
							SortingData: {
								Type: NMMapType,
								Map: map[string]*DataNode{
									Ratio: NewLeafNode(7.0),
									Load:  NewLeafNode(10.0),
								},
							},
						},
					},
				},
			},
		},
	}

	if rcv := sSpls.AsNavigableMap(); !reflect.DeepEqual(rcv, expNavMap) {
		t.Errorf("Expected %+v, received %+v", ToJSON(expNavMap), ToJSON(rcv))
	}
}

func TestSortedRoutesListRouteIDs(t *testing.T) {
	sr := SortedRoutesList{
		{
			Routes: []*SortedRoute{
				{
					RouteID: "sr1id1",
				},
				{
					RouteID: "sr1id2",
				},
			},
		},
		{
			Routes: []*SortedRoute{
				{
					RouteID: "sr2id1",
				},
				{
					RouteID: "sr2id2",
				},
			},
		},
	}

	val := sr.RouteIDs()
	sort.Slice(val, func(i, j int) bool {
		return val[i] < val[j]
	})
	exp := []string{"sr1id1", "sr1id2", "sr2id1", "sr2id2"}
	if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %v ,received %v", exp, val)
	}
}

func TestSortedRoutesListRoutesWithParams(t *testing.T) {
	sRs := SortedRoutesList{
		{
			Routes: []*SortedRoute{
				{
					RouteID:         "route1",
					RouteParameters: "params1",
				},
				{
					RouteID:         "route2",
					RouteParameters: "params2",
				},
			},
		},
		{
			Routes: []*SortedRoute{
				{
					RouteID:         "route3",
					RouteParameters: "params3",
				},
				{
					RouteID:         "route4",
					RouteParameters: "params4",
				},
			},
		},
	}
	val := sRs.RoutesWithParams()
	sort.Slice(val, func(i, j int) bool {
		return val[i] < val[j]
	})
	exp := []string{"route1:params1", "route2:params2", "route3:params3", "route4:params4"}

	if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %v ,received %v", val, exp)
	}

}

func TestSortedRoutesListAsNavigableMap(t *testing.T) {

	sRs := SortedRoutesList{

		{
			Routes: []*SortedRoute{
				{
					RouteID:         "route1",
					RouteParameters: "params1",
				},
				{
					RouteID:         "route2",
					RouteParameters: "params2",
				},
			},
		},
		{
			Routes: []*SortedRoute{
				{
					RouteID:         "route3",
					RouteParameters: "params3",
				},
				{
					RouteID:         "route4",
					RouteParameters: "params4",
				},
			},
		},
	}
	exp := &DataNode{Type: NMSliceType, Slice: make([]*DataNode, len(sRs))}
	for i, ss := range sRs {
		exp.Slice[i] = ss.AsNavigableMap()
	}
	if rcv := sRs.AsNavigableMap(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected <%v>, \n Received \n<%v>", ToJSON(exp), ToJSON(rcv))

	}

}

func TestLibRoutesSortSameWeight(t *testing.T) {
	tests := []struct {
		name string
		data map[string]*Decimal
		sort func(sr *SortedRoutes)
	}{
		{
			name: "SortWeight",
			data: map[string]*Decimal{
				Weight: NewDecimalFromFloat64(10.0),
			},
			sort: func(sr *SortedRoutes) { sr.SortWeight() },
		},
		{
			name: "SortLeastCost",
			data: map[string]*Decimal{
				Cost:   NewDecimalFromFloat64(0.1),
				Weight: NewDecimalFromFloat64(10.0),
			},
			sort: func(sr *SortedRoutes) { sr.SortLeastCost() },
		},
		{
			name: "SortHighestCost",
			data: map[string]*Decimal{
				Cost:   NewDecimalFromFloat64(0.1),
				Weight: NewDecimalFromFloat64(10.0),
			},
			sort: func(sr *SortedRoutes) { sr.SortHighestCost() },
		},
		{
			name: "SortResourceAscendent",
			data: map[string]*Decimal{
				ResourceUsageStr: NewDecimalFromFloat64(5.0),
				Weight:           NewDecimalFromFloat64(10.0),
			},
			sort: func(sr *SortedRoutes) { sr.SortResourceAscendent() },
		},
		{
			name: "SortResourceDescendent",
			data: map[string]*Decimal{
				ResourceUsageStr: NewDecimalFromFloat64(5.0),
				Weight:           NewDecimalFromFloat64(10.0),
			},
			sort: func(sr *SortedRoutes) { sr.SortResourceDescendent() },
		},
		{
			name: "SortLoadDistribution",
			data: map[string]*Decimal{
				Ratio:  NewDecimalFromFloat64(4.0),
				Load:   NewDecimalFromFloat64(3.0),
				Weight: NewDecimalFromFloat64(10.0),
			},
			sort: func(sr *SortedRoutes) { sr.SortLoadDistribution() },
		},
		{
			name: "SortQOS",
			data: map[string]*Decimal{
				Weight:  NewDecimalFromFloat64(10.0),
				MetaACD: NewDecimalFromFloat64(-1.0),
			},
			sort: func(sr *SortedRoutes) { sr.SortQOS([]string{MetaACD}) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sSpls := &SortedRoutes{}
			for i := 0; i <= 10; i++ {
				sSpls.Routes = append(sSpls.Routes, &SortedRoute{
					RouteID:            strconv.Itoa(i),
					sortingDataDecimal: tt.data,
				})
			}
			order := sSpls.RoutesWithParams()
			shuffled := false
			for i := 0; i < 100; i++ {
				tt.sort(sSpls)
				if !reflect.DeepEqual(sSpls.RoutesWithParams(), order) {
					shuffled = true
					break
				}
			}
			if !shuffled {
				t.Errorf("Routes kept their order after sorting")
			}
		})
	}
}

func TestSortResourceAscendentDescendent(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataDecimal: map[string]*Decimal{
					ResourceUsageStr: NewDecimalFromFloat64(10.0),
					Weight:           NewDecimalFromFloat64(10.0),
				},
				SortingData: map[string]any{
					ResourceUsageStr: 10.0,
					Weight:           10.0,
				},
			},
			{
				RouteID: "route2",
				sortingDataDecimal: map[string]*Decimal{
					ResourceUsageStr: NewDecimalFromFloat64(10.0),
					Weight:           NewDecimalFromFloat64(11.0),
				},
				SortingData: map[string]any{
					ResourceUsageStr: 10.0,
					Weight:           11.0,
				},
			},
		},
	}

	exp := []string{"route2", "route1"}
	sSpls.SortResourceAscendent()
	if rcv := sSpls.RoutesWithParams(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v, received %v", exp, rcv)
	}

	sSpls.SortResourceDescendent()
	if rcv := sSpls.RoutesWithParams(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v, received %v", exp, rcv)
	}

	sSpls = &SortedRoutes{Routes: []*SortedRoute{
		{
			RouteID: "route1",
			sortingDataDecimal: map[string]*Decimal{
				ResourceUsageStr: NewDecimalFromFloat64(11),
				Weight:           NewDecimalFromFloat64(10),
			},
			SortingData: map[string]any{
				ResourceUsageStr: 10.0,
				Weight:           10.0,
			},
		},
		{
			RouteID: "route2",
			sortingDataDecimal: map[string]*Decimal{
				ResourceUsageStr: NewDecimalFromFloat64(10),
				Weight:           NewDecimalFromFloat64(10),
			},
			SortingData: map[string]any{
				ResourceUsageStr: 10.0,
				Weight:           10.0,
			},
		},
	}}
	exp = []string{"route2", "route1"}
	sSpls.SortResourceAscendent()
	if rcv := sSpls.RoutesWithParams(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v, received %v", exp, rcv)
	}
	exp = []string{"route1", "route2"}
	sSpls.SortResourceDescendent()
	if rcv := sSpls.RoutesWithParams(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v, received %v", exp, rcv)
	}
}
