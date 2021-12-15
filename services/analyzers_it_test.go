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
	"log"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestAnalyzerSReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	if err := os.MkdirAll("/tmp/analyzers", 0700); err != nil {
		t.Fatal(err)
	}

	shdWg := new(sync.WaitGroup)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(shdWg, nil, cfg.GetReloadChan())
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	anzRPC := make(chan birpc.ClientConnector, 1)
	anz := NewAnalyzerService(cfg, server, filterSChan, anzRPC, srvDep)
	engine.NewConnManager(cfg)
	srvMngr.AddServices(anz,
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	ctx, cancel := context.WithCancel(context.TODO())
	srvMngr.StartServices(ctx, cancel)
	if anz.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "analyzers")
	if err := cfg.V1ReloadConfig(context.Background(), &config.ReloadArgs{
		Section: config.AnalyzerSJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	select {
	case d := <-anzRPC:
		anzRPC <- d
	case <-time.After(time.Second):
		t.Fatal("It took to long to reload the cache")
	}
	if !anz.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	err := anz.Start(ctx, cancel)
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err = anz.Reload(ctx, cancel)
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.AnalyzerSCfg().Enabled = false
	cfg.GetReloadChan() <- config.SectionToService[config.AnalyzerSJSON]
	time.Sleep(10 * time.Millisecond)

	if anz.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	cancel()
	time.Sleep(10 * time.Millisecond)
	if err := os.RemoveAll("/tmp/analyzers"); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzerSReload2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	if err := os.MkdirAll("/tmp/analyzers", 0700); err != nil {
		t.Fatal(err)
	}
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anzRPC := make(chan birpc.ClientConnector, 1)
	anz := NewAnalyzerService(cfg, server, filterSChan, anzRPC, srvDep)
	// anz.stopChan = make(chan struct{})
	// anz.start()
	// close(anz.stopChan)
	// anz.start()
	anz.anz = nil
	if err := os.RemoveAll("/tmp/analyzers"); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzerSReload3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	if err := os.MkdirAll("/tmp/analyzers_test3", 0700); err != nil {
		t.Fatal(err)
	}
	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers_test3"
	err := os.RemoveAll("/tmp/analyzers_test3")
	if err != nil {
		log.Fatal(err)
	}
	cfg.AnalyzerSCfg().IndexType = ""
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anzRPC := make(chan birpc.ClientConnector, 1)
	anz := NewAnalyzerService(cfg, server, filterSChan, anzRPC, srvDep)
	// anz.stopChan = make(chan struct{})
	ctx, cancel := context.WithCancel(context.TODO())
	anz.Start(ctx, cancel)

	anz.anz = nil
	// close(anz.stopChan)

}
