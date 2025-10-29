/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewDNSAgent returns the DNS Agent
func NewDNSAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	shdChan *utils.SyncedChan, connMgr *engine.ConnManager, caps *engine.Caps,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &DNSAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		connMgr:     connMgr,
		caps:        caps,
		srvDep:      srvDep,
	}
}

// DNSAgent implements Agent interface
type DNSAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	shdChan     *utils.SyncedChan

	stopChan chan struct{}

	dns     *agents.DNSAgent
	connMgr *engine.ConnManager
	caps    *engine.Caps
	srvDep  map[string]*sync.WaitGroup
}

// Start should handle the service start
func (dns *DNSAgent) Start() (err error) {
	if dns.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-dns.filterSChan
	dns.filterSChan <- filterS

	dns.Lock()
	defer dns.Unlock()
	dns.dns, err = agents.NewDNSAgent(dns.cfg, filterS, dns.connMgr, dns.caps)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed to initialize agent, error: <%s>", utils.DNSAgent, err.Error()))
		dns.dns = nil
		return
	}
	dns.stopChan = make(chan struct{})
	go dns.listenAndServe(dns.stopChan)
	return
}

// Reload handles the change of config
func (dns *DNSAgent) Reload() (err error) {
	filterS := <-dns.filterSChan
	dns.filterSChan <- filterS

	dns.Lock()
	defer dns.Unlock()

	if dns.dns != nil {
		close(dns.stopChan)
	}

	dns.dns, err = agents.NewDNSAgent(dns.cfg, filterS, dns.connMgr, dns.caps)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
		dns.dns = nil
		return
	}

	dns.dns.Lock()
	defer dns.dns.Unlock()
	dns.stopChan = make(chan struct{})
	go dns.listenAndServe(dns.stopChan)
	return
}

func (dns *DNSAgent) listenAndServe(stopChan chan struct{}) (err error) {
	dns.dns.RLock()
	defer dns.dns.RUnlock()
	if err = dns.dns.ListenAndServe(stopChan); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
		dns.shdChan.CloseOnce() // stop the engine here
	}
	return
}

// Shutdown stops the service
func (dns *DNSAgent) Shutdown() (err error) {
	if dns.dns == nil {
		return
	}
	close(dns.stopChan)
	dns.dns.Lock()
	defer dns.dns.Unlock()
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
