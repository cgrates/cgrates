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
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestNewHTTPTransport(t *testing.T) {
	opts := map[string]interface{}{
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
	}

	expDialer := &net.Dialer{
		DualStack:     true,
		Timeout:       30 * time.Second,
		FallbackDelay: 300 * time.Millisecond,
		KeepAlive:     30 * time.Second,
	}
	if dial, err := newDialer(opts); err != nil {
		t.Fatal(err)
	} else if !(expDialer != nil && dial != nil &&
		expDialer.DualStack == dial.DualStack &&
		expDialer.Timeout == dial.Timeout &&
		expDialer.FallbackDelay == dial.FallbackDelay &&
		expDialer.KeepAlive == dial.KeepAlive) {
		t.Errorf("Expected %+v, received %+v", expDialer, dial)
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
	if trsp, err := NewHTTPTransport(opts); err != nil {
		t.Fatal(err)
	} else if !(expTransport != nil && trsp != nil && // the dial options are not included
		expTransport.TLSClientConfig.InsecureSkipVerify == trsp.TLSClientConfig.InsecureSkipVerify &&
		expTransport.TLSHandshakeTimeout == trsp.TLSHandshakeTimeout &&
		expTransport.DisableKeepAlives == trsp.DisableKeepAlives &&
		expTransport.DisableCompression == trsp.DisableCompression &&
		expTransport.MaxIdleConns == trsp.MaxIdleConns &&
		expTransport.MaxIdleConnsPerHost == trsp.MaxIdleConnsPerHost &&
		expTransport.MaxConnsPerHost == trsp.MaxConnsPerHost &&
		expTransport.IdleConnTimeout == trsp.IdleConnTimeout &&
		expTransport.ResponseHeaderTimeout == trsp.ResponseHeaderTimeout &&
		expTransport.ExpectContinueTimeout == trsp.ExpectContinueTimeout &&
		expTransport.ForceAttemptHTTP2 == trsp.ForceAttemptHTTP2) {
		t.Errorf("Expected %+v, received %+v", expTransport, trsp)
	}

	opts[utils.HTTPClientDialKeepAliveCfg] = "30as"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientDialFallbackDelayCfg] = "300ams"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientDialTimeoutCfg] = "30as"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientForceAttemptHTTP2Cfg] = "string"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientExpectContinueTimeoutCfg] = "0a"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientResponseHeaderTimeoutCfg] = "0a"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientIdleConnTimeoutCfg] = "90as"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientMaxConnsPerHostCfg] = "not a number"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientMaxIdleConnsPerHostCfg] = "not a number"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientMaxIdleConnsCfg] = "not a number"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientDisableCompressionCfg] = "string"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientDisableKeepAlivesCfg] = "string"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientTLSHandshakeTimeoutCfg] = "10as"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
	opts[utils.HTTPClientTLSClientConfigCfg] = "string"
	if _, err := NewHTTPTransport(opts); err == nil {
		t.Error("Expected error but the transport was builded succesfully")
	}
}

func TestSetHTTPPstrTransport(t *testing.T) {
	tmp := httpPstrTransport
	SetHTTPPstrTransport(nil)
	if httpPstrTransport != nil {
		t.Error("Expected the transport to be nil", httpPstrTransport)
	}
	httpPstrTransport = tmp
}
