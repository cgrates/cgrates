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

func testCreateDirs(t *testing.T) {
	for _, dir := range []string{"/tmp/In", "/tmp/Out", "/tmp/LoaderIn", "/tmp/SubpathWithoutMove",
		"/tmp/SubpathLoaderWithMove", "/tmp/SubpathOut", "/tmp/templateLoaderIn", "/tmp/templateLoaderOut",
		"/tmp/customSepLoaderIn", "/tmp/customSepLoaderOut"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
	if err := os.WriteFile(path.Join("/tmp/In", utils.AttributesCsv), []byte(`
#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ALS1,con1,*string:~*req.Account:1001,2014-07-29T15:00:00Z,*string:~*req.Field1:Initial,*req.Field1,*variable,Sub1,true,20
cgrates.org,ALS1,con2;con3,,,,*req.Field2,*variable,Sub2,true,20
`), 0644); err != nil {
		t.Fatal(err.Error())
	}
}

func TestLoaderSReload(t *testing.T) {
	testCreateDirs(t)
	cfg := config.NewDefaultCGRConfig()
	cfg.TemplatesCfg()["attrTemplateLoader"] = []*config.FCTemplate{
		{
			Type:  utils.MetaVariable,
			Path:  "*req.Accounts",
			Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
		},
	}
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)

	shdChan := utils.NewSyncedChan()
	shdWg := new(sync.WaitGroup)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	conMngr := engine.NewConnManager(cfg, nil)
	srv := NewLoaderService(cfg, db, filterSChan,
		server, make(chan birpc.ClientConnector, 1),
		conMngr, anz, srvDep)
	srvMngr.AddServices(srv, db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}

	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	var reply string
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "loaders", "tutinternal"),
			Section: config.LoaderJson,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond)
	if !db.IsRunning() {
		t.Fatal("Expected service to be running")
	}
	time.Sleep(10 * time.Millisecond)
	if !srv.IsRunning() {
		t.Fatal("Expected service to be running")
	}

	err := srv.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	time.Sleep(10 * time.Millisecond)
	err = srv.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	time.Sleep(10 * time.Millisecond)
	for _, v := range cfg.LoaderCfg() {
		v.Enabled = false
	}
	time.Sleep(10 * time.Millisecond)
	cfg.GetReloadChan(config.LoaderJson) <- struct{}{}
	time.Sleep(10 * time.Millisecond)

	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
	testCleanupFiles(t)
}
func testCleanupFiles(t *testing.T) {
	for _, dir := range []string{"/tmp/In", "/tmp/Out", "/tmp/LoaderIn", "/tmp/SubpathWithoutMove",
		"/tmp/SubpathLoaderWithMove", "/tmp/SubpathOut"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}

func TestLoaderSReload2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	for _, ld := range cfg.LoaderCfg() {
		ld.Enabled = false
	}
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	db.dbchan <- new(engine.DataManager)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	srv := NewLoaderService(cfg, db, filterSChan,
		server, make(chan birpc.ClientConnector, 1),
		nil, anz, srvDep)
	err := srv.Start()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoaderSReload3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	for _, ld := range cfg.LoaderCfg() {
		ld.Enabled = false
	}
	cfg.LoaderCfg()[0].Enabled = true
	cfg.LoaderCfg()[0].TpInDir = "/tmp/TestLoaderSReload3"
	cfg.LoaderCfg()[0].RunDelay = -1
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	db.dbchan <- new(engine.DataManager)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	srv := NewLoaderService(cfg, db, filterSChan,
		server, make(chan birpc.ClientConnector, 1),
		nil, anz, srvDep)
	err := srv.Start()
	if err == nil || err.Error() != "no such file or directory" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "no such file or directory", err)
	}
	err = srv.Reload()
	if err == nil || err.Error() != "no such file or directory" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "no such file or directory", err)
	}
}
