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

package cores

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewCoreService(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = 1
	stopchan := make(chan struct{}, 1)
	caps := engine.NewCaps(1, utils.MetaBusy)
	sts := engine.NewCapsStats(cfgDflt.CoreSCfg().CapsStatsInterval, caps, stopchan)
	expected := &CoreService{
		cfg:       cfgDflt,
		CapsStats: sts,
	}

	rcv := NewCoreService(cfgDflt, caps, stopchan)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	//shut down the service
	rcv.Shutdown()
}

func TestCoreServiceStatus(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = 1
	caps := engine.NewCaps(1, utils.MetaBusy)
	stopChan := make(chan struct{}, 1)

	cores := NewCoreService(cfgDflt, caps, stopChan)
	args := &utils.TenantWithOpts{
		Tenant: "cgrates.org",
		Opts:   map[string]interface{}{},
	}

	var reply map[string]interface{}
	cfgVrs, err := utils.GetCGRVersion()
	if err != nil {
		t.Error(err)
	}

	expected := map[string]interface{}{
		utils.GoVersion:        runtime.Version(),
		utils.RunningSince:     "TIME_CHANGED",
		utils.VersionName:      cfgVrs,
		utils.ActiveGoroutines: runtime.NumGoroutine(),
		utils.MemoryUsage:      "CHANGED_MEMORY_USAGE",
		utils.NodeID:           cfgDflt.GeneralCfg().NodeID,
	}
	if err := cores.Status(args, &reply); err != nil {
		t.Error(err)
	} else {
		reply[utils.RunningSince] = "TIME_CHANGED"
		reply[utils.MemoryUsage] = "CHANGED_MEMORY_USAGE"
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}

	utils.GitLastLog = `Date: wrong format
`
	if err := cores.Status(args, &reply); err != nil {
		t.Error(err)
	}

	utils.GitLastLog = ""
}
