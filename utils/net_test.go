// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestLocalAddr(t *testing.T) {
	eOut := &NetAddr{network: Local, ip: Local}
	if rcv := LocalAddr(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
}

func TestNewNetAddr(t *testing.T) {
	eOut := &NetAddr{}
	if rcv := NewNetAddr(EmptyString, EmptyString); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
	eOut = &NetAddr{network: "network"}
	if rcv := NewNetAddr("network", EmptyString); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
	eOut = &NetAddr{ip: "127.0.0.1", port: 2012}
	if rcv := NewNetAddr(EmptyString, "127.0.0.1:2012"); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
	eOut = &NetAddr{network: "network", ip: "127.0.0.1", port: 2012}
	if rcv := NewNetAddr("network", "127.0.0.1:2012"); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
}

func TestNetAddrNetwork(t *testing.T) {
	lc := NetAddr{network: "network", ip: "127.0.0.1", port: 2012}
	if rcv := lc.Network(); !reflect.DeepEqual(rcv, lc.network) {
		t.Errorf("Expected: %+v, received: %+v ", lc.network, rcv)
	}
}

func TestNetAddrString(t *testing.T) {
	lc := NetAddr{network: "network", ip: "127.0.0.1", port: 2012}
	if rcv := lc.String(); !reflect.DeepEqual(rcv, lc.ip) {
		t.Errorf("Expected: %+v, received: %+v ", lc.ip, rcv)
	}
}

func TestNetAddrPort(t *testing.T) {
	lc := NetAddr{network: "network", ip: "127.0.0.1", port: 2012}
	if rcv := lc.Port(); !reflect.DeepEqual(rcv, lc.port) {
		t.Errorf("Expected: %+v, received: %+v ", lc.port, rcv)
	}
}

func TestNetAddrHost(t *testing.T) {
	lc := NetAddr{network: "network", ip: "127.0.0.1", port: 2012}
	if rcv := lc.Host(); !reflect.DeepEqual(rcv, "127.0.0.1:2012") {
		t.Errorf("Expected: '127.0.0.1:2012', received: %+v ", rcv)
	}
	lc = NetAddr{network: "network", ip: Local, port: 2012}
	if rcv := lc.Host(); !reflect.DeepEqual(rcv, Local) {
		t.Errorf("Expected: %+v, received: %+v ", Local, rcv)
	}
}

func TestGetRemoteIP(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:2080/json_rpc", bytes.NewBuffer(nil))
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1:2356"
	exp := "127.0.0.1"
	if rply, err := GetRemoteIP(req); err != nil {
		t.Fatal(err)
	} else if rply != exp {
		t.Errorf("Expected: %q ,received: %q", exp, rply)
	}
	req.RemoteAddr = "notAnIP"
	if _, err := GetRemoteIP(req); err == nil {
		t.Fatal("Expected error received nil")
	}
	req.RemoteAddr = "127.0.0:2012"
	if _, err := GetRemoteIP(req); err == nil {
		t.Fatal("Expected error received nil")
	}

	req.Header.Set("X-FORWARDED-FOR", "127.0.0.2,127.0.0.3")
	exp = "127.0.0.2"
	if rply, err := GetRemoteIP(req); err != nil {
		t.Fatal(err)
	} else if rply != exp {
		t.Errorf("Expected: %q ,received: %q", exp, rply)
	}
	req.Header.Set("X-FORWARDED-FOR", "127.0.0.")
	if _, err := GetRemoteIP(req); err == nil {
		t.Fatal("Expected error received nil")
	}

	req.Header.Set("X-REAL-IP", "127.0.0.4")
	exp = "127.0.0.4"
	if rply, err := GetRemoteIP(req); err != nil {
		t.Fatal(err)
	} else if rply != exp {
		t.Errorf("Expected: %q ,received: %q", exp, rply)
	}
}

func TestNewServerRequest(t *testing.T) {
	test := &serverRequest{
		Method: "1",
		Params: &json.RawMessage{'2'},
		Id:     &json.RawMessage{'3'},
	}
	a := NewServerRequest("1", json.RawMessage{'2'}, json.RawMessage{'3'})
	if !reflect.DeepEqual(a, test) {
		t.Errorf("Expecting: %+v, received: %+v", test, a)
	}
}

func TestDecodeServerRequest(t *testing.T) {
	test := strings.NewReader("{\"method\":\"APIerSv1.LoadTariffPlanFromFolder\",\"params\":[{\"FolderPath\":\"/usr/share/cgrates/tariffplans/tutorial\",\"DryRun\":false,\"Validate\":false,\"Opts\":null,\"Caching\":null}],\"id\":0}")
	test2 := strings.NewReader("{\"method\":\"APIerSv1.LoadTariffPlanFromFolder\",\"params\":[{\"FolderPath\":\"/usr/share/cgrates/tariffplans/tutorial\",\"DryRun\":false,\"Validate\":false,\"Opts\":null,\"Caching\":null}],\"id\":0}")
	req := new(serverRequest)
	err := json.NewDecoder(test).Decode(req)
	rcvReq, rcvErr := DecodeServerRequest(test2)
	if !reflect.DeepEqual(req, rcvReq) {
		t.Errorf("Expecting: %+v, received: %+v", req, rcvReq)
	}
	if err != rcvErr {
		t.Errorf("Expecting: %+v, received: %+v", err, rcvErr)
	}
}

func TestWriteServerResponse(t *testing.T) {
	writer := bytes.NewBufferString(EmptyString)
	var id *json.RawMessage
	var result any = "OK"
	var errMessage any
	slsByte := []byte("10")
	id = (*json.RawMessage)(&slsByte)

	if err := WriteServerResponse(writer, id, result, errMessage); err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	}
	if writer.String() != "{\"id\":10,\"result\":\"OK\",\"error\":null}\n" {
		t.Errorf("Expecting: <{\"id\":10,\"result\":\"OK\",\"error\":null}>, received: <%+v>", writer.String())
	}
}
