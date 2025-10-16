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

func TestTPEnewTPAccounts(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: &engine.DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1 string, str2 string) (*utils.Account, error) {
			acc := &utils.Account{
				Tenant: "cgrates.org",
				ID:     "Account_simple",
				Opts:   map[string]any{},
				Balances: map[string]*utils.Balance{
					"VoiceBalance": {
						ID:        "VoiceBalance",
						FilterIDs: []string{"*string:~*req.Account:1001"},
						Weights: utils.DynamicWeights{
							{
								Weight: 12,
							},
						},
						Type: "*abstract",
						Opts: map[string]any{
							"Destination": "10",
						},
						Units: utils.NewDecimal(0, 0),
					},
				},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
			}
			return acc, nil
		},
	}}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMng)
	exp := &TPAccounts{
		dm: dm,
	}
	rcv := newTPAccounts(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportItemsAccount(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	// connMng := engine.NewConnManager(cfg)
	// dataDB, err := engine.NewDataDBConn(cfg.DbCfg().DBConns[utils.MetaDefault].Type,
	// 	cfg.DbCfg().DBConns[utils.MetaDefault].Host, cfg.DbCfg().DBConns[utils.MetaDefault].Port,
	// 	cfg.DbCfg().DBConns[utils.MetaDefault].Name, cfg.DbCfg().DBConns[utils.MetaDefault].User,
	// 	cfg.DbCfg().DBConns[utils.MetaDefault].Password, cfg.GeneralCfg().DBDataEncoding,
	// 	cfg.DbCfg().Opts, cfg.DbCfg().Items)
	// if err != nil {
	// 	t.Error(err)
	// }
	// defer dataDB.Close()
	// dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	// dm := engine.NewDataManager(dbCM, config.CgrConfig().CacheCfg(), connMng)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	tpAcc := TPAccounts{
		dm: dm,
	}
	acc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "Account_simple",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						Weight: 12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": "10",
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	tpAcc.dm.SetAccount(context.Background(), acc, false)
	err := tpAcc.exportItems(context.Background(), wrtr, "cgrates.org", []string{"Account_simple"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsAccountNoDbConn(t *testing.T) {
	wrtr := new(bytes.Buffer)
	tpAcc := TPAccounts{
		dm: nil,
	}
	acc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "Account_simple",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						Weight: 12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": "10",
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	tpAcc.dm.SetAccount(context.Background(), acc, false)
	err := tpAcc.exportItems(context.Background(), wrtr, "cgrates.org", []string{"Account_simple"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

// type mockWrtr struct {
// 	io.Writer
// }

// func (mockWrtr) Close() error { return utils.ErrNotImplemented }

// func TestTPEExportAccountHeaderError(t *testing.T) {
// 	byteBuff := new(bytes.Buffer)
// 	wrtr := mockWrtr{byteBuff}
// 	cfg := config.NewDefaultCGRConfig()
// 	connMng := engine.NewConnManager(cfg)
// 	dm := engine.NewDataManager(&engine.DataDBMock{
// 		GetAccountDrvF: func(ctx *context.Context, str1 string, str2 string) (*utils.Account, error) {
// 			acc := &utils.Account{
// 				Tenant: "cgrates.org",
// 				ID:     "Account_simple",
// 				Opts:   map[string]any{},
// 				Balances: map[string]*utils.Balance{
// 					"VoiceBalance": {
// 						ID:        "VoiceBalance",
// 						FilterIDs: []string{"*string:~*req.Account:1001"},
// 						Weights: utils.DynamicWeights{
// 							{
// 								Weight: 12,
// 							},
// 						},
// 						Type: "*abstract",
// 						Opts: map[string]any{
// 							"Destination": "10",
// 						},
// 						Units: utils.NewDecimal(0, 0),
// 					},
// 				},
// 				Weights: utils.DynamicWeights{
// 					{
// 						Weight: 10,
// 					},
// 				},
// 			}
// 			return acc, nil
// 		},
// 	}, cfg, connMng)
// 	tpAcc := &TPAccounts{
// 		dm: dm,
// 	}
// 	err := tpAcc.exportItems(context.Background(), wrtr, "cgrates.org", []string{"Accountt_simple"})
// 	if err != nil {
// 		t.Errorf("Expected nil\n but received %v", err)
// 	}
// }

func TestTPEExportItemsAccountIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	tpAcc := TPAccounts{
		dm: dm,
	}
	acc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "Account_complicated",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						Weight: 12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": "10",
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	tpAcc.dm.SetAccount(context.Background(), acc, false)
	err := tpAcc.exportItems(context.Background(), wrtr, "cgrates.net", []string{"Account_simple"})
	errExpect := "<NOT_FOUND> cannot find Account with id: <Account_simple>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
