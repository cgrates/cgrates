/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package main

import "testing"

func TestCgrConsoleFlags(t *testing.T) {
	if err := cgrConsoleFlags.Parse([]string{"-version", "true"}); err != nil {
		t.Fatal(err)
	} else if *version != true {
		t.Errorf("Expected true, received %+v", *version)
	}

	if err := cgrConsoleFlags.Parse([]string{"-verbose", "true"}); err != nil {
		t.Fatal(err)
	} else if *verbose != true {
		t.Errorf("Expected true, received %+v", *version)
	}

	if err := cgrConsoleFlags.Parse([]string{"-server", "192.168.100.2:8080"}); err != nil {
		t.Fatal(err)
	} else if *server != "192.168.100.2:8080" {
		t.Errorf("Expected 192.168.100.2:8080 but received %+v", *server)
	}

	if err := cgrConsoleFlags.Parse([]string{"-rpc_encoding", "*birpc"}); err != nil {
		t.Fatal(err)
	} else if *rpcEncoding != "*birpc" {
		t.Errorf("Expected *birpc but received %+v", *rpcEncoding)
	}

	if err := cgrConsoleFlags.Parse([]string{"-crt_path", "/tmp"}); err != nil {
		t.Fatal(err)
	} else if *certificatePath != "/tmp" {
		t.Errorf("Expected /tmp but received %+v", *rpcEncoding)
	}

	if err := cgrConsoleFlags.Parse([]string{"-key_path", "/tmp"}); err != nil {
		t.Fatal(err)
	} else if *keyPath != "/tmp" {
		t.Errorf("Expected /tmp but received %+v", *rpcEncoding)
	}

	if err := cgrConsoleFlags.Parse([]string{"-ca_path", "/tmp"}); err != nil {
		t.Fatal(err)
	} else if *caPath != "/tmp" {
		t.Errorf("Expected /tmp but received %+v", *rpcEncoding)
	}

	if err := cgrConsoleFlags.Parse([]string{"-tls", "true"}); err != nil {
		t.Fatal(err)
	} else if *tls != true {
		t.Errorf("Expected true but received %+v", *rpcEncoding)
	}

	if err := cgrConsoleFlags.Parse([]string{"-reply_timeout", "200"}); err != nil {
		t.Fatal(err)
	} else if *replyTimeOut != 200 {
		t.Errorf("Expected 200 but received %+v", *rpcEncoding)
	}
}
