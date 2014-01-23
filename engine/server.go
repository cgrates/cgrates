/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package engine

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Server struct {
	rpcEnabled  bool
	httpEnabled bool
}

func (s *Server) RpcRegister(rcvr interface{}) {
	rpc.Register(rcvr)
	s.rpcEnabled = true
}

func (s *Server) RpcRegisterName(name string, rcvr interface{}) {
	rpc.RegisterName(name, rcvr)
	s.rpcEnabled = true
}

func (s *Server) RegisterHttpFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
	s.httpEnabled = true
}

func (s *Server) ServeJSON(addr string) {
	if !s.rpcEnabled {
		return
	}
	lJSON, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		conn, err := lJSON.Accept()
		if err != nil {
			Logger.Err(fmt.Sprintf("<CGRServer> Accept error: %v", conn))
			continue
		}

		//Logger.Info(fmt.Sprintf("<CGRServer> New incoming connection: %v", conn.RemoteAddr()))
		go jsonrpc.ServeConn(conn)
	}

}

func (s *Server) ServeGOB(addr string) {
	if !s.rpcEnabled {
		return
	}
	lGOB, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		conn, err := lGOB.Accept()
		if err != nil {
			Logger.Err(fmt.Sprintf("<CGRServer> Accept error: %v", conn))
			continue
		}

		//Logger.Info(fmt.Sprintf("<CGRServer> New incoming connection: %v", conn.RemoteAddr()))
		go rpc.ServeConn(conn)
	}
}

func (s *Server) ServeHTTP(addr string) {
	if !s.httpEnabled {
		return
	}
	http.ListenAndServe(addr, nil)
}
