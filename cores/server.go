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

package cores

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cenkalti/rpc2"
	"golang.org/x/net/websocket"
)

func NewServer(caps *engine.Caps) (s *Server) {
	return &Server{
		httpMux:         http.NewServeMux(),
		httpsMux:        http.NewServeMux(),
		stopbiRPCServer: make(chan struct{}, 1),
		caps:            caps,
	}
}

type Server struct {
	sync.RWMutex
	rpcEnabled      bool
	httpEnabled     bool
	birpcSrv        *rpc2.Server
	stopbiRPCServer chan struct{} // used in order to fully stop the biRPC
	httpsMux        *http.ServeMux
	httpMux         *http.ServeMux
	caps            *engine.Caps
	anz             *analyzers.AnalyzerService
}

func (s *Server) SetAnalyzer(anz *analyzers.AnalyzerService) {
	s.anz = anz
}

func (s *Server) RpcRegister(rcvr interface{}) {
	utils.RegisterRpcParams(utils.EmptyString, rcvr)
	rpc.Register(rcvr)
	s.Lock()
	s.rpcEnabled = true
	s.Unlock()
}

func (s *Server) RpcRegisterName(name string, rcvr interface{}) {
	utils.RegisterRpcParams(name, rcvr)
	rpc.RegisterName(name, rcvr)
	s.Lock()
	s.rpcEnabled = true
	s.Unlock()
}

func (s *Server) RegisterHttpFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if s.httpMux != nil {
		s.httpMux.HandleFunc(pattern, handler)
	}
	if s.httpsMux != nil {
		s.httpsMux.HandleFunc(pattern, handler)
	}
	s.Lock()
	s.httpEnabled = true
	s.Unlock()
}

