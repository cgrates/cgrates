// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
)

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

type mockConnector struct{}

func (c *mockConnector) Call(_ *context.Context, _ string, _, _ any) (err error) {
	return errors.New("error")
}

func TestNewBiJSONrpcClient(t *testing.T) {

	lis, err := net.Listen(TCP, "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	defer lis.Close()
	adr := lis.Addr().String()

	tests := []struct {
		name    string
		addr    string
		obj     birpc.ClientConnector
		wantErr bool
	}{
		{
			name:    "True connection",
			addr:    adr,
			obj:     new(mockConnector),
			wantErr: false,
		},
		{
			name:    "Nil obj",
			addr:    adr,
			obj:     nil,
			wantErr: false,
		},
		{
			name:    "Empty addr",
			addr:    "",
			obj:     new(mockConnector),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBiJSONrpcClient(tt.addr, tt.obj)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewBiJSONrpcClient() failed: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewBiJSONrpcClient() succeeded unexpectedly")
			}

			if got == nil {
				t.Error(err)
			}

		})
	}
}
func TestOnBiJSONConnectDisconnect(t *testing.T) {
	sClinents := NewServiceBiRPCClients()
	//connect BiJSON
	client := &birpc.Service{}
	sClinents.OnBiJSONConnect(client, 0)

	//we'll change the connection identifier just for testing
	sClinents.biJClnts[client] = "test_conn"
	sClinents.biJIDs = nil

	expected := NewServiceBiRPCClients()
	expected.biJClnts[client] = "test_conn"
	expected.biJIDs = nil

	if !reflect.DeepEqual(sClinents, expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, sClinents)
	}

	//Disconnect BiJSON
	sClinents.OnBiJSONDisconnect(client)
	delete(expected.biJClnts, client)
	if !reflect.DeepEqual(sClinents, expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, sClinents)
	}
}

type testRPCClientConnection struct{}

func (*testRPCClientConnection) Call(string, any, any) error { return nil }

type mockConnWarnDisconnect1 struct {
	*testRPCClientConnection
}

func (mk *mockConnWarnDisconnect1) Call(ctx *context.Context, method string, args any, rply any) error {
	return ErrNotImplemented
}

func TestBiJClntID(t *testing.T) {
	client := &mockConnWarnDisconnect1{}
	sClinents := NewServiceBiRPCClients()
	sClinents.biJClnts = map[birpc.ClientConnector]string{
		client: "First_connector",
	}
	expected := "First_connector"
	if rcv := sClinents.BiJClntID(client); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}
