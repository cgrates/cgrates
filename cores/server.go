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
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"golang.org/x/net/websocket"
)

func NewServer(caps *engine.Caps) (s *Server) {
	s = &Server{
		httpMux:         http.NewServeMux(),
		httpsMux:        http.NewServeMux(),
		stopbiRPCServer: make(chan struct{}, 1),
		caps:            caps,

		rpcStarted: utils.NewSyncedChan(),
		rpcServer:  birpc.NewServer(),
		birpcSrv:   birpc.NewBirpcServer(),
	}
	s.httpServer = &http.Server{Handler: s.httpMux}
	s.httpsServer = &http.Server{Handler: s.httpsMux}
	return
}

type Server struct {
	sync.RWMutex
	httpEnabled     bool
	birpcSrv        *birpc.BirpcServer
	stopbiRPCServer chan struct{} // used in order to fully stop the biRPC
	httpsMux        *http.ServeMux
	httpMux         *http.ServeMux
	caps            *engine.Caps
	anz             *analyzers.AnalyzerService

	rpcStarted  *utils.SyncedChan
	rpcServer   *birpc.Server
	rpcJSONl    net.Listener
	rpcGOBl     net.Listener
	rpcJSONlTLS net.Listener
	rpcGOBlTLS  net.Listener
	httpServer  *http.Server
	httpsServer *http.Server
	startSrv    sync.Once
}

func (s *Server) SetAnalyzer(anz *analyzers.AnalyzerService) {
	s.anz = anz
}

func (s *Server) RpcRegister(rcvr interface{}) {
	birpc.Register(rcvr)
}

func (s *Server) RpcRegisterName(name string, rcvr interface{}) {
	birpc.RegisterName(name, rcvr)
}

func (s *Server) RpcUnregisterName(name string) {
	birpc.DefaultServer.UnregisterName(name)
}

func (s *Server) RegisterHTTPFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.httpMux.HandleFunc(pattern, handler)
	s.httpsMux.HandleFunc(pattern, handler)
	s.Lock()
	s.httpEnabled = true
	s.Unlock()
}

func (s *Server) RegisterHttpHandler(pattern string, handler http.Handler) {
	s.httpMux.Handle(pattern, handler)
	s.httpsMux.Handle(pattern, handler)
	s.Lock()
	s.httpEnabled = true
	s.Unlock()
}

// Registers a new BiJsonRpc name
func (s *Server) BiRPCRegisterName(name string, rcv interface{}) {
	s.birpcSrv.RegisterName(name, rcv)
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

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rmtIP, _ := utils.GetRemoteIP(r)
	rmtAddr, _ := net.ResolveIPAddr(utils.EmptyString, rmtIP)
	res := newRPCRequest(r.Body, rmtAddr, s.caps, s.anz).Call()
	io.Copy(w, res)
	r.Body.Close()
}

func (s *Server) handleWebSocket(ws *websocket.Conn) {
	birpc.ServeCodec(newCapsJSONCodec(ws, s.caps, s.anz))
}

