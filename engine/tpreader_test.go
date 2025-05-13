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

func TestTPReaderCallCacheNoCaching(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil, nil)
	cM := NewConnManager(defaultCfg)
	args := map[string][]string{
		utils.CacheFilters:   {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
		utils.CacheResources: {},
	}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
	}
	ctx := context.Background()

	err := CallCache(cM, ctx, []string{}, utils.MetaNone, args, []string{}, opts, true, "cgrates.org")

	if err != nil {
		t.Error(err)
	}

}

func TestTPReaderCallCacheReloadCacheFirstCallErr(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan birpc.ClientConnector, 1)
	mCC := &ccMock{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.CacheSv1ReloadCache: func(_ *context.Context, args, reply any) error {
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
	client <- mCC

	cM := NewConnManager(defaultCfg)
	cM.AddInternalConn("cacheConn1", "", client)
	caching := utils.MetaReload
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
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

func TestTPReaderCallCacheReloadCacheSecondCallErr(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan birpc.ClientConnector, 1)
	mCC := &ccMock{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.CacheSv1ReloadCache: func(_ *context.Context, args, reply any) error {
				return nil
			},
			utils.CacheSv1Clear: func(_ *context.Context, args, reply any) error {
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
	client <- mCC

	cM := NewConnManager(defaultCfg)
	cM.AddInternalConn("cacheConn1", "", client)
	caching := utils.MetaReload
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
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

func TestTPReaderCallCacheLoadCache(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan birpc.ClientConnector, 1)
	mCC := &ccMock{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.CacheSv1LoadCache: func(_ *context.Context, args, reply any) error {
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
			utils.CacheSv1Clear: func(_ *context.Context, args, reply any) error {
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
	client <- mCC

	cM := NewConnManager(defaultCfg)
	cM.AddInternalConn("cacheConn1", "", client)
	caching := utils.MetaLoad
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
	}
	ctx := context.Background()

	err := CallCache(cM, ctx, cacheConns, caching, args, cacheIDs, opts, false, "cgrates.org")

	if err != nil {
		t.Error(err)
	}
}

func TestTPReaderCallCacheRemoveItems(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan birpc.ClientConnector, 1)
	mCC := &ccMock{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.CacheSv1RemoveItems: func(_ *context.Context, args, reply any) error {
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
			utils.CacheSv1Clear: func(_ *context.Context, args, reply any) error {
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
	client <- mCC

	cM := NewConnManager(defaultCfg)
	cM.AddInternalConn("cacheConn1", "", client)
	caching := utils.MetaRemove
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
	}
	ctx := context.Background()

	err := CallCache(cM, ctx, cacheConns, caching, args, cacheIDs, opts, false, "cgrates.org")

	if err != nil {
		t.Error(err)
	}
}

func TestTPReaderCallCacheClear(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan birpc.ClientConnector, 1)
	mCC := &ccMock{
		calls: map[string]func(_ *context.Context, args any, reply any) error{
			utils.CacheSv1Clear: func(_ *context.Context, args, reply any) error {
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
	client <- mCC

	cM := NewConnManager(defaultCfg)
	cM.AddInternalConn("cacheConn1", "", client)
	caching := utils.MetaClear
	args := map[string][]string{
		utils.CacheFilters: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]any{
		utils.MetaSubsys: utils.MetaChargers,
	}
	ctx := context.Background()

	err := CallCache(cM, ctx, cacheConns, caching, args, cacheIDs, opts, false, "cgrates.org")

	if err != nil {
		t.Error(err)
	}
}

func TestTPReaderGetLoadedIdsResourceProfiles(t *testing.T) {
	tpr := &TpReader{
		resProfiles: map[utils.TenantID]*utils.TPResourceProfile{
			{Tenant: "cgrates.org", ID: "ResGroup1"}: {
				TPid:              "tp_test",
				Tenant:            "cgrates.org",
				ID:                "ResGroup1",
				FilterIDs:         []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
				UsageTTL:          "1s",
				AllocationMessage: "call",
				Weights:           ";10",
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

func TestTPReaderGetLoadedIdsStatQueueProfiles(t *testing.T) {
	tpr := &TpReader{
		sqProfiles: map[utils.TenantID]*utils.TPStatProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:      "tp_test",
				Tenant:    "cgrates.org",
				ID:        "ResGroup1",
				FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
				Weights:   ";10",
				Blockers:  ";true",
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

func TestTPReaderGetLoadedIdsThresholdProfiles(t *testing.T) {
	tpr := &TpReader{
		thProfiles: map[utils.TenantID]*utils.TPThresholdProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:      "tp_test",
				Tenant:    "cgrates.org",
				ID:        "ResGroup1",
				FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
				Weights:   ";10",
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

func TestTPReaderGetLoadedIdsFilters(t *testing.T) {
	tpr := &TpReader{
		filters: map[utils.TenantID]*utils.TPFilterProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:   "tp_test",
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

func TestTPReaderGetLoadedIdsRouteProfiles(t *testing.T) {
	tpr := &TpReader{
		routeProfiles: map[utils.TenantID]*utils.TPRouteProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:      "tp_test",
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

func TestTPReaderGetLoadedIdsAttributeProfiles(t *testing.T) {
	tpr := &TpReader{
		attributeProfiles: map[utils.TenantID]*utils.TPAttributeProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:      "tp_test",
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

func TestTPReaderGetLoadedIdsChargerProfiles(t *testing.T) {
	tpr := &TpReader{
		chargerProfiles: map[utils.TenantID]*utils.TPChargerProfile{
			{Tenant: "cgrates.org", ID: "cgratesID"}: {
				TPid:      "tp_test",
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

func TestTPReaderGetLoadedIdsError(t *testing.T) {
	tpr := &TpReader{}
	errExpect := "Unsupported load category"
	if _, err := tpr.GetLoadedIds(""); err == nil || err.Error() != errExpect {
		t.Errorf("\nExpected error %v but received \n%v", errExpect, err)
	}
}

func TestTPReaderReloadCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	data.SetLoadIDsDrv(nil, make(map[string]int64))
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:              map[string]any{},
		Tenant:               "cgrates.org",
		ResourceProfileIDs:   []string{"cgrates.org:resourceProfilesID"},
		StatsQueueProfileIDs: []string{"cgrates.org:statProfilesID"},
		ThresholdProfileIDs:  []string{"cgrates.org:thresholdProfilesID"},
		FilterIDs:            []string{"cgrates.org:filtersID"},
		RouteProfileIDs:      []string{"cgrates.org:routeProfilesID"},
		AttributeProfileIDs:  []string{"cgrates.org:attributeProfilesID"},
		ChargerProfileIDs:    []string{"cgrates.org:chargerProfilesID"},
		ResourceIDs:          []string{"cgrates.org:resourceProfilesID"},
		StatsQueueIDs:        []string{"cgrates.org:statProfilesID"},
		ThresholdIDs:         []string{"cgrates.org:thresholdProfilesID"},

		RateProfileIDs:   []string{},
		ActionProfileIDs: []string{},
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
	cnMgr := NewConnManager(cfg)
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
	cnMgr.AddInternalConn(connID, utils.CacheSv1, rpcInternal)
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
		dm:         NewDataManager(data, cfg, cnMgr),
		cacheConns: []string{connID},
	}
	if err := tpr.ReloadCache(context.Background(), utils.MetaReload, false, make(map[string]any), "cgrates.org"); err != nil {
		t.Error(err)
	}
}

func TestTpReaderLoadAll(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	storeCSV := &CSVStorage{}
	db, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	tpr, err := NewTpReader(db, storeCSV, "", "", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	tprCopy, err := NewTpReader(db, storeCSV, "", "", nil, nil, true)
	if err != nil {
		t.Error(err)
	}
	if err = tpr.LoadAll(); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(tpr, tprCopy) {
		t.Errorf("Expected <%+v> , \nReceived <%+v>", tprCopy, tpr)
	}
}
