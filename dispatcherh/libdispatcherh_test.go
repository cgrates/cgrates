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
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestRegisterArgsAsDispatcherHosts(t *testing.T) {
	args := &RegisterArgs{
		Tenant: "cgrates.org",
		Hosts: []*RegisterHostCfg{
			{
				ID:        "Host1",
				Port:      "2012",
				TLS:       true,
				Transport: utils.MetaJSON,
			},
			{
				ID:        "Host2",
				Port:      "2013",
				TLS:       false,
				Transport: utils.MetaGOB,
			},
		},
		Opts: make(map[string]interface{}),
	}
	exp := []*engine.DispatcherHost{
		{
			Tenant: "cgrates.org",
			ID:     "Host1",
			Conn: &config.RemoteHost{
				Address:   "127.0.0.1:2012",
				TLS:       true,
				Transport: utils.MetaJSON,
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "Host2",
			Conn: &config.RemoteHost{
				Address:   "127.0.0.1:2013",
				TLS:       false,
				Transport: utils.MetaGOB,
			},
		},
	}
	if rply := args.AsDispatcherHosts("127.0.0.1"); !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func TestGetConnPort(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}

	cfg.ListenCfg().RPCJSONTLSListen = ":2072"
	cfg.ListenCfg().RPCJSONListen = ":2012"
	cfg.ListenCfg().RPCGOBTLSListen = ":2073"
	cfg.ListenCfg().RPCGOBListen = ":2013"
	cfg.ListenCfg().HTTPTLSListen = ":2081"
	cfg.ListenCfg().HTTPListen = ":2080"
	cfg.HTTPCfg().HTTPJsonRPCURL = "/json_rpc"

	if port, err := getConnPort(cfg, utils.MetaJSON, false); err != nil {
		t.Fatal(err)
	} else if port != "2012" {
		t.Errorf("Expected: %q ,received: %q", "2012", port)
	}
	if port, err := getConnPort(cfg, utils.MetaJSON, true); err != nil {
		t.Fatal(err)
	} else if port != "2072" {
		t.Errorf("Expected: %q ,received: %q", "2072", port)
	}
	if port, err := getConnPort(cfg, utils.MetaGOB, false); err != nil {
		t.Fatal(err)
	} else if port != "2013" {
		t.Errorf("Expected: %q ,received: %q", "2013", port)
	}
	if port, err := getConnPort(cfg, utils.MetaGOB, true); err != nil {
		t.Fatal(err)
	} else if port != "2073" {
		t.Errorf("Expected: %q ,received: %q", "2073", port)
	}
	if port, err := getConnPort(cfg, rpcclient.HTTPjson, false); err != nil {
		t.Fatal(err)
	} else if port != "2080/json_rpc" {
		t.Errorf("Expected: %q ,received: %q", "2080/json_rpc", port)
	}
	if port, err := getConnPort(cfg, rpcclient.HTTPjson, true); err != nil {
		t.Fatal(err)
	} else if port != "2081/json_rpc" {
		t.Errorf("Expected: %q ,received: %q", "2081/json_rpc", port)
	}
	cfg.ListenCfg().RPCJSONListen = "2012"
	if _, err := getConnPort(cfg, utils.MetaJSON, false); err == nil {
		t.Fatal("Expected error received nil")
	}
}

func TestRegister(t *testing.T) {
	ra := &RegisterArgs{
		Tenant: "cgrates.org",
		Hosts: []*RegisterHostCfg{
			{
				ID:        "Host1",
				Port:      "2012",
				TLS:       true,
				Transport: utils.MetaJSON,
			},
			{
				ID:        "Host2",
				Port:      "2013",
				TLS:       false,
				Transport: utils.MetaGOB,
			},
		},
		Opts: make(map[string]interface{}),
	}
	raJSON, err := json.Marshal([]interface{}{ra})
	id := json.RawMessage("1")
	if err != nil {
		t.Fatal(err)
	}
	args := utils.NewServerRequest(utils.DispatcherHv1RegisterHosts, raJSON, id)
	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/json_rpc", bytes.NewBuffer(argsJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:2356"
	engine.SetCache(engine.NewCacheS(config.CgrConfig(), nil))
	if rplyID, err := register(req); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(id, *rplyID) {
		t.Errorf("Expected: %q ,received: %q", string(id), string(*rplyID))
	}

	host1 := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		ID:     "Host1",
		Conn: &config.RemoteHost{
			Address:   "127.0.0.1:2012",
			TLS:       true,
			Transport: utils.MetaJSON,
		},
	}
	host2 := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		ID:     "Host2",
		Conn: &config.RemoteHost{
			Address:   "127.0.0.1:2013",
			TLS:       false,
			Transport: utils.MetaGOB,
		},
	}

	if x, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); !ok {
		t.Errorf("Expected to find Host1 in cache")
	} else if !reflect.DeepEqual(host1, x) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(host1), utils.ToJSON(x))
	}
	if x, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host2.TenantID()); !ok {
		t.Errorf("Expected to find Host2 in cache")
	} else if !reflect.DeepEqual(host2, x) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(host2), utils.ToJSON(x))
	}

	if _, err := register(req); err != io.EOF {
		t.Errorf("Expected error: %s ,received: %v", io.EOF, err)
	}

	ua := &UnregisterArgs{
		Tenant: "cgrates.org",
		IDs:    []string{"Host1", "Host2"},
		Opts:   make(map[string]interface{}),
	}
	uaJSON, err := json.Marshal([]interface{}{ua})
	id = json.RawMessage("2")
	if err != nil {
		t.Fatal(err)
	}
	uargs := utils.NewServerRequest(utils.DispatcherHv1UnregisterHosts, uaJSON, id)
	uargsJSON, err := json.Marshal(uargs)
	if err != nil {
		t.Fatal(err)
	}
	req, err = http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/json_rpc", bytes.NewBuffer(uargsJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:2356"
	if rplyID, err := register(req); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(id, *rplyID) {
		t.Errorf("Expected: %q ,received: %q", string(id), string(*rplyID))
	}
	if x, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host1.TenantID()); ok {
		t.Errorf("Expected to not find Host1 in cache %+v", x)
	}
	if x, ok := engine.Cache.Get(utils.CacheDispatcherHosts, host2.TenantID()); ok {
		t.Errorf("Expected to not find Host2 in cache %+v", x)
	}
	errCfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	engine.NewConnManager(errCfg, map[string]chan rpcclient.ClientConnector{})
	errCfg.CacheCfg().Partitions[utils.CacheDispatcherHosts].Replicate = true
	errCfg.RPCConns()["errCon"] = &config.RPCConn{
		Strategy: utils.MetaFirst,
		PoolSize: 1,
		Conns: []*config.RemoteHost{
			{
				Address:     "127.0.0.1:5612",
				Transport:   "*json",
				Synchronous: false,
				TLS:         false,
			},
		},
	}
	errCfg.CacheCfg().ReplicationConns = []string{"errCon"}
	engine.SetCache(engine.NewCacheS(errCfg, nil))
	req.Body = ioutil.NopCloser(bytes.NewBuffer(uargsJSON))
	if _, err := register(req); err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected error: %s ,received: %v", utils.ErrPartiallyExecuted, err)
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(argsJSON))
	if _, err := register(req); err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected error: %s ,received: %v", utils.ErrPartiallyExecuted, err)
	}

	req.RemoteAddr = "127.0.0"
	req.Body = ioutil.NopCloser(bytes.NewBuffer(argsJSON))
	if _, err := register(req); err == nil {
		t.Errorf("Expected error,received: nil")
	}
	args2 := utils.NewServerRequest(utils.DispatcherHv1RegisterHosts, id, id)
	args2JSON, err := json.Marshal(args2)
	if err != nil {
		t.Fatal(err)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(args2JSON))
	if _, err := register(req); err == nil {
		t.Errorf("Expected error,received: nil")
	}
	args2 = utils.NewServerRequest(utils.DispatcherHv1UnregisterHosts, id, id)
	args2JSON, err = json.Marshal(args2)
	if err != nil {
		t.Fatal(err)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(args2JSON))
	if _, err := register(req); err == nil {
		t.Errorf("Expected error,received: nil")
	}
	args2 = utils.NewServerRequest(utils.DispatcherSv1GetProfileForEvent, id, id)
	args2JSON, err = json.Marshal(args2)
	if err != nil {
		t.Fatal(err)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(args2JSON))
	if _, err := register(req); err == nil {
		t.Errorf("Expected error,received: nil")
	}

	engine.SetCache(engine.NewCacheS(config.CgrConfig(), nil))
}

