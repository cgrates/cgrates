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

	"github.com/cenkalti/rpc2"
	rpc2_jsonrpc "github.com/cenkalti/rpc2/jsonrpc"
	"github.com/cgrates/rpcclient"
)

// NewBiJSONrpcClient will create a bidirectional JSON client connection
func NewBiJSONrpcClient(addr string, handlers map[string]interface{}) (*rpc2.Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	clnt := rpc2.NewClientWithCodec(rpc2_jsonrpc.NewJSONCodec(conn))
	for method, handlerFunc := range handlers {
		clnt.Handle(method, handlerFunc)
	}
	go clnt.Run()
	return clnt, nil
}

// Interface which the server needs to work as BiRPCServer
type BiRPCServer interface {
	Call(string, interface{}, interface{}) error // So we can use it also as rpcclient.RpcClientConnection
	CallBiRPC(rpcclient.RpcClientConnection, string, interface{}, interface{}) error
}

func NewBiRPCInternalClient(serverConn BiRPCServer) *BiRPCInternalClient {
	return &BiRPCInternalClient{serverConn: serverConn}
}

// Need separate client from the original RpcClientConnection since diretly passing the server is not enough without passing the client's reference
type BiRPCInternalClient struct {
	serverConn BiRPCServer
	clntConn   rpcclient.RpcClientConnection // conn to reach client and do calls over it
}

// Used in case when clientConn is not available at init time (eg: SMGAsterisk who needs the biRPCConn at initialization)
func (clnt *BiRPCInternalClient) SetClientConn(clntConn rpcclient.RpcClientConnection) {
	clnt.clntConn = clntConn
}

// Part of rpcclient.RpcClientConnection interface
func (clnt *BiRPCInternalClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return clnt.serverConn.CallBiRPC(clnt.clntConn, serviceMethod, args, reply)
}
