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

package services

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"

	"github.com/cgrates/cgrates/analyzers"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	// used to build the connector for analyzers
	intAnzConn = func(c rpcclient.ClientConnector, to string) rpcclient.ClientConnector { return c }
)

// NewAnalyzerService returns the Analyzer Service
func NewAnalyzerService(cfg *config.CGRConfig, server *utils.Server, exitChan chan bool,
	internalAnalyzerSChan chan rpcclient.ClientConnector) *AnalyzerService {
	return &AnalyzerService{
		connChan: internalAnalyzerSChan,
		cfg:      cfg,
		server:   server,
		exitChan: exitChan,
	}
}

// AnalyzerService implements Service interface
type AnalyzerService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	server   *utils.Server
	exitChan chan bool

	anz      *analyzers.AnalyzerService
	rpc      *v1.AnalyzerSv1
	connChan chan rpcclient.ClientConnector
}

// Start should handle the sercive start
func (anz *AnalyzerService) Start() (err error) {
	if anz.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	if anz.anz, err = analyzers.NewAnalyzerService(anz.cfg); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.AnalyzerS, err.Error()))
		return
	}
	go func() {
		if err := anz.anz.ListenAndServe(anz.exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.AnalyzerS, err.Error()))
		}
		anz.anz.Shutdown()
		anz.exitChan <- true
		return
	}()
	intAnzConn = func(c rpcclient.ClientConnector, to string) rpcclient.ClientConnector {
		return anz.anz.NewAnalyzerConnector(c, utils.MetaInternal, utils.EmptyString, to)
	}
	utils.AnalizerWraperFunc = func(conn rpc.ServerCodec, enc string, from, to net.Addr) rpc.ServerCodec {
		fromstr := ""
		if from != nil {
			fromstr = from.String()
		}
		tostr := ""
		if to != nil {
			tostr = to.String()
		}
		return anz.anz.NewServerCodec(conn, enc, fromstr, tostr)
	}
	anz.rpc = v1.NewAnalyzerSv1(anz.anz)
	if !anz.cfg.DispatcherSCfg().Enabled {
		anz.server.RpcRegister(anz.rpc)
	}
	anz.connChan <- anz.rpc

	return
}

// Reload handles the change of config
func (anz *AnalyzerService) Reload() (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (anz *AnalyzerService) Shutdown() (err error) {
	anz.Lock()
	anz.anz.Shutdown()
	anz.anz = nil
	anz.rpc = nil
	<-anz.connChan
	anz.Unlock()
	return
}

// IsRunning returns if the service is running
func (anz *AnalyzerService) IsRunning() bool {
	anz.RLock()
	defer anz.RUnlock()
	return anz != nil && anz.anz != nil
}

// ServiceName returns the service name
func (anz *AnalyzerService) ServiceName() string {
	return utils.AnalyzerS
}

// ShouldRun returns if the service should be running
func (anz *AnalyzerService) ShouldRun() bool {
	return anz.cfg.AnalyzerSCfg().Enabled
}

func (anz *AnalyzerService) GetAnalyzerS() *analyzers.AnalyzerService {
	return anz.anz
}
