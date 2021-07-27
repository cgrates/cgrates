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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestHealthFilter(t *testing.T) {
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
				"*suffix:BrokenFilter:Invalid"},
			MaxHits: 1,
		},
	}
	if err := dm.SetThresholdProfile(context.Background(), thPrf.ThresholdProfile, false); err != nil {
		t.Error(err)
	}

	exp := &FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*string:*opts.ID:1002":                   {"TestHealthIndexThreshold"},
			"cgrates.org:*string:*opts.*eventType:AccountUpdate": {"TestHealthIndexThreshold"},
		},
		BrokenIndexes:  map[string][]string{},
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
			"cgrates.org:*string:*opts.ID:1002":                   {"TestHealthIndexThreshold"},
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
			"cgrates.org:*string:*opts.ID:1002":                   {"TestHealthIndexThreshold"},
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
