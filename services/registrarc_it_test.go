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
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestDispatcherHReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()["dispConn"] = &config.RPCConn{
		Strategy: rpcclient.PoolFirst,
		Conns: []*config.RemoteHost{{
			Address:   "http://127.0.0.1:2080/dispatchers_registrar",
			Transport: rpcclient.HTTPjson,
		}},
	}
	shdWg := new(sync.WaitGroup)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	cls := commonlisteners.NewCommonListenerS(nil)
	srvMngr := servmanager.NewServiceManager(shdWg, nil, cfg)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	anz := NewAnalyzerService(cfg, cls, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
	connMngr := engine.NewConnManager(cfg)
	srv := NewRegistrarCService(cfg, cls, connMngr, anz, srvDep)
	srvMngr.AddServices(srv,
		NewLoaderService(cfg, db, filterSChan, cls,
			make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	ctx, cancel := context.WithCancel(context.TODO())
	srvMngr.StartServices(ctx, cancel)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	cfgPath := t.TempDir()
	filePath := filepath.Join(cfgPath, "cgrates.json")
	if err := os.WriteFile(filePath, []byte(`{
"general": {
        "node_id": "ALL"
},
"listen": {
        "rpc_json": ":6012",
        "rpc_gob": ":6013",
        "http": ":6080"
},
"rpc_conns": {
        "dispConn": {
                "strategy": "*first",
                "conns": [{
                        "address": "http://127.0.0.1:2080/registrar",
                        "transport": "*http_jsonrpc"
                }]
        }
},
"registrarc": {
        "dispatchers": {
                "enabled": true,
                "registrars_conns": ["dispConn"],
                "hosts": [{
                        "Tenant": "*default",
                        "ID": "hostB",
                        "transport": "*json",
                        "tls": false
                }],
                "refresh_interval": "1s"
        }
}
}`), 0644); err != nil {
		t.Fatal(err)
	}

	var reply string
	cfg.ConfigPath = cfgPath
	if err := cfg.V1ReloadConfig(context.Background(), &config.ReloadArgs{
		Section: config.RegistrarCJSON,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	runtime.Gosched()
	runtime.Gosched()
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !srv.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	err := srv.Start(ctx, cancel)
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err = srv.Reload(ctx, cancel)
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.RegistrarCCfg().Dispatchers.RegistrarSConns = []string{}
	cfg.GetReloadChan() <- config.SectionToService[config.RegistrarCJSON]
	time.Sleep(10 * time.Millisecond)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	cancel()
	time.Sleep(10 * time.Millisecond)
}
