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

package agents

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/sm"
)

func NewDiameterAgent(cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	sessionS rpcclient.RpcClientConnection) (*DiameterAgent, error) {
	if sessionS != nil && reflect.ValueOf(sessionS).IsNil() {
		sessionS = nil
	}
	da := &DiameterAgent{
		cgrCfg: cgrCfg, filterS: filterS,
		sessionS: sessionS, connMux: new(sync.Mutex)}
	dictsPath := cgrCfg.DiameterAgentCfg().DictionariesPath
	if len(dictsPath) != 0 {
		if err := loadDictionaries(dictsPath, utils.DiameterAgent); err != nil {
			return nil, err
		}
	}
	return da, nil
}

type DiameterAgent struct {
	cgrCfg   *config.CGRConfig
	filterS  *engine.FilterS
	sessionS rpcclient.RpcClientConnection // Connection towards CGR-SessionS component
	connMux  *sync.Mutex                   // Protect connection for read/write
}

// ListenAndServe is called when DiameterAgent is started, usually from within cmd/cgr-engine
func (da *DiameterAgent) ListenAndServe() error {
	return diam.ListenAndServe(da.cgrCfg.DiameterAgentCfg().Listen, da.handlers(), nil)
}

// Creates the message handlers
func (da *DiameterAgent) handlers() diam.Handler {
	settings := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(da.cgrCfg.DiameterAgentCfg().OriginHost),
		OriginRealm:      datatype.DiameterIdentity(da.cgrCfg.DiameterAgentCfg().OriginRealm),
		VendorID:         datatype.Unsigned32(da.cgrCfg.DiameterAgentCfg().VendorId),
		ProductName:      datatype.UTF8String(da.cgrCfg.DiameterAgentCfg().ProductName),
		FirmwareRevision: datatype.Unsigned32(utils.DIAMETER_FIRMWARE_REVISION),
	}
	dSM := sm.New(settings)
	dSM.HandleFunc("ALL", da.handleMessage) // route all commands to one dispatcher
	go func() {
		for err := range dSM.ErrorReports() {
			utils.Logger.Err(fmt.Sprintf("<%s> sm error: %v", utils.DiameterAgent, err))
		}
	}()
	return dSM
}

// handleALL is the handler of all messages coming in via Diameter
func (da *DiameterAgent) handleMessage(c diam.Conn, m *diam.Message) {
	var processed bool
	var rply *diam.Message
	var err error
	for _, reqProcessor := range da.cgrCfg.DiameterAgentCfg().RequestProcessors {
		var lclProcessed bool
		lclProcessed, err = da.processRequest(
			reqProcessor,
			newAgentRequest(
				newDADataProvider(m),
				reqProcessor.Tenant, da.cgrCfg.DefaultTenant, da.filterS),
			rply)
		if lclProcessed {
			processed = lclProcessed
		}
		if err != nil ||
			(lclProcessed && !reqProcessor.ContinueOnSuccess) {
			break
		}
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing message: %s",
				utils.DiameterAgent, err.Error(), m))
		return // FixMe with returning some error on HTTP level
	} else if !processed {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> no request processor enabled, ignoring message %s from %s",
				utils.DiameterAgent, m, c.RemoteAddr()))
		return
	}
}

func (da *DiameterAgent) processRequest(reqProcessor *config.DARequestProcessor,
	agReq *AgentRequest, rply *diam.Message) (processed bool, err error) {
	return
}
