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
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewDNSAgent returns the DNS Agent
func NewDNSAgent() servmanager.Service {
	return new(DNSAgent)
}

// DNSAgent implements Agent interface
type DNSAgent struct {
	sync.RWMutex
	dns *agents.DNSAgent
}

// Start should handle the sercive start
func (dns *DNSAgent) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if dns.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	dns.Lock()
	defer dns.Unlock()
	// var sSInternal bool
	var sS rpcclient.RpcClientConnection
	utils.Logger.Info(fmt.Sprintf("starting %s service", utils.DNSAgent))
	if !sp.GetConfig().DispatcherSCfg().Enabled && sp.GetConfig().DNSAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		// sSInternal = true
		srvSessionS, has := sp.GetService(utils.SessionS)
		if !has {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
				utils.DNSAgent, utils.SessionS))
			return utils.ErrNotFound
		}
		sSIntConn := <-srvSessionS.GetIntenternalChan()
		srvSessionS.GetIntenternalChan() <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = sp.GetConnection(utils.SessionS, sp.GetConfig().DNSAgentCfg().SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DNSAgent, utils.SessionS, err.Error()))
			return
		}
	}
	dns.dns, err = agents.NewDNSAgent(sp.GetConfig(), sp.GetFilterS(), sS)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
		return
	}
	// if sSInternal { // bidirectional client backwards connection
	// 	sS.(*utils.BiRPCInternalClient).SetClientConn(da)
	// 	var rply string
	// 	if err := sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
	// 		utils.EmptyString, &rply); err != nil {
	// 		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
	// 			utils.DNSAgent, utils.SessionS, err.Error()))
	// 		exitChan <- true
	// 		return
	// 	}
	// }
	go func() {
		if err = dns.dns.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
			sp.GetExitChan() <- true // stop the engine here
		}
	}()
	return
}

// GetIntenternalChan returns the internal connection chanel
// no chanel for DNSAgent
func (dns *DNSAgent) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return nil
}

// Reload handles the change of config
func (dns *DNSAgent) Reload(sp servmanager.ServiceProvider) (err error) {
	var sS rpcclient.RpcClientConnection
	if !sp.GetConfig().DispatcherSCfg().Enabled && sp.GetConfig().DNSAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		// sSInternal = true
		srvSessionS, has := sp.GetService(utils.SessionS)
		if !has {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
				utils.DNSAgent, utils.SessionS))
			return utils.ErrNotFound
		}
		sSIntConn := <-srvSessionS.GetIntenternalChan()
		srvSessionS.GetIntenternalChan() <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = sp.GetConnection(utils.SessionS, sp.GetConfig().DNSAgentCfg().SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DNSAgent, utils.SessionS, err.Error()))
			return
		}
	}
	if err = dns.Shutdown(); err != nil {
		return
	}
	dns.Lock()
	defer dns.Unlock()
	dns.dns.SetSessionSConnection(sS)
	if err = dns.dns.Reload(); err != nil {
		return
	}
	go func() {
		if err := dns.dns.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
			sp.GetExitChan() <- true // stop the engine here
		}
	}()
	return
}

// Shutdown stops the service
func (dns *DNSAgent) Shutdown() (err error) {
	dns.Lock()
	defer dns.Unlock()
	if err = dns.dns.Shutdown(); err != nil {
		return
	}
	dns.dns = nil
	return
}

// GetRPCInterface returns the interface to register for server
func (dns *DNSAgent) GetRPCInterface() interface{} {
	return dns.dns
}

// IsRunning returns if the service is running
func (dns *DNSAgent) IsRunning() bool {
	dns.RLock()
	defer dns.RUnlock()
	return dns != nil && dns.dns != nil
}

// ServiceName returns the service name
func (dns *DNSAgent) ServiceName() string {
	return utils.DNSAgent
}
