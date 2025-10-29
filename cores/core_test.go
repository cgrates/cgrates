/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package cores

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
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

func TestWriteHeapProfile(t *testing.T) {
	tmpFilePath := "test_heap_profile.pprof"
	defer os.Remove(tmpFilePath)
	t.Run("Success", func(t *testing.T) {
		err := writeHeapProfile(tmpFilePath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, err := os.Stat(tmpFilePath); os.IsNotExist(err) {
			t.Fatal("expected file to be created")
		}
	})

	t.Run("FileCreationError", func(t *testing.T) {
		invalidPath := "/invalid/path/to/file.pprof"
		err := writeHeapProfile(invalidPath)
		if err == nil {
			t.Fatal("expected an error but got none")
		}
		if !strings.Contains(err.Error(), "could not create memory profile") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestNewMemProfNameFunc(t *testing.T) {
	t.Run("NoTimestamp", func(t *testing.T) {
		gen := newMemProfNameFunc(0, false)
		for i := 1; i <= 3; i++ {
			expected := fmt.Sprintf("mem_%d.prof", i)
			result := gen()
			if result != expected {
				t.Errorf("expected %s, got %s", expected, result)
			}
		}
	})

	t.Run("Timestamp1SecondOrMore", func(t *testing.T) {
		gen := newMemProfNameFunc(1*time.Second, true)
		now := time.Now()
		result := gen()
		expected := fmt.Sprintf("mem_%s.prof", now.Format("20060102150405"))
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})
}
