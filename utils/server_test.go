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
