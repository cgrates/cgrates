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
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewHTTPAgent returns the HTTP Agent
func NewHTTPAgent(cfg *config.CGRConfig,
	connMgr *engine.ConnManager,
	srvIndexer *servmanager.ServiceIndexer) *HTTPAgent {
	return &HTTPAgent{
		cfg:        cfg,
		connMgr:    connMgr,
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// HTTPAgent implements Agent interface
type HTTPAgent struct {
	sync.RWMutex

	cl *commonlisteners.CommonListenerS

	// we can realy stop the HTTPAgent so keep a flag
	// if we registerd the handlers
	started bool

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the sercive start
func (ha *HTTPAgent) Start(_ chan struct{}) (err error) {
	if ha.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.FilterS,
		},
		ha.srvIndexer, ha.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	fs := srvDeps[utils.FilterS].(*FilterService)

	ha.Lock()
	defer ha.Unlock()

	ha.started = true
	for _, agntCfg := range ha.cfg.HTTPAgentCfg() {
		cl.RegisterHttpHandler(agntCfg.URL,
			agents.NewHTTPAgent(ha.connMgr, agntCfg.SessionSConns, fs.FilterS(),
				ha.cfg.GeneralCfg().DefaultTenant, agntCfg.RequestPayload,
				agntCfg.ReplyPayload, agntCfg.RequestProcessors))
	}
	close(ha.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (ha *HTTPAgent) Reload(_ chan struct{}) (err error) {
	return // no reload
}

// Shutdown stops the service
func (ha *HTTPAgent) Shutdown() (err error) {
	ha.Lock()
	ha.started = false
	ha.Unlock()
	return // no shutdown for the momment
}

// IsRunning returns if the service is running
func (ha *HTTPAgent) IsRunning() bool {
	ha.RLock()
	defer ha.RUnlock()
	return ha.started
}

// ServiceName returns the service name
func (ha *HTTPAgent) ServiceName() string {
	return utils.HTTPAgent
}

// ShouldRun returns if the service should be running
func (ha *HTTPAgent) ShouldRun() bool {
	return len(ha.cfg.HTTPAgentCfg()) != 0
}

// StateChan returns signaling channel of specific state
func (ha *HTTPAgent) StateChan(stateID string) chan struct{} {
	return ha.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (ha *HTTPAgent) IntRPCConn() birpc.ClientConnector {
	return ha.intRPCconn
}
