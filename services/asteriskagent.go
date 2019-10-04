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
	"sync"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewAsteriskAgent returns the Asterisk Agent
func NewAsteriskAgent(cfg *config.CGRConfig, sSChan,
	dispatcherChan chan rpcclient.RpcClientConnection,
	exitChan chan bool) servmanager.Service {
	return &AsteriskAgent{
		cfg:            cfg,
		sSChan:         sSChan,
		dispatcherChan: dispatcherChan,
		exitChan:       exitChan,
	}
}

// AsteriskAgent implements Agent interface
type AsteriskAgent struct {
	sync.RWMutex
	cfg            *config.CGRConfig
	sSChan         chan rpcclient.RpcClientConnection
	dispatcherChan chan rpcclient.RpcClientConnection
	exitChan       chan bool

	smas []*agents.AsteriskAgent
}

// Start should handle the sercive start
func (ast *AsteriskAgent) Start() (err error) {
	if ast.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	ast.Lock()
	defer ast.Unlock()
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	utils.Logger.Info("Starting Asterisk agent")
	if !ast.cfg.DispatcherSCfg().Enabled && ast.cfg.AsteriskAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		sSIntConn := <-ast.sSChan
		ast.sSChan <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = NewConnection(ast.cfg, ast.sSChan, ast.dispatcherChan, ast.cfg.AsteriskAgentCfg().SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.AsteriskAgent, utils.SessionS, err.Error()))
			return
		}
	}

	listenAndServe := func(sma *agents.AsteriskAgent, exitChan chan bool) {
		if err = sma.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> runtime error: %s!", utils.AsteriskAgent, err))
		}
		exitChan <- true
	}
	ast.smas = make([]*agents.AsteriskAgent, len(ast.cfg.AsteriskAgentCfg().AsteriskConns))
	for connIdx := range ast.cfg.AsteriskAgentCfg().AsteriskConns { // Instantiate connections towards asterisk servers
		if ast.smas[connIdx], err = agents.NewAsteriskAgent(ast.cfg, connIdx, sS); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.AsteriskAgent, err))
			return
		}
		if sSInternal { // bidirectional client backwards connection
			sS.(*utils.BiRPCInternalClient).SetClientConn(ast.smas[connIdx])
			var rply string
			if err = sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
				utils.EmptyString, &rply); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
					utils.AsteriskAgent, utils.SessionS, err.Error()))
				return
			}
		}
		go listenAndServe(ast.smas[connIdx], ast.exitChan)
	}
	return
}

// GetIntenternalChan returns the internal connection chanel
func (ast *AsteriskAgent) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return nil
}

// Reload handles the change of config
func (ast *AsteriskAgent) Reload() (err error) {
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	if !ast.cfg.DispatcherSCfg().Enabled && ast.cfg.AsteriskAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		sSIntConn := <-ast.sSChan
		ast.sSChan <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = NewConnection(ast.cfg, ast.sSChan, ast.dispatcherChan, ast.cfg.AsteriskAgentCfg().SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.AsteriskAgent, utils.SessionS, err.Error()))
			return
		}
	}
	ast.Lock()
	defer ast.Unlock()
	for _, conn := range ast.smas {
		conn.SetSessionSConnection(sS)
		if sSInternal { // bidirectional client backwards connection
			sS.(*utils.BiRPCInternalClient).SetClientConn(conn)
			var rply string
			if err = sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
				utils.EmptyString, &rply); err != nil {
				utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
					utils.AsteriskAgent, utils.SessionS, err.Error()))
				return
			}
		}
	}
	return //partial reload
}

// Shutdown stops the service
func (ast *AsteriskAgent) Shutdown() (err error) {
	return // no shutdown for the momment
}

// IsRunning returns if the service is running
func (ast *AsteriskAgent) IsRunning() bool {
	ast.RLock()
	defer ast.RUnlock()
	return ast != nil && ast.smas != nil
}

// ServiceName returns the service name
func (ast *AsteriskAgent) ServiceName() string {
	return utils.AsteriskAgent
}

// ShouldRun returns if the service should be running
func (ast *AsteriskAgent) ShouldRun() bool {
	return ast.cfg.AsteriskAgentCfg().Enabled
}
