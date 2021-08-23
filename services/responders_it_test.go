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

	"github.com/cgrates/rpcclient"

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
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	internalChan := make(chan rpcclient.ClientConnector, 1)
	srv := NewResponderService(cfg, server, internalChan,
		shdChan, anz, srvDep)

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
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	internalChan := make(chan rpcclient.ClientConnector, 1)
	srv := NewResponderService(cfg, server, internalChan,
		shdChan, anz, srvDep)

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
