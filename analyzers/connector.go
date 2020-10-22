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

	"github.com/cgrates/rpcclient"
)

func NewAnalyzeConnector(sc rpcclient.ClientConnector) rpcclient.ClientConnector {
	return &AnalyzeConnector{conn: sc}
}

type AnalyzeConnector struct {
	conn  rpcclient.ClientConnector
	seq   uint64
	seqLk sync.Mutex
}

func (c *AnalyzeConnector) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	c.seqLk.Lock()
	id := c.seq
	c.seq++
	c.seqLk.Unlock()
	go h.handleRequest(id, serviceMethod, args)
	err = c.conn.Call(serviceMethod, args, reply)
	go h.handleResponse(id, reply, err)
	return
}
