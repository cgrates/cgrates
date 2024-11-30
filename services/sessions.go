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

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// NewSessionService returns the Session Service
func NewSessionService(cfg *config.CGRConfig, dm *DataDBService, filterSChan chan *engine.FilterS,
	clSChan chan *commonlisteners.CommonListenerS, internalChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anzChan chan *AnalyzerService,
	srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &SessionService{
		connChan:    internalChan,
		cfg:         cfg,
		dm:          dm,
		filterSChan: filterSChan,
		clSChan:     clSChan,
		connMgr:     connMgr,
		anzChan:     anzChan,
		srvDep:      srvDep,
		srvIndexer:  srvIndexer,
	}
}

// SessionService implements Service interface
type SessionService struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	dm          *DataDBService
	anzChan     chan *AnalyzerService
	filterSChan chan *engine.FilterS

	sm *sessions.SessionS
	cl *commonlisteners.CommonListenerS

	bircpEnabled bool // to stop birpc server if needed
	stopChan     chan struct{}
	connChan     chan birpc.ClientConnector
	connMgr      *engine.ConnManager
	cfg          *config.CGRConfig
	srvDep       map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the service start
func (smg *SessionService) Start(ctx *context.Context, shtDw context.CancelFunc) (err error) {
	if smg.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	smg.cl = <-smg.clSChan
	smg.clSChan <- smg.cl
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, smg.filterSChan); err != nil {
		return
	}
	var datadb *engine.DataManager
	if datadb, err = smg.dm.WaitForDM(ctx); err != nil {
		return
	}
	anz := <-smg.anzChan
	smg.anzChan <- anz

	smg.Lock()
	defer smg.Unlock()

	smg.sm = sessions.NewSessionS(smg.cfg, datadb, filterS, smg.connMgr)
	//start sync session in a separate goroutine
	smg.stopChan = make(chan struct{})
	go smg.sm.ListenAndServe(smg.stopChan)
	// Pass internal connection via BiRPCClient

	// Register RPC handler
	srv, _ := engine.NewServiceWithName(smg.sm, utils.SessionS, true) // methods with multiple options
	// srv, _ := birpc.NewService(apis.NewSessionSv1(smg.sm), utils.EmptyString, false) // methods with multiple options
	if !smg.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			smg.cl.RpcRegister(s)
		}
	}
	smg.connChan <- anz.GetInternalCodec(srv, utils.SessionS)
	// Register BiRpc handlers
	if smg.cfg.SessionSCfg().ListenBijson != utils.EmptyString {
		smg.bircpEnabled = true
		for n, s := range srv {
			smg.cl.BiRPCRegisterName(n, s)
		}
		// run this in it's own goroutine
		go smg.start(shtDw)
	}
	return
}

func (smg *SessionService) start(shtDw context.CancelFunc) (err error) {
	if err := smg.cl.ServeBiRPC(smg.cfg.SessionSCfg().ListenBijson,
		smg.cfg.SessionSCfg().ListenBigob, smg.sm.OnBiJSONConnect, smg.sm.OnBiJSONDisconnect); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> serve BiRPC error: %s!", utils.SessionS, err))
		smg.Lock()
		smg.bircpEnabled = false
		smg.Unlock()
		shtDw()
	}
	return
}

// Reload handles the change of config
func (smg *SessionService) Reload(*context.Context, context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (smg *SessionService) Shutdown() (err error) {
	smg.Lock()
	defer smg.Unlock()
	close(smg.stopChan)
	if err = smg.sm.Shutdown(); err != nil {
		return
	}
	if smg.bircpEnabled {
		smg.cl.StopBiRPC()
		smg.bircpEnabled = false
	}
	smg.sm = nil
	<-smg.connChan
	smg.cl.RpcUnregisterName(utils.SessionSv1)
	// smg.server.BiRPCUnregisterName(utils.SessionSv1)
	return
}

// IsRunning returns if the service is running
func (smg *SessionService) IsRunning() bool {
	smg.RLock()
	defer smg.RUnlock()
	return smg.sm != nil
}

// ServiceName returns the service name
func (smg *SessionService) ServiceName() string {
	return utils.SessionS
}

// ShouldRun returns if the service should be running
func (smg *SessionService) ShouldRun() bool {
	return smg.cfg.SessionSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (smg *SessionService) StateChan(stateID string) chan struct{} {
	return smg.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (smg *SessionService) IntRPCConn() birpc.ClientConnector {
	return smg.intRPCconn
}
