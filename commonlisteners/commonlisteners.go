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

package commonlisteners

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"golang.org/x/net/websocket"
)

func NewCommonListenerS(caps *engine.Caps) *CommonListenerS {
	c := &CommonListenerS{
		httpMux:         http.NewServeMux(),
		httpsMux:        http.NewServeMux(),
		stopbiRPCServer: make(chan struct{}, 1),
		caps:            caps,

		rpcServer: birpc.NewServer(),
		birpcSrv:  birpc.NewBirpcServer(),
	}
	c.httpServer = &http.Server{
		Handler: c.httpMux,
	}
	c.httpsServer = &http.Server{
		Handler: c.httpsMux,
	}
	return c
}

type CommonListenerS struct {
	mu          sync.Mutex // mutex for httpEnabled field
	httpEnabled bool

	birpcSrv        *birpc.BirpcServer
	stopbiRPCServer chan struct{} // used in order to fully stop the biRPC
	httpsMux        *http.ServeMux
	httpMux         *http.ServeMux
	caps            *engine.Caps
	anz             *analyzers.AnalyzerS

	rpcServer   *birpc.Server
	rpcJSONl    net.Listener
	rpcGOBl     net.Listener
	rpcJSONlTLS net.Listener
	rpcGOBlTLS  net.Listener
	httpServer  *http.Server
	httpsServer *http.Server
	startSrv    sync.Once
}

func (c *CommonListenerS) SetAnalyzer(anz *analyzers.AnalyzerS) {
	c.anz = anz
}

func (c *CommonListenerS) RpcRegister(rcvr any) {
	c.rpcServer.Register(rcvr)
}

func (c *CommonListenerS) RpcRegisterName(name string, rcvr any) {
	c.rpcServer.RegisterName(name, rcvr)
}

func (c *CommonListenerS) RpcUnregisterName(name string) {
	c.rpcServer.UnregisterName(name)
}

func (c *CommonListenerS) RegisterHTTPFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	c.httpMux.HandleFunc(pattern, handler)
	c.httpsMux.HandleFunc(pattern, handler)
	c.mu.Lock()
	c.httpEnabled = true
	c.mu.Unlock()
}

func (c *CommonListenerS) RegisterHttpHandler(pattern string, handler http.Handler) {
	c.httpMux.Handle(pattern, handler)
	c.httpsMux.Handle(pattern, handler)
	c.mu.Lock()
	c.httpEnabled = true
	c.mu.Unlock()
}

// Registers a new BiJsonRpc name
func (c *CommonListenerS) BiRPCRegisterName(name string, rcv any) {
	c.birpcSrv.RegisterName(name, rcv)
}

func (c *CommonListenerS) BiRPCUnregisterName(name string) {
	c.birpcSrv.UnregisterName(name)
}

func (c *CommonListenerS) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rmtIP, _ := utils.GetRemoteIP(r)
	rmtAddr, _ := net.ResolveIPAddr(utils.EmptyString, rmtIP)
	res := newRPCRequest(c.rpcServer, r.Body, rmtAddr, c.caps, c.anz).Call()
	io.Copy(w, res)
	r.Body.Close()
}

func (c *CommonListenerS) handleWebSocket(ws *websocket.Conn) {
	c.rpcServer.ServeCodec(newCapsJSONCodec(ws, c.caps, c.anz))
}

