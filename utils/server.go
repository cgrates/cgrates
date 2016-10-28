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
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
	"time"

	"github.com/cenk/rpc2"
	"golang.org/x/net/websocket"
)
import _ "net/http/pprof"

type Server struct {
	rpcEnabled  bool
	httpEnabled bool
	birpcSrv    *rpc2.Server
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

// Registers a new BiJsonRpc name
func (s *Server) BiRPCRegisterName(method string, handlerFunc interface{}) {
	if s.birpcSrv == nil {
		s.birpcSrv = rpc2.NewServer()
	}
	s.birpcSrv.Handle(method, handlerFunc)
}

func (s *Server) BiRPCRegister(rcvr interface{}) {
	if s.birpcSrv == nil {
		s.birpcSrv = rpc2.NewServer()
	}
	rcvType := reflect.TypeOf(rcvr)
	for i := 0; i < rcvType.NumMethod(); i++ {
		method := rcvType.Method(i)
		if method.Name != "Call" {
			s.birpcSrv.Handle("SMGenericV1."+method.Name, method.Func.Interface())
		}
	}
}

func (s *Server) ServeJSON(addr string) {
	if !s.rpcEnabled {
		return
	}
	lJSON, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("ServeJSON listen error:", e)
	}
	Logger.Info(fmt.Sprintf("Starting CGRateS JSON server at <%s>.", addr))
	errCnt := 0
	var lastErrorTime time.Time
	for {
		conn, err := lJSON.Accept()
		if err != nil {
			Logger.Err(fmt.Sprintf("<CGRServer> JSON accept error: <%s>", err.Error()))
			now := time.Now()
			if now.Sub(lastErrorTime) > time.Duration(5*time.Second) {
				errCnt = 0 // reset error count if last error was more than 5 seconds ago
			}
			lastErrorTime = time.Now()
			errCnt += 1
			if errCnt > 50 { // Too many errors in short interval, network buffer failure most probably
				break
			}
			continue
		}
		//utils.Logger.Info(fmt.Sprintf("<CGRServer> New incoming connection: %v", conn.RemoteAddr()))
		go jsonrpc.ServeConn(conn)
	}

}

func (s *Server) ServeGOB(addr string) {
	if !s.rpcEnabled {
		return
	}
	lGOB, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("ServeGOB listen error:", e)
	}
	Logger.Info(fmt.Sprintf("Starting CGRateS GOB server at <%s>.", addr))
	errCnt := 0
	var lastErrorTime time.Time
	for {
		conn, err := lGOB.Accept()
		if err != nil {
			Logger.Err(fmt.Sprintf("<CGRServer> GOB accept error: <%s>", err.Error()))
			now := time.Now()
			if now.Sub(lastErrorTime) > time.Duration(5*time.Second) {
				errCnt = 0 // reset error count if last error was more than 5 seconds ago
			}
			lastErrorTime = time.Now()
			errCnt += 1
			if errCnt > 50 { // Too many errors in short interval, network buffer failure most probably
				break
			}
			continue
		}

		//utils.Logger.Info(fmt.Sprintf("<CGRServer> New incoming connection: %v", conn.RemoteAddr()))
		go rpc.ServeConn(conn)
	}
}

func (s *Server) ServeHTTP(addr string) {
	if s.rpcEnabled {
		http.HandleFunc("/jsonrpc", func(w http.ResponseWriter, req *http.Request) {
			defer req.Body.Close()
			w.Header().Set("Content-Type", "application/json")
			res := NewRPCRequest(req.Body).Call()
			io.Copy(w, res)
		})
		http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
			jsonrpc.ServeConn(ws)
		}))
		s.httpEnabled = true
	}
	if !s.httpEnabled {
		return
	}
	Logger.Info(fmt.Sprintf("Starting CGRateS HTTP server at %s.", addr))
	http.ListenAndServe(addr, nil)
}

func (s *Server) ServeBiJSON(addr string) {
	if s.birpcSrv == nil {
		return
	}
	lBiJSON, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("ServeBiJSON listen error:", e)
	}
	Logger.Info(fmt.Sprintf("Starting CGRateS BiJSON server at %s.", addr))
	s.birpcSrv.Accept(lBiJSON)
}

// rpcRequest represents a RPC request.
// rpcRequest implements the io.ReadWriteCloser interface.
type rpcRequest struct {
	r    io.Reader     // holds the JSON formated RPC request
	rw   io.ReadWriter // holds the JSON formated RPC response
	done chan bool     // signals then end of the RPC request
}

// NewRPCRequest returns a new rpcRequest.
func NewRPCRequest(r io.Reader) *rpcRequest {
	var buf bytes.Buffer
	done := make(chan bool)
	return &rpcRequest{r, &buf, done}
}

func (r *rpcRequest) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *rpcRequest) Write(p []byte) (n int, err error) {
	n, err = r.rw.Write(p)
	r.done <- true
	return
}

func (r *rpcRequest) Close() error {
	//r.done <- true // seem to be called sometimes before the write command finishes!
	return nil
}

// Call invokes the RPC request, waits for it to complete, and returns the results.
func (r *rpcRequest) Call() io.Reader {
	go jsonrpc.ServeConn(r)
	<-r.done
	return r.rw
}
