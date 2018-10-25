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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
	"sync"
	"time"

	"github.com/cenkalti/rpc2"
	rpc2_jsonrpc "github.com/cenkalti/rpc2/jsonrpc"
	"golang.org/x/net/websocket"
	_ "net/http/pprof"
)

type Server struct {
	rpcEnabled  bool
	httpEnabled bool
	birpcSrv    *rpc2.Server
	sync.RWMutex
	httpsMux *http.ServeMux
}

func (s *Server) RpcRegister(rcvr interface{}) {
	rpc.Register(rcvr)
	s.Lock()
	s.rpcEnabled = true
	s.Unlock()
}

func (s *Server) RpcRegisterName(name string, rcvr interface{}) {
	rpc.RegisterName(name, rcvr)
	s.Lock()
	s.rpcEnabled = true
	s.Unlock()
}

func (s *Server) RegisterHttpFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
	if s.httpsMux != nil {
		s.httpsMux.HandleFunc(pattern, handler)
	}
	s.Lock()
	s.httpEnabled = true
	s.Unlock()
}

func (s *Server) RegisterHttpHandler(pattern string, handler http.Handler) {
	http.Handle(pattern, handler)
	if s.httpsMux != nil {
		s.httpsMux.Handle(pattern, handler)
	}
	s.Lock()
	s.httpEnabled = true
	s.Unlock()
}

// Registers a new BiJsonRpc name
func (s *Server) BiRPCRegisterName(method string, handlerFunc interface{}) {
	s.RLock()
	isNil := s.birpcSrv == nil
	s.RUnlock()
	if isNil {
		s.Lock()
		s.birpcSrv = rpc2.NewServer()
		s.Unlock()
	}
	s.birpcSrv.Handle(method, handlerFunc)
}

