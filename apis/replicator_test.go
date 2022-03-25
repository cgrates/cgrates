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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewReplicatorSv1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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
		Opts:   map[string]interface{}{},
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
				Opts: map[string]interface{}{
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

func TestReplicatorGetStatQueue(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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

func TestReplicatorGetFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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
	if !reflect.DeepEqual(fltr, &reply) {
		t.Errorf("Expected %v\n but received %v", fltr, reply)
	}
}

func TestReplicatorGetThreshold(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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

func TestReplicatorGetThresholdProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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

func TestReplicatorGetResource(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.Resource
	rp := NewReplicatorSv1(dm, v1)
	rsc := &engine.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup2",
		Usages: make(map[string]*engine.ResourceUsage),
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

func TestReplicatorGetResourceProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.ResourceProfile
	rp := NewReplicatorSv1(dm, v1)
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

func TestReplicatorGetRouteProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.RouteProfile
	rp := NewReplicatorSv1(dm, v1)
	rte := &engine.RouteProfile{
		ID:     "ROUTE_2003",
		Tenant: "cgrates.org",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*engine.Route{
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

func TestReplicatorGetAttributeProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.AttributeProfile
	rp := NewReplicatorSv1(dm, v1)
	attr := &engine.AttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_TEST",
		FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
		Attributes: []*engine.Attribute{
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

func TestGetChargerProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.ChargerProfile
	rp := NewReplicatorSv1(dm, v1)
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

func TestGetDispatcherProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.DispatcherProfile
	rp := NewReplicatorSv1(dm, v1)
	dsp := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Strategy:  utils.MetaFirst,
		StrategyParams: map[string]interface{}{
			utils.MetaDefaultRatio: "false",
		},
		Weight: 20,
		Hosts: engine.DispatcherHostProfiles{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]interface{}{"0": "192.168.54.203"},
				Blocker:   false,
			},
		},
	}
	rp.dm.SetDispatcherProfile(context.Background(), dsp, false)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:Dsp1"),
	}

	if err := rp.GetDispatcherProfile(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(dsp, &reply) {
		t.Errorf("Expected %v\n but received %v", dsp, reply)
	}
}

func TestGetDispatcherHost(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}

	var reply engine.DispatcherHost
	rp := NewReplicatorSv1(dm, v1)
	dsph := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:              "DSH1",
			Address:         "*internal",
			ConnectAttempts: 1,
			Reconnects:      3,
			ConnectTimeout:  time.Minute,
			ReplyTimeout:    2 * time.Minute,
		},
	}
	rp.dm.SetDispatcherHost(context.Background(), dsph)
	tntID := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:DSH1"),
	}

	if err := rp.GetDispatcherHost(context.Background(), tntID, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(dsph, &reply) {
		t.Errorf("Expected %v\n but received %v", dsph, reply)
	}
}

func TestGetItemLoadIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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

func TestReplicatorSetThresholdProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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
		APIOpts: map[string]interface{}{
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

func TestReplicatorSetAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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
			Opts:   map[string]interface{}{},
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
					Opts: map[string]interface{}{
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
		APIOpts: map[string]interface{}{
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
		Opts:   map[string]interface{}{},
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
				Opts: map[string]interface{}{
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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
		APIOpts: map[string]interface{}{
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

func TestReplicatorSetStatQueueProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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
		APIOpts: map[string]interface{}{
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

func TestReplicatorSetStatQueue(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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
		APIOpts: map[string]interface{}{
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
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
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
		APIOpts: map[string]interface{}{
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

func TestReplicatorSetResourceProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	v1 := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: engine.NewConnManager(cfg),
		ping:    struct{}{},
	}
	cfg.AdminSCfg().CachesConns = []string{"*internal"}
	var reply string
	rp := NewReplicatorSv1(dm, v1)
	rsp := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
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
		APIOpts: map[string]interface{}{
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
