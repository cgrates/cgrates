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

package apis

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewReplicatorSv1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	rcv := NewReplicatorSv1(dm, v1)
	exp := &ReplicatorSv1{
		dm: dm,
		v1: v1,
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}
}

func TestReplicatorGetAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.Account
	rp := NewReplicatorSv1(dm, v1)
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
	rp.dm.SetAccount(context.Background(), acc, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:Account_simple"),
	}

	if err := rp.GetAccount(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(acc, &reply) {
		t.Errorf("Expected %v\n but received %v", acc, reply)
	}
}

func TestReplicatorGetAccountError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.Account
	rp := NewReplicatorSv1(dm, v1)
	acc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "Account_simple_2",
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
	rp.dm.SetAccount(context.Background(), acc, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:Account_simple"),
	}

	if err := rp.GetAccount(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorGetStatQueue(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.StatQueue
	rp := NewReplicatorSv1(dm, v1)

	stq := &engine.StatQueue{
		Tenant: "cgrates.org",
		ID:     "sq1",
		SQMetrics: map[string]engine.StatMetric{
			utils.MetaACD: engine.NewACD(0, "", nil),
			utils.MetaTCD: engine.NewTCD(0, "", nil),
		},
	}
	rp.dm.SetStatQueue(context.Background(), stq)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:sq1"),
	}

	if err := rp.GetStatQueue(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(stq, &reply) {
		t.Errorf("Expected %v\n but received %v", stq, reply)
	}
}

func TestReplicatorGetStatQueueErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.StatQueue
	rp := NewReplicatorSv1(dm, v1)

	stq := &engine.StatQueue{
		Tenant: "cgrates.org",
		ID:     "sq2",
		SQMetrics: map[string]engine.StatMetric{
			utils.MetaACD: engine.NewACD(0, "", nil),
			utils.MetaTCD: engine.NewTCD(0, "", nil),
		},
	}
	rp.dm.SetStatQueue(context.Background(), stq)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:sq1"),
	}

	if err := rp.GetStatQueue(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorGetFilter(t *testing.T) {
	engine.Cache.Clear(nil)
	defer engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.Filter
	rp := NewReplicatorSv1(dm, v1)

	fltr := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_prf",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1004", "6774", "22312"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*opts.Subsystems",
				Values:  []string{"*attributes"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destinations",
				Values:  []string{"+0775", "+442"},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.NumberOfEvents",
			},
		},
	}
	rp.dm.SetFilter(context.Background(), fltr, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:fltr_for_prf"),
	}

	if err := rp.GetFilter(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if utils.ToJSON(fltr) != utils.ToJSON(reply) {
		t.Errorf("Expected %+v\n but received %+v", utils.ToJSON(fltr), utils.ToJSON(reply))
	}
}

func TestReplicatorGetFilterError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.Filter
	rp := NewReplicatorSv1(dm, v1)

	fltr := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_not_for_prf",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1004", "6774", "22312"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*opts.Subsystems",
				Values:  []string{"*attributes"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destinations",
				Values:  []string{"+0775", "+442"},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.NumberOfEvents",
			},
		},
	}
	rp.dm.SetFilter(context.Background(), fltr, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:fltr_for_prf"),
	}

	if err := rp.GetFilter(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorGetThreshold(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.Threshold
	rp := NewReplicatorSv1(dm, v1)

	thd := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_2",
		Hits:   0,
	}
	rp.dm.SetThreshold(context.Background(), thd)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:THD_2"),
	}

	if err := rp.GetThreshold(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(thd, &reply) {
		t.Errorf("Expected %v\n but received %v", thd, reply)
	}
}

