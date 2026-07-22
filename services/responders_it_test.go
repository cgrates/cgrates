//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestResponderSReload(t *testing.T) {
	// utils.Logger.SetLogLevel(7)
	cfg := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.ThresholdSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	internalChan := make(chan birpc.ClientConnector, 1)
	srv := NewResponderService(cfg, server, internalChan,
		shdChan, anz, srvDep, filterSChan)

	srvName := srv.ServiceName()
	if srvName != utils.ResponderS {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ResponderS, srvName)
	}

	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	err := srv.Start()
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	if !srv.IsRunning() {
		t.Errorf("Expected service to be running")
	}

	err = srv.Start()
	if err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrServiceAlreadyRunning, err)
	}

	err = srv.Reload()
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	srv.syncChans = map[string]chan *engine.Responder{
		"srv": make(chan *engine.Responder, 1),
	}
	err = srv.Shutdown()
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}

}
func TestResponderSReload2(t *testing.T) {
	// utils.Logger.SetLogLevel(7)
	cfg := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.ThresholdSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	internalChan := make(chan birpc.ClientConnector, 1)
	srv := NewResponderService(cfg, server, internalChan,
		shdChan, anz, srvDep, filterSChan)

	srvName := srv.ServiceName()
	if srvName != utils.ResponderS {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ResponderS, srvName)
	}

	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	err := srv.Start()
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	if !srv.IsRunning() {
		t.Errorf("Expected service to be running")
	}

	err = srv.Start()
	if err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrServiceAlreadyRunning, err)
	}

	srv.syncChans = map[string]chan *engine.Responder{
		"srv": make(chan *engine.Responder, 1),
	}
	srv.resp = &engine.Responder{
		ShdChan:          shdChan,
		Timeout:          1,
		Timezone:         "",
		MaxComputedUsage: nil,
	}
	srv.sync()
	srv.RegisterSyncChan("srv", make(chan *engine.Responder, 1))
	srv.resp = nil

}
