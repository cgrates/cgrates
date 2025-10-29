/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package agents

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJanusHTTPjsonDPString(t *testing.T) {
	req := httptest.NewRequest("GET", "http://janus.com/test", nil)
	jHj := &janusHTTPjsonDP{
		req: req,
	}
	result := jHj.String()
	expected := "GET http://janus.com/test HTTP/1.1\r\n\r\n"
	if result == "" || !strings.Contains(result, expected) {
		t.Errorf("Expected request string to contain %q, got %q", expected, result)
	}
}

func TestJanusHTTPjsonDPFieldAsString(t *testing.T) {
	jHj := &janusHTTPjsonDP{}

	result, err := jHj.FieldAsString([]string{"test", "path"})
	if err == nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := ""
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
func TestNewJanusHTTPjsonDP(t *testing.T) {
	reqBody := `{"key": "value"}`
	req := httptest.NewRequest("POST", "http://janus.com", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	dp, err := newJanusHTTPjsonDP(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if dp == nil {
		t.Fatalf("Expected DataProvider not to be nil")
	}
	jHj, ok := dp.(*janusHTTPjsonDP)
	if !ok {
		t.Fatalf("Expected DataProvider to be of type *janusHTTPjsonDP")
	}
	expectedReqBody := map[string]interface{}{"key": "value"}
	if len(jHj.reqBody) != len(expectedReqBody) {
		t.Errorf("Expected reqBody length %d, got %d", len(expectedReqBody), len(jHj.reqBody))
	}
	for k, v := range expectedReqBody {
		if jHj.reqBody[k] != v {
			t.Errorf("Expected reqBody[%q] to be %v, got %v", k, v, jHj.reqBody[k])
		}
	}
	invalidReqBody := `{invalid json}`
	req = httptest.NewRequest("POST", "http://janus.com", bytes.NewBufferString(invalidReqBody))
	req.Header.Set("Content-Type", "application/json")
	dp, err = newJanusHTTPjsonDP(req)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
func TestJanusAccessControlHeaders(t *testing.T) {
	tests := []struct {
		originHeader string
		expectCors   bool
	}{
		{"http://janus.com", true},
		{"", false},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", "http://janus.com", nil)
		if test.originHeader != "" {
			req.Header.Set("Origin", test.originHeader)
		}
		rr := httptest.NewRecorder()
		janusAccessControlHeaders(rr, req)
		result := rr.Result()
		if test.expectCors {
			if origin := result.Header.Get("Access-Control-Allow-Origin"); origin != test.originHeader {
				t.Errorf("Expected Access-Control-Allow-Origin to be %q, got %q", test.originHeader, origin)
			}
			if methods := result.Header.Get("Access-Control-Allow-Methods"); methods != "POST, GET, OPTIONS, PUT, DELETE" {
				t.Errorf("Expected Access-Control-Allow-Methods to be %q, got %q", "POST, GET, OPTIONS, PUT, DELETE", methods)
			}
			if headers := result.Header.Get("Access-Control-Allow-Headers"); headers != "Accept, Accept-Language, Content-Type" {
				t.Errorf("Expected Access-Control-Allow-Headers to be %q, got %q", "Accept, Accept-Language, Content-Type", headers)
			}
		} else {

			if origin := result.Header.Get("Access-Control-Allow-Origin"); origin != "" {
				t.Errorf("Expected no Access-Control-Allow-Origin header, got %q", origin)
			}
		}
	}
}
