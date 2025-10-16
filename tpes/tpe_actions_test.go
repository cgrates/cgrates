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

package tpes

import (
	"bytes"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestTPEnewTPActions(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: &engine.DataDBMock{
		GetActionProfileDrvF: func(ctx *context.Context, tenant string, ID string) (*utils.ActionProfile, error) {
			act := &utils.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "SET_BAL",
				FilterIDs: []string{
					"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
				Schedule: utils.MetaASAP,
				Actions: []*utils.APAction{
					{
						ID:   "SET_BAL",
						Type: utils.MetaSetBalance,
						Diktats: []*utils.APDiktat{
							{
								Opts: map[string]any{
									"*balancePath":  "MONETARY",
									"*balanceValue": "10",
								},
							}},
					},
				},
			}
			return act, nil
		},
	}}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMng)
	exp := &TPActions{
		dm: dm,
	}
	rcv := newTPActions(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportItemsActions(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	tpAct := TPActions{
		dm: dm,
	}
	act := &utils.ActionProfile{
		Tenant: "cgrates.org",
		ID:     "SET_BAL",
		FilterIDs: []string{
			"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Schedule: utils.MetaASAP,
		Actions: []*utils.APAction{
			{
				ID:   "SET_BAL",
				Type: utils.MetaSetBalance,
				Diktats: []*utils.APDiktat{
					{
						Opts: map[string]any{
							"*balancePath":  "MONETARY",
							"*balanceValue": "10",
						},
					}},
			},
		},
	}
	tpAct.dm.SetActionProfile(context.Background(), act, false)
	err := tpAct.exportItems(context.Background(), wrtr, "cgrates.org", []string{"SET_BAL"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsActionsEmpty(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	tpAct := TPActions{
		dm: dm,
	}
	act := &utils.ActionProfile{}
	tpAct.dm.SetActionProfile(context.Background(), act, false)
	err := tpAct.exportItems(context.Background(), wrtr, "cgrates.org", []string{})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsActionsNoDbConn(t *testing.T) {
	engine.Cache.Clear(nil)
	wrtr := new(bytes.Buffer)
	tpAct := TPActions{
		dm: nil,
	}
	act := &utils.ActionProfile{
		Tenant: "cgrates.org",
		ID:     "SET_BAL",
		FilterIDs: []string{
			"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Schedule: utils.MetaASAP,
		Actions: []*utils.APAction{
			{
				ID:   "SET_BAL",
				Type: utils.MetaSetBalance,
				Diktats: []*utils.APDiktat{
					{
						Opts: map[string]any{
							"*balancePath":  "MONETARY",
							"*balanceValue": "10",
						},
					}},
			},
		},
	}
	tpAct.dm.SetActionProfile(context.Background(), act, false)
	err := tpAct.exportItems(context.Background(), wrtr, "cgrates.org", []string{"SET_BAL"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

func TestTPEExportItemsActionsIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	tpAct := TPActions{
		dm: dm,
	}
	act := &utils.ActionProfile{
		Tenant: "cgrates.org",
		ID:     "SET_BAL",
		FilterIDs: []string{
			"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Schedule: utils.MetaASAP,
		Actions: []*utils.APAction{
			{
				ID:   "SET_BAL",
				Type: utils.MetaSetBalance,
				Diktats: []*utils.APDiktat{
					{
						Opts: map[string]any{
							"*balancePath":  "MONETARY",
							"*balanceValue": "10",
						},
					}},
			},
		},
	}
	tpAct.dm.SetActionProfile(context.Background(), act, false)
	err := tpAct.exportItems(context.Background(), wrtr, "cgrates.org", []string{"UNSET_BAL"})
	errExpect := "<NOT_FOUND> cannot find Actions id: <UNSET_BAL>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
