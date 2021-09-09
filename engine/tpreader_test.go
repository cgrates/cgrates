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
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestCallCacheNoCaching(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cM := NewConnManager(defaultCfg)
	args := map[string][]string{
		utils.CacheFilters:   {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
		utils.CacheResources: {},
	}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}
	ctx := context.Background()

	err := CallCache(cM, ctx, []string{}, utils.MetaNone, args, []string{}, opts, true, "cgrates.org")

	if err != nil {
		t.Error(err)
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
	client := make(chan birpc.ClientConnector, 1)
	mCC := &ccMock{
		calls: map[string]func(_ *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(_ *context.Context, args, reply interface{}) error {
				expArgs := &utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
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
	client <- mCC

	cM := NewConnManager(defaultCfg)
	cM.AddInternalConn("cacheConn1", "", client)
	caching := utils.MetaReload
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}
	ctx := context.Background()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	explog := "Reloading cache\n"
	experr := utils.ErrUnsupporteServiceMethod
	err := CallCache(cM, ctx, cacheConns, caching, args, cacheIDs, opts, true, "cgrates.org")

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
	client := make(chan birpc.ClientConnector, 1)
	mCC := &ccMock{
		calls: map[string]func(_ *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(_ *context.Context, args, reply interface{}) error {
				return nil
			},
			utils.CacheSv1Clear: func(_ *context.Context, args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
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
	client <- mCC

	cM := NewConnManager(defaultCfg)
	cM.AddInternalConn("cacheConn1", "", client)
	caching := utils.MetaReload
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}
	ctx := context.Background()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	explog1 := "Reloading cache"
	explog2 := "Clearing indexes"
	experr := utils.ErrUnsupporteServiceMethod
	explog3 := fmt.Sprintf("WARNING: Got error on cache clear: %s\n", experr)
	err := CallCache(cM, ctx, cacheConns, caching, args, cacheIDs, opts, true, "cgrates.org")

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
	client := make(chan birpc.ClientConnector, 1)
	mCC := &ccMock{
		calls: map[string]func(_ *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1LoadCache: func(_ *context.Context, args, reply interface{}) error {
				expArgs := &utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
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
			utils.CacheSv1Clear: func(_ *context.Context, args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
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
	client <- mCC

	cM := NewConnManager(defaultCfg)
	cM.AddInternalConn("cacheConn1", "", client)
	caching := utils.MetaLoad
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}
	ctx := context.Background()

	err := CallCache(cM, ctx, cacheConns, caching, args, cacheIDs, opts, false, "cgrates.org")

	if err != nil {
		t.Error(err)
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
	client := make(chan birpc.ClientConnector, 1)
	mCC := &ccMock{
		calls: map[string]func(_ *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1RemoveItems: func(_ *context.Context, args, reply interface{}) error {
				expArgs := &utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
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
			utils.CacheSv1Clear: func(_ *context.Context, args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
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
	client <- mCC

	cM := NewConnManager(defaultCfg)
	cM.AddInternalConn("cacheConn1", "", client)
	caching := utils.MetaRemove
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}
	ctx := context.Background()

	err := CallCache(cM, ctx, cacheConns, caching, args, cacheIDs, opts, false, "cgrates.org")

	if err != nil {
		t.Error(err)
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
	client := make(chan birpc.ClientConnector, 1)
	mCC := &ccMock{
		calls: map[string]func(_ *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1Clear: func(_ *context.Context, args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
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
	client <- mCC

	cM := NewConnManager(defaultCfg)
	cM.AddInternalConn("cacheConn1", "", client)
	caching := utils.MetaClear
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}
	ctx := context.Background()

	err := CallCache(cM, ctx, cacheConns, caching, args, cacheIDs, opts, false, "cgrates.org")

	if err != nil {
		t.Error(err)
	}
}

func TestGetLoadedIdsResourceProfiles(t *testing.T) {
	tpr := &TpReader{
		resProfiles: map[utils.TenantID]*utils.TPResourceProfile{
			{Tenant: "cgrates.org", ID: "ResGroup1"}: {
				TPid:              testTPID,
				Tenant:            "cgrates.org",
				ID:                "ResGroup1",
				FilterIDs:         []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
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

func TestGetLoadedIdsStatQueueProfiles(t *testing.T) {
	tpr := &TpReader{
		sqProfiles: map[utils.TenantID]*utils.TPStatProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:      testTPID,
				Tenant:    "cgrates.org",
				ID:        "ResGroup1",
				FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
				Weight:    10,
				Blocker:   true,
				Stored:    true,
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
				FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
				Weight:    10,
				Blocker:   true,
				MaxHits:   3,
				MinHits:   1,
				Async:     true,
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
				TPid:      testTPID,
				Tenant:    "cgrates.org",
				ID:        "ResGroup1",
				FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
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
				FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z", "*string:~*opts.*context:sessions"},
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
				TPid:      testTPID,
				Tenant:    "cgrates.org",
				ID:        "ResGroup1",
				FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
				RunID:     "RUN_ID",
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
				TPid:      testTPID,
				Tenant:    "cgrates.org",
				ID:        "ResGroup1",
				FilterIDs: []string{"*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
				Strategy:  utils.MetaMaxCostDisconnect,
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
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:              map[string]interface{}{},
		Tenant:               "cgrates.org",
		ResourceProfileIDs:   []string{"cgrates.org:resourceProfilesID"},
		StatsQueueProfileIDs: []string{"cgrates.org:statProfilesID"},
		ThresholdProfileIDs:  []string{"cgrates.org:thresholdProfilesID"},
		FilterIDs:            []string{"cgrates.org:filtersID"},
		RouteProfileIDs:      []string{"cgrates.org:routeProfilesID"},
		AttributeProfileIDs:  []string{"cgrates.org:attributeProfilesID"},
		ChargerProfileIDs:    []string{"cgrates.org:chargerProfilesID"},
		DispatcherProfileIDs: []string{"cgrates.org:dispatcherProfilesID"},
		DispatcherHostIDs:    []string{"cgrates.org:dispatcherHostsID"},
		ResourceIDs:          []string{"cgrates.org:resourceProfilesID"},
		StatsQueueIDs:        []string{"cgrates.org:statProfilesID"},
		ThresholdIDs:         []string{"cgrates.org:thresholdProfilesID"},

		RateProfileIDs:   []string{},
		ActionProfileIDs: []string{},
	}
	cM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(ctx *context.Context, args interface{}, reply interface{}) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(ctx *context.Context, args interface{}, reply interface{}) error {
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cM
	cnMgr := NewConnManager(cfg)
	cnMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, rpcInternal)
	tpr := &TpReader{
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
		dm: NewDataManager(data, config.CgrConfig().CacheCfg(), cnMgr),
	}
	tpr.cacheConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	if err := tpr.ReloadCache(context.Background(), utils.MetaReload, false, make(map[string]interface{}), "cgrates.org"); err != nil {
		t.Error(err)
	}
}
