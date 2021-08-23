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
	"os"
	"path"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestEventReaderSReload(t *testing.T) {
	for _, dir := range []string{"/tmp/ers/in", "/tmp/ers/out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
	cfg := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	defer func() {
		shdChan.CloseOnce()
		time.Sleep(10 * time.Millisecond)
	}()
	shdWg := new(sync.WaitGroup)
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	db := NewDataDBService(cfg, nil, srvDep)
	sS := NewSessionService(cfg, db, server, make(chan rpcclient.ClientConnector, 1), shdChan, nil, anz, srvDep)
	erS := NewEventReaderService(cfg, filterSChan, shdChan, nil, srvDep)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(erS, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
	if erS.IsRunning() {
		t.Fatal("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "ers_reload", "internal"),
		Section: config.ERsJson,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}
	runtime.Gosched()
	if !erS.IsRunning() {
		t.Fatalf("Expected service to be running")
	}

	runtime.Gosched()
	err := erS.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	time.Sleep(10 * time.Millisecond)
	runtime.Gosched()
	err = erS.Reload()
	if err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.ERsCfg().Enabled = false
	cfg.GetReloadChan(config.ERsJson) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if erS.IsRunning() {
		t.Fatal("Expected service to be down")
	}

}

func TestEventReaderSReload2(t *testing.T) {
	for _, dir := range []string{"/tmp/ers/in", "/tmp/ers/out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.SessionSCfg().Enabled = true
	cfg.ERsCfg().Enabled = true
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			Type: "bad_type",
		},
	}
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	erS := NewEventReaderService(cfg, filterSChan, shdChan, nil, srvDep)
	ers := ers.NewERService(cfg, nil, nil)

	runtime.Gosched()
	srv := erS.(*EventReaderService)
	srv.stopChan = make(chan struct{})
	srv.rldChan = make(chan struct{})
	err := srv.listenAndServe(ers, srv.stopChan, srv.rldChan)
	if err == nil || err.Error() != "unsupported reader type: <bad_type>" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "unsupported reader type: <bad_type>", err)
	}
}
