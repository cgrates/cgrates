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
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

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
	var result interface{} = "OK"
	var errMessage interface{}
	slsByte := []byte("10")
	id = (*json.RawMessage)(&slsByte)

	if err := WriteServerResponse(writer, id, result, errMessage); err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	}
	if writer.String() != "{\"id\":10,\"result\":\"OK\",\"error\":null}\n" {
		t.Errorf("Expecting: <{\"id\":10,\"result\":\"OK\",\"error\":null}>, received: <%+v>", writer.String())
	}
}
