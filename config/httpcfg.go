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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// HTTPCfg is the HTTP config section
type HTTPCfg struct {
	JsonRPCURL        string // JSON RPC relative URL ("" to disable)
	RegistrarSURL     string // registrar service relative URL
	PrometheusURL     string
	WSURL             string            // WebSocket relative URL ("" to disable)
	FreeswitchCDRsURL string            // Freeswitch CDRS relative URL ("" to disable)
	CDRsURL           string            // CDRS relative URL ("" to disable)
	PprofPath         string            // runtime profiling url path ("" to disable)
	UseBasicAuth      bool              // Use basic auth for HTTP API
	AuthUsers         map[string]string // Basic auth user:password map (base64 passwords)
	ClientOpts        *http.Transport
	dialer            *net.Dialer
}

// loadHTTPCfg loads the Http section of the configuration
func (httpcfg *HTTPCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnHTTPCfg := new(HTTPJsonCfg)
	if err = jsnCfg.GetSection(ctx, HTTPJSON, jsnHTTPCfg); err != nil {
		return
	}
	return httpcfg.loadFromJSONCfg(jsnHTTPCfg)
}

func newDialer(dialer *net.Dialer, jsnCfg *HTTPClientOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	dialer.DualStack = true
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
	if jsnCfg.SkipTLSVerification != nil {
		httpOpts.TLSClientConfig = &tls.Config{InsecureSkipVerify: *jsnCfg.SkipTLSVerification}
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
		return
	}
	if jsnHTTPCfg.JSONRPCURL != nil {
		httpcfg.JsonRPCURL = *jsnHTTPCfg.JSONRPCURL
	}
	if jsnHTTPCfg.RegistrarsURL != nil {
		httpcfg.RegistrarSURL = *jsnHTTPCfg.RegistrarsURL
	}
	if jsnHTTPCfg.PrometheusURL != nil {
		httpcfg.PrometheusURL = *jsnHTTPCfg.PrometheusURL
	}
	if jsnHTTPCfg.WSURL != nil {
		httpcfg.WSURL = *jsnHTTPCfg.WSURL
	}
	if jsnHTTPCfg.FreeswitchCDRsURL != nil {
		httpcfg.FreeswitchCDRsURL = *jsnHTTPCfg.FreeswitchCDRsURL
	}
	if jsnHTTPCfg.HTTPCDRs != nil {
		httpcfg.CDRsURL = *jsnHTTPCfg.HTTPCDRs
	}
	if jsnHTTPCfg.PprofPath != nil {
		httpcfg.PprofPath = *jsnHTTPCfg.PprofPath
	}
	if jsnHTTPCfg.UseBasicAuth != nil {
		httpcfg.UseBasicAuth = *jsnHTTPCfg.UseBasicAuth
	}
	if jsnHTTPCfg.AuthUsers != nil {
		httpcfg.AuthUsers = *jsnHTTPCfg.AuthUsers
	}
	if jsnHTTPCfg.ClientOpts != nil {
		if err = newDialer(httpcfg.dialer, jsnHTTPCfg.ClientOpts); err != nil {
			return
		}
		err = loadTransportFromJSONCfg(httpcfg.ClientOpts, httpcfg.dialer, jsnHTTPCfg.ClientOpts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (httpcfg HTTPCfg) AsMapInterface(string) any {
	clientOpts := map[string]any{
		utils.HTTPClientSkipTLSVerificationCfg:   false,
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
		clientOpts[utils.HTTPClientSkipTLSVerificationCfg] = httpcfg.ClientOpts.TLSClientConfig.InsecureSkipVerify
	}
	return map[string]any{
		utils.HTTPJsonRPCURLCfg:        httpcfg.JsonRPCURL,
		utils.RegistrarSURLCfg:         httpcfg.RegistrarSURL,
		utils.PrometheusURLCfg:         httpcfg.PrometheusURL,
		utils.HTTPWSURLCfg:             httpcfg.WSURL,
		utils.HTTPFreeswitchCDRsURLCfg: httpcfg.FreeswitchCDRsURL,
		utils.HTTPCDRsURLCfg:           httpcfg.CDRsURL,
		utils.PprofPathCfg:             httpcfg.PprofPath,
		utils.HTTPUseBasicAuthCfg:      httpcfg.UseBasicAuth,
		utils.HTTPAuthUsersCfg:         httpcfg.AuthUsers,
		utils.HTTPClientOptsCfg:        clientOpts,
	}
}

func (HTTPCfg) SName() string                 { return HTTPJSON }
func (httpcfg HTTPCfg) CloneSection() Section { return httpcfg.Clone() }

// Clone returns a deep copy of HTTPCfg
func (httpcfg HTTPCfg) Clone() (cln *HTTPCfg) {
	dialer := &net.Dialer{
		Timeout:       httpcfg.dialer.Timeout,
		DualStack:     httpcfg.dialer.DualStack,
		KeepAlive:     httpcfg.dialer.KeepAlive,
		FallbackDelay: httpcfg.dialer.FallbackDelay,
	}
	cln = &HTTPCfg{
		JsonRPCURL:        httpcfg.JsonRPCURL,
		RegistrarSURL:     httpcfg.RegistrarSURL,
		PrometheusURL:     httpcfg.PrometheusURL,
		WSURL:             httpcfg.WSURL,
		FreeswitchCDRsURL: httpcfg.FreeswitchCDRsURL,
		CDRsURL:           httpcfg.CDRsURL,
		PprofPath:         httpcfg.PprofPath,
		UseBasicAuth:      httpcfg.UseBasicAuth,
		AuthUsers:         make(map[string]string),
		ClientOpts:        httpcfg.ClientOpts.Clone(),
		dialer:            dialer,
	}
	for u, a := range httpcfg.AuthUsers {
		cln.AuthUsers[u] = a
	}
	return
}

type HTTPClientOptsJson struct {
	SkipTLSVerification   *bool   `json:"skipTLSVerification"`
	TLSHandshakeTimeout   *string `json:"tlsHandshakeTimeout"`
	DisableKeepAlives     *bool   `json:"disableKeepAlives"`
	DisableCompression    *bool   `json:"disableCompression"`
	MaxIdleConns          *int    `json:"maxIdleConns"`
	MaxIdleConnsPerHost   *int    `json:"maxIdleConnsPerHost"`
	MaxConnsPerHost       *int    `json:"maxConnsPerHost"`
	IdleConnTimeout       *string `json:"IdleConnTimeout"`
	ResponseHeaderTimeout *string `json:"responseHeaderTimeout"`
	ExpectContinueTimeout *string `json:"expectContinueTimeout"`
	ForceAttemptHTTP2     *bool   `json:"forceAttemptHttp2"`
	DialTimeout           *string `json:"dialTimeout"`
	DialFallbackDelay     *string `json:"dialFallbackDelay"`
	DialKeepAlive         *string `json:"dialKeepAlive"`
}

// HTTP config section
type HTTPJsonCfg struct {
	JSONRPCURL        *string             `json:"json_rpc_url"`
	RegistrarsURL     *string             `json:"registrars_url"`
	PrometheusURL     *string             `json:"prometheus_url"`
	WSURL             *string             `json:"ws_url"`
	FreeswitchCDRsURL *string             `json:"freeswitch_cdrs_url"`
	HTTPCDRs          *string             `json:"http_cdrs"`
	PprofPath         *string             `json:"pprof_path"`
	UseBasicAuth      *bool               `json:"use_basic_auth"`
	AuthUsers         *map[string]string  `json:"auth_users"`
	ClientOpts        *HTTPClientOptsJson `json:"client_opts"`
}

func diffHTTPClientOptsJsonCfgDialer(d *HTTPClientOptsJson, v1, v2 *net.Dialer) *HTTPClientOptsJson {
	if d == nil {
		d = new(HTTPClientOptsJson)
	}
	if v1.Timeout != v2.Timeout {
		d.DialTimeout = utils.StringPointer(v2.Timeout.String())
	}
	if v1.FallbackDelay != v2.FallbackDelay {
		d.DialFallbackDelay = utils.StringPointer(v2.FallbackDelay.String())
	}
	if v1.KeepAlive != v2.KeepAlive {
		d.DialKeepAlive = utils.StringPointer(v2.KeepAlive.String())
	}
	return d
}

func diffHTTPClientOptsJsonCfg(d *HTTPClientOptsJson, v1, v2 *http.Transport) *HTTPClientOptsJson {
	if d == nil {
		d = new(HTTPClientOptsJson)
	}
	if v1.TLSClientConfig != nil {
		if v1.TLSClientConfig.InsecureSkipVerify != v2.TLSClientConfig.InsecureSkipVerify {
			d.SkipTLSVerification = utils.BoolPointer(v2.TLSClientConfig.InsecureSkipVerify)
		}
	}
	if v1.TLSHandshakeTimeout != v2.TLSHandshakeTimeout {
		d.TLSHandshakeTimeout = utils.StringPointer(v2.TLSHandshakeTimeout.String())
	}
	if v1.DisableKeepAlives != v2.DisableKeepAlives {
		d.DisableKeepAlives = utils.BoolPointer(v2.DisableKeepAlives)
	}
	if v1.DisableCompression != v2.DisableCompression {
		d.DisableCompression = utils.BoolPointer(v2.DisableCompression)
	}
	if v1.MaxIdleConns != v2.MaxIdleConns {
		d.MaxIdleConns = utils.IntPointer(v2.MaxIdleConns)
	}
	if v1.MaxIdleConnsPerHost != v2.MaxIdleConnsPerHost {
		d.MaxIdleConnsPerHost = utils.IntPointer(v2.MaxIdleConnsPerHost)
	}
	if v1.MaxConnsPerHost != v2.MaxConnsPerHost {
		d.MaxConnsPerHost = utils.IntPointer(v2.MaxConnsPerHost)
	}
	if v1.IdleConnTimeout != v2.IdleConnTimeout {
		d.IdleConnTimeout = utils.StringPointer(v2.IdleConnTimeout.String())
	}
	if v1.ResponseHeaderTimeout != v2.ResponseHeaderTimeout {
		d.ResponseHeaderTimeout = utils.StringPointer(v2.ResponseHeaderTimeout.String())
	}
	if v1.ExpectContinueTimeout != v2.ExpectContinueTimeout {
		d.ExpectContinueTimeout = utils.StringPointer(v2.ExpectContinueTimeout.String())
	}
	if v1.ForceAttemptHTTP2 != v2.ForceAttemptHTTP2 {
		d.ForceAttemptHTTP2 = utils.BoolPointer(v2.ForceAttemptHTTP2)
	}
	return d
}
func diffHTTPJsonCfg(d *HTTPJsonCfg, v1, v2 *HTTPCfg) *HTTPJsonCfg {
	if d == nil {
		d = new(HTTPJsonCfg)
	}

	if v1.JsonRPCURL != v2.JsonRPCURL {
		d.JSONRPCURL = utils.StringPointer(v2.JsonRPCURL)
	}
	if v1.RegistrarSURL != v2.RegistrarSURL {
		d.RegistrarsURL = utils.StringPointer(v2.RegistrarSURL)
	}
	if v1.PrometheusURL != v2.PrometheusURL {
		d.PrometheusURL = utils.StringPointer(v2.PrometheusURL)
	}
	if v1.WSURL != v2.WSURL {
		d.WSURL = utils.StringPointer(v2.WSURL)
	}
	if v1.FreeswitchCDRsURL != v2.FreeswitchCDRsURL {
		d.FreeswitchCDRsURL = utils.StringPointer(v2.FreeswitchCDRsURL)
	}
	if v1.CDRsURL != v2.CDRsURL {
		d.HTTPCDRs = utils.StringPointer(v2.CDRsURL)
	}
	if v1.PprofPath != v2.PprofPath {
		d.PprofPath = utils.StringPointer(v2.PprofPath)
	}
	if v1.UseBasicAuth != v2.UseBasicAuth {
		d.UseBasicAuth = utils.BoolPointer(v2.UseBasicAuth)
	}
	if !utils.MapStringStringEqual(v1.AuthUsers, v2.AuthUsers) {
		d.AuthUsers = &v2.AuthUsers
	}
	d.ClientOpts = diffHTTPClientOptsJsonCfg(d.ClientOpts, v1.ClientOpts, v2.ClientOpts)
	d.ClientOpts = diffHTTPClientOptsJsonCfgDialer(d.ClientOpts, v1.dialer, v2.dialer)
	return d
}
