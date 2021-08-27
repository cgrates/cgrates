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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestHealthFilterAttributes(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	if err := dm.SetAttributeProfile(context.Background(), &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		FilterIDs: []string{"*string:~*req.Account:1001", "Fltr1"},
	}, false); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetIndexes(context.Background(), utils.CacheAttributeFilterIndexes, "cgrates.org",
		map[string]utils.StringSet{"*string:*req.Account:1002": {"ATTR1": {}, "ATTR2": {}}},
		true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*req.Account:1001": {"ATTR1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*string:*req.Account:1002": {"ATTR1"},
		},
		MissingFilters: map[string][]string{
			"cgrates.org:Fltr1": {"ATTR1"},
		},
		MissingObjects: []string{"cgrates.org:ATTR2"},
	}

	if rply, err := GetFltrIdxHealth(context.Background(), dm,
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

	if err := dm.SetAttributeProfile(context.Background(), &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		FilterIDs: []string{"*string:~*req.Account:1001", "Fltr1", "Fltr3"},
	}, false); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetFilter(context.Background(), &Filter{
		Tenant: "cgrates.org",
		ID:     "Fltr3",
	}, false); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetIndexes(context.Background(), utils.CacheReverseFilterIndexes, "cgrates.org:Fltr2",
		map[string]utils.StringSet{utils.CacheAttributeFilterIndexes: {"ATTR1": {}, "ATTR2": {}}},
		true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetRateProfile(context.Background(), &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID:        "RT1",
				FilterIDs: []string{"Fltr3"},
			},
		},
	}, false); err != nil {
		t.Fatal(err)
	}

	exp := map[string]*ReverseFilterIHReply{
		utils.CacheAttributeFilterIndexes: {
			MissingReverseIndexes: map[string][]string{
				"cgrates.org:ATTR1": {"Fltr3"},
			},
			MissingFilters: map[string][]string{
				"cgrates.org:Fltr1": {"ATTR1"},
			},
			BrokenReverseIndexes: map[string][]string{
				"cgrates.org:ATTR1": {"Fltr2"},
			},
			MissingObjects: []string{"cgrates.org:ATTR2"},
		},
		utils.CacheRateFilterIndexes: {
			MissingReverseIndexes: map[string][]string{
				"cgrates.org:RP1:RT1": {"Fltr3"},
			},
			BrokenReverseIndexes: make(map[string][]string),
			MissingFilters:       make(map[string][]string),
		},
	}
	objCaches := make(map[string]*ltcache.Cache)
	for indxType := range utils.CacheIndexesToPrefix {
		objCaches[indxType] = ltcache.NewCache(-1, 0, false, nil)
	}
	objCaches[utils.CacheRateFilterIndexes] = ltcache.NewCache(-1, 0, false, nil)
	if rply, err := GetRevFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(-1, 0, false, nil),
		ltcache.NewCache(-1, 0, false, nil),
		objCaches); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
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
				"*string:~*opts.ID:1002",
				"*suffix:BrokenFilter:Invalid", // this will not be indexed
			},
			MaxHits: 1,
		},
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf.ThresholdProfile, false); err != nil {
		t.Error(err)
	}

	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.ID:1002":                  {"TestHealthIndexThreshold"},
			"cgrates.org:*string:*opts.*eventType:AccountUpdate": {"TestHealthIndexThreshold"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
	}
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
	if err := dm.SetIndexes(context.Background(), utils.CacheThresholdFilterIndexes, "cgrates.org",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingThreshold"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.ID:1002":                  {"TestHealthIndexThreshold"},
			"cgrates.org:*string:*opts.*eventType:AccountUpdate": {"TestHealthIndexThreshold"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*prefix:req.InvalidIdx:10": {"TestHealthIndexThreshold"},
		},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
				"*string:~*opts.ID:1002",
				"FLTR_1_DOES_NOT_EXIST"},
			MaxHits: 1,
		},
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf.ThresholdProfile, false); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{"cgrates.org:InexistingThreshold"},
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.ID:1002":                  {"TestHealthIndexThreshold"},
			"cgrates.org:*string:*opts.*eventType:AccountUpdate": {"TestHealthIndexThreshold"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*prefix:req.InvalidIdx:10": {"TestHealthIndexThreshold"},
		},
		MissingFilters: map[string][]string{
			"cgrates.org:FLTR_1_DOES_NOT_EXIST": {"TestHealthIndexThreshold"},
		},
	}
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
			"*suffix:BrokenFilter:Invalid"}, // invalid (2 static types)
		RunID:        "raw",
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       20,
	}
	if err := dm.SetChargerProfile(context.Background(), chPrf, false); err != nil {
		t.Error(err)
	}

	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*eventType:ChargerAccountUpdate": {"Raw"},
			"cgrates.org:*string:*req.*Account:1234":                    {"Raw"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
	}
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
	if err := dm.SetIndexes(context.Background(), utils.CacheChargerFilterIndexes, "cgrates.org",
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
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
			"*suffix:BrokenFilter:Invalid",
			"FLTR_1_DOES_NOT_EXIST_CHRGR"},
		RunID:        "raw",
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       20,
	}
	if err := dm.SetChargerProfile(context.Background(), chPrf, false); err != nil {
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
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
	if err := dm.SetResourceProfile(context.Background(), rsPrf, false); err != nil {
		t.Error(err)
	}

	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"tenant.custom:*string:*opts.*eventType:ResourceAccountUpdate": {"RES_GRP1"},
			"tenant.custom:*string:*req.RequestType:*rated":                {"RES_GRP1"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
	}
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
	if err := dm.SetIndexes(context.Background(), utils.CacheResourceFilterIndexes, "tenant.custom",
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
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
	if err := dm.SetResourceProfile(context.Background(), rsPrf, false); err != nil {
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
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, false); err != nil {
		t.Error(err)
	}

	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*apikey:sts1234":      {"Stat_1"},
			"cgrates.org:*string:*req.RequestType:*postpaid": {"Stat_1"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
	}
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
	if err := dm.SetIndexes(context.Background(), utils.CacheStatFilterIndexes, "cgrates.org",
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
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
	if err := dm.SetStatQueueProfile(context.Background(), sqPrf, false); err != nil {
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
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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

	// we will set this routes but without indexing
	rtPrf := &RouteProfile{
		Tenant: "routes.com",
		ID:     "ROUTE_ACNT_1001",
		FilterIDs: []string{"*string:~*opts.*apikey:rts1234",
			"*string:~*req.Usage:160s",
			"*string:~*stats.STATS_VENDOR_2.*acd:1m", // *stats will not be indexing
			"*string:~*nothing.Denied:true",
			"*suffix:BrokenFilter:Invalid"},
		Sorting:           utils.MetaLC,
		SortingParameters: []string{},
		Routes: []*Route{
			{
				ID:              "route1",
				Weight:          10,
				Blocker:         false,
				RouteParameters: "",
			},
			{
				ID:             "route2",
				RateProfileIDs: []string{"RP_1002"},
				Weight:         20,
				Blocker:        false,
			},
		},
		Weight: 10,
	}
	if err := dm.SetRouteProfile(context.Background(), rtPrf, false); err != nil {
		t.Error(err)
	}

	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"routes.com:*string:*opts.*apikey:rts1234": {"ROUTE_ACNT_1001"},
			"routes.com:*string:*req.Usage:160s":       {"ROUTE_ACNT_1001"},
			"routes.com:*string:*nothing.Denied:true":  {"ROUTE_ACNT_1001"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
	}
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		utils.CacheRouteFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	indexes := map[string]utils.StringSet{
		"*string:*req.RequestType:*rated": { // obj exist but the index don't
			"ROUTE_ACNT_1001": {},
		},
		"*suffix:*req.Destination:+222": {
			"ROUTE_ACNT_1001": {},
			"ROUTE_ACNT_1002": {},
		},
		"*suffix:*req.Destination:+333": {
			"ROUTE_ACNT_1001": {},
			"ROUTE_ACNT_1002": {},
		},
		"*string:*req.ExtraField:Usage": { // index is valid but the obj does not exist
			"InexistingRoute": {},
		},
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	if err := dm.SetIndexes(context.Background(), utils.CacheRouteFilterIndexes, "routes.com",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{
			"routes.com:InexistingRoute",
			"routes.com:ROUTE_ACNT_1002",
		},
		MissingIndexes: map[string][]string{
			"routes.com:*string:*opts.*apikey:rts1234": {"ROUTE_ACNT_1001"},
			"routes.com:*string:*req.Usage:160s":       {"ROUTE_ACNT_1001"},
			"routes.com:*string:*nothing.Denied:true":  {"ROUTE_ACNT_1001"},
		},
		BrokenIndexes: map[string][]string{
			"routes.com:*suffix:*req.Destination:+222":   {"ROUTE_ACNT_1001"},
			"routes.com:*suffix:*req.Destination:+333":   {"ROUTE_ACNT_1001"},
			"routes.com:*string:*req.RequestType:*rated": {"ROUTE_ACNT_1001"},
		},
		MissingFilters: map[string][]string{},
	}
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		utils.CacheRouteFilterIndexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(rply.MissingObjects)
		sort.Strings(exp.MissingObjects)
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}

	//we will use an inexisting Filter(not inline) for the same RouteProfile
	rtPrf = &RouteProfile{
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
		Routes: []*Route{
			{
				ID:              "route1",
				Weight:          10,
				Blocker:         false,
				RouteParameters: "",
			},
			{
				ID:             "route2",
				RateProfileIDs: []string{"RP_1002"},
				Weight:         20,
				Blocker:        false,
			},
		},
		Weight: 10,
	}
	if err := dm.SetRouteProfile(context.Background(), rtPrf, false); err != nil {
		t.Error(err)
	}
	exp = &FilterIHReply{
		MissingObjects: []string{
			"routes.com:InexistingRoute",
			"routes.com:ROUTE_ACNT_1002",
		},
		MissingIndexes: map[string][]string{
			"routes.com:*string:*opts.*apikey:rts1234": {"ROUTE_ACNT_1001"},
			"routes.com:*string:*req.Usage:160s":       {"ROUTE_ACNT_1001"},
			"routes.com:*string:*nothing.Denied:true":  {"ROUTE_ACNT_1001"},
		},
		BrokenIndexes: map[string][]string{
			"routes.com:*suffix:*req.Destination:+222":   {"ROUTE_ACNT_1001"},
			"routes.com:*suffix:*req.Destination:+333":   {"ROUTE_ACNT_1001"},
			"routes.com:*string:*req.RequestType:*rated": {"ROUTE_ACNT_1001"},
		},
		MissingFilters: map[string][]string{
			"routes.com:FLTR_1_NOT_EXIST": {"ROUTE_ACNT_1001"},
		},
	}
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		utils.CacheRouteFilterIndexes); err != nil {
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
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this dispatcherProfile but without indexing
	dspPrf := &DispatcherProfile{
		Tenant: "cgrates.org",
		ID:     "Dsp1",
		FilterIDs: []string{
			"*string:~*opts.*apikey:dps1234|dsp9876",
			"*string:~*req.AnswerTime:2013-11-07T08:42:26Z",
			"*string:~*libphonenumber.<~*req.Destination>:+443234566", // *libphonenumber will not be indexing
			"*suffix:BrokenFilter:Invalid",
		},
		Strategy: utils.MetaRandom,
		Weight:   20,
		Hosts: DispatcherHostProfiles{
			{
				ID: "ALL",
			},
		},
	}
	if err := dm.SetDispatcherProfile(context.Background(), dspPrf, false); err != nil {
		t.Error(err)
	}

	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*apikey:dps1234":                {"Dsp1"},
			"cgrates.org:*string:*opts.*apikey:dsp9876":                {"Dsp1"},
			"cgrates.org:*string:*req.AnswerTime:2013-11-07T08:42:26Z": {"Dsp1"},
		},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
		MissingObjects: []string{},
	}

	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		utils.CacheDispatcherFilterIndexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	// we will set manually some indexes that points to an nil object or index is valid but the obj is missing
	indexes := map[string]utils.StringSet{
		"*string:*req.RequestType:*rated": { // obj exist but the index don't
			"Dsp1": {},
			"Dsp2": {},
		},
		"*suffix:*opts.Destination:+100": { // obj exist but the index don't
			"Dsp1": {},
			"Dsp2": {},
		},
		"*string:*req.ExtraField:Usage": { // index is valid but the obj does not exist
			"InexistingDispatcher":  {},
			"InexistingDispatcher2": {},
		},
	}
	if err := dm.SetIndexes(context.Background(), utils.CacheDispatcherFilterIndexes, "cgrates.org",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	//get the newIdxHealth for dispatchersProfile
	exp = &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*apikey:dps1234":                {"Dsp1"},
			"cgrates.org:*string:*opts.*apikey:dsp9876":                {"Dsp1"},
			"cgrates.org:*string:*req.AnswerTime:2013-11-07T08:42:26Z": {"Dsp1"},
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

	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
		Tenant: "cgrates.org",
		ID:     "Dsp1",
		FilterIDs: []string{
			"*string:~*opts.*apikey:dps1234|dsp9876",
			"*string:~*req.AnswerTime:2013-11-07T08:42:26Z",
			"*string:~*libphonenumber.<~*req.Destination>:+443234566", // *libphonenumber will not be indexing
			"*suffix:BrokenFilter:Invalid",
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
	if err := dm.SetDispatcherProfile(context.Background(), dspPrf, false); err != nil {
		t.Error(err)
	}

	//get the newIdxHealth for dispatchersProfile
	exp = &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*apikey:dps1234":                {"Dsp1"},
			"cgrates.org:*string:*opts.*apikey:dsp9876":                {"Dsp1"},
			"cgrates.org:*string:*req.AnswerTime:2013-11-07T08:42:26Z": {"Dsp1"},
		},
		BrokenIndexes: map[string][]string{
			"cgrates.org:*suffix:*opts.Destination:+100":  {"Dsp1"},
			"cgrates.org:*string:*req.RequestType:*rated": {"Dsp1"},
		},
		MissingFilters: map[string][]string{
			"cgrates.org:FLTR_1_NOT_EXIST": {"Dsp1"},
		},
		MissingObjects: []string{
			"cgrates.org:Dsp2",
			"cgrates.org:InexistingDispatcher",
			"cgrates.org:InexistingDispatcher2",
		},
	}
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
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
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this multiple chargers but without indexing(same and different indexes)
	chPrf1 := &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Raw",
		FilterIDs: []string{
			"*string:~*opts.*eventType:ChargerAccountUpdate",
			"*string:~*req.Account:1234",
			"*suffix:BrokenFilter:Invalid"},
		RunID:        "raw",
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       20,
	}
	chPrf2 := &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Default",
		FilterIDs: []string{
			"*string:~*opts.*eventType:ChargerAccountUpdate",
			"*prefix:~*req.Destination:+2234|~*req.CGRID",
			"*suffix:~*req.Usage:10",
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
	if err := dm.SetChargerProfile(context.Background(), chPrf1, false); err != nil {
		t.Error(err)
	}
	if err := dm.SetChargerProfile(context.Background(), chPrf2, false); err != nil {
		t.Error(err)
	}
	if err := dm.SetChargerProfile(context.Background(), chPrf3, false); err != nil {
		t.Error(err)
	}

	// check the indexes health
	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.*eventType:ChargerAccountUpdate": {"Raw", "Default"},
			"cgrates.org:*string:*req.Account:1234":                     {"Raw", "Default", "Call_Attr1"},
			"cgrates.org:*prefix:*req.Destination:+2234":                {"Default"},
			"cgrates.org:*suffix:*req.Usage:10":                         {"Default"},
		},
		BrokenIndexes: map[string][]string{},
		MissingFilters: map[string][]string{
			"cgrates.org:FLTR_1_NOT_EXIST2": {"Default", "Call_Attr1"},
			"cgrates.org:FLTR_1_NOT_EXIST":  {"Call_Attr1"},
		},
		MissingObjects: []string{},
	}
	if rply, err := GetFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		utils.CacheChargerFilterIndexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(exp.MissingIndexes["cgrates.org:*string:*opts.*eventType:ChargerAccountUpdate"])
		sort.Strings(exp.MissingIndexes["cgrates.org:*string:*req.Account:1234"])
		sort.Strings(exp.MissingFilters["cgrates.org:FLTR_1_NOT_EXIST2"])
		sort.Strings(rply.MissingIndexes["cgrates.org:*string:*opts.*eventType:ChargerAccountUpdate"])
		sort.Strings(rply.MissingIndexes["cgrates.org:*string:*req.Account:1234"])
		sort.Strings(rply.MissingFilters["cgrates.org:FLTR_1_NOT_EXIST2"])
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}
}

