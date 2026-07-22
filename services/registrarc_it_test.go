//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
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
	connMngr := engine.NewConnManager(cfg, nil)
	srv := NewRegistrarCService(cfg, server, connMngr, anz, srvDep)
	srvMngr.AddServices(srv, db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
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
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    cfgPath,
			Section: config.RegistrarCJson,
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
	err := srv.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err = srv.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.RegistrarCCfg().RPC.RegistrarSConns = nil
	cfg.RegistrarCCfg().Dispatchers.RegistrarSConns = nil
	cfg.GetReloadChan(config.RegistrarCJson) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}
