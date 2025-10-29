/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestV1LoadCache(t *testing.T) {
	defer func() {
		InitCache(nil)
	}()
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	args := utils.AttrReloadCacheWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{},
		AttrReloadCache: utils.AttrReloadCache{
			ArgsCache: utils.ArgsCache{
				ThresholdIDs: []string{"THD1"},
			},
		},
	}
	loadIds := map[string]int64{
		utils.CacheThresholdProfiles: time.Now().Unix(),
		utils.CacheStatQueueProfiles: time.Now().Unix(),
		utils.CacheChargerProfiles:   time.Now().Unix(),
	}
	if err := dm.SetLoadIDs(loadIds); err != nil {
		t.Error(err)
	}
	cacheS := NewCacheS(cfg, dm)
	var reply string
	if err := cacheS.V1LoadCache(args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expectd ok ,received %v", reply)
	}
}

func TestCacheV1ReloadCache(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		InitCache(nil)
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := NewCacheS(cfg, dm)
	attrs := utils.AttrReloadCacheWithArgDispatcher{
		AttrReloadCache: utils.AttrReloadCache{
			ArgsCache: utils.ArgsCache{
				FilterIDs:             []string{"DSP_FLT"},
				DestinationIDs:        []string{"CRUDDestination2"},
				ReverseDestinationIDs: []string{"CRUDDestination2"},
				RatingPlanIDs:         []string{"RP_DFLT"},
				RatingProfileIDs:      []string{"*out:cgrates:call:1001"},
			},
		},
	}
	dst := &Destination{Id: "CRUDDestination2", Prefixes: []string{"+491", "+492", "+493"}}
	dm.SetDestination(dst, utils.NonTransactional)
	dm.SetReverseDestination(dst, utils.NonTransactional)
	rp := &RatingPlan{
		Id: "RP_DFLT",
		Timings: map[string]*RITiming{
			"30eab301": {
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f861": {
				Rates: []*Rate{
					{
						GroupIntervalStart: 0,
						Value:              0.01,
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			"CRUDDestination2": []*RPRate{
				{
					Timing: "30eab301",
					Rating: "b457f861",
					Weight: 10,
				},
			},
		},
	}
	dm.SetRatingPlan(rp, utils.NonTransactional)
	rP := &RatingProfile{Id: utils.ConcatenatedKey(utils.META_OUT, "cgrates", "call", "1001"),
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:   rp.Id,
		}},
	}
	dm.SetRatingProfile(rP, utils.NonTransactional)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "DSP_FLT",
		Rules: []*FilterRule{
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"2009"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	if _, err := GetFilter(dm, "cgrates.org", "DSP_FLT", false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	var reply string
	if err := chS.V1ReloadCache(attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected ok ,received %v", reply)
	}
}

func TestCacheSV1FlushCache(t *testing.T) {
	defer func() {
		InitCache(nil)
	}()
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := NewCacheS(cfg, dm)
	args := utils.AttrReloadCacheWithArgDispatcher{
		AttrReloadCache: utils.AttrReloadCache{
			ArgsCache: utils.ArgsCache{
				DestinationIDs: []string{"DST1"},
				RatingPlanIDs:  []string{"RP1"},
				ResourceIDs:    []string{"RSC1"},
			},
		},
	}

	loadIds := map[string]int64{
		utils.CacheAttributeProfiles: time.Now().UnixNano(),
		utils.CacheSupplierProfiles:  time.Now().UnixNano(),
	}
	if err := dm.SetLoadIDs(loadIds); err != nil {
		t.Error(err)
	}
	var reply string
	if err := chS.V1FlushCache(args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected ok,received %v", reply)
	}
	for id := range loadIds {
		if _, has := Cache.Get(utils.CacheLoadIDs, id); !has {
			t.Error("Load not stored in cache")
		}
	}
}

func TestLoadDataDbCache(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	rp := &RatingPlan{
		Id:      "RT_PLAN1",
		Timings: map[string]*RITiming{},
		Ratings: map[string]*RIRate{
			"asjkilj": {
				ConnectFee:       10,
				RoundingMethod:   utils.ROUNDING_UP,
				RoundingDecimals: 1,
				MaxCost:          10,
			},
		},
		DestinationRates: map[string]RPRateList{},
	}
	if err := dm.SetRatingPlan(rp, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := dm.PreloadCacheForPrefix(utils.RATING_PLAN_PREFIX); err != nil {
		t.Error(nil)
	}
}

func TestPreCacheStatus(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := NewCacheS(cfg, dm)
	args := &utils.AttrCacheIDsWithArgDispatcher{
		CacheIDs: []string{utils.CacheChargerProfiles, utils.CacheDispatcherHosts},
	}
	go func() {
		chS.pcItems[utils.CacheChargerProfiles] <- struct{}{}
	}()
	time.Sleep(5 * time.Millisecond)
	exp := map[string]string{
		utils.CacheChargerProfiles: utils.MetaReady,
		utils.CacheDispatcherHosts: utils.MetaPrecaching,
	}
	var reply map[string]string
	if err := chS.V1PrecacheStatus(args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestCachesRPCCall(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := NewCacheS(cfg, dm)
	Cache.Clear(nil)
	Cache.Set(utils.CacheThresholds, "cgrates:TH1", &Threshold{}, []string{}, true, utils.NonTransactional)
	args := &utils.ArgsGetCacheItemIDsWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{},
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheThresholds,
		},
	}
	var reply []string
	if err := chS.Call(context.Background(), utils.CacheSv1GetItemIDs, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, []string{"cgrates:TH1"}) {
		t.Errorf("Expected %v", []string{"cgrates:TH1"})
	}
}

func TestCacheSPreCache(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.CacheCfg()[utils.CacheAttributeProfiles].Precache = true
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	chS := NewCacheS(cfg, dm)
	Cache.Clear(nil)
	if err := chS.Precache(); err != nil {
		t.Error(err)
	}
}

func TestCacheV1HasItem(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	chS := NewCacheS(cfg, dm)
	args := &utils.ArgsGetCacheItemWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheThresholds,
			ItemID:  "cgrates.org:TH1",
		},
	}
	thd := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}
	Cache.Set(utils.CacheThresholds, utils.ConcatenatedKey(thd.Tenant, thd.ID), thd, []string{}, true, utils.NonTransactional)
	var reply bool
	if err := chS.V1HasItem(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestCacheV1GetItemExpiryTime(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.CacheCfg()[utils.CacheThresholds].TTL = 24 * time.Hour
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	chS := NewCacheS(cfg, dm)
	args := &utils.ArgsGetCacheItemWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheThresholds,
			ItemID:  "cgrates.org:TH1",
		},
	}
	thd := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Hits:   2,
	}
	Cache.Set(utils.CacheThresholds, utils.ConcatenatedKey(thd.Tenant, thd.ID), thd, []string{}, true, utils.NonTransactional)
	var reply time.Time

	if err := chS.V1GetItemExpiryTime(args, &reply); err != nil {
		t.Error(err)
	} else if reply.Day() != time.Now().Add(24*time.Hour).Day() {
		t.Error("Expected 1 Day ttl")
	}

}

