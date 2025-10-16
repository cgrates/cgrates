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

package apis

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

func TestRatesGetRateProfileErrMandatoryIeMissing(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{},
	}
	rply := &utils.RateProfile{}
	err := admS.GetRateProfile(context.Background(), args, rply)
	expected := "MANDATORY_IE_MISSING: [ID]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, rply)
	}
}

func TestRatesGetRateProfile1(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "DefaultRate",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}

	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "DefaultRate",
		},
	}
	var result utils.RateProfile

	expected := &utils.RateProfile{
		ID:        "DefaultRate",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
			},
		},
	}
	err = admS.GetRateProfile(context.Background(), args, &result)
	rslt := &result

	if !reflect.DeepEqual(expected.Rates["RT_WEEK"].ID, rslt.Rates["RT_WEEK"].ID) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rslt))
	}
	if !reflect.DeepEqual(expected.Rates["RT_WEEK"].ActivationTimes, rslt.Rates["RT_WEEK"].ActivationTimes) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rslt))
	}
	expected.Rates = nil
	rslt.Rates = nil
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	} else if !reflect.DeepEqual(expected, rslt) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rslt))
	}

}

func TestRatesGetRateProfileErrorNotFound(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "2",
		},
	}
	var result utils.RateProfile
	err := admS.GetRateProfile(context.Background(), args, &result)
	expected := utils.ErrNotFound
	if err == nil || err != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestRatesGetRateProfileIDs(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "RP1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}

	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.ArgsItemIDs{}
	result := &[]string{}
	expected := &[]string{"RP1"}
	err = admS.GetRateProfileIDs(context.Background(), args, result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestRatesGetRateProfile2(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "RP2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}

	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.ArgsItemIDs{
		Tenant: "tenant",
	}
	result := &[]string{}
	expected := &[]string{"RP2"}
	err = admS.GetRateProfileIDs(context.Background(), args, result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestRatesGetRateProfileErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDBMock := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDBMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.ArgsItemIDs{
		Tenant: "tenant",
	}
	result := &[]string{}
	err := admS.GetRateProfileIDs(context.Background(), args, result)
	if err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestRatesGetRateProfileErr2(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDBMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(*context.Context, string) ([]string, error) {
			return []string{}, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDBMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.ArgsItemIDs{
		Tenant: "tenant",
	}
	result := &[]string{}
	err := admS.GetRateProfileIDs(context.Background(), args, result)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestRatesGetRateProfilesCount(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "RP3",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.ArgsItemIDs{
		Tenant: "tenant",
	}
	result := utils.IntPointer(0)
	expected := utils.IntPointer(1)
	err = admS.GetRateProfilesCount(context.Background(), args, result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(&result, &expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", &expected, &result)
	}
}

func TestRatesGetRateProfilesCountEmptyTenant(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "RP4",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}

	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.ArgsItemIDs{}
	result := utils.IntPointer(0)
	expected := utils.IntPointer(1)
	err = admS.GetRateProfilesCount(context.Background(), args, result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(&result, &expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", &expected, &result)
	}
}

func TestRatesGetRateProfilesCountGetKeysError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDBMock := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDBMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.ArgsItemIDs{}
	result := utils.IntPointer(0)
	err := admS.GetRateProfilesCount(context.Background(), args, result)
	if err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestRatesGetRateProfilesCountKeysLenError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDBMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(*context.Context, string) ([]string, error) {
			return []string{}, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDBMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.ArgsItemIDs{}
	result := utils.IntPointer(0)
	err := admS.GetRateProfilesCount(context.Background(), args, result)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestRatesSetRateProfileMissingStructFieldError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	expected := "MANDATORY_IE_MISSING: [ID]"
	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestRatesSetRateProfileEmptyTenant(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "2",
		},
	}
	var result utils.RateProfile
	err = admS.GetRateProfile(context.Background(), args, &result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expected := utils.RateProfile{
		ID:        "2",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
			},
		},
	}
	if !reflect.DeepEqual(result.Rates["RT_WEEK"].ID, expected.Rates["RT_WEEK"].ID) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.Rates["RT_WEEK"].ID, result.Rates["RT_WEEK"].ID)
	}
	if !reflect.DeepEqual(result.Rates["RT_WEEK"].ActivationTimes, expected.Rates["RT_WEEK"].ActivationTimes) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.Rates["RT_WEEK"].ActivationTimes, result.Rates["RT_WEEK"].ActivationTimes)
	}
	if !reflect.DeepEqual("cgrates.org:2:RT_WEEK", result.Rates["RT_WEEK"].UID()) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "cgrates.org:2:RT_WEEK", result.Rates["RT_WEEK"].UID())
	}
	result.Rates = nil
	expected.Rates = nil
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestRatesSetRateProfileError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "RP6",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}

	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "RP6",
		},
	}
	var result utils.RateProfile
	err = admS.GetRateProfile(context.Background(), args, &result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expected := utils.RateProfile{
		ID:        "RP6",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
			},
		},
	}
	if !reflect.DeepEqual(result.Rates["RT_WEEK"].ID, expected.Rates["RT_WEEK"].ID) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.Rates["RT_WEEK"].ID, result.Rates["RT_WEEK"].ID)
	}
	if !reflect.DeepEqual(result.Rates["RT_WEEK"].ActivationTimes, expected.Rates["RT_WEEK"].ActivationTimes) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.Rates["RT_WEEK"].ActivationTimes, result.Rates["RT_WEEK"].ActivationTimes)
	}
	if result.Rates["RT_WEEK"].UID() != "cgrates.org:RP6:RT_WEEK" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "cgrates.org:RP6:RT_WEEK", result.Rates["RT_WEEK"].UID())
	}
	result.Rates = nil
	expected.Rates = nil
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestRatesSetRateProfile(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}

	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
			ID:     "2",
		},
	}
	var result utils.RateProfile
	err = admS.GetRateProfile(context.Background(), args, &result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expected := utils.RateProfile{
		ID:        "2",
		Tenant:    "tenant",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
			},
		},
	}
	if !reflect.DeepEqual(result.Rates["RT_WEEK"].ID, expected.Rates["RT_WEEK"].ID) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.Rates["RT_WEEK"].ID, result.Rates["RT_WEEK"].ID)
	}
	if !reflect.DeepEqual(result.Rates["RT_WEEK"].ActivationTimes, expected.Rates["RT_WEEK"].ActivationTimes) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.Rates["RT_WEEK"].ActivationTimes, result.Rates["RT_WEEK"].ActivationTimes)
	}
	if !reflect.DeepEqual("tenant:2:RT_WEEK", result.Rates["RT_WEEK"].UID()) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "tenant:2:RT_WEEK", result.Rates["RT_WEEK"].UID())
	}
	result.Rates = nil
	expected.Rates = nil
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestRatesRemoveRateProfile(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
			ID:     "2",
		},
	}

	reply := utils.StringPointer("")
	err = admS.RemoveRateProfile(context.Background(), arg, reply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(reply), utils.ToJSON("OK")) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON("OK"), utils.ToJSON(reply))
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
			ID:     "2",
		},
	}
	var result utils.RateProfile
	err = admS.GetRateProfile(context.Background(), args, &result)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestRatesRemoveRateProfileMissing(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}

	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
		},
	}

	reply := utils.StringPointer("")
	err = admS.RemoveRateProfile(context.Background(), arg, reply)
	expectedErr := "MANDATORY_IE_MISSING: [ID]"
	if err == nil || err.Error() != expectedErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expectedErr, err)
	}

}

