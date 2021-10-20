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

type HTTPClientOpts struct {
	Transport *http.Transport
	Dialer    *net.Dialer
}

// HTTPCfg is the HTTP config section
type HTTPCfg struct {
	JsonRPCURL        string            // JSON RPC relative URL ("" to disable)
	RegistrarSURL     string            // registrar service relative URL
	WSURL             string            // WebSocket relative URL ("" to disable)
	FreeswitchCDRsURL string            // Freeswitch CDRS relative URL ("" to disable)
	CDRsURL           string            // CDRS relative URL ("" to disable)
	UseBasicAuth      bool              // Use basic auth for HTTP API
	AuthUsers         map[string]string // Basic auth user:password map (base64 passwords)
	ClientOpts        *HTTPClientOpts
}

// loadHTTPCfg loads the Http section of the configuration
func (httpcfg *HTTPCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnHTTPCfg := new(HTTPJsonCfg)
	if err = jsnCfg.GetSection(ctx, HTTPJSON, jsnHTTPCfg); err != nil {
		return
	}
	return httpcfg.loadFromJSONCfg(jsnHTTPCfg)
}

func newDialer(jsnCfg *HTTPClientOptsJson) (dialer *net.Dialer, err error) {
	if jsnCfg == nil {
		return
	}
	dialer = &net.Dialer{
		DualStack: true,
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

func (httpOpts *HTTPClientOpts) loadFromJSONCfg(jsnCfg *HTTPClientOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	httpOpts.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	if jsnCfg.SkipTLSVerification != nil {
		httpOpts.Transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: *jsnCfg.SkipTLSVerification}
	}
	if jsnCfg.TLSHandshakeTimeout != nil {
		if httpOpts.Transport.TLSHandshakeTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.TLSHandshakeTimeout); err != nil {
			return
		}
	}
	if jsnCfg.DisableKeepAlives != nil {
		httpOpts.Transport.DisableKeepAlives = *jsnCfg.DisableKeepAlives
	}
	if jsnCfg.DisableCompression != nil {
		httpOpts.Transport.DisableCompression = *jsnCfg.DisableCompression
	}
	if jsnCfg.MaxIdleConns != nil {
		httpOpts.Transport.MaxIdleConns = *jsnCfg.MaxIdleConns
	}
	if jsnCfg.MaxIdleConnsPerHost != nil {
		httpOpts.Transport.MaxIdleConnsPerHost = *jsnCfg.MaxIdleConnsPerHost
	}
	if jsnCfg.MaxConnsPerHost != nil {
		httpOpts.Transport.MaxConnsPerHost = *jsnCfg.MaxConnsPerHost
	}
	if jsnCfg.IdleConnTimeout != nil {
		if httpOpts.Transport.IdleConnTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.IdleConnTimeout); err != nil {
			return
		}
	}
	if jsnCfg.ResponseHeaderTimeout != nil {
		if httpOpts.Transport.ResponseHeaderTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.ResponseHeaderTimeout); err != nil {
			return
		}
	}
	if jsnCfg.ExpectContinueTimeout != nil {
		if httpOpts.Transport.ExpectContinueTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.ExpectContinueTimeout); err != nil {
			return
		}
	}
	if jsnCfg.ForceAttemptHTTP2 != nil {
		httpOpts.Transport.ForceAttemptHTTP2 = *jsnCfg.ForceAttemptHTTP2
	}
	if httpOpts.Dialer, err = newDialer(jsnCfg); err != nil {
		return
	}
	httpOpts.Transport.DialContext = httpOpts.Dialer.DialContext
	return
}

