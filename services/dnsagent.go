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

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewDNSAgent returns the DNS Agent
func NewDNSAgent(cfg *config.CGRConfig,
	connMgr *engine.ConnManager,
	srvIndexer *servmanager.ServiceIndexer) *DNSAgent {
	return &DNSAgent{
		cfg:        cfg,
		connMgr:    connMgr,
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// DNSAgent implements Agent interface
type DNSAgent struct {
	sync.RWMutex
	cfg *config.CGRConfig

	stopChan chan struct{}

	dns     *agents.DNSAgent
	connMgr *engine.ConnManager

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the service start
func (dns *DNSAgent) Start(shutdown chan struct{}) (err error) {
	if dns.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	fs := dns.srvIndexer.GetService(utils.FilterS).(*FilterService)
	if utils.StructChanTimeout(fs.StateChan(utils.StateServiceUP), dns.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.DNSAgent, utils.FilterS, utils.StateServiceUP)
	}

	dns.Lock()
	defer dns.Unlock()
	dns.dns, err = agents.NewDNSAgent(dns.cfg, fs.FilterS(), dns.connMgr)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
		dns.dns = nil
		return
	}
	dns.stopChan = make(chan struct{})
	go dns.listenAndServe(dns.stopChan, shutdown)
	close(dns.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (dns *DNSAgent) Reload(shutdown chan struct{}) (err error) {
	fs := dns.srvIndexer.GetService(utils.FilterS).(*FilterService)
	if utils.StructChanTimeout(fs.StateChan(utils.StateServiceUP), dns.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.DNSAgent, utils.FilterS, utils.StateServiceUP)
	}

	dns.Lock()
	defer dns.Unlock()

	if dns.dns != nil {
		close(dns.stopChan)
	}

	dns.dns, err = agents.NewDNSAgent(dns.cfg, fs.FilterS(), dns.connMgr)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
		dns.dns = nil
		return
	}

	dns.dns.Lock()
	defer dns.dns.Unlock()
	dns.stopChan = make(chan struct{})
	go dns.listenAndServe(dns.stopChan, shutdown)
	return
}

func (dns *DNSAgent) listenAndServe(stopChan chan struct{}, shutdown chan struct{}) (err error) {
	dns.dns.RLock()
	defer dns.dns.RUnlock()
	if err = dns.dns.ListenAndServe(stopChan); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
		close(shutdown) // stop the engine here
	}
	return
}

// Shutdown stops the service
func (dns *DNSAgent) Shutdown() (err error) {
	if dns.dns == nil {
		return
	}
	close(dns.stopChan)
	dns.Lock()
	defer dns.Unlock()
	dns.dns = nil
	return
}

// IsRunning returns if the service is running
func (dns *DNSAgent) IsRunning() bool {
	dns.RLock()
	defer dns.RUnlock()
	return dns.dns != nil
}

// ServiceName returns the service name
func (dns *DNSAgent) ServiceName() string {
	return utils.DNSAgent
}

// ShouldRun returns if the service should be running
func (dns *DNSAgent) ShouldRun() bool {
	return dns.cfg.DNSAgentCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (dns *DNSAgent) StateChan(stateID string) chan struct{} {
	return dns.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (dns *DNSAgent) IntRPCConn() birpc.ClientConnector {
	return dns.intRPCconn
}
