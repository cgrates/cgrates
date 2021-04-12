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

package registrarc

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestDispatcherHostsService(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Registrar))
	defer ts.Close()
	cfg := config.NewDefaultCGRConfig()

	cfg.RPCConns()["conn1"] = &config.RPCConn{
		Strategy: rpcclient.PoolFirst,
		Conns: []*config.RemoteHost{{
			Address:     ts.URL,
			Synchronous: true,
			TLS:         false,
			Transport:   rpcclient.HTTPjson,
		}},
	}
	cfg.RegistrarCCfg().Dispatcher.Enabled = true
	cfg.RegistrarCCfg().Dispatcher.Hosts = map[string][]*config.RemoteHost{
		utils.MetaDefault: {
			{
				ID:        "Host1",
				Transport: utils.MetaJSON,
			},
		},
	}
	cfg.RegistrarCCfg().Dispatcher.RefreshInterval = 100 * time.Millisecond
	cfg.RegistrarCCfg().Dispatcher.RegistrarSConns = []string{"conn1"}

	ds := NewRegistrarCService(cfg, engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{}))

	ds.registerDispHosts()

	host1 := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:        "Host1",
			Address:   "127.0.0.1:2012",
			Transport: utils.MetaJSON,
		},
	}

	if x, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); !ok {
		t.Errorf("Expected to find Host1 in cache")
	} else if !reflect.DeepEqual(host1, x) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(host1), utils.ToJSON(x))
	}
	cfg.RegistrarCCfg().Dispatcher.Hosts = map[string][]*config.RemoteHost{
		utils.MetaDefault: {
			{
				ID:        "Host2",
				Transport: utils.MetaJSON,
			},
		},
	}
	config.CgrConfig().CacheCfg().Partitions[utils.CacheDispatcherHosts].Replicate = true
	config.CgrConfig().CacheCfg().ReplicationConns = []string{"*localhost"}
	ds.registerDispHosts()
	host1.ID = "Host2"
	if x, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); !ok {
		t.Errorf("Expected to find Host2 in cache")
	} else if !reflect.DeepEqual(host1, x) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(host1), utils.ToJSON(x))
	}
	unregisterHosts(ds.connMgr, cfg.RegistrarCCfg().Dispatcher, "cgrates.org", utils.RegistrarSv1UnregisterDispatcherHosts)
	if _, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); ok {
		t.Errorf("Expected to not find Host2 in cache")
	}

	config.CgrConfig().CacheCfg().Partitions[utils.CacheDispatcherHosts].Replicate = false
	config.CgrConfig().CacheCfg().ReplicationConns = []string{}

	host1.ID = "Host1"
	cfg.RegistrarCCfg().Dispatcher.Hosts = map[string][]*config.RemoteHost{
		utils.MetaDefault: {
			{
				ID:        "Host1",
				Transport: utils.MetaJSON,
			},
		},
	}
	ds.Shutdown()
	if _, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); ok {
		t.Errorf("Expected to not find Host2 in cache")
	}

	cfg.ListenCfg().RPCJSONListen = "2012"
	ds.registerDispHosts()

	ds = NewRegistrarCService(cfg, engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{}))
	ds.Shutdown()
	stopChan := make(chan struct{})
	close(stopChan)
	ds.ListenAndServe(stopChan, make(chan struct{}))
}

func TestRegistrarcListenAndServe(t *testing.T) {
	//cover purposes only
	cfg := config.NewDefaultCGRConfig()
	cfg.RegistrarCCfg().Dispatcher.Enabled = true
	cfg.RegistrarCCfg().RPC.Enabled = true
	regStSrv := NewRegistrarCService(cfg, nil)
	stopChan := make(chan struct{}, 1)
	rldChan := make(chan struct{}, 1)
	rldChan <- struct{}{}
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(stopChan)
	}()
	regStSrv.ListenAndServe(stopChan, rldChan)
	regStSrv.Shutdown()
}

func TestRegistrarcregisterRPCHostsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RegistrarCCfg().RPC.RegistrarSConns = []string{"testConnID"}
	cfg.RegistrarCCfg().RPC.Hosts = map[string][]*config.RemoteHost{
		utils.MetaDefault: {
			{
				ID:          "",
				Address:     "",
				Transport:   "",
				Synchronous: false,
				TLS:         false,
			},
		},
	}
	regStSrv := NewRegistrarCService(cfg, nil)
	regStSrv.registerRPCHosts()
}

func TestRegisterRPCHosts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RegistrarCCfg().RPC.RegistrarSConns = []string{"errCon1"}
	cfg.RegistrarCCfg().RPC.Hosts = map[string][]*config.RemoteHost{
		"testHostKey": {},
	}
	cfg.RPCConns()["errCon1"] = &config.RPCConn{
		Strategy: utils.MetaFirst,
		PoolSize: 1,
		Conns: []*config.RemoteHost{
			{
				ID:          "errCon1",
				Address:     "127.0.0.1:9999",
				Transport:   "*json",
				Synchronous: true,
			},
		},
	}
	regist := &RegistrarCService{
		cfg:     cfg,
		connMgr: engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{}),
	}
	registCmp := &RegistrarCService{
		cfg:     cfg,
		connMgr: engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{}),
	}
	regist.registerRPCHosts()
	if !reflect.DeepEqual(regist, registCmp) {
		t.Errorf("Expected: %+v ,received: %+v", registCmp, regist)
	}
}

func TestRegistrarcListenAndServedTmCDispatcher(t *testing.T) {
	//cover purposes only
	cfg := config.NewDefaultCGRConfig()
	cfg.RegistrarCCfg().Dispatcher.Enabled = true
	cfg.RegistrarCCfg().Dispatcher.RefreshInterval = 1
	cfg.RegistrarCCfg().RPC.Enabled = true
	regStSrv := NewRegistrarCService(cfg, nil)
	stopChan := make(chan struct{}, 1)
	rldChan := make(chan struct{}, 1)
	go func() {
		time.Sleep(20 * time.Millisecond)
		close(stopChan)
	}()
	regStSrv.ListenAndServe(stopChan, rldChan)
	regStSrv.Shutdown()
}

func TestRegistrarcListenAndServedTmCRPC(t *testing.T) {
	//cover purposes only
	cfg := config.NewDefaultCGRConfig()
	cfg.RegistrarCCfg().Dispatcher.Enabled = true
	cfg.RegistrarCCfg().RPC.Enabled = true
	cfg.RegistrarCCfg().RPC.RefreshInterval = 1
	regStSrv := NewRegistrarCService(cfg, nil)
	stopChan := make(chan struct{}, 1)
	rldChan := make(chan struct{}, 1)
	go func() {
		time.Sleep(20 * time.Millisecond)
		close(stopChan)
	}()
	regStSrv.ListenAndServe(stopChan, rldChan)
	regStSrv.Shutdown()
}
