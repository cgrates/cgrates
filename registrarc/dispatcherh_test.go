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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestDispatcherHostsService(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Registar))
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
	cfg.DispatcherHCfg().Enabled = true
	cfg.DispatcherHCfg().Hosts = map[string][]*config.DispatcherHRegistarCfg{
		utils.MetaDefault: {
			{
				ID:                "Host1",
				RegisterTransport: utils.MetaJSON,
			},
		},
	}
	cfg.DispatcherHCfg().RefreshInterval = 100 * time.Millisecond
	cfg.DispatcherHCfg().RegistrarSConns = []string{"conn1"}

	ds := NewRegistrarCService(cfg, engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{}))

	ds.registerHosts()

	host1 := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		ID:     "Host1",
		Conn: &config.RemoteHost{
			Address:   "127.0.0.1:2012",
			Transport: utils.MetaJSON,
		},
	}

	if x, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); !ok {
		t.Errorf("Expected to find Host1 in cache")
	} else if !reflect.DeepEqual(host1, x) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(host1), utils.ToJSON(x))
	}
	cfg.DispatcherHCfg().Hosts = map[string][]*config.DispatcherHRegistarCfg{
		utils.MetaDefault: {
			{
				ID:                "Host2",
				RegisterTransport: utils.MetaJSON,
			},
		},
	}
	config.CgrConfig().CacheCfg().Partitions[utils.CacheDispatcherHosts].Replicate = true
	config.CgrConfig().CacheCfg().ReplicationConns = []string{"*localhost"}
	ds.registerHosts()
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
	cfg.DispatcherHCfg().Hosts = map[string][]*config.DispatcherHRegistarCfg{
		utils.MetaDefault: {
			{
				ID:                "Host1",
				RegisterTransport: utils.MetaJSON,
			},
		},
	}
	ds.Shutdown()
	if _, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); ok {
		t.Errorf("Expected to not find Host2 in cache")
	}

	cfg.ListenCfg().RPCJSONListen = "2012"
	ds.registerHosts()

	ds = NewRegistrarCService(cfg, engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{}))
	ds.Shutdown()
	stopChan := make(chan struct{})
	close(stopChan)
	ds.ListenAndServe(stopChan)
}
