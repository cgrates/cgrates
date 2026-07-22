// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package trends

import (
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewTrendService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewGuardianLocker(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cache := engine.NewCacheS(cfg, nil, nil, nil, locker)
	filterS := &engine.FilterS{}
	connMgr := &engine.ConnManager{}

	trendService := NewTrendService(dm, cache, cfg, filterS, connMgr)

	if trendService == nil {
		t.Errorf("Expected non-nil TrendS, got nil")
	}

	if trendService.dm != dm {
		t.Errorf("Expected dm to be %v, got %v", dm, trendService.dm)
	}

	if trendService.cache != cache {
		t.Errorf("Expected cache to be %v, got %v", cache, trendService.cache)
	}

	if trendService.cfg != cfg {
		t.Errorf("Expected cfg to be %v, got %v", cfg, trendService.cfg)
	}

	if trendService.fltrS != filterS {
		t.Errorf("Expected filterS to be %v, got %v", filterS, trendService.fltrS)
	}

	if trendService.connMgr != connMgr {
		t.Errorf("Expected connMgr to be %v, got %v", connMgr, trendService.connMgr)
	}

	if trendService.crnTQs == nil {
		t.Errorf("Expected crnTQs to be non-nil, got nil")
	}

	if trendService.crnTQsMux == nil {
		t.Errorf("Expected crnTQsMux to be non-nil, got nil")
	}

	if trendService.loopStopped == nil {
		t.Errorf("Expected loopStopped to be non-nil, got nil")
	}
}
