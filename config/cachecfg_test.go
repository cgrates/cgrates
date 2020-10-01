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
package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestAsTransCacheConfig(t *testing.T) {
	a := &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			"test": {
				Limit:     50,
				TTL:       time.Duration(60 * time.Second),
				StaticTTL: true,
				Precache:  true,
			},
		},
	}
	expected := map[string]*ltcache.CacheConfig{
		"test": {
			MaxItems:  50,
			TTL:       time.Duration(60 * time.Second),
			StaticTTL: true,
		},
	}
	reply := a.AsTransCacheConfig()
	if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %+v, received: %+v", expected, utils.ToJSON(reply))
	}
}

func TestCacheCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &CacheJsonCfg{
		Partitions: &map[string]*CacheParamJsonCfg{
			utils.MetaDestinations: {
				Limit:      utils.IntPointer(10),
				Ttl:        utils.StringPointer("2"),
				Static_ttl: utils.BoolPointer(true),
				Precache:   utils.BoolPointer(true),
				Replicate:  utils.BoolPointer(true),
			},
		},
		Replication_conns: &[]string{"conn1", "conn2"},
	}
	expected := &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			utils.MetaDestinations: {Limit: 10, TTL: time.Duration(2), StaticTTL: true, Precache: true, Replicate: true},
		},
		ReplicationConns: []string{"conn1", "conn2"},
	}
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.cacheCfg.loadFromJsonCfg(jsonCfg); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(expected.Partitions[utils.MetaDestinations], jsnCfg.cacheCfg.Partitions[utils.MetaDestinations]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected.Partitions[utils.MetaDestinations]),
				utils.ToJSON(jsnCfg.cacheCfg.Partitions[utils.MetaDestinations]))
		} else if !reflect.DeepEqual(jsnCfg.cacheCfg.ReplicationConns, expected.ReplicationConns) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected.ReplicationConns), utils.ToJSON(jsnCfg.cacheCfg.ReplicationConns))
		}
	}
}

func TestCacheParamCfgloadFromJsonCfg(t *testing.T) {
	json := &CacheParamJsonCfg{
		Limit:      utils.IntPointer(5),
		Ttl:        utils.StringPointer("1s"),
		Static_ttl: utils.BoolPointer(true),
		Precache:   utils.BoolPointer(true),
	}
	expected := &CacheParamCfg{
		Limit:     5,
		TTL:       time.Duration(time.Second),
		StaticTTL: true,
		Precache:  true,
	}
	rcv := new(CacheParamCfg)
	if err := rcv.loadFromJsonCfg(json); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestCachesCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"caches":{
			"partitions": {
				"*destinations": {"limit": 10000, "ttl": "", "static_ttl": false, "precache": true, "replicate": true},
				},
			},
		}`
	eMap := map[string]interface{}{
		utils.PartitionsCfg: map[string]interface{}{
			utils.MetaDestinations: map[string]interface{}{"limit": 10000, "ttl": "", "static_ttl": false, "precache": true, "replicate": true},
		},
		utils.ReplicationConnsCfg: []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		newMap := cgrCfg.cacheCfg.AsMapInterface()
		if !reflect.DeepEqual(newMap[utils.PartitionsCfg].(map[string]interface{})[utils.MetaDestinations],
			eMap[utils.PartitionsCfg].(map[string]interface{})[utils.MetaDestinations]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.PartitionsCfg].(map[string]interface{})[utils.MetaDestinations],
				newMap[utils.PartitionsCfg].(map[string]interface{})[utils.MetaDestinations])
		}
	}
}

func TestCachesCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
"caches":{
	"partitions": {
		"*rating_plans": {"limit": 10, "ttl": "", "static_ttl": true, "precache": true, "replicate": false},
		},
    "replication_conns": ["conn1", "conn2"],
	},
}`
	eMap := map[string]interface{}{
		utils.PartitionsCfg: map[string]interface{}{
			utils.MetaRatingPlans: map[string]interface{}{"limit": 10, "ttl": "", "static_ttl": true, "precache": true},
		},
		utils.ReplicationConnsCfg: []string{"conn1", "conn2"},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		newMap := cgrCfg.cacheCfg.AsMapInterface()
		if !reflect.DeepEqual(newMap[utils.PartitionsCfg].(map[string]interface{})[utils.MetaRatingPlans],
			newMap[utils.PartitionsCfg].(map[string]interface{})[utils.MetaRatingPlans]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.PartitionsCfg].(map[string]interface{})[utils.MetaRatingPlans],
				eMap[utils.PartitionsCfg].(map[string]interface{})[utils.MetaRatingPlans])
		}
		if !reflect.DeepEqual(newMap[utils.ReplicationConnsCfg], eMap[utils.ReplicationConnsCfg]) {
			t.Errorf("Expected %+v, received %+v", eMap[utils.ReplicationConnsCfg], newMap[utils.ReplicationConnsCfg])
		}
	}
}