func TestIndexHealthReverseChecking(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	// we will set this multiple chargers but without indexing(same and different indexes)
	chPrf1 := &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Raw",
		FilterIDs: []string{
			"*string:~*req.Account:1234",
			"FLTR_1",
			"FLTR_2",
		},
		RunID:        "raw",
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       20,
	}
	if err := dm.SetChargerProfile(context.Background(), chPrf1, false); err != nil {
		t.Error(err)
	}
	// get cachePrefix for every subsystem
	objCaches := make(map[string]*ltcache.Cache)
	for indxType := range utils.CacheIndexesToPrefix {
		objCaches[indxType] = ltcache.NewCache(-1, 0, false, nil)
	}

	// reverse indexes for charger
	exp := map[string]*ReverseFilterIHReply{
		utils.CacheChargerFilterIndexes: {
			MissingFilters: map[string][]string{
				"cgrates.org:FLTR_1": {"Raw"},
				"cgrates.org:FLTR_2": {"Raw"},
			},
			BrokenReverseIndexes:  map[string][]string{},
			MissingReverseIndexes: map[string][]string{},
		},
	}
	if rply, err := GetRevFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		objCaches); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	// set reverse indexes for Raw that is already set and 2 that does not exist
	indexes := map[string]utils.StringSet{
		utils.CacheChargerFilterIndexes: {
			"Raw":        {},
			"Default":    {},
			"Call_Attr1": {},
		},
	}
	if err := dm.SetIndexes(context.Background(), utils.CacheReverseFilterIndexes, "cgrates.org:FLTR_1",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	// reverse indexes for charger with the changes
	exp = map[string]*ReverseFilterIHReply{
		utils.CacheChargerFilterIndexes: {
			MissingFilters: map[string][]string{
				"cgrates.org:FLTR_1": {"Raw"},
				"cgrates.org:FLTR_2": {"Raw"},
			},
			BrokenReverseIndexes:  map[string][]string{},
			MissingReverseIndexes: map[string][]string{},
			MissingObjects:        []string{"cgrates.org:Default", "cgrates.org:Call_Attr1"},
		},
	}
	if rply, err := GetRevFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(0, 0, false, nil),
		ltcache.NewCache(0, 0, false, nil),
		objCaches); err != nil {
		t.Fatal(err)
	} else {
		sort.Strings(exp[utils.CacheChargerFilterIndexes].MissingObjects)
		sort.Strings(rply[utils.CacheChargerFilterIndexes].MissingObjects)
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}

	// reverse for a filter present in PROFILE but does not exist for the same indexes
	indexes = map[string]utils.StringSet{
		utils.CacheChargerFilterIndexes: {
			"Raw":        {},
			"Default":    {},
			"Call_Attr1": {},
		},
	}
	if err := dm.SetIndexes(context.Background(), utils.CacheReverseFilterIndexes,
		"cgrates.org:FLTR_NOT_IN_PROFILE",
		indexes, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	// reverse indexes for charger with the changes
	exp = map[string]*ReverseFilterIHReply{
		utils.CacheChargerFilterIndexes: {
			MissingFilters: map[string][]string{
				"cgrates.org:FLTR_1": {"Raw"},
				"cgrates.org:FLTR_2": {"Raw"},
			},
			BrokenReverseIndexes: map[string][]string{
				"cgrates.org:Raw": {"FLTR_NOT_IN_PROFILE"},
			},
			MissingReverseIndexes: map[string][]string{},
			MissingObjects:        []string{"cgrates.org:Default", "cgrates.org:Call_Attr1"},
		},
	}
	if rply, err := GetRevFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(-1, 0, false, nil),
		ltcache.NewCache(-1, 0, false, nil),
		objCaches); err != nil {
		t.Fatal(err)
	} else {
		sort.Strings(exp[utils.CacheChargerFilterIndexes].MissingObjects)
		sort.Strings(rply[utils.CacheChargerFilterIndexes].MissingObjects)
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}
}

