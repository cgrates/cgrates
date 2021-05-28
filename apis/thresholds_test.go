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
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestThresholdsSetGetRemThresholdProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	arg := &utils.TenantID{
		ID: "thdID",
	}
	var result engine.ThresholdProfile
	var reply string

	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:  "cgrates.org",
			ID:      "thdID",
			MaxHits: 10,
			Weight:  10,
		},
	}

	if err := adms.SetThresholdProfile(context.Background(), thPrf, &reply); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	if err := adms.GetThresholdProfile(context.Background(), arg, &result); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(result, *thPrf.ThresholdProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(thPrf.ThresholdProfile), utils.ToJSON(result))
	}

	var thPrfIDs []string
	expThPrfIDs := []string{"thdID"}

	if err := adms.GetThresholdProfileIDs(context.Background(), &utils.PaginatorWithTenant{},
		&thPrfIDs); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(thPrfIDs, expThPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expThPrfIDs, thPrfIDs)
	}

	var rplyCount int

	if err := adms.GetThresholdProfileCount(context.Background(), &utils.TenantWithAPIOpts{},
		&rplyCount); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if rplyCount != len(thPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", len(thPrfIDs), rplyCount)
	}

	argRem := &utils.TenantIDWithAPIOpts{
		TenantID: arg,
	}

	if err := adms.RemoveThresholdProfile(context.Background(), argRem, &reply); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if err := adms.GetThresholdProfile(context.Background(), arg, &result); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestThresholdsGetThresholdProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv engine.ThresholdProfile
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.GetThresholdProfile(context.Background(), &utils.TenantID{}, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.GetThresholdProfile(context.Background(), &utils.TenantID{
		ID: "TestThresholdsGetThresholdProfileCheckErrors",
	}, &rcv); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestThresholdsSetThresholdProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
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
	experr = "SERVER_ERROR: broken reference to filter: invalid_filter_format for item with ID: cgrates.org:TestThresholdsSetThresholdProfileCheckErrors"

	if err := adms.SetThresholdProfile(context.Background(), thPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	thPrf.FilterIDs = []string{}
	adms.connMgr = engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): make(chan birpc.ClientConnector),
	})
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
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
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
			Weight:  10,
		},
	}
	var reply string

	if err := adms.SetThresholdProfile(context.Background(), thPrf, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	adms.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	adms.connMgr = engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): make(chan birpc.ClientConnector),
	})
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

	if err := adms.GetThresholdProfile(context.Background(), &utils.TenantID{
		ID: "TestThresholdsRemoveThresholdProfileCheckErrors",
	}, &rcv); err == nil || err != utils.ErrNotFound {
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
	}

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
		&utils.PaginatorWithTenant{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestThresholdsGetThresholdProfileIDsErrKeys(t *testing.T) {
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
		&utils.PaginatorWithTenant{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestThresholdsGetThresholdProfileIDsCountErrMock(t *testing.T) {
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

	if err := adms.GetThresholdProfileCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestThresholdsGetThresholdProfileIDsCountErrKeys(t *testing.T) {
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

	if err := adms.GetThresholdProfileCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestThresholdsNewThresholdSv1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, true)
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
