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
