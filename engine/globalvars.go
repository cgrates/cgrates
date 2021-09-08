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

package engine

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// this file will contain all the global variable that are used by other subsystems

var (
	httpPstrTransport *http.Transport
	connMgr           *ConnManager
)

func init() {
	httpPstrTransport, _ = NewHTTPTransport(config.CgrConfig().HTTPCfg().ClientOpts)
}

// SetConnManager is the exported method to set the connectionManager used when operate on an account.
func SetConnManager(conMgr *ConnManager) {
	connMgr = conMgr
}

// SetHTTPPstrTransport sets the http transport to be used by the HTTP Poster
func SetHTTPPstrTransport(pstrTransport *http.Transport) {
	httpPstrTransport = pstrTransport
}

// GetHTTPPstrTransport gets the http transport to be used by the HTTP Poster
func GetHTTPPstrTransport() *http.Transport {
	return httpPstrTransport
}

// NewHTTPTransport will create a new transport for HTTP client
func NewHTTPTransport(opts map[string]interface{}) (trsp *http.Transport, err error) {
	trsp = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	if val, has := opts[utils.HTTPClientTLSClientConfigCfg]; has {
		var skipTLSVerify bool
		if skipTLSVerify, err = utils.IfaceAsBool(val); err != nil {
			return
		}
		trsp.TLSClientConfig = &tls.Config{InsecureSkipVerify: skipTLSVerify}
	}
	if val, has := opts[utils.HTTPClientTLSHandshakeTimeoutCfg]; has {
		var tlsHndTimeout time.Duration
		if tlsHndTimeout, err = utils.IfaceAsDuration(val); err != nil {
			return
		}
		trsp.TLSHandshakeTimeout = tlsHndTimeout
	}
	if val, has := opts[utils.HTTPClientDisableKeepAlivesCfg]; has {
		var disKeepAlives bool
		if disKeepAlives, err = utils.IfaceAsBool(val); err != nil {
			return
		}
		trsp.DisableKeepAlives = disKeepAlives
	}
	if val, has := opts[utils.HTTPClientDisableCompressionCfg]; has {
		var disCmp bool
		if disCmp, err = utils.IfaceAsBool(val); err != nil {
			return
		}
		trsp.DisableCompression = disCmp
	}
	if val, has := opts[utils.HTTPClientMaxIdleConnsCfg]; has {
		var maxIdleConns int64
		if maxIdleConns, err = utils.IfaceAsTInt64(val); err != nil {
			return
		}
		trsp.MaxIdleConns = int(maxIdleConns)
	}
	if val, has := opts[utils.HTTPClientMaxIdleConnsPerHostCfg]; has {
		var maxIdleConns int64
		if maxIdleConns, err = utils.IfaceAsTInt64(val); err != nil {
			return
		}
		trsp.MaxIdleConnsPerHost = int(maxIdleConns)
	}
	if val, has := opts[utils.HTTPClientMaxConnsPerHostCfg]; has {
		var maxConns int64
		if maxConns, err = utils.IfaceAsTInt64(val); err != nil {
			return
		}
		trsp.MaxConnsPerHost = int(maxConns)
	}
	if val, has := opts[utils.HTTPClientIdleConnTimeoutCfg]; has {
		var idleTimeout time.Duration
		if idleTimeout, err = utils.IfaceAsDuration(val); err != nil {
			return
		}
		trsp.IdleConnTimeout = idleTimeout
	}
	if val, has := opts[utils.HTTPClientResponseHeaderTimeoutCfg]; has {
		var responseTimeout time.Duration
		if responseTimeout, err = utils.IfaceAsDuration(val); err != nil {
			return
		}
		trsp.ResponseHeaderTimeout = responseTimeout
	}
	if val, has := opts[utils.HTTPClientExpectContinueTimeoutCfg]; has {
		var continueTimeout time.Duration
		if continueTimeout, err = utils.IfaceAsDuration(val); err != nil {
			return
		}
		trsp.ExpectContinueTimeout = continueTimeout
	}
	if val, has := opts[utils.HTTPClientForceAttemptHTTP2Cfg]; has {
		var forceHTTP2 bool
		if forceHTTP2, err = utils.IfaceAsBool(val); err != nil {
			return
		}
		trsp.ForceAttemptHTTP2 = forceHTTP2
	}
	var dial *net.Dialer
	if dial, err = newDialer(opts); err != nil {
		return
	}
	trsp.DialContext = dial.DialContext
	return
}

// newDialer returns the objects that creates the DialContext function
func newDialer(opts map[string]interface{}) (dial *net.Dialer, err error) {
	dial = &net.Dialer{
		DualStack: true,
	}
	if val, has := opts[utils.HTTPClientDialTimeoutCfg]; has {
		var timeout time.Duration
		if timeout, err = utils.IfaceAsDuration(val); err != nil {
			return
		}
		dial.Timeout = timeout
	}
	if val, has := opts[utils.HTTPClientDialFallbackDelayCfg]; has {
		var fallDelay time.Duration
		if fallDelay, err = utils.IfaceAsDuration(val); err != nil {
			return
		}
		dial.FallbackDelay = fallDelay
	}
	if val, has := opts[utils.HTTPClientDialKeepAliveCfg]; has {
		var keepAlive time.Duration
		if keepAlive, err = utils.IfaceAsDuration(val); err != nil {
			return
		}
		dial.KeepAlive = keepAlive
	}
	return
}
