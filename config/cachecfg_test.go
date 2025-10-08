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
				TTL:       60 * time.Second,
				StaticTTL: true,
				Precache:  true,
			},
		},
	}
	expected := map[string]*ltcache.CacheConfig{
		"test": {
			MaxItems:  50,
			TTL:       60 * time.Second,
			StaticTTL: true,
		},
	}
	reply := a.AsTransCacheConfig()
	if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %+v, received: %+v", expected, utils.ToJSON(reply))
	}
}

func TestReplicationConnsLoadFromJsonCfg(t *testing.T) {
	jsonCfg := &CacheJsonCfg{
		Replication_conns: &[]string{utils.MetaInternal},
	}
	expErrMessage := "replication connection ID needs to be different than *internal"
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.cacheCfg.loadFromJSONCfg(jsonCfg); err == nil || err.Error() != expErrMessage {
		t.Errorf("Expected %+v , recevied %+v", expErrMessage, err)
	}
}

func TestCacheParamCfgloadFromJsonCfg1(t *testing.T) {
	json := &CacheParamJsonCfg{
		Limit:      utils.IntPointer(5),
		Ttl:        utils.StringPointer("1s"),
		Static_ttl: utils.BoolPointer(true),
		Precache:   utils.BoolPointer(true),
	}
	expected := &CacheParamCfg{
		Limit:     5,
		TTL:       time.Second,
		StaticTTL: true,
		Precache:  true,
	}
	rcv := new(CacheParamCfg)
	if err := rcv.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	} else if err := rcv.loadFromJSONCfg(json); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestCacheParamCfgloadFromJsonCfg2(t *testing.T) {
	jsonCfg := &CacheJsonCfg{
		Partitions: map[string]*CacheParamJsonCfg{
			utils.MetaAttributes: {
				Ttl: utils.StringPointer("1ss"),
			},
		},
	}
	expErrMessage := "time: unknown unit \"ss\" in duration \"1ss\""
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.cacheCfg.loadFromJSONCfg(jsonCfg); err == nil || err.Error() != expErrMessage {
		t.Errorf("Expected %+v \n, recevied %+v", expErrMessage, err)
	}
}

func TestCacheCfgClone(t *testing.T) {
	cs := &CacheCfg{
		Partitions:       map[string]*CacheParamCfg{},
		ReplicationConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
	}
	rcv := cs.Clone()
	if !reflect.DeepEqual(cs, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cs), utils.ToJSON(rcv))
	}
	if rcv.ReplicationConns[0] = ""; cs.ReplicationConns[0] != utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffCacheParamsJsonCfg(t *testing.T) {
	var d map[string]*CacheParamJsonCfg
	v2 := map[string]*CacheParamCfg{
		"CACHE_2": {
			Limit:     3,
			TTL:       5 * time.Minute,
			StaticTTL: true,
			Precache:  false,
			Remote:    true,
			Replicate: true,
		},
	}

	expected := map[string]*CacheParamJsonCfg{
		"CACHE_2": {
			Limit:      utils.IntPointer(3),
			Ttl:        utils.StringPointer("5m0s"),
			Static_ttl: utils.BoolPointer(true),
			Precache:   utils.BoolPointer(false),
			Remote:     utils.BoolPointer(true),
			Replicate:  utils.BoolPointer(true),
		},
	}

	rcv := diffCacheParamsJsonCfg(d, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := map[string]*CacheParamCfg{
		"CACHE_1": {
			Limit:     2,
			TTL:       2 * time.Minute,
			StaticTTL: false,
			Precache:  true,
			Replicate: false,
		},
	}
	expected2 := map[string]*CacheParamJsonCfg{
		"CACHE_1": {
			Limit:      utils.IntPointer(2),
			Ttl:        utils.StringPointer("2m0s"),
			Static_ttl: utils.BoolPointer(false),
			Precache:   utils.BoolPointer(true),
			Remote:     utils.BoolPointer(false),
			Replicate:  utils.BoolPointer(false),
		},
	}
	rcv = diffCacheParamsJsonCfg(d, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestDiffCacheJsonCfg(t *testing.T) {
	var d *CacheJsonCfg

	v1 := &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			"CACHE_2": {
				Limit:     2,
				TTL:       2 * time.Minute,
				StaticTTL: false,
				Precache:  true,
				Remote:    false,
				Replicate: false,
			},
		},
		ReplicationConns: []string{},
	}

	v2 := &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			"CACHE_2": {
				Limit:     3,
				TTL:       5 * time.Minute,
				StaticTTL: true,
				Precache:  false,
				Remote:    true,
				Replicate: true,
			},
		},
		ReplicationConns: []string{"*repl_conn"},
	}

	expected := &CacheJsonCfg{
		Partitions: map[string]*CacheParamJsonCfg{
			"CACHE_2": {
				Limit:      utils.IntPointer(3),
				Ttl:        utils.StringPointer("5m0s"),
				Static_ttl: utils.BoolPointer(true),
				Precache:   utils.BoolPointer(false),
				Remote:     utils.BoolPointer(true),
				Replicate:  utils.BoolPointer(true),
			},
		},
		Replication_conns: &[]string{"*repl_conn"},
	}

	rcv := diffCacheJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &CacheJsonCfg{
		Partitions: map[string]*CacheParamJsonCfg{
			"CACHE_2": {
				Limit:      utils.IntPointer(2),
				Ttl:        utils.StringPointer("2m0s"),
				Static_ttl: utils.BoolPointer(false),
				Precache:   utils.BoolPointer(true),
				Remote:     utils.BoolPointer(false),
				Replicate:  utils.BoolPointer(false),
			},
		},
		Replication_conns: nil,
	}

	rcv = diffCacheJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestCacheCloneSection(t *testing.T) {
	cacheCfg := &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			"CACHE_2": {
				Limit:     2,
				TTL:       2 * time.Minute,
				StaticTTL: false,
				Precache:  true,
				Replicate: false,
			},
		},
		ReplicationConns: []string{},
	}

	exp := &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			"CACHE_2": {
				Limit:     2,
				TTL:       2 * time.Minute,
				StaticTTL: false,
				Precache:  true,
				Replicate: false,
			},
		},
		ReplicationConns: []string{},
	}
	rcv := cacheCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestCacheCfgloadFromJSONCfg(t *testing.T) {

	cCfg := &CacheCfg{}

	jsnCfg := &CacheJsonCfg{
		Remote_conns: &[]string{"remote", "test"},
	}

	exp := &CacheCfg{
		RemoteConns: []string{"remote", "test"},
	}

	if err := cCfg.loadFromJSONCfg(jsnCfg); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(utils.ToJSON(cCfg), utils.ToJSON(exp)) {
		t.Errorf("Expected <%v>, Received <%v>", utils.ToJSON(exp), utils.ToJSON(cCfg))

	}
}
func TestDiffCacheJsonCfgRemoteConn(t *testing.T) {
	var d *CacheJsonCfg

	v1 := &CacheCfg{}

	v2 := &CacheCfg{

		RemoteConns: []string{"test"},
	}

	expected := &CacheJsonCfg{
		Partitions:   make(map[string]*CacheParamJsonCfg),
		Remote_conns: &[]string{"test"},
	}

	if rcv := diffCacheJsonCfg(d, v1, v2); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

}
