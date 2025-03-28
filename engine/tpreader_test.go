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
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestTPReaderCallCacheNoCaching(t *testing.T) {
	tmp := connMgr
	defer func() {
		connMgr = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	cM := NewConnManager(cfg, nil)
	args := map[string][]string{
		utils.CacheFilters:   {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
		utils.CacheResources: {},
	}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
	}

	err := CallCache(cM, []string{}, utils.MetaNone, args, []string{}, opts, true, "cgrates.org")

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

}

func TestTPReaderCallCacheReloadCacheFirstCallErr(t *testing.T) {
	tmp := connMgr
	defer func() {
		connMgr = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan birpc.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReloadCache: func(ctx *context.Context, args, reply any) error {
				expArgs := &utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]any{
						utils.MetaSubsys: utils.MetaChargers,
					},
					FilterIDs: []string{"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
					Tenant:    "cgrates.org",
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(expArgs), utils.ToJSON(args),
					)
				}
				return utils.ErrUnsupporteServiceMethod
			},
		},
	}
	client <- ccM

	cM := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaReload
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	explog := "Reloading cache\n"
	experr := utils.ErrUnsupporteServiceMethod
	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, true, "cgrates.org")

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	rcvlog := buf.String()[20:]
	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}
}

func TestTPReaderCallCacheReloadCacheSecondCallErr(t *testing.T) {
	tmp := connMgr
	defer func() {
		connMgr = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan birpc.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReloadCache: func(ctx *context.Context, args, reply any) error {
				return nil
			},
			utils.CacheSv1Clear: func(ctx *context.Context, args, reply any) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]any{
						utils.MetaSubsys: utils.MetaChargers,
					},
					CacheIDs: []string{"cacheID"},
					Tenant:   "cgrates.org",
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return utils.ErrUnsupporteServiceMethod
			},
		},
	}
	client <- ccM

	cM := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaReload
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	explog1 := "Reloading cache"
	explog2 := "Clearing indexes"
	experr := utils.ErrUnsupporteServiceMethod
	explog3 := fmt.Sprintf("WARNING: Got error on cache clear: %s\n", experr)
	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, true, "cgrates.org")

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	rcvlog1 := buf.String()[20 : 20+len(explog1)]
	if rcvlog1 != explog1 {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog1, rcvlog1)
	}

	rcvlog2 := buf.String()[41+len(rcvlog1) : 41+len(rcvlog1)+len(explog2)]
	if rcvlog2 != explog2 {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog2, rcvlog2)
	}

	rcvlog3 := buf.String()[62+len(rcvlog1)+len(explog2):]
	if rcvlog3 != explog3 {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog3, rcvlog3)
	}
}

