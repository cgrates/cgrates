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
	"net/http"
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
