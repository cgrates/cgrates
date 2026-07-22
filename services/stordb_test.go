// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
