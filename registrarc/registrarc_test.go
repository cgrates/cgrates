// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package registrarc

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// func TestDispatcherHostsService(t *testing.T) {
// 	ts := httptest.NewServer(http.HandlerFunc(Registrar))
// 	defer ts.Close()
// 	cfg := config.NewDefaultCGRConfig()

// 	cfg.RPCConns()["conn1"] = &config.RPCConn{
// 		Strategy: rpcclient.PoolFirst,
// 		Conns: []*config.RemoteHost{{
// 			Address:   ts.URL,
// 			TLS:       false,
// 			Transport: rpcclient.HTTPjson,
// 		}},
// 	}
// 	cfg.RegistrarCCfg().Dispatchers.Hosts = map[string][]*config.RemoteHost{
// 		utils.MetaDefault: {
// 			{
// 				ID:        "Host1",
// 				Transport: utils.MetaJSON,
// 			},
// 		},
// 	}
// 	cfg.RegistrarCCfg().Dispatchers.RefreshInterval = 100 * time.Millisecond
// 	cfg.RegistrarCCfg().Dispatchers.RegistrarSConns = []string{"conn1"}

// 	ds := NewRegistrarCService(cfg, engine.NewConnManager(cfg))

// 	ds.registerDispHosts()

// 	host1 := &engine.DispatcherHost{
// 		Tenant: "cgrates.org",
// 		RemoteHost: &config.RemoteHost{
// 			ID:        "Host1",
// 			Address:   "127.0.0.1:2012",
// 			Transport: utils.MetaJSON,
// 		},
// 	}

// 	if x, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); !ok {
// 		t.Errorf("Expected to find Host1 in cache")
// 	} else if !reflect.DeepEqual(host1, x) {
// 		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(host1), utils.ToJSON(x))
// 	}
// 	cfg.RegistrarCCfg().Dispatchers.Hosts = map[string][]*config.RemoteHost{
// 		utils.MetaDefault: {
// 			{
// 				ID:        "Host2",
// 				Transport: utils.MetaJSON,
// 			},
// 		},
// 	}
// 	cfg.CacheCfg().Partitions[utils.CacheDispatcherHosts].Replicate = true
// 	cfg.CacheCfg().ReplicationConns = []string{"*localhost"}
// 	ds.registerDispHosts()
// 	host1.ID = "Host2"
// 	if x, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); !ok {
// 		t.Errorf("Expected to find Host2 in cache")
// 	} else if !reflect.DeepEqual(host1, x) {
// 		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(host1), utils.ToJSON(x))
// 	}
// 	unregisterHosts(ds.connMgr, cfg.RegistrarCCfg().Dispatchers, "cgrates.org", utils.RegistrarSv1UnregisterDispatcherHosts)
// 	if _, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); ok {
// 		t.Errorf("Expected to not find Host2 in cache")
// 	}

// 	cfg.CacheCfg().Partitions[utils.CacheDispatcherHosts].Replicate = false
// 	cfg.CacheCfg().ReplicationConns = []string{}

// 	host1.ID = "Host1"
// 	cfg.RegistrarCCfg().Dispatchers.Hosts = map[string][]*config.RemoteHost{
// 		utils.MetaDefault: {
// 			{
// 				ID:        "Host1",
// 				Transport: utils.MetaJSON,
// 			},
// 		},
// 	}
// 	ds.Shutdown()
// 	if _, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); ok {
// 		t.Errorf("Expected to not find Host2 in cache")
// 	}

// 	cfg.ListenCfg().RPCJSONListen = "2012"
// 	ds.registerDispHosts()

// 	ds = NewRegistrarCService(cfg, engine.NewConnManager(cfg))
// 	ds.Shutdown()
// 	stopChan := make(chan struct{})
// 	close(stopChan)
// 	ds.ListenAndServe(stopChan, make(chan struct{}))
// }

// func TestRegistrarcListenAndServe(t *testing.T) {
// 	//cover purposes only
// 	cfg := config.NewDefaultCGRConfig()
// 	regStSrv := NewRegistrarCService(cfg, nil)
// 	stopChan := make(chan struct{}, 1)
// 	rldChan := make(chan struct{}, 1)
// 	rldChan <- struct{}{}
// 	go func() {
// 		time.Sleep(10 * time.Millisecond)
// 		close(stopChan)
// 	}()
// 	regStSrv.ListenAndServe(stopChan, rldChan)
// 	regStSrv.Shutdown()
// }

func TestRegistrarcregisterRPCHostsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RegistrarCCfg().RPC.RegistrarSConns = []string{"testConnID"}
	cfg.RegistrarCCfg().RPC.Hosts = map[string][]*config.RemoteHost{
		utils.MetaDefault: {
			{
				ID:        "",
				Address:   "",
				Transport: "",
				TLS:       false,
			},
		},
	}
	regStSrv := NewRegistrarCService(cfg, nil)
	regStSrv.registerRPCHosts()
}

func TestRegisterRPCHosts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewGuardianLocker(cfg)
	cfg.RegistrarCCfg().RPC.RegistrarSConns = []string{"errCon1"}
	cfg.RegistrarCCfg().RPC.Hosts = map[string][]*config.RemoteHost{
		"testHostKey": {},
	}
	cfg.RPCConns()["errCon1"] = &config.RPCConn{
		Strategy: utils.MetaFirst,
		PoolSize: 1,
		Conns: []*config.RemoteHost{
			{
				ID:        "errCon1",
				Address:   "127.0.0.1:9999",
				Transport: "*json",
			},
		},
	}
	cache := engine.NewCacheS(cfg, nil, nil, nil, locker)
	regist := &RegistrarCService{
		cfg:     cfg,
		connMgr: engine.NewConnManager(cfg),
	}
	regist.connMgr.SetCache(cache)
	registCmp := &RegistrarCService{
		cfg:     cfg,
		connMgr: engine.NewConnManager(cfg),
	}
	registCmp.connMgr.SetCache(cache)
	regist.registerRPCHosts()
	if !reflect.DeepEqual(regist, registCmp) {
		t.Errorf("Expected: %+v ,received: %+v", registCmp, regist)
	}
}

// func TestRegistrarcListenAndServedTmCDispatcher(t *testing.T) {
// 	//cover purposes only
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.RegistrarCCfg().Dispatchers.RefreshInterval = 1
// 	regStSrv := NewRegistrarCService(cfg, nil)
// 	stopChan := make(chan struct{}, 1)
// 	rldChan := make(chan struct{}, 1)
// 	go func() {
// 		time.Sleep(20 * time.Millisecond)
// 		close(stopChan)
// 	}()
// 	regStSrv.ListenAndServe(stopChan, rldChan)
// 	regStSrv.Shutdown()
// }

// func TestRegistrarcListenAndServedTmCRPC(t *testing.T) {
// 	//cover purposes only
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.RegistrarCCfg().RPC.RefreshInterval = 1
// 	regStSrv := NewRegistrarCService(cfg, nil)
// 	stopChan := make(chan struct{}, 1)
// 	rldChan := make(chan struct{}, 1)
// 	go func() {
// 		time.Sleep(20 * time.Millisecond)
// 		close(stopChan)
// 	}()
// 	regStSrv.ListenAndServe(stopChan, rldChan)
// 	regStSrv.Shutdown()
// }
