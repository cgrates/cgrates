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

func TestAttributesSetGetAttributeProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			ID: "TestGetAttributeProfile",
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.RequestType",
					Type:  utils.MetaConstant,
					Value: utils.MetaPrepaid,
				},
			},
		},
	}
	var reply string

	if err := admS.SetAttributeProfile(context.Background(), attrPrf, &reply); err != nil {
		t.Error(err)
	}
	//get after set
	var rcv engine.APIAttributeProfile
	if err := admS.GetAttributeProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestGetAttributeProfile",
			}}, &rcv); err != nil {
		t.Error(err)
	} else {
		newRcv := &rcv
		if !reflect.DeepEqual(newRcv, attrPrf.APIAttributeProfile) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(attrPrf.APIAttributeProfile), utils.ToJSON(newRcv))
		}
	}

	//count the IDs
	var nmbr int
	if err := admS.GetAttributeProfilesCount(context.Background(), &utils.ArgsItemIDs{}, &nmbr); err != nil {
		t.Error(err)
	} else if nmbr != 1 {
		t.Errorf("Expected just one ID")
	}

	//get the IDs
	var ids []string
	expected := []string{"TestGetAttributeProfile"}
	if err := admS.GetAttributeProfileIDs(context.Background(),
		&utils.ArgsItemIDs{}, &ids); err != nil {
		t.Error(err)
	} else if len(ids) != len(expected) {
		t.Errorf("Expected %+v, received %+v", ids, expected)
	}

	if err := admS.RemoveAttributeProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestGetAttributeProfile",
			}}, &reply); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	dm.DataDB().Flush(utils.EmptyString)
}

func TestAttributesSetAttributeProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.RequestType",
					Type:  utils.MetaConstant,
					Value: utils.MetaConstant,
				},
			},
		},
	}

	var reply string
	expected := "MANDATORY_IE_MISSING: [ID]"
	if err := admS.SetAttributeProfile(context.Background(), attrPrf, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	attrPrf.ID = "TestSetAttributeProfileCheckErrors"

	attrPrf.FilterIDs = []string{"invalid_fltier_format"}
	expected = "SERVER_ERROR: broken reference to filter: <invalid_fltier_format> for item with ID: cgrates.org:TestSetAttributeProfileCheckErrors"
	if err := admS.SetAttributeProfile(context.Background(), attrPrf, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	attrPrf.FilterIDs = []string{}

	attrPrf.Attributes[0].Path = "*req.RequestType"

	admS.connMgr = engine.NewConnManager(cfg)
	admS.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	expected = "SERVER_ERROR: context deadline exceeded"
	cfg.GeneralCfg().DefaultCaching = utils.MetaRemove
	if err := admS.SetAttributeProfile(ctx, attrPrf, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	cancel()

	dbMock := &engine.DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.AttributeProfile, error) {
			attrPRf := &engine.AttributeProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return attrPRf, nil
		},
		SetAttributeProfileDrvF: func(ctx *context.Context, attr *engine.AttributeProfile) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, utils.ErrNotImplemented
		},
		RemoveAttributeProfileDrvF: func(ctx *context.Context, str1 string, str2 string) error {
			return nil
		},
	}

	admS.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	expected = "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := admS.SetAttributeProfile(context.Background(), attrPrf, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	engine.Cache.Clear(nil)
	dm.DataDB().Flush(utils.EmptyString)
}

func TestAttributesGetAttributeProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var rcv engine.APIAttributeProfile
	expected := "MANDATORY_IE_MISSING: [ID]"
	if err := admS.GetAttributeProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{}}, &rcv); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	admS.dm = nil
	expected = "SERVER_ERROR: NO_DATABASE_CONNECTION"
	if err := admS.GetAttributeProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestGetAttributeProfileCheckErrors",
			}}, &rcv); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	engine.Cache.Clear(nil)
	dm.DataDB().Flush(utils.EmptyString)
}

func TestAttributesRemoveAttributeProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			ID: "TestGetAttributeProfile",
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.RequestType",
					Type:  utils.MetaConstant,
					Value: utils.MetaConstant,
				},
			},
		},
	}
	var reply string
	if err := admS.SetAttributeProfile(context.Background(), attrPrf, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	admS.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	admS.connMgr = engine.NewConnManager(cfg)
	admS.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
	expected := "SERVER_ERROR: context deadline exceeded"
	if err := admS.RemoveAttributeProfile(ctx,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestGetAttributeProfile",
			},
			APIOpts: map[string]interface{}{
				utils.MetaUsage: 10,
			}}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	cancel()
	admS.cfg.GeneralCfg().DefaultCaching = utils.MetaNone

	var rcv engine.APIAttributeProfile
	if err := admS.GetAttributeProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestGetAttributeProfile",
			}}, &rcv); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	expected = "MANDATORY_IE_MISSING: [ID]"
	if err := admS.RemoveAttributeProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{}}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	admS.dm = nil
	expected = "SERVER_ERROR: NO_DATABASE_CONNECTION"
	if err := admS.RemoveAttributeProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestGetAttributeProfile",
			}}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	admS.dm = engine.NewDataManager(data, cfg.CacheCfg(), nil)

	engine.Cache.Clear(nil)
	dm.DataDB().Flush(utils.EmptyString)
}

func TestAttributesRemoveAttributeProfileMockErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.AttributeProfile, error) {
			attrPRf := &engine.AttributeProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return attrPRf, nil
		},
		SetAttributeProfileDrvF: func(ctx *context.Context, attr *engine.AttributeProfile) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, utils.ErrNotImplemented
		},
		RemoveAttributeProfileDrvF: func(ctx *context.Context, str1 string, str2 string) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{
				"testIndex": {},
			}, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply string
	expected := "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := admS.RemoveAttributeProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestGetAttributeProfile",
			}}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	engine.Cache.Clear(nil)
	dm.DataDB().Flush(utils.EmptyString)
}

func TestAttributesGetAttributeProfileIDsMockErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.AttributeProfile, error) {
			attrPRf := &engine.AttributeProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return attrPRf, nil
		},
		SetAttributeProfileDrvF: func(ctx *context.Context, attr *engine.AttributeProfile) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, utils.ErrNotImplemented
		},
		RemoveAttributeProfileDrvF: func(ctx *context.Context, str1 string, str2 string) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	expected := "NOT_IMPLEMENTED"
	if err := admS.GetAttributeProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	engine.Cache.Clear(nil)
	dm.DataDB().Flush(utils.EmptyString)
}

func TestAttributesGetAttributeProfileIDsMockErrKeys(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{}, nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	expected := "NOT_FOUND"
	if err := admS.GetAttributeProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	engine.Cache.Clear(nil)
	dm.DataDB().Flush(utils.EmptyString)
}

func TestAttributesGetAttributeProfilesCountMockErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.AttributeProfile, error) {
			attrPRf := &engine.AttributeProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return attrPRf, nil
		},
		SetAttributeProfileDrvF: func(ctx *context.Context, attr *engine.AttributeProfile) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, utils.ErrNotImplemented
		},
		RemoveAttributeProfileDrvF: func(ctx *context.Context, str1 string, str2 string) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int
	expected := "NOT_IMPLEMENTED"
	if err := admS.GetAttributeProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	dbMockNew := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{}, nil
		},
	}
	expected = "NOT_FOUND"
	admS.dm = engine.NewDataManager(dbMockNew, cfg.CacheCfg(), nil)
	if err := admS.GetAttributeProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	engine.Cache.Clear(nil)
	dm.DataDB().Flush(utils.EmptyString)
}

func TestAttributesGetAttributeForEvent(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	attrS := engine.NewAttributeService(dm, fltrs, cfg)
	admS := &AdminSv1{
		dm:  dm,
		cfg: cfg,
	}
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "TestGetAttributeProfile",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.RequestType",
					Type:  utils.MetaConstant,
					Value: utils.MetaPrepaid,
				},
			},
		},
	}
	var result string

	if err := admS.SetAttributeProfile(context.Background(), attrPrf, &result); err != nil {
		t.Error(err)
	}

	attsv1 := NewAttributeSv1(attrS)
	ev := &utils.CGREvent{
		Event: map[string]interface{}{
			utils.AccountField: "1002",
		},
	}
	var reply engine.APIAttributeProfile

	expAttrPrf := &engine.APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "TestGetAttributeProfile",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: utils.MetaPrepaid,
			},
		},
	}

	if err := attsv1.GetAttributeForEvent(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else {
		rplyPntr := &reply
		if !reflect.DeepEqual(expAttrPrf, rplyPntr) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expAttrPrf), utils.ToJSON(rplyPntr))
		}
	}

	var rplyev engine.AttrSProcessEventReply
	ev.Event[utils.RequestType] = utils.MetaPseudoPrepaid
	//now we will process the event for our attr
	expectedEv := engine.AttrSProcessEventReply{
		AlteredFields: []*engine.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:TestGetAttributeProfile",
				Fields:           []string{"*req.RequestType"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
				utils.RequestType:  utils.MetaPrepaid,
			},
			APIOpts: map[string]interface{}{},
		},
	}
	if err := attsv1.ProcessEvent(context.Background(), ev, &rplyev); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedEv, rplyev) {
		t.Errorf("Expected %+v ,received %+v", utils.ToJSON(expectedEv), utils.ToJSON(rplyev))
	}
}

