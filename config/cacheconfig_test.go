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
	a := &CacheConfig{
		"test": &CacheParamConfig{
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

func TestCacheConfigloadFromJsonCfg(t *testing.T) {
	var cachecfg, expected CacheConfig
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
	expected = CacheConfig{
		"destinations":         &CacheParamConfig{Limit: -1, TTL: time.Duration(0), StaticTTL: false, Precache: false},
		"reverse_destinations": &CacheParamConfig{Limit: -1, TTL: time.Duration(0), StaticTTL: false, Precache: false},
		"rating_plans":         &CacheParamConfig{Limit: -1, TTL: time.Duration(0), StaticTTL: false, Precache: false},
	}
	cachecfg = CacheConfig{}
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
