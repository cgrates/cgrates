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
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestThresholdsSetGetRemThresholdProfile(t *testing.T) {
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
			ID: "thdID",
		},
	}
	var result engine.ThresholdProfile
	var reply string

	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "thdID",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			MaxHits:   10,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}

	if err := adms.SetThresholdProfile(context.Background(), thPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	if err := adms.GetThresholdProfile(context.Background(), arg, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, *thPrf.ThresholdProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(thPrf.ThresholdProfile), utils.ToJSON(result))
	}

	var thPrfIDs []string
	expThPrfIDs := []string{"thdID"}

	if err := adms.GetThresholdProfileIDs(context.Background(), &utils.ArgsItemIDs{},
		&thPrfIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thPrfIDs, expThPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expThPrfIDs, thPrfIDs)
	}

	var rplyCount int

	if err := adms.GetThresholdProfilesCount(context.Background(), &utils.ArgsItemIDs{},
		&rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != len(thPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", len(thPrfIDs), rplyCount)
	}

	if err := adms.RemoveThresholdProfile(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	}

	engine.Cache.Clear(nil)
	if err := adms.GetThresholdProfile(context.Background(), arg, &result); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestThresholdsGetThresholdProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv engine.ThresholdProfile
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.GetThresholdProfile(context.Background(), &utils.TenantIDWithAPIOpts{}, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestThresholdsGetThresholdProfileCheckErrors",
		},
	}

	if err := adms.GetThresholdProfile(context.Background(), arg, &rcv); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestThresholdsSetThresholdProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{},
	}

	var reply string
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.SetThresholdProfile(context.Background(), thPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	thPrf.ID = "TestThresholdsSetThresholdProfileCheckErrors"
	thPrf.FilterIDs = []string{"invalid_filter_format"}
	experr = "SERVER_ERROR: broken reference to filter: <invalid_filter_format> for item with ID: cgrates.org:TestThresholdsSetThresholdProfileCheckErrors"

	if err := adms.SetThresholdProfile(context.Background(), thPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	thPrf.FilterIDs = []string{}
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	experr = "SERVER_ERROR: context deadline exceeded"
	cfg.GeneralCfg().DefaultCaching = utils.MetaRemove
	if err := adms.SetThresholdProfile(ctx, thPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>,\nreceived <%+v>", experr, err)
	}
	cancel()

	dbMock := &engine.DataDBMock{
		GetThresholdProfileDrvF: func(*context.Context, string, string) (*engine.ThresholdProfile, error) {
			thPrf := &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return thPrf, nil
		},
		SetThresholdProfileDrvF: func(*context.Context, *engine.ThresholdProfile) error {
			return nil
		},
		RemThresholdProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
	}

	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.SetThresholdProfile(context.Background(), thPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestThresholdsRemoveThresholdProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			ID:      "TestThresholdsRemoveThresholdProfileCheckErrors",
			Tenant:  "cgrates.org",
			MaxHits: 10,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}
	var reply string

	if err := adms.SetThresholdProfile(context.Background(), thPrf, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	adms.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
	experr := "SERVER_ERROR: context deadline exceeded"

	if err := adms.RemoveThresholdProfile(ctx, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestThresholdsRemoveThresholdProfileCheckErrors",
		},
	}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
	cancel()

	adms.cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	var rcv engine.ThresholdProfile

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestThresholdsRemoveThresholdProfileCheckErrors",
		},
	}

	if err := adms.GetThresholdProfile(context.Background(), arg, &rcv); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	experr = "MANDATORY_IE_MISSING: [ID]"

	if err := adms.RemoveThresholdProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.RemoveThresholdProfile(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestThresholdsRemoveThresholdProfileCheckErrors",
		}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dbMock := &engine.DataDBMock{
		GetThresholdProfileDrvF: func(*context.Context, string, string) (*engine.ThresholdProfile, error) {
			thPrf := &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return thPrf, nil
		},
		SetThresholdProfileDrvF: func(*context.Context, *engine.ThresholdProfile) error {
			return nil
		},
		RemThresholdProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		RemoveThresholdDrvF: func(ctx *context.Context, tnt, id string) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, nil
		},
	}

	engine.Cache.Clear(nil)
	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.RemoveThresholdProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestThresholdsRemoveThresholdProfileCheckErrors",
			}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestThresholdsGetThresholdProfileIDsErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetThresholdProfileDrvF: func(*context.Context, string, string) (*engine.ThresholdProfile, error) {
			thPrf := &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return thPrf, nil
		},
		SetThresholdProfileDrvF: func(*context.Context, *engine.ThresholdProfile) error {
			return nil
		},
		RemThresholdProfileDrvF: func(*context.Context, string, string) error {
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

	if err := adms.GetThresholdProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestThresholdsGetThresholdProfileIDsErrKeys(t *testing.T) {
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

	if err := adms.GetThresholdProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestThresholdsGetThresholdProfileIDsGetOptsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetThresholdProfileDrvF: func(*context.Context, string, string) (*engine.ThresholdProfile, error) {
			thPrf := &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return thPrf, nil
		},
		SetThresholdProfileDrvF: func(*context.Context, *engine.ThresholdProfile) error {
			return nil
		},
		RemThresholdProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"thp_cgrates.org:key1", "thp_cgrates.org:key2", "thp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "cannot convert field<bool>: true to int"

	if err := adms.GetThresholdProfileIDs(context.Background(),
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

func TestThresholdsGetThresholdProfileIDsPaginateErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetThresholdProfileDrvF: func(*context.Context, string, string) (*engine.ThresholdProfile, error) {
			thPrf := &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return thPrf, nil
		},
		SetThresholdProfileDrvF: func(*context.Context, *engine.ThresholdProfile) error {
			return nil
		},
		RemThresholdProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"rpp_cgrates.org:key1", "rpp_cgrates.org:key2", "rpp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := `SERVER_ERROR: maximum number of items exceeded`

	if err := adms.GetThresholdProfileIDs(context.Background(),
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

func TestThresholdsGetThresholdProfilesCountErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetThresholdProfileDrvF: func(*context.Context, string, string) (*engine.ThresholdProfile, error) {
			thPrf := &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return thPrf, nil
		},
		SetThresholdProfileDrvF: func(*context.Context, *engine.ThresholdProfile) error {
			return nil
		},
		RemThresholdProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetThresholdProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestThresholdsGetThresholdProfilesCountErrKeys(t *testing.T) {
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

	if err := adms.GetThresholdProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsNewThresholdSv1(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	tS := engine.NewThresholdService(dm, cfg, nil, nil)

	exp := &ThresholdSv1{
		tS: tS,
	}
	rcv := NewThresholdSv1(tS)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestThresholdsSv1Ping(t *testing.T) {
	thSv1 := new(ThresholdSv1)
	var reply string
	if err := thSv1.Ping(nil, nil, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply error")
	}
}

func TestThresholdsAPIs(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ThresholdSCfg().ActionSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions)}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"thd1", "thd2"},
			utils.OptsActionsProfileIDs:    []string{"actPrfID"},
		},
	}
	mCC := &mockClientConn{
		calls: map[string]func(*context.Context, any, any) error{
			utils.ActionSv1ExecuteActions: func(ctx *context.Context, args, reply any) error {
				if !reflect.DeepEqual(args, expEv) {
					return fmt.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expEv), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- mCC
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), utils.ActionSv1, rpcInternal)

	tS := engine.NewThresholdService(dm, cfg, fltrs, cM)
	adms := &AdminSv1{
		dm:  dm,
		cfg: cfg,
	}

	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant:    "cgrates.org",
			ID:        "actPrfID",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Actions: []*engine.APAction{
				{
					ID: "actID",
				},
			},
		},
	}

	var reply string
	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err != nil {
		t.Error(err)
	}

	thPrf1 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "thd1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			MaxHits:   10,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			ActionProfileIDs: []string{"actPrfID"},
		},
	}

	if err := adms.SetThresholdProfile(context.Background(), thPrf1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	thPrf2 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:  "cgrates.org",
			ID:      "thd2",
			MaxHits: 10,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			ActionProfileIDs: []string{"actPrfID"},
		},
	}

	if err := adms.SetThresholdProfile(context.Background(), thPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	tSv1 := NewThresholdSv1(tS)
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		ID: "EventTest",
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"thd1", "thd2"},
		},
	}

	expThresholds := engine.Thresholds{
		{
			Tenant: "cgrates.org",
			ID:     "thd1",
		},
		{
			Tenant: "cgrates.org",
			ID:     "thd2",
		},
	}

	var rplyThresholds engine.Thresholds
	if err := tSv1.GetThresholdsForEvent(context.Background(), args, &rplyThresholds); err != nil {
		t.Error(err)
	} else {
		sort.Slice(rplyThresholds, func(i, j int) bool {
			return rplyThresholds[i].ID < rplyThresholds[j].ID
		})
		// We compare JSONs because the received Thresholds have unexported fields
		if utils.ToJSON(expThresholds) != utils.ToJSON(rplyThresholds) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expThresholds), utils.ToJSON(rplyThresholds))
		}
	}

	expThreshold := engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "thd1",
	}

	var rplyThreshold engine.Threshold
	if err := tSv1.GetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "thd1",
	}}, &rplyThreshold); err != nil {
		t.Error(err)
	} else {
		// We compare JSONs because the received Threshold has unexported fields
		if utils.ToJSON(expThreshold) != utils.ToJSON(rplyThreshold) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expThreshold), utils.ToJSON(rplyThreshold))
		}
	}

	expIDs := []string{"thd1", "thd2"}
	tIDs := make([]string, 2)
	if err := tSv1.GetThresholdIDs(context.Background(), &utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
	}, &tIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(tIDs)
		if !reflect.DeepEqual(tIDs, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expIDs), utils.ToJSON(tIDs))
		}
	}

	if err := tSv1.ProcessEvent(context.Background(), args, &tIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(tIDs)
		if !reflect.DeepEqual(tIDs, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expIDs), utils.ToJSON(tIDs))
		}
	}

	if err := tSv1.ResetThreshold(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "thd1",
	}}, &reply); err != nil {
		t.Error(err)
	}
}

func TestThresholdsGetThresholdProfilesOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args1 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "test_ID1",
			MaxHits:          5,
			MinHits:          1,
			ActionProfileIDs: []string{utils.MetaNone},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetThresholdProfile(context.Background(), args1, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	args2 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "test_ID2",
			MaxHits:          4,
			MinHits:          2,
			ActionProfileIDs: []string{utils.MetaNone},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetThresholdProfile(context.Background(), args2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	// this profile will not match
	args3 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "test2_ID1",
			MaxHits:          5,
			MinHits:          1,
			ActionProfileIDs: []string{utils.MetaNone},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetThresholdProfile(context.Background(), args3, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
	}
	exp := []*engine.ThresholdProfile{
		{
			Tenant:           "cgrates.org",
			ID:               "test_ID1",
			MaxHits:          5,
			MinHits:          1,
			ActionProfileIDs: []string{utils.MetaNone},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		{
			Tenant:           "cgrates.org",
			ID:               "test_ID2",
			MaxHits:          4,
			MinHits:          2,
			ActionProfileIDs: []string{utils.MetaNone},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}

	var getReply []*engine.ThresholdProfile
	if err := admS.GetThresholdProfiles(context.Background(), argsGet, &getReply); err != nil {
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

func TestThresholdsGetThresholdProfilesGetIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "test_ID1",
			MaxHits:          5,
			MinHits:          1,
			ActionProfileIDs: []string{utils.MetaNone},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetThresholdProfile(context.Background(), args, &setReply); err != nil {
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
	var getReply []*engine.ThresholdProfile
	if err := admS.GetThresholdProfiles(context.Background(), argsGet, &getReply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestThresholdsGetThresholdProfilesGetProfileErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		SetThresholdProfileDrvF: func(*context.Context, *engine.ThresholdProfile) error {
			return nil
		},
		RemThresholdProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"thp_cgrates.org:TEST"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []*engine.ThresholdProfile
	experr := "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.GetThresholdProfiles(context.Background(),
		&utils.ArgsItemIDs{
			ItemsPrefix: "TEST",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}
