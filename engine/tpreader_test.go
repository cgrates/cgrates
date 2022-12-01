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
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
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
	opts := map[string]interface{}{
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
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(args, reply interface{}) error {
				expArgs := &utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
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

	cM := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaReload
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]interface{}{
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
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(args, reply interface{}) error {
				return nil
			},
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
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

	cM := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaReload
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
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
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1LoadCache: func(args, reply interface{}) error {
				expArgs := &utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
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
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
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

	cM := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaLoad
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
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
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1RemoveItems: func(args, reply interface{}) error {
				expArgs := &utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
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
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
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

	cM := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaRemove
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
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
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
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

	cM := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaClear
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]interface{}{
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
		APIOpts:               map[string]interface{}{},
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
		ReverseDestinationIDs: []string{},
	}
	rpcInternal := make(chan rpcclient.ClientConnector, 1)
	rpcInternal <- &ccMock{
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
	tmp := connMgr
	defer func() { connMgr = tmp }()
	connMgr = NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
		dm: NewDataManager(NewInternalDB(nil, nil, false, cfg.DataDbCfg().Items), config.CgrConfig().CacheCfg(), connMgr),
	}
	tpr.dm.SetLoadIDs(make(map[string]int64))
	tpr.cacheConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	if err := tpr.ReloadCache(utils.MetaReload, false, make(map[string]interface{}), "cgrates.org"); err != nil {
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
				OnEvicted: func(itmID string, value interface{}) {
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
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.SchedulerSv1Reload: func(args, reply interface{}) error {
				rpl := "reply"

				*reply.(*string) = rpl
				return nil
			},
		},
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- ccMocK

	tmp := connMgr
	defer func() { connMgr = tmp }()
	connMgr = NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
				OnEvicted: func(itmID string, value interface{}) {
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
