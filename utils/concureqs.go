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
	"net/rpc"
	"net/rpc/jsonrpc"
)

// ConcReqs the structure that allocs requests for API
type ConcReqs struct {
	strategy string
	aReqs    chan struct{}
}

// NewConReqs creates a new ConcReqs
func NewConReqs(reqs int, strategy string) *ConcReqs {
	cR := &ConcReqs{
		strategy: strategy,
		aReqs:    make(chan struct{}, reqs),
	}
	return cR
}

// IsLimited returns true if the limit is not 0
func (cR *ConcReqs) IsLimited() bool {
	return cap(cR.aReqs) != 0
}

// Allocated returns the number of requests actively serviced
func (cR *ConcReqs) Allocated() int {
	return len(cR.aReqs)
}

// Allocate will reserve a channel for the API call
func (cR *ConcReqs) Allocate() (err error) {
	switch cR.strategy {
	case MetaBusy:
		if len(cR.aReqs) == cap(cR.aReqs) {
			return ErrMaxConcurentRPCExceededNoCaps
		}
		fallthrough
	case MetaQueue:
		cR.aReqs <- struct{}{}
	}
	return
}

// Deallocate will free a channel for the API call
func (cR *ConcReqs) Deallocate() {
	<-cR.aReqs
	return
}

type conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

func newConcReqsGOBCodec(conn conn, conReqs *ConcReqs, anz anzWrapFunc) rpc.ServerCodec {
	return anz(newConcReqsServerCodec(newGobServerCodec(conn), conReqs), MetaGOB, conn.RemoteAddr(), conn.LocalAddr())
}

func newConcReqsJSONCodec(conn conn, conReqs *ConcReqs, anz anzWrapFunc) rpc.ServerCodec {
	return anz(newConcReqsServerCodec(jsonrpc.NewServerCodec(conn), conReqs), MetaJSON, conn.RemoteAddr(), conn.LocalAddr())
}

func newConcReqsServerCodec(sc rpc.ServerCodec, conReqs *ConcReqs) rpc.ServerCodec {
	if !conReqs.IsLimited() {
		return sc
	}
	return &concReqsServerCodec{
		sc:      sc,
		conReqs: conReqs,
	}
}

type concReqsServerCodec struct {
	sc      rpc.ServerCodec
	conReqs *ConcReqs
}

func (c *concReqsServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.sc.ReadRequestHeader(r)
}

func (c *concReqsServerCodec) ReadRequestBody(x interface{}) error {
	if err := c.conReqs.Allocate(); err != nil {
		return err
	}
	return c.sc.ReadRequestBody(x)
}
func (c *concReqsServerCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	if r.Error == ErrMaxConcurentRPCExceededNoCaps.Error() {
		r.Error = ErrMaxConcurentRPCExceeded.Error()
	} else {
		defer c.conReqs.Deallocate()
	}
	return c.sc.WriteResponse(r, x)
}
func (c *concReqsServerCodec) Close() error { return c.sc.Close() }
