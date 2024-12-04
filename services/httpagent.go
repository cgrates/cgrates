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
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewHTTPAgent returns the HTTP Agent
func NewHTTPAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	clSChan chan *commonlisteners.CommonListenerS, connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &HTTPAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		clSChan:     clSChan,
		connMgr:     connMgr,
		srvDep:      srvDep,
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// HTTPAgent implements Agent interface
type HTTPAgent struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	filterSChan chan *engine.FilterS

	cl *commonlisteners.CommonListenerS

	// we can realy stop the HTTPAgent so keep a flag
	// if we registerd the handlers
	started bool

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig
	srvDep  map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the sercive start
func (ha *HTTPAgent) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if ha.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	cl := <-ha.clSChan
	ha.clSChan <- cl
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, ha.filterSChan); err != nil {
		return
	}

	ha.Lock()
	ha.started = true
	utils.Logger.Info(fmt.Sprintf("<%s> successfully started HTTPAgent", utils.HTTPAgent))
	for _, agntCfg := range ha.cfg.HTTPAgentCfg() {
		cl.RegisterHttpHandler(agntCfg.URL,
			agents.NewHTTPAgent(ha.connMgr, agntCfg.SessionSConns, filterS,
				ha.cfg.GeneralCfg().DefaultTenant, agntCfg.RequestPayload,
				agntCfg.ReplyPayload, agntCfg.RequestProcessors))
	}
	ha.Unlock()
	close(ha.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (ha *HTTPAgent) Reload(*context.Context, context.CancelFunc) (err error) {
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
