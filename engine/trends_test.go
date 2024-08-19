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
)

func TestTrendProfileTenantID(t *testing.T) {
	profile := &TrendProfile{
		Tenant: "cgrates.org",
		ID:     "1",
	}
	result := profile.TenantID()
	expected := "cgrates.org:1"
	if result != expected {
		t.Errorf("TenantID() = %v; want %v", result, expected)
	}
}

func TestTrendTenantID(t *testing.T) {
	trend := &Trend{
		Tenant: "cgrates.org",
		ID:     "1",
	}
	result := trend.TenantID()
	expected := "cgrates.org:1"
	if result != expected {
		t.Errorf("TenantID() = %v; want %v", result, expected)
	}
}

func TestNewTrendService(t *testing.T) {
	dm := &DataManager{}
	cgrcfg := &config.CGRConfig{}
	filterS := &FilterS{}
	result := NewTrendService(dm, cgrcfg, filterS)
	if result.dm != dm {
		t.Errorf("Expected dm to be %v, got %v", dm, result.dm)
	}
	if result.cgrcfg != cgrcfg {
		t.Errorf("Expected cgrcfg to be %v, got %v", cgrcfg, result.cgrcfg)
	}
	if result.filterS != filterS {
		t.Errorf("Expected filterS to be %v, got %v", filterS, result.filterS)
	}
}