func TestRatesRemoveRateProfileEmptyTenant(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "2",
		},
	}

	reply := utils.StringPointer("")
	err = admS.RemoveRateProfile(context.Background(), arg, reply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(reply), utils.ToJSON("OK")) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON("OK"), utils.ToJSON(reply))
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "2",
		},
	}
	var result utils.RateProfile
	err = admS.GetRateProfile(context.Background(), args, &result)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestRatesSetGetRateProfileError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}

	expected := "SERVER_ERROR: NOT_IMPLEMENTED"
	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
			ID:     "2",
		},
	}
	var result utils.RateProfile
	err = admS.GetRateProfile(context.Background(), args, &result)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestRatesSetRemoveRateProfileError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}

	expected := "SERVER_ERROR: NOT_IMPLEMENTED"
	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "2",
		},
	}

	reply := utils.StringPointer("")
	err = admS.RemoveRateProfile(context.Background(), arg, reply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(reply), utils.ToJSON("")) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(""), utils.ToJSON(reply))
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
			ID:     "2",
		},
	}
	var result utils.RateProfile
	err = admS.GetRateProfile(context.Background(), args, &result)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestRatesSetRateProfileRates(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext1 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext1, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	ext2 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	expected := "OK"
	err = admS.SetRateProfile(context.Background(), ext2, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(rtRply), utils.ToJSON(expected)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rtRply))
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
			ID:     "2",
		},
	}
	var result utils.RateProfile
	err = admS.GetRateProfile(context.Background(), args, &result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expected2 := &utils.RateProfile{
		ID:        "2",
		Tenant:    "tenant",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
			},
		},
	}
	if !reflect.DeepEqual(utils.ToJSON(expected2), utils.ToJSON(&result)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected2), utils.ToJSON(&result))
	}

}