func TestIndexHealthMissingReverseIndexes(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	filter1 := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_1",
		Rules: []*FilterRule{
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			},
		},
	}
	filter2 := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
		Rules: []*FilterRule{
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
				Type:    utils.MetaString,
				Values:  []string{"1003"},
			},
		},
	}
	filter3 := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_3",
		Rules: []*FilterRule{
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
				Type:    utils.MetaString,
				Values:  []string{"1004"},
			},
		},
	}
	if err := dm.SetFilter(context.Background(), filter1, false); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(context.Background(), filter2, false); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(context.Background(), filter3, false); err != nil {
		t.Error(err)
	}

	// we will set this multiple chargers but without indexing(same and different indexes)
	chPrf1 := &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Raw",
		FilterIDs: []string{
			"*string:~*req.Account:1234",
			"FLTR_1",
			"FLTR_3",
		},
		RunID:        "raw",
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       20,
	}
	if err := dm.SetChargerProfile(context.Background(), chPrf1, false); err != nil {
		t.Error(err)
	}
	// get cachePrefix for every subsystem
	objCaches := make(map[string]*ltcache.Cache)
	for indxType := range utils.CacheIndexesToPrefix {
		objCaches[indxType] = ltcache.NewCache(-1, 0, false, nil)
	}

	exp := map[string]*ReverseFilterIHReply{
		utils.CacheChargerFilterIndexes: {
			MissingFilters:       map[string][]string{},
			BrokenReverseIndexes: map[string][]string{},
			MissingReverseIndexes: map[string][]string{
				"cgrates.org:Raw": {"FLTR_1", "FLTR_3"},
			},
		},
	}
	if rply, err := GetRevFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(-1, 0, false, nil),
		ltcache.NewCache(-1, 0, false, nil),
		objCaches); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}

	Cache.Clear(nil)
	if err := dm.SetFilter(context.Background(), filter1, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(context.Background(), filter2, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetFilter(context.Background(), filter3, true); err != nil {
		t.Error(err)
	}
	// initialized cachePrefix for every subsystem again for a new case
	objCaches = make(map[string]*ltcache.Cache)
	for indxType := range utils.CacheIndexesToPrefix {
		objCaches[indxType] = ltcache.NewCache(-1, 0, false, nil)
	}
	chPrf1 = &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "Raw",
		FilterIDs: []string{
			"*string:~*req.Account:1234",
			"FLTR_1",
			"FLTR_2",
			"FLTR_3",
		},
		RunID:        "raw",
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       20,
	}
	// now we set the charger with indexing and check the reverse indexes health
	if err := dm.SetChargerProfile(context.Background(), chPrf1, false); err != nil {
		t.Error(err)
	}
	exp = map[string]*ReverseFilterIHReply{
		utils.CacheChargerFilterIndexes: {
			MissingFilters:       map[string][]string{},
			BrokenReverseIndexes: map[string][]string{},
			MissingReverseIndexes: map[string][]string{
				"cgrates.org:Raw": {"FLTR_1", "FLTR_2", "FLTR_3"},
			},
		},
	}
	if rply, err := GetRevFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(-1, 0, false, nil),
		ltcache.NewCache(-1, 0, false, nil),
		objCaches); err != nil {
		t.Fatal(err)
	} else {
		sort.Strings(exp[utils.CacheChargerFilterIndexes].MissingReverseIndexes["cgrates.org:Raw"])
		sort.Strings(rply[utils.CacheChargerFilterIndexes].MissingReverseIndexes["cgrates.org:Raw"])
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}

	// delete the filters and check again the reverse index health
	if err := dm.RemoveFilter(context.Background(), "cgrates.org", "FLTR_1", true); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveFilter(context.Background(), "cgrates.org", "FLTR_2", true); err != nil {
		t.Error(err)
	}
	if err := dm.RemoveFilter(context.Background(), "cgrates.org", "FLTR_3", true); err != nil {
		t.Error(err)
	}
	// Nnow the exepcted should be on missing filters as those were removed lately
	exp = map[string]*ReverseFilterIHReply{
		utils.CacheChargerFilterIndexes: {
			MissingFilters: map[string][]string{
				"cgrates.org:FLTR_1": {"Raw"},
				"cgrates.org:FLTR_2": {"Raw"},
				"cgrates.org:FLTR_3": {"Raw"},
			},
			BrokenReverseIndexes:  map[string][]string{},
			MissingReverseIndexes: map[string][]string{},
		},
	}
	if rply, err := GetRevFltrIdxHealth(context.Background(), dm,
		ltcache.NewCache(-1, 0, false, nil),
		ltcache.NewCache(-1, 0, false, nil),
		objCaches); err != nil {
		t.Fatal(err)
	} else {
		sort.Strings(exp[utils.CacheChargerFilterIndexes].MissingReverseIndexes["cgrates.org:Raw"])
		sort.Strings(rply[utils.CacheChargerFilterIndexes].MissingReverseIndexes["cgrates.org:Raw"])
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(rply))
		}
	}
}
