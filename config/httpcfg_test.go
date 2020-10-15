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
package config

import (
	"crypto/tls"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestHTTPCfgloadFromJsonCfg(t *testing.T) {
	cfgJSONStr := &HTTPJsonCfg{
		Json_rpc_url:              utils.StringPointer("/jsonrpc"),
		Ws_url:                    utils.StringPointer("/ws"),
		Dispatchers_registrar_url: utils.StringPointer("/randomUrl"),
		Freeswitch_cdrs_url:       utils.StringPointer("/freeswitch_json"),
		Http_Cdrs:                 utils.StringPointer("/cdr_http"),
		Use_basic_auth:            utils.BoolPointer(false),
		Auth_users:                utils.MapStringStringPointer(map[string]string{}),
	}
	expected := &HTTPCfg{
		HTTPJsonRPCURL:          "/jsonrpc",
		HTTPWSURL:               "/ws",
		DispatchersRegistrarURL: "/randomUrl",
		HTTPFreeswitchCDRsURL:   "/freeswitch_json",
		HTTPCDRsURL:             "/cdr_http",
		HTTPUseBasicAuth:        false,
		HTTPAuthUsers:           map[string]string{},
		ClientOpts: map[string]interface{}{
			utils.HTTPClientTLSClientConfigCfg:       false,
			utils.HTTPClientTLSHandshakeTimeoutCfg:   "10s",
			utils.HTTPClientDisableKeepAlivesCfg:     false,
			utils.HTTPClientDisableCompressionCfg:    false,
			utils.HTTPClientMaxIdleConnsCfg:          100.,
			utils.HTTPClientMaxIdleConnsPerHostCfg:   2.,
			utils.HTTPClientMaxConnsPerHostCfg:       0.,
			utils.HTTPClientIdleConnTimeoutCfg:       "90s",
			utils.HTTPClientResponseHeaderTimeoutCfg: "0",
			utils.HTTPClientExpectContinueTimeoutCfg: "0",
			utils.HTTPClientForceAttemptHTTP2Cfg:     true,
			utils.HTTPClientDialTimeoutCfg:           "30s",
			utils.HTTPClientDialFallbackDelayCfg:     "300ms",
			utils.HTTPClientDialKeepAliveCfg:         "30s",
		},
	}
	if cfgJsn, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = cfgJsn.httpCfg.loadFromJsonCfg(cfgJSONStr); err != nil {
		t.Error(err)
	} else if cfgJsn.httpCfg.transport = nil; !reflect.DeepEqual(expected, cfgJsn.httpCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cfgJsn.httpCfg))
	}
}

func TestHTTPCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"http": {},
}`
	eMap := map[string]interface{}{
		utils.HTTPJsonRPCURLCfg:          "/jsonrpc",
		utils.DispatchersRegistrarURLCfg: "/dispatchers_registrar",
		utils.HTTPWSURLCfg:               "/ws",
		utils.HTTPFreeswitchCDRsURLCfg:   "/freeswitch_json",
		utils.HTTPCDRsURLCfg:             "/cdr_http",
		utils.HTTPUseBasicAuthCfg:        false,
		utils.HTTPAuthUsersCfg:           map[string]string{},
		utils.HTTPClientOptsCfg: map[string]interface{}{
			utils.HTTPClientTLSClientConfigCfg:       false,
			utils.HTTPClientTLSHandshakeTimeoutCfg:   "10s",
			utils.HTTPClientDisableKeepAlivesCfg:     false,
			utils.HTTPClientDisableCompressionCfg:    false,
			utils.HTTPClientMaxIdleConnsCfg:          100.,
			utils.HTTPClientMaxIdleConnsPerHostCfg:   2.,
			utils.HTTPClientMaxConnsPerHostCfg:       0.,
			utils.HTTPClientIdleConnTimeoutCfg:       "90s",
			utils.HTTPClientResponseHeaderTimeoutCfg: "0",
			utils.HTTPClientExpectContinueTimeoutCfg: "0",
			utils.HTTPClientForceAttemptHTTP2Cfg:     true,
			utils.HTTPClientDialTimeoutCfg:           "30s",
			utils.HTTPClientDialFallbackDelayCfg:     "300ms",
			utils.HTTPClientDialKeepAliveCfg:         "30s",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.httpCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestHTTPCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"http": {
       "json_rpc_url": "/rpc",					
	   "ws_url": "",	
	   "use_basic_auth": true,					
	   "auth_users": {"user1": "authenticated", "user2": "authenticated"},
     },
}`
	eMap := map[string]interface{}{
		utils.HTTPJsonRPCURLCfg:          "/rpc",
		utils.DispatchersRegistrarURLCfg: "/dispatchers_registrar",
		utils.HTTPWSURLCfg:               "",
		utils.HTTPFreeswitchCDRsURLCfg:   "/freeswitch_json",
		utils.HTTPCDRsURLCfg:             "/cdr_http",
		utils.HTTPUseBasicAuthCfg:        true,
		utils.HTTPAuthUsersCfg: map[string]string{
			"user1": "authenticated",
			"user2": "authenticated",
		},
		utils.HTTPClientOptsCfg: map[string]interface{}{
			utils.HTTPClientTLSClientConfigCfg:       false,
			utils.HTTPClientTLSHandshakeTimeoutCfg:   "10s",
			utils.HTTPClientDisableKeepAlivesCfg:     false,
			utils.HTTPClientDisableCompressionCfg:    false,
			utils.HTTPClientMaxIdleConnsCfg:          100.,
			utils.HTTPClientMaxIdleConnsPerHostCfg:   2.,
			utils.HTTPClientMaxConnsPerHostCfg:       0.,
			utils.HTTPClientIdleConnTimeoutCfg:       "90s",
			utils.HTTPClientResponseHeaderTimeoutCfg: "0",
			utils.HTTPClientExpectContinueTimeoutCfg: "0",
			utils.HTTPClientForceAttemptHTTP2Cfg:     true,
			utils.HTTPClientDialTimeoutCfg:           "30s",
			utils.HTTPClientDialFallbackDelayCfg:     "300ms",
			utils.HTTPClientDialKeepAliveCfg:         "30s",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.httpCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestHTTPCfgGetDefaultHTTPTransort(t *testing.T) {
	httpCfg := new(HTTPCfg)
	if rply := httpCfg.GetDefaultHTTPTransort(); rply != nil {
		t.Errorf("Expected %+v, received %+v", nil, rply)
	}
}

func TestHTTPCfgInitTransport(t *testing.T) {
	httpCfg := &HTTPCfg{
		ClientOpts: map[string]interface{}{
			utils.HTTPClientTLSClientConfigCfg:       false,
			utils.HTTPClientTLSHandshakeTimeoutCfg:   "10s",
			utils.HTTPClientDisableKeepAlivesCfg:     false,
			utils.HTTPClientDisableCompressionCfg:    false,
			utils.HTTPClientMaxIdleConnsCfg:          100.,
			utils.HTTPClientMaxIdleConnsPerHostCfg:   2.,
			utils.HTTPClientMaxConnsPerHostCfg:       0.,
			utils.HTTPClientIdleConnTimeoutCfg:       "90s",
			utils.HTTPClientResponseHeaderTimeoutCfg: "0",
			utils.HTTPClientExpectContinueTimeoutCfg: "0",
			utils.HTTPClientForceAttemptHTTP2Cfg:     true,
			utils.HTTPClientDialTimeoutCfg:           "30s",
			utils.HTTPClientDialFallbackDelayCfg:     "300ms",
			utils.HTTPClientDialKeepAliveCfg:         "30s",
		},
	}
	// the dial options are not included
	checkTransport := func(t1, t2 *http.Transport) bool {
		return t1 != nil && t2 != nil &&
			t1.TLSClientConfig.InsecureSkipVerify == t2.TLSClientConfig.InsecureSkipVerify &&
			t1.TLSHandshakeTimeout == t2.TLSHandshakeTimeout &&
			t1.DisableKeepAlives == t2.DisableKeepAlives &&
			t1.DisableCompression == t2.DisableCompression &&
			t1.MaxIdleConns == t2.MaxIdleConns &&
			t1.MaxIdleConnsPerHost == t2.MaxIdleConnsPerHost &&
			t1.MaxConnsPerHost == t2.MaxConnsPerHost &&
			t1.IdleConnTimeout == t2.IdleConnTimeout &&
			t1.ResponseHeaderTimeout == t2.ResponseHeaderTimeout &&
			t1.ExpectContinueTimeout == t2.ExpectContinueTimeout &&
			t1.ForceAttemptHTTP2 == t2.ForceAttemptHTTP2
	}
	expTransport := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 2,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   true,
	}
	if err := httpCfg.initTransport(); err != nil {
		t.Fatal(err)
	} else if !checkTransport(expTransport, httpCfg.GetDefaultHTTPTransort()) {
		t.Errorf("Expected %+v, received %+v", expTransport, httpCfg.GetDefaultHTTPTransort())
	}

	httpCfg.ClientOpts[utils.HTTPClientDialKeepAliveCfg] = "30as"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientDialFallbackDelayCfg] = "300ams"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientDialTimeoutCfg] = "30as"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientForceAttemptHTTP2Cfg] = "string"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientExpectContinueTimeoutCfg] = "0a"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientResponseHeaderTimeoutCfg] = "0a"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientIdleConnTimeoutCfg] = "90as"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientMaxConnsPerHostCfg] = "not a number"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientMaxIdleConnsPerHostCfg] = "not a number"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientMaxIdleConnsCfg] = "not a number"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientDisableCompressionCfg] = "string"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientDisableKeepAlivesCfg] = "string"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientTLSHandshakeTimeoutCfg] = "10as"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	httpCfg.ClientOpts[utils.HTTPClientTLSClientConfigCfg] = "string"
	if err := httpCfg.initTransport(); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
}
