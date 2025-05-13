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

func TestTPEnewTPRoutes(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetRouteProfileDrvF: func(ctx *context.Context, tnt string, id string) (*utils.RouteProfile, error) {
			rte := &utils.RouteProfile{
				ID:     "ROUTE_2003",
				Tenant: "cgrates.org",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*utils.Route{
					{
						ID: "route1",
						Weights: utils.DynamicWeights{
							{
								Weight: 20,
							},
						},
					},
				},
			}
			return rte, nil
		},
	}, cfg, connMng)
	exp := &TPRoutes{
		dm: dm,
	}
	rcv := newTPRoutes(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportRoutes(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpRte := TPRoutes{
		dm: dm,
	}
	rte := &utils.RouteProfile{
		ID:     "ROUTE_2003",
		Tenant: "cgrates.org",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*utils.Route{
			{
				ID: "route1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			},
		},
	}
	tpRte.dm.SetRouteProfile(context.Background(), rte, false)
	err := tpRte.exportItems(context.Background(), wrtr, "cgrates.org", []string{"ROUTE_2003"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsRoutesNoDbConn(t *testing.T) {
	engine.Cache.Clear(nil)
	wrtr := new(bytes.Buffer)
	tpRte := TPRoutes{
		dm: nil,
	}
	rte := &utils.RouteProfile{
		ID:     "ROUTE_2003",
		Tenant: "cgrates.org",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*utils.Route{
			{
				ID: "route1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			},
		},
	}
	tpRte.dm.SetRouteProfile(context.Background(), rte, false)
	err := tpRte.exportItems(context.Background(), wrtr, "cgrates.org", []string{"ROUTE_2003"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

func TestTPEExportItemsRoutesIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpRte := TPRoutes{
		dm: dm,
	}
	rte := &utils.RouteProfile{
		ID:     "ROUTE_2003",
		Tenant: "cgrates.org",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*utils.Route{
			{
				ID: "route1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			},
		},
	}
	tpRte.dm.SetRouteProfile(context.Background(), rte, false)
	err := tpRte.exportItems(context.Background(), wrtr, "cgrates.org", []string{"ROUTE_2004"})
	errExpect := "<NOT_FOUND> cannot find RouteProfile with id: <ROUTE_2004>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
