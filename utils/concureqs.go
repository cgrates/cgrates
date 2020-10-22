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
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
)

var ConReqs *ConcReqs

type ConcReqs struct {
	limit    int
	strategy string
	aReqs    chan struct{}
}

func NewConReqs(reqs int, strategy string) *ConcReqs {
	cR := &ConcReqs{
		limit:    reqs,
		strategy: strategy,
		aReqs:    make(chan struct{}, reqs),
	}
	for i := 0; i < reqs; i++ {
		cR.aReqs <- struct{}{}
	}
	return cR
}

// IsLimited returns true if the limit is not 0
func (cR *ConcReqs) IsLimited() bool {
	return ConReqs.limit != 0
}

// Allocate will reserve a channel for the API call
func (cR *ConcReqs) Allocate() (err error) {
	switch cR.strategy {
	case MetaBusy:
		if len(cR.aReqs) == 0 {
			return ErrMaxConcurentRPCExceededNoCaps
		}
		fallthrough
	case MetaQueue:
		<-cR.aReqs // get from channel
	}
	return
}

// Deallocate will free a channel for the API call
func (cR *ConcReqs) Deallocate() {
	cR.aReqs <- struct{}{}
	return
}

func newConcReqsGOBCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return newConcReqsServerCodec(newGobServerCodec(conn))
}

func newConcReqsJSONCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return newConcReqsServerCodec(jsonrpc.NewServerCodec(conn))
}

func newConcReqsServerCodec(sc rpc.ServerCodec) rpc.ServerCodec {
	if !ConReqs.IsLimited() {
		return sc
	}
	return &concReqsServerCodec{sc: sc}
}

type concReqsServerCodec struct {
	sc rpc.ServerCodec
}

func (c *concReqsServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.sc.ReadRequestHeader(r)
}

func (c *concReqsServerCodec) ReadRequestBody(x interface{}) error {
	if err := ConReqs.Allocate(); err != nil {
		return err
	}
	return c.sc.ReadRequestBody(x)
}
func (c *concReqsServerCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	if r.Error == ErrMaxConcurentRPCExceededNoCaps.Error() {
		r.Error = ErrMaxConcurentRPCExceeded.Error()
	} else {
		defer ConReqs.Deallocate()
	}
	return c.sc.WriteResponse(r, x)
}
func (c *concReqsServerCodec) Close() error { return c.sc.Close() }
