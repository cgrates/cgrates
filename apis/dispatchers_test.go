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

package apis

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestDispatchersGetDispatcherProfilesOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args1 := &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Hosts: engine.DispatcherHostProfiles{
				{
					ID: "Host1",
				},
			},
			Weight: 10,
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}

	var setReply string
	if err := admS.SetDispatcherProfile(context.Background(), args1, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	args2 := &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID2",
			Hosts: engine.DispatcherHostProfiles{
				{
					ID: "Host2",
				},
			},
			Weight: 10,
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := admS.SetDispatcherProfile(context.Background(), args2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	// this profile will not match
	args3 := &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant: "cgrates.org",
			ID:     "test2_ID1",
			Hosts: engine.DispatcherHostProfiles{
				{
					ID: "Host3",
				},
			},
			Weight: 10,
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := admS.SetDispatcherProfile(context.Background(), args3, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
	}
	exp := []*engine.DispatcherProfile{
		{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Hosts: engine.DispatcherHostProfiles{
				{
					ID: "Host1",
				},
			},
			Weight: 10,
		},
		{
			Tenant: "cgrates.org",
			ID:     "test_ID2",
			Hosts: engine.DispatcherHostProfiles{
				{
					ID: "Host2",
				},
			},
			Weight: 10,
		},
	}

	var getReply []*engine.DispatcherProfile
	if err := admS.GetDispatcherProfiles(context.Background(), argsGet, &getReply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(getReply, func(i, j int) bool {
			return getReply[i].ID < getReply[j].ID
		})
		if !reflect.DeepEqual(getReply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(exp), utils.ToJSON(getReply))
		}
	}
}