func (s *Server) RegisterHttpHandler(pattern string, handler http.Handler) {
	if s.httpMux != nil {
		s.httpMux.Handle(pattern, handler)
	}
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

func (s *Server) serveCodec(addr, codecName string, newCodec func(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) rpc.ServerCodec,
	shdChan *utils.SyncedChan) {
	s.RLock()
	enabled := s.rpcEnabled
	s.RUnlock()
	if !enabled {
		return
	}

	l, e := net.Listen(utils.TCP, addr)
	if e != nil {
		log.Printf("Serve%s listen error: %s", codecName, e)
		shdChan.CloseOnce()
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS %s server at <%s>.", codecName, addr))
	s.accept(l, codecName, newCodec, shdChan)
}

func (s *Server) accept(l net.Listener, codecName string, newCodec func(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) rpc.ServerCodec,
	shdChan *utils.SyncedChan) {
	errCnt := 0
	var lastErrorTime time.Time
	for {
		conn, err := l.Accept()
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<CGRServer> %s accept error: <%s>", codecName, err.Error()))
			now := time.Now()
			if now.Sub(lastErrorTime) > 5*time.Second {
				errCnt = 0 // reset error count if last error was more than 5 seconds ago
			}
			lastErrorTime = time.Now()
			errCnt++
			if errCnt > 50 { // Too many errors in short interval, network buffer failure most probably
				shdChan.CloseOnce()
				return
			}
			continue
		}
		go rpc.ServeCodec(newCodec(conn, s.caps, s.anz))
	}
}

func (s *Server) ServeJSON(addr string, shdChan *utils.SyncedChan) {
	s.serveCodec(addr, utils.JSONCaps, newCapsJSONCodec, shdChan)
}

func (s *Server) ServeGOB(addr string, shdChan *utils.SyncedChan) {
	s.serveCodec(addr, utils.GOBCaps, newCapsGOBCodec, shdChan)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	rmtIP, _ := utils.GetRemoteIP(r)
	rmtAddr, _ := net.ResolveIPAddr(utils.EmptyString, rmtIP)
	res := newRPCRequest(r.Body, rmtAddr, s.caps, s.anz).Call()
	io.Copy(w, res)
}

func registerProfiler(addr string, mux *http.ServeMux) {
	mux.HandleFunc(addr, pprof.Index)
	mux.HandleFunc(addr+"cmdline", pprof.Cmdline)
	mux.HandleFunc(addr+"profile", pprof.Profile)
	mux.HandleFunc(addr+"symbol", pprof.Symbol)
	mux.HandleFunc(addr+"trace", pprof.Trace)
	mux.Handle(addr+"goroutine", pprof.Handler("goroutine"))
	mux.Handle(addr+"heap", pprof.Handler("heap"))
	mux.Handle(addr+"threadcreate", pprof.Handler("threadcreate"))
	mux.Handle(addr+"block", pprof.Handler("block"))
	mux.Handle(addr+"allocs", pprof.Handler("allocs"))
	mux.Handle(addr+"mutex", pprof.Handler("mutex"))
}

func (s *Server) RegisterProfiler(addr string) {
	if addr[len(addr)-1] != '/' {
		addr = addr + "/"
	}
	registerProfiler(addr, s.httpMux)
	registerProfiler(addr, s.httpsMux)
}

func (s *Server) ServeHTTP(addr string, jsonRPCURL string, wsRPCURL string,
	useBasicAuth bool, userList map[string]string, shdChan *utils.SyncedChan) {
	s.RLock()
	enabled := s.rpcEnabled
	s.RUnlock()
	if !enabled {
		return
	}
	if jsonRPCURL != "" {
		s.Lock()
		s.httpEnabled = true
		s.Unlock()

		utils.Logger.Info("<HTTP> enabling handler for JSON-RPC")
		if useBasicAuth {
			s.httpMux.HandleFunc(jsonRPCURL, use(s.handleRequest, basicAuth(userList)))
		} else {
			s.httpMux.HandleFunc(jsonRPCURL, s.handleRequest)
		}
	}
	if wsRPCURL != "" {
		s.Lock()
		s.httpEnabled = true
		s.Unlock()
		utils.Logger.Info("<HTTP> enabling handler for WebSocket connections")
		wsHandler := websocket.Handler(s.handleWebSocket)
		if useBasicAuth {
			s.httpMux.HandleFunc(wsRPCURL, use(wsHandler.ServeHTTP, basicAuth(userList)))
		} else {
			s.httpMux.Handle(wsRPCURL, wsHandler)
		}
	}
	if !s.httpEnabled {
		return
	}
	if useBasicAuth {
		utils.Logger.Info("<HTTP> enabling basic auth")
	}
	utils.Logger.Info(fmt.Sprintf("<HTTP> start listening at <%s>", addr))
	if err := http.ListenAndServe(addr, s.httpMux); err != nil {
		log.Println(fmt.Sprintf("<HTTP>Error: %s when listening ", err))
		shdChan.CloseOnce()
	}
}

// ServeBiRPC create a goroutine to listen and serve as BiRPC server
func (s *Server) ServeBiRPC(addrJSON, addrGOB string, onConn func(*rpc2.Client), onDis func(*rpc2.Client)) (err error) {
	s.RLock()
	isNil := s.birpcSrv == nil
	s.RUnlock()
	if isNil {
		return fmt.Errorf("BiRPCServer should not be nil")
	}

	s.birpcSrv.OnConnect(onConn)
	s.birpcSrv.OnDisconnect(onDis)
	if addrJSON != utils.EmptyString {
		var ljson net.Listener
		if ljson, err = s.listenBiRPC(s.birpcSrv, addrJSON, utils.JSONCaps, newCapsBiRPCJSONCodec); err != nil {
			return
		}
		defer ljson.Close()
	}
	if addrGOB != utils.EmptyString {
		var lgob net.Listener
		if lgob, err = s.listenBiRPC(s.birpcSrv, addrGOB, utils.GOBCaps, newCapsBiRPCGOBCodec); err != nil {
			return
		}
		defer lgob.Close()
	}
	<-s.stopbiRPCServer // wait until server is stopped to close the listener
	return
}

func (s *Server) listenBiRPC(srv *rpc2.Server, addr, codecName string, newCodec func(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) rpc2.Codec) (lBiRPC net.Listener, err error) {
	if lBiRPC, err = net.Listen(utils.TCP, addr); err != nil {
		log.Printf("ServeBi%s listen error: %s \n", codecName, err)
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS Bi%s server at <%s>", codecName, addr))
	go s.acceptBiRPC(srv, lBiRPC, codecName, newCodec)
	return
}

func (s *Server) acceptBiRPC(srv *rpc2.Server, l net.Listener, codecName string, newCodec func(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) rpc2.Codec) {
	for {
		conn, err := l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") { // if closed by us do not log
				return
			}
			s.stopbiRPCServer <- struct{}{}
			utils.Logger.Crit(fmt.Sprintf("Stopped Bi%s server beacause %s", codecName, err))
			return // stop if we get Accept error
		}
		go srv.ServeCodec(newCodec(conn, s.caps, s.anz))
	}
}

