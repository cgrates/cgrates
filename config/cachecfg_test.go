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
	var cachecfg, expected *CacheCfg
	cachecfg = new(CacheCfg)
	expected = new(CacheCfg)
	if err := cachecfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cachecfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, cachecfg)
	}
	if err := cachecfg.loadFromJsonCfg(new(CacheJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cachecfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, cachecfg)
	}
	cfgJSONStr := `{
"caches":{
	"partitions": {
		"*destinations": {"limit": -1, "ttl": "", "static_ttl": false, "precache": false},			
		"*reverse_destinations": {"limit": -1, "ttl": "", "static_ttl": false, "precache": false},	
		"*rating_plans": {"limit": -1, "ttl": "", "static_ttl": false, "precache": false},
		},
	},
}`
	expected = &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			"*destinations":         {Limit: -1, TTL: time.Duration(0), StaticTTL: false, Precache: false},
			"*reverse_destinations": {Limit: -1, TTL: time.Duration(0), StaticTTL: false, Precache: false},
			"*rating_plans":         {Limit: -1, TTL: time.Duration(0), StaticTTL: false, Precache: false},
		},
	}
	cachecfg = new(CacheCfg)
	cachecfg.Partitions = make(map[string]*CacheParamCfg)
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnCacheCfg, err := jsnCfg.CacheJsonCfg(); err != nil {
		t.Error(err)
	} else if err = cachecfg.loadFromJsonCfg(jsnCacheCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cachecfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, cachecfg)
	}
}

func TestCacheParamCfgloadFromJsonCfg(t *testing.T) {
	var fscocfg, expected CacheParamCfg
	if err := fscocfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscocfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, fscocfg)
	}
	if err := fscocfg.loadFromJsonCfg(new(CacheParamJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscocfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, fscocfg)
	}
	json := &CacheParamJsonCfg{
		Limit:      utils.IntPointer(5),
		Ttl:        utils.StringPointer("1s"),
		Static_ttl: utils.BoolPointer(true),
		Precache:   utils.BoolPointer(true),
	}
	expected = CacheParamCfg{
		Limit:     5,
		TTL:       time.Duration(time.Second),
		StaticTTL: true,
		Precache:  true,
	}
	if err = fscocfg.loadFromJsonCfg(json); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, fscocfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(fscocfg))
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
