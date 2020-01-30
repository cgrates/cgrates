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
	"net"
	"reflect"
	"testing"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/rpcclient"
)

func TestNewBiJSONrpcClient(t *testing.T) {
	//empty check
	addr := "127.0.0.1:4024"
	handlers := map[string]interface{}{}
	rcv, err := NewBiJSONrpcClient(addr, handlers)
	if err == nil || rcv != nil {
		t.Error("Expencting: \"connection refused\", received : nil")
	}

	l, err := net.Listen(TCP, addr)
	if err != nil {
		t.Error(err)
	}
	handlers = map[string]interface{}{
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
	args  interface{}
	reply interface{}
}

func (*testBiRPCServer) Call(string, interface{}, interface{}) error { return nil }
func (t *testBiRPCServer) CallBiRPC(_ rpcclient.ClientConnector, metod string, args interface{}, reply interface{}) error {
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

	err := rcv.Call(APIerSv1ComputeActionPlanIndexes, "arg1", "reply")
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
