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
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*utils.ChargerProfile, error) {
			chgr := &utils.ChargerProfile{
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
	}, cfg, connMng)
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
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpChgr := TPChargers{
		dm: dm,
	}
	chgr := &utils.ChargerProfile{
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
	err := tpChgr.exportItems(context.Background(), wrtr, "cgrates.org", []string{"Chargers1"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsChargersNoDbConn(t *testing.T) {
	engine.Cache.Clear(nil)
	wrtr := new(bytes.Buffer)
	tpChgr := TPChargers{
		dm: nil,
	}
	chgr := &utils.ChargerProfile{
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
	err := tpChgr.exportItems(context.Background(), wrtr, "cgrates.org", []string{"Chargers1"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

func TestTPEExportItemsChargersIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpChgr := TPChargers{
		dm: dm,
	}
	chgr := &utils.ChargerProfile{
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
	err := tpChgr.exportItems(context.Background(), wrtr, "cgrates.org", []string{"Chargers2"})
	errExpect := "<NOT_FOUND> cannot find ChargerProfile with id: <Chargers2>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