func TestTPReaderCallCacheLoadCache(t *testing.T) {
	tmp := connMgr
	defer func() {
		connMgr = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan birpc.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1LoadCache: func(ctx *context.Context, args, reply any) error {
				expArgs := &utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]any{
						utils.MetaSubsys: utils.MetaChargers,
					},
					FilterIDs: []string{"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
					Tenant:    "cgrates.org",
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
			utils.CacheSv1Clear: func(ctx *context.Context, args, reply any) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]any{
						utils.MetaSubsys: utils.MetaChargers,
					},
					CacheIDs: []string{"cacheID"},
					Tenant:   "cgrates.org",
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
		},
	}
	client <- ccM

	cM := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaLoad
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
	}

	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, false, "cgrates.org")

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestTPReaderCallCacheRemoveItems(t *testing.T) {
	tmp := connMgr
	defer func() {
		connMgr = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan birpc.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1RemoveItems: func(ctx *context.Context, args, reply any) error {
				expArgs := &utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]any{
						utils.MetaSubsys: utils.MetaChargers,
					},
					FilterIDs: []string{"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
					Tenant:    "cgrates.org",
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
			utils.CacheSv1Clear: func(ctx *context.Context, args, reply any) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]any{
						utils.MetaSubsys: utils.MetaChargers,
					},
					CacheIDs: []string{"cacheID"},
					Tenant:   "cgrates.org",
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
		},
	}
	client <- ccM

	cM := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaRemove
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
	}

	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, false, "cgrates.org")

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestTPReaderCallCacheClear(t *testing.T) {
	tmp := connMgr
	defer func() {
		connMgr = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan birpc.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1Clear: func(ctx *context.Context, args, reply any) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]any{
						utils.MetaSubsys: utils.MetaChargers,
					},
					Tenant: "cgrates.org",
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
		},
	}
	client <- ccM

	cM := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaClear
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
	}

	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, false, "cgrates.org")

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestTPReaderGetLoadedIdsDestinations(t *testing.T) {
	tpr := &TpReader{
		destinations: map[string]*Destination{
			"1001": {
				Id:       "1001_ID",
				Prefixes: []string{"39", "40"},
			},
			"1002": {
				Id:       "1002_ID",
				Prefixes: []string{"75", "82"},
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.DestinationPrefix)
	if err != nil {
		t.Error(err)
	}
	//Test fails sometimes because of the order of the returned slice
	expRcv := []string{"1001", "1002"}
	sort.Strings(rcv)
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsReverseDestinations(t *testing.T) {
	tpr := &TpReader{
		destinations: map[string]*Destination{
			"1001": {
				Id:       "1001_ID",
				Prefixes: []string{"39", "75"},
			},
			"1002": {
				Id:       "1002_ID",
				Prefixes: []string{"87", "21"},
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.ReverseDestinationPrefix)
	if err != nil {
		t.Error(err)
	}
	sort.Strings(rcv)
	expRcv := []string{"21", "39", "75", "87"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsRatingPlans(t *testing.T) {
	tpr := &TpReader{
		ratingPlans: map[string]*RatingPlan{
			"RP_RETAIL1": {
				Id: "RP_1001",
				Timings: map[string]*RITiming{
					"TIMING_1001": {
						ID:         "PEAK",
						Years:      []int{2021},
						Months:     []time.Month{8},
						MonthDays:  []int{31},
						WeekDays:   []time.Weekday{5},
						StartTime:  "15:00:00Z",
						EndTime:    "17:00:00Z",
						cronString: "21 2 5 25 8 5 2021", //sec, min, hour, monthday, month, weekday, year
						tag:        utils.MetaAny,
					},
				},
				Ratings: map[string]*RIRate{
					"RATING_1001": {
						ConnectFee:       0.4,
						RoundingMethod:   utils.MetaRoundingUp,
						RoundingDecimals: 4,
						MaxCost:          0.60,
						MaxCostStrategy:  utils.MetaMaxCostDisconnect,
						Rates: []*RGRate{
							{
								GroupIntervalStart: 0,
								Value:              0.2,
								RateIncrement:      60 * time.Second,
								RateUnit:           60 * time.Second,
							},
						},
						tag: utils.MetaAny,
					},
				},
				DestinationRates: map[string]RPRateList{
					"DR_FS_40CNT": {
						{
							Timing: "TIMING_1001",
							Rating: "RATING_1001",
							Weight: 10,
						},
					},
				},
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.RatingPlanPrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := []string{"RP_RETAIL1"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsRatingProfiles(t *testing.T) {
	tpr := &TpReader{
		ratingProfiles: map[string]*RatingProfile{
			"1001": {
				Id: "RP_RETAIL1",
				RatingPlanActivations: []*RatingPlanActivation{
					{
						ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
						RatingPlanId:   "RETAIL1",
					},
				},
			},
			"1002": {
				Id: "RP_RETAIL2",
				RatingPlanActivations: []*RatingPlanActivation{
					{
						ActivationTime: time.Date(2014, 7, 21, 14, 15, 0, 0, time.UTC),
						RatingPlanId:   "RETAIL2",
					},
				},
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.RatingProfilePrefix)
	if err != nil {
		t.Error(err)
	}
	sort.Strings(rcv)
	expRcv := []string{"1001", "1002"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsActions(t *testing.T) {
	tpr := &TpReader{
		actions: map[string][]*Action{
			"TOPUP_RST_10": {
				{
					Id:               "ACTION_1001",
					ActionType:       utils.MetaSetBalance,
					ExtraParameters:  "",
					ExpirationString: "",
					Weight:           10,
					balanceValue:     9.45,
				},
			},
			"TOPUP_RST_5": {
				{
					Id:               "ACTION_1002",
					ActionType:       utils.MetaPublishAccount,
					ExtraParameters:  "",
					ExpirationString: "",
					Weight:           5,
					balanceValue:     12.32,
				},
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.ActionPrefix)
	if err != nil {
		t.Error(err)
	}
	sort.Strings(rcv)
	expRcv := []string{"TOPUP_RST_10", "TOPUP_RST_5"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsActionPlans(t *testing.T) {
	tpr := &TpReader{
		actionPlans: map[string]*ActionPlan{
			"PACKAGE_1001": {
				Id: "TOPUP_RST_10",
				AccountIDs: map[string]bool{
					"1001": true,
					"1002": false,
				},
			},
			"PACKAGE_1002": {
				Id: "TOPUP_RST_5",
				AccountIDs: map[string]bool{
					"1001": false,
					"1002": true,
				},
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.ActionPlanPrefix)
	if err != nil {
		t.Error(err)
	}
	sort.Strings(rcv)
	expRcv := []string{"PACKAGE_1001", "PACKAGE_1002"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsSharedGroup(t *testing.T) {
	tpr := &TpReader{
		sharedGroups: map[string]*SharedGroup{
			"SHARED_A": {
				Id: "SHARED_ID1",
				AccountParameters: map[string]*SharingParameters{
					"PARAM_1": {
						Strategy:      utils.MetaTopUp,
						RatingSubject: "1001",
					},
				},
				MemberIds: map[string]bool{
					"1001": true,
					"1002": false,
				},
			},
			"SHARED_B": {
				Id: "SHARED_ID2",
				AccountParameters: map[string]*SharingParameters{
					"PARAM_1": {
						Strategy:      utils.MetaTopUp,
						RatingSubject: "1002",
					},
				},
				MemberIds: map[string]bool{
					"1001": true,
					"1002": false,
				},
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.SharedGroupPrefix)
	if err != nil {
		t.Error(err)
	}
	sort.Strings(rcv)
	expRcv := []string{"SHARED_A", "SHARED_B"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsResourceProfiles(t *testing.T) {
	tpr := &TpReader{
		resProfiles: map[utils.TenantID]*utils.TPResourceProfile{
			{Tenant: "cgrates.org", ID: "ResGroup1"}: {
				TPid:      testTPID,
				Tenant:    "cgrates.org",
				ID:        "ResGroup1",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				ActivationInterval: &utils.TPActivationInterval{
					ActivationTime: "2014-07-29T15:00:00Z",
				},
				UsageTTL:          "1s",
				AllocationMessage: "call",
				Weight:            10,
				Limit:             "2",
				Blocker:           true,
				Stored:            true,
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.ResourceProfilesPrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := []string{"cgrates.org:ResGroup1"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsActionTriggers(t *testing.T) {
	tpr := &TpReader{
		actionsTriggers: map[string]ActionTriggers{
			"STANDARD_TRIGGERS": {
				{
					ID:             "ID1",
					UniqueID:       "",
					ThresholdType:  "*max_balance",
					ThresholdValue: 20,
					Recurrent:      false,
					MinSleep:       0,
					Weight:         10,
					ActionsID:      "LOG_WARNING",
				},
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.ActionTriggerPrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := []string{"STANDARD_TRIGGERS"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsStatQueueProfiles(t *testing.T) {
	tpr := &TpReader{
		sqProfiles: map[utils.TenantID]*utils.TPStatProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:      testTPID,
				Tenant:    "cgrates.org",
				ID:        "ResGroup1",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				ActivationInterval: &utils.TPActivationInterval{
					ActivationTime: "2014-07-29T15:00:00Z",
				},
				Weight:  10,
				Blocker: true,
				Stored:  true,
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.StatQueueProfilePrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := []string{"cgrates.org:cgratesID"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsThresholdProfiles(t *testing.T) {
	tpr := &TpReader{
		thProfiles: map[utils.TenantID]*utils.TPThresholdProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:      testTPID,
				Tenant:    "cgrates.org",
				ID:        "ResGroup1",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				ActivationInterval: &utils.TPActivationInterval{
					ActivationTime: "2014-07-29T15:00:00Z",
				},
				Weight:  10,
				Blocker: true,
				MaxHits: 3,
				MinHits: 1,
				Async:   true,
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.ThresholdProfilePrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := []string{"cgrates.org:cgratesID"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsFilters(t *testing.T) {
	tpr := &TpReader{
		filters: map[utils.TenantID]*utils.TPFilterProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:   testTPID,
				Tenant: "cgrates.org",
				ID:     "ResGroup1",
				ActivationInterval: &utils.TPActivationInterval{
					ActivationTime: "2014-07-29T15:00:00Z",
				},
				Filters: []*utils.TPFilter{
					{
						Type:    "~*req",
						Element: "Account",
						Values:  []string{"1001"},
					},
				},
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.FilterPrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := []string{"cgrates.org:cgratesID"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsRouteProfiles(t *testing.T) {
	tpr := &TpReader{
		routeProfiles: map[utils.TenantID]*utils.TPRouteProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:   testTPID,
				Tenant: "cgrates.org",
				ID:     "ResGroup1",
				ActivationInterval: &utils.TPActivationInterval{
					ActivationTime: "2014-07-29T15:00:00Z",
				},
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.RouteProfilePrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := []string{"cgrates.org:cgratesID"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsAttributeProfiles(t *testing.T) {
	tpr := &TpReader{
		attributeProfiles: map[utils.TenantID]*utils.TPAttributeProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:      testTPID,
				Tenant:    "cgrates.org",
				ID:        "ResGroup1",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				ActivationInterval: &utils.TPActivationInterval{
					ActivationTime: "2014-07-29T15:00:00Z",
				},
				Contexts: []string{"sessions"},
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.AttributeProfilePrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := []string{"cgrates.org:cgratesID"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsChargerProfiles(t *testing.T) {
	tpr := &TpReader{
		chargerProfiles: map[utils.TenantID]*utils.TPChargerProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:   testTPID,
				Tenant: "cgrates.org",
				ID:     "ResGroup1",
				ActivationInterval: &utils.TPActivationInterval{
					ActivationTime: "2014-07-29T15:00:00Z",
				},
				RunID: "RUN_ID",
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.ChargerProfilePrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := []string{"cgrates.org:cgratesID"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsDispatcherProfiles(t *testing.T) {
	tpr := &TpReader{
		dispatcherProfiles: map[utils.TenantID]*utils.TPDispatcherProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:   testTPID,
				Tenant: "cgrates.org",
				ID:     "ResGroup1",
				ActivationInterval: &utils.TPActivationInterval{
					ActivationTime: "2014-07-29T15:00:00Z",
				},
				Strategy: utils.MetaMaxCostDisconnect,
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.DispatcherProfilePrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := []string{"cgrates.org:cgratesID"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsEmptyObject(t *testing.T) {
	tpr := &TpReader{}
	rcv, err := tpr.GetLoadedIds(utils.DispatcherProfilePrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := make([]string, 0)
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsDispatcherHosts(t *testing.T) {
	tpr := &TpReader{
		dispatcherHosts: map[utils.TenantID]*utils.TPDispatcherHost{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:   testTPID,
				Tenant: "cgrates.org",
				ID:     "ResGroup1",
			},
		},
	}
	rcv, err := tpr.GetLoadedIds(utils.DispatcherHostPrefix)
	if err != nil {
		t.Error(err)
	}
	expRcv := []string{"cgrates.org:cgratesID"}
	if !reflect.DeepEqual(expRcv, rcv) {
		t.Errorf("\nExpected %v but received \n%v", expRcv, rcv)
	}
}

func TestTPReaderGetLoadedIdsError(t *testing.T) {
	tpr := &TpReader{}
	errExpect := "Unsupported load category"
	if _, err := tpr.GetLoadedIds(""); err == nil || err.Error() != errExpect {
		t.Errorf("\nExpected error %v but received \n%v", errExpect, err)
	}
}

func TestTPReaderReloadCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:               map[string]any{},
		Tenant:                "cgrates.org",
		ActionIDs:             []string{"ActionsID"},
		ActionPlanIDs:         []string{"ActionPlansID"},
		ActionTriggerIDs:      []string{"ActionTriggersID"},
		DestinationIDs:        []string{"DestinationsID"},
		TimingIDs:             []string{"TimingsID"},
		RatingPlanIDs:         []string{"RatingPlansID"},
		RatingProfileIDs:      []string{"RatingProfilesID"},
		SharedGroupIDs:        []string{"SharedGroupsID"},
		ResourceProfileIDs:    []string{"cgrates.org:resourceProfilesID"},
		StatsQueueProfileIDs:  []string{"cgrates.org:statProfilesID"},
		ThresholdProfileIDs:   []string{"cgrates.org:thresholdProfilesID"},
		FilterIDs:             []string{"cgrates.org:filtersID"},
		RouteProfileIDs:       []string{"cgrates.org:routeProfilesID"},
		AttributeProfileIDs:   []string{"cgrates.org:attributeProfilesID"},
		ChargerProfileIDs:     []string{"cgrates.org:chargerProfilesID"},
		DispatcherProfileIDs:  []string{"cgrates.org:dispatcherProfilesID"},
		DispatcherHostIDs:     []string{"cgrates.org:dispatcherHostsID"},
		ResourceIDs:           []string{"cgrates.org:resourceProfilesID"},
		StatsQueueIDs:         []string{"cgrates.org:statProfilesID"},
		ThresholdIDs:          []string{"cgrates.org:thresholdProfilesID"},
		AccountActionPlanIDs:  []string{"AccountActionPlansID"},
		RankingProfileIDs:     []string{"cgrates.org:rankingProfilesID"},
		ReverseDestinationIDs: []string{},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CacheSv1ReloadCache: func(ctx *context.Context, args any, reply any) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(ctx *context.Context, args any, reply any) error {
				return nil
			},
		},
	}
	tmp := connMgr
	defer func() { connMgr = tmp }()
	connMgr = NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	tpr := &TpReader{
		actions: map[string][]*Action{
			"ActionsID": {},
		},
		actionPlans: map[string]*ActionPlan{
			"ActionPlansID": {},
		},
		actionsTriggers: map[string]ActionTriggers{
			"ActionTriggersID": {},
		},
		destinations: map[string]*Destination{
			"DestinationsID": {},
		},
		timings: map[string]*utils.TPTiming{
			"TimingsID": {},
		},
		ratingPlans: map[string]*RatingPlan{
			"RatingPlansID": {},
		},
		ratingProfiles: map[string]*RatingProfile{
			"RatingProfilesID": {},
		},
		sharedGroups: map[string]*SharedGroup{
			"SharedGroupsID": {},
		},
		resProfiles: map[utils.TenantID]*utils.TPResourceProfile{
			{Tenant: "cgrates.org", ID: "resourceProfilesID"}: {},
		},
		sqProfiles: map[utils.TenantID]*utils.TPStatProfile{
			{Tenant: "cgrates.org", ID: "statProfilesID"}: {},
		},
		thProfiles: map[utils.TenantID]*utils.TPThresholdProfile{
			{Tenant: "cgrates.org", ID: "thresholdProfilesID"}: {},
		},
		filters: map[utils.TenantID]*utils.TPFilterProfile{
			{Tenant: "cgrates.org", ID: "filtersID"}: {},
		},
		routeProfiles: map[utils.TenantID]*utils.TPRouteProfile{
			{Tenant: "cgrates.org", ID: "routeProfilesID"}: {},
		},
		attributeProfiles: map[utils.TenantID]*utils.TPAttributeProfile{
			{Tenant: "cgrates.org", ID: "attributeProfilesID"}: {},
		},
		chargerProfiles: map[utils.TenantID]*utils.TPChargerProfile{
			{Tenant: "cgrates.org", ID: "chargerProfilesID"}: {},
		},
		dispatcherProfiles: map[utils.TenantID]*utils.TPDispatcherProfile{
			{Tenant: "cgrates.org", ID: "dispatcherProfilesID"}: {},
		},
		dispatcherHosts: map[utils.TenantID]*utils.TPDispatcherHost{
			{Tenant: "cgrates.org", ID: "dispatcherHostsID"}: {},
		},
		acntActionPlans: map[string][]string{
			"AccountActionPlansID": {},
		},
		rgProfiles: map[utils.TenantID]*utils.TPRankingProfile{
			{Tenant: "cgrates.org", ID: "rankingProfilesID"}: {},
		},
		dm: NewDataManager(NewInternalDB(nil, nil, false, cfg.DataDbCfg().Items), config.CgrConfig().CacheCfg(), connMgr),
	}
	tpr.dm.SetLoadIDs(make(map[string]int64))
	tpr.cacheConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	if err := tpr.ReloadCache(utils.MetaReload, false, make(map[string]any), "cgrates.org"); err != nil {
		t.Error(err)
	}
}

func TestTPReaderLoadDestinationsFiltered(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			utils.CacheTBLTPDestinations: {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value interface{}){
					func(itmID string, value any) {
					},
				},
			}},
	)
	tscache.Set(utils.CacheTBLTPDestinations, "itemId", &utils.TPDestination{
		TPid: "tpID",
		ID:   "prefixes",
	}, []string{"groupId"}, true, "tId")
	db.db = tscache

	tpr, err := NewTpReader(db, db, "itemId", "local", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	if b, err := tpr.LoadDestinationsFiltered(""); (err != nil) || !b {
		t.Errorf("expected nil ,received %v", err)
	}
	tscache.Remove(utils.CacheTBLTPDestinations, "itemId", true, utils.NonTransactional)

	if b, err := tpr.LoadDestinationsFiltered(""); err == nil || err != utils.ErrNotFound || b {
		t.Errorf("expected nil ,received %v", err)
	}
	tpr.dm = nil
	tscache.Set(utils.CacheTBLTPDestinations, "itemId", &utils.TPDestination{}, []string{"groupId"}, true, "tId")
	if b, err := tpr.LoadDestinationsFiltered(""); err != nil || !b {
		t.Errorf("expected nil ,received %v", err)
	}
}

func TestTPReaderLoadAll(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(nil, db, "", "local", nil, nil, false)
	if err != nil {
		t.Error(err)
	}
	if err = tpr.LoadAll(); err != nil {
		t.Error(err)
	}
}

func TestTpReaderReloadScheduler(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	ccMocK := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.SchedulerSv1Reload: func(ctx *context.Context, args, reply any) error {
				rpl := "reply"

				*reply.(*string) = rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMocK

	tmp := connMgr
	defer func() { connMgr = tmp }()
	connMgr = NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.SchedulerConnsCfg): clientconn,
	})

	tpr := &TpReader{
		actions: map[string][]*Action{
			"ActionsID": {},
		},
		actionPlans: map[string]*ActionPlan{
			"ActionPlansID":  {},
			"ActionPlansID2": {},
		},

		resProfiles: map[utils.TenantID]*utils.TPResourceProfile{
			{Tenant: "cgrates.org", ID: "resourceProfilesID"}: {},
		},
		sqProfiles: map[utils.TenantID]*utils.TPStatProfile{
			{Tenant: "cgrates.org", ID: "statProfilesID"}: {},
		},
		thProfiles: map[utils.TenantID]*utils.TPThresholdProfile{
			{Tenant: "cgrates.org", ID: "thresholdProfilesID"}: {},
		},
		filters: map[utils.TenantID]*utils.TPFilterProfile{
			{Tenant: "cgrates.org", ID: "filtersID"}: {},
		},
		routeProfiles: map[utils.TenantID]*utils.TPRouteProfile{
			{Tenant: "cgrates.org", ID: "routeProfilesID"}: {},
		},
		attributeProfiles: map[utils.TenantID]*utils.TPAttributeProfile{
			{Tenant: "cgrates.org", ID: "attributeProfilesID"}: {},
		},
		chargerProfiles: map[utils.TenantID]*utils.TPChargerProfile{
			{Tenant: "cgrates.org", ID: "chargerProfilesID"}: {},
		},
		dispatcherProfiles: map[utils.TenantID]*utils.TPDispatcherProfile{
			{Tenant: "cgrates.org", ID: "dispatcherProfilesID"}: {},
		},
		dispatcherHosts: map[utils.TenantID]*utils.TPDispatcherHost{
			{Tenant: "cgrates.org", ID: "dispatcherHostsID"}: {},
		},

		dm: NewDataManager(NewInternalDB(nil, nil, false, cfg.DataDbCfg().Items), config.CgrConfig().CacheCfg(), connMgr),
	}
	tpr.dm.SetLoadIDs(make(map[string]int64))
	tpr.schedulerConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.SchedulerConnsCfg)}

	if err := tpr.ReloadScheduler(false); err != nil {
		t.Error(err)
	}

}

func TestTpReaderIsValid(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(nil, db, "", "local", nil, nil, false)
	if err != nil {
		t.Error(err)
	}
	tpr.ratingPlans = map[string]*RatingPlan{
		"rate": {
			Timings: map[string]*RITiming{
				"timing": {
					StartTime: "00:00:00",
					Years:     utils.Years{},
					Months:    utils.Months{},
					MonthDays: utils.MonthDays{},
					WeekDays:  utils.WeekDays{},
				},
			},
		},
	}
	if valid := tpr.IsValid(); !valid {
		t.Error("expected true,received false")
	}

}

func TestTpReaderLoadAccountActions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			utils.CacheTBLTPAccountActions: {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {
					},
				},
			}},
	)
	tscache.Set(utils.CacheTBLTPAccountActions, "*prfitemId", &utils.TPAccountActions{
		TPid:    "tp_acc1",
		Account: "acc1",
		Tenant:  "tn1",
	}, []string{"groupId"}, true, "tId")
	tscache.Set(utils.CacheTBLTPAccountActions, "*prfitemId2", &utils.TPAccountActions{
		TPid:         "tp_acc2",
		Account:      "acc2",
		Tenant:       "tn2",
		ActionPlanId: "actionplans",
	}, []string{"groupId"}, true, "tId")
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	db.db = tscache
	tpr, err := NewTpReader(db, db, "*prf", "local", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	if err := tpr.LoadAccountActions(); err == nil || err.Error() != fmt.Sprintf("could not get action plan for tag %q", "actionplans") {

		t.Error(err)
	}
	tpr.dm = nil
	if err := tpr.LoadAccountActions(); err == nil {
		t.Error(err)
	}
}

func TestTPCSVImporterChargerProfiles(t *testing.T) {
	db := NewInternalDB(nil, nil, true, map[string]*config.ItemOpt{
		utils.CacheTBLTPChargers: {
			Limit: 2,
		},
	})
	tpCharger := &utils.TPChargerProfile{
		TPid:   "chargetp",
		Tenant: "cgrates",
		ID:     "id",
	}
	db.db.Set(utils.CacheTBLTPChargers, "tpid:itm1", tpCharger, []string{"grp"}, true, utils.NonTransactional)
	tpImp := &TPCSVImporter{
		TPid:    "tpid",
		Verbose: false,
		csvr:    db,
		StorDb:  db,
	}
	if err := tpImp.importChargerProfiles("fn"); err != nil {
		t.Error(err)
	}
	if val, has := db.db.Get(utils.CacheTBLTPChargers, utils.ConcatenatedKey(tpCharger.TPid, tpCharger.Tenant, tpCharger.ID)); !has {
		t.Error("has no value")
	} else if !reflect.DeepEqual(val, tpCharger) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(tpCharger), utils.ToJSON(val))
	}
}

func TestTPCSVImporterDispatcherProfiles(t *testing.T) {

	db := NewInternalDB(nil, nil, true, map[string]*config.ItemOpt{
		utils.CacheTBLTPDispatchers: {
			Limit: 3,
		},
	})
	dsP := &utils.TPDispatcherProfile{
		TPid:   "disTP",
		Tenant: "tnt",
		ID:     "id",
	}
	db.db.Set(utils.CacheTBLTPDispatchers, "tpid:dsp1", dsP, []string{"grp"}, true, utils.NonTransactional)
	tpImp := &TPCSVImporter{
		TPid:    "tpid",
		Verbose: false,
		csvr:    db,
		StorDb:  db,
	}
	if err := tpImp.importDispatcherProfiles("fn"); err != nil {
		t.Error(err)
	}
	if val, has := db.db.Get(utils.CacheTBLTPDispatchers, utils.ConcatenatedKey(dsP.TPid, dsP.Tenant, dsP.ID)); !has {
		t.Error("has no value")
	} else if !reflect.DeepEqual(val, dsP) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(dsP), utils.ToJSON(val))
	}

}

func TestTPCSVImporterDispatcherHosts(t *testing.T) {
	db := NewInternalDB(nil, nil, true, map[string]*config.ItemOpt{
		utils.CacheTBLTPDispatcherHosts: {
			Limit: 3,
		},
	})
	dsH := &utils.TPDispatcherHost{
		TPid:   "dshTp",
		Tenant: "host",
		ID:     "host_id",
	}
	db.db.Set(utils.CacheTBLTPDispatcherHosts, "tpid:dsp1", dsH, []string{"grp"}, true, utils.NonTransactional)
	tpImp := &TPCSVImporter{
		TPid:    "tpid",
		Verbose: false,
		csvr:    db,
		StorDb:  db,
	}
	if err := tpImp.importDispatcherHosts("fn"); err != nil {
		t.Error(err)
	}
	if val, has := db.db.Get(utils.CacheTBLTPDispatcherHosts, utils.ConcatenatedKey(dsH.TPid, dsH.Tenant, dsH.ID)); !has {
		t.Error("has no value")
	} else if !reflect.DeepEqual(val, dsH) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(dsH), utils.ToJSON(val))
	}

}

func TestTPCSVImporterErrs(t *testing.T) {
	db := NewInternalDB(nil, nil, true, map[string]*config.ItemOpt{
		utils.CacheTBLTPDispatcherHosts: {
			Limit: 3,
		},
	})
	tpImp := &TPCSVImporter{
		TPid:    "tpid",
		Verbose: false,
		csvr:    db,
		StorDb:  db,
	}
	fn := "test"
	if err := tpImp.importTimings(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importDestinations(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importRates(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importDestinationRates(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importRatingPlans(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importRatingProfiles(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importSharedGroups(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importActions(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importActionTimings(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importActionTriggers(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importAccountActions(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importResources(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importStats(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importThresholds(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importFilters(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importRoutes(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importAttributeProfiles(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importChargerProfiles(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importDispatcherProfiles(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := tpImp.importDispatcherHosts(fn); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestTpReaderLoadTimingsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			utils.CacheTBLTPAccountActions: {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {
					},
				},
			}},
	)
	duplicateId := "id"
	tscache.Set(utils.CacheTBLTPTimings, "*prfitemId", &utils.ApierTPTiming{
		TPid: "tpId2",
		ID:   duplicateId,
	}, []string{"groupId"}, true, "tId")
	tscache.Set(utils.CacheTBLTPTimings, "*prfitemId2", &utils.ApierTPTiming{
		TPid: "TpId3",
		ID:   duplicateId,
	}, []string{"groupId"}, true, "tId")
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	db.db = tscache
	tpr, err := NewTpReader(db, db, "*prf", "local", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	if err := tpr.LoadTimings(); err == nil || err.Error() != fmt.Sprintf("duplicate timing tag: %s", duplicateId) {
		t.Error(err)
	}
}

func TestLoadDestinationRatesErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			utils.CacheTBLTPDestinationRates: {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {
					},
				},
			},
		},
	)
	duplicateId := "id"
	tscache.Set(utils.CacheTBLTPDestinationRates, "*prfdest_rate1", &utils.TPDestinationRate{
		TPid: "tpId2",
		ID:   duplicateId,
	}, []string{"groupId"}, true, "tId")
	tscache.Set(utils.CacheTBLTPDestinationRates, "*prfdest_rate2", &utils.TPDestinationRate{
		TPid: "TpId3",
		ID:   duplicateId,
	}, []string{"groupId"}, true, "tId")
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	db.db = tscache
	tpr, err := NewTpReader(db, db, "*prf", "local", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	if err := tpr.LoadDestinationRates(); err == nil || err.Error() != fmt.Sprintf("Non unique ID %+s", duplicateId) {
		t.Error(err)
	}
	tpr.rates = map[string]*utils.TPRateRALs{
		"rate002": {},
	}
	tscache.Remove(utils.CacheTBLTPDestinationRates, "*prfdest_rate2", true, utils.NonTransactional)
	tpDestRate := &utils.TPDestinationRate{
		TPid: "tpId3",
		ID:   "tp_rate001",
		DestinationRates: []*utils.DestinationRate{
			{
				RateId:        "rate001",
				DestinationId: "val",
			},
		},
	}
	tscache.Set(utils.CacheTBLTPDestinationRates, "*prfdest_rate3", tpDestRate, []string{"grpId"}, true, utils.NonTransactional)
	if err := tpr.LoadDestinationRates(); err == nil || err.Error() != fmt.Sprintf("could not find rate for tag %q", tpDestRate.DestinationRates[0].RateId) {
		t.Error(err)
	}
	tpr.rates["rate001"] = &utils.TPRateRALs{
		TPid:      "tariff",
		ID:        "rals_id",
		RateSlots: []*utils.RateSlot{},
	}
	tpr.dm.dataDB = db
	if err := tpr.LoadDestinationRates(); err == nil || err.Error() != fmt.Sprintf("could not get destination for tag %q", tpDestRate.DestinationRates[0].DestinationId) {
		t.Error(err)
	}
}
func TestTpReaderLoadRatingPlansFilteredErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheTBLTPRatingPlans: {
			TTL:       time.Minute * 30,
			StaticTTL: false,
		},
		utils.CacheTBLTPTimings: {
			TTL:       time.Minute * 30,
			StaticTTL: false,
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)

	tpr, err := NewTpReader(db, db, "*prf", "local", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	if b, err := tpr.LoadRatingPlansFiltered("tag"); err == nil || b || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestLoadRatingProfilesFiltered(t *testing.T) {
	qriedRpf := &utils.TPRatingProfile{
		TPid:   "rate",
		Tenant: "cgr",
	}
	cfg := config.NewDefaultCGRConfig()
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			utils.CacheTBLTPRatingProfiles: {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
				OnEvicted: []func(itmID string, value any){
					func(itmID string, value any) {
					},
				},
			},
		},
	)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	db.db = tscache
	tpr, err := NewTpReader(db, db, "*prf", "local", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	if err := tpr.LoadRatingProfilesFiltered(qriedRpf); err == nil || err.Error() != fmt.Sprintf("no RatingProfile for filter %v, error: %v", qriedRpf, utils.ErrNotFound) {
		t.Error(err)
	}
	val := []*utils.TPRatingProfile{
		{
			LoadId:   "load",
			Tenant:   "cgrates",
			Category: "cat",
			Subject:  " subj",
			TPid:     "rating1",
		}, {
			LoadId:   "load",
			Tenant:   "cgrates",
			Category: "cat",
			Subject:  " subj",
			TPid:     "rating1",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					RatingPlanId:   "RP_1001",
					ActivationTime: "test",
				},
			},
		},
	}
	tscache.Set(utils.CacheTBLTPRatingProfiles, "rate:cgritm", val[0], []string{"grpId"}, true, utils.NonTransactional)
	tscache.Set(utils.CacheTBLTPRatingProfiles, "rate:cgritm2", val[1], []string{"grpId"}, true, utils.NonTransactional)
	if err := tpr.LoadRatingProfilesFiltered(qriedRpf); err == nil || err.Error() != fmt.Sprintf("Non unique id %+v", val[1].GetId()) {
		t.Error(err)
	}
	val[1].TPid, val[1].LoadId, val[1].Category = "rating2", "load2", "category2"
	if err := tpr.LoadRatingProfilesFiltered(qriedRpf); err == nil || err.Error() != fmt.Sprintf("cannot parse activation time from %v", val[1].RatingPlanActivations[0].ActivationTime) {
		t.Error(err)
	}
	tpr.timezone, val[1].RatingPlanActivations[0].ActivationTime = "UTC", "*monthly_estimated"
	if err := tpr.LoadRatingProfilesFiltered(qriedRpf); err == nil || err.Error() != fmt.Sprintf("could not load rating plans for tag: %q", val[1].RatingPlanActivations[0].RatingPlanId) {
		t.Error(err)
	}
}

func TestTpReaderLoadActionTriggers(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			utils.CacheTBLTPActionTriggers: {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
			},
		},
	)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	db.db = tscache
	tpr, err := NewTpReader(db, db, "*prf", "UTC", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	trVals := []*utils.TPActionTriggers{
		{
			TPid: "TPAct",
			ID:   "ID2",
			ActionTriggers: []*utils.TPActionTrigger{
				{
					BalanceId:             "id",
					Id:                    "STANDARD_TRIGGERS",
					ThresholdType:         "*min_balance",
					ThresholdValue:        2,
					Recurrent:             false,
					MinSleep:              "0",
					BalanceRatingSubject:  "rate",
					BalanceType:           "*monetary",
					BalanceDestinationIds: "FS_USERS",
					ActionsId:             "LOG_WARNING",
					Weight:                10,
					BalanceWeight:         "20.12",
					BalanceExpirationDate: "*monthly",
					BalanceCategories:     "monthly",
				},
				{
					BalanceId:             "id",
					Id:                    "STANDARD_TRIGGERS",
					ThresholdType:         "*max_event_counter",
					ThresholdValue:        5,
					Recurrent:             false,
					MinSleep:              "0",
					BalanceType:           "*monetary",
					BalanceDestinationIds: "FS_USERS",
					ActionsId:             "LOG_WARNING",
					Weight:                10,
					BalanceWeight:         "20.1",
					BalanceExpirationDate: "*yearly",
					ExpirationDate:        "date",
					BalanceSharedGroups:   "group1",
					BalanceTimingTags:     "timing",
					BalanceBlocker:        "true",
					BalanceDisabled:       "false",
				},
			},
		},
	}
	tscache.Set(utils.CacheTBLTPActionTriggers, "*prfitem1", trVals[0], []string{"*prfitem1", "*prfitem2"}, true, utils.NonTransactional)
	if err := tpr.LoadActionTriggers(); err == nil || err.Error() != "Unsupported time format" {
		t.Error(err)
	}
	trVals[0].ActionTriggers[0].ExpirationDate = "*monthly"
	trVals[0].ActionTriggers[0].ActivationDate = "value"
	if err := tpr.LoadActionTriggers(); err == nil || err.Error() != "Unsupported time format" {
		t.Error(err)
	}
	trVals[0].ActionTriggers[0].ActivationDate = "*monthly"
	trVals[0].ActionTriggers[0].MinSleep = "two"
	if err := tpr.LoadActionTriggers(); err == nil || err.Error() != fmt.Sprintf("time: invalid duration %q", trVals[0].ActionTriggers[0].MinSleep) {
		t.Error(err)
	}
	trVals[0].ActionTriggers[0].MinSleep = "5000ns"
	trVals[0].ActionTriggers[1].ExpirationDate = "*montly"
	if err := tpr.LoadActionTriggers(); err != nil {
		t.Error(err)
	}
}

func TestTpReaderSetDestination(t *testing.T) {
	dest := &Destination{
		Id:       "1001_ID",
		Prefixes: []string{"39", "75"},
	}
	cfg := config.NewDefaultCGRConfig()
	tscache := ltcache.NewTransCache(
		map[string]*ltcache.CacheConfig{
			utils.CacheDestinations: {
				MaxItems:  3,
				TTL:       time.Minute * 30,
				StaticTTL: false,
			},
		},
	)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	db.db = tscache
	tpr, err := NewTpReader(db, db, "*prf", "UTC", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	if err := tpr.setDestination(dest, true, ""); err != nil {
		t.Error(err)
	}
	if val, has := tscache.Get(utils.CacheDestinations, dest.Id); !has {
		t.Error("has no value")
	} else {
		if rcv, cancast := val.(*Destination); !cancast {
			t.Error("it's not type *Destination")
		} else if !reflect.DeepEqual(rcv, dest) {
			t.Errorf("exepcted %v,received %v", utils.ToJSON(dest), utils.ToJSON(rcv))
		}
	}
}

func TestTPReaderLoadAccountActionsFilteredErr(t *testing.T) {
	Cache.Clear(nil)
	qried := &utils.TPAccountActions{
		TPid:          "tp_Id",
		AllowNegative: false,
		Disabled:      true,
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheTBLTPAccountActions: {
			Limit:  3,
			Remote: true,
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "*prf", "UTC", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	if err := tpr.LoadAccountActionsFiltered(qried); err == nil || err.Error() != fmt.Sprintf("%v: %+v", utils.ErrNotFound.Error(), qried) {
		t.Error(err)
	}
	db.db.Set(utils.CacheTBLTPAccountActions, "tp_Id:item", &utils.TPAccountActions{TPid: utils.TestSQL, LoadId: utils.TestSQL, Tenant: "cgrates.org",
		Account: "1001", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}, []string{"grpId"}, true, utils.NonTransactional)
	db.db.Set(utils.CacheTBLTPAccountActions, "tp_Id:item2", &utils.TPAccountActions{TPid: utils.TestSQL, LoadId: utils.TestSQL, Tenant: "cgrates.org",
		Account: "1001", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}, []string{"grpId"}, true, utils.NonTransactional)
	if err := tpr.LoadAccountActionsFiltered(qried); err == nil || err.Error() != fmt.Sprintf("Non unique ID %+v", utils.ConcatenatedKey("cgrates.org", "1001")) {
		t.Error(err)
	}
	db.db.Remove(utils.CacheTBLTPAccountActions, "tp_Id:item2", true, utils.NonTransactional)
	if err := tpr.LoadAccountActionsFiltered(qried); err == nil || err.Error() != fmt.Sprint(utils.ErrNotFound.Error()+" (ActionPlan): "+"PREPAID_10") {
		t.Error(err)
	}
	db.db.Set(utils.CacheTBLTPActionPlans, "*prf:PREPAID_10", &utils.TPActionPlan{
		TPid: "TEST_TPID",
		ID:   "PACKAGE_10",
		ActionPlan: []*utils.TPActionTiming{
			{
				ActionsId: "TOPUP_RST_10",
				TimingId:  "ASAP",
				Weight:    10.0},
			{
				ActionsId: "TOPUP_RST_5",
				TimingId:  "ASAP",
				Weight:    20.0},
		},
	}, []string{"grpID"}, true, utils.NonTransactional)
}

func TestTprRemoveFromDatabase(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheSharedGroups: {
			Limit: 3,
		},
		utils.CacheChargerProfiles: {
			Limit: 2,
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "*prf", "UTC", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	db.db.Set(utils.CacheSharedGroups, "itmID", &SharedGroup{
		Id: "SG_TEST",
	}, []string{}, true, utils.NonTransactional)
	db.db.Set(utils.CacheChargerProfiles, "cgrates.org:DEFAULT", &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "DEFAULT",
		FilterIDs:    []string{},
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       0,
	}, []string{}, true, utils.NonTransactional)

	tpr.sharedGroups = map[string]*SharedGroup{
		"itmID": {
			Id: "SG_TEST",
		},
	}
	tpr.chargerProfiles = map[utils.TenantID]*utils.TPChargerProfile{
		{Tenant: "cgrates", ID: "item2"}: {
			Tenant:       "cgrates.org",
			ID:           "DEFAULT",
			FilterIDs:    []string{},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       0,
		},
	}
	if err := tpr.RemoveFromDatabase(false, true); err != nil {
		t.Error(err)
	}
	if _, has := db.db.Get(utils.CacheSharedGroups, "itmID"); has {
		t.Error("should been removed from the cache")
	} else if _, has := db.db.Get(utils.CacheSharedGroups, "cgrates.org:DEFAULT"); has {
		t.Error("should been removed from the cache")
	}
}

func TestLoadActionPlansErrs(t *testing.T) {
	tmp := Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheTBLTPActionPlans: {
			StaticTTL: true,
			Limit:     4,
		},
		utils.CacheActions: {
			Limit: 2,
		},
	}
	defer func() {
		Cache = tmp
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "tpr", "UTC", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	db.db.Set(utils.CacheTBLTPActionPlans, "tpr:item1", &utils.TPActionPlan{
		TPid: "TEST_TPID",
		ID:   "PACKAGE_10",
		ActionPlan: []*utils.TPActionTiming{
			{
				ActionsId: "TOPUP_RST_10",
				TimingId:  "ASAP",
				Weight:    10.0},
		},
	}, []string{}, true, utils.NonTransactional)
	tpr.actions = map[string][]*Action{
		"TOPUP_RST_*": {},
	}
	if err := tpr.LoadActionPlans(); err == nil || err.Error() != fmt.Sprintf("[ActionPlans] Could not load the action for tag: %q", "TOPUP_RST_10") {
		t.Error(err)
	}
	db.db.Set(utils.CacheActions, "TOPUP_RST_10", nil, []string{}, true, utils.NonTransactional)
	if err := tpr.LoadActionPlans(); err == nil || err.Error() != fmt.Sprintf("[ActionPlans] Could not load the timing for tag: %q", "ASAP") {
		t.Error(err)
	}
}

func TestLoadRatingPlansFiltered(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheSharedGroups: {
			Limit: 3,
		},
		utils.CacheTBLTPTimings: {
			Limit: 2,
		},
		utils.CacheTBLTPDestinationRates: {
			Limit: 3,
		},
		utils.CacheTBLTPRates: {
			Limit: 2,
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "*prf", "UTC", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	db.db.Set(utils.CacheTBLTPRatingPlans, "*prf:def:itmID", &utils.TPRatingPlan{
		ID:   "ID",
		TPid: "TPid",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			{
				DestinationRatesId: "DestinationRatesId",
				TimingId:           "TimingId",
				Weight:             0.7,
			},
		},
	}, []string{}, true, utils.NonTransactional)
	if _, err := tpr.LoadRatingPlansFiltered("def"); err == nil || !strings.Contains(err.Error(), "no timing with id ") {
		t.Error(err)
	}
	db.db.Set(utils.CacheTBLTPTimings, "*prf:TimingId", &utils.ApierTPTiming{

		TPid:      "testTPid",
		ID:        "TimingId",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "1;2;3;4;5",
		Time:      "01:00:00",
	}, []string{}, true, utils.NonTransactional)
	db.db.Set(utils.CacheTBLTPTimings, "*prf:TimingId2", &utils.ApierTPTiming{

		TPid:      "testTPid",
		ID:        "TimingId",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "1;2;3;4;5",
		Time:      "01:00:00",
	}, []string{}, true, utils.NonTransactional)
	if _, err := tpr.LoadRatingPlansFiltered("def"); err == nil {
		t.Error(err)
	}
	db.db.Remove(utils.CacheTBLTPTimings, "*prf:TimingId2", true, utils.NonTransactional)
	if _, err := tpr.LoadRatingPlansFiltered("def"); err == nil || !strings.Contains(err.Error(), "no DestinationRates profile with id") {
		t.Error(err)
	}
	db.db.Set(utils.CacheTBLTPDestinationRates, "*prf:DestinationRatesId1", &utils.TPDestinationRate{
		TPid: "TEST_TPID",
		ID:   "DestinationRatesId",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:  "TEST_DEST1",
				RateId:         "TEST_RATE1",
				RoundingMethod: "*up",
				Rate: &utils.TPRateRALs{
					TPid: "TPidTpRate",
					ID:   "RT_FS_USERS",
					RateSlots: []*utils.RateSlot{
						{
							ConnectFee:         12,
							Rate:               3,
							RateUnit:           "6s",
							RateIncrement:      "6s",
							GroupIntervalStart: "0s",
						},
					},
				},
				RoundingDecimals: 4},
			{
				DestinationId:    "TEST_DEST2",
				RateId:           "TEST_RATE2",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
		},
	}, []string{}, true, utils.NonTransactional)
	db.db.Set(utils.CacheTBLTPDestinationRates, "*prf:DestinationRatesId2", &utils.TPDestinationRate{
		TPid: "TEST_TPID",
		ID:   "DestinationRatesId",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:    "TEST_DEST1",
				RateId:           "TEST_RATE1",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
		},
	}, []string{}, true, utils.NonTransactional)
	if _, err := tpr.LoadRatingPlansFiltered("def"); err == nil {
		t.Error(err)
	}
	db.db.Remove(utils.CacheTBLTPDestinationRates, "*prf:DestinationRatesId2", true, utils.NonTransactional)
	if _, err := tpr.LoadRatingPlansFiltered("def"); err == nil || !strings.Contains(err.Error(), "no Rates profile with id ") {
		t.Error(err)
	}
	db.db.Set(utils.CacheTBLTPRates, "*prf:TEST_RATE1", &utils.TPRateRALs{
		TPid: "TPidTpRate",
		ID:   "RT_FS_USERS",
		RateSlots: []*utils.RateSlot{
			{
				ConnectFee:         12,
				Rate:               3,
				RateUnit:           "6s",
				RateIncrement:      "6s",
				GroupIntervalStart: "0s",
			},
		},
	}, []string{}, true, utils.NonTransactional)
	db.db.Set(utils.CacheTBLTPRates, "*prf:TEST_RATE12", &utils.TPRateRALs{
		TPid: "TPidTpRate",
		ID:   "RT_FS_USERS",
		RateSlots: []*utils.RateSlot{
			{
				ConnectFee:         12,
				Rate:               3,
				RateUnit:           "6s",
				RateIncrement:      "6s",
				GroupIntervalStart: "0s",
			},
		},
	}, []string{}, true, utils.NonTransactional)
	if _, err := tpr.LoadRatingPlansFiltered("def"); err == nil {
		t.Error(err)
	}
}

func TestTPRLoadRatingProfiles(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "RP1", "", nil, nil, false)

	if err != nil {
		t.Error(err)
	}
	rpS := []*utils.TPRatingProfile{
		{
			TPid: "RP1",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					ActivationTime:   "2014-07-29T15:00:00Z",
					RatingPlanId:     "PlanOne",
					FallbackSubjects: "FallBack",
				},
			},
		},
	}
	if err := db.SetTPRatingProfiles(rpS); err != nil {
		t.Error(err)
	}
	if err = tpr.LoadRatingProfiles(); err == nil || err.Error() != fmt.Sprintf("could not load rating plans for tag: %q", rpS[0].RatingPlanActivations[0].RatingPlanId) {
		t.Error(err)
	}
	rpS2 := []*utils.TPRatingProfile{
		{
			TPid: "RP2",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					ActivationTime:   "2014-07-29T15:00:00Z",
					RatingPlanId:     "PlanOne",
					FallbackSubjects: "FallBack",
				},
			},
		},
		{
			TPid: "RP2",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					ActivationTime:   "2012-01-01T00:00:00Z",
					RatingPlanId:     "RPl_SAMPLE_RATING_PLAN",
					FallbackSubjects: utils.EmptyString,
				},
				{
					ActivationTime:   "test",
					RatingPlanId:     "RPl_SAMPLE_RATING_PLAN2",
					FallbackSubjects: utils.EmptyString,
				},
			},
		}}
	if err := db.SetTPRatingProfiles(rpS2); err != nil {
		t.Error(err)
	}
	if err = tpr.LoadRatingProfiles(); err == nil {
		t.Error(err)
	}
}

func TestTPRLoadAccountActions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "", "", nil, nil, false)

	if err != nil {
		t.Error(err)
	}
	accAcs := []*utils.TPAccountActions{
		{
			TPid:             "testTPid",
			ActionPlanId:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersId: "STANDARD_TRIGGERS",
			AllowNegative:    true,
			Disabled:         true,
		},
	}
	if err := db.SetTPAccountActions(accAcs); err != nil {
		t.Error(err)
	}
	if err = tpr.LoadAccountActions(); err == nil || err.Error() != fmt.Sprintf("could not get action triggers for tag %q", accAcs[0].ActionTriggersId) {
		t.Error(err)
	}
}
func TestTpReaderRemoveFromDatabase(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "", "", nil, nil, false)
	if err != nil {
		t.Error(err)
	}
	dest := &Destination{
		Id:       "DST2",
		Prefixes: []string{"1001"},
	}
	tpr.destinations = map[string]*Destination{
		"GERMANY": dest,
	}
	if tpr.dm.SetDestination(dest, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	acc := &Account{
		ID: "cgrates.org:1001",
	}
	ap1 := &ActionPlan{
		Id:         "TestActionPlansRemoveMember1",
		AccountIDs: utils.StringMap{"one": true},
		ActionTimings: []*ActionTiming{
			{
				Uuid: "uuid1",
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.MetaASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
		},
	}
	if err := tpr.dm.SetAccount(acc); err != nil {
		t.Error(err)
	}
	if err := tpr.dm.SetActionPlan(ap1.Id, ap1, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err = tpr.dm.SetAccountActionPlans(acc.ID, []string{ap1.Id}, false); err != nil {
		t.Error(err)
	}
	tpr.acntActionPlans = map[string][]string{
		acc.ID: {ap1.Id},
	}
	if err := tpr.RemoveFromDatabase(false, false); err != nil {
		t.Error(err)
	}
	if err := tpr.RemoveFromDatabase(false, true); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestTpReaderRemoveFromDatabaseDspPrf(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "", "", nil, nil, false)
	if err != nil {
		t.Error(err)
	}
	dspPrf := &DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:~*req.Accont:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Strategy: utils.MetaRandom,
		Weight:   20,
	}
	if err := tpr.dm.SetDispatcherProfile(dspPrf, true); err != nil {
		t.Error(err)
	}
	tpr.dispatcherProfiles = map[utils.TenantID]*utils.TPDispatcherProfile{
		{
			Tenant: "cgrates.org",
			ID:     "Dsp1",
		}: {
			Tenant: "cgrates.org",
			ID:     "Dsp1",
		},
	}
	if err = tpr.RemoveFromDatabase(false, true); err != nil {
		t.Error(err)
	}
	if err = tpr.RemoveFromDatabase(false, true); err == nil || err != utils.ErrDSPProfileNotFound {
		t.Error(err)
	}
}

func TestTpReaderRemoveFromDatabaseDspHst(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, db, "", "", nil, nil, false)
	if err != nil {
		t.Error(err)
	}
	dspHst := &DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:        "Host2",
			Address:   "127.0.0.1:2013",
			TLS:       false,
			Transport: utils.MetaGOB,
		},
	}
	if err = tpr.dm.SetDispatcherHost(dspHst); err != nil {
		t.Error(err)
	}
	tpr.dispatcherHosts = map[utils.TenantID]*utils.TPDispatcherHost{
		{}: {
			ID:     "Host2",
			Tenant: "cgrates.org",
		},
	}
	if err = tpr.RemoveFromDatabase(false, true); err != nil {
		t.Error(err)
	}
	if err = tpr.RemoveFromDatabase(false, true); err == nil || err != utils.ErrDSPHostNotFound {
		t.Error(err)
	}
}

func TestTprLoadAccountActionFiltered(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpConn := connMgr
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		SetConnManager(tmpConn)
	}()
	dataDb := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	storDb := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	tpr, err := NewTpReader(dataDb, storDb, "TP1", "", []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}, nil, false)
	if err != nil {
		t.Error(err)
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- clMock(func(serviceMethod string, _, _ any) error {
		if serviceMethod == utils.CacheSv1ReloadCache {

			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): clientconn,
	})
	timings := &utils.ApierTPTiming{
		TPid:      "TP1",
		ID:        "ASAP",
		Years:     "",
		Months:    "05",
		MonthDays: "01",
		WeekDays:  "1",
		Time:      "15:00:00Z",
	}

	storDb.SetTPTimings([]*utils.ApierTPTiming{timings})

	actions := &utils.TPActions{
		TPid: "TP1",
		ID:   "TOPUP_RST_10",
		Actions: []*utils.TPAction{
			{
				Identifier:    "*topup_reset",
				BalanceId:     "BalID",
				BalanceType:   "*data",
				Units:         "10",
				ExpiryTime:    "*unlimited",
				TimingTags:    utils.MetaASAP,
				BalanceWeight: "10",
				Weight:        10,
			},
		},
	}

	storDb.SetTPActions([]*utils.TPActions{actions})
	actionplans := []*utils.TPActionPlan{
		{
			TPid: "TP1",
			ID:   "PREPAID_10",
			ActionPlan: []*utils.TPActionTiming{
				{
					ActionsId: "TOPUP_RST_10",
					TimingId:  "ASAP",
					Weight:    10.0},
			},
		},
	}
	tpatrs := &utils.TPActionTriggers{
		TPid: "TP1",
		ID:   "STANDARD_TRIGGERS",
		ActionTriggers: []*utils.TPActionTrigger{
			{
				Id:                "STANDARD_TRIGGERS",
				UniqueID:          "1",
				ThresholdType:     "*min_balance",
				ThresholdValue:    2.0,
				Recurrent:         false,
				BalanceType:       "*monetary",
				BalanceCategories: "call",
				ActionsId:         "LOG_WARNING",
				Weight:            10},
		},
	}
	storDb.SetTPActionTriggers([]*utils.TPActionTriggers{
		tpatrs,
	})
	storDb.SetTPActionPlans(actionplans)
	qriedAA := &utils.TPAccountActions{
		TPid:             "TP1",
		LoadId:           "ID",
		Tenant:           "cgrates.org",
		Account:          "1001",
		ActionPlanId:     "PREPAID_10",
		ActionTriggersId: "STANDARD_TRIGGERS",
		AllowNegative:    true,
		Disabled:         false,
	}
	storDb.SetTPAccountActions([]*utils.TPAccountActions{
		qriedAA,
	})

	SetConnManager(connMgr)
	if err := tpr.LoadAccountActionsFiltered(qriedAA); err == nil {
		t.Error(err)
	}
	//unfinished
}

func TestTprLoadRatingPlansFiltered(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	dataDb := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	storDb := NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)

	storDb.SetTPDestinations([]*utils.TPDestination{{
		TPid:     "TP1",
		ID:       "FS_USERS",
		Prefixes: []string{"+664", "+554"}},
	})

	tprt := &utils.TPRateRALs{
		TPid: "TP1",
		ID:   "RT_FS_USERS",
		RateSlots: []*utils.RateSlot{
			{
				ConnectFee:         12,
				Rate:               3,
				RateUnit:           "6s",
				RateIncrement:      "6s",
				GroupIntervalStart: "0s",
			},
			{
				ConnectFee:         12,
				Rate:               3,
				RateUnit:           "4s",
				RateIncrement:      "6s",
				GroupIntervalStart: "1s",
			},
		},
	}

	storDb.SetTPRates([]*utils.TPRateRALs{tprt})

	tpdr := &utils.TPDestinationRate{
		TPid: "TP1",
		ID:   "DR_FREESWITCH_USERS",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:    "FS_USERS",
				RateId:           "RT_FS_USERS",
				RoundingMethod:   "*up",
				RoundingDecimals: 2},
		},
	}
	storDb.SetTPDestinationRates([]*utils.TPDestinationRate{tpdr})
	storDb.SetTPTimings([]*utils.ApierTPTiming{
		{
			TPid:      "TP1",
			ID:        "ALWAYS",
			Years:     "*any",
			Months:    "*any",
			MonthDays: "*any",
			WeekDays:  "1;2;3;4;5",
			Time:      "00:00:00"},
	})
	rp := &utils.TPRatingPlan{
		TPid: "TP1",
		ID:   "RPl_SAMPLE_RATING_PLAN2",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			{
				DestinationRatesId: "DR_FREESWITCH_USERS",
				TimingId:           "ALWAYS",
				Weight:             10,
			},
		}}
	storDb.SetTPRatingPlans([]*utils.TPRatingPlan{rp})
	tpr, err := NewTpReader(dataDb, storDb, "TP1", "", []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}, nil, false)
	if err != nil {
		t.Error(err)
	}
	if load, err := tpr.LoadRatingPlansFiltered(""); err != nil || !load {
		t.Error(err)
	}
}
