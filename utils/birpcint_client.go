// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"net"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/jsonrpc"
)

// NewBiJSONrpcClient will create a bidirectional JSON client connection
func NewBiJSONrpcClient(addr string, obj birpc.ClientConnector) (*birpc.BirpcClient, error) {
	conn, err := net.Dial(TCP, addr)
	if err != nil {
		return nil, err
	}
	clnt := birpc.NewBirpcClientWithCodec(jsonrpc.NewJSONBirpcCodec(conn))
	if obj != nil {
		clnt.Register(obj)
	}
	return clnt, nil
}