func TestAttributesSv1Ping(t *testing.T) {
	attrsv := new(AttributeSv1)
	var reply string
	if err := attrsv.Ping(nil, nil, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply error")
	}
}

func TestAttributesGetAttributeProfilesOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	args1 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.RequestType",
					Type:  utils.MetaConstant,
					Value: utils.MetaPrepaid,
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
	if err := admS.SetAttributeProfile(context.Background(), args1, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	args2 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID2",
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.RequestType",
					Type:  utils.MetaConstant,
					Value: utils.MetaPrepaid,
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

	if err := admS.SetAttributeProfile(context.Background(), args2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	// this profile will not match
	args3 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant: "cgrates.org",
			ID:     "test2_ID1",
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.RequestType",
					Type:  utils.MetaConstant,
					Value: utils.MetaPrepaid,
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

	if err := admS.SetAttributeProfile(context.Background(), args3, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
	}
	exp := []*engine.APIAttributeProfile{
		{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.RequestType",
					Type:  utils.MetaConstant,
					Value: utils.MetaPrepaid,
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
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.RequestType",
					Type:  utils.MetaConstant,
					Value: utils.MetaPrepaid,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}

	var getReply []*engine.APIAttributeProfile
	if err := admS.GetAttributeProfiles(context.Background(), argsGet, &getReply); err != nil {
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

func TestAttributesGetAttributeProfilesGetIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	args := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.RequestType",
					Type:  utils.MetaConstant,
					Value: utils.MetaPrepaid,
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
	if err := admS.SetAttributeProfile(context.Background(), args, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
		APIOpts: map[string]interface{}{
			utils.PageLimitOpt:    2,
			utils.PageOffsetOpt:   4,
			utils.PageMaxItemsOpt: 5,
		},
	}

	experr := `SERVER_ERROR: maximum number of items exceeded`
	var getReply []*engine.APIAttributeProfile
	if err := admS.GetAttributeProfiles(context.Background(), argsGet, &getReply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesGetAttributeProfilesGetProfileErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		SetAttributeProfileDrvF: func(*context.Context, *engine.AttributeProfile) error {
			return nil
		},
		RemoveAttributeProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"alp_cgrates.org:TEST"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []*engine.APIAttributeProfile
	experr := "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.GetAttributeProfiles(context.Background(),
		&utils.ArgsItemIDs{
			ItemsPrefix: "TEST",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestAttributesGetAttributeProfileIDsGetOptsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetAttributeProfileDrvF: func(*context.Context, string, string) (*engine.AttributeProfile, error) {
			attrPrf := &engine.AttributeProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return attrPrf, nil
		},
		SetAttributeProfileDrvF: func(*context.Context, *engine.AttributeProfile) error {
			return nil
		},
		RemoveAttributeProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"alp_cgrates.org:key1", "alp_cgrates.org:key2", "alp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "cannot convert field<bool>: true to int"

	if err := adms.GetAttributeProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
			APIOpts: map[string]interface{}{
				utils.PageLimitOpt: true,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestAttributesGetAttributeProfileIDsPaginateErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetAttributeProfileDrvF: func(*context.Context, string, string) (*engine.AttributeProfile, error) {
			attrPrf := &engine.AttributeProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return attrPrf, nil
		},
		SetAttributeProfileDrvF: func(*context.Context, *engine.AttributeProfile) error {
			return nil
		},
		RemoveAttributeProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"alp_cgrates.org:key1", "alp_cgrates.org:key2", "alp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := `SERVER_ERROR: maximum number of items exceeded`

	if err := adms.GetAttributeProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
			APIOpts: map[string]interface{}{
				utils.PageLimitOpt:    2,
				utils.PageOffsetOpt:   4,
				utils.PageMaxItemsOpt: 5,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestAttributesSetAttributeProfileAsAttrPrfErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			ID:         "TestGetAttributeProfile",
			Attributes: []*engine.ExternalAttribute{},
		},
	}
	var reply string

	experr := `SERVER_ERROR: MANDATORY_IE_MISSING: [Attributes]`
	if err := admS.SetAttributeProfile(context.Background(), attrPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}
