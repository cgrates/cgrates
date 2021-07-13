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
	"sync"
	"time"

	"github.com/cgrates/birpc"
)

func NewAnalyzerServerCodec(sc birpc.ServerCodec, aS *AnalyzerService, enc, from, to string) birpc.ServerCodec {
	return &AnalyzerServerCodec{
		sc:   sc,
		reqs: make(map[uint64]*rpcAPI),
		aS:   aS,
		enc:  enc,
		from: from,
		to:   to,
	}
}

type AnalyzerServerCodec struct {
	sc birpc.ServerCodec

	// keep the API in memory because the write is async
	reqs   map[uint64]*rpcAPI
	reqIdx uint64
	reqsLk sync.RWMutex
	aS     *AnalyzerService
	enc    string
	from   string
	to     string
}

func (c *AnalyzerServerCodec) ReadRequestHeader(r *birpc.Request) (err error) {
	err = c.sc.ReadRequestHeader(r)
	c.reqsLk.Lock()
	c.reqIdx = r.Seq
	c.reqs[c.reqIdx] = &rpcAPI{
		ID:        r.Seq,
		Method:    r.ServiceMethod,
		StartTime: time.Now(),
	}
	c.reqsLk.Unlock()
	return
}

func (c *AnalyzerServerCodec) ReadRequestBody(x interface{}) (err error) {
	err = c.sc.ReadRequestBody(x)
	c.reqsLk.Lock()
	c.reqs[c.reqIdx].Params = x
	c.reqsLk.Unlock()
	return
}

func (c *AnalyzerServerCodec) WriteResponse(r *birpc.Response, x interface{}) error {
	c.reqsLk.Lock()
	api := c.reqs[r.Seq]
	delete(c.reqs, r.Seq)
	c.reqsLk.Unlock()
	go c.aS.logTrafic(api.ID, api.Method, api.Params, x, r.Error, c.enc, c.from, c.to, api.StartTime, time.Now())
	return c.sc.WriteResponse(r, x)
}

func (c *AnalyzerServerCodec) Close() error { return c.sc.Close() }

func NewAnalyzerBiRPCCodec(sc birpc.BirpcCodec, aS *AnalyzerService, enc, from, to string) birpc.BirpcCodec {
	return &AnalyzerBiRPCCodec{
		sc:   sc,
		reqs: make(map[uint64]*rpcAPI),
		reps: make(map[uint64]*rpcAPI),
		aS:   aS,
		enc:  enc,
		from: from,
		to:   to,
	}
}

type AnalyzerBiRPCCodec struct {
	sc birpc.BirpcCodec

	// keep the API in memory because the write is async
	reqs   map[uint64]*rpcAPI
	reqIdx uint64
	reqsLk sync.RWMutex
	reps   map[uint64]*rpcAPI
	repIdx uint64
	repsLk sync.RWMutex

	aS   *AnalyzerService
	enc  string
	from string
	to   string
}

// ReadHeader must read a message and populate either the request
// or the response by inspecting the incoming message.
func (c *AnalyzerBiRPCCodec) ReadHeader(req *birpc.Request, resp *birpc.Response) (err error) {
	err = c.sc.ReadHeader(req, resp)
	if req.ServiceMethod != "" {
		c.reqsLk.Lock()
		c.reqIdx = req.Seq
		c.reqs[c.reqIdx] = &rpcAPI{
			ID:        req.Seq,
			Method:    req.ServiceMethod,
			StartTime: time.Now(),
		}
		c.reqsLk.Unlock()
	} else {
		c.repsLk.Lock()
		c.repIdx = resp.Seq
		c.reps[c.repIdx].Error = resp.Error
		c.repsLk.Unlock()
	}
	return
}

// ReadRequestBody into args argument of handler function.
func (c *AnalyzerBiRPCCodec) ReadRequestBody(x interface{}) (err error) {
	err = c.sc.ReadRequestBody(x)
	c.reqsLk.Lock()
	c.reqs[c.reqIdx].Params = x
	c.reqsLk.Unlock()
	return
}

// ReadResponseBody into reply argument of handler function.
func (c *AnalyzerBiRPCCodec) ReadResponseBody(x interface{}) (err error) {
	err = c.sc.ReadResponseBody(x)
	c.repsLk.Lock()
	api := c.reqs[c.repIdx]
	delete(c.reqs, c.repIdx)
	c.repsLk.Unlock()
	go c.aS.logTrafic(api.ID, api.Method, api.Params, x, api.Error, c.enc, c.to, c.from, api.StartTime, time.Now())
	return
}

// WriteRequest must be safe for concurrent use by multiple goroutines.
func (c *AnalyzerBiRPCCodec) WriteRequest(req *birpc.Request, x interface{}) error {
	c.repsLk.Lock()
	c.reqIdx = req.Seq
	c.reqs[c.reqIdx] = &rpcAPI{
		ID:        req.Seq,
		Method:    req.ServiceMethod,
		StartTime: time.Now(),
	}
	c.repsLk.Unlock()
	return c.sc.WriteRequest(req, x)
}

// WriteResponse must be safe for concurrent use by multiple goroutines.
func (c *AnalyzerBiRPCCodec) WriteResponse(r *birpc.Response, x interface{}) error {
	c.reqsLk.Lock()
	api := c.reqs[r.Seq]
	delete(c.reqs, r.Seq)
	c.reqsLk.Unlock()
	go c.aS.logTrafic(api.ID, api.Method, api.Params, x, r.Error, c.enc, c.from, c.to, api.StartTime, time.Now())
	return c.sc.WriteResponse(r, x)
}

// Close is called when client/server finished with the connection.
func (c *AnalyzerBiRPCCodec) Close() error { return c.sc.Close() }
