/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"maps"
	"net"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/jsonrpc"
)

// NewBiJSONrpcClient will create a bidirectional JSON client connection
func NewBiJSONrpcClient(addr string, obj birpc.ClientConnector) (*birpc.BirpcClient, error) {
	conn, err := net.Dial(TCP, addr)
	if err != nil {
		return nil, err
	}
	clnt := birpc.NewBirpcClientWithCodec(jsonrpc.NewJSONBirpcCodec(conn))
	if obj != nil {
		clnt.Register(obj)
	}
	return clnt, nil
}

type ServiceBiRPCClients struct {
	biJMux   sync.RWMutex                     // mux protecting BI-JSON connections
	biJClnts map[birpc.ClientConnector]string // index BiJSONConnection so we can sync them later
	biJIDs   map[string]*biJClient            // identifiers of bidirectional JSON conns, used to call RPC based on connIDs
}

func NewServiceBiRPCClients() *ServiceBiRPCClients {
	return &ServiceBiRPCClients{
		biJClnts: make(map[birpc.ClientConnector]string),
		biJIDs:   make(map[string]*biJClient),
	}
}

// biJClient contains info we need to reach back a bidirectional json client
type biJClient struct {
	conn  birpc.ClientConnector // connection towards BiJ client
	proto float64               // client protocol version
}

// Proto exports proto field from biJClient b (only used for SessionS)
func (b *biJClient) Proto() float64 {
	return b.proto
}

// Conn exports conn field from biJClient b
func (b *biJClient) Conn() birpc.ClientConnector {
	return b.conn
}

// OnBiJSONConnect handles new client connections.
func (s *ServiceBiRPCClients) OnBiJSONConnect(c birpc.ClientConnector, clientProtocol float64) {
	nodeID := UUIDSha1Prefix() // connection identifier, should be later updated as login procedure
	s.biJMux.Lock()
	s.biJClnts[c] = nodeID
	s.biJIDs[nodeID] = &biJClient{conn: c, proto: clientProtocol}
	s.biJMux.Unlock()
}

// OnBiJSONDisconnect handles client disconnects.
func (s *ServiceBiRPCClients) OnBiJSONDisconnect(c birpc.ClientConnector) {
	s.biJMux.Lock()
	if nodeID, has := s.biJClnts[c]; has {
		delete(s.biJClnts, c)
		delete(s.biJIDs, nodeID)
	}
	s.biJMux.Unlock()
}

// RegisterIntBiJConn is called on internal BiJ connection towards Service
func (s *ServiceBiRPCClients) RegisterIntBiJConn(c birpc.ClientConnector, nodeID string, clientProtocol float64) {
	s.biJMux.Lock()
	s.biJClnts[c] = nodeID
	s.biJIDs[nodeID] = &biJClient{conn: c, proto: clientProtocol}
	s.biJMux.Unlock()
}

// biJClnt returns a bidirectional JSON client based on connection ID
func (s *ServiceBiRPCClients) BiJClnt(connID string) (clnt *biJClient) {
	if connID == "" {
		return
	}
	s.biJMux.RLock()
	clnt = s.biJIDs[connID]
	s.biJMux.RUnlock()
	return
}

// biJClnt returns connection ID based on bidirectional connection received
func (s *ServiceBiRPCClients) BiJClntID(c birpc.ClientConnector) (clntConnID string) {
	if c == nil {
		return
	}
	s.biJMux.RLock()
	clntConnID = s.biJClnts[c]
	s.biJMux.RUnlock()
	return
}

// biJClnts is a thread-safe method to return the list of active clients for BiJson
func (s *ServiceBiRPCClients) BiJClients() (clnts []*biJClient) {
	s.biJMux.RLock()
	clnts = make([]*biJClient, len(s.biJIDs))
	i := 0
	for _, clnt := range s.biJIDs {
		clnts[i] = clnt
		i++
	}
	s.biJMux.RUnlock()
	return
}

// BiJClientsMap is a thread-safe method that returns a copy of biJIDs map of biJClients
func (s *ServiceBiRPCClients) BiJClientsMap() map[string]*biJClient {
	clients := make(map[string]*biJClient)
	s.biJMux.RLock()
	maps.Copy(clients, s.biJIDs)
	s.biJMux.RUnlock()
	return clients
}
