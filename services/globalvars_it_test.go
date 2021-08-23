//go:build integration
// +build integration

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
	"sync"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestGlobalVarsReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewGlobalVarS(cfg, srvDep)
	err := srv.Start()
	if !srv.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	err = srv.Reload()
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}

	err2 := srv.ServiceName()
	if err2 != utils.GlobalVarS {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.GlobalVarS, err2)
	}

	err3 := srv.ShouldRun()
	if err3 != true {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", true, err3)
	}
	cfg.HTTPCfg().ClientOpts = map[string]interface{}{
		utils.HTTPClientTLSHandshakeTimeoutCfg: false,
	}
	err = srv.(*GlobalVarS).initHTTPTransport()
	if err == nil || err.Error() != "cannot convert field: false to time.Duration" {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	err = srv.Shutdown()
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}

}
