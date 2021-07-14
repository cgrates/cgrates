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

	"github.com/cenkalti/rpc2"
	jsonrpc2 "github.com/cenkalti/rpc2/jsonrpc"
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
		return analyzers.NewAnalyzerServerCodec(r, anz, utils.MetaGOB, fromstr, tostr)
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
		return analyzers.NewAnalyzerServerCodec(r, anz, utils.MetaJSON, fromstr, tostr)
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

func newCapsBiRPCGOBCodec(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) (r rpc2.Codec) {
	r = newCapsBiRPCCodec(rpc2.NewGobCodec(conn), caps)
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
		return analyzers.NewAnalyzerBiRPCCodec(r, anz, utils.MetaGOB, fromstr, tostr)
	}
	return
}

func newCapsBiRPCJSONCodec(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) (r rpc2.Codec) {
	r = newCapsBiRPCCodec(jsonrpc2.NewJSONCodec(conn), caps)
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
		return analyzers.NewAnalyzerBiRPCCodec(r, anz, utils.MetaJSON, fromstr, tostr)
	}
	return
}

func newCapsBiRPCCodec(sc rpc2.Codec, caps *engine.Caps) rpc2.Codec {
	if !caps.IsLimited() {
		return sc
	}
	return &capsBiRPCCodec{
		sc:   sc,
		caps: caps,
	}
}

type capsBiRPCCodec struct {
	sc   rpc2.Codec
	caps *engine.Caps
}

// ReadHeader must read a message and populate either the request
// or the response by inspecting the incoming message.
func (c *capsBiRPCCodec) ReadHeader(req *rpc2.Request, resp *rpc2.Response) (err error) {
	return c.sc.ReadHeader(req, resp)
}

// ReadRequestBody into args argument of handler function.
func (c *capsBiRPCCodec) ReadRequestBody(x interface{}) (err error) {
	if err = c.caps.Allocate(); err != nil {
		return
	}
	return c.sc.ReadRequestBody(x)
}

// ReadResponseBody into reply argument of handler function.
func (c *capsBiRPCCodec) ReadResponseBody(x interface{}) error {
	return c.sc.ReadResponseBody(x)
}

// WriteRequest must be safe for concurrent use by multiple goroutines.
func (c *capsBiRPCCodec) WriteRequest(req *rpc2.Request, x interface{}) error {
	return c.sc.WriteRequest(req, x)
}

// WriteResponse must be safe for concurrent use by multiple goroutines.
func (c *capsBiRPCCodec) WriteResponse(r *rpc2.Response, x interface{}) error {
	if r.Error == utils.ErrMaxConcurentRPCExceededNoCaps.Error() {
		r.Error = utils.ErrMaxConcurentRPCExceeded.Error()
	} else {
		defer c.caps.Deallocate()
	}
	return c.sc.WriteResponse(r, x)
}

// Close is called when client/server finished with the connection.
func (c *capsBiRPCCodec) Close() error { return c.sc.Close() }
