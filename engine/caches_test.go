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
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestV1LoadCache(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	dm.cacheCfg[utils.CacheThresholds].Precache = true
	thd := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH_1",
		Hits:   0,
	}
	if err := dm.SetThreshold(thd); err != nil {
		t.Error(err)
	}
	args := utils.AttrReloadCacheWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{},
		AttrReloadCache: utils.AttrReloadCache{
			ArgsCache: utils.ArgsCache{
				ThresholdIDs: []string{"THD1"},
			},
		},
	}
	cacheS := NewCacheS(cfg, dm)
	var reply string
	if err := cacheS.V1LoadCache(args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expectd ok ,received %v", reply)
	}
}

func TestCacheV1ReloadCache(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		InitCache(cfg.CacheCfg())
	}()
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	chS := NewCacheS(cfg, dm)
	attrs := utils.AttrReloadCacheWithArgDispatcher{
		AttrReloadCache: utils.AttrReloadCache{
			ArgsCache: utils.ArgsCache{
				FilterIDs: []string{"DSP_FLTR"},
			},
		},
	}
	fltr := &Filter{
		ID:     "DSP_FLTR",
		Tenant: "cgrates.org",
		Rules: []*FilterRule{
			{Type: utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
	}
	if err := dm.SetFilter(fltr); err != nil {
		t.Error(err)
	}
	Cache = db.db
	var reply string
	if err := chS.V1ReloadCache(attrs, &reply); err != nil {
		t.Error(err)
	}

}
