/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package services

import (
	"reflect"
	"sync"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestStorDBServiceCoverage for cover testing
func TestStorDBServiceCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewStorDBService(cfg, false, srvDep)
	err := srv.IsRunning()
	if err == true {
		t.Errorf("Expected service to be down")
	}
	var dErr error
	srv.db, dErr = engine.NewInternalDB([]string{"test"}, []string{"test2"}, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	err = srv.IsRunning()
	if err == false {
		t.Errorf("Expected service to be running")
	}
	srv.oldDBCfg = &config.StorDbCfg{
		Type:     utils.MetaInternal,
		Host:     "test_host",
		Port:     "test_port",
		Name:     "test_name",
		User:     "test_user",
		Password: "test_pass",
	}
	serviceName := srv.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.StorDB) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.StorDB, serviceName)
	}
	shouldRun := srv.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	srv.Shutdown()
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
