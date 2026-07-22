//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func TestEventExporterSReload(t *testing.T) {
	for _, dir := range []string{"/tmp/testCSV", "/tmp/testComposedCSV", "/tmp/testFWV", "/tmp/testCSVMasked",
		"/tmp/testCSVfromVirt", "/tmp/testCSVExpTemp"} {
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
	cfg.AttributeSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	shdWg := new(sync.WaitGroup)
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	chS := engine.NewCacheS(cfg, nil, nil)
	close(chS.GetPrecacheChannel(utils.CacheAttributeProfiles))
	close(chS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes))
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	attrS := NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan birpc.ClientConnector, 1),
		anz, srvDep)
	ees := NewEventExporterService(cfg, filterSChan, engine.NewConnManager(cfg, nil),
		server, make(chan birpc.ClientConnector, 2), anz, srvDep)
	srvMngr.AddServices(ees, attrS, db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
	if ees.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	fcTmp := &config.FCTemplate{Tag: "TenantID",
		Path:      "Tenant",
		Type:      utils.MetaVariable,
		Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
		Mandatory: true,
		Layout:    time.RFC3339,
	}
	fcTmp.ComputePath()
	cfg.TemplatesCfg()["requiredFields"] = []*config.FCTemplate{fcTmp}
	var reply string
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "ees"),
			Section: config.EEsJson,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !ees.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	err := ees.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err = ees.Reload()
	if err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.EEsCfg().Enabled = false
	cfg.GetReloadChan(config.EEsJson) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if ees.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}

func TestEventExporterSReload2(t *testing.T) {
	for _, dir := range []string{"/tmp/testCSV", "/tmp/testComposedCSV", "/tmp/testFWV", "/tmp/testCSVMasked",
		"/tmp/testCSVfromVirt", "/tmp/testCSVExpTemp"} {
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
	cfg.AttributeSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	ees := NewEventExporterService(cfg, filterSChan, engine.NewConnManager(cfg, nil),
		server, make(chan birpc.ClientConnector, 2), anz, srvDep)
	if ees.IsRunning() {
		t.Fatalf("Expected service to be down")
	}

}
