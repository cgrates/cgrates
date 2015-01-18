/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package config

import (
	"flag"
	"testing"
)

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, disabled by default.") // This flag will be passed here via "go test -local" args

var mfCgrCfg *CGRConfig

func TestInitConfig(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	if mfCgrCfg, err = NewCGRConfigFromFolder("/usr/share/cgrates/conf/samples/multifiles"); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}
