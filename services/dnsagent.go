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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewDNSAgent returns the DNS Agent
func NewDNSAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &DNSAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		connMgr:     connMgr,
		srvDep:      srvDep,
		srvIndexer:  srvIndexer,
	}
}

// DNSAgent implements Agent interface
type DNSAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS

	stopChan chan struct{}

	dns     *agents.DNSAgent
	connMgr *engine.ConnManager
	srvDep  map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the service start
func (dns *DNSAgent) Start(ctx *context.Context, shtDwn context.CancelFunc) (err error) {
	if dns.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, dns.filterSChan); err != nil {
		return
	}

	dns.Lock()
	defer dns.Unlock()
	dns.dns, err = agents.NewDNSAgent(dns.cfg, filterS, dns.connMgr)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
		dns.dns = nil
		return
	}
	dns.stopChan = make(chan struct{})
	go dns.listenAndServe(dns.stopChan, shtDwn)
	return
}

// Reload handles the change of config
func (dns *DNSAgent) Reload(ctx *context.Context, shtDwn context.CancelFunc) (err error) {
	filterS := <-dns.filterSChan
	dns.filterSChan <- filterS

	dns.Lock()
	defer dns.Unlock()

	if dns.dns != nil {
		close(dns.stopChan)
	}

	dns.dns, err = agents.NewDNSAgent(dns.cfg, filterS, dns.connMgr)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
		dns.dns = nil
		return
	}

	dns.dns.Lock()
	defer dns.dns.Unlock()
	dns.stopChan = make(chan struct{})
	go dns.listenAndServe(dns.stopChan, shtDwn)
	return
}

func (dns *DNSAgent) listenAndServe(stopChan chan struct{}, shtDwn context.CancelFunc) (err error) {
	dns.dns.RLock()
	defer dns.dns.RUnlock()
	if err = dns.dns.ListenAndServe(stopChan); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
		shtDwn() // stop the engine here
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
