/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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
		JSONRPCURL:        utils.StringPointer("/jsonrpc"),
		WSURL:             utils.StringPointer("/ws"),
		RegistrarsURL:     utils.StringPointer("/randomUrl"),
		FreeswitchCDRsURL: utils.StringPointer("/freeswitch_json"),
		HTTPCDRs:          utils.StringPointer("/cdr_http"),
		UseBasicAuth:      utils.BoolPointer(false),
		AuthUsers:         utils.MapStringStringPointer(map[string]string{}),
	}
	expected := &HTTPCfg{
		JsonRPCURL:        "/jsonrpc",
		WSURL:             "/ws",
		RegistrarSURL:     "/randomUrl",
		FreeswitchCDRsURL: "/freeswitch_json",
		CDRsURL:           "/cdr_http",
		PprofPath:         "/debug/pprof/",
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
	if err := cfgJsn.httpCfg.loadFromJSONCfg(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.AsMapInterface(),
		cfgJsn.httpCfg.AsMapInterface()) {
		t.Errorf("Expected %+v \n, received %+v",
			expected.AsMapInterface(),
			cfgJsn.httpCfg.AsMapInterface())
	}

	cfgJSONStr = nil
	if err := cfgJsn.httpCfg.loadFromJSONCfg(cfgJSONStr); err != nil {
		t.Error(err)
	}
}

func TestHTTPCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"http": {},
}`
	eMap := map[string]any{
		utils.HTTPJsonRPCURLCfg:        "/jsonrpc",
		utils.RegistrarSURLCfg:         "/registrar",
		utils.HTTPWSURLCfg:             "/ws",
		utils.HTTPFreeswitchCDRsURLCfg: "/freeswitch_json",
		utils.HTTPCDRsURLCfg:           "/cdr_http",
		utils.PprofPathCfg:             "/debug/pprof/",
		utils.HTTPUseBasicAuthCfg:      false,
		utils.HTTPAuthUsersCfg:         map[string]string{},
		utils.HTTPClientOptsCfg: map[string]any{
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
	eMap := map[string]any{
		utils.HTTPJsonRPCURLCfg:        "/rpc",
		utils.RegistrarSURLCfg:         "/registrar",
		utils.HTTPWSURLCfg:             "",
		utils.HTTPFreeswitchCDRsURLCfg: "/freeswitch_json",
		utils.HTTPCDRsURLCfg:           "/cdr_http",
		utils.PprofPathCfg:             "/debug/pprof/",
		utils.HTTPUseBasicAuthCfg:      true,
		utils.HTTPAuthUsersCfg: map[string]string{
			"user1": "authenticated",
			"user2": "authenticated",
		},
		utils.HTTPClientOptsCfg: map[string]any{
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
	} else if rcv := cgrCfg.httpCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
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
	if !reflect.DeepEqual(rcv.AsMapInterface(), ban.AsMapInterface()) {
		t.Errorf("Expected: %+v\nReceived: %+v", ban.AsMapInterface(),
			rcv.AsMapInterface())
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
		JSONRPCURL:        utils.StringPointer("JsonRpcUrl2"),
		RegistrarsURL:     utils.StringPointer("RegistrarSUrl2"),
		WSURL:             utils.StringPointer("WsUrl2"),
		FreeswitchCDRsURL: utils.StringPointer("FsCdrsUrl2"),
		HTTPCDRs:          utils.StringPointer("CdrsUrl2"),
		UseBasicAuth:      utils.BoolPointer(false),
		AuthUsers: &map[string]string{
			"User2": "passUser2",
		},
		ClientOpts: &HTTPClientOptsJson{
			MaxIdleConns: utils.IntPointer(100),
		},
	}

	rcv := diffHTTPJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &HTTPJsonCfg{
		ClientOpts: &HTTPClientOptsJson{},
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
	if !reflect.DeepEqual(rcv.AsMapInterface(), exp.AsMapInterface()) {
		t.Errorf("Expected %v \n but received \n %v",
			utils.ToJSON(exp.AsMapInterface()),
			utils.ToJSON(rcv.AsMapInterface()))
	}
}

func TestNewDialerJsonCfgNil(t *testing.T) {
	var jsnCfg *HTTPClientOptsJson

	nDialer := NewDefaultCGRConfig().httpCfg.dialer
	nDialer.DualStack = false
	if err := newDialer(nDialer, jsnCfg); err != nil {
		t.Errorf("Expected error <nil> \n but received error <%v>", err)
	} else if nDialer.DualStack {
		t.Errorf("Dialer DualStack shouldnt have changed, was <false>, now is <%v>",
			nDialer.DualStack)
	}

}
func TestNewDialerJsonCfgDialTimeout(t *testing.T) {
	jsnCfg := &HTTPClientOptsJson{
		DialTimeout: utils.StringPointer("invalid time"),
	}

	nDialer := NewDefaultCGRConfig().httpCfg.dialer
	expErr := `time: invalid duration "invalid time"`
	if err := newDialer(nDialer, jsnCfg); err.Error() != expErr {
		t.Errorf("Expected error <%v> \n but received error <%v>", expErr, err)
	}

}
func TestNewDialerJsonCfgDialFallbackDelay(t *testing.T) {
	jsnCfg := &HTTPClientOptsJson{
		DialFallbackDelay: utils.StringPointer("invalid time"),
	}

	nDialer := NewDefaultCGRConfig().httpCfg.dialer
	expErr := `time: invalid duration "invalid time"`
	if err := newDialer(nDialer, jsnCfg); err.Error() != expErr {
		t.Errorf("Expected error <%v> \n but received error <%v>", expErr, err)
	}

}
func TestNewDialerJsonCfgDialKeepAlive(t *testing.T) {
	jsnCfg := &HTTPClientOptsJson{
		DialKeepAlive: utils.StringPointer("invalid time"),
	}

	nDialer := NewDefaultCGRConfig().httpCfg.dialer
	expErr := `time: invalid duration "invalid time"`
	if err := newDialer(nDialer, jsnCfg); err.Error() != expErr {
		t.Errorf("Expected error <%v> \n but received error <%v>", expErr, err)
	}

}

func TestLoadTransportFromJSONCfgNilJson(t *testing.T) {

	EqualHttpOpts := func(httpOpts1, httpOpts2 *http.Transport) bool {
		if httpOpts1.TLSHandshakeTimeout != httpOpts2.TLSHandshakeTimeout {
			return false
		}
		if httpOpts1.DisableKeepAlives != httpOpts2.DisableKeepAlives {
			return false
		}
		if httpOpts1.DisableCompression != httpOpts2.DisableCompression {
			return false
		}
		if httpOpts1.MaxIdleConns != httpOpts2.MaxIdleConns {
			return false
		}
		if httpOpts1.MaxIdleConnsPerHost != httpOpts2.MaxIdleConnsPerHost {
			return false
		}
		if httpOpts1.IdleConnTimeout != httpOpts2.IdleConnTimeout {
			return false
		}
		if httpOpts1.MaxConnsPerHost != httpOpts2.MaxConnsPerHost {
			return false
		}
		return true
	}

	httpOpts := &http.Transport{
		TLSHandshakeTimeout: time.Duration(2),
		DisableKeepAlives:   false,
		DisableCompression:  false,
		MaxIdleConns:        2,
		MaxIdleConnsPerHost: 2,
		MaxConnsPerHost:     2,
		IdleConnTimeout:     time.Duration(2),
	}

	httpOptsCopy := httpOpts

	var jsnCfg *HTTPClientOptsJson

	if err := loadTransportFromJSONCfg(httpOpts,
		NewDefaultCGRConfig().httpCfg.dialer,
		jsnCfg); err != nil {
		t.Errorf("Expected error <nil> \n but received error <%v>", err)
	} else if !EqualHttpOpts(httpOpts, httpOptsCopy) {
		t.Errorf("Expected HttpOpts not to change, was <%+v>,\n Now is <%+v>",
			httpOpts, httpOptsCopy)
	}

}

func TestLoadTransportFromJSONCfgTLSHandshakeTimeout(t *testing.T) {
	httpOpts := NewDefaultCGRConfig().httpCfg.ClientOpts

	jsnCfg := &HTTPClientOptsJson{
		TLSHandshakeTimeout: utils.StringPointer("invalid time"),
	}
	expErr := `time: invalid duration "invalid time"`
	if err := loadTransportFromJSONCfg(httpOpts,
		NewDefaultCGRConfig().httpCfg.dialer,
		jsnCfg); err.Error() != expErr {
		t.Errorf("Expected error <%v> \n but received error <%v>", expErr, err.Error())
	}

}
func TestLoadTransportFromJSONCfgIdleConnTimeout(t *testing.T) {
	httpOpts := NewDefaultCGRConfig().httpCfg.ClientOpts

	jsnCfg := &HTTPClientOptsJson{
		IdleConnTimeout: utils.StringPointer("invalid time"),
	}
	expErr := `time: invalid duration "invalid time"`
	if err := loadTransportFromJSONCfg(httpOpts,
		NewDefaultCGRConfig().httpCfg.dialer,
		jsnCfg); err.Error() != expErr {
		t.Errorf("Expected error <%v> \n but received error <%v>", expErr, err.Error())
	}

}
func TestLoadTransportFromJSONCfgResponseHeaderTimeout(t *testing.T) {
	httpOpts := NewDefaultCGRConfig().httpCfg.ClientOpts

	jsnCfg := &HTTPClientOptsJson{
		ResponseHeaderTimeout: utils.StringPointer("invalid time"),
	}
	expErr := `time: invalid duration "invalid time"`
	if err := loadTransportFromJSONCfg(httpOpts,
		NewDefaultCGRConfig().httpCfg.dialer,
		jsnCfg); err.Error() != expErr {
		t.Errorf("Expected error <%v> \n but received error <%v>", expErr, err.Error())
	}
}
func TestLoadTransportFromJSONCfgExpectContinueTimeout(t *testing.T) {
	httpOpts := NewDefaultCGRConfig().httpCfg.ClientOpts

	jsnCfg := &HTTPClientOptsJson{
		ExpectContinueTimeout: utils.StringPointer("invalid time"),
	}
	expErr := `time: invalid duration "invalid time"`
	if err := loadTransportFromJSONCfg(httpOpts,
		NewDefaultCGRConfig().httpCfg.dialer,
		jsnCfg); err.Error() != expErr {
		t.Errorf("Expected error <%v> \n but received error <%v>", expErr, err.Error())
	}
}

func TestHTTPCfgloadFromJsonCfgClientOptsErr(t *testing.T) {
	cfgJSONStr := &HTTPJsonCfg{
		ClientOpts: &HTTPClientOptsJson{
			DialTimeout: utils.StringPointer("invalid value"),
		},
	}
	expErr := `time: invalid duration "invalid value"`
	cfgJsn := NewDefaultCGRConfig()
	if err := cfgJsn.httpCfg.loadFromJSONCfg(cfgJSONStr); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err.Error())
	}
}

func TestDiffHTTPClientOptsJsonCfgDialer(t *testing.T) {
	var d *HTTPClientOptsJson

	v1 := &net.Dialer{
		Timeout:       time.Duration(2),
		FallbackDelay: time.Duration(2),
		KeepAlive:     time.Duration(2),
	}

	v2 := &net.Dialer{
		Timeout:       time.Duration(3),
		FallbackDelay: time.Duration(3),
		KeepAlive:     time.Duration(3),
	}

	expected := &HTTPClientOptsJson{
		DialTimeout:       utils.StringPointer("3ns"),
		DialFallbackDelay: utils.StringPointer("3ns"),
		DialKeepAlive:     utils.StringPointer("3ns"),
	}

	rcv := diffHTTPClientOptsJsonCfgDialer(d, v1, v2)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &HTTPClientOptsJson{}

	rcv = diffHTTPClientOptsJsonCfgDialer(d, v1, v2_2)
	if !reflect.DeepEqual(expected2, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestDiffHTTPClientOptsJsonCfg(t *testing.T) {
	var d *HTTPClientOptsJson

	v1 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		TLSHandshakeTimeout:   time.Duration(2),
		DisableKeepAlives:     false,
		DisableCompression:    false,
		MaxIdleConns:          2,
		MaxIdleConnsPerHost:   2,
		MaxConnsPerHost:       2,
		IdleConnTimeout:       time.Duration(2),
		ResponseHeaderTimeout: time.Duration(2),
		ExpectContinueTimeout: time.Duration(2),
		ForceAttemptHTTP2:     false,
	}

	v2 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSHandshakeTimeout:   time.Duration(3),
		DisableKeepAlives:     true,
		DisableCompression:    true,
		MaxIdleConns:          3,
		MaxIdleConnsPerHost:   3,
		MaxConnsPerHost:       3,
		IdleConnTimeout:       time.Duration(3),
		ResponseHeaderTimeout: time.Duration(3),
		ExpectContinueTimeout: time.Duration(3),
		ForceAttemptHTTP2:     true,
	}

	expected := &HTTPClientOptsJson{
		SkipTLSVerification:   utils.BoolPointer(true),
		TLSHandshakeTimeout:   utils.StringPointer("3ns"),
		DisableKeepAlives:     utils.BoolPointer(true),
		DisableCompression:    utils.BoolPointer(true),
		MaxIdleConns:          utils.IntPointer(3),
		MaxIdleConnsPerHost:   utils.IntPointer(3),
		MaxConnsPerHost:       utils.IntPointer(3),
		IdleConnTimeout:       utils.StringPointer("3ns"),
		ResponseHeaderTimeout: utils.StringPointer("3ns"),
		ExpectContinueTimeout: utils.StringPointer("3ns"),
		ForceAttemptHTTP2:     utils.BoolPointer(true),
	}

	rcv := diffHTTPClientOptsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &HTTPClientOptsJson{}

	rcv = diffHTTPClientOptsJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(expected2, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}