// loadFromJSONCfg loads Database config from JsonCfg
func (httpcfg *HTTPCfg) loadFromJSONCfg(jsnHTTPCfg *HTTPJsonCfg) (err error) {
	if jsnHTTPCfg == nil {
		return nil
	}
	if jsnHTTPCfg.Json_rpc_url != nil {
		httpcfg.JsonRPCURL = *jsnHTTPCfg.Json_rpc_url
	}
	if jsnHTTPCfg.Registrars_url != nil {
		httpcfg.RegistrarSURL = *jsnHTTPCfg.Registrars_url
	}
	if jsnHTTPCfg.Ws_url != nil {
		httpcfg.WSURL = *jsnHTTPCfg.Ws_url
	}
	if jsnHTTPCfg.Freeswitch_cdrs_url != nil {
		httpcfg.FreeswitchCDRsURL = *jsnHTTPCfg.Freeswitch_cdrs_url
	}
	if jsnHTTPCfg.Http_Cdrs != nil {
		httpcfg.CDRsURL = *jsnHTTPCfg.Http_Cdrs
	}
	if jsnHTTPCfg.Use_basic_auth != nil {
		httpcfg.UseBasicAuth = *jsnHTTPCfg.Use_basic_auth
	}
	if jsnHTTPCfg.Auth_users != nil {
		httpcfg.AuthUsers = *jsnHTTPCfg.Auth_users
	}
	if jsnHTTPCfg.Client_opts != nil {
		err = httpcfg.ClientOpts.loadFromJSONCfg(jsnHTTPCfg.Client_opts)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (httpcfg HTTPCfg) AsMapInterface(string) interface{} {
	clientOpts := map[string]interface{}{
		utils.HTTPClientSkipTLSVerificationCfg:   httpcfg.ClientOpts.Transport.TLSClientConfig.InsecureSkipVerify,
		utils.HTTPClientTLSHandshakeTimeoutCfg:   httpcfg.ClientOpts.Transport.TLSHandshakeTimeout,
		utils.HTTPClientDisableKeepAlivesCfg:     httpcfg.ClientOpts.Transport.DisableKeepAlives,
		utils.HTTPClientDisableCompressionCfg:    httpcfg.ClientOpts.Transport.DisableCompression,
		utils.HTTPClientMaxIdleConnsCfg:          httpcfg.ClientOpts.Transport.MaxIdleConns,
		utils.HTTPClientMaxIdleConnsPerHostCfg:   httpcfg.ClientOpts.Transport.MaxIdleConnsPerHost,
		utils.HTTPClientMaxConnsPerHostCfg:       httpcfg.ClientOpts.Transport.MaxConnsPerHost,
		utils.HTTPClientIdleConnTimeoutCfg:       httpcfg.ClientOpts.Transport.IdleConnTimeout,
		utils.HTTPClientResponseHeaderTimeoutCfg: httpcfg.ClientOpts.Transport.ResponseHeaderTimeout,
		utils.HTTPClientExpectContinueTimeoutCfg: httpcfg.ClientOpts.Transport.ExpectContinueTimeout,
		utils.HTTPClientForceAttemptHTTP2Cfg:     httpcfg.ClientOpts.Transport.ForceAttemptHTTP2,
		utils.HTTPClientDialTimeoutCfg:           httpcfg.ClientOpts.Dialer.Timeout,
		utils.HTTPClientDialFallbackDelayCfg:     httpcfg.ClientOpts.Dialer.FallbackDelay,
		utils.HTTPClientDialKeepAliveCfg:         httpcfg.ClientOpts.Dialer.KeepAlive,
	}
	return map[string]interface{}{
		utils.HTTPJsonRPCURLCfg:        httpcfg.JsonRPCURL,
		utils.RegistrarSURLCfg:         httpcfg.RegistrarSURL,
		utils.HTTPWSURLCfg:             httpcfg.WSURL,
		utils.HTTPFreeswitchCDRsURLCfg: httpcfg.FreeswitchCDRsURL,
		utils.HTTPCDRsURLCfg:           httpcfg.CDRsURL,
		utils.HTTPUseBasicAuthCfg:      httpcfg.UseBasicAuth,
		utils.HTTPAuthUsersCfg:         httpcfg.AuthUsers,
		utils.HTTPClientOptsCfg:        clientOpts,
	}
}

func (HTTPCfg) SName() string                 { return HTTPJSON }
func (httpcfg HTTPCfg) CloneSection() Section { return httpcfg.Clone() }

func (httpOpts *HTTPClientOpts) Clone() (cln *HTTPClientOpts) {
	transport := &http.Transport{
		Proxy:                 httpOpts.Transport.Proxy,
		TLSClientConfig:       httpOpts.Transport.TLSClientConfig,
		TLSHandshakeTimeout:   httpOpts.Transport.TLSHandshakeTimeout,
		DisableKeepAlives:     httpOpts.Transport.DisableKeepAlives,
		DisableCompression:    httpOpts.Transport.DisableCompression,
		MaxIdleConns:          httpOpts.Transport.MaxIdleConns,
		MaxIdleConnsPerHost:   httpOpts.Transport.MaxIdleConnsPerHost,
		MaxConnsPerHost:       httpOpts.Transport.MaxConnsPerHost,
		IdleConnTimeout:       httpOpts.Transport.IdleConnTimeout,
		ResponseHeaderTimeout: httpOpts.Transport.ResponseHeaderTimeout,
		ExpectContinueTimeout: httpOpts.Transport.ExpectContinueTimeout,
		ForceAttemptHTTP2:     httpOpts.Transport.ForceAttemptHTTP2,
	}
	dialer := &net.Dialer{
		Timeout:       httpOpts.Dialer.Timeout,
		DualStack:     httpOpts.Dialer.DualStack,
		KeepAlive:     httpOpts.Dialer.KeepAlive,
		FallbackDelay: httpOpts.Dialer.FallbackDelay,
	}
	return &HTTPClientOpts{
		Transport: transport,
		Dialer:    dialer,
	}
}

// Clone returns a deep copy of HTTPCfg
func (httpcfg HTTPCfg) Clone() (cln *HTTPCfg) {
	cln = &HTTPCfg{
		JsonRPCURL:        httpcfg.JsonRPCURL,
		RegistrarSURL:     httpcfg.RegistrarSURL,
		WSURL:             httpcfg.WSURL,
		FreeswitchCDRsURL: httpcfg.FreeswitchCDRsURL,
		CDRsURL:           httpcfg.CDRsURL,
		UseBasicAuth:      httpcfg.UseBasicAuth,
		AuthUsers:         make(map[string]string),
		ClientOpts:        httpcfg.ClientOpts.Clone(),
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
	MaxIdleConnsPerHost   *int    `json:"maxConnsPerHost"`
	MaxConnsPerHost       *int    `json:"maxConnsPerHost"`
	IdleConnTimeout       *string `json:"IdleConnTimeout"`
	ResponseHeaderTimeout *string `json:"responseHeaderTimeout"`
	ExpectContinueTimeout *string `json:"expectContinueTransport"`
	ForceAttemptHTTP2     *bool   `json:"forceAttemptHttp2"`
	DialTimeout           *string `json:"dialTimeout"`
	DialFallbackDelay     *string `json:"dialFallbackDelay"`
	DialKeepAlive         *string `json:"dialKeepAlive"`
}

// HTTP config section
type HTTPJsonCfg struct {
	Json_rpc_url        *string
	Registrars_url      *string
	Ws_url              *string
	Freeswitch_cdrs_url *string
	Http_Cdrs           *string
	Use_basic_auth      *bool
	Auth_users          *map[string]string
	Client_opts         *HTTPClientOptsJson
}

func diffHTTPClientOptsJsonCfg(d *HTTPClientOptsJson, v1, v2 *HTTPClientOpts) *HTTPClientOptsJson {
	if d == nil {
		d = new(HTTPClientOptsJson)
	}
	if v1.Transport.TLSClientConfig.InsecureSkipVerify != v2.Transport.TLSClientConfig.InsecureSkipVerify {
		d.SkipTLSVerification = utils.BoolPointer(v2.Transport.TLSClientConfig.InsecureSkipVerify)
	}
	if v1.Transport.TLSHandshakeTimeout != v2.Transport.TLSHandshakeTimeout {
		d.TLSHandshakeTimeout = utils.StringPointer(v2.Transport.TLSHandshakeTimeout.String())
	}
	if v1.Transport.DisableKeepAlives != v2.Transport.DisableKeepAlives {
		d.DisableKeepAlives = utils.BoolPointer(v2.Transport.DisableKeepAlives)
	}
	if v1.Transport.DisableCompression != v2.Transport.DisableCompression {
		d.DisableCompression = utils.BoolPointer(v2.Transport.DisableCompression)
	}
	if v1.Transport.MaxIdleConns != v2.Transport.MaxIdleConns {
		d.MaxIdleConns = utils.IntPointer(v2.Transport.MaxIdleConns)
	}
	if v1.Transport.MaxIdleConnsPerHost != v2.Transport.MaxIdleConnsPerHost {
		d.MaxIdleConnsPerHost = utils.IntPointer(v2.Transport.MaxIdleConnsPerHost)
	}
	if v1.Transport.MaxConnsPerHost != v2.Transport.MaxConnsPerHost {
		d.MaxConnsPerHost = utils.IntPointer(v2.Transport.MaxConnsPerHost)
	}
	if v1.Transport.IdleConnTimeout != v2.Transport.IdleConnTimeout {
		d.IdleConnTimeout = utils.StringPointer(v2.Transport.IdleConnTimeout.String())
	}
	if v1.Transport.ResponseHeaderTimeout != v2.Transport.ResponseHeaderTimeout {
		d.ResponseHeaderTimeout = utils.StringPointer(v2.Transport.ResponseHeaderTimeout.String())
	}
	if v1.Transport.ExpectContinueTimeout != v2.Transport.ExpectContinueTimeout {
		d.ExpectContinueTimeout = utils.StringPointer(v2.Transport.ExpectContinueTimeout.String())
	}
	if v1.Transport.ForceAttemptHTTP2 != v2.Transport.ForceAttemptHTTP2 {
		d.ForceAttemptHTTP2 = utils.BoolPointer(v2.Transport.ForceAttemptHTTP2)
	}
	if v1.Dialer.Timeout != v2.Dialer.Timeout {
		d.DialTimeout = utils.StringPointer(v2.Dialer.Timeout.String())
	}
	if v1.Dialer.FallbackDelay != v2.Dialer.FallbackDelay {
		d.DialFallbackDelay = utils.StringPointer(v2.Dialer.FallbackDelay.String())
	}
	if v1.Dialer.KeepAlive != v2.Dialer.KeepAlive {
		d.DialKeepAlive = utils.StringPointer(v2.Dialer.KeepAlive.String())
	}
	return d
}
func diffHTTPJsonCfg(d *HTTPJsonCfg, v1, v2 *HTTPCfg) *HTTPJsonCfg {
	if d == nil {
		d = new(HTTPJsonCfg)
	}

	if v1.JsonRPCURL != v2.JsonRPCURL {
		d.Json_rpc_url = utils.StringPointer(v2.JsonRPCURL)
	}
	if v1.RegistrarSURL != v2.RegistrarSURL {
		d.Registrars_url = utils.StringPointer(v2.RegistrarSURL)
	}
	if v1.WSURL != v2.WSURL {
		d.Ws_url = utils.StringPointer(v2.WSURL)
	}
	if v1.FreeswitchCDRsURL != v2.FreeswitchCDRsURL {
		d.Freeswitch_cdrs_url = utils.StringPointer(v2.FreeswitchCDRsURL)
	}
	if v1.CDRsURL != v2.CDRsURL {
		d.Http_Cdrs = utils.StringPointer(v2.CDRsURL)
	}
	if v1.UseBasicAuth != v2.UseBasicAuth {
		d.Use_basic_auth = utils.BoolPointer(v2.UseBasicAuth)
	}
	if !utils.MapStringStringEqual(v1.AuthUsers, v2.AuthUsers) {
		d.Auth_users = &v2.AuthUsers
	}
	d.Client_opts = diffHTTPClientOptsJsonCfg(d.Client_opts, v1.ClientOpts, v2.ClientOpts)
	return d
}

func diffMap(d, v1, v2 map[string]interface{}) map[string]interface{} {
	if d == nil {
		d = make(map[string]interface{})
	}
	for k, v := range v2 {
		if val, has := v1[k]; !has || val != v {
			d[k] = v
		}
	}
	return d
}