func (c *CommonListenerS) ServeJSON(addr string, shutdown *utils.SyncedChan) (err error) {
	if c.rpcJSONl, err = net.Listen(utils.TCP, addr); err != nil {
		log.Printf("Serve%s listen error: %s", utils.JSONCaps, err)
		shutdown.CloseOnce()
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS %s server at <%s>.", utils.JSONCaps, addr))
	return acceptRPC(shutdown, c.rpcServer, c.rpcJSONl, utils.JSONCaps, func(conn conn) birpc.ServerCodec {
		return newCapsJSONCodec(conn, c.caps, c.anz)
	})
}

func (c *CommonListenerS) ServeGOB(addr string, shutdown *utils.SyncedChan) (err error) {
	if c.rpcGOBl, err = net.Listen(utils.TCP, addr); err != nil {
		log.Printf("Serve%s listen error: %s", utils.GOBCaps, err)
		shutdown.CloseOnce()
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS %s server at <%s>.", utils.GOBCaps, addr))
	return acceptRPC(shutdown, c.rpcServer, c.rpcGOBl, utils.GOBCaps, func(conn conn) birpc.ServerCodec {
		return newCapsGOBCodec(conn, c.caps, c.anz)
	})
}

func (c *CommonListenerS) ServeHTTP(addr, jsonRPCURL, wsRPCURL, pprofPath string,
	useBasicAuth bool, userList map[string]string, shutdown *utils.SyncedChan) {
	c.mu.Lock()
	c.httpEnabled = c.httpEnabled || jsonRPCURL != "" || wsRPCURL != "" || pprofPath != ""
	enabled := c.httpEnabled
	c.mu.Unlock()
	if !enabled {
		return
	}
	if jsonRPCURL != "" {
		utils.Logger.Info("<HTTP> enabling handler for JSON-RPC")
		if useBasicAuth {
			c.httpMux.HandleFunc(jsonRPCURL, use(c.handleRequest, basicAuth(userList)))
		} else {
			c.httpMux.HandleFunc(jsonRPCURL, c.handleRequest)
		}
	}
	if wsRPCURL != "" {
		utils.Logger.Info("<HTTP> enabling handler for WebSocket connections")
		wsHandler := websocket.Handler(c.handleWebSocket)
		if useBasicAuth {
			c.httpMux.HandleFunc(wsRPCURL, use(wsHandler.ServeHTTP, basicAuth(userList)))
		} else {
			c.httpMux.Handle(wsRPCURL, wsHandler)
		}
	}
	if pprofPath != "" {
		if !strings.HasSuffix(pprofPath, "/") {
			pprofPath += "/"
		}
		utils.Logger.Info(fmt.Sprintf("<HTTP> profiling endpoints registered at %q", pprofPath))
		if useBasicAuth {
			c.httpMux.HandleFunc(pprofPath, use(pprof.Index, basicAuth(userList)))
			c.httpMux.HandleFunc(pprofPath+"cmdline", use(pprof.Cmdline, basicAuth(userList)))
			c.httpMux.HandleFunc(pprofPath+"profile", use(pprof.Profile, basicAuth(userList)))
			c.httpMux.HandleFunc(pprofPath+"symbol", use(pprof.Symbol, basicAuth(userList)))
			c.httpMux.HandleFunc(pprofPath+"trace", use(pprof.Trace, basicAuth(userList)))
		} else {
			c.httpMux.HandleFunc(pprofPath, pprof.Index)
			c.httpMux.HandleFunc(pprofPath+"cmdline", pprof.Cmdline)
			c.httpMux.HandleFunc(pprofPath+"profile", pprof.Profile)
			c.httpMux.HandleFunc(pprofPath+"symbol", pprof.Symbol)
			c.httpMux.HandleFunc(pprofPath+"trace", pprof.Trace)
		}
	}
	if useBasicAuth {
		utils.Logger.Info("<HTTP> enabling basic auth")
	}
	utils.Logger.Info(fmt.Sprintf("<HTTP> start listening at <%s>", addr))
	c.httpServer.Addr = addr
	if err := c.httpServer.ListenAndServe(); err != nil {
		log.Println(fmt.Sprintf("<HTTP>Error: %s when listening ", err))
		shutdown.CloseOnce()
	}
}

// ServeBiRPC create a goroutine to listen and serve as BiRPC server
func (c *CommonListenerS) ServeBiRPC(addrJSON, addrGOB string, onConn, onDis func(birpc.ClientConnector)) (err error) {
	c.birpcSrv.OnConnect(onConn)
	c.birpcSrv.OnDisconnect(onDis)
	if addrJSON != utils.EmptyString {
		var ljson net.Listener
		if ljson, err = listenBiRPC(c.birpcSrv, addrJSON, utils.JSONCaps, func(conn conn) birpc.BirpcCodec {
			return newCapsBiRPCJSONCodec(conn, c.caps, c.anz)
		}, c.stopbiRPCServer); err != nil {
			return
		}
		defer ljson.Close()
	}
	if addrGOB != utils.EmptyString {
		var lgob net.Listener
		if lgob, err = listenBiRPC(c.birpcSrv, addrGOB, utils.GOBCaps, func(conn conn) birpc.BirpcCodec {
			return newCapsBiRPCGOBCodec(conn, c.caps, c.anz)
		}, c.stopbiRPCServer); err != nil {
			return
		}
		defer lgob.Close()
	}
	<-c.stopbiRPCServer // wait until server is stopped to close the listener
	return
}

func (c *CommonListenerS) ServeGOBTLS(addr, serverCrt, serverKey, caCert string, serverPolicy int,
	serverName string, shutdown *utils.SyncedChan) (err error) {
	config, err := loadTLSConfig(serverCrt, serverKey, caCert, serverPolicy, serverName)
	if err != nil {
		shutdown.CloseOnce()
		return
	}
	c.rpcGOBlTLS, err = tls.Listen(utils.TCP, addr, config)
	if err != nil {
		log.Println(fmt.Sprintf("Error: %s when listening", err))
		shutdown.CloseOnce()
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS %s TLS server at <%s>.", utils.GOBCaps, addr))

	return acceptRPC(shutdown, c.rpcServer, c.rpcGOBlTLS, utils.GOBCaps, func(conn conn) birpc.ServerCodec {
		return newCapsGOBCodec(conn, c.caps, c.anz)
	})
}

func (c *CommonListenerS) ServeJSONTLS(addr, serverCrt, serverKey, caCert string, serverPolicy int,
	serverName string, shutdown *utils.SyncedChan) (err error) {
	config, err := loadTLSConfig(serverCrt, serverKey, caCert, serverPolicy, serverName)
	if err != nil {
		shutdown.CloseOnce()
		return
	}
	c.rpcJSONlTLS, err = tls.Listen(utils.TCP, addr, config)
	if err != nil {
		log.Println(fmt.Sprintf("Error: %s when listening", err))
		shutdown.CloseOnce()
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS %s TLS server at <%s>.", utils.JSONCaps, addr))

	return acceptRPC(shutdown, c.rpcServer, c.rpcJSONlTLS, utils.JSONCaps, func(conn conn) birpc.ServerCodec {
		return newCapsGOBCodec(conn, c.caps, c.anz)
	})
}

func (c *CommonListenerS) ServeHTTPS(addr, serverCrt, serverKey, caCert string, serverPolicy int,
	serverName, jsonRPCURL, wsRPCURL, pprofPath string, useBasicAuth bool, userList map[string]string,
	shutdown *utils.SyncedChan) {
	c.mu.Lock()
	c.httpEnabled = c.httpEnabled || jsonRPCURL != "" || wsRPCURL != "" || pprofPath != ""
	enabled := c.httpEnabled
	c.mu.Unlock()
	if !enabled {
		return
	}
	if jsonRPCURL != "" {
		utils.Logger.Info("<HTTPS> enabling handler for JSON-RPC")
		if useBasicAuth {
			c.httpsMux.HandleFunc(jsonRPCURL, use(c.handleRequest, basicAuth(userList)))
		} else {
			c.httpsMux.HandleFunc(jsonRPCURL, c.handleRequest)
		}
	}
	if wsRPCURL != "" {
		utils.Logger.Info("<HTTPS> enabling handler for WebSocket connections")
		wsHandler := websocket.Handler(c.handleWebSocket)
		if useBasicAuth {
			c.httpsMux.HandleFunc(wsRPCURL, use(wsHandler.ServeHTTP, basicAuth(userList)))
		} else {
			c.httpsMux.Handle(wsRPCURL, wsHandler)
		}
	}
	if pprofPath != "" {
		if !strings.HasSuffix(pprofPath, "/") {
			pprofPath += "/"
		}
		utils.Logger.Info(fmt.Sprintf("<HTTPS> profiling endpoints registered at %q", pprofPath))
		if useBasicAuth {
			c.httpsMux.HandleFunc(pprofPath, use(pprof.Index, basicAuth(userList)))
			c.httpsMux.HandleFunc(pprofPath+"cmdline", use(pprof.Cmdline, basicAuth(userList)))
			c.httpsMux.HandleFunc(pprofPath+"profile", use(pprof.Profile, basicAuth(userList)))
			c.httpsMux.HandleFunc(pprofPath+"symbol", use(pprof.Symbol, basicAuth(userList)))
			c.httpsMux.HandleFunc(pprofPath+"trace", use(pprof.Trace, basicAuth(userList)))
		} else {
			c.httpsMux.HandleFunc(pprofPath, pprof.Index)
			c.httpsMux.HandleFunc(pprofPath+"cmdline", pprof.Cmdline)
			c.httpsMux.HandleFunc(pprofPath+"profile", pprof.Profile)
			c.httpsMux.HandleFunc(pprofPath+"symbol", pprof.Symbol)
			c.httpsMux.HandleFunc(pprofPath+"trace", pprof.Trace)
		}
	}
	if useBasicAuth {
		utils.Logger.Info("<HTTPS> enabling basic auth")
	}
	config, err := loadTLSConfig(serverCrt, serverKey, caCert, serverPolicy, serverName)
	if err != nil {
		shutdown.CloseOnce()
		return
	}
	c.httpsServer.Addr = addr
	c.httpsServer.TLSConfig = config
	utils.Logger.Info(fmt.Sprintf("<HTTPS> start listening at <%s>", addr))

	if err := c.httpsServer.ListenAndServeTLS(serverCrt, serverKey); err != nil {
		log.Println(fmt.Sprintf("<HTTPS>Error: %s when listening ", err))
		shutdown.CloseOnce()
	}
}

func (c *CommonListenerS) Stop() {
	if c.rpcJSONl != nil {
		c.rpcJSONl.Close()
	}
	if c.rpcGOBl != nil {
		c.rpcGOBl.Close()
	}
	if c.rpcJSONlTLS != nil {
		c.rpcJSONlTLS.Close()
	}
	if c.rpcGOBlTLS != nil {
		c.rpcGOBlTLS.Close()
	}
	if c.httpServer != nil {
		c.httpServer.Shutdown(context.Background())
	}
	if c.httpsServer != nil {
		c.httpsServer.Shutdown(context.Background())
	}
	c.StopBiRPC()
}

// StopBiRPC stops the go routine create with ServeBiJSON
func (c *CommonListenerS) StopBiRPC() {
	c.stopbiRPCServer <- struct{}{}
	c.birpcSrv = birpc.NewBirpcServer()
}

func (c *CommonListenerS) StartServer(cfg *config.CGRConfig, shutdown *utils.SyncedChan) {
	c.startSrv.Do(func() {
		go c.ServeJSON(cfg.ListenCfg().RPCJSONListen, shutdown)
		go c.ServeGOB(cfg.ListenCfg().RPCGOBListen, shutdown)
		go c.ServeHTTP(
			cfg.ListenCfg().HTTPListen,
			cfg.HTTPCfg().JsonRPCURL,
			cfg.HTTPCfg().WSURL,
			cfg.HTTPCfg().PprofPath,
			cfg.HTTPCfg().UseBasicAuth,
			cfg.HTTPCfg().AuthUsers,
			shutdown,
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
			go c.ServeGOBTLS(
				cfg.ListenCfg().RPCGOBTLSListen,
				cfg.TLSCfg().ServerCerificate,
				cfg.TLSCfg().ServerKey,
				cfg.TLSCfg().CaCertificate,
				cfg.TLSCfg().ServerPolicy,
				cfg.TLSCfg().ServerName,
				shutdown,
			)
		}
		if cfg.ListenCfg().RPCJSONTLSListen != utils.EmptyString {
			go c.ServeJSONTLS(
				cfg.ListenCfg().RPCJSONTLSListen,
				cfg.TLSCfg().ServerCerificate,
				cfg.TLSCfg().ServerKey,
				cfg.TLSCfg().CaCertificate,
				cfg.TLSCfg().ServerPolicy,
				cfg.TLSCfg().ServerName,
				shutdown,
			)
		}
		if cfg.ListenCfg().HTTPTLSListen != utils.EmptyString {
			go c.ServeHTTPS(
				cfg.ListenCfg().HTTPTLSListen,
				cfg.TLSCfg().ServerCerificate,
				cfg.TLSCfg().ServerKey,
				cfg.TLSCfg().CaCertificate,
				cfg.TLSCfg().ServerPolicy,
				cfg.TLSCfg().ServerName,
				cfg.HTTPCfg().JsonRPCURL,
				cfg.HTTPCfg().WSURL,
				cfg.HTTPCfg().PprofPath,
				cfg.HTTPCfg().UseBasicAuth,
				cfg.HTTPCfg().AuthUsers,
				shutdown,
			)
		}
	})
}
