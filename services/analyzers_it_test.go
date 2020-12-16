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
	"reflect"
	"sync"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestNewAnalyzerService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	close(chS.GetPrecacheChannel(utils.CacheActionProfiles))
	close(chS.GetPrecacheChannel(utils.CacheActionProfilesFilterIndexes))
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	connChan := make(chan rpcclient.ClientConnector, 1)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, connChan, srvDep)
	expected := &AnalyzerService{
		connChan:    connChan,
		cfg:         cfg,
		server:      server,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		srvDep:      srvDep,
	}
	if !reflect.DeepEqual(anz, expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expected), utils.ToJSON(anz))
	}
}

func TestAnalyzerSNotRunning(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	close(chS.GetPrecacheChannel(utils.CacheActionProfiles))
	close(chS.GetPrecacheChannel(utils.CacheActionProfilesFilterIndexes))
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	connChan := make(chan rpcclient.ClientConnector, 1)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, connChan, srvDep)
	if anz.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	serviceName := anz.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.AnalyzerS) {
		t.Errorf("\nExpecting <ActionS>,\n Received <%+v>", serviceName)
	}
	shouldRun := anz.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	getAnalyzerS := anz.GetAnalyzerS()
	if !reflect.DeepEqual(anz.anz, getAnalyzerS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(anz.anz), utils.ToJSON(getAnalyzerS))
	}
}
