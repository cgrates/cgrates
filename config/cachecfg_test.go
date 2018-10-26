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
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestAsTransCacheConfig(t *testing.T) {
	a := &CacheCfg{
		"test": &CacheParamCfg{
			Limit:     50,
			TTL:       time.Duration(60 * time.Second),
			StaticTTL: true,
			Precache:  true,
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
	var cachecfg, expected CacheCfg
	if err := cachecfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cachecfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, cachecfg)
	}
	if err := cachecfg.loadFromJsonCfg(new(CacheJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cachecfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, cachecfg)
	}
	cfgJSONStr := `{
"cache":{
	"destinations": {"limit": -1, "ttl": "", "static_ttl": false, "precache": false},			
	"reverse_destinations": {"limit": -1, "ttl": "", "static_ttl": false, "precache": false},	
	"rating_plans": {"limit": -1, "ttl": "", "static_ttl": false, "precache": false},
	}		
}`
	expected = CacheCfg{
		"destinations":         &CacheParamCfg{Limit: -1, TTL: time.Duration(0), StaticTTL: false, Precache: false},
		"reverse_destinations": &CacheParamCfg{Limit: -1, TTL: time.Duration(0), StaticTTL: false, Precache: false},
		"rating_plans":         &CacheParamCfg{Limit: -1, TTL: time.Duration(0), StaticTTL: false, Precache: false},
	}
	cachecfg = CacheCfg{}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnCacheCfg, err := jsnCfg.CacheJsonCfg(); err != nil {
		t.Error(err)
	} else if err = cachecfg.loadFromJsonCfg(jsnCacheCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cachecfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, cachecfg)
	}
}

func TestCacheParamCfgloadFromJsonCfg(t *testing.T) {
	var fscocfg, expected CacheParamCfg
	if err := fscocfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscocfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, fscocfg)
	}
	if err := fscocfg.loadFromJsonCfg(new(CacheParamJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscocfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, fscocfg)
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
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(fscocfg))
	}
}
