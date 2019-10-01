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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewHTTPAgent returns the HTTP Agent
func NewHTTPAgent() servmanager.Service {
	return new(HTTPAgent)
}

// HTTPAgent implements Agent interface
type HTTPAgent struct {
	sync.RWMutex
	ha *agents.HTTPAgent
}

// Start should handle the sercive start
func (ha *HTTPAgent) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if ha.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	ha.Lock()
	defer ha.Unlock()
	utils.Logger.Info("Starting HTTP agent")
	for _, agntCfg := range sp.GetConfig().HttpAgentCfg() {
		var sS rpcclient.RpcClientConnection
		if sS, err = sp.NewConnection(utils.SessionS, agntCfg.SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> could not connect to %s, error: %s",
				utils.HTTPAgent, utils.SessionS, err.Error()))
			return
		}
		sp.GetServer().RegisterHttpHandler(agntCfg.Url,
			agents.NewHTTPAgent(sS, sp.GetFilterS(), sp.GetConfig().GeneralCfg().DefaultTenant, agntCfg.RequestPayload,
				agntCfg.ReplyPayload, agntCfg.RequestProcessors))
	}
	return
}

// GetIntenternalChan returns the internal connection chanel
func (ha *HTTPAgent) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return nil
}

// Reload handles the change of config
func (ha *HTTPAgent) Reload(sp servmanager.ServiceProvider) (err error) {
	return // no reload
}

// Shutdown stops the service
func (ha *HTTPAgent) Shutdown() (err error) {
	return // no shutdown for the momment
}

// GetRPCInterface returns the interface to register for server
func (ha *HTTPAgent) GetRPCInterface() interface{} {
	return ha.ha
}

// IsRunning returns if the service is running
func (ha *HTTPAgent) IsRunning() bool {
	ha.RLock()
	defer ha.RUnlock()
	return ha != nil && ha.ha != nil
}

// ServiceName returns the service name
func (ha *HTTPAgent) ServiceName() string {
	return utils.HTTPAgent
}
