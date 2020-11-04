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
package cores

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUse(t *testing.T) {
	runned := 0
	mid := func(h http.HandlerFunc) http.HandlerFunc {
		runned++
		return h
	}
	g := use(func(http.ResponseWriter, *http.Request) {}, mid, mid)
	g(nil, nil)
	if runned != 2 {
		t.Error("Expecting something")
	}
}
func TestBasicAuth(t *testing.T) {
	midle := basicAuth(map[string]string{"1001": "MTIzNA=="})
	var runned bool
	toTest := midle(func(http.ResponseWriter, *http.Request) {
		runned = true
	})

	req, err := http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Get "+base64.StdEncoding.EncodeToString([]byte("1001:1234")))
	rr := httptest.NewRecorder()

	toTest.ServeHTTP(rr, req)
	if !runned {
		t.Error("ResponseWrite error")
	}
	if rr.Result().Header.Get("WWW-Authenticate") != `Basic realm="Restricted"` {
		t.Error("Expecting: Basic realm=Restricted, received: ", rr.Result().Header.Get("WWW-Authenticate"))
	}
	//part 1 -> <BasicAuth> Missing authorization header value
	runned = false
	req, err = http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("missing", "Get "+base64.StdEncoding.EncodeToString([]byte("1001:1234")))
	rr = httptest.NewRecorder()

	toTest.ServeHTTP(rr, req)
	if runned {
		t.Error("ResponseWrite error")
	}
	if rr.Result().Header.Get("WWW-Authenticate") != `Basic realm="Restricted"` {
		t.Error("Expecting: Basic realm=Restricted, received: ", rr.Result().Header.Get("WWW-Authenticate"))
	}
	//part 2 -> <BasicAuth> Unable to decode authorization header
	runned = false
	req, err = http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Get WRONG STRING "+base64.StdEncoding.EncodeToString([]byte("1001:1234")))
	rr = httptest.NewRecorder()

	toTest.ServeHTTP(rr, req)
	if runned {
		t.Error("ResponseWrite error")
	}
	if rr.Result().Header.Get("WWW-Authenticate") != `Basic realm="Restricted"` {
		t.Error("Expecting: Basic realm=Restricted, received: ", rr.Result().Header.Get("WWW-Authenticate"))
	}
	//part 3 -> <BasicAuth> Unauthorized API access. Missing or extra credential components
	runned = false
	req, err = http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Get "+base64.StdEncoding.EncodeToString([]byte("10011234")))
	rr = httptest.NewRecorder()

	toTest.ServeHTTP(rr, req)
	if runned {
		t.Error("ResponseWrite error")
	}
	if rr.Result().Header.Get("WWW-Authenticate") != `Basic realm="Restricted"` {
		t.Error("Expecting: Basic realm=Restricted, received: ", rr.Result().Header.Get("WWW-Authenticate"))
	}

	//part 4 -> <BasicAuth> Unauthorized API access by user
	runned = false
	req, err = http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Get "+base64.StdEncoding.EncodeToString([]byte("1001:1235")))
	rr = httptest.NewRecorder()

	toTest.ServeHTTP(rr, req)
	if runned {
		t.Error("ResponseWrite error")
	}
	if rr.Result().Header.Get("WWW-Authenticate") != `Basic realm="Restricted"` {
		t.Error("Expecting: Basic realm=Restricted, received: ", rr.Result().Header.Get("WWW-Authenticate"))
	}
}

func TestVerifyCredential(t *testing.T) {
	var hashedPasswords = map[string]string{
		"1234": "MTIzNA==",
		"bar":  "YmFy",
	}

	var verifyCredentialTests = []struct {
		username string
		password string
		userList map[string]string
		result   bool
	}{
		{"test", "1234", map[string]string{"test": hashedPasswords["1234"]}, true},
		{"test", "0000", map[string]string{"test": hashedPasswords["1234"]}, false},
		{"foo", "bar", map[string]string{"test": "1234", "foo": hashedPasswords["bar"]}, true},
		{"foo", "1234", map[string]string{"test": "1234", "foo": hashedPasswords["bar"]}, false},
		{"none", "1234", map[string]string{"test": "1234", "foo": hashedPasswords["bar"]}, false},
		{"test", "1234", map[string]string{"test": "12340", "foo": hashedPasswords["bar"]}, false},
	}

	for _, tt := range verifyCredentialTests {
		r := verifyCredential(tt.username, tt.password, tt.userList)
		if r != tt.result {
			t.Errorf("verifyCredential(%s, %s, %v) => %t, want %t", tt.username, tt.password, tt.userList, r, tt.result)
		}
	}
}
