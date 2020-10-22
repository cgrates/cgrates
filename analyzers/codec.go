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

package analyzers

import (
	"net/rpc"
)

func NewAnalyzeServerCodec(sc rpc.ServerCodec) rpc.ServerCodec {
	return &AnalyzeServerCodec{sc: sc, req: new(RPCServerRequest)}
}

type AnalyzeServerCodec struct {
	sc rpc.ServerCodec
	// keep the information about the header so we handle this when the body is readed
	// the ReadRequestHeader and ReadRequestBody are called in pairs
	req *RPCServerRequest
}

func (c *AnalyzeServerCodec) ReadRequestHeader(r *rpc.Request) (err error) {
	c.req.reset()
	err = c.sc.ReadRequestHeader(r)
	c.req.Method = r.ServiceMethod
	c.req.ID = r.Seq
	return
}

func (c *AnalyzeServerCodec) ReadRequestBody(x interface{}) (err error) {
	err = c.sc.ReadRequestBody(x)
	go h.handleRequest(c.req.ID, c.req.Method, x)
	return
}
func (c *AnalyzeServerCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	go h.handleResponse(r.Seq, x, r.Error)
	return c.sc.WriteResponse(r, x)
}
func (c *AnalyzeServerCodec) Close() error { return c.sc.Close() }
