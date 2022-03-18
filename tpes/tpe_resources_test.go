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

func TestTPEnewTPResources(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetResourceProfileDrvF: func(ctx *context.Context, tnt string, id string) (*engine.ResourceProfile, error) {
			rsc := &engine.ResourceProfile{
				Tenant:            "cgrates.org",
				ID:                "ResGroup1",
				FilterIDs:         []string{"*string:~*req.Account:1001"},
				Limit:             10,
				AllocationMessage: "Approved",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					}},
				ThresholdIDs: []string{utils.MetaNone},
			}
			return rsc, nil
		},
	}, nil, connMng)
	exp := &TPResources{
		dm: dm,
	}
	rcv := newTPResources(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportItemsResources(t *testing.T) {
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
	tpRsc := TPResources{
		dm: dm,
	}
	rsc := &engine.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResGroup1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		Limit:             10,
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			}},
		ThresholdIDs: []string{utils.MetaNone},
	}
	tpRsc.dm.SetResourceProfile(context.Background(), rsc, false)
	err = tpRsc.exportItems(context.Background(), wrtr, "cgrates.org", []string{"ResGroup1"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsResourcesIDNotFound(t *testing.T) {
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
	tpRsc := TPResources{
		dm: dm,
	}
	rsc := &engine.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResGroup1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		Limit:             10,
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			}},
		ThresholdIDs: []string{utils.MetaNone},
	}
	tpRsc.dm.SetResourceProfile(context.Background(), rsc, false)
	err = tpRsc.exportItems(context.Background(), wrtr, "cgrates.org", []string{"ResGroup2"})
	errExpect := "<NOT_FOUND> cannot find ResourceProfile with id: <ResGroup2>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