func TestReplicatorGetThresholdError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.Threshold
	rp := NewReplicatorSv1(dm, v1)

	thd := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_3",
		Hits:   0,
	}
	rp.dm.SetThreshold(context.Background(), thd)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:THD_2"),
	}

	if err := rp.GetThreshold(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorGetThresholdProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.ThresholdProfile
	rp := NewReplicatorSv1(dm, v1)

	thd := &engine.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_2",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}
	rp.dm.SetThresholdProfile(context.Background(), thd, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:THD_2"),
	}

	if err := rp.GetThresholdProfile(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(thd, &reply) {
		t.Errorf("Expected %v\n but received %v", thd, reply)
	}
}

func TestReplicatorGetThresholdProfileError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.ThresholdProfile
	rp := NewReplicatorSv1(dm, v1)

	thd := &engine.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_4",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}
	rp.dm.SetThresholdProfile(context.Background(), thd, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:THD_2"),
	}

	if err := rp.GetThresholdProfile(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorGetResource(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.Resource
	rp := NewReplicatorSv1(dm, v1)
	rsc := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup2",
		Usages: make(map[string]*utils.ResourceUsage),
	}
	rp.dm.SetResource(context.Background(), rsc)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ResGroup2"),
	}

	if err := rp.GetResource(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rsc, &reply) {
		t.Errorf("Expected %v\n but received %v", rsc, reply)
	}
}

func TestReplicatorGetResourceError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.Resource
	rp := NewReplicatorSv1(dm, v1)
	rsc := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup3",
		Usages: make(map[string]*utils.ResourceUsage),
	}
	rp.dm.SetResource(context.Background(), rsc)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ResGroup2"),
	}

	if err := rp.GetResource(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorGetResourceProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.ResourceProfile
	rp := NewReplicatorSv1(dm, v1)
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
	rp.dm.SetResourceProfile(context.Background(), rsc, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ResGroup1"),
	}

	if err := rp.GetResourceProfile(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rsc, &reply) {
		t.Errorf("Expected %v\n but received %v", rsc, reply)
	}
}

func TestReplicatorGetResourceProfileError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.ResourceProfile
	rp := NewReplicatorSv1(dm, v1)
	rsc := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResGroup10",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		Limit:             10,
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			}},
		ThresholdIDs: []string{utils.MetaNone},
	}
	rp.dm.SetResourceProfile(context.Background(), rsc, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ResGroup1"),
	}

	if err := rp.GetResourceProfile(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorGetRouteProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.RouteProfile
	rp := NewReplicatorSv1(dm, v1)
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
	rp.dm.SetRouteProfile(context.Background(), rte, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ROUTE_2003"),
	}

	if err := rp.GetRouteProfile(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rte, &reply) {
		t.Errorf("Expected %v\n but received %v", rte, reply)
	}
}

