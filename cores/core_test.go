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
	"errors"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
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
	shdWg := new(sync.WaitGroup)
	shdChan := utils.NewSyncedChan()
	expected := &CoreService{
		shdWg:     shdWg,
		shdChan:   shdChan,
		cfg:       cfgDflt,
		CapsStats: sts,
		caps:      caps,
	}

	rcv := NewCoreService(cfgDflt, caps, nil, stopchan, shdWg, shdChan)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
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

	cores := NewCoreService(cfgDflt, caps, nil, stopChan, nil, nil)
	args := &utils.TenantWithAPIOpts{
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}

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
		utils.NodeID:           cfgDflt.GeneralCfg().NodeID,
	}
	if err := cores.V1Status(context.Background(), args, &reply); err != nil {
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
	if err := cores.V1Status(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	utils.GitCommitDate = ""
	utils.GitCommitHash = ""
}

func TestV1Panic(t *testing.T) {
	coreService := &CoreService{}
	expectedMessage := "test panic message"
	args := &utils.PanicMessageArgs{Message: expectedMessage}
	defer func() {
		if r := recover(); r != nil {
			if r != expectedMessage {
				t.Errorf("Expected panic message %v, got %v", expectedMessage, r)
			}
		} else {
			t.Error("Expected panic but did not get one")
		}
	}()
	err := coreService.V1Panic(nil, args, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestV1StopMemoryProfiling(t *testing.T) {
	coreService := &CoreService{}
	var reply string

	t.Run("Success", func(t *testing.T) {
		err := coreService.V1StopMemoryProfiling(nil, utils.TenantWithAPIOpts{}, &reply)
		if err == nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if reply == utils.OK {
			t.Errorf("Expected reply %s, got %s", utils.OK, reply)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		expectedError := errors.New("stop memory profiling error")
		err := coreService.V1StopMemoryProfiling(nil, utils.TenantWithAPIOpts{}, &reply)
		if err == nil {
			t.Error("Expected error but got nil")
		} else if err == expectedError {
			t.Errorf("Expected error %v, got %v", expectedError, err)
		}
	})
}

func TestV1StartCPUProfiling(t *testing.T) {
	coreService := &CoreService{}
	tests := []struct {
		name          string
		args          *utils.DirectoryArgs
		expectedReply string
		expectedError error
	}{
		{
			name: "Valid Directory Path",
			args: &utils.DirectoryArgs{
				DirPath: "/valid/path",
			},
			expectedReply: utils.OK,
			expectedError: nil,
		},
		{
			name: "Invalid Directory Path",
			args: &utils.DirectoryArgs{
				DirPath: "/invalid/path",
			},
			expectedReply: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reply string
			err := coreService.V1StartCPUProfiling(nil, tt.args, &reply)
			if err == nil && tt.expectedError != nil {
				t.Errorf("Expected error %v, but got nil", tt.expectedError)
			}
		})
	}
}

func TestV1StopCPUProfiling(t *testing.T) {
	coreService := &CoreService{}
	tests := []struct {
		name          string
		initialStatus bool
		expectedReply string
		expectedError error
	}{
		{
			name:          "Successful StopCPUProfiling",
			initialStatus: true,
			expectedReply: utils.OK,
			expectedError: nil,
		},
		{
			name:          "No CPUProfiling to Stop",
			initialStatus: false,
			expectedReply: utils.OK,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reply string
			err := coreService.V1StopCPUProfiling(nil, nil, &reply)
			if err == nil && err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}
			if err == nil && tt.expectedError != nil {
				t.Errorf("Expected error %v, but got nil", tt.expectedError)
			}
			if reply == tt.expectedReply {
				t.Errorf("Expected reply %s, got %s", tt.expectedReply, reply)
			}
		})
	}
}
