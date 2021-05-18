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

	"github.com/cgrates/cgrates/rates"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

func TestRatesGetRateProfileErrMandatoryIeMissing(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
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

func TestRatesGetRateProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &utils.APIRateProfile{
		ID:        "DefaultRate",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
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

func TestApisRatesGetRateProfileErrorNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
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

func TestApisRatesGetRateProfileIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &utils.APIRateProfile{
		ID:        "DefaultRate",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
			},
		},
	}

	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.PaginatorWithTenant{
		Tenant: utils.EmptyString,
	}
	result := &[]string{}
	expected := &[]string{"DefaultRate"}
	err = admS.GetRateProfileIDs(context.Background(), args, result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestApisRatesGetRateProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &utils.APIRateProfile{
		ID:        "DefaultRate",
		Tenant:    "tenant",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
			},
		},
	}

	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.PaginatorWithTenant{
		Tenant: "tenant",
	}
	result := &[]string{}
	expected := &[]string{"DefaultRate"}
	err = admS.GetRateProfileIDs(context.Background(), args, result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestApisRatesGetRateProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDBMock := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDBMock, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	args := &utils.PaginatorWithTenant{
		Tenant: "tenant",
	}
	result := &[]string{}
	err := admS.GetRateProfileIDs(context.Background(), args, result)
	if err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestApisRatesGetRateProfileErr2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDBMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(*context.Context, string) ([]string, error) {
			return []string{}, nil
		},
	}
	dm := engine.NewDataManager(dataDBMock, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	args := &utils.PaginatorWithTenant{
		Tenant: "tenant",
	}
	result := &[]string{}
	err := admS.GetRateProfileIDs(context.Background(), args, result)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestApisRatesGetRateProfileIDsCount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &utils.APIRateProfile{
		ID:        "DefaultRate",
		Tenant:    "tenant",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
			},
		},
	}

	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.TenantWithAPIOpts{
		Tenant: "tenant",
	}
	result := utils.IntPointer(0)
	expected := utils.IntPointer(1)
	err = admS.GetRateProfileIDsCount(context.Background(), args, result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(&result, &expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", &expected, &result)
	}
}

func TestApisRatesGetRateProfileIDsCountEmptyTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &utils.APIRateProfile{
		ID:        "DefaultRate",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
			},
		},
	}

	var rtRply string
	err := admS.SetRateProfile(context.Background(), ext, &rtRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	args := &utils.TenantWithAPIOpts{
		Tenant: "",
	}
	result := utils.IntPointer(0)
	expected := utils.IntPointer(1)
	err = admS.GetRateProfileIDsCount(context.Background(), args, result)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(&result, &expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", &expected, &result)
	}
}

func TestApisRatesGetRateProfileIDsCountGetKeysError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDBMock := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDBMock, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	args := &utils.TenantWithAPIOpts{
		Tenant: "",
	}
	result := utils.IntPointer(0)
	err := admS.GetRateProfileIDsCount(context.Background(), args, result)
	if err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestApisRatesGetRateProfileIDsCountKeysLenError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDBMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(*context.Context, string) ([]string, error) {
			return []string{}, nil
		},
	}
	dm := engine.NewDataManager(dataDBMock, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	args := &utils.TenantWithAPIOpts{
		Tenant: "",
	}
	result := utils.IntPointer(0)
	err := admS.GetRateProfileIDsCount(context.Background(), args, result)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestApisRateSetRateProfileMissingStructFieldError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &utils.APIRateProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
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

func TestApisRateSetRateProfileEmptyTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &utils.APIRateProfile{
		ID:        "2",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
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

func TestApisRateSetRateProfileError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &utils.APIRateProfile{
		ID:        "2",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
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

func TestApisRateSetRateProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &utils.APIRateProfile{
		ID:        "2",
		Tenant:    "tenant",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
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

func TestApisRateNewRateSv1(t *testing.T) {
	rateS := &rates.RateS{}
	expected := &RateSv1{
		rS: rateS,
	}
	result := NewRateSv1(rateS)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

/*
func TestApisRateCostForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	rateS := rates.NewRateS(cfg, nil, dm)
	expected := &RateSv1{
		rS: rateS,
	}
	rateSv1 := NewRateSv1(rateS)
	if !reflect.DeepEqual(rateSv1, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, rateSv1)
	}
	args := &utils.ArgsCostForEvent{
		RateProfileIDs: []string{"rtID"},
		CGREvent: &utils.CGREvent{
			Tenant:  "tenant",
			ID:      "ID",
			Time:    nil,
			Event:   nil,
			APIOpts: nil,
		},
	}
	rpCost := &utils.RateProfileCost{}
	err := rateSv1.CostForEvent(context.Background(), args, rpCost)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
	expected2 := &utils.RateProfileCost{}
	if !reflect.DeepEqual(rpCost, expected2) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected2, rpCost)
	}
}
*/
