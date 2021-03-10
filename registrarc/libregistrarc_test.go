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
			RemoteHost: &config.RemoteHost{
				ID:        "Host1",
				Address:   "127.0.0.1:2012",
				TLS:       true,
				Transport: utils.MetaJSON,
			},
		},
		{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:        "Host2",
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
	cfg := config.NewDefaultCGRConfig()

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
	args := utils.NewServerRequest(utils.RegistrarSv1RegisterDispatcherHosts, raJSON, id)
	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/json_rpc", bytes.NewBuffer(argsJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:2356"
	engine.Cache = engine.NewCacheS(config.CgrConfig(), nil, nil)
	if rplyID, err := register(req); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(id, *rplyID) {
		t.Errorf("Expected: %q ,received: %q", string(id), string(*rplyID))
	}

	host1 := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:        "Host1",
			Address:   "127.0.0.1:2012",
			TLS:       true,
			Transport: utils.MetaJSON,
		},
	}
	host2 := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:        "Host2",
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
	uargs := utils.NewServerRequest(utils.RegistrarSv1UnregisterDispatcherHosts, uaJSON, id)
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
	errCfg := config.NewDefaultCGRConfig()

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
	engine.Cache = engine.NewCacheS(errCfg, nil, nil)
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
	args2 := utils.NewServerRequest(utils.RegistrarSv1RegisterDispatcherHosts, id, id)
	args2JSON, err := json.Marshal(args2)
	if err != nil {
		t.Fatal(err)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(args2JSON))
	if _, err := register(req); err == nil {
		t.Errorf("Expected error,received: nil")
	}
	args2 = utils.NewServerRequest(utils.RegistrarSv1UnregisterDispatcherHosts, id, id)
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
	args2 = utils.NewServerRequest(utils.DispatcherSv1GetProfileForEvent, id, id)
	args2JSON, err = json.Marshal(args2)
	if err != nil {
		t.Fatal(err)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(args2JSON))
	if _, err := register(req); err == nil {
		t.Errorf("Expected error,received: nil")
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(argsJSON))
	if _, err := register(req); err == nil {
		t.Errorf("Expected error,received: nil")
	}
	args2 = utils.NewServerRequest(utils.RegistrarSv1RegisterRPCHosts, id, id)
	args2JSON, err = json.Marshal(args2)
	if err != nil {
		t.Fatal(err)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(args2JSON))
	if _, err := register(req); err == nil {
		t.Errorf("Expected error,received: nil")
	}
	args2 = utils.NewServerRequest(utils.RegistrarSv1UnregisterRPCHosts, id, id)
	args2JSON, err = json.Marshal(args2)
	if err != nil {
		t.Fatal(err)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(args2JSON))
	if _, err := register(req); err == nil {
		t.Errorf("Expected error,received: nil")
	}
	engine.Cache = engine.NewCacheS(config.CgrConfig(), nil, nil)
}

type errRecorder struct{}

func (*errRecorder) Header() http.Header        { return make(http.Header) }
func (*errRecorder) Write([]byte) (int, error)  { return 0, io.EOF }
func (*errRecorder) WriteHeader(statusCode int) {}

func TestRegistrar(t *testing.T) {
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
	args := utils.NewServerRequest(utils.RegistrarSv1RegisterDispatcherHosts, raJSON, id)
	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/json_rpc", bytes.NewBuffer(argsJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:2356"

	Registrar(w, req)
	exp := "{\"id\":1,\"result\":\"OK\",\"error\":null}\n"
	if w.Body.String() != exp {
		t.Errorf("Expected: %q ,received: %q", exp, w.Body.String())
	}

	w = httptest.NewRecorder()
	Registrar(w, req)
	exp = "{\"id\":0,\"result\":null,\"error\":\"EOF\"}\n"
	if w.Body.String() != exp {
		t.Errorf("Expected: %q ,received: %q", exp, w.Body.String())
	}

	Registrar(new(errRecorder), req)
}

func TestLibRegistrarcRegister(t *testing.T) {
	req := &http.Request{
		Method:           "",
		URL:              nil,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             http.NoBody,
		GetBody:          nil,
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Host:             "",
		Form:             nil,
		PostForm:         nil,
		MultipartForm:    nil,
		Trailer:          nil,
		RemoteAddr:       "",
		RequestURI:       "",
		TLS:              nil,
		Cancel:           nil,
		Response:         nil,
	}
	result, err := register(req)
	expected := &json.RawMessage{}
	if reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
	if err == nil || err.Error() != "EOF" {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", "EOF", err)
	}
}

