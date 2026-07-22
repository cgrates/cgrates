//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	err = srv.Shutdown()
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}

}
