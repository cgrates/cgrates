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
		PrometheusURL:       utils.StringPointer("/prometheus"),
		Ws_url:              utils.StringPointer("/ws"),
		Registrars_url:      utils.StringPointer("/randomUrl"),
		Freeswitch_cdrs_url: utils.StringPointer("/freeswitch_json"),
		Http_Cdrs:           utils.StringPointer("/cdr_http"),
		Use_basic_auth:      utils.BoolPointer(false),
		Auth_users:          utils.MapStringStringPointer(map[string]string{}),
	}
	expected := &HTTPCfg{
		HTTPJsonRPCURL:        "/jsonrpc",
		PrometheusURL:         "/prometheus",
		HTTPWSURL:             "/ws",
		RegistrarSURL:         "/randomUrl",
		HTTPFreeswitchCDRsURL: "/freeswitch_json",
		HTTPCDRsURL:           "/cdr_http",
		HTTPUseBasicAuth:      false,
		HTTPAuthUsers:         map[string]string{},
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
	cfgJsn := NewDefaultCGRConfig()
	if err := cfgJsn.httpCfg.loadFromJSONCfg(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.AsMapInterface(), cfgJsn.httpCfg.AsMapInterface()) {
		t.Errorf("Expected %+v \n, received %+v", expected.AsMapInterface(), cfgJsn.httpCfg.AsMapInterface())
	}

}

func TestHTTPCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"http": {},
}`
	eMap := map[string]any{
		utils.HTTPJsonRPCURLCfg:        "/jsonrpc",
		utils.RegistrarSURLCfg:         "/registrar",
		utils.PrometheusURLCfg:         "/prometheus",
		utils.HTTPWSURLCfg:             "/ws",
		utils.HTTPFreeswitchCDRsURLCfg: "/freeswitch_json",
		utils.HTTPCDRsURLCfg:           "/cdr_http",
		utils.HTTPUseBasicAuthCfg:      false,
		utils.HTTPAuthUsersCfg:         map[string]string{},
		utils.HTTPClientOptsCfg: map[string]any{
			utils.HTTPClientTLSClientConfigCfg:       false,
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
	} else if rcv := cgrCfg.httpCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestHTTPCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"http": {
		"json_rpc_url": "/rpc",					
		"ws_url": "",	
		"prometheus_url": "/metrics",	
		"use_basic_auth": true,					
		"auth_users": {"user1": "authenticated", "user2": "authenticated"},
	},
}`
	eMap := map[string]any{
		utils.HTTPJsonRPCURLCfg:        "/rpc",
		utils.RegistrarSURLCfg:         "/registrar",
		utils.PrometheusURLCfg:         "/metrics",
		utils.HTTPWSURLCfg:             "",
		utils.HTTPFreeswitchCDRsURLCfg: "/freeswitch_json",
		utils.HTTPCDRsURLCfg:           "/cdr_http",
		utils.HTTPUseBasicAuthCfg:      true,
		utils.HTTPAuthUsersCfg: map[string]string{
			"user1": "authenticated",
			"user2": "authenticated",
		},
		utils.HTTPClientOptsCfg: map[string]any{
			utils.HTTPClientTLSClientConfigCfg:       false,
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
	} else if rcv := cgrCfg.httpCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestHTTPCfgClone(t *testing.T) {
	ban := &HTTPCfg{
		HTTPJsonRPCURL:        "/jsonrpc",
		PrometheusURL:         "/prometheus",
		HTTPWSURL:             "/ws",
		RegistrarSURL:         "/randomUrl",
		HTTPFreeswitchCDRsURL: "/freeswitch_json",
		HTTPCDRsURL:           "/cdr_http",
		HTTPUseBasicAuth:      false,
		HTTPAuthUsers: map[string]string{
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
	if !reflect.DeepEqual(rcv.AsMapInterface(), ban.AsMapInterface()) {
		t.Errorf("Expected: %+v\nReceived: %+v", ban.AsMapInterface(),
			rcv.AsMapInterface())
	}
	if rcv.ClientOpts.MaxIdleConns = 50; ban.ClientOpts.MaxIdleConns != 100 {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.HTTPAuthUsers["user"] = ""; ban.HTTPAuthUsers["user"] != "pass" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestHTTPNewDialer(t *testing.T) {
	dialer := &net.Dialer{
		DualStack: true,
	}
	var jsnCfg *HTTPClientOptsJson
	if err := newDialer(dialer, jsnCfg); err != nil {
		t.Error(err)
	}
	jsnCfg = &HTTPClientOptsJson{
		DialTimeout:       utils.StringPointer("test"),
		DialFallbackDelay: utils.StringPointer("test2"),
	}
	if err := newDialer(dialer, jsnCfg); err == nil {
		t.Error(err)
	}
	jsnCfg = &HTTPClientOptsJson{
		DialFallbackDelay: utils.StringPointer("test2"),
	}
	if err := newDialer(dialer, jsnCfg); err == nil {
		t.Error(err)
	}
	jsnCfg = &HTTPClientOptsJson{
		DialKeepAlive: utils.StringPointer("test3"),
	}
	if err := newDialer(dialer, jsnCfg); err == nil {
		t.Error(err)
	}
}

func TestHTTPLoadTransportFromJSONCfg(t *testing.T) {

	var jsnCfg *HTTPClientOptsJson
	httpOpts := &http.Transport{}
	httpDialer := &net.Dialer{}
	if err := loadTransportFromJSONCfg(httpOpts, httpDialer, jsnCfg); err != nil {
		t.Error(err)
	}
	jsnCfg = &HTTPClientOptsJson{
		TLSHandshakeTimeout: utils.StringPointer("test"),
	}
	if err := loadTransportFromJSONCfg(httpOpts, httpDialer, jsnCfg); err == nil {
		t.Error(err)
	}
	jsnCfg = &HTTPClientOptsJson{
		IdleConnTimeout: utils.StringPointer("test2"),
	}
	if err := loadTransportFromJSONCfg(httpOpts, httpDialer, jsnCfg); err == nil {
		t.Error(err)
	}
	jsnCfg = &HTTPClientOptsJson{
		ResponseHeaderTimeout: utils.StringPointer("test3"),
	}
	if err := loadTransportFromJSONCfg(httpOpts, httpDialer, jsnCfg); err == nil {
		t.Error(err)
	}
	jsnCfg = &HTTPClientOptsJson{
		ExpectContinueTimeout: utils.StringPointer("test4"),
	}
	if err := loadTransportFromJSONCfg(httpOpts, httpDialer, jsnCfg); err == nil {
		t.Error(err)
	}

}

func TestHTTPLoadFromJSONCfg(t *testing.T) {
	var jsonHTTPCfg *HTTPJsonCfg
	httpcfg := &HTTPCfg{}
	if err := httpcfg.loadFromJSONCfg(jsonHTTPCfg); err != nil {
		t.Error(err)
	}

	jsnHTTPCfg := &HTTPJsonCfg{
		Client_opts: &HTTPClientOptsJson{
			DialTimeout: utils.StringPointer("test")},
	}
	httpcg := &HTTPCfg{
		dialer:     &net.Dialer{},
		ClientOpts: &http.Transport{},
	}
	if err := httpcg.loadFromJSONCfg(jsnHTTPCfg); err == nil {
		t.Error(err)
	}
}