// StopBiRPC stops the go routine create with ServeBiJSON
func (s *Server) StopBiRPC() {
	s.stopbiRPCServer <- struct{}{}
	s.Lock()
	s.birpcSrv = nil
	s.Unlock()
}

// rpcRequest represents a RPC request.
// rpcRequest implements the io.ReadWriteCloser interface.
type rpcRequest struct {
	r          io.ReadCloser // holds the JSON formated RPC request
	rw         io.ReadWriter // holds the JSON formated RPC response
	remoteAddr net.Addr
	caps       *engine.Caps
	anzWarpper *analyzers.AnalyzerService
}

// newRPCRequest returns a new rpcRequest.
func newRPCRequest(r io.ReadCloser, remoteAddr net.Addr, caps *engine.Caps, anz *analyzers.AnalyzerService) *rpcRequest {
	return &rpcRequest{
		r:          r,
		rw:         new(bytes.Buffer),
		remoteAddr: remoteAddr,
		caps:       caps,
		anzWarpper: anz,
	}
}

func (r *rpcRequest) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *rpcRequest) Write(p []byte) (n int, err error) {
	return r.rw.Write(p)
}

func (r *rpcRequest) LocalAddr() net.Addr {
	return utils.LocalAddr()
}
func (r *rpcRequest) RemoteAddr() net.Addr {
	return r.remoteAddr
}

func (r *rpcRequest) Close() error {
	return r.r.Close()
}

// Call invokes the RPC request, waits for it to complete, and returns the results.
func (r *rpcRequest) Call() io.Reader {
	rpc.ServeCodec(newCapsJSONCodec(r, r.caps, r.anzWarpper))
	return r.rw
}

func loadTLSConfig(serverCrt, serverKey, caCert string, serverPolicy int,
	serverName string) (config *tls.Config, err error) {
	cert, err := tls.LoadX509KeyPair(serverCrt, serverKey)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("Error: %s when load server keys", err))
		return nil, err
	}

	rootCAs, err := x509.SystemCertPool()
	//This will only happen on windows
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("Error: %s when load SystemCertPool", err))
		return nil, err
	}

	if caCert != "" {
		ca, err := os.ReadFile(caCert)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("Error: %s when read CA", err))
			return config, err
		}

		if ok := rootCAs.AppendCertsFromPEM(ca); !ok {
			utils.Logger.Crit(fmt.Sprintf("Cannot append certificate authority"))
			return config, errors.New("Cannot append certificate authority")
		}
	}

	config = &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.ClientAuthType(serverPolicy),
		ClientCAs:    rootCAs,
	}
	if serverName != "" {
		config.ServerName = serverName
	}
	return
}

