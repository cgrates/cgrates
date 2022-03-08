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
)

// NewBiJSONrpcClient will create a bidirectional JSON client connection
func NewBiJSONrpcClient(addr string, handlers map[string]interface{}) (*rpc2.Client, error) {
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
