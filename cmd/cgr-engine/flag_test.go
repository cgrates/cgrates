// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
			name:       "cpuProfDir",
			flags:      []string{"-cpuProfDir", "/tmp/profiling"},
			flagVar:    &flags.profiling.cpu.dir,
			defaultVal: "",
			want:       "/tmp/profiling",
		},
		{
			name:       "memProfDir",
			flags:      []string{"-memProfDir", "/tmp/profiling"},
			flagVar:    &flags.profiling.mem.dir,
			defaultVal: "",
			want:       "/tmp/profiling",
		},
		{
			name:       "memProfInterval",
			flags:      []string{"-memProfInterval", "1s"},
			flagVar:    &flags.profiling.mem.interval,
			defaultVal: 15 * time.Second,
			want:       time.Second,
		},
		{
			name:       "memProfMaxFiles",
			flags:      []string{"-memProfMaxFiles", "3"},
			flagVar:    &flags.profiling.mem.maxFiles,
			defaultVal: 1,
			want:       3,
		},
		{
			name:       "memProfTimestamp",
			flags:      []string{"-memProfTimestamp"},
			flagVar:    &flags.profiling.mem.useTS,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "scheduledShutdown",
			flags:      []string{"-scheduledShutdown", "1h"},
			flagVar:    &flags.process.scheduledShutdown,
			defaultVal: time.Duration(0),
			want:       time.Hour,
		},
		{
			name:       "singleCPU",
			flags:      []string{"-singleCPU"},
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
			name:       "nodeID",
			flags:      []string{"-nodeID", "CGRateS.org"},
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
			name:       "checkConfig",
			flags:      []string{"-checkConfig", "true"},
			flagVar:    &flags.config.check,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "printConfig",
			flags:      []string{"-printConfig", "true"},
			flagVar:    &flags.config.print,
			defaultVal: false,
			want:       true,
		},
		{
			name:       "setVersions",
			flags:      []string{"-setVersions"},
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