func TestGetConnPortHTTPJson(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	expected := "2014"
	result, err := getConnPort(cfg, rpcclient.BiRPCJSON, false)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
	if err != nil {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}

func TestGetConnPortBiRPCGOB(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	expected := ""
	result, err := getConnPort(cfg, rpcclient.BiRPCGOB, false)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
	if err == nil || err.Error() != "missing port in address" {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", "missing port in address", err)
	}
}

func TestRegisterRegistrarSv1UnregisterRPCHosts(t *testing.T) {
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
	args := utils.NewServerRequest(utils.RegistrarSv1UnregisterRPCHosts, raJSON, id)
	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/json_rpc", bytes.NewBuffer(argsJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:2356"
	engine.Cache = engine.NewCacheS(config.CgrConfig(), nil, nil)
	if rplyID, err := register(req); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(id, *rplyID) {
		t.Errorf("Expected: %q ,received: %q", string(id), string(*rplyID))
	}
}

func TestRegisterRegistrarSv1UnregisterRPCHostsError(t *testing.T) {
	ra := &UnregisterArgs{
		IDs:    []string{"Host1"},
		Tenant: "cgrates.org",
		Opts:   make(map[string]interface{}),
	}
	raJSON, err := json.Marshal([]interface{}{ra})
	id := json.RawMessage("1")
	if err != nil {
		t.Fatal(err)
	}
	args := utils.NewServerRequest(utils.RegistrarSv1UnregisterRPCHosts, raJSON, id)
	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/json_rpc", bytes.NewBuffer(argsJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:2356"
	cfg := config.NewDefaultCGRConfig()
	config.CgrConfig().RPCConns()["errCon"] = &config.RPCConn{
		Strategy: utils.MetaFirst,
		PoolSize: 1,
		Conns: []*config.RemoteHost{
			{
				ID:          "Host1",
				Address:     "127.0.0.1:9999",
				Transport:   "*json",
				Synchronous: true,
			},
		},
	}
	engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{})
	cfg.RPCConns()["errCon"] = config.CgrConfig().RPCConns()["errCon"]
	cfg.CacheCfg().ReplicationConns = []string{"errCon"}
	cfg.CacheCfg().Partitions[utils.CacheRPCConnections].Replicate = true
	engine.Cache = engine.NewCacheS(cfg, nil, nil)
	_, err = register(req)
	if err == nil || err != utils.ErrPartiallyExecuted {
		t.Fatal(err)
	}
	delete(config.CgrConfig().RPCConns(), "errCon")
	engine.Cache = engine.NewCacheS(config.CgrConfig(), nil, nil)
}

func TestRegisterRegistrarSv1RegisterRPCHosts(t *testing.T) {
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
	args := utils.NewServerRequest(utils.RegistrarSv1RegisterRPCHosts, raJSON, id)
	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/json_rpc", bytes.NewBuffer(argsJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:2356"
	if rplyID, err := register(req); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(id, *rplyID) {
		t.Errorf("Expected: %q ,received: %q", string(id), string(*rplyID))
	}
	engine.Cache = engine.NewCacheS(config.CgrConfig(), nil, nil)
}

func TestRegisterRegistrarSv1RegisterRPCHostsError(t *testing.T) {
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
	args := utils.NewServerRequest(utils.RegistrarSv1RegisterRPCHosts, raJSON, id)
	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:2080/json_rpc", bytes.NewBuffer(argsJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:3000"
	cfg := config.NewDefaultCGRConfig()
	config.CgrConfig().RPCConns()["errCon1"] = &config.RPCConn{
		Strategy: utils.MetaFirst,
		PoolSize: 1,
		Conns: []*config.RemoteHost{
			{
				ID:          "Host1",
				Address:     "127.0.0.1:9999",
				Transport:   "*json",
				Synchronous: true,
			},
		},
	}
	engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{})
	cfg.RPCConns()["errCon1"] = config.CgrConfig().RPCConns()["errCon1"]
	cfg.CacheCfg().ReplicationConns = []string{"errCon1"}
	cfg.CacheCfg().Partitions[utils.CacheRPCConnections].Replicate = true
	engine.Cache = engine.NewCacheS(cfg, nil, nil)
	_, err = register(req)
	if err == nil || err != utils.ErrPartiallyExecuted {
		t.Fatal(err)
	}
	delete(config.CgrConfig().RPCConns(), "errCon1")
	engine.Cache = engine.NewCacheS(config.CgrConfig(), nil, nil)
}
