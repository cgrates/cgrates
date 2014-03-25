/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package config

import (
	"flag"
	"path"
	"testing"
)

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")

func TestLoadXmlCfg(t *testing.T) {
	if !*testLocal {
		return
	}
	cfgPath := path.Join(*dataDir, "conf", "samples", "config_local_test.cfg")
	cfg, err := NewCGRConfig(&cfgPath)
	if err != nil {
		t.Error(err)
	}
	if cfg.XmlCfgDocument == nil {
		t.Error("Did not load the XML Config Document")
	}
	if cdreFWCfg, err := cfg.XmlCfgDocument.GetCdreFWCfg("CDREFW-A"); err != nil {
		t.Error(err)
	} else if cdreFWCfg == nil {
		t.Error("Could not retrieve CDRExporter FixedWidth config instance")
	}
}
