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
	"reflect"
	"testing"
)

func TestFFNNewFallbackFileNameFronString(t *testing.T) {
	fileName := "cdr|*http_json_cdr|http%3A%2F%2F127.0.0.1%3A12080%2Finvalid_json|1acce2c9-3f2d-4774-8662-c28872dad515.json"
	eFFN := &FallbackFileName{Module: "cdr",
		Transport:  MetaHTTPjsonCDR,
		Address:    "http://127.0.0.1:12080/invalid_json",
		RequestID:  "1acce2c9-3f2d-4774-8662-c28872dad515",
		FileSuffix: JSNSuffix}
	if ffn, err := NewFallbackFileNameFronString(fileName); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFFN, ffn) {
		t.Errorf("Expecting: %+v, received: %+v", eFFN, ffn)
	}
	fileName = "cdr|*http_post|http%3A%2F%2F127.0.0.1%3A12080%2Finvalid|70c53d6d-dbd7-452e-a5bd-36bab59bb9ff.form"
	eFFN = &FallbackFileName{Module: "cdr",
		Transport:  META_HTTP_POST,
		Address:    "http://127.0.0.1:12080/invalid",
		RequestID:  "70c53d6d-dbd7-452e-a5bd-36bab59bb9ff",
		FileSuffix: FormSuffix}
	if ffn, err := NewFallbackFileNameFronString(fileName); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFFN, ffn) {
		t.Errorf("Expecting: %+v, received: %+v", eFFN, ffn)
	}
	fileName = "act>*call_url|*http_json|http%3A%2F%2Flocalhost%3A2080%2Flog_warning|f52cf23e-da2f-4675-b36b-e8fcc3869270.json"
	eFFN = &FallbackFileName{Module: "act>*call_url",
		Transport:  MetaHTTPjson,
		Address:    "http://localhost:2080/log_warning",
		RequestID:  "f52cf23e-da2f-4675-b36b-e8fcc3869270",
		FileSuffix: JSNSuffix}
	if ffn, err := NewFallbackFileNameFronString(fileName); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFFN, ffn) {
		t.Errorf("Expecting: %+v, received: %+v", eFFN, ffn)
	}
}

func TestFFNFallbackFileNameAsString(t *testing.T) {
	eFn := "cdr|*http_json_cdr|http%3A%2F%2F127.0.0.1%3A12080%2Finvalid_json|1acce2c9-3f2d-4774-8662-c28872dad515.json"
	ffn := &FallbackFileName{
		Module:     "cdr",
		Transport:  MetaHTTPjsonCDR,
		Address:    "http://127.0.0.1:12080/invalid_json",
		RequestID:  "1acce2c9-3f2d-4774-8662-c28872dad515",
		FileSuffix: JSNSuffix}
	if ffnStr := ffn.AsString(); ffnStr != eFn {
		t.Errorf("Expecting: <%q>, received: <%q>", eFn, ffnStr)
	}
}
