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

package tpes

import (
	"bytes"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestTPEnewTPAttributes(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetAttributeProfileDrvF: func(ctx *context.Context, str1 string, str2 string) (*utils.AttributeProfile, error) {
			attr := &utils.AttributeProfile{
				Tenant:    utils.CGRateSorg,
				ID:        "TEST_ATTRIBUTES_TEST",
				FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
				Attributes: []*utils.Attribute{
					{
						Path:  utils.AccountField,
						Type:  utils.MetaConstant,
						Value: nil,
					},
					{
						Path:  "*tenant",
						Type:  utils.MetaConstant,
						Value: nil,
					},
				},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			}
			return attr, nil
		},
	}, cfg, connMng)
	exp := &TPAttributes{
		dm: dm,
	}
	rcv := newTPAttributes(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportItemsAttributes(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpAttr := TPAttributes{
		dm: dm,
	}
	attr := &utils.AttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_TEST",
		FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.AccountField,
				Type:  utils.MetaConstant,
				Value: nil,
			},
			{
				Path:  "*tenant",
				Type:  utils.MetaConstant,
				Value: nil,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	tpAttr.dm.SetAttributeProfile(context.Background(), attr, false)
	err := tpAttr.exportItems(context.Background(), wrtr, "cgrates.org", []string{"TEST_ATTRIBUTES_TEST"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsAttributesNoDbConn(t *testing.T) {
	engine.Cache.Clear(nil)
	wrtr := new(bytes.Buffer)
	tpAttr := TPAttributes{
		dm: nil,
	}
	attr := &utils.AttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_TEST",
		FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.AccountField,
				Type:  utils.MetaConstant,
				Value: nil,
			},
			{
				Path:  "*tenant",
				Type:  utils.MetaConstant,
				Value: nil,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	tpAttr.dm.SetAttributeProfile(context.Background(), attr, false)
	err := tpAttr.exportItems(context.Background(), wrtr, "cgrates.org", []string{"TEST_ATTRIBUTES_TEST"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

func TestTPEExportItemsAttributesIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpAct := TPAttributes{
		dm: dm,
	}
	attr := &utils.AttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_TEST",
		FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
		Attributes: []*utils.Attribute{
			{
				Path:  utils.AccountField,
				Type:  utils.MetaConstant,
				Value: nil,
			},
			{
				Path:  "*tenant",
				Type:  utils.MetaConstant,
				Value: nil,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	tpAct.dm.SetAttributeProfile(context.Background(), attr, false)
	err := tpAct.exportItems(context.Background(), wrtr, "cgrates.org", []string{"TEST_ATTRIBUTES"})
	errExpect := "<NOT_FOUND> cannot find AttributeProfile with id: <TEST_ATTRIBUTES>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
