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
	"sync"
	"time"
)

func NewAnalyzerServerCodec(sc rpc.ServerCodec, aS *AnalyzerService, enc, from, to string) rpc.ServerCodec {
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
	sc rpc.ServerCodec

	// keep the API in memory because the write is async
	reqs   map[uint64]*rpcAPI
	reqIdx uint64
	reqsLk sync.RWMutex
	aS     *AnalyzerService
	enc    string
	from   string
	to     string
}

func (c *AnalyzerServerCodec) ReadRequestHeader(r *rpc.Request) (err error) {
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

func (c *AnalyzerServerCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	c.reqsLk.Lock()
	api := c.reqs[r.Seq]
	delete(c.reqs, r.Seq)
	c.reqsLk.Unlock()
	go c.aS.logTrafic(api.ID, api.Method, api.Params, x, r.Error, c.enc, c.from, c.to, api.StartTime, time.Now())
	return c.sc.WriteResponse(r, x)
}

func (c *AnalyzerServerCodec) Close() error { return c.sc.Close() }
