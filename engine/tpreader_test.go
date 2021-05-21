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
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestCallCacheNoCaching(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cM := NewConnManager(defaultCfg, nil)
	cacheConns := []string{}
	caching := utils.MetaNone
	args := map[string][]string{
		utils.FilterIDs:   {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
		utils.ResourceIDs: {},
	}
	cacheIDs := []string{}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}

	expArgs := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, true)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(args, expArgs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expArgs, args)
	}
}

func TestCallCacheReloadCacheFirstCallErr(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(args, reply interface{}) error {
				expArgs := utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					ArgsCache: map[string][]string{
						utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
					},
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

	cM := NewConnManager(defaultCfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaReload
	args := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	explog := "Reloading cache\n"
	experr := utils.ErrUnsupporteServiceMethod
	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, true)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	rcvlog := buf.String()[20:]
	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}
}

func TestCallCacheReloadCacheSecondCallErr(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(args, reply interface{}) error {
				return nil
			},
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					CacheIDs: []string{"cacheID"},
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

	cM := NewConnManager(defaultCfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaReload
	args := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
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
	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, true)

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

func TestCallCacheLoadCache(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1LoadCache: func(args, reply interface{}) error {
				expArgs := utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					ArgsCache: map[string][]string{
						utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
					},
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					CacheIDs: []string{"cacheID"},
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

	cM := NewConnManager(defaultCfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaLoad
	args := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}

	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, false)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestCallCacheRemoveItems(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1RemoveItems: func(args, reply interface{}) error {
				expArgs := utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					ArgsCache: map[string][]string{
						utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
					},
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					CacheIDs: []string{"cacheID"},
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

	cM := NewConnManager(defaultCfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaRemove
	args := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}

	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, false)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestCallCacheClear(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
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

	cM := NewConnManager(defaultCfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaClear
	args := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}

	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, false)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestGetLoadedIdsDestinations(t *testing.T) {
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

func TestGetLoadedIdsReverseDestinations(t *testing.T) {
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

func TestGetLoadedIdsRatingPlans(t *testing.T) {
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
						RoundingMethod:   utils.MetaUp,
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

func TestGetLoadedIdsRatingProfiles(t *testing.T) {
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

func TestGetLoadedIdsActions(t *testing.T) {
	tpr := &TpReader{
		actions: map[string][]*Action{
			"TOPUP_RST_10": {
				{
					Id:               "ACTION_1001",
					ActionType:       utils.MetaSetBalance,
					ExtraParameters:  "",
					Filter:           "",
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
					Filter:           "",
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

func TestGetLoadedIdsActionPlans(t *testing.T) {
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

func TestGetLoadedIdsSharedGroup(t *testing.T) {
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

func TestGetLoadedIdsResourceProfiles(t *testing.T) {
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

func TestGetLoadedIdsActionTriggers(t *testing.T) {
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

func TestGetLoadedIdsStatQueueProfiles(t *testing.T) {
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

func TestGetLoadedIdsThresholdProfiles(t *testing.T) {
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

func TestGetLoadedIdsFilters(t *testing.T) {
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

func TestGetLoadedIdsRouteProfiles(t *testing.T) {
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

func TestGetLoadedIdsAttributeProfiles(t *testing.T) {
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

func TestGetLoadedIdsChargerProfiles(t *testing.T) {
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

func TestGetLoadedIdsDispatcherProfiles(t *testing.T) {
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

func TestGetLoadedIdsEmptyObject(t *testing.T) {
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

func TestGetLoadedIdsDispatcherHosts(t *testing.T) {
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

func TestGetLoadedIdsError(t *testing.T) {
	tpr := &TpReader{}
	errExpect := "Unsupported load category"
	if _, err := tpr.GetLoadedIds(""); err == nil || err.Error() != errExpect {
		t.Errorf("\nExpected error %v but received \n%v", errExpect, err)
	}
}

func TestReloadCache(t *testing.T) {
	data := NewInternalDB(nil, nil, false)
	cfg := config.NewDefaultCGRConfig()
	argExpect := utils.AttrReloadCacheWithAPIOpts{
		APIOpts: map[string]interface{}{},
		Tenant:  "",
		ArgsCache: map[string][]string{
			"ActionIDs":            {"ActionsID"},
			"ActionPlanIDs":        {"ActionPlansID"},
			"ActionTriggerIDs":     {"ActionTriggersID"},
			"DestinationIDs":       {"DestinationsID"},
			"TimingIDs":            {"TimingsID"},
			"RatingPlanIDs":        {"RatingPlansID"},
			"RatingProfileIDs":     {"RatingProfilesID"},
			"SharedGroupIDs":       {"SharedGroupsID"},
			"ResourceProfileIDs":   {"cgrates.org:resourceProfilesID"},
			"StatsQueueProfileIDs": {"cgrates.org:statProfilesID"},
			"ThresholdProfileIDs":  {"cgrates.org:thresholdProfilesID"},
			"FilterIDs":            {"cgrates.org:filtersID"},
			"RouteProfileIDs":      {"cgrates.org:routeProfilesID"},
			"AttributeProfileIDs":  {"cgrates.org:attributeProfilesID"},
			"ChargerProfileIDs":    {"cgrates.org:chargerProfilesID"},
			"DispatcherProfileIDs": {"cgrates.org:dispatcherProfilesID"},
			"DispatcherHostIDs":    {"cgrates.org:dispatcherHostsID"},
			"ResourceIDs":          {"cgrates.org:resourceProfilesID"},
			"StatsQueueIDs":        {"cgrates.org:statProfilesID"},
			"ThresholdIDs":         {"cgrates.org:thresholdProfilesID"},
			"AccountActionPlanIDs": {"AccountActionPlansID"},
		},
	}
	cM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(args interface{}, reply interface{}) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(args interface{}, reply interface{}) error {
				return nil
			},
		},
	}
	rpcInternal := make(chan rpcclient.ClientConnector, 1)
	rpcInternal <- cM
	cnMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
		dm: NewDataManager(data, config.CgrConfig().CacheCfg(), cnMgr),
	}
	tpr.cacheConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	if err := tpr.ReloadCache(utils.MetaReload, false, make(map[string]interface{})); err != nil {
		t.Error(err)
	}
}
