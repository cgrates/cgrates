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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestGlobalVarS(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	exp := &GlobalVarS{cfg: cfg}
	if gv := NewGlobalVarS(cfg, nil); !reflect.DeepEqual(gv, exp) {
		t.Errorf("Expected %+v, received %+v", exp, gv)
	}
	if exp.ServiceName() != utils.GlobalVarS {
		t.Errorf("Unexpected service name %q", exp.ServiceName())
	}
	if !exp.ShouldRun() {
		t.Errorf("This service should allways run")
	}
	if !exp.IsRunning() {
		t.Errorf("This service needs to be running")
	}
	cfg.HTTPCfg().ClientOpts[utils.HTTPClientDialTimeoutCfg] = "30as"

	if err := exp.Shutdown(); err != nil {
		t.Fatal(err)
	}
}