func TestRatesSetRateProfileRatesNoTenant(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext1 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext1, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	ext2 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	expected := "OK"
	err = admS.SetRateProfile(context.Background(), ext2, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(rtRply), utils.ToJSON(expected)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rtRply))
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "2",
		},
	}
	var result utils.RateProfile
	err = admS.GetRateProfile(context.Background(), args, &result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expected2 := &utils.RateProfile{
		ID:        "2",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
			},
		},
	}
	if !reflect.DeepEqual(utils.ToJSON(expected2), utils.ToJSON(&result)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected2), utils.ToJSON(&result))
	}

}

func TestRatesSetRateProfileRatesMissingField(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext2 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	expected := ""
	err := admS.SetRateProfile(context.Background(), ext2, &rtRply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
	if !reflect.DeepEqual(utils.ToJSON(rtRply), utils.ToJSON(expected)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rtRply))
	}

}

func TestRatesSetRateProfileRatesErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext2 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	expected := ""
	err := admS.SetRateProfile(context.Background(), ext2, &rtRply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
	if !reflect.DeepEqual(utils.ToJSON(rtRply), utils.ToJSON(expected)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rtRply))
	}

}

func TestRatesRemoveRateProfileRate(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext1 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext1, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	ext2 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	expected := "OK"
	err = admS.SetRateProfile(context.Background(), ext2, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(rtRply), utils.ToJSON(expected)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rtRply))
	}
	args1 := &utils.RemoveRPrfRates{
		Tenant:  "cgrates.org",
		ID:      "2",
		RateIDs: []string{"RT_WEEK"},
		APIOpts: nil,
	}
	err = admS.RemoveRateProfileRates(context.Background(), args1, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(rtRply), utils.ToJSON(expected)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rtRply))
	}

	args2 := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "2",
		},
	}
	var result utils.RateProfile
	err = admS.GetRateProfile(context.Background(), args2, &result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	expected2 := &utils.RateProfile{
		ID:        "2",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates:     map[string]*utils.Rate{},
	}
	if !reflect.DeepEqual(utils.ToJSON(expected2), utils.ToJSON(&result)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected2), utils.ToJSON(&result))
	}
}

func TestRatesRemoveRateProfileRateEmptyTenant(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext1 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext1, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expected := "OK"
	if !reflect.DeepEqual(utils.ToJSON(rtRply), utils.ToJSON(expected)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rtRply))
	}
	args1 := &utils.RemoveRPrfRates{
		ID:      "2",
		RateIDs: []string{"RT_WEEK"},
		APIOpts: nil,
	}
	err = admS.RemoveRateProfileRates(context.Background(), args1, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(rtRply), utils.ToJSON(expected)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rtRply))
	}

	args2 := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "2",
		},
	}
	var result utils.RateProfile
	err = admS.GetRateProfile(context.Background(), args2, &result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	expected2 := &utils.RateProfile{
		ID:        "2",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates:     map[string]*utils.Rate{},
	}
	if !reflect.DeepEqual(utils.ToJSON(expected2), utils.ToJSON(&result)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected2), utils.ToJSON(&result))
	}
}

func TestRatesRemoveRateProfileRateError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	var rtRply string
	args1 := &utils.RemoveRPrfRates{
		ID:      "2",
		RateIDs: []string{"RT_WEEK"},
		APIOpts: nil,
	}
	expected := "SERVER_ERROR: NOT_IMPLEMENTED"
	err := admS.RemoveRateProfileRates(context.Background(), args1, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}