func TestReplicatorGetRouteProfileError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.RouteProfile
	rp := NewReplicatorSv1(dm, v1)
	rte := &utils.RouteProfile{
		ID:     "ROUTE_2001",
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
	rp.dm.SetRouteProfile(context.Background(), rte, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ROUTE_2003"),
	}

	if err := rp.GetRouteProfile(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorGetAttributeProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.AttributeProfile
	rp := NewReplicatorSv1(dm, v1)
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
	rp.dm.SetAttributeProfile(context.Background(), attr, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:TEST_ATTRIBUTES_TEST"),
	}

	if err := rp.GetAttributeProfile(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(attr, &reply) {
		t.Errorf("Expected %v\n but received %v", attr, reply)
	}
}

func TestReplicatorGetAttributeProfileError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.AttributeProfile
	rp := NewReplicatorSv1(dm, v1)
	attr := &utils.AttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES",
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
	rp.dm.SetAttributeProfile(context.Background(), attr, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:TEST_ATTRIBUTES_TEST"),
	}

	if err := rp.GetAttributeProfile(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorGetChargerProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.ChargerProfile
	rp := NewReplicatorSv1(dm, v1)
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
	rp.dm.SetChargerProfile(context.Background(), chgr, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:Chargers1"),
	}

	if err := rp.GetChargerProfile(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(chgr, &reply) {
		t.Errorf("Expected %v\n but received %v", chgr, reply)
	}
}

func TestReplicatorGetChargerProfileError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply utils.ChargerProfile
	rp := NewReplicatorSv1(dm, v1)
	chgr := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Chargers100",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	rp.dm.SetChargerProfile(context.Background(), chgr, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:Chargers1"),
	}

	if err := rp.GetChargerProfile(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorGetItemLoadIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply map[string]int64
	rp := NewReplicatorSv1(dm, v1)
	itmLIDs := map[string]int64{
		"ID_1": 21,
	}
	rp.dm.SetLoadIDs(context.Background(), itmLIDs)
	tntID := &utils.StringWithAPIOpts{
		Tenant: "cgrates.org",
	}

	if err := rp.GetItemLoadIDs(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(&itmLIDs, &reply) {
		t.Errorf("Expected %v\n but received %v", itmLIDs, reply)
	}
}

func TestReplicatorGetItemLoadIDsError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply map[string]int64
	rp := NewReplicatorSv1(dm, v1)
	itmLIDs := map[string]int64{
		"ID_1": 31,
	}
	rp.dm.SetLoadIDs(context.Background(), itmLIDs)
	tntID := &utils.StringWithAPIOpts{
		Tenant: "cgrates.org",
	}

	if err := rp.GetItemLoadIDs(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(&itmLIDs, &reply) {
		t.Errorf("Expected %v\n but received %v", itmLIDs, reply)
	}
}

func TestReplicatorSetThresholdProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	th := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "THD_100",
			FilterIDs:        []string{"*string:~*req.Account:1001"},
			ActionProfileIDs: []string{"actPrfID"},
			MaxHits:          7,
			MinHits:          0,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Async: true,
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetThresholdProfile(context.Background(), th, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetThresholdProfile(context.Background(), "cgrates.org", "THD_100", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	exp := &engine.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_100",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}
}

func TestReplicatorSetThresholdProfileErr1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	th := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "THD_100",
			FilterIDs:        []string{"*string:~*req.Account:1001"},
			ActionProfileIDs: []string{"actPrfID"},
			MaxHits:          7,
			MinHits:          0,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Async: true,
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.SetThresholdProfile(context.Background(), th, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorSetAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	acc := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "Account_simple1",
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
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetAccount(context.Background(), acc, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetAccount(context.Background(), "cgrates.org", "Account_simple1")
	if err != nil {
		t.Error(err)
	}
	exp := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "Account_simple1",
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
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}
}

func TestReplicatorSetThreshold(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	thd := &engine.ThresholdWithAPIOpts{
		Threshold: &engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_20",
			Hits:   0,
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetThreshold(context.Background(), thd, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetThreshold(context.Background(), "cgrates.org", "THD_20", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	exp := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_20",
		Hits:   0,
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}
}

func TestReplicatorSetThresholdErr1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	thd := &engine.ThresholdWithAPIOpts{
		Threshold: &engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_20",
			Hits:   0,
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.SetThreshold(context.Background(), thd, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorSetStatQueueProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	sq := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "SQ_20",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			QueueLength: 14,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
				{
					MetricID: utils.MetaTCD,
				},
				{
					MetricID: utils.MetaPDD,
				},
				{
					MetricID: utils.MetaTCC,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetStatQueueProfile(context.Background(), sq, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetStatQueueProfile(context.Background(), "cgrates.org", "SQ_20", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, sq.StatQueueProfile) {
		t.Errorf("Expected %v\n but received %v", sq.StatQueueProfile, rcv)
	}
}

func TestReplicatorSetStatQueueProfileErr1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	sq := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "SQ_20",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			QueueLength: 14,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
				{
					MetricID: utils.MetaTCD,
				},
				{
					MetricID: utils.MetaPDD,
				},
				{
					MetricID: utils.MetaTCC,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.SetStatQueueProfile(context.Background(), sq, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}
func TestReplicatorSetStatQueue(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	sq := &engine.StatQueueWithAPIOpts{
		StatQueue: &engine.StatQueue{
			Tenant: "cgrates.org",
			ID:     "sq11",
			SQMetrics: map[string]engine.StatMetric{
				utils.MetaACD: engine.NewACD(0, "", nil),
				utils.MetaTCD: engine.NewTCD(0, "", nil),
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetStatQueue(context.Background(), sq, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetStatQueue(context.Background(), "cgrates.org", "sq11", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, sq.StatQueue) {
		t.Errorf("Expected %v\n but received %v", sq.StatQueue, rcv)
	}
}

func TestReplicatorSetFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_prf",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Subject",
					Values:  []string{"1004", "6774", "22312"},
				},
				{
					Type:    utils.MetaString,
					Element: "~*opts.Subsystems",
					Values:  []string{"*attributes"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destinations",
					Values:  []string{"+0775", "+442"},
				},
				{
					Type:    utils.MetaExists,
					Element: "~*req.NumberOfEvents",
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetFilter(context.Background(), fltr, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetFilter(context.Background(), "cgrates.org", "fltr_for_prf", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, fltr.Filter) {
		t.Errorf("Expected %v\n but received %v", fltr.Filter, rcv)
	}
}

func TestReplicatorSetFilterErr1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_prf",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Subject",
					Values:  []string{"1004", "6774", "22312"},
				},
				{
					Type:    utils.MetaString,
					Element: "~*opts.Subsystems",
					Values:  []string{"*attributes"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destinations",
					Values:  []string{"+0775", "+442"},
				},
				{
					Type:    utils.MetaExists,
					Element: "~*req.NumberOfEvents",
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.SetFilter(context.Background(), fltr, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorSetResourceProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rsp := &utils.ResourceProfileWithAPIOpts{
		ResourceProfile: &utils.ResourceProfile{
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
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetResourceProfile(context.Background(), rsp, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetResourceProfile(context.Background(), "cgrates.org", "ResGroup1", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, rsp.ResourceProfile) {
		t.Errorf("Expected %v\n but received %v", rsp.ResourceProfile, rcv)
	}
}

func TestReplicatorSetResourceProfileErr1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rsp := &utils.ResourceProfileWithAPIOpts{
		ResourceProfile: &utils.ResourceProfile{
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
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.SetResourceProfile(context.Background(), rsp, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorSetResource(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rs := &utils.ResourceWithAPIOpts{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
			Usages: make(map[string]*utils.ResourceUsage),
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetResource(context.Background(), rs, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetResource(context.Background(), "cgrates.org", "ResGroup2", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, rs.Resource) {
		t.Errorf("Expected %v\n but received %v", rs.Resource, rcv)
	}
}

func TestReplicatorSetResourceErr1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rs := &utils.ResourceWithAPIOpts{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
			Usages: make(map[string]*utils.ResourceUsage),
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.SetResource(context.Background(), rs, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}
func TestReplicatorSetRouteProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rtp := &utils.RouteProfileWithAPIOpts{
		RouteProfile: &utils.RouteProfile{
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
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetRouteProfile(context.Background(), rtp, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetRouteProfile(context.Background(), "cgrates.org", "ROUTE_2003", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, rtp.RouteProfile) {
		t.Errorf("Expected %v\n but received %v", rtp.RouteProfile, rcv)
	}
}

func TestReplicatorSetRouteProfileErr1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rtp := &utils.RouteProfileWithAPIOpts{
		RouteProfile: &utils.RouteProfile{
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
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.SetRouteProfile(context.Background(), rtp, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorSetAttributeProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	attrPrf := &utils.AttributeProfileWithAPIOpts{
		AttributeProfile: &utils.AttributeProfile{
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
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetAttributeProfile(context.Background(), attrPrf, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetAttributeProfile(context.Background(), "cgrates.org", "TEST_ATTRIBUTES_TEST", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, attrPrf.AttributeProfile) {
		t.Errorf("Expected %v\n but received %v", attrPrf.AttributeProfile, rcv)
	}
}

func TestReplicatorSetAttributeProfileErr1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	attrPrf := &utils.AttributeProfileWithAPIOpts{
		AttributeProfile: &utils.AttributeProfile{
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
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.SetAttributeProfile(context.Background(), attrPrf, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorChargerProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	chgrPrf := &utils.ChargerProfileWithAPIOpts{
		ChargerProfile: &utils.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Chargers1",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetChargerProfile(context.Background(), chgrPrf, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetChargerProfile(context.Background(), "cgrates.org", "Chargers1", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, chgrPrf.ChargerProfile) {
		t.Errorf("Expected %v\n but received %v", chgrPrf.ChargerProfile, rcv)
	}
}

func TestReplicatorChargerProfileErr1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	chgrPrf := &utils.ChargerProfileWithAPIOpts{
		ChargerProfile: &utils.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Chargers1",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.SetChargerProfile(context.Background(), chgrPrf, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveThreshold(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	thd := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_2",
		Hits:   0,
	}
	if err := rp.dm.SetThreshold(context.Background(), thd); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:THD_2"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveThreshold(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveThresholdErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	thd := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_2",
		Hits:   0,
	}
	if err := rp.dm.SetThreshold(context.Background(), thd); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:THD_2"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveThreshold(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
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
	if err := rp.dm.SetAccount(context.Background(), acc, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:Account_simple"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveAccount(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveStatQueue(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	sq := &engine.StatQueue{
		Tenant: "cgrates.org",
		ID:     "sq11",
		SQMetrics: map[string]engine.StatMetric{
			utils.MetaACD: engine.NewACD(0, "", nil),
			utils.MetaTCD: engine.NewTCD(0, "", nil),
		},
	}
	if err := rp.dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:sq11"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveStatQueue(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveStatQueueErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	sq := &engine.StatQueue{
		Tenant: "cgrates.org",
		ID:     "sq11",
		SQMetrics: map[string]engine.StatMetric{
			utils.MetaACD: engine.NewACD(0, "", nil),
			utils.MetaTCD: engine.NewTCD(0, "", nil),
		},
	}
	if err := rp.dm.SetStatQueue(context.Background(), sq); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:sq11"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveStatQueue(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	fltr := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_prf",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1004", "6774", "22312"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*opts.Subsystems",
				Values:  []string{"*attributes"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destinations",
				Values:  []string{"+0775", "+442"},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.NumberOfEvents",
			},
		},
	}
	if err := rp.dm.SetFilter(context.Background(), fltr, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:fltr_for_prf"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveFilter(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveFilterErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	fltr := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_prf",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1004", "6774", "22312"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*opts.Subsystems",
				Values:  []string{"*attributes"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destinations",
				Values:  []string{"+0775", "+442"},
			},
			{
				Type:    utils.MetaExists,
				Element: "~*req.NumberOfEvents",
			},
		},
	}
	if err := rp.dm.SetFilter(context.Background(), fltr, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:fltr_for_prf"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveFilter(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveThresholdProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	thd := &engine.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_2",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}
	if err := rp.dm.SetThresholdProfile(context.Background(), thd, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:THD_2"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveThresholdProfile(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveThresholdProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	thd := &engine.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_2",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Async: true,
	}
	if err := rp.dm.SetThresholdProfile(context.Background(), thd, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:THD_2"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveThresholdProfile(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveStatQueueProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	sq := &engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ_20",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		QueueLength: 14,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaASR,
			},
			{
				MetricID: utils.MetaTCD,
			},
			{
				MetricID: utils.MetaPDD,
			},
			{
				MetricID: utils.MetaTCC,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	if err := rp.dm.SetStatQueueProfile(context.Background(), sq, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:SQ_20"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveStatQueueProfile(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveStatQueueProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	sq := &engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ_20",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		QueueLength: 14,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaASR,
			},
			{
				MetricID: utils.MetaTCD,
			},
			{
				MetricID: utils.MetaPDD,
			},
			{
				MetricID: utils.MetaTCC,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	if err := rp.dm.SetStatQueueProfile(context.Background(), sq, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:SQ_20"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveStatQueueProfile(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveResource(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rsc := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup2",
		Usages: make(map[string]*utils.ResourceUsage),
	}
	if err := rp.dm.SetResource(context.Background(), rsc); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ResGroup2"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveResource(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveResourceErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rsc := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup2",
		Usages: make(map[string]*utils.ResourceUsage),
	}
	if err := rp.dm.SetResource(context.Background(), rsc); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ResGroup2"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}

	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveResource(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveResourceProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rscPrf := &utils.ResourceProfile{
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
	if err := rp.dm.SetResourceProfile(context.Background(), rscPrf, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ResGroup1"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveResourceProfile(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveResourceProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rscPrf := &utils.ResourceProfile{
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
	if err := rp.dm.SetResourceProfile(context.Background(), rscPrf, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ResGroup1"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}

	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveResourceProfile(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveRouteProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rtPf := &utils.RouteProfile{
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
	if err := rp.dm.SetRouteProfile(context.Background(), rtPf, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ROUTE_2003"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveRouteProfile(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveRouteProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rtPf := &utils.RouteProfile{
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
	if err := rp.dm.SetRouteProfile(context.Background(), rtPf, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:ROUTE_2003"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}

	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveRouteProfile(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveAttributeProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	attrPrf := &utils.AttributeProfile{
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
	if err := rp.dm.SetAttributeProfile(context.Background(), attrPrf, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:TEST_ATTRIBUTES_TEST"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveAttributeProfile(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveAttributeProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	attrPrf := &utils.AttributeProfile{
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
	if err := rp.dm.SetAttributeProfile(context.Background(), attrPrf, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:TEST_ATTRIBUTES_TEST"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}

	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveAttributeProfile(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveChargerProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	chgrPrf := &utils.ChargerProfile{
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
	if err := rp.dm.SetChargerProfile(context.Background(), chgrPrf, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:Chargers1"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveChargerProfile(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveChargerProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	chgrPrf := &utils.ChargerProfile{
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
	if err := rp.dm.SetChargerProfile(context.Background(), chgrPrf, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:Chargers1"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}

	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveChargerProfile(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorGetRateProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply utils.RateProfile
	rp := NewReplicatorSv1(dm, v1)
	rtPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(int64(0*time.Second), 0),
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(12, 2),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Minute), 0),
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:      utils.NewDecimal(1234, 3),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}
	rp.dm.SetRateProfile(context.Background(), rtPrf, false, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:RP1"),
	}

	if err := rp.GetRateProfile(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rtPrf, &reply) {
		t.Errorf("Expected %v\n but received %v", rtPrf, reply)
	}
}

func TestReplicatorGetRateProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply utils.RateProfile
	rp := NewReplicatorSv1(dm, v1)
	rtPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(int64(0*time.Second), 0),
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(12, 2),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Minute), 0),
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:      utils.NewDecimal(1234, 3),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}
	rp.dm.SetRateProfile(context.Background(), rtPrf, false, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:RP2"),
	}

	if err := rp.GetRateProfile(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorGetActionProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply utils.ActionProfile
	rp := NewReplicatorSv1(dm, v1)
	actPrf := &utils.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
		Actions: []*utils.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*utils.APDiktat{{
					Opts: map[string]any{
						"*balancePath":  "~*balance.TestBalance.Value",
						"*balanceValue": "10",
					},
				}},
			},
		},
	}
	rp.dm.SetActionProfile(context.Background(), actPrf, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:AP1"),
	}

	if err := rp.GetActionProfile(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actPrf, &reply) {
		t.Errorf("Expected %v\n but received %v", actPrf, reply)
	}
}

func TestReplicatorGetActionProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply utils.ActionProfile
	rp := NewReplicatorSv1(dm, v1)
	actPrf := &utils.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
		Actions: []*utils.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*utils.APDiktat{{
					Opts: map[string]any{
						"*balancePath":  "~*balance.TestBalance.Value",
						"*balanceValue": "10",
					},
				}},
			},
		},
	}
	rp.dm.SetActionProfile(context.Background(), actPrf, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:AP2"),
	}

	if err := rp.GetActionProfile(context.Background(), tntID, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestReplicatorSetRateProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rpp := &utils.RateProfileWithAPIOpts{
		RateProfile: &utils.RateProfile{
			Tenant:    "cgrates.org",
			ID:        "RP1",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			MinCost:         utils.NewDecimal(1, 1),
			MaxCost:         utils.NewDecimal(6, 1),
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
					Weights: utils.DynamicWeights{
						{
							Weight: 0,
						},
					},
					ActivationTimes: "* * * * 1-5",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(int64(0*time.Second), 0),
							FixedFee:      utils.NewDecimal(0, 0),
							RecurrentFee:  utils.NewDecimal(12, 2),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
							Increment:     utils.NewDecimal(int64(time.Minute), 0),
						},
						{
							IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
							FixedFee:      utils.NewDecimal(1234, 3),
							RecurrentFee:  utils.NewDecimal(6, 2),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetRateProfile(context.Background(), rpp, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetRateProfile(context.Background(), "cgrates.org", "RP1", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, rpp.RateProfile) {
		t.Errorf("Expected %v\n but received %v", rpp.RateProfile, rcv)
	}
}

func TestReplicatorSetRateProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rpp := &utils.RateProfileWithAPIOpts{
		RateProfile: &utils.RateProfile{
			Tenant:    "cgrates.org",
			ID:        "RP1",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			MinCost:         utils.NewDecimal(1, 1),
			MaxCost:         utils.NewDecimal(6, 1),
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
					Weights: utils.DynamicWeights{
						{
							Weight: 0,
						},
					},
					ActivationTimes: "* * * * 1-5",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(int64(0*time.Second), 0),
							FixedFee:      utils.NewDecimal(0, 0),
							RecurrentFee:  utils.NewDecimal(12, 2),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
							Increment:     utils.NewDecimal(int64(time.Minute), 0),
						},
						{
							IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
							FixedFee:      utils.NewDecimal(1234, 3),
							RecurrentFee:  utils.NewDecimal(6, 2),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}

	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.SetRateProfile(context.Background(), rpp, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorSetActionProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant:    "cgrates.org",
			ID:        "AP1",
			FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
			Actions: []*utils.APAction{
				{
					ID:        "TOPUP",
					FilterIDs: []string{},
					Type:      "*topup",
					Diktats: []*utils.APDiktat{{
						Opts: map[string]any{
							"*balancePath":  "~*balance.TestBalance.Value",
							"*balanceValue": "10",
						},
					}},
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	if err := rp.SetActionProfile(context.Background(), actPrf, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := rp.dm.GetActionProfile(context.Background(), "cgrates.org", "AP1", false, false, utils.GenUUID())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, actPrf.ActionProfile) {
		t.Errorf("Expected %v\n but received %v", actPrf.ActionProfile, rcv)
	}
}

func TestReplicatorSetActionProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant:    "cgrates.org",
			ID:        "AP1",
			FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
			Actions: []*utils.APAction{
				{
					ID:        "TOPUP",
					FilterIDs: []string{},
					Type:      "*topup",
					Diktats: []*utils.APDiktat{{
						Opts: map[string]any{
							"*balancePath":  "~*balance.TestBalance.Value",
							"*balanceValue": "10",
						},
					}},
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}
	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.SetActionProfile(context.Background(), actPrf, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveRateProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rtp := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(int64(0*time.Second), 0),
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(12, 2),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Minute), 0),
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:      utils.NewDecimal(1234, 3),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}
	if err := rp.dm.SetRateProfile(context.Background(), rtp, false, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:RP1"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveRateProfile(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveRateProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rtp := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(int64(0*time.Second), 0),
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(12, 2),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Minute), 0),
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:      utils.NewDecimal(1234, 3),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}
	if err := rp.dm.SetRateProfile(context.Background(), rtp, false, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:RP1"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}

	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveRateProfile(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}

func TestReplicatorRemoveActionProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	actPrf := &utils.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
		Actions: []*utils.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*utils.APDiktat{{
					Opts: map[string]any{
						"*balancePath":  "~*balance.TestBalance.Value",
						"*balanceValue": "10",
					},
				}},
			},
		},
	}
	if err := rp.dm.SetActionProfile(context.Background(), actPrf, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:AP1"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}
	err := rp.RemoveActionProfile(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestReplicatorRemoveActionProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	actPrf := &utils.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "AP1",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
		Actions: []*utils.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*utils.APDiktat{{
					Opts: map[string]any{
						"*balancePath":  "~*balance.TestBalance.Value",
						"*balanceValue": "10",
					},
				}},
			},
		},
	}
	if err := rp.dm.SetActionProfile(context.Background(), actPrf, false); err != nil {
		t.Error(err)
	}
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:AP1"),
		APIOpts: map[string]any{
			utils.MetaCache: utils.OK,
		},
	}

	errExpect := "nil rpc in argument method:  in: <nil> out"
	if err := rp.RemoveActionProfile(context.Background(), args, &reply); !strings.Contains(err.Error(), errExpect) {
		t.Errorf("Expected error to include %v", errExpect)
	}
}
