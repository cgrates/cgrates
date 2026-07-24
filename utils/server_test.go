// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"testing"
)

func TestServerTestMock(t *testing.T) {
	ln, err := net.Listen("tcp", ":0") // will pick a free port number automatically
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for range 2 {
			conn, err := ln.Accept()
			if err != nil {
				t.Error(err)
			}
			rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()
	client, err := jsonrpc.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	wantErr := "rpc: can't find service _goRPC_.Cancel"
	serviceMethod := "_goRPC_.Cancel"
	var reply bool
	if err = client.Call(serviceMethod, nil, &reply); err == nil || err.Error() != wantErr {
		t.Errorf("client.Call(%q) err = %v, want %v", serviceMethod, err, wantErr)
	}
	NewServer()
	if err = client.Call(serviceMethod, nil, &reply); err != nil {
		t.Fatalf("client.Call(%q) returned unexpected error: %v", serviceMethod, err)
	}
}
