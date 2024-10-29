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
	expected := &CoreS{
		cfg:       cfgDflt,
		CapsStats: sts,
		caps:      caps,
	}
	rcv := NewCoreService(cfgDflt, caps, nil, stopchan, nil, nil)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
	//shut down the service
	rcv.shtDw = func() {}
	rcv.Shutdown()
	rcv.ShutdownEngine()
}

func TestCoreServiceStatus(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CoreSCfg().CapsStatsInterval = 1
	cores := NewCoreService(cfg, engine.NewCaps(1, utils.MetaBusy), nil, nil, nil, func() {})

	var reply map[string]any
	cfgVrs, err := utils.GetCGRVersion()
	if err != nil {
		t.Error(err)
	}

	expected := map[string]any{
		utils.GoVersion:        runtime.Version(),
		utils.RunningSince:     "TIME_CHANGED",
		utils.VersionName:      cfgVrs,
		utils.ActiveGoroutines: runtime.NumGoroutine(),
		utils.MemoryUsage:      "CHANGED_MEMORY_USAGE",
		utils.NodeID:           cfg.GeneralCfg().NodeID,
	}
	if err := cores.V1Status(nil, nil, &reply); err != nil {
		t.Error(err)
	} else {
		reply[utils.RunningSince] = "TIME_CHANGED"
		reply[utils.MemoryUsage] = "CHANGED_MEMORY_USAGE"
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
	utils.GitCommitDate = "wrong format"
	utils.GitCommitHash = "73014DAA0C1D7EDCB532D5FE600B8A20D588CDF8"
	if err := cores.V1Status(nil, nil, &reply); err != nil {
		t.Error(err)
	}

	utils.GitCommitDate = ""
	utils.GitCommitHash = ""
}
