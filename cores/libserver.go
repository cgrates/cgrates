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
	"os"
	"strings"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// rpcRequest represents a RPC request.
// rpcRequest implements the io.ReadWriteCloser interface.
type rpcRequest struct {
	r          io.ReadCloser // holds the JSON formated RPC request
	rw         io.ReadWriter // holds the JSON formated RPC response
	remoteAddr net.Addr
	caps       *engine.Caps
	anzWarpper *analyzers.AnalyzerService
	srv        *birpc.Server
}

// newRPCRequest returns a new rpcRequest.
func newRPCRequest(srv *birpc.Server, r io.ReadCloser, remoteAddr net.Addr, caps *engine.Caps, anz *analyzers.AnalyzerService) *rpcRequest {
	return &rpcRequest{
		r:          r,
		rw:         new(bytes.Buffer),
		remoteAddr: remoteAddr,
		caps:       caps,
		anzWarpper: anz,
		srv:        srv,
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
	r.srv.ServeCodec(newCapsJSONCodec(r, r.caps, r.anzWarpper))
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
			utils.Logger.Crit("Cannot append certificate authority")
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

func acceptRPC(ctx *context.Context, shtdwnEngine context.CancelFunc,
	srv *birpc.Server, l net.Listener, codecName string, newCodec func(conn conn) birpc.ServerCodec) (err error) {
	var errCnt int
	var lastErrorTime time.Time
	for {
		var conn net.Conn
		if conn, err = l.Accept(); err != nil {
			select {
			case <-ctx.Done():
				return
			default:
			}
			utils.Logger.Err(fmt.Sprintf("<CGRServer> %s accept error: <%s>", codecName, err.Error()))
			now := time.Now()
			if now.Sub(lastErrorTime) > 5*time.Second {
				errCnt = 0 // reset error count if last error was more than 5 seconds ago
			}
			lastErrorTime = time.Now()
			errCnt++
			if errCnt > 50 { // Too many errors in short interval, network buffer failure most probably
				shtdwnEngine()
				return
			}
			continue
		}
		go srv.ServeCodec(newCodec(conn))
	}
}

func acceptBiRPC(srv *birpc.BirpcServer, l net.Listener, codecName string, newCodec func(conn conn) birpc.BirpcCodec, stopbiRPCServer chan struct{}) {
	for {
		conn, err := l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") { // if closed by us do not log
				return
			}
			stopbiRPCServer <- struct{}{}
			utils.Logger.Crit(fmt.Sprintf("Stopped Bi%s server beacause %s", codecName, err))
			return // stop if we get Accept error
		}
		go srv.ServeCodec(newCodec(conn))
	}
}

func listenBiRPC(srv *birpc.BirpcServer, addr, codecName string, newCodec func(conn conn) birpc.BirpcCodec, stopbiRPCServer chan struct{}) (lBiRPC net.Listener, err error) {
	if lBiRPC, err = net.Listen(utils.TCP, addr); err != nil {
		log.Printf("ServeBi%s listen error: %s \n", codecName, err)
		return
	}
	utils.Logger.Info(fmt.Sprintf("Starting CGRateS Bi%s server at <%s>", codecName, addr))
	go acceptBiRPC(srv, lBiRPC, codecName, newCodec, stopbiRPCServer)
	return
}
