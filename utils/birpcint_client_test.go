// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"net"
	"reflect"
	"testing"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
)

func TestNewBiJSONrpcClient(t *testing.T) {
	//empty check
	addr := "127.0.0.1:4024"
	handlers := map[string]any{}
	rcv, err := NewBiJSONrpcClient(addr, handlers)
	if err == nil || rcv != nil {
		t.Error("Expencting: \"connection refused\", received : nil")
	}

	l, err := net.Listen(TCP, addr)
	if err != nil {
		t.Error(err)
	}
	handlers = map[string]any{
		"": func(*rpc2.Client, *struct{}, *string) error { return nil },
	}

	rcv, err = NewBiJSONrpcClient(addr, handlers)
	if err != nil {
		t.Error(err)
	}
	l.Close()
}

type testBiRPCServer struct {
	metod string
	args  any
	reply any
}

func (*testBiRPCServer) Call(*context.Context, string, any, any) error { return nil }
func (t *testBiRPCServer) CallBiRPC(_ birpc.ClientConnector, metod string, args any, reply any) error {
	t.metod = metod
	t.args = args
	t.reply = reply
	return nil
}

func TestNewBiRPCInternalClient(t *testing.T) {
	//empty check

	rpc := &testBiRPCServer{}
	eOut := &BiRPCInternalClient{serverConn: rpc}
	rcv := NewBiRPCInternalClient(rpc)
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}

	rcv.SetClientConn(&testBiRPCServer{})

	if rcv.clntConn == nil {
		t.Error("Client Connection must be not nil")
	}

	err := rcv.Call(context.Background(), APIerSv1ComputeActionPlanIndexes, "arg1", "reply")
	if err != nil {
		t.Error(err)
	}
	testrpc := &testBiRPCServer{
		metod: APIerSv1ComputeActionPlanIndexes,
		args:  "arg1",
		reply: "reply",
	}
	if !reflect.DeepEqual(testrpc, rpc) {
		t.Errorf("Expecting: %+v, received: %+v", testrpc, rpc)
	}

}
