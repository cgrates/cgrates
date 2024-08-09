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

	"github.com/cgrates/cgrates/utils"
)

// HTTPCfg is the HTTP config section
type HTTPCfg struct {
	HTTPJsonRPCURL        string            // JSON RPC relative URL ("" to disable)
	RegistrarSURL         string            // registrar service relative URL
	PrometheusURL         string            // endpoint for prometheus metrics ("" to disable)
	HTTPWSURL             string            // WebSocket relative URL ("" to disable)
	HTTPFreeswitchCDRsURL string            // Freeswitch CDRS relative URL ("" to disable)
	HTTPCDRsURL           string            // CDRS relative URL ("" to disable)
	PprofPath             string            // runtime profiling url path ("" to disable)
	HTTPUseBasicAuth      bool              // Use basic auth for HTTP API
	HTTPAuthUsers         map[string]string // Basic auth user:password map (base64 passwords)
	ClientOpts            *http.Transport
	dialer                *net.Dialer
}

func newDialer(dialer *net.Dialer, jsnCfg *HTTPClientOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.DialTimeout != nil {
		if dialer.Timeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.DialTimeout); err != nil {
			return
		}
	}
	if jsnCfg.DialFallbackDelay != nil {
		if dialer.FallbackDelay, err = utils.ParseDurationWithNanosecs(*jsnCfg.DialFallbackDelay); err != nil {
			return
		}
	}
	if jsnCfg.DialKeepAlive != nil {
		if dialer.KeepAlive, err = utils.ParseDurationWithNanosecs(*jsnCfg.DialKeepAlive); err != nil {
			return
		}
	}
	return
}

func loadTransportFromJSONCfg(httpOpts *http.Transport, httpDialer *net.Dialer, jsnCfg *HTTPClientOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	httpOpts.Proxy = http.ProxyFromEnvironment
	if jsnCfg.SkipTLSVerify != nil {
		httpOpts.TLSClientConfig = &tls.Config{InsecureSkipVerify: *jsnCfg.SkipTLSVerify}
	}
	if jsnCfg.TLSHandshakeTimeout != nil {
		if httpOpts.TLSHandshakeTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.TLSHandshakeTimeout); err != nil {
			return
		}
	}
	if jsnCfg.DisableKeepAlives != nil {
		httpOpts.DisableKeepAlives = *jsnCfg.DisableKeepAlives
	}
	if jsnCfg.DisableCompression != nil {
		httpOpts.DisableCompression = *jsnCfg.DisableCompression
	}
	if jsnCfg.MaxIdleConns != nil {
		httpOpts.MaxIdleConns = *jsnCfg.MaxIdleConns
	}
	if jsnCfg.MaxIdleConnsPerHost != nil {
		httpOpts.MaxIdleConnsPerHost = *jsnCfg.MaxIdleConnsPerHost
	}
	if jsnCfg.MaxConnsPerHost != nil {
		httpOpts.MaxConnsPerHost = *jsnCfg.MaxConnsPerHost
	}
	if jsnCfg.IdleConnTimeout != nil {
		if httpOpts.IdleConnTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.IdleConnTimeout); err != nil {
			return
		}
	}
	if jsnCfg.ResponseHeaderTimeout != nil {
		if httpOpts.ResponseHeaderTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.ResponseHeaderTimeout); err != nil {
			return
		}
	}
	if jsnCfg.ExpectContinueTimeout != nil {
		if httpOpts.ExpectContinueTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.ExpectContinueTimeout); err != nil {
			return
		}
	}
	if jsnCfg.ForceAttemptHTTP2 != nil {
		httpOpts.ForceAttemptHTTP2 = *jsnCfg.ForceAttemptHTTP2
	}
	httpOpts.DialContext = httpDialer.DialContext
	return
}

