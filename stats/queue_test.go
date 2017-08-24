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
package stats

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"
)

func TestStatsInstancesSort(t *testing.T) {
	sInsts := StatsInstances{
		&StatsInstance{cfg: &engine.StatsConfig{ID: "FIRST", Weight: 30.0}},
		&StatsInstance{cfg: &engine.StatsConfig{ID: "SECOND", Weight: 40.0}},
		&StatsInstance{cfg: &engine.StatsConfig{ID: "THIRD", Weight: 30.0}},
		&StatsInstance{cfg: &engine.StatsConfig{ID: "FOURTH", Weight: 35.0}},
	}
	sInsts.Sort()
	eSInst := StatsInstances{
		&StatsInstance{cfg: &engine.StatsConfig{ID: "SECOND", Weight: 40.0}},
		&StatsInstance{cfg: &engine.StatsConfig{ID: "FOURTH", Weight: 35.0}},
		&StatsInstance{cfg: &engine.StatsConfig{ID: "FIRST", Weight: 30.0}},
		&StatsInstance{cfg: &engine.StatsConfig{ID: "THIRD", Weight: 30.0}},
	}
	if !reflect.DeepEqual(eSInst, sInsts) {
		t.Errorf("expecting: %+v, received: %+v", eSInst, sInsts)
	}
}