func TestCachesV1GetItemIDs(t *testing.T) {
	c := CacheS{}

	err := c.V1GetItemIDs(&utils.ArgsGetCacheItemIDsWithArgDispatcher{}, &[]string{})

	if err != nil {
		if err.Error() != "NOT_FOUND" {
			t.Error(err)
		}
	}
}

func TestCachesV1GetItemExpiryTime(t *testing.T) {
	c := CacheS{}
	tm := time.Now()

	err := c.V1GetItemExpiryTime(&utils.ArgsGetCacheItemWithArgDispatcher{}, &tm)

	if err != nil {
		if err.Error() != "NOT_FOUND" {
			t.Error(err)
		}
	}
}

func TestV1RemoveItem(t *testing.T) {
	ch := CacheS{}
	str := "test"

	err := ch.V1RemoveItem(&utils.ArgsGetCacheItemWithArgDispatcher{}, &str)

	if err != nil {
		t.Error(err)
	}

	if str != "OK" {
		t.Error(str)
	}
}

func TestV1Clear(t *testing.T) {
	ch := CacheS{}
	str := "test"

	err := ch.V1Clear(&utils.AttrCacheIDsWithArgDispatcher{}, &str)

	if err != nil {
		t.Error(err)
	}

	if str != "OK" {
		t.Error(str)
	}
}

func TestV1GetCacheStats(t *testing.T) {
	ch := CacheS{}
	cs := map[string]*ltcache.CacheStats{}

	err := ch.V1GetCacheStats(&utils.AttrCacheIDsWithArgDispatcher{}, &cs)

	if err != nil {
		t.Error(err)
	}

	if cs == nil {
		t.Error("didnt get cache stats")
	}
}

func TestV1PrecacheStatus(t *testing.T) {
	c := CacheS{}

	err := c.V1PrecacheStatus(&utils.AttrCacheIDsWithArgDispatcher{}, &map[string]string{})

	if err == nil {
		t.Error("didn't receive an error")
	}
}

