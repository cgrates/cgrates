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
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestHealthAccountAction(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAccountActionPlans("1001", []string{"AP1", "AP2"}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP2", &ActionPlan{
		Id:            "AP2",
		AccountIDs:    utils.NewStringMap("1002"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{"1002": {"AP2"}},
		BrokenReferences:          map[string][]string{"AP2": {"1001"}, "AP1": nil},
	}
	if rply, err := GetAccountActionPlansIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthAccountAction2(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAccountActionPlans("1001", []string{"AP1", "AP2"}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP2", &ActionPlan{
		Id:            "AP2",
		AccountIDs:    utils.NewStringMap("1001"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{},
		BrokenReferences:          map[string][]string{"AP1": nil},
	}
	if rply, err := GetAccountActionPlansIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthAccountAction3(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAccountActionPlans("1002", []string{"AP1"}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP1", &ActionPlan{
		Id:            "AP1",
		AccountIDs:    utils.NewStringMap("1002"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP2", &ActionPlan{
		Id:            "AP2",
		AccountIDs:    utils.NewStringMap("1002"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{"1002": {"AP2"}},
		BrokenReferences:          map[string][]string{},
	}
	if rply, err := GetAccountActionPlansIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthAccountAction4(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAccountActionPlans("1002", []string{"AP2", "AP1"}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetAccountActionPlans("1001", []string{"AP2"}, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP1", &ActionPlan{
		Id:            "AP1",
		AccountIDs:    utils.NewStringMap("1002"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetActionPlan("AP2", &ActionPlan{
		Id:            "AP2",
		AccountIDs:    utils.NewStringMap("1001"),
		ActionTimings: []*ActionTiming{{}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{},
		BrokenReferences:          map[string][]string{"AP2": {"1002"}},
	}
	if rply, err := GetAccountActionPlansIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthReverseDestination(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination("DST1", []string{"1001", "1002"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetReverseDestination("DST2", []string{"1001"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST2",
		Prefixes: []string{"1002"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{"1002": {"DST2"}},
		BrokenReferences:           map[string][]string{"DST1": nil, "DST2": {"1001"}},
	}
	if rply, err := GetReverseDestinationsIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthReverseDestination2(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination("DST1", []string{"1001"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetReverseDestination("DST2", []string{"1001"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST2",
		Prefixes: []string{"1001"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{},
		BrokenReferences:           map[string][]string{"DST1": nil},
	}
	if rply, err := GetReverseDestinationsIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthReverseDestination3(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination("DST1", []string{"1002"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST1",
		Prefixes: []string{"1002"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST2",
		Prefixes: []string{"1002"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{"1002": {"DST2"}},
		BrokenReferences:           map[string][]string{},
	}
	if rply, err := GetReverseDestinationsIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthReverseDestination4(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination("DST1", []string{"1002"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetReverseDestination("DST2", []string{"1001", "1002"}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST1",
		Prefixes: []string{"1002"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetDestination(&Destination{
		Id:       "DST2",
		Prefixes: []string{"1001"},
	}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	exp := &ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{},
		BrokenReferences:           map[string][]string{"DST2": {"1002"}},
	}
	if rply, err := GetReverseDestinationsIndexHealth(dm, -1, -1, -1, -1, false, false); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthFilter(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAttributeProfile(&AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		Contexts:  []string{utils.MetaAny},
		FilterIDs: []string{"*string:~*req.Account:1001", "Fltr1"},
	}, false); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetIndexes(utils.CacheAttributeFilterIndexes, "cgrates.org:*any",
		map[string]utils.StringSet{"*string:*req.Account:1002": {
			"ATTR1": {},
			"ATTR2": {},
		}},
		true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*any:*string:*req.Account:1001": {"ATTR1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*string:*req.Account:1002": {"ATTR1"},
		},
		MissingFilters: map[string][]string{
			"cgrates.org:Fltr1": {"ATTR1"},
		},
		MissingObjects: []string{"cgrates.org:ATTR2"},
	}

	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(-1, 0, false, nil),
		ltcache.NewCache(-1, 0, false, nil),
		ltcache.NewCache(-1, 0, false, nil),
		utils.CacheAttributeFilterIndexes); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthReverseFilter(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAttributeProfile(&AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		Contexts:  []string{utils.MetaAny},
		FilterIDs: []string{"*string:~*req.Account:1001", "Fltr1"},
	}, false); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetIndexes(utils.CacheReverseFilterIndexes, "cgrates.org:Fltr2",
		map[string]utils.StringSet{utils.CacheAttributeFilterIndexes: {"ATTR1:*cdrs": {}, "ATTR2:*any": {}}},
		true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetIndexes(utils.CacheReverseFilterIndexes, "cgrates.org:Fltr1",
		map[string]utils.StringSet{utils.CacheAttributeFilterIndexes: {"ATTR1:*cdrs": {}}},
		true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	exp := map[string]*ReverseFilterIHReply{
		utils.CacheAttributeFilterIndexes: {
			MissingReverseIndexes: map[string][]string{
				// "cgrates.org:ATTR1": {"Fltr1:*any"},
			},
			MissingFilters: map[string][]string{"cgrates.org:Fltr1": {"ATTR1"}},
			BrokenReverseIndexes: map[string][]string{
				"cgrates.org:ATTR1:*cdrs": {"Fltr1", "Fltr2"},
			},
			MissingObjects: []string{"cgrates.org:ATTR2"},
		},
	}

	objCaches := make(map[string]*ltcache.Cache)
	for indxType := range utils.CacheIndexesToPrefix {
		objCaches[indxType] = ltcache.NewCache(-1, 0, false, nil)
	}
	if rply, err := GetRevFltrIdxHealth(dm,
		ltcache.NewCache(-1, 0, false, nil),
		ltcache.NewCache(-1, 0, false, nil),
		objCaches); err != nil {
		t.Fatal(err)
	} else {
		sort.Strings(rply[utils.CacheAttributeFilterIndexes].BrokenReverseIndexes["cgrates.org:ATTR1:*cdrs"])
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}
}

func TestHealthIndexThreshold(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this threshold but without indexing
	thPrf := &ThresholdProfileWithAPIOpts{
		ThresholdProfile: &ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "TestHealthIndexThreshold",
			FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.ID:1002",         // *asm will not be indexing
				"*suffix:BrokenFilter:Invalid"}, // static value, won't index
			MaxHits: 1,
		},
	}
	if err := dm.SetThresholdProfile(thPrf.ThresholdProfile, false); err != nil {
		t.Error(err)
	}

	args := &IndexHealthArgsWith3Ch{}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*eventType:AccountUpdate": {"TestHealthIndexThreshold"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheThresholdFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	indexes := map[string]utils.StringSet{
		"*prefix:req.InvalidIdx:10": { // obj exist but the index don't
			"TestHealthIndexThreshold": {},
		},
		"*string:*req.Destination:123": { // index is valid but the obj does not exist
			"InexistingThreshold": {},
		},
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	if err := dm.SetIndexes(utils.CacheThresholdFilterIndexes, "cgrates.org",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingThreshold"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*eventType:AccountUpdate": {"TestHealthIndexThreshold"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*prefix:req.InvalidIdx:10": {"TestHealthIndexThreshold"},
		},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheThresholdFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	//we will use an inexisting Filter(not inline) for the same ThresholdProfile
	thPrf = &ThresholdProfileWithAPIOpts{
		ThresholdProfile: &ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "TestHealthIndexThreshold",
			FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.ID:1002",
				"FLTR_1_DOES_NOT_EXIST"},
			MaxHits: 1,
		},
	}
	if err := dm.SetThresholdProfile(thPrf.ThresholdProfile, false); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingThreshold"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*eventType:AccountUpdate": {"TestHealthIndexThreshold"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*prefix:req.InvalidIdx:10": {"TestHealthIndexThreshold"},
		},
		MissingFilters: map[string][]string{
			"cgrates.org:FLTR_1_DOES_NOT_EXIST": {"TestHealthIndexThreshold"},
		},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheThresholdFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthIndexCharger(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this charger but without indexing
	chPrf := &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Raw",
		FilterIDs: []string{
			"*string:~*opts.*eventType:ChargerAccountUpdate",
			"*string:~*req.*Account:1234",
			"*string:~*asm.ID:1002", // *asm will not be indexing
			"*suffix:BrokenFilter:Invalid"},
		RunID:        "raw",
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       20,
	}
	if err := dm.SetChargerProfile(chPrf, false); err != nil {
		t.Error(err)
	}

	args := &IndexHealthArgsWith3Ch{}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*eventType:ChargerAccountUpdate": {"Raw"},
			"cgrates.org:*string:*req.*Account:1234":                    {"Raw"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheChargerFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	indexes := map[string]utils.StringSet{
		"*prefix:req.Destination:+10": { // obj exist but the index don't
			"Raw": {},
		},
		"*string:*req.Destination:123": { // index is valid but the obj does not exist
			"InexistingCharger": {},
		},
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	if err := dm.SetIndexes(utils.CacheChargerFilterIndexes, "cgrates.org",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingCharger"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*eventType:ChargerAccountUpdate": {"Raw"},
			"cgrates.org:*string:*req.*Account:1234":                    {"Raw"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*prefix:req.Destination:+10": {"Raw"},
		},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheChargerFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	//we will use an inexisting Filter(not inline) for the same ChargerProfile
	chPrf = &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Raw",
		FilterIDs: []string{
			"*string:~*opts.*eventType:ChargerAccountUpdate",
			"*string:~*req.*Account:1234",
			"*string:~*asm.ID:1002", // *asm will not be indexing
			"*suffix:BrokenFilter:Invalid",
			"FLTR_1_DOES_NOT_EXIST_CHRGR"},
		RunID:        "raw",
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       20,
	}
	if err := dm.SetChargerProfile(chPrf, false); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingCharger"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*eventType:ChargerAccountUpdate": {"Raw"},
			"cgrates.org:*string:*req.*Account:1234":                    {"Raw"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*prefix:req.Destination:+10": {"Raw"},
		},
		MissingFilters: map[string][]string{
			"cgrates.org:FLTR_1_DOES_NOT_EXIST_CHRGR": {"Raw"},
		},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheChargerFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthIndexResources(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this resource but without indexing
	rsPrf := &ResourceProfile{
		Tenant: "tenant.custom",
		ID:     "RES_GRP1",
		FilterIDs: []string{
			"*string:~*opts.*eventType:ResourceAccountUpdate",
			"*string:~*req.RequestType:*rated",
			"*prefix:~*accounts.RES_GRP1.Available:10", // *accounts will not be indexing
			"*suffix:BrokenFilter:Invalid",
		},
		UsageTTL:          10 * time.Microsecond,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
	}
	if err := dm.SetResourceProfile(rsPrf, false); err != nil {
		t.Error(err)
	}

	args := &IndexHealthArgsWith3Ch{}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"tenant.custom:*string:*opts.*eventType:ResourceAccountUpdate": {"RES_GRP1"},
			"tenant.custom:*string:*req.RequestType:*rated":                {"RES_GRP1"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheResourceFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	indexes := map[string]utils.StringSet{
		"*suffix:*req.Destination:+10": { // obj exist but the index don't
			"RES_GRP1": {},
		},
		"*string:*req.CGRID:not_an_id": { // index is valid but the obj does not exist
			"InexistingResource": {},
		},
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	if err := dm.SetIndexes(utils.CacheResourceFilterIndexes, "tenant.custom",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"tenant.custom:InexistingResource"},
		MissingIndexes: map[string][]string{
			"tenant.custom:*string:*opts.*eventType:ResourceAccountUpdate": {"RES_GRP1"},
			"tenant.custom:*string:*req.RequestType:*rated":                {"RES_GRP1"},
		},
		BrokenIndexes: map[string][]string{
			"tenant.custom:*suffix:*req.Destination:+10": {"RES_GRP1"},
		},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheResourceFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	//we will use an inexisting Filter(not inline) for the same ResourceProfile
	rsPrf = &ResourceProfile{
		Tenant: "tenant.custom",
		ID:     "RES_GRP1",
		FilterIDs: []string{
			"*string:~*opts.*eventType:ResourceAccountUpdate",
			"*string:~*req.RequestType:*rated",
			"*prefix:~*accounts.RES_GRP1.Available:10", // *asm will not be indexing
			"*suffix:BrokenFilter:Invalid",
			"FLTR_1_NOT_EXIST",
		},
		UsageTTL:          10 * time.Microsecond,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
	}
	if err := dm.SetResourceProfile(rsPrf, false); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"tenant.custom:InexistingResource"},
		MissingIndexes: map[string][]string{
			"tenant.custom:*string:*opts.*eventType:ResourceAccountUpdate": {"RES_GRP1"},
			"tenant.custom:*string:*req.RequestType:*rated":                {"RES_GRP1"},
		},
		BrokenIndexes: map[string][]string{
			"tenant.custom:*suffix:*req.Destination:+10": {"RES_GRP1"},
		},
		MissingFilters: map[string][]string{
			"tenant.custom:FLTR_1_NOT_EXIST": {"RES_GRP1"},
		},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheResourceFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthIndexStats(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this statQueue but without indexing
	sqPrf := &StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "Stat_1",
		FilterIDs: []string{
			"*string:~*opts.*apikey:sts1234",
			"*string:~*req.RequestType:*postpaid",
			"*prefix:~*resources.RES_GRP1.Available:10", // *resources will not be indexing
			"*suffix:BrokenFilter:Invalid",
		},
		Weight:      30,
		QueueLength: 100,
		TTL:         10 * time.Second,
		MinItems:    0,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*tcd",
			},
			{
				MetricID: "*asr",
			},
			{
				MetricID: "*acd",
			},
		},
		Blocker:      true,
		ThresholdIDs: []string{utils.MetaNone},
	}
	if err := dm.SetStatQueueProfile(sqPrf, false); err != nil {
		t.Error(err)
	}

	args := &IndexHealthArgsWith3Ch{}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*apikey:sts1234":      {"Stat_1"},
			"cgrates.org:*string:*req.RequestType:*postpaid": {"Stat_1"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheStatFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	indexes := map[string]utils.StringSet{
		"*suffix:*req.Destination:+60": { // obj exist but the index don't
			"Stat_1": {},
		},
		"*string:*req.ExtraField:Usage": { // index is valid but the obj does not exist
			"InexistingStats": {},
		},
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	if err := dm.SetIndexes(utils.CacheStatFilterIndexes, "cgrates.org",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingStats"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*apikey:sts1234":      {"Stat_1"},
			"cgrates.org:*string:*req.RequestType:*postpaid": {"Stat_1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*suffix:*req.Destination:+60": {"Stat_1"},
		},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheStatFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	//we will use an inexisting Filter(not inline) for the same StatQueueProfile
	sqPrf = &StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "Stat_1",
		FilterIDs: []string{
			"*string:~*opts.*apikey:sts1234",
			"*string:~*req.RequestType:*postpaid",
			"*prefix:~*resources.RES_GRP1.Available:10", // *resources will not be indexing
			"*suffix:BrokenFilter:Invalid",
			"FLTR_1_NOT_EXIST",
		},
		Weight:      30,
		QueueLength: 100,
		TTL:         10 * time.Second,
		MinItems:    0,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*tcd",
			},
			{
				MetricID: "*asr",
			},
			{
				MetricID: "*acd",
			},
		},
		Blocker:      true,
		ThresholdIDs: []string{utils.MetaNone},
	}
	if err := dm.SetStatQueueProfile(sqPrf, false); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingStats"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*apikey:sts1234":      {"Stat_1"},
			"cgrates.org:*string:*req.RequestType:*postpaid": {"Stat_1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*suffix:*req.Destination:+60": {"Stat_1"},
		},
		MissingFilters: map[string][]string{
			"cgrates.org:FLTR_1_NOT_EXIST": {"Stat_1"},
		},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheStatFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestHealthIndexRoutes(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this statQueue but without indexing
	sqPrf := &StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "Stat_1",
		FilterIDs: []string{
			"*string:~*opts.*apikey:sts1234",
			"*string:~*req.RequestType:*postpaid",
			"*prefix:~*resources.RES_GRP1.Available:10", // *resources will not be indexing
			"*suffix:BrokenFilter:Invalid",
		},
		Weight:      30,
		QueueLength: 100,
		TTL:         10 * time.Second,
		MinItems:    0,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*tcd",
			},
			{
				MetricID: "*asr",
			},
			{
				MetricID: "*acd",
			},
		},
		Blocker:      true,
		ThresholdIDs: []string{utils.MetaNone},
	}
	if err := dm.SetStatQueueProfile(sqPrf, false); err != nil {
		t.Error(err)
	}

	args := &IndexHealthArgsWith3Ch{}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*apikey:sts1234":      {"Stat_1"},
			"cgrates.org:*string:*req.RequestType:*postpaid": {"Stat_1"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheStatFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	indexes := map[string]utils.StringSet{
		"*suffix:*req.Destination:+60": { // obj exist but the index don't
			"Stat_1": {},
		},
		"*string:*req.ExtraField:Usage": { // index is valid but the obj does not exist
			"InexistingStats": {},
		},
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	if err := dm.SetIndexes(utils.CacheStatFilterIndexes, "cgrates.org",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingStats"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*apikey:sts1234":      {"Stat_1"},
			"cgrates.org:*string:*req.RequestType:*postpaid": {"Stat_1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*suffix:*req.Destination:+60": {"Stat_1"},
		},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheStatFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	//we will use an inexisting Filter(not inline) for the same StatQueueProfile
	sqPrf = &StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "Stat_1",
		FilterIDs: []string{
			"*string:~*opts.*apikey:sts1234",
			"*string:~*req.RequestType:*postpaid",
			"*prefix:~*resources.RES_GRP1.Available:10", // *resources will not be indexing
			"*suffix:BrokenFilter:Invalid",
			"FLTR_1_NOT_EXIST",
		},
		Weight:      30,
		QueueLength: 100,
		TTL:         10 * time.Second,
		MinItems:    0,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*tcd",
			},
			{
				MetricID: "*asr",
			},
			{
				MetricID: "*acd",
			},
		},
		Blocker:      true,
		ThresholdIDs: []string{utils.MetaNone},
	}
	if err := dm.SetStatQueueProfile(sqPrf, false); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingStats"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*apikey:sts1234":      {"Stat_1"},
			"cgrates.org:*string:*req.RequestType:*postpaid": {"Stat_1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*suffix:*req.Destination:+60": {"Stat_1"},
		},
		MissingFilters: map[string][]string{
			"cgrates.org:FLTR_1_NOT_EXIST": {"Stat_1"},
		},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheStatFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}