func TestDispatchersSetGetRemDispatcherProfile(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "dspID",
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	var result engine.DispatcherProfile
	var reply string

	dspPrf := &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:    "cgrates.org",
			ID:        "dspID",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Hosts: engine.DispatcherHostProfiles{
				{
					ID: "Host1",
				},
			},
			Weight: 10,
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}

	if err := adms.SetDispatcherProfile(context.Background(), dspPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	if err := adms.GetDispatcherProfile(context.Background(), arg, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, *dspPrf.DispatcherProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(dspPrf.DispatcherProfile), utils.ToJSON(result))
	}

	var dspPrfIDs []string
	expDspPrfIDs := []string{"dspID"}

	if err := adms.GetDispatcherProfileIDs(context.Background(), &utils.ArgsItemIDs{},
		&dspPrfIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dspPrfIDs, expDspPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expDspPrfIDs, dspPrfIDs)
	}

	var rplyCount int

	if err := adms.GetDispatcherProfilesCount(context.Background(), &utils.ArgsItemIDs{},
		&rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != len(dspPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", len(dspPrfIDs), rplyCount)
	}

	if err := adms.RemoveDispatcherProfile(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	}

	engine.Cache.Clear(nil)
	if err := adms.GetDispatcherProfile(context.Background(), arg, &result); err == nil ||
		err != utils.ErrDSPProfileNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv engine.DispatcherProfile
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.GetDispatcherProfile(context.Background(), &utils.TenantIDWithAPIOpts{}, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestDispatchersGetDispatcherProfileCheckErrors",
		},
	}

	if err := adms.GetDispatcherProfile(context.Background(), arg, &rcv); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersSetDispatcherProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	dspPrf := &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{},
	}

	var reply string
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.SetDispatcherProfile(context.Background(), dspPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dspPrf.ID = "TestDispatchersSetDispatcherProfileCheckErrors"
	dspPrf.FilterIDs = []string{"invalid_filter_format"}
	experr = "SERVER_ERROR: broken reference to filter: <invalid_filter_format> for item with ID: cgrates.org:TestDispatchersSetDispatcherProfileCheckErrors"

	if err := adms.SetDispatcherProfile(context.Background(), dspPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dspPrf.FilterIDs = []string{}
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	experr = "SERVER_ERROR: context deadline exceeded"
	cfg.GeneralCfg().DefaultCaching = utils.MetaRemove
	if err := adms.SetDispatcherProfile(ctx, dspPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>,\nreceived <%+v>", experr, err)
	}
	cancel()

	dbMock := &engine.DataDBMock{
		GetDispatcherProfileDrvF: func(*context.Context, string, string) (*engine.DispatcherProfile, error) {
			dspPrf := &engine.DispatcherProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return dspPrf, nil
		},
		SetDispatcherProfileDrvF: func(*context.Context, *engine.DispatcherProfile) error {
			return nil
		},
		RemoveDispatcherProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
	}

	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.SetDispatcherProfile(context.Background(), dspPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersRemoveDispatcherProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	dspPrf := &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			ID:     "TestDispatchersRemoveDispatcherProfileCheckErrors",
			Tenant: "cgrates.org",
			Hosts: engine.DispatcherHostProfiles{
				{
					ID: "HOST",
				},
			},
			Weight: 10,
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	var reply string

	if err := adms.SetDispatcherProfile(context.Background(), dspPrf, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	adms.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
	experr := "SERVER_ERROR: context deadline exceeded"

	if err := adms.RemoveDispatcherProfile(ctx, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestDispatchersRemoveDispatcherProfileCheckErrors",
		},
	}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
	cancel()

	adms.cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	var rcv engine.DispatcherProfile

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestDispatchersRemoveDispatcherProfileCheckErrors",
		},
	}

	if err := adms.GetDispatcherProfile(context.Background(), arg, &rcv); err == nil || err != utils.ErrDSPProfileNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrDSPProfileNotFound, err)
	}

	experr = "MANDATORY_IE_MISSING: [ID]"

	if err := adms.RemoveDispatcherProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.RemoveDispatcherProfile(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestDispatchersRemoveDispatcherProfileCheckErrors",
		}, APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dbMock := &engine.DataDBMock{
		GetDispatcherProfileDrvF: func(*context.Context, string, string) (*engine.DispatcherProfile, error) {
			dspPrf := &engine.DispatcherProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return dspPrf, nil
		},
		SetDispatcherProfileDrvF: func(*context.Context, *engine.DispatcherProfile) error {
			return nil
		},
		RemoveDispatcherProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, nil
		},
		RemoveDispatcherHostDrvF: func(ctx *context.Context, tnt, id string) error {
			return nil
		},
	}

	engine.Cache.Clear(nil)
	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.RemoveDispatcherProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestDispatchersRemoveDispatcherProfileCheckErrors",
			}, APIOpts: map[string]any{
				utils.MetaCache: utils.MetaNone,
			}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherProfileIDsErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetDispatcherProfileDrvF: func(*context.Context, string, string) (*engine.DispatcherProfile, error) {
			dspPrf := &engine.DispatcherProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return dspPrf, nil
		},
		SetDispatcherProfileDrvF: func(*context.Context, *engine.DispatcherProfile) error {
			return nil
		},
		RemoveDispatcherProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "NOT_IMPLEMENTED"

	if err := adms.GetDispatcherProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherProfileIDsErrKeys(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{}, nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string

	if err := adms.GetDispatcherProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherProfilesCountErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetDispatcherProfileDrvF: func(*context.Context, string, string) (*engine.DispatcherProfile, error) {
			dspPrf := &engine.DispatcherProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return dspPrf, nil
		},
		SetDispatcherProfileDrvF: func(*context.Context, *engine.DispatcherProfile) error {
			return nil
		},
		RemoveDispatcherProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetDispatcherProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestDispatchersGetDispatcherProfilesCountErrKeys(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{}, nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetDispatcherProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestDispatchersGetDispatcherProfileIDsGetOptsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetDispatcherProfileDrvF: func(*context.Context, string, string) (*engine.DispatcherProfile, error) {
			dspPrf := &engine.DispatcherProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return dspPrf, nil
		},
		SetDispatcherProfileDrvF: func(*context.Context, *engine.DispatcherProfile) error {
			return nil
		},
		RemoveDispatcherProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"dpp_cgrates.org:key1", "dpp_cgrates.org:key2", "dpp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "cannot convert field<bool>: true to int"

	if err := adms.GetDispatcherProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
			APIOpts: map[string]any{
				utils.PageLimitOpt: true,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherProfileIDsPaginateErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetDispatcherProfileDrvF: func(*context.Context, string, string) (*engine.DispatcherProfile, error) {
			dspPrf := &engine.DispatcherProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return dspPrf, nil
		},
		SetDispatcherProfileDrvF: func(*context.Context, *engine.DispatcherProfile) error {
			return nil
		},
		RemoveDispatcherProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"dpp_cgrates.org:key1", "dpp_cgrates.org:key2", "dpp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := `SERVER_ERROR: maximum number of items exceeded`

	if err := adms.GetDispatcherProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
			APIOpts: map[string]any{
				utils.PageLimitOpt:    2,
				utils.PageOffsetOpt:   4,
				utils.PageMaxItemsOpt: 5,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherHostsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args1 := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:         "test_ID1",
				Reconnects: -1,
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetDispatcherHost(context.Background(), args1, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	args2 := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:         "test_ID2",
				Reconnects: -1,
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetDispatcherHost(context.Background(), args2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	// this Host will not match
	args3 := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:         "test2_ID1",
				Reconnects: -1,
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetDispatcherHost(context.Background(), args3, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
	}
	exp := []*engine.DispatcherHost{
		{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:         "test_ID1",
				Reconnects: -1,
			},
		},
		{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:         "test_ID2",
				Reconnects: -1,
			},
		},
	}

	var getReply []*engine.DispatcherHost
	if err := admS.GetDispatcherHosts(context.Background(), argsGet, &getReply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(getReply, func(i, j int) bool {
			return getReply[i].ID < getReply[j].ID
		})
		if !reflect.DeepEqual(getReply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(exp), utils.ToJSON(getReply))
		}
	}
}

func TestDispatchersSetGetRemDispatcherHost(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "dspHost1",
		},
	}
	var result engine.DispatcherHost
	var reply string

	dspHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:         "dspHost1",
				Reconnects: -1,
			},
		},
		APIOpts: nil,
	}

	if err := adms.SetDispatcherHost(context.Background(), dspHost, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	if err := adms.GetDispatcherHost(context.Background(), arg, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, *dspHost.DispatcherHost) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(dspHost.DispatcherHost), utils.ToJSON(result))
	}

	var dspHostIDs []string
	expDspHostIDs := []string{"dspHost1"}

	if err := adms.GetDispatcherHostIDs(context.Background(), &utils.ArgsItemIDs{},
		&dspHostIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dspHostIDs, expDspHostIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expDspHostIDs, dspHostIDs)
	}

	var rplyCount int

	if err := adms.GetDispatcherHostsCount(context.Background(), &utils.ArgsItemIDs{},
		&rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != len(dspHostIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", len(dspHostIDs), rplyCount)
	}

	if err := adms.RemoveDispatcherHost(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	}

	engine.Cache.Clear(nil)
	if err := adms.GetDispatcherHost(context.Background(), arg, &result); err == nil ||
		err != utils.ErrDSPHostNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrDSPHostNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherHostCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv engine.DispatcherHost
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.GetDispatcherHost(context.Background(), &utils.TenantIDWithAPIOpts{}, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestDispatchersGetDispatcherHostCheckErrors",
		},
	}

	if err := adms.GetDispatcherHost(context.Background(), arg, &rcv); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersSetDispatcherHostCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	dspHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			RemoteHost: &config.RemoteHost{},
		},
	}

	var reply string
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.SetDispatcherHost(context.Background(), dspHost, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dspHost.ID = "TestDispatchersSetDispatcherHostCheckErrors"
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	experr = "SERVER_ERROR: context deadline exceeded"
	cfg.GeneralCfg().DefaultCaching = utils.MetaRemove
	if err := adms.SetDispatcherHost(ctx, dspHost, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>,\nreceived <%+v>", experr, err)
	}
	cancel()

	dbMock := &engine.DataDBMock{
		GetDispatcherHostDrvF: func(*context.Context, string, string) (*engine.DispatcherHost, error) {
			dspHost := &engine.DispatcherHost{
				Tenant: "cgrates.org",
				RemoteHost: &config.RemoteHost{
					ID: "TEST",
				},
			}
			return dspHost, nil
		},
		SetDispatcherHostDrvF: func(*context.Context, *engine.DispatcherHost) error {
			return nil
		},
		RemoveDispatcherHostDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
	}

	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.SetDispatcherHost(context.Background(), dspHost, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersRemoveDispatcherHostCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	dspHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			RemoteHost: &config.RemoteHost{
				ID: "TestDispatchersRemoveDispatcherHostCheckErrors",
			},
			Tenant: "cgrates.org",
		},
	}
	var reply string

	if err := adms.SetDispatcherHost(context.Background(), dspHost, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	adms.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
	experr := "SERVER_ERROR: context deadline exceeded"

	if err := adms.RemoveDispatcherHost(ctx, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestDispatchersRemoveDispatcherHostCheckErrors",
		},
	}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
	cancel()

	adms.cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	var rcv engine.DispatcherHost

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestDispatchersRemoveDispatcherHostCheckErrors",
		},
	}

	if err := adms.GetDispatcherHost(context.Background(), arg, &rcv); err == nil || err != utils.ErrDSPHostNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrDSPHostNotFound, err)
	}

	experr = "MANDATORY_IE_MISSING: [ID]"

	if err := adms.RemoveDispatcherHost(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.RemoveDispatcherHost(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestDispatchersRemoveDispatcherHostCheckErrors",
		}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dbMock := &engine.DataDBMock{
		GetDispatcherHostDrvF: func(*context.Context, string, string) (*engine.DispatcherHost, error) {
			dspHost := &engine.DispatcherHost{
				Tenant: "cgrates.org",
				RemoteHost: &config.RemoteHost{
					ID: "TEST",
				},
			}
			return dspHost, nil
		},
		SetDispatcherHostDrvF: func(*context.Context, *engine.DispatcherHost) error {
			return nil
		},
		RemoveDispatcherHostDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, nil
		},
	}

	engine.Cache.Clear(nil)
	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.RemoveDispatcherHost(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestDispatchersRemoveDispatcherHostCheckErrors",
			}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherHostIDsErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetDispatcherHostDrvF: func(*context.Context, string, string) (*engine.DispatcherHost, error) {
			thPrf := &engine.DispatcherHost{
				Tenant: "cgrates.org",
				RemoteHost: &config.RemoteHost{
					ID: "TEST",
				},
			}
			return thPrf, nil
		},
		SetDispatcherHostDrvF: func(*context.Context, *engine.DispatcherHost) error {
			return nil
		},
		RemoveDispatcherHostDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "NOT_IMPLEMENTED"

	if err := adms.GetDispatcherHostIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherHostIDsErrKeys(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{}, nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string

	if err := adms.GetDispatcherHostIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherHostIDsGetOptsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetDispatcherHostDrvF: func(*context.Context, string, string) (*engine.DispatcherHost, error) {
			dspHost := &engine.DispatcherHost{
				Tenant: "cgrates.org",
				RemoteHost: &config.RemoteHost{
					ID: "TEST",
				},
			}
			return dspHost, nil
		},
		SetDispatcherHostDrvF: func(*context.Context, *engine.DispatcherHost) error {
			return nil
		},
		RemoveDispatcherHostDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"dpp_cgrates.org:key1", "dpp_cgrates.org:key2", "dpp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "cannot convert field<bool>: true to int"

	if err := adms.GetDispatcherHostIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
			APIOpts: map[string]any{
				utils.PageLimitOpt: true,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherHostIDsPaginateErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetDispatcherHostDrvF: func(*context.Context, string, string) (*engine.DispatcherHost, error) {
			dspHost := &engine.DispatcherHost{
				Tenant: "cgrates.org",
				RemoteHost: &config.RemoteHost{
					ID: "TEST",
				},
			}
			return dspHost, nil
		},
		SetDispatcherHostDrvF: func(*context.Context, *engine.DispatcherHost) error {
			return nil
		},
		RemoveDispatcherHostDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"dpp_cgrates.org:key1", "dpp_cgrates.org:key2", "dpp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := `SERVER_ERROR: maximum number of items exceeded`

	if err := adms.GetDispatcherHostIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
			APIOpts: map[string]any{
				utils.PageLimitOpt:    2,
				utils.PageOffsetOpt:   4,
				utils.PageMaxItemsOpt: 5,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherHostsCountErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetDispatcherHostDrvF: func(*context.Context, string, string) (*engine.DispatcherHost, error) {
			thPrf := &engine.DispatcherHost{
				Tenant: "cgrates.org",
				RemoteHost: &config.RemoteHost{
					ID: "TEST",
				},
			}
			return thPrf, nil
		},
		SetDispatcherHostDrvF: func(*context.Context, *engine.DispatcherHost) error {
			return nil
		},
		RemoveDispatcherHostDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetDispatcherHostsCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestDispatchersGetDispatcherHostsCountErrKeys(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{}, nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetDispatcherHostsCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestDispatchersGetDispatcherProfilesGetIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args := &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Hosts: engine.DispatcherHostProfiles{
				{
					ID: "Host1",
				},
			},
			Weight: 10,
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}

	var setReply string
	if err := admS.SetDispatcherProfile(context.Background(), args, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
		APIOpts: map[string]any{
			utils.PageLimitOpt:    2,
			utils.PageOffsetOpt:   4,
			utils.PageMaxItemsOpt: 5,
		},
	}

	experr := `SERVER_ERROR: maximum number of items exceeded`
	var getReply []*engine.DispatcherProfile
	if err := admS.GetDispatcherProfiles(context.Background(), argsGet, &getReply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestDispatchersGetDispatcherProfilesGetProfileErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		SetDispatcherProfileDrvF: func(*context.Context, *engine.DispatcherProfile) error {
			return nil
		},
		RemoveDispatcherProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"dpp_cgrates.org:TEST"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []*engine.DispatcherProfile
	experr := "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.GetDispatcherProfiles(context.Background(),
		&utils.ArgsItemIDs{
			ItemsPrefix: "TEST",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersGetDispatcherHostsGetIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:         "test_ID1",
				Reconnects: -1,
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetDispatcherHost(context.Background(), args, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
		APIOpts: map[string]any{
			utils.PageLimitOpt:    2,
			utils.PageOffsetOpt:   4,
			utils.PageMaxItemsOpt: 5,
		},
	}

	experr := `SERVER_ERROR: maximum number of items exceeded`
	var getReply []*engine.DispatcherHost
	if err := admS.GetDispatcherHosts(context.Background(), argsGet, &getReply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestDispatchersGetDispatcherHostsGetHostErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		SetDispatcherHostDrvF: func(*context.Context, *engine.DispatcherHost) error {
			return nil
		},
		RemoveDispatcherHostDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"dph_cgrates.org:TEST"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []*engine.DispatcherHost
	experr := "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.GetDispatcherHosts(context.Background(),
		&utils.ArgsItemIDs{
			ItemsPrefix: "TEST",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestDispatchersSetDispatcherHostErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	adms := &AdminSv1{
		cfg: cfg,
	}

	dspHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			RemoteHost: &config.RemoteHost{
				ID: "TEST",
			},
		},
	}

	var reply string
	experr := "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.SetDispatcherHost(context.Background(), dspHost, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

}
