// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	exitChan chan bool, connMgr *engine.ConnManager) servmanager.Service {
	return &DNSAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		exitChan:    exitChan,
		connMgr:     connMgr,
	}
}

// DNSAgent implements Agent interface
type DNSAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	exitChan    chan bool

	dns     *agents.DNSAgent
	connMgr *engine.ConnManager

	oldListen string
}

// Start should handle the sercive start
func (dns *DNSAgent) Start() (err error) {
	if dns.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-dns.filterSChan
	dns.filterSChan <- filterS

	dns.Lock()
	defer dns.Unlock()
	dns.oldListen = dns.cfg.DNSAgentCfg().Listen
	dns.dns, err = agents.NewDNSAgent(dns.cfg, filterS, dns.connMgr)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
		return
	}
	go func() {
		if err = dns.dns.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
			dns.exitChan <- true // stop the engine here
		}
	}()
	return
}

// Reload handles the change of config
func (dns *DNSAgent) Reload() (err error) {
	if dns.oldListen == dns.cfg.DNSAgentCfg().Listen {
		return
	}
	if err = dns.Shutdown(); err != nil {
		return
	}
	dns.Lock()
	dns.oldListen = dns.cfg.DNSAgentCfg().Listen
	defer dns.Unlock()
	if err = dns.dns.Reload(); err != nil {
		return
	}
	go func() {
		if err := dns.dns.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.DNSAgent, err.Error()))
			dns.exitChan <- true // stop the engine here
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

// ShouldRun returns if the service should be running
func (dns *DNSAgent) ShouldRun() bool {
	return dns.cfg.DNSAgentCfg().Enabled
}
