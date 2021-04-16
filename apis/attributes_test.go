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

func TestSetGetAttributeProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	attrPrf := &engine.AttributeWithAPIOpts{
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
	if err := admS.GetAttributeProfileIDsCount(context.Background(), &utils.TenantWithAPIOpts{}, &nmbr); err != nil {
		t.Error(err)
	} else if nmbr != 1 {
		t.Errorf("Expected just one ID")
	}

	//get the IDs
	var ids []string
	expected := []string{"TestGetAttributeProfile"}
	if err := admS.GetAttributeProfileIDs(context.Background(),
		&utils.PaginatorWithTenant{}, &ids); err != nil {
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
	dm.DataDB().Flush(utils.EmptyString)
}

type dataDBKeys struct {
	*engine.DataDBMock
}

func (dbM *dataDBKeys) GetAttributeProfileDrv(*context.Context, string, string) (*engine.AttributeProfile, error) {
	attrPRf := &engine.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "TEST",
	}
	return attrPRf, nil
}

func (dbM *dataDBKeys) SetAttributeProfileDrv(*context.Context, *engine.AttributeProfile) error {
	return nil
}

func (dbM *dataDBKeys) RemoveAttributeProfileDrv(*context.Context, string, string) error {
	return nil
}

func TestSetAttributeProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	attrPrf := &engine.AttributeWithAPIOpts{
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
	expected = "SERVER_ERROR: broken reference to filter: invalid_fltier_format for item with ID: cgrates.org:TestSetAttributeProfileCheckErrors"
	if err := admS.SetAttributeProfile(context.Background(), attrPrf, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	attrPrf.FilterIDs = []string{}

	attrPrf.Attributes[0].Path = utils.EmptyString
	expected = "SERVER_ERROR: MANDATORY_IE_MISSING: [Path]"
	if err := admS.SetAttributeProfile(context.Background(), attrPrf, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	attrPrf.Attributes[0].Path = "*req.RequestType"

	admS.connMgr = engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): make(chan birpc.ClientConnector),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	expected = "SERVER_ERROR: context deadline exceeded"
	cfg.GeneralCfg().DefaultCaching = utils.MetaRemove
	if err := admS.SetAttributeProfile(ctx, attrPrf, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	cancel()

	dbMock := new(dataDBKeys)
	admS.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	expected = "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := admS.SetAttributeProfile(context.Background(), attrPrf, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestGetAttributeProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, true)
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

	dm.DataDB().Flush(utils.EmptyString)
}

func TestRemoveAttributeProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	attrPrf := &engine.AttributeWithAPIOpts{
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
	admS.connMgr = engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): make(chan birpc.ClientConnector),
	})
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

	dm.DataDB().Flush(utils.EmptyString)
}

func TestRemoveAttributeProfileMockErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := new(dataDBKeys)
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

	dm.DataDB().Flush(utils.EmptyString)
}

func TestGetAttributeProfileIDsMockErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := new(dataDBKeys)
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	expected := "NOT_IMPLEMENTED"
	if err := admS.GetAttributeProfileIDs(context.Background(),
		&utils.PaginatorWithTenant{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	dm.DataDB().Flush(utils.EmptyString)
}

type dataDBKeysFill struct {
	*engine.DataDBMock
}

func (dbM *dataDBKeysFill) GetKeysForPrefix(_ *context.Context, _ string) ([]string, error) {
	return []string{}, nil
}

func TestGetAttributeProfileIDsMockErrKeys(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := new(dataDBKeysFill)
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	expected := "NOT_FOUND"
	if err := admS.GetAttributeProfileIDs(context.Background(),
		&utils.PaginatorWithTenant{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	dm.DataDB().Flush(utils.EmptyString)
}

func TestGetAttributeProfileIDscOUNTMockErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := new(dataDBKeys)
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	admS := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int
	expected := "NOT_IMPLEMENTED"
	if err := admS.GetAttributeProfileIDsCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	dbMockNew := new(dataDBKeysFill)
	expected = "NOT_FOUND"
	admS.dm = engine.NewDataManager(dbMockNew, cfg.CacheCfg(), nil)
	if err := admS.GetAttributeProfileIDsCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	dm.DataDB().Flush(utils.EmptyString)
}

func TestGetAttributeForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	attrS := engine.NewAttributeService(dm, fltrs, cfg)
	admS := &AdminSv1{
		dm:  dm,
		cfg: cfg,
	}
	attrPrf := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    "cgrates.org",
			Contexts:  []string{"*any"},
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
	args := &engine.AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.AccountField: "1002",
			},
		},
	}
	var reply engine.AttributeProfile

	rsr, err := config.NewRSRParsers(utils.MetaPrepaid, utils.InInFieldSep)
	if err != nil {
		t.Error(err)
	}
	expAttrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		Contexts:  []string{"*any"},
		ID:        "TestGetAttributeProfile",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Attributes: []*engine.Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: rsr,
			},
		},
	}

	if err := attsv1.GetAttributeForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else {
		rplyPntr := &reply
		if !reflect.DeepEqual(expAttrPrf, rplyPntr) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expAttrPrf), utils.ToJSON(rplyPntr))
		}
	}

	var rplyev engine.AttrSProcessEventReply
	args.Event[utils.RequestType] = utils.MetaPseudoPrepaid
	//now we will process the event for our attr
	expectedEv := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"TestGetAttributeProfile"},
		AlteredFields:   []string{"*req.RequestType"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
				utils.RequestType:  utils.MetaPrepaid,
			},
			APIOpts: map[string]interface{}{},
		},
	}
	if err := attsv1.ProcessEvent(context.Background(), args, &rplyev); err != nil {
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
