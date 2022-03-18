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

func TestTPEnewTPChargers(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*engine.ChargerProfile, error) {
			chgr := &engine.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "Chargers1",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{"*none"},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			}
			return chgr, nil
		},
	}, nil, connMng)
	exp := &TPChargers{
		dm: dm,
	}
	rcv := newTPChargers(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportItemsChargers(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dataDB, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	tpChgr := TPChargers{
		dm: dm,
	}
	chgr := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Chargers1",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	tpChgr.dm.SetChargerProfile(context.Background(), chgr, false)
	err = tpChgr.exportItems(context.Background(), wrtr, "cgrates.org", []string{"Chargers1"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsChargersIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dataDB, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	tpChgr := TPChargers{
		dm: dm,
	}
	chgr := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Chargers1",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	tpChgr.dm.SetChargerProfile(context.Background(), chgr, false)
	err = tpChgr.exportItems(context.Background(), wrtr, "cgrates.org", []string{"Chargers2"})
	errExpect := "<NOT_FOUND> cannot find ChargerProfile with id: <Chargers2>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
