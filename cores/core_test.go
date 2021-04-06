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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewCoreService(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = time.Second
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
	close(stopchan)
	//shut down the service
	rcv.Shutdown()
}

func TestCoreServiceStatus(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = 1
	caps := engine.NewCaps(1, utils.MetaBusy)
	stopChan := make(chan struct{}, 1)

	cores := NewCoreService(cfgDflt, caps, stopChan)
	args := &utils.TenantWithAPIOpts{
		Tenant:  "cgrates.org",
		APIOpts: map[string]interface{}{},
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
	}
	goRoutinesInt := (reply[utils.ActiveGoroutines]).(int)
	if err != nil {
		t.Error(err)
	}
	if goRoutinesInt < 18 {
		t.Errorf("Expected %+v to be larger than 18", reply[utils.GoVersion])
	}
	if !reflect.DeepEqual(expected[utils.GoVersion], reply[utils.GoVersion]) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected[utils.GoVersion]), utils.ToJSON(reply[utils.GoVersion]))
	}
	if !reflect.DeepEqual(expected[utils.RunningSince], reply[utils.RunningSince]) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected[utils.RunningSince]), utils.ToJSON(reply[utils.RunningSince]))
	}
	if !reflect.DeepEqual(expected[utils.VersionName], reply[utils.VersionName]) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected[utils.VersionName]), utils.ToJSON(reply[utils.VersionName]))
	}
	if !reflect.DeepEqual(expected[utils.MemoryUsage], reply[utils.MemoryUsage]) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected[utils.MemoryUsage]), utils.ToJSON(reply[utils.MemoryUsage]))
	}
	if !reflect.DeepEqual(expected[utils.NodeID], reply[utils.NodeID]) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected[utils.NodeID]), utils.ToJSON(reply[utils.NodeID]))
	}
	utils.GitLastLog = `Date: wrong format
`
	if err := cores.Status(args, &reply); err != nil {
		t.Error(err)
	}

	utils.GitLastLog = ""
}
