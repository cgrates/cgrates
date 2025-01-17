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
func TestFlags(t *testing.T) {
	flags := newFlags()
	tests := []struct {
		name       string
		flags      []string
		flagVar    any
		defaultVal any
		want       any
	}{
		{
			name:       "config_path",
			flags:      []string{"-config_path", path.Join("/usr", "share", "cgrates", "conf", "samples", "tutorial")},
			flagVar:    &flags.config.path,
			defaultVal: "/etc/cgrates/",
			want:       "/usr/share/cgrates/conf/samples/tutorial",
		},
		{
			name:       "version",
			flags:      []string{"-version"},
			flagVar:    &flags.config.version,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "pid",
			flags:      []string{"-pid", "/run/cgrates/cgrates.pid"},
			flagVar:    &flags.process.pidFile,
			defaultVal: "",
			want:       "/run/cgrates/cgrates.pid",
		},
		{
			name:       "cpuprof_dir",
			flags:      []string{"-cpuprof_dir", "/tmp/profiling"},
			flagVar:    &flags.profiling.cpu.dir,
			defaultVal: "",
			want:       "/tmp/profiling",
		},
		{
			name:       "memprof_dir",
			flags:      []string{"-memprof_dir", "/tmp/profiling"},
			flagVar:    &flags.profiling.mem.dir,
			defaultVal: "",
			want:       "/tmp/profiling",
		},
		{
			name:       "memprof_interval",
			flags:      []string{"-memprof_interval", "1s"},
			flagVar:    &flags.profiling.mem.interval,
			defaultVal: 15 * time.Second,
			want:       time.Second,
		},
		{
			name:       "memprof_maxfiles",
			flags:      []string{"-memprof_maxfiles", "3"},
			flagVar:    &flags.profiling.mem.maxFiles,
			defaultVal: 1,
			want:       3,
		},
		{
			name:       "memprof_timestamp",
			flags:      []string{"-memprof_timestamp"},
			flagVar:    &flags.profiling.mem.useTS,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "scheduled_shutdown",
			flags:      []string{"-scheduled_shutdown", "1h"},
			flagVar:    &flags.process.scheduledShutdown,
			defaultVal: time.Duration(0),
			want:       time.Hour,
		},
		{
			name:       "single_cpu",
			flags:      []string{"-single_cpu"},
			flagVar:    &flags.process.singleCPU,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "logger",
			flags:      []string{"-logger", "*stdout"},
			flagVar:    &flags.logger.typ,
			defaultVal: "",
			want:       "*stdout",
		},
		{
			name:       "node_id",
			flags:      []string{"-node_id", "CGRateS.org"},
			flagVar:    &flags.logger.nodeID,
			defaultVal: "",
			want:       "CGRateS.org",
		},
		{
			name:       "log_level",
			flags:      []string{"-log_level", "7"},
			flagVar:    &flags.logger.level,
			defaultVal: -1,
			want:       7,
		},
		{
			name:       "check_config",
			flags:      []string{"-check_config", "true"},
			flagVar:    &flags.config.check,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "set_versions",
			flags:      []string{"-set_versions"},
			flagVar:    &flags.data.setVersions,
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
			if err := flags.Parse(tt.flags); err != nil {
				t.Errorf("flags.Parse(%v) returned unexpected error: %v", tt.flags, err)
			}
			flagVal = reflect.ValueOf(tt.flagVar).Elem().Interface()
			if flagVal != tt.want {
				t.Errorf("%s=%v, want %v", tt.name, flagVal, tt.want)
			}
		})
	}
}

func TestPreloadFlag(t *testing.T) {
	flg := newFlags()
	if err := flg.Parse([]string{"-preload", "loader1,loader2"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"loader1", "loader2"}
	if !reflect.DeepEqual(flg.data.preloadIDs, want) {
		t.Errorf("preload IDs = %v, want %v", flg.data.preloadIDs, want)
	}
}
