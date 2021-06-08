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

package main

import (
	"path"
	"testing"
	"time"
)

// if the flag change this should fail
// do not use constants in this test
func TestCgrEngineFlags(t *testing.T) {
	if err := cgrEngineFlags.Parse([]string{"-config_path", path.Join("/conf", "samples", "tutorial")}); err != nil {
		t.Fatal(err)
	} else if *cfgPath != "/conf/samples/tutorial" {
		t.Errorf("Expected /conf/samples/tutorial, received %+v", *cfgPath)
	}

	if err := cgrEngineFlags.Parse([]string{"-version", "true"}); err != nil {
		t.Fatal(err)
	} else if *version != true {
		t.Errorf("Expected true, received %+v", *version)
	}

	if err := cgrEngineFlags.Parse([]string{"-check_config", "true"}); err != nil {
		t.Fatal(err)
	} else if *checkConfig != true {
		t.Errorf("Expected true, received %+v", *checkConfig)
	}

	if err := cgrEngineFlags.Parse([]string{"-pid", "usr/share/cgrates/cgrates.json"}); err != nil {
		t.Fatal(err)
	} else if *pidFile != "usr/share/cgrates/cgrates.json" {
		t.Errorf("Expected usr/share/cgrates/cgrates.json, received %+v", *pidFile)
	}

	if err := cgrEngineFlags.Parse([]string{"-httprof_path", "http://example.com/"}); err != nil {
		t.Fatal(err)
	} else if *httpPprofPath!= "http://example.com/" {
		t.Errorf("Expected http://example.com/, received %+v", *httpPprofPath)
	}

	if err := cgrEngineFlags.Parse([]string{"-cpuprof_dir", "1"}); err != nil {
		t.Fatal(err)
	} else if *cpuProfDir != "1" {
		t.Errorf("Expected 1, received %+v", *httpPprofPath)
	}

	if err := cgrEngineFlags.Parse([]string{"-memprof_dir", "true"}); err != nil {
		t.Fatal(err)
	} else if *memProfDir != "true" {
		t.Errorf("Expected true received %+v", *memProfDir)
	}

	if err := cgrEngineFlags.Parse([]string{"-memprof_interval", "1s"}); err != nil {
		t.Fatal(err)
	} else if *memProfInterval != time.Second {
		t.Errorf("Expected 1s, received %+v", *memProfInterval)
	}

	if err := cgrEngineFlags.Parse([]string{"-memprof_nrfiles", "3"}); err != nil {
		t.Fatal(err)
	} else if *memProfNrFiles!= 3 {
		t.Errorf("Expected 3, received %+v", *memProfNrFiles)
	}

	if err := cgrEngineFlags.Parse([]string{"-scheduled_shutdown", "1h"}); err != nil {
		t.Fatal(err)
	} else if *scheduledShutdown != "1h" {
		t.Errorf("Expected 1h, received %+v", *scheduledShutdown)
	}

	if err := cgrEngineFlags.Parse([]string{"-singlecpu"}); err != nil {
		t.Fatal(err)
	} else if *singlecpu != true {
		t.Errorf("Expected true, received %+v", *singlecpu)
	}

	if err := cgrEngineFlags.Parse([]string{"-logger", "*stdout"}); err != nil {
		t.Fatal(err)
	} else if *syslogger != "*stdout" {
		t.Errorf("Expected *stdout, received %+v", *syslogger)
	}

	if err := cgrEngineFlags.Parse([]string{"-node_id", "CGRates.org"}); err != nil {
		t.Fatal(err)
	} else if *nodeID != "CGRates.org" {
		t.Errorf("Expected CGRates.org, received %+v", *nodeID)
	}

	if err := cgrEngineFlags.Parse([]string{"-log_level", "7"}); err != nil {
		t.Fatal(err)
	} else if *logLevel != 7 {
		t.Errorf("Expected 7, received %+v", *logLevel)
	}

	if err := cgrEngineFlags.Parse([]string{"-preload", "TestPreloadID"}); err != nil {
		t.Fatal(err)
	} else if *preload != "TestPreloadID" {
		t.Errorf("Expected 7, received %+v", *preload)
	}
}