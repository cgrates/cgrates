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
		GetResourceProfileDrvF: func(ctx *context.Context, tnt string, id string) (*utils.ResourceProfile, error) {
			rsc := &utils.ResourceProfile{
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
	}, cfg, connMng)
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpRsc := TPResources{
		dm: dm,
	}
	rsc := &utils.ResourceProfile{
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
	err := tpRsc.exportItems(context.Background(), wrtr, "cgrates.org", []string{"ResGroup1"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsResourcesNoDbConn(t *testing.T) {
	engine.Cache.Clear(nil)
	wrtr := new(bytes.Buffer)
	tpRsc := TPResources{
		dm: nil,
	}
	rsc := &utils.ResourceProfile{
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
	err := tpRsc.exportItems(context.Background(), wrtr, "cgrates.org", []string{"ResGroup1"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

func TestTPEExportItemsResourcesIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpRsc := TPResources{
		dm: dm,
	}
	rsc := &utils.ResourceProfile{
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
	err := tpRsc.exportItems(context.Background(), wrtr, "cgrates.org", []string{"ResGroup2"})
	errExpect := "<NOT_FOUND> cannot find ResourceProfile with id: <ResGroup2>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
