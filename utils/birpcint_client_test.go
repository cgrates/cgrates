// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"net"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
)

type mockClientConn struct{}

func (mCC *mockClientConn) Call(ctx *context.Context, serviceMethod string, args, reply any) error {
	return nil
}
func TestBiRpcNewBiJSONrpcClient(t *testing.T) {
	rcv, err := NewBiJSONrpcClient("test", nil)

	if err != nil {
		if err.Error() != "dial tcp: address test: missing port in address" {
			t.Error(err)
		}
	}

	if rcv != nil {
		t.Error(err)
	}
}

func TestNewBiJSONrpcClient(t *testing.T) {
	lis, err := net.Listen(TCP, "127.0.0.1:2012")
	if err != nil {
		t.Fatal(err)
	}

	defer lis.Close()
	adr := lis.Addr().String()
	tests := []struct {
		name   string
		addr   string
		obj    birpc.ClientConnector
		expErr string
	}{
		{
			addr: adr,
			obj:  new(mockClientConn),
		},
		{
			name: "Nil obj",
			addr: adr,
			obj:  nil,
		},
		{
			name:   "Empty addr",
			addr:   "",
			obj:    new(mockClientConn),
			expErr: "dial tcp: missing address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBiJSONrpcClient(tt.addr, tt.obj)
			if err != nil {
				if err.Error() != tt.expErr {
					t.Errorf("Exected %v, recieved %v", tt.expErr, err)
				}
				return
			}
		})
	}
}
