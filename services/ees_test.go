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
	"path"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

//TestEventExporterSCoverage for cover testing
func TestEventExporterSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.AttributeSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	shdWg := new(sync.WaitGroup)
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	chS := engine.NewCacheS(cfg, nil, nil)
	close(chS.GetPrecacheChannel(utils.CacheAttributeProfiles))
	close(chS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes))
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	attrS := NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1),
		anz, srvDep)
	ees := NewEventExporterService(cfg, filterSChan, engine.NewConnManager(cfg, nil), server, make(chan rpcclient.ClientConnector, 1), anz, srvDep)
	srvMngr.AddServices(ees, attrS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if ees.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	fcTmp := &config.FCTemplate{Tag: "TenantID",
		Path:      "Tenant",
		Type:      utils.MetaVariable,
		Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
		Mandatory: true,
		Layout:    time.RFC3339,
	}
	fcTmp.ComputePath()
	cfg.TemplatesCfg()["requiredFields"] = []*config.FCTemplate{fcTmp}
	var reply string
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "ees"),
		Section: config.EEsJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !ees.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	err := ees.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err = ees.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.EEsCfg().Enabled = false
	cfg.GetReloadChan(config.EEsJson) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if ees.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}
