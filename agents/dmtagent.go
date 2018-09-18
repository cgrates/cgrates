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
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/sm"
)

func NewDiameterAgent(cgrCfg *config.CGRConfig,
	sessionS, thdS rpcclient.RpcClientConnection) (*DiameterAgent, error) {
	if sessionS != nil && reflect.ValueOf(sessionS).IsNil() {
		sessionS = nil
	}
	if thdS != nil && reflect.ValueOf(thdS).IsNil() {
		thdS = nil
	}
	da := &DiameterAgent{
		cgrCfg: cgrCfg, sessionS: sessionS,
		thdS: thdS, connMux: new(sync.Mutex)}
	dictsDir := cgrCfg.DiameterAgentCfg().DictionariesDir
	if len(dictsDir) != 0 {
		if err := loadDictionaries(dictsDir, utils.DiameterAgent); err != nil {
			return nil, err
		}
	}
	return da, nil
}

type DiameterAgent struct {
	cgrCfg   *config.CGRConfig
	sessionS rpcclient.RpcClientConnection // Connection towards CGR-SessionS component
	thdS     rpcclient.RpcClientConnection // Connection towards CGR-ThresholdS component
	connMux  *sync.Mutex                   // Protect connection for read/write
}

// ListenAndServe is called when DiameterAgent is started, usually from within cmd/cgr-engine
func (self *DiameterAgent) ListenAndServe() error {
	return diam.ListenAndServe(self.cgrCfg.DiameterAgentCfg().Listen, self.handlers(), nil)
}

// Creates the message handlers
func (self *DiameterAgent) handlers() diam.Handler {
	settings := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(self.cgrCfg.DiameterAgentCfg().OriginHost),
		OriginRealm:      datatype.DiameterIdentity(self.cgrCfg.DiameterAgentCfg().OriginRealm),
		VendorID:         datatype.Unsigned32(self.cgrCfg.DiameterAgentCfg().VendorId),
		ProductName:      datatype.UTF8String(self.cgrCfg.DiameterAgentCfg().ProductName),
		FirmwareRevision: datatype.Unsigned32(utils.DIAMETER_FIRMWARE_REVISION),
	}
	dSM := sm.New(settings)
	dSM.HandleFunc("ALL", self.handleALL) // route all commands to one dispatcher
	go func() {
		for err := range dSM.ErrorReports() {
			utils.Logger.Err(fmt.Sprintf("<%s> sm error: %v", utils.DiameterAgent, err))
		}
	}()
	return dSM
}

func (self *DiameterAgent) handleALL(c diam.Conn, m *diam.Message) {
	utils.Logger.Warning(fmt.Sprintf("<%s> received unexpected message from %s:\n%s", utils.DiameterAgent, c.RemoteAddr(), m))
}