func TestCacheV1HasGroup(t *testing.T) {
	c := CacheS{}
	bl := true
	err := c.V1HasGroup(&utils.ArgsGetGroupWithArgDispatcher{}, &bl)

	if err != nil {
		if err.Error() != "unknown cacheID: *resources" {
			t.Error(err)
		}
	}
}

func TestV1GetGroupItemIDs(t *testing.T) {
	c := CacheS{}
	slc := []string{"test"}

	err := c.V1GetGroupItemIDs(&utils.ArgsGetGroupWithArgDispatcher{}, &slc)

	if err.Error() != "NOT_FOUND" {
		t.Error(err)
	}
}

func TestV1RemoveGroup(t *testing.T) {
	c := CacheS{}
	str := ""

	err := c.V1RemoveGroup(&utils.ArgsGetGroupWithArgDispatcher{}, &str)

	if err != nil {
		t.Error(err)
	}
}

func TestCachetoStringSlice(t *testing.T) {
	s := []string{}

	slc := toStringSlice(&s)

	if !reflect.DeepEqual(s, slc) {
		t.Errorf("received %v, expected %v", slc, s)
	}

	slc = toStringSlice(nil)

	if slc != nil {
		t.Error(slc)
	}
}

func TestCacheSV1ReloadCacheErrors(t *testing.T) {
	str := "test"
	slc := []string{str}
	type args struct {
		attrs utils.AttrReloadCacheWithArgDispatcher
		reply *string
	}

	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	chS := CacheS{
		cfg: &config.CGRConfig{},
		dm:  dm,
	}

	tests := []struct {
		name string
		args args
		err  string
	}{
		{
			name: "FlushAll",
			args: args{attrs: utils.AttrReloadCacheWithArgDispatcher{
				ArgDispatcher: &utils.ArgDispatcher{},
				TenantArg:     utils.TenantArg{},
				AttrReloadCache: utils.AttrReloadCache{
					FlushAll: true,
				},
			}, reply: &str},
			err: "",
		},
		{
			name: "Reload destination error",
			args: args{attrs: utils.AttrReloadCacheWithArgDispatcher{
				ArgDispatcher: &utils.ArgDispatcher{},
				TenantArg:     utils.TenantArg{},
				AttrReloadCache: utils.AttrReloadCache{
					ArgsCache: utils.ArgsCache{
						DestinationIDs:        slc,
						ReverseDestinationIDs: slc,
						RatingPlanIDs:         slc,
						RatingProfileIDs:      slc,
						ActionIDs:             slc,
						ActionPlanIDs:         slc,
						AccountActionPlanIDs:  slc,
						ActionTriggerIDs:      slc,
						SharedGroupIDs:        slc,
						ResourceProfileIDs:    slc,
						ResourceIDs:           slc,
						StatsQueueIDs:         slc,
						StatsQueueProfileIDs:  slc,
						ThresholdIDs:          slc,
						ThresholdProfileIDs:   slc,
						FilterIDs:             slc,
						SupplierProfileIDs:    slc,
						AttributeProfileIDs:   slc,
						ChargerProfileIDs:     slc,
						DispatcherProfileIDs:  slc,
						DispatcherHostIDs:     slc,
						DispatcherRoutesIDs:   slc,
					},
					FlushAll: false,
				},
			}, reply: &str},
			err: "err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := chS.V1ReloadCache(tt.args.attrs, tt.args.reply)

			if err != nil {
				if err.Error() != tt.err {
					t.Error(err)
				}
			}
		})
	}
}

func TestCacheSV1LoadCache(t *testing.T) {
	str := "test"
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	chS := CacheS{
		cfg: &config.CGRConfig{},
		dm:  dm,
	}

	err := chS.V1LoadCache(utils.AttrReloadCacheWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{},
		TenantArg:     utils.TenantArg{},
		AttrReloadCache: utils.AttrReloadCache{
			FlushAll: true,
		},
	}, &str)

	if err != nil {
		t.Error(err)
	}
}

func TestCacheSV1FlushCacheFlushAll(t *testing.T) {
	str := "test"
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	chS := CacheS{
		cfg: &config.CGRConfig{},
		dm:  dm,
	}

	err := chS.V1FlushCache(utils.AttrReloadCacheWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{},
		TenantArg:     utils.TenantArg{},
		AttrReloadCache: utils.AttrReloadCache{
			FlushAll: true,
		},
	}, &str)

	if err != nil {
		t.Error(err)
	}
}
