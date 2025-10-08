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