func (s *Server) ServeJSON(ctx *context.Context, shtdwnEngine context.CancelFunc, addr string) (err error) {
	if s.rpcJSONl, err = net.Listen(utils.TCP, addr); err != nil {
		log.Printf("Serve%s listen error: %s", utils.JSONCaps, err)
		shtdwnEngine()
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS %s server at <%s>.", utils.JSONCaps, addr))
	return acceptRPC(ctx, shtdwnEngine, s.rpcServer, s.rpcJSONl, utils.JSONCaps, func(conn conn) birpc.ServerCodec {
		return newCapsJSONCodec(conn, s.caps, s.anz)
	})
}

func (s *Server) ServeGOB(ctx *context.Context, shtdwnEngine context.CancelFunc, addr string) (err error) {
	if s.rpcGOBl, err = net.Listen(utils.TCP, addr); err != nil {
		log.Printf("Serve%s listen error: %s", utils.GOBCaps, err)
		shtdwnEngine()
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS %s server at <%s>.", utils.GOBCaps, addr))
	return acceptRPC(ctx, shtdwnEngine, s.rpcServer, s.rpcGOBl, utils.GOBCaps, func(conn conn) birpc.ServerCodec {
		return newCapsGOBCodec(conn, s.caps, s.anz)
	})
}

func (s *Server) ServeHTTP(shtdwnEngine context.CancelFunc, addr, jsonRPCURL, wsRPCURL string,
	useBasicAuth bool, userList map[string]string) {
	s.Lock()
	s.httpEnabled = s.httpEnabled || jsonRPCURL != "" || wsRPCURL != ""
	enabled := s.httpEnabled
	s.Unlock()
	if !enabled {
		return
	}
	if jsonRPCURL != "" {
		utils.Logger.Info("<HTTP> enabling handler for JSON-RPC")
		if useBasicAuth {
			s.httpMux.HandleFunc(jsonRPCURL, use(s.handleRequest, basicAuth(userList)))
		} else {
			s.httpMux.HandleFunc(jsonRPCURL, s.handleRequest)
		}
	}
	if wsRPCURL != "" {
		utils.Logger.Info("<HTTP> enabling handler for WebSocket connections")
		wsHandler := websocket.Handler(s.handleWebSocket)
		if useBasicAuth {
			s.httpMux.HandleFunc(wsRPCURL, use(wsHandler.ServeHTTP, basicAuth(userList)))
		} else {
			s.httpMux.Handle(wsRPCURL, wsHandler)
		}
	}
	if useBasicAuth {
		utils.Logger.Info("<HTTP> enabling basic auth")
	}
	utils.Logger.Info(fmt.Sprintf("<HTTP> start listening at <%s>", addr))
	s.httpServer.Addr = addr
	if err := s.httpServer.ListenAndServe(); err != nil {
		log.Println(fmt.Sprintf("<HTTP>Error: %s when listening ", err))
		shtdwnEngine()
	}
}

// ServeBiRPC create a goroutine to listen and serve as BiRPC server
func (s *Server) ServeBiRPC2(addrJSON, addrGOB string, onConn, onDis func(birpc.ClientConnector)) (err error) {
	s.birpcSrv.OnConnect(onConn)
	s.birpcSrv.OnDisconnect(onDis)
	if addrJSON != utils.EmptyString {
		var ljson net.Listener
		if ljson, err = listenBiRPC(s.birpcSrv, addrJSON, utils.JSONCaps, func(conn conn) birpc.BirpcCodec {
			return newCapsBiRPCJSONCodec(conn, s.caps, s.anz)
		}, s.stopbiRPCServer); err != nil {
			return
		}
		defer ljson.Close()
	}
	if addrGOB != utils.EmptyString {
		var lgob net.Listener
		if lgob, err = listenBiRPC(s.birpcSrv, addrGOB, utils.GOBCaps, func(conn conn) birpc.BirpcCodec {
			return newCapsBiRPCGOBCodec(conn, s.caps, s.anz)
		}, s.stopbiRPCServer); err != nil {
			return
		}
		defer lgob.Close()
	}
	<-s.stopbiRPCServer // wait until server is stopped to close the listener
	return
}

func (s *Server) ServeGOBTLS(ctx *context.Context, shtdwnEngine context.CancelFunc,
	addr, serverCrt, serverKey, caCert string, serverPolicy int, serverName string) (err error) {
	config, err := loadTLSConfig(serverCrt, serverKey, caCert, serverPolicy, serverName)
	if err != nil {
		shtdwnEngine()
		return
	}
	s.rpcGOBlTLS, err = tls.Listen(utils.TCP, addr, config)
	if err != nil {
		log.Println(fmt.Sprintf("Error: %s when listening", err))
		shtdwnEngine()
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS %s TLS server at <%s>.", utils.GOBCaps, addr))

	return acceptRPC(ctx, shtdwnEngine, s.rpcServer, s.rpcGOBlTLS, utils.GOBCaps, func(conn conn) birpc.ServerCodec {
		return newCapsGOBCodec(conn, s.caps, s.anz)
	})
}

func (s *Server) ServeJSONTLS(ctx *context.Context, shtdwnEngine context.CancelFunc,
	addr, serverCrt, serverKey, caCert string, serverPolicy int, serverName string) (err error) {
	config, err := loadTLSConfig(serverCrt, serverKey, caCert, serverPolicy, serverName)
	if err != nil {
		shtdwnEngine()
		return
	}
	s.rpcJSONlTLS, err = tls.Listen(utils.TCP, addr, config)
	if err != nil {
		log.Println(fmt.Sprintf("Error: %s when listening", err))
		shtdwnEngine()
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS %s TLS server at <%s>.", utils.JSONCaps, addr))

	return acceptRPC(ctx, shtdwnEngine, s.rpcServer, s.rpcJSONlTLS, utils.JSONCaps, func(conn conn) birpc.ServerCodec {
		return newCapsGOBCodec(conn, s.caps, s.anz)
	})
}

func (s *Server) ServeHTTPS(shtdwnEngine context.CancelFunc,
	addr, serverCrt, serverKey, caCert string, serverPolicy int,
	serverName string, jsonRPCURL string, wsRPCURL string,
	useBasicAuth bool, userList map[string]string) {
	s.Lock()
	s.httpEnabled = s.httpEnabled || jsonRPCURL != "" || wsRPCURL != ""
	enabled := s.httpEnabled
	s.Unlock()
	if !enabled {
		return
	}
	if jsonRPCURL != "" {
		utils.Logger.Info("<HTTPS> enabling handler for JSON-RPC")
		if useBasicAuth {
			s.httpsMux.HandleFunc(jsonRPCURL, use(s.handleRequest, basicAuth(userList)))
		} else {
			s.httpsMux.HandleFunc(jsonRPCURL, s.handleRequest)
		}
	}
	if wsRPCURL != "" {
		utils.Logger.Info("<HTTPS> enabling handler for WebSocket connections")
		wsHandler := websocket.Handler(s.handleWebSocket)
		if useBasicAuth {
			s.httpsMux.HandleFunc(wsRPCURL, use(wsHandler.ServeHTTP, basicAuth(userList)))
		} else {
			s.httpsMux.Handle(wsRPCURL, wsHandler)
		}
	}
	if useBasicAuth {
		utils.Logger.Info("<HTTPS> enabling basic auth")
	}
	config, err := loadTLSConfig(serverCrt, serverKey, caCert, serverPolicy, serverName)
	if err != nil {
		shtdwnEngine()
		return
	}
	s.httpsServer.Addr = addr
	s.httpsServer.TLSConfig = config
	utils.Logger.Info(fmt.Sprintf("<HTTPS> start listening at <%s>", addr))

	if err := s.httpsServer.ListenAndServeTLS(serverCrt, serverKey); err != nil {
		log.Println(fmt.Sprintf("<HTTPS>Error: %s when listening ", err))
		shtdwnEngine()
	}
}

func (s *Server) Stop() {
	s.rpcJSONl.Close()
	s.rpcGOBl.Close()
	s.rpcJSONlTLS.Close()
	s.rpcGOBlTLS.Close()
	s.httpServer.Shutdown(context.Background())
	s.httpsServer.Shutdown(context.Background())
	s.StopBiRPC()
}

// StopBiRPC stops the go routine create with ServeBiJSON
func (s *Server) StopBiRPC() {
	s.stopbiRPCServer <- struct{}{}
	s.birpcSrv = birpc.NewBirpcServer()
}

func (s *Server) StartServer(ctx *context.Context, shtdwnEngine context.CancelFunc, cfg *config.CGRConfig) {
	s.startSrv.Do(func() {
		go s.ServeJSON(ctx, shtdwnEngine, cfg.ListenCfg().RPCJSONListen)
		go s.ServeGOB(ctx, shtdwnEngine, cfg.ListenCfg().RPCGOBListen)
		go s.ServeHTTP(
			shtdwnEngine,
			cfg.ListenCfg().HTTPListen,
			cfg.HTTPCfg().JsonRPCURL,
			cfg.HTTPCfg().WSURL,
			cfg.HTTPCfg().UseBasicAuth,
			cfg.HTTPCfg().AuthUsers,
		)
		if (len(cfg.ListenCfg().RPCGOBTLSListen) != 0 ||
			len(cfg.ListenCfg().RPCJSONTLSListen) != 0 ||
			len(cfg.ListenCfg().HTTPTLSListen) != 0) &&
			(len(cfg.TLSCfg().ServerCerificate) == 0 ||
				len(cfg.TLSCfg().ServerKey) == 0) {
			utils.Logger.Warning("WARNING: missing TLS certificate/key file!")
			return
		}
		if cfg.ListenCfg().RPCGOBTLSListen != utils.EmptyString {
			go s.ServeGOBTLS(
				ctx, shtdwnEngine,
				cfg.ListenCfg().RPCGOBTLSListen,
				cfg.TLSCfg().ServerCerificate,
				cfg.TLSCfg().ServerKey,
				cfg.TLSCfg().CaCertificate,
				cfg.TLSCfg().ServerPolicy,
				cfg.TLSCfg().ServerName,
			)
		}
		if cfg.ListenCfg().RPCJSONTLSListen != utils.EmptyString {
			go s.ServeJSONTLS(
				ctx, shtdwnEngine,
				cfg.ListenCfg().RPCJSONTLSListen,
				cfg.TLSCfg().ServerCerificate,
				cfg.TLSCfg().ServerKey,
				cfg.TLSCfg().CaCertificate,
				cfg.TLSCfg().ServerPolicy,
				cfg.TLSCfg().ServerName,
			)
		}
		if cfg.ListenCfg().HTTPTLSListen != utils.EmptyString {
			go s.ServeHTTPS(
				shtdwnEngine,
				cfg.ListenCfg().HTTPTLSListen,
				cfg.TLSCfg().ServerCerificate,
				cfg.TLSCfg().ServerKey,
				cfg.TLSCfg().CaCertificate,
				cfg.TLSCfg().ServerPolicy,
				cfg.TLSCfg().ServerName,
				cfg.HTTPCfg().JsonRPCURL,
				cfg.HTTPCfg().WSURL,
				cfg.HTTPCfg().UseBasicAuth,
				cfg.HTTPCfg().AuthUsers,
			)
		}
	})
}
