// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"net"

	"github.com/cenkalti/rpc2"
	rpc2_jsonrpc "github.com/cenkalti/rpc2/jsonrpc"
	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
)

// NewBiJSONrpcClient will create a bidirectional JSON client connection
func NewBiJSONrpcClient(addr string, handlers map[string]any) (*rpc2.Client, error) {
	conn, err := net.Dial(TCP, addr)
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
	Call(*context.Context, string, any, any) error // So we can use it also as birpc.ClientConnector
	CallBiRPC(birpc.ClientConnector, string, any, any) error
}

type BiRPCClient interface {
	Call(*context.Context, string, any, any) error // So we can use it also as birpc.ClientConnector
	ID() string
}

func NewBiRPCInternalClient(serverConn BiRPCServer) *BiRPCInternalClient {
	return &BiRPCInternalClient{serverConn: serverConn}
}

// Need separate client from the original RpcClientConnection since diretly passing the server is not enough without passing the client's reference
type BiRPCInternalClient struct {
	serverConn BiRPCServer
	clntConn   birpc.ClientConnector // conn to reach client and do calls over it
}

// Used in case when clientConn is not available at init time (eg: SMGAsterisk who needs the biRPCConn at initialization)
func (clnt *BiRPCInternalClient) SetClientConn(clntConn birpc.ClientConnector) {
	clnt.clntConn = clntConn
}

// Part of birpc.ClientConnector interface
func (clnt *BiRPCInternalClient) Call(ctx *context.Context, serviceMethod string, args any, reply any) error {
	return clnt.serverConn.CallBiRPC(clnt.clntConn, serviceMethod, args, reply)
}
