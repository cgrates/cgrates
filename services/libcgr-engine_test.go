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

package services

import (
	"path"
	"reflect"
	"testing"
	"time"
)

// If any flag changes, this test should fail.
// Do not use constants in this test to ensure these changes are detected.
func TestCgrEngineFlags(t *testing.T) {
	ngFlags := NewCGREngineFlags()
	tests := []struct {
		name       string
		flags      []string
		flagVar    any
		defaultVal any
		want       any
	}{
		{
			name:       "cfgPath",
			flags:      []string{"-config_path", path.Join("/usr", "share", "cgrates", "conf", "samples", "tutorial")},
			flagVar:    ngFlags.CfgPath,
			defaultVal: "/etc/cgrates/",
			want:       "/usr/share/cgrates/conf/samples/tutorial",
		},
		{
			name:       "version",
			flags:      []string{"-version"},
			flagVar:    ngFlags.Version,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "pidFile",
			flags:      []string{"-pid", "/run/cgrates/cgrates.pid"},
			flagVar:    ngFlags.PidFile,
			defaultVal: "",
			want:       "/run/cgrates/cgrates.pid",
		},
		{
			name:       "cpuProfDir",
			flags:      []string{"-cpuprof_dir", "/tmp/profiling"},
			flagVar:    ngFlags.CpuPrfDir,
			defaultVal: "",
			want:       "/tmp/profiling",
		},
		{
			name:       "memProfDir",
			flags:      []string{"-memprof_dir", "/tmp/profiling"},
			flagVar:    ngFlags.MemPrfDir,
			defaultVal: "",
			want:       "/tmp/profiling",
		},
		{
			name:       "memProfInterval",
			flags:      []string{"-memprof_interval", "1s"},
			flagVar:    ngFlags.MemPrfInterval,
			defaultVal: 15 * time.Second,
			want:       time.Second,
		},
		{
			name:       "memProfMaxFiles",
			flags:      []string{"-memprof_maxfiles", "3"},
			flagVar:    ngFlags.MemPrfMaxF,
			defaultVal: 1,
			want:       3,
		},
		{
			name:       "memProfTimestamp",
			flags:      []string{"-memprof_timestamp"},
			flagVar:    ngFlags.MemPrfTS,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "scheduledShutdown",
			flags:      []string{"-scheduled_shutdown", "1h"},
			flagVar:    ngFlags.ScheduledShutdown,
			defaultVal: "",
			want:       "1h",
		},
		{
			name:       "singleCPU",
			flags:      []string{"-single_cpu"},
			flagVar:    ngFlags.SingleCPU,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "syslogger",
			flags:      []string{"-logger", "*stdout"},
			flagVar:    ngFlags.Logger,
			defaultVal: "",
			want:       "*stdout",
		},
		{
			name:       "nodeID",
			flags:      []string{"-node_id", "CGRateS.org"},
			flagVar:    ngFlags.NodeID,
			defaultVal: "",
			want:       "CGRateS.org",
		},
		{
			name:       "logLevel",
			flags:      []string{"-log_level", "7"},
			flagVar:    ngFlags.LogLevel,
			defaultVal: -1,
			want:       7,
		},
		{
			name:       "preload",
			flags:      []string{"-preload", "TestPreloadID"},
			flagVar:    ngFlags.Preload,
			defaultVal: "",
			want:       "TestPreloadID",
		},
		{
			name:       "check_config",
			flags:      []string{"-check_config", "true"},
			flagVar:    ngFlags.CheckConfig,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "setVersions",
			flags:      []string{"-set_versions"},
			flagVar:    ngFlags.SetVersions,
			defaultVal: false,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagVal := reflect.ValueOf(tt.flagVar).Elem().Interface()
			if flagVal != tt.defaultVal {
				t.Errorf("%s=%v, want default value %v", tt.name, flagVal, tt.defaultVal)
			}
			if err := ngFlags.Parse(tt.flags); err != nil {
				t.Errorf("cgrEngineFlags.Parse(%v) returned unexpected error: %v", tt.flags, err)
			}
			flagVal = reflect.ValueOf(tt.flagVar).Elem().Interface()
			if flagVal != tt.want {
				t.Errorf("%s=%v, want %v", tt.name, flagVal, tt.want)
			}
		})
	}
}
