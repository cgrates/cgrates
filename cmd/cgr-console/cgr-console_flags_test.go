// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	} else if *replyTimeout != 200 {
		t.Errorf("Expected 200 but received %+v", *rpcEncoding)
	}
}
