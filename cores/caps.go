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

package cores

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

func newCapsGOBCodec(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) (r rpc.ServerCodec) {
	r = newCapsServerCodec(newGobServerCodec(conn), caps)
	if anz != nil {
		from := conn.RemoteAddr()
		var fromstr string
		if from != nil {
			fromstr = from.String()
		}
		to := conn.LocalAddr()
		var tostr string
		if to != nil {
			tostr = to.String()
		}
		return analyzers.NewServerCodec(r, anz, utils.MetaGOB, fromstr, tostr)
	}
	return
}

func newCapsJSONCodec(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) (r rpc.ServerCodec) {
	r = newCapsServerCodec(jsonrpc.NewServerCodec(conn), caps)
	if anz != nil {
		from := conn.RemoteAddr()
		var fromstr string
		if from != nil {
			fromstr = from.String()
		}
		to := conn.LocalAddr()
		var tostr string
		if to != nil {
			tostr = to.String()
		}
		return analyzers.NewServerCodec(r, anz, utils.MetaJSON, fromstr, tostr)
	}
	return
}

func newCapsServerCodec(sc rpc.ServerCodec, caps *engine.Caps) rpc.ServerCodec {
	if !caps.IsLimited() {
		return sc
	}
	return &capsServerCodec{
		sc:   sc,
		caps: caps,
	}
}

type capsServerCodec struct {
	sc   rpc.ServerCodec
	caps *engine.Caps
}

func (c *capsServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.sc.ReadRequestHeader(r)
}

func (c *capsServerCodec) ReadRequestBody(x interface{}) error {
	if err := c.caps.Allocate(); err != nil {
		return err
	}
	return c.sc.ReadRequestBody(x)
}
func (c *capsServerCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	if r.Error == utils.ErrMaxConcurentRPCExceededNoCaps.Error() {
		r.Error = utils.ErrMaxConcurentRPCExceeded.Error()
	} else {
		defer c.caps.Deallocate()
	}
	return c.sc.WriteResponse(r, x)
}
func (c *capsServerCodec) Close() error { return c.sc.Close() }
