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

package engine

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/ltcache"
)

var Cache *ltcache.TransCache

func init() {
	InitCache(nil)
}

// InitCache will instantiate the cache with specific or default configuraiton
func InitCache(cfg config.CacheConfig) {
	if cfg == nil {
		cfg = config.CgrConfig().CacheCfg()
	}
	Cache = ltcache.NewTransCache(cfg.AsTransCacheConfig())
}

func NewCacheS(cfg *config.CGRConfig, dm *DataManager) (c *CacheS) {
	InitCache(cfg.CacheCfg()) // make sure we use the same config as package shared one
	c = &CacheS{cfg: cfg, dm: dm,
		cItems: make(map[string]chan struct{})}
	for k := range cfg.CacheCfg() {
		c.cItems[k] = make(chan struct{})
	}
	return
}

// CacheS deals with cache preload and other cache related tasks
type CacheS struct {
	cfg    *config.CGRConfig
	dm     *DataManager
	cItems map[string]chan struct{} // signal precaching done
}
