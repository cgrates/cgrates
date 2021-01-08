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

package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type Server interface {
	RpcRegister(rcvr interface{})
	RpcRegisterName(name string, rcvr interface{})
	RegisterHttpFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	RegisterHttpHandler(pattern string, handler http.Handler)
	BiRPCRegisterName(method string, handlerFunc interface{})
	BiRPCRegister(rcvr interface{})
}

func LocalAddr() *NetAddr {
	return &NetAddr{network: Local, ip: Local}
}

func NewNetAddr(network, host string) *NetAddr {
	ip, port, err := net.SplitHostPort(host)
	if err != nil {
		Logger.Warning(fmt.Sprintf("failed parsing RemoteAddr: %s, err: %s",
			host, err.Error()))
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		Logger.Warning(fmt.Sprintf("failed converting port : %+v, err: %s",
			port, err.Error()))
	}
	return &NetAddr{network: network, ip: ip, port: p}
}

type NetAddr struct {
	network string
	ip      string
	port    int
}

// Network is part of net.Addr interface
func (lc *NetAddr) Network() string {
	return lc.network
}

// String is part of net.Addr interface
func (lc *NetAddr) String() string {
	return lc.ip
}

// Port .
func (lc *NetAddr) Port() int {
	return lc.port
}

// Host .
func (lc *NetAddr) Host() string {
	if lc.ip == Local {
		return Local
	}
	return ConcatenatedKey(lc.ip, strconv.Itoa(lc.port))
}

// GetRemoteIP returns the IP from http request
func GetRemoteIP(r *http.Request) (ip string, err error) {
	ip = r.Header.Get("X-REAL-IP")
	if net.ParseIP(ip) != nil {
		return
	}
	for _, ip = range strings.Split(r.Header.Get("X-FORWARDED-FOR"), FieldsSep) {
		if net.ParseIP(ip) != nil {
			return
		}
	}
	if ip, _, err = net.SplitHostPort(r.RemoteAddr); err != nil {
		return
	}
	if net.ParseIP(ip) != nil {
		return
	}
	ip = EmptyString
	err = fmt.Errorf("no valid ip found")
	return
}

func DecodeServerRequest(r io.Reader) (req *serverRequest, err error) {
	req = new(serverRequest)
	err = json.NewDecoder(r).Decode(req)
	return
}

// NewServerRequest used in dispatcherh tests
func NewServerRequest(method string, params, id json.RawMessage) *serverRequest {
	return &serverRequest{
		Method: method,
		Params: &params,
		Id:     &id,
	}
}

type serverRequest struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	Id     *json.RawMessage `json:"id"`
}

func WriteServerResponse(w io.Writer, id *json.RawMessage, result, err interface{}) error {
	return json.NewEncoder(w).Encode(
		serverResponse{
			Id:     id,
			Result: result,
			Error:  err,
		})
}

type serverResponse struct {
	Id     *json.RawMessage `json:"id"`
	Result interface{}      `json:"result"`
	Error  interface{}      `json:"error"`
}