func (s *Server) BiRPCRegister(rcvr interface{}) {
	s.RLock()
	isNil := s.birpcSrv == nil
	s.RUnlock()
	if isNil {
		s.Lock()
		s.birpcSrv = rpc2.NewServer()
		s.Unlock()
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
	s.RLock()
	enabled := s.rpcEnabled
	s.RUnlock()
	if !enabled {
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
	s.RLock()
	enabled := s.rpcEnabled
	s.RUnlock()
	if !enabled {
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

func handleRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	res := NewRPCRequest(r.Body).Call()
	io.Copy(w, res)
}

func (s *Server) ServeHTTP(addr string, jsonRPCURL string, wsRPCURL string,
	useBasicAuth bool, userList map[string]string, exitChan chan bool) {
	s.RLock()
	enabled := s.rpcEnabled
	s.RUnlock()
	if !enabled {
		return
	}
	if enabled && jsonRPCURL != "" {
		s.Lock()
		s.httpEnabled = true
		s.Unlock()

		Logger.Info("<HTTP> enabling handler for JSON-RPC")
		if useBasicAuth {
			http.HandleFunc(jsonRPCURL, use(handleRequest, basicAuth(userList)))
		} else {
			http.HandleFunc(jsonRPCURL, handleRequest)
		}
	}
	if enabled && wsRPCURL != "" {
		s.Lock()
		s.httpEnabled = true
		s.Unlock()
		Logger.Info("<HTTP> enabling handler for WebSocket connections")
		wsHandler := websocket.Handler(func(ws *websocket.Conn) {
			jsonrpc.ServeConn(ws)
		})
		if useBasicAuth {
			http.HandleFunc(wsRPCURL, use(func(w http.ResponseWriter, r *http.Request) {
				wsHandler.ServeHTTP(w, r)
			}, basicAuth(userList)))
		} else {
			http.Handle(wsRPCURL, wsHandler)
		}
	}
	if !s.httpEnabled {
		return
	}
	if useBasicAuth {
		Logger.Info("<HTTP> enabling basic auth")
	}
	Logger.Info(fmt.Sprintf("<HTTP> start listening at <%s>", addr))
	http.ListenAndServe(addr, nil)
	exitChan <- true
}

func (s *Server) ServeBiJSON(addr string, onConn func(*rpc2.Client), onDis func(*rpc2.Client)) {
	s.RLock()
	isNil := s.birpcSrv == nil
	s.RUnlock()
	if isNil {
		return
	}
	lBiJSON, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("ServeBiJSON listen error:", e)
	}
	s.birpcSrv.OnConnect(onConn)
	s.birpcSrv.OnDisconnect(onDis)
	Logger.Info(fmt.Sprintf("Starting CGRateS BiJSON server at <%s>", addr))
	for {
		conn, err := lBiJSON.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.birpcSrv.ServeCodec(rpc2_jsonrpc.NewJSONCodec(conn))
	}
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

func loadTLSConfig(serverCrt, serverKey, caCert string, serverPolicy int,
	serverName string) (config tls.Config, err error) {
	cert, err := tls.LoadX509KeyPair(serverCrt, serverKey)
	if err != nil {
		log.Fatalf("Error: %s when load server keys", err)
	}
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Fatalf("Error: %s when load SystemCertPool", err)
		return
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if caCert != "" {
		ca, err := ioutil.ReadFile(caCert)
		if err != nil {
			log.Fatalf("Error: %s when read CA", err)
			return config, err
		}

		if ok := rootCAs.AppendCertsFromPEM(ca); !ok {
			log.Fatalf("Cannot append certificate authority")
			return config, err
		}
	}

	config = tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.ClientAuthType(serverPolicy),
		ClientCAs:    rootCAs,
	}
	if serverName != "" {
		config.ServerName = serverName
	}
	return
}

func (s *Server) ServeGOBTLS(addr, serverCrt, serverKey, caCert string,
	serverPolicy int, serverName string) {
	s.RLock()
	enabled := s.rpcEnabled
	s.RUnlock()
	if !enabled {
		return
	}
	config, err := loadTLSConfig(serverCrt, serverKey, caCert, serverPolicy, serverName)
	if err != nil {
		return
	}
	listener, err := tls.Listen("tcp", addr, &config)
	if err != nil {
		log.Fatalf("Error: %s when listening", err)
	}

	Logger.Info(fmt.Sprintf("Starting CGRateS GOB TLS server at <%s>.", addr))
	errCnt := 0
	var lastErrorTime time.Time
	for {
		conn, err := listener.Accept()
		defer conn.Close()
		if err != nil {
			Logger.Err(fmt.Sprintf("<CGRServer> TLS accept error: <%s>", err.Error()))
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

func (s *Server) ServeJSONTLS(addr, serverCrt, serverKey, caCert string,
	serverPolicy int, serverName string) {
	s.RLock()
	enabled := s.rpcEnabled
	s.RUnlock()
	if !enabled {
		return
	}
	config, err := loadTLSConfig(serverCrt, serverKey, caCert, serverPolicy, serverName)
	if err != nil {
		return
	}
	listener, err := tls.Listen("tcp", addr, &config)
	if err != nil {
		log.Fatalf("Error: %s when listening", err)
	}
	Logger.Info(fmt.Sprintf("Starting CGRateS JSON TLS server at <%s>.", addr))
	errCnt := 0
	var lastErrorTime time.Time
	for {
		conn, err := listener.Accept()
		defer conn.Close()
		if err != nil {
			Logger.Err(fmt.Sprintf("<CGRServer> TLS accept error: <%s>", err.Error()))
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
		go jsonrpc.ServeConn(conn)
	}
}

func (s *Server) ServeHTTPTLS(addr, serverCrt, serverKey, caCert string, serverPolicy int,
	serverName string, jsonRPCURL string, wsRPCURL string,
	useBasicAuth bool, userList map[string]string) {
	s.RLock()
	enabled := s.rpcEnabled
	s.RUnlock()
	if !enabled {
		return
	}
	s.httpsMux = http.NewServeMux()
	if enabled && jsonRPCURL != "" {
		s.Lock()
		s.httpEnabled = true
		s.Unlock()
		Logger.Info("<HTTPTLS> enabling handler for JSON-RPC")
		if useBasicAuth {
			s.httpsMux.HandleFunc(jsonRPCURL, use(handleRequest, basicAuth(userList)))
		} else {
			s.httpsMux.HandleFunc(jsonRPCURL, handleRequest)
		}
	}
	if enabled && wsRPCURL != "" {
		s.Lock()
		s.httpEnabled = true
		s.Unlock()
		Logger.Info("<HTTPTLS> enabling handler for WebSocket connections")
		wsHandler := websocket.Handler(func(ws *websocket.Conn) {
			jsonrpc.ServeConn(ws)
		})
		if useBasicAuth {
			s.httpsMux.HandleFunc(wsRPCURL, use(func(w http.ResponseWriter, r *http.Request) {
				wsHandler.ServeHTTP(w, r)
			}, basicAuth(userList)))
		} else {
			s.httpsMux.Handle(wsRPCURL, wsHandler)
		}
	}
	if !s.httpEnabled {
		return
	}
	if useBasicAuth {
		Logger.Info("<HTTPTLS> enabling basic auth")
	}
	config, err := loadTLSConfig(serverCrt, serverKey, caCert, serverPolicy, serverName)
	if err != nil {
		return
	}
	httpSrv := http.Server{
		Addr:      addr,
		Handler:   s.httpsMux,
		TLSConfig: &config,
	}
	Logger.Info(fmt.Sprintf("<HTTPTLS> start listening at <%s>", addr))
	httpSrv.ListenAndServeTLS(serverCrt, serverKey)
}
