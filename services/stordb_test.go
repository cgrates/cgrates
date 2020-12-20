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

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

//TestStorDBServiceCoverage for cover testing
func TestStorDBServiceCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewStorDBService(cfg, srvDep)
	err := srv.IsRunning()
	if err == true {
		t.Errorf("Expected service to be down")
	}
	srv.db = engine.NewInternalDB([]string{"test"}, []string{"test2"}, true)
	err = srv.IsRunning()
	if err == false {
		t.Errorf("Expected service to be running")
	}
	err2 := srv.Start()
	if err2 == nil || err2 != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err2)
	}
	srv.oldDBCfg = &config.StorDbCfg{
		Type:     utils.INTERNAL,
		Host:     "test_host",
		Port:     "test_port",
		Name:     "test_name",
		User:     "test_user",
		Password: "test_pass",
	}
	err2 = srv.Reload()
	if err2 == nil {
		t.Errorf("\nExpecting <Error 1045: Access denied for user 'cgrates'@'localhost' (using password: NO)>,\n Received <%+v>", err2)
	}
}
