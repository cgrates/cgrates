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
	"net"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestHTTPCfgloadFromJsonCfg(t *testing.T) {
	cfgJSONStr := &HTTPJsonCfg{
		Json_rpc_url:        utils.StringPointer("/jsonrpc"),
		Ws_url:              utils.StringPointer("/ws"),
		Registrars_url:      utils.StringPointer("/randomUrl"),
		Freeswitch_cdrs_url: utils.StringPointer("/freeswitch_json"),
		Http_Cdrs:           utils.StringPointer("/cdr_http"),
		Use_basic_auth:      utils.BoolPointer(false),
		Auth_users:          utils.MapStringStringPointer(map[string]string{}),
	}
	expected := &HTTPCfg{
		JsonRPCURL:        "/jsonrpc",
		WSURL:             "/ws",
		RegistrarSURL:     "/randomUrl",
		FreeswitchCDRsURL: "/freeswitch_json",
		CDRsURL:           "/cdr_http",
		UseBasicAuth:      false,
		AuthUsers:         map[string]string{},
		ClientOpts: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			TLSHandshakeTimeout:   10 * time.Second,
			DisableKeepAlives:     false,
			DisableCompression:    false,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   2,
			MaxConnsPerHost:       0,
			IdleConnTimeout:       90 * time.Second,
			ResponseHeaderTimeout: 0,
			ExpectContinueTimeout: 0,
			ForceAttemptHTTP2:     true,
		},
		dialer: &net.Dialer{
			Timeout:       30 * time.Second,
			FallbackDelay: 300 * time.Millisecond,
			KeepAlive:     30 * time.Second,
			DualStack:     true,
		},
	}
	expected.ClientOpts.DialContext = expected.dialer.DialContext
	cfgJsn := NewDefaultCGRConfig()
	if err = cfgJsn.httpCfg.loadFromJSONCfg(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.AsMapInterface(utils.InfieldSep),
		cfgJsn.httpCfg.AsMapInterface(utils.InfieldSep)) {
		t.Errorf("Expected %+v \n, received %+v",
			expected.AsMapInterface(utils.InfieldSep),
			cfgJsn.httpCfg.AsMapInterface(utils.InfieldSep))
	}

	cfgJSONStr = nil
	if err = cfgJsn.httpCfg.loadFromJSONCfg(cfgJSONStr); err != nil {
		t.Error(err)
	}
}

func TestHTTPCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"http": {},
}`
	eMap := map[string]interface{}{
		utils.HTTPJsonRPCURLCfg:        "/jsonrpc",
		utils.RegistrarSURLCfg:         "/registrar",
		utils.HTTPWSURLCfg:             "/ws",
		utils.HTTPFreeswitchCDRsURLCfg: "/freeswitch_json",
		utils.HTTPCDRsURLCfg:           "/cdr_http",
		utils.HTTPUseBasicAuthCfg:      false,
		utils.HTTPAuthUsersCfg:         map[string]string{},
		utils.HTTPClientOptsCfg: map[string]interface{}{
			utils.HTTPClientSkipTLSVerificationCfg:   false,
			utils.HTTPClientTLSHandshakeTimeoutCfg:   "10s",
			utils.HTTPClientDisableKeepAlivesCfg:     false,
			utils.HTTPClientDisableCompressionCfg:    false,
			utils.HTTPClientMaxIdleConnsCfg:          100,
			utils.HTTPClientMaxIdleConnsPerHostCfg:   2,
			utils.HTTPClientMaxConnsPerHostCfg:       0,
			utils.HTTPClientIdleConnTimeoutCfg:       "1m30s",
			utils.HTTPClientResponseHeaderTimeoutCfg: "0s",
			utils.HTTPClientExpectContinueTimeoutCfg: "0s",
			utils.HTTPClientForceAttemptHTTP2Cfg:     true,
			utils.HTTPClientDialTimeoutCfg:           "30s",
			utils.HTTPClientDialFallbackDelayCfg:     "300ms",
			utils.HTTPClientDialKeepAliveCfg:         "30s",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.httpCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
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
		utils.HTTPJsonRPCURLCfg:        "/rpc",
		utils.RegistrarSURLCfg:         "/registrar",
		utils.HTTPWSURLCfg:             "",
		utils.HTTPFreeswitchCDRsURLCfg: "/freeswitch_json",
		utils.HTTPCDRsURLCfg:           "/cdr_http",
		utils.HTTPUseBasicAuthCfg:      true,
		utils.HTTPAuthUsersCfg: map[string]string{
			"user1": "authenticated",
			"user2": "authenticated",
		},
		utils.HTTPClientOptsCfg: map[string]interface{}{
			utils.HTTPClientSkipTLSVerificationCfg:   false,
			utils.HTTPClientTLSHandshakeTimeoutCfg:   "10s",
			utils.HTTPClientDisableKeepAlivesCfg:     false,
			utils.HTTPClientDisableCompressionCfg:    false,
			utils.HTTPClientMaxIdleConnsCfg:          100,
			utils.HTTPClientMaxIdleConnsPerHostCfg:   2,
			utils.HTTPClientMaxConnsPerHostCfg:       0,
			utils.HTTPClientIdleConnTimeoutCfg:       "1m30s",
			utils.HTTPClientResponseHeaderTimeoutCfg: "0s",
			utils.HTTPClientExpectContinueTimeoutCfg: "0s",
			utils.HTTPClientForceAttemptHTTP2Cfg:     true,
			utils.HTTPClientDialTimeoutCfg:           "30s",
			utils.HTTPClientDialFallbackDelayCfg:     "300ms",
			utils.HTTPClientDialKeepAliveCfg:         "30s",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.httpCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestHTTPCfgClone(t *testing.T) {
	ban := &HTTPCfg{
		JsonRPCURL:        "/jsonrpc",
		WSURL:             "/ws",
		RegistrarSURL:     "/randomUrl",
		FreeswitchCDRsURL: "/freeswitch_json",
		CDRsURL:           "/cdr_http",
		UseBasicAuth:      false,
		AuthUsers: map[string]string{
			"user": "pass",
		},
		ClientOpts: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			TLSHandshakeTimeout:   10 * time.Second,
			DisableKeepAlives:     false,
			DisableCompression:    false,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   2,
			MaxConnsPerHost:       0,
			IdleConnTimeout:       90 * time.Second,
			ResponseHeaderTimeout: 0,
			ExpectContinueTimeout: 0,
			ForceAttemptHTTP2:     true,
		},
		dialer: &net.Dialer{
			Timeout:       30 * time.Second,
			FallbackDelay: 300 * time.Millisecond,
			KeepAlive:     30 * time.Second,
			DualStack:     true,
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(rcv.AsMapInterface(utils.InfieldSep), ban.AsMapInterface(utils.InfieldSep)) {
		t.Errorf("Expected: %+v\nReceived: %+v", ban.AsMapInterface(utils.InfieldSep),
			rcv.AsMapInterface(utils.InfieldSep))
	}
	if rcv.ClientOpts.MaxIdleConns = 50; ban.ClientOpts.MaxIdleConns != 100 {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.AuthUsers["user"] = ""; ban.AuthUsers["user"] != "pass" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffHTTPJsonCfg(t *testing.T) {
	var d *HTTPJsonCfg

	v1 := &HTTPCfg{
		JsonRPCURL:        "JsonRpcUrl",
		RegistrarSURL:     "RegistrarSUrl",
		WSURL:             "WSUrl",
		FreeswitchCDRsURL: "FsCdrsUrl",
		CDRsURL:           "CdrsUrl",
		UseBasicAuth:      true,
		AuthUsers: map[string]string{
			"User1": "passUser1",
		},
		ClientOpts: &http.Transport{},
		dialer:     &net.Dialer{},
	}

	v2 := &HTTPCfg{
		JsonRPCURL:        "JsonRpcUrl2",
		RegistrarSURL:     "RegistrarSUrl2",
		WSURL:             "WsUrl2",
		FreeswitchCDRsURL: "FsCdrsUrl2",
		CDRsURL:           "CdrsUrl2",
		UseBasicAuth:      false,
		AuthUsers: map[string]string{
			"User2": "passUser2",
		},
		ClientOpts: &http.Transport{
			MaxIdleConns: 100,
		},
		dialer: &net.Dialer{},
	}

	expected := &HTTPJsonCfg{
		Json_rpc_url:        utils.StringPointer("JsonRpcUrl2"),
		Registrars_url:      utils.StringPointer("RegistrarSUrl2"),
		Ws_url:              utils.StringPointer("WsUrl2"),
		Freeswitch_cdrs_url: utils.StringPointer("FsCdrsUrl2"),
		Http_Cdrs:           utils.StringPointer("CdrsUrl2"),
		Use_basic_auth:      utils.BoolPointer(false),
		Auth_users: &map[string]string{
			"User2": "passUser2",
		},
		Client_opts: &HTTPClientOptsJson{
			MaxIdleConns: utils.IntPointer(100),
		},
	}

	rcv := diffHTTPJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &HTTPJsonCfg{
		Client_opts: &HTTPClientOptsJson{},
	}
	rcv = diffHTTPJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

}

func TestHttpCfgCloneSection(t *testing.T) {
	httpCfg := &HTTPCfg{
		JsonRPCURL:        "JsonRpcUrl",
		RegistrarSURL:     "RegistrarSUrl",
		WSURL:             "WSUrl",
		FreeswitchCDRsURL: "FsCdrsUrl",
		CDRsURL:           "CdrsUrl",
		UseBasicAuth:      true,
		AuthUsers: map[string]string{
			"User1": "passUser1",
		},
		ClientOpts: &http.Transport{},
		dialer:     &net.Dialer{},
	}

	exp := &HTTPCfg{
		JsonRPCURL:        "JsonRpcUrl",
		RegistrarSURL:     "RegistrarSUrl",
		WSURL:             "WSUrl",
		FreeswitchCDRsURL: "FsCdrsUrl",
		CDRsURL:           "CdrsUrl",
		UseBasicAuth:      true,
		AuthUsers: map[string]string{
			"User1": "passUser1",
		},
		ClientOpts: &http.Transport{},
		dialer:     &net.Dialer{},
	}

	rcv := httpCfg.CloneSection()
	if !reflect.DeepEqual(rcv.AsMapInterface(utils.InfieldSep),
		exp.AsMapInterface(utils.InfieldSep)) {
		t.Errorf("Expected %v \n but received \n %v",
			utils.ToJSON(exp.AsMapInterface(utils.InfieldSep)),
			utils.ToJSON(rcv.AsMapInterface(utils.InfieldSep)))
	}
}
