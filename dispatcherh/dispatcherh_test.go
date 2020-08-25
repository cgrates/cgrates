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

package dispatcherh

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestDispatcherHostsServiceCall(t *testing.T) {
	if err := new(DispatcherHostsService).Call("", nil, nil); err != utils.ErrNotImplemented {
		t.Errorf("Expected error: %s ,received: %v", utils.ErrNotImplemented, err)
	}
}

func TestDispatcherHostsService(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Registar))
	defer ts.Close()
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}

	cfg.RPCConns()["conn1"] = &config.RPCConn{
		Strategy: rpcclient.PoolFirst,
		Conns: []*config.RemoteHost{{
			Address:     ts.URL,
			Synchronous: true,
			TLS:         false,
			Transport:   rpcclient.HTTPjson,
		}},
	}
	cfg.DispatcherHCfg().Enabled = true
	cfg.DispatcherHCfg().Hosts = map[string][]string{utils.MetaDefault: {"Host1"}}
	cfg.DispatcherHCfg().RegisterInterval = 100 * time.Millisecond
	cfg.DispatcherHCfg().RegisterTransport = utils.MetaJSON
	cfg.DispatcherHCfg().DispatchersConns = []string{"conn1"}

	ds := NewDispatcherHService(cfg, engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{}))

	if err = ds.registerHosts(); err != nil {
		t.Fatal(err)
	}

	host1 := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		ID:     "Host1",
		Conns: []*config.RemoteHost{{
			Address:   "127.0.0.1:2012",
			Transport: utils.MetaJSON,
		}},
	}

	if x, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); !ok {
		t.Errorf("Expected to find Host1 in cache")
	} else if !reflect.DeepEqual(host1, x) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(host1), utils.ToJSON(x))
	}
	cfg.DispatcherHCfg().Hosts = map[string][]string{utils.MetaDefault: {"Host2"}}
	config.CgrConfig().CacheCfg().Partitions[utils.CacheDispatcherHosts].Replicate = true
	config.CgrConfig().CacheCfg().ReplicationConns = []string{"*localhost"}
	if err = ds.registerHosts(); err != nil {
		t.Fatal(err)
	}
	host1.ID = "Host2"
	if x, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); !ok {
		t.Errorf("Expected to find Host2 in cache")
	} else if !reflect.DeepEqual(host1, x) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(host1), utils.ToJSON(x))
	}
	ds.unregisterHosts()
	if _, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); ok {
		t.Errorf("Expected to not find Host2 in cache")
	}

	config.CgrConfig().CacheCfg().Partitions[utils.CacheDispatcherHosts].Replicate = false
	config.CgrConfig().CacheCfg().ReplicationConns = []string{}

	host1.ID = "Host1"
	cfg.DispatcherHCfg().Hosts = map[string][]string{utils.MetaDefault: {"Host1"}}
	if err = ds.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if _, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); ok {
		t.Errorf("Expected to not find Host2 in cache")
	}

	cfg.ListenCfg().RPCJSONListen = "2012"
	if err = ds.registerHosts(); err == nil {
		t.Fatal("Expected error received nil")
	}

	ds = NewDispatcherHService(cfg, engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{}))
	config.CgrConfig().CacheCfg().Partitions[utils.CacheDispatcherHosts].Replicate = true
	config.CgrConfig().CacheCfg().ReplicationConns = []string{"*localhost"}
	if err = ds.ListenAndServe(); err == nil {
		t.Fatal("Expected error received nil")
	}

	config.CgrConfig().CacheCfg().Partitions[utils.CacheDispatcherHosts].Replicate = false
	config.CgrConfig().CacheCfg().ReplicationConns = []string{}
	cfg.ListenCfg().RPCJSONListen = "127.0.0.1:2012"

	ds = NewDispatcherHService(cfg, engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{}))
	ds.Shutdown()
	if err = ds.ListenAndServe(); err != nil {
		t.Fatal(err)
	}
}
