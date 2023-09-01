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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

func newCapsGOBCodec(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) (r birpc.ServerCodec) {
	r = newCapsServerCodec(birpc.NewServerCodec(conn), caps)
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

func newCapsJSONCodec(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) (r birpc.ServerCodec) {
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

func newCapsServerCodec(sc birpc.ServerCodec, caps *engine.Caps) birpc.ServerCodec {
	if !caps.IsLimited() {
		return sc
	}
	return &capsServerCodec{
		sc:   sc,
		caps: caps,
	}
}

type capsServerCodec struct {
	sc   birpc.ServerCodec
	caps *engine.Caps
}

func (c *capsServerCodec) ReadRequestHeader(r *birpc.Request) error {
	return c.sc.ReadRequestHeader(r)
}

func (c *capsServerCodec) ReadRequestBody(x any) error {
	if err := c.caps.Allocate(); err != nil {
		return err
	}
	return c.sc.ReadRequestBody(x)
}
func (c *capsServerCodec) WriteResponse(r *birpc.Response, x any) error {
	if r.Error == utils.ErrMaxConcurrentRPCExceededNoCaps.Error() {
		r.Error = utils.ErrMaxConcurrentRPCExceeded.Error()
	} else {
		defer c.caps.Deallocate()
	}
	return c.sc.WriteResponse(r, x)
}
func (c *capsServerCodec) Close() error { return c.sc.Close() }

func newCapsBiRPCGOBCodec(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) (r birpc.BirpcCodec) {
	r = newCapsBiRPCCodec(birpc.NewGobBirpcCodec(conn), caps)
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
		return analyzers.NewAnalyzerBiRPCCodec(r, anz, rpcclient.BiRPCGOB, fromstr, tostr)
	}
	return
}

func newCapsBiRPCJSONCodec(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) (r birpc.BirpcCodec) {
	r = newCapsBiRPCCodec(jsonrpc.NewJSONBirpcCodec(conn), caps)
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
		return analyzers.NewAnalyzerBiRPCCodec(r, anz, rpcclient.BiRPCJSON, fromstr, tostr)
	}
	return
}

func newCapsBiRPCCodec(sc birpc.BirpcCodec, caps *engine.Caps) birpc.BirpcCodec {
	if !caps.IsLimited() {
		return sc
	}
	return &capsBiRPCCodec{
		sc:   sc,
		caps: caps,
	}
}

type capsBiRPCCodec struct {
	sc   birpc.BirpcCodec
	caps *engine.Caps
}

// ReadHeader must read a message and populate either the request
// or the response by inspecting the incoming message.
func (c *capsBiRPCCodec) ReadHeader(req *birpc.Request, resp *birpc.Response) (err error) {
	if err = c.sc.ReadHeader(req, resp); err != nil ||
		req.ServiceMethod == utils.EmptyString { // caps will not process replies
		return
	}
	if err = c.caps.Allocate(); err != nil {
		req.ServiceMethod = utils.SessionSv1CapsError
		err = nil
	}
	return
}

// ReadRequestBody into args argument of handler function.
func (c *capsBiRPCCodec) ReadRequestBody(x any) (err error) {
	return c.sc.ReadRequestBody(x)
}

// ReadResponseBody into reply argument of handler function.
func (c *capsBiRPCCodec) ReadResponseBody(x any) error {
	return c.sc.ReadResponseBody(x)
}

// WriteRequest must be safe for concurrent use by multiple goroutines.
func (c *capsBiRPCCodec) WriteRequest(req *birpc.Request, x any) error {
	return c.sc.WriteRequest(req, x)
}

// WriteResponse must be safe for concurrent use by multiple goroutines.
func (c *capsBiRPCCodec) WriteResponse(r *birpc.Response, x any) error {
	if r.Error == utils.ErrMaxConcurrentRPCExceededNoCaps.Error() {
		r.Error = utils.ErrMaxConcurrentRPCExceeded.Error()
	} else {
		defer c.caps.Deallocate()
	}
	return c.sc.WriteResponse(r, x)
}

// Close is called when client/server finished with the connection.
func (c *capsBiRPCCodec) Close() error { return c.sc.Close() }