func (s *Server) serveCodecTLS(addr, codecName, serverCrt, serverKey, caCert string,
	serverPolicy int, serverName string, newCodec func(conn conn, caps *engine.Caps, anz *analyzers.AnalyzerService) rpc.ServerCodec,
	shdChan *utils.SyncedChan) {
	s.RLock()
	enabled := s.rpcEnabled
	s.RUnlock()
	if !enabled {
		return
	}
	config, err := loadTLSConfig(serverCrt, serverKey, caCert, serverPolicy, serverName)
	if err != nil {
		shdChan.CloseOnce()
		return
	}
	listener, err := tls.Listen(utils.TCP, addr, config)
	if err != nil {
		log.Println(fmt.Sprintf("Error: %s when listening", err))
		shdChan.CloseOnce()
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS %s TLS server at <%s>.", codecName, addr))
	s.accept(listener, codecName+" "+utils.TLS, newCodec, shdChan)
}

func (s *Server) ServeGOBTLS(addr, serverCrt, serverKey, caCert string,
	serverPolicy int, serverName string, shdChan *utils.SyncedChan) {
	s.serveCodecTLS(addr, utils.GOBCaps, serverCrt, serverKey, caCert, serverPolicy, serverName, newCapsGOBCodec, shdChan)
}

func (s *Server) ServeJSONTLS(addr, serverCrt, serverKey, caCert string,
	serverPolicy int, serverName string, shdChan *utils.SyncedChan) {
	s.serveCodecTLS(addr, utils.JSONCaps, serverCrt, serverKey, caCert, serverPolicy, serverName, newCapsJSONCodec, shdChan)
}

func (s *Server) handleWebSocket(ws *websocket.Conn) {
	rpc.ServeCodec(newCapsJSONCodec(ws, s.caps, s.anz))
}

func (s *Server) ServeHTTPTLS(addr, serverCrt, serverKey, caCert string, serverPolicy int,
	serverName string, jsonRPCURL string, wsRPCURL string,
	useBasicAuth bool, userList map[string]string, shdChan *utils.SyncedChan) {
	s.RLock()
	enabled := s.rpcEnabled
	s.RUnlock()
	if !enabled {
		return
	}
	if jsonRPCURL != "" {
		s.Lock()
		s.httpEnabled = true
		s.Unlock()
		utils.Logger.Info("<HTTPS> enabling handler for JSON-RPC")
		if useBasicAuth {
			s.httpsMux.HandleFunc(jsonRPCURL, use(s.handleRequest, basicAuth(userList)))
		} else {
			s.httpsMux.HandleFunc(jsonRPCURL, s.handleRequest)
		}
	}
	if wsRPCURL != "" {
		s.Lock()
		s.httpEnabled = true
		s.Unlock()
		utils.Logger.Info("<HTTPS> enabling handler for WebSocket connections")
		wsHandler := websocket.Handler(s.handleWebSocket)
		if useBasicAuth {
			s.httpsMux.HandleFunc(wsRPCURL, use(wsHandler.ServeHTTP, basicAuth(userList)))
		} else {
			s.httpsMux.Handle(wsRPCURL, wsHandler)
		}
	}
	if !s.httpEnabled {
		return
	}
	if useBasicAuth {
		utils.Logger.Info("<HTTPS> enabling basic auth")
	}
	config, err := loadTLSConfig(serverCrt, serverKey, caCert, serverPolicy, serverName)
	if err != nil {
		shdChan.CloseOnce()
		return
	}
	httpSrv := http.Server{
		Addr:      addr,
		Handler:   s.httpsMux,
		TLSConfig: config,
	}
	utils.Logger.Info(fmt.Sprintf("<HTTPS> start listening at <%s>", addr))
	if err := httpSrv.ListenAndServeTLS(serverCrt, serverKey); err != nil {
		log.Println(fmt.Sprintf("<HTTPS>Error: %s when listening ", err))
		shdChan.CloseOnce()
	}
}
