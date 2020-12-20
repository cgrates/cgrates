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

	"github.com/cgrates/cgrates/loaders"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

//TestLoaderSCoverage for cover testing
func TestLoaderSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	internalLoaderSChan := make(chan rpcclient.ClientConnector, 1)
	rpcInternal := map[string]chan rpcclient.ClientConnector{}
	cM := engine.NewConnManager(cfg, rpcInternal)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srv := NewLoaderService(cfg, db,
		filterSChan, server, internalLoaderSChan,
		cM, anz, srvDep)
	if srv == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(srv))
	}
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv.ldrs = loaders.NewLoaderService(&engine.DataManager{}, []*config.LoaderSCfg{{
		ID:      "test_id",
		Enabled: true,
	}},
		"test", &engine.FilterS{}, &engine.ConnManager{})
	if !srv.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !reflect.DeepEqual(srv.GetLoaderS(), srv.ldrs) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", srv.ldrs, srv.GetLoaderS())
	}
	errStart := srv.Start()
	if errStart == nil || errStart != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, errStart)
	}

}
