// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
			defaultVal: 15 * time.Second,
			want:       time.Second,
		},
		{
			name:       "memProfMaxFiles",
			flags:      []string{"-memprof_maxfiles", "3"},
			flagVar:    memProfMaxFiles,
			defaultVal: 1,
			want:       3,
		},
		{
			name:       "memProfTimestamp",
			flags:      []string{"-memprof_timestamp"},
			flagVar:    memProfTimestamp,
			defaultVal: false,
			want:       true,
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
