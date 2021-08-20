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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestHealthAccountAction(t *testing.T) {
	Cache.Clear(nil)
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination(&Destination{Id: "DST1", Prefixes: []string{"1001", "1002"}}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetReverseDestination(&Destination{Id: "DST2", Prefixes: []string{"1001"}}, utils.NonTransactional); err != nil {
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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination(&Destination{Id: "DST1", Prefixes: []string{"1001"}}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetReverseDestination(&Destination{Id: "DST2", Prefixes: []string{"1001"}}, utils.NonTransactional); err != nil {
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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination(&Destination{Id: "DST1", Prefixes: []string{"1002"}}, utils.NonTransactional); err != nil {
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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetReverseDestination(&Destination{Id: "DST1", Prefixes: []string{"1002"}}, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetReverseDestination(&Destination{Id: "DST2", Prefixes: []string{"1001", "1002"}}, utils.NonTransactional); err != nil {
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

func TestHealthFilterAttributes(t *testing.T) {
	Cache.Clear(nil)
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAttributeProfile(&AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		Contexts:  []string{utils.META_ANY},
		FilterIDs: []string{"*string:~*req.Account:1001", "Fltr1"},
	}, false); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetFilterIndexes(utils.CacheAttributeFilterIndexes, "cgrates.org:*any",
		map[string]utils.StringMap{"*string:~*req.Account:1002": {"ATTR1": true, "ATTR2": true}},
		true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*any:*string:~*req.Account:1001": {"ATTR1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*string:~*req.Account:1002": {"ATTR1"},
		},
		MissingFilters: map[string][]string{
			"Fltr1": {"ATTR1"},
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

func TestHealthIndexThreshold(t *testing.T) {
	Cache.Clear(nil)
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this threshold but without indexing
	thPrf := &ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     "TestHealthIndexThreshold",
		FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate",
			"*string:~*asm.ID:1002",
			"*prefix:StaticValue:AlwaysTrue",
		},
		MaxHits: 1,
	}

	if err := dm.SetThresholdProfile(thPrf, false); err != nil {
		t.Error(err)
	}

	args := &IndexHealthArgsWith3Ch{}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:~*opts.*eventType:AccountUpdate": {"TestHealthIndexThreshold"},
			"cgrates.org:*string:~*asm.ID:1002":                   {"TestHealthIndexThreshold"},
			"cgrates.org:*prefix:StaticValue:AlwaysTrue":          {"TestHealthIndexThreshold"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
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

	indexes := map[string]utils.StringMap{
		"*prefix:req.InvalidIdx:10": { // obj exist but the index don't
			"TestHealthIndexThreshold": true,
		},
		"*string:*req.Destination:123": { // index is valid but the obj does not exist
			"InexistingThreshold": true,
		},
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	if err := dm.SetFilterIndexes(utils.CacheThresholdFilterIndexes, "cgrates.org",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingThreshold"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:~*opts.*eventType:AccountUpdate": {"TestHealthIndexThreshold"},
			"cgrates.org:*string:~*asm.ID:1002":                   {"TestHealthIndexThreshold"},
			"cgrates.org:*prefix:StaticValue:AlwaysTrue":          {"TestHealthIndexThreshold"},
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
	thPrf = &ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     "TestHealthIndexThreshold",
		FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate",
			"*string:~*asm.ID:1002",
			"FLTR_1_DOES_NOT_EXIST",
		},
		MaxHits: 1,
	}
	if err := dm.SetThresholdProfile(thPrf, false); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingThreshold"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:~*opts.*eventType:AccountUpdate": {"TestHealthIndexThreshold"},
			"cgrates.org:*string:~*asm.ID:1002":                   {"TestHealthIndexThreshold"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*prefix:req.InvalidIdx:10": {"TestHealthIndexThreshold"},
		},
		MissingFilters: map[string][]string{
			"FLTR_1_DOES_NOT_EXIST": {"TestHealthIndexThreshold"},
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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this charger but without indexing
	chPrf := &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Raw",
		FilterIDs: []string{
			"*string:~*opts.*eventType:ChargerAccountUpdate",
			"*string:~*req.*Account:1234",
			"*string:~*asm.ID:1002",
			"*suffix:BrokenIndex:Invalid"}, // suffix is not indexed
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
			"cgrates.org:*string:~*opts.*eventType:ChargerAccountUpdate": {"Raw"},
			"cgrates.org:*string:~*req.*Account:1234":                    {"Raw"},
			"cgrates.org:*string:~*asm.ID:1002":                          {"Raw"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
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

	indexes := map[string]utils.StringMap{
		"*prefix:req.Destination:+10": { // obj exist but the index don't
			"Raw": true,
		},
		"*string:*req.Destination:123": { // index is valid but the obj does not exist
			"InexistingCharger": true,
		},
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	if err := dm.SetFilterIndexes(utils.CacheChargerFilterIndexes, "cgrates.org",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingCharger"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:~*opts.*eventType:ChargerAccountUpdate": {"Raw"},
			"cgrates.org:*string:~*req.*Account:1234":                    {"Raw"},
			"cgrates.org:*string:~*asm.ID:1002":                          {"Raw"},
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
			"*string:~*asm.ID:1002",
			"*suffix:BrokenFilter:Invalid",
			"FLTR_1_DOES_NOT_EXIST_CHRGR",
		},
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
			"cgrates.org:*string:~*opts.*eventType:ChargerAccountUpdate": {"Raw"},
			"cgrates.org:*string:~*req.*Account:1234":                    {"Raw"},
			"cgrates.org:*string:~*asm.ID:1002":                          {"Raw"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*prefix:req.Destination:+10": {"Raw"},
		},
		MissingFilters: map[string][]string{
			"FLTR_1_DOES_NOT_EXIST_CHRGR": {"Raw"},
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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this resource but without indexing
	rsPrf := &ResourceProfile{
		Tenant: "tenant.custom",
		ID:     "RES_GRP1",
		FilterIDs: []string{
			"*string:~*opts.*eventType:ResourceAccountUpdate",
			"*string:~*req.RequestType:*rated",
			"*prefix:~*accounts.RES_GRP1.Available:10",
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
			"tenant.custom:*string:~*opts.*eventType:ResourceAccountUpdate": {"RES_GRP1"},
			"tenant.custom:*string:~*req.RequestType:*rated":                {"RES_GRP1"},
			"tenant.custom:*prefix:~*accounts.RES_GRP1.Available:10":        {"RES_GRP1"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
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

	indexes := map[string]utils.StringMap{
		"*suffix:*req.Destination:+10": { // obj exist but the index don't
			"RES_GRP1": true,
		},
		"*string:*req.CGRID:not_an_id": { // index is valid but the obj does not exist
			"InexistingResource": true,
		},
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	if err := dm.SetFilterIndexes(utils.CacheResourceFilterIndexes, "tenant.custom",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"tenant.custom:InexistingResource"},
		MissingIndexes: map[string][]string{
			"tenant.custom:*string:~*opts.*eventType:ResourceAccountUpdate": {"RES_GRP1"},
			"tenant.custom:*string:~*req.RequestType:*rated":                {"RES_GRP1"},
			"tenant.custom:*prefix:~*accounts.RES_GRP1.Available:10":        {"RES_GRP1"},
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
			"*prefix:~*accounts.RES_GRP1.Available:10",
			"*suffix:BrokenFilter:Invalid", // suffix will not be indexed
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
			"tenant.custom:*string:~*opts.*eventType:ResourceAccountUpdate": {"RES_GRP1"},
			"tenant.custom:*string:~*req.RequestType:*rated":                {"RES_GRP1"},
			"tenant.custom:*prefix:~*accounts.RES_GRP1.Available:10":        {"RES_GRP1"},
		},
		BrokenIndexes: map[string][]string{
			"tenant.custom:*suffix:*req.Destination:+10": {"RES_GRP1"},
		},
		MissingFilters: map[string][]string{
			"FLTR_1_NOT_EXIST": {"RES_GRP1"},
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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this statQueue but without indexing
	sqPrf := &StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "Stat_1",
		FilterIDs: []string{
			"*string:~*opts.*apikey:sts1234",
			"*string:~*req.RequestType:*postpaid",
			"*prefix:~*resources.RES_GRP1.Available:10",
			"*suffix:BrokenFilter:Invalid", // suffix will not be indexed
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
		ThresholdIDs: []string{utils.META_NONE},
	}
	if err := dm.SetStatQueueProfile(sqPrf, false); err != nil {
		t.Error(err)
	}

	args := &IndexHealthArgsWith3Ch{}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:~*opts.*apikey:sts1234":            {"Stat_1"},
			"cgrates.org:*string:~*req.RequestType:*postpaid":       {"Stat_1"},
			"cgrates.org:*prefix:~*resources.RES_GRP1.Available:10": {"Stat_1"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
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

	indexes := map[string]utils.StringMap{
		"*suffix:*req.Destination:+60": { // obj exist but the index don't
			"Stat_1": true,
		},
		"*string:*req.ExtraField:Usage": { // index is valid but the obj does not exist
			"InexistingStats": true,
		},
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	if err := dm.SetFilterIndexes(utils.CacheStatFilterIndexes, "cgrates.org",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingStats"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:~*opts.*apikey:sts1234":            {"Stat_1"},
			"cgrates.org:*string:~*req.RequestType:*postpaid":       {"Stat_1"},
			"cgrates.org:*prefix:~*resources.RES_GRP1.Available:10": {"Stat_1"},
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
		ThresholdIDs: []string{utils.META_NONE},
	}
	if err := dm.SetStatQueueProfile(sqPrf, false); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingStats"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:~*opts.*apikey:sts1234":            {"Stat_1"},
			"cgrates.org:*string:~*req.RequestType:*postpaid":       {"Stat_1"},
			"cgrates.org:*prefix:~*resources.RES_GRP1.Available:10": {"Stat_1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*suffix:*req.Destination:+60": {"Stat_1"},
		},
		MissingFilters: map[string][]string{
			"FLTR_1_NOT_EXIST": {"Stat_1"},
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

func TestHealthIndexSupplier(t *testing.T) {
	Cache.Clear(nil)
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this routes but without indexing
	rtPrf := &SupplierProfile{
		Tenant: "routes.com",
		ID:     "ROUTE_ACNT_1001",
		FilterIDs: []string{"*string:~*opts.*apikey:rts1234",
			"*string:~*req.Usage:160s",
			"*string:~*stats.STATS_VENDOR_2.*acd:1m",
			"*string:~*nothing.Denied:true",
			"*suffix:BrokenFilter:Invalid", // suffix will not be indexes
		},
		Sorting:           utils.MetaLC,
		SortingParameters: []string{},
		Suppliers: []*Supplier{
			{
				ID:      "route1",
				Weight:  10,
				Blocker: false,
			},
			{
				ID:            "route2",
				RatingPlanIDs: []string{"RP_1002"},
				Weight:        20,
				Blocker:       false,
			},
		},
		Weight: 10,
	}
	if err := dm.SetSupplierProfile(rtPrf, false); err != nil {
		t.Error(err)
	}

	args := &IndexHealthArgsWith3Ch{}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"routes.com:*string:~*opts.*apikey:rts1234":         {"ROUTE_ACNT_1001"},
			"routes.com:*string:~*req.Usage:160s":               {"ROUTE_ACNT_1001"},
			"routes.com:*string:~*nothing.Denied:true":          {"ROUTE_ACNT_1001"},
			"routes.com:*string:~*stats.STATS_VENDOR_2.*acd:1m": {"ROUTE_ACNT_1001"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheSupplierFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	indexes := map[string]utils.StringMap{
		"*string:*req.RequestType:*rated": { // obj exist but the index don't
			"ROUTE_ACNT_1001": true,
		},
		"*suffix:*req.Destination:+222": {
			"ROUTE_ACNT_1001": true,
			"ROUTE_ACNT_1002": true,
		},
		"*suffix:*req.Destination:+333": {
			"ROUTE_ACNT_1001": true,
			"ROUTE_ACNT_1002": true,
		},
		"*string:*req.ExtraField:Usage": { // index is valid but the obj does not exist
			"InexistingRoute": true,
		},
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	if err := dm.SetFilterIndexes(utils.CacheSupplierFilterIndexes, "routes.com",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{
			"routes.com:InexistingRoute",
			"routes.com:ROUTE_ACNT_1002",
		},
		MissingIndexes: map[string][]string{
			"routes.com:*string:~*opts.*apikey:rts1234":         {"ROUTE_ACNT_1001"},
			"routes.com:*string:~*req.Usage:160s":               {"ROUTE_ACNT_1001"},
			"routes.com:*string:~*nothing.Denied:true":          {"ROUTE_ACNT_1001"},
			"routes.com:*string:~*stats.STATS_VENDOR_2.*acd:1m": {"ROUTE_ACNT_1001"},
		},
		BrokenIndexes: map[string][]string{
			"routes.com:*suffix:*req.Destination:+222":   {"ROUTE_ACNT_1001"},
			"routes.com:*suffix:*req.Destination:+333":   {"ROUTE_ACNT_1001"},
			"routes.com:*string:*req.RequestType:*rated": {"ROUTE_ACNT_1001"},
		},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheSupplierFilterIndexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(rply.MissingObjects)
		sort.Strings(exp.MissingObjects)
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}

	//we will use an inexisting Filter(not inline) for the same RouteProfile
	rtPrf = &SupplierProfile{
		Tenant: "routes.com",
		ID:     "ROUTE_ACNT_1001",
		FilterIDs: []string{"*string:~*opts.*apikey:rts1234",
			"*string:~*req.Usage:160s",
			"*string:~*stats.STATS_VENDOR_2.*acd:1m", // *stats will not be indexing
			"*string:~*nothing.Denied:true",
			"*suffix:BrokenFilter:Invalid",
			"FLTR_1_NOT_EXIST",
		},
		Sorting:           utils.MetaLC,
		SortingParameters: []string{},
		Suppliers: []*Supplier{
			{
				ID:      "route1",
				Weight:  10,
				Blocker: false,
			},
			{
				ID:            "route2",
				RatingPlanIDs: []string{"RP_1002"},
				Weight:        20,
				Blocker:       false,
			},
		},
		Weight: 10,
	}
	if err := dm.SetSupplierProfile(rtPrf, false); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{
			"routes.com:InexistingRoute",
			"routes.com:ROUTE_ACNT_1002",
		},
		MissingIndexes: map[string][]string{
			"routes.com:*string:~*opts.*apikey:rts1234":         {"ROUTE_ACNT_1001"},
			"routes.com:*string:~*req.Usage:160s":               {"ROUTE_ACNT_1001"},
			"routes.com:*string:~*nothing.Denied:true":          {"ROUTE_ACNT_1001"},
			"routes.com:*string:~*stats.STATS_VENDOR_2.*acd:1m": {"ROUTE_ACNT_1001"},
		},
		BrokenIndexes: map[string][]string{
			"routes.com:*suffix:*req.Destination:+222":   {"ROUTE_ACNT_1001"},
			"routes.com:*suffix:*req.Destination:+333":   {"ROUTE_ACNT_1001"},
			"routes.com:*string:*req.RequestType:*rated": {"ROUTE_ACNT_1001"},
		},
		MissingFilters: map[string][]string{
			"FLTR_1_NOT_EXIST": {"ROUTE_ACNT_1001"},
		},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheSupplierFilterIndexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(rply.MissingObjects)
		sort.Strings(exp.MissingObjects)
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}
}

func TestHealthIndexDispatchers(t *testing.T) {
	Cache.Clear(nil)
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this dispatcherProfile but without indexing
	dspPrf := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "Dsp1",
		Subsystems: []string{utils.META_ANY, utils.MetaSessionS},
		FilterIDs: []string{
			"*string:~*opts.*apikey:dps1234;dsp9876",
			"*string:~*req.AnswerTime:2013-11-07T08:42:26Z",
			"*string:~*libphonenumber.<~*req.Destination>:+443234566",
			"*suffix:BrokenFilter:Invalid", // suffix will not be indexing
		},
		Strategy: utils.MetaRandom,
		Weight:   20,
		Hosts: DispatcherHostProfiles{
			{
				ID: "ALL",
			},
		},
	}
	if err := dm.SetDispatcherProfile(dspPrf, false); err != nil {
		t.Error(err)
	}

	args := &IndexHealthArgsWith3Ch{}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*any:*string:~*opts.*apikey:dps1234":                               {"Dsp1"},
			"cgrates.org:*any:*string:~*opts.*apikey:dsp9876":                               {"Dsp1"},
			"cgrates.org:*any:*string:~*req.AnswerTime:2013-11-07T08:42:26Z":                {"Dsp1"},
			"cgrates.org:*any:*string:~*libphonenumber.<~*req.Destination>:+443234566":      {"Dsp1"},
			"cgrates.org:*sessions:*string:~*opts.*apikey:dps1234":                          {"Dsp1"},
			"cgrates.org:*sessions:*string:~*opts.*apikey:dsp9876":                          {"Dsp1"},
			"cgrates.org:*sessions:*string:~*req.AnswerTime:2013-11-07T08:42:26Z":           {"Dsp1"},
			"cgrates.org:*sessions:*string:~*libphonenumber.<~*req.Destination>:+443234566": {"Dsp1"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
	}

	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheDispatcherFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	indexes := map[string]utils.StringMap{
		"*string:*req.RequestType:*rated": { // obj exist but the index don't
			"Dsp1": true,
			"Dsp2": true,
		},
		"*suffix:*opts.Destination:+100": { // obj exist but the index don't
			"Dsp1": true,
			"Dsp2": true,
		},
		"*string:*req.ExtraField:Usage": { // index is valid but the obj does not exist
			"InexistingDispatcher":  true,
			"InexistingDispatcher2": true,
		},
	}
	if err := dm.SetFilterIndexes(utils.CacheDispatcherFilterIndexes, "cgrates.org",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	//get the newIdxHealth for dispatchersProfile
	exp = &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*any:*string:~*opts.*apikey:dps1234":                               {"Dsp1"},
			"cgrates.org:*any:*string:~*opts.*apikey:dsp9876":                               {"Dsp1"},
			"cgrates.org:*any:*string:~*req.AnswerTime:2013-11-07T08:42:26Z":                {"Dsp1"},
			"cgrates.org:*any:*string:~*libphonenumber.<~*req.Destination>:+443234566":      {"Dsp1"},
			"cgrates.org:*sessions:*string:~*opts.*apikey:dps1234":                          {"Dsp1"},
			"cgrates.org:*sessions:*string:~*opts.*apikey:dsp9876":                          {"Dsp1"},
			"cgrates.org:*sessions:*string:~*req.AnswerTime:2013-11-07T08:42:26Z":           {"Dsp1"},
			"cgrates.org:*sessions:*string:~*libphonenumber.<~*req.Destination>:+443234566": {"Dsp1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*suffix:*opts.Destination:+100":  {"Dsp1"},
			"cgrates.org:*string:*req.RequestType:*rated": {"Dsp1"},
		},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{
			"cgrates.org:Dsp2",
			"cgrates.org:InexistingDispatcher",
			"cgrates.org:InexistingDispatcher2",
		},
	}

	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheDispatcherFilterIndexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(rply.MissingObjects)
		sort.Strings(exp.MissingObjects)
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}

	//we will use an inexisting Filter(not inline) for the same DispatcherProfile
	dspPrf = &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "Dsp1",
		Subsystems: []string{utils.META_ANY, utils.MetaSessionS},
		FilterIDs: []string{
			"*string:~*opts.*apikey:dps1234;dsp9876",
			"*string:~*req.AnswerTime:2013-11-07T08:42:26Z",
			"*string:~*libphonenumber.<~*req.Destination>:+443234566",
			"*suffix:BrokenFilter:Invalid", // suffix will not be indexing
			"FLTR_1_NOT_EXIST",
		},
		Strategy: utils.MetaRandom,
		Weight:   20,
		Hosts: DispatcherHostProfiles{
			{
				ID: "ALL",
			},
		},
	}
	if err := dm.SetDispatcherProfile(dspPrf, false); err != nil {
		t.Error(err)
	}

	//get the newIdxHealth for dispatchersProfile
	exp = &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*any:*string:~*opts.*apikey:dps1234":                               {"Dsp1"},
			"cgrates.org:*any:*string:~*opts.*apikey:dsp9876":                               {"Dsp1"},
			"cgrates.org:*any:*string:~*req.AnswerTime:2013-11-07T08:42:26Z":                {"Dsp1"},
			"cgrates.org:*any:*string:~*libphonenumber.<~*req.Destination>:+443234566":      {"Dsp1"},
			"cgrates.org:*sessions:*string:~*opts.*apikey:dps1234":                          {"Dsp1"},
			"cgrates.org:*sessions:*string:~*opts.*apikey:dsp9876":                          {"Dsp1"},
			"cgrates.org:*sessions:*string:~*req.AnswerTime:2013-11-07T08:42:26Z":           {"Dsp1"},
			"cgrates.org:*sessions:*string:~*libphonenumber.<~*req.Destination>:+443234566": {"Dsp1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*suffix:*opts.Destination:+100":  {"Dsp1"},
			"cgrates.org:*string:*req.RequestType:*rated": {"Dsp1"},
		},
		MissingFilters: map[string][]string{
			"FLTR_1_NOT_EXIST": {"Dsp1"},
		},
		MissingObjects: []string{
			"cgrates.org:Dsp2",
			"cgrates.org:InexistingDispatcher",
			"cgrates.org:InexistingDispatcher2",
		},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheDispatcherFilterIndexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(rply.MissingObjects)
		sort.Strings(exp.MissingObjects)
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}
}

func TestIndexHealthMultipleProfiles(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this multiple chargers but without indexing(same and different indexes)
	chPrf1 := &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Raw",
		FilterIDs: []string{
			"*string:~*opts.*eventType:ChargerAccountUpdate",
			"*string:~*req.Account:1234",
			"*string:~*asm.ID:1002",
			"*suffix:BrokenFilter:Invalid"}, // suffix will not be indexing
		RunID:        "raw",
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       20,
	}
	chPrf2 := &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Default",
		FilterIDs: []string{
			"*string:~*opts.*eventType:ChargerAccountUpdate",
			"*prefix:~*req.Destination:+2234;~*req.CGRID",
			"*prefix:~*req.Usage:10",
			"*string:~*req.Account:1234",
			"FLTR_1_NOT_EXIST2",
		},
		RunID:  "*default",
		Weight: 10,
	}
	chPrf3 := &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Call_Attr1",
		FilterIDs: []string{
			"*string:~*req.Account:1234",
			"*string:*broken:index",
			"FLTR_1_NOT_EXIST",
			"FLTR_1_NOT_EXIST2",
		},
		AttributeIDs: []string{"Attr1"},
		RunID:        "*attribute",
		Weight:       0,
	}
	if err := dm.SetChargerProfile(chPrf1, false); err != nil {
		t.Error(err)
	}
	if err := dm.SetChargerProfile(chPrf2, false); err != nil {
		t.Error(err)
	}
	if err := dm.SetChargerProfile(chPrf3, false); err != nil {
		t.Error(err)
	}

	// check the indexes health
	args := &IndexHealthArgsWith3Ch{}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:~*asm.ID:1002":                          {"Raw"},
			"cgrates.org:*string:~*opts.*eventType:ChargerAccountUpdate": {"Default", "Raw"},
			"cgrates.org:*string:~*req.Account:1234":                     {"Call_Attr1", "Default","Raw"},
			"cgrates.org:*prefix:~*req.Destination:+2234":                {"Default"},
			"cgrates.org:*prefix:~*req.Destination:~*req.CGRID":          {"Default"},
			"cgrates.org:*prefix:~*req.Usage:10":                         {"Default"},
			"cgrates.org:*string:*broken:index":                          {"Call_Attr1"},
		},
		BrokenIndexes: map[string][]string{},
		MissingFilters: map[string][]string{
			"FLTR_1_NOT_EXIST2": {"Call_Attr1", "Default"},
			"FLTR_1_NOT_EXIST":  {"Call_Attr1"},
		},
		MissingObjects: []string{},
	}
	if rply, err := GetFltrIdxHealth(dm,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheChargerFilterIndexes); err != nil {
		t.Error(err)
	} else {
		for _, slice := range rply.MissingIndexes{sort.Strings(slice)}
		for _, slice := range rply.MissingFilters{sort.Strings(slice)}
		for _, slice := range rply.BrokenIndexes{sort.Strings(slice)}
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}
}
