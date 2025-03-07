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

package trends

import (
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

func TestNewTrendService(t *testing.T) {
	dm := &engine.DataManager{}
	cfg := &config.CGRConfig{}
	filterS := &engine.FilterS{}
	connMgr := &engine.ConnManager{}

	trendService := NewTrendService(dm, cfg, filterS, connMgr)

	if trendService == nil {
		t.Errorf("Expected non-nil TrendS, got nil")
	}

	if trendService.dm != dm {
		t.Errorf("Expected dm to be %v, got %v", dm, trendService.dm)
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