// loadFromJSONCfg loads Database config from JsonCfg
func (httpcfg *HTTPCfg) loadFromJSONCfg(jsnHTTPCfg *HTTPJsonCfg) (err error) {
	if jsnHTTPCfg == nil {
		return nil
	}
	if jsnHTTPCfg.Json_rpc_url != nil {
		httpcfg.HTTPJsonRPCURL = *jsnHTTPCfg.Json_rpc_url
	}
	if jsnHTTPCfg.Registrars_url != nil {
		httpcfg.RegistrarSURL = *jsnHTTPCfg.Registrars_url
	}
	if jsnHTTPCfg.PrometheusURL != nil {
		httpcfg.PrometheusURL = *jsnHTTPCfg.PrometheusURL
	}
	if jsnHTTPCfg.Ws_url != nil {
		httpcfg.HTTPWSURL = *jsnHTTPCfg.Ws_url
	}
	if jsnHTTPCfg.Freeswitch_cdrs_url != nil {
		httpcfg.HTTPFreeswitchCDRsURL = *jsnHTTPCfg.Freeswitch_cdrs_url
	}
	if jsnHTTPCfg.Http_Cdrs != nil {
		httpcfg.HTTPCDRsURL = *jsnHTTPCfg.Http_Cdrs
	}
	if jsnHTTPCfg.PprofPath != nil {
		httpcfg.PprofPath = *jsnHTTPCfg.PprofPath
	}
	if jsnHTTPCfg.Use_basic_auth != nil {
		httpcfg.HTTPUseBasicAuth = *jsnHTTPCfg.Use_basic_auth
	}
	if jsnHTTPCfg.Auth_users != nil {
		httpcfg.HTTPAuthUsers = *jsnHTTPCfg.Auth_users
	}
	if jsnHTTPCfg.Client_opts != nil {
		if err = newDialer(httpcfg.dialer, jsnHTTPCfg.Client_opts); err != nil {
			return
		}
		err = loadTransportFromJSONCfg(httpcfg.ClientOpts, httpcfg.dialer, jsnHTTPCfg.Client_opts)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (httpcfg *HTTPCfg) AsMapInterface() map[string]any {
	clientOpts := map[string]any{
		utils.HTTPClientTLSClientConfigCfg:       false,
		utils.HTTPClientTLSHandshakeTimeoutCfg:   httpcfg.ClientOpts.TLSHandshakeTimeout.String(),
		utils.HTTPClientDisableKeepAlivesCfg:     httpcfg.ClientOpts.DisableKeepAlives,
		utils.HTTPClientDisableCompressionCfg:    httpcfg.ClientOpts.DisableCompression,
		utils.HTTPClientMaxIdleConnsCfg:          httpcfg.ClientOpts.MaxIdleConns,
		utils.HTTPClientMaxIdleConnsPerHostCfg:   httpcfg.ClientOpts.MaxIdleConnsPerHost,
		utils.HTTPClientMaxConnsPerHostCfg:       httpcfg.ClientOpts.MaxConnsPerHost,
		utils.HTTPClientIdleConnTimeoutCfg:       httpcfg.ClientOpts.IdleConnTimeout.String(),
		utils.HTTPClientResponseHeaderTimeoutCfg: httpcfg.ClientOpts.ResponseHeaderTimeout.String(),
		utils.HTTPClientExpectContinueTimeoutCfg: httpcfg.ClientOpts.ExpectContinueTimeout.String(),
		utils.HTTPClientForceAttemptHTTP2Cfg:     httpcfg.ClientOpts.ForceAttemptHTTP2,
		utils.HTTPClientDialTimeoutCfg:           httpcfg.dialer.Timeout.String(),
		utils.HTTPClientDialFallbackDelayCfg:     httpcfg.dialer.FallbackDelay.String(),
		utils.HTTPClientDialKeepAliveCfg:         httpcfg.dialer.KeepAlive.String(),
	}
	if httpcfg.ClientOpts.TLSClientConfig != nil {
		clientOpts[utils.HTTPClientTLSClientConfigCfg] = httpcfg.ClientOpts.TLSClientConfig.InsecureSkipVerify
	}
	return map[string]any{
		utils.HTTPJsonRPCURLCfg:        httpcfg.HTTPJsonRPCURL,
		utils.RegistrarSURLCfg:         httpcfg.RegistrarSURL,
		utils.PrometheusURLCfg:         httpcfg.PrometheusURL,
		utils.HTTPWSURLCfg:             httpcfg.HTTPWSURL,
		utils.HTTPFreeswitchCDRsURLCfg: httpcfg.HTTPFreeswitchCDRsURL,
		utils.HTTPCDRsURLCfg:           httpcfg.HTTPCDRsURL,
		utils.PprofPathCfg:             httpcfg.PprofPath,
		utils.HTTPUseBasicAuthCfg:      httpcfg.HTTPUseBasicAuth,
		utils.HTTPAuthUsersCfg:         httpcfg.HTTPAuthUsers,
		utils.HTTPClientOptsCfg:        clientOpts,
	}
}

// Clone returns a deep copy of HTTPCfg
func (httpcfg HTTPCfg) Clone() (cln *HTTPCfg) {
	dialer := &net.Dialer{
		Timeout:       httpcfg.dialer.Timeout,
		KeepAlive:     httpcfg.dialer.KeepAlive,
		FallbackDelay: httpcfg.dialer.FallbackDelay,
	}
	cln = &HTTPCfg{
		HTTPJsonRPCURL:        httpcfg.HTTPJsonRPCURL,
		RegistrarSURL:         httpcfg.RegistrarSURL,
		PrometheusURL:         httpcfg.PrometheusURL,
		HTTPWSURL:             httpcfg.HTTPWSURL,
		HTTPFreeswitchCDRsURL: httpcfg.HTTPFreeswitchCDRsURL,
		HTTPCDRsURL:           httpcfg.HTTPCDRsURL,
		PprofPath:             httpcfg.PprofPath,
		HTTPUseBasicAuth:      httpcfg.HTTPUseBasicAuth,
		HTTPAuthUsers:         make(map[string]string),
		ClientOpts:            httpcfg.ClientOpts.Clone(),
		dialer:                dialer,
	}
	for u, a := range httpcfg.HTTPAuthUsers {
		cln.HTTPAuthUsers[u] = a
	}
	return
}
