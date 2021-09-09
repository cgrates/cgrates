/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"testing"
	"time"
)

// if the flag change this should fail
// do not use constants in this test
func TestCgrEngineFlags(t *testing.T) {
	cgrEngineFlags := NewCGREngineFlags()
	if err := cgrEngineFlags.Parse([]string{"-config_path", path.Join("/conf", "samples", "tutorial")}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.CfgPath != "/conf/samples/tutorial" {
		t.Errorf("Expected /conf/samples/tutorial, received %+v", *cgrEngineFlags.CfgPath)
	}

	if err := cgrEngineFlags.Parse([]string{"-version", "true"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.Version != true {
		t.Errorf("Expected true, received %+v", *cgrEngineFlags.Version)
	}

	if err := cgrEngineFlags.Parse([]string{"-check_config", "true"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.CheckConfig != true {
		t.Errorf("Expected true, received %+v", *cgrEngineFlags.CheckConfig)
	}

	if err := cgrEngineFlags.Parse([]string{"-pid", "usr/share/cgrates/cgrates.json"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.PidFile != "usr/share/cgrates/cgrates.json" {
		t.Errorf("Expected usr/share/cgrates/cgrates.json, received %+v", *cgrEngineFlags.PidFile)
	}

	if err := cgrEngineFlags.Parse([]string{"-httprof_path", "http://example.com/"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.HttpPrfPath != "http://example.com/" {
		t.Errorf("Expected http://example.com/, received %+v", *cgrEngineFlags.HttpPrfPath)
	}

	if err := cgrEngineFlags.Parse([]string{"-cpuprof_dir", "1"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.CpuPrfDir != "1" {
		t.Errorf("Expected 1, received %+v", *cgrEngineFlags.CpuPrfDir)
	}

	if err := cgrEngineFlags.Parse([]string{"-memprof_dir", "true"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.MemPrfDir != "true" {
		t.Errorf("Expected true received %+v", *cgrEngineFlags.MemPrfDir)
	}

	if err := cgrEngineFlags.Parse([]string{"-memprof_interval", "1s"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.MemPrfInterval != time.Second {
		t.Errorf("Expected 1s, received %+v", *cgrEngineFlags.MemPrfInterval)
	}

	if err := cgrEngineFlags.Parse([]string{"-memprof_nrfiles", "3"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.MemPrfNoF != 3 {
		t.Errorf("Expected 3, received %+v", *cgrEngineFlags.MemPrfNoF)
	}

	if err := cgrEngineFlags.Parse([]string{"-scheduled_shutdown", "1h"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.ScheduledShutDown != "1h" {
		t.Errorf("Expected 1h, received %+v", *cgrEngineFlags.ScheduledShutDown)
	}

	if err := cgrEngineFlags.Parse([]string{"-singlecpu"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.Singlecpu != true {
		t.Errorf("Expected true, received %+v", *cgrEngineFlags.Singlecpu)
	}

	if err := cgrEngineFlags.Parse([]string{"-logger", "*cgrEngineFlags.stdout"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.SysLogger != "*cgrEngineFlags.stdout" {
		t.Errorf("Expected *cgrEngineFlags.stdout, received %+v", *cgrEngineFlags.SysLogger)
	}

	if err := cgrEngineFlags.Parse([]string{"-node_id", "CGRates.org"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.NodeID != "CGRates.org" {
		t.Errorf("Expected CGRates.org, received %+v", *cgrEngineFlags.NodeID)
	}

	if err := cgrEngineFlags.Parse([]string{"-log_level", "7"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.LogLevel != 7 {
		t.Errorf("Expected 7, received %+v", *cgrEngineFlags.LogLevel)
	}

	if err := cgrEngineFlags.Parse([]string{"-preload", "TestPreloadID"}); err != nil {
		t.Fatal(err)
	} else if *cgrEngineFlags.Preload != "TestPreloadID" {
		t.Errorf("Expected 7, received %+v", *cgrEngineFlags.Preload)
	}
}