type errRecorder struct{}

func (*errRecorder) Header() http.Header        { return make(http.Header) }
func (*errRecorder) Write([]byte) (int, error)  { return 0, io.EOF }
func (*errRecorder) WriteHeader(statusCode int) {}

func TestRegistar(t *testing.T) {
	w := httptest.NewRecorder()
	ra := &RegisterArgs{
		Tenant: "cgrates.org",
		Hosts: []*RegisterHostCfg{
			{
				ID:        "Host1",
				Port:      "2012",
				TLS:       true,
				Transport: utils.MetaJSON,
			},
			{
				ID:        "Host2",
				Port:      "2013",
				TLS:       false,
				Transport: utils.MetaGOB,
			},
		},
		Opts: make(map[string]interface{}),
	}
	raJSON, err := json.Marshal([]interface{}{ra})
	id := json.RawMessage("1")
	if err != nil {
		t.Fatal(err)
	}
	args := utils.NewServerRequest(utils.DispatcherHv1RegisterHosts, raJSON, id)
	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/json_rpc", bytes.NewBuffer(argsJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:2356"

	Registar(w, req)
	exp := "{\"id\":1,\"result\":\"OK\",\"error\":null}\n"
	if w.Body.String() != exp {
		t.Errorf("Expected: %q ,received: %q", exp, w.Body.String())
	}

	w = httptest.NewRecorder()
	Registar(w, req)
	exp = "{\"id\":0,\"result\":null,\"error\":\"EOF\"}\n"
	if w.Body.String() != exp {
		t.Errorf("Expected: %q ,received: %q", exp, w.Body.String())
	}

	Registar(new(errRecorder), req)
}
