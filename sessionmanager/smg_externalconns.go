/*
Real-time Charging System for Telecom & ISP environments
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

package sessionmanager

import (
	"errors"
	"sync"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/utils"
)

const CGR_CONNUUID = "cgr_connid"

var ErrConnectionNotFound = errors.New("CONNECTION_NOT_FOUND")

// Attempts to get the connId previously set in the client state container
func getClientConnId(clnt *rpc2.Client) string {
	uuid, hasIt := clnt.State.Get(CGR_CONNUUID)
	if !hasIt {
		return ""
	}
	return uuid.(string)
}

func NewSMGExternalConnections() *SMGExternalConnections {
	return &SMGExternalConnections{conns: make(map[string]*rpc2.Client), connMux: new(sync.Mutex)}
}

type SMGExternalConnections struct {
	conns   map[string]*rpc2.Client
	connMux *sync.Mutex
}

// Index the client connection so we can use it to communicate back
func (self *SMGExternalConnections) OnClientConnect(clnt *rpc2.Client) {
	self.connMux.Lock()
	defer self.connMux.Unlock()
	connId := utils.GenUUID()
	clnt.State.Set(CGR_CONNUUID, connId) // Set unique id for the connection so we can identify it later in requests
	self.conns[connId] = clnt
}

// Unindex the client connection so we can use it to communicate back
func (self *SMGExternalConnections) OnClientDisconnect(clnt *rpc2.Client) {
	self.connMux.Lock()
	defer self.connMux.Unlock()
	if connId := getClientConnId(clnt); connId != "" {
		delete(self.conns, connId)
	}
}

func (self *SMGExternalConnections) GetConnection(connId string) *rpc2.Client {
	self.connMux.Lock()
	defer self.connMux.Unlock()
	return self.conns[connId]
}
