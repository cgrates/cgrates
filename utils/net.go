// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	RpcRegister(rcvr any)
	RpcRegisterName(name string, rcvr any)
	RegisterHTTPFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	RegisterHttpHandler(pattern string, handler http.Handler)
	BiRPCRegisterName(method string, handlerFunc any)
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

// NewServerRequest used in registrarc tests
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

func WriteServerResponse(w io.Writer, id *json.RawMessage, result, err any) error {
	return json.NewEncoder(w).Encode(
		serverResponse{
			Id:     id,
			Result: result,
			Error:  err,
		})
}

type serverResponse struct {
	Id     *json.RawMessage `json:"id"`
	Result any              `json:"result"`
	Error  any              `json:"error"`
}
