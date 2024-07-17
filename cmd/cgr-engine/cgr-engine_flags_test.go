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

package main

import (
	"path"
	"reflect"
	"testing"
	"time"
)

// If any flag changes, this test should fail.
// Do not use constants in this test to ensure these changes are detected.
func TestCgrEngineFlags(t *testing.T) {
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
			flagVar:    cfgPath,
			defaultVal: "/etc/cgrates/",
			want:       "/usr/share/cgrates/conf/samples/tutorial",
		},
		{
			name:       "version",
			flags:      []string{"-version"},
			flagVar:    version,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "printConfig",
			flags:      []string{"-print_config"},
			flagVar:    printConfig,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "pidFile",
			flags:      []string{"-pid", "/run/cgrates/cgrates.pid"},
			flagVar:    pidFile,
			defaultVal: "",
			want:       "/run/cgrates/cgrates.pid",
		},
		{
			name:       "httpPprof",
			flags:      []string{"-http_pprof"},
			flagVar:    httpPprof,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "cpuProfDir",
			flags:      []string{"-cpuprof_dir", "/tmp/profiling"},
			flagVar:    cpuProfDir,
			defaultVal: "",
			want:       "/tmp/profiling",
		},
		{
			name:       "memProfDir",
			flags:      []string{"-memprof_dir", "/tmp/profiling"},
			flagVar:    memProfDir,
			defaultVal: "",
			want:       "/tmp/profiling",
		},
		{
			name:       "memProfInterval",
			flags:      []string{"-memprof_interval", "1s"},
			flagVar:    memProfInterval,
			defaultVal: 5 * time.Second,
			want:       time.Second,
		},
		{
			name:       "memProfNrFiles",
			flags:      []string{"-memprof_nrfiles", "3"},
			flagVar:    memProfNrFiles,
			defaultVal: 1,
			want:       3,
		},
		{
			name:       "scheduledShutdown",
			flags:      []string{"-scheduled_shutdown", "1h"},
			flagVar:    scheduledShutdown,
			defaultVal: time.Duration(0),
			want:       time.Hour,
		},
		{
			name:       "singleCPU",
			flags:      []string{"-singlecpu"},
			flagVar:    singleCPU,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "syslogger",
			flags:      []string{"-logger", "*stdout"},
			flagVar:    syslogger,
			defaultVal: "",
			want:       "*stdout",
		},
		{
			name:       "nodeID",
			flags:      []string{"-node_id", "CGRateS.org"},
			flagVar:    nodeID,
			defaultVal: "",
			want:       "CGRateS.org",
		},
		{
			name:       "logLevel",
			flags:      []string{"-log_level", "7"},
			flagVar:    logLevel,
			defaultVal: -1,
			want:       7,
		},
		{
			name:       "preload",
			flags:      []string{"-preload", "TestPreloadID"},
			flagVar:    preload,
			defaultVal: "",
			want:       "TestPreloadID",
		},
		{
			name:       "setVersions",
			flags:      []string{"-set_versions"},
			flagVar:    setVersions,
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
			if err := cgrEngineFlags.Parse(tt.flags); err != nil {
				t.Errorf("cgrEngineFlags.Parse(%v) returned unexpected error: %v", tt.flags, err)
			}
			flagVal = reflect.ValueOf(tt.flagVar).Elem().Interface()
			if flagVal != tt.want {
				t.Errorf("%s=%v, want %v", tt.name, flagVal, tt.want)
			}
		})
	}
}