func TestRatesRemoveRateProfileRateErrorMissingField(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	var rtRply string
	args1 := &utils.RemoveRPrfRates{

		RateIDs: []string{"RT_WEEK"},
		APIOpts: nil,
	}
	expected := "MANDATORY_IE_MISSING: [ID]"
	err := admS.RemoveRateProfileRates(context.Background(), args1, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesSetRateProfileErrorSetLoadIDs(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		GetRateProfileDrvF: func(c *context.Context, s string, s2 string) (*utils.RateProfile, error) {
			return nil, utils.ErrNotFound
		},
		SetRateProfileDrvF: func(c *context.Context, profile *utils.RateProfile, overWrite bool) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	expected := "SERVER_ERROR: NOT_IMPLEMENTED"
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesSetRateProfileRatesErrorSetLoadIDs(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		GetRateProfileDrvF: func(c *context.Context, s string, s2 string) (*utils.RateProfile, error) {
			return &utils.RateProfile{
				ID:        "2",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.Subject:1001"},
				Rates:     map[string]*utils.Rate{},
			}, nil
		},
		SetRateProfileDrvF: func(c *context.Context, profile *utils.RateProfile, overWrite bool) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	expected := "SERVER_ERROR: NOT_IMPLEMENTED"
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesRemoveRateProfileRatesErrorSetLoadIDs(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		RemoveRateProfileDrvF: func(ctx *context.Context, str1 string, str2 string, rtIDs *[]string) error {
			return nil
		},
		GetRateProfileDrvF: func(c *context.Context, s string, s2 string) (*utils.RateProfile, error) {
			return &utils.RateProfile{
				Tenant: "tenant",
			}, nil
		},
		SetRateProfileDrvF: func(c *context.Context, profile *utils.RateProfile, overWrite bool) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.RemoveRPrfRates{
		ID:      "2",
		Tenant:  "tenant",
		RateIDs: []string{"RT_WEEK"},
	}
	var rtRply string
	expected := "SERVER_ERROR: NOT_IMPLEMENTED"
	err := admS.RemoveRateProfileRates(context.Background(), ext, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesRemoveRateProfileErrorSetLoadIDs(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		RemoveRateProfileDrvF: func(ctx *context.Context, str1 string, str2 string, rateIDs *[]string) error {
			return nil
		},
		GetRateProfileDrvF: func(c *context.Context, s string, s2 string) (*utils.RateProfile, error) {
			return &utils.RateProfile{
				Tenant: "tenant",
			}, nil
		},
		SetRateProfileDrvF: func(c *context.Context, profile *utils.RateProfile, overWrite bool) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
			ID:     "2",
		},
	}
	var rtRply string
	expected := "SERVER_ERROR: NOT_IMPLEMENTED"
	err := admS.RemoveRateProfile(context.Background(), ext, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesSetRateProfileErrorCache(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.AdminSCfg().CachesConns = []string{}
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		GetRateProfileDrvF: func(c *context.Context, s string, s2 string) (*utils.RateProfile, error) {
			return nil, utils.ErrNotFound
		},
		SetRateProfileDrvF: func(c *context.Context, profile *utils.RateProfile, overWrite bool) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: "1",
		},
	}
	var rtRply string
	expected := "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]"
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesSetRateProfileRatesErrorCache(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.AdminSCfg().CachesConns = []string{}

	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		GetRateProfileDrvF: func(c *context.Context, s string, s2 string) (*utils.RateProfile, error) {
			return &utils.RateProfile{
				ID:        "2",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.Subject:1001"},
				Rates:     map[string]*utils.Rate{},
			}, nil
		},
		SetRateProfileDrvF: func(c *context.Context, profile *utils.RateProfile, overWrite bool) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "2",
			Tenant:    "tenant",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
		},
	}
	var rtRply string
	expected := "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]"
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesRemoveRateProfileRatesErrorCache(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.AdminSCfg().CachesConns = []string{}
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		RemoveRateProfileDrvF: func(ctx *context.Context, str1 string, str2 string, rateIDs *[]string) error {
			return nil
		},
		GetRateProfileDrvF: func(c *context.Context, s string, s2 string) (*utils.RateProfile, error) {
			return &utils.RateProfile{
				Tenant: "tenant",
			}, nil
		},
		SetRateProfileDrvF: func(c *context.Context, profile *utils.RateProfile, overWrite bool) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.RemoveRPrfRates{
		ID:      "2",
		Tenant:  "tenant",
		RateIDs: []string{"RT_WEEK"},
	}
	var rtRply string
	expected := "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]"
	err := admS.RemoveRateProfileRates(context.Background(), ext, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesRemoveRateProfileErrorSetCache(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.AdminSCfg().CachesConns = []string{}
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		RemoveRateProfileDrvF: func(ctx *context.Context, str1 string, str2 string, rateIDs *[]string) error {
			return nil
		},
		GetRateProfileDrvF: func(c *context.Context, s string, s2 string) (*utils.RateProfile, error) {
			return &utils.RateProfile{
				Tenant: "tenant",
			}, nil
		},
		SetRateProfileDrvF: func(c *context.Context, profile *utils.RateProfile, overWrite bool) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ext := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "tenant",
			ID:     "2",
		},
	}
	var rtRply string
	expected := "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]"
	err := admS.RemoveRateProfile(context.Background(), ext, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesGetRateProfilesOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args1 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Rates: map[string]*utils.Rate{
				"RATE1": {
					ID: "RATE1",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetRateProfile(context.Background(), args1, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	args2 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID2",
			Rates: map[string]*utils.Rate{
				"RATE2": {
					ID: "RATE2",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetRateProfile(context.Background(), args2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	// this profile will not match
	args3 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			Tenant: "cgrates.org",
			ID:     "test2_ID1",
			Rates: map[string]*utils.Rate{
				"RATE1": {
					ID: "RATE1",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetRateProfile(context.Background(), args3, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
	}
	exp := []*utils.RateProfile{
		{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Rates: map[string]*utils.Rate{
				"RATE1": {
					ID: "RATE1",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "test_ID2",
			Rates: map[string]*utils.Rate{
				"RATE2": {
					ID: "RATE2",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}

	var getReply []*utils.RateProfile
	if err := admS.GetRateProfiles(context.Background(), argsGet, &getReply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(getReply, func(i, j int) bool {
			return getReply[i].ID < getReply[j].ID
		})
		if utils.ToJSON(getReply) != utils.ToJSON(exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(exp), utils.ToJSON(getReply))
		}
	}
}

func TestRatesGetRateProfilesGetIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Rates: map[string]*utils.Rate{
				"RATE1": {
					ID: "RATE1",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetRateProfile(context.Background(), args, &setReply); err != nil {
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
	var getReply []*utils.RateProfile
	if err := admS.GetRateProfiles(context.Background(), argsGet, &getReply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestRatesGetRateProfilesGetProfileErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		SetRateProfileDrvF: func(*context.Context, *utils.RateProfile, bool) error {
			return nil
		},
		RemoveRateProfileDrvF: func(*context.Context, string, string, *[]string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"rtp_cgrates.org:TEST"}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []*utils.RateProfile
	experr := "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.GetRateProfiles(context.Background(),
		&utils.ArgsItemIDs{
			ItemsPrefix: "TEST",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesGetRateProfileIDsGetOptsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetRateProfileDrvF: func(*context.Context, string, string) (*utils.RateProfile, error) {
			ratePrf := &utils.RateProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return ratePrf, nil
		},
		SetRateProfileDrvF: func(*context.Context, *utils.RateProfile, bool) error {
			return nil
		},
		RemoveRateProfileDrvF: func(*context.Context, string, string, *[]string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"rtp_cgrates.org:key1", "rtp_cgrates.org:key2", "rtp_cgrates.org:key3"}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "cannot convert field<bool>: true to int"

	if err := adms.GetRateProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
			APIOpts: map[string]any{
				utils.PageLimitOpt: true,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesGetRateProfileIDsPaginateErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetRateProfileDrvF: func(*context.Context, string, string) (*utils.RateProfile, error) {
			ratePrf := &utils.RateProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return ratePrf, nil
		},
		SetRateProfileDrvF: func(*context.Context, *utils.RateProfile, bool) error {
			return nil
		},
		RemoveRateProfileDrvF: func(*context.Context, string, string, *[]string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"rtp_cgrates.org:key1", "rtp_cgrates.org:key2", "rtp_cgrates.org:key3"}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := `SERVER_ERROR: maximum number of items exceeded`

	if err := adms.GetRateProfileIDs(context.Background(),
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

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesSetGetRemRateProfileRates(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	arg := &utils.ArgsSubItemIDs{
		Tenant:      "cgrates.org",
		ProfileID:   "test_ID1",
		ItemsPrefix: "RATE",
	}
	var result []*utils.Rate
	var reply string

	ratePrf := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Rates: map[string]*utils.Rate{
				"RATE1": {
					ID: "RATE1",
				},
				"RATE2": {
					ID: "RATE2",
				},
				"RATE3": {
					ID: "RATE3",
				},
				"INVALID": {
					ID: "INVALID",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	if err := adms.SetRateProfile(context.Background(), ratePrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	exp := []*utils.Rate{
		{
			ID: "RATE1",
		},
		{
			ID: "RATE2",
		},
		{
			ID: "RATE3",
		},
	}
	if err := adms.GetRateProfileRates(context.Background(), arg, &result); err != nil {
		t.Error(err)
	} else {
		sort.Slice(result, func(i, j int) bool {
			return result[i].ID < result[j].ID
		})
		if utils.ToJSON(result) != utils.ToJSON(exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(exp), utils.ToJSON(result))
		}
	}

	var rateIDs []string
	expRateIDs := []string{"RATE1", "RATE2", "RATE3"}

	if err := adms.GetRateProfileRateIDs(context.Background(), &utils.ArgsSubItemIDs{
		Tenant:      "cgrates.org",
		ProfileID:   "test_ID1",
		ItemsPrefix: "RATE",
	},
		&rateIDs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(rateIDs, func(i, j int) bool {
			return rateIDs[i] < rateIDs[j]
		})
		if !reflect.DeepEqual(rateIDs, expRateIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expRateIDs, rateIDs)
		}
	}

	var rplyCount int

	if err := adms.GetRateProfileRatesCount(context.Background(), &utils.ArgsSubItemIDs{
		Tenant:      "cgrates.org",
		ProfileID:   "test_ID1",
		ItemsPrefix: "RATE",
	},
		&rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != len(rateIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", len(rateIDs), rplyCount)
	}

	argsRemove := &utils.RemoveRPrfRates{
		Tenant:  "cgrates.org",
		ID:      "test_ID1",
		RateIDs: []string{"RATE1", "RATE2", "RATE3"},
	}

	if err := adms.RemoveRateProfileRates(context.Background(), argsRemove, &reply); err != nil {
		t.Error(err)
	}

	engine.Cache.Clear(nil)
	if err := adms.GetRateProfileRates(context.Background(), arg, &result); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesGetRateProfileRatesCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv []*utils.Rate
	experr := "MANDATORY_IE_MISSING: [ProfileID]"

	if err := adms.GetRateProfileRates(context.Background(), &utils.ArgsSubItemIDs{
		Tenant: "cgrates.org",
	}, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "NO_DATABASE_CONNECTION"

	arg := &utils.ArgsSubItemIDs{
		ProfileID: "RatePrf1",
	}

	if err := adms.GetRateProfileRates(context.Background(), arg, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesGetRateProfileRatesCountErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetRateProfileDrvF: func(*context.Context, string, string) (*utils.RateProfile, error) {
			ratePrf := &utils.RateProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return ratePrf, nil
		},
		SetRateProfileDrvF: func(*context.Context, *utils.RateProfile, bool) error {
			return nil
		},
		RemoveRateProfileDrvF: func(*context.Context, string, string, *[]string) error {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetRateProfileRatesCount(context.Background(),
		&utils.ArgsSubItemIDs{
			Tenant:    "cgrates.org",
			ProfileID: "prfID",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestRatesGetRateProfileRatesCountErrKeys(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetRateProfileRatesDrvF: func(*context.Context, string, string, string, bool) ([]string, []*utils.Rate, error) {
			return []string{}, nil, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetRateProfileRatesCount(context.Background(),
		&utils.ArgsSubItemIDs{
			ProfileID: "prfID",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestRatesGetRateProfileRatesCountErrMissing(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, nil)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	experr := `MANDATORY_IE_MISSING: [ProfileID]`
	var reply int
	if err := adms.GetRateProfileRatesCount(context.Background(),
		&utils.ArgsSubItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestRatesGetRateProfileRateIDsErrNotFound(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	idb, err := engine.NewInternalDB(nil, nil, nil, nil)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string

	if err := adms.GetRateProfileRateIDs(context.Background(),
		&utils.ArgsSubItemIDs{
			Tenant:    "cgrates.org",
			ProfileID: "prfID",
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesGetRateProfileRateIDsErrKeys(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetRateProfileRatesDrvF: func(ctx *context.Context, s1, s2, s3 string, b bool) ([]string, []*utils.Rate, error) {
			return []string{}, nil, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string

	if err := adms.GetRateProfileRateIDs(context.Background(),
		&utils.ArgsSubItemIDs{
			Tenant:    "cgrates.org",
			ProfileID: "prfID",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesGetRateProfileRateIDsGetOptsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetRateProfileRatesDrvF: func(ctx *context.Context, s1, s2, s3 string, b bool) ([]string, []*utils.Rate, error) {
			return []string{"RATE1", "RATE2"}, []*utils.Rate{
				{
					ID: "RATE1",
				},
				{
					ID: "RATE2",
				},
			}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "cannot convert field<bool>: true to int"

	if err := adms.GetRateProfileRateIDs(context.Background(),
		&utils.ArgsSubItemIDs{
			Tenant:    "cgrates.org",
			ProfileID: "prfID",
			APIOpts: map[string]any{
				utils.PageLimitOpt: true,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesGetRateProfileRateIDsPaginateErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetRateProfileRatesDrvF: func(ctx *context.Context, s1, s2, s3 string, b bool) ([]string, []*utils.Rate, error) {
			return []string{"RATE1", "RATE2"}, []*utils.Rate{
				{
					ID: "RATE1",
				},
				{
					ID: "RATE2",
				},
			}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := `SERVER_ERROR: maximum number of items exceeded`

	if err := adms.GetRateProfileRateIDs(context.Background(),
		&utils.ArgsSubItemIDs{
			ProfileID: "prfID",
			APIOpts: map[string]any{
				utils.PageLimitOpt:    2,
				utils.PageOffsetOpt:   4,
				utils.PageMaxItemsOpt: 5,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesGetRateProfileRateIDsErrMissing(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, nil)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := `MANDATORY_IE_MISSING: [ProfileID]`

	if err := adms.GetRateProfileRateIDs(context.Background(),
		&utils.ArgsSubItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestRatesSetRateProfileErrConvertOverwriteOpt(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					ActivationTimes: "* * * * *",
				},
			},
			ID: "RateProfile",
		},
		APIOpts: map[string]any{
			utils.MetaRateSOverwrite: "invalid_opt",
		},
	}
	expected := `strconv.ParseBool: parsing "invalid_opt": invalid syntax`
	var rtRply string
	err := admS.SetRateProfile(context.Background(), args, &rtRply)
	if err == nil || err.Error() != expected {
		t.Errorf("expected <%+v>, \nreceived <%+v>", expected, err)
	}
}

func TestRatesGetRateProfilePagination(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	ratePrf := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "RATE_PROFILE",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Rates: map[string]*utils.Rate{
				"RateA1": {
					ID: "RateA1",
					Weights: utils.DynamicWeights{
						{
							Weight: 35,
						},
					},
				},
				"RateA2": {
					ID: "RateA2",
					Weights: utils.DynamicWeights{
						{
							Weight: 5,
						},
					},
				},
				"RateA3": {
					ID: "RateA3",
					Weights: utils.DynamicWeights{
						{
							Weight: 40,
						},
					},
				},
				"RateB5": {
					ID: "RateB5",
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
				},
				"RateB1": {
					ID: "RateB1",
					Weights: utils.DynamicWeights{
						{
							Weight: 25,
						},
					},
				},
				"RateB3": {
					ID: "RateB3",
					Weights: utils.DynamicWeights{
						{
							Weight: 15,
						},
					},
				},
				"RateB2": {
					ID: "RateB2",
					Weights: utils.DynamicWeights{
						{
							Weight: 5,
						},
					},
				},
				"RateB6": {
					ID: "RateB6",
					Weights: utils.DynamicWeights{
						{
							Weight: 30,
						},
					},
				},
				"RateB4": {
					ID: "RateB4",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
				},
			},
		},
	}

	var reply string
	if err := admS.SetRateProfile(context.Background(), ratePrf, &reply); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RATE_PROFILE",
		},
		APIOpts: map[string]any{
			utils.PageLimitOpt:   4,
			utils.PageOffsetOpt:  1,
			utils.ItemsPrefixOpt: "RateB",
		},
	}

	expected := utils.RateProfile{
		ID:        "RATE_PROFILE",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RateB2": {
				ID: "RateB2",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
			},
			"RateB3": {
				ID: "RateB3",
				Weights: utils.DynamicWeights{
					{
						Weight: 15,
					},
				},
			},
			"RateB4": {
				ID: "RateB4",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			},
			"RateB5": {
				ID: "RateB5",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
			},
		},
	}
	var replyRatePrf utils.RateProfile
	if err := admS.GetRateProfile(context.Background(), args, &replyRatePrf); err != nil {
		t.Error(err)
	} else if len(replyRatePrf.Rates) != len(expected.Rates) {
		t.Errorf("expected: %+v Rates, \nreceived: %+v Rates",
			len(expected.Rates), len(replyRatePrf.Rates))
	} else {
		for rateID := range expected.Rates {
			if _, ok := replyRatePrf.Rates[rateID]; !ok {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>",
					utils.ToJSON(expected), utils.ToJSON(replyRatePrf))
				t.Fatalf("rate <%+v> could not be found in reply", rateID)
			}
		}
	}

	engine.Cache.Clear(nil)
	args = &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RATE_PROFILE",
		},
		APIOpts: map[string]any{
			utils.PageLimitOpt:   4,
			utils.PageOffsetOpt:  1,
			utils.ItemsPrefixOpt: "RateA",
		},
	}

	expected = utils.RateProfile{
		ID:        "RATE_PROFILE",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RateA2": {
				ID: "RateA2",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
			},
			"RateA3": {
				ID: "RateA3",
				Weights: utils.DynamicWeights{
					{
						Weight: 40,
					},
				},
			},
		},
	}
	if err := admS.GetRateProfile(context.Background(), args, &replyRatePrf); err != nil {
		t.Error(err)
	} else if len(replyRatePrf.Rates) != len(expected.Rates) {
		t.Errorf("expected: %+v Rates, \nreceived: %+v Rates",
			len(expected.Rates), len(replyRatePrf.Rates))
	} else {
		for rateID := range expected.Rates {
			if _, ok := replyRatePrf.Rates[rateID]; !ok {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>",
					utils.ToJSON(expected), utils.ToJSON(replyRatePrf))
				t.Fatalf("rate <%+v> could not be found in reply", rateID)
			}
		}
	}

	args = &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RATE_PROFILE",
		},
		APIOpts: map[string]any{
			utils.PageOffsetOpt: 1,
		},
	}

	expected = utils.RateProfile{
		ID:        "RATE_PROFILE",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RateA2": {
				ID: "RateA2",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
			},
			"RateA3": {
				ID: "RateA3",
				Weights: utils.DynamicWeights{
					{
						Weight: 40,
					},
				},
			},
			"RateB5": {
				ID: "RateB5",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
			},
			"RateB1": {
				ID: "RateB1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
			},
			"RateB3": {
				ID: "RateB3",
				Weights: utils.DynamicWeights{
					{
						Weight: 15,
					},
				},
			},
			"RateB2": {
				ID: "RateB2",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
			},
			"RateB6": {
				ID: "RateB6",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
			},
			"RateB4": {
				ID: "RateB4",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			},
		},
	}
	if err := admS.GetRateProfile(context.Background(), args, &replyRatePrf); err != nil {
		t.Error(err)
	} else if len(replyRatePrf.Rates) != len(expected.Rates) {
		t.Errorf("expected: %+v Rates, \nreceived: %+v Rates",
			len(expected.Rates), len(replyRatePrf.Rates))
	} else {
		for rateID := range expected.Rates {
			if _, ok := replyRatePrf.Rates[rateID]; !ok {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>",
					utils.ToJSON(expected), utils.ToJSON(replyRatePrf))
				t.Fatalf("rate <%+v> could not be found in reply", rateID)
			}
		}
	}

	args = &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RATE_PROFILE",
		},
		APIOpts: map[string]any{
			utils.PageLimitOpt:    4,
			utils.PageOffsetOpt:   1,
			utils.PageMaxItemsOpt: 4,
			utils.ItemsPrefixOpt:  "RateB",
		},
	}

	experr := `SERVER_ERROR: maximum number of items exceeded`
	if err := admS.GetRateProfile(context.Background(), args, &replyRatePrf); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RATE_PROFILE",
		},
		APIOpts: map[string]any{
			utils.PageLimitOpt:  true,
			utils.PageOffsetOpt: 1,
		},
	}

	experr = `cannot convert field<bool>: true to int`
	if err := admS.GetRateProfile(context.Background(), args, &replyRatePrf); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}
